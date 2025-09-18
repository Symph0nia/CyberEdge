package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
	"cyberedge/pkg/services"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// 自动迁移测试表
	err = db.AutoMigrate(
		&models.ProjectOptimized{},
		&models.ScanTarget{},
		&models.ScanResultOptimized{},
		&models.VulnerabilityOptimized{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupTestHandler() (*ScanFrameworkHandler, *gorm.DB, error) {
	db, err := setupTestDB()
	if err != nil {
		return nil, nil, err
	}

	scanDAO := dao.NewScanDAO(db)

	scanService, err := services.NewScanService(scanDAO)
	if err != nil {
		return nil, nil, err
	}

	handler := NewScanFrameworkHandler(scanService)
	return handler, db, nil
}

func TestScanFrameworkHandler_StartScan(t *testing.T) {
	handler, db, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}

	// 创建测试项目
	project := &models.ProjectOptimized{
		Name:        "Test Project",
		Description: "Test Description",
		CreatedAt:   time.Now(),
	}
	db.Create(project)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterScanFrameworkRoutes(router.Group("/api"))

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedFields []string
	}{
		{
			name: "Valid scan request",
			requestBody: map[string]interface{}{
				"project_id": project.ID,
				"target":     "example.com",
				"pipeline":   "quick",
			},
			expectedStatus: http.StatusCreated,
			expectedFields: []string{"success", "data", "message"},
		},
		{
			name: "Missing project_id",
			requestBody: map[string]interface{}{
				"target":   "example.com",
				"pipeline": "quick",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name: "Missing target",
			requestBody: map[string]interface{}{
				"project_id": project.ID,
				"pipeline":   "quick",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name: "Missing pipeline",
			requestBody: map[string]interface{}{
				"project_id": project.ID,
				"target":     "example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
		{
			name: "Invalid pipeline",
			requestBody: map[string]interface{}{
				"project_id": project.ID,
				"target":     "example.com",
				"pipeline":   "nonexistent",
			},
			expectedStatus: http.StatusBadRequest,
			expectedFields: []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/scan-framework/start", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			for _, field := range tt.expectedFields {
				if _, exists := response[field]; !exists {
					t.Errorf("Expected field '%s' in response", field)
				}
			}

			// 验证成功响应的数据结构
			if tt.expectedStatus == http.StatusCreated {
				if success, ok := response["success"].(bool); !ok || !success {
					t.Error("Expected success to be true")
				}

				if data, ok := response["data"].(map[string]interface{}); ok {
					requiredFields := []string{"scan_id", "project_id", "target_id", "state", "created_at", "pipeline"}
					for _, field := range requiredFields {
						if _, exists := data[field]; !exists {
							t.Errorf("Expected field '%s' in data", field)
						}
					}
				} else {
					t.Error("Expected data field to be an object")
				}
			}
		})
	}
}

func TestScanFrameworkHandler_GetScanStatus(t *testing.T) {
	handler, db, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}

	// 创建测试扫描目标
	scanTarget := &models.ScanTarget{
		ProjectID: 1,
		Address:   "example.com",
		Type:      "domain",
		CreatedAt: time.Now(),
	}
	db.Create(scanTarget)

	// 创建测试扫描结果
	scanResult := &models.ScanResultOptimized{
		ProjectID:   1,
		TargetID:    scanTarget.ID,
		Port:        80,
		Protocol:    "tcp",
		State:       "completed",
		ServiceName: "test-scanner",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(time.Minute),
	}
	db.Create(scanResult)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterScanFrameworkRoutes(router.Group("/api"))

	tests := []struct {
		name           string
		scanID         string
		expectedStatus int
		checkFields    []string
	}{
		{
			name:           "Valid scan ID",
			scanID:         "1",
			expectedStatus: http.StatusOK,
			checkFields:    []string{"scan_id", "project_id", "target_id", "state"},
		},
		{
			name:           "Invalid scan ID",
			scanID:         "invalid",
			expectedStatus: http.StatusBadRequest,
			checkFields:    []string{"error"},
		},
		{
			name:           "Non-existent scan ID",
			scanID:         "999",
			expectedStatus: http.StatusNotFound,
			checkFields:    []string{"error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/scan-framework/status/"+tt.scanID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			for _, field := range tt.checkFields {
				if tt.expectedStatus == http.StatusOK {
					if data, ok := response["data"].(map[string]interface{}); ok {
						if _, exists := data[field]; !exists {
							t.Errorf("Expected field '%s' in data", field)
						}
					}
				} else {
					if _, exists := response[field]; !exists {
						t.Errorf("Expected field '%s' in response", field)
					}
				}
			}
		})
	}
}

func TestScanFrameworkHandler_GetAvailableTools(t *testing.T) {
	handler, _, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterScanFrameworkRoutes(router.Group("/api"))

	req, _ := http.NewRequest("GET", "/api/scan-framework/tools", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if data, ok := response["data"].(map[string]interface{}); ok {
		// 检查是否包含扫描工具类别
		expectedCategories := []string{"subdomain", "port", "webtech", "vulnerability", "webpath"}
		for _, category := range expectedCategories {
			if tools, exists := data[category]; exists {
				if toolsList, ok := tools.([]interface{}); ok {
					// 验证工具列表结构
					for _, tool := range toolsList {
						if toolMap, ok := tool.(map[string]interface{}); ok {
							if _, hasName := toolMap["name"]; !hasName {
								t.Error("Tool should have name field")
							}
							if _, hasAvailable := toolMap["available"]; !hasAvailable {
								t.Error("Tool should have available field")
							}
						}
					}
				}
			}
		}
	} else {
		t.Error("Expected data field to be an object")
	}
}

func TestScanFrameworkHandler_GetAvailablePipelines(t *testing.T) {
	handler, _, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterScanFrameworkRoutes(router.Group("/api"))

	req, _ := http.NewRequest("GET", "/api/scan-framework/pipelines", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if data, ok := response["data"].([]interface{}); ok {
		if len(data) == 0 {
			t.Error("Expected at least one pipeline")
		}

		// 验证流水线结构
		for _, pipeline := range data {
			if pipelineMap, ok := pipeline.(map[string]interface{}); ok {
				requiredFields := []string{"name", "key", "parallel", "continue_on_error", "stages"}
				for _, field := range requiredFields {
					if _, exists := pipelineMap[field]; !exists {
						t.Errorf("Pipeline should have field '%s'", field)
					}
				}

				// 验证stages结构
				if stages, ok := pipelineMap["stages"].([]interface{}); ok {
					for _, stage := range stages {
						if stageMap, ok := stage.(map[string]interface{}); ok {
							stageFields := []string{"name", "scanner_names", "parallel", "depends_on"}
							for _, field := range stageFields {
								if _, exists := stageMap[field]; !exists {
									t.Errorf("Stage should have field '%s'", field)
								}
							}
						}
					}
				}
			}
		}
	} else {
		t.Error("Expected data field to be an array")
	}
}

func TestScanFrameworkHandler_GetProjectScanResults(t *testing.T) {
	handler, db, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}

	// 创建测试数据
	project := &models.ProjectOptimized{
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	db.Create(project)

	target := &models.ScanTarget{
		ProjectID: project.ID,
		Address:   "example.com",
		Type:      "domain",
		CreatedAt: time.Now(),
	}
	db.Create(target)

	scanResult := &models.ScanResultOptimized{
		ProjectID:   project.ID,
		TargetID:    target.ID,
		Port:        80,
		Protocol:    "tcp",
		State:       "completed",
		ServiceName: "subdomain",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now().Add(time.Minute),
	}
	db.Create(scanResult)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterScanFrameworkRoutes(router.Group("/api"))

	tests := []struct {
		name           string
		projectID      string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Valid project ID",
			projectID:      "1",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "With scan_type filter",
			projectID:      "1",
			queryParams:    "?scan_type=subdomain",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "With status filter",
			projectID:      "1",
			queryParams:    "?status=completed",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Invalid project ID",
			projectID:      "invalid",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/scan-framework/results/project/" + tt.projectID + tt.queryParams
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse response: %v", err)
			}

			if tt.expectedStatus == http.StatusOK {
				if data, ok := response["data"].([]interface{}); ok {
					if len(data) != tt.expectedCount {
						t.Errorf("Expected %d results, got %d", tt.expectedCount, len(data))
					}

					if tt.expectedCount > 0 {
						// 验证结果结构
						if result, ok := data[0].(map[string]interface{}); ok {
							requiredFields := []string{"id", "target_id", "port", "protocol", "state", "service_name"}
							for _, field := range requiredFields {
								if _, exists := result[field]; !exists {
									t.Errorf("Result should have field '%s'", field)
								}
							}
						}
					}
				}

				if count, ok := response["count"].(float64); ok {
					if int(count) != tt.expectedCount {
						t.Errorf("Expected count %d, got %d", tt.expectedCount, int(count))
					}
				}
			}
		})
	}
}

