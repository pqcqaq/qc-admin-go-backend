# Branch Nodes 数据结构重构

## 修改概述

将 `branch_nodes` 字段从简单的 ID 映射改为存储完整的分支配置信息，使其与 `parallel_config.threads` 的设计模式保持一致。

## 修改内容

### 1. 后端 Schema 修改

**文件：** `database/schema/workflow.go`

**修改前：**
```go
field.JSON("branch_nodes", map[string]uint64{}).Optional().Comment("分支节点映射")
```

**修改后：**
```go
field.JSON("branch_nodes", map[string]interface{}{}).Optional().Comment("分支配置映射（存储完整的分支配置：name, condition, handlerId, targetNodeId）")
```

### 2. 前端类型定义修改

**文件：** `src/workflow/types.ts`

**新增接口：**
```typescript
export interface BranchNodeConfig {
  name: string;
  condition?: string;
  handlerId?: string;
  targetNodeId?: number;
}
```

**修改 WorkflowNodeResponse：**
```typescript
branchNodes?: Record<string, BranchNodeConfig>;  // 之前是 Record<string, number>
```

### 3. 前端数据转换逻辑

**文件：** `src/views/test/composables/useWorkflowApplication.ts`

#### 3.1 读取数据（后端 → 前端）

在 `convertNodeResponseToVueFlowNode` 函数中：

```typescript
// 从 branchNodes 转换为 branches 数组（用于 UI 显示）
let branches: any[] | undefined;
if (node.branchNodes && Object.keys(node.branchNodes).length > 0) {
  branches = Object.values(node.branchNodes).map((branchConfig: any) => ({
    name: branchConfig.name,
    condition: branchConfig.condition,
    handlerId: branchConfig.handlerId,
    // targetNodeId 会在 computed 中从 edges 读取，这里不需要
  }));
}
```

**说明：**
- 从后端的 `branchNodes`（Record 格式）转换为前端的 `branches`（Array 格式）
- 不包含 `targetNodeId`，因为它会在 UI 的 computed 属性中从 edges 动态读取

#### 3.2 保存数据（前端 → 后端）

新增 `calculateBranchNodesFromNode` 函数：

```typescript
const calculateBranchNodesFromNode = (
  node: Node,
  nodeIdMapping?: Map<string, string>
): Record<string, any> | undefined => {
  // 从 node.data.branches 读取分支配置
  const branches = node.data.branches;
  if (!branches || branches.length === 0) return undefined;

  const edges = workflow.getAllEdges();
  const branchNodes: Record<string, any> = {};

  branches.forEach((branch: any) => {
    const branchName = branch.name;
    
    // 查找对应的 edge 获取 targetNodeId
    const expectedSourceHandle = `${node.id}-branch-${branchName}`;
    const edge = edges.find(
      e => e.source === node.id && e.sourceHandle === expectedSourceHandle
    );

    let targetNodeId: number | undefined;
    if (edge) {
      let targetId = edge.target;
      if (nodeIdMapping && nodeIdMapping.has(edge.target)) {
        targetId = nodeIdMapping.get(edge.target)!;
      }
      const targetIdNum = parseInt(targetId, 10);
      if (!isNaN(targetIdNum)) {
        targetNodeId = targetIdNum;
      }
    }

    // 构建完整的分支配置
    branchNodes[branchName] = {
      name: branchName,
      condition: branch.condition || "",
      handlerId: branch.handlerId,
      targetNodeId
    };
  });

  return Object.keys(branchNodes).length > 0 ? branchNodes : undefined;
};
```

**说明：**
- 从 `node.data.branches` 读取分支配置（name, condition, handlerId）
- 从 edges 中查找对应的连接，获取 `targetNodeId`
- 合并成完整的 `branchNodes` 配置保存到数据库

## 数据流说明

### 创建新的条件节点

1. **用户拖拽创建节点**
   - 节点带有默认的 `branches` 配置（来自 `nodeConfig.ts`）
   - 例如：`[{ name: "true", condition: "result === true" }, { name: "false", condition: "result === false" }]`

2. **用户连接分支**
   - 创建 edge，设置 `sourceHandle = "nodeId-branch-true"`
   - edge 中包含 `branchName = "true"`

3. **保存到数据库**
   - 调用 `calculateBranchNodesFromNode` 计算完整配置
   - 保存到 `branch_nodes` 字段：
     ```json
     {
       "true": {
         "name": "true",
         "condition": "result === true",
         "handlerId": null,
         "targetNodeId": 123
       },
       "false": {
         "name": "false",
         "condition": "result === false",
         "handlerId": null,
         "targetNodeId": 456
       }
     }
     ```

### 加载已有的条件节点

1. **从数据库读取**
   - `branchNodes` 包含完整配置

2. **转换为前端格式**
   - 调用 `convertNodeResponseToVueFlowNode`
   - 将 `branchNodes` 转换为 `branches` 数组（不包含 targetNodeId）

3. **UI 显示**
   - `branches` computed 属性从 `node.data.branches` 和 edges 合并数据
   - 动态计算 `targetNodeId` 用于显示

## 与 parallelConfig 的对比

| 特性 | parallelConfig.threads | branch_nodes |
|------|----------------------|--------------|
| 存储位置 | `parallel_config.threads` | `branch_nodes` |
| 数据格式 | Array | Record (Map) |
| 配置信息 | id, name, handlerId | name, condition, handlerId |
| 连接信息 | targetNodeId | targetNodeId |
| UI 显示 | parallelChildren (computed) | branches (computed) |
| Handle ID | `nodeId-parallel-${thread.id}` | `nodeId-branch-${branch.name}` |

## 优势

1. **数据完整性**：分支配置信息（name, condition, handlerId）和连接信息（targetNodeId）都保存在数据库中
2. **设计一致性**：与 `parallelConfig.threads` 的设计模式保持一致
3. **职责清晰**：`config` 字段只用于节点的公共配置，不混入特定节点类型的配置
4. **易于扩展**：未来可以轻松添加更多分支配置字段

## 注意事项

1. **向后兼容**：需要迁移旧数据，将简单的 `{branchName: nodeId}` 格式转换为新格式
2. **数据同步**：删除分支时，需要同时删除对应的 edge
3. **Handle ID**：必须使用 `${nodeId}-branch-${branchName}` 格式，确保分支的稳定性

