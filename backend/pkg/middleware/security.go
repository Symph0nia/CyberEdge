package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 限流配置
	RateLimitRPS    int           // 每秒请求数限制
	RateLimitBurst  int           // 突发请求数
	RateLimitWindow time.Duration // 限流窗口

	// 请求大小限制
	MaxRequestSize int64 // 最大请求体大小

	// 安全头配置
	EnableSecurityHeaders bool
	CSPPolicy            string

	// 输入验证
	MaxHeaderSize   int
	MaxURLLength    int
	AllowedMethods  []string
	BlockedUserAgents []string
}

// DefaultSecurityConfig 默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimitRPS:    100,
		RateLimitBurst:  200,
		RateLimitWindow: time.Minute,
		MaxRequestSize:  50 * 1024 * 1024, // 50MB
		EnableSecurityHeaders: true,
		CSPPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'",
		MaxHeaderSize: 8192,  // 8KB
		MaxURLLength:  2048,  // 2KB
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		BlockedUserAgents: []string{
			"sqlmap", "nikto", "nmap", "masscan", "zap", "w3af",
			"burpsuite", "acunetix", "nessus", "openvas",
		},
	}
}

// RateLimiter 简单的令牌桶限流器
type RateLimiter struct {
	mu       sync.RWMutex
	buckets  map[string]*TokenBucket
	config   *SecurityConfig
	cleanup  *time.Ticker
}

// TokenBucket 令牌桶
type TokenBucket struct {
	tokens    int
	capacity  int
	refillRate int
	lastRefill time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(config *SecurityConfig) *RateLimiter {
	limiter := &RateLimiter{
		buckets: make(map[string]*TokenBucket),
		config:  config,
		cleanup: time.NewTicker(5 * time.Minute), // 每5分钟清理一次
	}

	// 启动清理协程
	go limiter.cleanupRoutine()

	return limiter
}

// cleanupRoutine 清理过期的令牌桶
func (rl *RateLimiter) cleanupRoutine() {
	for range rl.cleanup.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastRefill) > rl.config.RateLimitWindow*2 {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[clientIP]
	if !exists {
		bucket = &TokenBucket{
			tokens:     rl.config.RateLimitBurst,
			capacity:   rl.config.RateLimitBurst,
			refillRate: rl.config.RateLimitRPS,
			lastRefill: time.Now(),
		}
		rl.buckets[clientIP] = bucket
	}

	// 重新填充令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * bucket.refillRate
	bucket.tokens = min(bucket.capacity, bucket.tokens+tokensToAdd)
	bucket.lastRefill = now

	// 检查是否有可用令牌
	if bucket.tokens > 0 {
		bucket.tokens--
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SecurityMiddleware 安全中间件
func SecurityMiddleware(config *SecurityConfig) gin.HandlerFunc {
	rateLimiter := NewRateLimiter(config)

	return gin.HandlerFunc(func(c *gin.Context) {
		// 1. 请求大小限制
		if c.Request.ContentLength > config.MaxRequestSize {
			log.Printf("请求体过大: %d bytes from %s", c.Request.ContentLength, c.ClientIP())
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "请求体过大"})
			c.Abort()
			return
		}

		// 2. URL长度限制
		if len(c.Request.URL.String()) > config.MaxURLLength {
			log.Printf("URL过长: %s from %s", c.Request.URL.String(), c.ClientIP())
			c.JSON(http.StatusBadRequest, gin.H{"error": "URL过长"})
			c.Abort()
			return
		}

		// 3. 方法白名单检查
		methodAllowed := false
		for _, allowedMethod := range config.AllowedMethods {
			if c.Request.Method == allowedMethod {
				methodAllowed = true
				break
			}
		}
		if !methodAllowed {
			log.Printf("不允许的HTTP方法: %s from %s", c.Request.Method, c.ClientIP())
			c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "不允许的HTTP方法"})
			c.Abort()
			return
		}

		// 4. User-Agent黑名单检查
		userAgent := strings.ToLower(c.GetHeader("User-Agent"))
		for _, blockedAgent := range config.BlockedUserAgents {
			if strings.Contains(userAgent, blockedAgent) {
				log.Printf("被阻止的User-Agent: %s from %s", userAgent, c.ClientIP())
				c.JSON(http.StatusForbidden, gin.H{"error": "访问被拒绝"})
				c.Abort()
				return
			}
		}

		// 5. 请求头大小检查
		totalHeaderSize := 0
		for name, values := range c.Request.Header {
			for _, value := range values {
				totalHeaderSize += len(name) + len(value)
			}
		}
		if totalHeaderSize > config.MaxHeaderSize {
			log.Printf("请求头过大: %d bytes from %s", totalHeaderSize, c.ClientIP())
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求头过大"})
			c.Abort()
			return
		}

		// 6. 基本的注入攻击检测
		if detectInjectionAttempt(c) {
			log.Printf("检测到潜在注入攻击 from %s: %s", c.ClientIP(), c.Request.URL.String())
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求包含非法内容"})
			c.Abort()
			return
		}

		// 7. 限流检查
		clientIP := c.ClientIP()
		if !rateLimiter.Allow(clientIP) {
			log.Printf("限流触发: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "请求过于频繁，请稍后重试"})
			c.Abort()
			return
		}

		// 8. 设置安全响应头
		if config.EnableSecurityHeaders {
			setSecurityHeaders(c, config)
		}

		c.Next()
	})
}

