package handlers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
)

var configCollection *mongo.Collection

// SetConfigCollection 设置MongoDB配置集合
func SetConfigCollection(collection *mongo.Collection) {
	configCollection = collection
}

// GetQRCodeStatus 从数据库获取二维码接口的状态
func GetQRCodeStatus() (bool, error) {
	var config struct {
		QRCodeEnabled bool `bson:"qrcode_enabled"`
	}

	err := configCollection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 默认值，如果没有文档
			return true, nil
		}
		return false, fmt.Errorf("无法获取二维码状态: %v", err)
	}
	return config.QRCodeEnabled, nil
}

// SetQRCodeStatus 更新二维码接口的状态
func SetQRCodeStatus(enabled bool) error {
	_, err := configCollection.UpdateOne(
		context.Background(),
		bson.M{},
		bson.M{"$set": bson.M{"qrcode_enabled": enabled}},
		options.Update().SetUpsert(true), // 如果没有文档则插入新文档
	)
	return err
}

// SetQRCodeStatusHandler 设置或查询二维码接口的状态
func SetQRCodeStatusHandler(c *gin.Context) {
	if c.Request.Method == http.MethodGet {
		enabled, err := GetQRCodeStatus() // 从数据库获取二维码状态
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "无法获取二维码状态"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"enabled": enabled})
		return
	}

	var request struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
		return
	}

	if err := SetQRCodeStatus(request.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法更新二维码状态"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "二维码接口状态已更新", "enabled": request.Enabled})
}
