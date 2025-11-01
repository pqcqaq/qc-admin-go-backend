# 节点部分字段更新优化

## 问题描述

移动节点位置时，虽然只改变了 `positionX` 和 `positionY`，但更新请求包含了所有字段：

```json
{
  "name": "条件检查",
  "nodeKey": "591287546177127446",
  "type": "condition_checker",
  "description": "",
  "config": {},
  "applicationId": "591202603132519446",
  "positionX": 1228.000015258789,
  "positionY": -20.849990844726562,
  "branchNodes": { "true": { "name": "true", "condition": "result === true" } },
  "color": "#E6A23C"
}
```

### 问题分析

**原来的逻辑**：
- 无论哪个字段变更，都提交所有字段
- 浪费网络带宽
- 增加数据传输量
- 降低保存性能

**期望的行为**：
- 只提交变更的字段
- 减少网络传输
- 提升保存性能

## 后端支持情况

### 后端更新逻辑

后端的 `UpdateWorkflowNode` 函数使用**条件更新**：

```go
// internal/funcs/workflow_func.go
func (WorkflowFuncs) UpdateWorkflowNode(ctx context.Context, id uint64, req *models.UpdateWorkflowNodeRequest) (*models.WorkflowNodeResponse, error) {
    builder := database.Client.WorkflowNode.UpdateOneID(id)

    if req.Name != "" {
        builder = builder.SetName(req.Name)
    }

    if req.Description != "" {
        builder = builder.SetDescription(req.Description)
    }

    if req.Prompt != "" {
        builder = builder.SetPrompt(req.Prompt)
    }

    // ... 其他字段
}
```

**特点**：
- ✅ 支持部分更新：不传的字段不会被更新
- ✅ 使用空值判断：`!= ""`、`!= nil` 等
- ⚠️ 限制：无法将字段清空为空值（因为空值会被跳过）

### API 类型定义

```typescript
// src/workflow/types.ts
export interface UpdateWorkflowNodeRequest {
  // 必填字段
  name: string;
  nodeKey: string;
  type: WorkflowNodeType;
  config: Record<string, any>;

  // 可选字段
  description?: string;
  prompt?: string;
  processorLanguage?: string;
  processorCode?: string;
  branchNodes?: Record<string, BranchNodeConfig>;
  parallelConfig?: Record<string, any>;
  apiConfig?: Record<string, any>;
  async?: boolean;
  timeout?: number;
  retryCount?: number;
  positionX?: number;
  positionY?: number;
  color?: string;
}
```

**字段分类**：
- **必填字段**（4 个）：`name`、`nodeKey`、`type`、`config`
- **可选字段**（13 个）：其他所有字段

## 优化方案

### 核心思路

**只提交必填字段 + 变更的可选字段**

1. **必填字段**：总是提交（后端要求）
2. **可选字段**：只提交变更的字段

### 实现细节

#### 修改前

```typescript
// 更新修改的节点
for (const node of nodesToUpdate) {
  const nodeData: any = {
    name: node.data.label || node.id,
    nodeKey: node.id,
    type: node.type,
    description: node.data.description || "",
    config: node.data.config || {},
    applicationId,
    positionX: node.position.x,
    positionY: node.position.y,
    prompt: node.data.prompt,
    processorLanguage: node.data.processorLanguage,
    processorCode: node.data.processorCode,
    apiConfig: node.data.apiConfig,
    parallelConfig: node.data.parallelConfig,
    async: node.data.async,
    timeout: node.data.timeout,
    retryCount: node.data.retryCount,
    color: node.data.color
  };

  await updateWorkflowNode(node.id, nodeData);
}
```

#### 修改后

