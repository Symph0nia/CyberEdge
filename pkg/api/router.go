package api

import (
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/task"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// SetupRouter 设置API路由
func SetupRouter(userCollection *mongo.Collection, mongoClient *mongo.Client, dbName string) *gin.Engine {
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

	// 从数据库加载配置
	var configResult struct {
		Secrets struct {
			SessionSecret string `bson:"sessionSecret"`
			JWTSecret     string `bson:"jwtSecret"`
		} `bson:"secrets"`
	}

	// 设置Session中间件
	store := cookie.NewStore([]byte(configResult.Secrets.SessionSecret))
	router.Use(sessions.Sessions("mysession", store))

	// 验证JWT的API
	router.GET("/auth/check", handlers.CheckAuth(configResult.Secrets.JWTSecret))

	router.GET("/auth/qrcode", handlers.GenerateQRCode)
	router.POST("/auth/validate", handlers.ValidateTOTP(configResult.Secrets.JWTSecret))

	// 使用中间件进行鉴权
	authenticated := router.Group("/")
	authenticated.Use(AuthMiddleware(configResult.Secrets.JWTSecret))

	{
		// 控制二维码接口状态的API
		authenticated.GET("/auth/qrcode/status", handlers.SetQRCodeStatusHandler)  // 查询二维码状态
		authenticated.POST("/auth/qrcode/status", handlers.SetQRCodeStatusHandler) // 更新二维码状态

		// 用户管理API
		authenticated.GET("/users", handlers.HandleUsers)
		authenticated.GET("/users/:account", handlers.HandleUsers)
		authenticated.POST("/users", handlers.HandleUsers)
		authenticated.DELETE("/users/:account", handlers.HandleUsers)

		// 创建任务调度器
		scheduler, err := task.NewScheduler("amqp://guest:guest@localhost:5672/", "task_queue", mongoClient, dbName)
		if err != nil {
			panic(err) // 或者适当处理错误
		}
		taskHandler := handlers.NewTaskHandler(scheduler)

		// 任务管理API
		authenticated.GET("/tasks", taskHandler.GetAllTasks) // 获取所有任务
		authenticated.POST("/tasks", taskHandler.CreateTask)
		authenticated.GET("/tasks/:id", taskHandler.GetTask)          // 获取单个任务
		authenticated.POST("/tasks/:id/start", taskHandler.StartTask) // 开始单个任务
		authenticated.POST("/tasks/:id/stop", taskHandler.StopTask)   // 停止单个任务
		authenticated.DELETE("/tasks/:id", taskHandler.DeleteTask)    // 删除单个任务
	}

	return router
}
