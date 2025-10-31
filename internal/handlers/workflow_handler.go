package handlers

import (
	"net/http"
	"strconv"

	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// WorkflowHandler 工作流处理器
type WorkflowHandler struct{}

// NewWorkflowHandler 创建新的工作流处理器
func NewWorkflowHandler() *WorkflowHandler {
	return &WorkflowHandler{}
}

// ============ WorkflowApplication Handlers ============

// GetWorkflowApplications 获取所有工作流应用
// @Summary      获取所有工作流应用
// @Description  获取系统中所有工作流应用的列表
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]models.WorkflowApplicationResponse,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/applications [get]
func (h *WorkflowHandler) GetWorkflowApplications(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	apps, err := funcs.WorkflowFuncs{}.GetAllWorkflowApplications(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取工作流应用列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apps,
		"count":   len(apps),
	})
}

// GetWorkflowApplicationsWithPagination 分页获取工作流应用列表
// @Summary      分页获取工作流应用列表
// @Description  根据分页参数获取工作流应用列表
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Param        page      query     int     false  "页码"         default(1)
// @Param        pageSize  query     int     false  "每页数量"      default(10)
// @Param        order     query     string  false  "排序方式"      default(desc)
// @Param        orderBy   query     string  false  "排序字段"      default(createTime)
// @Param        name      query     string  false  "应用名称"
// @Param        status    query     string  false  "状态"
// @Success      200  {object}  object{success=bool,data=[]models.WorkflowApplicationResponse,pagination=models.Pagination}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/applications/page [get]
func (h *WorkflowHandler) GetWorkflowApplicationsWithPagination(c *gin.Context) {
	var req models.PageWorkflowApplicationRequest

	// 设置默认值
	req.Page = 1
	req.PageSize = 10
	req.Order = "desc"
	req.OrderBy = "createTime"

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("查询参数格式错误", err.Error()))
		return
	}

	// 调用服务层方法
	ctx := middleware.GetRequestContext(c)
	result, err := funcs.WorkflowFuncs{}.GetWorkflowApplicationsWithPagination(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取工作流应用列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

// GetWorkflowApplication 根据ID获取工作流应用
// @Summary      根据ID获取工作流应用
// @Description  根据工作流应用ID获取详细信息
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "工作流应用ID"
// @Success      200  {object}  object{success=bool,data=models.WorkflowApplicationResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/applications/{id} [get]
func (h *WorkflowHandler) GetWorkflowApplication(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	app, err := funcs.WorkflowFuncs{}.GetWorkflowApplicationByID(ctx, id)
	if err != nil {
		if err.Error() == "workflow application not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流应用未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询工作流应用失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    app,
	})
}

// CreateWorkflowApplication 创建工作流应用
// @Summary      创建工作流应用
// @Description  创建新的工作流应用
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Param        application  body      models.CreateWorkflowApplicationRequest  true  "工作流应用信息"
// @Success      201   {object}  object{success=bool,data=models.WorkflowApplicationResponse}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/applications [post]
func (h *WorkflowHandler) CreateWorkflowApplication(c *gin.Context) {
	var req models.CreateWorkflowApplicationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用名称不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	app, err := funcs.WorkflowFuncs{}.CreateWorkflowApplication(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建工作流应用失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    app,
		"message": "工作流应用创建成功",
	})
}

// UpdateWorkflowApplication 更新工作流应用
// @Summary      更新工作流应用
// @Description  根据ID更新工作流应用信息
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Param        id            path      string                                   true  "工作流应用ID"
// @Param        application   body      models.UpdateWorkflowApplicationRequest  true  "工作流应用信息"
// @Success      200   {object}  object{success=bool,data=models.WorkflowApplicationResponse}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/applications/{id} [put]
func (h *WorkflowHandler) UpdateWorkflowApplication(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateWorkflowApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	app, err := funcs.WorkflowFuncs{}.UpdateWorkflowApplication(ctx, id, &req)
	if err != nil {
		if err.Error() == "workflow application not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流应用未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新工作流应用失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    app,
		"message": "工作流应用更新成功",
	})
}

