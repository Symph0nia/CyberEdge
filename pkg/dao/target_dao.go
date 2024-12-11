package dao

import (
	"context"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type TargetDAO struct {
	collection        *mongo.Collection
	resultsCollection *mongo.Collection // 添加 results 集合
}

func NewTargetDAO(db *mongo.Database) *TargetDAO {
	return &TargetDAO{
		collection:        db.Collection("targets"),
		resultsCollection: db.Collection("results"),
	}
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

// GetTargetDetailsById 获取目标详情，包含实时统计数据
func (dao *TargetDAO) GetTargetDetailsById(id string) (*models.TargetDetails, error) {
	logging.Info("正在获取目标详情: %s", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的目标 ID: %s, 错误: %v", id, err)
		return nil, err
	}

	// 获取基本目标信息
	var target models.Target
	err = dao.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&target)
	if err != nil {
		logging.Error("获取目标失败: %s, 错误: %v", id, err)
		return nil, err
	}

	// 从 results 集合中获取统计数据
	subdomainCount, err := dao.getSubdomainCount(objID)
	if err != nil {
		logging.Error("获取子域名数量失败: %v", err)
		subdomainCount = 0
	}

	portCount, err := dao.getPortCount(objID)
	if err != nil {
		logging.Error("获取端口数量失败: %v", err)
		portCount = 0
	}

	pathCount, err := dao.getPathCount(objID)
	if err != nil {
		logging.Error("获取路径数量失败: %v", err)
		pathCount = 0
	}

	vulnCount, err := dao.getVulnerabilityCount(objID)
	if err != nil {
		logging.Error("获取漏洞数量失败: %v", err)
		vulnCount = 0
	}

	details := &models.TargetDetails{
		Target: &target,
		Stats: models.TargetStats{
			SubdomainCount:     subdomainCount,
			PortCount:          portCount,
			PathCount:          pathCount,
			VulnerabilityCount: vulnCount,
		},
	}

	logging.Info("成功获取目标详情: %s", id)
	return details, nil
}

// 子域名数量统计
func (dao *TargetDAO) getSubdomainCount(targetID primitive.ObjectID) (int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id": targetID,
				"type":      "Subdomain",
			},
		},
		{
			"$unwind": "$data.subdomains",
		},
		{
			"$count": "count",
		},
	}

	var result []bson.M
	cursor, err := dao.resultsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	count := result[0]["count"]
	switch v := count.(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected count type: %T", count)
	}
}

// 端口数量统计
func (dao *TargetDAO) getPortCount(targetID primitive.ObjectID) (int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id": targetID,
				"type":      "Port",
			},
		},
		{
			"$unwind": "$data.ports",
		},
		{
			"$count": "count",
		},
	}

	var result []bson.M
	cursor, err := dao.resultsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	// 处理不同的整数类型
	count := result[0]["count"]
	switch v := count.(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected count type: %T", count)
	}
}

// 路径数量统计
func (dao *TargetDAO) getPathCount(targetID primitive.ObjectID) (int, error) {
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id": targetID,
				"type":      "Path",
			},
		},
		{
			"$unwind": "$data.paths",
		},
		{
			"$count": "count",
		},
	}

	var result []bson.M
	cursor, err := dao.resultsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	count := result[0]["count"]
	switch v := count.(type) {
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unexpected count type: %T", count)
	}
}

// 漏洞数量统计（预留）
func (dao *TargetDAO) getVulnerabilityCount(targetID primitive.ObjectID) (int, error) {
	// TODO: 实现漏洞统计逻辑
	return 0, nil
}
