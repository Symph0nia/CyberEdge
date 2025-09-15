package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
)

type Task struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id,omitempty"`
	Type        string              `bson:"type" json:"type"`
	Status      TaskStatus          `bson:"status" json:"status"`
	Payload     string              `bson:"payload" json:"payload"`
	TargetID    *primitive.ObjectID `bson:"target_id,omitempty" json:"target_id,omitempty"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at" json:"updated_at"`
	CompletedAt *time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	Result      string              `bson:"result,omitempty" json:"result,omitempty"`
}
