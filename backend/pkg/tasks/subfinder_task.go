package tasks

import (
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SubfinderTask struct {
	TaskTemplate
	resultDAO *dao.ResultDAO
	targetDAO *dao.TargetDAO
	configDAO *dao.ConfigDAO // 新增配置DAO
}

func NewSubfinderTask(taskDAO *dao.TaskDAO, targetDAO *dao.TargetDAO, resultDAO *dao.ResultDAO, configDAO *dao.ConfigDAO) *SubfinderTask {
	return &SubfinderTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
		resultDAO:    resultDAO,
		targetDAO:    targetDAO,
		configDAO:    configDAO, // 初始化配置DAO
	}
}

func (s *SubfinderTask) Handle(ctx context.Context, t *asynq.Task) error {
	return s.Execute(ctx, t, s.runSubfinder)
}

func (s *SubfinderTask) runSubfinder(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Domain   string `json:"target"`
		TaskID   string `json:"task_id"`
		TargetID string `json:"target_id,omitempty"`
	}

	// 解析任务载荷
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Domain == "" {
		return fmt.Errorf("无效的域名")
	}

	logging.Info("开始执行 Subfinder 任务: %s", payload.Domain)

	// 从数据库获取默认工具配置
	toolConfig, err := s.configDAO.GetDefaultToolConfig()
	if err != nil {
		logging.Error("获取默认工具配置失败: %v", err)
		return fmt.Errorf("获取默认工具配置失败: %v", err)
	}

	// 检查 Subfinder 是否启用
	if !toolConfig.SubfinderConfig.Enabled {
		logging.Warn("Subfinder 工具未启用，跳过任务执行")
		return nil
	}

	// 创建临时文件
	tempFile, err := ioutil.TempFile("", "subfinder-result-*.txt")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name()) // 确保在函数结束时删除临时文件

	// 准备 Subfinder 命令参数
	subfinderArgs := []string{
		"-d", payload.Domain,
		"-silent",
		"-o", tempFile.Name(),
	}

	// 如果配置了配置文件路径，添加参数
	if toolConfig.SubfinderConfig.ConfigPath != "" {
		subfinderArgs = append(subfinderArgs, "-config", toolConfig.SubfinderConfig.ConfigPath)
	}

	// 如果配置了线程数，添加参数
	if toolConfig.SubfinderConfig.Threads > 0 {
		subfinderArgs = append(subfinderArgs, "-t", strconv.Itoa(toolConfig.SubfinderConfig.Threads))
	}

	// 如果配置了超时时间，添加参数
	if toolConfig.SubfinderConfig.Timeout > 0 {
		subfinderArgs = append(subfinderArgs, "-timeout", strconv.Itoa(toolConfig.SubfinderConfig.Timeout))
	}

	logging.Info("执行 Subfinder 命令，参数: %v", subfinderArgs)

	// 执行 Subfinder 命令，将结果输出到临时文件
	cmd := exec.Command("subfinder", subfinderArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 subfinder 命令失败: %v", err)
	}

	// 读取临时文件内容
	result, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("读取 Subfinder 结果文件失败: %v", err)
	}

	logging.Info("Subfinder 任务完成，结果已保存到文件: %s", tempFile.Name())

	// 解析子域名列表
	subdomains := strings.Split(strings.TrimSpace(string(result)), "\n")

	// 处理 TargetID
	var targetID *primitive.ObjectID
	if payload.TargetID != "" {
		objID, err := primitive.ObjectIDFromHex(payload.TargetID)
		if err == nil {
			targetID = &objID
		}
	}

	// 创建子域名条目
	var subdomainEntries []models.SubdomainEntry
	for _, subdomain := range subdomains {
		if subdomain != "" {
			entry := models.SubdomainEntry{
				ID:     primitive.NewObjectID(),
				Domain: subdomain,
				IsRead: false,
			}
			// 只有在 targetID 存在时才设置 TargetID
			if targetID != nil {
				entry.TargetID = targetID
			}
			subdomainEntries = append(subdomainEntries, entry)
		}
	}

	// 创建 SubdomainData 对象
	subdomainData := &models.SubdomainData{
		Subdomains: subdomainEntries,
	}

	// 创建扫描结果记录
	scanResult := &models.Result{
		ID:        primitive.NewObjectID(),
		Type:      "Subdomain",
		Target:    payload.Domain,
		Timestamp: time.Now(),
		Data:      subdomainData,
		TargetID:  targetID,
		IsRead:    false,
	}

	// 存储扫描结果
	if err := s.resultDAO.CreateResult(scanResult); err != nil {
		logging.Error("存储扫描结果失败: %v", err)
		return err
	}

	logging.Info("成功处理并存储 Subfinder 结果，共找到 %d 个子域名", len(subdomainEntries))

	return nil
}
