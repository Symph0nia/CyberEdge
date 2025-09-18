-- CyberEdge Database Migration Script
-- 从基础用户表迁移到完整扫描功能schema
-- 版本: v1.0 -> v2.0 (完整扫描功能支持)

-- 检查当前数据库版本
-- 如果migrations表不存在，则创建它
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    description TEXT
);

-- 检查是否已经应用了此迁移
SELECT COUNT(*) as already_applied FROM schema_migrations WHERE version = '20240918_complete_scan_schema';

-- 如果已经应用过，跳过迁移
-- (这部分需要在应用逻辑中处理，SQL不支持复杂条件控制)

-- =============================================================================
-- 开始迁移：添加扫描功能相关表
-- =============================================================================

-- 1. 项目表 (ProjectOptimized)
CREATE TABLE IF NOT EXISTS project_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    INDEX idx_name (name),
    INDEX idx_deleted_at (deleted_at),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB;

-- 2. 扫描目标表 (ScanTarget)
CREATE TABLE IF NOT EXISTS scan_targets (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    type VARCHAR(20) NOT NULL,
    address VARCHAR(255) NOT NULL,
    parent_id INT UNSIGNED NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_project_id (project_id),
    INDEX idx_type (type),
    INDEX idx_address (address),
    INDEX idx_parent_id (parent_id),
    UNIQUE KEY unique_project_address (project_id, address)
) ENGINE=InnoDB;

-- 3. 扫描结果表 (ScanResultOptimized)
CREATE TABLE IF NOT EXISTS scan_result_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    target_id INT UNSIGNED NOT NULL,
    port INT NOT NULL,
    protocol VARCHAR(10) NOT NULL,
    state VARCHAR(20),
    service_name VARCHAR(50),
    version VARCHAR(100),
    fingerprint VARCHAR(255),
    banner TEXT,
    is_web_service BOOLEAN DEFAULT FALSE,
    http_title VARCHAR(255),
    http_status INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_project_id (project_id),
    INDEX idx_target_id (target_id),
    INDEX idx_port (port),
    INDEX idx_service_name (service_name),
    INDEX idx_is_web_service (is_web_service),
    UNIQUE KEY unique_target_port_protocol (target_id, port, protocol)
) ENGINE=InnoDB;

-- 4. Web路径表 (WebPathOptimized)
CREATE TABLE IF NOT EXISTS web_path_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    scan_result_id INT UNSIGNED NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INT,
    title VARCHAR(255),
    length INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_scan_result_id (scan_result_id),
    INDEX idx_path (path(255)),
    INDEX idx_status_code (status_code),
    UNIQUE KEY unique_scan_result_path (scan_result_id, path(255))
) ENGINE=InnoDB;

-- 5. 漏洞表 (VulnerabilityOptimized)
CREATE TABLE IF NOT EXISTS vulnerability_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    scan_result_id INT UNSIGNED NOT NULL,
    web_path_id INT UNSIGNED NULL,
    cve_id VARCHAR(50),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    cvss DECIMAL(3,1),
    location VARCHAR(255),
    parameter VARCHAR(100),
    payload TEXT,
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_scan_result_id (scan_result_id),
    INDEX idx_web_path_id (web_path_id),
    INDEX idx_cve_id (cve_id),
    INDEX idx_severity (severity),
    INDEX idx_cvss (cvss),
    INDEX idx_status (status)
) ENGINE=InnoDB;

-- 6. 技术栈表 (TechnologyOptimized)
CREATE TABLE IF NOT EXISTS technology_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_name (name),
    INDEX idx_category (category)
) ENGINE=InnoDB;

-- 7. 扫描结果与技术栈关联表
CREATE TABLE IF NOT EXISTS scan_result_technologies (
    scan_result_id INT UNSIGNED NOT NULL,
    technology_id INT UNSIGNED NOT NULL,
    version VARCHAR(100),

    PRIMARY KEY (scan_result_id, technology_id)
) ENGINE=InnoDB;

-- 8. 扫描框架结果表
CREATE TABLE IF NOT EXISTS scan_framework_results (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    scan_target_id INT UNSIGNED NOT NULL,
    target VARCHAR(255) NOT NULL,
    scan_type VARCHAR(50) NOT NULL,
    scanner_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL,
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP NULL,
    raw_data TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_project_id (project_id),
    INDEX idx_scan_target_id (scan_target_id),
    INDEX idx_scan_type (scan_type),
    INDEX idx_status (status),
    INDEX idx_start_time (start_time)
) ENGINE=InnoDB;

-- 9. 扫描框架目标表
CREATE TABLE IF NOT EXISTS scan_framework_targets (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    target VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_project_id (project_id),
    INDEX idx_target (target),
    UNIQUE KEY unique_project_target (project_id, target)
) ENGINE=InnoDB;

-- =============================================================================
-- 添加外键约束 (在所有表创建完成后添加)
-- =============================================================================

-- 检查并添加外键约束 (避免重复添加)
SET @exist_fk = 0;

