# 修复 Vue 组件卸载时的 vnode null 错误

## 问题描述

在使用工作流编辑器时，偶尔会出现以下错误：

```
Uncaught (in promise) TypeError: Cannot destructure property 'type' of 'vnode' as it is null.
    at unmount (runtime-core.esm-bundler.js:5847:7)
    at unmountComponent (runtime-core.esm-bundler.js:5999:7)
    at unmount (runtime-core.esm-bundler.js:5879:7)
    at unmountChildren (runtime-core.esm-bundler.js:6019:7)
    ...
```

这是一个 **Vue 3 运行时错误**，发生在组件卸载时。

## 问题分析

### 根本原因

**问题 1：watch 未清理**

在 `vueflow.vue` 的 `onMounted` 中创建了 watch，但没有在组件卸载时停止：

```typescript
// ❌ 问题代码
onMounted(async () => {
  workflowRef.value = useWorkflow(workflowOptions);
  nodeTypes.value = workflowRef.value.nodeTypes;

  // 创建 watch，但没有保存停止函数
  watch(
    () => workflowRef.value?.nodes?.value,
    newNodes => {
      if (newNodes) {
        nodes.value = newNodes;
      }
    },
    { immediate: true, deep: true }
  );

  watch(
    () => workflowRef.value?.edges?.value,
    newEdges => {
      if (newEdges) {
        edges.value = newEdges;
      }
    },
    { immediate: true, deep: true }
  );

  await initData();
});

// ❌ 没有 onBeforeUnmount 清理
```

**问题**：
- 组件卸载后，watch 仍然在运行
- watch 回调尝试更新已销毁组件的响应式数据
- Vue 尝试更新已销毁的 vnode，导致 vnode 为 null

**问题 2：DecisionNode 中的空值处理**

在 `DecisionNode.vue` 中，如果 `branchNodes` 为 `undefined` 或 `null`，`Object.values()` 可能导致错误：

```typescript
// ❌ 问题代码
const branches = computed<BranchConfig[]>(() => {
  return Object.values(props.data.branchNodes); // 如果 branchNodes 为 undefined，会报错
});
```

**问题**：
- 节点初始化时，`branchNodes` 可能还未设置
- 节点卸载时，`props.data` 可能已被清空
- `Object.values(undefined)` 会抛出错误

## 修复方案

### 修复 1：添加 watch 清理逻辑

**文件**：`src/views/test/vueflow.vue`

**修改前**：
```typescript
import { ref, onMounted, shallowRef, computed, watch } from "vue";

// ...

onMounted(async () => {
  workflowRef.value = useWorkflow(workflowOptions);
  nodeTypes.value = workflowRef.value.nodeTypes;

  watch(
    () => workflowRef.value?.nodes?.value,
    newNodes => {
      if (newNodes) {
        nodes.value = newNodes;
      }
    },
    { immediate: true, deep: true }
  );

  watch(
    () => workflowRef.value?.edges?.value,
    newEdges => {
      if (newEdges) {
        edges.value = newEdges;
      }
    },
    { immediate: true, deep: true }
  );

  await initData();
});
```

**修改后**：
```typescript
import { ref, onMounted, onBeforeUnmount, shallowRef, computed, watch } from "vue";

// ...

// 存储 watch 停止函数
const watchStopHandles: (() => void)[] = [];

onMounted(async () => {
  workflowRef.value = useWorkflow(workflowOptions);
  nodeTypes.value = workflowRef.value.nodeTypes;

  // 保存 watch 停止函数
  const stopNodesWatch = watch(
    () => workflowRef.value?.nodes?.value,
    newNodes => {
      if (newNodes) {
        nodes.value = newNodes;
      }
    },
    { immediate: true, deep: true }
  );
  watchStopHandles.push(stopNodesWatch);

  const stopEdgesWatch = watch(
    () => workflowRef.value?.edges?.value,
    newEdges => {
      if (newEdges) {
        edges.value = newEdges;
      }
    },
    { immediate: true, deep: true }
  );
  watchStopHandles.push(stopEdgesWatch);

  await initData();
});

// 在组件卸载前清理
onBeforeUnmount(() => {
  // 停止所有 watch
  watchStopHandles.forEach(stop => stop());
  watchStopHandles.length = 0;
});
```

