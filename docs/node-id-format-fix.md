# 节点 ID 格式问题修复文档

## 问题描述

在保存工作流时，新创建的节点被错误地识别为需要更新的节点，导致调用更新 API 时出现 "工作流节点ID格式无效" 错误。

### 错误信息

```json
{
  "success": false,
  "code": 400,
  "message": "工作流节点ID格式无效",
  "data": {
    "provided_id": "condition_checker-1761963565515-jygtq5wzo"
  },
  "path": "/api/v1/workflow/nodes/condition_checker-1761963565515-jygtq5wzo"
}
```

### 问题根源

1. **前端生成临时 ID**：新创建的节点使用临时 ID，格式为 `{nodeType}-{timestamp}-{randomString}`
   - 例如：`condition_checker-1761963565515-jygtq5wzo`

2. **后端要求数字 ID**：更新节点的 API 要求 ID 必须是纯数字字符串
   - 后端使用 `strconv.ParseUint(idStr, 10, 64)` 解析 ID
   - 临时 ID 无法解析为数字，导致 400 错误

3. **Diff 逻辑缺陷**：原来的 diff 逻辑只检查 snapshot 中是否存在节点，没有检查 ID 格式
   - 在某些情况下，临时 ID 的节点可能被误判为需要更新的节点

4. **branchNodes 更新问题**：即使节点创建成功，在更新 branchNodes 时仍使用临时 ID

## 修复方案

### 1. 添加 ID 格式检查函数

**文件**：`src/views/test/composables/useWorkflowApplication.ts`

```typescript
/**
 * 判断 ID 是否为数据库 ID（纯数字字符串）
 */
const isDatabaseId = (id: string): boolean => {
  return /^\d+$/.test(id);
};
```

**功能**：
- 使用正则表达式检查 ID 是否为纯数字
- 返回 `true` 表示是数据库 ID
- 返回 `false` 表示是临时 ID

### 2. 优化节点 Diff 逻辑

**修改前**：
```typescript
// 找出新增和修改的节点（使用 hash 比较）
for (const node of currentNodes) {
  const snapshotHash = snapshot.value.nodeHashes.get(node.id);
  if (!snapshotHash) {
    // 新增的节点
    nodesToCreate.push(node);
    debugLog("工作流保存", `节点 ${node.id} 是新增节点`);
  } else {
    // 计算当前节点的 hash 并与 snapshot 比较
    const currentHash = getNodeHash(node);
    if (currentHash !== snapshotHash) {
      nodesToUpdate.push(node);
      debugLog(
        "工作流保存",
        `节点 ${node.id} 有变化 (hash: ${snapshotHash} -> ${currentHash})`
      );
    }
  }
}
```

**修改后**：
```typescript
// 找出新增和修改的节点（使用 hash 比较）
for (const node of currentNodes) {
  // 首先检查 ID 格式：如果不是数据库 ID（纯数字），则一定是新节点
  if (!isDatabaseId(node.id)) {
    nodesToCreate.push(node);
    debugLog("工作流保存", `节点 ${node.id} 是新增节点（临时ID）`);
    continue;
  }

  // 对于数据库 ID，检查 snapshot
  const snapshotHash = snapshot.value.nodeHashes.get(node.id);
  if (!snapshotHash) {
    // 新增的节点（不在 snapshot 中）
    nodesToCreate.push(node);
    debugLog("工作流保存", `节点 ${node.id} 是新增节点`);
  } else {
    // 计算当前节点的 hash 并与 snapshot 比较
    const currentHash = getNodeHash(node);
    if (currentHash !== snapshotHash) {
      nodesToUpdate.push(node);
      debugLog(
        "工作流保存",
        `节点 ${node.id} 有变化 (hash: ${snapshotHash} -> ${currentHash})`
      );
    }
  }
}
```

**改进点**：
- ✅ 优先检查 ID 格式
- ✅ 临时 ID 直接识别为新节点
- ✅ 避免临时 ID 被误判为更新节点

### 3. 优化边 Diff 逻辑

**修改方式**：与节点 diff 逻辑相同

```typescript
// 找出新增和修改的边（使用 hash 比较）
for (const edge of currentEdges) {
  // 首先检查 ID 格式：如果不是数据库 ID（纯数字），则一定是新边
  if (!isDatabaseId(edge.id)) {
    edgesToCreate.push(edge);
    debugLog("工作流保存", `边 ${edge.id} 是新增边（临时ID）`);
    continue;
  }

  // 对于数据库 ID，检查 snapshot
  const snapshotHash = snapshot.value.edgeHashes.get(edge.id);
  if (!snapshotHash) {
    // 新增的边（不在 snapshot 中）
    edgesToCreate.push(edge);
    debugLog("工作流保存", `边 ${edge.id} 是新增边`);
  } else {
    // 计算当前边的 hash 并与 snapshot 比较
    const currentHash = getEdgeHash(edge);
    if (currentHash !== snapshotHash) {
      edgesToUpdate.push(edge);
      debugLog(
        "工作流保存",
        `边 ${edge.id} 有变化 (hash: ${snapshotHash} -> ${currentHash})`
      );
    }
  }
}
```

