package setup

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/tasks"
)

// InitTaskHandler 初始化任务处理器
func InitTaskHandler(taskDAO *dao.TaskDAO, targetDAO *dao.TargetDAO, resultDAO *dao.ResultDAO, configDAO *dao.ConfigDAO) *tasks.TaskHandler {
	taskHandler := tasks.NewTaskHandler()

	subfinderTask := tasks.NewSubfinderTask(taskDAO, targetDAO, resultDAO, configDAO)
	nmapTask := tasks.NewNmapTask(taskDAO, targetDAO, resultDAO, configDAO)
	FfufTask := tasks.NewFfufTask(taskDAO, targetDAO, resultDAO, configDAO)

	taskHandler.RegisterHandler(tasks.TaskTypeSubfinder, subfinderTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeNmap, nmapTask.Handle)
	taskHandler.RegisterHandler(tasks.TaskTypeFfuf, FfufTask.Handle)

	return taskHandler
}
