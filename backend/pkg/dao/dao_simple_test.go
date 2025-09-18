package dao

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 简化的用户模型，兼容SQLite
type SimpleUser struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Username     string `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash string `gorm:"size:255;not null" json:"-"`
	Role         string `gorm:"size:20;default:'user'" json:"role"`  // 使用string而不是enum
	CreatedAt    int64  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime" json:"updated_at"`
}

func (SimpleUser) TableName() string {
	return "simple_users"
}

// 简化的UserDAO用于测试
type SimpleUserDAO struct {
	db *gorm.DB
}

func NewSimpleUserDAO(db *gorm.DB) *SimpleUserDAO {
	return &SimpleUserDAO{db: db}
}

func (dao *SimpleUserDAO) Create(user *SimpleUser) error {
	return dao.db.Create(user).Error
}

func (dao *SimpleUserDAO) GetByUsername(username string) (*SimpleUser, error) {
	var user SimpleUser
	err := dao.db.Where("username = ?", username).First(&user).Error
	return &user, err
}

func (dao *SimpleUserDAO) GetByEmail(email string) (*SimpleUser, error) {
	var user SimpleUser
	err := dao.db.Where("email = ?", email).First(&user).Error
	return &user, err
}

func (dao *SimpleUserDAO) GetAll() ([]*SimpleUser, error) {
	var users []*SimpleUser
	err := dao.db.Find(&users).Error
	return users, err
}

func (dao *SimpleUserDAO) Update(user *SimpleUser) error {
	return dao.db.Save(user).Error
}

func (dao *SimpleUserDAO) Delete(id uint) error {
	return dao.db.Delete(&SimpleUser{}, id).Error
}

// 创建测试数据库
func setupSimpleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移简化的用户表
	err = db.AutoMigrate(&SimpleUser{})
	require.NoError(t, err)

	return db
}

// 简化的DAO测试
func TestSimpleUserDAO(t *testing.T) {
	db := setupSimpleTestDB(t)
	userDAO := NewSimpleUserDAO(db)

	t.Run("Create user", func(t *testing.T) {
		user := &SimpleUser{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		err := userDAO.Create(user)
		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("Get user by username", func(t *testing.T) {
		// 先创建用户
		user := &SimpleUser{
			Username:     "gettest",
			Email:        "gettest@example.com",
			PasswordHash: "hashed_password",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}
		err := userDAO.Create(user)
		require.NoError(t, err)

		// 获取用户
		foundUser, err := userDAO.GetByUsername("gettest")
		assert.NoError(t, err)
		assert.Equal(t, "gettest", foundUser.Username)
		assert.Equal(t, "gettest@example.com", foundUser.Email)
	})

	t.Run("Get user by email", func(t *testing.T) {
		// 先创建用户
		user := &SimpleUser{
			Username:     "emailtest",
			Email:        "emailtest@example.com",
			PasswordHash: "hashed_password",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}
		err := userDAO.Create(user)
		require.NoError(t, err)

		// 通过邮箱获取用户
		foundUser, err := userDAO.GetByEmail("emailtest@example.com")
		assert.NoError(t, err)
		assert.Equal(t, "emailtest", foundUser.Username)
		assert.Equal(t, "emailtest@example.com", foundUser.Email)
	})

	t.Run("Update user", func(t *testing.T) {
		// 创建用户
		user := &SimpleUser{
			Username:     "updatetest",
			Email:        "updatetest@example.com",
			PasswordHash: "old_hash",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}
		err := userDAO.Create(user)
		require.NoError(t, err)

		// 更新用户
		user.PasswordHash = "new_hash"
		user.Role = "admin"
		err = userDAO.Update(user)
		assert.NoError(t, err)

		// 验证更新
		updatedUser, err := userDAO.GetByUsername("updatetest")
		assert.NoError(t, err)
		assert.Equal(t, "new_hash", updatedUser.PasswordHash)
		assert.Equal(t, "admin", updatedUser.Role)
	})

	t.Run("Delete user", func(t *testing.T) {
		// 创建用户
		user := &SimpleUser{
			Username:     "deletetest",
			Email:        "deletetest@example.com",
			PasswordHash: "hash",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}
		err := userDAO.Create(user)
		require.NoError(t, err)

		// 删除用户
		err = userDAO.Delete(user.ID)
		assert.NoError(t, err)

		// 验证删除
		_, err = userDAO.GetByUsername("deletetest")
		assert.Error(t, err) // 应该找不到用户
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("Handle duplicate username", func(t *testing.T) {
		user1 := &SimpleUser{
			Username:     "duplicate",
			Email:        "dup1@example.com",
			PasswordHash: "hash1",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		user2 := &SimpleUser{
			Username:     "duplicate", // 重复用户名
			Email:        "dup2@example.com",
			PasswordHash: "hash2",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		// 第一个用户应该成功
		err := userDAO.Create(user1)
		assert.NoError(t, err)

		// 第二个用户应该失败（重复用户名）
		err = userDAO.Create(user2)
		assert.Error(t, err)
	})

	t.Run("Get all users", func(t *testing.T) {
		// 创建多个用户
		users := []*SimpleUser{
			{Username: "user1", Email: "user1@test.com", PasswordHash: "hash1", Role: "user", CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
			{Username: "user2", Email: "user2@test.com", PasswordHash: "hash2", Role: "admin", CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()},
		}

		for _, user := range users {
			err := userDAO.Create(user)
			require.NoError(t, err)
		}

		// 获取所有用户
		allUsers, err := userDAO.GetAll()
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(allUsers), 2)
	})
}

func TestDatabaseOperations(t *testing.T) {
	db := setupSimpleTestDB(t)

	t.Run("Transaction test", func(t *testing.T) {
		err := db.Transaction(func(tx *gorm.DB) error {
			user := &SimpleUser{
				Username:     "txtest",
				Email:        "txtest@example.com",
				PasswordHash: "hash",
				Role:         "user",
			}
			return tx.Create(user).Error
		})
		assert.NoError(t, err)

		// 验证用户已创建
		var user SimpleUser
		err = db.Where("username = ?", "txtest").First(&user).Error
		assert.NoError(t, err)
	})

	t.Run("Batch operations", func(t *testing.T) {
		users := []*SimpleUser{
			{Username: "batch1", Email: "batch1@test.com", PasswordHash: "hash", Role: "user"},
			{Username: "batch2", Email: "batch2@test.com", PasswordHash: "hash", Role: "user"},
			{Username: "batch3", Email: "batch3@test.com", PasswordHash: "hash", Role: "user"},
		}

		err := db.CreateInBatches(users, 2).Error
		assert.NoError(t, err)

		// 验证批量创建
		var count int64
		db.Model(&SimpleUser{}).Where("username LIKE 'batch%'").Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("Query optimization", func(t *testing.T) {
		// 创建测试数据
		user := &SimpleUser{
			Username:     "querytest",
			Email:        "querytest@example.com",
			PasswordHash: "hash",
			Role:         "admin",
		}
		err := db.Create(user).Error
		require.NoError(t, err)

		// 测试不同查询方式
		var users1 []SimpleUser
		err = db.Where("role = ?", "admin").Find(&users1).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users1), 1)

		var users2 []SimpleUser
		err = db.Where("username LIKE ?", "query%").Find(&users2).Error
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users2), 1)
	})
}