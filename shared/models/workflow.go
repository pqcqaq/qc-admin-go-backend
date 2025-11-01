package models

import (
	"time"
)

// ============ WorkflowApplication Models ============

// WorkflowApplicationResponse 工作流应用响应结构
type WorkflowApplicationResponse struct {
	ID             string                  `json:"id"`
	CreateTime     string                  `json:"createTime"`
	UpdateTime     string                  `json:"updateTime"`
	Name           string                  `json:"name"`
	Description    string                  `json:"description,omitempty"`
	StartNodeID    string                  `json:"startNodeId,omitempty"` // 旧架构，保留兼容
	ClientSecret   string                  `json:"clientSecret"`
	Variables      map[string]interface{}  `json:"variables,omitempty"`
	Version        uint                    `json:"version"`
	Status         string                  `json:"status"`                   // draft, published, archived
	GraphData      string                  `json:"graphData,omitempty"`      // 新架构：完整的工作流图JSON
	ViewportConfig map[string]interface{}  `json:"viewportConfig,omitempty"` // 画布视口配置
	Nodes          []*WorkflowNodeResponse `json:"nodes,omitempty"`          // 旧架构，保留兼容
}

// CreateWorkflowApplicationRequest 创建工作流应用请求结构
type CreateWorkflowApplicationRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description,omitempty"`
	StartNodeID string                 `json:"startNodeId,omitempty"` // 可选，如果不提供则自动创建默认开始节点
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Status      string                 `json:"status,omitempty"` // draft, published, archived
}

// UpdateWorkflowApplicationRequest 更新工作流应用请求结构
type UpdateWorkflowApplicationRequest struct {
	Name           string                 `json:"name,omitempty"` // 改为可选
	Description    string                 `json:"description,omitempty"`
	StartNodeID    string                 `json:"startNodeId,omitempty"` // 旧架构，改为可选
	Variables      map[string]interface{} `json:"variables,omitempty"`
	Version        uint                   `json:"version,omitempty"`
	Status         string                 `json:"status,omitempty"`         // draft, published, archived
	GraphData      string                 `json:"graphData,omitempty"`      // 新架构：完整的工作流图JSON
	ViewportConfig map[string]interface{} `json:"viewportConfig,omitempty"` // 画布视口配置
}

// PageWorkflowApplicationRequest 分页查询工作流应用请求结构
type PageWorkflowApplicationRequest struct {
	PaginationRequest
	Name      string `form:"name" json:"name"`           // 按名称模糊搜索
	Status    string `form:"status" json:"status"`       // 按状态过滤
	BeginTime string `form:"beginTime" json:"beginTime"` // 开始时间
	EndTime   string `form:"endTime" json:"endTime"`     // 结束时间
}

// PageWorkflowApplicationResponse 分页查询工作流应用响应结构
type PageWorkflowApplicationResponse struct {
	Data       []*WorkflowApplicationResponse `json:"data"`
	Pagination Pagination                     `json:"pagination"`
}

// ============ WorkflowEdge Models ============

// WorkflowEdgeResponse 工作流边响应结构
type WorkflowEdgeResponse struct {
	ID            string                 `json:"id"`
	CreateTime    string                 `json:"createTime"`
	UpdateTime    string                 `json:"updateTime"`
	ApplicationID string                 `json:"applicationId"`
	SourceNodeID  string                 `json:"source"` // 源节点数据库ID
	TargetNodeID  string                 `json:"target"` // 目标节点数据库ID
	SourceHandle  string                 `json:"sourceHandle,omitempty"`
	TargetHandle  string                 `json:"targetHandle,omitempty"`
	Type          string                 `json:"type"` // default, branch, parallel
	Label         string                 `json:"label,omitempty"`
	BranchName    string                 `json:"branchName,omitempty"`
	Animated      bool                   `json:"animated"`
	Style         map[string]interface{} `json:"style,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// CreateWorkflowEdgeRequest 创建工作流边请求结构
type CreateWorkflowEdgeRequest struct {
	ApplicationID string                 `json:"applicationId" binding:"required"`
	SourceNodeID  string                 `json:"source" binding:"required"` // 源节点数据库ID
	TargetNodeID  string                 `json:"target" binding:"required"` // 目标节点数据库ID
	SourceHandle  string                 `json:"sourceHandle,omitempty"`
	TargetHandle  string                 `json:"targetHandle,omitempty"`
	Type          string                 `json:"type,omitempty"` // default, branch, parallel
	Label         string                 `json:"label,omitempty"`
	BranchName    string                 `json:"branchName,omitempty"`
	Animated      bool                   `json:"animated,omitempty"`
	Style         map[string]interface{} `json:"style,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
}

