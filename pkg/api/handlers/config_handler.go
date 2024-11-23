package handlers

import (
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ConfigHandler struct {
	configService *service.ConfigService
}

func NewConfigHandler(configService *service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configService: configService}
}

func (h *ConfigHandler) GetQRCodeStatus(c *gin.Context) {
	enabled, err := h.configService.GetQRCodeStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取二维码状态"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"enabled": enabled})
}

func (h *ConfigHandler) SetQRCodeStatus(c *gin.Context) {
	var request struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := h.configService.SetQRCodeStatus(request.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新二维码状态"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "二维码接口状态已更新", "enabled": request.Enabled})
}

// GetSystemInfo 获取系统信息的API接口
func (h *ConfigHandler) GetSystemInfo(c *gin.Context) {
	// 创建一个map来存储所有系统信息
	systemInfo := make(map[string]interface{})

	// 获取程序运行目录
	currentDir, err := h.configService.GetCurrentDirectory()
	if err != nil {
		systemInfo["currentDirectory"] = "获取失败: " + err.Error()
	} else {
		systemInfo["currentDirectory"] = currentDir
	}

	// 获取本机IP
	localIP, err := h.configService.GetLocalIP()
	if err != nil {
		systemInfo["localIP"] = "获取失败: " + err.Error()
	} else {
		systemInfo["localIP"] = localIP
	}

	// 获取外网IP
	publicIP, err := h.configService.GetPublicIP()
	if err != nil {
		systemInfo["publicIP"] = "获取失败: " + err.Error()
	} else {
		systemInfo["publicIP"] = publicIP
	}

	// 获取系统内核版本
	kernelVersion, err := h.configService.GetKernelVersion()
	if err != nil {
		systemInfo["kernelVersion"] = "获取失败: " + err.Error()
	} else {
		systemInfo["kernelVersion"] = kernelVersion
	}

	// 获取系统发行版信息
	osDistribution, err := h.configService.GetOSDistribution()
	if err != nil {
		systemInfo["osDistribution"] = "获取失败: " + err.Error()
	} else {
		systemInfo["osDistribution"] = osDistribution
	}

	// 获取程序运行权限
	privileges, err := h.configService.GetCurrentPrivileges()
	if err != nil {
		systemInfo["privileges"] = "获取失败: " + err.Error()
	} else {
		systemInfo["privileges"] = privileges
	}

	// 返回所有系统信息
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"systemInfo": systemInfo,
		},
		"message": "系统信息获取成功",
	})
}

// GetToolsStatus 获取工具安装状态的API接口
func (h *ConfigHandler) GetToolsStatus(c *gin.Context) {
	// 获取工具安装状态
	toolStatus, err := h.configService.CheckToolsInstallation()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "获取工具状态失败",
			"error":   err.Error(),
		})
		return
	}

	// 创建工具版本信息映射
	toolVersions := make(map[string]string)
	tools := []string{"nmap", "ffuf", "subfinder", "httpx"}

	// 尝试获取已安装工具的版本信息
	for _, tool := range tools {
		if toolStatus.GetToolStatus(tool) {
			version, err := h.configService.GetToolVersion(tool)
			if err == nil {
				toolVersions[tool] = version
			}
		}
	}

	// 构造响应数据
	toolsInfo := map[string]interface{}{
		"installedStatus": map[string]bool{
			"Nmap":      toolStatus.Nmap,
			"Ffuf":      toolStatus.Ffuf,
			"Subfinder": toolStatus.Subfinder,
			"HttpX":     toolStatus.HttpX,
		},
	}

	// 如果有版本信息，添加到响应中
	if len(toolVersions) > 0 {
		toolsInfo["versions"] = toolVersions
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"toolsInfo": toolsInfo,
		},
		"message": "工具状态检测成功",
	})
}
