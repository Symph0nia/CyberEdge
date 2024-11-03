package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"errors"
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

// GetResultsByType 根据类型获取扫描结果列表
func (dao *ResultDAO) GetResultsByType(resultType string) ([]*models.Result, error) {
	logging.Info("正在获取类型为 %s 的扫描结果", resultType)

	var results []*models.Result

	cursor, err := dao.collection.Find(context.Background(), bson.M{"type": resultType})
	if err != nil {
		logging.Error("获取类型为 %s 的扫描结果失败: %v", resultType, err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var result models.Result
		if err := cursor.Decode(&result); err != nil {
			logging.Error("解析扫描结果失败: %v", err)
			return nil, err
		}
		results = append(results, &result)
	}

	if err := cursor.Err(); err != nil {
		logging.Error("游标错误: %v", err)
		return nil, err
	}

	logging.Info("成功获取类型为 %s 的扫描结果，共 %d 个", resultType, len(results))
	return results, nil
}

// UpdateResult 更新指定 ID 的扫描结果
func (dao *ResultDAO) UpdateResult(id string, updatedResult *models.Result) error {
	logging.Info("正在更新扫描结果: %s", id)

	// 将字符串 ID 转换为 ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的结果 ID: %s, 错误: %v", id, err)
		return err
	}

	// 构造 MongoDB 更新操作，确保包含 IsRead 字段
	update := bson.M{
		"$set": bson.M{
			"type":       updatedResult.Type,
			"target":     updatedResult.Target,
			"data":       updatedResult.Data,
			"is_read":    updatedResult.IsRead, // 包含 IsRead 字段
			"updated_at": time.Now(),           // 手动更新 updated_at 字段
		},
	}

	// 执行更新操作
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

// UpdateEntryReadStatus 更新指定任务中的指定条目的已读状态
func (dao *ResultDAO) UpdateEntryReadStatus(resultID string, entryID string, isRead bool) error {
	logging.Info("正在更新任务 %s 中条目 %s 的已读状态", resultID, entryID)

	// 将任务 ID 和条目 ID 转换为 ObjectID
	objID, err := primitive.ObjectIDFromHex(resultID)
	if err != nil {
		logging.Error("无效的任务 ID: %s, 错误: %v", resultID, err)
		return err
	}

	entryObjID, err := primitive.ObjectIDFromHex(entryID)
	if err != nil {
		logging.Error("无效的条目 ID: %s, 错误: %v", entryID, err)
		return err
	}

	// 获取任务
	result, err := dao.GetResultByID(resultID)
	if err != nil {
		logging.Error("无法获取任务: %v", err)
		return err
	}

	// 根据任务类型处理不同的数据结构
	switch result.Type {
	case "Port":
		var portData models.PortData
		if err := unmarshalData(result.Data, &portData); err != nil {
			return err
		}
		for _, port := range portData.Ports {
			if port.ID == entryObjID {
				port.IsRead = isRead
				break
			}
		}
		result.Data = portData

	case "Fingerprint":
		var fingerprints []*models.Fingerprint
		if err := unmarshalData(result.Data, &fingerprints); err != nil {
			return err
		}
		for _, fingerprint := range fingerprints {
			if fingerprint.ID == entryObjID {
				fingerprint.IsRead = isRead
				break
			}
		}
		result.Data = fingerprints

	case "Path":
		var paths []*models.Path
		if err := unmarshalData(result.Data, &paths); err != nil {
			return err
		}
		for _, path := range paths {
			if path.ID == entryObjID {
				path.IsRead = isRead
				break
			}
		}
		result.Data = paths

	default:
		return errors.New("未知的数据类型")
	}

	// 构造 MongoDB 更新操作，更新整个任务
	update := bson.M{
		"$set": bson.M{
			"data":       result.Data,
			"updated_at": time.Now(),
		},
	}

	// 执行更新操作
	updateResult, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)

	if err != nil {
		logging.Error("更新任务 %s 的条目 %s 的已读状态失败: %v", resultID, entryID, err)
		return err
	}

	if updateResult.ModifiedCount == 0 {
		logging.Warn("未找到匹配的任务 %s 进行更新", resultID)
	} else {
		logging.Info("成功更新任务 %s 中条目 %s 的已读状态", resultID, entryID)
	}

	return nil
}

// unmarshalData 是一个辅助函数，用于将 interface{} 类型的数据解析为指定的结构体
func unmarshalData(data interface{}, target interface{}) error {
	bsonData, err := bson.Marshal(data)
	if err != nil {
		logging.Error("序列化数据失败: %v", err)
		return errors.New("序列化数据失败")
	}

	if err := bson.Unmarshal(bsonData, target); err != nil {
		logging.Error("解析数据失败: %v", err)
		return errors.New("解析数据失败")
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
