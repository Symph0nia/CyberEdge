package handlers

import (
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		Type     string              `json:"type" binding:"required"`
		Payload  interface{}         `json:"payload" binding:"required"`
		TargetID *primitive.ObjectID `json:"target_id,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := h.taskService.CreateTask(request.Type, request.Payload, request.TargetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建任务失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "任务创建成功"})
}

// StartTasks 批量启动任务
func (h *TaskHandler) StartTasks(c *gin.Context) {
	var request struct {
		TaskIDs []string `json:"taskIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 批量启动任务
	result, err := h.taskService.StartTasks(request.TaskIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量启动任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务启动完成",
		"result":  result,
	})
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

// DeleteTasks 处理批量删除任务的请求
func (h *TaskHandler) DeleteTasks(c *gin.Context) {
	var request struct {
		TaskIDs []string `json:"taskIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.taskService.DeleteTasks(request.TaskIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量删除任务失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "任务删除完成",
		"result":  result,
	})
}
