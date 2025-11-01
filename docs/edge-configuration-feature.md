# 边配置功能实现总结

## 功能概述

实现了选中边（Edge）并修改其配置数据的功能，用户可以通过点击连线来编辑连线的各种属性。

## 实现的功能

### 1. 边选中状态管理

**文件：** `src/views/test/composables/useWorkflow.ts`

**新增状态：**
- `selectedEdgeId` - 当前选中的边 ID
- `selectedEdge` - 当前选中的边对象（computed）

**新增方法：**
- `setSelectedEdgeId(edgeId: string | null)` - 设置选中的边

**行为：**
- 选中边时，自动清除节点的选中状态
- 选中节点时，自动清除边的选中状态
- 更新 Vue Flow 中边的 `selected` 属性，实现视觉高亮

**关键代码：**
```typescript
const selectedEdgeId = ref<string | null>(null);

const selectedEdge = computed(() => {
  if (!selectedEdgeId.value) return null;
  return findEdge(selectedEdgeId.value);
});

const setSelectedEdgeId = (edgeId: string | null) => {
  selectedEdgeId.value = edgeId;
  // 清除节点的选中状态
  selectedNodeId.value = null;

  // 清除所有节点的选中状态
  const allNodes = getNodes.value;
  const updatedNodes = allNodes.map(node => ({
    ...node,
    selected: false
  }));
  setNodes(updatedNodes);

  // 更新所有边的选中状态
  const allEdges = getEdges.value;
  const updatedEdges = allEdges.map(edge => ({
    ...edge,
    selected: edge.id === edgeId
  }));
  setEdges(updatedEdges);
};
```

---

### 2. 边点击事件处理

**文件：** `src/views/test/components/WorkflowEditor.vue`

**修改：**
- 在 `onEdgeClick` 事件中调用 `setSelectedEdgeId`
- 添加 `selectedEdge` computed 属性
- 传递 `selectedEdge` 给 PropertiesPanel
- 添加 `handleUpdateEdge` 和 `handleDeleteEdge` 方法

**关键代码：**
```typescript
// 获取选中的边
const selectedEdge = computed(() => props.workflow.selectedEdge.value || null);

// 边点击事件
function onEdgeClick({ edge }: { edge: any }) {
  props.workflow?.setSelectedEdgeId(edge.id);
}

// 更新边
async function handleUpdateEdge(edgeId: string, updates: Partial<Edge>) {
  await props.workflow?.updateEdge(edgeId, updates);
}

// 删除边
async function handleDeleteEdge(edgeId: string) {
  await props.workflow?.deleteEdge(edgeId);
}
```

---

### 3. 边配置 Section 组件

**文件：** `src/views/test/components/PropertiesPanel/sections/EdgeConfigSection.vue`

**功能：**
- ✅ **标签（label）** - 连线上显示的文本
- ✅ **分支名称（branchName）** - 条件节点的分支标识（只读）
- ✅ **动画效果（animated）** - 启用/禁用流动动画
- ✅ **连线类型（type）** - 平滑阶梯/直线/贝塞尔曲线/阶梯
- ✅ **样式配置（style）** - JSON 格式的样式配置（stroke, strokeWidth 等）
- ✅ **自定义数据（data）** - JSON 格式的自定义数据
- ✅ **连线信息** - 显示连线 ID、源节点、目标节点、Handle 等信息

**UI 组件：**
- 文本输入框（标签）
- 开关（动画效果）
- 下拉选择（连线类型）
- 多行文本框（样式配置、自定义数据）
- 信息展示区（连线详细信息）

**JSON 验证：**
- 样式配置和自定义数据支持 JSON 格式
- 输入错误的 JSON 会显示错误提示
- 空值会清除对应的配置

**关键代码：**
```vue
<template>
  <el-collapse-item name="edge-config" title="连线配置">
    <!-- 标签 -->
    <el-form-item label="标签">
      <el-input
        :model-value="edge.label as string"
        placeholder="请输入连线标签"
        @update:modelValue="updateEdgeData('label', $event)"
      />
    </el-form-item>

    <!-- 动画效果 -->
    <el-form-item label="动画效果">
      <el-switch
        :model-value="edge.animated"
        @change="updateEdgeData('animated', $event)"
      />
    </el-form-item>

    <!-- 连线类型 -->
    <el-form-item label="连线类型">
      <el-select
        :model-value="edge.type || 'smoothstep'"
        @change="updateEdgeData('type', $event)"
      >
        <el-option label="平滑阶梯" value="smoothstep" />
        <el-option label="直线" value="straight" />
        <el-option label="贝塞尔曲线" value="default" />
        <el-option label="阶梯" value="step" />
      </el-select>
    </el-form-item>

    <!-- 样式配置 -->
    <el-form-item label="样式配置">
      <el-input
        :model-value="styleJson"
        type="textarea"
        :rows="4"
        placeholder='{"stroke": "#ff0000", "strokeWidth": 2}'
        @update:modelValue="updateStyle"
      />
    </el-form-item>

    <!-- 自定义数据 -->
    <el-form-item label="自定义数据">
      <el-input
        :model-value="dataJson"
        type="textarea"
        :rows="4"
        placeholder='{"key": "value"}'
        @update:modelValue="updateData"
      />
    </el-form-item>
  </el-collapse-item>
</template>
```

