package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
)

// 优化后的测试 - 基于Linus审查建议，专注业务逻辑测试，不测试数据转换

// MockScanDAOOptimized for testing
type MockScanDAOOptimized struct {
	mock.Mock
}

// 确保Mock实现接口
var _ dao.ScanDAOOptimizedInterface = (*MockScanDAOOptimized)(nil)

func (m *MockScanDAOOptimized) CreateProject(project *models.ProjectOptimized) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockScanDAOOptimized) GetProjectByID(id uint) (*models.ProjectOptimized, error) {
	args := m.Called(id)
	if project := args.Get(0); project != nil {
		return project.(*models.ProjectOptimized), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScanDAOOptimized) GetProjectByName(name string) (*models.ProjectOptimized, error) {
	args := m.Called(name)
	if project := args.Get(0); project != nil {
		return project.(*models.ProjectOptimized), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScanDAOOptimized) ListProjects() ([]models.ProjectOptimized, error) {
	args := m.Called()
	return args.Get(0).([]models.ProjectOptimized), args.Error(1)
}

func (m *MockScanDAOOptimized) DeleteProject(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockScanDAOOptimized) GetProjectDetails(projectID uint) (*models.ProjectStatsOptimized, []models.ScanTarget, error) {
	args := m.Called(projectID)
	return args.Get(0).(*models.ProjectStatsOptimized), args.Get(1).([]models.ScanTarget), args.Error(2)
}

func (m *MockScanDAOOptimized) GetProjectStatsOptimized(projectID uint) (*models.ProjectStatsOptimized, error) {
	args := m.Called(projectID)
	return args.Get(0).(*models.ProjectStatsOptimized), args.Error(1)
}

func (m *MockScanDAOOptimized) ImportScanData(data *models.ScanDataImport) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockScanDAOOptimized) GetVulnerabilities(projectID uint, filters map[string]interface{}) ([]models.VulnerabilityOptimized, error) {
	args := m.Called(projectID, filters)
	return args.Get(0).([]models.VulnerabilityOptimized), args.Error(1)
}

func (m *MockScanDAOOptimized) GetProjectHierarchy(projectID uint) ([]models.ScanTarget, error) {
	args := m.Called(projectID)
	return args.Get(0).([]models.ScanTarget), args.Error(1)
}

func (m *MockScanDAOOptimized) SearchTargets(projectID uint, searchTerm string) ([]models.ScanTarget, error) {
	args := m.Called(projectID, searchTerm)
	return args.Get(0).([]models.ScanTarget), args.Error(1)
}

// 测试核心业务规则：项目管理
func TestOptimizedCreateProject(t *testing.T) {
	t.Run("Create project with valid data", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		mockDAO.On("GetProjectByName", "Test Project").Return(nil, errors.New("not found"))
		mockDAO.On("CreateProject", mock.AnythingOfType("*models.ProjectOptimized")).Return(nil)

		project, err := service.CreateProject("Test Project", "Test Description")

		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, "Test Project", project.Name)
		mockDAO.AssertExpectations(t)
	})

	t.Run("Reject empty project name", func(t *testing.T) {
		service := NewScanServiceOptimized(new(MockScanDAOOptimized))

		project, err := service.CreateProject("", "Description")

		assert.Error(t, err)
		assert.Nil(t, project)
		assert.Contains(t, err.Error(), "项目名称不能为空")
	})

	t.Run("Reject duplicate project name", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		existingProject := &models.ProjectOptimized{Name: "Existing Project"}
		mockDAO.On("GetProjectByName", "Existing Project").Return(existingProject, nil)

		project, err := service.CreateProject("Existing Project", "Description")

		assert.Error(t, err)
		assert.Nil(t, project)
		assert.Contains(t, err.Error(), "项目名称已存在")
		mockDAO.AssertExpectations(t)
	})
}

