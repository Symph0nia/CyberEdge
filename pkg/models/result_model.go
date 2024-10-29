package models

import (
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Result 表示一次完整的扫描结果
type Result struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Targets []*Target          `json:"Targets"`
}

// Target 表示一个扫描目标，可以是子域名或IP地址
type Target struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Identifier  string             `json:"Identifier"`            // 子域名或IP地址
	Type        string             `json:"Type"`                  // "Subdomain" 或 "IP"
	Subdomains  []*Subdomain       `json:"Subdomains,omitempty"`  // 子域名列表
	IPAddresses []*IPAddress       `json:"IPAddresses,omitempty"` // IP地址列表
	LastScan    time.Time          `json:"LastScan"`
}

// Subdomain 表示一个子域名及其相关信息
type Subdomain struct {
	ID    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name  string             `json:"Name"`
	Ports []*Port            `json:"Ports,omitempty"` // 扫描到的端口信息
}

// IPAddress 表示一个IP地址及其相关信息
type IPAddress struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Address net.IP             `json:"Address"`
	Ports   []*Port            `json:"Ports,omitempty"` // 扫描到的端口信息
}

// Port 表示一个开放端口及其相关信息
type Port struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Number       int                `json:"Number"`
	Protocol     string             `json:"Protocol"`
	Service      string             `json:"Service"`
	Banner       string             `json:"Banner"`
	Fingerprints []*Fingerprint     `json:"Fingerprints,omitempty"` // 指纹信息
	Paths        []*Path            `json:"Paths,omitempty"`        // URL路径信息
}

// Fingerprint 表示在端口上识别到的服务或应用的特征
type Fingerprint struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Type    string             `json:"Type"`
	Name    string             `json:"Name"`
	Version string             `json:"Version"`
}

// Path 表示在Web服务中发现的URL路径
type Path struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	URL         string             `json:"URL"`
	StatusCode  int                `json:"StatusCode"`
	ContentType string             `json:"ContentType"`
	Size        int64              `json:"Size"`
}
