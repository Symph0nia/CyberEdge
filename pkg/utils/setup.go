// utils/setup.go

package utils

import (
	"cyberedge/pkg/api"
	"cyberedge/pkg/api/handlers"
	"cyberedge/pkg/factory"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/service/task"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupCollections 设置各个集合的处理器
func SetupCollections(db *mongo.Database) error {
	if db == nil {
		return fmt.Errorf("数据库连接为空")
	}
	handlers.SetTOTPCollection(db.Collection("totp"))
	handlers.SetUserCollection(db.Collection("users"))
	handlers.SetConfigCollection(db.Collection("config"))
	handlers.SetTaskCollection(db.Collection("tasks"))
	return nil
}

// SetupScheduler 初始化任务调度器
func SetupScheduler(rabbitConn *amqp.Connection, rabbitChannel *amqp.Channel, client *mongo.Client, taskCollection *mongo.Collection) (*models.Scheduler, error) {
	if rabbitConn == nil || rabbitChannel == nil || client == nil || taskCollection == nil {
		return nil, fmt.Errorf("初始化调度器的参数不完整")
	}
	scheduler := factory.CreateScheduler(rabbitConn, rabbitChannel, "taskqueue", client, taskCollection)
	if scheduler == nil {
		return nil, fmt.Errorf("创建调度器失败")
	}
	return scheduler, nil
}

// SetupAndStartTaskProcessor 初始化并启动任务处理器
func SetupAndStartTaskProcessor(scheduler *models.Scheduler) error {
	taskService := task.NewTaskService(scheduler)
	taskProcessor := task.NewTaskProcessor(scheduler, taskService)

	// 启动任务处理器
	go taskProcessor.StartTaskProcessor()
	logging.Info("任务处理器启动")

	return nil
}

// SetupGlobalVariables 设置全局变量
func SetupGlobalVariables(scheduler *models.Scheduler, jwtSecret string) error {
	if scheduler == nil {
		return fmt.Errorf("调度器为空")
	}
	if jwtSecret == "" {
		return fmt.Errorf("JWT secret 为空")
	}
	api.GlobalScheduler = scheduler
	api.GlobalJWTSecret = jwtSecret
	return nil
}

// SetupRouter 设置API路由
func SetupRouter() (*gin.Engine, error) {
	router := api.SetupRouter()
	if router == nil {
		return nil, fmt.Errorf("设置路由失败")
	}
	return router, nil
}
