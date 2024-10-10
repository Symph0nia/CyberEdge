package main

import (
	"context"
	"cyberedge/pkg/task"
	"fmt"
	"time"

	"cyberedge/pkg/api"
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/logging"     // 引入自定义日志组件
	"github.com/streadway/amqp" // 引入 RabbitMQ 客户端库
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, err := connectToMongoDB("mongodb://localhost:27017")
	if err != nil {
		logging.LogError(err)
		return
	}
	defer disconnectMongoDB(client)

	userCollection := client.Database("cyberedgeDB").Collection("users")
	totpCollection := client.Database("cyberedgeDB").Collection("totp")
	configCollection := client.Database("cyberedgeDB").Collection("config") // 新增配置集合
	taskCollection := client.Database("cyberedgeDB").Collection("tasks")    // 新增任务集合

	handlers.SetTOTPCollection(totpCollection)
	handlers.SetUserCollection(userCollection)
	handlers.SetConfigCollection(configCollection) // 设置配置集合
	handlers.SetTaskCollection(taskCollection)     // 设置任务集合

	if err := ensureCollectionExists(userCollection); err != nil {
		logging.LogError(err)
		return
	}

	// 连接到 RabbitMQ
	rabbitConn, err := connectToRabbitMQ("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logging.LogError(err)
		return
	}
	defer rabbitConn.Close()

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		logging.LogError(err)
		return
	}
	defer rabbitChannel.Close()

	scheduler, err := task.NewScheduler("amqp://guest:guest@localhost:5672/", "task_queue", client, "cyberedgeDB")
	if err != nil {
		logging.LogError(err)
		return
	}

	// 启动任务处理器
	go scheduler.StartTaskProcessor()

	router := api.SetupRouter(userCollection, client, "cyberedgeDB") // 传递 userCollection、client 和数据库名称
	if err := router.Run(":8081"); err != nil {
		logging.LogError(fmt.Errorf("启动API服务失败: %v", err))
	}
}

// connectToMongoDB 连接到MongoDB并返回客户端
func connectToMongoDB(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("无法连接到MongoDB: %v", err)
	}
	return client, nil
}

// disconnectMongoDB 断开MongoDB连接
func disconnectMongoDB(client *mongo.Client) {
	if err := client.Disconnect(context.Background()); err != nil {
		logging.LogError(fmt.Errorf("断开MongoDB连接失败: %v", err))
	}
}

// ensureCollectionExists 确保集合存在
func ensureCollectionExists(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := collection.InsertOne(ctx, bson.M{"placeholder": "value"}); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil // 集合已存在
		}
		return fmt.Errorf("插入文档失败: %v", err)
	}

	return nil
}

// connectToRabbitMQ 连接到RabbitMQ并返回连接实例
func connectToRabbitMQ(uri string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("无法连接到RabbitMQ: %v", err)
	}
	return conn, nil
}
