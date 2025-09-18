package service

import (
	"errors"
	"fmt"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
)

type ScanService struct {
	scanDAO *dao.ScanDAO
}

func NewScanService(scanDAO *dao.ScanDAO) *ScanService {
	return &ScanService{
		scanDAO: scanDAO,
	}
}

// Project 管理
func (s *ScanService) CreateProject(name, description string) (*models.Project, error) {
	if name == "" {
		return nil, errors.New("项目名称不能为空")
	}

	// 检查重名
	if _, err := s.scanDAO.GetProjectByName(name); err == nil {
		return nil, errors.New("项目名称已存在")
	}

	project := &models.Project{
		Name:        name,
		Description: description,
	}

	if err := s.scanDAO.CreateProject(project); err != nil {
		return nil, fmt.Errorf("创建项目失败: %w", err)
	}

	return project, nil
}

func (s *ScanService) GetProject(id uint) (*models.Project, error) {
	return s.scanDAO.GetProjectByID(id)
}

func (s *ScanService) ListProjects() ([]models.Project, error) {
	return s.scanDAO.ListProjects()
}

func (s *ScanService) DeleteProject(id uint) error {
	return s.scanDAO.DeleteProject(id)
}

// 扫描结果导入 - 核心业务逻辑
func (s *ScanService) ImportScanResults(projectID uint, scanData *ScanResultData) error {
	// 验证项目存在
	project, err := s.scanDAO.GetProjectByID(projectID)
	if err != nil {
		return fmt.Errorf("项目不存在: %w", err)
	}

	// 构建完整的层次结构
	hierarchyData := s.buildHierarchyFromScanData(project, scanData)

	// 批量保存
	return s.scanDAO.CreateOrUpdateHierarchy(hierarchyData)
}

// 从扫描数据构建层次结构
func (s *ScanService) buildHierarchyFromScanData(project *models.Project, scanData *ScanResultData) *models.Project {
	domainMap := make(map[string]*models.Domain)
	subdomainMap := make(map[string]*models.Subdomain)
	ipMap := make(map[string]*models.IPAddress)

	// 处理每个扫描结果
	for _, result := range scanData.Results {
		// 获取或创建IP地址
		ip, exists := ipMap[result.IP]
		if !exists {
			ip = &models.IPAddress{
				Address: result.IP,
			}
			ipMap[result.IP] = ip
		}

		// 如果有域名信息，处理域名层次
		if result.Domain != "" {
			domain := s.getOrCreateDomain(domainMap, result.Domain, project)

			subdomainName := result.Subdomain
			if subdomainName == "" {
				subdomainName = "@" // 根域
			}

			subdomain := s.getOrCreateSubdomain(subdomainMap, domain, subdomainName)

			// 关联IP到子域名
			subdomain.IPAddresses = append(subdomain.IPAddresses, *ip)
		}

		// 处理端口和服务
		for _, portData := range result.Ports {
			port := &models.Port{
				Number:   portData.Number,
				Protocol: portData.Protocol,
				State:    portData.State,
			}

			// 如果检测到服务，创建服务记录
			if portData.Service != nil {
				service := &models.Service{
					Type:        portData.Service.Name,
					Name:        portData.Service.Name,
					Version:     portData.Service.Version,
					Fingerprint: portData.Service.Fingerprint,
					Banner:      portData.Service.Banner,
				}

				// 处理Web服务特有的数据
				if service.IsWebService() {
					s.processWebServiceData(service, portData.Service.WebData)
				}

				// 处理漏洞
				s.processVulnerabilities(service, portData.Service.Vulnerabilities)

				port.Service = service
			}

			ip.Ports = append(ip.Ports, *port)
		}
	}

	return project
}

