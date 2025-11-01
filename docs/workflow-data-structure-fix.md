# 工作流数据结构修复说明

## 问题描述

当前系统存在历史数据结构遗留问题：

### 1. 条件节点（Condition Checker）

**问题：**
- 前端在节点配置中保存 `branches` 数组
- 后端在节点表中保存 `branch_nodes` 字段（旧架构）
- 但实际上，分支连接关系应该通过 `edge` 表的 `branch_name` 字段管理

**正确的架构：**
- `node.data.branches` 只用于 UI 显示（渲染 handle）
- 实际的分支连接通过 `edge` 表管理，每个分支对应一条 edge，带有 `branch_name` 字段
- 不需要在节点配置中保存 `branches` 或 `branch_nodes`

### 2. 并行节点（Parallel Executor）

**问题：**
- 前端在节点配置中保存 `parallelChildren` 数组
- 但如果用户添加了并行任务但还没有连接节点，这些任务信息就会丢失
- 无法保存没有连接的任务

**正确的架构：**
- `node.data.parallelConfig.threads` 保存任务信息（任务 ID、名称、handler ID 等），**会保存到数据库**
- `node.data.parallelChildren` 从 `parallelConfig.threads` 和 edges 计算得出，只用于 UI 显示
- 实际的并行子节点关系通过 `parent_node_id` 管理
- Handle ID 使用 `parallel-${thread.id}`，而不是 `parallel-${index}`

## 已完成的修复

### 1. 前端保存逻辑修复

**文件：** `src/views/test/composables/useWorkflowApplication.ts`

**修改内容：**
- ✅ 移除了 `branches` 从节点保存数据中（只用于 UI 显示）
- ✅ 移除了 `parallelChildren` 从节点保存数据中（只用于 UI 显示）
- ✅ 添加了 `calculateBranchNodesFromEdges()` 函数，从 edges 中计算 `branchNodes`
- ✅ 在保存节点时，自动从 edges 计算并保存 `branchNodes` 字段（后端需要）
- ✅ 添加了注释说明数据流和保存逻辑

**修改位置：**
1. 新增 `calculateBranchNodesFromEdges()` 函数 - 从 edges 中提取分支信息
2. `getNodeHash()` 函数 - 计算节点 hash 时包含从 edges 计算的 `branchNodes`
3. 创建节点时 - 从 edges 计算 `branchNodes` 并保存
4. 更新节点时 - 从 edges 计算 `branchNodes` 并保存

**核心逻辑：**
```typescript
// 从 edges 中计算 branchNodes
const calculateBranchNodesFromEdges = (nodeId: string): Record<string, number> | undefined => {
  const edges = workflow.getAllEdges();
  const branchEdges = edges.filter(
    e => e.source === nodeId && e.data?.branchName
  );

  if (branchEdges.length === 0) return undefined;

  const branchNodes: Record<string, number> = {};
  branchEdges.forEach(edge => {
    const branchName = edge.data.branchName;
    const targetId = parseInt(edge.target, 10);
    if (branchName && !isNaN(targetId)) {
      branchNodes[branchName] = targetId;
    }
  });

  return Object.keys(branchNodes).length > 0 ? branchNodes : undefined;
};
```

### 2. 类型定义修复

**文件：** `src/views/test/components/types.ts`

**修改内容：**
- ✅ 新增 `ParallelThreadConfig` 接口，定义并行任务的数据结构
- ✅ 更新 `ParallelConfig` 接口，添加 `threads` 字段
- ✅ 更新 `ParallelChildConfig` 接口，说明从 `parallelConfig.threads` 和 edges 计算得出
- ✅ 为 `BranchConfig` 添加注释，说明只用于 UI 显示
- ✅ 为 `NodeData` 接口添加注释，说明哪些字段会保存到数据库

**新增的类型定义：**
```typescript
/**
 * 并行任务线程配置
 * 这个配置会保存到数据库的 parallel_config.threads 字段
 */
export interface ParallelThreadConfig {
  id: string; // 任务唯一标识符（用于关联 edge 的 sourceHandle）
  name: string; // 任务名称
  handlerId?: string; // 处理器节点 ID（可选，用于指定特定的处理逻辑）
  [key: string]: any;
}

/**
 * 并行配置（用于并行节点）
 * 注意：这个配置会保存到数据库的 parallel_config 字段
 */
export interface ParallelConfig {
  mode?: "all" | "any" | "race"; // 并行模式：全部完成、任意完成、竞速
  timeout?: number; // 超时时间
  threads?: ParallelThreadConfig[]; // 并行任务列表（保存任务信息）
  [key: string]: any;
}
```

### 3. 并行节点数据结构重构

**文件：** `src/views/test/components/PropertiesPanel/index.vue`

**修改内容：**
- ✅ 修改 `handleAddParallelChild()` - 在 `parallelConfig.threads` 中添加任务
- ✅ 修改 `handleRemoveParallelChild()` - 从 `parallelConfig.threads` 中删除任务
- ✅ 修改 `handleUpdateParallelChildName()` - 更新 `parallelConfig.threads` 中的任务名称

