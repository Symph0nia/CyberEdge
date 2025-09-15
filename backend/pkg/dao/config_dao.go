package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
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

// 获取工具配置列表
func (d *ConfigDAO) GetToolConfigs() ([]*models.ToolConfig, error) {
	logging.Info("正在获取所有工具配置")

	cursor, err := d.collection.Find(context.Background(), bson.M{"name": bson.M{"$exists": true}})
	if err != nil {
		logging.Error("获取工具配置列表失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var configs []*models.ToolConfig
	if err := cursor.All(context.Background(), &configs); err != nil {
		logging.Error("解析工具配置列表失败: %v", err)
		return nil, err
	}

	logging.Info("成功获取工具配置列表，共 %d 条记录", len(configs))
	return configs, nil
}

// 获取默认工具配置
func (d *ConfigDAO) GetDefaultToolConfig() (*models.ToolConfig, error) {
	logging.Info("正在获取默认工具配置")

	var config models.ToolConfig
	err := d.collection.FindOne(
		context.Background(),
		bson.M{"is_default": true},
	).Decode(&config)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Info("未找到默认工具配置，将创建一个新的默认配置")
			defaultConfig := models.GetDefaultConfig()
			defaultConfig.BeforeCreate()

			result, err := d.collection.InsertOne(context.Background(), defaultConfig)
			if err != nil {
				logging.Error("创建默认工具配置失败: %v", err)
				return nil, err
			}

			id, ok := result.InsertedID.(primitive.ObjectID)
			if !ok {
				logging.Error("无法获取创建的默认配置ID")
				return nil, errors.New("无法获取创建的默认配置ID")
			}

			defaultConfig.ID = id
			logging.Info("成功创建默认工具配置")
			return defaultConfig, nil
		}

		logging.Error("获取默认工具配置失败: %v", err)
		return nil, err
	}

	logging.Info("成功获取默认工具配置: %s", config.Name)
	return &config, nil
}

// 根据ID获取工具配置
func (d *ConfigDAO) GetToolConfigByID(id string) (*models.ToolConfig, error) {
	logging.Info("正在获取ID为 %s 的工具配置", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的工具配置ID: %v", err)
		return nil, err
	}

	var config models.ToolConfig
	err = d.collection.FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&config)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Error("未找到ID为 %s 的工具配置", id)
			return nil, err
		}
		logging.Error("获取工具配置失败: %v", err)
		return nil, err
	}

	logging.Info("成功获取ID为 %s 的工具配置", id)
	return &config, nil
}

// 创建工具配置
func (d *ConfigDAO) CreateToolConfig(config *models.ToolConfig) (*models.ToolConfig, error) {
	logging.Info("正在创建工具配置: %s", config.Name)

	config.BeforeCreate()

	// 如果要创建的配置是默认配置，先将其他默认配置设为非默认
	if config.IsDefault {
		if err := d.clearDefaultStatus(); err != nil {
			logging.Error("清除其他默认配置状态失败: %v", err)
			return nil, err
		}
	}

	result, err := d.collection.InsertOne(context.Background(), config)
	if err != nil {
		logging.Error("创建工具配置失败: %v", err)
		return nil, err
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		logging.Error("无法获取创建的配置ID")
		return nil, errors.New("无法获取创建的配置ID")
	}

	config.ID = id
	logging.Info("成功创建工具配置: %s, ID: %s", config.Name, config.ID.Hex())
	return config, nil
}

// 更新工具配置
func (d *ConfigDAO) UpdateToolConfig(config *models.ToolConfig) error {
	logging.Info("正在更新ID为 %s 的工具配置", config.ID.Hex())

	config.BeforeUpdate()

	// 如果要更新的配置是默认配置，先将其他默认配置设为非默认
	if config.IsDefault {
		if err := d.clearDefaultStatus(); err != nil {
			logging.Error("清除其他默认配置状态失败: %v", err)
			return err
		}
	}

	_, err := d.collection.ReplaceOne(
		context.Background(),
		bson.M{"_id": config.ID},
		config,
	)

	if err != nil {
		logging.Error("更新工具配置失败: %v", err)
		return err
	}

	logging.Info("成功更新ID为 %s 的工具配置", config.ID.Hex())
	return nil
}

// 删除工具配置
func (d *ConfigDAO) DeleteToolConfig(id string) error {
	logging.Info("正在删除ID为 %s 的工具配置", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的工具配置ID: %v", err)
		return err
	}

	// 检查是否是默认配置
	var config models.ToolConfig
	err = d.collection.FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&config)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			logging.Error("未找到ID为 %s 的工具配置", id)
			return err
		}
		logging.Error("获取工具配置失败: %v", err)
		return err
	}

	// 不允许删除默认配置
	if config.IsDefault {
		logging.Error("无法删除默认工具配置")
		return errors.New("无法删除默认工具配置")
	}

	_, err = d.collection.DeleteOne(
		context.Background(),
		bson.M{"_id": objID},
	)

	if err != nil {
		logging.Error("删除工具配置失败: %v", err)
		return err
	}

	logging.Info("成功删除ID为 %s 的工具配置", id)
	return nil
}

// 设置配置为默认配置
func (d *ConfigDAO) SetDefaultToolConfig(id string) error {
	logging.Info("正在将ID为 %s 的工具配置设置为默认配置", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的工具配置ID: %v", err)
		return err
	}

	// 先将所有配置设为非默认
	if err := d.clearDefaultStatus(); err != nil {
		logging.Error("清除其他默认配置状态失败: %v", err)
		return err
	}

	// 将指定配置设为默认
	// 修复：使用 time.Now() 代替 config.UpdatedAt
	_, err = d.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$set": bson.M{"is_default": true, "updated_at": primitive.NewDateTimeFromTime(time.Now())}},
	)

	if err != nil {
		logging.Error("设置默认工具配置失败: %v", err)
		return err
	}

	logging.Info("成功将ID为 %s 的工具配置设置为默认配置", id)
	return nil
}

// 辅助方法：清除所有默认配置状态
func (d *ConfigDAO) clearDefaultStatus() error {
	_, err := d.collection.UpdateMany(
		context.Background(),
		bson.M{"is_default": true},
		bson.M{"$set": bson.M{"is_default": false}},
	)
	return err
}
