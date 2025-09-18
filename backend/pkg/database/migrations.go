package database

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// AutoMigrateScanModels 自动迁移扫描相关的数据表
func AutoMigrateScanModels(db *gorm.DB) error {
	// 按照依赖关系顺序迁移表 - 使用简化的模型结构
	return db.AutoMigrate(
		&models.User{},                   // 先迁移用户表
		&models.Project{},                // 项目表
		&models.ScanTarget{},             // 扫描目标（域名、子域名、IP统一）
		&models.ScanResult{},             // 扫描结果（端口+服务）
		&models.WebPath{},                // Web路径
		&models.Technology{},             // 技术栈
		&models.Vulnerability{},          // 漏洞信息
		&models.ScanResultTechnology{},   // 扫描结果与技术栈关联表
	)
}

// CreateIndexes 创建性能优化索引
func CreateIndexes(db *gorm.DB) error {
	// Project表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name)").Error; err != nil {
		return err
	}

	// ScanTarget表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_targets_project_type ON scan_targets(project_id, type)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_targets_address ON scan_targets(address)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_targets_parent ON scan_targets(parent_id)").Error; err != nil {
		return err
	}

	// ScanResult表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_results_target_port_protocol ON scan_results(target_id, port, protocol)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_results_project ON scan_results(project_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_results_service ON scan_results(service_name)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_results_web_service ON scan_results(is_web_service)").Error; err != nil {
		return err
	}

	// WebPath表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_web_paths_scan_result_path ON web_paths(scan_result_id, path)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_web_paths_status_code ON web_paths(status_code)").Error; err != nil {
		return err
	}

	// Vulnerability表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_scan_result ON vulnerabilities(scan_result_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_web_path ON vulnerabilities(web_path_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_severity ON vulnerabilities(severity)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_cve ON vulnerabilities(cve_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_status ON vulnerabilities(status)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_cvss ON vulnerabilities(cvss)").Error; err != nil {
		return err
	}

	// Technology表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_technologies_name ON technologies(name)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_technologies_category ON technologies(category)").Error; err != nil {
		return err
	}

	// ScanResultTechnology表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_result_tech_result ON scan_result_technologies(scan_result_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_scan_result_tech_tech ON scan_result_technologies(technology_id)").Error; err != nil {
		return err
	}

	return nil
}

// CreateConstraints 创建数据约束（如果数据库支持）
func CreateConstraints(db *gorm.DB) error {
	// 这里可以添加额外的约束，GORM已经会根据模型定义创建基本约束

	// 确保端口号在有效范围内
	if err := db.Exec("ALTER TABLE scan_results ADD CONSTRAINT check_port_number CHECK (port >= 1 AND port <= 65535)").Error; err != nil {
		// 忽略约束已存在的错误（SQLite不支持IF NOT EXISTS for constraints）
	}

	// 确保漏洞严重级别有效
	if err := db.Exec("ALTER TABLE vulnerabilities ADD CONSTRAINT check_severity CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info'))").Error; err != nil {
		// 忽略约束已存在的错误
	}

	// 确保协议类型有效
	if err := db.Exec("ALTER TABLE scan_results ADD CONSTRAINT check_protocol CHECK (protocol IN ('tcp', 'udp', 'sctp'))").Error; err != nil {
		// 忽略约束已存在的错误
	}

	// 确保扫描目标类型有效
	if err := db.Exec("ALTER TABLE scan_targets ADD CONSTRAINT check_target_type CHECK (type IN ('domain', 'subdomain', 'ip'))").Error; err != nil {
		// 忽略约束已存在的错误
	}

	// 确保漏洞状态有效
	if err := db.Exec("ALTER TABLE vulnerabilities ADD CONSTRAINT check_status CHECK (status IN ('open', 'fixed', 'false_positive'))").Error; err != nil {
		// 忽略约束已存在的错误
	}

	return nil
}