// UpdateWorkflowEdgeRequest 更新工作流边请求结构
type UpdateWorkflowEdgeRequest struct {
	SourceHandle string                 `json:"sourceHandle,omitempty"`
	TargetHandle string                 `json:"targetHandle,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Label        string                 `json:"label,omitempty"`
	BranchName   string                 `json:"branchName,omitempty"`
	Animated     *bool                  `json:"animated,omitempty"`
	Style        map[string]interface{} `json:"style,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

// PageWorkflowEdgeRequest 分页查询工作流边请求结构
type PageWorkflowEdgeRequest struct {
	PaginationRequest
	ApplicationID string `json:"applicationId,omitempty" form:"applicationId"`
	SourceNodeID  string `json:"sourceNodeId,omitempty" form:"sourceNodeId"`
	TargetNodeID  string `json:"targetNodeId,omitempty" form:"targetNodeId"`
	Type          string `json:"type,omitempty" form:"type"`
	BeginTime     string `json:"beginTime,omitempty" form:"beginTime"`
	EndTime       string `json:"endTime,omitempty" form:"endTime"`
}

// PageWorkflowEdgeResponse 分页查询工作流边响应结构
type PageWorkflowEdgeResponse struct {
	Data       []*WorkflowEdgeResponse `json:"data"`
	Pagination Pagination              `json:"pagination"`
}

// BatchCreateWorkflowEdgesRequest 批量创建工作流边请求结构
type BatchCreateWorkflowEdgesRequest struct {
	Edges []CreateWorkflowEdgeRequest `json:"edges" binding:"required"`
}

// BatchDeleteWorkflowEdgesRequest 批量删除工作流边请求结构
type BatchDeleteWorkflowEdgesRequest struct {
	EdgeIDs []string `json:"edgeIds" binding:"required"`
}

// ============ WorkflowNode Models ============

// WorkflowNodeResponse 工作流节点响应结构
type WorkflowNodeResponse struct {
	ID                string                 `json:"id"`
	CreateTime        string                 `json:"createTime"`
	UpdateTime        string                 `json:"updateTime"`
	Name              string                 `json:"name"`
	Type              string                 `json:"type"` // user_input, todo_task_generator, condition_checker, api_caller, data_processor, while_loop, end_node, parallel_executor, llm_caller
	Description       string                 `json:"description,omitempty"`
	Prompt            string                 `json:"prompt,omitempty"`
	Config            map[string]interface{} `json:"config"`
	ApplicationID     string                 `json:"applicationId"`
	ProcessorLanguage string                 `json:"processorLanguage,omitempty"`
	ProcessorCode     string                 `json:"processorCode,omitempty"`
	BranchNodes       map[string]interface{} `json:"branchNodes,omitempty"`
	ParallelConfig    map[string]interface{} `json:"parallelConfig,omitempty"`
	APIConfig         map[string]interface{} `json:"apiConfig,omitempty"`
	Async             bool                   `json:"async"`
	Timeout           int                    `json:"timeout"`
	RetryCount        int                    `json:"retryCount"`
	PositionX         float64                `json:"positionX"`
	PositionY         float64                `json:"positionY"`
	Color             string                 `json:"color,omitempty"`
}

// CreateWorkflowNodeRequest 创建工作流节点请求结构
type CreateWorkflowNodeRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Type              string                 `json:"type" binding:"required"` // user_input, todo_task_generator, condition_checker, api_caller, data_processor, while_loop, end_node, parallel_executor, llm_caller
	Description       string                 `json:"description,omitempty"`
	Prompt            string                 `json:"prompt,omitempty"`
	Config            map[string]interface{} `json:"config" binding:"required"`
	ApplicationID     string                 `json:"applicationId" binding:"required"`
	ProcessorLanguage string                 `json:"processorLanguage,omitempty"`
	ProcessorCode     string                 `json:"processorCode,omitempty"`
	BranchNodes       map[string]interface{} `json:"branchNodes,omitempty"`
	ParallelConfig    map[string]interface{} `json:"parallelConfig,omitempty"`
	APIConfig         map[string]interface{} `json:"apiConfig,omitempty"`
	Async             *bool                  `json:"async,omitempty"`
	Timeout           *int                   `json:"timeout,omitempty"`
	RetryCount        *int                   `json:"retryCount,omitempty"`
	PositionX         *float64               `json:"positionX,omitempty"`
	PositionY         *float64               `json:"positionY,omitempty"`
	Color             string                 `json:"color,omitempty"`
}

