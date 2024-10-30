package models

import (
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Result 表示一次扫描结果，使用 Type 字段区分不同类型
type Result struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	ParentID  *primitive.ObjectID `json:"parent_id,omitempty" bson:"parent_id,omitempty"` // 上级 ID，可为空
	Type      string              `json:"Type"`                                           // "Subdomain", "IP", "Port" 等
	Target    string              `json:"Target"`                                         // 扫描目标，如域名或 IP 地址
	Timestamp time.Time           `json:"Timestamp"`                                      // 扫描时间
	Data      interface{}         `json:"Data"`                                           // 存储具体的扫描数据
}

// SubdomainData 表示子域名扫描结果的数据结构
type SubdomainData struct {
	Subdomains []string `json:"Subdomains"`
}

// IPAddressData 表示 IP 地址扫描结果的数据结构
type IPAddressData struct {
	IPAddresses []net.IP `json:"IPAddresses"`
}

// PortData 表示端口扫描结果的数据结构
type PortData struct {
	Ports []*Port `json:"Ports"`
}

// Port 表示一个开放端口及其相关信息
type Port struct {
	Number       int            `json:"Number"`
	Protocol     string         `json:"Protocol"`
	Service      string         `json:"Service"`
	Banner       string         `json:"Banner"`
	Fingerprints []*Fingerprint `json:"Fingerprints,omitempty"` // 指纹信息
	Paths        []*Path        `json:"Paths,omitempty"`        // URL路径信息
}

// Fingerprint 表示在端口上识别到的服务或应用的特征
type Fingerprint struct {
	Type    string `json:"Type"`
	Name    string `json:"Name"`
	Version string `json:"Version"`
}

// Path 表示在 Web 服务中发现的 URL 路径
type Path struct {
	URL         string `json:"URL"`
	StatusCode  int    `json:"StatusCode"`
	ContentType string `json:"ContentType"`
	Size        int64  `json:"Size"`
}
