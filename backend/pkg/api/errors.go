package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ErrorCode 错误代码
type ErrorCode int

const (
	// 认证相关错误
	ErrInvalidCredentials ErrorCode = 1001
	ErrUnauthorized      ErrorCode = 1002
	ErrInvalidToken      ErrorCode = 1003

	// 用户相关错误
	ErrUserNotFound     ErrorCode = 2001
	ErrUserExists       ErrorCode = 2002
	ErrEmailExists      ErrorCode = 2003
	ErrWeakPassword     ErrorCode = 2004

	// 2FA相关错误
	Err2FANotSetup      ErrorCode = 3001
	ErrInvalid2FACode   ErrorCode = 3002

	// 通用错误
	ErrInvalidRequest   ErrorCode = 4001
	ErrInternalServer   ErrorCode = 5001
)

// ErrorResponse 统一错误响应格式
type ErrorResponse struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Success bool      `json:"success"`
}

// getErrorMessage 获取用户友好的错误消息
func getErrorMessage(code ErrorCode) string {
	switch code {
	case ErrInvalidCredentials:
		return "用户名或密码错误"
	case ErrUnauthorized:
		return "未授权访问"
	case ErrInvalidToken:
		return "认证令牌无效"
	case ErrUserNotFound:
		return "用户不存在"
	case ErrUserExists:
		return "用户名已存在"
	case ErrEmailExists:
		return "邮箱已被使用"
	case ErrWeakPassword:
		return "密码强度不足"
	case Err2FANotSetup:
		return "请先设置双因子认证"
	case ErrInvalid2FACode:
		return "验证码无效"
	case ErrInvalidRequest:
		return "请求参数无效"
	case ErrInternalServer:
		return "服务器内部错误"
	default:
		return "未知错误"
	}
}

// RespondWithError 统一错误响应
func RespondWithError(c *gin.Context, statusCode int, errorCode ErrorCode) {
	c.JSON(statusCode, ErrorResponse{
		Code:    errorCode,
		Message: getErrorMessage(errorCode),
		Success: false,
	})
}

// RespondWithSuccess 统一成功响应
func RespondWithSuccess(c *gin.Context, data interface{}) {
	response := gin.H{
		"success": true,
	}

	if data != nil {
		// 如果data是map，则合并到response中
		if dataMap, ok := data.(map[string]interface{}); ok {
			for k, v := range dataMap {
				response[k] = v
			}
		} else {
			response["data"] = data
		}
	}

	c.JSON(http.StatusOK, response)
}