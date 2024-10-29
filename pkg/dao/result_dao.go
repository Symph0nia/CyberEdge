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

type ResultDAO struct {
	collection *mongo.Collection
}

// NewResultDAO 创建一个新的 ResultDAO 实例
func NewResultDAO(collection *mongo.Collection) *ResultDAO {
	return &ResultDAO{collection: collection}
}

// CreateResult 创建新的扫描结果
func (dao *ResultDAO) CreateResult(result *models.Result) error {
	logging.Info("正在创建新的扫描结果")

	_, err := dao.collection.InsertOne(context.Background(), result)
	if err != nil {
		logging.Error("创建扫描结果失败: %v", err)
		return err
	}

	logging.Info("扫描结果创建成功")
	return nil
}

// GetResultByID 根据 ID 获取扫描结果
func (dao *ResultDAO) GetResultByID(id string) (*models.Result, error) {
	logging.Info("正在获取扫描结果: %s", id)

	var result models.Result
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的结果 ID: %s, 错误: %v", id, err)
		return nil, err
	}

	err = dao.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&result)
	if err != nil {
		logging.Error("获取扫描结果失败: %s, 错误: %v", id, err)
		return nil, err
	}

	logging.Info("成功获取扫描结果: %s", id)
	return &result, nil
}

// UpdateResult 更新指定 ID 的扫描结果
func (dao *ResultDAO) UpdateResult(id string, updatedResult *models.Result) error {
	logging.Info("正在更新扫描结果: %s", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的结果 ID: %s, 错误: %v", id, err)
		return err
	}

	update := bson.M{
		"$set": updatedResult,
	}

	update["$set"].(bson.M)["updated_at"] = time.Now()

	updateResult, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)

	if err != nil {
		logging.Error("更新扫描结果失败: %s, 错误: %v", id, err)
		return err
	}

	if updateResult.ModifiedCount == 0 {
		logging.Warn("未找到匹配的扫描结果进行更新: %s", id)
	} else {
		logging.Info("成功更新扫描结果: %s", id)
	}

	return nil
}

// DeleteResult 删除指定 ID 的扫描结果
func (dao *ResultDAO) DeleteResult(id string) error {
	logging.Info("正在删除扫描结果: %s", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的结果 ID: %s, 错误: %v", id, err)
		return err
	}

	result, err := dao.collection.DeleteOne(context.Background(), bson.M{"_id": objID})

	if result.DeletedCount == 0 {
		logging.Warn("未找到要删除的扫描结果: %s", id)
		return mongo.ErrNoDocuments
	}

	if err != nil {
		logging.Error("删除扫描结果失败: %s, 错误: %v", id, err)
		return err
	}

	logging.Info("成功删除扫描结果: %s", id)
	return nil
}
