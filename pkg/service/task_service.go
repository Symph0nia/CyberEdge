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

// DeleteTask 删除指定的任务
func (s *TaskService) DeleteTask(id string) error {
	logging.Info("正在删除任务: %s", id)

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
		return nil
	}

	// 遍历所有队列
	for _, queue := range queues {
		// 检查各种状态的任务列表
		taskLists := []struct {
			name     string
			listFunc func(string, ...asynq.ListOption) ([]*asynq.TaskInfo, error)
		}{
			{"待处理", inspector.ListPendingTasks},
			{"进行中", inspector.ListActiveTasks},
			{"已调度", inspector.ListScheduledTasks},
			{"重试中", inspector.ListRetryTasks},
			{"已归档", inspector.ListArchivedTasks},
			{"已完成", inspector.ListCompletedTasks},
		}

		// 遍历所有状态的任务
		for _, tl := range taskLists {
			tasks, err := tl.listFunc(queue)
			if err != nil {
				logging.Error("获取队列[%s]%s任务列表失败: %v", queue, tl.name, err)
				continue
			}

			// 遍历任务找到匹配的任务ID
			for _, t := range tasks {
				var payloadMap map[string]interface{}
				if err := json.Unmarshal(t.Payload, &payloadMap); err != nil {
					continue
				}

				// 检查任务ID是否匹配
				if taskID, ok := payloadMap["task_id"].(string); ok && taskID == id {
					err = inspector.DeleteTask(queue, t.ID)
					if err != nil {
						logging.Error("从队列[%s]删除%s任务失败: %v", queue, tl.name, err)
						continue
					}
					logging.Info("成功从队列[%s]删除%s任务: %s", queue, tl.name, id)
				}
			}
		}
	}

	logging.Info("成功删除任务: %s", id)
	return nil
}
