package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"cyberedge/pkg/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"image/png"
	"math/big"
	"net/http"
	"strings"
	"time"
)

// 随机生成一个16位的字符串
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("无法生成随机字符串: %w", err)
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

// GenerateQRCode 生成二维码并返回给客户端
func GenerateQRCode(c *gin.Context) {
	qrcodeEnabled, err := GetQRCodeStatus() // 从数据库获取二维码状态
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取二维码状态"})
		return
	}

	if !qrcodeEnabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "二维码接口已关闭"})
		return
	}

	accountName, err := generateRandomString(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成账户名称"})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CyberEdgeAdmin",
		AccountName: accountName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成密钥"})
		return
	}

	_, err = userCollection.InsertOne(context.Background(), bson.M{
		"account":    accountName,
		"secret":     key.Secret(),
		"loginCount": 0,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法存储密钥"})
		return
	}

	img, err := key.Image(200, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成二维码"})
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法编码二维码"})
		return
	}

	c.Data(http.StatusOK, "image/png", buf.Bytes())
}

// ValidateTOTP 验证用户输入的TOTP代码
func ValidateTOTP(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Code    string `json:"code" binding:"required"`
			Account string `json:"account" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "验证码和账户是必需的"})
			return
		}

		var user models.User
		err := userCollection.FindOne(context.Background(), bson.M{"account": request.Account}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无法找到密钥"})
			return
		}

		if totp.Validate(request.Code, user.Secret) {
			_, err := userCollection.UpdateOne(
				context.Background(),
				bson.M{"account": request.Account},
				bson.M{"$inc": bson.M{"loginCount": 1}},
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新登录次数"})
				return
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
				"account": request.Account,
				"exp":     time.Now().Add(time.Hour * 72).Unix(),
			})

			tokenString, err := token.SignedString([]byte(jwtSecret))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法生成令牌"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"status": "验证码有效", "token": tokenString, "loginCount": user.LoginCount + 1})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "验证码无效"})
		}
	}
}

// CheckAuth 验证JWT并返回认证状态
func CheckAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取JWT
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "未提供令牌"})
			return
		}

		// 去掉 "Bearer " 前缀
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "无效的令牌格式"})
			return
		}

		// 验证JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// 确保使用的是预期的签名方法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "无效的令牌"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.JSON(http.StatusOK, gin.H{"authenticated": true, "account": claims["account"]})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"authenticated": false, "error": "无效的令牌"})
		}
	}
}
