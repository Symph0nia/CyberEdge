package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConfigDAO struct {
	collection *mongo.Collection
}

func NewConfigDAO(collection *mongo.Collection) *ConfigDAO {
	return &ConfigDAO{collection: collection}
}

func (d *ConfigDAO) GetQRCodeStatus() (bool, error) {
	logging.Info("正在获取二维码状态")
	var config struct {
		QRCodeEnabled bool `bson:"qrcode_enabled"`
	}

	err := d.collection.FindOne(context.Background(), bson.M{}).Decode(&config)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Info("未找到二维码状态配置，使用默认值 true")
			return true, nil // 默认值，如果没有文档
		}
		logging.Error("获取二维码状态失败: %v", err)
		return false, err
	}
	logging.Info("成功获取二维码状态: %v", config.QRCodeEnabled)
	return config.QRCodeEnabled, nil
}

func (d *ConfigDAO) SetQRCodeStatus(enabled bool) error {
	logging.Info("正在设置二维码状态为: %v", enabled)
	_, err := d.collection.UpdateOne(
		context.Background(),
		bson.M{},
		bson.M{"$set": bson.M{"qrcode_enabled": enabled}},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		logging.Error("设置二维码状态失败: %v", err)
		return err
	}
	logging.Info("成功设置二维码状态为: %v", enabled)
	return nil
}
