package dao

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

type ScanDAO struct {
	db *gorm.DB
}

func NewScanDAO(db *gorm.DB) *ScanDAO {
	return &ScanDAO{db: db}
}

// Project 相关操作
func (d *ScanDAO) CreateProject(project *models.Project) error {
	return d.db.Create(project).Error
}

func (d *ScanDAO) GetProjectByID(id uint) (*models.Project, error) {
	var project models.Project
	err := d.db.Preload("Domains.Subdomains.IPAddresses.Ports.Service.Vulnerabilities").
		Preload("Domains.Subdomains.IPAddresses.Ports.Service.WebPaths.Vulnerabilities").
		Preload("Domains.Subdomains.IPAddresses.Ports.Service.Technologies").
		First(&project, id).Error
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

func (d *ScanDAO) UpdateProject(project *models.Project) error {
	return d.db.Save(project).Error
}

func (d *ScanDAO) DeleteProject(id uint) error {
	return d.db.Delete(&models.Project{}, id).Error
}

// Domain 相关操作
func (d *ScanDAO) CreateDomain(domain *models.Domain) error {
	return d.db.Create(domain).Error
}

func (d *ScanDAO) GetDomainByName(projectID uint, name string) (*models.Domain, error) {
	var domain models.Domain
	err := d.db.Where("project_id = ? AND name = ?", projectID, name).First(&domain).Error
	return &domain, err
}

// Subdomain 相关操作
func (d *ScanDAO) CreateSubdomain(subdomain *models.Subdomain) error {
	return d.db.Create(subdomain).Error
}

func (d *ScanDAO) GetSubdomainByName(domainID uint, name string) (*models.Subdomain, error) {
	var subdomain models.Subdomain
	err := d.db.Where("domain_id = ? AND name = ?", domainID, name).First(&subdomain).Error
	return &subdomain, err
}

// IPAddress 相关操作
func (d *ScanDAO) CreateIPAddress(ip *models.IPAddress) error {
	return d.db.Create(ip).Error
}

func (d *ScanDAO) GetIPAddressByAddress(address string) (*models.IPAddress, error) {
	var ip models.IPAddress
	err := d.db.Where("address = ?", address).
		Preload("Ports.Service.Vulnerabilities").
		Preload("Ports.Service.WebPaths.Vulnerabilities").
		First(&ip).Error
	return &ip, err
}

// Port 相关操作
func (d *ScanDAO) CreatePort(port *models.Port) error {
	return d.db.Create(port).Error
}

func (d *ScanDAO) GetPortByIPAndNumber(ipID uint, number int, protocol string) (*models.Port, error) {
	var port models.Port
	err := d.db.Where("ip_address_id = ? AND number = ? AND protocol = ?", ipID, number, protocol).
		Preload("Service").First(&port).Error
	return &port, err
}

// Service 相关操作
func (d *ScanDAO) CreateService(service *models.Service) error {
	return d.db.Create(service).Error
}

func (d *ScanDAO) UpdateService(service *models.Service) error {
	return d.db.Save(service).Error
}

func (d *ScanDAO) GetServiceByPortID(portID uint) (*models.Service, error) {
	var service models.Service
	err := d.db.Where("port_id = ?", portID).
		Preload("Vulnerabilities").
		Preload("WebPaths.Vulnerabilities").
		Preload("Technologies").
		First(&service).Error
	return &service, err
}

// WebPath 相关操作
func (d *ScanDAO) CreateWebPath(webPath *models.WebPath) error {
	return d.db.Create(webPath).Error
}

func (d *ScanDAO) GetWebPathByServiceAndPath(serviceID uint, path string) (*models.WebPath, error) {
	var webPath models.WebPath
	err := d.db.Where("service_id = ? AND path = ?", serviceID, path).First(&webPath).Error
	return &webPath, err
}

// Vulnerability 相关操作
func (d *ScanDAO) CreateVulnerability(vuln *models.Vulnerability) error {
	return d.db.Create(vuln).Error
}

func (d *ScanDAO) GetVulnerabilitiesByService(serviceID uint) ([]models.Vulnerability, error) {
	var vulns []models.Vulnerability
	err := d.db.Where("service_id = ?", serviceID).Find(&vulns).Error
	return vulns, err
}

func (d *ScanDAO) GetVulnerabilitiesByWebPath(webPathID uint) ([]models.Vulnerability, error) {
	var vulns []models.Vulnerability
	err := d.db.Where("web_path_id = ?", webPathID).Find(&vulns).Error
	return vulns, err
}

// Technology 相关操作
func (d *ScanDAO) CreateTechnology(tech *models.Technology) error {
	return d.db.Create(tech).Error
}

func (d *ScanDAO) GetTechnologyByName(name string) (*models.Technology, error) {
	var tech models.Technology
	err := d.db.Where("name = ?", name).First(&tech).Error
	return &tech, err
}

// 统计查询
func (d *ScanDAO) GetProjectVulnerabilityStats(projectID uint) (map[string]int, error) {
	stats := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	rows, err := d.db.Raw(`
		SELECT v.severity, COUNT(*) as count
		FROM vulnerabilities v
		JOIN services s ON v.service_id = s.id OR v.web_path_id IN (
			SELECT wp.id FROM web_paths wp WHERE wp.service_id = s.id
		)
		JOIN ports p ON s.port_id = p.id
		JOIN ip_addresses ip ON p.ip_address_id = ip.id
		JOIN subdomains sd ON ip.subdomain_id = sd.id
		JOIN domains d ON sd.domain_id = d.id
		WHERE d.project_id = ?
		GROUP BY v.severity
	`, projectID).Rows()

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			continue
		}
		stats[severity] = count
	}

	return stats, nil
}