// 测试数据验证业务逻辑
func TestOptimizedValidateScanData(t *testing.T) {
	service := NewScanServiceOptimized(new(MockScanDAOOptimized))

	t.Run("Valid scan data passes validation", func(t *testing.T) {
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results: []models.ScanTargetImport{
				{
					Type:    "domain",
					Address: "example.com",
					Ports: []models.PortScanImport{
						{
							Number:   80,
							Protocol: "tcp",
							State:    "open",
						},
					},
				},
			},
		}

		err := service.validateScanData(data)
		assert.NoError(t, err)
	})

	t.Run("Reject empty results", func(t *testing.T) {
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results:   []models.ScanTargetImport{},
		}

		err := service.validateScanData(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "扫描结果不能为空")
	})

	t.Run("Reject invalid target type", func(t *testing.T) {
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results: []models.ScanTargetImport{
				{
					Type:    "invalid_type",
					Address: "example.com",
					Ports:   []models.PortScanImport{},
				},
			},
		}

		err := service.validateScanData(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "目标类型无效")
	})

	t.Run("Reject invalid port number", func(t *testing.T) {
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results: []models.ScanTargetImport{
				{
					Type:    "ip",
					Address: "192.168.1.1",
					Ports: []models.PortScanImport{
						{
							Number:   70000, // Invalid port
							Protocol: "tcp",
							State:    "open",
						},
					},
				},
			},
		}

		err := service.validateScanData(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "端口号无效")
	})

	t.Run("Reject invalid protocol", func(t *testing.T) {
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results: []models.ScanTargetImport{
				{
					Type:    "ip",
					Address: "192.168.1.1",
					Ports: []models.PortScanImport{
						{
							Number:   80,
							Protocol: "invalid_protocol",
							State:    "open",
						},
					},
				},
			},
		}

		err := service.validateScanData(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "端口协议无效")
	})
}

// 测试漏洞过滤验证逻辑
func TestOptimizedValidateVulnerabilityFilters(t *testing.T) {
	service := NewScanServiceOptimized(new(MockScanDAOOptimized))

	t.Run("Valid filters pass validation", func(t *testing.T) {
		filters := map[string]interface{}{
			"severity": "critical",
			"status":   "open",
			"search":   "sql injection",
		}

		err := service.validateVulnerabilityFilters(filters)
		assert.NoError(t, err)
	})

	t.Run("Invalid severity rejected", func(t *testing.T) {
		filters := map[string]interface{}{
			"severity": "invalid_severity",
		}

		err := service.validateVulnerabilityFilters(filters)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的严重级别")
	})

	t.Run("Invalid status rejected", func(t *testing.T) {
		filters := map[string]interface{}{
			"status": "invalid_status",
		}

		err := service.validateVulnerabilityFilters(filters)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无效的状态")
	})

	t.Run("Empty values are allowed", func(t *testing.T) {
		filters := map[string]interface{}{
			"severity": "",
			"status":   "",
		}

		err := service.validateVulnerabilityFilters(filters)
		assert.NoError(t, err)
	})
}

// 测试风险评分算法
func TestOptimizedRiskCalculation(t *testing.T) {
	service := NewScanServiceOptimized(new(MockScanDAOOptimized))

	t.Run("Calculate risk score correctly", func(t *testing.T) {
		vulnStats := map[string]int{
			"critical": 2,  // 2 * 10 = 20
			"high":     3,  // 3 * 7 = 21
			"medium":   5,  // 5 * 4 = 20
			"low":      10, // 10 * 1 = 10
			"info":     0,  // 0 * 0 = 0
		}

		score := service.calculateRiskScore(vulnStats)
		assert.Equal(t, 71.0, score) // 20 + 21 + 20 + 10 = 71
	})

	t.Run("Risk score capped at 100", func(t *testing.T) {
		vulnStats := map[string]int{
			"critical": 20, // 20 * 10 = 200
			"high":     0,
			"medium":   0,
			"low":      0,
		}

		score := service.calculateRiskScore(vulnStats)
		assert.Equal(t, 100.0, score) // Capped at 100
	})

	t.Run("Get correct risk level", func(t *testing.T) {
		testCases := []struct {
			score    float64
			expected string
		}{
			{90, "critical"},
			{70, "high"},
			{40, "medium"},
			{10, "low"},
			{0, "none"},
		}

		for _, tc := range testCases {
			level := service.getRiskLevel(tc.score)
			assert.Equal(t, tc.expected, level, "Score %f should be %s", tc.score, tc.expected)
		}
	})
}

