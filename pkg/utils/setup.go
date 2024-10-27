// utils/setup.go

package utils

import (
	"cyberedge/pkg/api"
	"cyberedge/pkg/api/handlers"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupCollections 设置各个集合的处理器
func SetupCollections(db *mongo.Database) error {
	if db == nil {
		return fmt.Errorf("数据库连接为空")
	}
	handlers.SetTOTPCollection(db.Collection("totp"))
	handlers.SetUserCollection(db.Collection("users"))
	handlers.SetConfigCollection(db.Collection("config"))
	handlers.SetTaskCollection(db.Collection("tasks"))
	return nil
}

// SetupGlobalVariables 设置全局变量
func SetupGlobalVariables(jwtSecret string) error {
	if jwtSecret == "" {
		return fmt.Errorf("JWT secret 为空")
	}
	api.GlobalJWTSecret = jwtSecret
	return nil
}

// SetupRouter 设置API路由
func SetupRouter() (*gin.Engine, error) {
	router := api.SetupRouter()
	if router == nil {
		return nil, fmt.Errorf("设置路由失败")
	}
	return router, nil
}
