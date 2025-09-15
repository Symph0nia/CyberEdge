package models

import (
	"gorm.io/gorm"
)

// TargetType 目标类型
type TargetType string

const (
	TargetTypeDomain TargetType = "domain"
	TargetTypeIP     TargetType = "ip"
	TargetTypeURL    TargetType = "url"
)

// TargetStatus 目标状态
type TargetStatus string

const (
	TargetStatusActive   TargetStatus = "active"
	TargetStatusInactive TargetStatus = "inactive"
	TargetStatusArchived TargetStatus = "archived"
)

// Target 目标模型
type Target struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"size:255;not null" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	Type        TargetType   `gorm:"type:enum('domain','ip','url');not null" json:"type"`
	Target      string       `gorm:"size:500;not null;index" json:"target"`
	Status      TargetStatus `gorm:"type:enum('active','inactive','archived');default:'active';index" json:"status"`

	// 统计字段
	SubdomainCount     int `gorm:"default:0" json:"subdomain_count"`
	PortCount          int `gorm:"default:0" json:"port_count"`
	PathCount          int `gorm:"default:0" json:"path_count"`
	VulnerabilityCount int `gorm:"default:0" json:"vulnerability_count"`

	CreatedAt int64 `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime" json:"updated_at"`

	// 关联关系
	Tasks      []Task      `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"tasks,omitempty"`
	Subdomains []Subdomain `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"subdomains,omitempty"`
	Ports      []Port      `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"ports,omitempty"`
	Paths      []Path      `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"paths,omitempty"`
}

// TargetStats 目标统计
type TargetStats struct {
	SubdomainCount     int              `json:"subdomain_count"`
	PortCount          int              `json:"port_count"`
	PathCount          int              `json:"path_count"`
	VulnerabilityCount int              `json:"vulnerability_count"`
	TopPorts           []PortStat       `json:"top_ports"`
	HTTPStatusStats    []HTTPStatusStat `json:"http_status_stats"`
}

// TargetDetails 目标详情
type TargetDetails struct {
	Target *Target     `json:"target"`
	Stats  TargetStats `json:"stats"`
}

// PortStat 端口统计
type PortStat struct {
	Port  int `json:"port"`
	Count int `json:"count"`
}

// HTTPStatusStat HTTP状态码统计
type HTTPStatusStat struct {
	Status int    `json:"status"`
	Count  int    `json:"count"`
	Label  string `json:"label"`
}

// TableName 指定表名
func (Target) TableName() string {
	return "targets"
}

// BeforeCreate 创建前钩子
func (t *Target) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前钩子
func (t *Target) BeforeUpdate(tx *gorm.DB) error {
	return nil
}