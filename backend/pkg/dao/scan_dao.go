package dao

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"cyberedge/pkg/models"
)

// ScanDAO - 基于Linus审查建议，消除N+1查询，使用高效的JOIN查询

type ScanDAO struct {
	db *gorm.DB
}

func NewScanDAO(db *gorm.DB) *ScanDAO {
	return &ScanDAO{db: db}
}

// Project 管理 - 保持简单
func (d *ScanDAO) CreateProject(project *models.Project) error {
	return d.db.Create(project).Error
}

func (d *ScanDAO) GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	err := d.db.First(&project, id).Error
	return &project, err
}

func (d *ScanDAO) GetProjectByName(name string) (*models.Project, error) {
	var project models.Project
	err := d.db.Where("name = ?", name).First(&project).Error
	return &project, err
}

func (d *ScanDAO) ListProjects() ([]models.Project, error) {
	var projects []models.Project
	err := d.db.Find(&projects).Error
	return projects, err
}

func (d *ScanDAO) DeleteProject(id uint) error {
	return d.db.Delete(&models.Project{}, id).Error
}

// 高效的项目详情查询 - 单次JOIN代替多次Preload
func (d *ScanDAO) GetProjectDetails(projectID uint) (*models.ProjectStats, []models.ScanTarget, error) {
	// 获取项目基本信息和统计
	stats, err := d.GetProjectStats(projectID)
	if err != nil {
		return nil, nil, err
	}

	// 单次查询获取所有目标和扫描结果
	var targets []models.ScanTarget
	err = d.db.Where("project_id = ?", projectID).
		Order("type, address").
		Find(&targets).Error
	if err != nil {
		return stats, nil, err
	}

	return stats, targets, nil
}

// 高效的统计查询 - 使用原生SQL避免ORM复杂性
func (d *ScanDAO) GetProjectStats(projectID uint) (*models.ProjectStats, error) {
	var stats models.ProjectStats

	// 基本项目信息
	var project models.Project
	if err := d.db.First(&project, projectID).Error; err != nil {
		return nil, err
	}

	stats.ProjectID = projectID
	stats.ProjectName = project.Name

	// 一次查询获取所有统计数据
	var counts struct {
		TargetCount     int `json:"target_count"`
		DomainCount     int `json:"domain_count"`
		IPCount         int `json:"ip_count"`
		PortCount       int `json:"port_count"`
		ServiceCount    int `json:"service_count"`
		WebServiceCount int `json:"web_service_count"`
	}

	err := d.db.Raw(`
		SELECT
			COUNT(DISTINCT t.id) as target_count,
			COUNT(DISTINCT CASE WHEN t.type = 'domain' THEN t.id END) as domain_count,
			COUNT(DISTINCT CASE WHEN t.type = 'ip' THEN t.id END) as ip_count,
			COUNT(DISTINCT sr.id) as port_count,
			COUNT(DISTINCT CASE WHEN sr.service_name != '' THEN sr.id END) as service_count,
			COUNT(DISTINCT CASE WHEN sr.is_web_service = true THEN sr.id END) as web_service_count
		FROM scan_targets t
		LEFT JOIN scan_results sr ON t.id = sr.target_id
		WHERE t.project_id = ?
	`, projectID).Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	stats.TargetCount = counts.TargetCount
	stats.DomainCount = counts.DomainCount
	stats.IPCount = counts.IPCount
	stats.PortCount = counts.PortCount
	stats.ServiceCount = counts.ServiceCount
	stats.WebServiceCount = counts.WebServiceCount

	// 漏洞统计
	vulnStats, err := d.GetVulnerabilityStats(projectID)
	if err != nil {
		return nil, err
	}
	stats.VulnerabilityStats = vulnStats

	// 最后扫描时间
	var lastScan time.Time
	d.db.Raw(`
		SELECT MAX(created_at)
		FROM scan_results sr
		JOIN scan_targets t ON sr.target_id = t.id
		WHERE t.project_id = ?
	`, projectID).Scan(&lastScan)
	stats.LastScanTime = lastScan

	return &stats, nil
}

