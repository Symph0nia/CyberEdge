package task

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"

	"cyberedge/pkg/models" // 导入模型包
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Task 代表一个可调度的任务
type Task struct {
	ID          primitive.ObjectID // 使用 ObjectID 类型
	TaskID      string             // 任务的字符串 ID
	Description string             // 描述
	Interval    time.Duration      // 执行间隔
	IsRunning   bool               // 是否正在运行
	StopCh      chan bool          // 停止信号通道
}

// NewTask 创建一个新的任务
func NewTask(id primitive.ObjectID, description string, interval time.Duration) *Task {
	return &Task{
		ID:          id,
		TaskID:      id.Hex(), // 将 ObjectID 转换为字符串并赋值给 TaskID
		Description: description,
		Interval:    interval,
		IsRunning:   false,
		StopCh:      make(chan bool),
	}
}

// Start 启动任务
func (t *Task) Start() {
	if t.IsRunning {
		return // 如果已经在运行，则不再启动
	}

	t.IsRunning = true

	go func() {
		for {
			select {
			case <-t.StopCh:
				fmt.Printf("Task %s stopped\n", t.ID.Hex())
				t.IsRunning = false
				return
			default:
				t.execute()
				time.Sleep(t.Interval)
			}
		}
	}()
}

// Stop 停止任务
func (t *Task) Stop() {
	t.StopCh <- true
}

// execute 执行任务的具体逻辑
func (t *Task) execute() {
	fmt.Printf("Executing task %s: %s\n", t.ID.Hex(), t.Description)
}

// TaskManager 管理所有任务的结构体
type TaskManager struct {
	tasks           map[string]*Task
	mongoCollection *mongo.Collection // MongoDB集合用于存储任务信息
}

// NewTaskManager 创建一个新的任务管理器
func NewTaskManager(collection *mongo.Collection) *TaskManager {
	return &TaskManager{
		tasks:           make(map[string]*Task),
		mongoCollection: collection,
	}
}

// AddTask 添加新任务到管理器中并保存到MongoDB
func (tm *TaskManager) AddTask(task *Task) error {
	taskModel := models.Task{
		ID:          task.ID, // 使用 ObjectID 类型的 ID
		TaskID:      task.TaskID,
		Description: task.Description,
		Interval:    task.Interval,
		IsRunning:   task.IsRunning,
		CreatedAt:   time.Now(),
	}

	tm.tasks[task.TaskID] = task

	if _, err := tm.mongoCollection.InsertOne(context.Background(), taskModel); err != nil {
		return err
	}

	return nil
}

// GetAllTasks 获取所有任务的状态信息并从MongoDB加载
func (tm *TaskManager) GetAllTasks() ([]models.Task, error) {
	var tasks []models.Task

	cursor, err := tm.mongoCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var task models.Task
		if err := cursor.Decode(&task); err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTask 获取单个任务的状态信息并从MongoDB加载，并返回 Task 类型以便调用 Start 和 Stop 方法。
func (tm *TaskManager) GetTask(id string) (*Task, error) {
	var model models.Task

	err := tm.mongoCollection.FindOne(context.Background(), bson.M{"task_id": id}).Decode(&model)
	if err != nil {
		return nil, err // 如果没有找到返回错误
	}

	task := NewTask(model.ID, model.Description, model.Interval)
	task.IsRunning = model.IsRunning // 设置当前运行状态

	return task, nil
}