// 处理Web服务数据
func (s *ScanService) processWebServiceData(service *models.Service, webData *WebServiceData) {
	if webData == nil {
		return
	}

	// 处理Web路径
	for _, pathData := range webData.Paths {
		webPath := &models.WebPath{
			Path:       pathData.Path,
			StatusCode: pathData.StatusCode,
			Title:      pathData.Title,
			Length:     pathData.Length,
		}

		// 处理路径级漏洞
		for _, vulnData := range pathData.Vulnerabilities {
			vuln := &models.Vulnerability{
				CVEID:       vulnData.CVEID,
				Title:       vulnData.Title,
				Description: vulnData.Description,
				Severity:    vulnData.Severity,
				CVSS:        vulnData.CVSS,
				Location:    pathData.Path,
				Parameter:   vulnData.Parameter,
				Payload:     vulnData.Payload,
			}
			webPath.Vulnerabilities = append(webPath.Vulnerabilities, *vuln)
		}

		service.WebPaths = append(service.WebPaths, *webPath)
	}

	// 处理技术栈（这里简化处理，实际应该通过DAO查询或创建）
	for _, techName := range webData.Technologies {
		tech := &models.Technology{
			Name: techName,
		}
		service.Technologies = append(service.Technologies, *tech)
	}
}

// 处理漏洞数据
func (s *ScanService) processVulnerabilities(service *models.Service, vulnDataList []VulnerabilityData) {
	for _, vulnData := range vulnDataList {
		vuln := &models.Vulnerability{
			CVEID:       vulnData.CVEID,
			Title:       vulnData.Title,
			Description: vulnData.Description,
			Severity:    vulnData.Severity,
			CVSS:        vulnData.CVSS,
			Location:    vulnData.Location,
			Parameter:   vulnData.Parameter,
			Payload:     vulnData.Payload,
		}
		service.Vulnerabilities = append(service.Vulnerabilities, *vuln)
	}
}

// 辅助方法
func (s *ScanService) getOrCreateDomain(domainMap map[string]*models.Domain, domainName string, project *models.Project) *models.Domain {
	if domain, exists := domainMap[domainName]; exists {
		return domain
	}

	domain := &models.Domain{
		Name: domainName,
	}
	domainMap[domainName] = domain

	// 添加到项目中
	project.Domains = append(project.Domains, *domain)
	return domain
}

func (s *ScanService) getOrCreateSubdomain(subdomainMap map[string]*models.Subdomain, domain *models.Domain, subdomainName string) *models.Subdomain {
	key := fmt.Sprintf("%s.%s", subdomainName, domain.Name)
	if subdomain, exists := subdomainMap[key]; exists {
		return subdomain
	}

	subdomain := &models.Subdomain{
		Name: subdomainName,
	}
	subdomainMap[key] = subdomain
	domain.Subdomains = append(domain.Subdomains, *subdomain)
	return subdomain
}