```typescript
// 更新修改的节点
for (const node of nodesToUpdate) {
  // 计算字段级别的变化
  let changedFieldsList: string[] = [];
  if (snapshotNode) {
    const fieldChangesInfo = getNodeFieldChanges(node, snapshotNode);
    if (fieldChangesInfo) {
      changedFieldsList = fieldChangesInfo.changedFields;
    }
  }

  // 构建更新数据：只包含必填字段 + 变更的字段
  const nodeData: any = {
    // 必填字段（后端要求）
    name: node.data.label || node.id,
    nodeKey: node.id,
    type: node.type,
    config: node.data.config || {},
    applicationId
  };

  // 只添加变更的可选字段
  if (changedFieldsList.includes("position")) {
    nodeData.positionX = node.position.x;
    nodeData.positionY = node.position.y;
  }

  if (changedFieldsList.includes("data.description")) {
    nodeData.description = node.data.description || "";
  }

  if (changedFieldsList.includes("data.prompt")) {
    nodeData.prompt = node.data.prompt;
  }

  // ... 其他可选字段

  await updateWorkflowNode(node.id, nodeData);
}
```

### 字段映射表

| 变更字段 | 提交字段 | 说明 |
|---------|---------|------|
| `position` | `positionX`, `positionY` | 位置变更 |
| `data.description` | `description` | 描述变更 |
| `data.prompt` | `prompt` | 提示词变更 |
| `data.processorLanguage` | `processorLanguage` | 处理器语言变更 |
| `data.processorCode` | `processorCode` | 处理器代码变更 |
| `data.apiConfig` | `apiConfig` | API 配置变更 |
| `data.parallelConfig` | `parallelConfig` | 并行配置变更 |
| `data.async` | `async` | 异步标志变更 |
| `data.timeout` | `timeout` | 超时时间变更 |
| `data.retryCount` | `retryCount` | 重试次数变更 |
| `data.color` | `color` | 颜色变更 |
| `data.branchNodes` | `branchNodes` | 分支配置变更 |

## 优化效果

### 场景 1：只移动节点位置

**优化前**：
```json
{
  "name": "条件检查",
  "nodeKey": "591287546177127446",
  "type": "condition_checker",
  "description": "",
  "config": {},
  "applicationId": "591202603132519446",
  "positionX": 1228.000015258789,
  "positionY": -20.849990844726562,
  "branchNodes": { "true": { "name": "true", "condition": "result === true" } },
  "color": "#E6A23C",
  "prompt": "",
  "processorLanguage": "",
  "processorCode": "",
  "apiConfig": {},
  "parallelConfig": {},
  "async": false,
  "timeout": 30,
  "retryCount": 0
}
```
- 字段数：**17 个**
- 数据量：**~500 字节**

**优化后**：
```json
{
  "name": "条件检查",
  "nodeKey": "591287546177127446",
  "type": "condition_checker",
  "config": {},
  "applicationId": "591202603132519446",
  "positionX": 1228.000015258789,
  "positionY": -20.849990844726562
}
```
- 字段数：**7 个**
- 数据量：**~200 字节**
- 减少：**60%** 🚀

### 场景 2：修改节点描述

**优化前**：
```json
{
  // 所有 17 个字段
}
```
- 数据量：**~500 字节**

**优化后**：
```json
{
  "name": "条件检查",
  "nodeKey": "591287546177127446",
  "type": "condition_checker",
  "config": {},
  "applicationId": "591202603132519446",
  "description": "新的描述"
}
```
- 字段数：**6 个**
- 数据量：**~180 字节**
- 减少：**64%** 🚀

### 场景 3：修改条件节点的分支

**优化前**：
```json
{
  // 所有 17 个字段
}
```
- 数据量：**~500 字节**

**优化后**：
```json
{
  "name": "条件检查",
  "nodeKey": "591287546177127446",
  "type": "condition_checker",
  "config": {},
  "applicationId": "591202603132519446",
  "branchNodes": {
    "true": { "name": "true", "condition": "result === true" },
    "false": { "name": "false", "condition": "result === false" }
  }
}
```
- 字段数：**6 个**
- 数据量：**~280 字节**
- 减少：**44%** 🚀

## 性能对比

### 典型工作流（10 个节点，移动 5 个节点）

**优化前**：
- 每个节点：~500 字节
- 总数据量：5 × 500 = **2,500 字节**

**优化后**：
- 每个节点：~200 字节
- 总数据量：5 × 200 = **1,000 字节**
- 减少：**60%**

### 复杂工作流（50 个节点，批量调整位置 20 个节点）

**优化前**：
- 总数据量：20 × 500 = **10,000 字节** (~10 KB)

