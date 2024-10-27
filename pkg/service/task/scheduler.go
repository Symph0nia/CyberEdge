package task

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TaskService 封装所有任务相关的服务
type TaskService struct {
	scheduler    *models.Scheduler
	taskMutex    sync.Mutex
	runningTasks map[string]context.CancelFunc
}

// NewTaskService 创建新的 TaskService 实例
func NewTaskService(scheduler *models.Scheduler) *TaskService {
	return &TaskService{
		scheduler:    scheduler,
		runningTasks: make(map[string]context.CancelFunc),
	}
}

// ScheduleTask 将任务发送到RabbitMQ队列并存储到MongoDB
func (s *TaskService) ScheduleTask(task models.Task) error {
	task.UpdateStatus(models.TaskStatusWaiting)

	body, err := json.Marshal(task)
	if err != nil {
		logging.Error("无法序列化任务: %v", err)
		return fmt.Errorf("无法序列化任务: %v", err)
	}

	err = s.scheduler.AMQPChannel.Publish("", s.scheduler.QueueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   task.ID,
		Body:        body,
	})
	if err != nil {
		logging.Error("无法发布消息到RabbitMQ队列: %v", err)
		return fmt.Errorf("无法发布消息到队列: %v", err)
	}

	filter := bson.M{"_id": task.ID}
	update := bson.M{"$set": task}
	opts := options.Update().SetUpsert(true)

	_, err = s.scheduler.TaskCollection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		logging.Error("无法插入或更新任务到MongoDB: %v", err)
		return fmt.Errorf("无法插入或更新任务到MongoDB: %v", err)
	}

	logging.Info("成功调度任务: %s", task.ID)
	return nil
}

// GetAllTasks 获取所有任务
func (s *TaskService) GetAllTasks() ([]models.Task, error) {
	var tasks []models.Task
	cursor, err := s.scheduler.TaskCollection.Find(context.Background(), bson.M{})
	if err != nil {
		logging.Error("无法从MongoDB获取所有任务: %v", err)
		return nil, fmt.Errorf("无法获取所有任务: %v", err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var task models.Task
		if err := cursor.Decode(&task); err != nil {
			logging.Error("解码MongoDB任务失败: %v", err)
			return nil, fmt.Errorf("解码任务失败: %v", err)
		}
		tasks = append(tasks, task)
	}

	logging.Info("成功获取所有任务，共 %d 个", len(tasks))
	return tasks, nil
}

// GetTask 获取单个任务
func (s *TaskService) GetTask(id string) (models.Task, error) {
	var task models.Task
	err := s.scheduler.TaskCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&task)
	if err != nil {
		logging.Error("从MongoDB获取任务失败，任务ID: %s, 错误: %v", id, err)
		return models.Task{}, fmt.Errorf("未找到任务: %v", err)
	}
	logging.Info("成功获取任务，任务ID: %s", id)
	return task, nil
}

// StartTask 开始执行单个任务
func (s *TaskService) StartTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	task.UpdateStatus(models.TaskStatusRunning)
	if err := s.ScheduleTask(task); err != nil {
		logging.Error("重新调度任务失败，任务ID: %s, 错误: %v", id, err)
		return fmt.Errorf("重新调度任务失败: %v", err)
	}

	logging.Info("成功开始执行任务，任务ID: %s", id)
	return nil
}

// StopTask 停止单个任务
func (s *TaskService) StopTask(id string) error {
	s.taskMutex.Lock()
	cancelFunc, exists := s.runningTasks[id]
	s.taskMutex.Unlock()

	if exists {
		// 调用取消函数来停止任务
		cancelFunc()
		logging.Info("发送停止信号到任务，任务ID: %s", id)

		// 等待一段时间，确保任务有机会停止
		time.Sleep(1 * time.Second)

		// 从 runningTasks 中移除任务
		s.taskMutex.Lock()
		delete(s.runningTasks, id)
		s.taskMutex.Unlock()
	} else {
		logging.Warn("尝试停止不存在或已完成的任务，任务ID: %s", id)
	}

	// 更新数据库中的任务状态
	update := bson.M{
		"$set": bson.M{
			"status":     models.TaskStatusStopped,
			"updated_at": time.Now(),
		},
	}
	_, err := s.scheduler.TaskCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		update,
	)
	if err != nil {
		logging.Error("更新MongoDB中的任务状态失败，任务ID: %s, 错误: %v", id, err)
		return fmt.Errorf("更新MongoDB中的任务状态失败: %v", err)
	}

	logging.Info("成功停止执行任务，任务ID: %s", id)
	return nil
}

// DeleteTask 删除单个任务
func (s *TaskService) DeleteTask(id string) error {
	_, err := s.scheduler.TaskCollection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		logging.Error("删除MongoDB中的任务失败，任务ID: %s, 错误: %v", id, err)
		return fmt.Errorf("删除MongoDB中的任务失败: %v", err)
	}

	logging.Info("成功删除任务，任务ID: %s", id)
	return nil
}

// CloseScheduler 关闭调度器连接和通道
func (s *TaskService) CloseScheduler() {
	if err := s.scheduler.AMQPChannel.Close(); err != nil {
		logging.Error("关闭RabbitMQ通道失败: %v", err)
	}
	if err := s.scheduler.AMQPConn.Close(); err != nil {
		logging.Error("关闭RabbitMQ连接失败: %v", err)
	}
	logging.Info("成功关闭调度器连接和通道")
}

// UpdateTaskStatus 更新任务状态
func (s *TaskService) UpdateTaskStatus(id string, status models.TaskStatus) error {
	_, err := s.scheduler.TaskCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}},
	)
	if err != nil {
		logging.Error("更新MongoDB中的任务状态失败，任务ID: %s, 错误: %v", id, err)
		return fmt.Errorf("更新MongoDB中的任务状态失败: %v", err)
	}
	logging.Info("成功更新任务状态，任务ID: %s, 新状态: %s", id, status)
	return nil
}
