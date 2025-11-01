# 修复边类型变更检测问题

## 问题描述

修改边的类型（如从 "default" 改为 "smoothstep"）后保存，控制台显示：

```
[10:47:27] [工作流保存] ⚠️ 边 591285685432878102 没有实际变化，跳过更新
[10:47:27] [工作流保存] ✅ 边保存完成
[10:47:27] [工作流保存] 📊 保存统计: 节点: +0 ~0 -0 | 边: +0 ~0 -0 | 共更新 0 个字段
```

**问题**：边的类型确实改变了，但系统没有检测到变化，导致更新被跳过。

## 问题分析

### 边的类型系统

Vue Flow 中的边有两种类型概念：

1. **Vue Flow 边类型**（`edge.type`）
   - 控制边的视觉样式
   - 可选值：`"default"`, `"smoothstep"`, `"step"`, `"straight"`, `"simplebezier"` 等
   - 存储位置：`edge.type`
   - 用途：前端渲染

2. **后端业务类型**（`type` 字段）
   - 控制边的业务逻辑
   - 可选值：`"default"`, `"branch"`, `"parallel"`
   - 计算方式：根据 `edge.data.isParallelChild` 和 `edge.data.branchName` 判断
   - 用途：后端业务逻辑

### 数据流

**创建边时**：
```typescript
const edgeData = {
  type: backendType,  // 后端业务类型
  data: {
    ...edge.data,
    vueFlowType: edge.type  // Vue Flow 边类型存储在 data 中
  }
};
```

**加载边时**：
```typescript
// 后端返回的边数据
{
  type: "default",  // 后端业务类型
  data: {
    vueFlowType: "smoothstep"  // Vue Flow 边类型
  }
}

// 前端恢复
edge.type = edgeData.data.vueFlowType || "default";
```

### 问题根源

**Hash 检测**：
```typescript
const getEdgeHash = (edge: Edge): string => {
  const businessData = {
    source: edge.source,
    target: edge.target,
    sourceHandle: edge.sourceHandle,
    targetHandle: edge.targetHandle,
    type: edge.type,  // ✅ 包含 Vue Flow 类型
    label: edge.label,
    animated: edge.animated,
    style: edge.style,
    data: edge.data
  };
  return hashString(JSON.stringify(businessData));
};
```

**字段级别 Diff**（修复前）：
```typescript
const getEdgeFieldChanges = (currentEdge, snapshotEdge) => {
  // ... 比较其他字段

  // ❌ 只比较 data 对象，但 JSON.stringify 可能因为属性顺序不同而判断为相同
  if (JSON.stringify(currentEdge.data) !== JSON.stringify(snapshotEdge.data)) {
    changes.data = {
      ...currentEdge.data,
      vueFlowType: currentEdge.type
    };
    changedFields.push("data");
    hasChanges = true;
  }

  // ❌ 没有单独比较 edge.type
};
```

**问题**：
1. Hash 检测到了变化（因为 `edge.type` 改变了）
2. 边被加入到 `edgesToUpdate` 数组
3. 但 `getEdgeFieldChanges` 没有单独比较 `edge.type`
4. `JSON.stringify(currentEdge.data)` 可能因为对象属性顺序、引用等原因判断为"相同"
5. 函数返回 `null`，更新被跳过

### 时序图

```
用户修改边类型: "default" -> "smoothstep"
    ↓
edge.type = "smoothstep"
    ↓
保存工作流
    ↓
计算 hash: 包含 edge.type ✅
    ↓
hash 不同，加入 edgesToUpdate ✅
    ↓
调用 getEdgeFieldChanges
    ↓
比较 edge.type? ❌ 没有
    ↓
比较 JSON.stringify(data)? ❌ 判断为相同
    ↓
返回 null
    ↓
跳过更新 ❌
```

## 修复方案

### 核心思路

**在 `getEdgeFieldChanges` 中单独比较 Vue Flow 边类型**

### 实现细节

**文件**：`src/views/test/composables/useWorkflowApplication.ts`

