// pkg/logging/logging.go

package logging

import (
	"log"
	"os"
)

// 日志记录器
var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
)

func init() {
	// 创建日志文件并初始化日志记录器
	if err := initializeLoggers("cyberedge.log"); err != nil {
		log.Fatalf("无法初始化日志记录器: %v", err)
	}
}

// initializeLoggers 创建日志文件并初始化信息和错误日志记录器
func initializeLoggers(logFilePath string) error {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err // 返回错误以供上层处理
	}

	InfoLogger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(logFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return nil
}

// LogInfo 记录信息日志
func LogInfo(message string) {
	InfoLogger.Println(message)
}

// LogError 记录错误日志
func LogError(err error) {
	ErrorLogger.Println(err)
}
