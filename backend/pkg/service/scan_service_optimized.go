package service

import (
	"errors"
	"fmt"

	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
)

// 优化后的ScanService - 基于Linus审查建议，简化业务逻辑，专注核心功能

type ScanServiceOptimized struct {
	scanDAO dao.ScanDAOOptimizedInterface
}

func NewScanServiceOptimized(scanDAO dao.ScanDAOOptimizedInterface) *ScanServiceOptimized {
	return &ScanServiceOptimized{
		scanDAO: scanDAO,
	}
}

// Project 管理 - 保持简单直接
func (s *ScanServiceOptimized) CreateProject(name, description string) (*models.ProjectOptimized, error) {
	if name == "" {
		return nil, errors.New("项目名称不能为空")
	}

	// 检查重名 - 直接查询，不复杂化
	if _, err := s.scanDAO.GetProjectByName(name); err == nil {
		return nil, errors.New("项目名称已存在")
	}

	project := &models.ProjectOptimized{
		Name:        name,
		Description: description,
	}

	if err := s.scanDAO.CreateProject(project); err != nil {
		return nil, fmt.Errorf("创建项目失败: %w", err)
	}

	return project, nil
}

func (s *ScanServiceOptimized) GetProject(id uint) (*models.ProjectOptimized, error) {
	return s.scanDAO.GetProjectByID(id)
}

func (s *ScanServiceOptimized) ListProjects() ([]models.ProjectOptimized, error) {
	return s.scanDAO.ListProjects()
}

func (s *ScanServiceOptimized) DeleteProject(id uint) error {
	return s.scanDAO.DeleteProject(id)
}

// 项目详情 - 单次查询获取所有必要数据
func (s *ScanServiceOptimized) GetProjectDetails(projectID uint) (*models.ProjectStatsOptimized, []models.ScanTarget, error) {
	return s.scanDAO.GetProjectDetails(projectID)
}

// 项目统计 - 委托给DAO的高效查询
func (s *ScanServiceOptimized) GetProjectStats(projectID uint) (*models.ProjectStatsOptimized, error) {
	return s.scanDAO.GetProjectStatsOptimized(projectID)
}

// 扫描数据导入 - 简化的业务逻辑，委托复杂性给DAO
func (s *ScanServiceOptimized) ImportScanData(data *models.ScanDataImport) error {
	// 验证项目存在
	_, err := s.scanDAO.GetProjectByID(data.ProjectID)
	if err != nil {
		return fmt.Errorf("项目不存在: %w", err)
	}

	// 基本数据验证
	if err := s.validateScanData(data); err != nil {
		return fmt.Errorf("数据验证失败: %w", err)
	}

	// 委托给DAO执行批量导入
	return s.scanDAO.ImportScanData(data)
}

// 数据验证 - 专注业务规则，不处理数据转换
func (s *ScanServiceOptimized) validateScanData(data *models.ScanDataImport) error {
	if len(data.Results) == 0 {
		return errors.New("扫描结果不能为空")
	}

	for i, target := range data.Results {
		if target.Address == "" {
			return fmt.Errorf("第%d个目标地址不能为空", i+1)
		}

		if target.Type != "domain" && target.Type != "subdomain" && target.Type != "ip" {
			return fmt.Errorf("第%d个目标类型无效: %s", i+1, target.Type)
		}

		for j, port := range target.Ports {
			if port.Number < 1 || port.Number > 65535 {
				return fmt.Errorf("第%d个目标的第%d个端口号无效: %d", i+1, j+1, port.Number)
			}

			if port.Protocol != "tcp" && port.Protocol != "udp" {
				return fmt.Errorf("第%d个目标的第%d个端口协议无效: %s", i+1, j+1, port.Protocol)
			}
		}
	}

	return nil
}

