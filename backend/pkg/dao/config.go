package dao

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// ConfigDAO 配置数据访问对象
type ConfigDAO struct {
	*BaseDAO
}

// NewConfigDAO 创建配置DAO
func NewConfigDAO(db *gorm.DB) *ConfigDAO {
	return &ConfigDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建配置
func (d *ConfigDAO) Create(config *models.ToolConfig) error {
	if config.IsDefault {
		if err := d.ClearDefault(); err != nil {
			return err
		}
	}
	return d.db.Create(config).Error
}

// Update 更新配置
func (d *ConfigDAO) Update(config *models.ToolConfig) error {
	if config.IsDefault {
		if err := d.ClearDefault(); err != nil {
			return err
		}
	}
	return d.db.Save(config).Error
}

// GetByID 根据ID获取配置
func (d *ConfigDAO) GetByID(id uint) (*models.ToolConfig, error) {
	var config models.ToolConfig
	err := d.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetDefault 获取默认配置
func (d *ConfigDAO) GetDefault() (*models.ToolConfig, error) {
	var config models.ToolConfig
	err := d.db.Where("is_default = ?", true).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			defaultConfig := models.GetDefaultConfig()
			if err := d.Create(defaultConfig); err != nil {
				return nil, err
			}
			return defaultConfig, nil
		}
		return nil, err
	}
	return &config, nil
}

// GetAll 获取所有配置
func (d *ConfigDAO) GetAll() ([]*models.ToolConfig, error) {
	var configs []*models.ToolConfig
	err := d.db.Order("is_default DESC, created_at DESC").Find(&configs).Error
	return configs, err
}

// Delete 删除配置
func (d *ConfigDAO) Delete(id uint) error {
	var config models.ToolConfig
	if err := d.db.First(&config, id).Error; err != nil {
		return err
	}

	if config.IsDefault {
		return gorm.ErrCheckConstraintViolated
	}

	return d.db.Delete(&config).Error
}

// SetDefault 设置默认配置
func (d *ConfigDAO) SetDefault(id uint) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.ToolConfig{}).
			Where("is_default = ?", true).
			Update("is_default", false).Error; err != nil {
			return err
		}

		return tx.Model(&models.ToolConfig{}).
			Where("id = ?", id).
			Update("is_default", true).Error
	})
}

// ClearDefault 清除所有默认配置标记
func (d *ConfigDAO) ClearDefault() error {
	return d.db.Model(&models.ToolConfig{}).
		Where("is_default = ?", true).
		Update("is_default", false).Error
}

// Count 获取配置总数
func (d *ConfigDAO) Count() (int64, error) {
	var count int64
	err := d.db.Model(&models.ToolConfig{}).Count(&count).Error
	return count, err
}