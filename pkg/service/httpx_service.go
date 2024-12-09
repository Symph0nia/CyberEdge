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

// 探测单个域名
func (s *HTTPXService) ProbeAndUpdateSubdomain(resultID, entryID string) error {
	// 获取扫描结果
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	result, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return err
	}

	// 检查任务类型
	if result.Type != "Subdomain" {
		return errors.New("任务类型不是子域名扫描")
	}

	// 解析 SubdomainData
	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(result.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return err
	}

	// 查找指定的子域名
	var targetSubdomain *models.SubdomainEntry
	for i := range subdomainData.Subdomains {
		if subdomainData.Subdomains[i].ID.Hex() == entryID {
			targetSubdomain = &subdomainData.Subdomains[i]
			break
		}
	}

	if targetSubdomain == nil {
		return errors.New("未找到指定的子域名")
	}

	probeResult, err := s.probeWithHTTPX(ctx, targetSubdomain.Domain)
	if err != nil {
		logging.Error("HTTPX探测失败: %v", err)
		return err
	}

	// 更新数据库
	err = s.resultDAO.UpdateSubdomainHTTPInfo(resultID, entryID, probeResult.StatusCode, probeResult.Title)
	if err != nil {
		logging.Error("更新子域名HTTP信息失败: %v", err)
		return err
	}

	logging.Info("成功更新子域名 %s 的HTTP信息: Status=%d, Title=%s",
		targetSubdomain.Domain, probeResult.StatusCode, probeResult.Title)

	return nil
}

// 批量探测域名
func (s *HTTPXService) BatchProbeAndUpdateSubdomains(resultID string, entryIDs []string) (*ResolveResult, error) {
	logging.Info("开始批量HTTP探测: %v", entryIDs)

	result := &ResolveResult{
		Success: make([]string, 0),
		Failed:  make(map[string]string),
	}

	// 获取扫描结果
	scanResult, err := s.resultDAO.GetResultByID(resultID)
	if err != nil {
		logging.Error("获取扫描结果失败: %v", err)
		return nil, err
	}

	// 检查任务类型
	if scanResult.Type != "Subdomain" {
		return nil, errors.New("任务类型不是子域名扫描")
	}

	// 解析 SubdomainData
	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(scanResult.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return nil, err
	}

	// 创建entryID到subdomain的映射
	subdomainMap := make(map[string]*models.SubdomainEntry)
	for i := range subdomainData.Subdomains {
		subdomainMap[subdomainData.Subdomains[i].ID.Hex()] = &subdomainData.Subdomains[i]
	}

	type probeTask struct {
		entryID  string
		domain   string
		resultCh chan *HTTPXResult
		errCh    chan error
	}

	taskCh := make(chan probeTask, len(entryIDs))
	var wg sync.WaitGroup
	var resultMutex sync.Mutex

	// 启动工作线程
	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
				probeResult, err := s.probeWithHTTPX(ctx, task.domain)
				cancel()

				if err != nil {
					task.errCh <- err
					continue
				}
				task.resultCh <- probeResult
			}
		}()
	}

	// 发送任务
	for _, entryID := range entryIDs {
		subdomain, exists := subdomainMap[entryID]
		if !exists {
			result.Failed[entryID] = "未找到指定的子域名"
			continue
		}

		resultCh := make(chan *HTTPXResult, 1)
		errCh := make(chan error, 1)

		taskCh <- probeTask{
			entryID:  entryID,
			domain:   subdomain.Domain,
			resultCh: resultCh,
			errCh:    errCh,
		}

		// 处理结果
		go func(entryID string, domain string, resultCh chan *HTTPXResult, errCh chan error) {
			select {
			case probeResult := <-resultCh:
				err := s.resultDAO.UpdateSubdomainHTTPInfo(resultID, entryID, probeResult.StatusCode, probeResult.Title)
				resultMutex.Lock()
				if err != nil {
					result.Failed[entryID] = fmt.Sprintf("更新HTTP信息失败: %v", err)
				} else {
					result.Success = append(result.Success, entryID)
					logging.Info("成功更新子域名 %s 的HTTP信息: Status=%d, Title=%s",
						domain, probeResult.StatusCode, probeResult.Title)
				}
				resultMutex.Unlock()

			case err := <-errCh:
				resultMutex.Lock()
				result.Failed[entryID] = fmt.Sprintf("HTTP探测失败: %v", err)
				resultMutex.Unlock()
			}
			close(resultCh)
			close(errCh)
		}(entryID, subdomain.Domain, resultCh, errCh)
	}

	close(taskCh)
	wg.Wait()

	logging.Info("批量HTTP探测完成，成功: %d, 失败: %d",
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
