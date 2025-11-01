# 实时模式功能实现总结

## 功能概述

实现了工作流编辑器的实时模式功能，当启用实时模式时，系统会每 500ms 自动计算工作流的变更（diff），并将变更数据发送到 Socket.IO（预留 TODO）。

## 实现的功能

### 1. 实时模式状态管理

**文件：** `src/views/test/composables/useWorkflowApplication.ts`

**新增状态：**
- `realtimeMode` - 实时模式是否启用（ref<boolean>）
- `realtimeTimer` - 定时器引用（用于清理）

**关键代码：**
```typescript
// 实时模式状态
const realtimeMode = ref(false);
let realtimeTimer: ReturnType<typeof setInterval> | null = null;
```

---

### 2. Diff 计算逻辑

**函数：** `calculateWorkflowDiff()`

**功能：**
- 计算当前工作流与 snapshot 的差异
- 返回节点和边的创建、更新、删除列表

**返回数据结构：**
```typescript
{
  nodes: {
    created: Node[],    // 新增的节点
    updated: Node[],    // 更新的节点
    deleted: string[]   // 删除的节点 ID
  },
  edges: {
    created: Edge[],    // 新增的边
    updated: Edge[],    // 更新的边
    deleted: string[]   // 删除的边 ID
  }
}
```

**实现逻辑：**
1. 获取当前的节点和边
2. 与 snapshot 中的数据进行对比
3. 使用 hash 函数检测节点/边的业务数据是否变更
4. 分类为创建、更新、删除三种操作

**关键代码：**
```typescript
const calculateWorkflowDiff = () => {
  const currentNodes = workflow.getAllNodes();
  const currentEdges = workflow.getAllEdges();

  const diff = {
    nodes: {
      created: [] as Node[],
      updated: [] as Node[],
      deleted: [] as string[]
    },
    edges: {
      created: [] as Edge[],
      updated: [] as Edge[],
      deleted: [] as string[]
    }
  };

  // 计算节点的 diff
  const currentNodeIds = new Set(currentNodes.map(n => n.id));
  const snapshotNodeIds = new Set(snapshot.value.nodes.keys());

  // 新增的节点
  for (const node of currentNodes) {
    if (!snapshot.value.nodes.has(node.id)) {
      diff.nodes.created.push(node);
    } else {
      // 检查是否更新
      const nodeHash = getNodeHash(node);
      const snapshotHash = snapshot.value.nodeHashes.get(node.id);
      if (nodeHash !== snapshotHash) {
        diff.nodes.updated.push(node);
      }
    }
  }

  // 删除的节点
  for (const nodeId of snapshotNodeIds) {
    if (!currentNodeIds.has(nodeId)) {
      diff.nodes.deleted.push(nodeId);
    }
  }

  // 计算边的 diff（类似逻辑）
  // ...

  return diff;
};
```

---

### 3. 实时模式控制函数

#### `startRealtimeMode()`

**功能：** 启动实时模式，开始定期计算 diff

**实现：**
```typescript
const startRealtimeMode = () => {
  if (realtimeTimer) {
    return; // 已经启动
  }

  debugLog("实时模式", "✅ 启动实时模式");
  realtimeMode.value = true;

  realtimeTimer = setInterval(() => {
    const diff = calculateWorkflowDiff();

    // 检查是否有变更
    const hasChanges =
      diff.nodes.created.length > 0 ||
      diff.nodes.updated.length > 0 ||
      diff.nodes.deleted.length > 0 ||
      diff.edges.created.length > 0 ||
      diff.edges.updated.length > 0 ||
      diff.edges.deleted.length > 0;

    if (hasChanges) {
      debugLog("实时模式", "检测到变更", diff);

      // TODO: 将 diff 数据发送到 Socket.IO
      // Example:
      // socket.emit('workflow:update', {
      //   applicationId: currentApplication.value?.id,
      //   diff: diff
      // });
    }
  }, 500);
};
```

**特点：**
- 每 500ms 执行一次 diff 计算
- 只有检测到变更时才输出日志
- 预留了 Socket.IO 发送逻辑的 TODO

#### `stopRealtimeMode()`

**功能：** 停止实时模式，清理定时器

**实现：**
```typescript
const stopRealtimeMode = () => {
  if (realtimeTimer) {
    clearInterval(realtimeTimer);
    realtimeTimer = null;
    realtimeMode.value = false;
    debugLog("实时模式", "❌ 停止实时模式");
  }
};
```

#### `toggleRealtimeMode(enabled: boolean)`

**功能：** 切换实时模式开关

**实现：**
```typescript
const toggleRealtimeMode = (enabled: boolean) => {
  if (enabled) {
    startRealtimeMode();
  } else {
    stopRealtimeMode();
  }
};
```

---

### 4. EditorToolbar 组件更新

**文件：** `src/views/test/components/EditorToolbar.vue`

