package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"cyberedge/pkg/models"
)

// 验证GORM模型与数据库schema一致性的测试

func TestSchemaConsistency(t *testing.T) {
	// 使用测试数据库
	dsn := "root:password@tcp(localhost:3306)/cyberedge_test?charset=utf8mb4&parseTime=True&loc=Local"

	// 创建测试数据库
	createTestDB()

	// 连接GORM
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移所有模型
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
	if err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	// 验证表结构
	t.Run("ValidateTableStructure", func(t *testing.T) {
		validateTableStructure(t, db)
	})

	// 验证索引
	t.Run("ValidateIndexes", func(t *testing.T) {
		validateIndexes(t, db)
	})

	// 验证外键
	t.Run("ValidateForeignKeys", func(t *testing.T) {
		validateForeignKeys(t, db)
	})

	// 清理测试数据库
	dropTestDB()
}

func createTestDB() {
	dsn := "root:password@tcp(localhost:3306)/"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS cyberedge_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	if err != nil {
		log.Fatalf("Failed to create test database: %v", err)
	}
}

func dropTestDB() {
	dsn := "root:password@tcp(localhost:3306)/"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Printf("Failed to connect to MySQL for cleanup: %v", err)
		return
	}
	defer db.Close()

	_, err = db.Exec("DROP DATABASE IF EXISTS cyberedge_test")
	if err != nil {
		log.Printf("Failed to drop test database: %v", err)
	}
}

func validateTableStructure(t *testing.T, db *gorm.DB) {
	// 期望的表列表
	expectedTables := []string{
		"users",
		"project_optimizeds",
		"scan_targets",
		"scan_result_optimizeds",
		"vulnerability_optimizeds",
		"web_path_optimizeds",
		"technology_optimizeds",
		"scan_result_technologies",
		"scan_framework_results",
		"scan_framework_targets",
	}

	// 检查表是否存在
	for _, tableName := range expectedTables {
		if !db.Migrator().HasTable(tableName) {
			t.Errorf("Table %s does not exist", tableName)
		}
	}

	// 检查关键字段
	validateUsersTable(t, db)
	validateProjectTable(t, db)
	validateScanTargetTable(t, db)
	validateScanResultTable(t, db)
	validateVulnerabilityTable(t, db)
}

func validateUsersTable(t *testing.T, db *gorm.DB) {
	requiredColumns := []string{"id", "username", "email", "password_hash", "is_2fa_enabled", "totp_secret", "role", "created_at", "updated_at"}

	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&models.User{}, column) {
			t.Errorf("users table missing column: %s", column)
		}
	}
}

func validateProjectTable(t *testing.T, db *gorm.DB) {
	requiredColumns := []string{"id", "name", "description", "created_at", "updated_at", "deleted_at"}

	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&models.ProjectOptimized{}, column) {
			t.Errorf("project_optimizeds table missing column: %s", column)
		}
	}
}

func validateScanTargetTable(t *testing.T, db *gorm.DB) {
	requiredColumns := []string{"id", "project_id", "type", "address", "parent_id", "created_at", "updated_at"}

	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&models.ScanTarget{}, column) {
			t.Errorf("scan_targets table missing column: %s", column)
		}
	}
}

func validateScanResultTable(t *testing.T, db *gorm.DB) {
	requiredColumns := []string{
		"id", "project_id", "target_id", "port", "protocol", "state",
		"service_name", "version", "fingerprint", "banner",
		"is_web_service", "http_title", "http_status",
		"created_at", "updated_at",
	}

	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&models.ScanResultOptimized{}, column) {
			t.Errorf("scan_result_optimizeds table missing column: %s", column)
		}
	}
}

func validateVulnerabilityTable(t *testing.T, db *gorm.DB) {
	requiredColumns := []string{
		"id", "scan_result_id", "web_path_id", "cve_id", "title", "description",
		"severity", "cvss", "location", "parameter", "payload", "status",
		"created_at", "updated_at",
	}

	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&models.VulnerabilityOptimized{}, column) {
			t.Errorf("vulnerability_optimizeds table missing column: %s", column)
		}
	}
}

