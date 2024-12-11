package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// Target 模型增加统计字段
type Target struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Type        string             `json:"type" bson:"type"`
	Target      string             `json:"target" bson:"target"`
	Status      string             `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updated_at"`
	// 新增统计字段
	SubdomainCount     int `json:"subdomain_count" bson:"subdomain_count"`
	PortCount          int `json:"port_count" bson:"port_count"`
	PathCount          int `json:"path_count" bson:"path_count"`
	VulnerabilityCount int `json:"vulnerability_count" bson:"vulnerability_count"`
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

// TargetStats 目标统计信息
type TargetStats struct {
	SubdomainCount     int `json:"subdomain_count"`
	PortCount          int `json:"port_count"`
	PathCount          int `json:"path_count"`
	VulnerabilityCount int `json:"vulnerability_count"`
}

// TargetDetails 目标详细信息
type TargetDetails struct {
	Target *Target     `json:"target"`
	Stats  TargetStats `json:"stats"`
}
