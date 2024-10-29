// task_types.go

package tasks

// TaskType 定义了任务类型
type TaskType string

const (
	// TaskTypePing 定义 Ping 任务类型
	TaskTypePing  TaskType = "ping"
	TaskTypeHttpx TaskType = "httpx"
	// 在这里添加其他任务类型
	// TaskTypeExample TaskType = "example"
)
