package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
	"cyberedge/pkg/scanner"
	"cyberedge/pkg/scanner/tools"
)

// ScanService 扫描服务 - 使用ScanJob管理扫描任务，消除端口0的设计缺陷
type ScanService struct {
	scanDAO  dao.ScanDAOInterface
	registry *scanner.ScannerRegistry
}

// NewScanService 创建扫描服务
func NewScanService(scanDAO dao.ScanDAOInterface) (*ScanService, error) {
	registry := scanner.NewScannerRegistry()

	// 注册所有扫描工具
	if err := registerAllTools(registry); err != nil {
		return nil, fmt.Errorf("注册扫描工具失败: %w", err)
	}

	return &ScanService{
		scanDAO:  scanDAO,
		registry: registry,
	}, nil
}

// StartScan 启动扫描任务 - 返回ScanJob而不是ScanResult
func (s *ScanService) StartScan(ctx context.Context, projectID uint, target string, pipelineName string) (*models.ScanJob, error) {
	// 验证流水线是否存在
	pipeline, exists := scanner.GetPipeline(pipelineName)
	if !exists {
		return nil, fmt.Errorf("流水线配置不存在: %s", pipelineName)
	}

	// 创建扫描任务记录
	scanJob := &models.ScanJob{
		ProjectID:    projectID,
		Target:       target,
		PipelineName: pipelineName,
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	// 保存扫描任务到数据库
	if err := s.scanDAO.CreateScanJob(scanJob); err != nil {
		return nil, fmt.Errorf("创建扫描任务失败: %w", err)
	}

	// 设置流水线参数
	pipeline.ProjectID = projectID
	pipeline.Target = target

	// 异步执行扫描
	go s.executeScanPipeline(ctx, scanJob, pipeline)

	return scanJob, nil
}

// executeScanPipeline 执行扫描流水线 - 使用ScanJob管理状态
func (s *ScanService) executeScanPipeline(ctx context.Context, scanJob *models.ScanJob, pipeline scanner.ScanPipeline) {
	// 更新任务状态为运行中
	scanJob.Status = "running"
	scanJob.StartTime = time.Now()
	if err := s.scanDAO.UpdateScanJob(scanJob); err != nil {
		// 记录错误但继续执行
		fmt.Printf("更新扫描任务状态失败: %v\n", err)
	}

	manager := s.registry.GetManager()

	// 执行流水线
	results, err := manager.ExecutePipeline(ctx, pipeline)

	// 处理扫描结果
	var processingErr error
	if err == nil {
		processingErr = s.processScanResults(scanJob.ProjectID, results)
	}

	// 更新最终状态
	endTime := time.Now()
	scanJob.EndTime = &endTime
	scanJob.UpdatedAt = time.Now()

	if err != nil {
		scanJob.Status = "failed"
		scanJob.ErrorMessage = err.Error()
	} else if processingErr != nil {
		scanJob.Status = "failed"
		scanJob.ErrorMessage = processingErr.Error()
	} else {
		scanJob.Status = "completed"
		scanJob.ErrorMessage = ""
	}

	// 保存最终状态
	if updateErr := s.scanDAO.UpdateScanJob(scanJob); updateErr != nil {
		fmt.Printf("更新扫描任务最终状态失败: %v\n", updateErr)
	}
}

// processScanResults 处理扫描结果 - 简化的数据处理逻辑
func (s *ScanService) processScanResults(projectID uint, results []scanner.ScanResult) error {
	for _, result := range results {
		switch result.Category {
		case scanner.CategorySubdomain:
			if err := s.processSubdomainResults(projectID, result); err != nil {
				return fmt.Errorf("处理子域名结果失败: %w", err)
			}
		case scanner.CategoryPort:
			if err := s.processPortResults(projectID, result); err != nil {
				return fmt.Errorf("处理端口结果失败: %w", err)
			}
		case scanner.CategoryWebTech:
			if err := s.processWebTechResults(projectID, result); err != nil {
				return fmt.Errorf("处理Web技术结果失败: %w", err)
			}
		case scanner.CategoryWebPath:
			if err := s.processWebPathResults(projectID, result); err != nil {
				return fmt.Errorf("处理Web路径结果失败: %w", err)
			}
		case scanner.CategoryVulnerability:
			if err := s.processVulnerabilityResults(projectID, result); err != nil {
				return fmt.Errorf("处理漏洞结果失败: %w", err)
			}
		}
	}

	return nil
}

// processSubdomainResults 处理子域名扫描结果
func (s *ScanService) processSubdomainResults(projectID uint, result scanner.ScanResult) error {
	subdomainData, ok := result.Data.(scanner.SubdomainData)
	if !ok {
		return fmt.Errorf("无效的子域名数据格式")
	}

	for _, subdomain := range subdomainData.Subdomains {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(projectID, subdomain.Subdomain, "subdomain")
		if err != nil {
			fmt.Printf("创建子域名目标失败: %v\n", err)
			continue
		}

		// 创建子域名发现记录 - 使用DNS解析结果而不是端口0
		if len(subdomain.IPs) > 0 {
			subResult := &models.ScanResult{
				ProjectID:   projectID,
				TargetID:    target.ID,
				Port:        53, // DNS端口
				Protocol:    "udp",
				State:       "resolved",
				ServiceName: "dns",
				Banner:      fmt.Sprintf("Resolved to %s", subdomain.IPs[0]),
				CreatedAt:   result.EndTime,
			}

			if err := s.scanDAO.CreateScanResult(subResult); err != nil {
				fmt.Printf("保存子域名解析结果失败: %v\n", err)
			}
		}
	}

	return nil
}

// processPortResults 处理端口扫描结果
func (s *ScanService) processPortResults(projectID uint, result scanner.ScanResult) error {
	portData, ok := result.Data.(scanner.PortData)
	if !ok {
		return fmt.Errorf("无效的端口数据格式")
	}

	for _, port := range portData.Ports {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(projectID, result.Target, "host")
		if err != nil {
			fmt.Printf("创建主机目标失败: %v\n", err)
			continue
		}

		// 创建端口扫描结果 - 使用真实端口号
		portResult := &models.ScanResult{
			ProjectID:   projectID,
			TargetID:    target.ID,
			Port:        port.Port,
			Protocol:    port.Protocol,
			State:       port.State,
			ServiceName: port.Service.Name,
			Version:     port.Service.Version,
			Banner:      port.Service.Banner,
			CreatedAt:   result.EndTime,
		}

		if err := s.scanDAO.CreateScanResult(portResult); err != nil {
			fmt.Printf("保存端口扫描结果失败: %v\n", err)
		}
	}

	return nil
}

// processWebTechResults 处理Web技术扫描结果
func (s *ScanService) processWebTechResults(projectID uint, result scanner.ScanResult) error {
	webTechData, ok := result.Data.(scanner.WebTechData)
	if !ok {
		return fmt.Errorf("无效的Web技术数据格式")
	}

	// 查找或创建扫描目标
	target, err := s.findOrCreateScanTarget(projectID, webTechData.URL, "web")
	if err != nil {
		return err
	}

	// 创建Web服务扫描结果 - 使用实际的HTTP端口
	port := 80
	// 从URL中推断端口
	if strings.Contains(webTechData.URL, ":443") || strings.HasPrefix(webTechData.URL, "https://") {
		port = 443
	}

	webResult := &models.ScanResult{
		ProjectID:    projectID,
		TargetID:     target.ID,
		Port:         port,
		Protocol:     "tcp",
		State:        "open",
		ServiceName:  "http",
		IsWebService: true,
		HTTPTitle:    webTechData.Title,
		HTTPStatus:   webTechData.StatusCode,
		CreatedAt:    result.EndTime,
	}

	return s.scanDAO.CreateScanResult(webResult)
}

// processWebPathResults 处理Web路径扫描结果
func (s *ScanService) processWebPathResults(projectID uint, result scanner.ScanResult) error {
	webPathData, ok := result.Data.(scanner.WebPathData)
	if !ok {
		return fmt.Errorf("无效的Web路径数据格式")
	}

	for _, path := range webPathData.Paths {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(projectID, path.URL, "web")
		if err != nil {
			fmt.Printf("创建Web路径目标失败: %v\n", err)
			continue
		}

		// 创建Web路径扫描结果 - 使用实际的HTTP端口
		port := 80
		// 从URL中推断端口
		if strings.Contains(path.URL, ":443") || strings.HasPrefix(path.URL, "https://") {
			port = 443
		}

		pathResult := &models.ScanResult{
			ProjectID:    projectID,
			TargetID:     target.ID,
			Port:         port,
			Protocol:     "tcp",
			State:        "accessible",
			ServiceName:  "http",
			IsWebService: true,
			HTTPTitle:    path.Title,
			HTTPStatus:   path.StatusCode,
			CreatedAt:    result.EndTime,
		}

		if err := s.scanDAO.CreateScanResult(pathResult); err != nil {
			fmt.Printf("保存Web路径结果失败: %v\n", err)
		}
	}

	return nil
}

// processVulnerabilityResults 处理漏洞扫描结果
func (s *ScanService) processVulnerabilityResults(projectID uint, result scanner.ScanResult) error {
	vulnData, ok := result.Data.(scanner.VulnerabilityData)
	if !ok {
		return fmt.Errorf("无效的漏洞数据格式")
	}

	for _, vuln := range vulnData.Vulnerabilities {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(projectID, vuln.Target, "host")
		if err != nil {
			fmt.Printf("创建漏洞目标失败: %v\n", err)
			continue
		}

		// 从漏洞的Location推断端口号
		port := 80 // 默认HTTP端口
		if strings.Contains(vuln.Location, ":443") || strings.Contains(vuln.Location, "https://") {
			port = 443
		} else if strings.Contains(vuln.Location, ":22") {
			port = 22
		} else if strings.Contains(vuln.Location, ":21") {
			port = 21
		}

		// 查找相关的ScanResult记录
		var scanResultID uint
		targetResults, err := s.scanDAO.GetTargetScanResults(target.ID)
		if err == nil {
			for _, sr := range targetResults {
				if sr.Port == port {
					scanResultID = sr.ID
					break
				}
			}
		}

		// 如果没找到对应的ScanResult，创建一个基础的
		if scanResultID == 0 {

			baseResult := &models.ScanResult{
				ProjectID:   projectID,
				TargetID:    target.ID,
				Port:        port,
				Protocol:    "tcp",
				State:       "vulnerable",
				ServiceName: "unknown",
				CreatedAt:   result.EndTime,
			}

			if err := s.scanDAO.CreateScanResult(baseResult); err != nil {
				fmt.Printf("创建基础扫描结果失败: %v\n", err)
				continue
			}
			scanResultID = baseResult.ID
		}

		// 直接使用ImportScanData批量创建漏洞记录
		vulnImport := &models.ScanDataImport{
			ProjectID: projectID,
			Results: []models.ScanTargetImport{
				{
					Type:    target.Type,
					Address: target.Address,
					Ports: []models.PortScanImport{
						{
							Number:   port,
							Protocol: "tcp",
							State:    "vulnerable",
							Service: &models.ServiceScanImport{
								Name: "vulnerable-service",
								Vulnerabilities: []models.VulnerabilityImport{
									{
										CVEID:       vuln.CVEID,
										Title:       vuln.Title,
										Description: vuln.Description,
										Severity:    vuln.Severity,
										CVSS:        vuln.CVSS,
										Location:    vuln.Location,
										Parameter:   vuln.Parameter,
										Payload:     vuln.Payload,
									},
								},
							},
						},
					},
				},
			},
		}

		if err := s.scanDAO.ImportScanData(vulnImport); err != nil {
			fmt.Printf("导入漏洞数据失败: %v\n", err)
		}
	}

	return nil
}

// findOrCreateScanTarget 查找或创建扫描目标 - 移除TODO，实现完整逻辑
func (s *ScanService) findOrCreateScanTarget(projectID uint, address string, targetType string) (*models.ScanTarget, error) {
	// 输入验证 - 核心安全检查
	if projectID == 0 {
		return nil, fmt.Errorf("无效的项目ID: %d", projectID)
	}

	if strings.TrimSpace(address) == "" {
		return nil, fmt.Errorf("扫描目标地址不能为空")
	}

	// 地址长度限制
	if len(address) > 255 {
		return nil, fmt.Errorf("扫描目标地址过长，最大255字符")
	}

	// 过滤恶意输入 - SQL注入防护
	address = strings.TrimSpace(address)
	if containsSQLInjection(address) {
		return nil, fmt.Errorf("检测到恶意输入，拒绝处理")
	}

	// 首先尝试查找已存在的目标
	existing, err := s.scanDAO.GetScanTargetByAddress(projectID, address)
	if err == nil && existing != nil {
		return existing, nil
	}

	// 创建新目标
	newTarget := &models.ScanTarget{
		ProjectID: projectID,
		Address:   address,
		Type:      targetType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.scanDAO.CreateScanTarget(newTarget); err != nil {
		return nil, fmt.Errorf("创建扫描目标失败: %w", err)
	}

	return newTarget, nil
}

// GetScanStatus 获取扫描状态 - 使用ScanJob而不是ScanResult
func (s *ScanService) GetScanStatus(scanJobID uint) (*models.ScanJob, error) {
	return s.scanDAO.GetScanJobByID(scanJobID)
}

// GetAvailableTools 获取可用工具
func (s *ScanService) GetAvailableTools() map[scanner.ScanCategory][]scanner.ScannerInfo {
	return s.registry.GetAvailableTools()
}

// GetAvailablePipelines 获取可用流水线
func (s *ScanService) GetAvailablePipelines() []string {
	return scanner.ListPipelines()
}

// GetProjectScanResults 获取项目扫描结果 - 直接使用DAO方法
func (s *ScanService) GetProjectScanResults(projectID uint, filters map[string]interface{}) ([]models.ScanResult, error) {
	return s.scanDAO.GetProjectScanResults(projectID, filters)
}

// GetProjectScanJobs 获取项目扫描任务
func (s *ScanService) GetProjectScanJobs(projectID uint, filters map[string]interface{}) ([]models.ScanJob, error) {
	return s.scanDAO.GetProjectScanJobs(projectID, filters)
}

// GetProjectVulnerabilities 获取项目漏洞
func (s *ScanService) GetProjectVulnerabilities(projectID uint, severity string) ([]models.Vulnerability, error) {
	filters := make(map[string]interface{})
	if severity != "" {
		filters["severity"] = severity
	}
	return s.scanDAO.GetVulnerabilities(projectID, filters)
}

// GetVulnerabilityStats 获取漏洞统计
func (s *ScanService) GetVulnerabilityStats(projectID uint) (map[string]int, error) {
	return s.scanDAO.GetVulnerabilityStats(projectID)
}

// registerAllTools 注册所有可用的扫描工具
func registerAllTools(registry *scanner.ScannerRegistry) error {
	// 注册子域名扫描工具
	if err := registry.RegisterScanner(tools.NewSubfinderScanner()); err != nil {
		return err
	}

	// 注册端口扫描工具
	if err := registry.RegisterScanner(tools.NewNmapScanner()); err != nil {
		return err
	}

	// 注册Web技术探测工具
	if err := registry.RegisterScanner(tools.NewHttpxScanner()); err != nil {
		return err
	}

	// 注册漏洞扫描工具
	if err := registry.RegisterScanner(tools.NewNucleiScanner()); err != nil {
		return err
	}

	// 注册目录扫描工具
	if err := registry.RegisterScanner(tools.NewGobusterScanner()); err != nil {
		return err
	}

	return nil
}

// containsSQLInjection 检测SQL注入攻击模式
func containsSQLInjection(input string) bool {
	// 常见SQL注入关键词 - 简单但有效的检测
	sqlPatterns := []string{
		"DROP TABLE", "DELETE FROM", "INSERT INTO", "UPDATE SET",
		"--", "/*", "*/", "UNION SELECT", "OR 1=1", "OR '1'='1",
		"'; ", "' OR ", "\" OR ", "\"; ", "\" AND ", "' AND ",
	}

	upperInput := strings.ToUpper(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(upperInput, pattern) {
			return true
		}
	}
	return false
}