// UpdateWorkflowNodeRequest 更新工作流节点请求结构
// 注意：所有字段都是可选的，只更新提交的字段
type UpdateWorkflowNodeRequest struct {
	Name              string                 `json:"name,omitempty"`
	Type              string                 `json:"type,omitempty"` // user_input, todo_task_generator, condition_checker, api_caller, data_processor, while_loop, end_node, parallel_executor, llm_caller
	Description       string                 `json:"description,omitempty"`
	Prompt            string                 `json:"prompt,omitempty"`
	Config            map[string]interface{} `json:"config,omitempty"`
	ProcessorLanguage string                 `json:"processorLanguage,omitempty"`
	ProcessorCode     string                 `json:"processorCode,omitempty"`
	BranchNodes       map[string]interface{} `json:"branchNodes,omitempty"`
	ParallelConfig    map[string]interface{} `json:"parallelConfig,omitempty"`
	APIConfig         map[string]interface{} `json:"apiConfig,omitempty"`
	Async             *bool                  `json:"async,omitempty"`
	Timeout           *int                   `json:"timeout,omitempty"`
	RetryCount        *int                   `json:"retryCount,omitempty"`
	PositionX         *float64               `json:"positionX,omitempty"`
	PositionY         *float64               `json:"positionY,omitempty"`
	Color             string                 `json:"color,omitempty"`
}

// PageWorkflowNodeRequest 分页查询工作流节点请求结构
type PageWorkflowNodeRequest struct {
	PaginationRequest
	Name          string `form:"name" json:"name"`                   // 按名称模糊搜索
	Type          string `form:"type" json:"type"`                   // 按类型过滤
	ApplicationID string `form:"applicationId" json:"applicationId"` // 按应用ID过滤
	BeginTime     string `form:"beginTime" json:"beginTime"`         // 开始时间
	EndTime       string `form:"endTime" json:"endTime"`             // 结束时间
}

// PageWorkflowNodeResponse 分页查询工作流节点响应结构
type PageWorkflowNodeResponse struct {
	Data       []*WorkflowNodeResponse `json:"data"`
	Pagination Pagination              `json:"pagination"`
}

// ============ WorkflowExecution Models ============

// WorkflowExecutionResponse 工作流执行记录响应结构
type WorkflowExecutionResponse struct {
	ID             string                           `json:"id"`
	CreateTime     string                           `json:"createTime"`
	UpdateTime     string                           `json:"updateTime"`
	ExecutionID    string                           `json:"executionId"`
	ApplicationID  string                           `json:"applicationId"`
	Status         string                           `json:"status"` // pending, running, completed, failed, cancelled, timeout
	Input          map[string]interface{}           `json:"input,omitempty"`
	Output         map[string]interface{}           `json:"output,omitempty"`
	Context        map[string]interface{}           `json:"context,omitempty"`
	StartedAt      *time.Time                       `json:"startedAt,omitempty"`
	FinishedAt     *time.Time                       `json:"finishedAt,omitempty"`
	DurationMs     int                              `json:"durationMs"`
	TotalTokens    int                              `json:"totalTokens"`
	TotalCost      float64                          `json:"totalCost"`
	ErrorMessage   string                           `json:"errorMessage,omitempty"`
	ErrorStack     string                           `json:"errorStack,omitempty"`
	TriggeredBy    string                           `json:"triggeredBy,omitempty"`
	TriggerSource  string                           `json:"triggerSource,omitempty"`
	NodeExecutions []*WorkflowNodeExecutionResponse `json:"nodeExecutions,omitempty"`
}

