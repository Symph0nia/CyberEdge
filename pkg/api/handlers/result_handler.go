package handlers

import (
	"cyberedge/pkg/models"
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type ResultHandler struct {
	resultService *service.ResultService
}

// NewResultHandler 创建一个新的 ResultHandler 实例
func NewResultHandler(resultService *service.ResultService) *ResultHandler {
	return &ResultHandler{resultService: resultService}
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

// MarkEntryAsRead 根据任务 ID 和条目 ID 修改条目的已读状态
func (h *ResultHandler) MarkEntryAsRead(c *gin.Context) {
	resultID := c.Param("result_id")
	entryID := c.Param("entry_id")

	if err := h.resultService.MarkEntryAsRead(resultID, entryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新条目的已读状态"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "条目已成功标记为已读"})
}
