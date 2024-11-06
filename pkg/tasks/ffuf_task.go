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
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FfufTask struct {
	TaskTemplate
	resultDAO *dao.ResultDAO
}

func NewFfufTask(taskDAO *dao.TaskDAO, resultDAO *dao.ResultDAO) *FfufTask {
	return &FfufTask{
		TaskTemplate: TaskTemplate{TaskDAO: taskDAO},
		resultDAO:    resultDAO,
	}
}

func (f *FfufTask) Handle(ctx context.Context, t *asynq.Task) error {
	return f.Execute(ctx, t, f.runFfuf)
}

func (f *FfufTask) runFfuf(ctx context.Context, t *asynq.Task) error {
	var payload struct {
		Target   string `json:"target"`
		TaskID   string `json:"task_id"`
		ParentID string `json:"parent_id,omitempty"`
	}

	// 解析任务载荷
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %v", err)
	}

	if payload.Target == "" {
		return fmt.Errorf("无效的目标 URL")
	}

	logging.Info("开始执行 ffuf 任务: %s", payload.Target)

	// 创建临时文件
	tempFile, err := ioutil.TempFile("", "ffuf-result-*.json")
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tempFile.Name()) // 确保在函数结束时删除临时文件

	// 执行 ffuf 命令，将结果输出到临时文件
	cmd := exec.Command("ffuf", "-u", payload.Target+"/FUZZ", "-w", "/path/to/wordlist.txt", "-o", tempFile.Name(), "-of", "json")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("执行 ffuf 命令失败: %v", err)
	}

	// 读取临时文件内容
	result, err := ioutil.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("读取 ffuf 结果文件失败: %v", err)
	}

	logging.Info("ffuf 任务完成，结果已保存到文件: %s", tempFile.Name())

	// 解析 ffuf 结果
	var ffufResult struct {
		Results []struct {
			Status int    `json:"status"`
			Length int    `json:"length"`
			Words  int    `json:"words"`
			Lines  int    `json:"lines"`
			URL    string `json:"url"`
		} `json:"results"`
	}
	if err := json.Unmarshal(result, &ffufResult); err != nil {
		return fmt.Errorf("解析 ffuf 结果失败: %v", err)
	}

	// 创建 Path 列表
	pathData := &models.PathData{
		Paths: []*models.Path{},
	}
	for _, r := range ffufResult.Results {
		pathData.Paths = append(pathData.Paths, &models.Path{
			ID:     primitive.NewObjectID(),
			Path:   strings.TrimPrefix(r.URL, payload.Target),
			Status: r.Status,
			Length: r.Length,
			Words:  r.Words,
			Lines:  r.Lines,
			IsRead: false, // 默认未读
		})
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
		Type:      "Path",
		Target:    payload.Target,
		Timestamp: time.Now(),
		Data:      pathData,
		ParentID:  parentID,
		IsRead:    false, // 初始任务记录默认未读
	}

	// 存储扫描结果
	if err := f.resultDAO.CreateResult(scanResult); err != nil {
		logging.Error("存储扫描结果失败: %v", err)
		return err
	}

	logging.Info("成功处理并存储 ffuf 结果，共找到 %d 个路径", len(pathData.Paths))

	return nil
}