-- scan_targets外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_scan_targets_project'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_targets ADD CONSTRAINT fk_scan_targets_project FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_scan_targets_project already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- scan_targets自引用外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_scan_targets_parent'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_targets ADD CONSTRAINT fk_scan_targets_parent FOREIGN KEY (parent_id) REFERENCES scan_targets(id) ON DELETE CASCADE',
    'SELECT "FK fk_scan_targets_parent already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- scan_result_optimizeds外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_scan_results_project'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_result_optimizeds ADD CONSTRAINT fk_scan_results_project FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_scan_results_project already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_scan_results_target'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_result_optimizeds ADD CONSTRAINT fk_scan_results_target FOREIGN KEY (target_id) REFERENCES scan_targets(id) ON DELETE CASCADE',
    'SELECT "FK fk_scan_results_target already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- web_path_optimizeds外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_web_paths_scan_result'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE web_path_optimizeds ADD CONSTRAINT fk_web_paths_scan_result FOREIGN KEY (scan_result_id) REFERENCES scan_result_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_web_paths_scan_result already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- vulnerability_optimizeds外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_vulnerabilities_scan_result'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE vulnerability_optimizeds ADD CONSTRAINT fk_vulnerabilities_scan_result FOREIGN KEY (scan_result_id) REFERENCES scan_result_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_vulnerabilities_scan_result already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_vulnerabilities_web_path'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE vulnerability_optimizeds ADD CONSTRAINT fk_vulnerabilities_web_path FOREIGN KEY (web_path_id) REFERENCES web_path_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_vulnerabilities_web_path already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- scan_result_technologies外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_srt_scan_result'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_result_technologies ADD CONSTRAINT fk_srt_scan_result FOREIGN KEY (scan_result_id) REFERENCES scan_result_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_srt_scan_result already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_srt_technology'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_result_technologies ADD CONSTRAINT fk_srt_technology FOREIGN KEY (technology_id) REFERENCES technology_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_srt_technology already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- scan_framework_results外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_sfr_project'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_framework_results ADD CONSTRAINT fk_sfr_project FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_sfr_project already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_sfr_scan_target'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_framework_results ADD CONSTRAINT fk_sfr_scan_target FOREIGN KEY (scan_target_id) REFERENCES scan_targets(id) ON DELETE CASCADE',
    'SELECT "FK fk_sfr_scan_target already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- scan_framework_targets外键
SELECT COUNT(*) INTO @exist_fk
FROM information_schema.TABLE_CONSTRAINTS
WHERE CONSTRAINT_NAME = 'fk_sft_project'
AND TABLE_SCHEMA = DATABASE();

SET @sql = IF(@exist_fk = 0,
    'ALTER TABLE scan_framework_targets ADD CONSTRAINT fk_sft_project FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE',
    'SELECT "FK fk_sft_project already exists"');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- =============================================================================
-- 创建视图
-- =============================================================================

-- 删除现有视图（如果存在）
DROP VIEW IF EXISTS v_project_stats;
DROP VIEW IF EXISTS v_vulnerability_overview;

-- 项目统计视图
CREATE VIEW v_project_stats AS
SELECT
    p.id as project_id,
    p.name as project_name,
    COUNT(DISTINCT st.id) as target_count,
    COUNT(DISTINCT CASE WHEN st.type = 'domain' THEN st.id END) as domain_count,
    COUNT(DISTINCT CASE WHEN st.type = 'ip' THEN st.id END) as ip_count,
    COUNT(DISTINCT sr.id) as port_count,
    COUNT(DISTINCT CASE WHEN sr.service_name IS NOT NULL THEN sr.id END) as service_count,
    COUNT(DISTINCT CASE WHEN sr.is_web_service = 1 THEN sr.id END) as web_service_count,
    COUNT(DISTINCT v.id) as vulnerability_count,
    COUNT(DISTINCT CASE WHEN v.severity = 'critical' THEN v.id END) as critical_count,
    COUNT(DISTINCT CASE WHEN v.severity = 'high' THEN v.id END) as high_count,
    MAX(sr.created_at) as last_scan_time
FROM project_optimizeds p
LEFT JOIN scan_targets st ON p.id = st.project_id
LEFT JOIN scan_result_optimizeds sr ON st.id = sr.target_id
LEFT JOIN vulnerability_optimizeds v ON sr.id = v.scan_result_id
WHERE p.deleted_at IS NULL
GROUP BY p.id, p.name;

-- 漏洞概览视图
CREATE VIEW v_vulnerability_overview AS
SELECT
    v.id,
    v.title,
    v.severity,
    v.cvss,
    v.status,
    p.name as project_name,
    st.address as target_address,
    sr.port,
    sr.service_name,
    v.created_at
FROM vulnerability_optimizeds v
JOIN scan_result_optimizeds sr ON v.scan_result_id = sr.id
JOIN scan_targets st ON sr.target_id = st.id
JOIN project_optimizeds p ON st.project_id = p.id
WHERE p.deleted_at IS NULL
ORDER BY
    CASE v.severity
        WHEN 'critical' THEN 1
        WHEN 'high' THEN 2
        WHEN 'medium' THEN 3
        WHEN 'low' THEN 4
        ELSE 5
    END,
    v.cvss DESC,
    v.created_at DESC;

-- =============================================================================
-- 记录迁移完成
-- =============================================================================

INSERT INTO schema_migrations (version, description)
VALUES ('20240918_complete_scan_schema', 'Added complete scanning functionality tables and views')
ON DUPLICATE KEY UPDATE applied_at = CURRENT_TIMESTAMP;

-- 输出迁移完成信息
SELECT 'Migration completed successfully! Added scanning functionality to CyberEdge database.' as migration_status;