// 测试导入业务逻辑
func TestOptimizedImportScanData(t *testing.T) {
	t.Run("Import successful with valid data", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		project := &models.ProjectOptimized{ID: 1, Name: "Test Project"}
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results: []models.ScanTargetImport{
				{
					Type:    "domain",
					Address: "example.com",
					Ports: []models.PortScanImport{
						{Number: 80, Protocol: "tcp", State: "open"},
					},
				},
			},
		}

		mockDAO.On("GetProjectByID", uint(1)).Return(project, nil)
		mockDAO.On("ImportScanData", data).Return(nil)

		err := service.ImportScanData(data)

		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("Reject import for non-existent project", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		data := &models.ScanDataImport{ProjectID: 999}

		mockDAO.On("GetProjectByID", uint(999)).Return(nil, errors.New("not found"))

		err := service.ImportScanData(data)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "项目不存在")
		mockDAO.AssertExpectations(t)
	})

	t.Run("Reject invalid scan data", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		project := &models.ProjectOptimized{ID: 1}
		data := &models.ScanDataImport{
			ProjectID: 1,
			Results:   []models.ScanTargetImport{}, // Empty results
		}

		mockDAO.On("GetProjectByID", uint(1)).Return(project, nil)

		err := service.ImportScanData(data)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "数据验证失败")
		mockDAO.AssertExpectations(t)
	})
}

// 测试搜索业务逻辑
func TestOptimizedSearch(t *testing.T) {
	t.Run("Search with valid parameters", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		project := &models.ProjectOptimized{ID: 1}
		targets := []models.ScanTarget{
			{Address: "example.com", Type: "domain"},
		}

		mockDAO.On("GetProjectByID", uint(1)).Return(project, nil)
		mockDAO.On("SearchTargets", uint(1), "example").Return(targets, nil)

		results, err := service.SearchTargets(1, "example")

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		mockDAO.AssertExpectations(t)
	})

	t.Run("Reject empty search term", func(t *testing.T) {
		service := NewScanServiceOptimized(new(MockScanDAOOptimized))

		results, err := service.SearchTargets(1, "")

		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Contains(t, err.Error(), "搜索条件不能为空")
	})
}

// 测试项目概览业务逻辑
func TestOptimizedProjectOverview(t *testing.T) {
	t.Run("Generate project overview correctly", func(t *testing.T) {
		mockDAO := new(MockScanDAOOptimized)
		service := NewScanServiceOptimized(mockDAO)

		stats := &models.ProjectStatsOptimized{
			ProjectID:       1,
			ProjectName:     "Test Project",
			TargetCount:     10,
			ServiceCount:    5,
			WebServiceCount: 3,
			VulnerabilityStats: map[string]int{
				"critical": 2,
				"high":     3,
				"medium":   1,
				"low":      0,
			},
			LastScanTime: time.Now(),
		}

		mockDAO.On("GetProjectStatsOptimized", uint(1)).Return(stats, nil)

		overview, err := service.GetProjectOverview(1)

		assert.NoError(t, err)
		assert.NotNil(t, overview)
		assert.Equal(t, "Test Project", overview["project_name"])
		assert.Equal(t, 10, overview["target_count"])
		assert.Equal(t, 6, overview["vulnerability_count"]) // 2+3+1+0
		assert.Equal(t, "medium", overview["risk_level"])   // Score: 2*10+3*7+1*4 = 45
		mockDAO.AssertExpectations(t)
	})
}

// 性能相关的边界测试
func TestOptimizedPerformanceBoundaries(t *testing.T) {
	service := NewScanServiceOptimized(new(MockScanDAOOptimized))

	t.Run("Handle large vulnerability counts", func(t *testing.T) {
		vulnStats := map[string]int{
			"critical": 1000,
			"high":     2000,
			"medium":   5000,
			"low":      10000,
		}

		score := service.calculateRiskScore(vulnStats)
		level := service.getRiskLevel(score)

		assert.Equal(t, 100.0, score)       // Should be capped
		assert.Equal(t, "critical", level)  // Should be critical
	})

	t.Run("Handle zero vulnerabilities", func(t *testing.T) {
		vulnStats := map[string]int{
			"critical": 0,
			"high":     0,
			"medium":   0,
			"low":      0,
		}

		score := service.calculateRiskScore(vulnStats)
		level := service.getRiskLevel(score)

		assert.Equal(t, 0.0, score)
		assert.Equal(t, "none", level)
	})
}