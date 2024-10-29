// ping.go

package tasks

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"net/http"
	"time"
)

type PingTask struct {
	TaskTemplate
}

func NewPingTask(taskDAO *dao.TaskDAO) *PingTask {
	return &PingTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
	}
}

func (p *PingTask) Handle(ctx context.Context, t *asynq.Task) error {
	return p.Execute(ctx, t, p.doPing)
}

func (p *PingTask) doPing(ctx context.Context, t *asynq.Task) error {
	// 定义一个结构体来解析 JSON 载荷
	var payload struct {
		Target string `json:"target"`
		TaskID string `json:"task_id"`
	}

	// 解析 JSON 载荷
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		logging.Error("解析任务载荷失败: %v", err)
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Target == "" {
		return fmt.Errorf("无效的 Ping 目标")
	}

	logging.Info("开始执行 Ping 任务: %s", payload.Target)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	start := time.Now()
	resp, err := client.Get("http://" + payload.Target)
	duration := time.Since(start)

	if err != nil {
		logging.Error("Ping 失败: %v", err)
		// 更新状态为失败
		_ = p.TaskDAO.UpdateTaskStatus(payload.TaskID, models.TaskStatusFailed, fmt.Sprintf("错误: %v", err))
		return err
	}

	defer resp.Body.Close()

	logging.Info("Ping 成功: %s, 耗时: %v", payload.Target, duration)

	return nil
}
