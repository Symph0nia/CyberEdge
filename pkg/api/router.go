package api

import (
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/middleware"
	"cyberedge/pkg/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"time"
)

type Router struct {
	userHandler    *handlers.UserHandler
	configHandler  *handlers.ConfigHandler
	taskHandler    *handlers.TaskHandler
	resultHandler  *handlers.ResultHandler
	jwtSecret      string
	sessionSecret  string
	allowedOrigins []string
}

func NewRouter(
	userService *service.UserService,
	configService *service.ConfigService,
	taskService *service.TaskService,
	resultService *service.ResultService,
	dnsService *service.DNSService,
	jwtSecret string,
	sessionSecret string,
	allowedOrigins []string,
) *Router {
	return &Router{
		userHandler:    handlers.NewUserHandler(userService),
		configHandler:  handlers.NewConfigHandler(configService),
		taskHandler:    handlers.NewTaskHandler(taskService),
		resultHandler:  handlers.NewResultHandler(resultService, dnsService),
		jwtSecret:      jwtSecret,
		sessionSecret:  sessionSecret,
		allowedOrigins: allowedOrigins,
	}
}

// SetupRouter 设置并返回 gin.Engine
func (r *Router) SetupRouter() *gin.Engine {
	router := gin.Default()

	// 配置CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     r.allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 设置Session中间件
	store := cookie.NewStore([]byte(r.sessionSecret))
	router.Use(sessions.Sessions("mysession", store))

	// 验证JWT的API
	router.GET("/auth/check", r.userHandler.CheckAuth)
	router.GET("/auth/qrcode", r.userHandler.GenerateQRCode)
	router.POST("/auth/validate", r.userHandler.ValidateTOTP)

	// 使用中间件进行鉴权
	authenticated := router.Group("/")
	authenticated.Use(middleware.AuthMiddleware(r.jwtSecret))

	{
		// 控制二维码接口状态的API
		authenticated.GET("/auth/qrcode/status", r.configHandler.GetQRCodeStatus)  // 获取二维码状态
		authenticated.POST("/auth/qrcode/status", r.configHandler.SetQRCodeStatus) // 设置二维码状态

		// 系统配置相关API
		authenticated.GET("/system/info", r.configHandler.GetSystemInfo)   // 获取系统信息
		authenticated.GET("/system/tools", r.configHandler.GetToolsStatus) // 获取工具安装状态

		// 用户管理API
		authenticated.GET("/users", r.userHandler.GetUsers)               // 获取所有用户
		authenticated.GET("/users/:account", r.userHandler.GetUser)       // 获取单个用户
		authenticated.POST("/users", r.userHandler.CreateUser)            // 创建新用户
		authenticated.DELETE("/users/:account", r.userHandler.DeleteUser) // 删除用户

		// 任务管理API
		authenticated.POST("/tasks", r.taskHandler.CreateTask)          // 创建任务
		authenticated.GET("/tasks", r.taskHandler.GetAllTasks)          // 获取所有任务
		authenticated.DELETE("/tasks/:id", r.taskHandler.DeleteTask)    // 删除任务
		authenticated.POST("/tasks/:id/start", r.taskHandler.StartTask) // 启动单个任务

		// 扫描结果管理API
		authenticated.GET("/results/:id", r.resultHandler.GetResultByID)                          // 获取单个扫描结果
		authenticated.GET("/results/type/:type", r.resultHandler.GetResultsByType)                // 获取指定类型的扫描结果
		authenticated.PUT("/results/:id", r.resultHandler.UpdateResult)                           // 更新扫描结果
		authenticated.DELETE("/results/:id", r.resultHandler.DeleteResult)                        // 删除扫描结果
		authenticated.PUT("/results/:id/read", r.resultHandler.MarkResultAsRead)                  // 根据任务 ID 修改任务的已读状态
		authenticated.PUT("/results/:id/entries/:entry_id/read", r.resultHandler.MarkEntryAsRead) // 根据任务 ID 和条目 ID 修改条目的已读状态
		authenticated.PUT("/results/:id/entries/:entry_id/resolve", r.resultHandler.ResolveSubdomainIPHandler)
	}

	return router
}
