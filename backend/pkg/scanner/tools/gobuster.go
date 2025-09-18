package tools

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cyberedge/pkg/scanner"
)

// GobusterScanner Gobuster目录/路径扫描工具
type GobusterScanner struct {
	*BaseScannerTool
}

// NewGobusterScanner 创建Gobuster扫描工具
func NewGobusterScanner() scanner.Scanner {
	return &GobusterScanner{
		BaseScannerTool: NewBaseScannerToolWithVersionArgs("gobuster", scanner.CategoryWebPath, "gobuster", []string{"--version"}),
	}
}

// Scan 执行目录/路径扫描
func (g *GobusterScanner) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.ScanResult, error) {
	// 验证配置
	if err := g.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 准备目标
	targets := g.prepareTargets(config)
	if len(targets) == 0 {
		return nil, scanner.ErrInvalidTarget
	}

	// 扫描所有目标
	var allPaths []scanner.WebPathInfo
	for _, target := range targets {
		// 构建命令参数
		args := g.buildArgs(target, config.Options)

		// 设置超时
		timeout := config.Timeout
		if timeout == 0 {
			timeout = 10 * time.Minute // 默认10分钟超时
		}

		// 执行命令
		output, err := g.ExecuteCommand(ctx, args, timeout)
		if err != nil {
			// 记录错误但继续扫描其他目标
			continue
		}

		// 解析结果
		paths, err := g.parseOutput(output, target)
		if err != nil {
			continue
		}

		allPaths = append(allPaths, paths...)
	}

	// 构建扫描结果
	result := &scanner.ScanResult{
		Data: scanner.WebPathData{
			Paths: allPaths,
		},
	}

	return result, nil
}

// ValidateConfig 验证扫描配置
func (g *GobusterScanner) ValidateConfig(config scanner.ScanConfig) error {
	if err := g.BaseScannerTool.ValidateConfig(config); err != nil {
		return err
	}

	// 验证目标格式
	targets := g.prepareTargets(config)
	if len(targets) == 0 {
		return fmt.Errorf("无有效的Web目标: %s", config.Target)
	}

	return nil
}

