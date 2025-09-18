package dao

import "cyberedge/pkg/models"

// UserDAOInterface 用户DAO接口
type UserDAOInterface interface {
	Create(user *models.User) error
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	GetByID(id uint) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	GetAll() ([]*models.User, error)
}

// ScanDAOInterface 扫描数据DAO接口
type ScanDAOInterface interface {
	// Project 管理
	CreateProject(project *models.Project) error
	GetProjectByID(id uint) (*models.Project, error)
	GetProjectByName(name string) (*models.Project, error)
	ListProjects() ([]models.Project, error)
	DeleteProject(id uint) error

	// 项目详情和统计
	GetProjectDetails(projectID uint) (*models.ProjectStats, []models.ScanTarget, error)
	GetProjectStats(projectID uint) (*models.ProjectStats, error)

	// 扫描数据导入
	ImportScanData(data *models.ScanDataImport) error

	// 查询功能
	GetVulnerabilities(projectID uint, filters map[string]interface{}) ([]models.Vulnerability, error)
	GetVulnerabilityStats(projectID uint) (map[string]int, error)
	GetProjectHierarchy(projectID uint) ([]models.ScanTarget, error)
	SearchTargets(projectID uint, searchTerm string) ([]models.ScanTarget, error)
}