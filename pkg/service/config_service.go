package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"fmt"
	"github.com/StackExchange/wmi"
	"net"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"
)

type ConfigService struct {
	configDAO *dao.ConfigDAO
}

// ToolStatus 存储工具的安装状态
type ToolStatus struct {
	Nmap      bool `json:"nmap"`
	Ffuf      bool `json:"ffuf"`
	Subfinder bool `json:"subfinder"`
	HttpX     bool `json:"httpx"`
}

func NewConfigService(configDAO *dao.ConfigDAO) *ConfigService {
	return &ConfigService{configDAO: configDAO}
}

// 原有的函数保持不变
func (s *ConfigService) GetQRCodeStatus() (bool, error) {
	logging.Info("正在获取二维码状态")
	status, err := s.configDAO.GetQRCodeStatus()
	if err != nil {
		logging.Error("获取二维码状态失败: %v", err)
		return false, err
	}
	logging.Info("成功获取二维码状态: %v", status)
	return status, nil
}

func (s *ConfigService) SetQRCodeStatus(enabled bool) error {
	logging.Info("正在设置二维码状态为: %v", enabled)
	err := s.configDAO.SetQRCodeStatus(enabled)
	if err != nil {
		logging.Error("设置二维码状态失败: %v", err)
		return err
	}
	logging.Info("成功设置二维码状态为: %v", enabled)
	return nil
}

// GetLocalIP 获取本机IP地址
func (s *ConfigService) GetLocalIP() (string, error) {
	logging.Info("正在获取本机IP地址")

	// 获取所有网络接口
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		logging.Error("获取网络接口失败: %v", err)
		return "", err
	}

	// 遍历所有网络接口，查找非回环地址的IPv4地址
	for _, addr := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				logging.Info("成功获取本机IP地址: %v", ipnet.IP.String())
				return ipnet.IP.String(), nil
			}
		}
	}

	logging.Error("未找到合适的本机IP地址")
	return "", nil
}

// GetPublicIP 获取外网IP地址
func (s *ConfigService) GetPublicIP() (string, error) {
	logging.Info("正在获取外网IP地址")

	// 执行 curl 命令获取外网IP
	cmd := exec.Command("curl", "-s", "ip.me")
	output, err := cmd.Output()
	if err != nil {
		logging.Error("获取外网IP地址失败: %v", err)
		return "", err
	}

	// 处理输出，获取IP地址
	ip := strings.TrimSpace(string(output))
	logging.Info("成功获取外网IP地址: %v", ip)

	return ip, nil
}

// GetCurrentDirectory 获取程序运行目录
func (s *ConfigService) GetCurrentDirectory() (string, error) {
	logging.Info("正在获取程序运行目录")

	// 获取当前工作目录
	dir, err := os.Getwd()
	if err != nil {
		logging.Error("获取程序运行目录失败: %v", err)
		return "", err
	}

	logging.Info("成功获取程序运行目录: %s", dir)
	return dir, nil
}

// GetKernelVersion 获取系统内核版本
func (s *ConfigService) GetKernelVersion() (string, error) {
	logging.Info("正在获取系统内核版本")

	if runtime.GOOS == "windows" {
		type Win32_OperatingSystem struct {
			Version     string
			BuildNumber string
		}

		var operatingSystems []Win32_OperatingSystem
		err := wmi.Query("SELECT Version, BuildNumber FROM Win32_OperatingSystem", &operatingSystems)
		if err != nil {
			logging.Error("WMI查询系统版本失败: %v", err)
			return "", err
		}

		if len(operatingSystems) > 0 {
			version := fmt.Sprintf("%s (Build %s)",
				operatingSystems[0].Version,
				operatingSystems[0].BuildNumber)
			logging.Info("成功获取Windows系统版本: %s", version)
			return version, nil
		}
		return "", fmt.Errorf("未找到系统版本信息")
	} else {
		// Linux/Unix 系统
		cmd := exec.Command("uname", "-r")
		output, err := cmd.Output()
		if err != nil {
			logging.Error("获取系统内核版本失败: %v", err)
			return "", err
		}
		version := strings.TrimSpace(string(output))
		logging.Info("成功获取系统内核版本: %s", version)
		return version, nil
	}
}

// GetOSDistribution 获取系统发行版信息
func (s *ConfigService) GetOSDistribution() (string, error) {
	logging.Info("正在获取系统发行版信息")

	if runtime.GOOS == "windows" {
		type Win32_OperatingSystem struct {
			Caption        string
			Version        string
			OSArchitecture string
		}

		var operatingSystems []Win32_OperatingSystem
		err := wmi.Query("SELECT Caption, Version, OSArchitecture FROM Win32_OperatingSystem", &operatingSystems)
		if err != nil {
			logging.Error("WMI查询操作系统信息失败: %v", err)
			return "", err
		}

		if len(operatingSystems) > 0 {
			info := fmt.Sprintf("%s %s (%s)",
				operatingSystems[0].Caption,
				operatingSystems[0].Version,
				operatingSystems[0].OSArchitecture)
			logging.Info("成功获取Windows系统信息: %s", info)
			return info, nil
		}
		return "", fmt.Errorf("未找到操作系统信息")
	} else {
		// Linux/Unix 系统代码保持不变
		content, err := os.ReadFile("/etc/os-release")
		if err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "PRETTY_NAME=") {
					info := strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
					logging.Info("成功获取Linux发行版信息: %s", info)
					return info, nil
				}
			}
		}

		cmd := exec.Command("lsb_release", "-d")
		output, err := cmd.Output()
		if err != nil {
			logging.Error("获取Linux发行版信息失败: %v", err)
			return "", err
		}
		info := strings.TrimSpace(strings.TrimPrefix(string(output), "Description:"))
		logging.Info("成功获取Linux发行版信息: %s", info)
		return info, nil
	}
}