// CreateWorkflowExecutionRequest 创建工作流执行请求结构
type CreateWorkflowExecutionRequest struct {
	ApplicationID string                 `json:"applicationId" binding:"required"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	TriggeredBy   string                 `json:"triggeredBy,omitempty"`
	TriggerSource string                 `json:"triggerSource,omitempty"`
}

// UpdateWorkflowExecutionRequest 更新工作流执行请求结构
type UpdateWorkflowExecutionRequest struct {
	Status       string                 `json:"status,omitempty"` // pending, running, completed, failed, cancelled, timeout
	Output       map[string]interface{} `json:"output,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	ErrorMessage string                 `json:"errorMessage,omitempty"`
	ErrorStack   string                 `json:"errorStack,omitempty"`
}

// PageWorkflowExecutionRequest 分页查询工作流执行请求结构
type PageWorkflowExecutionRequest struct {
	PaginationRequest
	ExecutionID   string `form:"executionId" json:"executionId"`     // 按执行ID搜索
	ApplicationID string `form:"applicationId" json:"applicationId"` // 按应用ID过滤
	Status        string `form:"status" json:"status"`               // 按状态过滤
	TriggeredBy   string `form:"triggeredBy" json:"triggeredBy"`     // 按触发者过滤
	BeginTime     string `form:"beginTime" json:"beginTime"`         // 开始时间
	EndTime       string `form:"endTime" json:"endTime"`             // 结束时间
}

// PageWorkflowExecutionResponse 分页查询工作流执行响应结构
type PageWorkflowExecutionResponse struct {
	Data       []*WorkflowExecutionResponse `json:"data"`
	Pagination Pagination                   `json:"pagination"`
}

// ============ WorkflowNodeExecution Models ============

// WorkflowNodeExecutionResponse 节点执行记录响应结构
type WorkflowNodeExecutionResponse struct {
	ID                string                 `json:"id"`
	CreateTime        string                 `json:"createTime"`
	UpdateTime        string                 `json:"updateTime"`
	ExecutionID       string                 `json:"executionId"`
	NodeID            string                 `json:"nodeId"`
	NodeName          string                 `json:"nodeName"`
	NodeType          string                 `json:"nodeType"`
	Status            string                 `json:"status"` // pending, running, completed, failed, skipped, timeout
	Input             map[string]interface{} `json:"input,omitempty"`
	Output            map[string]interface{} `json:"output,omitempty"`
	Extra             map[string]interface{} `json:"extra,omitempty"`
	StartedAt         *time.Time             `json:"startedAt,omitempty"`
	FinishedAt        *time.Time             `json:"finishedAt,omitempty"`
	DurationMs        int                    `json:"durationMs"`
	PromptTokens      int                    `json:"promptTokens"`
	CompletionTokens  int                    `json:"completionTokens"`
	TotalTokens       int                    `json:"totalTokens"`
	Cost              float64                `json:"cost"`
	Model             string                 `json:"model,omitempty"`
	ErrorMessage      string                 `json:"errorMessage,omitempty"`
	ErrorStack        string                 `json:"errorStack,omitempty"`
	RetryCount        int                    `json:"retryCount"`
	IsAsync           bool                   `json:"isAsync"`
	ParentExecutionID string                 `json:"parentExecutionId,omitempty"`
}

// CreateWorkflowNodeExecutionRequest 创建节点执行请求结构
type CreateWorkflowNodeExecutionRequest struct {
	ExecutionID       string                 `json:"executionId" binding:"required"`
	NodeID            string                 `json:"nodeId" binding:"required"`
	NodeName          string                 `json:"nodeName" binding:"required"`
	NodeType          string                 `json:"nodeType" binding:"required"`
	Input             map[string]interface{} `json:"input,omitempty"`
	IsAsync           *bool                  `json:"isAsync,omitempty"`
	ParentExecutionID string                 `json:"parentExecutionId,omitempty"`
}

// UpdateWorkflowNodeExecutionRequest 更新节点执行请求结构
type UpdateWorkflowNodeExecutionRequest struct {
	Status           string                 `json:"status,omitempty"` // pending, running, completed, failed, skipped, timeout
	Output           map[string]interface{} `json:"output,omitempty"`
	Extra            map[string]interface{} `json:"extra,omitempty"`
	PromptTokens     *int                   `json:"promptTokens,omitempty"`
	CompletionTokens *int                   `json:"completionTokens,omitempty"`
	TotalTokens      *int                   `json:"totalTokens,omitempty"`
	Cost             *float64               `json:"cost,omitempty"`
	Model            string                 `json:"model,omitempty"`
	ErrorMessage     string                 `json:"errorMessage,omitempty"`
	ErrorStack       string                 `json:"errorStack,omitempty"`
	RetryCount       *int                   `json:"retryCount,omitempty"`
}

