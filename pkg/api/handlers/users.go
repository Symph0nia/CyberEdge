package handlers

import (
	"context"
	"net/http"

	"cyberedge/pkg/models" // 导入用户模型包
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// HandleUsers 处理用户的CRUD操作
func HandleUsers(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		account := c.Param("account")
		if account == "" {
			cursor, err := userCollection.Find(context.Background(), bson.M{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取用户"})
				return
			}
			defer cursor.Close(context.Background())

			var users []models.User // 使用新模型
			if err := cursor.All(context.Background(), &users); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析用户数据"})
				return
			}
			c.JSON(http.StatusOK, users)
		} else {
			var user models.User // 使用新模型
			err := userCollection.FindOne(context.Background(), bson.M{"account": account}).Decode(&user)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
				return
			}
			c.JSON(http.StatusOK, user)
		}

	case http.MethodPost:
		var newUser models.User // 使用新模型
		if err := c.ShouldBindJSON(&newUser); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		_, err := userCollection.InsertOne(context.Background(), newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法添加用户"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"status": "用户已添加"})

	case http.MethodDelete:
		account := c.Param("account")
		if account == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "需要提供用户名"})
			return
		}
		_, err := userCollection.DeleteOne(context.Background(), bson.M{"account": account})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法删除用户"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "用户已删除"})

	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "不支持的方法"})
	}
}
