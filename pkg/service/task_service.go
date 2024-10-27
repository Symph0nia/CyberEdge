package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/tasks"
	"github.com/hibiken/asynq"
)

type TaskService struct {
	taskDAO     *dao.TaskDAO
	asynqClient *asynq.Client
}

// NewTaskService 创建一个新的 TaskService 实例
func NewTaskService(taskDAO *dao.TaskDAO, asynqClient *asynq.Client) *TaskService {
	return &TaskService{
		taskDAO:     taskDAO,
		asynqClient: asynqClient,
	}
}

// CreatePingTask 创建一个新的 Ping 任务并将其添加到队列中
func (s *TaskService) CreatePingTask(target string) error {
	logging.Info("正在创建 Ping 任务: %s", target)

	// 创建任务对象
	task := &models.Task{
		Type:   tasks.TaskTypePing,
		Target: target,
		Status: models.TaskStatusPending,
	}

	// 将任务保存到数据库
	if err := s.taskDAO.CreateTask(task); err != nil {
		logging.Error("创建任务失败: %v", err)
		return err
	}

	// 创建 Asynq 任务
	asynqTask, err := tasks.NewPingTask(target)
	if err != nil {
		logging.Error("创建 Asynq 任务失败: %v", err)
		return err
	}

	// 将任务加入队列
	_, err = s.asynqClient.Enqueue(asynqTask)
	if err != nil {
		logging.Error("将任务加入队列失败: %v", err)
		return err
	}

	logging.Info("成功创建并加入队列 Ping 任务: %s", target)
	return nil
}

// GetAllTasks 获取所有任务
func (s *TaskService) GetAllTasks() ([]models.Task, error) {
	logging.Info("正在获取所有任务")

	tasks, err := s.taskDAO.GetAllTasks()
	if err != nil {
		logging.Error("获取所有任务失败: %v", err)
		return nil, err
	}

	logging.Info("成功获取所有任务，共 %d 个", len(tasks))
	return tasks, nil
}

// UpdateTaskStatus 更新指定任务的状态
func (s *TaskService) UpdateTaskStatus(id string, status models.TaskStatus) error {
	logging.Info("正在更新任务状态: %s 到 %s", id, status)

	if err := s.taskDAO.UpdateTaskStatus(id, status); err != nil {
		logging.Error("更新任务状态失败: %s, 错误: %v", id, err)
		return err
	}

	logging.Info("成功更新任务状态: %s 到 %s", id, status)
	return nil
}

// DeleteTask 删除指定的任务
func (s *TaskService) DeleteTask(id string) error {
	logging.Info("正在删除任务: %s", id)

	if err := s.taskDAO.DeleteTask(id); err != nil {
		logging.Error("删除任务失败: %s, 错误: %v", id, err)
		return err
	}

	logging.Info("成功删除任务: %s", id)
	return nil
}
