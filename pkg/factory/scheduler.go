package factory

import (
	"cyberedge/pkg/models"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateScheduler 创建一个新的调度器
func CreateScheduler(amqpConn *amqp.Connection, amqpChannel *amqp.Channel, queueName string, mongoClient *mongo.Client, taskCollection *mongo.Collection) *models.Scheduler {
	return &models.Scheduler{
		AMQPConn:       amqpConn,
		AMQPChannel:    amqpChannel,
		QueueName:      queueName,
		MongoClient:    mongoClient,
		TaskCollection: taskCollection,
	}
}
