package utils

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/service"
	"cyberedge/pkg/tasks"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/mongo"
)

// InitTaskComponents 初始化任务相关组件
func InitTaskComponents(db *mongo.Database, redisAddr string) (*service.TaskService, *asynq.Server, error) {
	// 初始化 Task DAO
	taskDAO := dao.NewTaskDAO(db.Collection("tasks"))

	// 初始化 Asynq 客户端
	asynqClient, err := InitAsynqClient(redisAddr)
	if err != nil {
		logging.Error("初始化 Asynq 客户端失败: %v", err)
		return nil, nil, err
	}
	logging.Info("Asynq 客户端初始化成功")

	// 初始化 TaskService
	taskService := service.NewTaskService(taskDAO, asynqClient)

	// 初始化 Asynq 服务器
	asynqServer, err := InitAsynqServer(redisAddr)
	if err != nil {
		asynqClient.Close()
		logging.Error("初始化 Asynq 服务器失败: %v", err)
		return nil, nil, err
	}
	logging.Info("Asynq 服务器初始化成功")

	return taskService, asynqServer, nil
}

// InitTaskHandler 初始化任务处理器
func InitTaskHandler(taskDAO *dao.TaskDAO) *tasks.TaskHandler {
	taskHandler := tasks.NewTaskHandler()

	// 注册 Ping 任务处理函数
	pingTask := tasks.NewPingTask(taskDAO)
	taskHandler.RegisterHandler(tasks.TaskTypePing, pingTask.Handle)

	return taskHandler
}

// StartAsynqServer 启动 Asynq 服务器
func StartAsynqServer(server *asynq.Server, handler *tasks.TaskHandler) {
	go func() {
		if err := server.Run(handler); err != nil {
			logging.Error("运行 Asynq 服务器失败: %v", err)
		}
	}()
	logging.Info("Asynq 服务器已启动")
}
