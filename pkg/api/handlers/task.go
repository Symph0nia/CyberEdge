package handlers

import (
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

// CreateTask 处理创建通用任务的请求
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var request struct {
		Type    string      `json:"type" binding:"required"`
		Payload interface{} `json:"payload" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := h.taskService.CreateTask(request.Type, request.Payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "任务创建成功"})
}

// StartTask 启动单个任务
func (h *TaskHandler) StartTask(c *gin.Context) {
	// 从URL参数中获取任务ID
	taskID := c.Param("id")

	// 从数据库中获取任务
	task, err := h.taskService.GetTaskByID(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "任务未找到"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取任务失败"})
		}
		return
	}

	// 启动任务
	err = h.taskService.StartTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "启动任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务启动成功"})
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

// DeleteTask 处理删除任务的请求
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")

	if err := h.taskService.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}