**改进点**：
- ✅ 导入 `onBeforeUnmount`
- ✅ 保存 watch 返回的停止函数
- ✅ 在组件卸载前调用所有停止函数
- ✅ 清空停止函数数组

### 修复 2：添加空值检查

**文件**：`src/views/test/components/nodes/DecisionNode.vue`

**修改前**：
```typescript
const branches = computed<BranchConfig[]>(() => {
  return Object.values(props.data.branchNodes);
});
```

**修改后**：
```typescript
const branches = computed<BranchConfig[]>(() => {
  if (!props.data?.branchNodes) {
    return [];
  }
  return Object.values(props.data.branchNodes);
});
```

**改进点**：
- ✅ 使用可选链 `?.` 检查 `props.data`
- ✅ 检查 `branchNodes` 是否存在
- ✅ 不存在时返回空数组，避免错误

## 技术细节

### Vue 3 watch 的生命周期

**watch 的特性**：
1. `watch()` 返回一个停止函数
2. 在组件卸载时，watch **不会自动停止**（如果在 setup 外部创建）
3. 在 `onMounted` 中创建的 watch 需要手动清理

**正确的 watch 使用模式**：

```typescript
// ✅ 方式 1：在 setup 顶层创建（自动清理）
const stopWatch = watch(source, callback);

// ✅ 方式 2：在生命周期钩子中创建（手动清理）
onMounted(() => {
  const stopWatch = watch(source, callback);
  
  onBeforeUnmount(() => {
    stopWatch();
  });
});

// ✅ 方式 3：使用数组管理多个 watch
const watchStops: (() => void)[] = [];

onMounted(() => {
  watchStops.push(watch(source1, callback1));
  watchStops.push(watch(source2, callback2));
});

onBeforeUnmount(() => {
  watchStops.forEach(stop => stop());
  watchStops.length = 0;
});
```

### 为什么会导致 vnode null 错误？

**错误链路**：

1. **组件开始卸载**
   - Vue 开始销毁组件
   - 组件的 vnode 被标记为待销毁

2. **watch 仍在运行**
   - watch 检测到数据变化
   - 触发回调函数

3. **尝试更新已销毁的组件**
   - 回调中更新响应式数据（如 `nodes.value = newNodes`）
   - Vue 尝试更新对应的 vnode

4. **vnode 已被销毁**
   - vnode 已经为 null
   - 解构 `vnode.type` 时报错

**时序图**：

```
组件卸载开始
    ↓
vnode 标记为待销毁
    ↓
watch 检测到变化 ← 数据更新
    ↓
watch 回调执行
    ↓
尝试更新 nodes.value
    ↓
Vue 尝试更新 vnode
    ↓
vnode 已为 null ← ❌ 错误发生
    ↓
Cannot destructure property 'type' of 'vnode' as it is null
```

### 其他可能导致此错误的原因

1. **异步操作未清理**
   ```typescript
   // ❌ 问题代码
   onMounted(() => {
     setTimeout(() => {
       // 组件可能已卸载
       someRef.value = newValue;
     }, 1000);
   });
   
   // ✅ 正确做法
   let timeoutId: number;
   
   onMounted(() => {
     timeoutId = setTimeout(() => {
       someRef.value = newValue;
     }, 1000);
   });
   
   onBeforeUnmount(() => {
     clearTimeout(timeoutId);
   });
   ```

2. **事件监听器未移除**
   ```typescript
   // ❌ 问题代码
   onMounted(() => {
     window.addEventListener('resize', handleResize);
   });
   
   // ✅ 正确做法
   onMounted(() => {
     window.addEventListener('resize', handleResize);
   });
   
   onBeforeUnmount(() => {
     window.removeEventListener('resize', handleResize);
   });
   ```

