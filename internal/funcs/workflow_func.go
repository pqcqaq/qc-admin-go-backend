package funcs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/workflowapplication"
	"go-backend/database/ent/workflowedge"
	"go-backend/database/ent/workflownode"
	"go-backend/database/ent/workflowversion"
	"go-backend/pkg/database"
	"go-backend/pkg/utils"
	"go-backend/shared/models"
)

// WorkflowFuncs 工作流服务函数
type WorkflowFuncs struct{}

// ============ WorkflowApplication CRUD ============

// GetAllWorkflowApplications 获取所有工作流应用
func (WorkflowFuncs) GetAllWorkflowApplications(ctx context.Context) ([]*models.WorkflowApplicationResponse, error) {
	apps, err := database.Client.WorkflowApplication.Query().
		WithNodes().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	appResponses := make([]*models.WorkflowApplicationResponse, 0, len(apps))
	for _, app := range apps {
		appResponses = append(appResponses, WorkflowFuncs{}.ConvertWorkflowApplicationToResponse(app))
	}

	return appResponses, nil
}

// GetWorkflowApplicationByID 根据ID获取工作流应用
func (WorkflowFuncs) GetWorkflowApplicationByID(ctx context.Context, id uint64) (*models.WorkflowApplicationResponse, error) {
	app, err := database.Client.WorkflowApplication.Query().
		Where(workflowapplication.ID(id)).
		// WithNodes().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow application not found")
		}
		return nil, err
	}
	return WorkflowFuncs{}.ConvertWorkflowApplicationToResponse(app), nil
}

// CreateWorkflowApplication 创建工作流应用
func (WorkflowFuncs) CreateWorkflowApplication(ctx context.Context, req *models.CreateWorkflowApplicationRequest) (*models.WorkflowApplicationResponse, error) {
	// 生成客户端密钥
	clientSecret, err := generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	// 使用事务创建应用和默认开始节点
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	// 如果没有提供 startNodeId，创建默认的开始节点
	var startNodeID uint64
	if req.StartNodeID != "" {
		startNodeID = utils.StringToUint64(req.StartNodeID)
	}

	// 先创建应用（使用临时的 startNodeID = 0）
	appBuilder := tx.WorkflowApplication.Create().
		SetName(req.Name).
		SetStartNodeID(0). // 临时值
		SetClientSecret(clientSecret)

	if req.Description != "" {
		appBuilder = appBuilder.SetDescription(req.Description)
	}

	if req.Variables != nil {
		appBuilder = appBuilder.SetVariables(req.Variables)
	}

	if req.Status != "" {
		appBuilder = appBuilder.SetStatus(workflowapplication.Status(req.Status))
	}

	app, err := appBuilder.Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 如果没有提供 startNodeId，创建默认的开始节点
	if req.StartNodeID == "" {
		startNode, err := tx.WorkflowNode.Create().
			SetName("开始").
			SetType(workflownode.TypeUserInput).
			SetDescription("工作流开始节点").
			SetConfig(map[string]interface{}{}).
			SetApplicationID(app.ID).
			SetPositionX(250).
			SetPositionY(50).
			SetColor("#67C23A").
			Save(ctx)

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create start node: %w", err)
		}

		startNodeID = startNode.ID

		// 更新应用的 startNodeID
		app, err = app.Update().
			SetStartNodeID(startNodeID).
			Save(ctx)

		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update application start node: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return WorkflowFuncs{}.GetWorkflowApplicationByID(ctx, app.ID)
}

// UpdateWorkflowApplication 更新工作流应用
func (WorkflowFuncs) UpdateWorkflowApplication(ctx context.Context, id uint64, req *models.UpdateWorkflowApplicationRequest) (*models.WorkflowApplicationResponse, error) {
	builder := database.Client.WorkflowApplication.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.StartNodeID != "" {
		startNodeID := utils.StringToUint64(req.StartNodeID)
		builder = builder.SetStartNodeID(startNodeID)
	}

	if req.Variables != nil {
		builder = builder.SetVariables(req.Variables)
	}

	if req.Version > 0 {
		builder = builder.SetVersion(req.Version)
	}

	if req.Status != "" {
		builder = builder.SetStatus(workflowapplication.Status(req.Status))
	}

	if req.ViewportConfig != nil {
		builder = builder.SetViewportConfig(req.ViewportConfig)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow application not found")
		}
		return nil, err
	}

	return WorkflowFuncs{}.GetWorkflowApplicationByID(ctx, id)
}

// DeleteWorkflowApplication 删除工作流应用(软删除)
func (WorkflowFuncs) DeleteWorkflowApplication(ctx context.Context, id uint64) error {
	err := database.Client.WorkflowApplication.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("workflow application not found")
		}
		return err
	}
	return nil
}

// GetWorkflowApplicationsWithPagination 分页获取工作流应用列表
func (WorkflowFuncs) GetWorkflowApplicationsWithPagination(ctx context.Context, req *models.PageWorkflowApplicationRequest) (*models.PageWorkflowApplicationResponse, error) {
	query := database.Client.WorkflowApplication.Query().
		WithNodes()

	// 添加搜索条件
	if req.Name != "" {
		query = query.Where(workflowapplication.NameContains(req.Name))
	}

	if req.Status != "" {
		query = query.Where(workflowapplication.StatusEQ(workflowapplication.Status(req.Status)))
	}

	if req.BeginTime != "" {
		beginTime, err := time.Parse(time.RFC3339, req.BeginTime)
		if err == nil {
			query = query.Where(workflowapplication.CreateTimeGTE(beginTime))
		}
	}

	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			query = query.Where(workflowapplication.CreateTimeLTE(endTime))
		}
	}

	// 获取总数
	total, err := query.Count(ctx)
	if err != nil {
		return nil, err
	}

	// 计算分页
	offset := (req.Page - 1) * req.PageSize
	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	// 设置排序
	if req.OrderBy != "" {
		switch req.OrderBy {
		case "name":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(workflowapplication.FieldName))
			} else {
				query = query.Order(ent.Asc(workflowapplication.FieldName))
			}
		case "createTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(workflowapplication.FieldCreateTime))
			} else {
				query = query.Order(ent.Asc(workflowapplication.FieldCreateTime))
			}
		case "updateTime":
			if req.Order == "desc" {
				query = query.Order(ent.Desc(workflowapplication.FieldUpdateTime))
			} else {
				query = query.Order(ent.Asc(workflowapplication.FieldUpdateTime))
			}
		}
	} else {
		// 默认按创建时间降序
		query = query.Order(ent.Desc(workflowapplication.FieldCreateTime))
	}

	// 执行查询
	apps, err := query.Offset(offset).Limit(req.PageSize).All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	appResponses := make([]*models.WorkflowApplicationResponse, 0, len(apps))
	for _, app := range apps {
		appResponses = append(appResponses, WorkflowFuncs{}.ConvertWorkflowApplicationToResponse(app))
	}

	return &models.PageWorkflowApplicationResponse{
		Data: appResponses,
		Pagination: models.Pagination{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      int64(total),
			TotalPages: totalPages,
			HasNext:    req.Page < totalPages,
			HasPrev:    req.Page > 1,
		},
	}, nil
}