// 高效的漏洞统计
func (d *ScanDAO) GetVulnerabilityStats(projectID uint) (map[string]int, error) {
	type VulnStat struct {
		Severity string `json:"severity"`
		Count    int    `json:"count"`
	}

	var vulnStats []VulnStat
	err := d.db.Raw(`
		SELECT v.severity, COUNT(*) as count
		FROM vulnerabilities v
		JOIN scan_results sr ON v.scan_result_id = sr.id
		JOIN scan_targets t ON sr.target_id = t.id
		WHERE t.project_id = ? AND v.status = 'open'
		GROUP BY v.severity
	`, projectID).Scan(&vulnStats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, stat := range vulnStats {
		result[stat.Severity] = stat.Count
	}

	// 确保所有严重级别都有值
	severities := []string{"critical", "high", "medium", "low", "info"}
	for _, severity := range severities {
		if _, exists := result[severity]; !exists {
			result[severity] = 0
		}
	}

	return result, nil
}

// 批量导入扫描数据 - 使用事务和批量插入优化性能
func (d *ScanDAO) ImportScanData(data *models.ScanDataImport) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		// 构建目标映射，处理层次关系
		targetMap := make(map[string]*models.ScanTarget)

		// 第一遍：创建所有目标
		for _, targetData := range data.Results {
			target := &models.ScanTarget{
				ProjectID: data.ProjectID,
				Type:      targetData.Type,
				Address:   targetData.Address,
			}

			// 查找或创建目标
			err := tx.Where("project_id = ? AND address = ?", data.ProjectID, targetData.Address).
				FirstOrCreate(target).Error
			if err != nil {
				return fmt.Errorf("创建目标失败 %s: %w", targetData.Address, err)
			}

			targetMap[targetData.Address] = target
		}

		// 第二遍：设置父子关系
		for _, targetData := range data.Results {
			if targetData.Parent != "" {
				target := targetMap[targetData.Address]
				parent := targetMap[targetData.Parent]
				if parent != nil {
					target.ParentID = &parent.ID
					tx.Save(target)
				}
			}
		}

		// 第三遍：批量创建扫描结果
		var vulnerabilities []models.Vulnerability
		var scanResultTechs []models.ScanResultTechnology

		for _, targetData := range data.Results {
			target := targetMap[targetData.Address]

			for _, portData := range targetData.Ports {
				scanResult := models.ScanResult{
					ProjectID:   data.ProjectID,
					TargetID:    target.ID,
					Port:        portData.Number,
					Protocol:    portData.Protocol,
					State:       portData.State,
				}

				if portData.Service != nil {
					service := portData.Service
					scanResult.ServiceName = service.Name
					scanResult.Version = service.Version
					scanResult.Fingerprint = service.Fingerprint
					scanResult.Banner = service.Banner
					scanResult.IsWebService = service.IsWebService
					scanResult.HTTPTitle = service.HTTPTitle
					scanResult.HTTPStatus = service.HTTPStatus
				}

				// 先创建扫描结果以获取ID
				if err := tx.Create(&scanResult).Error; err != nil {
					return fmt.Errorf("创建扫描结果失败: %w", err)
				}

				// 处理漏洞
				if portData.Service != nil {
					for _, vulnData := range portData.Service.Vulnerabilities {
						vuln := models.Vulnerability{
							ScanResultID: scanResult.ID,
							CVEID:        vulnData.CVEID,
							Title:        vulnData.Title,
							Description:  vulnData.Description,
							Severity:     vulnData.Severity,
							CVSS:         vulnData.CVSS,
							Location:     vulnData.Location,
							Parameter:    vulnData.Parameter,
							Payload:      vulnData.Payload,
							Status:       "open",
						}
						vulnerabilities = append(vulnerabilities, vuln)
					}

					// 处理Web路径
					for _, pathData := range portData.Service.WebPaths {
						webPath := models.WebPath{
							ScanResultID: scanResult.ID,
							Path:         pathData.Path,
							StatusCode:   pathData.StatusCode,
							Title:        pathData.Title,
							Length:       pathData.Length,
						}

						if err := tx.Create(&webPath).Error; err != nil {
							return fmt.Errorf("创建Web路径失败: %w", err)
						}

						// 处理路径级漏洞
						for _, vulnData := range pathData.Vulnerabilities {
							vuln := models.Vulnerability{
								ScanResultID: scanResult.ID,
								WebPathID:    &webPath.ID,
								CVEID:        vulnData.CVEID,
								Title:        vulnData.Title,
								Description:  vulnData.Description,
								Severity:     vulnData.Severity,
								CVSS:         vulnData.CVSS,
								Location:     vulnData.Location,
								Parameter:    vulnData.Parameter,
								Payload:      vulnData.Payload,
								Status:       "open",
							}
							vulnerabilities = append(vulnerabilities, vuln)
						}
					}

					// 处理技术栈
					for _, techData := range portData.Service.Technologies {
						// 查找或创建技术
						tech := models.Technology{
							Name:     techData.Name,
							Category: techData.Category,
						}
						tx.Where("name = ?", techData.Name).FirstOrCreate(&tech)

						// 创建关联
						scanResultTech := models.ScanResultTechnology{
							ScanResultID: scanResult.ID,
							TechnologyID: tech.ID,
							Version:      techData.Version,
						}
						scanResultTechs = append(scanResultTechs, scanResultTech)
					}
				}
			}
		}

		// 批量插入漏洞
		if len(vulnerabilities) > 0 {
			if err := tx.CreateInBatches(vulnerabilities, 100).Error; err != nil {
				return fmt.Errorf("批量创建漏洞失败: %w", err)
			}
		}

		// 批量插入技术栈关联
		if len(scanResultTechs) > 0 {
			if err := tx.CreateInBatches(scanResultTechs, 100).Error; err != nil {
				return fmt.Errorf("批量创建技术栈关联失败: %w", err)
			}
		}

		return nil
	})
}

