package tasks

import (
	"bytes"
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NaabuTask struct {
	TaskTemplate
	resultDAO *dao.ResultDAO
}

func NewNaabuTask(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *NaabuTask {
	return &NaabuTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
		resultDAO:    resultDAO,
	}
}

func (n *NaabuTask) Handle(ctx context.Context, t *asynq.Task) error {
	return n.Execute(ctx, t, n.runNaabu)
}

func (n *NaabuTask) runNaabu(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Host     string `json:"target"`
		TaskID   string `json:"task_id"`
		ParentID string `json:"parent_id,omitempty"` // 可选参数，用于关联现有记录
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Host == "" {
		return fmt.Errorf("无效的主机地址")
	}

	logging.Info("开始执行 Naabu 任务: %s", payload.Host)

	cmd := exec.Command("naabu", "-host", payload.Host)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 naabu 命令失败: %v", err)
	}

	result := out.String()
	logging.Info("Naabu 任务完成，结果: %s", result)

	portList := make([]*models.Port, 0)
	for _, portStr := range strings.Split(strings.TrimSpace(result), "\n") {
		if portStr != "" {
			hostPort := strings.Split(portStr, ":")
			if len(hostPort) != 2 {
				logging.Warn("无效的端口格式: %s", portStr)
				continue
			}

			portNumber, err := strconv.Atoi(hostPort[1]) // 提取并解析端口号
			if err != nil {
				logging.Warn("解析端口失败: %s, 错误: %v", portStr, err)
				continue // 如果解析失败，则跳过该端口
			}

			portList = append(portList, &models.Port{
				Number:   portNumber,
				Protocol: "tcp",     // 假设使用 TCP 协议，可以根据需要调整
				Service:  "unknown", // 可以根据需要设置服务名称
			})
		}
	}

	portData := &models.PortData{
		Ports: portList,
	}

	var parentID *primitive.ObjectID
	if payload.ParentID != "" {
		objID, err := primitive.ObjectIDFromHex(payload.ParentID)
		if err == nil {
			parentID = &objID
		}
	}

	scanResult := &models.Result{
		ID:        primitive.NewObjectID(),
		Type:      "Port",
		Target:    payload.Host,
		Timestamp: time.Now(),
		Data:      portData,
		ParentID:  parentID,
	}

	if err := n.resultDAO.CreateResult(scanResult); err != nil {
		logging.Error("存储扫描结果失败: %v", err)
		return err
	}

	return nil
}