// ConvertWorkflowApplicationToResponse 将工作流应用实体转换为响应格式
func (WorkflowFuncs) ConvertWorkflowApplicationToResponse(app *ent.WorkflowApplication) *models.WorkflowApplicationResponse {
	resp := &models.WorkflowApplicationResponse{
		ID:             utils.Uint64ToString(app.ID),
		CreateTime:     utils.FormatDateTime(app.CreateTime),
		UpdateTime:     utils.FormatDateTime(app.UpdateTime),
		Name:           app.Name,
		Description:    app.Description,
		StartNodeID:    utils.Uint64ToString(app.StartNodeID),
		ClientSecret:   app.ClientSecret,
		Variables:      app.Variables,
		Version:        app.Version,
		Status:         string(app.Status),
		ViewportConfig: app.ViewportConfig,
	}

	// // 转换节点列表（旧架构，保留兼容）
	// if len(app.Edges.Nodes) > 0 {
	// 	resp.Nodes = make([]*models.WorkflowNodeResponse, 0, len(app.Edges.Nodes))
	// 	for _, node := range app.Edges.Nodes {
	// 		resp.Nodes = append(resp.Nodes, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
	// 	}
	// }

	return resp
}

// generateClientSecret 生成客户端密钥
func generateClientSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ============ WorkflowNode CRUD ============

// GetAllWorkflowNodes 获取所有工作流节点
func (WorkflowFuncs) GetAllWorkflowNodes(ctx context.Context) ([]*models.WorkflowNodeResponse, error) {
	nodes, err := database.Client.WorkflowNode.Query().
		WithApplication().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	nodeResponses := make([]*models.WorkflowNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		nodeResponses = append(nodeResponses, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
	}

	return nodeResponses, nil
}

// GetWorkflowNodeByID 根据ID获取工作流节点
func (WorkflowFuncs) GetWorkflowNodeByID(ctx context.Context, id uint64) (*models.WorkflowNodeResponse, error) {
	node, err := database.Client.WorkflowNode.Query().
		Where(workflownode.ID(id)).
		WithApplication().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow node not found")
		}
		return nil, err
	}
	return WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node), nil
}

// GetWorkflowNodesByApplicationID 根据应用ID获取所有节点
func (WorkflowFuncs) GetWorkflowNodesByApplicationID(ctx context.Context, applicationID uint64) ([]*models.WorkflowNodeResponse, error) {
	nodes, err := database.Client.WorkflowNode.Query().
		Where(workflownode.ApplicationIDEQ(applicationID)).
		WithApplication().
		Order(ent.Asc(workflownode.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	nodeResponses := make([]*models.WorkflowNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		nodeResponses = append(nodeResponses, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
	}

	return nodeResponses, nil
}

// CreateWorkflowNode 创建工作流节点
func (WorkflowFuncs) CreateWorkflowNode(ctx context.Context, req *models.CreateWorkflowNodeRequest) (*models.WorkflowNodeResponse, error) {
	applicationID := utils.StringToUint64(req.ApplicationID)

	builder := database.Client.WorkflowNode.Create().
		SetName(req.Name).
		SetType(workflownode.Type(req.Type)).
		SetConfig(req.Config).
		SetApplicationID(applicationID)

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.Prompt != "" {
		builder = builder.SetPrompt(req.Prompt)
	}

	if req.ProcessorLanguage != "" {
		builder = builder.SetProcessorLanguage(req.ProcessorLanguage)
	}

	if req.ProcessorCode != "" {
		builder = builder.SetProcessorCode(req.ProcessorCode)
	}

	if req.BranchNodes != nil {
		builder = builder.SetBranchNodes(req.BranchNodes)
	}

	if req.ParallelConfig != nil {
		builder = builder.SetParallelConfig(req.ParallelConfig)
	}

	if req.APIConfig != nil {
		builder = builder.SetAPIConfig(req.APIConfig)
	}

	if req.Async != nil {
		builder = builder.SetAsync(*req.Async)
	}

	if req.Timeout != nil {
		builder = builder.SetTimeout(*req.Timeout)
	}

	if req.RetryCount != nil {
		builder = builder.SetRetryCount(*req.RetryCount)
	}

	if req.PositionX != nil {
		builder = builder.SetPositionX(*req.PositionX)
	}

	if req.PositionY != nil {
		builder = builder.SetPositionY(*req.PositionY)
	}

	if req.Color != "" {
		builder = builder.SetColor(req.Color)
	}

	node, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return WorkflowFuncs{}.GetWorkflowNodeByID(ctx, node.ID)
}

// UpdateWorkflowNode 更新工作流节点
func (WorkflowFuncs) UpdateWorkflowNode(ctx context.Context, id uint64, req *models.UpdateWorkflowNodeRequest) (*models.WorkflowNodeResponse, error) {
	builder := database.Client.WorkflowNode.UpdateOneID(id)

	if req.Name != "" {
		builder = builder.SetName(req.Name)
	}

	if req.Type != "" {
		builder = builder.SetType(workflownode.Type(req.Type))
	}

	if req.Description != "" {
		builder = builder.SetDescription(req.Description)
	}

	if req.Prompt != "" {
		builder = builder.SetPrompt(req.Prompt)
	}

	if req.Config != nil {
		builder = builder.SetConfig(req.Config)
	}

	if req.ProcessorLanguage != "" {
		builder = builder.SetProcessorLanguage(req.ProcessorLanguage)
	}

	if req.ProcessorCode != "" {
		builder = builder.SetProcessorCode(req.ProcessorCode)
	}

	if req.BranchNodes != nil {
		builder = builder.SetBranchNodes(req.BranchNodes)
	}

	if req.ParallelConfig != nil {
		builder = builder.SetParallelConfig(req.ParallelConfig)
	}

	if req.APIConfig != nil {
		builder = builder.SetAPIConfig(req.APIConfig)
	}

	if req.Async != nil {
		builder = builder.SetAsync(*req.Async)
	}

	if req.Timeout != nil {
		builder = builder.SetTimeout(*req.Timeout)
	}

	if req.RetryCount != nil {
		builder = builder.SetRetryCount(*req.RetryCount)
	}

	if req.PositionX != nil {
		builder = builder.SetPositionX(*req.PositionX)
	}

	if req.PositionY != nil {
		builder = builder.SetPositionY(*req.PositionY)
	}

	if req.Color != "" {
		builder = builder.SetColor(req.Color)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow node not found")
		}
		return nil, err
	}

	return WorkflowFuncs{}.GetWorkflowNodeByID(ctx, id)
}

// DeleteWorkflowNode 删除工作流节点(软删除)
func (WorkflowFuncs) DeleteWorkflowNode(ctx context.Context, id uint64) error {
	err := database.Client.WorkflowNode.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("workflow node not found")
		}
		return err
	}
	return nil
}

