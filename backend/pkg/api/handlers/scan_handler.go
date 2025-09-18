package handlers

import (
	"net/http"
	"strconv"

	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
)

type ScanHandler struct {
	scanService *service.ScanService
}

func NewScanHandler(scanService *service.ScanService) *ScanHandler {
	return &ScanHandler{
		scanService: scanService,
	}
}

// RegisterScanRoutes 注册扫描相关路由
func (h *ScanHandler) RegisterScanRoutes(router *gin.RouterGroup) {
	scan := router.Group("/scan")
	{
		// 项目管理
		scan.POST("/projects", h.CreateProject)
		scan.GET("/projects", h.ListProjects)
		scan.GET("/projects/:id", h.GetProject)
		scan.DELETE("/projects/:id", h.DeleteProject)

		// 扫描结果导入
		scan.POST("/projects/:id/import", h.ImportScanResults)

		// 统计信息
		scan.GET("/projects/:id/stats", h.GetProjectStats)

		// 示例数据
		scan.POST("/projects/:id/sample", h.CreateSampleData)
	}
}

// CreateProject 创建扫描项目
func (h *ScanHandler) CreateProject(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=1,max=100"`
		Description string `json:"description" binding:"max=500"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效: " + err.Error()})
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

// ListProjects 获取项目列表
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

// GetProject 获取项目详情
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

// DeleteProject 删除项目
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

// ImportScanResults 导入扫描结果
func (h *ScanHandler) ImportScanResults(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	var scanData service.ScanResultData
	if err := c.ShouldBindJSON(&scanData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "扫描数据格式无效: " + err.Error()})
		return
	}

	// 基本验证
	if len(scanData.Results) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "扫描结果不能为空"})
		return
	}

	if err := h.scanService.ImportScanResults(uint(id), &scanData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入扫描结果失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "扫描结果导入成功",
		"count":   len(scanData.Results),
	})
}

// GetProjectStats 获取项目统计信息
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

// CreateSampleData 创建示例数据
func (h *ScanHandler) CreateSampleData(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的项目ID"})
		return
	}

	if err := h.scanService.CreateSampleData(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建示例数据失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "示例数据创建成功",
	})
}