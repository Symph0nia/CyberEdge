package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/golang-jwt/jwt/v5"
)

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CORS headers are set correctly", func(t *testing.T) {
		router := gin.New()
		allowedOrigins := []string{"http://localhost:3000", "http://localhost:8080"}
		router.Use(CORS(allowedOrigins))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 测试预检请求
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("Reject non-allowed origins", func(t *testing.T) {
		router := gin.New()
		allowedOrigins := []string{"http://localhost:3000"}
		router.Use(CORS(allowedOrigins))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://malicious-site.com")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// 不应该设置CORS头
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("Allow all origins with wildcard", func(t *testing.T) {
		router := gin.New()
		allowedOrigins := []string{"*"}
		router.Use(CORS(allowedOrigins))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "http://any-origin.com")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	jwtSecret := "test-secret-key-for-unit-tests-only"

	// 生成测试JWT令牌
	generateTestToken := func(username string, exp time.Time) string {
		claims := jwt.MapClaims{
			"username": username,
			"exp":      exp.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(jwtSecret))
		return tokenString
	}

	t.Run("Valid JWT token allows access", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(jwtSecret))

		router.GET("/protected", func(c *gin.Context) {
			username, exists := c.Get("username")
			assert.True(t, exists)
			c.JSON(200, gin.H{"username": username})
		})

		validToken := generateTestToken("testuser", time.Now().Add(time.Hour))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "testuser")
	})

	t.Run("Missing token returns 401", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(jwtSecret))

		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header required")
	})

	t.Run("Invalid token format returns 401", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(jwtSecret))

		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("Expired token returns 401", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(jwtSecret))

		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		expiredToken := generateTestToken("testuser", time.Now().Add(-time.Hour))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("Invalid signature returns 401", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(jwtSecret))

		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		// 使用错误的密钥生成token
		claims := jwt.MapClaims{
			"username": "testuser",
			"exp":      time.Now().Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		wrongToken, _ := token.SignedString([]byte("wrong-secret"))

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+wrongToken)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})

	t.Run("Malformed JWT token returns 401", func(t *testing.T) {
		router := gin.New()
		router.Use(JWTMiddleware(jwtSecret))

		router.GET("/protected", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid or expired token")
	})
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Logging middleware logs requests", func(t *testing.T) {
		router := gin.New()
		router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			Output: gin.DefaultWriter,
		}))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// 注意：实际的日志输出测试需要捕获输出，这里只是验证中间件不会破坏正常流程
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Recovery middleware handles panics", func(t *testing.T) {
		router := gin.New()
		router.Use(gin.Recovery())

		router.GET("/panic", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/panic", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// Recovery中间件应该捕获panic并返回500错误
	})
}

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 假设有一个安全头中间件
	securityHeaders := func() gin.HandlerFunc {
		return func(c *gin.Context) {
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "DENY")
			c.Header("X-XSS-Protection", "1; mode=block")
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Next()
		}
	}

	t.Run("Security headers are set", func(t *testing.T) {
		router := gin.New()
		router.Use(securityHeaders())

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	})
}

func TestRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 简单的内存限流中间件示例
	rateLimiter := func(maxRequests int) gin.HandlerFunc {
		requests := make(map[string]int)
		return func(c *gin.Context) {
			clientIP := c.ClientIP()
			if requests[clientIP] >= maxRequests {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
				c.Abort()
				return
			}
			requests[clientIP]++
			c.Next()
		}
	}

	t.Run("Rate limiter allows requests under limit", func(t *testing.T) {
		router := gin.New()
		router.Use(rateLimiter(5))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 发送3个请求，应该都成功
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("Rate limiter blocks requests over limit", func(t *testing.T) {
		router := gin.New()
		router.Use(rateLimiter(2))

		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		// 发送2个请求，应该成功
		for i := 0; i < 2; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		// 第3个请求应该被限制
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusTooManyRequests, w.Code)
		assert.Contains(t, w.Body.String(), "Rate limit exceeded")
	})
}

func TestRequestValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 请求大小限制中间件
	requestSizeLimit := func(maxSize int64) gin.HandlerFunc {
		return func(c *gin.Context) {
			if c.Request.ContentLength > maxSize {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too large"})
				c.Abort()
				return
			}
			c.Next()
		}
	}

	t.Run("Request size validation allows small requests", func(t *testing.T) {
		router := gin.New()
		router.Use(requestSizeLimit(1024)) // 1KB limit

		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		smallPayload := strings.NewReader("small data")
		req := httptest.NewRequest("POST", "/test", smallPayload)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Request size validation blocks large requests", func(t *testing.T) {
		router := gin.New()
		router.Use(requestSizeLimit(10)) // Very small limit

		router.POST("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		largePayload := strings.NewReader("This is a large payload that exceeds the limit")
		req := httptest.NewRequest("POST", "/test", largePayload)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
		assert.Contains(t, w.Body.String(), "Request too large")
	})
}