// ConvertWorkflowNodeToResponse 将工作流节点实体转换为响应格式
func (WorkflowFuncs) ConvertWorkflowNodeToResponse(node *ent.WorkflowNode) *models.WorkflowNodeResponse {
	resp := &models.WorkflowNodeResponse{
		ID:                utils.Uint64ToString(node.ID),
		CreateTime:        utils.FormatDateTime(node.CreateTime),
		UpdateTime:        utils.FormatDateTime(node.UpdateTime),
		Name:              node.Name,
		Type:              string(node.Type),
		Description:       node.Description,
		Prompt:            node.Prompt,
		Config:            node.Config,
		ApplicationID:     utils.Uint64ToString(node.ApplicationID),
		ProcessorLanguage: node.ProcessorLanguage,
		ProcessorCode:     node.ProcessorCode,
		BranchNodes:       node.BranchNodes,
		ParallelConfig:    node.ParallelConfig,
		APIConfig:         node.APIConfig,
		Async:             node.Async,
		Timeout:           node.Timeout,
		RetryCount:        node.RetryCount,
		PositionX:         node.PositionX,
		PositionY:         node.PositionY,
		Color:             node.Color,
	}

	return resp
}

// // ============ Workflow Graph Operations ============

// // NodeConnectionRule 节点连接规则
// type NodeConnectionRule struct {
// 	CanHaveNextNode      bool // 是否可以有next_node_id
// 	CanHaveBranches      bool // 是否可以有分支
// 	CanBeParallel        bool // 是否可以作为并行节点的子节点
// 	RequiresBranchName   bool // 连接时是否需要分支名称
// 	MaxOutputConnections int  // 最大输出连接数 (-1表示无限制)
// }

// // getNodeConnectionRule 获取节点类型的连接规则
// func getNodeConnectionRule(nodeType string) NodeConnectionRule {
// 	rules := map[string]NodeConnectionRule{
// 		"user_input": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        false,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 		"todo_task_generator": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 		"condition_checker": {
// 			CanHaveNextNode:      false,
// 			CanHaveBranches:      true,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   true,
// 			MaxOutputConnections: -1, // 可以有多个分支
// 		},
// 		"api_caller": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 		"data_processor": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 		"while_loop": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        false,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 		"end_node": {
// 			CanHaveNextNode:      false,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 0,
// 		},
// 		"parallel_executor": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        false,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 		"llm_caller": {
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		},
// 	}

// 	rule, exists := rules[nodeType]
// 	if !exists {
// 		// 默认规则
// 		return NodeConnectionRule{
// 			CanHaveNextNode:      true,
// 			CanHaveBranches:      false,
// 			CanBeParallel:        true,
// 			RequiresBranchName:   false,
// 			MaxOutputConnections: 1,
// 		}
// 	}
// 	return rule
// }

// // ConnectNodes 连接两个节点（普通连接，用于next_node_id）
// func (WorkflowFuncs) ConnectNodes(ctx context.Context, fromNodeID, toNodeID uint64) error {
// 	// 获取源节点
// 	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("source node not found")
// 		}
// 		return err
// 	}

// 	// 检查目标节点是否存在
// 	_, err = database.Client.WorkflowNode.Get(ctx, toNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("target node not found")
// 		}
// 		return err
// 	}

// 	// 获取节点连接规则
// 	rule := getNodeConnectionRule(string(fromNode.Type))

// 	// 检查源节点是否可以有next_node_id
// 	if !rule.CanHaveNextNode {
// 		return fmt.Errorf("node type '%s' cannot have next_node connection, use branch connection instead", fromNode.Type)
// 	}

// 	// 检查是否已经有连接
// 	if fromNode.NextNodeID != 0 {
// 		return fmt.Errorf("node already has a next_node connection, disconnect first")
// 	}

// 	// 更新源节点的 next_node_id
// 	err = database.Client.WorkflowNode.UpdateOneID(fromNodeID).
// 		SetNextNodeID(toNodeID).
// 		Exec(ctx)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // DisconnectNodes 断开节点的next_node_id连接
// func (WorkflowFuncs) DisconnectNodes(ctx context.Context, fromNodeID uint64) error {
// 	// 清除源节点的 next_node_id
// 	err := database.Client.WorkflowNode.UpdateOneID(fromNodeID).
// 		ClearNextNodeID().
// 		Exec(ctx)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("source node not found")
// 		}
// 		return err
// 	}
// 	return nil
// }

// // ConnectBranch 为分支节点（如condition_checker）添加分支连接
// func (WorkflowFuncs) ConnectBranch(ctx context.Context, fromNodeID, toNodeID uint64, branchName string) error {
// 	if branchName == "" {
// 		return fmt.Errorf("branch name is required")
// 	}

// 	// 获取源节点
// 	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("source node not found")
// 		}
// 		return err
// 	}

// 	// 检查目标节点是否存在
// 	_, err = database.Client.WorkflowNode.Get(ctx, toNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("target node not found")
// 		}
// 		return err
// 	}

// 	// 获取节点连接规则
// 	rule := getNodeConnectionRule(string(fromNode.Type))

// 	// 检查源节点是否可以有分支
// 	if !rule.CanHaveBranches {
// 		return fmt.Errorf("node type '%s' cannot have branch connections", fromNode.Type)
// 	}

// 	// 获取现有分支
// 	branchNodes := fromNode.BranchNodes
// 	if branchNodes == nil {
// 		branchNodes = make(map[string]interface{})
// 	}

// 	// 添加或更新分支
// 	branchNodes[branchName] = map[string]interface{}{
// 		"targetNodeId": toNodeID,
// 	}

// 	// 更新节点
// 	err = database.Client.WorkflowNode.UpdateOneID(fromNodeID).
// 		SetBranchNodes(branchNodes).
// 		Exec(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // DisconnectBranch 删除分支节点的某个分支连接
// func (WorkflowFuncs) DisconnectBranch(ctx context.Context, fromNodeID uint64, branchName string) error {
// 	if branchName == "" {
// 		return fmt.Errorf("branch name is required")
// 	}

// 	// 获取源节点
// 	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("source node not found")
// 		}
// 		return err
// 	}

// 	// 获取现有分支
// 	branchNodes := fromNode.BranchNodes
// 	if branchNodes == nil {
// 		return fmt.Errorf("node has no branches")
// 	}

// 	// 检查分支是否存在
// 	if _, exists := branchNodes[branchName]; !exists {
// 		return fmt.Errorf("branch '%s' not found", branchName)
// 	}

// 	// 删除分支
// 	delete(branchNodes, branchName)

