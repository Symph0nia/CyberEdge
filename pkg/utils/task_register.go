package utils

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/tasks"
)

// InitTaskHandler 初始化任务处理器
func InitTaskHandler(taskDAO *dao.TaskDAO) *tasks.TaskHandler {
	taskHandler := tasks.NewTaskHandler()

	// 注册 Ping 任务处理函数
	pingTask := tasks.NewPingTask(taskDAO)
	httpxTask := tasks.NewHttpxTask(taskDAO)
	taskHandler.RegisterHandler(tasks.TaskTypePing, pingTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeHttpx, httpxTask.Handle)

	return taskHandler
}
