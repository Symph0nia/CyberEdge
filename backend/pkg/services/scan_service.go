package services

import (
	"context"
	"fmt"
	"time"

	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
	"cyberedge/pkg/scanner"
	"cyberedge/pkg/scanner/tools"
)

// ScanService 扫描服务
type ScanService struct {
	scanDAO     dao.ScanDAOInterface
	registry    *scanner.ScannerRegistry
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

// StartScan 启动扫描任务
func (s *ScanService) StartScan(ctx context.Context, projectID uint, target string, pipelineName string) (*models.ScanResult, error) {
	// 获取扫描流水线
	pipeline, exists := scanner.GetPipeline(pipelineName)
	if !exists {
		return nil, fmt.Errorf("流水线配置不存在: %s", pipelineName)
	}

	// 设置流水线参数
	pipeline.ProjectID = projectID
	pipeline.Target = target

	// 创建扫描目标
	scanTarget := &models.ScanTarget{
		ProjectID: projectID,
		Type:      "pipeline",
		Address:   target,
		CreatedAt: time.Now(),
	}

	// TODO: 扫描框架需要重新设计以配合新的DAO接口
	// if err := s.scanDAO.CreateScanTarget(scanTarget); err != nil {
	//     return nil, fmt.Errorf("创建扫描目标失败: %w", err)
	// }
	_ = scanTarget // 避免编译错误

	// 创建扫描记录 (使用端口0表示流水线扫描)
	scanResult := &models.ScanResult{
		ProjectID:   projectID,
		TargetID:    scanTarget.ID,
		Port:        0, // 流水线扫描使用端口0
		Protocol:    "pipeline",
		State:       "running",
		ServiceName: pipelineName,
		CreatedAt:   time.Now(),
	}

	// 保存初始扫描记录
	// TODO: 扫描框架需要重新设计以配合新的DAO接口
	// if err := s.scanDAO.CreateScanResult(scanResult); err != nil {
	//     return nil, fmt.Errorf("创建扫描记录失败: %w", err)
	// }

	// 异步执行扫描 (使用传入的context)
	go s.executeScanPipeline(ctx, scanResult, pipeline)

	return scanResult, nil
}

// executeScanPipeline 执行扫描流水线
func (s *ScanService) executeScanPipeline(ctx context.Context, scanRecord *models.ScanResult, pipeline scanner.ScanPipeline) {
	manager := s.registry.GetManager()

	// 执行流水线
	results, err := manager.ExecutePipeline(ctx, pipeline)

	// 更新扫描记录状态
	scanRecord.UpdatedAt = time.Now()
	if err != nil {
		scanRecord.State = "failed"
		// Note: ErrorMessage field doesn't exist in ScanResultOptimized
	} else {
		scanRecord.State = "completed"
	}

	// 保存扫描结果
	if err := s.processScanResults(scanRecord, results); err != nil {
		scanRecord.State = "failed"
		// Note: ErrorMessage field doesn't exist in ScanResultOptimized
	}

	// 更新数据库记录
	// TODO: s.scanDAO.UpdateScanResult(scanRecord)
}

// processScanResults 处理扫描结果
func (s *ScanService) processScanResults(scanRecord *models.ScanResult, results []scanner.ScanResult) error {
	for _, result := range results {
		switch result.Category {
		case scanner.CategorySubdomain:
			if err := s.processSubdomainResults(scanRecord, result); err != nil {
				return err
			}
		case scanner.CategoryPort:
			if err := s.processPortResults(scanRecord, result); err != nil {
				return err
			}
		case scanner.CategoryWebTech:
			if err := s.processWebTechResults(scanRecord, result); err != nil {
				return err
			}
		case scanner.CategoryWebPath:
			if err := s.processWebPathResults(scanRecord, result); err != nil {
				return err
			}
		case scanner.CategoryVulnerability:
			if err := s.processVulnerabilityResults(scanRecord, result); err != nil {
				return err
			}
		}
	}

	return nil
}

// processSubdomainResults 处理子域名扫描结果
func (s *ScanService) processSubdomainResults(scanRecord *models.ScanResult, result scanner.ScanResult) error {
	subdomainData, ok := result.Data.(scanner.SubdomainData)
	if !ok {
		return fmt.Errorf("无效的子域名数据格式")
	}

	for _, subdomain := range subdomainData.Subdomains {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(scanRecord.ProjectID, subdomain.Subdomain, "subdomain")
		if err != nil {
			continue
		}

		// 创建子域名扫描结果
		subResult := &models.ScanResult{
			ProjectID:   scanRecord.ProjectID,
			TargetID:    target.ID,
			Port:        0, // 子域名扫描不涉及端口
			Protocol:    "dns",
			State:       "discovered",
			ServiceName: "subdomain-" + result.ScannerName,
			CreatedAt:   result.EndTime,
		}

		// TODO: s.scanDAO.CreateScanResult(subResult)
		_ = subResult // 避免编译错误
	}

	return nil
}

// processPortResults 处理端口扫描结果
func (s *ScanService) processPortResults(scanRecord *models.ScanResult, result scanner.ScanResult) error {
	portData, ok := result.Data.(scanner.PortData)
	if !ok {
		return fmt.Errorf("无效的端口数据格式")
	}

	for _, port := range portData.Ports {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(scanRecord.ProjectID, result.Target, "host")
		if err != nil {
			continue
		}

		// 创建端口扫描结果
		portResult := &models.ScanResult{
			ProjectID:   scanRecord.ProjectID,
			TargetID:    target.ID,
			Port:        port.Port,
			Protocol:    port.Protocol,
			State:       port.State,
			ServiceName: port.Service.Name,
			Version:     port.Service.Version,
			Banner:      port.Service.Banner,
			CreatedAt:   result.EndTime,
		}

		// TODO: s.scanDAO.CreateScanResult(portResult)
		_ = portResult // 避免编译错误
	}

	return nil
}

// processWebTechResults 处理Web技术扫描结果
func (s *ScanService) processWebTechResults(scanRecord *models.ScanResult, result scanner.ScanResult) error {
	webTechData, ok := result.Data.(scanner.WebTechData)
	if !ok {
		return fmt.Errorf("无效的Web技术数据格式")
	}

	// 查找或创建扫描目标
	target, err := s.findOrCreateScanTarget(scanRecord.ProjectID, webTechData.URL, "web")
	if err != nil {
		return err
	}

	// 创建Web技术扫描结果
	webResult := &models.ScanResult{
		ProjectID:    scanRecord.ProjectID,
		TargetID:     target.ID,
		Port:         80, // 默认HTTP端口
		Protocol:     "tcp",
		State:        "open",
		ServiceName:  "http",
		IsWebService: true,
		HTTPTitle:    webTechData.Title,
		HTTPStatus:   webTechData.StatusCode,
		CreatedAt:    result.EndTime,
	}

	// TODO: return s.scanDAO.CreateScanResult(webResult)
	_ = webResult // 避免编译错误
	return nil
}

// processWebPathResults 处理Web路径扫描结果
func (s *ScanService) processWebPathResults(scanRecord *models.ScanResult, result scanner.ScanResult) error {
	webPathData, ok := result.Data.(scanner.WebPathData)
	if !ok {
		return fmt.Errorf("无效的Web路径数据格式")
	}

	for _, path := range webPathData.Paths {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(scanRecord.ProjectID, path.URL, "path")
		if err != nil {
			continue
		}

		// 创建路径扫描结果
		pathResult := &models.ScanResult{
			ProjectID:    scanRecord.ProjectID,
			TargetID:     target.ID,
			Port:         80, // 默认HTTP端口
			Protocol:     "tcp",
			State:        "open",
			ServiceName:  "http-path",
			IsWebService: true,
			HTTPTitle:    path.Title,
			HTTPStatus:   path.StatusCode,
			CreatedAt:    result.EndTime,
		}

		// TODO: s.scanDAO.CreateScanResult(pathResult)
		_ = pathResult // 避免编译错误
	}

	return nil
}

// processVulnerabilityResults 处理漏洞扫描结果
func (s *ScanService) processVulnerabilityResults(scanRecord *models.ScanResult, result scanner.ScanResult) error {
	vulnData, ok := result.Data.(scanner.VulnerabilityData)
	if !ok {
		return fmt.Errorf("无效的漏洞数据格式")
	}

	for _, vuln := range vulnData.Vulnerabilities {
		// 查找或创建扫描目标
		target, err := s.findOrCreateScanTarget(scanRecord.ProjectID, vuln.Target, "vulnerability")
		if err != nil {
			continue
		}

		// 首先创建漏洞扫描结果记录
		vulnResult := &models.ScanResult{
			ProjectID:   scanRecord.ProjectID,
			TargetID:    target.ID,
			Port:        0, // 漏洞扫描端口待定
			Protocol:    "tcp",
			State:       "vulnerable",
			ServiceName: "vulnerability-" + result.ScannerName,
			CreatedAt:   result.EndTime,
		}

		// TODO: if err := s.scanDAO.CreateScanResult(vulnResult); err != nil {
		//     continue
		// }
		_ = vulnResult // 避免编译错误

		// 创建漏洞记录
		vulnRecord := &models.Vulnerability{
			ScanResultID: vulnResult.ID,
			CVEID:        vuln.CVEID,
			Title:        vuln.Title,
			Description:  vuln.Description,
			Severity:     vuln.Severity,
			CVSS:         vuln.CVSS,
			Location:     vuln.Location,
			Parameter:    vuln.Parameter,
			Payload:      vuln.Payload,
			Status:       "open",
			CreatedAt:    time.Now(),
		}

		// TODO: 使用 ImportScanData 批量创建漏洞，而不是单独创建
		_ = vulnRecord // 避免编译错误

	}

	return nil
}

// findOrCreateScanTarget 查找或创建扫描目标
func (s *ScanService) findOrCreateScanTarget(projectID uint, target string, targetType string) (*models.ScanTarget, error) {
	// TODO: 扫描框架需要重新设计以配合新的DAO接口
	// existing, err := s.scanDAO.GetScanTargetByTarget(projectID, target)
	// if err == nil {
	//     return existing, nil
	// }

	// 创建新目标
	newTarget := &models.ScanTarget{
		ProjectID: projectID,
		Address:   target,
		Type:      targetType,
		CreatedAt: time.Now(),
	}

	// TODO: if err := s.scanDAO.CreateScanTarget(newTarget); err != nil {
	//     return nil, err
	// }

	return newTarget, nil
}


// GetScanStatus 获取扫描状态
func (s *ScanService) GetScanStatus(scanID uint) (*models.ScanResult, error) {
	// TODO: 实现扫描状态查询 - 当前DAO接口不支持按ID查询扫描结果
	return nil, fmt.Errorf("扫描状态查询功能未实现")
}

// GetAvailableTools 获取可用工具
func (s *ScanService) GetAvailableTools() map[scanner.ScanCategory][]scanner.ScannerInfo {
	return s.registry.GetAvailableTools()
}

// GetAvailablePipelines 获取可用流水线
func (s *ScanService) GetAvailablePipelines() []string {
	return scanner.ListPipelines()
}

// GetProjectScanResults 获取项目扫描结果
func (s *ScanService) GetProjectScanResults(projectID uint, scanType string, status string) ([]models.ScanResult, error) {
	// TODO: 扫描框架需要重新设计以配合新的DAO接口
	return nil, fmt.Errorf("扫描结果查询功能未实现")
}

// GetProjectVulnerabilities 获取项目漏洞
func (s *ScanService) GetProjectVulnerabilities(projectID uint, severity string) ([]models.Vulnerability, error) {
	// 使用新的DAO接口方法
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