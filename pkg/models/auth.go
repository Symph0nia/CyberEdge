package models

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