// 	// 更新节点
// 	err = database.Client.WorkflowNode.UpdateOneID(fromNodeID).
// 		SetBranchNodes(branchNodes).
// 		Exec(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // AddNodeToParallel 将节点添加到并行执行节点
// func (WorkflowFuncs) AddNodeToParallel(ctx context.Context, parallelNodeID, childNodeID uint64) error {
// 	// 获取并行节点
// 	parallelNode, err := database.Client.WorkflowNode.Get(ctx, parallelNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("parallel node not found")
// 		}
// 		return err
// 	}

// 	// 检查是否是并行执行节点
// 	if parallelNode.Type != "parallel_executor" {
// 		return fmt.Errorf("node is not a parallel_executor, got type '%s'", parallelNode.Type)
// 	}

// 	// 获取子节点
// 	childNode, err := database.Client.WorkflowNode.Get(ctx, childNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("child node not found")
// 		}
// 		return err
// 	}

// 	// 检查子节点类型是否可以作为并行节点
// 	rule := getNodeConnectionRule(string(childNode.Type))
// 	if !rule.CanBeParallel {
// 		return fmt.Errorf("node type '%s' cannot be added to parallel executor", childNode.Type)
// 	}

// 	// 检查子节点是否已经有父节点
// 	if childNode.ParentNodeID != 0 {
// 		return fmt.Errorf("child node already has a parent node")
// 	}

// 	// 设置子节点的parent_node_id
// 	err = database.Client.WorkflowNode.UpdateOneID(childNodeID).
// 		SetParentNodeID(parallelNodeID).
// 		Exec(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // RemoveNodeFromParallel 从并行执行节点中移除子节点
// func (WorkflowFuncs) RemoveNodeFromParallel(ctx context.Context, childNodeID uint64) error {
// 	// 获取子节点
// 	childNode, err := database.Client.WorkflowNode.Get(ctx, childNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("child node not found")
// 		}
// 		return err
// 	}

// 	// 检查是否有父节点
// 	if childNode.ParentNodeID == 0 {
// 		return fmt.Errorf("node has no parent node")
// 	}

// 	// 清除parent_node_id
// 	err = database.Client.WorkflowNode.UpdateOneID(childNodeID).
// 		ClearParentNodeID().
// 		Exec(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// // GetParallelChildren 获取并行节点的所有子节点
// func (WorkflowFuncs) GetParallelChildren(ctx context.Context, parallelNodeID uint64) ([]*models.WorkflowNodeResponse, error) {
// 	// 获取并行节点
// 	parallelNode, err := database.Client.WorkflowNode.Get(ctx, parallelNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return nil, fmt.Errorf("parallel node not found")
// 		}
// 		return nil, err
// 	}

// 	// 检查是否是并行执行节点
// 	if parallelNode.Type != "parallel_executor" {
// 		return nil, fmt.Errorf("node is not a parallel_executor")
// 	}

// 	// 查询所有parent_node_id为该节点的子节点
// 	children, err := database.Client.WorkflowNode.Query().
// 		Where(workflownode.ParentNodeIDEQ(parallelNodeID)).
// 		All(ctx)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// 转换为响应格式
// 	childResponses := make([]*models.WorkflowNodeResponse, 0, len(children))
// 	for _, child := range children {
// 		childResponses = append(childResponses, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(child))
// 	}

// 	return childResponses, nil
// }

// // ValidateNodeConnection 验证两个节点是否可以连接
// func (WorkflowFuncs) ValidateNodeConnection(ctx context.Context, fromNodeID, toNodeID uint64, connectionType string) error {
// 	// 获取源节点和目标节点
// 	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("source node not found")
// 		}
// 		return err
// 	}

// 	toNode, err := database.Client.WorkflowNode.Get(ctx, toNodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("target node not found")
// 		}
// 		return err
// 	}

// 	// 检查是否在同一个应用中
// 	if fromNode.ApplicationID != toNode.ApplicationID {
// 		return fmt.Errorf("nodes must be in the same workflow application")
// 	}

// 	// 检查是否形成循环（简单检查：不能连接到自己）
// 	if fromNodeID == toNodeID {
// 		return fmt.Errorf("cannot connect node to itself")
// 	}

// 	// 根据连接类型验证
// 	fromRule := getNodeConnectionRule(string(fromNode.Type))
// 	toRule := getNodeConnectionRule(string(toNode.Type))

// 	switch connectionType {
// 	case "next":
// 		if !fromRule.CanHaveNextNode {
// 			return fmt.Errorf("source node type '%s' cannot have next_node connection", fromNode.Type)
// 		}
// 	case "branch":
// 		if !fromRule.CanHaveBranches {
// 			return fmt.Errorf("source node type '%s' cannot have branch connections", fromNode.Type)
// 		}
// 	case "parallel":
// 		if fromNode.Type != "parallel_executor" {
// 			return fmt.Errorf("source node must be parallel_executor for parallel connection")
// 		}
// 		if !toRule.CanBeParallel {
// 			return fmt.Errorf("target node type '%s' cannot be added to parallel executor", toNode.Type)
// 		}
// 	default:
// 		return fmt.Errorf("unknown connection type: %s", connectionType)
// 	}

// 	return nil
// }

// // GetNodeConnections 获取节点的所有连接信息
// func (WorkflowFuncs) GetNodeConnections(ctx context.Context, nodeID uint64) (map[string]interface{}, error) {
// 	node, err := database.Client.WorkflowNode.Get(ctx, nodeID)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return nil, fmt.Errorf("node not found")
// 		}
// 		return nil, err
// 	}

// 	connections := make(map[string]interface{})

// 	// Next node connection
// 	if node.NextNodeID != 0 {
// 		connections["next_node_id"] = utils.Uint64ToString(node.NextNodeID)
// 	}

// 	// Parent node (for parallel children)
// 	if node.ParentNodeID != 0 {
// 		connections["parent_node_id"] = utils.Uint64ToString(node.ParentNodeID)
// 	}

// 	// Branch connections
// 	if len(node.BranchNodes) > 0 {
// 		branches := make(map[string]string)
// 		for branchName, targetID := range node.BranchNodes {
// 			branches[branchName] = utils.Uint64ToString(targetID)
// 		}
// 		connections["branches"] = branches
// 	}

// 	// Parallel children (if this is a parallel executor)
// 	if node.Type == "parallel_executor" {
// 		children, err := WorkflowFuncs{}.GetParallelChildren(ctx, nodeID)
// 		if err == nil && len(children) > 0 {
// 			childIDs := make([]string, 0, len(children))
// 			for _, child := range children {
// 				childIDs = append(childIDs, child.ID)
// 			}
// 			connections["parallel_children"] = childIDs
// 		}
// 	}

// 	return connections, nil
// }

// // UpdateNodePosition 更新节点位置
// func (WorkflowFuncs) UpdateNodePosition(ctx context.Context, nodeID uint64, x, y float64) error {
// 	err := database.Client.WorkflowNode.UpdateOneID(nodeID).
// 		SetPositionX(x).
// 		SetPositionY(y).
// 		Exec(ctx)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return fmt.Errorf("node not found")
// 		}
// 		return err
// 	}
// 	return nil
// }