// PageWorkflowNodeExecutionRequest 分页查询节点执行请求结构
type PageWorkflowNodeExecutionRequest struct {
	PaginationRequest
	ExecutionID       string `form:"executionId" json:"executionId"`             // 按工作流执行ID过滤
	NodeID            string `form:"nodeId" json:"nodeId"`                       // 按节点ID过滤
	NodeType          string `form:"nodeType" json:"nodeType"`                   // 按节点类型过滤
	Status            string `form:"status" json:"status"`                       // 按状态过滤
	ParentExecutionID string `form:"parentExecutionId" json:"parentExecutionId"` // 按父执行ID过滤
	BeginTime         string `form:"beginTime" json:"beginTime"`                 // 开始时间
	EndTime           string `form:"endTime" json:"endTime"`                     // 结束时间
}

// PageWorkflowNodeExecutionResponse 分页查询节点执行响应结构
type PageWorkflowNodeExecutionResponse struct {
	Data       []*WorkflowNodeExecutionResponse `json:"data"`
	Pagination Pagination                       `json:"pagination"`
}

// ============ WorkflowVersion Models ============

// WorkflowVersionSnapshot 版本快照数据结构
type WorkflowVersionSnapshot struct {
	Nodes []*WorkflowNodeResponse `json:"nodes"`
	Edges []*WorkflowEdgeResponse `json:"edges"`
}

// WorkflowVersionResponse 工作流版本响应结构
type WorkflowVersionResponse struct {
	ID            string                  `json:"id"`
	CreateTime    string                  `json:"createTime"`
	UpdateTime    string                  `json:"updateTime"`
	ApplicationID string                  `json:"applicationId"`
	Version       uint                    `json:"version"`
	Snapshot      WorkflowVersionSnapshot `json:"snapshot"`
	ChangeLog     string                  `json:"changeLog,omitempty"`
}

// CreateWorkflowVersionRequest 创建工作流版本请求结构
type CreateWorkflowVersionRequest struct {
	ApplicationID string `json:"applicationId" binding:"required"`
	ChangeLog     string `json:"changeLog,omitempty"` // 可选的变更日志
}

// PageWorkflowVersionRequest 分页查询工作流版本请求结构
type PageWorkflowVersionRequest struct {
	PaginationRequest
	ApplicationID string `form:"applicationId" json:"applicationId"` // 按应用ID过滤
	Version       uint   `form:"version" json:"version"`             // 按版本号过滤
	BeginTime     string `form:"beginTime" json:"beginTime"`         // 开始时间
	EndTime       string `form:"endTime" json:"endTime"`             // 结束时间
}

// PageWorkflowVersionResponse 分页查询工作流版本响应结构
type PageWorkflowVersionResponse struct {
	Data       []*WorkflowVersionResponse `json:"data"`
	Pagination Pagination                 `json:"pagination"`
}

// WorkflowVersionResult 工作流版本结果
type WorkflowVersionResult struct {
	Success bool                     `json:"success"`
	Data    *WorkflowVersionResponse `json:"data,omitempty"`
	Message string                   `json:"message,omitempty"`
}

// WorkflowVersionListResult 工作流版本列表结果
type WorkflowVersionListResult struct {
	Success bool                       `json:"success"`
	Data    []*WorkflowVersionResponse `json:"data,omitempty"`
	Message string                     `json:"message,omitempty"`
}

// ============ WorkflowExecutionLog Models ============

