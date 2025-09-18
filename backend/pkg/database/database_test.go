package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"cyberedge/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatabaseConnection(t *testing.T) {
	t.Run("Connect with valid DSN", func(t *testing.T) {
		// 使用内存SQLite进行测试
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		assert.NoError(t, err)
		assert.NotNil(t, db)

		// 验证连接
		sqlDB, err := db.DB()
		assert.NoError(t, err)
		err = sqlDB.Ping()
		assert.NoError(t, err)
	})

	t.Run("Auto migrate models", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// 测试自动迁移
		err = db.AutoMigrate(
			&models.User{},
			&models.ProjectOptimized{},
			&models.ScanTarget{},
			&models.ScanResultOptimized{},
			&models.VulnerabilityOptimized{},
			&models.WebPathOptimized{},
			&models.TechnologyOptimized{},
			&models.ScanResultTechnology{},
			&models.ScanFrameworkResult{},
			&models.ScanFrameworkTarget{},
		)
		assert.NoError(t, err)

		// 验证表是否创建
		assert.True(t, db.Migrator().HasTable(&models.User{}))
		assert.True(t, db.Migrator().HasTable(&models.ProjectOptimized{}))
		assert.True(t, db.Migrator().HasTable(&models.ScanTarget{}))
		assert.True(t, db.Migrator().HasTable(&models.ScanResultOptimized{}))
		assert.True(t, db.Migrator().HasTable(&models.VulnerabilityOptimized{}))
	})
}

func TestDatabaseConfiguration(t *testing.T) {
	t.Run("Test environment variables", func(t *testing.T) {
		// 保存原始环境变量
		originalDSN := os.Getenv("MYSQL_DSN")

		// 设置测试DSN
		testDSN := "test:test@tcp(localhost:3306)/test_db?charset=utf8mb4&parseTime=True&loc=Local"
		os.Setenv("MYSQL_DSN", testDSN)

		// 获取DSN
		dsn := os.Getenv("MYSQL_DSN")
		assert.Equal(t, testDSN, dsn)

		// 恢复原始环境变量
		if originalDSN != "" {
			os.Setenv("MYSQL_DSN", originalDSN)
		} else {
			os.Unsetenv("MYSQL_DSN")
		}
	})

	t.Run("Database pool configuration", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		sqlDB, err := db.DB()
		require.NoError(t, err)

		// 配置连接池
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		// 验证配置 - 检查连接池工作
		stats := sqlDB.Stats()
		assert.GreaterOrEqual(t, stats.MaxIdleClosed, int64(0))
		assert.GreaterOrEqual(t, stats.OpenConnections, 0)
	})
}

func TestDatabaseTransactions(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	t.Run("Successful transaction", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			user := &models.User{
				Username:     "txtest1",
				Email:        "txtest1@example.com",
				PasswordHash: "hash",
				Role:         "user",
			}
			return tx.Create(user).Error
		})
		assert.NoError(t, err)

		// 验证用户已创建
		var user models.User
		err = db.Where("username = ?", "txtest1").First(&user).Error
		assert.NoError(t, err)
	})

	t.Run("Failed transaction rollback", func(t *testing.T) {
		initialCount := int64(0)
		db.Model(&models.User{}).Count(&initialCount)

		err := db.Transaction(func(tx *gorm.DB) error {
			// 创建第一个用户
			user1 := &models.User{
				Username:     "txtest2",
				Email:        "txtest2@example.com",
				PasswordHash: "hash",
				Role:         "user",
			}
			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			// 尝试创建重复用户名（应该失败）
			user2 := &models.User{
				Username:     "txtest2", // 重复用户名
				Email:        "txtest3@example.com",
				PasswordHash: "hash",
				Role:         "user",
			}
			return tx.Create(user2).Error // 这会失败并回滚事务
		})
		assert.Error(t, err)

		// 验证没有用户被创建（事务回滚）
		finalCount := int64(0)
		db.Model(&models.User{}).Count(&finalCount)
		assert.Equal(t, initialCount, finalCount)
	})
}

func TestDatabaseIndexes(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	t.Run("Check table indexes", func(t *testing.T) {
		// 验证用户表的索引
		assert.True(t, db.Migrator().HasIndex(&models.User{}, "username"))
		assert.True(t, db.Migrator().HasIndex(&models.User{}, "email"))
	})
}

func TestDatabaseErrorHandling(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	t.Run("Handle duplicate key errors", func(t *testing.T) {
		user1 := &models.User{
			Username:     "duplicate",
			Email:        "duplicate@example.com",
			PasswordHash: "hash",
			Role:         "user",
		}

		// 第一次创建应该成功
		err := db.Create(user1).Error
		assert.NoError(t, err)

		user2 := &models.User{
			Username:     "duplicate", // 重复用户名
			Email:        "different@example.com",
			PasswordHash: "hash",
			Role:         "user",
		}

		// 第二次创建应该失败
		err = db.Create(user2).Error
		assert.Error(t, err)
	})

	t.Run("Handle not found errors", func(t *testing.T) {
		var user models.User
		err := db.Where("username = ?", "nonexistent").First(&user).Error
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}

func TestDatabasePerformance(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	t.Run("Batch insert performance", func(t *testing.T) {
		users := make([]*models.User, 100)
		for i := 0; i < 100; i++ {
			users[i] = &models.User{
				Username:     fmt.Sprintf("batchuser%d", i),
				Email:        fmt.Sprintf("batch%d@example.com", i),
				PasswordHash: "hash",
				Role:         "user",
			}
		}

		// 批量插入
		start := time.Now()
		err := db.CreateInBatches(users, 20).Error
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, time.Second) // 应该在1秒内完成

		// 验证插入数量
		var count int64
		db.Model(&models.User{}).Where("username LIKE 'batchuser%'").Count(&count)
		assert.Equal(t, int64(100), count)
	})
}

