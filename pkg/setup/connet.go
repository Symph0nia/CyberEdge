package setup

import (
	"context"
	"cyberedge/pkg/logging"
	"fmt"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectToMongoDB 连接到MongoDB
func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	logging.Info("尝试连接到 MongoDB: %s", uri)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		logging.Error("连接 MongoDB 失败: %v", err)
		return nil, fmt.Errorf("连接 MongoDB 失败: %v", err)
	}
	logging.Info("成功连接到 MongoDB")
	return client, nil
}

// DisconnectMongoDB 断开MongoDB连接
func DisconnectMongoDB(client *mongo.Client) error {
	logging.Info("尝试断开 MongoDB 连接")
	if err := client.Disconnect(context.Background()); err != nil {
		logging.Error("MongoDB 断开连接失败: %v", err)
		return fmt.Errorf("MongoDB 断开连接失败: %v", err)
	}
	logging.Info("成功断开 MongoDB 连接")
	return nil
}

// InitAsynqClient 初始化Asynq客户端
func InitAsynqClient(redisAddr string) (*asynq.Client, error) {
	logging.Info("初始化 Asynq 客户端，Redis 地址: %s", redisAddr)
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	logging.Info("成功初始化 Asynq 客户端")
	return client, nil
}

// InitAsynqServer 初始化Asynq服务器
func InitAsynqServer(redisAddr string) (*asynq.Server, error) {
	logging.Info("初始化 Asynq 服务器，Redis 地址: %s", redisAddr)
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				"default":  5,
				"critical": 10,
			},
		},
	)
	logging.Info("成功初始化 Asynq 服务器")
	return server, nil
}
