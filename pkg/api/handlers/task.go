// CyberEdge/pkg/api/handlers/task.go

package handlers

import (
	"cyberedge/pkg/service/task"
	"fmt"
	"net/http"
	"time"

	"cyberedge/pkg/models"
	"github.com/gin-gonic/gin"
)

// TaskHandler 处理任务相关的请求
type TaskHandler struct {
	scheduler *models.Scheduler
}

// NewTaskHandler 创建新的任务处理器
func NewTaskHandler(scheduler *models.Scheduler) *TaskHandler {
	return &TaskHandler{scheduler: scheduler}
}

// GetAllTasks 获取所有任务
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := task.GetAllTasks(h.scheduler)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTask 获取单个任务
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := task.GetTask(h.scheduler, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// StartTask 开始执行单个任务
func (h *TaskHandler) StartTask(c *gin.Context) {
	id := c.Param("id")
	if err := task.StartTask(h.scheduler, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已开始执行"})
}

// StopTask 停止单个任务
func (h *TaskHandler) StopTask(c *gin.Context) {
	id := c.Param("id")
	if err := task.StopTask(h.scheduler, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已停止"})
}

// DeleteTask 删除单个任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := task.DeleteTask(h.scheduler, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已删除"})
}

// CreateTask 创建新任务
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var request struct {
		Type        string `json:"type"`        // 任务类型
		Description string `json:"description"` // 任务描述
		Interval    int    `json:"interval"`    // 运行间隔（分钟）
		Address     string `json:"address"`     // Ping 任务的目标地址
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	var newTask models.Task

	switch request.Type {
	case "normal":
		newTask = models.Task{
			ID:          generateID(),
			Type:        models.TaskType(request.Type),
			Description: request.Description,
			Status:      models.TaskStatusScheduled,
			Interval:    request.Interval,
			RunCount:    0,
			CreatedAt:   time.Now(),
		}

	case "ping":
		if request.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ping 任务必须提供地址"})
			return
		}

		newTask = models.Task{
			ID:          generateID(),
			Type:        models.TaskType(request.Type),
			Description: request.Address,
			Status:      models.TaskStatusScheduled,
			RunCount:    0,
			CreatedAt:   time.Now(),
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务类型"})
		return
	}

	if err := task.ScheduleTask(h.scheduler, newTask); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "任务已创建"})
}

// generateID 生成唯一 ID 的函数（示例）
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