func TestScanFrameworkHandler_GetVulnerabilityStats(t *testing.T) {
	handler, db, err := setupTestHandler()
	if err != nil {
		t.Fatalf("Failed to setup test handler: %v", err)
	}

	// 创建测试数据
	project := &models.ProjectOptimized{
		Name:      "Test Project",
		CreatedAt: time.Now(),
	}
	db.Create(project)

	target := &models.ScanTarget{
		ProjectID: project.ID,
		Address:   "example.com",
		Type:      "domain",
		CreatedAt: time.Now(),
	}
	db.Create(target)

	scanResult := &models.ScanResultOptimized{
		ProjectID:   project.ID,
		TargetID:    target.ID,
		Port:        80,
		Protocol:    "tcp",
		State:       "completed",
		ServiceName: "nuclei",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	db.Create(scanResult)

	// 创建不同严重程度的漏洞
	vulnerabilities := []*models.VulnerabilityOptimized{
		{
			ScanResultID: scanResult.ID,
			Title:        "Critical Vuln",
			Severity:     "critical",
			CreatedAt:    time.Now(),
		},
		{
			ScanResultID: scanResult.ID,
			Title:        "High Vuln",
			Severity:     "high",
			CreatedAt:    time.Now(),
		},
		{
			ScanResultID: scanResult.ID,
			Title:        "Medium Vuln",
			Severity:     "medium",
			CreatedAt:    time.Now(),
		},
	}

	for _, vuln := range vulnerabilities {
		db.Create(vuln)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler.RegisterScanFrameworkRoutes(router.Group("/api"))

	req, _ := http.NewRequest("GET", "/api/scan-framework/vulnerabilities/stats/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if success, ok := response["success"].(bool); !ok || !success {
		t.Error("Expected success to be true")
	}

	if data, ok := response["data"].(map[string]interface{}); ok {
		if projectID, ok := data["project_id"].(float64); !ok || int(projectID) != 1 {
			t.Error("Expected project_id to be 1")
		}

		if stats, ok := data["stats"].(map[string]interface{}); ok {
			expectedSeverities := []string{"critical", "high", "medium", "low", "info"}
			for _, severity := range expectedSeverities {
				if _, exists := stats[severity]; !exists {
					t.Errorf("Expected severity '%s' in stats", severity)
				}
			}

			// 验证统计数字
			if critical, ok := stats["critical"].(float64); ok {
				if int(critical) != 1 {
					t.Errorf("Expected 1 critical vulnerability, got %d", int(critical))
				}
			}
		}

		if total, ok := data["total"].(float64); ok {
			if int(total) != 3 {
				t.Errorf("Expected total 3, got %d", int(total))
			}
		}
	}
}