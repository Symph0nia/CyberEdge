package setup

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/tasks"
)

// InitTaskHandler 初始化任务处理器
func InitTaskHandler(taskDAO *dao.TaskDAO, targetDAO *dao.TargetDAO, resultDAO *dao.ResultDAO) *tasks.TaskHandler {
	taskHandler := tasks.NewTaskHandler()

	subfinderTask := tasks.NewSubfinderTask(taskDAO, targetDAO, resultDAO)
	nmapTask := tasks.NewNmapTask(taskDAO, targetDAO, resultDAO)
	FfufTask := tasks.NewFfufTask(taskDAO, targetDAO, resultDAO)

	taskHandler.RegisterHandler(tasks.TaskTypeSubfinder, subfinderTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeNmap, nmapTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeFfuf, FfufTask.Handle)

	return taskHandler
}
