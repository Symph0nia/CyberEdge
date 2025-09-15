// httpx_service.go

package service

import (
	"bytes"
	"context"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/utils"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// 补充获取目标类型名称的方法
func (s *HTTPXService) getTargetTypeName(taskType string) string {
	switch taskType {
	case "Subdomain":
		return "子域名"
	case "Port":
		return "端口"
	case "Path":
		return "路径"
	default:
		return "目标"
	}
}

type HTTPXService struct {
	resultDAO *dao.ResultDAO
	workers   int           // 工作线程数
	timeout   time.Duration // 超时时间
}

// HTTPX结果结构体
type HTTPXResult struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"` // 修改为正确的 JSON 字段名
	Title      string `json:"title"`
}

func NewHTTPXService(resultDAO *dao.ResultDAO) *HTTPXService {
	return &HTTPXService{
		resultDAO: resultDAO,
		workers:   10,              // 默认10个工作线程
		timeout:   3 * time.Second, // 默认3秒超时
	}
}

// 定义通用接口
type ProbeTarget interface {
	GetID() string       // 获取目标的 ID
	GetProbeURL() string // 获取探测 URL
}

// HTTPXService 的方法
func (s *HTTPXService) getTargetMap(result *models.Result) (map[string]ProbeTarget, error) {
	switch result.Type {
	case "Subdomain":
		var data models.SubdomainData
		if err := utils.UnmarshalData(result.Data, &data); err != nil {
			return nil, err
		}
		targetMap := make(map[string]ProbeTarget, len(data.Subdomains))
		for i := range data.Subdomains {
			targetMap[data.Subdomains[i].ID.Hex()] = data.Subdomains[i]
		}
		return targetMap, nil

	case "Port":
		var data models.PortData
		if err := utils.UnmarshalData(result.Data, &data); err != nil {
			return nil, err
		}
		targetMap := make(map[string]ProbeTarget, len(data.Ports))
		for i := range data.Ports {
			targetMap[data.Ports[i].ID.Hex()] = data.Ports[i]
		}
		return targetMap, nil

	case "Path":
		var data models.PathData
		if err := utils.UnmarshalData(result.Data, &data); err != nil {
			return nil, err
		}
		targetMap := make(map[string]ProbeTarget, len(data.Paths))
		for i := range data.Paths {
			targetMap[data.Paths[i].ID.Hex()] = data.Paths[i]
		}
		return targetMap, nil

	default:
		return nil, fmt.Errorf("不支持的任务类型: %s", result.Type)
	}
}

func (s *HTTPXService) ProbeTargets(resultID string, entryIDs []string) (*ResolveResult, error) {
	logging.Info("开始HTTP探测: %v", entryIDs)

	result := &ResolveResult{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
	}

	// 获取和验证扫描结果
	scanResult, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return nil, err
	}

	// 获取目标映射
	targetMap, err := s.getTargetMap(scanResult)
	if err != nil {
		return nil, err
	}

	// 创建工作池
	maxWorkers := s.workers
	resultChan := make(chan struct {
		entryID string
		target  ProbeTarget
		result  *HTTPXResult
		err     error
	}, len(entryIDs))
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	// 启动探测协程
	for _, entryID := range entryIDs {
		target, exists := targetMap[entryID]
		if !exists {
			result.Failed[entryID] = fmt.Sprintf("未找到指定的%s", s.getTargetTypeName(scanResult.Type))
			continue
		}

		wg.Add(1)
		go func(entryID string, target ProbeTarget) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
			defer cancel()

			httpResult, err := s.probeWithHTTPX(ctx, target.GetProbeURL())
			resultChan <- struct {
				entryID string
				target  ProbeTarget
				result  *HTTPXResult
				err     error
			}{entryID, target, httpResult, err}
		}(entryID, target)
	}

	// 启动结果收集协程
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 处理结果
	var mu sync.Mutex
	for res := range resultChan {
		if res.err != nil {
			mu.Lock()
			result.Failed[res.entryID] = fmt.Sprintf("探测失败: %v", res.err)
			mu.Unlock()
			continue
		}

		// 更新数据库
		err := s.resultDAO.UpdateHTTPInfo(
			resultID,
			res.entryID,
			scanResult.Type,
			res.result.StatusCode,
			res.result.Title,
		)

		mu.Lock()
		if err != nil {
			result.Failed[res.entryID] = fmt.Sprintf("更新HTTP信息失败: %v", err)
		} else {
			result.Success = append(result.Success, res.entryID)
			logging.Info("成功更新%s的HTTP信息: Target=%s, Status=%d, Title=%s",
				s.getTargetTypeName(scanResult.Type),
				res.target.GetProbeURL(),
				res.result.StatusCode,
				res.result.Title)
		}
		mu.Unlock()
	}

	logging.Info("探测完成，成功: %d, 失败: %d",
		len(result.Success), len(result.Failed))
	return result, nil
}

// 调用 httpx 工具进行探测
func (s *HTTPXService) probeWithHTTPX(ctx context.Context, domain string) (*HTTPXResult, error) {
	// 准备命令
	cmd := exec.CommandContext(ctx, "httpx",
		"-u", domain,
		"-silent",
		"-json",
		"-title",
		"-status-code",
		"-no-color",
		"-timeout", "3") // 设置 httpx 自身的超时时间为3秒

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// 执行命令
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("探测超时: %s", domain)
		}
		return nil, fmt.Errorf("执行httpx失败: %v, stderr: %s", err, stderr.String())
	}

	// 解析结果
	output := strings.TrimSpace(stdout.String())
	if output == "" {
		return nil, errors.New("无效的httpx输出")
	}

	var result HTTPXResult
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("解析httpx输出失败: %v", err)
	}

	// 验证状态码
	if result.StatusCode == 0 && len(result.URL) > 0 {
		return nil, fmt.Errorf("无效的状态码: %s", domain)
	}

	return &result, nil
}
