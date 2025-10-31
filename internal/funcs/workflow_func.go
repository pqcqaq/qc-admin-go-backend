package funcs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"time"

	"go-backend/database/ent"
	"go-backend/database/ent/workflowapplication"
	"go-backend/database/ent/workflownode"
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
		WithNodes().
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
			SetNodeKey("start_node").
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
		ID:           utils.Uint64ToString(app.ID),
		CreateTime:   utils.FormatDateTime(app.CreateTime),
		UpdateTime:   utils.FormatDateTime(app.UpdateTime),
		Name:         app.Name,
		Description:  app.Description,
		StartNodeID:  utils.Uint64ToString(app.StartNodeID),
		ClientSecret: app.ClientSecret,
		Variables:    app.Variables,
		Version:      app.Version,
		Status:       string(app.Status),
	}

	// 转换节点列表
	if len(app.Edges.Nodes) > 0 {
		resp.Nodes = make([]*models.WorkflowNodeResponse, 0, len(app.Edges.Nodes))
		for _, node := range app.Edges.Nodes {
			resp.Nodes = append(resp.Nodes, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(node))
		}
	}

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
		SetNodeKey(req.NodeKey).
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

	if req.NextNodeID != "" {
		nextNodeID := utils.StringToUint64(req.NextNodeID)
		builder = builder.SetNextNodeID(nextNodeID)
	}

	if req.ParentNodeID != "" {
		parentNodeID := utils.StringToUint64(req.ParentNodeID)
		builder = builder.SetParentNodeID(parentNodeID)
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

	if req.NodeKey != "" {
		builder = builder.SetNodeKey(req.NodeKey)
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

	if req.NextNodeID != "" {
		nextNodeID := utils.StringToUint64(req.NextNodeID)
		builder = builder.SetNextNodeID(nextNodeID)
	}

	if req.ParentNodeID != "" {
		parentNodeID := utils.StringToUint64(req.ParentNodeID)
		builder = builder.SetParentNodeID(parentNodeID)
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
		NodeKey:           node.NodeKey,
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
	}

	if node.NextNodeID != 0 {
		resp.NextNodeID = utils.Uint64ToString(node.NextNodeID)
	}

	if node.ParentNodeID != 0 {
		resp.ParentNodeID = utils.Uint64ToString(node.ParentNodeID)
	}

	return resp
}

// ============ Workflow Graph Operations ============

// NodeConnectionRule 节点连接规则
type NodeConnectionRule struct {
	CanHaveNextNode      bool // 是否可以有next_node_id
	CanHaveBranches      bool // 是否可以有分支
	CanBeParallel        bool // 是否可以作为并行节点的子节点
	RequiresBranchName   bool // 连接时是否需要分支名称
	MaxOutputConnections int  // 最大输出连接数 (-1表示无限制)
}

// getNodeConnectionRule 获取节点类型的连接规则
func getNodeConnectionRule(nodeType string) NodeConnectionRule {
	rules := map[string]NodeConnectionRule{
		"user_input": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        false,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
		"todo_task_generator": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        true,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
		"condition_checker": {
			CanHaveNextNode:      false,
			CanHaveBranches:      true,
			CanBeParallel:        true,
			RequiresBranchName:   true,
			MaxOutputConnections: -1, // 可以有多个分支
		},
		"api_caller": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        true,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
		"data_processor": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        true,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
		"while_loop": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        false,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
		"end_node": {
			CanHaveNextNode:      false,
			CanHaveBranches:      false,
			CanBeParallel:        true,
			RequiresBranchName:   false,
			MaxOutputConnections: 0,
		},
		"parallel_executor": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        false,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
		"llm_caller": {
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        true,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		},
	}

	rule, exists := rules[nodeType]
	if !exists {
		// 默认规则
		return NodeConnectionRule{
			CanHaveNextNode:      true,
			CanHaveBranches:      false,
			CanBeParallel:        true,
			RequiresBranchName:   false,
			MaxOutputConnections: 1,
		}
	}
	return rule
}

// ConnectNodes 连接两个节点（普通连接，用于next_node_id）
func (WorkflowFuncs) ConnectNodes(ctx context.Context, fromNodeID, toNodeID uint64) error {
	// 获取源节点
	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("source node not found")
		}
		return err
	}

	// 检查目标节点是否存在
	_, err = database.Client.WorkflowNode.Get(ctx, toNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("target node not found")
		}
		return err
	}

	// 获取节点连接规则
	rule := getNodeConnectionRule(string(fromNode.Type))

	// 检查源节点是否可以有next_node_id
	if !rule.CanHaveNextNode {
		return fmt.Errorf("node type '%s' cannot have next_node connection, use branch connection instead", fromNode.Type)
	}

	// 检查是否已经有连接
	if fromNode.NextNodeID != 0 {
		return fmt.Errorf("node already has a next_node connection, disconnect first")
	}

	// 更新源节点的 next_node_id
	err = database.Client.WorkflowNode.UpdateOneID(fromNodeID).
		SetNextNodeID(toNodeID).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