// prepareTargets 准备扫描目标
func (g *GobusterScanner) prepareTargets(config scanner.ScanConfig) []string {
	var targets []string

	// 从父级结果中获取Web技术信息
	if len(config.ParentResults) > 0 {
		for _, parentResult := range config.ParentResults {
			if parentResult.Category == scanner.CategoryWebTech {
				if webTechData, ok := parentResult.Data.(scanner.WebTechData); ok {
					if webTechData.URL != "" && webTechData.StatusCode >= 200 && webTechData.StatusCode < 400 {
						targets = append(targets, webTechData.URL)
					}
				}
			}
		}
	}

	// 如果没有父级结果，使用配置的目标
	if len(targets) == 0 {
		target := g.SanitizeTarget(config.Target)
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
func (g *GobusterScanner) buildArgs(target string, options map[string]string) []string {
	args := []string{
		"dir",                    // 目录模式
		"-u", target,             // 目标URL
		"-q",                     // 静默模式
		"-r",                     // 跟随重定向
		"--timeout", "10s",       // 10秒超时
		"--wildcard",             // 通配符检测
		"-t", "50",               // 50线程
	}

	// 字典文件
	if wordlist, exists := options["wordlist"]; exists && wordlist != "" {
		args = append(args, "-w", wordlist)
	} else {
		// 默认字典
		args = append(args, "-w", "/usr/share/wordlists/dirb/common.txt")
	}

	// 文件扩展名
	if extensions, exists := options["extensions"]; exists && extensions != "" {
		args = append(args, "-x", extensions)
	} else {
		// 默认扩展名
		args = append(args, "-x", "php,html,js,txt,xml,json,bak,backup")
	}

	// 状态码过滤
	if statusCodes, exists := options["status_codes"]; exists && statusCodes != "" {
		args = append(args, "-s", statusCodes)
	} else {
		// 默认状态码
		args = append(args, "-s", "200,204,301,302,307,403")
	}

	// 排除状态码
	if excludeStatus, exists := options["exclude_status"]; exists && excludeStatus != "" {
		args = append(args, "-b", excludeStatus)
	}

	// 排除长度
	if excludeLength, exists := options["exclude_length"]; exists && excludeLength != "" {
		args = append(args, "--exclude-length", excludeLength)
	}

	// 用户代理
	if userAgent, exists := options["user_agent"]; exists && userAgent != "" {
		args = append(args, "-a", userAgent)
	}

	// 自定义头
	if headers, exists := options["headers"]; exists && headers != "" {
		for _, header := range strings.Split(headers, ",") {
			args = append(args, "-H", strings.TrimSpace(header))
		}
	}

	// 代理
	if proxy, exists := options["proxy"]; exists && proxy != "" {
		args = append(args, "--proxy", proxy)
	}

	// Cookie
	if cookies, exists := options["cookies"]; exists && cookies != "" {
		args = append(args, "-c", cookies)
	}

	// 递归扫描
	if recursive, exists := options["recursive"]; exists && recursive == "true" {
		args = append(args, "-R")
		if depth, exists := options["depth"]; exists && depth != "" {
			args = append(args, "--depth", depth)
		}
	}

	return args
}

// parseOutput 解析gobuster输出
func (g *GobusterScanner) parseOutput(output []byte, baseURL string) ([]scanner.WebPathInfo, error) {
	lines := g.ParseLines(output)
	var paths []scanner.WebPathInfo

	for _, line := range lines {
		// 跳过状态和错误信息
		if strings.HasPrefix(line, "===============================================================") ||
			strings.HasPrefix(line, "Gobuster") ||
			strings.HasPrefix(line, "by") ||
			strings.HasPrefix(line, "===============================================================") ||
			strings.Contains(line, "Starting gobuster") ||
			strings.Contains(line, "===============================================================") {
			continue
		}

		// 解析路径信息
		pathInfo := g.parsePathLine(line, baseURL)
		if pathInfo.Path != "" {
			paths = append(paths, pathInfo)
		}
	}

	return paths, nil
}

// parsePathLine 解析单个路径行
func (g *GobusterScanner) parsePathLine(line, baseURL string) scanner.WebPathInfo {
	// Gobuster输出格式: /path (Status: 200) [Size: 1234]
	// 或者: /path.php (Status: 301) -> /newpath [Size: 0]

	var pathInfo scanner.WebPathInfo

	// 提取路径
	pathPattern := regexp.MustCompile(`^(/[^\s]*?)`)
	if matches := pathPattern.FindStringSubmatch(line); len(matches) > 1 {
		pathInfo.Path = matches[1]
		pathInfo.URL = strings.TrimSuffix(baseURL, "/") + pathInfo.Path
	}

	// 提取状态码
	statusPattern := regexp.MustCompile(`\(Status:\s*(\d+)\)`)
	if matches := statusPattern.FindStringSubmatch(line); len(matches) > 1 {
		if code, err := strconv.Atoi(matches[1]); err == nil {
			pathInfo.StatusCode = code
		}
	}

	// 提取大小
	sizePattern := regexp.MustCompile(`\[Size:\s*(\d+)\]`)
	if matches := sizePattern.FindStringSubmatch(line); len(matches) > 1 {
		if size, err := strconv.Atoi(matches[1]); err == nil {
			pathInfo.Length = size
		}
	}

	// 检查重定向
	if strings.Contains(line, "->") {
		redirectPattern := regexp.MustCompile(`->\s*([^\s\[]+)`)
		if matches := redirectPattern.FindStringSubmatch(line); len(matches) > 1 {
			pathInfo.Title = "Redirect to: " + matches[1]
		}
	}

	return pathInfo
}

// parseGobusterVHost 解析vhost模式输出（扩展功能）
func (g *GobusterScanner) parseGobusterVHost(output []byte, baseDomain string) ([]scanner.SubdomainInfo, error) {
	lines := g.ParseLines(output)
	var subdomains []scanner.SubdomainInfo

	for _, line := range lines {
		// VHost输出格式: Found: subdomain.example.com (Status: 200) [Size: 1234]
		vhostPattern := regexp.MustCompile(`Found:\s*([^\s]+)\s*\(Status:\s*(\d+)\)`)
		if matches := vhostPattern.FindStringSubmatch(line); len(matches) > 2 {
			subdomain := scanner.SubdomainInfo{
				Domain:    baseDomain,
				Subdomain: matches[1],
				Source:    "gobuster-vhost",
			}

			// 解析IP地址
			if ips, err := g.resolveIPs(matches[1]); err == nil {
				subdomain.IPs = ips
			}

			subdomains = append(subdomains, subdomain)
		}
	}

	return subdomains, nil
}

// resolveIPs 解析域名对应的IP地址（重用subfinder的方法）
func (g *GobusterScanner) resolveIPs(domain string) ([]string, error) {
	// 这里可以重用BaseScannerTool的IP解析功能
	// 或者直接调用net.LookupIP
	return []string{}, nil // 简化实现
}