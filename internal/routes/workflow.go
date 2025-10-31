package routes

import (
	"go-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

// setupWorkflowRoutes 设置工作流相关路由
func (r *Router) setupWorkflowRoutes(rg *gin.RouterGroup) {

	workflowHandler := handlers.NewWorkflowHandler()

	workflow := rg.Group("/workflow")
	{
		// WorkflowApplication 路由
		applications := workflow.Group("/applications")
		{
			// 基本CRUD操作
			applications.GET("", workflowHandler.GetWorkflowApplications)                    // 获取所有工作流应用
			applications.GET("/page", workflowHandler.GetWorkflowApplicationsWithPagination) // 分页获取工作流应用列表
			applications.GET("/:id", workflowHandler.GetWorkflowApplication)                 // 根据ID获取工作流应用
			applications.POST("", workflowHandler.CreateWorkflowApplication)                 // 创建工作流应用
			applications.PUT("/:id", workflowHandler.UpdateWorkflowApplication)              // 更新工作流应用
			applications.DELETE("/:id", workflowHandler.DeleteWorkflowApplication)           // 删除工作流应用

			// 特殊操作
			applications.POST("/:id/clone", workflowHandler.CloneWorkflowApplication) // 克隆工作流应用
		}

		// WorkflowNode 路由
		nodes := workflow.Group("/nodes")
		{
			// 基本CRUD操作
			nodes.GET("", workflowHandler.GetWorkflowNodes)                               // 获取所有工作流节点
			nodes.GET("/by-application", workflowHandler.GetWorkflowNodesByApplicationID) // 根据应用ID获取节点
			nodes.GET("/:id", workflowHandler.GetWorkflowNode)                            // 根据ID获取工作流节点
			nodes.POST("", workflowHandler.CreateWorkflowNode)                            // 创建工作流节点
			nodes.PUT("/:id", workflowHandler.UpdateWorkflowNode)                         // 更新工作流节点
			nodes.DELETE("/:id", workflowHandler.DeleteWorkflowNode)                      // 删除工作流节点
		}

		// Workflow Graph 操作路由
		graph := workflow.Group("/graph")
		{
			// 基本连接操作
			graph.POST("/connect", workflowHandler.ConnectNodes)       // 连接两个节点（next_node_id）
			graph.POST("/disconnect", workflowHandler.DisconnectNodes) // 断开节点连接

			// 分支连接操作
			graph.POST("/connect-branch", workflowHandler.ConnectBranch)       // 添加分支连接
			graph.POST("/disconnect-branch", workflowHandler.DisconnectBranch) // 删除分支连接

			// 并行节点操作
			graph.POST("/add-to-parallel", workflowHandler.AddNodeToParallel)           // 添加节点到并行执行
			graph.POST("/remove-from-parallel", workflowHandler.RemoveNodeFromParallel) // 从并行执行中移除节点
			graph.GET("/parallel-children", workflowHandler.GetParallelChildren)        // 获取并行节点的子节点

			// 节点信息查询
			graph.GET("/connections", workflowHandler.GetNodeConnections) // 获取节点的所有连接信息

			// 位置和批量操作
			graph.PUT("/position", workflowHandler.UpdateNodePosition)        // 更新节点位置
			graph.PUT("/positions", workflowHandler.BatchUpdateNodePositions) // 批量更新节点位置
			graph.POST("/batch-delete", workflowHandler.BatchDeleteNodes)     // 批量删除节点
		}
	}
}
