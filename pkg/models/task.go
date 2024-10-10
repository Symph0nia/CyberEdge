package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Task 代表一个可存储的任务结构体
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"` // 使用 ObjectID 类型
	TaskID      string             `bson:"task_id"`       // 使用字符串作为任务ID
	Description string             `bson:"description"`
	Interval    time.Duration      `bson:"interval"`
	IsRunning   bool               `bson:"is_running"`
	CreatedAt   time.Time          `bson:"created_at"`
}
