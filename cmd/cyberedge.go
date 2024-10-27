// CyberEdge/cmd/cyberedge.go

package main

import (
	"cyberedge/pkg/logging"
	"cyberedge/pkg/utils"
	"log"
	"path/filepath"
	"time"
)

func main() {
	// 初始化日志系统
	logPath := filepath.Join("logs", "cyberedge.log")
	if err := logging.InitializeLoggers(logPath); err != nil {
		// 使用标准库的 log 包，因为我们的日志系统可能还没有初始化
		log.Fatalf("初始化日志系统失败: %v", err)
	}
	logging.Info("日志系统初始化成功")

	// 启动日志轮换（每24小时轮换一次）
	logging.StartLogRotation(24 * time.Hour)
	defer logging.StopLogRotation()

	// 连接MongoDB数据库
	client, err := utils.ConnectToMongoDB("mongodb://localhost:27017")
	if err != nil {
		logging.Error("连接MongoDB失败: %v", err)
		return
	}
	defer func() {
		if err := utils.DisconnectMongoDB(client); err != nil {
			logging.Error("断开MongoDB连接失败: %v", err)
		}
	}()
	logging.Info("MongoDB连接成功")

	// 初始化数据库集合
	db := client.Database("cyberedgeDB")
	userCollection := db.Collection("users")

	// 确保集合存在
	if err := utils.EnsureCollectionExists(userCollection); err != nil {
		logging.Error("确保用户集合存在失败: %v", err)
		return
	}
	logging.Info("用户集合确认存在")

	// 连接RabbitMQ
	rabbitConn, err := utils.ConnectToRabbitMQ("amqp://guest:guest@localhost:5672")
	if err != nil {
		logging.Error("连接RabbitMQ失败: %v", err)
		return
	}
	defer rabbitConn.Close()
	logging.Info("RabbitMQ连接成功")

	rabbitChannel, err := utils.CreateRabbitMQChannel(rabbitConn)
	if err != nil {
		logging.Error("创建RabbitMQ通道失败: %v", err)
		return
	}
	defer rabbitChannel.Close()
	logging.Info("RabbitMQ通道创建成功")

	// 设置集合
	if err := utils.SetupCollections(db); err != nil {
		logging.Error("设置数据库集合失败: %v", err)
		return
	}
	logging.Info("数据库集合设置完成")

	// 设置全局变量
	if err := utils.SetupGlobalVariables("your-jwt-secret"); err != nil {
		logging.Error("设置全局变量失败: %v", err)
		return
	}
	logging.Info("全局变量设置完成")

	// 设置API路由
	router, err := utils.SetupRouter()
	if err != nil {
		logging.Error("设置API路由失败: %v", err)
		return
	}
	logging.Info("API路由设置完成")

	// 启动API服务器
	logging.Info("正在启动API服务器...")
	if err := router.Run(":8081"); err != nil {
		logging.Error("启动API服务器失败: %v", err)
	}
}
