package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Target 表示一个扫描目标
type Target struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`               // 目标名称
	Description string             `json:"description" bson:"description"` // 目标描述
	Type        string             `json:"type" bson:"type"`               // "domain" 或 "ip"
	Target      string             `json:"target" bson:"target"`           // 具体的域名或IP地址
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`    // 创建时间
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`    // 更新时间
}

// BeforeCreate 创建前的钩子
func (t *Target) BeforeCreate() {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
}

// BeforeUpdate 更新前的钩子
func (t *Target) BeforeUpdate() {
	t.UpdatedAt = time.Now()
}
