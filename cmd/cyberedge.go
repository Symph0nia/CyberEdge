// cmd/cyberedge.go

package main

import (
	"context"
	"cyberedge/pkg/api"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/service"
	"cyberedge/pkg/setup"
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

	// 连接MongoDB数据库
	client, err := setup.ConnectToMongoDB("mongodb://localhost:27017")
	if err != nil {
		logging.Error("连接MongoDB失败: %v", err)
		return
	}
	defer setup.DisconnectMongoDB(client)
	logging.Info("MongoDB连接成功")

	// 初始化数据库和集合
	db := client.Database("cyberedgeDB")

	// 初始化任务相关组件
	taskService, asynqServer, err := setup.InitTaskComponents(db, "localhost:6379")
	if err != nil {
		logging.Error("初始化任务组件失败: %v", err)
		return
	}
	defer taskService.Close()

	// 初始化 DAO
	taskDAO := dao.NewTaskDAO(db.Collection("tasks"))
	resultDAO := dao.NewResultDAO(db.Collection("results"))
	userDAO := dao.NewUserDAO(db.Collection("users"))
	configDAO := dao.NewConfigDAO(db.Collection("config"))
	targetDAO := dao.NewTargetDAO(db)

	// 初始化任务处理器
	taskHandler := setup.InitTaskHandler(taskDAO, targetDAO, resultDAO)

	// 启动 Asynq 服务器
	setup.StartAsynqServer(asynqServer, taskHandler)

	// 初始化 Service
	jwtSecret := "your-jwt-secret" // 应从配置文件或环境变量中读取
	userService := service.NewUserService(userDAO, configDAO, jwtSecret)
	configService := service.NewConfigService(configDAO)
	resultService := service.NewResultService(resultDAO)
	dnsService := service.NewDNSService(resultDAO)
	httpxService := service.NewHTTPXService(resultDAO)
	targetService := service.NewTargetService(targetDAO)

	// 设置API路由，包括任务管理的路由
	router := api.NewRouter(
		userService,
		configService,
		taskService,
		resultService,
		dnsService,
		httpxService,
		targetService,
		jwtSecret,
		"your-session-secret",             // 应从配置文件或环境变量中读取
		[]string{"http://localhost:8080"}, // 允许的源
	)
	engine := router.SetupRouter()
	logging.Info("API路由设置完成")

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:    ":8081",
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

	// 关闭 Asynq 服务器
	asynqServer.Shutdown()

	logging.Info("服务器已关闭")
}
