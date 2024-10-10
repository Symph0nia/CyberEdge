package task

import (
	"bytes"
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"os/exec"
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
func (s *Scheduler) ProcessPingTask(task Task) error {
	result, err := ExecutePing(task.Description) // 使用 Description 字段作为目标地址
	if err != nil {
		task.Status = "error" // 更新状态为错误
	} else {
		task.Status = "completed" // 更新状态为完成
	}

	if _, err := s.taskCollection.UpdateOne(context.Background(), bson.M{"id": task.ID}, bson.M{"$set": bson.M{"status": task.Status}}); err != nil {
		return fmt.Errorf("更新 MongoDB 中的 Ping 任务状态失败: %v", err)
	}

	log.Printf("Ping 任务结果:\n%s\n", result)
	return nil
}
