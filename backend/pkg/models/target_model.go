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
	SubdomainCount     int              `json:"subdomain_count"`
	PortCount          int              `json:"port_count"`
	PathCount          int              `json:"path_count"`
	VulnerabilityCount int              `json:"vulnerability_count"`
	TopPorts           []PortStat       `json:"top_ports"`         // 新增端口排名
	HTTPStatusStats    []HTTPStatusStat `json:"http_status_stats"` // 新增HTTP状态码统计
}

// TargetDetails 目标详细信息
type TargetDetails struct {
	Target *Target     `json:"target"`
	Stats  TargetStats `json:"stats"`
}

// PortStat 端口统计信息
type PortStat struct {
	Port  int `json:"port" bson:"port"`
	Count int `json:"count" bson:"count"`
}

// HTTPStatusStat HTTP状态码统计信息
type HTTPStatusStat struct {
	Status int    `json:"status" bson:"status"`
	Count  int    `json:"count" bson:"count"`
	Label  string `json:"label" bson:"label"` // 用于展示的标签
}