**文件：** `src/views/test/components/PropertiesPanel/composables/useNodeOperations.ts`

**修改内容：**
- ✅ 修改 `parallelChildren` computed - 从 `parallelConfig.threads` 和 edges 计算得出

**文件：** `src/views/test/components/nodes/ParallelNode.vue`

**修改内容：**
- ✅ 修改 handle 渲染逻辑 - 从 `parallelConfig.threads` 读取任务信息
- ✅ 修改 handle ID - 使用 `parallel-${thread.id}` 而不是 `parallel-${index}`

**文件：** `src/views/test/components/nodeConfig.ts`

**修改内容：**
- ✅ 更新默认配置 - 在 `parallelConfig` 中添加 `threads` 字段

### 4. 属性面板注释

**文件：** `src/views/test/components/PropertiesPanel/index.vue`

**修改内容：**
- ✅ 添加了详细的注释说明分支和并行任务的操作逻辑
- ✅ 标记了 TODO 项，提醒需要在删除分支/并行任务时同步删除 edge

## 待完成的工作

### 1. 删除分支时同步删除 Edge

**位置：** `PropertiesPanel/index.vue` - `handleRemoveBranch()`

**需要实现：**
```typescript
function handleRemoveBranch(index: number) {
  if (!props.selectedNode) return;

  const currentBranches = props.selectedNode.data.branches || [];
  if (currentBranches.length <= 1) return;

  const branchName = currentBranches[index].name;
  
  // 1. 找到对应的 edge
  const edgeToDelete = getEdges.value.find(
    e => e.source === props.selectedNode.id && 
         e.sourceHandle?.includes(`branch-${branchName}`)
  );
  
  // 2. 删除 edge
  if (edgeToDelete) {
    await deleteEdge(edgeToDelete.id);
  }
  
  // 3. 更新节点配置
  const newBranches = currentBranches.filter((_, i) => i !== index);
  updateNodeData("branches", newBranches);
}
```

### 2. 删除并行任务时同步删除 Edge 和清除 parent_node_id

**位置：** `PropertiesPanel/index.vue` - `handleRemoveParallelChild()`

**需要实现：**
```typescript
function handleRemoveParallelChild(index: number) {
  if (!props.selectedNode) return;

  const currentChildren = props.selectedNode.data.parallelChildren || [];
  if (currentChildren.length <= 1) return;

  // 1. 找到对应的 edge
  const edgeToDelete = getEdges.value.find(
    e => e.source === props.selectedNode.id && 
         e.sourceHandle?.includes(`parallel-${index}`)
  );
  
  // 2. 如果有连接的子节点，清除其 parent_node_id
  if (edgeToDelete) {
    const childNodeId = edgeToDelete.target;
    await removeNodeFromParallel(childNodeId);
    await deleteEdge(edgeToDelete.id);
  }
  
  // 3. 更新节点配置
  const newChildren = currentChildren.filter((_, i) => i !== index);
  updateNodeData("parallelChildren", newChildren);
}
```

### 3. 后端清理旧字段

**当前状态：**
- ✅ `branch_nodes` 字段**仍在使用**
- ✅ 后端有 `ConnectBranch` 和 `DisconnectBranch` API，会更新 `branch_nodes` 字段
- ⚠️ 这说明后端同时维护了两套数据：
  - 旧架构：`node.branch_nodes` 字段（map[string]uint64）
  - 新架构：`edge.branch_name` 字段

**问题：**
- 前端现在通过 edge 来管理分支连接
- 但后端的 `ConnectBranch` API 会同时更新 `branch_nodes` 字段
- 这导致数据冗余，可能出现不一致

**建议方案：**

**方案 A：继续使用 `branch_nodes`（推荐）**
1. 前端在保存时也保存 `branchNodes` 字段
2. 从 `edges` 中提取分支信息，构建 `branchNodes` map
3. 保持与后端 API 的一致性
4. 优点：不需要修改后端，向后兼容
5. 缺点：数据冗余

**方案 B：完全迁移到 edge 表**
1. 废弃 `branch_nodes` 字段
2. 废弃 `ConnectBranch` 和 `DisconnectBranch` API
3. 所有分支连接通过 edge 表管理
4. 优点：数据结构清晰，无冗余
5. 缺点：需要修改后端，可能影响现有功能

**当前采用方案 A**，因为：
- 后端已经有完整的 `branch_nodes` 逻辑
- 修改后端影响范围大
- 可以保持向后兼容

### 4. 数据迁移

**如果有历史数据：**
1. 将 `branch_nodes` 中的数据迁移到 `edge` 表
2. 将并行子节点关系迁移到 `parent_node_id`
3. 清理节点配置中的 `branches` 和 `parallelChildren`

## 数据流说明

### 条件节点的数据流

1. **创建节点时：**
   - 节点带有默认的 `branches` 配置（如 `[{name: "true"}, {name: "false"}]`）
   - 这些配置只用于渲染 handle，不保存到数据库

2. **用户连接分支时：**
   - 用户从 `branch-true` handle 拖拽连接线到目标节点
   - 创建一个 edge，带有 `branchName: "true"`
   - edge 保存到数据库

