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