// // BatchUpdateNodePositions 批量更新节点位置
// func (WorkflowFuncs) BatchUpdateNodePositions(ctx context.Context, positions map[string]map[string]float64) error {
// 	// positions 格式: {"nodeId": {"x": 100, "y": 200}}
// 	tx, err := database.Client.Tx(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	for nodeIDStr, pos := range positions {
// 		nodeID := utils.StringToUint64(nodeIDStr)
// 		x, xOk := pos["x"]
// 		y, yOk := pos["y"]

// 		if !xOk || !yOk {
// 			tx.Rollback()
// 			return fmt.Errorf("invalid position data for node %s", nodeIDStr)
// 		}

// 		err = tx.WorkflowNode.UpdateOneID(nodeID).
// 			SetPositionX(x).
// 			SetPositionY(y).
// 			Exec(ctx)
// 		if err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	}

// 	return tx.Commit()
// }

// // BatchDeleteNodes 批量删除节点
// func (WorkflowFuncs) BatchDeleteNodes(ctx context.Context, nodeIDs []uint64) error {
// 	tx, err := database.Client.Tx(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	for _, nodeID := range nodeIDs {
// 		err = tx.WorkflowNode.DeleteOneID(nodeID).Exec(ctx)
// 		if err != nil {
// 			tx.Rollback()
// 			return err
// 		}
// 	}

// 	return tx.Commit()
// }

// CloneWorkflowApplication 克隆工作流应用（包括所有节点）
func (WorkflowFuncs) CloneWorkflowApplication(ctx context.Context, applicationID uint64, newName string) (*models.WorkflowApplicationResponse, error) {
	// 获取原应用
	originalApp, err := database.Client.WorkflowApplication.Query().
		Where(workflowapplication.ID(applicationID)).
		WithNodes().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow application not found")
		}
		return nil, err
	}

	// 生成新的客户端密钥
	clientSecret, err := generateClientSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client secret: %w", err)
	}

	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, err
	}

	// 创建新应用
	newApp, err := tx.WorkflowApplication.Create().
		SetName(newName).
		SetDescription(originalApp.Description).
		SetStartNodeID(originalApp.StartNodeID).
		SetClientSecret(clientSecret).
		SetVariables(originalApp.Variables).
		SetStatus(workflowapplication.StatusDraft).
		Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// 克隆所有节点
	nodeIDMap := make(map[uint64]uint64) // 旧ID -> 新ID
	for _, oldNode := range originalApp.Edges.Nodes {
		newNode, err := tx.WorkflowNode.Create().
			SetName(oldNode.Name).
			SetType(oldNode.Type).
			SetDescription(oldNode.Description).
			SetPrompt(oldNode.Prompt).
			SetConfig(oldNode.Config).
			SetApplicationID(newApp.ID).
			SetProcessorLanguage(oldNode.ProcessorLanguage).
			SetProcessorCode(oldNode.ProcessorCode).
			SetBranchNodes(oldNode.BranchNodes).
			SetParallelConfig(oldNode.ParallelConfig).
			SetAPIConfig(oldNode.APIConfig).
			SetAsync(oldNode.Async).
			SetTimeout(oldNode.Timeout).
			SetRetryCount(oldNode.RetryCount).
			SetPositionX(oldNode.PositionX).
			SetPositionY(oldNode.PositionY).
			SetColor(oldNode.Color).
			Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		nodeIDMap[oldNode.ID] = newNode.ID
	}

	// 更新新应用的起始节点ID
	if newStartNodeID, ok := nodeIDMap[originalApp.StartNodeID]; ok {
		err = tx.WorkflowApplication.UpdateOneID(newApp.ID).
			SetStartNodeID(newStartNodeID).
			Exec(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return WorkflowFuncs{}.GetWorkflowApplicationByID(ctx, newApp.ID)
}

// // ============ WorkflowEdge CRUD ============

// GetAllWorkflowEdges 获取所有工作流边
func (WorkflowFuncs) GetAllWorkflowEdges(ctx context.Context) ([]*models.WorkflowEdgeResponse, error) {
	edges, err := database.Client.WorkflowEdge.Query().
		WithApplication().
		WithSourceNode().
		WithTargetNode().
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	edgeResponses := make([]*models.WorkflowEdgeResponse, 0, len(edges))
	for _, edge := range edges {
		edgeResponses = append(edgeResponses, WorkflowFuncs{}.ConvertWorkflowEdgeToResponse(edge))
	}

	return edgeResponses, nil
}

// GetWorkflowEdgeByID 根据ID获取工作流边
func (WorkflowFuncs) GetWorkflowEdgeByID(ctx context.Context, id uint64) (*models.WorkflowEdgeResponse, error) {
	edge, err := database.Client.WorkflowEdge.Query().
		Where(workflowedge.ID(id)).
		WithApplication().
		WithSourceNode().
		WithTargetNode().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow edge not found")
		}
		return nil, err
	}
	return WorkflowFuncs{}.ConvertWorkflowEdgeToResponse(edge), nil
}

// GetWorkflowEdgesByApplicationID 根据应用ID获取所有边
func (WorkflowFuncs) GetWorkflowEdgesByApplicationID(ctx context.Context, applicationID uint64) ([]*models.WorkflowEdgeResponse, error) {
	edges, err := database.Client.WorkflowEdge.Query().
		Where(workflowedge.ApplicationIDEQ(applicationID)).
		WithApplication().
		WithSourceNode().
		WithTargetNode().
		Order(ent.Asc(workflowedge.FieldCreateTime)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	edgeResponses := make([]*models.WorkflowEdgeResponse, 0, len(edges))
	for _, edge := range edges {
		edgeResponses = append(edgeResponses, WorkflowFuncs{}.ConvertWorkflowEdgeToResponse(edge))
	}

	return edgeResponses, nil
}

// CreateWorkflowEdge 创建工作流边
func (WorkflowFuncs) CreateWorkflowEdge(ctx context.Context, req *models.CreateWorkflowEdgeRequest) (*models.WorkflowEdgeResponse, error) {
	applicationID := utils.StringToUint64(req.ApplicationID)
	sourceNodeID := utils.StringToUint64(req.SourceNodeID)
	targetNodeID := utils.StringToUint64(req.TargetNodeID)

	builder := database.Client.WorkflowEdge.Create().
		SetApplicationID(applicationID).
		SetSourceNodeID(sourceNodeID).
		SetTargetNodeID(targetNodeID)

	if req.SourceHandle != "" {
		builder = builder.SetSourceHandle(req.SourceHandle)
	}

	if req.TargetHandle != "" {
		builder = builder.SetTargetHandle(req.TargetHandle)
	}

	if req.Type != "" {
		builder = builder.SetType(workflowedge.Type(req.Type))
	}

	if req.Label != "" {
		builder = builder.SetLabel(req.Label)
	}

	if req.BranchName != "" {
		builder = builder.SetBranchName(req.BranchName)
	}

	builder = builder.SetAnimated(req.Animated)

	if req.Style != nil {
		builder = builder.SetStyle(req.Style)
	}

	if req.Data != nil {
		builder = builder.SetData(req.Data)
	}

	edge, err := builder.Save(ctx)
	if err != nil {
		return nil, err
	}

	return WorkflowFuncs{}.GetWorkflowEdgeByID(ctx, edge.ID)
}

// UpdateWorkflowEdge 更新工作流边
func (WorkflowFuncs) UpdateWorkflowEdge(ctx context.Context, id uint64, req *models.UpdateWorkflowEdgeRequest) (*models.WorkflowEdgeResponse, error) {
	builder := database.Client.WorkflowEdge.UpdateOneID(id)

	if req.SourceHandle != "" {
		builder = builder.SetSourceHandle(req.SourceHandle)
	}

	if req.TargetHandle != "" {
		builder = builder.SetTargetHandle(req.TargetHandle)
	}

	if req.Type != "" {
		builder = builder.SetType(workflowedge.Type(req.Type))
	}

	if req.Label != "" {
		builder = builder.SetLabel(req.Label)
	}

	if req.BranchName != "" {
		builder = builder.SetBranchName(req.BranchName)
	}

	if req.Animated != nil {
		builder = builder.SetAnimated(*req.Animated)
	}

	if req.Style != nil {
		builder = builder.SetStyle(req.Style)
	}

	if req.Data != nil {
		builder = builder.SetData(req.Data)
	}

	err := builder.Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow edge not found")
		}
		return nil, err
	}

	return WorkflowFuncs{}.GetWorkflowEdgeByID(ctx, id)
}

