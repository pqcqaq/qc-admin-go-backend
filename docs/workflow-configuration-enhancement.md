# 工作流配置增强 - 修复总结

## 修复的问题

### 1. branchNodes 保存逻辑修复

**问题描述：**
- 条件节点的 `branchNodes` 字段无法正确保存分支和连线关系
- 在创建新节点时，目标节点可能还没有数据库 ID，导致 `branchNodes` 映射错误

**解决方案：**
- 修改 `calculateBranchNodesFromEdges` 函数，支持传入 `nodeIdMapping` 参数
- 在所有节点和边保存完成后，统一更新条件节点的 `branchNodes`
- 使用 `nodeIdMapping` 将临时 ID 映射到数据库 ID

**修改的文件：**
- `src/views/test/composables/useWorkflowApplication.ts`

**关键代码：**
```typescript
// 更新条件节点的 branchNodes（在所有节点和边都保存完成后）
const conditionNodes = currentNodes.filter(
  n => n.type === NodeTypeEnum.CONDITION_CHECKER
);

for (const node of conditionNodes) {
  // 使用 nodeIdMapping 计算 branchNodes
  const branchNodes = calculateBranchNodesFromEdges(node.id, nodeIdMapping);
  
  if (branchNodes && Object.keys(branchNodes).length > 0) {
    // 更新节点的 branchNodes 字段
    await updateWorkflowNode(node.id, nodeData);
  }
}
```

---

### 2. 新增节点 ID 更新问题修复

**问题描述：**
- 新增节点保存成功后，前端的节点 ID（临时 ID）没有更新为后端返回的数据库 ID
- 导致第二次保存时出现错误

**解决方案：**
- 在保存成功后，使用 `workflow.updateNodeId()` 方法更新节点 ID
- 该方法会自动更新相关边的 source 和 target

**修改的文件：**
- `src/views/test/composables/useWorkflowApplication.ts`

**关键代码：**
```typescript
// 更新 Vue Flow 中的节点和边 ID（将临时 ID 替换为数据库 ID）
if (nodeIdMapping.size > 0) {
  debugLog("工作流保存", "开始更新节点和边的 ID...");

  // 使用 updateNodeId 方法更新节点 ID（会自动更新相关的边）
  for (const [tempId, dbId] of nodeIdMapping) {
    workflow.updateNodeId(tempId, dbId);
    debugLog("工作流保存", `✅ 更新节点 ID: ${tempId} -> ${dbId}`);
  }

  debugLog("工作流保存", `✅ ID 更新完成`);
}
```

---

### 3. 基础信息配置增强

**新增字段：**
- `async` - 异步执行开关
- `retryCount` - 重试次数
- `prompt` - 提示词（仅 LLM 节点显示）

**修改的文件：**
- `src/views/test/components/PropertiesPanel/sections/BasicInfoSection.vue`

**效果：**
- 用户可以在基础信息面板中配置节点的异步执行、重试次数等通用属性
- LLM 节点可以直接在基础信息中配置提示词

---

### 4. API 配置面板

**新增文件：**
- `src/views/test/components/PropertiesPanel/sections/ApiConfigSection.vue`

**功能：**
- API URL 配置
- HTTP 方法选择（GET/POST/PUT/PATCH/DELETE）
- 请求头配置（JSON 格式）
- 请求体配置（JSON 格式，仅 POST/PUT/PATCH）
- 查询参数配置（JSON 格式）
- 请求超时配置
- 响应路径配置（用于提取响应数据）

**适用节点：**
- `API_CALLER` 节点

---

### 5. 数据处理器配置面板

**新增文件：**
- `src/views/test/components/PropertiesPanel/sections/ProcessorConfigSection.vue`

**功能：**
- 处理器语言选择（JavaScript/Python/Go/Java）
- 处理器代码编辑器（带语法高亮）
- 代码模板（根据选择的语言自动切换）
- 代码说明（输入参数、返回值）

