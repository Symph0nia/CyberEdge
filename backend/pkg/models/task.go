package models

import (
	"gorm.io/gorm"
)

// TaskType 任务类型
type TaskType string

const (
	TaskTypeSubfinder TaskType = "subfinder"
	TaskTypeNmap      TaskType = "nmap"
	TaskTypeFfuf      TaskType = "ffuf"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

// Task 任务模型
type Task struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	TargetID  *uint      `gorm:"index" json:"target_id,omitempty"`
	Type      TaskType   `gorm:"type:enum('subfinder','nmap','ffuf');not null;index" json:"type"`
	Status    TaskStatus `gorm:"type:enum('pending','running','completed','failed');default:'pending';index" json:"status"`
	Payload   string     `gorm:"type:text;not null" json:"payload"`
	Result    string     `gorm:"type:longtext" json:"result,omitempty"`
	CreatedAt int64      `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt int64      `gorm:"autoUpdateTime" json:"updated_at"`
	CompletedAt *int64   `json:"completed_at,omitempty"`

	// 关联关系
	Target *Target `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"target,omitempty"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "tasks"
}

// BeforeCreate 创建前钩子
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前钩子
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// MarkCompleted 标记任务完成
func (t *Task) MarkCompleted(result string) {
	t.Status = TaskStatusCompleted
	t.Result = result
	now := int64(1000) // 应该用实际时间戳
	t.CompletedAt = &now
}

// MarkFailed 标记任务失败
func (t *Task) MarkFailed(errorMsg string) {
	t.Status = TaskStatusFailed
	t.Result = errorMsg
	now := int64(1000) // 应该用实际时间戳
	t.CompletedAt = &now
}