// 批量创建或更新
func (d *ScanDAO) CreateOrUpdateHierarchy(project *models.Project) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		// 创建或更新项目
		if err := tx.Save(project).Error; err != nil {
			return err
		}

		// 递归处理域名及其子结构
		for i := range project.Domains {
			domain := &project.Domains[i]
			domain.ProjectID = project.ID

			if err := tx.Save(domain).Error; err != nil {
				return err
			}

			// 处理子域名
			for j := range domain.Subdomains {
				subdomain := &domain.Subdomains[j]
				subdomain.DomainID = domain.ID

				if err := tx.Save(subdomain).Error; err != nil {
					return err
				}

				// 处理IP地址
				for k := range subdomain.IPAddresses {
					ip := &subdomain.IPAddresses[k]
					ip.SubdomainID = &subdomain.ID

					if err := tx.Save(ip).Error; err != nil {
						return err
					}

					// 处理端口
					for l := range ip.Ports {
						port := &ip.Ports[l]
						port.IPAddressID = ip.ID

						if err := tx.Save(port).Error; err != nil {
							return err
						}

						// 处理服务
						if port.Service != nil {
							service := port.Service
							service.PortID = port.ID

							if err := tx.Save(service).Error; err != nil {
								return err
							}

							// 处理Web路径和漏洞
							if service.IsWebService() {
								for m := range service.WebPaths {
									webPath := &service.WebPaths[m]
									webPath.ServiceID = service.ID

									if err := tx.Save(webPath).Error; err != nil {
										return err
									}

									// 处理路径级漏洞
									for n := range webPath.Vulnerabilities {
										vuln := &webPath.Vulnerabilities[n]
										vuln.WebPathID = &webPath.ID

										if err := tx.Save(vuln).Error; err != nil {
											return err
										}
									}
								}
							}

							// 处理服务级漏洞
							for m := range service.Vulnerabilities {
								vuln := &service.Vulnerabilities[m]
								vuln.ServiceID = &service.ID

								if err := tx.Save(vuln).Error; err != nil {
									return err
								}
							}
						}
					}
				}
			}
		}

		return nil
	})
}

// ========== 优化模型支持 ==========

// ScanTarget 相关操作
func (d *ScanDAO) CreateScanTarget(target *models.ScanTarget) error {
	return d.db.Create(target).Error
}

func (d *ScanDAO) GetScanTargetByTarget(projectID uint, target string) (*models.ScanTarget, error) {
	var scanTarget models.ScanTarget
	err := d.db.Where("project_id = ? AND target = ?", projectID, target).First(&scanTarget).Error
	return &scanTarget, err
}

func (d *ScanDAO) GetScanTargetByID(id uint) (*models.ScanTarget, error) {
	var target models.ScanTarget
	err := d.db.First(&target, id).Error
	return &target, err
}

// ScanResultOptimized 相关操作
func (d *ScanDAO) CreateScanResult(result *models.ScanResultOptimized) error {
	return d.db.Create(result).Error
}

func (d *ScanDAO) UpdateScanResult(result *models.ScanResultOptimized) error {
	return d.db.Save(result).Error
}

func (d *ScanDAO) GetScanResultByID(id uint) (*models.ScanResultOptimized, error) {
	var result models.ScanResultOptimized
	err := d.db.Preload("ScanTarget").First(&result, id).Error
	return &result, err
}

func (d *ScanDAO) GetScanResultsByProject(projectID uint) ([]models.ScanResultOptimized, error) {
	var results []models.ScanResultOptimized
	err := d.db.Where("project_id = ?", projectID).
		Preload("ScanTarget").
		Find(&results).Error
	return results, err
}

func (d *ScanDAO) GetScanResultsByStatus(status string) ([]models.ScanResultOptimized, error) {
	var results []models.ScanResultOptimized
	err := d.db.Where("status = ?", status).
		Preload("ScanTarget").
		Find(&results).Error
	return results, err
}

// VulnerabilityOptimized 相关操作
func (d *ScanDAO) CreateVulnerabilityOptimized(vuln *models.VulnerabilityOptimized) error {
	return d.db.Create(vuln).Error
}

func (d *ScanDAO) GetVulnerabilitiesByProject(projectID uint) ([]models.VulnerabilityOptimized, error) {
	var vulns []models.VulnerabilityOptimized
	err := d.db.Where("project_id = ?", projectID).
		Preload("ScanTarget").
		Find(&vulns).Error
	return vulns, err
}

func (d *ScanDAO) GetVulnerabilitiesBySeverity(projectID uint, severity string) ([]models.VulnerabilityOptimized, error) {
	var vulns []models.VulnerabilityOptimized
	err := d.db.Where("project_id = ? AND severity = ?", projectID, severity).
		Preload("ScanTarget").
		Find(&vulns).Error
	return vulns, err
}

func (d *ScanDAO) GetVulnerabilityStats(projectID uint) (map[string]int, error) {
	stats := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	rows, err := d.db.Raw(`
		SELECT severity, COUNT(*) as count
		FROM vulnerability_optimizeds
		WHERE project_id = ?
		GROUP BY severity
	`, projectID).Rows()

	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			continue
		}
		stats[severity] = count
	}

	return stats, nil
}