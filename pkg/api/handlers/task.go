// CyberEdge/pkg/api/handles/task.go

package handlers

import (
	"fmt"
	"net/http"
	"time"

	"cyberedge/pkg/task"
	"github.com/gin-gonic/gin"
)

// TaskHandler 处理任务相关的请求
type TaskHandler struct {
	scheduler *task.Scheduler
}

// NewTaskHandler 创建新的任务处理器
func NewTaskHandler(scheduler *task.Scheduler) *TaskHandler {
	return &TaskHandler{scheduler: scheduler}
}

// GetAllTasks 获取所有任务
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	tasks, err := h.scheduler.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}

// GetTask 获取单个任务
func (h *TaskHandler) GetTask(c *gin.Context) {
	id := c.Param("id")
	task, err := h.scheduler.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, task)
}

// StartTask 开始执行单个任务
func (h *TaskHandler) StartTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.scheduler.StartTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已开始执行"})
}

// StopTask 停止单个任务
func (h *TaskHandler) StopTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.scheduler.StopTask(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已停止"})
}

// DeleteTask 删除单个任务
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id := c.Param("id")
	if err := h.scheduler.DeleteTask(id); err != nil {
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

	switch request.Type {
	case "normal":
		task := task.Task{
			ID:          generateID(), // 生成唯一 ID 的函数
			Type:        request.Type,
			Description: request.Description,
			Status:      "scheduled",
			Interval:    request.Interval,
			RunCount:    0,          // 初始化运行次数为0
			CreatedAt:   time.Now(), // 设置创建时间
		}

		if err := h.scheduler.ScheduleTask(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "ping":
		if request.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ping 任务必须提供地址"})
			return
		}

		pingTask := task.Task{
			ID:          generateID(), // 生成唯一 ID 的函数
			Type:        request.Type,
			Description: fmt.Sprintf(request.Address),
			Status:      "scheduled",
			RunCount:    0,          // 初始化运行次数为0
			CreatedAt:   time.Now(), // 设置创建时间
		}

		if err := h.scheduler.ScheduleTask(pingTask); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务类型"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "任务已创建"})
}

// generateID 生成唯一 ID 的函数（示例）
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
