package task

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"sync"
)

// TaskProcessor 任务处理器结构
type TaskProcessor struct {
	scheduler    *models.Scheduler
	taskService  *TaskService
	runningTasks map[string]context.CancelFunc
	taskMutex    sync.Mutex
}

// NewTaskProcessor 创建新的任务处理器
func NewTaskProcessor(s *models.Scheduler, ts *TaskService) *TaskProcessor {
	return &TaskProcessor{
		scheduler:    s,
		taskService:  ts,
		runningTasks: make(map[string]context.CancelFunc),
	}
}

// StartTaskProcessor 启动任务处理器，监听 RabbitMQ 队列并处理消息
func (tp *TaskProcessor) StartTaskProcessor() {
	msgs, err := tp.scheduler.AMQPChannel.Consume(
		tp.scheduler.QueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logging.Fatal("无法启动RabbitMQ消费者: %v", err)
	}

	logging.Info("任务处理器已启动，正在监听队列: %s", tp.scheduler.QueueName)

	for msg := range msgs {
		var task models.Task
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			logging.Error("解码任务失败: %v", err)
			continue
		}

		logging.Info("收到新任务，任务ID: %s，类型: %s", task.ID, task.Type)

		// 创建一个可取消的上下文
		ctx, cancel := context.WithCancel(context.Background())

		// 将取消函数存储到 runningTasks 中
		tp.taskMutex.Lock()
		tp.runningTasks[task.ID] = cancel
		tp.taskMutex.Unlock()

		// 启动一个新的 goroutine 来处理任务
		go func(task models.Task) {
			defer tp.removeRunningTask(task.ID)

			// 根据任务类型处理任务
			switch task.Type {
			case models.TaskTypePing:
				if err := ProcessPingTask(ctx, tp.scheduler, task); err != nil {
					logging.Error("处理Ping任务失败，任务ID: %s，错误: %v", task.ID, err)
				}
			default:
				logging.Warn("未知任务类型，任务ID: %s，类型: %s", task.ID, task.Type)
			}

			// 任务完成后，更新任务状态
			tp.taskService.UpdateTaskStatus(task.ID, models.TaskStatusCompleted)
		}(task)
	}
}

// StopTask 停止指定的任务
func (tp *TaskProcessor) StopTask(taskID string) {
	tp.taskMutex.Lock()
	defer tp.taskMutex.Unlock()

	if cancel, exists := tp.runningTasks[taskID]; exists {
		cancel() // 调用取消函数来停止任务
		logging.Info("已发送停止信号到任务，任务ID: %s", taskID)
	} else {
		logging.Warn("尝试停止不存在或已完成的任务，任务ID: %s", taskID)
	}
}

// removeRunningTask 从 runningTasks 中移除指定的任务
func (tp *TaskProcessor) removeRunningTask(taskID string) {
	tp.taskMutex.Lock()
	defer tp.taskMutex.Unlock()
	delete(tp.runningTasks, taskID)
}
