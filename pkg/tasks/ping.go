// ping.go

package tasks

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
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
	target := string(t.Payload())

	if target == "" {
		return fmt.Errorf("无效的 Ping 目标")
	}

	logging.Info("开始执行 Ping 任务: %s", target)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	start := time.Now()
	resp, err := client.Get("http://" + target)
	duration := time.Since(start)

	if err != nil {
		logging.Error("Ping 失败: %v", err)
		// 更新状态为失败
		_ = p.TaskDAO.UpdateTaskStatus(t.ResultWriter().TaskID(), models.TaskStatusFailed, fmt.Sprintf("错误: %v", err))
		return err
	}

	defer resp.Body.Close()

	logging.Info("Ping 成功: %s, 耗时: %v", target, duration)

	// 更新状态为完成
	_ = p.TaskDAO.UpdateTaskStatus(t.ResultWriter().TaskID(), models.TaskStatusCompleted, fmt.Sprintf("耗时: %v", duration))

	return nil
}