// DeleteWorkflowEdge 删除工作流边(软删除)
func (WorkflowFuncs) DeleteWorkflowEdge(ctx context.Context, id uint64) error {
	err := database.Client.WorkflowEdge.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("workflow edge not found")
		}
		return err
	}
	return nil
}

// BatchCreateWorkflowEdges 批量创建工作流边
func (WorkflowFuncs) BatchCreateWorkflowEdges(ctx context.Context, req *models.BatchCreateWorkflowEdgesRequest) ([]*models.WorkflowEdgeResponse, error) {
	responses := make([]*models.WorkflowEdgeResponse, 0, len(req.Edges))

	for _, edgeReq := range req.Edges {
		edge, err := WorkflowFuncs{}.CreateWorkflowEdge(ctx, &edgeReq)
		if err != nil {
			return nil, fmt.Errorf("failed to create edge: %w", err)
		}
		responses = append(responses, edge)
	}

	return responses, nil
}

// BatchDeleteWorkflowEdges 批量删除工作流边
func (WorkflowFuncs) BatchDeleteWorkflowEdges(ctx context.Context, req *models.BatchDeleteWorkflowEdgesRequest) error {
	for _, edgeIDStr := range req.EdgeIDs {
		edgeID := utils.StringToUint64(edgeIDStr)
		err := WorkflowFuncs{}.DeleteWorkflowEdge(ctx, edgeID)
		if err != nil {
			return fmt.Errorf("failed to delete edge %s: %w", edgeIDStr, err)
		}
	}
	return nil
}

// ConvertWorkflowEdgeToResponse 将工作流边实体转换为响应格式
func (WorkflowFuncs) ConvertWorkflowEdgeToResponse(edge *ent.WorkflowEdge) *models.WorkflowEdgeResponse {
	resp := &models.WorkflowEdgeResponse{
		ID:            utils.Uint64ToString(edge.ID),
		CreateTime:    utils.FormatDateTime(edge.CreateTime),
		UpdateTime:    utils.FormatDateTime(edge.UpdateTime),
		ApplicationID: utils.Uint64ToString(edge.ApplicationID),
		SourceNodeID:  utils.Uint64ToString(edge.SourceNodeID), // 返回数据库 ID
		TargetNodeID:  utils.Uint64ToString(edge.TargetNodeID), // 返回数据库 ID
		SourceHandle:  edge.SourceHandle,
		TargetHandle:  edge.TargetHandle,
		Type:          string(edge.Type),
		Label:         edge.Label,
		BranchName:    edge.BranchName,
		Animated:      edge.Animated,
		Style:         edge.Style,
		Data:          edge.Data,
	}

	return resp
}

// ============ WorkflowVersion CRUD ============