**适用节点：**
- `DATA_PROCESSOR` 节点

---

### 6. LLM 配置面板

**新增文件：**
- `src/views/test/components/PropertiesPanel/sections/LlmConfigSection.vue`

**功能：**
- 提示词配置（支持变量替换）
- 模型名称配置
- 温度（Temperature）配置（0-2）
- 最大 Token 数配置
- Top P 配置（0-1）
- 系统提示词配置
- 变量替换说明（`{{input}}`, `{{context.xxx}}`, `{{env.xxx}}`）

**适用节点：**
- `LLM_CALLER` 节点

---

## 数据结构

### 后端返回的节点数据类型

```typescript
export interface WorkflowNodeResponse {
  id: string;
  createTime: string;
  updateTime: string;
  name: string;
  nodeKey: string;
  type: WorkflowNodeType;
  description?: string;
  prompt?: string;                      // LLM 提示词
  config: Record<string, any>;          // 通用配置
  applicationId: string;
  processorLanguage?: string;           // 处理器语言
  processorCode?: string;               // 处理器代码
  nextNodeId?: string;
  parentNodeId?: string;
  branchNodes?: Record<string, number>; // 分支节点映射
  parallelConfig?: Record<string, any>; // 并行配置
  apiConfig?: Record<string, any>;      // API 配置
  async: boolean;                       // 异步执行
  timeout: number;                      // 超时时间
  retryCount: number;                   // 重试次数
  positionX: number;
  positionY: number;
  color?: string;
}
```

### 前端节点数据映射

所有后端字段都已正确映射到前端的 `NodeData` 接口，并在相应的配置面板中提供编辑功能。

---

## 保存流程

1. **创建新节点**（不包含 branchNodes）
2. **更新修改的节点**（不包含 branchNodes）
3. **删除节点**
4. **创建新边**（使用 nodeIdMapping 映射临时 ID）
5. **更新修改的边**
6. **删除边**
7. **更新条件节点的 branchNodes**（使用 nodeIdMapping）
8. **更新前端节点和边的 ID**（将临时 ID 替换为数据库 ID）
9. **更新 snapshot**（用于检测未保存的更改）

---

## 测试建议

1. **测试新增节点保存**
   - 创建新节点 → 保存 → 再次保存（验证 ID 更新）
   
2. **测试条件节点分支**
   - 创建条件节点 → 添加分支 → 连接到新节点 → 保存 → 验证 branchNodes

3. **测试并行节点任务**
   - 创建并行节点 → 添加任务 → 连接到新节点 → 保存 → 验证 parallelConfig.threads

4. **测试配置面板**
   - LLM 节点：配置 prompt、model、temperature 等
   - API 节点：配置 URL、method、headers 等
   - 数据处理器节点：配置 language、code 等

5. **测试基础信息**
   - 配置 async、retryCount、timeout 等通用属性

---

## 注意事项

1. **ID 映射**：新创建的节点在保存成功后会自动更新 ID，无需手动处理
2. **branchNodes 计算**：条件节点的 branchNodes 会在所有节点和边保存完成后自动计算和更新
3. **JSON 格式**：API 配置中的 headers、body、query 需要使用有效的 JSON 格式
4. **变量替换**：LLM 提示词支持 `{{input}}`、`{{context.xxx}}`、`{{env.xxx}}` 等变量

---

## 相关文件

### 修改的文件
- `src/views/test/composables/useWorkflowApplication.ts`
- `src/views/test/components/PropertiesPanel/sections/BasicInfoSection.vue`
- `src/views/test/components/PropertiesPanel/index.vue`

### 新增的文件
- `src/views/test/components/PropertiesPanel/sections/ApiConfigSection.vue`
- `src/views/test/components/PropertiesPanel/sections/ProcessorConfigSection.vue`
- `src/views/test/components/PropertiesPanel/sections/LlmConfigSection.vue`

---

## 完成时间

2025-10-31

