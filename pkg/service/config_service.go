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
