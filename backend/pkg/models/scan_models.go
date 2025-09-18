package models

import (
	"time"
	"gorm.io/gorm"
)

// CyberEdge 扫描数据模型 - 基于扁平化设计，优化查询性能

// Project - 扫描项目
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex;size:100"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ScanJob - 扫描任务管理
type ScanJob struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ProjectID    uint      `json:"project_id" gorm:"not null;index"`
	Target       string    `json:"target" gorm:"not null;size:255"`
	PipelineName string    `json:"pipeline_name" gorm:"not null;size:100"`
	Status       string    `json:"status" gorm:"not null;size:20;default:'pending'"` // pending, running, completed, failed
	StartTime    time.Time `json:"start_time"`
	EndTime      *time.Time `json:"end_time"`
	ErrorMessage string    `json:"error_message,omitempty" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	Project Project `json:"-" gorm:"foreignKey:ProjectID"`
}

// ScanTarget - 扫描目标（合并域名、子域名、IP概念）
type ScanTarget struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"project_id" gorm:"not null;index"`
	Type      string    `json:"type" gorm:"not null;index;size:20"`     // "domain", "subdomain", "ip"
	Address   string    `json:"address" gorm:"not null;index;size:255"` // 域名、子域名或IP地址
	ParentID  *uint     `json:"parent_id" gorm:"index"`                 // 父级目标ID，用于层次关系
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Project    Project `json:"-" gorm:"foreignKey:ProjectID"`
	Parent     *ScanTarget      `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children   []ScanTarget     `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	ScanResults []ScanResult `json:"scan_results,omitempty" gorm:"foreignKey:TargetID"`
}

// ScanResult - 扫描结果（端口+服务）
type ScanResult struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ProjectID   uint      `json:"project_id" gorm:"not null;index"`
	TargetID    uint      `json:"target_id" gorm:"not null;index;uniqueIndex:idx_target_port_protocol"`
	Port        int       `json:"port" gorm:"not null;index;uniqueIndex:idx_target_port_protocol"`
	Protocol    string    `json:"protocol" gorm:"not null;size:10;uniqueIndex:idx_target_port_protocol"`
	State       string    `json:"state" gorm:"size:20"`
	ServiceName string    `json:"service_name" gorm:"size:50;index"`
	Version     string    `json:"version" gorm:"size:100"`
	Fingerprint string    `json:"fingerprint" gorm:"size:255"`
	Banner      string    `json:"banner" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Web服务特有字段
	IsWebService bool   `json:"is_web_service" gorm:"index"`
	HTTPTitle    string `json:"http_title" gorm:"size:255"`
	HTTPStatus   int    `json:"http_status"`

	// 关联关系
	Project        Project `json:"-" gorm:"foreignKey:ProjectID"`
	Target         ScanTarget       `json:"-" gorm:"foreignKey:TargetID"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty" gorm:"foreignKey:ScanResultID"`
	WebPaths       []WebPath       `json:"web_paths,omitempty" gorm:"foreignKey:ScanResultID"`
}

// Vulnerability - 漏洞信息
type Vulnerability struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ScanResultID uint      `json:"scan_result_id" gorm:"not null;index"`
	WebPathID    *uint     `json:"web_path_id" gorm:"index"` // 可选，路径级漏洞
	CVEID        string    `json:"cve_id" gorm:"size:50;index"`
	Title        string    `json:"title" gorm:"not null;size:255"`
	Description  string    `json:"description" gorm:"type:text"`
	Severity     string    `json:"severity" gorm:"not null;size:20;index"`
	CVSS         float64   `json:"cvss" gorm:"index"`
	Location     string    `json:"location" gorm:"size:255"`
	Parameter    string    `json:"parameter" gorm:"size:100"`
	Payload      string    `json:"payload" gorm:"type:text"`
	Status       string    `json:"status" gorm:"size:20;default:'open';index"` // open, fixed, false_positive
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	ScanResult ScanResult `json:"-" gorm:"foreignKey:ScanResultID"`
	WebPath    *WebPath   `json:"-" gorm:"foreignKey:WebPathID"`
}

