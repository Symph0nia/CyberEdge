// subfinder_task.go

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

type SubfinderTask struct {
	TaskTemplate
}

func NewSubfinderTask(taskDAO *dao.TaskDAO) *SubfinderTask {
	return &SubfinderTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
	}
}

func (s *SubfinderTask) Handle(ctx context.Context, t *asynq.Task) error {
	return s.Execute(ctx, t, s.runSubfinder)
}

func (s *SubfinderTask) runSubfinder(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Domain string `json:"target"`
		TaskID string `json:"task_id"`
	}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Domain == "" {
		return fmt.Errorf("无效的域名")
	}

	logging.Info("开始执行 Subfinder 任务: %s", payload.Domain)

	cmd := exec.Command("subfinder", "-d", payload.Domain)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 subfinder 命令失败: %v", err)
	}

	result := out.String()
	logging.Info("Subfinder 任务完成，结果: %s", result)

	// 返回结果供 Execute 方法使用
	return nil
}
