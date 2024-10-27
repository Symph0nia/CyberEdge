// utils/connect.go

package utils

import (
	"context"
	"cyberedge/pkg/logging"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectToMongoDB 连接到MongoDB
func ConnectToMongoDB(uri string) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	return client, nil
}

// DisconnectMongoDB 断开MongoDB连接
func DisconnectMongoDB(client *mongo.Client) {
	if err := client.Disconnect(context.Background()); err != nil {
		logging.Error("MongoDB断开连接失败: %v", err)
	}
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
		return err
	}
	return nil
}

// ConnectToRabbitMQ 连接到RabbitMQ
func ConnectToRabbitMQ(uri string) (*amqp.Connection, error) {
	return amqp.Dial(uri)
}
