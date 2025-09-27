package handlers

import (
	"strconv"

	"go-backend/database/ent"
	"go-backend/database/ent/loginrecord"
	"go-backend/internal/funcs"
	"go-backend/internal/middleware"
	"go-backend/pkg/database"
	"go-backend/shared/models"

	"github.com/gin-gonic/gin"
)

// LoginRecordHandler 登录记录处理器
type LoginRecordHandler struct {
}

// NewLoginRecordHandler 创建新的登录记录处理器
func NewLoginRecordHandler() *LoginRecordHandler {
	return &LoginRecordHandler{}
}

// GetLoginRecords 获取登录记录列表
// @Summary      获取登录记录列表
// @Description  获取系统登录记录列表（管理员权限）
// @Tags         admin
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page     query    int    false "页码" default(1)
// @Param        limit    query    int    false "每页数量" default(20)
// @Param        user_id  query    string false "用户ID"
// @Param        status   query    string false "登录状态" Enums(success,failed,locked)
// @Success      200 {object} object{success=bool,data=object{records=[]models.LoginRecordResponse,total=int}}
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      403 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /admin/login-records [get]
func (h *LoginRecordHandler) GetLoginRecords(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")
	userIDStr := c.Query("user_id")
	status := c.Query("status")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var records []*ent.LoginRecord
	var total int

	if userIDStr != "" {
		// 查询指定用户的登录记录
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			middleware.ThrowError(c, middleware.ValidationError("用户ID格式错误", ""))
			return
		}

		records, err = funcs.LoginRecordFuncs{}.GetUserLoginRecords(middleware.GetRequestContext(c), userID, limit, offset)
		if err != nil {
			middleware.ThrowError(c, middleware.InternalServerError("查询用户登录记录失败", err.Error()))
			return
		}

		// 获取总数
		totalCount, err := database.Client.LoginRecord.Query().
			Where(loginrecord.UserIDEQ(userID)).
			Count(middleware.GetRequestContext(c))
		if err != nil {
			middleware.ThrowError(c, middleware.InternalServerError("查询记录总数失败", err.Error()))
			return
		}
		total = totalCount

	} else if status != "" {
		// 根据状态查询登录记录
		records, err = funcs.LoginRecordFuncs{}.GetLoginRecordsByStatus(middleware.GetRequestContext(c), status, limit, offset)
		if err != nil {
			middleware.ThrowError(c, middleware.InternalServerError("查询登录记录失败", err.Error()))
			return
		}

		// 获取总数
		totalCount, err := database.Client.LoginRecord.Query().
			Where(loginrecord.StatusEQ(loginrecord.Status(status))).
			Count(middleware.GetRequestContext(c))
		if err != nil {
			middleware.ThrowError(c, middleware.InternalServerError("查询记录总数失败", err.Error()))
			return
		}
		total = totalCount

	} else {
		// 查询所有登录记录
		records, err = database.Client.LoginRecord.Query().
			Order(ent.Desc(loginrecord.FieldCreateTime)).
			Limit(limit).
			Offset(offset).
			All(middleware.GetRequestContext(c))
		if err != nil {
			middleware.ThrowError(c, middleware.InternalServerError("查询登录记录失败", err.Error()))
			return
		}

		// 获取总数
		totalCount, err := database.Client.LoginRecord.Query().
			Count(middleware.GetRequestContext(c))
		if err != nil {
			middleware.ThrowError(c, middleware.InternalServerError("查询记录总数失败", err.Error()))
			return
		}
		total = totalCount
	}

	// 转换为响应格式
	recordResponses := make([]*models.LoginRecordResponse, len(records))
	for i, record := range records {
		recordResponses[i] = convertLoginRecordToResponse(record)
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"records": recordResponses,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// GetUserLoginRecords 获取当前用户的登录记录
// @Summary      获取当前用户的登录记录
// @Description  获取当前登录用户的登录记录
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page  query    int false "页码" default(1)
// @Param        limit query    int false "每页数量" default(10)
// @Success      200 {object} object{success=bool,data=object{records=[]models.LoginRecordResponse,total=int}}
// @Failure      400 {object} object{success=bool,message=string}
// @Failure      401 {object} object{success=bool,message=string}
// @Failure      500 {object} object{success=bool,message=string}
// @Router       /auth/login-records [get]
func (h *LoginRecordHandler) GetUserLoginRecords(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		middleware.ThrowError(c, middleware.UnauthorizedError("未找到用户信息", ""))
		return
	}

	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	offset := (page - 1) * limit

	// 查询用户的登录记录
	records, err := funcs.LoginRecordFuncs{}.GetUserLoginRecords(middleware.GetRequestContext(c), userID, limit, offset)
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("查询登录记录失败", err.Error()))
		return
	}

	// 获取总数
	total, err := database.Client.LoginRecord.Query().
		Where(loginrecord.UserIDEQ(userID)).
		Count(middleware.GetRequestContext(c))
	if err != nil {
		middleware.ThrowError(c, middleware.InternalServerError("查询记录总数失败", err.Error()))
		return
	}

	// 转换为响应格式
	recordResponses := make([]*models.LoginRecordResponse, len(records))
	for i, record := range records {
		recordResponses[i] = convertLoginRecordToResponse(record)
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"records": recordResponses,
			"total":   total,
			"page":    page,
			"limit":   limit,
		},
	})
}

// convertLoginRecordToResponse 转换登录记录为响应格式
func convertLoginRecordToResponse(record *ent.LoginRecord) *models.LoginRecordResponse {
	response := &models.LoginRecordResponse{
		ID:             record.ID,
		UserID:         record.UserID,
		Identifier:     record.Identifier,
		CredentialType: string(record.CredentialType),
		IPAddress:      record.IPAddress,
		Status:         string(record.Status),
		LoginTime:      record.CreateTime.Format("2006-01-02 15:04:05"),
	}

	if record.UserAgent != "" {
		response.UserAgent = record.UserAgent
	}
	if record.DeviceInfo != "" {
		response.DeviceInfo = record.DeviceInfo
	}
	if record.Location != "" {
		response.Location = record.Location
	}
	if record.FailureReason != "" {
		response.FailureReason = record.FailureReason
	}
	if record.SessionID != "" {
		response.SessionID = record.SessionID
	}
	if record.LogoutTime != nil {
		response.LogoutTime = record.LogoutTime.Format("2006-01-02 15:04:05")
	}
	if record.Duration != 0 {
		response.Duration = record.Duration
	}
	if record.Metadata != nil {
		response.Metadata = record.Metadata
	}

	return response
}
