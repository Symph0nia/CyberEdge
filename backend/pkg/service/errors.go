package service

import "errors"

// 用户相关的错误定义
var (
	ErrUserExists         = errors.New("USER_EXISTS")
	ErrEmailExists        = errors.New("EMAIL_EXISTS")
	ErrInvalidCredentials = errors.New("INVALID_CREDENTIALS")
	ErrUserNotFound       = errors.New("USER_NOT_FOUND")
	ErrWeakPassword       = errors.New("WEAK_PASSWORD")
	ErrInvalidPassword    = errors.New("INVALID_PASSWORD")
)