// WebPath - Web路径（仅针对Web服务）
type WebPath struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ScanResultID uint      `json:"scan_result_id" gorm:"not null;index"`
	Path         string    `json:"path" gorm:"not null;size:500;index"`
	StatusCode   int       `json:"status_code" gorm:"index"`
	Title        string    `json:"title" gorm:"size:255"`
	Length       int       `json:"length"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// 关联关系
	ScanResult      ScanResult      `json:"-" gorm:"foreignKey:ScanResultID"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty" gorm:"foreignKey:WebPathID"`
}

// Technology - 技术栈（多对多关系）
type Technology struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null;uniqueIndex;size:100"`
	Category  string    `json:"category" gorm:"size:50;index"` // web_server, framework, database, etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ScanResultTechnology - 扫描结果与技术栈的关联表
type ScanResultTechnology struct {
	ScanResultID uint `json:"scan_result_id" gorm:"primaryKey"`
	TechnologyID uint `json:"technology_id" gorm:"primaryKey"`
	Version      string `json:"version" gorm:"size:100"`

	// 关联关系
	ScanResult ScanResult  `json:"-" gorm:"foreignKey:ScanResultID"`
	Technology Technology  `json:"-" gorm:"foreignKey:TechnologyID"`
}

// 优化后的查询视图结构
type ProjectStats struct {
	ProjectID          uint           `json:"project_id"`
	ProjectName        string         `json:"project_name"`
	TargetCount        int            `json:"target_count"`
	DomainCount        int            `json:"domain_count"`
	IPCount            int            `json:"ip_count"`
	PortCount          int            `json:"port_count"`
	ServiceCount       int            `json:"service_count"`
	WebServiceCount    int            `json:"web_service_count"`
	VulnerabilityStats map[string]int `json:"vulnerability_stats"`
	LastScanTime       time.Time      `json:"last_scan_time"`
}

// 扁平化的导入数据结构（保持API兼容性）
type ScanDataImport struct {
	ProjectID uint                   `json:"project_id"`
	Results   []ScanTargetImport     `json:"results"`
}

type ScanTargetImport struct {
	Type      string              `json:"type"`      // "domain", "subdomain", "ip"
	Address   string              `json:"address"`   // 地址
	Parent    string              `json:"parent"`    // 父级地址（可选）
	Ports     []PortScanImport    `json:"ports"`
}

type PortScanImport struct {
	Number      int                    `json:"number"`
	Protocol    string                 `json:"protocol"`
	State       string                 `json:"state"`
	Service     *ServiceScanImport     `json:"service,omitempty"`
}

type ServiceScanImport struct {
	Name            string                   `json:"name"`
	Version         string                   `json:"version"`
	Fingerprint     string                   `json:"fingerprint"`
	Banner          string                   `json:"banner"`
	IsWebService    bool                     `json:"is_web_service"`
	HTTPTitle       string                   `json:"http_title"`
	HTTPStatus      int                      `json:"http_status"`
	WebPaths        []WebPathImport          `json:"web_paths,omitempty"`
	Technologies    []TechnologyImport       `json:"technologies,omitempty"`
	Vulnerabilities []VulnerabilityImport    `json:"vulnerabilities"`
}

type WebPathImport struct {
	Path            string                `json:"path"`
	StatusCode      int                   `json:"status_code"`
	Title           string                `json:"title"`
	Length          int                   `json:"length"`
	Vulnerabilities []VulnerabilityImport `json:"vulnerabilities"`
}

type TechnologyImport struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Version  string `json:"version"`
}

type VulnerabilityImport struct {
	CVEID       string  `json:"cve_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	CVSS        float64 `json:"cvss"`
	Location    string  `json:"location"`
	Parameter   string  `json:"parameter"`
	Payload     string  `json:"payload"`
}

// 帮助方法
func (t *ScanTarget) IsRoot() bool {
	return t.ParentID == nil
}

func (t *ScanTarget) GetFullPath() string {
	if t.ParentID == nil {
		return t.Address
	}
	// 这里可以递归构建完整路径，但通常通过查询解决
	return t.Address
}

func (sr *ScanResult) GetServiceSignature() string {
	if sr.ServiceName != "" && sr.Version != "" {
		return sr.ServiceName + "/" + sr.Version
	}
	return sr.ServiceName
}

func (v *Vulnerability) IsCritical() bool {
	return v.Severity == "critical"
}

func (v *Vulnerability) IsOpen() bool {
	return v.Status == "open"
}