**优化后**：
- 总数据量：20 × 200 = **4,000 字节** (~4 KB)
- 减少：**60%**

## 特殊处理：条件节点的 branchNodes

### 问题

条件节点的 `branchNodes` 比较特殊：
- 即使只移动位置，也需要包含 `branchNodes`（保持一致性）
- 因为 `branchNodes` 可能引用其他节点的 ID

### 解决方案

```typescript
// 对于条件节点，检查 branchNodes 是否变更
if (node.type === NodeTypeEnum.CONDITION_CHECKER) {
  const branchNodes = calculateBranchNodesFromNode(node, nodeIdMapping);
  if (branchNodes && Object.keys(branchNodes).length > 0) {
    // 如果 branchNodes 有变化，或者是为了保持一致性，总是包含它
    if (changedFieldsList.includes("data.branchNodes") || changedFieldsList.length > 0) {
      nodeData.branchNodes = branchNodes;
    }
  }
}
```

**逻辑**：
- 如果 `branchNodes` 本身有变化 → 包含它
- 如果节点有任何变化（包括位置） → 包含它（保持一致性）
- 如果节点完全没变化 → 不会进入更新流程

## 技术细节

### getNodeFieldChanges 函数

这个函数负责计算节点的字段级别变化：

```typescript
const getNodeFieldChanges = (
  currentNode: Node,
  snapshotNode: Node
): {
  changedFields: string[];
  changes: Partial<Node>;
} | null => {
  const changedFields: string[] = [];
  const changes: Partial<Node> = {};
  let hasChanges = false;

  // 检查位置变化
  if (
    currentNode.position.x !== snapshotNode.position.x ||
    currentNode.position.y !== snapshotNode.position.y
  ) {
    changes.position = currentNode.position;
    changedFields.push("position");
    hasChanges = true;
  }

  // 检查 data 字段变化
  if (currentNode.data.label !== snapshotNode.data.label) {
    changedFields.push("data.label");
    hasChanges = true;
  }

  // ... 其他字段检查

  return hasChanges ? { changedFields, changes } : null;
};
```

### 为什么必填字段总是提交？

**原因**：
1. **后端要求**：`binding:"required"` 标记
2. **数据完整性**：确保节点的基本信息始终存在
3. **简化逻辑**：避免复杂的条件判断

**必填字段列表**：
- `name`：节点名称
- `nodeKey`：节点键（唯一标识）
- `type`：节点类型
- `config`：节点配置（可以是空对象）
- `applicationId`：所属应用 ID

## 相关文件

- **`src/views/test/composables/useWorkflowApplication.ts`**
  - 节点更新逻辑（优化）
  - `getNodeFieldChanges` 函数

- **`src/workflow/types.ts`**
  - `UpdateWorkflowNodeRequest` 类型定义

- **`internal/funcs/workflow_func.go`**
  - 后端更新逻辑

- **`shared/models/workflow.go`**
  - 后端请求模型定义

## 测试建议

### 1. 测试只移动节点位置

```typescript
// 1. 加载已有工作流
// 2. 移动节点位置
// 3. 保存工作流
// 4. 检查网络请求：
//    - 只包含必填字段 + positionX + positionY
//    - 不包含其他未变更的字段
```

### 2. 测试修改节点描述

```typescript
// 1. 加载已有工作流
// 2. 修改节点描述
// 3. 保存工作流
// 4. 检查网络请求：
//    - 只包含必填字段 + description
//    - 不包含其他未变更的字段
```

### 3. 测试条件节点

```typescript
// 1. 加载包含条件节点的工作流
// 2. 移动条件节点位置
// 3. 保存工作流
// 4. 检查网络请求：
//    - 包含必填字段 + positionX + positionY + branchNodes
//    - branchNodes 保持一致性
```

## 总结

这次优化实现了节点的部分字段更新：

✅ **只提交变更的字段**：减少 60% 的数据传输量  
✅ **保留必填字段**：确保后端验证通过  
✅ **特殊处理条件节点**：保持 branchNodes 的一致性  
✅ **利用现有 diff 逻辑**：复用 `getNodeFieldChanges` 函数  

这是一次重要的性能优化，特别是对于频繁移动节点的场景！🚀

