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
	"strings"
	"time"
)

type SubfinderTask struct {
	TaskTemplate
	resultDAO *dao.ResultDAO
}

func NewSubfinderTask(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *SubfinderTask {
	return &SubfinderTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
		resultDAO:    resultDAO,
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

	cmd := exec.Command("subfinder", "-d", payload.Domain, "-silent")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 subfinder 命令失败: %v", err)
	}

	result := out.String()
	logging.Info("Subfinder 任务完成，结果: %s", result)

	// 解析结果并存储到数据库
	subdomains := strings.Split(strings.TrimSpace(result), "\n")
	target := &models.Target{
		Identifier: payload.Domain,
		Type:       "Subdomain",
		LastScan:   time.Now(),
	}

	for _, subdomain := range subdomains {
		if subdomain != "" {
			target.Subdomains = append(target.Subdomains, &models.Subdomain{Name: subdomain})
		}
	}

	scanResult := &models.Result{
		Targets: []*models.Target{target},
	}

	if err := s.resultDAO.CreateResult(scanResult); err != nil {
		logging.Error("存储扫描结果失败: %v", err)
		return err
	}

	return nil
}
