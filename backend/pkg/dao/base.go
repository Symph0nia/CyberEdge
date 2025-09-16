package dao

import (
	"gorm.io/gorm"
)

// BaseDAO 基础DAO，包含通用的数据库操作
type BaseDAO struct {
	db *gorm.DB
}

// NewBaseDAO 创建基础DAO
func NewBaseDAO(db *gorm.DB) *BaseDAO {
	return &BaseDAO{db: db}
}

// GetDB 获取数据库连接
func (d *BaseDAO) GetDB() *gorm.DB {
	return d.db
}

// Transaction 执行事务
func (d *BaseDAO) Transaction(fn func(*gorm.DB) error) error {
	return d.db.Transaction(fn)
}