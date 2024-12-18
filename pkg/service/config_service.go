package service

import (
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"fmt"
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
	Nmap      bool
	Ffuf      bool
	Subfinder bool
	HttpX     bool
	Fscan     bool // 新增
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

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_OperatingSystem).Version")
		output, err := cmd.Output()
		if err != nil {
			logging.Error("获取Windows系统版本失败: %v", err)
			return "", err
		}

		version := strings.TrimSpace(string(output))

		// 获取 Build 号
		buildCmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_OperatingSystem).BuildNumber")
		buildOutput, err := buildCmd.Output()
		if err == nil {
			buildNumber := strings.TrimSpace(string(buildOutput))
			version = fmt.Sprintf("%s (Build %s)", version, buildNumber)
		}

		logging.Info("成功获取Windows系统版本: %s", version)
		return version, nil

	case "darwin": // macOS
		cmd := exec.Command("uname", "-r")
		output, err := cmd.Output()
		if err != nil {
			logging.Error("获取macOS内核版本失败: %v", err)
			return "", err
		}
		version := strings.TrimSpace(string(output))
		logging.Info("成功获取macOS内核版本: %s", version)
		return version, nil

	default: // Linux 和其他 Unix 系统
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

	switch runtime.GOOS {
	case "windows":
		// 获取操作系统名称
		captionCmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_OperatingSystem).Caption")
		captionOutput, err := captionCmd.Output()
		if err != nil {
			logging.Error("获取Windows系统名称失败: %v", err)
			return "", err
		}
		caption := strings.TrimSpace(string(captionOutput))

		// 获取版本号
		versionCmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_OperatingSystem).Version")
		versionOutput, err := versionCmd.Output()
		if err != nil {
			logging.Error("获取Windows版本号失败: %v", err)
			return "", err
		}
		version := strings.TrimSpace(string(versionOutput))

		// 获取系统架构
		archCmd := exec.Command("powershell", "-Command", "(Get-WmiObject Win32_OperatingSystem).OSArchitecture")
		archOutput, err := archCmd.Output()
		if err != nil {
			logging.Error("获取Windows架构信息失败: %v", err)
			return "", err
		}
		arch := strings.TrimSpace(string(archOutput))

		info := fmt.Sprintf("%s %s (%s)", caption, version, arch)
		logging.Info("成功获取Windows系统信息: %s", info)
		return info, nil

	case "darwin":
		// 获取 macOS 版本名称
		productCmd := exec.Command("sw_vers", "-productName")
		productOutput, err := productCmd.Output()
		if err != nil {
			logging.Error("获取macOS产品名称失败: %v", err)
			return "", err
		}
		productName := strings.TrimSpace(string(productOutput))

		// 获取 macOS 版本号
		versionCmd := exec.Command("sw_vers", "-productVersion")
		versionOutput, err := versionCmd.Output()
		if err != nil {
			logging.Error("获取macOS版本号失败: %v", err)
			return "", err
		}
		version := strings.TrimSpace(string(versionOutput))

		// 获取系统架构
		archCmd := exec.Command("uname", "-m")
		archOutput, err := archCmd.Output()
		if err != nil {
			logging.Error("获取macOS架构信息失败: %v", err)
			return "", err
		}
		arch := strings.TrimSpace(string(archOutput))

		info := fmt.Sprintf("%s %s (%s)", productName, version, arch)
		logging.Info("成功获取macOS系统信息: %s", info)
		return info, nil

	default: // Linux 和其他 Unix 系统
		// 首先尝试读取 /etc/os-release
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

		// 如果读取 /etc/os-release 失败，尝试 lsb_release
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
		// Windows系统使用where命令检查
		status.Nmap = s.checkWindowsToolExists("nmap.exe")
		status.Ffuf = s.checkWindowsToolExists("ffuf.exe")
		status.Subfinder = s.checkWindowsToolExists("subfinder.exe")
		status.HttpX = s.checkWindowsToolExists("httpx.exe")
		status.Fscan = s.checkWindowsToolExists("fscan.exe") // 新增
	} else {
		// Unix-based系统使用which命令检查
		status.Nmap = s.checkUnixToolExists("nmap")
		status.Ffuf = s.checkUnixToolExists("ffuf")
		status.Subfinder = s.checkUnixToolExists("subfinder")
		status.HttpX = s.checkUnixToolExists("httpx")
		status.Fscan = s.checkUnixToolExists("fscan") // 新增
	}

	logging.Info("工具安装状态检查完成: Nmap=%v, Ffuf=%v, Subfinder=%v, HttpX=%v, Fscan=%v",
		status.Nmap, status.Ffuf, status.Subfinder, status.HttpX, status.Fscan)

	return status, nil
}

// checkWindowsToolExists 检查Windows系统中是否存在指定工具
func (s *ConfigService) checkWindowsToolExists(toolName string) bool {
	logging.Info("检查Windows工具: %s", toolName)

	// 首先使用 where 命令检查
	whereCmd := exec.Command("where", toolName)
	if output, err := whereCmd.Output(); err == nil && len(output) > 0 {
		logging.Info("工具[%s]已安装", toolName)
		return true
	}

	// 如果 where 命令失败，尝试使用 PowerShell 检查
	psCmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-Command %s -ErrorAction SilentlyContinue", toolName))
	if output, err := psCmd.Output(); err == nil && len(output) > 0 {
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
			"Fscan":     status.Fscan,
		},
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
	case "fscan":
		return ts.Fscan
	default:
		return false
	}
}
