package setup

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/tasks"
)

// InitTaskHandler 初始化任务处理器
func InitTaskHandler(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *tasks.TaskHandler {
	taskHandler := tasks.NewTaskHandler()

	// 注册 Ping 任务处理函数
	pingTask := tasks.NewPingTask(taskDAO)
	httpxTask := tasks.NewHttpxTask(taskDAO)
	subfinderTask := tasks.NewSubfinderTask(taskDAO, resultDAO) // 传入 resultDAO
	nmapTask := tasks.NewNmapTask(taskDAO, resultDAO)
	FfufTask := tasks.NewFfufTask(taskDAO, resultDAO)

	taskHandler.RegisterHandler(tasks.TaskTypePing, pingTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeHttpx, httpxTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeSubfinder, subfinderTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeNmap, nmapTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeFfuf, FfufTask.Handle)

	return taskHandler
}
