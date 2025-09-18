package scanner

import (
	"context"
	"time"
)

// Scanner - 所有扫描工具的统一接口
type Scanner interface {
	// GetName 获取扫描工具名称
	GetName() string

	// GetCategory 获取扫描类别
	GetCategory() ScanCategory

	// IsAvailable 检查工具是否可用
	IsAvailable() bool

	// Scan 执行扫描
	Scan(ctx context.Context, config ScanConfig) (*ScanResult, error)

	// ValidateConfig 验证配置参数
	ValidateConfig(config ScanConfig) error
}

// ScanCategory 扫描类别枚举
type ScanCategory string

const (
	CategorySubdomain     ScanCategory = "subdomain"     // 子域名发现
	CategoryPort          ScanCategory = "port"          // 端口扫描
	CategoryService       ScanCategory = "service"       // 服务识别
	CategoryWebTech       ScanCategory = "webtech"       // Web技术栈
	CategoryWebPath       ScanCategory = "webpath"       // Web路径发现
	CategoryVulnerability ScanCategory = "vulnerability" // 漏洞扫描
)

// ScanConfig 统一扫描配置
type ScanConfig struct {
	// 通用参数
	ProjectID uint              `json:"project_id"`
	Target    string            `json:"target"`      // 扫描目标
	Options   map[string]string `json:"options"`     // 工具特定选项
	Timeout   time.Duration     `json:"timeout"`     // 超时时间

	// 上下文参数（用于级联扫描）
	ParentResults []ScanResult `json:"parent_results,omitempty"`
}

// ScanResult 统一扫描结果
type ScanResult struct {
	// 元数据
	ScannerName string        `json:"scanner_name"`
	Category    ScanCategory  `json:"category"`
	Target      string        `json:"target"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     time.Time     `json:"end_time"`
	Duration    time.Duration `json:"duration"`
	Status      ScanStatus    `json:"status"`
	Error       string        `json:"error,omitempty"`

	// 结果数据（根据category类型解析不同字段）
	Data interface{} `json:"data"`
}

// ScanStatus 扫描状态
type ScanStatus string

const (
	StatusPending   ScanStatus = "pending"   // 等待中
	StatusRunning   ScanStatus = "running"   // 执行中
	StatusCompleted ScanStatus = "completed" // 已完成
	StatusFailed    ScanStatus = "failed"    // 执行失败
	StatusTimeout   ScanStatus = "timeout"   // 执行超时
)

// 具体结果数据结构

// SubdomainData 子域名扫描结果
type SubdomainData struct {
	Subdomains []SubdomainInfo `json:"subdomains"`
}

type SubdomainInfo struct {
	Domain    string   `json:"domain"`    // example.com
	Subdomain string   `json:"subdomain"` // api.example.com
	IPs       []string `json:"ips"`       // 解析的IP地址
	Source    string   `json:"source"`    // 发现来源
}

// PortData 端口扫描结果
type PortData struct {
	Ports []PortInfo `json:"ports"`
}

type PortInfo struct {
	Port     int          `json:"port"`
	Protocol string       `json:"protocol"` // tcp/udp
	State    string       `json:"state"`    // open/closed/filtered
	Service  *ServiceInfo `json:"service,omitempty"`
}

type ServiceInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Fingerprint string `json:"fingerprint"`
	Banner      string `json:"banner"`
}

// WebTechData Web技术栈识别结果
type WebTechData struct {
	URL          string           `json:"url"`
	StatusCode   int              `json:"status_code"`
	Title        string           `json:"title"`
	Technologies []TechnologyInfo `json:"technologies"`
}

type TechnologyInfo struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Version  string `json:"version"`
}

// WebPathData Web路径发现结果
type WebPathData struct {
	Paths []WebPathInfo `json:"paths"`
}

type WebPathInfo struct {
	URL        string `json:"url"`
	Path       string `json:"path"`
	StatusCode int    `json:"status_code"`
	Length     int    `json:"length"`
	Title      string `json:"title"`
}

// VulnerabilityData 漏洞扫描结果
type VulnerabilityData struct {
	Vulnerabilities []VulnerabilityInfo `json:"vulnerabilities"`
}

type VulnerabilityInfo struct {
	Target      string  `json:"target"`
	CVEID       string  `json:"cve_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	CVSS        float64 `json:"cvss"`
	Location    string  `json:"location"`
	Parameter   string  `json:"parameter"`
	Payload     string  `json:"payload"`
}

// ScanManager 扫描管理器接口
type ScanManager interface {
	// RegisterScanner 注册扫描工具
	RegisterScanner(scanner Scanner) error

	// GetScanner 获取指定扫描工具
	GetScanner(name string) (Scanner, error)

	// ListScanners 列出所有可用扫描工具
	ListScanners() []Scanner

	// ListByCategory 按类别列出扫描工具
	ListByCategory(category ScanCategory) []Scanner

	// ExecuteScan 执行扫描任务
	ExecuteScan(ctx context.Context, config ScanConfig) (*ScanResult, error)

	// ExecutePipeline 执行扫描流水线
	ExecutePipeline(ctx context.Context, pipeline ScanPipeline) ([]ScanResult, error)
}

// ScanPipeline 扫描流水线配置
type ScanPipeline struct {
	Name        string        `json:"name"`
	ProjectID   uint          `json:"project_id"`
	Target      string        `json:"target"`
	Stages      []ScanStage   `json:"stages"`
	Parallel    bool          `json:"parallel"`    // 是否并行执行
	ContinueOnError bool      `json:"continue_on_error"` // 出错时是否继续
}

// ScanStage 扫描阶段
type ScanStage struct {
	Name         string            `json:"name"`
	ScannerNames []string          `json:"scanner_names"` // 使用的扫描工具
	Options      map[string]string `json:"options"`       // 阶段特定选项
	Parallel     bool              `json:"parallel"`      // 本阶段内是否并行
	DependsOn    []string          `json:"depends_on"`    // 依赖的前置阶段
}