// CreateSampleData 创建示例扫描数据用于演示
func (s *ScanService) CreateSampleData(projectID uint) error {
	sampleData := &ScanResultData{
		Results: []ScanResult{
			{
				IP:        "192.168.1.100",
				Domain:    "example.com",
				Subdomain: "www",
				Ports: []PortData{
					{
						Number:   80,
						Protocol: "tcp",
						State:    "open",
						Service: &ServiceData{
							Name:        "http",
							Version:     "Apache/2.4.41",
							Fingerprint: "Apache httpd 2.4.41 ((Ubuntu))",
							Banner:      "HTTP/1.1 200 OK Server: Apache/2.4.41",
							WebData: &WebServiceData{
								Paths: []WebPathData{
									{
										Path:       "/",
										StatusCode: 200,
										Title:      "Welcome to Example.com",
										Length:     2048,
									},
									{
										Path:       "/admin",
										StatusCode: 200,
										Title:      "Admin Panel",
										Length:     1024,
										Vulnerabilities: []VulnerabilityData{
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
								Technologies: []string{"Apache", "PHP", "MySQL"},
							},
							Vulnerabilities: []VulnerabilityData{
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
						Service: &ServiceData{
							Name:        "https",
							Version:     "Apache/2.4.41",
							Fingerprint: "Apache httpd 2.4.41 ((Ubuntu)) OpenSSL/1.1.1f",
							Banner:      "HTTP/1.1 200 OK Server: Apache/2.4.41",
						},
					},
					{
						Number:   22,
						Protocol: "tcp",
						State:    "open",
						Service: &ServiceData{
							Name:        "ssh",
							Version:     "OpenSSH 8.2p1",
							Fingerprint: "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
							Banner:      "SSH-2.0-OpenSSH_8.2p1 Ubuntu-4ubuntu0.5",
						},
					},
				},
			},
			{
				IP:        "192.168.1.101",
				Domain:    "example.com",
				Subdomain: "api",
				Ports: []PortData{
					{
						Number:   8080,
						Protocol: "tcp",
						State:    "open",
						Service: &ServiceData{
							Name:        "http",
							Version:     "nginx/1.18.0",
							Fingerprint: "nginx/1.18.0 (Ubuntu)",
							Banner:      "HTTP/1.1 200 OK Server: nginx/1.18.0",
							WebData: &WebServiceData{
								Paths: []WebPathData{
									{
										Path:       "/api/v1",
										StatusCode: 200,
										Title:      "API Documentation",
										Length:     4096,
									},
									{
										Path:       "/api/v1/users",
										StatusCode: 401,
										Title:      "Unauthorized",
										Length:     256,
										Vulnerabilities: []VulnerabilityData{
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
								Technologies: []string{"nginx", "Node.js", "Express"},
							},
						},
					},
				},
			},
		},
	}

	return s.ImportScanResults(projectID, sampleData)
}

// 统计和查询服务
func (s *ScanService) GetProjectStats(projectID uint) (*ProjectStats, error) {
	project, err := s.scanDAO.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	vulnStats, err := s.scanDAO.GetProjectVulnerabilityStats(projectID)
	if err != nil {
		return nil, err
	}

	stats := &ProjectStats{
		ProjectID:           projectID,
		ProjectName:         project.Name,
		DomainCount:         len(project.Domains),
		VulnerabilityStats:  vulnStats,
	}

	// 计算其他统计信息
	s.calculateProjectStats(project, stats)

	return stats, nil
}

func (s *ScanService) calculateProjectStats(project *models.Project, stats *ProjectStats) {
	subdomainCount := 0
	ipCount := 0
	portCount := 0
	serviceCount := 0

	for _, domain := range project.Domains {
		subdomainCount += len(domain.Subdomains)

		for _, subdomain := range domain.Subdomains {
			ipCount += len(subdomain.IPAddresses)

			for _, ip := range subdomain.IPAddresses {
				portCount += len(ip.Ports)

				for _, port := range ip.Ports {
					if port.Service != nil {
						serviceCount++
					}
				}
			}
		}
	}

	stats.SubdomainCount = subdomainCount
	stats.IPCount = ipCount
	stats.PortCount = portCount
	stats.ServiceCount = serviceCount
}

// 数据传输对象定义
type ScanResultData struct {
	Results []ScanResult `json:"results"`
}

type ScanResult struct {
	IP        string     `json:"ip"`
	Domain    string     `json:"domain,omitempty"`
	Subdomain string     `json:"subdomain,omitempty"`
	Ports     []PortData `json:"ports"`
}

type PortData struct {
	Number   int          `json:"number"`
	Protocol string       `json:"protocol"`
	State    string       `json:"state"`
	Service  *ServiceData `json:"service,omitempty"`
}

type ServiceData struct {
	Name            string              `json:"name"`
	Version         string              `json:"version"`
	Fingerprint     string              `json:"fingerprint"`
	Banner          string              `json:"banner"`
	WebData         *WebServiceData     `json:"web_data,omitempty"`
	Vulnerabilities []VulnerabilityData `json:"vulnerabilities"`
}

type WebServiceData struct {
	Paths        []WebPathData `json:"paths"`
	Technologies []string      `json:"technologies"`
}

type WebPathData struct {
	Path            string              `json:"path"`
	StatusCode      int                 `json:"status_code"`
	Title           string              `json:"title"`
	Length          int                 `json:"length"`
	Vulnerabilities []VulnerabilityData `json:"vulnerabilities"`
}

type VulnerabilityData struct {
	CVEID       string  `json:"cve_id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Severity    string  `json:"severity"`
	CVSS        float64 `json:"cvss"`
	Location    string  `json:"location"`
	Parameter   string  `json:"parameter"`
	Payload     string  `json:"payload"`
}

type ProjectStats struct {
	ProjectID          uint           `json:"project_id"`
	ProjectName        string         `json:"project_name"`
	DomainCount        int            `json:"domain_count"`
	SubdomainCount     int            `json:"subdomain_count"`
	IPCount            int            `json:"ip_count"`
	PortCount          int            `json:"port_count"`
	ServiceCount       int            `json:"service_count"`
	VulnerabilityStats map[string]int `json:"vulnerability_stats"`
}