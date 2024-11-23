// CyberEdge/pkg/models/auth.go

package models

// User 代表用户的结构体
type User struct {
	Account    string `bson:"account" json:"account"`
	Secret     string `bson:"secret" json:"-"`
	LoginCount int    `bson:"loginCount" json:"loginCount"`
}

// TOTPValidationRequest 定义了TOTP验证请求的结构
type TOTPValidationRequest struct {
	Code    string `json:"code" binding:"required"`
	Account string `json:"account" binding:"required"`
}

// TOTPValidationResponse 定义了TOTP验证响应的结构
type TOTPValidationResponse struct {
	Status     string `json:"status"`
	Token      string `json:"token"`
	LoginCount int    `json:"loginCount"`
}