**修改前**：
```typescript
const getEdgeFieldChanges = (currentEdge, snapshotEdge) => {
  // ... 比较其他字段

  if (
    JSON.stringify(currentEdge.data) !== JSON.stringify(snapshotEdge.data)
  ) {
    changes.data = {
      ...currentEdge.data,
      vueFlowType: currentEdge.type
    };
    changedFields.push("data");
    hasChanges = true;
  }

  return hasChanges ? { changedFields, changes } : null;
};
```

**修改后**：
```typescript
const getEdgeFieldChanges = (currentEdge, snapshotEdge) => {
  // ... 比较其他字段

  // 比较 Vue Flow 的边类型（存储在 data.vueFlowType 中）
  const currentVueFlowType = currentEdge.type;
  const snapshotVueFlowType = snapshotEdge.type;
  
  if (currentVueFlowType !== snapshotVueFlowType) {
    // Vue Flow 类型变化，需要更新 data
    changes.data = {
      ...currentEdge.data,
      vueFlowType: currentEdge.type
    };
    changedFields.push("data.vueFlowType");
    hasChanges = true;
  } else if (
    JSON.stringify(currentEdge.data) !== JSON.stringify(snapshotEdge.data)
  ) {
    // 其他 data 字段变化
    changes.data = {
      ...currentEdge.data,
      vueFlowType: currentEdge.type
    };
    changedFields.push("data");
    hasChanges = true;
  }

  return hasChanges ? { changedFields, changes } : null;
};
```

**改进点**：
- ✅ 单独比较 `edge.type`（Vue Flow 边类型）
- ✅ 类型变化时，标记为 `data.vueFlowType` 变更
- ✅ 其他 data 字段变化时，标记为 `data` 变更
- ✅ 确保 `vueFlowType` 总是包含在 `changes.data` 中

## 修复效果

### 场景：修改边的类型

**修改前**：
```
用户操作：将边类型从 "default" 改为 "smoothstep"
保存结果：
  [工作流保存] 边 591285685432878102 有变化 (hash: xxx -> yyy)
  [工作流保存] ⚠️ 边 591285685432878102 没有实际变化，跳过更新
  [工作流保存] 📊 保存统计: 边: +0 ~0 -0
```
- ❌ Hash 检测到变化
- ❌ 但字段级别 diff 没有检测到
- ❌ 更新被跳过
- ❌ 边类型没有保存

**修改后**：
```
用户操作：将边类型从 "default" 改为 "smoothstep"
保存结果：
  [工作流保存] 边 591285685432878102 有变化 (hash: xxx -> yyy)
  [工作流保存] 边 591285685432878102 的变更字段: data.vueFlowType
  [工作流保存] ✅ 更新边: 591285685432878102
  [工作流保存] 📊 保存统计: 边: +0 ~1 -0 | 共更新 1 个字段
```
- ✅ Hash 检测到变化
- ✅ 字段级别 diff 检测到 `data.vueFlowType` 变化
- ✅ 更新成功
- ✅ 边类型正确保存

### 场景：修改边的其他属性

**修改边的标签**：
```
[工作流保存] 边 591285685432878102 的变更字段: label
[工作流保存] ✅ 更新边: 591285685432878102
```

**修改边的动画**：
```
[工作流保存] 边 591285685432878102 的变更字段: animated
[工作流保存] ✅ 更新边: 591285685432878102
```

**同时修改类型和标签**：
```
[工作流保存] 边 591285685432878102 的变更字段: data.vueFlowType, label
[工作流保存] ✅ 更新边: 591285685432878102
```

## 技术细节

### 为什么 JSON.stringify 比较不可靠？

**问题 1：属性顺序**
```javascript
const obj1 = { a: 1, b: 2 };
const obj2 = { b: 2, a: 1 };

JSON.stringify(obj1);  // '{"a":1,"b":2}'
JSON.stringify(obj2);  // '{"b":2,"a":1}'

JSON.stringify(obj1) === JSON.stringify(obj2);  // false ❌
```

**问题 2：undefined 值**
```javascript
const obj1 = { a: 1, b: undefined };
const obj2 = { a: 1 };

JSON.stringify(obj1);  // '{"a":1}'
JSON.stringify(obj2);  // '{"a":1}'

JSON.stringify(obj1) === JSON.stringify(obj2);  // true ✅
// 但实际上 obj1 和 obj2 不同
```

