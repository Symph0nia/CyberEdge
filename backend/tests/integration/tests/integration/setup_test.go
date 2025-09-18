package integration

import (
	"cyberedge/pkg/api"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/service"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/setup"
	"cyberedge/pkg/models"
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

	// Connect to test MySQL database
	testDB, err = setup.ConnectToMySQL()
	if err != nil {
		log.Printf("Warning: Failed to connect to MySQL, using in-memory SQLite: %v", err)
		// TODO: Fallback to in-memory SQLite for tests
		return
	}

	// Drop and recreate tables for clean testing
	testDB.Migrator().DropTable(&dao.SystemConfig{}, &models.User{}, &models.ToolConfig{})

	// Auto-migrate database tables for testing
	err = testDB.AutoMigrate(
		&dao.SystemConfig{},
		&models.User{},
		&models.ToolConfig{},
		// Add other models as they are migrated
	)
	if err != nil {
		log.Printf("Warning: Failed to migrate database tables: %v", err)
	}

	// Initialize DAOs
	userDAO := dao.NewUserDAO(testDB)
	configDAO := dao.NewConfigDAO(testDB)

	// Initialize Services
	userService := service.NewUserService(userDAO, configDAO, "test-jwt-secret")
	configService := service.NewConfigService(configDAO)

	// Initialize Router with minimal services
	router := api.NewRouter(
		userService,
		configService,
		nil, // taskService
		nil, // resultService
		nil, // dnsService
		nil, // httpxService
		nil, // targetService
		"test-jwt-secret",
		"test-session-secret",
		[]string{"http://localhost:3000"},
	)

	testRouter = router.SetupRouter()
}

func teardownTest() {
	if testDB != nil {
		// Clean up test database - disconnect MySQL
		setup.DisconnectMySQL(testDB)
	}
}

// GetTestRouter returns the test router instance
func GetTestRouter() *gin.Engine {
	return testRouter
}

// GetTestDB returns the test database instance
func GetTestDB() *gorm.DB {
	return testDB
}