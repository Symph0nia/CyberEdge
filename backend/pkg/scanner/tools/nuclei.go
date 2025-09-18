package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cyberedge/pkg/scanner"
)

// NucleiScanner Nuclei漏洞扫描工具
type NucleiScanner struct {
	*BaseScannerTool
}

// NewNucleiScanner 创建Nuclei扫描工具
func NewNucleiScanner() scanner.Scanner {
	return &NucleiScanner{
		BaseScannerTool: NewBaseScannerToolWithVersionArgs("nuclei", scanner.CategoryVulnerability, "nuclei", []string{"-version"}),
	}
}

// Scan 执行漏洞扫描
func (n *NucleiScanner) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.ScanResult, error) {
	// 验证配置
	if err := n.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 准备目标
	targets := n.prepareTargets(config)
	if len(targets) == 0 {
		return nil, scanner.ErrInvalidTarget
	}

	// 构建命令参数
	args := n.buildArgs(targets, config.Options)

	// 设置超时
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 15 * time.Minute // 默认15分钟超时
	}

	// 执行命令
	output, err := n.ExecuteCommand(ctx, args, timeout)
	if err != nil {
		return nil, fmt.Errorf("nuclei执行失败: %w", err)
	}

	// 解析结果
	vulnerabilities, err := n.parseOutput(output)
	if err != nil {
		return nil, fmt.Errorf("解析nuclei输出失败: %w", err)
	}

	// 构建扫描结果
	result := &scanner.ScanResult{
		Data: scanner.VulnerabilityData{
			Vulnerabilities: vulnerabilities,
		},
	}

	return result, nil
}

// ValidateConfig 验证扫描配置
func (n *NucleiScanner) ValidateConfig(config scanner.ScanConfig) error {
	if err := n.BaseScannerTool.ValidateConfig(config); err != nil {
		return err
	}

	// 验证目标格式
	targets := n.prepareTargets(config)
	if len(targets) == 0 {
		return fmt.Errorf("无有效的扫描目标: %s", config.Target)
	}

	return nil
}

