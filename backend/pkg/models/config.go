package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
)

// ToolConfig 工具配置模型
type ToolConfig struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:255;not null" json:"name"`
	IsDefault bool       `gorm:"index" json:"is_default"`
	Config    ConfigJSON `gorm:"type:json;not null" json:"config"`
	CreatedAt int64      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt int64      `gorm:"autoUpdateTime" json:"updated_at"`
}

// ConfigJSON 自定义JSON类型，用于处理复杂配置
type ConfigJSON map[string]interface{}

// Value 实现driver.Valuer接口
func (c ConfigJSON) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan 实现sql.Scanner接口
func (c *ConfigJSON) Scan(value interface{}) error {
	if value == nil {
		*c = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ConfigJSON", value)
	}

	return json.Unmarshal(bytes, c)
}

// ToolConfigDetail 详细的工具配置结构（用于类型安全的配置操作）
type ToolConfigDetail struct {
	ID        uint                `json:"id"`
	Name      string              `json:"name"`
	IsDefault bool                `json:"is_default"`
	Nmap      NmapConfig          `json:"nmap"`
	Ffuf      FfufConfig          `json:"ffuf"`
	Subfinder SubfinderConfig     `json:"subfinder"`
	CreatedAt int64               `json:"created_at"`
	UpdatedAt int64               `json:"updated_at"`
}

// NmapConfig Nmap配置
type NmapConfig struct {
	Enabled     bool   `json:"enabled"`
	Ports       string `json:"ports"`
	Timeout     int    `json:"timeout"`
	Concurrency int    `json:"concurrency"`
}

// FfufConfig Ffuf配置
type FfufConfig struct {
	Enabled    bool   `json:"enabled"`
	Wordlist   string `json:"wordlist"`
	Extensions string `json:"extensions"`
	Threads    int    `json:"threads"`
}

// SubfinderConfig Subfinder配置
type SubfinderConfig struct {
	Enabled bool `json:"enabled"`
	Threads int  `json:"threads"`
	Timeout int  `json:"timeout"`
}

// TableName 指定表名
func (ToolConfig) TableName() string {
	return "tool_configs"
}

// BeforeCreate 创建前钩子
func (t *ToolConfig) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前钩子
func (t *ToolConfig) BeforeUpdate(tx *gorm.DB) error {
	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *ToolConfig {
	defaultConfig := ConfigJSON{
		"nmap": map[string]interface{}{
			"enabled":     true,
			"ports":       "21,22,23,25,53,80,110,135,139,143,443,445,993,995,1433,3306,3389,5432,6379,8080",
			"timeout":     300,
			"concurrency": 100,
		},
		"ffuf": map[string]interface{}{
			"enabled":    true,
			"wordlist":   "/usr/share/wordlists/dirb/common.txt",
			"extensions": "php,asp,jsp,html,js",
			"threads":    50,
		},
		"subfinder": map[string]interface{}{
			"enabled": true,
			"threads": 10,
			"timeout": 60,
		},
	}

	return &ToolConfig{
		Name:      "默认配置",
		IsDefault: true,
		Config:    defaultConfig,
	}
}

// ToDetail 转换为详细配置结构
func (t *ToolConfig) ToDetail() (*ToolConfigDetail, error) {
	detail := &ToolConfigDetail{
		ID:        t.ID,
		Name:      t.Name,
		IsDefault: t.IsDefault,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	// 解析nmap配置
	if nmapData, ok := t.Config["nmap"].(map[string]interface{}); ok {
		detail.Nmap = NmapConfig{
			Enabled:     getBool(nmapData, "enabled"),
			Ports:       getString(nmapData, "ports"),
			Timeout:     getInt(nmapData, "timeout"),
			Concurrency: getInt(nmapData, "concurrency"),
		}
	}

	// 解析ffuf配置
	if ffufData, ok := t.Config["ffuf"].(map[string]interface{}); ok {
		detail.Ffuf = FfufConfig{
			Enabled:    getBool(ffufData, "enabled"),
			Wordlist:   getString(ffufData, "wordlist"),
			Extensions: getString(ffufData, "extensions"),
			Threads:    getInt(ffufData, "threads"),
		}
	}

	// 解析subfinder配置
	if subfinderData, ok := t.Config["subfinder"].(map[string]interface{}); ok {
		detail.Subfinder = SubfinderConfig{
			Enabled: getBool(subfinderData, "enabled"),
			Threads: getInt(subfinderData, "threads"),
			Timeout: getInt(subfinderData, "timeout"),
		}
	}

	return detail, nil
}

// 辅助函数
func getBool(data map[string]interface{}, key string) bool {
	if val, ok := data[key].(bool); ok {
		return val
	}
	return false
}

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	if val, ok := data[key].(int); ok {
		return val
	}
	return 0
}