// task_template.go

package tasks

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
)

// TaskTemplate 提供通用任务处理逻辑的模板
type TaskTemplate struct {
	TaskDAO *dao.TaskDAO
}

func (t *TaskTemplate) Execute(ctx context.Context, task *asynq.Task, handler func(context.Context, *asynq.Task) error) error {
	var payload map[string]string
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		logging.Error("解析任务载荷失败: %v", err)
		return err
	}

	taskID, ok := payload["task_id"]
	if !ok {
		logging.Error("任务载荷中缺少 task_id")
		return fmt.Errorf("缺少 task_id")
	}

	// 更新任务状态为进行中
	if err := t.TaskDAO.UpdateTaskStatus(taskID, models.TaskStatusRunning, ""); err != nil {
		logging.Error("更新任务状态为进行中失败: %s, 错误: %v", taskID, err)
		return err
	}

	// 执行具体的任务处理逻辑
	err := handler(ctx, task)

	if err != nil {
		// 如果任务执行失败，更新状态为失败并记录错误信息
		if updateErr := t.TaskDAO.UpdateTaskStatus(taskID, models.TaskStatusFailed, err.Error()); updateErr != nil {
			logging.Error("更新任务状态为失败时出错: %s, 错误: %v", taskID, updateErr)
		}
		return err
	}

	// 任务成功完成，更新状态为完成并记录结果
	if err := t.TaskDAO.UpdateTaskStatus(taskID, models.TaskStatusCompleted, "成功完成"); err != nil {
		logging.Error("更新任务状态为完成失败: %s, 错误: %v", taskID, err)
		return err
	}

	return nil
}
