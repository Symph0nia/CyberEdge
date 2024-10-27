// utils/connect.go

package utils

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"

	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectToMongoDB 连接到MongoDB
func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("连接MongoDB失败: %v", err)
	}
	return client, nil
}

// DisconnectMongoDB 断开MongoDB连接
func DisconnectMongoDB(client *mongo.Client) error {
	if err := client.Disconnect(context.Background()); err != nil {
		return fmt.Errorf("MongoDB断开连接失败: %v", err)
	}
	return nil
}

// EnsureCollectionExists 确保集合存在
func EnsureCollectionExists(collection *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, bson.M{"placeholder": "value"})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil // 集合已存在，忽略重复键错误
		}
		return fmt.Errorf("确保集合存在失败: %v", err)
	}
	return nil
}

// ConnectToRedis 连接到Redis
func ConnectToRedis(addr string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("连接Redis失败: %v", err)
	}

	return client, nil
}

// InitAsynqClient 初始化Asynq客户端
func InitAsynqClient(redisAddr string) (*asynq.Client, error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return client, nil
}

// InitAsynqServer 初始化Asynq服务器
func InitAsynqServer(redisAddr string) (*asynq.Server, error) {
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
	return server, nil
}