**新增 Props：**
- `realtimeMode?: boolean` - 实时模式状态

**新增 Emits：**
- `toggle-realtime: [enabled: boolean]` - 切换实时模式事件

**UI 更新：**
```vue
<div class="toolbar-right">
  <!-- 未保存变更提示 -->
  <el-tag v-if="hasUnsavedChanges" type="warning" effect="dark">
    <el-icon><WarningFilled /></el-icon>
    有未保存的变更
  </el-tag>

  <!-- 实时模式开关 -->
  <div class="realtime-mode-switch">
    <span class="switch-label">实时模式</span>
    <el-switch
      :model-value="realtimeMode"
      @change="handleRealtimeModeChange"
    />
  </div>

  <!-- 保存按钮 -->
  <el-button
    type="success"
    :loading="saving"
    :disabled="!hasUnsavedChanges"
    @click="handleSave"
  >
    <el-icon><DocumentChecked /></el-icon>
    保存
  </el-button>
</div>
```

**样式：**
```scss
.toolbar-right {
  display: flex;
  gap: 12px;
  align-items: center;

  .realtime-mode-switch {
    display: flex;
    gap: 8px;
    align-items: center;
    padding: 0 12px;
    border-radius: 4px;
    background: #f5f7fa;

    .switch-label {
      font-size: 14px;
      color: #606266;
      white-space: nowrap;
    }
  }
}
```

---

### 5. flowapp.vue 集成

**文件：** `src/views/test/flowapp.vue`

**导入状态和方法：**
```typescript
const {
  // ... 其他状态
  realtimeMode,
  // ... 其他方法
  toggleRealtimeMode,
  workflow
} = workflowApp;
```

**传递给 EditorToolbar：**
```vue
<EditorToolbar
  :application-name="currentApplication.name"
  :has-unsaved-changes="hasUnsavedChanges"
  :saving="saving"
  :dark-mode="darkMode"
  :realtime-mode="realtimeMode"
  @back="handleBackToList"
  @save="handleSaveWorkflow"
  @toggle-realtime="toggleRealtimeMode"
/>
```

---

## 使用流程

1. **启用实时模式**
   - 点击工具栏中的"实时模式"开关
   - 系统开始每 500ms 计算一次 diff
   - 控制台会输出检测到的变更

2. **编辑工作流**
   - 添加/删除/修改节点
   - 添加/删除/修改连线
   - 实时模式会自动检测这些变更

3. **查看变更**
   - 打开浏览器控制台
   - 查看 `[实时模式]` 相关的日志输出
   - 可以看到详细的 diff 数据

4. **停止实时模式**
   - 关闭"实时模式"开关
   - 定时器被清理，停止 diff 计算

---

## 待完成的工作（TODO）

### Socket.IO 集成

在 `startRealtimeMode()` 函数中，需要完成 Socket.IO 的发送逻辑：

```typescript
// TODO: 将 diff 数据发送到 Socket.IO
// Example:
// socket.emit('workflow:update', {
//   applicationId: currentApplication.value?.id,
//   diff: diff
// });
```

**建议的实现步骤：**

1. **安装 Socket.IO 客户端**
   ```bash
   npm install socket.io-client
   ```

2. **创建 Socket.IO 连接**
   ```typescript
   import { io } from 'socket.io-client';
   
   const socket = io('http://your-backend-url');
   ```

3. **发送 diff 数据**
   ```typescript
   if (hasChanges) {
     socket.emit('workflow:update', {
       applicationId: currentApplication.value?.id,
       timestamp: Date.now(),
       diff: diff
     });
   }
   ```

4. **处理连接状态**
   ```typescript
   socket.on('connect', () => {
     console.log('Socket.IO 已连接');
   });
   
   socket.on('disconnect', () => {
     console.log('Socket.IO 已断开');
   });
   ```

---

## 技术细节

### Diff 计算优化

- 使用 `Set` 数据结构快速查找节点/边是否存在
- 使用 hash 函数检测业务数据是否变更（避免深度比较）
- 只在有变更时才输出日志和发送数据

### 定时器管理

- 使用 `setInterval` 实现定期执行
- 使用 `clearInterval` 清理定时器，避免内存泄漏
- 在启动前检查是否已经启动，避免重复启动

### 状态同步

- `realtimeMode` 状态与 UI 开关双向绑定
- 通过 `toggleRealtimeMode` 方法统一管理启动/停止逻辑

---

## 相关文件

### 修改的文件
- `src/views/test/composables/useWorkflowApplication.ts` - 添加实时模式逻辑
- `src/views/test/components/EditorToolbar.vue` - 添加实时模式开关
- `src/views/test/flowapp.vue` - 集成实时模式状态和方法

### 新增的文件
- `docs/realtime-mode-feature.md` - 功能文档

---

## 完成时间

2025-10-31

