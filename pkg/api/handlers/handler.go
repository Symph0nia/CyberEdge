package handlers

import (
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	userCollection   *mongo.Collection
	totpCollection   *mongo.Collection
	configCollection *mongo.Collection
	taskCollection   *mongo.Collection // 新增任务集合
)

// SetUserCollection 设置用户集合
func SetUserCollection(collection *mongo.Collection) {
	userCollection = collection
}

// SetTOTPCollection 设置TOTP集合
func SetTOTPCollection(collection *mongo.Collection) {
	totpCollection = collection
}

// SetConfigCollection 设置MongoDB配置集合
func SetConfigCollection(collection *mongo.Collection) {
	configCollection = collection
}

// SetTaskCollection 设置MongoDB任务集合
func SetTaskCollection(collection *mongo.Collection) {
	taskCollection = collection
}
