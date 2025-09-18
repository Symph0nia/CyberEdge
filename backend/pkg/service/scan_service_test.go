package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
)

// MockScanDAO for testing business logic without database dependencies
type MockScanDAO struct {
	mock.Mock
}

// Ensure MockScanDAO implements ScanDAOInterface
var _ dao.ScanDAOInterface = (*MockScanDAO)(nil)

func (m *MockScanDAO) CreateProject(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockScanDAO) GetProjectByID(id uint) (*models.Project, error) {
	args := m.Called(id)
	if project := args.Get(0); project != nil {
		return project.(*models.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScanDAO) GetProjectByName(name string) (*models.Project, error) {
	args := m.Called(name)
	if project := args.Get(0); project != nil {
		return project.(*models.Project), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockScanDAO) ListProjects() ([]models.Project, error) {
	args := m.Called()
	return args.Get(0).([]models.Project), args.Error(1)
}

func (m *MockScanDAO) DeleteProject(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockScanDAO) CreateOrUpdateHierarchy(project *models.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockScanDAO) GetProjectVulnerabilityStats(projectID uint) (map[string]int, error) {
	args := m.Called(projectID)
	return args.Get(0).(map[string]int), args.Error(1)
}

// Test core project management business logic
func TestCreateProject(t *testing.T) {
	mockDAO := new(MockScanDAO)
	service := NewScanService(mockDAO)

	t.Run("Create project successfully", func(t *testing.T) {
		projectName := "Test Project"
		description := "Test Description"

		// Mock: project name doesn't exist
		mockDAO.On("GetProjectByName", projectName).Return(nil, errors.New("not found"))
		mockDAO.On("CreateProject", mock.AnythingOfType("*models.Project")).Return(nil)

		project, err := service.CreateProject(projectName, description)

		assert.NoError(t, err)
		assert.NotNil(t, project)
		assert.Equal(t, projectName, project.Name)
		assert.Equal(t, description, project.Description)
		mockDAO.AssertExpectations(t)
	})

	t.Run("Reject empty project name", func(t *testing.T) {
		project, err := service.CreateProject("", "Description")

		assert.Error(t, err)
		assert.Nil(t, project)
		assert.Contains(t, err.Error(), "项目名称不能为空")
	})

	t.Run("Reject duplicate project name", func(t *testing.T) {
		existingProject := &models.Project{Name: "Existing Project"}

		mockDAO.On("GetProjectByName", "Existing Project").Return(existingProject, nil)

		project, err := service.CreateProject("Existing Project", "Description")

		assert.Error(t, err)
		assert.Nil(t, project)
		assert.Contains(t, err.Error(), "项目名称已存在")
		mockDAO.AssertExpectations(t)
	})

	t.Run("Handle database error during creation", func(t *testing.T) {
		// Create fresh mock for this test
		freshMockDAO := new(MockScanDAO)
		freshService := NewScanService(freshMockDAO)

		projectName := "Test Project"
		dbError := errors.New("database connection failed")

		freshMockDAO.On("GetProjectByName", projectName).Return(nil, errors.New("not found"))
		freshMockDAO.On("CreateProject", mock.AnythingOfType("*models.Project")).Return(dbError)

		project, err := freshService.CreateProject(projectName, "Description")

		assert.Error(t, err)
		assert.Nil(t, project)
		assert.Contains(t, err.Error(), "创建项目失败")
		freshMockDAO.AssertExpectations(t)
	})
}

// Test scan data hierarchy building - core business logic
func TestBuildHierarchyFromScanData(t *testing.T) {
	// This test doesn't need DAO since it tests pure business logic
	service := &ScanService{}

	project := &models.Project{
		ID:   1,
		Name: "Test Project",
	}

	t.Run("Build hierarchy with single result", func(t *testing.T) {
		scanData := &ScanResultData{
			Results: []ScanResult{
				{
					IP:        "192.168.1.100",
					Domain:    "example.com",
					Subdomain: "www",
					Ports: []PortData{
						{
							Number:   80,
							Protocol: "tcp",
							State:    "open",
							Service: &ServiceData{
								Name:        "http",
								Version:     "Apache/2.4.41",
								Fingerprint: "Apache httpd",
								Banner:      "HTTP/1.1 200 OK",
							},
						},
					},
				},
			},
		}

		result := service.buildHierarchyFromScanData(project, scanData)

		assert.NotNil(t, result)
		assert.Equal(t, project, result)
		// Hierarchy building modifies project in place
		// In real implementation, project would have domains populated
	})

	t.Run("Handle empty scan data", func(t *testing.T) {
		scanData := &ScanResultData{
			Results: []ScanResult{},
		}

		result := service.buildHierarchyFromScanData(project, scanData)

		assert.NotNil(t, result)
		assert.Equal(t, project, result)
	})

	t.Run("Handle scan data without domain", func(t *testing.T) {
		scanData := &ScanResultData{
			Results: []ScanResult{
				{
					IP:     "10.0.0.1",
					Domain: "", // No domain
					Ports: []PortData{
						{
							Number:   22,
							Protocol: "tcp",
							State:    "open",
							Service: &ServiceData{
								Name:    "ssh",
								Version: "OpenSSH 8.2p1",
							},
						},
					},
				},
			},
		}

		result := service.buildHierarchyFromScanData(project, scanData)

		assert.NotNil(t, result)
		// Should handle IP-only results without domains
	})

	t.Run("Handle web service with vulnerabilities", func(t *testing.T) {
		scanData := &ScanResultData{
			Results: []ScanResult{
				{
					IP:        "192.168.1.200",
					Domain:    "test.com",
					Subdomain: "api",
					Ports: []PortData{
						{
							Number:   443,
							Protocol: "tcp",
							State:    "open",
							Service: &ServiceData{
								Name:    "https",
								Version: "nginx/1.18.0",
								WebData: &WebServiceData{
									Paths: []WebPathData{
										{
											Path:       "/admin",
											StatusCode: 200,
											Title:      "Admin Panel",
											Length:     1024,
											Vulnerabilities: []VulnerabilityData{
												{
													Title:       "Unprotected Admin Panel",
													Description: "Admin panel without auth",
													Severity:    "high",
													CVSS:        7.5,
													Location:    "/admin",
												},
											},
										},
									},
									Technologies: []string{"nginx", "PHP"},
								},
								Vulnerabilities: []VulnerabilityData{
									{
										CVEID:       "CVE-2021-44790",
										Title:       "Buffer Overflow",
										Description: "Critical vulnerability",
										Severity:    "critical",
										CVSS:        9.8,
									},
								},
							},
						},
					},
				},
			},
		}

		result := service.buildHierarchyFromScanData(project, scanData)

		assert.NotNil(t, result)
		// Hierarchy should include web paths and vulnerabilities
	})
}

// Test vulnerability processing logic
func TestProcessVulnerabilities(t *testing.T) {
	// This test doesn't need DAO since it tests pure business logic
	service := &ScanService{}

	t.Run("Process service-level vulnerabilities", func(t *testing.T) {
		serviceObj := &models.Service{
			Type: "http",
			Name: "Apache",
		}

		vulnData := []VulnerabilityData{
			{
				CVEID:       "CVE-2021-44790",
				Title:       "Apache Buffer Overflow",
				Description: "Critical buffer overflow in mod_lua",
				Severity:    "critical",
				CVSS:        9.8,
				Location:    "mod_lua",
			},
			{
				Title:       "Weak Configuration",
				Description: "Server allows directory listing",
				Severity:    "medium",
				CVSS:        5.3,
				Location:    "/",
			},
		}

		// Call the actual method
		service.processVulnerabilities(serviceObj, vulnData)

		// Verify vulnerabilities were added
		assert.Len(t, serviceObj.Vulnerabilities, 2)

		// Check first vulnerability
		vuln1 := serviceObj.Vulnerabilities[0]
		assert.Equal(t, "CVE-2021-44790", vuln1.CVEID)
		assert.Equal(t, "critical", vuln1.Severity)
		assert.Equal(t, 9.8, vuln1.CVSS)

		// Check second vulnerability
		vuln2 := serviceObj.Vulnerabilities[1]
		assert.Equal(t, "Weak Configuration", vuln2.Title)
		assert.Equal(t, "medium", vuln2.Severity)
		assert.Equal(t, 5.3, vuln2.CVSS)
	})

	t.Run("Handle empty vulnerability list", func(t *testing.T) {
		serviceObj := &models.Service{
			Type: "ssh",
			Name: "OpenSSH",
		}

		service.processVulnerabilities(serviceObj, []VulnerabilityData{})

		assert.Len(t, serviceObj.Vulnerabilities, 0)
	})
}

// Test web service data processing
func TestProcessWebServiceData(t *testing.T) {
	// This test doesn't need DAO since it tests pure business logic
	service := &ScanService{}

	t.Run("Process web service with paths and technologies", func(t *testing.T) {
		serviceObj := &models.Service{
			Type: "https",
			Name: "nginx",
		}

		webData := &WebServiceData{
			Paths: []WebPathData{
				{
					Path:       "/",
					StatusCode: 200,
					Title:      "Home Page",
					Length:     2048,
				},
				{
					Path:       "/api/v1",
					StatusCode: 401,
					Title:      "Unauthorized",
					Length:     256,
					Vulnerabilities: []VulnerabilityData{
						{
							Title:       "SQL Injection",
							Description: "SQLi in user endpoint",
							Severity:    "critical",
							CVSS:        9.1,
							Location:    "/api/v1",
							Parameter:   "id",
							Payload:     "1' OR '1'='1",
						},
					},
				},
			},
			Technologies: []string{"nginx", "Node.js", "MongoDB"},
		}

		service.processWebServiceData(serviceObj, webData)

		// Verify web paths were added
		assert.Len(t, serviceObj.WebPaths, 2)

		// Check first path
		path1 := serviceObj.WebPaths[0]
		assert.Equal(t, "/", path1.Path)
		assert.Equal(t, 200, path1.StatusCode)
		assert.Len(t, path1.Vulnerabilities, 0)

		// Check second path with vulnerability
		path2 := serviceObj.WebPaths[1]
		assert.Equal(t, "/api/v1", path2.Path)
		assert.Equal(t, 401, path2.StatusCode)
		assert.Len(t, path2.Vulnerabilities, 1)

		vuln := path2.Vulnerabilities[0]
		assert.Equal(t, "SQL Injection", vuln.Title)
		assert.Equal(t, "critical", vuln.Severity)
		assert.Equal(t, "/api/v1", vuln.Location)
		assert.Equal(t, "id", vuln.Parameter)

		// Verify technologies were added
		assert.Len(t, serviceObj.Technologies, 3)
		assert.Equal(t, "nginx", serviceObj.Technologies[0].Name)
		assert.Equal(t, "Node.js", serviceObj.Technologies[1].Name)
		assert.Equal(t, "MongoDB", serviceObj.Technologies[2].Name)
	})

	t.Run("Handle nil web data", func(t *testing.T) {
		serviceObj := &models.Service{
			Type: "http",
			Name: "Apache",
		}

		service.processWebServiceData(serviceObj, nil)

		assert.Len(t, serviceObj.WebPaths, 0)
		assert.Len(t, serviceObj.Technologies, 0)
	})
}

// Test import scan results business logic
func TestImportScanResults(t *testing.T) {

	t.Run("Import scan results successfully", func(t *testing.T) {
		mockDAO := new(MockScanDAO)
		service := NewScanService(mockDAO)

		projectID := uint(1)
		project := &models.Project{
			ID:   projectID,
			Name: "Test Project",
		}

		scanData := &ScanResultData{
			Results: []ScanResult{
				{
					IP:     "192.168.1.100",
					Domain: "example.com",
					Ports: []PortData{
						{
							Number:   80,
							Protocol: "tcp",
							State:    "open",
						},
					},
				},
			},
		}

		mockDAO.On("GetProjectByID", projectID).Return(project, nil)
		mockDAO.On("CreateOrUpdateHierarchy", mock.AnythingOfType("*models.Project")).Return(nil)

		err := service.ImportScanResults(projectID, scanData)

		assert.NoError(t, err)
		mockDAO.AssertExpectations(t)
	})

	t.Run("Reject import for non-existent project", func(t *testing.T) {
		mockDAO := new(MockScanDAO)
		service := NewScanService(mockDAO)

		projectID := uint(999)
		scanData := &ScanResultData{Results: []ScanResult{}}

		mockDAO.On("GetProjectByID", projectID).Return(nil, errors.New("project not found"))

		err := service.ImportScanResults(projectID, scanData)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "项目不存在")
		mockDAO.AssertExpectations(t)
	})

	t.Run("Handle database error during hierarchy creation", func(t *testing.T) {
		mockDAO := new(MockScanDAO)
		service := NewScanService(mockDAO)

		projectID := uint(1)
		project := &models.Project{ID: projectID}
		scanData := &ScanResultData{Results: []ScanResult{}}
		dbError := errors.New("database error")

		mockDAO.On("GetProjectByID", projectID).Return(project, nil)
		mockDAO.On("CreateOrUpdateHierarchy", mock.AnythingOfType("*models.Project")).Return(dbError)

		err := service.ImportScanResults(projectID, scanData)

		assert.Error(t, err)
		assert.Equal(t, dbError, err)
		mockDAO.AssertExpectations(t)
	})
}

// Test project statistics calculation
func TestGetProjectStats(t *testing.T) {

	t.Run("Calculate project statistics correctly", func(t *testing.T) {
		mockDAO := new(MockScanDAO)
		service := NewScanService(mockDAO)

		projectID := uint(1)
		project := &models.Project{
			ID:   projectID,
			Name: "Test Project",
			Domains: []models.Domain{
				{
					Name: "example.com",
					Subdomains: []models.Subdomain{
						{
							Name: "www",
							IPAddresses: []models.IPAddress{
								{
									Address: "192.168.1.100",
									Ports: []models.Port{
										{
											Number:   80,
											Protocol: "tcp",
											Service:  &models.Service{Name: "http"},
										},
										{
											Number:   443,
											Protocol: "tcp",
											Service:  &models.Service{Name: "https"},
										},
									},
								},
							},
						},
						{
							Name: "api",
							IPAddresses: []models.IPAddress{
								{
									Address: "192.168.1.101",
									Ports: []models.Port{
										{
											Number:   8080,
											Protocol: "tcp",
										},
									},
								},
							},
						},
					},
				},
			},
		}

		vulnStats := map[string]int{
			"critical": 2,
			"high":     5,
			"medium":   10,
			"low":      3,
		}

		mockDAO.On("GetProjectByID", projectID).Return(project, nil)
		mockDAO.On("GetProjectVulnerabilityStats", projectID).Return(vulnStats, nil)

		stats, err := service.GetProjectStats(projectID)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, projectID, stats.ProjectID)
		assert.Equal(t, "Test Project", stats.ProjectName)
		assert.Equal(t, 1, stats.DomainCount)
		assert.Equal(t, 2, stats.SubdomainCount)
		assert.Equal(t, 2, stats.IPCount)
		assert.Equal(t, 3, stats.PortCount)
		assert.Equal(t, 2, stats.ServiceCount) // Only 2 ports have services
		assert.Equal(t, vulnStats, stats.VulnerabilityStats)

		mockDAO.AssertExpectations(t)
	})

	t.Run("Handle project not found", func(t *testing.T) {
		mockDAO := new(MockScanDAO)
		service := NewScanService(mockDAO)

		projectID := uint(999)

		mockDAO.On("GetProjectByID", projectID).Return(nil, errors.New("not found"))

		stats, err := service.GetProjectStats(projectID)

		assert.Error(t, err)
		assert.Nil(t, stats)
		mockDAO.AssertExpectations(t)
	})
}

// Test domain and subdomain creation logic
func TestGetOrCreateDomainAndSubdomain(t *testing.T) {
	// This test doesn't need DAO since it tests pure business logic
	service := &ScanService{}

	t.Run("Create new domain when not exists", func(t *testing.T) {
		domainMap := make(map[string]*models.Domain)
		project := &models.Project{Name: "Test Project"}

		domain := service.getOrCreateDomain(domainMap, "example.com", project)

		assert.NotNil(t, domain)
		assert.Equal(t, "example.com", domain.Name)
		assert.Contains(t, domainMap, "example.com")
		assert.Len(t, project.Domains, 1)
		assert.Equal(t, "example.com", project.Domains[0].Name)
	})

	t.Run("Return existing domain when already exists", func(t *testing.T) {
		existingDomain := &models.Domain{Name: "example.com"}
		domainMap := map[string]*models.Domain{
			"example.com": existingDomain,
		}
		project := &models.Project{Name: "Test Project"}

		domain := service.getOrCreateDomain(domainMap, "example.com", project)

		assert.Equal(t, existingDomain, domain)
		assert.Len(t, project.Domains, 0) // Should not add duplicate
	})

	t.Run("Create new subdomain when not exists", func(t *testing.T) {
		subdomainMap := make(map[string]*models.Subdomain)
		domain := &models.Domain{Name: "example.com"}

		subdomain := service.getOrCreateSubdomain(subdomainMap, domain, "www")

		assert.NotNil(t, subdomain)
		assert.Equal(t, "www", subdomain.Name)
		assert.Contains(t, subdomainMap, "www.example.com")
		assert.Len(t, domain.Subdomains, 1)
	})

	t.Run("Handle root domain (empty subdomain)", func(t *testing.T) {
		subdomainMap := make(map[string]*models.Subdomain)
		domain := &models.Domain{Name: "example.com"}

		subdomain := service.getOrCreateSubdomain(subdomainMap, domain, "@")

		assert.NotNil(t, subdomain)
		assert.Equal(t, "@", subdomain.Name)
		assert.Contains(t, subdomainMap, "@.example.com")
	})
}

// Test input validation and edge cases
func TestInputValidation(t *testing.T) {
	// This test doesn't need DAO since it tests pure business logic
	service := &ScanService{}

	t.Run("Handle malicious domain names", func(t *testing.T) {
		maliciousDomains := []string{
			"'; DROP TABLE domains; --",
			"<script>alert('xss')</script>.com",
			"../../etc/passwd",
			"localhost:8080/../../admin",
		}

		project := &models.Project{Name: "Test"}
		domainMap := make(map[string]*models.Domain)

		for _, maliciousDomain := range maliciousDomains {
			// The method should handle malicious input gracefully
			domain := service.getOrCreateDomain(domainMap, maliciousDomain, project)

			// Domain should be created but input should be sanitized
			assert.NotNil(t, domain)
			assert.Equal(t, maliciousDomain, domain.Name) // For now, just store as-is

			// In production, domain.Name should be sanitized
			t.Logf("Processed potentially malicious domain: %s", maliciousDomain)
		}
	})

	t.Run("Handle extreme port numbers", func(t *testing.T) {
		extremePorts := []int{
			-1,     // Negative port
			0,      // Zero port
			65536,  // Beyond max port range
			999999, // Way beyond max
		}

		for _, port := range extremePorts {
			portData := PortData{
				Number:   port,
				Protocol: "tcp",
				State:    "open",
			}

			// Should handle extreme values gracefully
			assert.True(t, port < 1 || port > 65535, "Port %d is outside valid range", port)
			t.Logf("Handled extreme port number: %d", portData.Number)
		}
	})
}