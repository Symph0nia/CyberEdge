package models

import (
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Result 表示一次扫描结果
type Result struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TargetID  *primitive.ObjectID `json:"target_id,omitempty" bson:"target_id,omitempty"`
	Type      string              `json:"type" bson:"type"`
	Target    string              `json:"target" bson:"target"`
	Timestamp time.Time           `json:"timestamp" bson:"timestamp"`
	Data      interface{}         `json:"data" bson:"data"`
	IsRead    bool                `json:"is_read" bson:"is_read"`
}

// SubdomainData 子域名扫描结果
type SubdomainData struct {
	Subdomains []SubdomainEntry `json:"subdomains" bson:"subdomains"`
}

// 修改各个条目的结构，添加目标归属字段
type SubdomainEntry struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TargetID   *primitive.ObjectID `json:"target_id,omitempty" bson:"target_id,omitempty"`
	Domain     string              `json:"domain" bson:"domain"`
	IP         string              `json:"ip" bson:"ip"`
	IsRead     bool                `json:"is_read" bson:"is_read"`
	HTTPStatus int                 `json:"http_status" bson:"http_status"`
	HTTPTitle  string              `json:"http_title" bson:"http_title"`
}

// PortData 端口扫描结果
type PortData struct {
	Ports []PortEntry `json:"ports" bson:"ports"`
}

// PortEntry 端口条目
type PortEntry struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TargetID   *primitive.ObjectID `json:"target_id,omitempty" bson:"target_id,omitempty"`
	Host       string              `json:"host" bson:"host"`
	Number     int                 `json:"number" bson:"number"`
	Protocol   string              `json:"protocol" bson:"protocol"`
	State      string              `json:"state" bson:"state"`
	Service    string              `json:"service" bson:"service"`
	IsRead     bool                `json:"is_read" bson:"is_read"`
	HTTPStatus int                 `json:"http_status" bson:"http_status"`
	HTTPTitle  string              `json:"http_title" bson:"http_title"`
}

// PathData 路径扫描结果
type PathData struct {
	Paths []PathEntry `json:"paths" bson:"paths"`
}

// PathEntry 路径条目
type PathEntry struct {
	ID         primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	TargetID   *primitive.ObjectID `json:"target_id,omitempty" bson:"target_id,omitempty"`
	Path       string              `json:"path" bson:"path"`
	Status     int                 `json:"status" bson:"status"`
	Length     int                 `json:"length" bson:"length"`
	Words      int                 `json:"words" bson:"words"`
	Lines      int                 `json:"lines" bson:"lines"`
	IsRead     bool                `json:"is_read" bson:"is_read"`
	HTTPStatus int                 `json:"http_status" bson:"http_status"`
	HTTPTitle  string              `json:"http_title" bson:"http_title"`
}

// 为 SubdomainEntry 实现接口
func (e SubdomainEntry) GetID() string {
	return e.ID.Hex()
}

func (e SubdomainEntry) GetProbeURL() string {
	return e.Domain
}

// 为 PortEntry 实现接口
func (e PortEntry) GetID() string {
	return e.ID.Hex()
}

func (e PortEntry) GetProbeURL() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Number) // 这里需要在 PortEntry 中添加 Host 字段
}

// 为 PathEntry 实现接口
func (e PathEntry) GetID() string {
	return e.ID.Hex()
}

func (e PathEntry) GetProbeURL() string {
	return e.Path
}
