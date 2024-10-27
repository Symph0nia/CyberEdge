package handlers

import (
	"cyberedge/pkg/models"
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler 创建一个新的 TaskHandler 实例
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// CreatePingTask 处理创建 Ping 任务的请求
func (h *TaskHandler) CreatePingTask(c *gin.Context) {
	var request struct {
		Target string `json:"target" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := h.taskService.CreatePingTask(request.Target); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Ping 任务创建成功"})
}

// GetAllTasks 处理获取所有任务的请求
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取任务"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// UpdateTaskStatus 处理更新任务状态的请求
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id := c.Param("id")
	var request struct {
		Status models.TaskStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := h.taskService.UpdateTaskStatus(id, request.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新任务状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务状态更新成功"})
}

// DeleteTask 处理删除任务的请求
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.taskService.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}
