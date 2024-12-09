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

type TaskDAO struct {
	collection *mongo.Collection
}

// NewTaskDAO 创建一个新的 TaskDAO 实例
func NewTaskDAO(collection *mongo.Collection) *TaskDAO {
	return &TaskDAO{collection: collection}
}

// CreateTask 创建新任务
func (dao *TaskDAO) CreateTask(task *models.Task) error {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	logging.Info("正在创建新任务: %s", task.Type)

	_, err := dao.collection.InsertOne(context.Background(), task)
	if err != nil {
		logging.Error("创建任务失败: %s, 错误: %v", task.Type, err)
		return err
	}

	logging.Info("任务创建成功: %s", task.Type)
	return nil
}

// GetTaskByID 根据任务 ID 获取任务
func (dao *TaskDAO) GetTaskByID(id string) (*models.Task, error) {
	logging.Info("正在获取任务信息: %s", id)

	var task models.Task
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的任务 ID: %s, 错误: %v", id, err)
		return nil, err
	}

	err = dao.collection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&task)
	if err != nil {
		logging.Error("获取任务失败: %s, 错误: %v", id, err)
		return nil, err
	}

	logging.Info("成功获取任务信息: %s", id)
	return &task, nil
}

// GetTasksByIDs 批量获取任务信息
func (dao *TaskDAO) GetTasksByIDs(ids []string) ([]*models.Task, error) {
	var objectIDs []primitive.ObjectID
	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	var tasks []*models.Task
	cursor, err := dao.collection.Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": objectIDs}},
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &tasks); err != nil {
		return nil, err
	}

	return tasks, nil
}

// GetAllTasks 获取所有任务
func (dao *TaskDAO) GetAllTasks() ([]models.Task, error) {
	logging.Info("正在获取所有任务")

	cursor, err := dao.collection.Find(context.Background(), bson.M{})
	if err != nil {
		logging.Error("获取所有任务失败: %v", err)
		return nil, err
	}

	defer func() {
		if err := cursor.Close(context.Background()); err != nil {
			logging.Error("关闭游标失败: %v", err)
		}
	}()

	var tasks []models.Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		logging.Error("解析任务数据失败: %v", err) // 更正日志信息
		return nil, err
	}

	logging.Info("成功获取所有任务，共 %d 个", len(tasks))
	return tasks, nil
}

// UpdateTaskStatus 更新任务状态并记录结果
func (dao *TaskDAO) UpdateTaskStatus(id string, status models.TaskStatus, result string) error {
	logging.Info("正在更新任务状态: %s 到 %s", id, status)

	// 将字符串 ID 转换为 ObjectID
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的任务 ID: %s, 错误: %v", id, err)
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// 如果任务完成或失败，记录完成时间和结果
	if status == models.TaskStatusCompleted || status == models.TaskStatusFailed {
		update["$set"].(bson.M)["completed_at"] = time.Now()
		update["$set"].(bson.M)["result"] = result
	}

	updateResult, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID}, // 使用 ObjectID 进行查询
		update,
	)

	if err != nil {
		logging.Error("更新任务状态失败: %s, 错误: %v", id, err)
		return err
	}

	if updateResult.ModifiedCount == 0 {
		logging.Warn("未找到匹配的任务进行更新: %s", id)
	} else {
		logging.Info("成功更新任务状态: %s 到 %s", id, status)
	}

	return nil
}

type DeleteTasksResult struct {
	DeletedIDs []string
	FailedIDs  map[string]string
}

// DeleteTasks 批量删除任务
func (dao *TaskDAO) DeleteTasks(ids []string) (*DeleteTasksResult, error) {
	logging.Info("正在批量删除任务: %v", ids)

	result := &DeleteTasksResult{
		DeletedIDs: make([]string, 0),
		FailedIDs:  make(map[string]string),
	}

	// 转换所有有效的ObjectID
	var objectIDs []primitive.ObjectID
	invalidIDs := make(map[string]bool)

	for _, id := range ids {
		objID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			result.FailedIDs[id] = "无效的任务ID"
			invalidIDs[id] = true
			continue
		}
		objectIDs = append(objectIDs, objID)
	}

	if len(objectIDs) == 0 {
		return result, nil
	}

	// 执行批量删除
	_, err := dao.collection.DeleteMany(
		context.Background(),
		bson.M{"_id": bson.M{"$in": objectIDs}},
	)

	if err != nil {
		logging.Error("批量删除任务失败: %v", err)
		return nil, err
	}

	// 查询剩余的任务ID（未被删除的）
	cursor, err := dao.collection.Find(
		context.Background(),
		bson.M{"_id": bson.M{"$in": objectIDs}},
	)
	if err != nil {
		logging.Error("查询剩余任务失败: %v", err)
		return nil, err
	}
	defer cursor.Close(context.Background())

	var remainingTasks []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := cursor.All(context.Background(), &remainingTasks); err != nil {
		return nil, err
	}

	// 标记删除失败的任务
	remainingIDMap := make(map[string]bool)
	for _, task := range remainingTasks {
		remainingIDMap[task.ID.Hex()] = true
	}

	// 计算成功删除的任务
	for _, id := range ids {
		if invalidIDs[id] {
			continue
		}
		if remainingIDMap[id] {
			result.FailedIDs[id] = "删除失败"
		} else {
			result.DeletedIDs = append(result.DeletedIDs, id)
		}
	}

	logging.Info("批量删除任务完成，成功删除: %d, 删除失败: %d",
		len(result.DeletedIDs), len(result.FailedIDs))

	return result, nil
}

// UpdateTaskResult 更新指定 ID 的任务结果
func (dao *TaskDAO) UpdateTaskResult(id string, result string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"result":     result,
			"updated_at": time.Now(),
		},
	}

	_, err = dao.collection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		update,
	)

	return err
}
