// CyberEdge/pkg/models/tasks.go

package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// 定义任务状态常量
const (
	StatusRunning = "运行中"
	StatusWaiting = "等待中"
	StatusStopped = "停止中"
)

// Task 代表一个可存储的任务结构体
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // 使用 ObjectID 类型
	TaskID      string             `bson:"task_id"`       // 使用字符串作为任务ID
	Description string             `bson:"description"`
	Interval    int                `bson:"interval"` // 使用 int 类型表示以分钟为单位
	Status      string             `bson:"status"`   // 当前状态
	CreatedAt   time.Time          `bson:"created_at"`
	RunCount    int                `bson:"run_count"` // 新增字段：运行次数
}
