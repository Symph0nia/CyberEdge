package models

import (
	"time"
	"gorm.io/gorm"
)

// Project - 扫描项目根节点
type Project struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null;uniqueIndex;size:100"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// 关联关系
	Domains []Domain `json:"domains" gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

// Domain - 域名节点
type Domain struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ProjectID uint      `json:"project_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null;index;size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Project    Project    `json:"-" gorm:"foreignKey:ProjectID"`
	Subdomains []Subdomain `json:"subdomains" gorm:"foreignKey:DomainID;constraint:OnDelete:CASCADE"`
}

// Subdomain - 子域名节点
type Subdomain struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DomainID  uint      `json:"domain_id" gorm:"not null;index"`
	Name      string    `json:"name" gorm:"not null;index;size:100"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Domain     Domain      `json:"-" gorm:"foreignKey:DomainID"`
	IPAddresses []IPAddress `json:"ip_addresses" gorm:"foreignKey:SubdomainID;constraint:OnDelete:CASCADE"`
}

// IPAddress - IP地址节点
type IPAddress struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SubdomainID *uint     `json:"subdomain_id" gorm:"index"` // 可选，直接扫描IP时为空
	Address     string    `json:"address" gorm:"not null;index;size:45"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Subdomain *Subdomain `json:"-" gorm:"foreignKey:SubdomainID"`
	Ports     []Port     `json:"ports" gorm:"foreignKey:IPAddressID;constraint:OnDelete:CASCADE"`
}

// Port - 端口节点
type Port struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	IPAddressID uint      `json:"ip_address_id" gorm:"not null;index"`
	Number      int       `json:"number" gorm:"not null"`
	Protocol    string    `json:"protocol" gorm:"not null;size:10"` // tcp, udp
	State       string    `json:"state" gorm:"size:20"`                    // open, closed, filtered
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系 - 每个端口只有一个服务
	IPAddress IPAddress `json:"-" gorm:"foreignKey:IPAddressID"`
	Service   *Service  `json:"service" gorm:"foreignKey:PortID;constraint:OnDelete:CASCADE"`
}

// Service - 服务基础表（统一存储所有服务类型）
type Service struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	PortID      uint      `json:"port_id" gorm:"not null;uniqueIndex"`
	Type        string    `json:"type" gorm:"not null;size:50"`        // http, https, ssh, ftp, mysql, etc.
	Name        string    `json:"name" gorm:"size:100"`                        // 服务名称
	Version     string    `json:"version" gorm:"size:100"`                     // 服务版本
	Fingerprint string    `json:"fingerprint" gorm:"type:text"` // 服务指纹
	Banner      string    `json:"banner" gorm:"type:text"`     // 服务横幅
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Port            Port             `json:"-" gorm:"foreignKey:PortID"`
	Vulnerabilities []Vulnerability  `json:"vulnerabilities" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`

	// Web服务特有的关联（只有当Type为http/https时才有数据）
	WebPaths       []WebPath        `json:"web_paths,omitempty" gorm:"foreignKey:ServiceID;constraint:OnDelete:CASCADE"`
	Technologies   []Technology     `json:"technologies,omitempty" gorm:"many2many:service_technologies;constraint:OnDelete:CASCADE"`
}

// WebPath - Web路径（仅用于Web服务）
type WebPath struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	ServiceID  uint      `json:"service_id" gorm:"not null;index"`
	Path       string    `json:"path" gorm:"not null;size:500"`
	StatusCode int       `json:"status_code"`
	Title      string    `json:"title" gorm:"size:200"`
	Length     int       `json:"length"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联关系
	Service         Service         `json:"-" gorm:"foreignKey:ServiceID"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities" gorm:"foreignKey:WebPathID;constraint:OnDelete:CASCADE"`
}

// Technology - Web技术栈
type Technology struct {
	ID      uint   `json:"id" gorm:"primaryKey"`
	Name    string `json:"name" gorm:"not null;uniqueIndex;size:100"`
	Version string `json:"version" gorm:"size:100"`

	// 多对多关系
	Services []Service `json:"-" gorm:"many2many:service_technologies"`
}

// Vulnerability - 漏洞信息
type Vulnerability struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	ServiceID   *uint     `json:"service_id" gorm:"index"`   // 服务级漏洞
	WebPathID   *uint     `json:"web_path_id" gorm:"index"` // 路径级漏洞

	// 漏洞基本信息
	CVEID       string    `json:"cve_id" gorm:"index;size:20"`
	Title       string    `json:"title" gorm:"not null;size:200"`
	Description string    `json:"description" gorm:"type:text"`
	Severity    string    `json:"severity" gorm:"not null;size:20"` // critical, high, medium, low, info
	CVSS        float64   `json:"cvss"`

	// 位置信息
	Location    string    `json:"location" gorm:"size:500"`    // 具体位置描述
	Parameter   string    `json:"parameter" gorm:"size:100"`   // 漏洞参数
	Payload     string    `json:"payload" gorm:"type:text"`     // 测试载荷

	// 状态信息
	Status      string    `json:"status" gorm:"default:open;size:20"` // open, fixed, false_positive
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联关系
	Service *Service `json:"-" gorm:"foreignKey:ServiceID"`
	WebPath *WebPath `json:"-" gorm:"foreignKey:WebPathID"`
}

// ServiceTechnology - 服务与技术的多对多关联表
type ServiceTechnology struct {
	ServiceID    uint `gorm:"primaryKey"`
	TechnologyID uint `gorm:"primaryKey"`
}

// 辅助方法 - 检查服务是否为Web服务
func (s *Service) IsWebService() bool {
	return s.Type == "http" || s.Type == "https"
}

// 辅助方法 - 获取所有漏洞（包括路径级漏洞）
func (s *Service) GetAllVulnerabilities() []Vulnerability {
	var allVulns []Vulnerability

	// 服务级漏洞
	allVulns = append(allVulns, s.Vulnerabilities...)

	// 路径级漏洞（仅Web服务）
	if s.IsWebService() {
		for _, path := range s.WebPaths {
			allVulns = append(allVulns, path.Vulnerabilities...)
		}
	}

	return allVulns
}

// 辅助方法 - 获取高危漏洞统计
func (s *Service) GetVulnerabilityStats() map[string]int {
	stats := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	for _, vuln := range s.GetAllVulnerabilities() {
		if count, exists := stats[vuln.Severity]; exists {
			stats[vuln.Severity] = count + 1
		}
	}

	return stats
}