func validateIndexes(t *testing.T, db *gorm.DB) {
	// 获取数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database connection: %v", err)
	}

	// 检查关键索引
	expectedIndexes := map[string][]string{
		"users": {"idx_username", "idx_email"},
		"project_optimizeds": {"idx_name"},
		"scan_targets": {"idx_project_id", "idx_type", "idx_address"},
		"scan_result_optimizeds": {"idx_project_id", "idx_target_id", "idx_port", "idx_service_name"},
		"vulnerability_optimizeds": {"idx_scan_result_id", "idx_severity", "idx_status"},
	}

	for table, indexes := range expectedIndexes {
		for _, index := range indexes {
			query := `
				SELECT COUNT(*)
				FROM information_schema.statistics
				WHERE table_schema = DATABASE()
				AND table_name = ?
				AND index_name = ?`

			var count int
			err := sqlDB.QueryRow(query, table, index).Scan(&count)
			if err != nil {
				t.Errorf("Error checking index %s on table %s: %v", index, table, err)
			}
			if count == 0 {
				t.Errorf("Index %s missing on table %s", index, table)
			}
		}
	}
}

func validateForeignKeys(t *testing.T, db *gorm.DB) {
	// 获取数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("Failed to get database connection: %v", err)
	}

	// 检查关键外键约束
	expectedForeignKeys := map[string]string{
		"scan_targets": "project_id",
		"scan_result_optimizeds": "project_id,target_id",
		"vulnerability_optimizeds": "scan_result_id",
		"web_path_optimizeds": "scan_result_id",
		"scan_framework_results": "project_id,scan_target_id",
	}

	for table, columns := range expectedForeignKeys {
		columnList := strings.Split(columns, ",")
		for _, column := range columnList {
			query := `
				SELECT COUNT(*)
				FROM information_schema.key_column_usage
				WHERE table_schema = DATABASE()
				AND table_name = ?
				AND column_name = ?
				AND referenced_table_name IS NOT NULL`

			var count int
			err := sqlDB.QueryRow(query, table, column).Scan(&count)
			if err != nil {
				t.Errorf("Error checking foreign key %s on table %s: %v", column, table, err)
			}
			if count == 0 {
				t.Errorf("Foreign key constraint missing for column %s on table %s", column, table)
			}
		}
	}
}

// 辅助函数：检查GORM模型与实际schema的一致性
func validateModelSchemaConsistency(t *testing.T, db *gorm.DB) {
	// 这里可以添加更详细的一致性检查
	// 比如字段类型、长度限制等

	// 检查用户模型
	stmt := &gorm.Statement{DB: db}
	stmt.Parse(&models.User{})

	if stmt.Schema.Table != "users" {
		t.Errorf("User model table name mismatch: expected 'users', got '%s'", stmt.Schema.Table)
	}

	// 检查项目模型
	stmt.Parse(&models.ProjectOptimized{})
	if stmt.Schema.Table != "project_optimizeds" {
		t.Errorf("ProjectOptimized model table name mismatch: expected 'project_optimizeds', got '%s'", stmt.Schema.Table)
	}
}

// 运行测试的主函数（用于手动执行）
func main() {
	if len(os.Args) > 1 && os.Args[1] == "test" {
		fmt.Println("Running schema validation tests...")

		// 运行测试
		testing.Main(func(pat, str string) (bool, error) { return true, nil },
			[]testing.InternalTest{
				{"TestSchemaConsistency", TestSchemaConsistency},
			},
			[]testing.InternalBenchmark{},
			[]testing.InternalExample{})
	} else {
		fmt.Println("Schema validation test file created.")
		fmt.Println("Run with: go run schema_validation_test.go test")
	}
}