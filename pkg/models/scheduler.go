// CyberEdge/models/scheduler.go

package models

import (
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

// Scheduler 任务调度器结构
type Scheduler struct {
	AMQPConn       *amqp.Connection
	AMQPChannel    *amqp.Channel
	QueueName      string
	MongoClient    *mongo.Client
	TaskCollection *mongo.Collection
}

// NewScheduler 创建新的任务调度器
func NewScheduler(amqpConn *amqp.Connection, amqpChannel *amqp.Channel, queueName string, mongoClient *mongo.Client, taskCollection *mongo.Collection) *Scheduler {
	return &Scheduler{
		AMQPConn:       amqpConn,
		AMQPChannel:    amqpChannel,
		QueueName:      queueName,
		MongoClient:    mongoClient,
		TaskCollection: taskCollection,
	}
}
