package factory

import (
	"cyberedge/pkg/models"
	"github.com/google/uuid"
	"time"
)

// CreateTask 创建一个新的任务
func CreateTask(taskType models.TaskType, description string, interval int) *models.Task {
	return &models.Task{
		ID:          uuid.New().String(),
		Type:        taskType,
		Description: description,
		Status:      models.TaskStatusWaiting,
		Interval:    interval,
		RunCount:    0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// CreatePingTask 创建一个新的 Ping 任务
func CreatePingTask(description string, interval int) *models.Task {
	return CreateTask(models.TaskTypePing, description, interval)
}