// WorkflowExecutionLogResponse 执行日志响应结构
type WorkflowExecutionLogResponse struct {
	ID              string                 `json:"id"`
	CreateTime      string                 `json:"createTime"`
	UpdateTime      string                 `json:"updateTime"`
	ExecutionID     string                 `json:"executionId"`
	NodeExecutionID string                 `json:"nodeExecutionId,omitempty"`
	Level           string                 `json:"level"` // debug, info, warn, error
	Message         string                 `json:"message"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	LoggedAt        time.Time              `json:"loggedAt"`
}

// CreateWorkflowExecutionLogRequest 创建执行日志请求结构
type CreateWorkflowExecutionLogRequest struct {
	ExecutionID     string                 `json:"executionId" binding:"required"`
	NodeExecutionID string                 `json:"nodeExecutionId,omitempty"`
	Level           string                 `json:"level" binding:"required"` // debug, info, warn, error
	Message         string                 `json:"message" binding:"required"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// PageWorkflowExecutionLogRequest 分页查询执行日志请求结构
type PageWorkflowExecutionLogRequest struct {
	PaginationRequest
	ExecutionID     string `form:"executionId" json:"executionId"`         // 按执行ID过滤
	NodeExecutionID string `form:"nodeExecutionId" json:"nodeExecutionId"` // 按节点执行ID过滤
	Level           string `form:"level" json:"level"`                     // 按日志级别过滤
	BeginTime       string `form:"beginTime" json:"beginTime"`             // 开始时间
	EndTime         string `form:"endTime" json:"endTime"`                 // 结束时间
}

// PageWorkflowExecutionLogResponse 分页查询执行日志响应结构
type PageWorkflowExecutionLogResponse struct {
	Data       []*WorkflowExecutionLogResponse `json:"data"`
	Pagination Pagination                      `json:"pagination"`
}

// ============ Batch Save Models ============

// BatchSaveWorkflowRequest 批量保存工作流请求结构
type BatchSaveWorkflowRequest struct {
	ApplicationID   string                      `json:"applicationId" binding:"required"`
	NodeTempIDs     []string                    `json:"nodeTempIds"` // 要创建的节点的临时ID列表（与 NodesToCreate 一一对应）
	EdgeTempIDs     []string                    `json:"edgeTempIds"` // 要创建的边的临时ID列表（与 EdgesToCreate 一一对应）
	NodesToCreate   []CreateWorkflowNodeRequest `json:"nodesToCreate"`
	NodesToUpdate   []UpdateWorkflowNodeWithID  `json:"nodesToUpdate"`
	NodeIDsToDelete []string                    `json:"nodeIdsToDelete"`
	EdgesToCreate   []CreateWorkflowEdgeRequest `json:"edgesToCreate"`
	EdgesToUpdate   []UpdateWorkflowEdgeWithID  `json:"edgesToUpdate"`
	EdgeIDsToDelete []string                    `json:"edgeIdsToDelete"`
}

// UpdateWorkflowNodeWithID 带ID的节点更新请求
type UpdateWorkflowNodeWithID struct {
	ID   string                    `json:"id" binding:"required"`
	Data UpdateWorkflowNodeRequest `json:"data" binding:"required"`
}

// UpdateWorkflowEdgeWithID 带ID的边更新请求
type UpdateWorkflowEdgeWithID struct {
	ID   string                    `json:"id" binding:"required"`
	Data UpdateWorkflowEdgeRequest `json:"data" binding:"required"`
}

// BatchSaveWorkflowStats 批量保存统计信息
type BatchSaveWorkflowStats struct {
	NodesCreated int `json:"nodesCreated"`
	NodesUpdated int `json:"nodesUpdated"`
	NodesDeleted int `json:"nodesDeleted"`
	EdgesCreated int `json:"edgesCreated"`
	EdgesUpdated int `json:"edgesUpdated"`
	EdgesDeleted int `json:"edgesDeleted"`
}

// BatchSaveWorkflowData 批量保存返回数据
type BatchSaveWorkflowData struct {
	NodeIDMapping  map[string]string       `json:"nodeIdMapping"` // 临时ID -> 数据库ID 映射
	EdgeIDMapping  map[string]string       `json:"edgeIdMapping"` // 临时ID -> 数据库ID 映射
	CreatedNodes   []*WorkflowNodeResponse `json:"createdNodes"`
	UpdatedNodes   []*WorkflowNodeResponse `json:"updatedNodes"`
	DeletedNodeIDs []string                `json:"deletedNodeIds"`
	CreatedEdges   []*WorkflowEdgeResponse `json:"createdEdges"`
	UpdatedEdges   []*WorkflowEdgeResponse `json:"updatedEdges"`
	DeletedEdgeIDs []string                `json:"deletedEdgeIds"`
	Stats          BatchSaveWorkflowStats  `json:"stats"`
}

// BatchSaveWorkflowResponse 批量保存工作流响应结构
type BatchSaveWorkflowResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    *BatchSaveWorkflowData `json:"data"`
}
