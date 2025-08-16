package middleware

// 使用示例：
//
// 在处理器中使用自定义错误：
//
// 1. 使用c.Error()方式（推荐）：
//    func (h *UserHandler) GetUser(c *gin.Context) {
//        id := c.Param("id")
//        if id == "" {
//            middleware.ThrowError(c, middleware.BadRequestError("用户ID不能为空", nil))
//            return
//        }
//
//        user, err := h.userService.GetUser(id)
//        if err != nil {
//            middleware.ThrowError(c, middleware.UserNotFoundError(map[string]any{"id": id}))
//            return
//        }
//
//        c.JSON(200, gin.H{"success": true, "data": user})
//    }
//
// 2. 使用panic方式：
//    func (h *UserHandler) CreateUser(c *gin.Context) {
//        var user User
//        if err := c.ShouldBindJSON(&user); err != nil {
//            middleware.PanicWithError(middleware.ValidationError("请求数据格式错误", err.Error()))
//        }
//
//        // ... 业务逻辑
//    }
//
// 在服务层中抛出错误：
//    func (s *UserService) GetUser(id string) (*User, error) {
//        user, err := s.repo.FindByID(id)
//        if err != nil {
//            // 如果是数据库错误
//            middleware.PanicWithError(middleware.DatabaseError("查询用户失败", err.Error()))
//        }
//        if user == nil {
//            return nil, middleware.UserNotFoundError(map[string]any{"id": id})
//        }
//        return user, nil
//    }
//
// 在路由中注册中间件：
//    func (r *Router) SetupRoutes(engine *gin.Engine) {
//        // 注册错误处理中间件
//        engine.Use(middleware.ErrorHandlerMiddleware()) // 处理panic
//        engine.Use(middleware.ErrorHandler())           // 处理gin.Error
//
//        // ... 其他路由设置
//    }