// CreateWorkflowVersion 创建工作流版本快照
func (WorkflowFuncs) CreateWorkflowVersion(ctx context.Context, req *models.CreateWorkflowVersionRequest) (*models.WorkflowVersionResponse, error) {
	applicationID := utils.StringToUint64(req.ApplicationID)

	// 1. 查询应用是否存在且为激活状态
	_, err := database.Client.WorkflowApplication.Query().Where(
		workflowapplication.ID(applicationID),
		workflowapplication.StatusEQ(workflowapplication.StatusArchived),
	).Count(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow application not found")
		}
		return nil, err
	}

	// 2. 查询当前应用的最大版本号
	maxVersion, err := database.Client.WorkflowVersion.Query().
		Where(workflowversion.ApplicationID(applicationID)).
		Aggregate(ent.Max(workflowversion.FieldVersion)).
		Int(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	// 新版本号 = 最大版本号 + 1
	newVersion := uint(maxVersion + 1)

	// 3. 查询所有节点
	nodes, err := database.Client.WorkflowNode.Query().
		Where(workflownode.ApplicationID(applicationID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 4. 查询所有边
	edges, err := database.Client.WorkflowEdge.Query().
		Where(workflowedge.ApplicationID(applicationID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 5. 构建快照数据
	nodeResponses := make([]*models.WorkflowNodeResponse, 0, len(nodes))
	for _, node := range nodes {
		nodeResponses = append(nodeResponses, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
	}

	edgeResponses := make([]*models.WorkflowEdgeResponse, 0, len(edges))
	for _, edge := range edges {
		edgeResponses = append(edgeResponses, WorkflowFuncs{}.ConvertWorkflowEdgeToResponse(edge))
	}

	snapshot := models.WorkflowVersionSnapshot{
		Nodes: nodeResponses,
		Edges: edgeResponses,
	}

	// 6. 将快照转换为 map[string]interface{} 以存储到数据库
	snapshotMap := map[string]interface{}{
		"nodes": nodeResponses,
		"edges": edgeResponses,
	}

	// 7. 创建版本记录
	version, err := database.Client.WorkflowVersion.Create().
		SetApplicationID(applicationID).
		SetVersion(newVersion).
		SetSnapshot(snapshotMap).
		SetNillableChangeLog(&req.ChangeLog).
		Save(ctx)
	if err != nil {
		return nil, err
	}

	// 8. 返回响应
	return &models.WorkflowVersionResponse{
		ID:            utils.Uint64ToString(version.ID),
		CreateTime:    utils.FormatDateTime(version.CreateTime),
		UpdateTime:    utils.FormatDateTime(version.UpdateTime),
		ApplicationID: utils.Uint64ToString(version.ApplicationID),
		Version:       version.Version,
		Snapshot:      snapshot,
		ChangeLog:     version.ChangeLog,
	}, nil
}

// GetWorkflowVersionsByApplicationID 根据应用ID获取所有版本
func (WorkflowFuncs) GetWorkflowVersionsByApplicationID(ctx context.Context, applicationID uint64) ([]*models.WorkflowVersionResponse, error) {
	versions, err := database.Client.WorkflowVersion.Query().
		Where(workflowversion.ApplicationID(applicationID)).
		Order(ent.Desc(workflowversion.FieldVersion)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*models.WorkflowVersionResponse, 0, len(versions))
	for _, version := range versions {
		responses = append(responses, WorkflowFuncs{}.ConvertWorkflowVersionToResponse(version))
	}

	return responses, nil
}

// GetWorkflowVersionByID 根据ID获取版本
func (WorkflowFuncs) GetWorkflowVersionByID(ctx context.Context, id uint64) (*models.WorkflowVersionResponse, error) {
	version, err := database.Client.WorkflowVersion.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("workflow version not found")
		}
		return nil, err
	}

	return WorkflowFuncs{}.ConvertWorkflowVersionToResponse(version), nil
}

// ConvertWorkflowVersionToResponse 将工作流版本实体转换为响应格式
func (WorkflowFuncs) ConvertWorkflowVersionToResponse(version *ent.WorkflowVersion) *models.WorkflowVersionResponse {
	// 从 snapshot map 中提取节点和边
	var snapshot models.WorkflowVersionSnapshot

	// 使用 JSON 序列化/反序列化来转换类型
	snapshotBytes, err := json.Marshal(version.Snapshot)
	if err == nil {
		json.Unmarshal(snapshotBytes, &snapshot)
	}

	return &models.WorkflowVersionResponse{
		ID:            utils.Uint64ToString(version.ID),
		CreateTime:    utils.FormatDateTime(version.CreateTime),
		UpdateTime:    utils.FormatDateTime(version.UpdateTime),
		ApplicationID: utils.Uint64ToString(version.ApplicationID),
		Version:       version.Version,
		Snapshot:      snapshot,
		ChangeLog:     version.ChangeLog,
	}
}

// ============ Batch Save ============

// BatchSaveWorkflow 批量保存工作流（节点和边的增删改）
func (WorkflowFuncs) BatchSaveWorkflow(ctx context.Context, req *models.BatchSaveWorkflowRequest) (*models.BatchSaveWorkflowData, error) {
	applicationID := utils.StringToUint64(req.ApplicationID)

	// 验证应用是否存在
	exists, err := database.Client.WorkflowApplication.Query().
		Where(workflowapplication.ID(applicationID)).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check application existence: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("workflow application not found")
	}

	// 使用事务确保所有操作要么全部成功，要么全部失败
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}

	result := &models.BatchSaveWorkflowData{
		NodeIDMapping:  make(map[string]string),
		EdgeIDMapping:  make(map[string]string),
		CreatedNodes:   make([]*models.WorkflowNodeResponse, 0),
		UpdatedNodes:   make([]*models.WorkflowNodeResponse, 0),
		DeletedNodeIDs: make([]string, 0),
		CreatedEdges:   make([]*models.WorkflowEdgeResponse, 0),
		UpdatedEdges:   make([]*models.WorkflowEdgeResponse, 0),
		DeletedEdgeIDs: make([]string, 0),
		Stats: models.BatchSaveWorkflowStats{
			NodesCreated: 0,
			NodesUpdated: 0,
			NodesDeleted: 0,
			EdgesCreated: 0,
			EdgesUpdated: 0,
			EdgesDeleted: 0,
		},
	}

	// 临时ID到数据库ID的映射表（用于边的创建）
	// 注意：我们需要从前端请求中获取临时ID，这里通过请求数组的顺序来建立映射
	tempIDToDBID := make(map[string]uint64)

	// 1. 创建节点
	for i, nodeReq := range req.NodesToCreate {
		builder := tx.WorkflowNode.Create().
			SetName(nodeReq.Name).
			SetType(workflownode.Type(nodeReq.Type)).
			SetConfig(nodeReq.Config).
			SetApplicationID(applicationID)

		if nodeReq.Description != "" {
			builder = builder.SetDescription(nodeReq.Description)
		}
		if nodeReq.Prompt != "" {
			builder = builder.SetPrompt(nodeReq.Prompt)
		}
		if nodeReq.ProcessorLanguage != "" {
			builder = builder.SetProcessorLanguage(nodeReq.ProcessorLanguage)
		}
		if nodeReq.ProcessorCode != "" {
			builder = builder.SetProcessorCode(nodeReq.ProcessorCode)
		}
		if nodeReq.APIConfig != nil {
			builder = builder.SetAPIConfig(nodeReq.APIConfig)
		}
		if nodeReq.ParallelConfig != nil {
			builder = builder.SetParallelConfig(nodeReq.ParallelConfig)
		}
		if nodeReq.BranchNodes != nil {
			builder = builder.SetBranchNodes(nodeReq.BranchNodes)
		}
		if nodeReq.Async != nil {
			builder = builder.SetAsync(*nodeReq.Async)
		}
		if nodeReq.Timeout != nil {
			builder = builder.SetTimeout(*nodeReq.Timeout)
		}
		if nodeReq.RetryCount != nil {
			builder = builder.SetRetryCount(*nodeReq.RetryCount)
		}
		if nodeReq.Color != "" {
			builder = builder.SetColor(nodeReq.Color)
		}
		if nodeReq.PositionX != nil {
			builder = builder.SetPositionX(*nodeReq.PositionX)
		}
		if nodeReq.PositionY != nil {
			builder = builder.SetPositionY(*nodeReq.PositionY)
		}

		node, err := builder.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create node: %w", err)
		}

		// 记录临时ID到数据库ID的映射
		if i < len(req.NodeTempIDs) {
			tempID := req.NodeTempIDs[i]
			dbID := utils.Uint64ToString(node.ID)
			tempIDToDBID[tempID] = node.ID
			result.NodeIDMapping[tempID] = dbID
		}

		result.CreatedNodes = append(result.CreatedNodes, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
		result.Stats.NodesCreated++
	}

	// 2. 更新节点
	for _, nodeUpdate := range req.NodesToUpdate {
		nodeID := utils.StringToUint64(nodeUpdate.ID)
		nodeReq := nodeUpdate.Data

		builder := tx.WorkflowNode.UpdateOneID(nodeID)

		if nodeReq.Name != "" {
			builder = builder.SetName(nodeReq.Name)
		}
		if nodeReq.Description != "" {
			builder = builder.SetDescription(nodeReq.Description)
		}
		if nodeReq.Config != nil {
			builder = builder.SetConfig(nodeReq.Config)
		}
		if nodeReq.Prompt != "" {
			builder = builder.SetPrompt(nodeReq.Prompt)
		}
		if nodeReq.ProcessorLanguage != "" {
			builder = builder.SetProcessorLanguage(nodeReq.ProcessorLanguage)
		}
		if nodeReq.ProcessorCode != "" {
			builder = builder.SetProcessorCode(nodeReq.ProcessorCode)
		}
		if nodeReq.APIConfig != nil {
			builder = builder.SetAPIConfig(nodeReq.APIConfig)
		}
		if nodeReq.ParallelConfig != nil {
			builder = builder.SetParallelConfig(nodeReq.ParallelConfig)
		}
		if nodeReq.BranchNodes != nil {
			builder = builder.SetBranchNodes(nodeReq.BranchNodes)
		}
		if nodeReq.Async != nil {
			builder = builder.SetAsync(*nodeReq.Async)
		}
		if nodeReq.Timeout != nil {
			builder = builder.SetTimeout(*nodeReq.Timeout)
		}
		if nodeReq.RetryCount != nil {
			builder = builder.SetRetryCount(*nodeReq.RetryCount)
		}
		if nodeReq.Color != "" {
			builder = builder.SetColor(nodeReq.Color)
		}
		if nodeReq.PositionX != nil {
			builder = builder.SetPositionX(*nodeReq.PositionX)
		}
		if nodeReq.PositionY != nil {
			builder = builder.SetPositionY(*nodeReq.PositionY)
		}

		node, err := builder.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update node %s: %w", nodeUpdate.ID, err)
		}

		result.UpdatedNodes = append(result.UpdatedNodes, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
		result.Stats.NodesUpdated++
	}

	// 3. 删除节点
	for _, nodeIDStr := range req.NodeIDsToDelete {
		nodeID := utils.StringToUint64(nodeIDStr)

		err := tx.WorkflowNode.DeleteOneID(nodeID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete node %s: %w", nodeIDStr, err)
		}

		result.DeletedNodeIDs = append(result.DeletedNodeIDs, nodeIDStr)
		result.Stats.NodesDeleted++
	}

	// 4. 创建边
	for i, edgeReq := range req.EdgesToCreate {
		// 解析节点ID，优先从临时ID映射表查找，如果找不到再尝试解析为数据库ID
		var sourceNodeID uint64
		if dbID, ok := tempIDToDBID[edgeReq.SourceNodeID]; ok {
			// 从临时ID映射表找到
			sourceNodeID = dbID
		} else {
			// 尝试解析为数据库ID
			sourceNodeID = utils.StringToUint64(edgeReq.SourceNodeID)
			if sourceNodeID == 0 {
				tx.Rollback()
				return nil, fmt.Errorf("source node ID not found: %s (neither database ID nor temp ID)", edgeReq.SourceNodeID)
			}
		}

		var targetNodeID uint64
		if dbID, ok := tempIDToDBID[edgeReq.TargetNodeID]; ok {
			// 从临时ID映射表找到
			targetNodeID = dbID
		} else {
			// 尝试解析为数据库ID
			targetNodeID = utils.StringToUint64(edgeReq.TargetNodeID)
			if targetNodeID == 0 {
				tx.Rollback()
				return nil, fmt.Errorf("target node ID not found: %s (neither database ID nor temp ID)", edgeReq.TargetNodeID)
			}
		}

		builder := tx.WorkflowEdge.Create().
			SetApplicationID(applicationID).
			SetSourceNodeID(sourceNodeID).
			SetTargetNodeID(targetNodeID)

		if edgeReq.SourceHandle != "" {
			builder = builder.SetSourceHandle(edgeReq.SourceHandle)
		}
		if edgeReq.TargetHandle != "" {
			builder = builder.SetTargetHandle(edgeReq.TargetHandle)
		}
		if edgeReq.Type != "" {
			builder = builder.SetType(workflowedge.Type(edgeReq.Type))
		}
		if edgeReq.Label != "" {
			builder = builder.SetLabel(edgeReq.Label)
		}
		if edgeReq.BranchName != "" {
			builder = builder.SetBranchName(edgeReq.BranchName)
		}
		// Animated 在 CreateWorkflowEdgeRequest 中是 bool 类型，直接设置
		builder = builder.SetAnimated(edgeReq.Animated)
		if edgeReq.Style != nil {
			builder = builder.SetStyle(edgeReq.Style)
		}
		if edgeReq.Data != nil {
			builder = builder.SetData(edgeReq.Data)
		}

		edge, err := builder.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create edge: %w", err)
		}

		// 记录临时ID到数据库ID的映射
		if i < len(req.EdgeTempIDs) {
			tempID := req.EdgeTempIDs[i]
			dbID := utils.Uint64ToString(edge.ID)
			result.EdgeIDMapping[tempID] = dbID
		}

		result.CreatedEdges = append(result.CreatedEdges, WorkflowFuncs{}.ConvertWorkflowEdgeToResponse(edge))
		result.Stats.EdgesCreated++
	}

	// 5. 更新边
	for _, edgeUpdate := range req.EdgesToUpdate {
		edgeID := utils.StringToUint64(edgeUpdate.ID)
		edgeReq := edgeUpdate.Data

		builder := tx.WorkflowEdge.UpdateOneID(edgeID)

		if edgeReq.SourceHandle != "" {
			builder = builder.SetSourceHandle(edgeReq.SourceHandle)
		}
		if edgeReq.TargetHandle != "" {
			builder = builder.SetTargetHandle(edgeReq.TargetHandle)
		}
		if edgeReq.Type != "" {
			builder = builder.SetType(workflowedge.Type(edgeReq.Type))
		}
		if edgeReq.Label != "" {
			builder = builder.SetLabel(edgeReq.Label)
		}
		if edgeReq.BranchName != "" {
			builder = builder.SetBranchName(edgeReq.BranchName)
		}
		if edgeReq.Animated != nil {
			builder = builder.SetAnimated(*edgeReq.Animated)
		}
		if edgeReq.Style != nil {
			builder = builder.SetStyle(edgeReq.Style)
		}
		if edgeReq.Data != nil {
			builder = builder.SetData(edgeReq.Data)
		}

		edge, err := builder.Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to update edge %s: %w", edgeUpdate.ID, err)
		}

		result.UpdatedEdges = append(result.UpdatedEdges, WorkflowFuncs{}.ConvertWorkflowEdgeToResponse(edge))
		result.Stats.EdgesUpdated++
	}

	// 6. 删除边
	for _, edgeIDStr := range req.EdgeIDsToDelete {
		edgeID := utils.StringToUint64(edgeIDStr)

		err := tx.WorkflowEdge.DeleteOneID(edgeID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete edge %s: %w", edgeIDStr, err)
		}

		result.DeletedEdgeIDs = append(result.DeletedEdgeIDs, edgeIDStr)
		result.Stats.EdgesDeleted++
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}