// 漏洞查询 - 委托给DAO，专注过滤逻辑
func (s *ScanServiceOptimized) GetVulnerabilities(projectID uint, filters map[string]interface{}) ([]models.VulnerabilityOptimized, error) {
	// 验证项目存在
	_, err := s.scanDAO.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在: %w", err)
	}

	// 验证过滤参数
	if err := s.validateVulnerabilityFilters(filters); err != nil {
		return nil, fmt.Errorf("过滤参数无效: %w", err)
	}

	return s.scanDAO.GetVulnerabilities(projectID, filters)
}

func (s *ScanServiceOptimized) validateVulnerabilityFilters(filters map[string]interface{}) error {
	validSeverities := map[string]bool{
		"critical": true,
		"high":     true,
		"medium":   true,
		"low":      true,
		"info":     true,
	}

	validStatuses := map[string]bool{
		"open":           true,
		"fixed":          true,
		"false_positive": true,
	}

	if severity, ok := filters["severity"]; ok {
		if severityStr, isString := severity.(string); isString && severityStr != "" {
			if !validSeverities[severityStr] {
				return fmt.Errorf("无效的严重级别: %s", severityStr)
			}
		}
	}

	if status, ok := filters["status"]; ok {
		if statusStr, isString := status.(string); isString && statusStr != "" {
			if !validStatuses[statusStr] {
				return fmt.Errorf("无效的状态: %s", statusStr)
			}
		}
	}

	return nil
}

// 更新漏洞状态 - 简单的业务逻辑
func (s *ScanServiceOptimized) UpdateVulnerabilityStatus(vulnID uint, status string) error {
	validStatuses := map[string]bool{
		"open":           true,
		"fixed":          true,
		"false_positive": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("无效的漏洞状态: %s", status)
	}

	// TODO: 这里应该在DAO中添加UpdateVulnerabilityStatus方法
	// 暂时返回未实现错误
	return errors.New("更新漏洞状态功能未实现")
}

// 项目层次结构 - 委托给DAO
func (s *ScanServiceOptimized) GetProjectHierarchy(projectID uint) ([]models.ScanTarget, error) {
	_, err := s.scanDAO.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在: %w", err)
	}

	return s.scanDAO.GetProjectHierarchy(projectID)
}

// 搜索功能 - 简单委托
func (s *ScanServiceOptimized) SearchTargets(projectID uint, searchTerm string) ([]models.ScanTarget, error) {
	if searchTerm == "" {
		return nil, errors.New("搜索条件不能为空")
	}

	_, err := s.scanDAO.GetProjectByID(projectID)
	if err != nil {
		return nil, fmt.Errorf("项目不存在: %w", err)
	}

	return s.scanDAO.SearchTargets(projectID, searchTerm)
}

