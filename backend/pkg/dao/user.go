package dao

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// UserDAO 用户数据访问对象
type UserDAO struct {
	*BaseDAO
}

// NewUserDAO 创建用户DAO
func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建用户
func (d *UserDAO) Create(user *models.User) error {
	return d.db.Create(user).Error
}

// GetByUsername 根据用户名获取用户
func (d *UserDAO) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := d.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail 根据邮箱获取用户
func (d *UserDAO) GetByEmail(email string) (*models.User, error) {
	var user models.User
	err := d.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByID 根据ID获取用户
func (d *UserDAO) GetByID(id uint) (*models.User, error) {
	var user models.User
	err := d.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (d *UserDAO) Update(user *models.User) error {
	return d.db.Save(user).Error
}

// Delete 删除用户
func (d *UserDAO) Delete(id uint) error {
	return d.db.Delete(&models.User{}, id).Error
}

// GetAll 获取所有用户
func (d *UserDAO) GetAll() ([]*models.User, error) {
	var users []*models.User
	err := d.db.Find(&users).Error
	return users, err
}

// Count 获取用户总数
func (d *UserDAO) Count() (int64, error) {
	var count int64
	err := d.db.Model(&models.User{}).Count(&count).Error
	return count, err
}