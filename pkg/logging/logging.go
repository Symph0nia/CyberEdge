// CyberEdge/pkg/logging/logging.go

package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel 定义日志级别
type LogLevel int

const (
	// 定义不同的日志级别
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

var (
	// 日志记录器
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	fatalLogger *log.Logger

	// 日志文件
	logFile *os.File

	// 互斥锁，用于保护日志文件的并发写入
	logMutex sync.Mutex

	// 当前日志级别
	currentLogLevel LogLevel = DEBUG

	// 用于停止日志轮换的通道
	stopRotation chan struct{}

	// 最大日志文件大小（100 MB）
	maxLogSize int64 = 100 * 1024 * 1024
)

// InitializeLoggers 创建日志文件并初始化各级别的日志记录器
func InitializeLoggers(logFilePath string) error {
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 创建一个多重写入器，同时写入文件和标准输出
	multiWriter := io.MultiWriter(logFile, os.Stdout)

	// 初始化各级别的日志记录器
	debugLogger = log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	infoLogger = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	warnLogger = log.New(multiWriter, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	errorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds)
	fatalLogger = log.New(multiWriter, "FATAL: ", log.Ldate|log.Ltime|log.Lmicroseconds)

	return nil
}

// SetLogLevel 设置当前日志级别
func SetLogLevel(level LogLevel) {
	currentLogLevel = level
}

// logWithLevel 根据给定的日志级别记录日志
func logWithLevel(level LogLevel, format string, v ...interface{}) {
	if level < currentLogLevel {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	// 获取调用者的文件名和行号
	_, file, line, _ := runtime.Caller(2)
	message := fmt.Sprintf(format, v...)
	logEntry := fmt.Sprintf("%s:%d %s", filepath.Base(file), line, message)

	// 检查日志文件大小
	if fi, err := logFile.Stat(); err == nil {
		if fi.Size() > maxLogSize {
			if err := RotateLogFile(); err != nil {
				fmt.Fprintf(os.Stderr, "日志轮换失败: %v\n", err)
			}
		}
	}

	switch level {
	case DEBUG:
		debugLogger.Println(logEntry)
	case INFO:
		infoLogger.Println(logEntry)
	case WARN:
		warnLogger.Println(logEntry)
	case ERROR:
		errorLogger.Println(logEntry)
	case FATAL:
		fatalLogger.Println(logEntry)
		os.Exit(1)
	}
}

// Debug 记录调试级别的日志
func Debug(format string, v ...interface{}) {
	logWithLevel(DEBUG, format, v...)
}

// Info 记录信息级别的日志
func Info(format string, v ...interface{}) {
	logWithLevel(INFO, format, v...)
}

// Warn 记录警告级别的日志
func Warn(format string, v ...interface{}) {
	logWithLevel(WARN, format, v...)
}

// Error 记录错误级别的日志
func Error(format string, v ...interface{}) {
	logWithLevel(ERROR, format, v...)
}

// Fatal 记录致命错误级别的日志，并终止程序
func Fatal(format string, v ...interface{}) {
	logWithLevel(FATAL, format, v...)
}

// RotateLogFile 轮换日志文件
func RotateLogFile() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	// 关闭当前日志文件
	if err := logFile.Close(); err != nil {
		return fmt.Errorf("关闭当前日志文件失败: %v", err)
	}

	// 生成新的日志文件名
	timestamp := time.Now().Format("20060102-150405")
	newLogFileName := fmt.Sprintf("cyberedge-%s.log", timestamp)
	newLogFilePath := filepath.Join(filepath.Dir(logFile.Name()), newLogFileName)

	// 重命名当前日志文件
	if err := os.Rename(logFile.Name(), newLogFilePath); err != nil {
		return fmt.Errorf("重命名日志文件失败: %v", err)
	}

	// 创建新的日志文件
	return InitializeLoggers(logFile.Name())
}

// StartLogRotation 开始定期日志轮换
func StartLogRotation(interval time.Duration) {
	stopRotation = make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := RotateLogFile(); err != nil {
					Error("日志轮换失败: %v", err)
				} else {
					Info("日志轮换成功")
				}
			case <-stopRotation:
				return
			}
		}
	}()
}

// StopLogRotation 停止日志轮换
func StopLogRotation() {
	if stopRotation != nil {
		close(stopRotation)
	}
}

// 在程序退出时关闭日志文件
func cleanup() {
	if logFile != nil {
		logFile.Close()
	}
}

func init() {
	// 注册清理函数
	runtime.SetFinalizer(new(int), func(_ *int) {
		cleanup()
	})
}
