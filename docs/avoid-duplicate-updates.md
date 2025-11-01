# 避免条件节点重复更新优化

## 问题描述

在移动条件节点位置时，节点被更新了两次：

**第一次更新**（不含 branchNodes）：
```json
{
  "name": "条件检查",
  "nodeKey": "591286479313962006",
  "type": "condition_checker",
  "positionX": 1359.000015258789,
  "positionY": -38.84999084472656,
  "color": "#E6A23C"
  // 没有 branchNodes
}
```

**第二次更新**（包含 branchNodes）：
```json
{
  "name": "条件检查",
  "nodeKey": "591286479313962006",
  "type": "condition_checker",
  "positionX": 1359.000015258789,
  "positionY": -38.84999084472656,
  "color": "#E6A23C",
  "branchNodes": {
    "true": { "name": "true", "condition": "result === true" }
  }
}
```

### 问题分析

**原来的逻辑**：
1. **第一阶段**：更新所有修改的节点（不含 branchNodes）
2. **第二阶段**：更新所有条件节点（包含 branchNodes）

**导致的问题**：
- 已存在的条件节点即使只改了位置，也会被更新两次
- 浪费网络请求和数据库操作
- 降低保存性能

## 优化方案

### 核心思路

**区分新创建和已存在的条件节点**：

1. **新创建的条件节点**：
   - 第一次创建时不包含 branchNodes（因为目标节点可能还没创建）
   - 第二次更新时包含 branchNodes（使用映射后的数据库 ID）

2. **已存在的条件节点**：
   - 第一次更新时就包含 branchNodes（所有节点都已存在，不需要 ID 映射）
   - 不需要第二次更新

### 实现细节

#### 1. 第一阶段：更新已存在的节点

**修改前**：
```typescript
// 更新修改的节点（不包含 branchNodes，将在后面统一更新）
for (const node of nodesToUpdate) {
  const nodeData = {
    name: node.data.label || node.id,
    type: node.type,
    // ... 其他字段
    // 注意：branchNodes 将在所有节点和边保存完成后更新
  };

  await updateWorkflowNode(node.id, nodeData);
}
```

**修改后**：
```typescript
// 更新修改的节点
for (const node of nodesToUpdate) {
  const nodeData: any = {
    name: node.data.label || node.id,
    type: node.type,
    // ... 其他字段
  };

  // 对于条件节点，直接包含 branchNodes（避免二次更新）
  if (node.type === NodeTypeEnum.CONDITION_CHECKER) {
    const branchNodes = calculateBranchNodesFromNode(node, nodeIdMapping);
    if (branchNodes && Object.keys(branchNodes).length > 0) {
      nodeData.branchNodes = branchNodes;
    }
  }

  await updateWorkflowNode(node.id, nodeData);
}
```

**改进点**：
- ✅ 条件节点在第一次更新时就包含 branchNodes
- ✅ 不需要第二次更新
- ✅ 减少 50% 的更新请求

#### 2. 第二阶段：只更新新创建的条件节点

**修改前**：
```typescript
// 更新条件节点的 branchNodes（在所有节点和边都保存完成后）
const conditionNodes = currentNodes.filter(
  n => n.type === NodeTypeEnum.CONDITION_CHECKER
);

for (const node of conditionNodes) {
  // 所有条件节点都会被更新
  const actualNodeId = nodeIdMapping.get(node.id) || node.id;
  const branchNodes = calculateBranchNodesFromNode(node, nodeIdMapping);
  
  if (branchNodes && Object.keys(branchNodes).length > 0) {
    await updateWorkflowNode(actualNodeId, { branchNodes, ... });
  }
}
```

**修改后**：
```typescript
// 更新新创建的条件节点的 branchNodes（在所有节点和边都保存完成后）
// 注意：已存在的条件节点在第一次更新时已经包含了 branchNodes，不需要再次更新
const newConditionNodes = nodesToCreate.filter(
  n => n.type === NodeTypeEnum.CONDITION_CHECKER
);

if (newConditionNodes.length > 0) {
  for (const node of newConditionNodes) {
    // 只更新新创建的条件节点
    const actualNodeId = nodeIdMapping.get(node.id);
    if (!actualNodeId) continue;
    
    const branchNodes = calculateBranchNodesFromNode(node, nodeIdMapping);
    
    if (branchNodes && Object.keys(branchNodes).length > 0) {
      await updateWorkflowNode(actualNodeId, { branchNodes, ... });
    }
  }
}
```

**改进点**：
- ✅ 只更新新创建的条件节点
- ✅ 已存在的条件节点不会被重复更新
- ✅ 日志更清晰

## 优化效果

### 场景 1：移动已存在的条件节点

**优化前**：
```
[工作流保存] 节点 591286479313962006 有变化
✅ 更新节点: 591286479313962006 (不含 branchNodes)
✅ 更新条件节点 591286479313962006 的 branchNodes (含 branchNodes)
```
- 更新次数：**2 次**

**优化后**：
```
[工作流保存] 节点 591286479313962006 有变化
✅ 更新节点: 591286479313962006 (含 branchNodes)
```
- 更新次数：**1 次**
- 性能提升：**50%**

### 场景 2：创建新的条件节点

