package dao

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// SubdomainDAO 子域名数据访问对象
type SubdomainDAO struct {
	*BaseDAO
}

// NewSubdomainDAO 创建子域名DAO
func NewSubdomainDAO(db *gorm.DB) *SubdomainDAO {
	return &SubdomainDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建子域名
func (d *SubdomainDAO) Create(subdomain *models.Subdomain) error {
	return d.db.Create(subdomain).Error
}

// CreateBatch 批量创建子域名
func (d *SubdomainDAO) CreateBatch(subdomains []*models.Subdomain) error {
	if len(subdomains) == 0 {
		return nil
	}
	return d.db.Create(&subdomains).Error
}

// GetByTargetID 根据目标ID获取子域名
func (d *SubdomainDAO) GetByTargetID(targetID uint) ([]*models.Subdomain, error) {
	var subdomains []*models.Subdomain
	err := d.db.Where("target_id = ?", targetID).
		Order("created_at DESC").Find(&subdomains).Error
	return subdomains, err
}

// DeleteByTargetID 删除目标相关的所有子域名
func (d *SubdomainDAO) DeleteByTargetID(targetID uint) error {
	return d.db.Where("target_id = ?", targetID).Delete(&models.Subdomain{}).Error
}

// PortDAO 端口数据访问对象
type PortDAO struct {
	*BaseDAO
}

// NewPortDAO 创建端口DAO
func NewPortDAO(db *gorm.DB) *PortDAO {
	return &PortDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建端口
func (d *PortDAO) Create(port *models.Port) error {
	return d.db.Create(port).Error
}

// CreateBatch 批量创建端口
func (d *PortDAO) CreateBatch(ports []*models.Port) error {
	if len(ports) == 0 {
		return nil
	}
	return d.db.Create(&ports).Error
}

// GetByTargetID 根据目标ID获取端口
func (d *PortDAO) GetByTargetID(targetID uint) ([]*models.Port, error) {
	var ports []*models.Port
	err := d.db.Where("target_id = ?", targetID).
		Order("port ASC").Find(&ports).Error
	return ports, err
}

// DeleteByTargetID 删除目标相关的所有端口
func (d *PortDAO) DeleteByTargetID(targetID uint) error {
	return d.db.Where("target_id = ?", targetID).Delete(&models.Port{}).Error
}

// PathDAO 路径数据访问对象
type PathDAO struct {
	*BaseDAO
}

// NewPathDAO 创建路径DAO
func NewPathDAO(db *gorm.DB) *PathDAO {
	return &PathDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建路径
func (d *PathDAO) Create(path *models.Path) error {
	return d.db.Create(path).Error
}

// CreateBatch 批量创建路径
func (d *PathDAO) CreateBatch(paths []*models.Path) error {
	if len(paths) == 0 {
		return nil
	}
	return d.db.Create(&paths).Error
}

// GetByTargetID 根据目标ID获取路径
func (d *PathDAO) GetByTargetID(targetID uint) ([]*models.Path, error) {
	var paths []*models.Path
	err := d.db.Where("target_id = ?", targetID).
		Order("status_code ASC").Find(&paths).Error
	return paths, err
}

// DeleteByTargetID 删除目标相关的所有路径
func (d *PathDAO) DeleteByTargetID(targetID uint) error {
	return d.db.Where("target_id = ?", targetID).Delete(&models.Path{}).Error
}