package models

import (
	"net"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Result 表示一次扫描结果，使用 Type 字段区分不同类型
type Result struct {
	ID        primitive.ObjectID  `json:"id" bson:"_id,omitempty"`
	ParentID  *primitive.ObjectID `json:"ParentID,omitempty" bson:"ParentID,omitempty"` // 上级 ID，可为空
	Type      string              `json:"Type"`                                         // "Subdomain", "IP", "Port" 等
	Target    string              `json:"Target"`                                       // 扫描目标，如域名或 IP 地址
	Timestamp time.Time           `json:"Timestamp"`                                    // 扫描时间
	Data      interface{}         `json:"Data"`                                         // 存储具体的扫描数据
	IsRead    bool                `json:"IsRead" bson:"is_read"`                        // 是否已读，默认未读
}

// SubdomainData 表示子域名扫描结果的数据结构
type SubdomainData struct {
	Subdomains []SubdomainEntry `json:"Subdomains"`
}

// SubdomainEntry 表示每个子域名的条目
type SubdomainEntry struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Domain     string             `json:"domain" bson:"domain"`
	IP         string             `json:"ip" bson:"ip"`
	IsRead     bool               `json:"is_read" bson:"is_read"`
	HTTPStatus int                `json:"http_status" bson:"http_status"` // 注意字段名变化
	HTTPTitle  string             `json:"http_title" bson:"http_title"`   // 注意字段名变化
}

// IPAddressData 表示 IP 地址扫描结果的数据结构
type IPAddressData struct {
	IPAddresses []net.IP `json:"IPAddresses"`
}

// PortData 表示端口扫描结果的数据结构
type PortData struct {
	Ports []*Port `json:"Ports"`
}

// Port 表示一个端口及其相关信息
type Port struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"` // 唯一标识符
	Number    int                `json:"number" bson:"number"`
	Protocol  string             `json:"protocol" bson:"protocol"`
	State     string             `json:"state" bson:"state"`                    // 端口状态（如 open, closed, filtered）
	Service   string             `json:"service" bson:"service"`                // 服务名称
	Version   string             `json:"version" bson:"version,omitempty"`      // 服务版本
	Banner    string             `json:"banner" bson:"banner,omitempty"`        // 服务横幅信息
	Product   string             `json:"product" bson:"product,omitempty"`      // 产品名称
	ExtraInfo string             `json:"extraInfo" bson:"extra_info,omitempty"` // 额外信息
	IsRead    bool               `json:"isRead" bson:"is_read"`                 // 是否已读，默认未读
}

// Fingerprint 表示在端口上识别到的服务或应用的特征
type Fingerprint struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"` // 唯一标识符
	Type    string             `json:"Type"`
	Name    string             `json:"Name"`
	Version string             `json:"Version"`
	IsRead  bool               `json:"IsRead" bson:"is_read"` // 是否已读，默认未读
}

// PathEntry 表示单个路径扫描结果
type PathEntry struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Path   string             `bson:"path" json:"path"`
	Status int                `bson:"status" json:"status"`
	Length int                `bson:"length" json:"length"`
	Words  int                `bson:"words" json:"words"`
	Lines  int                `bson:"lines" json:"lines"`
	IsRead bool               `bson:"is_read" json:"is_read"`
}

// PathData 表示路径扫描结果的数据结构
type PathData struct {
	Paths []*Path `json:"Paths" bson:"paths"`
}

// Path 表示一个路径及其相关信息
type Path struct {
	ID     primitive.ObjectID `json:"id" bson:"_id,omitempty"` // 唯一标识符
	Path   string             `json:"path" bson:"path"`
	Status int                `json:"status" bson:"status"`
	Length int                `json:"length" bson:"length"`
	Words  int                `json:"words" bson:"words"`
	Lines  int                `json:"lines" bson:"lines"`
	IsRead bool               `json:"isRead" bson:"is_read"` // 是否已读，默认未读
}
