package handlers

import (
	"cyberedge/pkg/service"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type UserHandler struct {
	userService service.UserServiceInterface
}

func NewUserHandler(userService service.UserServiceInterface) *UserHandler {
	return &UserHandler{userService: userService}
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   token,
		"message": "登录成功",
	})
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8,max=128"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	err := h.userService.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		// 根据错误类型返回相应的错误码
		switch {
		case errors.Is(err, service.ErrUserExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		case errors.Is(err, service.ErrEmailExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被使用"})
		case errors.Is(err, service.ErrWeakPassword):
			c.JSON(http.StatusBadRequest, gin.H{"error": "密码强度不足"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "用户注册成功",
	})
}

// CheckAuth 检查认证状态
func (h *UserHandler) CheckAuth(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	// 安全地提取Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.userService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// GetUsers 获取所有用户
func (h *UserHandler) GetUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户列表失败"})
		return
	}

	// 转换用户数据为前端期望的格式
	var userList []map[string]interface{}
	for _, user := range users {
		userList = append(userList, map[string]interface{}{
			"account":    user.Username, // 映射 username 到 account
			"loginCount": 0,             // 临时设置为0，后续可以添加登录次数追踪
		})
	}

	c.JSON(http.StatusOK, userList)
}

// GetUser 根据ID获取用户
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// CreateUser 创建用户
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8,max=128"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	err := h.userService.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "用户创建成功"})
}

// DeleteUser 删除用户
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	err = h.userService.DeleteUser(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除用户失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// Setup2FA 设置双因子认证
func (h *UserHandler) Setup2FA(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	secret, qrCode, err := h.userService.Setup2FA(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":  secret,
		"qr_code": qrCode,
	})
}

// Verify2FA 验证双因子认证
func (h *UserHandler) Verify2FA(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Code     string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	err := h.userService.Verify2FA(req.Username, req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "验证成功"})
}

// Disable2FA 禁用双因子认证
func (h *UserHandler) Disable2FA(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	err := h.userService.Disable2FA(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "双因子认证已禁用"})
}

// GenerateQRCode 生成二维码（占位符方法）
func (h *UserHandler) GenerateQRCode(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "请使用 Setup2FA 接口设置双因子认证"})
}

// ValidateTOTP 验证TOTP（占位符方法）
func (h *UserHandler) ValidateTOTP(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "请使用 Verify2FA 接口验证双因子认证"})
}

// GetQRCodeStatus 获取二维码设置状态
func (h *UserHandler) GetQRCodeStatus(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
		return
	}

	// 安全地提取Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证格式"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.userService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is2FAEnabled": user.Is2FAEnabled,
		"username":     user.Username,
	})
}

// ChangePassword 修改密码
func (h *UserHandler) ChangePassword(c *gin.Context) {
	// 先检查认证
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供认证token"})
		return
	}

	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8,max=128"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的认证格式"})
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := h.userService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的token"})
		return
	}

	// 修改密码
	err = h.userService.ChangePassword(user.Username, req.CurrentPassword, req.NewPassword)
	if err != nil {
		if err.Error() == "INVALID_PASSWORD" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "当前密码错误"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "修改密码失败"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "密码修改成功",
	})
}