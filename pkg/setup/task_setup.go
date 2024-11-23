package setup

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
	taskDAO := dao.NewTaskDAO(db.Collection("tasks"))

	asynqClient, err := InitAsynqClient(redisAddr)
	if err != nil {
		logging.Error("初始化 Asynq 客户端失败: %v", err)
		return nil, nil, err
	}
	logging.Info("Asynq 客户端初始化成功")

	// 传入 redisAddr
	taskService := service.NewTaskService(taskDAO, asynqClient, redisAddr)

	asynqServer, err := InitAsynqServer(redisAddr)
	if err != nil {
		asynqClient.Close()
		logging.Error("初始化 Asynq 服务器失败: %v", err)
		return nil, nil, err
	}
	logging.Info("Asynq 服务器初始化成功")

	return taskService, asynqServer, nil
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
