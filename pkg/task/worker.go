// CyberEdge/pkg/task/worker.go

package task

import (
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"encoding/json"
)

// StartTaskProcessor 启动任务处理器，监听 RabbitMQ 队列并处理消息
func StartTaskProcessor(s *models.Scheduler) {
	msgs, err := s.AMQPChannel.Consume(
		s.QueueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logging.Fatal("无法启动RabbitMQ消费者: %v", err)
	}

	logging.Info("任务处理器已启动，正在监听队列: %s", s.QueueName)

	for msg := range msgs {
		var task models.Task
		if err := json.Unmarshal(msg.Body, &task); err != nil {
			logging.Error("解码任务失败: %v", err)
			continue
		}

		logging.Info("收到新任务，任务ID: %s，类型: %s", task.ID, task.Type)

		// 根据任务类型处理任务
		switch task.Type {
		case models.TaskTypePing:
			if err := ProcessPingTask(s, task); err != nil {
				logging.Error("处理Ping任务失败，任务ID: %s，错误: %v", task.ID, err)
			}
		default:
			logging.Warn("未知任务类型，任务ID: %s，类型: %s", task.ID, task.Type)
		}
	}
}
