// CyberEdge/cmd/cyberedge.go

package main

import (
	"context"
	"cyberedge/pkg/models"
	"cyberedge/pkg/task"
	"fmt"
	"time"

	"cyberedge/pkg/api"
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/logging"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// 连接MongoDB数据库
	client, err := connectToMongoDB("mongodb://localhost:27017")
	if err != nil {
		logging.Error("连接MongoDB失败: %v", err)
		return
	}
	defer disconnectMongoDB(client)

	// 初始化数据库集合
	db := client.Database("cyberedgeDB")
	userCollection := db.Collection("users")
	totpCollection := db.Collection("totp")
	configCollection := db.Collection("config")
	taskCollection := db.Collection("tasks")

	// 确保集合存在
	if err := ensureCollectionExists(userCollection); err != nil {
		logging.Error("确保用户集合存在失败: %v", err)
		return
	}

	// 设置API路由
	router := api.SetupRouter(userCollection, client, "cyberedgeDB")

	// 设置各个集合的处理器
	handlers.SetTOTPCollection(totpCollection)
	handlers.SetUserCollection(userCollection)
	handlers.SetConfigCollection(configCollection)
	handlers.SetTaskCollection(taskCollection)

	// 连接RabbitMQ
	rabbitConn, err := connectToRabbitMQ("amqp://guest:guest@localhost:5672")
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

	// 初始化任务调度器
	scheduler := models.NewScheduler(rabbitConn, rabbitChannel, "taskqueue", client, taskCollection)

	// 启动任务处理器
	go task.StartTaskProcessor(scheduler)

	// 启动API服务器
	if err := router.Run(":8081"); err != nil {
		logging.Error("启动API服务器失败: %v", err)
	}
}

// 连接到MongoDB
func connectToMongoDB(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("MongoDB连接失败: %v", err)
	}
	return client, nil
}

// 断开MongoDB连接
func disconnectMongoDB(client *mongo.Client) {
	if err := client.Disconnect(context.Background()); err != nil {
		logging.Error("MongoDB断开连接失败: %v", err)
	}
}

// 确保集合存在
func ensureCollectionExists(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 插入一个占位文档以确保集合存在
	_, err := collection.InsertOne(ctx, bson.M{"placeholder": "value"})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil // 集合已存在，忽略重复键错误
		}
		return fmt.Errorf("确保集合存在失败: %v", err)
	}
	return nil
}

// 连接到RabbitMQ
func connectToRabbitMQ(uri string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("RabbitMQ连接失败: %v", err)
	}
	return conn, nil
}
