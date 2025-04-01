package handlers

import (
	"cyberedge/pkg/models"
	"cyberedge/pkg/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// 构造响应数据
	toolsInfo := map[string]interface{}{
		"installedStatus": map[string]bool{
			"Nmap":      toolStatus.Nmap,
			"Ffuf":      toolStatus.Ffuf,
			"Subfinder": toolStatus.Subfinder,
			"HttpX":     toolStatus.HttpX,
			"Fscan":     toolStatus.Fscan,
			"Afrog":     toolStatus.Afrog,  // 新增 Afrog
			"Nuclei":    toolStatus.Nuclei, // 新增 Nuclei
		},
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

// GetToolConfigs 获取所有工具配置
func (h *ConfigHandler) GetToolConfigs(c *gin.Context) {
	configs, err := h.configService.GetToolConfigs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "获取工具配置列表失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    configs,
		"message": "获取工具配置列表成功",
	})
}

// GetDefaultToolConfig 获取默认工具配置
func (h *ConfigHandler) GetDefaultToolConfig(c *gin.Context) {
	config, err := h.configService.GetDefaultToolConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "获取默认工具配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    config,
		"message": "获取默认工具配置成功",
	})
}

// GetToolConfigByID 根据ID获取工具配置
func (h *ConfigHandler) GetToolConfigByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "缺少ID参数",
		})
		return
	}

	config, err := h.configService.GetToolConfigByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "获取工具配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"data":    config,
		"message": "获取工具配置成功",
	})
}

// CreateToolConfig 创建工具配置
func (h *ConfigHandler) CreateToolConfig(c *gin.Context) {
	var config models.ToolConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 验证配置有效性
	if err := h.configService.ValidateToolConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "配置验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 创建新配置
	createdConfig, err := h.configService.CreateToolConfig(&config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "创建工具配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"data":    createdConfig,
		"message": "创建工具配置成功",
	})
}

// UpdateToolConfig 更新工具配置
func (h *ConfigHandler) UpdateToolConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "缺少ID参数",
		})
		return
	}

	var config models.ToolConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请求格式错误",
			"error":   err.Error(),
		})
		return
	}

	// 确保ID一致
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "无效的ID格式",
			"error":   err.Error(),
		})
		return
	}
	config.ID = objID

	// 验证配置有效性
	if err := h.configService.ValidateToolConfig(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "配置验证失败",
			"error":   err.Error(),
		})
		return
	}

	// 更新配置
	if err := h.configService.UpdateToolConfig(&config); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "更新工具配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "更新工具配置成功",
	})
}

// DeleteToolConfig 删除工具配置
func (h *ConfigHandler) DeleteToolConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "缺少ID参数",
		})
		return
	}

	if err := h.configService.DeleteToolConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "删除工具配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "删除工具配置成功",
	})
}

// SetDefaultToolConfig 设置默认工具配置
func (h *ConfigHandler) SetDefaultToolConfig(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "缺少ID参数",
		})
		return
	}

	if err := h.configService.SetDefaultToolConfig(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "设置默认工具配置失败",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "设置默认工具配置成功",
	})
}