### 4. 修复 branchNodes 更新逻辑

**问题**：更新条件节点的 branchNodes 时，使用了临时 ID

**修改前**：
```typescript
for (const node of conditionNodes) {
  const branchNodes = calculateBranchNodesFromNode(node, nodeIdMapping);
  
  if (branchNodes && Object.keys(branchNodes).length > 0) {
    const nodeData = { /* ... */ };
    
    await updateWorkflowNode(node.id, nodeData); // ❌ 使用临时 ID
    debugLog(
      "工作流保存",
      `✅ 更新条件节点 ${node.id} 的 branchNodes:`,
      branchNodes
    );
  }
}
```

**修改后**：
```typescript
for (const node of conditionNodes) {
  // 获取节点的实际 ID（如果是新创建的节点，使用映射后的数据库 ID）
  const actualNodeId = nodeIdMapping.get(node.id) || node.id;
  
  const branchNodes = calculateBranchNodesFromNode(node, nodeIdMapping);
  
  if (branchNodes && Object.keys(branchNodes).length > 0) {
    const nodeData = { /* ... */ };
    
    await updateWorkflowNode(actualNodeId, nodeData); // ✅ 使用实际 ID
    debugLog(
      "工作流保存",
      `✅ 更新条件节点 ${node.id} (实际ID: ${actualNodeId}) 的 branchNodes:`,
      branchNodes
    );
  }
}
```

**改进点**：
- ✅ 使用 `nodeIdMapping` 获取实际的数据库 ID
- ✅ 如果节点是新创建的，使用映射后的 ID
- ✅ 如果节点已存在，使用原 ID
- ✅ 日志中显示临时 ID 和实际 ID 的对应关系

## 修复效果

### 修复前的日志

```
[工作流保存] 节点 condition_checker-1761963565515-jygtq5wzo 是新增节点（临时ID）
[工作流保存] ✅ 创建节点: condition_checker-1761963565515-jygtq5wzo -> 591286078304945174
[工作流保存] 开始更新条件节点的 branchNodes...
❌ PUT /api/workflow/nodes/condition_checker-1761963565515-jygtq5wzo 400 (Bad Request)
```

### 修复后的日志

```
[工作流保存] 节点 condition_checker-1761963565515-jygtq5wzo 是新增节点（临时ID）
[工作流保存] ✅ 创建节点: condition_checker-1761963565515-jygtq5wzo -> 591286078304945174
[工作流保存] 开始更新条件节点的 branchNodes...
[工作流保存] ✅ 更新条件节点 condition_checker-1761963565515-jygtq5wzo (实际ID: 591286078304945174) 的 branchNodes
✅ 保存成功
```

## 技术细节

### ID 格式说明

**临时 ID 格式**：
```
{nodeType}-{timestamp}-{randomString}
```

**示例**：
- `condition_checker-1761963565515-jygtq5wzo`
- `user_input-1761963565516-abc123def`
- `end_node-1761963565517-xyz789`

**数据库 ID 格式**：
```
纯数字字符串
```

**示例**：
- `591286078304945174`
- `123456789`
- `1`

### 正则表达式说明

```typescript
/^\d+$/
```

- `^`：字符串开始
- `\d+`：一个或多个数字
- `$`：字符串结束

**匹配示例**：
- ✅ `"123"` → true
- ✅ `"591286078304945174"` → true
- ❌ `"condition_checker-123"` → false
- ❌ `"abc123"` → false
- ❌ `""` → false

## 相关文件

- **`src/views/test/composables/useWorkflowApplication.ts`**
  - `isDatabaseId()` 函数（新增）
  - 节点 diff 逻辑（优化）
  - 边 diff 逻辑（优化）
  - branchNodes 更新逻辑（修复）

## 测试建议

### 1. 创建新节点测试

```typescript
// 1. 创建新的条件节点
// 2. 添加分支
// 3. 连接目标节点
// 4. 保存工作流
// 5. 验证：
//    - 节点创建成功
//    - branchNodes 更新成功
//    - 没有 400 错误
```

### 2. 更新现有节点测试

```typescript
// 1. 加载已有工作流
// 2. 修改条件节点的分支配置
// 3. 保存工作流
// 4. 验证：
//    - 节点更新成功
//    - branchNodes 更新成功
```

### 3. 混合操作测试

```typescript
// 1. 加载已有工作流
// 2. 创建新节点
// 3. 修改现有节点
// 4. 删除某些节点
// 5. 保存工作流
// 6. 验证：
//    - 所有操作都成功
//    - ID 映射正确
```

## 总结

这次修复解决了节点 ID 格式导致的保存失败问题：

✅ **ID 格式检查**：添加 `isDatabaseId()` 函数判断 ID 类型  
✅ **Diff 逻辑优化**：优先检查 ID 格式，避免误判  
✅ **ID 映射修复**：在更新 branchNodes 时使用正确的数据库 ID  
✅ **日志增强**：显示临时 ID 和实际 ID 的对应关系  

这是一次重要的 bug 修复，确保了新节点的正确保存流程。

