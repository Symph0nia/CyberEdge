package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
)

type TaskService struct {
	taskDAO     *dao.TaskDAO
	asynqClient *asynq.Client
	redisAddr   string // 添加Redis地址存储
}

// NewTaskService 创建一个新的 TaskService 实例
func NewTaskService(taskDAO *dao.TaskDAO, asynqClient *asynq.Client, redisAddr string) *TaskService {
	return &TaskService{
		taskDAO:     taskDAO,
		asynqClient: asynqClient,
		redisAddr:   redisAddr,
	}
}

func (s *TaskService) Close() {
	if s.asynqClient != nil {
		s.asynqClient.Close()
	}
}

// CreateTask 创建一个新的通用任务并保存到数据库
func (s *TaskService) CreateTask(taskType string, payload interface{}, parentID *primitive.ObjectID) error {
	logging.Info("正在创建任务: 类型 %s", taskType)

	var payloadBytes []byte
	var err error

	// 检查 payload 是否已经是字符串
	if payloadStr, ok := payload.(string); ok {
		payloadBytes = []byte(payloadStr)
	} else {
		// 如果不是字符串，则尝试 JSON 序列化
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			logging.Error("序列化任务载荷失败: %v", err)
			return err
		}
	}

	// 创建任务对象
	task := &models.Task{
		Type:     taskType,
		Payload:  string(payloadBytes),
		Status:   models.TaskStatusPending,
		ParentID: parentID,
	}

	// 将任务保存到数据库
	if err := s.taskDAO.CreateTask(task); err != nil {
		logging.Error("创建任务失败: %v", err)
		return err
	}

	logging.Info("成功创建任务: 类型 %s", taskType)
	return nil
}

func (s *TaskService) StartTask(task *models.Task) error {
	logging.Info("正在启动任务: ID %s, 类型 %s", task.ID.Hex(), task.Type)

	payloadMap := map[string]interface{}{
		"task_id": task.ID.Hex(),
		"target":  task.Payload,
	}

	if task.ParentID != nil {
		payloadMap["parent_id"] = task.ParentID.Hex()
	}

	payload, err := json.Marshal(payloadMap)
	if err != nil {
		logging.Error("序列化任务载荷失败: %v", err)
		return err
	}

	// 创建 Asynq 任务
	asynqTask := asynq.NewTask(task.Type, payload)

	// 将任务加入队列
	_, err = s.asynqClient.Enqueue(asynqTask)
	if err != nil {
		logging.Error("将任务加入队列失败: %v", err)
		return err
	}

	logging.Info("成功将任务加入队列: ID %s, 类型 %s", task.ID.Hex(), task.Type)
	return nil
}

// GetTaskByID 根据ID获取任务
func (s *TaskService) GetTaskByID(id string) (*models.Task, error) {
	logging.Info("正在获取任务: ID %s", id)

	task, err := s.taskDAO.GetTaskByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			logging.Warn("任务不存在: ID %s", id)
			return nil, fmt.Errorf("task not found")
		}
		logging.Error("获取任务失败: ID %s, 错误: %v", id, err)
		return nil, err
	}

	logging.Info("成功获取任务: ID %s", id)
	return task, nil
}

// GetAllTasks 获取所有任务
func (s *TaskService) GetAllTasks() ([]models.Task, error) {
	logging.Info("正在获取所有任务")

	tasks, err := s.taskDAO.GetAllTasks()
	if err != nil {
		logging.Error("获取所有任务失败: %v", err)
		return nil, err
	}

	logging.Info("成功获取所有任务，共 %d 个", len(tasks))
	return tasks, nil
}

func (s *TaskService) DeleteTask(id string) error {
	logging.Info("正在删除任务: %s", id)

	// 1. 先获取任务信息
	task, err := s.taskDAO.GetTaskByID(id)
	if err != nil {
		logging.Error("获取任务信息失败: %s, 错误: %v", id, err)
		return err
	}

	// 2. 从数据库中删除任务
	if err := s.taskDAO.DeleteTask(id); err != nil {
		logging.Error("从数据库删除任务失败: %s, 错误: %v", id, err)
		return err
	}

	// 3. 从Asynq队列中删除任务
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: s.redisAddr,
	})

	// 获取所有可用队列
	queues, err := inspector.Queues()
	if err != nil {
		logging.Error("获取队列列表失败: %v", err)
		return nil // 即使获取队列失败，我们也已经从数据库删除了任务，所以返回nil
	}

	// 构造任务key
	taskKey := fmt.Sprintf("%s", task.ID.Hex()) // 修改任务key的构造方式

	// 遍历所有存在的队列
	for _, queue := range queues {
		err = inspector.DeleteTask(queue, taskKey)
		if err != nil {
			if strings.Contains(err.Error(), "task not found") {
				logging.Info("任务在队列[%s]中未找到: %s", queue, id)
				continue
			}
			logging.Error("从Asynq队列[%s]删除任务失败: %s, 错误: %v", queue, id, err)
			// 继续尝试其他队列，不要因为一个队列失败就中断
			continue
		}
		logging.Info("成功从Asynq队列[%s]删除任务: %s", queue, id)
	}

	logging.Info("成功删除任务: %s", id)
	return nil
}
