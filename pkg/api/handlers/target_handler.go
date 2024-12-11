package handlers

import (
	"cyberedge/pkg/models"
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TargetHandler struct {
	targetService *service.TargetService
}

// NewTargetHandler 创建一个新的 TargetHandler 实例
func NewTargetHandler(targetService *service.TargetService) *TargetHandler {
	return &TargetHandler{
		targetService: targetService,
	}
}

// CreateTarget 处理创建目标的请求
func (h *TargetHandler) CreateTarget(c *gin.Context) {
	var target models.Target
	if err := c.ShouldBindJSON(&target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := h.targetService.CreateTarget(&target); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建目标失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "目标创建成功", "data": target})
}

// GetTargetByID 根据 ID 获取单个目标
func (h *TargetHandler) GetTargetByID(c *gin.Context) {
	id := c.Param("id")

	target, err := h.targetService.GetTargetByID(id)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到该目标"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目标失败"})
		return
	}

	c.JSON(http.StatusOK, target)
}

// GetAllTargets 获取所有目标
func (h *TargetHandler) GetAllTargets(c *gin.Context) {
	targets, err := h.targetService.GetAllTargets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目标列表失败"})
		return
	}

	c.JSON(http.StatusOK, targets)
}

// UpdateTarget 更新指定 ID 的目标
func (h *TargetHandler) UpdateTarget(c *gin.Context) {
	id := c.Param("id")
	var updatedTarget models.Target

	if err := c.ShouldBindJSON(&updatedTarget); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	// 确保目标存在
	existingTarget, err := h.targetService.GetTargetByID(id)
	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到该目标"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取目标失败"})
		return
	}

	// 保持原有的创建时间
	updatedTarget.CreatedAt = existingTarget.CreatedAt

	if err := h.targetService.UpdateTarget(id, &updatedTarget); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新目标失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "目标更新成功", "data": updatedTarget})
}

// DeleteTarget 删除指定 ID 的目标
func (h *TargetHandler) DeleteTarget(c *gin.Context) {
	id := c.Param("id")

	if err := h.targetService.DeleteTarget(id); err != nil {
		if err.Error() == "mongo: no documents in result" {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到该目标"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除目标失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "目标已删除"})
}
