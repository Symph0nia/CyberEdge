package service

import "cyberedge/pkg/models"

// UserServiceInterface 用户服务接口
type UserServiceInterface interface {
	CreateUser(username, email, password string) error
	Login(username, password string) (string, error)
	ValidateToken(token string) (*models.User, error)
	GetAllUsers() ([]*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	DeleteUser(id uint) error
	ChangePassword(username, currentPassword, newPassword string) error
	Setup2FA(username string) (string, []byte, error)
	Verify2FA(username, code string) error
	Disable2FA(username string) error
	ValidatePassword(password string) error
}