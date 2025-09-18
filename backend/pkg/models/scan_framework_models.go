package models

import (
	"time"
)

// ScanFrameworkResult 扫描框架专用结果模型
type ScanFrameworkResult struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProjectID   uint      `json:"project_id" gorm:"not null;index"`
	ScanTargetID uint     `json:"scan_target_id" gorm:"not null;index"`
	Target      string    `json:"target" gorm:"not null;size:255"`
	ScanType    string    `json:"scan_type" gorm:"not null;size:50;index"`
	ScannerName string    `json:"scanner_name" gorm:"not null;size:100"`
	Status      string    `json:"status" gorm:"not null;size:20;index"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	RawData     string    `json:"raw_data" gorm:"type:text"`
	ErrorMessage string   `json:"error_message" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Project    ProjectOptimized `json:"-" gorm:"foreignKey:ProjectID"`
	ScanTarget *ScanTarget      `json:"scan_target,omitempty" gorm:"foreignKey:ScanTargetID"`
}

// TableName 指定表名
func (ScanFrameworkResult) TableName() string {
	return "scan_framework_results"
}

// ScanFrameworkTarget 扫描框架目标模型
type ScanFrameworkTarget struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ProjectID  uint      `json:"project_id" gorm:"not null;index"`
	Target     string    `json:"target" gorm:"not null;size:255;index"`
	TargetType string    `json:"target_type" gorm:"not null;size:50"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联关系
	Project ProjectOptimized `json:"-" gorm:"foreignKey:ProjectID"`
}

// TableName 指定表名
func (ScanFrameworkTarget) TableName() string {
	return "scan_framework_targets"
}

// ScanFrameworkVulnerability 扫描框架漏洞模型
type ScanFrameworkVulnerability struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ProjectID    uint      `json:"project_id" gorm:"not null;index"`
	ScanTargetID uint      `json:"scan_target_id" gorm:"not null;index"`
	CVEID        string    `json:"cve_id" gorm:"size:50;index"`
	Title        string    `json:"title" gorm:"not null;size:255"`
	Description  string    `json:"description" gorm:"type:text"`
	Severity     string    `json:"severity" gorm:"not null;size:20;index"`
	CVSS         float64   `json:"cvss" gorm:"index"`
	Location     string    `json:"location" gorm:"size:255"`
	Parameter    string    `json:"parameter" gorm:"size:100"`
	Payload      string    `json:"payload" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Project    ProjectOptimized        `json:"-" gorm:"foreignKey:ProjectID"`
	ScanTarget *ScanFrameworkTarget    `json:"scan_target,omitempty" gorm:"foreignKey:ScanTargetID"`
}

// TableName 指定表名
func (ScanFrameworkVulnerability) TableName() string {
	return "scan_framework_vulnerabilities"
}