package tasks

import (
	"bytes"
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"os/exec"
)

type HttpxTask struct {
	TaskTemplate
}

func NewHttpxTask(taskDAO *dao.TaskDAO) *HttpxTask {
	return &HttpxTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
	}
}

func (h *HttpxTask) Handle(ctx context.Context, t *asynq.Task) error {
	return h.Execute(ctx, t, h.runHttpx)
}

func (h *HttpxTask) runHttpx(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Target string `json:"target"`
		TaskID string `json:"task_id"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Target == "" {
		return fmt.Errorf("无效的目标")
	}

	logging.Info("开始执行 Httpx 任务: %s", payload.Target)

	cmd := exec.Command("httpx", "-u", payload.Target, "-sc", "-cl", "-ct", "-location", "-rt", "-title", "-method", "-ip", "-cname")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 httpx 命令失败: %v", err)
	}

	result := out.String()
	logging.Info("Httpx 任务完成，结果: %s", result)

	// 返回结果供 Execute 方法使用
	return nil
}
