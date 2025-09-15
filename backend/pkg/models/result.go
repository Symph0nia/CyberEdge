package models

import (
	"gorm.io/gorm"
)

// Subdomain 子域名结果模型
type Subdomain struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TargetID   uint   `gorm:"not null;index" json:"target_id"`
	TaskID     uint   `gorm:"not null;index" json:"task_id"`
	Domain     string `gorm:"size:500;not null;index" json:"domain"`
	IP         string `gorm:"size:45" json:"ip"` // 支持IPv6
	HTTPStatus int    `gorm:"default:0" json:"http_status"`
	HTTPTitle  string `gorm:"size:1000" json:"http_title"`
	IsAlive    bool   `gorm:"default:false;index" json:"is_alive"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at"`

	// 关联关系
	Target *Target `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"target,omitempty"`
	Task   *Task   `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
}

// Port 端口扫描结果模型
type Port struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	TargetID   uint        `gorm:"not null;index" json:"target_id"`
	TaskID     uint        `gorm:"not null;index" json:"task_id"`
	Host       string      `gorm:"size:500;not null" json:"host"`
	Port       int         `gorm:"not null;index" json:"port"`
	Protocol   Protocol    `gorm:"type:enum('tcp','udp');default:'tcp'" json:"protocol"`
	State      PortState   `gorm:"type:enum('open','closed','filtered');not null;index" json:"state"`
	Service    string      `gorm:"size:100" json:"service"`
	HTTPStatus int         `gorm:"default:0" json:"http_status"`
	HTTPTitle  string      `gorm:"size:1000" json:"http_title"`
	CreatedAt  int64       `gorm:"autoCreateTime" json:"created_at"`

	// 关联关系
	Target *Target `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"target,omitempty"`
	Task   *Task   `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
}

// Path 路径扫描结果模型
type Path struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	TargetID      uint   `gorm:"not null;index" json:"target_id"`
	TaskID        uint   `gorm:"not null;index" json:"task_id"`
	URL           string `gorm:"size:1000;not null" json:"url"`
	Path          string `gorm:"size:500;not null;index" json:"path"`
	StatusCode    int    `gorm:"not null;index" json:"status_code"`
	ContentLength int    `gorm:"default:0" json:"content_length"`
	ContentWords  int    `gorm:"default:0" json:"content_words"`
	ContentLines  int    `gorm:"default:0" json:"content_lines"`
	Title         string `gorm:"size:1000" json:"title"`
	CreatedAt     int64  `gorm:"autoCreateTime" json:"created_at"`

	// 关联关系
	Target *Target `gorm:"foreignKey:TargetID;constraint:OnDelete:CASCADE" json:"target,omitempty"`
	Task   *Task   `gorm:"foreignKey:TaskID;constraint:OnDelete:CASCADE" json:"task,omitempty"`
}

// Protocol 协议类型
type Protocol string

const (
	ProtocolTCP Protocol = "tcp"
	ProtocolUDP Protocol = "udp"
)

// PortState 端口状态
type PortState string

const (
	PortStateOpen     PortState = "open"
	PortStateClosed   PortState = "closed"
	PortStateFiltered PortState = "filtered"
)

// TableName 指定表名
func (Subdomain) TableName() string {
	return "subdomains"
}

func (Port) TableName() string {
	return "ports"
}

func (Path) TableName() string {
	return "paths"
}

// BeforeCreate 钩子
func (s *Subdomain) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (p *Port) BeforeCreate(tx *gorm.DB) error {
	return nil
}

func (p *Path) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// 实现接口方法（保持向后兼容）
func (s Subdomain) GetID() string {
	return string(rune(s.ID))
}

func (s Subdomain) GetProbeURL() string {
	return s.Domain
}

func (p Port) GetID() string {
	return string(rune(p.ID))
}

func (p Port) GetProbeURL() string {
	return p.Host
}

func (pa Path) GetID() string {
	return string(rune(pa.ID))
}

func (pa Path) GetProbeURL() string {
	return pa.URL
}