package main

import (
	"cyberedge/pkg/logging"
	"cyberedge/pkg/task"
	"cyberedge/pkg/utils"
)

func main() {
	// 连接MongoDB数据库
	client, err := utils.ConnectToMongoDB("mongodb://localhost:27017")
	if err != nil {
		logging.Error("连接MongoDB失败: %v", err)
		return
	}
	defer utils.DisconnectMongoDB(client)

	// 初始化数据库集合
	db := client.Database("cyberedgeDB")
	userCollection := db.Collection("users")

	// 确保集合存在
	if err := utils.EnsureCollectionExists(userCollection); err != nil {
		logging.Error("确保用户集合存在失败: %v", err)
		return
	}

	// 连接RabbitMQ
	rabbitConn, err := utils.ConnectToRabbitMQ("amqp://guest:guest@localhost:5672")
	if err != nil {
		logging.Error("连接RabbitMQ失败: %v", err)
		return
	}
	defer rabbitConn.Close()

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		logging.Error("创建RabbitMQ通道失败: %v", err)
		return
	}
	defer rabbitChannel.Close()

	// 设置集合
	utils.SetupCollections(db)

	// 初始化任务调度器
	scheduler := utils.SetupScheduler(rabbitConn, rabbitChannel, client, db.Collection("tasks"))

	// 设置全局变量
	utils.SetupGlobalVariables(scheduler, "your-jwt-secret")

	// 启动任务处理器
	go task.StartTaskProcessor(scheduler)

	// 设置API路由
	router := utils.SetupRouter()

	// 启动API服务器
	if err := router.Run(":8081"); err != nil {
		logging.Error("启动API服务器失败: %v", err)
	}
}
