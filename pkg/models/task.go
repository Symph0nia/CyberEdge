package models

import (
	"time"
)

// TaskStatus 定义任务状态的类型
type TaskStatus string

const (
	TaskStatusWaiting   TaskStatus = "waiting"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusError     TaskStatus = "error"
	TaskStatusStopped   TaskStatus = "stopped"
	TaskStatusScheduled TaskStatus = "scheduled"
)

// TaskType 定义任务类型的类型
type TaskType string

const (
	TaskTypePing TaskType = "ping"
	// 可以在此添加其他任务类型
)

// Task 定义任务结构
type Task struct {
	ID          string     `json:"id" bson:"_id,omitempty"`
	Type        TaskType   `json:"type" bson:"type"`
	Description string     `json:"description" bson:"description"`
	Status      TaskStatus `json:"status" bson:"status"`
	Interval    int        `json:"interval" bson:"interval"`
	RunCount    int        `json:"run_count" bson:"run_count"`
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`
	Result      string     `json:"result,omitempty" bson:"result"`
}

// UpdateStatus 更新任务状态
func (t *Task) UpdateStatus(status TaskStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()
}

// IncrementRunCount 增加任务运行次数
func (t *Task) IncrementRunCount() {
	t.RunCount++
	t.UpdatedAt = time.Now()
}

// SetResult 设置任务执行结果
func (t *Task) SetResult(result string) {
	t.Result = result
	t.UpdatedAt = time.Now()
}