// prepareTargets 准备扫描目标
func (n *NucleiScanner) prepareTargets(config scanner.ScanConfig) []string {
	var targets []string

	// 从父级结果中获取Web技术信息
	if len(config.ParentResults) > 0 {
		for _, parentResult := range config.ParentResults {
			if parentResult.Category == scanner.CategoryWebTech {
				if webTechData, ok := parentResult.Data.(scanner.WebTechData); ok {
					if webTechData.URL != "" {
						targets = append(targets, webTechData.URL)
					}
				}
			} else if parentResult.Category == scanner.CategorySubdomain {
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
		target := n.SanitizeTarget(config.Target)
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
func (n *NucleiScanner) buildArgs(targets []string, options map[string]string) []string {
	args := []string{
		"-json",            // JSON输出
		"-silent",          // 静默模式
		"-timeout", "10",   // 10秒超时
		"-retries", "2",    // 重试2次
		"-c", "50",         // 50并发
		"-rate-limit", "150", // 速率限制
	}

	// 模板选择
	if templates, exists := options["templates"]; exists && templates != "" {
		for _, template := range strings.Split(templates, ",") {
			args = append(args, "-t", strings.TrimSpace(template))
		}
	} else {
		// 默认使用高严重性模板
		args = append(args, "-s", "high,critical")
	}

	// 严重性过滤
	if severity, exists := options["severity"]; exists && severity != "" {
		args = append(args, "-s", severity)
	}

	// 标签过滤
	if tags, exists := options["tags"]; exists && tags != "" {
		args = append(args, "-tags", tags)
	}

	// 排除标签
	if excludeTags, exists := options["exclude_tags"]; exists && excludeTags != "" {
		args = append(args, "-etags", excludeTags)
	}

	// 自定义头
	if headers, exists := options["headers"]; exists && headers != "" {
		for _, header := range strings.Split(headers, ",") {
			args = append(args, "-H", strings.TrimSpace(header))
		}
	}

	// 跟随重定向
	if followRedirect, exists := options["follow_redirect"]; exists && followRedirect == "true" {
		args = append(args, "-follow-redirects")
	}

	// 禁用更新检查
	args = append(args, "-duc")

	// 添加目标
	for _, target := range targets {
		args = append(args, "-u", target)
	}

	return args
}

// parseOutput 解析nuclei输出
func (n *NucleiScanner) parseOutput(output []byte) ([]scanner.VulnerabilityInfo, error) {
	lines := strings.Split(string(output), "\n")
	var vulnerabilities []scanner.VulnerabilityInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 尝试JSON解析
		var nucleiResult NucleiResult
		if err := json.Unmarshal([]byte(line), &nucleiResult); err != nil {
			// 如果不是JSON，尝试简单文本解析
			if vuln := n.parseSimpleLine(line); vuln.Target != "" {
				vulnerabilities = append(vulnerabilities, vuln)
			}
			continue
		}

		// 转换为VulnerabilityInfo
		vuln := n.convertNucleiResult(nucleiResult)
		if vuln.Target != "" {
			vulnerabilities = append(vulnerabilities, vuln)
		}
	}

	return vulnerabilities, nil
}

// parseSimpleLine 解析简单文本行
func (n *NucleiScanner) parseSimpleLine(line string) scanner.VulnerabilityInfo {
	// 简单格式：[template-id] [severity] target description
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		return scanner.VulnerabilityInfo{
			Target:      parts[len(parts)-1], // 最后一个字段通常是目标
			Title:       strings.Join(parts[2:len(parts)-1], " "),
			Description: line,
			Severity:    n.extractSeverity(line),
		}
	}

	return scanner.VulnerabilityInfo{}
}

// extractSeverity 从文本中提取严重性
func (n *NucleiScanner) extractSeverity(line string) string {
	severities := []string{"critical", "high", "medium", "low", "info"}
	lineLower := strings.ToLower(line)

	for _, severity := range severities {
		if strings.Contains(lineLower, severity) {
			return severity
		}
	}

	return "unknown"
}

// convertNucleiResult 转换Nuclei结果到VulnerabilityInfo
func (n *NucleiScanner) convertNucleiResult(result NucleiResult) scanner.VulnerabilityInfo {
	vuln := scanner.VulnerabilityInfo{
		Target:      result.Host,
		Title:       result.Info.Name,
		Description: result.Info.Description,
		Severity:    result.Info.Severity,
		Location:    result.MatchedAt,
	}

	// 设置CVSS分数
	if result.Info.Classification.CVSS != nil {
		if score := result.Info.Classification.CVSS.Score; score > 0 {
			vuln.CVSS = score
		}
	}

	// 设置CVE ID
	if len(result.Info.Classification.CVE) > 0 {
		vuln.CVEID = result.Info.Classification.CVE[0]
	}

	// 提取参数和载荷信息
	if len(result.ExtractedResults) > 0 {
		vuln.Parameter = strings.Join(result.ExtractedResults, ", ")
	}

	if result.Request != "" {
		vuln.Payload = result.Request
	}

	return vuln
}

// NucleiResult Nuclei JSON输出结构
type NucleiResult struct {
	TemplateID       string               `json:"template-id"`
	TemplateURL      string               `json:"template-url"`
	TemplatePath     string               `json:"template-path"`
	Info             NucleiInfo           `json:"info"`
	Type             string               `json:"type"`
	Host             string               `json:"host"`
	MatchedAt        string               `json:"matched-at"`
	ExtractedResults []string             `json:"extracted-results"`
	Request          string               `json:"request"`
	Response         string               `json:"response"`
	IP               string               `json:"ip"`
	Timestamp        string               `json:"timestamp"`
}

type NucleiInfo struct {
	Name           string                  `json:"name"`
	Author         []string                `json:"author"`
	Tags           []string                `json:"tags"`
	Description    string                  `json:"description"`
	Reference      []string                `json:"reference"`
	Severity       string                  `json:"severity"`
	Classification NucleiClassification    `json:"classification"`
	Metadata       map[string]interface{}  `json:"metadata"`
}

type NucleiClassification struct {
	CVE            []string    `json:"cve-id"`
	CWE            []string    `json:"cwe-id"`
	CVSS           *NucleiCVSS `json:"cvss-metrics"`
	CVSSV3Vector   string      `json:"cvss-vector"`
}

type NucleiCVSS struct {
	Score  float64 `json:"cvss-score"`
	Vector string  `json:"cvss-vector"`
}