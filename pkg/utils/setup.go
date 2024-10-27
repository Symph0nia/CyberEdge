// utils/setup.go
package utils

import (
	"cyberedge/pkg/api"
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/models"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupCollections 设置各个集合的处理器
func SetupCollections(db *mongo.Database) {
	handlers.SetTOTPCollection(db.Collection("totp"))
	handlers.SetUserCollection(db.Collection("users"))
	handlers.SetConfigCollection(db.Collection("config"))
	handlers.SetTaskCollection(db.Collection("tasks"))
}

// SetupScheduler 初始化任务调度器
func SetupScheduler(rabbitConn *amqp.Connection, rabbitChannel *amqp.Channel, client *mongo.Client, taskCollection *mongo.Collection) *models.Scheduler {
	return models.NewScheduler(rabbitConn, rabbitChannel, "taskqueue", client, taskCollection)
}

// SetupGlobalVariables 设置全局变量
func SetupGlobalVariables(scheduler *models.Scheduler, jwtSecret string) {
	api.GlobalScheduler = scheduler
	api.GlobalJWTSecret = jwtSecret
}

// SetupRouter 设置API路由
func SetupRouter() *gin.Engine {
	return api.SetupRouter()
}
