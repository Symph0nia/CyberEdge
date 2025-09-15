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

	// 获取端口排名数据
	topPorts, err := dao.GetTopPortStats(id)
	if err != nil {
		logging.Error("获取端口排名失败: %v", err)
		topPorts = []models.PortStat{}
	}

	// 获取HTTP状态码统计
	httpStats, err := dao.GetHTTPStatusStats(id)
	if err != nil {
		logging.Error("获取HTTP状态码统计失败: %v", err)
		httpStats = []models.HTTPStatusStat{}
	}

	details := &models.TargetDetails{
		Target: &target,
		Stats: models.TargetStats{
			SubdomainCount:     subdomainCount,
			PortCount:          portCount,
			PathCount:          pathCount,
			VulnerabilityCount: vulnCount,
			TopPorts:           topPorts,  // 添加端口排名
			HTTPStatusStats:    httpStats, // 添加HTTP状态码统计
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

// GetTopPortStats 获取目标的前10个最常见端口及其数量
func (dao *TargetDAO) GetTopPortStats(targetID string) ([]models.PortStat, error) {
	logging.Info("正在获取目标 %s 的端口统计信息", targetID)

	objID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		logging.Error("无效的目标 ID: %s, 错误: %v", targetID, err)
		return nil, err
	}

	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id": objID,
				"type":      "Port",
			},
		},
		{
			"$unwind": "$data.ports",
		},
		{
			"$group": bson.M{
				"_id":   "$data.ports.number",
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": 10,
		},
	}

	cursor, err := dao.resultsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		logging.Error("聚合端口统计失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stats []models.PortStat
	for cursor.Next(context.Background()) {
		var result struct {
			ID    int `bson:"_id"`
			Count int `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			logging.Error("解析端口统计结果失败: %v", err)
			return nil, err
		}
		stats = append(stats, models.PortStat{
			Port:  result.ID,
			Count: result.Count,
		})
	}

	if err := cursor.Err(); err != nil {
		logging.Error("游标错误: %v", err)
		return nil, err
	}

	logging.Info("成功获取目标 %s 的端口统计信息，共 %d 条记录", targetID, len(stats))
	return stats, nil
}

// GetHTTPStatusStats 获取目标所有HTTP状态码的统计信息
func (dao *TargetDAO) GetHTTPStatusStats(targetID string) ([]models.HTTPStatusStat, error) {
	logging.Info("正在获取目标 %s 的HTTP状态码统计信息", targetID)

	objID, err := primitive.ObjectIDFromHex(targetID)
	if err != nil {
		logging.Error("无效的目标 ID: %s, 错误: %v", targetID, err)
		return nil, err
	}

	// 构建聚合管道
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"target_id": objID,
				"$or": []bson.M{
					{"type": "Subdomain"},
					{"type": "Port"},
					{"type": "Path"},
				},
			},
		},
		{
			"$facet": bson.M{
				"subdomains": []bson.M{
					{"$match": bson.M{"type": "Subdomain"}},
					{"$unwind": "$data.subdomains"},
					{"$group": bson.M{
						"_id":   "$data.subdomains.http_status",
						"count": bson.M{"$sum": 1},
					}},
				},
				"ports": []bson.M{
					{"$match": bson.M{"type": "Port"}},
					{"$unwind": "$data.ports"},
					{"$group": bson.M{
						"_id":   "$data.ports.http_status",
						"count": bson.M{"$sum": 1},
					}},
				},
				"paths": []bson.M{
					{"$match": bson.M{"type": "Path"}},
					{"$unwind": "$data.paths"},
					{"$group": bson.M{
						"_id":   "$data.paths.http_status",
						"count": bson.M{"$sum": 1},
					}},
				},
			},
		},
		{
			"$project": bson.M{
				"all_stats": bson.M{
					"$concatArrays": []string{"$subdomains", "$ports", "$paths"},
				},
			},
		},
		{
			"$unwind": "$all_stats",
		},
		{
			"$group": bson.M{
				"_id":   "$all_stats._id",
				"count": bson.M{"$sum": "$all_stats.count"},
			},
		},
		{
			"$match": bson.M{
				"_id": bson.M{"$ne": nil},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
	}

	cursor, err := dao.resultsCollection.Aggregate(context.Background(), pipeline)
	if err != nil {
		logging.Error("聚合HTTP状态码统计失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var stats []models.HTTPStatusStat
	for cursor.Next(context.Background()) {
		var result struct {
			ID    int `bson:"_id"`
			Count int `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			logging.Error("解析HTTP状态码统计结果失败: %v", err)
			return nil, err
		}

		// 生成状态码描述标签
		label := getHTTPStatusLabel(result.ID)

		stats = append(stats, models.HTTPStatusStat{
			Status: result.ID,
			Count:  result.Count,
			Label:  label,
		})
	}

	if err := cursor.Err(); err != nil {
		logging.Error("游标错误: %v", err)
		return nil, err
	}

	logging.Info("成功获取目标 %s 的HTTP状态码统计信息，共 %d 条记录", targetID, len(stats))
	return stats, nil
}

// getHTTPStatusLabel 根据状态码生成描述标签
func getHTTPStatusLabel(status int) string {
	switch {
	case status >= 500:
		return fmt.Sprintf("%d 服务器错误", status)
	case status >= 400:
		return fmt.Sprintf("%d 客户端错误", status)
	case status >= 300:
		return fmt.Sprintf("%d 重定向", status)
	case status >= 200:
		return fmt.Sprintf("%d 成功", status)
	case status >= 100:
		return fmt.Sprintf("%d 信息", status)
	default:
		return fmt.Sprintf("%d 未知状态", status)
	}
}