// DeleteWorkflowApplication 删除工作流应用
// @Summary      删除工作流应用
// @Description  根据ID删除工作流应用
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "工作流应用ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/applications/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflowApplication(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.DeleteWorkflowApplication(ctx, id)
	if err != nil {
		if err.Error() == "workflow application not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流应用未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除工作流应用失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工作流应用删除成功",
	})
}

// CloneWorkflowApplication 克隆工作流应用
// @Summary      克隆工作流应用
// @Description  克隆一个工作流应用及其所有节点
// @Tags         workflow-applications
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "工作流应用ID"
// @Param        body  body      object{name=string}  true  "新应用名称"
// @Success      201   {object}  object{success=bool,data=models.WorkflowApplicationResponse}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/applications/{id}/clone [post]
func (h *WorkflowHandler) CloneWorkflowApplication(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	app, err := funcs.WorkflowFuncs{}.CloneWorkflowApplication(ctx, id, req.Name)
	if err != nil {
		if err.Error() == "workflow application not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流应用未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("克隆工作流应用失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    app,
		"message": "工作流应用克隆成功",
	})
}

// ============ WorkflowNode Handlers ============

// GetWorkflowNodes 获取所有工作流节点
// @Summary      获取所有工作流节点
// @Description  获取系统中所有工作流节点的列表
// @Tags         workflow-nodes
// @Accept       json
// @Produce      json
// @Success      200  {object}  object{success=bool,data=[]models.WorkflowNodeResponse,count=int}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/nodes [get]
func (h *WorkflowHandler) GetWorkflowNodes(c *gin.Context) {
	ctx := middleware.GetRequestContext(c)
	nodes, err := funcs.WorkflowFuncs{}.GetAllWorkflowNodes(ctx)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取工作流节点列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nodes,
		"count":   len(nodes),
	})
}

// GetWorkflowNodesByApplicationID 根据应用ID获取工作流节点
// @Summary      根据应用ID获取工作流节点
// @Description  获取指定工作流应用的所有节点
// @Tags         workflow-nodes
// @Accept       json
// @Produce      json
// @Param        applicationId  query     string  true  "工作流应用ID"
// @Success      200  {object}  object{success=bool,data=[]models.WorkflowNodeResponse,count=int}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/nodes/by-application [get]
func (h *WorkflowHandler) GetWorkflowNodesByApplicationID(c *gin.Context) {
	applicationIDStr := c.Query("applicationId")

	if applicationIDStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID不能为空", nil))
		return
	}

	applicationID, err := strconv.ParseUint(applicationIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流应用ID格式无效", map[string]any{
			"provided_id": applicationIDStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	nodes, err := funcs.WorkflowFuncs{}.GetWorkflowNodesByApplicationID(ctx, applicationID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取工作流节点列表失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    nodes,
		"count":   len(nodes),
	})
}

// GetWorkflowNode 根据ID获取工作流节点
// @Summary      根据ID获取工作流节点
// @Description  根据工作流节点ID获取详细信息
// @Tags         workflow-nodes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "工作流节点ID"
// @Success      200  {object}  object{success=bool,data=models.WorkflowNodeResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/nodes/{id} [get]
func (h *WorkflowHandler) GetWorkflowNode(c *gin.Context) {
	idStr := c.Param("id")

	if idStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("工作流节点ID不能为空", nil))
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流节点ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	node, err := funcs.WorkflowFuncs{}.GetWorkflowNodeByID(ctx, id)
	if err != nil {
		if err.Error() == "workflow node not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流节点未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("查询工作流节点失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    node,
	})
}

// CreateWorkflowNode 创建工作流节点
// @Summary      创建工作流节点
// @Description  创建新的工作流节点
// @Tags         workflow-nodes
// @Accept       json
// @Produce      json
// @Param        node  body      models.CreateWorkflowNodeRequest  true  "工作流节点信息"
// @Success      201   {object}  object{success=bool,data=models.WorkflowNodeResponse}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/nodes [post]
func (h *WorkflowHandler) CreateWorkflowNode(c *gin.Context) {
	var req models.CreateWorkflowNodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	if req.Name == "" {
		middleware.ThrowError(c, middleware.BadRequestError("工作流节点名称不能为空", nil))
		return
	}

	ctx := middleware.GetRequestContext(c)
	node, err := funcs.WorkflowFuncs{}.CreateWorkflowNode(ctx, &req)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("创建工作流节点失败", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    node,
		"message": "工作流节点创建成功",
	})
}