3. **加载工作流时：**
   - 从数据库加载节点（包含 `branches` 配置）
   - 从数据库加载 edges（包含 `branchName`）
   - UI 根据 `branches` 渲染 handle
   - UI 根据 edges 显示连接线

4. **保存工作流时：**
   - 节点数据**不包含** `branches`（只保存其他配置）
   - edges 正常保存（包含 `branchName`）

### 并行节点的数据流

1. **创建节点时：**
   - 节点带有默认的 `parallelConfig.threads` 配置（如 `[{id: "thread-1", name: "任务1"}, {id: "thread-2", name: "任务2"}]`）
   - 这些配置**会保存到数据库**

2. **用户添加并行任务时：**
   - 用户点击"添加并行任务"按钮
   - 在 `parallelConfig.threads` 中添加一个新任务（带有唯一 ID）
   - UI 根据 `threads` 渲染 handle（使用 `parallel-${thread.id}` 作为 handle ID）
   - **即使没有连接节点，任务信息也会保存**

3. **用户连接并行任务时：**
   - 用户从 `parallel-thread-1` handle 拖拽连接线到目标节点
   - 创建一个 edge，sourceHandle 为 `parallel-thread-1`
   - 调用 API 设置目标节点的 `parent_node_id` 为当前并行节点的 ID

4. **加载工作流时：**
   - 从数据库加载节点（包含 `parallelConfig.threads`）
   - 从数据库加载 edges
   - UI 根据 `parallelConfig.threads` 渲染 handle
   - UI 根据 edges 显示连接线
   - `parallelChildren` 从 `parallelConfig.threads` 和 edges 计算得出（用于属性面板显示）

5. **保存工作流时：**
   - 节点数据**包含** `parallelConfig`（包括 `threads`）
   - edges 正常保存
   - `parent_node_id` 关系已经在连接时设置，不需要再次保存

## 总结

**核心原则：**

1. **UI 显示数据（不保存到数据库）：**
   - `node.data.branches` - 只用于渲染条件节点的 handle
   - `node.data.parallelChildren` - 从 `parallelConfig.threads` 和 edges 计算得出，用于属性面板显示

2. **数据库保存数据：**
   - `node.branch_nodes` - 从 edges 中计算得出，保存到数据库（后端需要）
   - `edge.branch_name` - 分支连接的分支名称
   - `node.parallelConfig.threads` - 并行任务列表（任务 ID、名称、handler ID 等）
   - `node.parent_node_id` - 并行子节点的父节点 ID

3. **条件节点数据流：**
   - 用户在 UI 中连接分支 → 创建 edge（带 branchName）
   - 保存工作流时 → 从 edges 计算 branchNodes → 保存到数据库
   - 加载工作流时 → 从数据库读取 branches 配置 → 渲染 handle

4. **并行节点数据流：**
   - 用户添加并行任务 → 在 `parallelConfig.threads` 中添加任务
   - 用户连接并行任务 → 创建 edge（sourceHandle 为 `parallel-${thread.id}`）
   - 保存工作流时 → 保存 `parallelConfig`（包括 threads）
   - 加载工作流时 → 从 `parallelConfig.threads` 渲染 handle → 从 edges 获取连接关系

**好处：**

1. **数据一致性：** `branchNodes` 始终从 edges 计算得出，不会出现不一致
2. **职责分离：** UI 配置（branches）和数据库数据（branchNodes）分离
3. **向后兼容：** 保持与后端 API 的兼容性
4. **易于维护：** 数据流清晰，逻辑简单

**注意事项：**

1. 前端不应该直接修改 `node.data.branchNodes`，这个字段由系统自动计算
2. 删除分支时，需要同时删除对应的 edge
3. 删除并行任务时，需要同时删除对应的 edge 和清除子节点的 parent_node_id
4. Handle ID 必须使用 `${nodeId}-parallel-${thread.id}` 格式，而不是基于索引，以确保任务的稳定性
5. **所有 sourceHandle 和 targetHandle 的匹配必须使用精确匹配（`===`），不能使用 `includes()`**

## 数据结构示例

### 并行节点的数据结构

**保存到数据库的数据：**
```json
{
  "id": "123",
  "type": "parallel_executor",
  "label": "并行处理",
  "parallelConfig": {
    "mode": "all",
    "timeout": 30000,
    "threads": [
      {
        "id": "thread-1",
        "name": "任务1",
        "handlerId": "456"
      },
      {
        "id": "thread-2",
        "name": "任务2"
      }
    ]
  }
}
```

**UI 显示的数据（从 parallelConfig.threads 和 edges 计算得出）：**
```json
{
  "parallelChildren": [
    {
      "id": "thread-1",
      "name": "任务1",
      "handlerId": "456",
      "targetNodeId": "789"
    },
    {
      "id": "thread-2",
      "name": "任务2",
      "targetNodeId": null
    }
  ]
}
```

**对应的 edges：**
```json
[
  {
    "id": "edge-1",
    "source": "123",
    "target": "789",
    "sourceHandle": "123-parallel-thread-1",
    "targetHandle": "789-input"
  }
]
```

