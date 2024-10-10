package handlers

import (
	"cyberedge/pkg/task"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var taskManager = task.NewTaskManager(taskCollection)

// CreateTaskHandler 创建新任务的处理函数，并保存到MongoDB中
func CreateTaskHandler(c *gin.Context) {
	var json struct {
		Description string        `json:"description"`
		Interval    time.Duration `json:"interval"` // 例如：5s, 1m等
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	taskID := primitive.NewObjectID() // 创建新的 ObjectID

	schedulerTask := task.NewTask(taskID, json.Description, json.Interval) // 使用 scheduler.Task

	if err := taskManager.AddTask(schedulerTask); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法保存任务"})
		return
	}

	schedulerTask.Start() // 启动任务

	c.JSON(http.StatusCreated, gin.H{"message": "任务已创建", "id": taskID.Hex()}) // 返回 ID 的字符串表示形式
}

// GetAllTasksHandler 获取所有任务状态的处理函数，并从MongoDB加载数据
func GetAllTasksHandler(c *gin.Context) {
	tasks, err := taskManager.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取任务"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// GetSingleTaskHandler 获取单个任务状态的处理函数，并从MongoDB加载数据
func GetSingleTaskHandler(c *gin.Context) {
	id := c.Param("id")

	task, err := taskManager.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务未找到"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// StartSingleTaskHandler 启动单个任务的处理函数
func StartSingleTaskHandler(c *gin.Context) {
	id := c.Param("id")

	task, err := taskManager.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务未找到"})
		return
	}

	task.Start() // 启动任务

	c.JSON(http.StatusOK, gin.H{"status": "任务已启动", "id": id})
}

// StopSingleTaskHandler 停止单个任务的处理函数
func StopSingleTaskHandler(c *gin.Context) {
	id := c.Param("id")

	task, err := taskManager.GetTask(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务未找到"})
		return
	}

	task.Stop() // 停止任务

	c.JSON(http.StatusOK, gin.H{"status": "任务已停止", "id": id})
}
