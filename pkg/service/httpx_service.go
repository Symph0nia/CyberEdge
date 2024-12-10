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

func (s *HTTPXService) ProbeSubdomains(resultID string, entryIDs []string) (*ResolveResult, error) {
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

	if scanResult.Type != "Subdomain" {
		return nil, errors.New("任务类型不是子域名扫描")
	}

	// 解析数据
	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(scanResult.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return nil, err
	}

	// 创建域名映射
	subdomainMap := make(map[string]*models.SubdomainEntry)
	for i := range subdomainData.Subdomains {
		subdomainMap[subdomainData.Subdomains[i].ID.Hex()] = &subdomainData.Subdomains[i]
	}

	// 定义探测任务结构
	type probeResult struct {
		EntryID    string
		Domain     string
		HTTPResult *HTTPXResult
		Error      error
	}

	// 创建工作池
	maxWorkers := s.workers
	resultChan := make(chan probeResult, len(entryIDs))
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup

	// 启动探测协程
	for _, entryID := range entryIDs {
		subdomain, exists := subdomainMap[entryID]
		if !exists {
			result.Failed[entryID] = "未找到指定的子域名"
			continue
		}

		wg.Add(1)
		go func(entryID string, domain string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 创建上下文
			ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
			defer cancel()

			// 执行探测
			httpResult, err := s.probeWithHTTPX(ctx, domain)
			resultChan <- probeResult{
				EntryID:    entryID,
				Domain:     domain,
				HTTPResult: httpResult,
				Error:      err,
			}
		}(entryID, subdomain.Domain)
	}

	// 启动结果收集协程
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 处理结果
	var mu sync.Mutex
	for res := range resultChan {
		if res.Error != nil {
			mu.Lock()
			result.Failed[res.EntryID] = fmt.Sprintf("探测失败: %v", res.Error)
			mu.Unlock()
			continue
		}

		// 更新数据库
		err := s.resultDAO.UpdateSubdomainHTTPInfo(
			resultID,
			res.EntryID,
			res.HTTPResult.StatusCode,
			res.HTTPResult.Title,
		)

		mu.Lock()
		if err != nil {
			result.Failed[res.EntryID] = fmt.Sprintf("更新HTTP信息失败: %v", err)
		} else {
			result.Success = append(result.Success, res.EntryID)
			logging.Info("成功更新子域名 %s 的HTTP信息: Status=%d, Title=%s",
				res.Domain, res.HTTPResult.StatusCode, res.HTTPResult.Title)
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
