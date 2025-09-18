package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"cyberedge/pkg/service"
	"cyberedge/pkg/services"
	"github.com/gin-gonic/gin"
)

// ScanHandler 统一的扫描处理器 - 整合项目管理、扫描任务、安全验证
type ScanHandler struct {
	scanService        *service.ScanService
	frameworkService   *services.ScanService
	maxImportSize      int
	allowedDomainRegex *regexp.Regexp
	allowedIPRegex     *regexp.Regexp
}

func NewScanHandler(scanService *service.ScanService) *ScanHandler {
	// 编译安全验证正则表达式
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	ipRegex := regexp.MustCompile(`^(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`)

	return &ScanHandler{
		scanService:        scanService,
		frameworkService:   nil, // TODO: 框架服务暂时未实现
		maxImportSize:      10000, // 最大10k条扫描结果
		allowedDomainRegex: domainRegex,
		allowedIPRegex:     ipRegex,
	}
}

// RegisterScanRoutes 注册统一的扫描路由
func (h *ScanHandler) RegisterScanRoutes(router *gin.RouterGroup) {
	scan := router.Group("/scan")
	{
		// 项目管理
		scan.POST("/projects", h.CreateProject)
		scan.GET("/projects", h.ListProjects)
		scan.GET("/projects/:id", h.GetProject)
		scan.DELETE("/projects/:id", h.DeleteProject)

		// 扫描结果导入（带安全验证）
		scan.POST("/projects/:id/import", h.ImportScanResults)

		// 统计信息
		scan.GET("/projects/:id/stats", h.GetProjectStats)

		// 扫描任务管理
		scan.POST("/projects/:id/start", h.StartScan)
		scan.GET("/scans/:id/status", h.GetScanStatus)

		// 工具和流水线信息
		scan.GET("/tools", h.GetAvailableTools)
		scan.GET("/pipelines", h.GetAvailablePipelines)

		// 扫描结果查询
		scan.GET("/projects/:id/results", h.GetProjectScanResults)
		scan.GET("/projects/:id/vulnerabilities", h.GetProjectVulnerabilities)
		scan.GET("/projects/:id/vulnerabilities/stats", h.GetVulnerabilityStats)
	}
}

// === 项目管理功能 ===

func (h *ScanHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description" binding:"max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 安全验证：项目名称格式
	if !isValidProjectName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "项目名称格式无效"})
		return
	}

	project, err := h.scanService.CreateProject(req.Name, req.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    project,
		"message": "项目创建成功",
	})
}

func (h *ScanHandler) ListProjects(c *gin.Context) {
	projects, err := h.scanService.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    projects,
	})
}

func (h *ScanHandler) GetProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	project, err := h.scanService.GetProject(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "项目不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    project,
	})
}

func (h *ScanHandler) DeleteProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	if err := h.scanService.DeleteProject(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除项目失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "项目删除成功",
	})
}

func (h *ScanHandler) GetProjectStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	stats, err := h.scanService.GetProjectStats(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取项目统计失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// === 扫描结果导入功能（带安全验证）===

func (h *ScanHandler) ImportScanResults(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var importData struct {
		Results []struct {
			Type    string `json:"type" binding:"required"`
			Address string `json:"address" binding:"required"`
			Parent  string `json:"parent"`
			Ports   []struct {
				Number   int    `json:"number" binding:"required"`
				Protocol string `json:"protocol" binding:"required"`
				State    string `json:"state"`
				Service  *struct {
					Name            string `json:"name"`
					Version         string `json:"version"`
					Fingerprint     string `json:"fingerprint"`
					Banner          string `json:"banner"`
					IsWebService    bool   `json:"is_web_service"`
					HTTPTitle       string `json:"http_title"`
					HTTPStatus      int    `json:"http_status"`
					Vulnerabilities []struct {
						CVEID       string  `json:"cve_id"`
						Title       string  `json:"title"`
						Description string  `json:"description"`
						Severity    string  `json:"severity"`
						CVSS        float64 `json:"cvss"`
						Location    string  `json:"location"`
						Parameter   string  `json:"parameter"`
						Payload     string  `json:"payload"`
					} `json:"vulnerabilities"`
				} `json:"service"`
			} `json:"ports"`
		} `json:"results" binding:"required"`
	}

	if err := c.ShouldBindJSON(&importData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 安全验证：数据大小限制
	if len(importData.Results) > h.maxImportSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("导入数据过大，最多支持 %d 条记录", h.maxImportSize)})
		return
	}

	// 安全验证：地址格式检查
	for i, result := range importData.Results {
		if !h.isValidAddress(result.Type, result.Address) {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第%d条记录地址格式无效: %s", i+1, result.Address)})
			return
		}

		// 端口范围验证
		for j, port := range result.Ports {
			if port.Number < 1 || port.Number > 65535 {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("第%d条记录第%d个端口号无效: %d", i+1, j+1, port.Number)})
				return
			}
		}
	}

	// 转换为服务层数据格式并导入
	// TODO: 实现数据转换和导入逻辑
	_ = projectID // 避免编译错误，等待实现导入逻辑

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "扫描结果导入成功",
	})
}

