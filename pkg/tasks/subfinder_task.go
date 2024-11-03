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

	"go.mongodb.org/mongo-driver/bson/primitive"
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
		Domain   string `json:"target"`
		TaskID   string `json:"task_id"`
		ParentID string `json:"parent_id,omitempty"` // 可选参数，用于关联现有记录
	}

	// 解析任务载荷
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Domain == "" {
		return fmt.Errorf("无效的域名")
	}

	logging.Info("开始执行 Subfinder 任务: %s", payload.Domain)

	// 执行 Subfinder 命令
	cmd := exec.Command("subfinder", "-d", payload.Domain, "-silent")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 subfinder 命令失败: %v", err)
	}

	// 获取执行结果
	result := out.String()
	logging.Info("Subfinder 任务完成，结果: %s", result)

	// 解析子域名并创建 SubdomainEntry 列表
	subdomains := strings.Split(strings.TrimSpace(result), "\n")
	var subdomainEntries []models.SubdomainEntry
	for _, subdomain := range subdomains {
		subdomainEntries = append(subdomainEntries, models.SubdomainEntry{
			ID:     primitive.NewObjectID(),
			Domain: subdomain,
			IsRead: false, // 默认未读
		})
	}

	// 创建 SubdomainData 对象
	subdomainData := &models.SubdomainData{
		Subdomains: subdomainEntries,
	}

	var parentID *primitive.ObjectID
	if payload.ParentID != "" {
		objID, err := primitive.ObjectIDFromHex(payload.ParentID)
		if err == nil {
			parentID = &objID
		}
	}

	// 创建扫描结果记录
	scanResult := &models.Result{
		ID:        primitive.NewObjectID(),
		Type:      "Subdomain",
		Target:    payload.Domain,
		Timestamp: time.Now(),
		Data:      subdomainData,
		ParentID:  parentID,
		IsRead:    false, // 初始任务记录默认未读
	}

	// 存储扫描结果
	if err := s.resultDAO.CreateResult(scanResult); err != nil {
		logging.Error("存储扫描结果失败: %v", err)
		return err
	}

	return nil
}
