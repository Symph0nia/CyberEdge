package handlers

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
)

// SecureScanHandler 安全增强版的扫描处理器
type SecureScanHandler struct {
	scanService         *service.ScanService
	maxImportSize       int    // 最大导入数据大小
	allowedDomainRegex  *regexp.Regexp // 允许的域名格式
	allowedIPRegex      *regexp.Regexp // 允许的IP格式
}

func NewSecureScanHandler(scanService *service.ScanService) *SecureScanHandler {
	// 编译安全验证的正则表达式
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	ipRegex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	return &SecureScanHandler{
		scanService:        scanService,
		maxImportSize:     10000, // 最大10k条扫描结果
		allowedDomainRegex: domainRegex,
		allowedIPRegex:     ipRegex,
	}
}

// RegisterSecureScanRoutes 注册安全的扫描路由
func (h *SecureScanHandler) RegisterSecureScanRoutes(router *gin.RouterGroup) {
	scan := router.Group("/scan")
	{
		// 项目管理
		scan.POST("/projects", h.CreateProject)
		scan.GET("/projects", h.ListProjects)
		scan.GET("/projects/:id", h.GetProject)
		scan.DELETE("/projects/:id", h.DeleteProject)

		// 扫描结果导入 - 增强安全验证
		scan.POST("/projects/:id/import", h.ImportScanResultsSecure)

		// 统计信息
		scan.GET("/projects/:id/stats", h.GetProjectStats)

		// 示例数据
		scan.POST("/projects/:id/sample", h.CreateSampleData)
	}
}

