package integration

import (
	"cyberedge/pkg/api"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/service"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/database"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"testing"
	"os"
	"log"
)

var (
	testRouter *gin.Engine
	testDB     *gorm.DB
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Setup
	setupTest()

	// Run tests
	code := m.Run()

	// Teardown
	teardownTest()

	os.Exit(code)
}

func setupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize logging to discard output in tests
	err := logging.InitializeLoggers("/dev/null")
	if err != nil {
		log.Printf("Warning: Failed to initialize logging: %v", err)
	}

	// Connect to test database
	dsn := "root:password@tcp(localhost:3306)/cyberedge_test?charset=utf8mb4&parseTime=True&loc=Local"
	testDB, err = database.Connect(dsn)
	if err != nil {
		log.Printf("Warning: Failed to connect to MySQL, skipping integration tests: %v", err)
		return
	}

	// Auto-migrate test tables
	err = testDB.AutoMigrate(
		&dao.User{},
		// Add other models as needed
	)
	if err != nil {
		log.Printf("Warning: Failed to migrate test database: %v", err)
	}

	// Initialize DAOs
	userDAO := dao.NewUserDAO(testDB)

	// Initialize services
	jwtSecret := "test-secret-key"
	userService := service.NewUserService(userDAO, jwtSecret)

	// Create scan service (simplified for testing)
	scanService := service.NewScanService(testDB)

	// Initialize router
	allowedOrigins := []string{"http://localhost:8080"}
	testRouter = api.NewRouter(userService, scanService, jwtSecret, allowedOrigins)
}

func teardownTest() {
	if testDB != nil {
		// Clean up test data
		testDB.Exec("DROP DATABASE IF EXISTS cyberedge_test")
	}
}