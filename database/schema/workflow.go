package schema

import (
	"go-backend/database/mixins"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// WorkflowApplication 工作流应用
type WorkflowApplication struct {
	ent.Schema
}

func (WorkflowApplication) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_applications"},
	}
}

func (WorkflowApplication) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (WorkflowApplication) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Comment("工作流应用名称"),
		field.String("description").Optional().Comment("工作流应用描述"),
		field.Uint64("start_node_id").Optional().Comment("起始节点ID（旧架构，保留兼容）"),
		field.String("client_secret").Comment("客户端密钥"),
		field.JSON("variables", map[string]interface{}{}).Optional().Comment("全局变量定义"),
		field.Uint("version").Default(1).Comment("版本号"),
		field.Enum("status").Values("draft", "published", "archived").Default("draft").Comment("状态"),
	}
}

func (WorkflowApplication) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("client_secret").Unique(),
		index.Fields("status"),
	}
}

func (WorkflowApplication) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("nodes", WorkflowNode.Type),
		edge.To("edges", WorkflowEdge.Type),
		edge.To("executions", WorkflowExecution.Type),
	}
}

// WorkflowNode 工作流节点
type WorkflowNode struct {
	ent.Schema
}

func (WorkflowNode) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_nodes"},
	}
}

func (WorkflowNode) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

/*
graph TD

	Start[用户输入节点] --> Process[数据处理节点]

	Process --> Parallel{并行执行节点}

	Parallel -.并行.-> API1[API调用1<br/>async]
	Parallel -.并行.-> API2[API调用2<br/>async]
	Parallel -.并行.-> Todo[待办生成<br/>async]

	API1 --> Merge[汇聚点]
	API2 --> Merge
	Todo --> Merge

	Merge --> Condition{条件检查}

	Condition -->|满足| Loop[循环节点]
	Condition -->|不满足| End[结束节点]

	Loop --> API3[API调用节点]
	Loop --> Condition
	API3 --> Loop

	style Start fill:#a8e6cf
	style Parallel fill:#dda0dd
	style Condition fill:#ffd3b6
	style Loop fill:#ffaaa5
	style End fill:#ff8b94
*/
func (WorkflowNode) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().Comment("节点名称"),
		field.String("node_key").NotEmpty().Comment("节点唯一标识符"),
		field.Enum("type").Values(
			"user_input",          // 用户输入节点
			"todo_task_generator", // 待办任务生成器节点
			"condition_checker",   // 条件检查节点
			"api_caller",          // API调用节点
			"data_processor",      // 数据处理节点
			"while_loop",          // 循环节点
			"end_node",            // 结束节点
			"parallel_executor",   // 并行执行节点（可选择部分进行处理）
			"llm_caller",          // LLM调用节点
		).Comment("节点类型"),
		field.String("description").Optional().Comment("节点描述"),
		field.Text("prompt").Optional().Comment("节点提示词"),
		field.JSON("config", map[string]interface{}{}).Comment("节点配置"),
		field.Uint64("application_id").Comment("所属工作流应用ID"),
		field.String("processor_language").Optional().Comment("处理器语言"),
		field.Text("processor_code").Optional().Comment("代码处理器"),
		field.Uint64("next_node_id").Optional().Comment("下一个节点ID"),
		field.Uint64("parent_node_id").Optional().Comment("父节点ID"),
		field.JSON("branch_nodes", map[string]uint64{}).Optional().Comment("分支节点映射"),
		field.JSON("parallel_config", map[string]interface{}{}).Optional().Comment("并行执行配置"),
		field.JSON("api_config", map[string]interface{}{}).Optional().Comment("API调用配置"),
		field.Bool("async").Default(false).Comment("是否异步执行"),
		field.Int("timeout").Default(30).Comment("超时时间(秒)"),
		field.Int("retry_count").Default(0).Comment("重试次数"),
		field.Float("position_x").Default(0).Comment("画布X坐标"),
		field.Float("position_y").Default(0).Comment("画布Y坐标"),
		field.String("color").Optional().Comment("节点颜色"),
	}
}