// ImportScanResultsSecure 安全增强的扫描结果导入
func (h *SecureScanHandler) ImportScanResultsSecure(c *gin.Context) {
	// 1. 验证项目ID
	idStr := c.Param("id")
	id, err := h.validateProjectID(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 2. 限制请求体大小（防止DoS攻击）
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 50*1024*1024) // 50MB限制

	// 3. 安全地绑定和验证扫描数据
	var scanData service.ScanResultData
	if err := c.ShouldBindJSON(&scanData); err != nil {
		log.Printf("扫描数据绑定失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "扫描数据格式无效，请检查JSON格式"})
		return
	}

	// 4. 深度验证扫描数据安全性
	if err := h.validateScanDataSecurity(&scanData); err != nil {
		log.Printf("扫描数据安全验证失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("数据验证失败: %s", err.Error())})
		return
	}

	// 5. 执行导入
	if err := h.scanService.ImportScanResults(id, &scanData); err != nil {
		// 记录详细错误到日志，但不暴露给客户端
		log.Printf("扫描结果导入失败 - ProjectID: %d, Error: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入扫描结果失败，请联系管理员"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "扫描结果导入成功",
		"imported_count": len(scanData.Results),
	})
}

// validateProjectID 验证项目ID的安全性
func (h *SecureScanHandler) validateProjectID(idStr string) (uint, error) {
	if idStr == "" {
		return 0, fmt.Errorf("项目ID不能为空")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("项目ID必须是有效的数字")
	}

	if id == 0 || id > 4294967295 { // uint32最大值
		return 0, fmt.Errorf("项目ID超出有效范围")
	}

	return uint(id), nil
}

// validateScanDataSecurity 深度验证扫描数据的安全性
func (h *SecureScanHandler) validateScanDataSecurity(scanData *service.ScanResultData) error {
	if len(scanData.Results) == 0 {
		return fmt.Errorf("扫描结果不能为空")
	}

	if len(scanData.Results) > h.maxImportSize {
		return fmt.Errorf("扫描结果数量超出限制 (最大 %d 条)", h.maxImportSize)
	}

	for i, result := range scanData.Results {
		if err := h.validateScanResult(i, &result); err != nil {
			return fmt.Errorf("第 %d 条扫描结果: %w", i+1, err)
		}
	}

	return nil
}

// validateScanResult 验证单个扫描结果
func (h *SecureScanHandler) validateScanResult(index int, result *service.ScanResult) error {
	// 验证IP地址格式
	if result.IP == "" {
		return fmt.Errorf("IP地址不能为空")
	}

	if !h.allowedIPRegex.MatchString(result.IP) {
		return fmt.Errorf("IP地址格式无效: %s", result.IP)
	}

	// 验证域名格式（如果提供）
	if result.Domain != "" {
		if len(result.Domain) > 253 { // DNS标准最大长度
			return fmt.Errorf("域名长度超出限制")
		}

		if !h.allowedDomainRegex.MatchString(result.Domain) {
			return fmt.Errorf("域名格式无效: %s", result.Domain)
		}
	}

	// 验证子域名格式
	if result.Subdomain != "" {
		if len(result.Subdomain) > 63 { // DNS标签最大长度
			return fmt.Errorf("子域名长度超出限制")
		}

		// 防止路径遍历攻击
		if strings.Contains(result.Subdomain, "..") || strings.Contains(result.Subdomain, "/") {
			return fmt.Errorf("子域名包含非法字符")
		}
	}

	// 验证端口信息
	if len(result.Ports) == 0 {
		return fmt.Errorf("端口信息不能为空")
	}

	if len(result.Ports) > 100 { // 限制单个IP的端口数量
		return fmt.Errorf("端口数量超出限制 (最大 100 个)")
	}

	for j, port := range result.Ports {
		if err := h.validatePortData(j, &port); err != nil {
			return fmt.Errorf("端口 %d: %w", j+1, err)
		}
	}

	return nil
}

// validatePortData 验证端口数据
func (h *SecureScanHandler) validatePortData(index int, port *service.PortData) error {
	// 验证端口号
	if port.Number < 1 || port.Number > 65535 {
		return fmt.Errorf("端口号无效: %d", port.Number)
	}

	// 验证协议
	validProtocols := map[string]bool{"tcp": true, "udp": true, "sctp": true}
	if !validProtocols[strings.ToLower(port.Protocol)] {
		return fmt.Errorf("协议无效: %s", port.Protocol)
	}

	// 验证状态
	validStates := map[string]bool{"open": true, "closed": true, "filtered": true}
	if port.State != "" && !validStates[strings.ToLower(port.State)] {
		return fmt.Errorf("端口状态无效: %s", port.State)
	}

	// 验证服务数据
	if port.Service != nil {
		if err := h.validateServiceData(port.Service); err != nil {
			return fmt.Errorf("服务数据: %w", err)
		}
	}

	return nil
}

// validateServiceData 验证服务数据
func (h *SecureScanHandler) validateServiceData(service *service.ServiceData) error {
	// 限制字符串长度，防止存储攻击
	if len(service.Name) > 100 {
		return fmt.Errorf("服务名称过长")
	}

	if len(service.Version) > 100 {
		return fmt.Errorf("服务版本过长")
	}

	if len(service.Fingerprint) > 500 {
		return fmt.Errorf("服务指纹过长")
	}

	if len(service.Banner) > 2000 {
		return fmt.Errorf("服务横幅过长")
	}

	// 验证是否包含潜在的XSS或脚本内容
	dangerousPatterns := []string{
		"<script", "</script>", "javascript:", "data:text/html",
		"vbscript:", "onload=", "onerror=", "onclick=",
	}

	fields := []struct {
		name  string
		value string
	}{
		{"服务名称", service.Name},
		{"版本", service.Version},
		{"指纹", service.Fingerprint},
		{"横幅", service.Banner},
	}

	for _, field := range fields {
		lowerValue := strings.ToLower(field.value)
		for _, pattern := range dangerousPatterns {
			if strings.Contains(lowerValue, pattern) {
				return fmt.Errorf("%s包含潜在恶意内容", field.name)
			}
		}
	}

	// 验证Web数据
	if service.WebData != nil {
		if err := h.validateWebServiceData(service.WebData); err != nil {
			return fmt.Errorf("Web服务数据: %w", err)
		}
	}

	// 验证漏洞数据
	if len(service.Vulnerabilities) > 50 { // 限制单个服务的漏洞数量
		return fmt.Errorf("漏洞数量超出限制")
	}

	for i, vuln := range service.Vulnerabilities {
		if err := h.validateVulnerabilityData(&vuln); err != nil {
			return fmt.Errorf("漏洞 %d: %w", i+1, err)
		}
	}

	return nil
}

// validateWebServiceData 验证Web服务数据
func (h *SecureScanHandler) validateWebServiceData(webData *service.WebServiceData) error {
	if len(webData.Paths) > 100 { // 限制路径数量
		return fmt.Errorf("Web路径数量超出限制")
	}

	for i, path := range webData.Paths {
		if err := h.validateWebPath(&path); err != nil {
			return fmt.Errorf("Web路径 %d: %w", i+1, err)
		}
	}

	if len(webData.Technologies) > 20 { // 限制技术栈数量
		return fmt.Errorf("技术栈数量超出限制")
	}

	return nil
}

// validateWebPath 验证Web路径
func (h *SecureScanHandler) validateWebPath(path *service.WebPathData) error {
	if len(path.Path) > 500 {
		return fmt.Errorf("路径长度过长")
	}

	if len(path.Title) > 200 {
		return fmt.Errorf("页面标题过长")
	}

	// 验证路径格式，防止路径遍历
	if strings.Contains(path.Path, "..") {
		return fmt.Errorf("路径包含非法字符")
	}

	// 验证HTTP状态码
	if path.StatusCode < 100 || path.StatusCode > 999 {
		return fmt.Errorf("HTTP状态码无效: %d", path.StatusCode)
	}

	return nil
}

// validateVulnerabilityData 验证漏洞数据
func (h *SecureScanHandler) validateVulnerabilityData(vuln *service.VulnerabilityData) error {
	if len(vuln.Title) > 200 {
		return fmt.Errorf("漏洞标题过长")
	}

	if len(vuln.Description) > 2000 {
		return fmt.Errorf("漏洞描述过长")
	}

	// 验证严重级别
	validSeverities := map[string]bool{
		"critical": true, "high": true, "medium": true, "low": true, "info": true,
	}
	if vuln.Severity != "" && !validSeverities[strings.ToLower(vuln.Severity)] {
		return fmt.Errorf("漏洞严重级别无效: %s", vuln.Severity)
	}

	// 验证CVSS评分
	if vuln.CVSS < 0 || vuln.CVSS > 10 {
		return fmt.Errorf("CVSS评分无效: %f", vuln.CVSS)
	}

	return nil
}

// CreateProject 创建项目（重用原有逻辑，增加日志记录）
func (h *SecureScanHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description" binding:"max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("创建项目参数验证失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	project, err := h.scanService.CreateProject(req.Name, req.Description)
	if err != nil {
		log.Printf("创建项目失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建项目失败"})
		return
	}

	log.Printf("项目创建成功: ID=%d, Name=%s", project.ID, project.Name)
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"project": project,
	})
}

