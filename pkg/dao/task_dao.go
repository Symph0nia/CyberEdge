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

	_, err := dao.collection.UpdateOne(
		context.Background(),
		bson.M{"id": id}, // 使用字符串 ID 进行查询
		update,
	)

	if err != nil {
		logging.Error("更新任务状态失败: %s, 错误: %v", id, err)
		return err
	}

	logging.Info("成功更新任务状态: %s 到 %s", id, status)
	return nil
}

// DeleteTask 删除任务
func (dao *TaskDAO) DeleteTask(id string) error {
	logging.Info("正在删除任务: %s", id)

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		logging.Error("无效的任务 ID: %s, 错误: %v", id, err)
		return err
	}

	result, err := dao.collection.DeleteOne(context.Background(), bson.M{"_id": objID})

	if result.DeletedCount == 0 {
		logging.Warn("未找到要删除的任务: %s", id)
		return mongo.ErrNoDocuments
	}

	if err != nil {
		logging.Error("删除任务失败: %s, 错误: %v", id, err)
		return err
	}

	logging.Info("成功删除任务: %s", id)
	return nil
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
