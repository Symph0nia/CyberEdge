package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type TargetDAO struct {
	collection *mongo.Collection
}

// NewTargetDAO 创建一个新的 TargetDAO 实例
func NewTargetDAO(collection *mongo.Collection) *TargetDAO {
	return &TargetDAO{collection: collection}
}

// CreateTarget 创建新的目标
func (dao *TargetDAO) CreateTarget(target *models.Target) error {
	logging.Info("正在创建新的目标")

	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()

	_, err := dao.collection.InsertOne(context.Background(), target)
	if err != nil {
		logging.Error("创建目标失败: %v", err)
		return err
	}

	logging.Info("目标创建成功")
	return nil
}

// GetTargetByID 根据 ID 获取目标
func (dao *TargetDAO) GetTargetByID(id string) (*models.Target, error) {
	logging.Info("正在获取目标: %s", id)

	var target models.Target
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的目标 ID: %s, 错误: %v", id, err)
		return nil, err
	}

	err = dao.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&target)
	if err != nil {
		logging.Error("获取目标失败: %s, 错误: %v", id, err)
		return nil, err
	}

	logging.Info("成功获取目标: %s", id)
	return &target, nil
}

// GetAllTargets 获取所有目标
func (dao *TargetDAO) GetAllTargets() ([]*models.Target, error) {
	logging.Info("正在获取所有目标")

	var targets []*models.Target
	cursor, err := dao.collection.Find(context.Background(), bson.M{})
	if err != nil {
		logging.Error("获取目标列表失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var target models.Target
		if err := cursor.Decode(&target); err != nil {
			logging.Error("解析目标失败: %v", err)
			return nil, err
		}
		targets = append(targets, &target)
	}

	if err := cursor.Err(); err != nil {
		logging.Error("游标错误: %v", err)
		return nil, err
	}

	logging.Info("成功获取目标列表，共 %d 个", len(targets))
	return targets, nil
}

// UpdateTarget 更新目标信息
func (dao *TargetDAO) UpdateTarget(id string, updatedTarget *models.Target) error {
	logging.Info("正在更新目标: %s", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的目标 ID: %s, 错误: %v", id, err)
		return err
	}

	updatedTarget.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":        updatedTarget.Name,
			"description": updatedTarget.Description,
			"type":        updatedTarget.Type,
			"target":      updatedTarget.Target,
			"status":      updatedTarget.Status, // 添加 status 字段
			"updated_at":  updatedTarget.UpdatedAt,
		},
	}

	result, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)

	if err != nil {
		logging.Error("更新目标失败: %s, 错误: %v", id, err)
		return err
	}

	if result.ModifiedCount == 0 {
		logging.Warn("未找到匹配的目标进行更新: %s", id)
	} else {
		logging.Info("成功更新目标: %s", id)
	}

	return nil
}

// DeleteTarget 删除目标
func (dao *TargetDAO) DeleteTarget(id string) error {
	logging.Info("正在删除目标: %s", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的目标 ID: %s, 错误: %v", id, err)
		return err
	}

	result, err := dao.collection.DeleteOne(context.Background(), bson.M{"_id": objID})

	if err != nil {
		logging.Error("删除目标失败: %s, 错误: %v", id, err)
		return err
	}

	if result.DeletedCount == 0 {
		logging.Warn("未找到要删除的目标: %s", id)
		return mongo.ErrNoDocuments
	}

	logging.Info("成功删除目标: %s", id)
	return nil
}