// GetCurrentPrivileges 获取当前程序运行权限
func (s *ConfigService) GetCurrentPrivileges() (string, error) {
	logging.Info("正在获取程序运行权限")

	// 获取当前用户信息
	currentUser, err := user.Current()
	if err != nil {
		logging.Error("获取当前用户信息失败: %v", err)
		return "", err
	}

	var privilege string
	if runtime.GOOS == "windows" {
		// Windows 系统下检查管理员权限
		cmd := exec.Command("net", "session")
		err = cmd.Run()
		if err == nil {
			privilege = "Administrator"
		} else {
			privilege = "Normal User"
		}
	} else {
		// Linux/Unix 系统下检查 root 权限
		if currentUser.Uid == "0" {
			privilege = "root"
		} else {
			privilege = "normal user"
		}
	}

	// 构建详细的权限信息
	info := fmt.Sprintf("User: %s, UID: %s, Privilege: %s",
		currentUser.Username,
		currentUser.Uid,
		privilege)

	logging.Info("成功获取程序运行权限: %s", info)
	return info, nil
}

// CheckToolsInstallation 检查所有工具的安装状态
func (s *ConfigService) CheckToolsInstallation() (*ToolStatus, error) {
	logging.Info("正在检查工具安装状态")

	status := &ToolStatus{}

	if runtime.GOOS == "windows" {
		// Windows系统使用WMI检查
		status.Nmap = s.checkWindowsToolExists("nmap.exe")
		status.Ffuf = s.checkWindowsToolExists("ffuf.exe")
		status.Subfinder = s.checkWindowsToolExists("subfinder.exe")
		status.HttpX = s.checkWindowsToolExists("httpx.exe")
	} else {
		// Unix-based系统使用which命令检查
		status.Nmap = s.checkUnixToolExists("nmap")
		status.Ffuf = s.checkUnixToolExists("ffuf")
		status.Subfinder = s.checkUnixToolExists("subfinder")
		status.HttpX = s.checkUnixToolExists("httpx")
	}

	logging.Info("工具安装状态检查完成: Nmap=%v, Ffuf=%v, Subfinder=%v, HttpX=%v",
		status.Nmap, status.Ffuf, status.Subfinder, status.HttpX)

	return status, nil
}

// checkWindowsToolExists 使用WMI检查Windows系统中是否存在指定工具
func (s *ConfigService) checkWindowsToolExists(toolName string) bool {
	type Win32_Process struct {
		ExecutablePath string
	}

	logging.Info("检查Windows工具: %s", toolName)

	query := fmt.Sprintf("SELECT ExecutablePath FROM Win32_Process WHERE Name LIKE '%%%s%%'", toolName)
	var processes []Win32_Process
	err := wmi.Query(query, &processes)

	if err != nil {
		logging.Error("WMI查询工具[%s]失败: %v", toolName, err)
		return false
	}

	// 尝试执行命令验证
	cmd := exec.Command("where", toolName)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		logging.Info("工具[%s]已安装", toolName)
		return true
	}

	logging.Info("工具[%s]未安装", toolName)
	return false
}

// checkUnixToolExists 检查Unix-based系统中是否存在指定工具
func (s *ConfigService) checkUnixToolExists(toolName string) bool {
	logging.Info("检查Unix工具: %s", toolName)

	cmd := exec.Command("which", toolName)
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		logging.Info("工具[%s]已安装", toolName)
		return true
	}

	logging.Info("工具[%s]未安装", toolName)
	return false
}

// GetToolVersion 获取指定工具的版本信息
func (s *ConfigService) GetToolVersion(toolName string) (string, error) {
	var cmd *exec.Cmd

	switch toolName {
	case "nmap":
		cmd = exec.Command("nmap", "--version")
	case "ffuf":
		cmd = exec.Command("ffuf", "-V")
	case "subfinder":
		cmd = exec.Command("subfinder", "-version")
	case "httpx":
		cmd = exec.Command("httpx", "-version")
	default:
		return "", fmt.Errorf("unsupported tool: %s", toolName)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %v", err)
	}

	// 提取第一行作为版本信息
	version := strings.Split(string(output), "\n")[0]
	return strings.TrimSpace(version), nil
}

// 为HTTP处理程序准备的辅助方法
func (s *ConfigService) GetToolsStatus() (map[string]interface{}, error) {
	status, err := s.CheckToolsInstallation()
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"toolsStatus": map[string]interface{}{
			"Nmap":      status.Nmap,
			"Ffuf":      status.Ffuf,
			"Subfinder": status.Subfinder,
			"HttpX":     status.HttpX,
		},
	}

	// 获取已安装工具的版本信息
	versions := make(map[string]string)
	tools := []string{"nmap", "ffuf", "subfinder", "httpx"}

	for _, tool := range tools {
		if status.GetToolStatus(tool) {
			if version, err := s.GetToolVersion(tool); err == nil {
				versions[tool] = version
			}
		}
	}

	if len(versions) > 0 {
		result["versions"] = versions
	}

	return result, nil
}

// GetToolStatus 辅助方法，用于根据工具名获取状态
func (ts *ToolStatus) GetToolStatus(toolName string) bool {
	switch strings.ToLower(toolName) {
	case "nmap":
		return ts.Nmap
	case "ffuf":
		return ts.Ffuf
	case "subfinder":
		return ts.Subfinder
	case "httpx":
		return ts.HttpX
	default:
		return false
	}
}