// DisconnectNodes 断开节点的next_node_id连接
func (WorkflowFuncs) DisconnectNodes(ctx context.Context, fromNodeID uint64) error {
	// 清除源节点的 next_node_id
	err := database.Client.WorkflowNode.UpdateOneID(fromNodeID).
		ClearNextNodeID().
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("source node not found")
		}
		return err
	}
	return nil
}

// ConnectBranch 为分支节点（如condition_checker）添加分支连接
func (WorkflowFuncs) ConnectBranch(ctx context.Context, fromNodeID, toNodeID uint64, branchName string) error {
	if branchName == "" {
		return fmt.Errorf("branch name is required")
	}

	// 获取源节点
	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("source node not found")
		}
		return err
	}

	// 检查目标节点是否存在
	_, err = database.Client.WorkflowNode.Get(ctx, toNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("target node not found")
		}
		return err
	}

	// 获取节点连接规则
	rule := getNodeConnectionRule(string(fromNode.Type))

	// 检查源节点是否可以有分支
	if !rule.CanHaveBranches {
		return fmt.Errorf("node type '%s' cannot have branch connections", fromNode.Type)
	}

	// 获取现有分支
	branchNodes := fromNode.BranchNodes
	if branchNodes == nil {
		branchNodes = make(map[string]uint64)
	}

	// 添加或更新分支
	branchNodes[branchName] = toNodeID

	// 更新节点
	err = database.Client.WorkflowNode.UpdateOneID(fromNodeID).
		SetBranchNodes(branchNodes).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// DisconnectBranch 删除分支节点的某个分支连接
func (WorkflowFuncs) DisconnectBranch(ctx context.Context, fromNodeID uint64, branchName string) error {
	if branchName == "" {
		return fmt.Errorf("branch name is required")
	}

	// 获取源节点
	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("source node not found")
		}
		return err
	}

	// 获取现有分支
	branchNodes := fromNode.BranchNodes
	if branchNodes == nil {
		return fmt.Errorf("node has no branches")
	}

	// 检查分支是否存在
	if _, exists := branchNodes[branchName]; !exists {
		return fmt.Errorf("branch '%s' not found", branchName)
	}

	// 删除分支
	delete(branchNodes, branchName)

	// 更新节点
	err = database.Client.WorkflowNode.UpdateOneID(fromNodeID).
		SetBranchNodes(branchNodes).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// AddNodeToParallel 将节点添加到并行执行节点
