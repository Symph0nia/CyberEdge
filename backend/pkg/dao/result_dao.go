package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"cyberedge/pkg/utils"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type ResultDAO struct {
	collection *mongo.Collection
}

// getUpdatePath 获取不同类型的更新路径
func getUpdatePath(taskType string) string {
	switch taskType {
	case "Subdomain":
		return "data.subdomains"
	case "Port":
		return "data.ports"
	case "Path":
		return "data.paths"
	default:
		return ""
	}
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

func (dao *ResultDAO) UpdateEntryReadStatus(resultID string, entryID string, isRead bool) error {
	logging.Info("正在更新任务 %s 中条目 %s 的已读状态", resultID, entryID)

	// 转换 ObjectID
	objID, err := primitive.ObjectIDFromHex(resultID)
	if err != nil {
		return fmt.Errorf("无效的任务 ID: %v", err)
	}

	entryObjID, err := primitive.ObjectIDFromHex(entryID)
	if err != nil {
		return fmt.Errorf("无效的条目 ID: %v", err)
	}

	// 获取任务以确定类型
	result, err := dao.GetResultByID(resultID)
	if err != nil {
		return fmt.Errorf("无法获取任务: %v", err)
	}

	// 根据任务类型获取更新路径
	updatePath := getUpdatePath(result.Type)
	if updatePath == "" {
		return fmt.Errorf("不支持的任务类型: %s", result.Type)
	}

	// 执行更新
	update := bson.M{
		"$set": bson.M{
			fmt.Sprintf("%s.$[elem].is_read", updatePath): isRead,
			"updated_at": time.Now(),
		},
	}

	arrayFilters := []interface{}{
		bson.M{"elem._id": entryObjID},
	}

	updateResult, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilters,
		}),
	)

	if err != nil {
		return fmt.Errorf("更新失败: %v", err)
	}

	if updateResult.ModifiedCount == 0 {
		return fmt.Errorf("未找到匹配的记录")
	}

	logging.Info("成功更新已读状态")
	return nil
}

// UpdateSubdomainIP 更新指定任务中的子域名解析的 IP 地址
func (dao *ResultDAO) UpdateSubdomainIP(resultID string, entryID string, ip string) error {
	logging.Info("正在更新任务 %s 中子域名 %s 的 IP 地址", resultID, entryID)

	// 将任务 ID 和子域名 ID 转换为 ObjectID
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

	// 检查任务类型为 Subdomain
	if result.Type != "Subdomain" {
		return errors.New("任务类型不匹配，无法更新子域名 IP")
	}

	// 解析 result.Data 为 SubdomainData 结构
	var subdomainData models.SubdomainData
	if err := utils.UnmarshalData(result.Data, &subdomainData); err != nil {
		logging.Error("解析子域名数据失败: %v", err)
		return err
	}

	// 遍历子域名数据，找到匹配的子域名并更新 IP 地址
	for i, subdomain := range subdomainData.Subdomains {
		if subdomain.ID == entryObjID {
			subdomainData.Subdomains[i].IP = ip // 更新 IP 地址
			break
		}
	}

	// 将更新后的数据赋值回 result.Data
	result.Data = subdomainData

	// 构造 MongoDB 更新操作，更新整个任务
	update := bson.M{
		"$set": bson.M{
			"data":       result.Data,
			"updated_at": time.Now(), // 更新更新时间
		},
	}

	// 执行更新操作
	updateResult, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)

	if err != nil {
		logging.Error("更新任务 %s 的子域名 %s 的 IP 地址失败: %v", resultID, entryID, err)
		return err
	}

	if updateResult.ModifiedCount == 0 {
		logging.Warn("未找到匹配的任务 %s 进行更新", resultID)
	} else {
		logging.Info("成功更新任务 %s 中子域名 %s 的 IP 地址", resultID, entryID)
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

// UpdateHTTPInfo 通用的 HTTP 信息更新函数
func (dao *ResultDAO) UpdateHTTPInfo(resultID string, entryID string, taskType string, statusCode int, title string) error {
	logging.Info("正在更新任务 %s 中 %s 类型条目 %s 的HTTP信息", resultID, taskType, entryID)

	// 转换 ObjectID
	objID, err := primitive.ObjectIDFromHex(resultID)
	if err != nil {
		return fmt.Errorf("无效的任务 ID: %v", err)
	}

	entryObjID, err := primitive.ObjectIDFromHex(entryID)
	if err != nil {
		return fmt.Errorf("无效的条目 ID: %v", err)
	}

	// 构造更新操作
	updatePath := getUpdatePath(taskType)
	if updatePath == "" {
		return fmt.Errorf("不支持的任务类型: %s", taskType)
	}

	// 执行更新
	update := bson.M{
		"$set": bson.M{
			fmt.Sprintf("%s.$[elem].http_status", updatePath): statusCode,
			fmt.Sprintf("%s.$[elem].http_title", updatePath):  title,
			"updated_at": time.Now(),
		},
	}

	arrayFilters := []interface{}{
		bson.M{"elem._id": entryObjID},
	}

	result, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID, "type": taskType},
		update,
		options.Update().SetArrayFilters(options.ArrayFilters{
			Filters: arrayFilters,
		}),
	)

	if err != nil {
		return fmt.Errorf("更新失败: %v", err)
	}

	if result.ModifiedCount == 0 {
		return fmt.Errorf("未找到匹配的记录")
	}

	logging.Info("成功更新HTTP信息")
	return nil
}
