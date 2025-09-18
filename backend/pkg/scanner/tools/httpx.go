package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cyberedge/pkg/scanner"
)

// HttpxScanner Httpx Web技术探测工具
type HttpxScanner struct {
	*BaseScannerTool
}

// NewHttpxScanner 创建Httpx扫描工具
func NewHttpxScanner() scanner.Scanner {
	return &HttpxScanner{
		BaseScannerTool: NewBaseScannerToolWithVersionArgs("httpx", scanner.CategoryWebTech, "httpx", []string{"-version"}),
	}
}

// Scan 执行Web技术探测
func (h *HttpxScanner) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.ScanResult, error) {
	// 验证配置
	if err := h.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 准备目标
	targets := h.prepareTargets(config)
	if len(targets) == 0 {
		return nil, scanner.ErrInvalidTarget
	}

	// 构建命令参数
	args := h.buildArgs(targets, config.Options)

	// 设置超时
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 3 * time.Minute // 默认3分钟超时
	}

	// 执行命令
	output, err := h.ExecuteCommand(ctx, args, timeout)
	if err != nil {
		return nil, fmt.Errorf("httpx执行失败: %w", err)
	}

	// 解析结果
	webTechData, err := h.parseOutput(output)
	if err != nil {
		return nil, fmt.Errorf("解析httpx输出失败: %w", err)
	}

	// 构建扫描结果
	result := &scanner.ScanResult{
		Data: webTechData,
	}

	return result, nil
}

// ValidateConfig 验证扫描配置
func (h *HttpxScanner) ValidateConfig(config scanner.ScanConfig) error {
	if err := h.BaseScannerTool.ValidateConfig(config); err != nil {
		return err
	}

	// 验证目标格式
	targets := h.prepareTargets(config)
	if len(targets) == 0 {
		return fmt.Errorf("无有效的Web目标: %s", config.Target)
	}

	return nil
}

// prepareTargets 准备扫描目标
func (h *HttpxScanner) prepareTargets(config scanner.ScanConfig) []string {
	var targets []string

	// 从父级结果中获取子域名
	if len(config.ParentResults) > 0 {
		for _, parentResult := range config.ParentResults {
			if parentResult.Category == scanner.CategorySubdomain {
				if subdomainData, ok := parentResult.Data.(scanner.SubdomainData); ok {
					for _, subdomain := range subdomainData.Subdomains {
						// 添加HTTP和HTTPS变体
						targets = append(targets, "http://"+subdomain.Subdomain)
						targets = append(targets, "https://"+subdomain.Subdomain)
					}
				}
			}
		}
	}

	// 如果没有父级结果，使用配置的目标
	if len(targets) == 0 {
		target := h.SanitizeTarget(config.Target)
		if target != "" {
			// 添加协议前缀
			if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") {
				targets = append(targets, "http://"+target)
				targets = append(targets, "https://"+target)
			} else {
				targets = append(targets, target)
			}
		}
	}

	return targets
}

// buildArgs 构建命令参数
func (h *HttpxScanner) buildArgs(targets []string, options map[string]string) []string {
	args := []string{
		"-json",           // JSON输出
		"-silent",         // 静默模式
		"-timeout", "10",  // 10秒超时
		"-retries", "2",   // 重试2次
		"-threads", "100", // 100线程
	}

	// 技术检测
	if detectTech, exists := options["detect_tech"]; exists && detectTech == "true" {
		args = append(args, "-tech-detect")
	}

	// 标题提取
	if extractTitle, exists := options["extract_title"]; exists && extractTitle == "true" {
		args = append(args, "-title")
	}

	// 状态码
	if showStatus, exists := options["show_status"]; exists && showStatus == "true" {
		args = append(args, "-status-code")
	}

	// 内容长度
	if showLength, exists := options["show_length"]; exists && showLength == "true" {
		args = append(args, "-content-length")
	}

	// Web服务器识别
	if showServer, exists := options["show_server"]; exists && showServer == "true" {
		args = append(args, "-web-server")
	}

	// 响应时间
	if showTime, exists := options["show_time"]; exists && showTime == "true" {
		args = append(args, "-response-time")
	}

	// 跟随重定向
	if followRedirect, exists := options["follow_redirect"]; exists && followRedirect == "true" {
		args = append(args, "-follow-redirects")
	}

	// 自定义头
	if headers, exists := options["headers"]; exists && headers != "" {
		for _, header := range strings.Split(headers, ",") {
			args = append(args, "-H", strings.TrimSpace(header))
		}
	}

	// 从标准输入读取目标
	if len(targets) > 0 {
		args = append(args, "-l", "-") // 使用-l标志从stdin读取
	}

	return args
}