3. **第三方库未销毁**
   ```typescript
   // ❌ 问题代码
   onMounted(() => {
     const chart = new Chart(canvas, config);
   });
   
   // ✅ 正确做法
   let chart: Chart;
   
   onMounted(() => {
     chart = new Chart(canvas, config);
   });
   
   onBeforeUnmount(() => {
     chart?.destroy();
   });
   ```

## 测试建议

### 1. 测试组件卸载

```typescript
// 1. 打开工作流编辑器
// 2. 加载一个工作流
// 3. 快速切换到其他页面
// 4. 检查控制台是否有 vnode null 错误
```

### 2. 测试快速操作

```typescript
// 1. 打开工作流编辑器
// 2. 快速添加/删除多个节点
// 3. 立即关闭页面
// 4. 检查控制台是否有错误
```

### 3. 测试条件节点

```typescript
// 1. 创建一个条件节点
// 2. 不设置分支配置
// 3. 检查节点是否正常显示（应显示 "0 个分支"）
// 4. 删除节点
// 5. 检查控制台是否有错误
```

## 相关文件

### 修改的文件

- **`src/views/test/vueflow.vue`**
  - 添加 `onBeforeUnmount` 导入
  - 添加 watch 停止函数管理
  - 添加组件卸载清理逻辑

- **`src/views/test/components/nodes/DecisionNode.vue`**
  - 添加 `branchNodes` 空值检查

### 相关文件（未修改，但需注意）

- **`src/views/test/components/WorkflowEditor.vue`**
  - 检查后未发现类似问题

- **`src/views/test/components/nodes/ParallelNode.vue`**
  - 已使用可选链 `?.`，无需修改

## 最佳实践

### Vue 3 组件清理检查清单

在编写 Vue 3 组件时，确保清理以下资源：

- [ ] **watch**：保存停止函数，在 `onBeforeUnmount` 中调用
- [ ] **定时器**：`setTimeout`、`setInterval`
- [ ] **事件监听器**：`addEventListener`
- [ ] **第三方库实例**：图表、地图、编辑器等
- [ ] **WebSocket 连接**
- [ ] **动画帧**：`requestAnimationFrame`
- [ ] **Intersection Observer**、**Mutation Observer** 等
- [ ] **异步请求**：取消未完成的请求

### 推荐的组件结构

```typescript
<script setup lang="ts">
import { ref, watch, onMounted, onBeforeUnmount } from 'vue';

// 1. 响应式数据
const data = ref();

// 2. 清理函数集合
const cleanupFunctions: (() => void)[] = [];

// 3. 组件挂载
onMounted(() => {
  // 创建 watch
  const stopWatch = watch(data, callback);
  cleanupFunctions.push(stopWatch);
  
  // 添加事件监听
  window.addEventListener('resize', handleResize);
  cleanupFunctions.push(() => {
    window.removeEventListener('resize', handleResize);
  });
  
  // 创建定时器
  const timerId = setInterval(doSomething, 1000);
  cleanupFunctions.push(() => {
    clearInterval(timerId);
  });
});

// 4. 组件卸载
onBeforeUnmount(() => {
  // 执行所有清理函数
  cleanupFunctions.forEach(cleanup => cleanup());
  cleanupFunctions.length = 0;
});
</script>
```

## 总结

这次修复解决了两个问题：

✅ **watch 未清理**：在组件卸载时停止所有 watch  
✅ **空值处理**：在 DecisionNode 中添加空值检查  

这些修复可以有效防止组件卸载时的 vnode null 错误，提高应用的稳定性。

**关键要点**：
- 在生命周期钩子中创建的 watch 需要手动清理
- 使用 `onBeforeUnmount` 清理所有副作用
- 在访问可能为空的对象属性时，使用可选链 `?.` 和空值检查
- 遵循 Vue 3 组件清理最佳实践

