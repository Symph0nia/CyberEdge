package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

// 假设 userCollection 是一个已初始化的 MongoDB 集合
var userCollection *mongo.Collection

type User struct {
	Account    string `bson:"account" json:"account"`
	Secret     string `bson:"secret" json:"-"`
	LoginCount int    `bson:"loginCount" json:"loginCount"`
}

// SetUserCollection 设置用户集合
func SetUserCollection(collection *mongo.Collection) {
	userCollection = collection
}

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

			var users []User
			if err := cursor.All(context.Background(), &users); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "无法解析用户数据"})
				return
			}
			c.JSON(http.StatusOK, users)
		} else {
			var user User
			err := userCollection.FindOne(context.Background(), bson.M{"account": account}).Decode(&user)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
				return
			}
			c.JSON(http.StatusOK, user)
		}

	case http.MethodPost:
		var newUser User
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
