// CyberEdge/pkg/task/ping.go

package task

import (
	"bytes"
	"context"
	"cyberedge/pkg/models"
	"fmt"
	"os/exec"
	"time"

	"cyberedge/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
)

// ExecutePing 执行 Ping 命令并返回结果
func ExecutePing(address string) (string, error) {
	cmd := exec.Command("ping", "-c", "4", address) // Linux/MacOS 使用 -c，Windows 使用 -n
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("执行 ping 命令失败: %v", err)
	}
	return out.String(), nil
}

// ProcessPingTask 处理 Ping 任务逻辑
func ProcessPingTask(s *models.Scheduler, task models.Task) error {
	logging.Info("开始处理Ping任务，任务ID: %s，目标地址: %s", task.ID, task.Description)

	startTime := time.Now()

	result, err := ExecutePing(task.Description)
	if err != nil {
		logging.Error("Ping操作失败，任务ID: %s，目标: %s，错误: %v", task.ID, task.Description, err)
		task.UpdateStatus(models.TaskStatusError)
	} else {
		task.UpdateStatus(models.TaskStatusCompleted)
		logging.Info("Ping操作成功，任务ID: %s，目标: %s", task.ID, task.Description)
	}

	task.IncrementRunCount()

	// 更新 MongoDB 中的任务状态
	update := bson.M{
		"$set": bson.M{
			"status":    task.Status,
			"run_count": task.RunCount,
		},
	}
	if _, err := s.TaskCollection.UpdateOne(context.Background(), bson.M{"_id": task.ID}, update); err != nil {
		logging.Error("更新MongoDB中的Ping任务状态失败，任务ID: %s，错误: %v", task.ID, err)
		return fmt.Errorf("更新MongoDB中的Ping任务状态失败: %v", err)
	}

	elapsedTime := time.Since(startTime)

	logging.Info("Ping任务完成，任务ID: %s，耗时: %v", task.ID, elapsedTime)
	logging.Debug("Ping任务结果:\n%s", result)

	return nil
}
