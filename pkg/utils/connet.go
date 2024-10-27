// utils/connect.go

package utils

import (
	"context"
	"fmt"
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

// ConnectToRabbitMQ 连接到RabbitMQ
func ConnectToRabbitMQ(uri string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("连接RabbitMQ失败: %v", err)
	}
	return conn, nil
}

// CreateRabbitMQChannel 创建RabbitMQ通道
func CreateRabbitMQChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("创建RabbitMQ通道失败: %v", err)
	}
	return channel, nil
}