func (WorkflowNode) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "node_key").Unique(),
		index.Fields("next_node_id"),
		index.Fields("parent_node_id"),
		index.Fields("type"),
	}
}

func (WorkflowNode) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", WorkflowApplication.Type).
			Ref("nodes").Unique().
			Field("application_id").
			Required(),
		edge.To("executions", WorkflowNodeExecution.Type),
		edge.To("outgoing_edges", WorkflowEdge.Type),
		edge.To("incoming_edges", WorkflowEdge.Type),
	}
}

// WorkflowEdge 工作流边（连接）
type WorkflowEdge struct {
	ent.Schema
}

func (WorkflowEdge) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_edges"},
	}
}

func (WorkflowEdge) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (WorkflowEdge) Fields() []ent.Field {
	return []ent.Field{
		field.String("edge_key").NotEmpty().Comment("边唯一标识符"),
		field.Uint64("application_id").Comment("所属工作流应用ID"),
		field.Uint64("source_node_id").Comment("源节点ID"),
		field.Uint64("target_node_id").Comment("目标节点ID"),
		field.String("source_handle").Optional().Comment("源节点连接点ID"),
		field.String("target_handle").Optional().Comment("目标节点连接点ID"),
		field.Enum("type").Values(
			"default",  // 默认连接
			"branch",   // 分支连接
			"parallel", // 并行连接
		).Default("default").Comment("边类型"),
		field.String("label").Optional().Comment("边标签"),
		field.String("branch_name").Optional().Comment("分支名称（用于 condition_checker）"),
		field.Bool("animated").Default(false).Comment("是否动画"),
		field.JSON("style", map[string]interface{}{}).Optional().Comment("边样式"),
		field.JSON("data", map[string]interface{}{}).Optional().Comment("边数据"),
	}
}

func (WorkflowEdge) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "edge_key").Unique(),
		index.Fields("source_node_id"),
		index.Fields("target_node_id"),
		index.Fields("type"),
	}
}

func (WorkflowEdge) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", WorkflowApplication.Type).
			Ref("edges").Unique().
			Field("application_id").
			Required(),
		edge.From("source_node", WorkflowNode.Type).
			Ref("outgoing_edges").Unique().
			Field("source_node_id").
			Required(),
		edge.From("target_node", WorkflowNode.Type).
			Ref("incoming_edges").Unique().
			Field("target_node_id").
			Required(),
	}
}

// WorkflowExecution 工作流执行记录
type WorkflowExecution struct {
	ent.Schema
}

func (WorkflowExecution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_executions"},
	}
}

func (WorkflowExecution) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (WorkflowExecution) Fields() []ent.Field {
	return []ent.Field{
		field.String("execution_id").NotEmpty().Unique().Comment("执行ID"),
		field.Uint64("application_id").Comment("工作流应用ID"),
		field.Enum("status").Values(
			"pending",   // 等待中
			"running",   // 执行中
			"completed", // 已完成
			"failed",    // 失败
			"cancelled", // 已取消
			"timeout",   // 超时
		).Default("pending").Comment("执行状态"),
		field.JSON("input", map[string]interface{}{}).Optional().Comment("执行输入"),
		field.JSON("output", map[string]interface{}{}).Optional().Comment("执行输出"),
		field.JSON("context", map[string]interface{}{}).Optional().Comment("执行上下文"),
		field.Time("started_at").Optional().Comment("开始时间"),
		field.Time("finished_at").Optional().Comment("结束时间"),
		field.Int("duration_ms").Default(0).Comment("执行时长(毫秒)"),
		field.Int("total_tokens").Default(0).Comment("总Token消耗"),
		field.Float("total_cost").Default(0).Comment("总成本"),
		field.String("error_message").Optional().Comment("错误信息"),
		field.Text("error_stack").Optional().Comment("错误堆栈"),
		field.String("triggered_by").Optional().Comment("触发者"),
		field.String("trigger_source").Optional().Comment("触发源"),
	}
}

func (WorkflowExecution) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("execution_id").Unique(),
		index.Fields("application_id", "status"),
		index.Fields("started_at"),
		index.Fields("triggered_by"),
	}
}

