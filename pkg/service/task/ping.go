package task

import (
	"bytes"
	"context"
	"cyberedge/pkg/models"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"cyberedge/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
)

// ExecutePing 执行单次 Ping 命令并返回结果
func ExecutePing(ctx context.Context, address string) (string, error) {
	cmd := exec.CommandContext(ctx, "ping", "-c", "1", address)
	var out bytes.Buffer
	cmd.Stdout = &out

	errCh := make(chan error, 1)
	go func() {
		errCh <- cmd.Run()
	}()

	select {
	case <-ctx.Done():
		// 立即杀死进程
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "", fmt.Errorf("ping 操作被取消")
	case err := <-errCh:
		if err != nil {
			return "", fmt.Errorf("执行 ping 命令失败: %v", err)
		}
		return out.String(), nil
	}
}

// ProcessPingTask 处理 Ping 任务逻辑
func ProcessPingTask(ctx context.Context, s *models.Scheduler, task models.Task) error {
	logging.Info("开始处理Ping任务，任务ID: %s，目标地址: %s", task.ID, task.Description)

	startTime := time.Now()
	var results []string

	for i := 0; i < 4; i++ {
		select {
		case <-ctx.Done():
			logging.Info("Ping任务被取消，任务ID: %s", task.ID)
			task.UpdateStatus(models.TaskStatusStopped)
			task.SetResult(strings.Join(results, "\n"))
			return updateTaskInDB(s, task)
		default:
			result, err := ExecutePing(ctx, task.Description)
			if err != nil {
				if err.Error() == "ping 操作被取消" {
					logging.Info("Ping操作被取消，任务ID: %s", task.ID)
					task.UpdateStatus(models.TaskStatusStopped)
					task.SetResult(strings.Join(results, "\n"))
					return updateTaskInDB(s, task)
				}
				logging.Error("Ping操作失败，任务ID: %s，目标: %s，错误: %v", task.ID, task.Description, err)
				results = append(results, fmt.Sprintf("Ping失败: %v", err))
			} else {
				logging.Info("Ping操作成功，任务ID: %s，目标: %s", task.ID, task.Description)
				results = append(results, result)
			}
		}

		// 在每次 ping 操作之间添加一秒的延迟，同时检查取消信号
		select {
		case <-ctx.Done():
			logging.Info("Ping任务在等待期间被取消，任务ID: %s", task.ID)
			task.UpdateStatus(models.TaskStatusStopped)
			task.SetResult(strings.Join(results, "\n"))
			return updateTaskInDB(s, task)
		case <-time.After(1 * time.Second):
			// 继续下一次 ping
		}
	}

	task.UpdateStatus(models.TaskStatusCompleted)
	task.IncrementRunCount()
	task.SetResult(strings.Join(results, "\n"))

	elapsedTime := time.Since(startTime)
	logging.Info("Ping任务完成，任务ID: %s，耗时: %v", task.ID, elapsedTime)
	logging.Debug("Ping任务结果:\n%s", task.Result)

	return updateTaskInDB(s, task)
}

// updateTaskInDB 函数保持不变
func updateTaskInDB(s *models.Scheduler, task models.Task) error {
	update := bson.M{
		"$set": bson.M{
			"status":     task.Status,
			"run_count":  task.RunCount,
			"result":     task.Result,
			"updated_at": time.Now(),
		},
	}
	_, err := s.TaskCollection.UpdateOne(context.Background(), bson.M{"_id": task.ID}, update)
	if err != nil {
		logging.Error("更新MongoDB中的Ping任务状态失败，任务ID: %s，错误: %v", task.ID, err)
		return fmt.Errorf("更新MongoDB中的Ping任务状态失败: %v", err)
	}
	return nil
}
