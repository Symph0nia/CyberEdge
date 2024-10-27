package handlers

import (
	"cyberedge/pkg/factory"
	"cyberedge/pkg/models"
	"cyberedge/pkg/service/task"
	"github.com/gin-gonic/gin"
	"net/http"
)

// TaskHandler 处理任务相关的请求
type TaskHandler struct {
	taskService *task.TaskService
}

// NewTaskHandler 创建新的任务处理器
func NewTaskHandler(scheduler *models.Scheduler) *TaskHandler {
	return &TaskHandler{
		taskService: task.NewTaskService(scheduler),
	}
}

// GetAllTasks 获取所有任务
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTask 获取单个任务
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := h.taskService.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// StartTask 开始执行单个任务
func (h *TaskHandler) StartTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.taskService.StartTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已开始执行"})
}

// StopTask 停止单个任务
func (h *TaskHandler) StopTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.taskService.StopTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已停止"})
}

// DeleteTask 删除单个任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.taskService.DeleteTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}

// CreateTask 创建新任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var request struct {
		Type        string `json:"type"`
		Description string `json:"description"`
		Interval    int    `json:"interval"`
		Address     string `json:"address"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	var newTask *models.Task

	if request.Type != "ping" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务类型，只支持 ping 类型任务"})
		return
	}

	if request.Address == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ping 任务必须提供地址"})
		return
	}

	newTask = factory.CreatePingTask(request.Address, request.Interval)

	if err := h.taskService.ScheduleTask(*newTask); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "任务已创建", "task": newTask})
}
