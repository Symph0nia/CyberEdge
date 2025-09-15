package dao

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// TargetDAO 目标数据访问对象
type TargetDAO struct {
	*BaseDAO
}

// NewTargetDAO 创建目标DAO
func NewTargetDAO(db *gorm.DB) *TargetDAO {
	return &TargetDAO{
		BaseDAO: NewBaseDAO(db),
	}
}

// Create 创建目标
func (d *TargetDAO) Create(target *models.Target) error {
	return d.db.Create(target).Error
}

// GetByID 根据ID获取目标
func (d *TargetDAO) GetByID(id uint) (*models.Target, error) {
	var target models.Target
	err := d.db.First(&target, id).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

// GetByIDWithRelations 获取目标及其关联数据
func (d *TargetDAO) GetByIDWithRelations(id uint) (*models.Target, error) {
	var target models.Target
	err := d.db.Preload("Tasks").
		Preload("Subdomains").
		Preload("Ports").
		Preload("Paths").
		First(&target, id).Error
	if err != nil {
		return nil, err
	}
	return &target, nil
}

// Update 更新目标
func (d *TargetDAO) Update(target *models.Target) error {
	return d.db.Save(target).Error
}

// UpdateStats 更新目标统计信息
func (d *TargetDAO) UpdateStats(id uint, stats map[string]int) error {
	return d.db.Model(&models.Target{}).Where("id = ?", id).Updates(stats).Error
}

// GetAll 获取所有目标
func (d *TargetDAO) GetAll() ([]*models.Target, error) {
	var targets []*models.Target
	err := d.db.Order("created_at DESC").Find(&targets).Error
	return targets, err
}

// GetByType 根据类型获取目标
func (d *TargetDAO) GetByType(targetType models.TargetType) ([]*models.Target, error) {
	var targets []*models.Target
	err := d.db.Where("type = ?", targetType).
		Order("created_at DESC").Find(&targets).Error
	return targets, err
}

// GetByStatus 根据状态获取目标
func (d *TargetDAO) GetByStatus(status models.TargetStatus) ([]*models.Target, error) {
	var targets []*models.Target
	err := d.db.Where("status = ?", status).
		Order("created_at DESC").Find(&targets).Error
	return targets, err
}

// Search 搜索目标
func (d *TargetDAO) Search(keyword string) ([]*models.Target, error) {
	var targets []*models.Target
	err := d.db.Where("name LIKE ? OR target LIKE ? OR description LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Order("created_at DESC").Find(&targets).Error
	return targets, err
}

// Delete 删除目标
func (d *TargetDAO) Delete(id uint) error {
	// GORM会自动处理CASCADE删除
	return d.db.Delete(&models.Target{}, id).Error
}

// Count 获取目标总数
func (d *TargetDAO) Count() (int64, error) {
	var count int64
	err := d.db.Model(&models.Target{}).Count(&count).Error
	return count, err
}

// GetStats 获取目标统计信息
func (d *TargetDAO) GetStats(id uint) (*models.TargetStats, error) {
	var stats models.TargetStats

	// 获取基本统计
	target, err := d.GetByID(id)
	if err != nil {
		return nil, err
	}

	stats.SubdomainCount = target.SubdomainCount
	stats.PortCount = target.PortCount
	stats.PathCount = target.PathCount
	stats.VulnerabilityCount = target.VulnerabilityCount

	// 获取端口统计
	var portStats []models.PortStat
	err = d.db.Model(&models.Port{}).
		Select("port, COUNT(*) as count").
		Where("target_id = ?", id).
		Group("port").
		Order("count DESC").
		Limit(10).
		Scan(&portStats).Error
	if err == nil {
		stats.TopPorts = portStats
	}

	// 获取HTTP状态码统计
	var httpStats []models.HTTPStatusStat
	err = d.db.Model(&models.Path{}).
		Select("status_code as status, COUNT(*) as count").
		Where("target_id = ?", id).
		Group("status_code").
		Order("count DESC").
		Scan(&httpStats).Error
	if err == nil {
		// 添加状态码标签
		for i := range httpStats {
			httpStats[i].Label = getStatusLabel(httpStats[i].Status)
		}
		stats.HTTPStatusStats = httpStats
	}

	return &stats, nil
}

// RecalculateStats 重新计算目标统计信息
func (d *TargetDAO) RecalculateStats(id uint) error {
	var counts struct {
		SubdomainCount int64
		PortCount      int64
		PathCount      int64
	}

	// 统计子域名数量
	d.db.Model(&models.Subdomain{}).Where("target_id = ?", id).Count(&counts.SubdomainCount)

	// 统计端口数量
	d.db.Model(&models.Port{}).Where("target_id = ?", id).Count(&counts.PortCount)

	// 统计路径数量
	d.db.Model(&models.Path{}).Where("target_id = ?", id).Count(&counts.PathCount)

	// 更新统计信息
	return d.db.Model(&models.Target{}).Where("id = ?", id).Updates(map[string]interface{}{
		"subdomain_count": counts.SubdomainCount,
		"port_count":      counts.PortCount,
		"path_count":      counts.PathCount,
	}).Error
}

// 辅助函数：获取HTTP状态码标签
func getStatusLabel(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "成功"
	case status >= 300 && status < 400:
		return "重定向"
	case status >= 400 && status < 500:
		return "客户端错误"
	case status >= 500:
		return "服务器错误"
	default:
		return "其他"
	}
}