// CyberEdge/pkg/api/router.go

package api

import (
	"cyberedge/pkg/api/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"time"
)

// 全局变量
var (
	GlobalJWTSecret string
)

// SetupRouter 设置API路由
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// 配置CORS中间件
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // 允许的前端域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // 允许跨域请求携带认证信息
		MaxAge:           12 * time.Hour,
	}))

	// 设置Session中间件
	store := cookie.NewStore([]byte("your-session-secret")) // 使用一个固定的session secret
	router.Use(sessions.Sessions("mysession", store))

	// 验证JWT的API
	router.GET("/auth/check", handlers.CheckAuth(GlobalJWTSecret))

	router.GET("/auth/qrcode", handlers.GenerateQRCode)
	router.POST("/auth/validate", handlers.ValidateTOTP(GlobalJWTSecret))

	// 使用中间件进行鉴权
	authenticated := router.Group("/")
	authenticated.Use(AuthMiddleware(GlobalJWTSecret))

	{
		// 控制二维码接口状态的API
		authenticated.GET("/auth/qrcode/status", handlers.SetQRCodeStatusHandler)
		authenticated.POST("/auth/qrcode/status", handlers.SetQRCodeStatusHandler)

		// 用户管理API
		authenticated.GET("/users", handlers.HandleUsers)
		authenticated.GET("/users/:account", handlers.HandleUsers)
		authenticated.POST("/users", handlers.HandleUsers)
		authenticated.DELETE("/users/:account", handlers.HandleUsers)
	}

	return router
}