func (WorkflowExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("application", WorkflowApplication.Type).
			Ref("executions").Unique().
			Field("application_id").
			Required(),
		edge.To("node_executions", WorkflowNodeExecution.Type),
	}
}

// WorkflowNodeExecution 节点执行记录
type WorkflowNodeExecution struct {
	ent.Schema
}

func (WorkflowNodeExecution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_node_executions"},
	}
}

func (WorkflowNodeExecution) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (WorkflowNodeExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("execution_id").Comment("工作流执行ID"),
		field.Uint64("node_id").Comment("节点ID"),
		field.String("node_name").Comment("节点名称"),
		field.String("node_type").Comment("节点类型"),
		field.Enum("status").Values(
			"pending",
			"running",
			"completed",
			"failed",
			"skipped",
			"timeout",
		).Default("pending").Comment("执行状态"),
		field.JSON("input", map[string]interface{}{}).Optional().Comment("输入数据"),
		field.JSON("output", map[string]interface{}{}).Optional().Comment("输出数据"),
		field.JSON("extra", map[string]interface{}{}).Optional().Comment("额外信息"),
		field.Time("started_at").Optional().Comment("开始时间"),
		field.Time("finished_at").Optional().Comment("结束时间"),
		field.Int("duration_ms").Default(0).Comment("执行时长(毫秒)"),
		field.Int("prompt_tokens").Default(0).Comment("提示词Token"),
		field.Int("completion_tokens").Default(0).Comment("补全Token"),
		field.Int("total_tokens").Default(0).Comment("总Token"),
		field.Float("cost").Default(0).Comment("成本"),
		field.String("model").Optional().Comment("使用的模型"),
		field.String("error_message").Optional().Comment("错误信息"),
		field.Text("error_stack").Optional().Comment("错误堆栈"),
		field.Int("retry_count").Default(0).Comment("重试次数"),
		field.Bool("is_async").Default(false).Comment("是否异步执行"),
		field.Uint64("parent_execution_id").Optional().Comment("父节点执行ID"),
	}
}

func (WorkflowNodeExecution) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("execution_id", "node_id"),
		index.Fields("execution_id", "status"),
		index.Fields("started_at"),
		index.Fields("parent_execution_id"),
	}
}

func (WorkflowNodeExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow_execution", WorkflowExecution.Type).
			Ref("node_executions").Unique().
			Field("execution_id").
			Required(),
		edge.From("node", WorkflowNode.Type).
			Ref("executions").Unique().
			Field("node_id").
			Required(),
	}
}

// WorkflowVersion 工作流版本管理
type WorkflowVersion struct {
	ent.Schema
}

func (WorkflowVersion) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_versions"},
	}
}

func (WorkflowVersion) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (WorkflowVersion) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("application_id").Comment("工作流应用ID"),
		field.Uint("version").Comment("版本号"), // 在保存时自动增加
		field.JSON("snapshot", map[string]interface{}{}).Comment("版本快照"),
		field.String("change_log").Optional().Comment("变更日志"), // autosave或者手动保存时填写
	}
}

func (WorkflowVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("application_id", "version").Unique(),
	}
}

// WorkflowExecutionLog 执行日志
type WorkflowExecutionLog struct {
	ent.Schema
}

func (WorkflowExecutionLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_execution_logs"},
	}
}

func (WorkflowExecutionLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.BaseMixin{},
	}
}

func (WorkflowExecutionLog) Fields() []ent.Field {
	return []ent.Field{
		field.Uint64("execution_id").Comment("工作流执行ID"),
		field.Uint64("node_execution_id").Optional().Comment("节点执行ID"),
		field.Enum("level").Values("debug", "info", "warn", "error").Default("info").Comment("日志级别"),
		field.Text("message").Comment("日志消息"),
		field.JSON("metadata", map[string]interface{}{}).Optional().Comment("元数据"),
		field.Time("logged_at").Default(time.Now).Comment("记录时间"),
	}
}

func (WorkflowExecutionLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("execution_id", "logged_at"),
		index.Fields("node_execution_id"),
		index.Fields("level"),
	}
}
