package database

import (
	"cyberedge/pkg/models"
	"gorm.io/gorm"
)

// AutoMigrateScanModels 自动迁移扫描相关的数据表
func AutoMigrateScanModels(db *gorm.DB) error {
	// 按照依赖关系顺序迁移表
	return db.AutoMigrate(
		&models.User{},             // 先迁移用户表
		&models.Project{},
		&models.Domain{},
		&models.Subdomain{},
		&models.IPAddress{},
		&models.Port{},
		&models.Service{},
		&models.WebPath{},
		&models.Technology{},
		&models.Vulnerability{},
		&models.ServiceTechnology{},
	)
}

// CreateIndexes 创建性能优化索引
func CreateIndexes(db *gorm.DB) error {
	// Project表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name)").Error; err != nil {
		return err
	}

	// Domain表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_domains_project_name ON domains(project_id, name)").Error; err != nil {
		return err
	}

	// Subdomain表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_subdomains_domain_name ON subdomains(domain_id, name)").Error; err != nil {
		return err
	}

	// IPAddress表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ip_addresses_address ON ip_addresses(address)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ip_addresses_subdomain ON ip_addresses(subdomain_id)").Error; err != nil {
		return err
	}

	// Port表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_ports_ip_number_protocol ON ports(ip_address_id, number, protocol)").Error; err != nil {
		return err
	}

	// Service表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_services_port ON services(port_id)").Error; err != nil {
		return err
	}
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_services_type ON services(type)").Error; err != nil {
		return err
	}

	// WebPath表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_web_paths_service_path ON web_paths(service_id, path)").Error; err != nil {
		return err
	}

	// Vulnerability表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_vulnerabilities_service ON vulnerabilities(service_id)").Error; err != nil {
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

	// Technology表索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_technologies_name ON technologies(name)").Error; err != nil {
		return err
	}

	return nil
}

// CreateConstraints 创建数据约束（如果数据库支持）
func CreateConstraints(db *gorm.DB) error {
	// 这里可以添加额外的约束，GORM已经会根据模型定义创建基本约束

	// 确保端口号在有效范围内
	if err := db.Exec("ALTER TABLE ports ADD CONSTRAINT check_port_number CHECK (number >= 1 AND number <= 65535)").Error; err != nil {
		// 忽略约束已存在的错误（SQLite不支持IF NOT EXISTS for constraints）
	}

	// 确保漏洞严重级别有效
	if err := db.Exec("ALTER TABLE vulnerabilities ADD CONSTRAINT check_severity CHECK (severity IN ('critical', 'high', 'medium', 'low', 'info'))").Error; err != nil {
		// 忽略约束已存在的错误
	}

	// 确保协议类型有效
	if err := db.Exec("ALTER TABLE ports ADD CONSTRAINT check_protocol CHECK (protocol IN ('tcp', 'udp', 'sctp'))").Error; err != nil {
		// 忽略约束已存在的错误
	}

	return nil
}