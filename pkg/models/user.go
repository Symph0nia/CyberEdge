// CyberEdge/pkg/models/auth.go

package models

// User 代表用户的结构体
type User struct {
	Account    string `bson:"account" json:"account"`
	Secret     string `bson:"secret" json:"-"`
	LoginCount int    `bson:"loginCount" json:"loginCount"`
}
