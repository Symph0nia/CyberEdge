package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"cyberedge/pkg/scanner"
)

// SubfinderScanner Subfinder子域名扫描工具
type SubfinderScanner struct {
	*BaseScannerTool
}

// NewSubfinderScanner 创建Subfinder扫描工具
func NewSubfinderScanner() scanner.Scanner {
	return &SubfinderScanner{
		BaseScannerTool: NewBaseScannerToolWithVersionArgs("subfinder", scanner.CategorySubdomain, "subfinder", []string{"-version"}),
	}
}

// Scan 执行子域名扫描
func (s *SubfinderScanner) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.ScanResult, error) {
	// 验证配置
	if err := s.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 清理目标
	target := s.SanitizeTarget(config.Target)
	if target == "" {
		return nil, scanner.ErrInvalidTarget
	}

	// 构建命令参数
	args := s.buildArgs(target, config.Options)

	// 设置超时
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Minute // 默认10分钟超时
	}

	// 执行命令
	output, err := s.ExecuteCommand(ctx, args, timeout)
	if err != nil {
		return nil, fmt.Errorf("subfinder执行失败: %w", err)
	}

	// 解析结果
	subdomains, err := s.parseOutput(output, target)
	if err != nil {
		return nil, fmt.Errorf("解析subfinder输出失败: %w", err)
	}

	// 构建扫描结果
	result := &scanner.ScanResult{
		Data: scanner.SubdomainData{
			Subdomains: subdomains,
		},
	}

	return result, nil
}

// ValidateConfig 验证扫描配置
func (s *SubfinderScanner) ValidateConfig(config scanner.ScanConfig) error {
	if err := s.BaseScannerTool.ValidateConfig(config); err != nil {
		return err
	}

	// 验证目标是否为有效域名
	if !s.isValidDomain(config.Target) {
		return fmt.Errorf("目标不是有效的域名: %s", config.Target)
	}

	return nil
}

// buildArgs 构建命令参数
func (s *SubfinderScanner) buildArgs(target string, options map[string]string) []string {
	args := []string{
		"-d", target,        // 指定域名
		"-silent",           // 静默模式，只输出结果
		"-o", "/dev/stdout", // 输出到标准输出
	}

	// 添加可选参数
	if sources, exists := options["sources"]; exists && sources != "" {
		args = append(args, "-sources", sources)
	}

	if recursive, exists := options["recursive"]; exists && recursive == "true" {
		args = append(args, "-recursive")
	}

	if timeout, exists := options["timeout"]; exists && timeout != "" {
		args = append(args, "-timeout", timeout)
	}

	// 添加配置文件路径（如果存在）
	if configPath, exists := options["config"]; exists && configPath != "" {
		args = append(args, "-config", configPath)
	}

	return args
}

// parseOutput 解析subfinder输出
func (s *SubfinderScanner) parseOutput(output []byte, domain string) ([]scanner.SubdomainInfo, error) {
	lines := s.ParseLines(output)
	var subdomains []scanner.SubdomainInfo

	for _, line := range lines {
		subdomain := strings.TrimSpace(line)
		if subdomain == "" || !strings.Contains(subdomain, domain) {
			continue
		}

		// 解析IP地址
		ips, err := s.resolveIPs(subdomain)
		if err != nil {
			// DNS解析失败不应该阻止整个扫描，只记录空IP
			ips = []string{}
		}

		subdomainInfo := scanner.SubdomainInfo{
			Domain:    domain,
			Subdomain: subdomain,
			IPs:       ips,
			Source:    "subfinder",
		}

		subdomains = append(subdomains, subdomainInfo)
	}

	return subdomains, nil
}

// resolveIPs 解析域名对应的IP地址
func (s *SubfinderScanner) resolveIPs(domain string) ([]string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, ip := range ips {
		result = append(result, ip.String())
	}

	return result, nil
}

// isValidDomain 验证是否为有效域名
func (s *SubfinderScanner) isValidDomain(domain string) bool {
	// 基础域名格式验证
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	// 不能以点开头或结尾
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	// 必须包含至少一个点
	if !strings.Contains(domain, ".") {
		return false
	}

	// 检查是否包含非法字符
	for _, char := range domain {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '.' || char == '-') {
			return false
		}
	}

	return true
}

// SubfinderResult Subfinder原始输出结果（用于JSON解析模式）
type SubfinderResult struct {
	Host   string `json:"host"`
	Source string `json:"source"`
}

// parseJSONOutput 解析JSON格式输出（备用方案）
func (s *SubfinderScanner) parseJSONOutput(output []byte, domain string) ([]scanner.SubdomainInfo, error) {
	lines := strings.Split(string(output), "\n")
	var subdomains []scanner.SubdomainInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var result SubfinderResult
		if err := json.Unmarshal([]byte(line), &result); err != nil {
			// 如果不是JSON格式，跳过
			continue
		}

		// 解析IP地址
		ips, err := s.resolveIPs(result.Host)
		if err != nil {
			ips = []string{}
		}

		subdomainInfo := scanner.SubdomainInfo{
			Domain:    domain,
			Subdomain: result.Host,
			IPs:       ips,
			Source:    fmt.Sprintf("subfinder/%s", result.Source),
		}

		subdomains = append(subdomains, subdomainInfo)
	}

	return subdomains, nil
}