package tools

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"cyberedge/pkg/scanner"
)

// BaseScannerTool 基础扫描工具实现
type BaseScannerTool struct {
	name        string
	category    scanner.ScanCategory
	cmdPath     string
	available   bool
	versionArgs []string // 用于检查可用性的参数
}

// NewBaseScannerTool 创建基础扫描工具
func NewBaseScannerTool(name string, category scanner.ScanCategory, cmdPath string) *BaseScannerTool {
	return NewBaseScannerToolWithVersionArgs(name, category, cmdPath, []string{"--version"})
}

// NewBaseScannerToolWithVersionArgs 创建带自定义版本检查参数的基础扫描工具
func NewBaseScannerToolWithVersionArgs(name string, category scanner.ScanCategory, cmdPath string, versionArgs []string) *BaseScannerTool {
	tool := &BaseScannerTool{
		name:        name,
		category:    category,
		cmdPath:     cmdPath,
		versionArgs: versionArgs,
	}

	// 检查工具是否可用
	tool.available = tool.checkAvailability()
	return tool
}

// GetName 获取工具名称
func (b *BaseScannerTool) GetName() string {
	return b.name
}

// GetCategory 获取工具类别
func (b *BaseScannerTool) GetCategory() scanner.ScanCategory {
	return b.category
}

// IsAvailable 检查工具是否可用
func (b *BaseScannerTool) IsAvailable() bool {
	return b.available
}

// checkAvailability 检查命令是否存在
func (b *BaseScannerTool) checkAvailability() bool {
	if b.cmdPath == "" {
		return false
	}

	// 尝试执行指定的版本检查参数
	if len(b.versionArgs) == 0 {
		// 如果没有版本参数，只检查命令是否存在
		_, err := exec.LookPath(b.cmdPath)
		return err == nil
	}

	cmd := exec.Command(b.cmdPath, b.versionArgs...)
	return cmd.Run() == nil
}

// ExecuteCommand 执行命令行工具
func (b *BaseScannerTool) ExecuteCommand(ctx context.Context, cmdArgs []string, timeout time.Duration) ([]byte, error) {
	// 设置超时上下文
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// 创建命令
	cmd := exec.CommandContext(ctx, b.cmdPath, cmdArgs...)

	// 执行命令
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

// ValidateConfig 基础配置验证
func (b *BaseScannerTool) ValidateConfig(config scanner.ScanConfig) error {
	if config.Target == "" {
		return scanner.ErrInvalidTarget
	}

	if config.ProjectID == 0 {
		return scanner.ErrInvalidProjectID
	}

	return nil
}

// SanitizeTarget 清理和验证目标
func (b *BaseScannerTool) SanitizeTarget(target string) string {
	// 基础清理：去除空格和特殊字符
	target = strings.TrimSpace(target)
	target = strings.ToLower(target)

	// 移除协议前缀
	if strings.HasPrefix(target, "http://") {
		target = strings.TrimPrefix(target, "http://")
	}
	if strings.HasPrefix(target, "https://") {
		target = strings.TrimPrefix(target, "https://")
	}

	// 移除路径部分
	if idx := strings.Index(target, "/"); idx != -1 {
		target = target[:idx]
	}

	return target
}

// ParseLines 解析命令输出行
func (b *BaseScannerTool) ParseLines(output []byte) []string {
	lines := strings.Split(string(output), "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}

	return result
}