// 其他方法保持不变，但增加错误日志记录...
func (h *SecureScanHandler) ListProjects(c *gin.Context) {
	projects, err := h.scanService.ListProjects()
	if err != nil {
		log.Printf("获取项目列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"projects": projects,
	})
}

func (h *SecureScanHandler) GetProject(c *gin.Context) {
	id, err := h.validateProjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	project, err := h.scanService.GetProject(id)
	if err != nil {
		log.Printf("获取项目失败: ID=%d, Error=%v", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"project": project,
	})
}

func (h *SecureScanHandler) DeleteProject(c *gin.Context) {
	id, err := h.validateProjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	if err := h.scanService.DeleteProject(id); err != nil {
		log.Printf("删除项目失败: ID=%d, Error=%v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除项目失败"})
		return
	}

	log.Printf("项目删除成功: ID=%d", id)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "项目删除成功",
	})
}

func (h *SecureScanHandler) GetProjectStats(c *gin.Context) {
	id, err := h.validateProjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	stats, err := h.scanService.GetProjectStats(id)
	if err != nil {
		log.Printf("获取项目统计失败: ID=%d, Error=%v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

func (h *SecureScanHandler) CreateSampleData(c *gin.Context) {
	id, err := h.validateProjectID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	if err := h.scanService.CreateSampleData(id); err != nil {
		log.Printf("创建示例数据失败: ID=%d, Error=%v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建示例数据失败"})
		return
	}

	log.Printf("示例数据创建成功: ProjectID=%d", id)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "示例数据创建成功",
	})
}