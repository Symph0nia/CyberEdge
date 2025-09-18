package handlers

import (
	"net/http"
	"strconv"

	"cyberedge/pkg/scanner"
	"cyberedge/pkg/services"
	"github.com/gin-gonic/gin"
)

type ScanFrameworkHandler struct {
	scanService *services.ScanService
}

func NewScanFrameworkHandler(scanService *services.ScanService) *ScanFrameworkHandler {
	return &ScanFrameworkHandler{
		scanService: scanService,
	}
}

// RegisterScanFrameworkRoutes 注册扫描框架相关路由
func (h *ScanFrameworkHandler) RegisterScanFrameworkRoutes(router *gin.RouterGroup) {
	scan := router.Group("/scan-framework")
	{
		// 扫描任务管理
		scan.POST("/start", h.StartScan)
		scan.GET("/status/:id", h.GetScanStatus)

		// 工具和流水线信息
		scan.GET("/tools", h.GetAvailableTools)
		scan.GET("/pipelines", h.GetAvailablePipelines)

		// 扫描结果查询
		scan.GET("/results/project/:id", h.GetProjectScanResults)
		scan.GET("/vulnerabilities/project/:id", h.GetProjectVulnerabilities)
		scan.GET("/vulnerabilities/stats/:id", h.GetVulnerabilityStats)
	}
}

// StartScan 启动扫描任务
func (h *ScanFrameworkHandler) StartScan(c *gin.Context) {
	var req struct {
		ProjectID    uint   `json:"project_id" binding:"required"`
		Target       string `json:"target" binding:"required"`
		PipelineName string `json:"pipeline" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
		return
	}

	// 启动扫描
	scanResult, err := h.scanService.StartScan(c.Request.Context(), req.ProjectID, req.Target, req.PipelineName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"scan_id":     scanResult.ID,
			"project_id":  scanResult.ProjectID,
			"target":      scanResult.Target,
			"status":      scanResult.Status,
			"start_time":  scanResult.StartTime,
			"pipeline":    req.PipelineName,
		},
		"message": "扫描任务已启动",
	})
}

// GetScanStatus 获取扫描状态
func (h *ScanFrameworkHandler) GetScanStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的扫描ID"})
		return
	}

	scanResult, err := h.scanService.GetScanStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "扫描任务不存在"})
		return
	}

	response := gin.H{
		"scan_id":    scanResult.ID,
		"project_id": scanResult.ProjectID,
		"target":     scanResult.Target,
		"scan_type":  scanResult.ScanType,
		"status":     scanResult.Status,
		"start_time": scanResult.StartTime,
	}

	if !scanResult.EndTime.IsZero() {
		response["end_time"] = scanResult.EndTime
		response["duration"] = scanResult.EndTime.Sub(scanResult.StartTime).String()
	}

	if scanResult.ErrorMessage != "" {
		response["error"] = scanResult.ErrorMessage
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetAvailableTools 获取可用扫描工具
func (h *ScanFrameworkHandler) GetAvailableTools(c *gin.Context) {
	tools := h.scanService.GetAvailableTools()

	// 重组数据为更友好的格式
	response := make(map[string][]gin.H)
	for category, scanners := range tools {
		categoryTools := make([]gin.H, 0, len(scanners))
		for _, scanner := range scanners {
			categoryTools = append(categoryTools, gin.H{
				"name":      scanner.Name,
				"available": scanner.Available,
			})
		}
		response[string(category)] = categoryTools
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetAvailablePipelines 获取可用扫描流水线
func (h *ScanFrameworkHandler) GetAvailablePipelines(c *gin.Context) {
	pipelines := h.scanService.GetAvailablePipelines()

	// 获取详细的流水线配置
	pipelineDetails := make([]gin.H, 0, len(pipelines))
	for _, name := range pipelines {
		if pipeline, exists := scanner.GetPipeline(name); exists {
			stages := make([]gin.H, 0, len(pipeline.Stages))
			for _, stage := range pipeline.Stages {
				stages = append(stages, gin.H{
					"name":          stage.Name,
					"scanner_names": stage.ScannerNames,
					"parallel":      stage.Parallel,
					"depends_on":    stage.DependsOn,
				})
			}

			pipelineDetails = append(pipelineDetails, gin.H{
				"name":        pipeline.Name,
				"key":         name,
				"parallel":    pipeline.Parallel,
				"continue_on_error": pipeline.ContinueOnError,
				"stages":      stages,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    pipelineDetails,
	})
}

// GetProjectScanResults 获取项目扫描结果
func (h *ScanFrameworkHandler) GetProjectScanResults(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 查询参数
	scanType := c.Query("scan_type")
	status := c.Query("status")

	// 获取扫描结果
	results, err := h.scanService.GetProjectScanResults(uint(id), scanType, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取扫描结果失败: " + err.Error()})
		return
	}

	// 转换为响应格式
	responseData := make([]gin.H, 0, len(results))
	for _, result := range results {
		item := gin.H{
			"id":           result.ID,
			"target":       result.Target,
			"scan_type":    result.ScanType,
			"scanner_name": result.ScannerName,
			"status":       result.Status,
			"start_time":   result.StartTime,
		}

		if !result.EndTime.IsZero() {
			item["end_time"] = result.EndTime
			item["duration"] = result.EndTime.Sub(result.StartTime).String()
		}

		if result.ErrorMessage != "" {
			item["error"] = result.ErrorMessage
		}

		if result.ScanTarget != nil {
			item["target_type"] = result.ScanTarget.TargetType
		}

		responseData = append(responseData, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
		"count":   len(results),
	})
}

// GetProjectVulnerabilities 获取项目漏洞
func (h *ScanFrameworkHandler) GetProjectVulnerabilities(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 查询参数
	severity := c.Query("severity")

	// 获取漏洞列表
	vulnerabilities, err := h.scanService.GetProjectVulnerabilities(uint(id), severity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取漏洞列表失败: " + err.Error()})
		return
	}

	// 转换为响应格式
	responseData := make([]gin.H, 0, len(vulnerabilities))
	for _, vuln := range vulnerabilities {
		item := gin.H{
			"id":          vuln.ID,
			"cve_id":      vuln.CVEID,
			"title":       vuln.Title,
			"description": vuln.Description,
			"severity":    vuln.Severity,
			"cvss":        vuln.CVSS,
			"location":    vuln.Location,
			"parameter":   vuln.Parameter,
			"payload":     vuln.Payload,
			"created_at":  vuln.CreatedAt,
		}

		if vuln.ScanTarget != nil {
			item["target"] = vuln.ScanTarget.Target
			item["target_type"] = vuln.ScanTarget.TargetType
		}

		responseData = append(responseData, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responseData,
		"count":   len(vulnerabilities),
	})
}

// GetVulnerabilityStats 获取漏洞统计
func (h *ScanFrameworkHandler) GetVulnerabilityStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	// 获取漏洞统计
	stats, err := h.scanService.GetVulnerabilityStats(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取漏洞统计失败: " + err.Error()})
		return
	}

	// 计算总数
	total := 0
	for _, count := range stats {
		total += count
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"project_id": uint(id),
			"stats":      stats,
			"total":      total,
		},
	})
}