---

### 4. PropertiesPanel 更新

**文件：** `src/views/test/components/PropertiesPanel/index.vue`

**修改：**
- 添加 `selectedEdge` prop（可选）
- 添加 `updateEdge` 和 `deleteEdge` emit 事件
- 导入 `EdgeConfigSection` 组件
- 更新面板标题（节点配置/连线配置/属性配置）
- 添加删除边按钮
- 添加边配置表单显示逻辑
- 添加 `handleUpdateEdge` 和 `handleDeleteEdge` 方法

**关键代码：**
```vue
<template>
  <div class="properties-panel">
    <div class="panel-header">
      <h3 class="panel-title">
        {{ selectedNode ? "节点配置" : selectedEdge ? "连线配置" : "属性配置" }}
      </h3>
      <!-- 删除节点按钮 -->
      <el-button v-if="selectedNode" ... @click="handleDeleteNode" />
      <!-- 删除边按钮 -->
      <el-button v-if="selectedEdge" ... @click="handleDeleteEdge" />
    </div>

    <!-- 未选中节点或边时的提示 -->
    <div v-if="!selectedNode && !selectedEdge" class="empty-state">
      <p class="empty-text">请选择一个节点或连线</p>
    </div>

    <!-- 边属性表单 -->
    <div v-else-if="selectedEdge" class="properties-form">
      <el-collapse v-model="activeCollapse">
        <EdgeConfigSection
          :edge="selectedEdge"
          @update-edge="handleUpdateEdge"
        />
      </el-collapse>
    </div>

    <!-- 节点属性表单 -->
    <div v-else-if="selectedNode" class="properties-form">
      ...
    </div>
  </div>
</template>
```

---

## 数据结构

### 后端接口

```typescript
export interface UpdateWorkflowEdgeRequest {
  edgeKey?: string;
  sourceHandle?: string;
  targetHandle?: string;
  type?: WorkflowEdgeType;
  label?: string;
  branchName?: string;
  animated?: boolean;
  style?: Record<string, any>;
  data?: Record<string, any>;
}
```

### 可编辑字段

| 字段 | 类型 | 说明 | 编辑方式 |
|------|------|------|----------|
| `label` | string | 连线标签 | 文本输入框 |
| `branchName` | string | 分支名称 | 只读（条件节点专用） |
| `animated` | boolean | 动画效果 | 开关 |
| `type` | string | 连线类型 | 下拉选择 |
| `style` | object | 样式配置 | JSON 文本框 |
| `data` | object | 自定义数据 | JSON 文本框 |

---

## 使用流程

1. **选中边**
   - 点击画布中的连线
   - 连线高亮显示
   - 右侧属性面板显示"连线配置"

2. **编辑边属性**
   - 修改标签、动画、类型等属性
   - 输入 JSON 格式的样式和数据
   - 实时保存到 Vue Flow

3. **删除边**
   - 点击属性面板右上角的删除按钮
   - 或使用右键菜单删除

4. **取消选中**
   - 点击画布空白处
   - 或选中其他节点/边

---

## 样式配置示例

### 修改连线颜色和宽度

```json
{
  "stroke": "#ff0000",
  "strokeWidth": 3
}
```

### 修改连线为虚线

```json
{
  "stroke": "#409eff",
  "strokeWidth": 2,
  "strokeDasharray": "5,5"
}
```

---

## 自定义数据示例

### 添加业务数据

```json
{
  "priority": "high",
  "condition": "value > 100",
  "description": "高优先级分支"
}
```

---

## 测试建议

1. **基础功能测试**
   - 点击连线，验证选中状态
   - 修改标签，验证显示效果
   - 切换动画，验证流动效果
   - 修改连线类型，验证样式变化

2. **JSON 配置测试**
   - 输入有效的 JSON，验证应用成功
   - 输入无效的 JSON，验证错误提示
   - 清空 JSON，验证配置清除

3. **交互测试**
   - 选中边后选中节点，验证状态切换
   - 选中节点后选中边，验证状态切换
   - 删除边，验证删除成功

4. **条件节点测试**
   - 选中条件节点的分支连线
   - 验证 branchName 字段显示且不可编辑

---

## 相关文件

### 修改的文件
- `src/views/test/composables/useWorkflow.ts` - 添加边选中状态管理
- `src/views/test/components/WorkflowEditor.vue` - 添加边点击事件处理
- `src/views/test/components/PropertiesPanel/index.vue` - 支持边配置显示

### 新增的文件
- `src/views/test/components/PropertiesPanel/sections/EdgeConfigSection.vue` - 边配置组件

---

## 完成时间

2025-10-31

