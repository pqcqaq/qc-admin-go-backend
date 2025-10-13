package middleware

import (
	"context"
	"go-backend/database/ent"
	"go-backend/database/ent/apiauth"
	"go-backend/pkg/database"
	"sync"

	"github.com/gin-gonic/gin"
)

type ApiAuthKey string

const (
	ApiAuthRecord ApiAuthKey = "api_auth_record"
)

// APIAuthMiddleware API认证中间件
func APIAuthMiddleware(engine *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		requestPath := c.Request.URL.Path

		// 首先查找匹配的路由模式
		routePattern := findMatchingRoutePattern(engine, method, requestPath)
		if routePattern == "" {
			// 路由不存在，直接返回404
			ThrowError(c, NotFoundError("请求的API路由不存在", nil))
			c.Abort()
			return
		}

		// 使用路由模式查询数据库中的API认证记录
		record, err := QueryAPIAuthRecord(method, routePattern)
		if err != nil {
			ThrowError(c, InternalServerError("查询API认证记录失败", err.Error()))
			c.Abort()
			return
		}

		if record == nil {
			// 路由存在但认证记录不存在，创建默认的公开记录
			newRecord, err := CreatePublicAPIAuthRecord(method, routePattern)
			if err != nil {
				ThrowError(c, InternalServerError("创建默认公开API认证记录失败", err.Error()))
				c.Abort()
				return
			}
			c.Set(string(ApiAuthRecord), newRecord)
			c.Next()
			return
		}

		if record.IsPublic {
			c.Set(string(ApiAuthRecord), record)
			c.Next()
			return
		}

		c.Set(string(ApiAuthRecord), record)
		c.Next()
	}
}

// findMatchingRoutePattern 查找匹配的路由模式
func findMatchingRoutePattern(engine *gin.Engine, method, requestPath string) string {
	routes := engine.Routes()

	// 首先尝试精确匹配（处理静态路由）
	for _, route := range routes {
		if route.Method == method && route.Path == requestPath {
			return route.Path
		}
	}

	// 如果没有精确匹配，则创建一个临时的gin context来测试路由匹配
	// 这是通过gin内部路由树来找到正确的路由模式
	for _, route := range routes {
		if route.Method == method {
			// 使用gin的路由匹配逻辑
			if isPathMatch(route.Path, requestPath) {
				return route.Path
			}
		}
	}

	return "" // 没有找到匹配的路由
}

// isPathMatch 检查请求路径是否匹配路由模式
// 这个函数实现了简单的路由匹配逻辑，处理 :param 和 *wildcard
func isPathMatch(routePattern, requestPath string) bool {
	// 分割路径段
	routeSegments := splitPath(routePattern)
	requestSegments := splitPath(requestPath)

	// 如果段数不同且路由模式没有通配符，则不匹配
	if len(routeSegments) != len(requestSegments) {
		// 检查是否有通配符
		hasWildcard := false
		for _, segment := range routeSegments {
			if len(segment) > 0 && segment[0] == '*' {
				hasWildcard = true
				break
			}
		}
		if !hasWildcard {
			return false
		}
	}

	// 逐段比较
	for i, routeSegment := range routeSegments {
		if i >= len(requestSegments) {
			return false
		}

		// 通配符匹配剩余所有段
		if len(routeSegment) > 0 && routeSegment[0] == '*' {
			return true
		}

		// 参数匹配任意段
		if len(routeSegment) > 0 && routeSegment[0] == ':' {
			continue
		}

		// 精确匹配
		if routeSegment != requestSegments[i] {
			return false
		}
	}

	return true
}

// splitPath 分割路径为段
func splitPath(path string) []string {
	if path == "/" || path == "" {
		return []string{}
	}

	// 移除开头的斜杠并分割
	if path[0] == '/' {
		path = path[1:]
	}

	if path == "" {
		return []string{}
	}

	segments := make([]string, 0)
	current := ""

	for _, char := range path {
		if char == '/' {
			if current != "" {
				segments = append(segments, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		segments = append(segments, current)
	}

	return segments
}

// 全局锁，用于控制API认证记录创建的并发
var apiAuthCreateMutex sync.Mutex

// 改进的创建方法，支持幂等性和线程安全
func CreatePublicAPIAuthRecord(method, routePattern string) (*APIAuthRecord, error) {
	// 使用全局锁确保同一时间只有一个goroutine能创建记录
	apiAuthCreateMutex.Lock()
	defer apiAuthCreateMutex.Unlock()

	ctx := context.Background()

	// 先尝试再次查询，避免并发创建
	existingRecord, err := database.Client.APIAuth.
		Query().
		Where(
			apiauth.MethodEQ(method),
			apiauth.PathEQ(routePattern), // 使用路由模式而不是请求路径
		).
		Only(ctx)

	if err == nil {
		// 记录已存在，返回现有记录
		return &APIAuthRecord{
			IsPublic:    existingRecord.IsPublic,
			Permissions: make([]string, 0),
		}, nil
	}

	if !ent.IsNotFound(err) {
		return nil, err // 查询出错
	}

	// 记录不存在，创建新记录
	_, err = database.Client.APIAuth.
		Create().
		SetName(method + " " + routePattern). // 使用路由模式作为名称
		SetDescription("Automatically created public API record").
		SetMethod(method).
		SetPath(routePattern). // 存储路由模式
		SetIsPublic(true).
		SetIsActive(true).
		SetMetadata(map[string]interface{}{
			"auto_created":  true,
			"route_pattern": routePattern, // 额外存储路由模式信息
		}).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	return &APIAuthRecord{
		IsPublic:    true,
		Permissions: make([]string, 0),
	}, nil
}

// QueryAPIAuthRecord 使用entgo查询API认证记录
func QueryAPIAuthRecord(method, routePattern string) (*APIAuthRecord, error) {
	ctx := context.Background()
	record, err := database.Client.APIAuth.
		Query().
		Where(
			apiauth.MethodEQ(method),
			apiauth.PathEQ(routePattern), // 使用路由模式查询
			apiauth.TypeEQ(apiauth.TypeHTTP),
		).
		WithPermissions().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil // 未找到记录
		}
		return nil, err // 查询出错
	}

	var permissionsList []string
	for _, perm := range record.Edges.Permissions {
		permissionsList = append(permissionsList, perm.Action)
	}

	return &APIAuthRecord{
		IsPublic:    record.IsPublic,
		Permissions: permissionsList,
	}, nil
}

// APIAuthRecord 模拟的API认证记录结构体
type APIAuthRecord struct {
	IsPublic    bool
	Permissions []string
}
