package service

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
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
func (s *TaskService) CreateTask(taskType string, payload interface{}, targetID *primitive.ObjectID) error {
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
		TargetID: targetID,
	}

	// 将任务保存到数据库
	if err := s.taskDAO.CreateTask(task); err != nil {
		logging.Error("创建任务失败: %v", err)
		return err
	}

	logging.Info("成功创建任务: 类型 %s", taskType)
	return nil
}

type StartTaskResult struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"` // taskId -> error message
}

func (s *TaskService) StartTasks(taskIDs []string) (*StartTaskResult, error) {
	logging.Info("正在批量启动任务: %v", taskIDs)

	result := &StartTaskResult{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
	}

	// 批量获取任务信息
	tasks, err := s.taskDAO.GetTasksByIDs(taskIDs)
	if err != nil {
		logging.Error("批量获取任务信息失败: %v", err)
		return nil, err
	}

	// 将找到的任务ID映射到任务对象
	taskMap := make(map[string]*models.Task)
	for _, task := range tasks {
		taskMap[task.ID.Hex()] = task
	}

	// 创建批量任务
	taskGroups := make(map[string][]*asynq.Task)

	for _, taskID := range taskIDs {
		task, exists := taskMap[taskID]
		if !exists {
			result.Failed[taskID] = "任务未找到"
			continue
		}

		payloadMap := map[string]interface{}{
			"task_id": task.ID.Hex(),
			"target":  task.Payload,
		}

		if task.TargetID != nil {
			payloadMap["target_id"] = task.TargetID.Hex()
		}

		payload, err := json.Marshal(payloadMap)
		if err != nil {
			result.Failed[taskID] = "序列化任务载荷失败"
			continue
		}

		// 按任务类型分组
		asynqTask := asynq.NewTask(task.Type, payload)
		taskGroups[task.Type] = append(taskGroups[task.Type], asynqTask)
	}

	// 批量提交任务到队列
	for taskType, tasks := range taskGroups {
		// 使用事务批量提交同类型的任务
		err := s.enqueueBatch(context.Background(), tasks)
		if err != nil {
			logging.Error("批量提交任务类型 %s 失败: %v", taskType, err)
			// 查找失败的任务ID
			for _, task := range tasks {
				var payloadMap map[string]interface{}
				if err := json.Unmarshal(task.Payload(), &payloadMap); err != nil {
					continue
				}
				if taskID, ok := payloadMap["task_id"].(string); ok {
					result.Failed[taskID] = "加入队列失败"
				}
			}
			continue
		}

		// 记录成功的任务
		for _, task := range tasks {
			var payloadMap map[string]interface{}
			if err := json.Unmarshal(task.Payload(), &payloadMap); err != nil {
				continue
			}
			if taskID, ok := payloadMap["task_id"].(string); ok {
				result.Success = append(result.Success, taskID)
			}
		}
	}

	logging.Info("批量启动任务完成，成功: %d, 失败: %d",
		len(result.Success), len(result.Failed))

	return result, nil
}

// enqueueBatch 批量将任务加入队列
func (s *TaskService) enqueueBatch(ctx context.Context, tasks []*asynq.Task) error {
	// 不使用 Pipeline，直接批量提交任务
	for _, task := range tasks {
		if _, err := s.asynqClient.EnqueueContext(ctx, task); err != nil {
			return err
		}
	}
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

type DeleteTaskResult struct {
	Success []string          `json:"success"`
	Failed  map[string]string `json:"failed"` // taskId -> error message
}

// DeleteTasks 批量删除任务
func (s *TaskService) DeleteTasks(ids []string) (*DeleteTaskResult, error) {
	logging.Info("正在批量删除任务: %v", ids)

	result := &DeleteTaskResult{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
	}

	// 1. 从数据库中批量删除任务
	dbResult, err := s.taskDAO.DeleteTasks(ids)
	if err != nil {
		logging.Error("批量删除任务失败: %v", err)
		return nil, err
	}

	// 2. 从Asynq队列中删除任务
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: s.redisAddr,
	})

	// 获取所有可用队列
	queues, err := inspector.Queues()
	if err != nil {
		logging.Error("获取队列列表失败: %v", err)
		// 继续处理，因为数据库删除可能已经成功
	} else {
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

			// 创建任务ID到状态的映射
			idMap := make(map[string]bool)
			for _, id := range ids {
				idMap[id] = true
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

					if taskID, ok := payloadMap["task_id"].(string); ok && idMap[taskID] {
						err = inspector.DeleteTask(queue, t.ID)
						if err != nil {
							logging.Error("从队列[%s]删除%s任务失败: %v", queue, tl.name, err)
							continue
						}
						logging.Info("成功从队列[%s]删除%s任务: %s", queue, tl.name, taskID)
					}
				}
			}
		}
	}

	// 合并数据库删除结果
	result.Success = dbResult.DeletedIDs
	result.Failed = dbResult.FailedIDs

	logging.Info("批量删除任务完成，成功: %d, 失败: %d",
		len(result.Success), len(result.Failed))

	return result, nil
}