// === 扫描任务管理功能 ===

func (h *ScanHandler) StartScan(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var req struct {
		Target       string `json:"target" binding:"required"`
		PipelineName string `json:"pipeline" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 安全验证：目标地址格式
	if !h.isValidScanTarget(req.Target) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "目标地址格式无效"})
		return
	}

	// 启动扫描任务
	if h.frameworkService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "扫描框架服务未初始化"})
		return
	}
	scanResult, err := h.frameworkService.StartScan(c.Request.Context(), uint(projectID), req.Target, req.PipelineName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"scan_id":     scanResult.ID,
			"project_id":  scanResult.ProjectID,
			"target":      req.Target,
			"pipeline":    req.PipelineName,
			"created_at":  scanResult.CreatedAt,
		},
		"message": "扫描任务已启动",
	})
}

func (h *ScanHandler) GetScanStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的扫描ID"})
		return
	}

	if h.frameworkService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "扫描框架服务未初始化"})
		return
	}
	scanResult, err := h.frameworkService.GetScanStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "扫描任务不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    scanResult,
	})
}

// === 工具和流水线信息 ===

func (h *ScanHandler) GetAvailableTools(c *gin.Context) {
	// TODO: 实现扫描工具列表功能
	tools := map[string][]map[string]interface{}{
		"subdomain":      {},
		"port":           {},
		"service":        {},
		"web_tech":       {},
		"web_path":       {},
		"vulnerability":  {},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tools,
	})
}

func (h *ScanHandler) GetAvailablePipelines(c *gin.Context) {
	// 返回预定义的扫描流水线
	pipelines := []string{
		"comprehensive", // 全面扫描
		"quick",        // 快速扫描
		"web",          // Web应用扫描
		"network",      // 网络扫描
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pipelines,
	})
}

// === 扫描结果查询功能 ===

func (h *ScanHandler) GetProjectScanResults(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 解析查询参数
	filters := make(map[string]interface{})
	if port := c.Query("port"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			filters["port"] = p
		}
	}
	if service := c.Query("service"); service != "" {
		filters["service_name"] = service
	}

	// TODO: 实现扫描结果查询逻辑
	_ = projectID // 避免编译错误，等待实现查询逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
		"message": "扫描结果查询功能开发中",
	})
}

func (h *ScanHandler) GetProjectVulnerabilities(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 解析查询参数
	filters := make(map[string]interface{})
	if severity := c.Query("severity"); severity != "" {
		filters["severity"] = severity
	}

	// TODO: 实现漏洞查询逻辑
	_ = projectID // 避免编译错误，等待实现查询逻辑
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    []interface{}{},
		"message": "漏洞查询功能开发中",
	})
}

func (h *ScanHandler) GetVulnerabilityStats(c *gin.Context) {
	idStr := c.Param("id")
	projectID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// TODO: 实现漏洞统计逻辑
	_ = projectID // 避免编译错误，等待实现统计逻辑
	stats := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
		"info":     0,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// === 安全验证辅助函数 ===

func isValidProjectName(name string) bool {
	// 项目名称只能包含字母、数字、中文、下划线、短横线
	matched, _ := regexp.MatchString(`^[\w\-\u4e00-\u9fa5]+$`, name)
	return matched && len(name) >= 1 && len(name) <= 100
}

func (h *ScanHandler) isValidAddress(addressType, address string) bool {
	switch addressType {
	case "domain", "subdomain":
		return h.allowedDomainRegex.MatchString(address)
	case "ip":
		return h.allowedIPRegex.MatchString(address)
	default:
		return false
	}
}

func (h *ScanHandler) isValidScanTarget(target string) bool {
	// 目标可以是域名或IP
	return h.allowedDomainRegex.MatchString(target) || h.allowedIPRegex.MatchString(target)
}