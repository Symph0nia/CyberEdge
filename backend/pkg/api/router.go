package api

import (
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/middleware"
	"cyberedge/pkg/service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

type Router struct {
	userHandler    *handlers.UserHandler
	jwtSecret      string
	allowedOrigins []string
}

func NewRouter(
	userService *service.UserService,
	jwtSecret string,
	allowedOrigins []string,
) *Router {
	return &Router{
		userHandler:    handlers.NewUserHandler(userService),
		jwtSecret:      jwtSecret,
		allowedOrigins: allowedOrigins,
	}
}

// SetupRouter 设置并返回 gin.Engine - 只保留用户管理功能
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


	// 公开的认证API
	router.GET("/auth/check", r.userHandler.CheckAuth)
	router.GET("/auth/qrcode", r.userHandler.GenerateQRCode)
	router.POST("/auth/validate", r.userHandler.ValidateTOTP)
	router.POST("/auth/login", r.userHandler.Login)
	router.POST("/auth/register", r.userHandler.Register)

	// 使用中间件进行鉴权
	authenticated := router.Group("/")
	authenticated.Use(middleware.AuthMiddleware(r.jwtSecret))
	{
		// 用户管理API
		authenticated.GET("/users", r.userHandler.GetUsers)
		authenticated.GET("/users/:id", r.userHandler.GetUser)
		authenticated.POST("/users", r.userHandler.CreateUser)
		authenticated.DELETE("/users/:id", r.userHandler.DeleteUser)

		// 2FA管理
		authenticated.POST("/auth/2fa/setup", r.userHandler.Setup2FA)
		authenticated.POST("/auth/2fa/verify", r.userHandler.Verify2FA)
		authenticated.DELETE("/auth/2fa", r.userHandler.Disable2FA)
	}

	return router
}