// detectInjectionAttempt 检测注入攻击尝试
func detectInjectionAttempt(c *gin.Context) bool {
	// SQL注入检测模式
	sqlPatterns := []string{
		"union", "select", "insert", "update", "delete", "drop",
		"exec", "execute", "sp_", "xp_", "0x", "concat",
		"char(", "ascii(", "substring(", "length(",
		"'", "\"", ";", "--", "/*", "*/",
		"1=1", "1' or '1'='1", "admin'--",
	}

	// XSS检测模式
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "vbscript:",
		"onload=", "onerror=", "onclick=", "onmouseover=",
		"alert(", "confirm(", "prompt(", "document.cookie",
		"window.location", "eval(", "setTimeout(",
	}

	// 路径遍历检测模式
	pathTraversalPatterns := []string{
		"../", "..\\", "..", "%2e%2e", "%252e%252e",
		"/etc/passwd", "/proc/", "\\windows\\",
		"boot.ini", "win.ini",
	}

	// 命令注入检测模式
	cmdInjectionPatterns := []string{
		";", "|", "&", "`", "$(",
		"wget", "curl", "nc", "netcat",
		"/bin/", "/usr/bin/", "cmd.exe", "powershell",
	}

	allPatterns := append(append(append(sqlPatterns, xssPatterns...), pathTraversalPatterns...), cmdInjectionPatterns...)

	// 检查URL参数
	for _, param := range c.Request.URL.Query() {
		for _, value := range param {
			lowerValue := strings.ToLower(value)
			for _, pattern := range allPatterns {
				if strings.Contains(lowerValue, pattern) {
					return true
				}
			}
		}
	}

	// 检查路径
	lowerPath := strings.ToLower(c.Request.URL.Path)
	for _, pattern := range allPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}

	// 检查关键请求头
	headers := []string{"User-Agent", "Referer", "X-Forwarded-For", "X-Real-IP"}
	for _, headerName := range headers {
		headerValue := strings.ToLower(c.GetHeader(headerName))
		for _, pattern := range allPatterns {
			if strings.Contains(headerValue, pattern) {
				return true
			}
		}
	}

	return false
}

// setSecurityHeaders 设置安全响应头
func setSecurityHeaders(c *gin.Context, config *SecurityConfig) {
	// 防止点击劫持
	c.Header("X-Frame-Options", "DENY")

	// 防止MIME类型嗅探
	c.Header("X-Content-Type-Options", "nosniff")

	// XSS保护
	c.Header("X-XSS-Protection", "1; mode=block")

	// 强制HTTPS（如果是HTTPS请求）
	if c.Request.TLS != nil {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
	}

	// 内容安全策略
	if config.CSPPolicy != "" {
		c.Header("Content-Security-Policy", config.CSPPolicy)
	}

	// 引用者策略
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	// 权限策略
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

	// 防止缓存敏感页面
	if strings.Contains(c.Request.URL.Path, "/api/") {
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
	}
}

// InputSanitizer 输入清理中间件
func InputSanitizer() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 对于POST/PUT请求，清理请求体中的潜在恶意内容
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			// 这里可以添加更复杂的请求体清理逻辑
			// 例如，如果是JSON请求，可以解析并清理每个字段
		}

		c.Next()
	})
}

// CSRFProtection CSRF保护中间件（简化版）
func CSRFProtection() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 对于状态改变的请求（POST, PUT, DELETE），检查CSRF token
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" && c.Request.Method != "OPTIONS" {
			// 简化的CSRF检查：要求Referer头
			referer := c.GetHeader("Referer")
			origin := c.GetHeader("Origin")

			// 如果没有Referer和Origin，可能是CSRF攻击
			if referer == "" && origin == "" {
				log.Printf("CSRF保护: 缺少Referer和Origin头 from %s", c.ClientIP())
				c.JSON(http.StatusForbidden, gin.H{"error": "CSRF保护：请求被拒绝"})
				c.Abort()
				return
			}

			// 这里可以添加更严格的CSRF token验证
		}

		c.Next()
	})
}

// RequestLogger 安全请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// 记录请求开始
		log.Printf("请求开始: %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.Next()

		// 记录请求结束
		duration := time.Since(start)
		status := c.Writer.Status()

		log.Printf("请求完成: %s %s - %d (%v) from %s",
			c.Request.Method, c.Request.URL.Path, status, duration, c.ClientIP())

		// 记录异常状态码
		if status >= 400 {
			log.Printf("异常响应: %d - %s %s from %s",
				status, c.Request.Method, c.Request.URL.Path, c.ClientIP())
		}
	})
}

// RecoveryWithLogging 带日志记录的恢复中间件
func RecoveryWithLogging() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// 记录panic详情
		log.Printf("Panic recovered: %v - %s %s from %s",
			recovered, c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "服务器内部错误",
		})
	})
}