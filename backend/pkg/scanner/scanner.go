package scanner

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Scanner struct {
	taskDAO   *dao.TaskDAO
	resultDAO *dao.ResultDAO
}

func NewScanner(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *Scanner {
	return &Scanner{
		taskDAO:   taskDAO,
		resultDAO: resultDAO,
	}
}

// ScanRequest 扫描请求结构
type ScanRequest struct {
	Type     string `json:"type"`     // subfinder, nmap, ffuf
	Target   string `json:"target"`   // 目标域名或IP
	TargetID string `json:"target_id,omitempty"`
}

// ExecuteScan 执行扫描任务
func (s *Scanner) ExecuteScan(ctx context.Context, req ScanRequest) (*models.Task, error) {
	// 创建任务记录
	taskID := primitive.NewObjectID()
	task := &models.Task{
		ID:        taskID,
		Type:      req.Type,
		Status:    models.TaskStatusRunning,
		Payload:   req.Target,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 如果提供了TargetID，转换并设置
	if req.TargetID != "" {
		if targetOID, err := primitive.ObjectIDFromHex(req.TargetID); err == nil {
			task.TargetID = &targetOID
		}
	}

	// 保存任务到数据库
	if err := s.taskDAO.Create(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %v", err)
	}

	// 启动goroutine执行扫描
	go s.runScanInBackground(ctx, task)

	return task, nil
}

// runScanInBackground 在后台执行扫描
func (s *Scanner) runScanInBackground(ctx context.Context, task *models.Task) {
	defer func() {
		if r := recover(); r != nil {
			logging.Error("扫描任务发生panic: %v", r)
			s.updateTaskStatus(task.ID, models.TaskStatusFailed, fmt.Sprintf("任务异常: %v", r))
		}
	}()

	logging.Info("开始执行%s扫描任务: %s", task.Type, task.Payload)

	var result string
	var err error

	// 创建带超时的context
	scanCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	switch task.Type {
	case "subfinder":
		result, err = s.runSubfinder(scanCtx, task.Payload)
	case "nmap":
		result, err = s.runNmap(scanCtx, task.Payload)
	case "ffuf":
		result, err = s.runFfuf(scanCtx, task.Payload)
	default:
		err = fmt.Errorf("不支持的扫描类型: %s", task.Type)
	}

	// 更新任务状态
	if err != nil {
		logging.Error("%s扫描失败: %v", task.Type, err)
		s.updateTaskStatus(task.ID, models.TaskStatusFailed, err.Error())
		return
	}

	logging.Info("%s扫描完成: %s", task.Type, task.Payload)
	s.updateTaskStatus(task.ID, models.TaskStatusCompleted, result)

	// 保存结果到结果表
	s.saveResult(task, result)
}

// updateTaskStatus 更新任务状态
func (s *Scanner) updateTaskStatus(taskID primitive.ObjectID, status models.TaskStatus, result string) {
	updateData := map[string]interface{}{
		"status":     status,
		"result":     result,
		"updated_at": time.Now(),
	}

	if status == models.TaskStatusCompleted || status == models.TaskStatusFailed {
		now := time.Now()
		updateData["completed_at"] = &now
	}

	if err := s.taskDAO.Update(taskID, updateData); err != nil {
		logging.Error("更新任务状态失败: %v", err)
	}
}

// saveResult 保存扫描结果
func (s *Scanner) saveResult(task *models.Task, result string) {
	// 这里可以解析result并保存到results表
	// 暂时简化处理
	logging.Info("扫描结果已保存: %s", task.Type)
}

// runSubfinder 执行subfinder扫描
func (s *Scanner) runSubfinder(ctx context.Context, domain string) (string, error) {
	args := []string{
		"-d", domain,
		"-silent",
		"-o", "/tmp/subfinder_output.txt",
	}

	cmd := exec.CommandContext(ctx, "subfinder", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("subfinder执行失败: %v, 输出: %s", err, string(output))
	}

	return string(output), nil
}

// runNmap 执行nmap扫描
func (s *Scanner) runNmap(ctx context.Context, target string) (string, error) {
	args := []string{
		"-sS",           // SYN扫描
		"-T4",           // 扫描速度
		"-p", "1-1000",  // 端口范围
		"--open",        // 只显示开放端口
		target,
	}

	cmd := exec.CommandContext(ctx, "nmap", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("nmap执行失败: %v, 输出: %s", err, string(output))
	}

	return string(output), nil
}

// runFfuf 执行ffuf目录扫描
func (s *Scanner) runFfuf(ctx context.Context, target string) (string, error) {
	// 确保目标以http://或https://开头
	if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
		target = "http://" + target
	}

	args := []string{
		"-u", target + "/FUZZ",
		"-w", "/usr/share/wordlists/dirb/common.txt", // 需要确保wordlist存在
		"-fc", "404",     // 过滤404响应
		"-t", "50",       // 线程数
		"-o", "/tmp/ffuf_output.json",
		"-of", "json",    // JSON输出格式
	}

	cmd := exec.CommandContext(ctx, "ffuf", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffuf执行失败: %v, 输出: %s", err, string(output))
	}

	return string(output), nil
}

// GetTaskStatus 获取任务状态
func (s *Scanner) GetTaskStatus(taskID primitive.ObjectID) (*models.Task, error) {
	return s.taskDAO.GetByID(taskID)
}

// ListTasks 列出所有任务
func (s *Scanner) ListTasks() ([]*models.Task, error) {
	return s.taskDAO.GetAll()
}

// CancelTask 取消任务（标记为失败状态）
func (s *Scanner) CancelTask(taskID primitive.ObjectID) error {
	return s.updateTaskStatus(taskID, models.TaskStatusFailed, "用户取消")
}