**问题 3：函数和 Symbol**
```javascript
const obj1 = { a: 1, fn: () => {} };
const obj2 = { a: 1 };

JSON.stringify(obj1);  // '{"a":1}'
JSON.stringify(obj2);  // '{"a":1}'
```

### 更好的对象比较方法

**方式 1：深度比较库**
```typescript
import isEqual from 'lodash/isEqual';

if (!isEqual(currentEdge.data, snapshotEdge.data)) {
  // 有变化
}
```

**方式 2：单独比较关键字段**（本次采用）
```typescript
// 单独比较重要字段
if (currentEdge.type !== snapshotEdge.type) {
  // Vue Flow 类型变化
}

if (currentEdge.data?.branchName !== snapshotEdge.data?.branchName) {
  // 分支名称变化
}

// 其他字段用 JSON.stringify 作为兜底
if (JSON.stringify(currentEdge.data) !== JSON.stringify(snapshotEdge.data)) {
  // 其他 data 字段变化
}
```

**方式 3：规范化后比较**
```typescript
const normalize = (obj: any) => {
  return JSON.stringify(obj, Object.keys(obj).sort());
};

if (normalize(currentEdge.data) !== normalize(snapshotEdge.data)) {
  // 有变化
}
```

### Vue Flow 边类型列表

**内置类型**：
- `"default"` - 默认直线
- `"straight"` - 直线
- `"step"` - 阶梯线
- `"smoothstep"` - 平滑阶梯线
- `"simplebezier"` - 简单贝塞尔曲线

**自定义类型**：
可以通过 `edgeTypes` 注册自定义边组件。

## 测试建议

### 1. 测试修改边类型

```typescript
// 1. 创建一个工作流，添加两个节点和一条边
// 2. 选中边，修改类型为 "smoothstep"
// 3. 保存工作流
// 4. 检查控制台输出：
//    - 应显示 "边 xxx 的变更字段: data.vueFlowType"
//    - 应显示 "✅ 更新边: xxx"
//    - 保存统计应显示 "边: +0 ~1 -0"
// 5. 刷新页面，检查边类型是否保持为 "smoothstep"
```

### 2. 测试修改边的其他属性

```typescript
// 1. 修改边的标签
// 2. 保存工作流
// 3. 检查控制台输出：
//    - 应显示 "边 xxx 的变更字段: label"
// 4. 刷新页面，检查标签是否保存
```

### 3. 测试同时修改多个属性

```typescript
// 1. 同时修改边的类型、标签、动画
// 2. 保存工作流
// 3. 检查控制台输出：
//    - 应显示 "边 xxx 的变更字段: data.vueFlowType, label, animated"
// 4. 刷新页面，检查所有修改是否保存
```

### 4. 测试边缘情况

```typescript
// 1. 创建边后立即修改类型（不刷新页面）
// 2. 保存工作流
// 3. 检查是否正确保存

// 4. 修改边类型后撤销（改回原类型）
// 5. 保存工作流
// 6. 检查是否跳过更新（因为没有实际变化）
```

## 相关文件

### 修改的文件

- **`src/views/test/composables/useWorkflowApplication.ts`**
  - `getEdgeFieldChanges` 函数（第 385-496 行）
  - 添加 Vue Flow 边类型的单独比较

### 相关文件（未修改）

- **`src/views/test/composables/useWorkflowApplication.ts`**
  - `getEdgeHash` 函数（第 280-293 行）
  - 创建边逻辑（第 1063-1106 行）

## 总结

这次修复解决了边类型变更检测的问题：

✅ **单独比较 Vue Flow 边类型**：不依赖 JSON.stringify  
✅ **准确的变更字段标记**：`data.vueFlowType` vs `data`  
✅ **完整的日志输出**：清楚显示哪些字段变更了  
✅ **正确保存边类型**：用户修改的边类型能正确保存到后端  

**关键要点**：
- 对于重要字段，应该单独比较，不要依赖 JSON.stringify
- JSON.stringify 只适合作为兜底方案，检测未知字段的变化
- 清晰的日志输出有助于调试和理解系统行为

