package setup

import (
	"cyberedge/pkg/logging"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
)

// ConnectToMySQL 连接到MySQL
func ConnectToMySQL(defaultDSN string) (*gorm.DB, error) {
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = defaultDSN
	}

	logging.Info("尝试连接到 MySQL: %s", maskPassword(dsn))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logging.Error("连接 MySQL 失败: %v", err)
		return nil, fmt.Errorf("连接 MySQL 失败: %v", err)
	}

	logging.Info("成功连接到 MySQL")
	return db, nil
}

// maskPassword 隐藏DSN中的密码信息用于日志输出
func maskPassword(dsn string) string {
	// 简单的密码遮盖，实际项目中可能需要更复杂的处理
	return "[DSN with password masked]"
}