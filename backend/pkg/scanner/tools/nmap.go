package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cyberedge/pkg/scanner"
)

// NmapScanner Nmap端口扫描工具
type NmapScanner struct {
	*BaseScannerTool
}

// NewNmapScanner 创建Nmap扫描工具
func NewNmapScanner() scanner.Scanner {
	return &NmapScanner{
		BaseScannerTool: NewBaseScannerToolWithVersionArgs("nmap", scanner.CategoryPort, "nmap", []string{"--version"}),
	}
}

// Scan 执行端口扫描
func (n *NmapScanner) Scan(ctx context.Context, config scanner.ScanConfig) (*scanner.ScanResult, error) {
	// 验证配置
	if err := n.ValidateConfig(config); err != nil {
		return nil, err
	}

	// 清理目标
	target := n.SanitizeTarget(config.Target)
	if target == "" {
		return nil, scanner.ErrInvalidTarget
	}

	// 构建命令参数
	args := n.buildArgs(target, config.Options)

	// 设置超时
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 5 * time.Minute // 默认5分钟超时
	}

	// 执行命令
	output, err := n.ExecuteCommand(ctx, args, timeout)
	if err != nil {
		return nil, fmt.Errorf("nmap执行失败: %w", err)
	}

	// 解析结果
	ports, err := n.parseOutput(output, target)
	if err != nil {
		return nil, fmt.Errorf("解析nmap输出失败: %w", err)
	}

	// 构建扫描结果
	result := &scanner.ScanResult{
		Data: scanner.PortData{
			Ports: ports,
		},
	}

	return result, nil
}

// ValidateConfig 验证扫描配置
func (n *NmapScanner) ValidateConfig(config scanner.ScanConfig) error {
	if err := n.BaseScannerTool.ValidateConfig(config); err != nil {
		return err
	}

	// 验证目标格式
	target := n.SanitizeTarget(config.Target)
	if target == "" {
		return fmt.Errorf("目标格式无效: %s", config.Target)
	}

	return nil
}

// buildArgs 构建命令参数
func (n *NmapScanner) buildArgs(target string, options map[string]string) []string {
	args := []string{
		"-oX", "-",    // XML输出到标准输出
		"--max-rate", "1000", // 限制扫描速率
		"-T4",         // 时序模板4（快速）
	}

	// 扫描类型
	scanType := "connect"
	if sType, exists := options["scan_type"]; exists {
		scanType = sType
	}

	switch scanType {
	case "syn":
		args = append(args, "-sS") // SYN扫描
	case "udp":
		args = append(args, "-sU") // UDP扫描
	case "connect":
		args = append(args, "-sT") // TCP connect扫描
	case "version":
		args = append(args, "-sV") // 版本扫描
	default:
		args = append(args, "-sS") // 默认SYN扫描
	}

	// 端口范围
	if ports, exists := options["ports"]; exists && ports != "" {
		args = append(args, "-p", ports)
	} else {
		args = append(args, "--top-ports", "1000") // 默认扫描top 1000端口
	}

	// 服务检测
	if detectService, exists := options["detect_service"]; exists && detectService == "true" {
		args = append(args, "-sV")
	}

	// 操作系统检测
	if detectOS, exists := options["detect_os"]; exists && detectOS == "true" {
		args = append(args, "-O")
	}

	// 脚本扫描
	if scripts, exists := options["scripts"]; exists && scripts != "" {
		args = append(args, "--script", scripts)
	}

	// 并发度
	if parallelism, exists := options["parallelism"]; exists && parallelism != "" {
		args = append(args, "--min-parallelism", parallelism)
	}

	// 添加目标
	args = append(args, target)

	return args
}

// parseOutput 解析nmap XML输出
func (n *NmapScanner) parseOutput(output []byte, target string) ([]scanner.PortInfo, error) {
	// 使用简化的XML解析，提取端口信息
	lines := strings.Split(string(output), "\n")
	var ports []scanner.PortInfo

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "<port ") && strings.Contains(line, "state=") {
			port, err := n.parsePortLine(line)
			if err != nil {
				continue // 跳过解析失败的行
			}
			ports = append(ports, port)
		}
	}

	return ports, nil
}

// parsePortLine 解析单个端口行
func (n *NmapScanner) parsePortLine(line string) (scanner.PortInfo, error) {
	var port scanner.PortInfo

	// 解析端口号
	if portStr := n.extractAttribute(line, "portid"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			port.Port = p
		}
	}

	// 解析协议
	if protocol := n.extractAttribute(line, "protocol"); protocol != "" {
		port.Protocol = protocol
	}

	// 解析状态
	if state := n.extractAttribute(line, "state"); state != "" {
		port.State = state
	}

	// 如果包含服务信息
	if strings.Contains(line, "<service ") {
		service := &scanner.ServiceInfo{}
		if name := n.extractAttribute(line, "name"); name != "" {
			service.Name = name
		}
		if version := n.extractAttribute(line, "version"); version != "" {
			service.Version = version
		}
		if product := n.extractAttribute(line, "product"); product != "" {
			service.Fingerprint = product
		}
		port.Service = service
	}

	return port, nil
}

// extractAttribute 从XML行中提取属性值
func (n *NmapScanner) extractAttribute(line, attr string) string {
	attrPattern := attr + `="`
	start := strings.Index(line, attrPattern)
	if start == -1 {
		return ""
	}
	start += len(attrPattern)
	end := strings.Index(line[start:], `"`)
	if end == -1 {
		return ""
	}
	return line[start : start+end]
}

// NmapResult Nmap原始结果结构（用于JSON解析）
type NmapResult struct {
	Hosts []NmapHost `json:"hosts"`
}

type NmapHost struct {
	IP    string     `json:"ip"`
	Ports []NmapPort `json:"ports"`
}

type NmapPort struct {
	Port     int         `json:"port"`
	Protocol string      `json:"protocol"`
	State    string      `json:"state"`
	Service  NmapService `json:"service"`
}

type NmapService struct {
	Name    string `json:"name"`
	Product string `json:"product"`
	Version string `json:"version"`
}

// parseJSONOutput 解析JSON格式输出（备用方案）
func (n *NmapScanner) parseJSONOutput(output []byte) ([]scanner.PortInfo, error) {
	var result NmapResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	var ports []scanner.PortInfo
	for _, host := range result.Hosts {
		for _, nmapPort := range host.Ports {
			port := scanner.PortInfo{
				Port:     nmapPort.Port,
				Protocol: nmapPort.Protocol,
				State:    nmapPort.State,
			}

			if nmapPort.Service.Name != "" {
				port.Service = &scanner.ServiceInfo{
					Name:        nmapPort.Service.Name,
					Version:     nmapPort.Service.Version,
					Fingerprint: nmapPort.Service.Product,
				}
			}

			ports = append(ports, port)
		}
	}

	return ports, nil
}