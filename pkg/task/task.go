package task

import (
	"context"
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"

	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Task 定义任务结构
type Task struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // 任务类型
	Description string    `json:"description"` // 任务描述
	Status      string    `json:"status"`      // 任务状态
	Interval    int       `json:"interval"`    // 运行间隔（分钟）
	RunCount    int       `json:"run_count"`   // 运行次数
	CreatedAt   time.Time `json:"created_at"`  // 创建时间
}

// Scheduler 任务调度器
type Scheduler struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	queue          string
	mongoClient    *mongo.Client
	taskCollection *mongo.Collection
}

// NewScheduler 创建新的任务调度器
func NewScheduler(amqpURI, queueName string, mongoClient *mongo.Client, dbName string) (*Scheduler, error) {
	conn, err := amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("无法连接到RabbitMQ: %v", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("无法创建RabbitMQ通道: %v", err)
	}

	// 声明队列
	_, err = channel.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("无法声明队列: %v", err)
	}

	taskCollection := mongoClient.Database(dbName).Collection("tasks")

	return &Scheduler{
		conn:           conn,
		channel:        channel,
		queue:          queueName,
		mongoClient:    mongoClient,
		taskCollection: taskCollection,
	}, nil
}

// ScheduleTask 将任务发送到RabbitMQ队列并存储到MongoDB
func (s *Scheduler) ScheduleTask(task Task) error {
	task.CreatedAt = time.Now() // 设置创建时间

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("无法序列化任务: %v", err)
	}

	err = s.channel.Publish(
		"",
		s.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			MessageId:   task.ID,
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("无法发布消息到队列: %v", err)
	}

	task.Status = "waiting" // 设置状态为等待

	filter := bson.M{"id": task.ID}
	update := bson.M{"$set": task}
	opts := options.Update().SetUpsert(true) // 使用 Upsert 选项

	if _, err := s.taskCollection.UpdateOne(context.Background(), filter, update, opts); err != nil {
		return fmt.Errorf("无法插入或更新任务到MongoDB: %v", err)
	}

	log.Printf("已调度任务: %s", task.ID)
	return nil
}

// GetAllTasks 获取所有任务
func (s *Scheduler) GetAllTasks() ([]Task, error) {
	var tasks []Task
	cursor, err := s.taskCollection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("无法获取所有任务: %v", err)
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var task Task
		if err := cursor.Decode(&task); err != nil {
			return nil, fmt.Errorf("解码任务失败: %v", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTask 获取单个任务
func (s *Scheduler) GetTask(id string) (Task, error) {
	var task Task
	err := s.taskCollection.FindOne(context.Background(), bson.M{"id": id}).Decode(&task)
	if err != nil {
		return Task{}, fmt.Errorf("未找到任务: %v", err)
	}
	return task, nil
}

// StartTask 开始执行单个任务（通过RabbitMQ）
func (s *Scheduler) StartTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	task.Status = "running" // 更新状态为运行中

	if _, err := s.taskCollection.UpdateOne(context.Background(), bson.M{"id": id}, bson.M{"$set": bson.M{"status": "running"}}); err != nil {
		return fmt.Errorf("更新MongoDB中的任务状态失败: %v", err)
	}

	err = s.ScheduleTask(task) // 重新调度该任务
	if err != nil {
		return fmt.Errorf("重新调度任务失败: %v", err)
	}

	log.Printf("已开始执行任务: %s", id)
	return nil
}

// StopTask 停止单个任务（这里可以根据需求实现）
func (s *Scheduler) StopTask(id string) error {
	task, err := s.GetTask(id)
	if err != nil {
		return err
	}

	task.Status = "stopped" // 更新状态为已停止

	if _, err := s.taskCollection.UpdateOne(context.Background(), bson.M{"id": id}, bson.M{"$set": bson.M{"status": "stopped"}}); err != nil {
		return fmt.Errorf("更新MongoDB中的任务状态失败: %v", err)
	}

	log.Printf("已停止执行任务: %s", id)
	return nil
}

// DeleteTask 删除单个任务
func (s *Scheduler) DeleteTask(id string) error {
	_, err := s.taskCollection.DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return fmt.Errorf("删除MongoDB中的任务失败: %v", err)
	}

	log.Printf("已删除任务: %s", id)
	return nil
}

// Close 关闭调度器连接和通道
func (s *Scheduler) Close() {
	if err := s.channel.Close(); err != nil {
		log.Printf("关闭RabbitMQ通道失败: %v", err)
	}
	if err := s.conn.Close(); err != nil {
		log.Printf("关闭RabbitMQ连接失败: %v", err)
	}
}

// StartTaskProcessor 启动任务处理器，监听 RabbitMQ 队列并处理消息
func (s *Scheduler) StartTaskProcessor() {
	msgs, err := s.channel.Consume(
		s.queue,
		"",    // 消费者名称（空字符串表示自动生成）
		true,  // 自动确认消息已被消费
		false, // 不独占消费者通道
		false, // 不等待其他消费者连接
		false, // 不阻塞消费者通道
		nil,   // 可选参数，通常为nil
	)
	if err != nil {
		log.Fatalf("无法启动消费者: %v", err)
	}

	for msg := range msgs {
		var task Task
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			log.Printf("解码任务失败: %v", err)
			continue // 跳过此消息，继续处理下一个消息
		}

		switch task.Type {
		case "ping":
			if err := s.ProcessPingTask(task); err != nil {
				log.Printf("处理 Ping 任务失败: %v", err)
			}
		default:
			log.Printf("未知任务类型: %s", task.Type)
		}
	}
}
