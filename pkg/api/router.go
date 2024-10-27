// pkg/api/router.go

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
	jwtSecret      string
	sessionSecret  string
	allowedOrigins []string
}

func NewRouter(
	userService *service.UserService,
	configService *service.ConfigService,
	jwtSecret string,
	sessionSecret string,
	allowedOrigins []string,
) *Router {
	return &Router{
		userHandler:    handlers.NewUserHandler(userService),
		configHandler:  handlers.NewConfigHandler(configService),
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
		authenticated.GET("/auth/qrcode/status", r.configHandler.GetQRCodeStatus)
		authenticated.POST("/auth/qrcode/status", r.configHandler.SetQRCodeStatus)

		// 用户管理API
		authenticated.GET("/users", r.userHandler.GetUsers)
		authenticated.GET("/users/:account", r.userHandler.GetUser)
		authenticated.POST("/users", r.userHandler.CreateUser)
		authenticated.DELETE("/users/:account", r.userHandler.DeleteUser)
	}

	return router
}