**优化前**：
```
[工作流保存] 节点 condition-123 是新增节点（临时ID）
✅ 创建节点: condition-123 -> 1001
✅ 更新条件节点 condition-123 (数据库ID: 1001) 的 branchNodes
```
- 操作次数：**2 次**（创建 + 更新）

**优化后**：
```
[工作流保存] 节点 condition-123 是新增节点（临时ID）
✅ 创建节点: condition-123 -> 1001
✅ 更新新创建的条件节点 condition-123 (数据库ID: 1001) 的 branchNodes
```
- 操作次数：**2 次**（创建 + 更新）
- 保持不变（必须两次）

### 场景 3：修改条件节点的分支配置

**优化前**：
```
[工作流保存] 节点 591286479313962006 有变化
[工作流保存] 节点 591286479313962006 的变更字段: data.branchNodes
✅ 更新节点: 591286479313962006 (不含 branchNodes)
✅ 更新条件节点 591286479313962006 的 branchNodes (含 branchNodes)
```
- 更新次数：**2 次**

**优化后**：
```
[工作流保存] 节点 591286479313962006 有变化
[工作流保存] 节点 591286479313962006 的变更字段: data.branchNodes
✅ 更新节点: 591286479313962006 (含 branchNodes)
```
- 更新次数：**1 次**
- 性能提升：**50%**

## 性能对比

### 典型工作流（10 个节点，3 个条件节点）

**优化前**：
- 移动 3 个条件节点的位置
- 更新请求：3 × 2 = **6 次**

**优化后**：
- 移动 3 个条件节点的位置
- 更新请求：3 × 1 = **3 次**
- 性能提升：**50%**

### 复杂工作流（50 个节点，15 个条件节点）

**优化前**：
- 修改 15 个条件节点
- 更新请求：15 × 2 = **30 次**

**优化后**：
- 修改 15 个条件节点
- 更新请求：15 × 1 = **15 次**
- 性能提升：**50%**

## 技术细节

### 为什么已存在的节点可以一次更新？

**关键点**：已存在的节点的 `branchNodes.targetNodeId` 都是数据库 ID，不需要映射。

```typescript
// 已存在的条件节点
{
  "branchNodes": {
    "branch1": {
      "targetNodeId": "1002"  // ✅ 已经是数据库 ID
    }
  }
}

// 新创建的条件节点
{
  "branchNodes": {
    "branch1": {
      "targetNodeId": "node-456"  // ❌ 临时 ID，需要映射
    }
  }
}
```

### calculateBranchNodesFromNode 的作用

这个函数会自动处理 ID 映射：

```typescript
const calculateBranchNodesFromNode = (
  node: Node,
  nodeIdMapping?: Map<string, string>
): Record<string, any> | undefined => {
  // ...
  
  let targetNodeId: string | undefined;
  if (edge) {
    let targetId = edge.target;
    
    // 如果有映射表且目标 ID 在映射表中，使用映射后的 ID
    if (nodeIdMapping && nodeIdMapping.has(edge.target)) {
      targetId = nodeIdMapping.get(edge.target)!;
    }
    
    targetNodeId = targetId;
  }
  
  // ...
};
```

**对于已存在的节点**：
- `edge.target` 已经是数据库 ID（如 `"1002"`）
- `nodeIdMapping.has("1002")` 返回 `false`
- 直接使用 `edge.target`

**对于新创建的节点**：
- `edge.target` 是临时 ID（如 `"node-456"`）
- `nodeIdMapping.has("node-456")` 返回 `true`
- 使用映射后的数据库 ID（如 `"1002"`）

## 相关文件

- **`src/views/test/composables/useWorkflowApplication.ts`**
  - 第一阶段更新逻辑（优化）
  - 第二阶段更新逻辑（优化）
  - `calculateBranchNodesFromNode` 函数

## 测试建议

### 1. 测试移动已存在的条件节点

```typescript
// 1. 加载已有工作流（包含条件节点）
// 2. 移动条件节点的位置
// 3. 保存工作流
// 4. 验证：
//    - 只发送 1 次更新请求
//    - 请求包含 branchNodes
//    - 保存成功
```

### 2. 测试创建新的条件节点

```typescript
// 1. 创建新的条件节点
// 2. 添加分支和目标节点
// 3. 保存工作流
// 4. 验证：
//    - 发送 1 次创建请求 + 1 次更新请求
//    - 更新请求包含正确的 branchNodes
//    - targetNodeId 是数据库 ID
```

### 3. 测试修改条件节点的分支

```typescript
// 1. 加载已有工作流
// 2. 修改条件节点的分支配置
// 3. 保存工作流
// 4. 验证：
//    - 只发送 1 次更新请求
//    - 请求包含更新后的 branchNodes
//    - 保存成功
```

## 总结

这次优化解决了条件节点重复更新的问题：

✅ **已存在的条件节点**：一次更新即可（包含 branchNodes）  
✅ **新创建的条件节点**：仍需两次操作（创建 + 更新 branchNodes）  
✅ **性能提升**：减少 50% 的更新请求  
✅ **代码清晰**：逻辑更明确，日志更清晰  

这是一次重要的性能优化，特别是对于包含大量条件节点的复杂工作流！🚀