// parseOutput 解析httpx输出
func (h *HttpxScanner) parseOutput(output []byte) (scanner.WebTechData, error) {
	lines := strings.Split(string(output), "\n")
	var allTechData []scanner.WebTechData

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 尝试JSON解析
		var httpxResult HttpxResult
		if err := json.Unmarshal([]byte(line), &httpxResult); err != nil {
			// 如果不是JSON，尝试简单文本解析
			if techData := h.parseSimpleLine(line); techData.URL != "" {
				allTechData = append(allTechData, techData)
			}
			continue
		}

		// 转换为WebTechData
		techData := h.convertHttpxResult(httpxResult)
		if techData.URL != "" {
			allTechData = append(allTechData, techData)
		}
	}

	// 如果有多个结果，合并它们
	if len(allTechData) > 1 {
		// 返回第一个作为主要结果，或者可以实现更复杂的合并逻辑
		return allTechData[0], nil
	} else if len(allTechData) == 1 {
		return allTechData[0], nil
	}

	return scanner.WebTechData{}, nil
}

// parseSimpleLine 解析简单文本行
func (h *HttpxScanner) parseSimpleLine(line string) scanner.WebTechData {
	// 简单URL匹配
	urlPattern := regexp.MustCompile(`https?://[^\s]+`)
	if matches := urlPattern.FindString(line); matches != "" {
		return scanner.WebTechData{
			URL:        matches,
			StatusCode: h.extractStatusCode(line),
			Title:      h.extractTitle(line),
		}
	}

	return scanner.WebTechData{}
}

// extractStatusCode 从文本中提取状态码
func (h *HttpxScanner) extractStatusCode(line string) int {
	statusPattern := regexp.MustCompile(`\[(\d{3})\]`)
	if matches := statusPattern.FindStringSubmatch(line); len(matches) > 1 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			return code
		}
	}
	return 0
}

// extractTitle 从文本中提取标题
func (h *HttpxScanner) extractTitle(line string) string {
	titlePattern := regexp.MustCompile(`\[([^\]]+)\]`)
	matches := titlePattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		if len(match) > 1 && !regexp.MustCompile(`^\d+$`).MatchString(match[1]) {
			return match[1]
		}
	}
	return ""
}

// convertHttpxResult 转换Httpx结果到WebTechData
func (h *HttpxScanner) convertHttpxResult(result HttpxResult) scanner.WebTechData {
	techData := scanner.WebTechData{
		URL:        result.URL,
		StatusCode: result.StatusCode,
		Title:      result.Title,
	}

	// 转换技术信息
	for _, tech := range result.Technologies {
		techInfo := scanner.TechnologyInfo{
			Name:     tech.Name,
			Category: tech.Category,
			Version:  tech.Version,
		}
		techData.Technologies = append(techData.Technologies, techInfo)
	}

	// 从服务器头中提取技术信息
	if result.WebServer != "" {
		techInfo := scanner.TechnologyInfo{
			Name:     result.WebServer,
			Category: "Web Server",
		}
		techData.Technologies = append(techData.Technologies, techInfo)
	}

	return techData
}

// HttpxResult Httpx JSON输出结构
type HttpxResult struct {
	URL           string               `json:"url"`
	StatusCode    int                  `json:"status_code"`
	Title         string               `json:"title"`
	ContentLength int                  `json:"content_length"`
	WebServer     string               `json:"webserver"`
	Technologies  []HttpxTechnology    `json:"technologies"`
	ResponseTime  string               `json:"response_time"`
	Headers       map[string]string    `json:"headers"`
}

type HttpxTechnology struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Version  string `json:"version"`
}