func (WorkflowFuncs) AddNodeToParallel(ctx context.Context, parallelNodeID, childNodeID uint64) error {
	// 获取并行节点
	parallelNode, err := database.Client.WorkflowNode.Get(ctx, parallelNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("parallel node not found")
		}
		return err
	}

	// 检查是否是并行执行节点
	if parallelNode.Type != "parallel_executor" {
		return fmt.Errorf("node is not a parallel_executor, got type '%s'", parallelNode.Type)
	}

	// 获取子节点
	childNode, err := database.Client.WorkflowNode.Get(ctx, childNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("child node not found")
		}
		return err
	}

	// 检查子节点类型是否可以作为并行节点
	rule := getNodeConnectionRule(string(childNode.Type))
	if !rule.CanBeParallel {
		return fmt.Errorf("node type '%s' cannot be added to parallel executor", childNode.Type)
	}

	// 检查子节点是否已经有父节点
	if childNode.ParentNodeID != 0 {
		return fmt.Errorf("child node already has a parent node")
	}

	// 设置子节点的parent_node_id
	err = database.Client.WorkflowNode.UpdateOneID(childNodeID).
		SetParentNodeID(parallelNodeID).
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// RemoveNodeFromParallel 从并行执行节点中移除子节点
func (WorkflowFuncs) RemoveNodeFromParallel(ctx context.Context, childNodeID uint64) error {
	// 获取子节点
	childNode, err := database.Client.WorkflowNode.Get(ctx, childNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("child node not found")
		}
		return err
	}

	// 检查是否有父节点
	if childNode.ParentNodeID == 0 {
		return fmt.Errorf("node has no parent node")
	}

	// 清除parent_node_id
	err = database.Client.WorkflowNode.UpdateOneID(childNodeID).
		ClearParentNodeID().
		Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetParallelChildren 获取并行节点的所有子节点
func (WorkflowFuncs) GetParallelChildren(ctx context.Context, parallelNodeID uint64) ([]*models.WorkflowNodeResponse, error) {
	// 获取并行节点
	parallelNode, err := database.Client.WorkflowNode.Get(ctx, parallelNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("parallel node not found")
		}
		return nil, err
	}

	// 检查是否是并行执行节点
	if parallelNode.Type != "parallel_executor" {
		return nil, fmt.Errorf("node is not a parallel_executor")
	}

	// 查询所有parent_node_id为该节点的子节点
	children, err := database.Client.WorkflowNode.Query().
		Where(workflownode.ParentNodeIDEQ(parallelNodeID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 转换为响应格式
	childResponses := make([]*models.WorkflowNodeResponse, 0, len(children))
	for _, child := range children {
		childResponses = append(childResponses, WorkflowFuncs{}.ConvertWorkflowNodeToResponse(child))
	}

	return childResponses, nil
}

// ValidateNodeConnection 验证两个节点是否可以连接
func (WorkflowFuncs) ValidateNodeConnection(ctx context.Context, fromNodeID, toNodeID uint64, connectionType string) error {
	// 获取源节点和目标节点
	fromNode, err := database.Client.WorkflowNode.Get(ctx, fromNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("source node not found")
		}
		return err
	}

	toNode, err := database.Client.WorkflowNode.Get(ctx, toNodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("target node not found")
		}
		return err
	}

	// 检查是否在同一个应用中
	if fromNode.ApplicationID != toNode.ApplicationID {
		return fmt.Errorf("nodes must be in the same workflow application")
	}

	// 检查是否形成循环（简单检查：不能连接到自己）
	if fromNodeID == toNodeID {
		return fmt.Errorf("cannot connect node to itself")
	}

	// 根据连接类型验证
	fromRule := getNodeConnectionRule(string(fromNode.Type))
	toRule := getNodeConnectionRule(string(toNode.Type))

	switch connectionType {
	case "next":
		if !fromRule.CanHaveNextNode {
			return fmt.Errorf("source node type '%s' cannot have next_node connection", fromNode.Type)
		}
	case "branch":
		if !fromRule.CanHaveBranches {
			return fmt.Errorf("source node type '%s' cannot have branch connections", fromNode.Type)
		}
	case "parallel":
		if fromNode.Type != "parallel_executor" {
			return fmt.Errorf("source node must be parallel_executor for parallel connection")
		}
		if !toRule.CanBeParallel {
			return fmt.Errorf("target node type '%s' cannot be added to parallel executor", toNode.Type)
		}
	default:
		return fmt.Errorf("unknown connection type: %s", connectionType)
	}

	return nil
}

