package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// ToolConfig 工具配置主结构体
type ToolConfig struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`             // 配置名称
	IsDefault bool               `json:"is_default" bson:"is_default"` // 是否为默认配置
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`

	// 各工具配置
	NmapConfig      NmapConfig      `json:"nmap_config" bson:"nmap_config"`
	FfufConfig      FfufConfig      `json:"ffuf_config" bson:"ffuf_config"`
	SubfinderConfig SubfinderConfig `json:"subfinder_config" bson:"subfinder_config"`
	HttpxConfig     HttpxConfig     `json:"httpx_config" bson:"httpx_config"`
	FscanConfig     FscanConfig     `json:"fscan_config" bson:"fscan_config"`
	AfrogConfig     AfrogConfig     `json:"afrog_config" bson:"afrog_config"`
	NucleiConfig    NucleiConfig    `json:"nuclei_config" bson:"nuclei_config"`
}

// NmapConfig Nmap工具配置
type NmapConfig struct {
	Enabled     bool   `json:"enabled" bson:"enabled"`           // 是否启用
	Ports       string `json:"ports" bson:"ports"`               // 端口范围，如 "80,443,8080-8090"
	ScanTimeout int    `json:"scan_timeout" bson:"scan_timeout"` // 扫描超时时间(秒)
	Concurrency int    `json:"concurrency" bson:"concurrency"`   // 并发数
}

// FfufConfig Ffuf工具配置
type FfufConfig struct {
	Enabled       bool   `json:"enabled" bson:"enabled"`                 // 是否启用
	WordlistPath  string `json:"wordlist_path" bson:"wordlist_path"`     // 字典文件路径
	Extensions    string `json:"extensions" bson:"extensions"`           // 扩展名，如 "php,asp,aspx"
	Threads       int    `json:"threads" bson:"threads"`                 // 线程数
	MatchHttpCode string `json:"match_http_code" bson:"match_http_code"` // 匹配的HTTP状态码
}

// SubfinderConfig Subfinder工具配置
type SubfinderConfig struct {
	Enabled    bool   `json:"enabled" bson:"enabled"`         // 是否启用
	ConfigPath string `json:"config_path" bson:"config_path"` // 配置文件路径
	Threads    int    `json:"threads" bson:"threads"`         // 线程数
	Timeout    int    `json:"timeout" bson:"timeout"`         // 超时时间(秒)
}

// HttpxConfig HttpX工具配置 (虽然不做配置，但为了完整性仍定义结构体)
type HttpxConfig struct {
	Enabled bool `json:"enabled" bson:"enabled"` // 是否启用
	Threads int  `json:"threads" bson:"threads"` // 线程数
	Timeout int  `json:"timeout" bson:"timeout"` // 超时时间(秒)
}

// FscanConfig Fscan工具配置 (虽然不做配置，但为了完整性仍定义结构体)
type FscanConfig struct {
	Enabled bool `json:"enabled" bson:"enabled"` // 是否启用
	Threads int  `json:"threads" bson:"threads"` // 线程数
}

// AfrogConfig Afrog工具配置 (虽然不做配置，但为了完整性仍定义结构体)
type AfrogConfig struct {
	Enabled bool `json:"enabled" bson:"enabled"` // 是否启用
	Threads int  `json:"threads" bson:"threads"` // 线程数
}

// NucleiConfig Nuclei工具配置 (虽然不做配置，但为了完整性仍定义结构体)
type NucleiConfig struct {
	Enabled bool `json:"enabled" bson:"enabled"` // 是否启用
	Threads int  `json:"threads" bson:"threads"` // 线程数
}

// BeforeCreate 创建前的钩子
func (t *ToolConfig) BeforeCreate() {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
}

// BeforeUpdate 更新前的钩子
func (t *ToolConfig) BeforeUpdate() {
	t.UpdatedAt = time.Now()
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *ToolConfig {
	return &ToolConfig{
		Name:      "默认配置",
		IsDefault: true,
		NmapConfig: NmapConfig{
			Enabled:     true,
			Ports:       "21,22,23,25,53,80,110,111,135,139,143,443,445,465,587,993,995,1080,1433,1521,3306,3389,5432,5900,6379,8080,8443",
			ScanTimeout: 300,
			Concurrency: 100,
		},
		FfufConfig: FfufConfig{
			Enabled:       true,
			WordlistPath:  "/usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt",
			Extensions:    "php,asp,aspx,jsp,html,js",
			Threads:       50,
			MatchHttpCode: "200,204,301,302,307,401,403",
		},
		SubfinderConfig: SubfinderConfig{
			Enabled:    true,
			ConfigPath: "/etc/subfinder/config.yaml",
			Threads:    10,
			Timeout:    60,
		},
		HttpxConfig: HttpxConfig{
			Enabled: true,
			Threads: 50,
			Timeout: 10,
		},
		FscanConfig: FscanConfig{
			Enabled: true,
			Threads: 100,
		},
		AfrogConfig: AfrogConfig{
			Enabled: true,
			Threads: 50,
		},
		NucleiConfig: NucleiConfig{
			Enabled: true,
			Threads: 50,
		},
	}
}
