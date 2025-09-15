package setup

import (
	"context"
	"cyberedge/pkg/logging"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

// ConnectToMongoDB 连接到MongoDB
func ConnectToMongoDB(defaultURI string) (*mongo.Client, error) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = defaultURI
	}
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

// 删除了所有Asynq和Redis相关代码
// 现在使用简单的goroutine替代复杂的任务队列系统