// 创建示例数据 - 简化版本，专注演示
func (s *ScanServiceOptimized) CreateSampleData(projectID uint) error {
	sampleData := &models.ScanDataImport{
		ProjectID: projectID,
		Results: []models.ScanTargetImport{
			{
				Type:    "domain",
				Address: "example.com",
				Ports: []models.PortScanImport{
					{
						Number:   80,
						Protocol: "tcp",
						State:    "open",
						Service: &models.ServiceScanImport{
							Name:         "http",
							Version:      "Apache/2.4.41",
							Fingerprint:  "Apache httpd 2.4.41 ((Ubuntu))",
							Banner:       "HTTP/1.1 200 OK Server: Apache/2.4.41",
							IsWebService: true,
							HTTPTitle:    "Welcome to Example.com",
							HTTPStatus:   200,
							WebPaths: []models.WebPathImport{
								{
									Path:       "/admin",
									StatusCode: 200,
									Title:      "Admin Panel",
									Length:     1024,
									Vulnerabilities: []models.VulnerabilityImport{
										{
											Title:       "Unprotected Admin Panel",
											Description: "Admin panel accessible without authentication",
											Severity:    "high",
											CVSS:        7.5,
											Location:    "/admin",
										},
									},
								},
							},
							Technologies: []models.TechnologyImport{
								{Name: "Apache", Category: "web_server", Version: "2.4.41"},
								{Name: "PHP", Category: "language", Version: "7.4"},
							},
							Vulnerabilities: []models.VulnerabilityImport{
								{
									CVEID:       "CVE-2021-44790",
									Title:       "Apache HTTP Server Buffer Overflow",
									Description: "Buffer overflow vulnerability in mod_lua multipart parser",
									Severity:    "critical",
									CVSS:        9.8,
									Location:    "mod_lua",
								},
							},
						},
					},
					{
						Number:   443,
						Protocol: "tcp",
						State:    "open",
						Service: &models.ServiceScanImport{
							Name:         "https",
							Version:      "Apache/2.4.41",
							IsWebService: true,
							HTTPStatus:   200,
						},
					},
				},
			},
			{
				Type:    "subdomain",
				Address: "api.example.com",
				Parent:  "example.com",
				Ports: []models.PortScanImport{
					{
						Number:   8080,
						Protocol: "tcp",
						State:    "open",
						Service: &models.ServiceScanImport{
							Name:         "http",
							Version:      "nginx/1.18.0",
							IsWebService: true,
							HTTPTitle:    "API Documentation",
							HTTPStatus:   200,
							WebPaths: []models.WebPathImport{
								{
									Path:       "/api/v1/users",
									StatusCode: 401,
									Title:      "Unauthorized",
									Length:     256,
									Vulnerabilities: []models.VulnerabilityImport{
										{
											Title:       "SQL Injection",
											Description: "SQL injection vulnerability in user endpoint",
											Severity:    "critical",
											CVSS:        9.1,
											Location:    "/api/v1/users",
											Parameter:   "id",
											Payload:     "1' OR '1'='1",
										},
									},
								},
							},
							Technologies: []models.TechnologyImport{
								{Name: "nginx", Category: "web_server", Version: "1.18.0"},
								{Name: "Node.js", Category: "runtime", Version: "14.17.0"},
							},
						},
					},
				},
			},
			{
				Type:    "ip",
				Address: "192.168.1.200",
				Ports: []models.PortScanImport{
					{
						Number:   22,
						Protocol: "tcp",
						State:    "open",
						Service: &models.ServiceScanImport{
							Name:        "ssh",
							Version:     "OpenSSH 8.2p1",
							Fingerprint: "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
							Banner:      "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
						},
					},
				},
			},
		},
	}

	return s.ImportScanData(sampleData)
}

// 获取项目概览 - 为仪表板提供数据
func (s *ScanServiceOptimized) GetProjectOverview(projectID uint) (map[string]interface{}, error) {
	stats, err := s.GetProjectStats(projectID)
	if err != nil {
		return nil, err
	}

	// 计算风险评分（简化算法）
	riskScore := s.calculateRiskScore(stats.VulnerabilityStats)

	overview := map[string]interface{}{
		"project_name":        stats.ProjectName,
		"target_count":        stats.TargetCount,
		"vulnerability_count": s.getTotalVulnerabilities(stats.VulnerabilityStats),
		"risk_score":          riskScore,
		"risk_level":          s.getRiskLevel(riskScore),
		"last_scan":           stats.LastScanTime,
		"vulnerability_stats": stats.VulnerabilityStats,
		"service_stats": map[string]int{
			"total":       stats.ServiceCount,
			"web_service": stats.WebServiceCount,
		},
	}

	return overview, nil
}

// 简化的风险评分算法
func (s *ScanServiceOptimized) calculateRiskScore(vulnStats map[string]int) float64 {
	score := float64(vulnStats["critical"])*10 +
		float64(vulnStats["high"])*7 +
		float64(vulnStats["medium"])*4 +
		float64(vulnStats["low"])*1

	// 归一化到0-100
	if score > 100 {
		return 100
	}
	return score
}

func (s *ScanServiceOptimized) getRiskLevel(score float64) string {
	if score >= 80 {
		return "critical"
	} else if score >= 60 {
		return "high"
	} else if score >= 30 {
		return "medium"
	} else if score > 0 {
		return "low"
	}
	return "none"
}

func (s *ScanServiceOptimized) getTotalVulnerabilities(vulnStats map[string]int) int {
	total := 0
	for _, count := range vulnStats {
		total += count
	}
	return total
}