// 高效的漏洞查询
func (d *ScanDAO) GetVulnerabilities(projectID uint, filters map[string]interface{}) ([]models.Vulnerability, error) {
	query := d.db.Table("vulnerabilities v").
		Select("v.*, sr.port, sr.service_name, t.address, t.type").
		Joins("JOIN scan_results sr ON v.scan_result_id = sr.id").
		Joins("JOIN scan_targets t ON sr.target_id = t.id").
		Where("t.project_id = ?", projectID)

	// 应用过滤条件
	if severity, ok := filters["severity"]; ok && severity != "" {
		query = query.Where("v.severity = ?", severity)
	}

	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("v.status = ?", status)
	}

	if search, ok := filters["search"]; ok && search != "" {
		searchTerm := "%" + search.(string) + "%"
		query = query.Where("v.title LIKE ? OR v.description LIKE ? OR t.address LIKE ?",
			searchTerm, searchTerm, searchTerm)
	}

	var vulnerabilities []models.Vulnerability
	err := query.Order("v.cvss DESC, v.created_at DESC").Find(&vulnerabilities).Error

	return vulnerabilities, err
}

// 构建层次结构视图（用于前端显示）
func (d *ScanDAO) GetProjectHierarchy(projectID uint) ([]models.ScanTarget, error) {
	var targets []models.ScanTarget

	// 获取所有目标，按层次排序
	err := d.db.Where("project_id = ?", projectID).
		Order("CASE WHEN parent_id IS NULL THEN 0 ELSE 1 END, type, address").
		Find(&targets).Error

	if err != nil {
		return nil, err
	}

	// 构建层次关系
	targetMap := make(map[uint]*models.ScanTarget)
	var rootTargets []models.ScanTarget

	// 第一遍：建立映射
	for i := range targets {
		targetMap[targets[i].ID] = &targets[i]
		targets[i].Children = []models.ScanTarget{}
	}

	// 第二遍：构建层次关系
	for _, target := range targets {
		if target.ParentID == nil {
			rootTargets = append(rootTargets, target)
		} else {
			if parent, exists := targetMap[*target.ParentID]; exists {
				parent.Children = append(parent.Children, target)
			}
		}
	}

	return rootTargets, nil
}

// 搜索功能 - 添加分页支持
func (d *ScanDAO) SearchTargets(projectID uint, searchTerm string) ([]models.ScanTarget, error) {
	var targets []models.ScanTarget

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	err := d.db.Where("project_id = ? AND LOWER(address) LIKE ?", projectID, searchPattern).
		Order("type, address").
		Limit(1000). // 限制最大返回数量
		Find(&targets).Error

	return targets, err
}

// SearchTargetsWithPagination 带分页的搜索功能
func (d *ScanDAO) SearchTargetsWithPagination(projectID uint, searchTerm string, page, pageSize int) ([]models.ScanTarget, int64, error) {
	var targets []models.ScanTarget
	var total int64

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	query := d.db.Where("project_id = ? AND LOWER(address) LIKE ?", projectID, searchPattern)

	// 获取总数
	if err := query.Model(&models.ScanTarget{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	offset := (page - 1) * pageSize
	err := query.Order("type, address").
		Offset(offset).
		Limit(pageSize).
		Find(&targets).Error

	return targets, total, err
}