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
