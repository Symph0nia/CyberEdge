// naabu_task.go

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

type NaabuTask struct {
	TaskTemplate
}

func NewNaabuTask(taskDAO *dao.TaskDAO) *NaabuTask {
	return &NaabuTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
	}
}

func (n *NaabuTask) Handle(ctx context.Context, t *asynq.Task) error {
	return n.Execute(ctx, t, n.runNaabu)
}

func (n *NaabuTask) runNaabu(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Host   string `json:"target"`
		TaskID string `json:"task_id"`
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

	// 返回结果供 Execute 方法使用
	return nil
}