// GetNodeConnections 获取节点的所有连接信息
func (WorkflowFuncs) GetNodeConnections(ctx context.Context, nodeID uint64) (map[string]interface{}, error) {
	node, err := database.Client.WorkflowNode.Get(ctx, nodeID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fmt.Errorf("node not found")
		}
		return nil, err
	}

	connections := make(map[string]interface{})

	// Next node connection
	if node.NextNodeID != 0 {
		connections["next_node_id"] = utils.Uint64ToString(node.NextNodeID)
	}

	// Parent node (for parallel children)
	if node.ParentNodeID != 0 {
		connections["parent_node_id"] = utils.Uint64ToString(node.ParentNodeID)
	}

	// Branch connections
	if len(node.BranchNodes) > 0 {
		branches := make(map[string]string)
		for branchName, targetID := range node.BranchNodes {
			branches[branchName] = utils.Uint64ToString(targetID)
		}
		connections["branches"] = branches
	}

	// Parallel children (if this is a parallel executor)
	if node.Type == "parallel_executor" {
		children, err := WorkflowFuncs{}.GetParallelChildren(ctx, nodeID)
		if err == nil && len(children) > 0 {
			childIDs := make([]string, 0, len(children))
			for _, child := range children {
				childIDs = append(childIDs, child.ID)
			}
			connections["parallel_children"] = childIDs
		}
	}

	return connections, nil
}

// UpdateNodePosition 更新节点位置
func (WorkflowFuncs) UpdateNodePosition(ctx context.Context, nodeID uint64, x, y float64) error {
	err := database.Client.WorkflowNode.UpdateOneID(nodeID).
		SetPositionX(x).
		SetPositionY(y).
		Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fmt.Errorf("node not found")
		}
		return err
	}
	return nil
}

// BatchUpdateNodePositions 批量更新节点位置
func (WorkflowFuncs) BatchUpdateNodePositions(ctx context.Context, positions map[string]map[string]float64) error {
	// positions 格式: {"nodeId": {"x": 100, "y": 200}}
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return err
	}

	for nodeIDStr, pos := range positions {
		nodeID := utils.StringToUint64(nodeIDStr)
		x, xOk := pos["x"]
		y, yOk := pos["y"]

		if !xOk || !yOk {
			tx.Rollback()
			return fmt.Errorf("invalid position data for node %s", nodeIDStr)
		}

		err = tx.WorkflowNode.UpdateOneID(nodeID).
			SetPositionX(x).
			SetPositionY(y).
			Exec(ctx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// BatchDeleteNodes 批量删除节点
func (WorkflowFuncs) BatchDeleteNodes(ctx context.Context, nodeIDs []uint64) error {
	tx, err := database.Client.Tx(ctx)
	if err != nil {
		return err
	}

	for _, nodeID := range nodeIDs {
		err = tx.WorkflowNode.DeleteOneID(nodeID).Exec(ctx)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

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
			SetNodeKey(oldNode.NodeKey).
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
			Save(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		nodeIDMap[oldNode.ID] = newNode.ID
	}

	// 更新节点之间的连接关系
	for _, oldNode := range originalApp.Edges.Nodes {
		newNodeID := nodeIDMap[oldNode.ID]
		updateBuilder := tx.WorkflowNode.UpdateOneID(newNodeID)

		if oldNode.NextNodeID != 0 {
			if newNextNodeID, ok := nodeIDMap[oldNode.NextNodeID]; ok {
				updateBuilder = updateBuilder.SetNextNodeID(newNextNodeID)
			}
		}

		if oldNode.ParentNodeID != 0 {
			if newParentNodeID, ok := nodeIDMap[oldNode.ParentNodeID]; ok {
				updateBuilder = updateBuilder.SetParentNodeID(newParentNodeID)
			}
		}

		err = updateBuilder.Exec(ctx)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
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
