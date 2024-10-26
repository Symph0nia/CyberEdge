// CyberEdge/models/task.go

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
	TaskStatusScheduled TaskStatus = "scheduled" // 添加这个常量
)

// TaskType 定义任务类型的类型
type TaskType string

const (
	TaskTypePing TaskType = "ping"
	// 可以在此添加其他任务类型
)

// Task 定义任务结构
type Task struct {
	ID          string     `json:"id" bson:"_id,omitempty"`        // 任务唯一标识符
	Type        TaskType   `json:"type" bson:"type"`               // 任务类型
	Description string     `json:"description" bson:"description"` // 任务描述
	Status      TaskStatus `json:"status" bson:"status"`           // 任务状态
	Interval    int        `json:"interval" bson:"interval"`       // 运行间隔（分钟）
	RunCount    int        `json:"run_count" bson:"run_count"`     // 运行次数
	CreatedAt   time.Time  `json:"created_at" bson:"created_at"`   // 创建时间
	UpdatedAt   time.Time  `json:"updated_at" bson:"updated_at"`   // 最后更新时间
	Result      string     `json:"result,omitempty" bson:"result"` // 任务执行结果
}

// NewTask 创建一个新的任务
func NewTask(id string, taskType TaskType, description string, interval int) *Task {
	now := time.Now()
	return &Task{
		ID:          id,
		Type:        taskType,
		Description: description,
		Status:      TaskStatusWaiting,
		Interval:    interval,
		RunCount:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
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
