package tasks

import (
	"context"
	"cyberedge/pkg/logging"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"net/http"
	"time"
)

const TaskTypePing = "ping"

// PingPayload 定义 Ping 任务的负载结构
type PingPayload struct {
	Target string `json:"target"`
}

// NewPingTask 创建新的 Ping 任务
func NewPingTask(target string) (*asynq.Task, error) {
	payload, err := json.Marshal(PingPayload{Target: target})
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TaskTypePing, payload), nil
}

// HandlePingTask 处理 Ping 任务
func HandlePingTask(ctx context.Context, t *asynq.Task) error {
	var p PingPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	logging.Info("开始执行 Ping 任务: %s", p.Target)

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	start := time.Now()
	resp, err := client.Get("http://" + p.Target)
	duration := time.Since(start)

	if err != nil {
		logging.Error("Ping 失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	logging.Info("Ping 成功: %s, 耗时: %v", p.Target, duration)
	// 这里可以更新任务状态到 MongoDB
	return nil
}

// TaskHandler 结构体用于实现 asynq.Handler 接口
type TaskHandler struct{}

// NewTaskHandler 创建一个新的 TaskHandler
func NewTaskHandler() *TaskHandler {
	return &TaskHandler{}
}

// ProcessTask 处理任务的方法
func (h *TaskHandler) ProcessTask(ctx context.Context, task *asynq.Task) error {
	logging.Info("开始处理任务: %s", task.Type())

	var err error
	switch task.Type() {
	case TaskTypePing:
		err = HandlePingTask(ctx, task)
		// 可以添加其他任务类型的处理...
	default:
		err = fmt.Errorf("未知任务类型: %s", task.Type())
	}

	if err != nil {
		logging.Error("处理任务失败 [%s]: %v", task.Type(), err)
		return err
	}

	logging.Info("任务处理完成: %s", task.Type())
	return nil
}
