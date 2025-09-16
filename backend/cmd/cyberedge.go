// cmd/cyberedge.go

package main

import (
	"context"
	"cyberedge/pkg/api"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/service"
	"cyberedge/pkg/setup"
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	// 初始化日志系统
	logPath := filepath.Join("logs", "cyberedge.log")
	if err := logging.InitializeLoggers(logPath); err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}
	logging.Info("日志系统初始化成功")

	// 启动日志轮换（每24小时轮换一次）
	logging.StartLogRotation(24 * time.Hour)
	defer logging.StopLogRotation()

	// 创建一个用于优雅关闭的 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("加载 .env 文件失败: %v", err)
	}

	env := flag.String("env", "dev", "运行环境 (dev/prod)")
	flag.Parse()

	// 连接MySQL数据库
	db, err := setup.ConnectToMySQL("root:password@tcp(localhost:3306)/cyberedge?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		logging.Error("连接MySQL失败: %v", err)
		return
	}
	logging.Info("MySQL连接成功")

	// 初始化 DAO - 只保留用户相关
	userDAO := dao.NewUserDAO(db)

	// 初始化 Service - 只保留用户服务
	jwtSecret := os.Getenv("JWT_SECRET")

	userService := service.NewUserService(userDAO, jwtSecret)

	if *env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 根据环境设置 CORS 配置
	var allowedOrigins []string
	if *env == "prod" {
		// 生产环境必须指定具体域名，绝不能使用通配符
		prodOrigin := os.Getenv("ALLOWED_ORIGIN")
		if prodOrigin == "" {
			log.Fatal("生产环境必须设置 ALLOWED_ORIGIN 环境变量")
		}
		allowedOrigins = []string{prodOrigin}
	} else {
		// 开发环境允许更多源
		allowedOrigins = []string{
			"http://localhost:8080",
			"http://127.0.0.1:8080",
			"http://0.0.0.0:8080",
			"http://10.0.78.2:8080",      // 从启动脚本看到的网络地址
			"http://110.42.47.158:8080",  // 外部访问地址
		}
	}

	// 设置API路由 - 只保留用户相关
	router := api.NewRouter(
		userService,
		jwtSecret,
		allowedOrigins,
	)
	engine := router.SetupRouter()
	logging.Info("API路由设置完成")

	// 创建 HTTP 服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "31337"
	}
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: engine,
	}

	// 在后台启动 HTTP 服务器
	go func() {
		logging.Info("正在启动API服务器...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("启动API服务器失败: %v", err)
		}
	}()

	// 设置信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Info("正在关闭服务器...")

	// 创建一个5秒的超时上下文
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭 HTTP 服务器
	if err := srv.Shutdown(ctx); err != nil {
		logging.Error("服务器强制关闭: %v", err)
	}

	// 不再需要关闭Asynq服务器

	logging.Info("服务器已关闭")
}
