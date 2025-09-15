package models

import (
	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	Is2FAEnabled bool   `gorm:"default:false" json:"is_2fa_enabled"`
	TOTPSecret   string `gorm:"size:32" json:"-"`
	Role         string `gorm:"type:enum('admin','user');default:'user'" json:"role"`
	CreatedAt    int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

// TOTPValidationRequest TOTP验证请求
type TOTPValidationRequest struct {
	Code    string `json:"code" binding:"required"`
	Account string `json:"account" binding:"required"`
}

// TOTPValidationResponse TOTP验证响应
type TOTPValidationResponse struct {
	Status     string `json:"status"`
	Token      string `json:"token"`
	LoginCount int    `json:"login_count"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// BeforeCreate 创建前钩子
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return nil
}

// BeforeUpdate 更新前钩子
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	return nil
}