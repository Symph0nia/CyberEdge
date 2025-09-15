package service

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
	"cyberedge/pkg/scanner"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskService struct {
	scanner *scanner.Scanner
}

// NewTaskService 创建新的TaskService
func NewTaskService(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *TaskService {
	return &TaskService{
		scanner: scanner.NewScanner(taskDAO, resultDAO),
	}
}

// CreateTask 创建并执行扫描任务
func (s *TaskService) CreateTask(taskType string, target string, targetID *primitive.ObjectID) (*models.Task, error) {
	req := scanner.ScanRequest{
		Type:   taskType,
		Target: target,
	}

	if targetID != nil {
		req.TargetID = targetID.Hex()
	}

	return s.scanner.ExecuteScan(context.Background(), req)
}

// GetTaskByID 根据ID获取任务
func (s *TaskService) GetTaskByID(taskID primitive.ObjectID) (*models.Task, error) {
	return s.scanner.GetTaskStatus(taskID)
}

// GetAllTasks 获取所有任务
func (s *TaskService) GetAllTasks() ([]*models.Task, error) {
	return s.scanner.ListTasks()
}

// CancelTask 取消任务
func (s *TaskService) CancelTask(taskID primitive.ObjectID) error {
	return s.scanner.CancelTask(taskID)
}

// Close 关闭服务（现在不需要关闭任何连接）
func (s *TaskService) Close() {
	// 不再需要关闭Asynq客户端
}