// pkg/api/handlers/user_handler.go

package handlers

import (
	"cyberedge/pkg/models"
	"cyberedge/pkg/service"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GenerateQRCode(c *gin.Context) {
	qrCode, accountName, err := h.userService.GenerateQRCode()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 使用multipart响应返回二维码图片和账户名称
	c.JSON(http.StatusOK, gin.H{
		"qrcode":  base64.StdEncoding.EncodeToString(qrCode), // 将二维码图片转为base64
		"account": accountName,
	})
}

func (h *UserHandler) ValidateTOTP(c *gin.Context) {
	var request models.TOTPValidationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "验证码和账户是必需的"})
		return
	}

	token, loginCount, err := h.userService.ValidateTOTP(request.Code, request.Account)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := models.TOTPValidationResponse{
		Status:     "验证码有效",
		Token:      token,
		LoginCount: loginCount,
	}

	c.JSON(http.StatusOK, response)
}

func (h *UserHandler) CheckAuth(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	authenticated, account, err := h.userService.CheckAuth(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"authenticated": authenticated, "account": account})
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	account := c.Param("account")
	user, err := h.userService.GetUserByAccount(account)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var newUser models.User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}
	if err := h.userService.CreateUser(&newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法添加用户"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "用户已添加"})
}

func (h *UserHandler) DeleteUsers(c *gin.Context) {
	var request struct {
		Accounts []string `json:"accounts" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	result, err := h.userService.DeleteUsers(request.Accounts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "用户删除完成",
		"result": result,
	})
}