// UpdateWorkflowNode 更新工作流节点
// @Summary      更新工作流节点
// @Description  根据ID更新工作流节点信息
// @Tags         workflow-nodes
// @Accept       json
// @Produce      json
// @Param        id    path      string                            true  "工作流节点ID"
// @Param        node  body      models.UpdateWorkflowNodeRequest  true  "工作流节点信息"
// @Success      200   {object}  object{success=bool,data=models.WorkflowNodeResponse}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/nodes/{id} [put]
func (h *WorkflowHandler) UpdateWorkflowNode(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流节点ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	var req models.UpdateWorkflowNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	node, err := funcs.WorkflowFuncs{}.UpdateWorkflowNode(ctx, id, &req)
	if err != nil {
		if err.Error() == "workflow node not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流节点未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新工作流节点失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    node,
		"message": "工作流节点更新成功",
	})
}

// DeleteWorkflowNode 删除工作流节点
// @Summary      删除工作流节点
// @Description  根据ID删除工作流节点
// @Tags         workflow-nodes
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "工作流节点ID"
// @Success      200  {object}  object{success=bool,message=string}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      404  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/nodes/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflowNode(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("工作流节点ID格式无效", map[string]any{
			"provided_id": idStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.DeleteWorkflowNode(ctx, id)
	if err != nil {
		if err.Error() == "workflow node not found" {
			middleware.ThrowError(c, middleware.NotFoundError("工作流节点未找到", map[string]any{
				"id": id,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("删除工作流节点失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "工作流节点删除成功",
	})
}

// ============ Workflow Graph Operations Handlers ============

// ConnectNodes 连接两个节点（普通next_node_id连接）
// @Summary      连接两个节点
// @Description  在两个工作流节点之间创建普通连接（next_node_id）
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{fromNodeId=string,toNodeId=string}  true  "节点连接信息"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/connect [post]
func (h *WorkflowHandler) ConnectNodes(c *gin.Context) {
	var req struct {
		FromNodeID string `json:"fromNodeId" binding:"required"`
		ToNodeID   string `json:"toNodeId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	fromNodeID, err := strconv.ParseUint(req.FromNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("源节点ID格式无效", map[string]any{
			"provided_id": req.FromNodeID,
		}))
		return
	}

	toNodeID, err := strconv.ParseUint(req.ToNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("目标节点ID格式无效", map[string]any{
			"provided_id": req.ToNodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.ConnectNodes(ctx, fromNodeID, toNodeID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("连接节点失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点连接成功",
	})
}

// ConnectBranch 为分支节点添加分支连接
// @Summary      添加分支连接
// @Description  为分支节点（如condition_checker）添加分支连接
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{fromNodeId=string,toNodeId=string,branchName=string}  true  "分支连接信息"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/connect-branch [post]
func (h *WorkflowHandler) ConnectBranch(c *gin.Context) {
	var req struct {
		FromNodeID string `json:"fromNodeId" binding:"required"`
		ToNodeID   string `json:"toNodeId" binding:"required"`
		BranchName string `json:"branchName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	fromNodeID, err := strconv.ParseUint(req.FromNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("源节点ID格式无效", map[string]any{
			"provided_id": req.FromNodeID,
		}))
		return
	}

	toNodeID, err := strconv.ParseUint(req.ToNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("目标节点ID格式无效", map[string]any{
			"provided_id": req.ToNodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.ConnectBranch(ctx, fromNodeID, toNodeID, req.BranchName)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("添加分支连接失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "分支连接添加成功",
	})
}

// DisconnectBranch 删除分支连接
// @Summary      删除分支连接
// @Description  删除分支节点的某个分支连接
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{fromNodeId=string,branchName=string}  true  "分支信息"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/disconnect-branch [post]
func (h *WorkflowHandler) DisconnectBranch(c *gin.Context) {
	var req struct {
		FromNodeID string `json:"fromNodeId" binding:"required"`
		BranchName string `json:"branchName" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	fromNodeID, err := strconv.ParseUint(req.FromNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("源节点ID格式无效", map[string]any{
			"provided_id": req.FromNodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.DisconnectBranch(ctx, fromNodeID, req.BranchName)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("删除分支连接失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "分支连接已删除",
	})
}

// AddNodeToParallel 将节点添加到并行执行节点
// @Summary      添加节点到并行执行
// @Description  将节点添加到并行执行节点作为子节点
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{parallelNodeId=string,childNodeId=string}  true  "并行节点信息"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/add-to-parallel [post]
func (h *WorkflowHandler) AddNodeToParallel(c *gin.Context) {
	var req struct {
		ParallelNodeID string `json:"parallelNodeId" binding:"required"`
		ChildNodeID    string `json:"childNodeId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	parallelNodeID, err := strconv.ParseUint(req.ParallelNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("并行节点ID格式无效", map[string]any{
			"provided_id": req.ParallelNodeID,
		}))
		return
	}

	childNodeID, err := strconv.ParseUint(req.ChildNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("子节点ID格式无效", map[string]any{
			"provided_id": req.ChildNodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.AddNodeToParallel(ctx, parallelNodeID, childNodeID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("添加节点到并行执行失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点已添加到并行执行",
	})
}

// RemoveNodeFromParallel 从并行执行节点中移除子节点
// @Summary      从并行执行中移除节点
// @Description  从并行执行节点中移除子节点
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{childNodeId=string}  true  "子节点ID"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/remove-from-parallel [post]
func (h *WorkflowHandler) RemoveNodeFromParallel(c *gin.Context) {
	var req struct {
		ChildNodeID string `json:"childNodeId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	childNodeID, err := strconv.ParseUint(req.ChildNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("子节点ID格式无效", map[string]any{
			"provided_id": req.ChildNodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.RemoveNodeFromParallel(ctx, childNodeID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("从并行执行中移除节点失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点已从并行执行中移除",
	})
}

// GetParallelChildren 获取并行节点的所有子节点
// @Summary      获取并行节点的子节点
// @Description  获取并行执行节点的所有子节点列表
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        parallelNodeId  query     string  true  "并行节点ID"
// @Success      200  {object}  object{success=bool,data=[]models.WorkflowNodeResponse}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/graph/parallel-children [get]
func (h *WorkflowHandler) GetParallelChildren(c *gin.Context) {
	parallelNodeIDStr := c.Query("parallelNodeId")

	if parallelNodeIDStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("并行节点ID不能为空", nil))
		return
	}

	parallelNodeID, err := strconv.ParseUint(parallelNodeIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("并行节点ID格式无效", map[string]any{
			"provided_id": parallelNodeIDStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	children, err := funcs.WorkflowFuncs{}.GetParallelChildren(ctx, parallelNodeID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取并行子节点失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    children,
		"count":   len(children),
	})
}

// GetNodeConnections 获取节点的所有连接信息
// @Summary      获取节点连接信息
// @Description  获取节点的所有连接信息（包括next、branch、parallel等）
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        nodeId  query     string  true  "节点ID"
// @Success      200  {object}  object{success=bool,data=object}
// @Failure      400  {object}  object{success=bool,message=string}
// @Failure      500  {object}  object{success=bool,message=string}
// @Router       /workflow/graph/connections [get]
func (h *WorkflowHandler) GetNodeConnections(c *gin.Context) {
	nodeIDStr := c.Query("nodeId")

	if nodeIDStr == "" {
		middleware.ThrowError(c, middleware.BadRequestError("节点ID不能为空", nil))
		return
	}

	nodeID, err := strconv.ParseUint(nodeIDStr, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("节点ID格式无效", map[string]any{
			"provided_id": nodeIDStr,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	connections, err := funcs.WorkflowFuncs{}.GetNodeConnections(ctx, nodeID)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("获取节点连接信息失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    connections,
	})
}

// DisconnectNodes 断开节点连接
// @Summary      断开节点连接
// @Description  断开工作流节点的连接（删除边）
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{fromNodeId=string}  true  "源节点ID"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/disconnect [post]
func (h *WorkflowHandler) DisconnectNodes(c *gin.Context) {
	var req struct {
		FromNodeID string `json:"fromNodeId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	fromNodeID, err := strconv.ParseUint(req.FromNodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("源节点ID格式无效", map[string]any{
			"provided_id": req.FromNodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.DisconnectNodes(ctx, fromNodeID)
	if err != nil {
		if err.Error() == "source node not found" {
			middleware.ThrowError(c, middleware.NotFoundError("源节点未找到", map[string]any{
				"fromNodeId": fromNodeID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("断开节点连接失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点连接已断开",
	})
}

// UpdateNodePosition 更新节点位置
// @Summary      更新节点位置
// @Description  更新工作流节点在画布上的位置
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{nodeId=string,x=number,y=number}  true  "节点位置信息"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      404   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/position [put]
func (h *WorkflowHandler) UpdateNodePosition(c *gin.Context) {
	var req struct {
		NodeID string  `json:"nodeId" binding:"required"`
		X      float64 `json:"x" binding:"required"`
		Y      float64 `json:"y" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	nodeID, err := strconv.ParseUint(req.NodeID, 10, 64)
	if err != nil {
		middleware.ThrowError(c, middleware.BadRequestError("节点ID格式无效", map[string]any{
			"provided_id": req.NodeID,
		}))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err = funcs.WorkflowFuncs{}.UpdateNodePosition(ctx, nodeID, req.X, req.Y)
	if err != nil {
		if err.Error() == "node not found" {
			middleware.ThrowError(c, middleware.NotFoundError("节点未找到", map[string]any{
				"nodeId": nodeID,
			}))
		} else {
			middleware.ThrowError(c, middleware.DatabaseError("更新节点位置失败", err.Error()))
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点位置更新成功",
	})
}

// BatchUpdateNodePositions 批量更新节点位置
// @Summary      批量更新节点位置
// @Description  批量更新多个工作流节点在画布上的位置
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{positions=object}  true  "节点位置信息"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/positions [put]
func (h *WorkflowHandler) BatchUpdateNodePositions(c *gin.Context) {
	var req struct {
		Positions map[string]map[string]float64 `json:"positions" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	ctx := middleware.GetRequestContext(c)
	err := funcs.WorkflowFuncs{}.BatchUpdateNodePositions(ctx, req.Positions)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("批量更新节点位置失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点位置批量更新成功",
	})
}

// BatchDeleteNodes 批量删除节点
// @Summary      批量删除节点
// @Description  批量删除多个工作流节点
// @Tags         workflow-graph
// @Accept       json
// @Produce      json
// @Param        body  body      object{nodeIds=[]string}  true  "节点ID列表"
// @Success      200   {object}  object{success=bool,message=string}
// @Failure      400   {object}  object{success=bool,message=string}
// @Failure      500   {object}  object{success=bool,message=string}
// @Router       /workflow/graph/batch-delete [post]
func (h *WorkflowHandler) BatchDeleteNodes(c *gin.Context) {
	var req struct {
		NodeIDs []string `json:"nodeIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.ThrowError(c, middleware.ValidationError("请求数据格式错误", err.Error()))
		return
	}

	nodeIDs := make([]uint64, 0, len(req.NodeIDs))
	for _, idStr := range req.NodeIDs {
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			middleware.ThrowError(c, middleware.BadRequestError("节点ID格式无效", map[string]any{
				"provided_id": idStr,
			}))
			return
		}
		nodeIDs = append(nodeIDs, id)
	}

	ctx := middleware.GetRequestContext(c)
	err := funcs.WorkflowFuncs{}.BatchDeleteNodes(ctx, nodeIDs)
	if err != nil {
		middleware.ThrowError(c, middleware.DatabaseError("批量删除节点失败", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "节点批量删除成功",
	})
}
