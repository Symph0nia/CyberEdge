package handlers

import (
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ResultHandler struct {
	resultService *service.ResultService
	dnsService    *service.DNSService
	httpxService  *service.HTTPXService
}

// NewResultHandler 创建一个新的 ResultHandler 实例
func NewResultHandler(resultService *service.ResultService, dnsService *service.DNSService, httpxService *service.HTTPXService) *ResultHandler {
	return &ResultHandler{
		resultService: resultService,
		dnsService:    dnsService,
		httpxService:  httpxService,
	}
}

// CreateResult 处理创建扫描结果的请求
func (h *ResultHandler) CreateResult(c *gin.Context) {
	var request struct {
		Type    string      `json:"type" binding:"required"`
		Payload interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	result := &models.Result{
		Type:      request.Type,
		Timestamp: time.Now(),
		Data:      request.Payload,
	}

	if err := h.resultService.CreateResult(result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建扫描结果失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "扫描结果创建成功"})
}

// GetResultByID 根据 ID 获取单个扫描结果
func (h *ResultHandler) GetResultByID(c *gin.Context) {
	id := c.Param("id")

	result, err := h.resultService.GetResultByID(id)
	if err != nil {
		if err.Error() == "result not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到该扫描结果"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取扫描结果失败"})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetResultsByType 根据类型获取扫描结果列表
func (h *ResultHandler) GetResultsByType(c *gin.Context) {
	resultType := c.Param("type")

	results, err := h.resultService.GetResultsByType(resultType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取该类型的扫描结果"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// UpdateResult 更新指定 ID 的扫描结果
func (h *ResultHandler) UpdateResult(c *gin.Context) {
	id := c.Param("id")
	var updatedData struct {
		Type    string      `json:"type" binding:"required"`
		Payload interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	existingResult, err := h.resultService.GetResultByID(id)
	if err != nil {
		if err.Error() == "result not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到该扫描结果"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取扫描结果失败"})
		return
	}

	existingResult.Type = updatedData.Type
	existingResult.Data = updatedData.Payload

	if err := h.resultService.UpdateResult(id, existingResult); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新扫描结果失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "扫描结果更新成功"})
}

// DeleteResult 删除指定 ID 的扫描结果
func (h *ResultHandler) DeleteResult(c *gin.Context) {
	id := c.Param("id")

	if err := h.resultService.DeleteResult(id); err != nil {
		if err.Error() == "result not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到该扫描结果"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除扫描结果失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "扫描结果已删除"})
}

// MarkResultAsRead 根据任务 ID 修改任务的已读状态（支持已读/未读切换）
func (h *ResultHandler) MarkResultAsRead(c *gin.Context) {
	resultID := c.Param("id")

	// 从请求体获取新的 isRead 状态
	var request struct {
		IsRead bool `json:"isRead"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	// 调用服务层的 MarkResultAsRead 方法，传入 resultID 和新的 isRead 状态
	if err := h.resultService.MarkResultAsRead(resultID, request.IsRead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新任务的已读状态"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已成功标记为已读/未读"})
}

// MarkEntryAsRead 根据任务 ID 和条目 ID 修改条目的已读状态（支持已读/未读切换）
func (h *ResultHandler) MarkEntryAsRead(c *gin.Context) {
	resultID := c.Param("id")
	entryID := c.Param("entry_id")

	// 从请求体获取 isRead 状态
	var request struct {
		IsRead bool `json:"isRead"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	// 调用服务层方法，传入 isRead 状态
	if err := h.resultService.MarkEntryAsRead(resultID, entryID, request.IsRead); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新条目的已读状态"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "条目已成功标记为已读/未读"})
}

// ResolveSubdomainIPHandler 处理子域名 IP 解析请求
func (h *ResultHandler) ResolveSubdomainIPHandler(c *gin.Context) {
	// 从 URL 参数中获取 resultID 和 entryID
	resultID := c.Param("id")
	entryID := c.Param("entry_id")

	// 调用 DNSService 层方法进行子域名 IP 解析和更新
	err := h.dnsService.ResolveAndUpdateSubdomainIP(resultID, entryID)
	if err != nil {
		logging.Error("解析子域名 IP 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "解析子域名 IP 失败",
			"detail": err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "子域名 IP 解析成功并更新",
	})
}

// BatchResolveSubdomainIPHandler 处理批量子域名 IP 解析请求
func (h *ResultHandler) BatchResolveSubdomainIPHandler(c *gin.Context) {
	resultID := c.Param("id")

	var request struct {
		EntryIDs []string `json:"entryIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	result, err := h.dnsService.BatchResolveAndUpdateSubdomainIP(resultID, request.EntryIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "批量解析子域名IP失败",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量解析完成",
		"result":  result,
	})
}

func (h *ResultHandler) ProbeSubdomainHandler(c *gin.Context) {
	resultID := c.Param("id")
	entryID := c.Param("entry_id")

	err := h.httpxService.ProbeAndUpdateSubdomain(resultID, entryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "HTTP探测失败",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "HTTP探测成功",
	})
}

func (h *ResultHandler) BatchProbeSubdomainHandler(c *gin.Context) {
	resultID := c.Param("id")

	var request struct {
		EntryIDs []string `json:"entryIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	result, err := h.httpxService.BatchProbeAndUpdateSubdomains(resultID, request.EntryIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "批量HTTP探测失败",
			"detail": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "批量HTTP探测完成",
		"result":  result,
	})
}
