// utils/logging.go

package utils

import (
	"cyberedge/pkg/logging"
	"path/filepath"
	"time"
)

// InitializeLogging 初始化日志系统
func InitializeLogging(logDir string) error {
	logPath := filepath.Join(logDir, "cyberedge.log")
	if err := logging.InitializeLoggers(logPath); err != nil {
		return err
	}
	logging.Info("日志系统初始化成功")

	// 启动日志轮换（每24小时轮换一次）
	logging.StartLogRotation(24 * time.Hour)
	return nil
}

// StopLogging 停止日志轮换
func StopLogging() {
	logging.StopLogRotation()
}
