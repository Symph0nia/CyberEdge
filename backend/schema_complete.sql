-- CyberEdge Complete MySQL Schema
-- 基于优化后的数据模型，支持完整扫描功能

-- 创建数据库
CREATE DATABASE IF NOT EXISTS cyberedge CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cyberedge;

-- 用户表 (保持现有结构)
CREATE TABLE users (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    is_2fa_enabled BOOLEAN DEFAULT FALSE,
    totp_secret VARCHAR(32),
    role ENUM('admin', 'user') DEFAULT 'user',
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,

    INDEX idx_username (username),
    INDEX idx_email (email),
    INDEX idx_role (role),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB;

-- 项目表 (ProjectOptimized)
CREATE TABLE project_optimizeds (
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

-- 扫描目标表 (ScanTarget) - 合并域名、子域名、IP概念
CREATE TABLE scan_targets (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    type VARCHAR(20) NOT NULL,           -- "domain", "subdomain", "ip"
    address VARCHAR(255) NOT NULL,       -- 域名、子域名或IP地址
    parent_id INT UNSIGNED NULL,         -- 父级目标ID，用于层次关系
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES scan_targets(id) ON DELETE CASCADE,

    INDEX idx_project_id (project_id),
    INDEX idx_type (type),
    INDEX idx_address (address),
    INDEX idx_parent_id (parent_id),
    UNIQUE KEY unique_project_address (project_id, address)
) ENGINE=InnoDB;

-- 扫描结果表 (ScanResultOptimized) - 端口+服务
CREATE TABLE scan_result_optimizeds (
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

    -- Web服务特有字段
    is_web_service BOOLEAN DEFAULT FALSE,
    http_title VARCHAR(255),
    http_status INT,

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE,
    FOREIGN KEY (target_id) REFERENCES scan_targets(id) ON DELETE CASCADE,

    INDEX idx_project_id (project_id),
    INDEX idx_target_id (target_id),
    INDEX idx_port (port),
    INDEX idx_service_name (service_name),
    INDEX idx_is_web_service (is_web_service),
    UNIQUE KEY unique_target_port_protocol (target_id, port, protocol)
) ENGINE=InnoDB;

-- 漏洞表 (VulnerabilityOptimized)
CREATE TABLE vulnerability_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    scan_result_id INT UNSIGNED NOT NULL,
    web_path_id INT UNSIGNED NULL,       -- 可选，路径级漏洞
    cve_id VARCHAR(50),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,       -- critical, high, medium, low, info
    cvss DECIMAL(3,1),                   -- CVSS分数，如9.8
    location VARCHAR(255),
    parameter VARCHAR(100),
    payload TEXT,
    status VARCHAR(20) DEFAULT 'open',   -- open, fixed, false_positive
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (scan_result_id) REFERENCES scan_result_optimizeds(id) ON DELETE CASCADE,

    INDEX idx_scan_result_id (scan_result_id),
    INDEX idx_web_path_id (web_path_id),
    INDEX idx_cve_id (cve_id),
    INDEX idx_severity (severity),
    INDEX idx_cvss (cvss),
    INDEX idx_status (status)
) ENGINE=InnoDB;

-- Web路径表 (WebPathOptimized) - 仅针对Web服务
CREATE TABLE web_path_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    scan_result_id INT UNSIGNED NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INT,
    title VARCHAR(255),
    length INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (scan_result_id) REFERENCES scan_result_optimizeds(id) ON DELETE CASCADE,

    INDEX idx_scan_result_id (scan_result_id),
    INDEX idx_path (path(255)),
    INDEX idx_status_code (status_code),
    UNIQUE KEY unique_scan_result_path (scan_result_id, path(255))
) ENGINE=InnoDB;

-- 添加web_path_id外键约束到漏洞表
ALTER TABLE vulnerability_optimizeds
ADD FOREIGN KEY (web_path_id) REFERENCES web_path_optimizeds(id) ON DELETE CASCADE;

-- 技术栈表 (TechnologyOptimized)
CREATE TABLE technology_optimizeds (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    category VARCHAR(50),                -- web_server, framework, database, etc.
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_name (name),
    INDEX idx_category (category)
) ENGINE=InnoDB;

-- 扫描结果与技术栈关联表 (ScanResultTechnology)
CREATE TABLE scan_result_technologies (
    scan_result_id INT UNSIGNED NOT NULL,
    technology_id INT UNSIGNED NOT NULL,
    version VARCHAR(100),

    PRIMARY KEY (scan_result_id, technology_id),
    FOREIGN KEY (scan_result_id) REFERENCES scan_result_optimizeds(id) ON DELETE CASCADE,
    FOREIGN KEY (technology_id) REFERENCES technology_optimizeds(id) ON DELETE CASCADE
) ENGINE=InnoDB;

-- 扫描框架结果表 (ScanFrameworkResult) - 用于扫描工具原始输出
CREATE TABLE scan_framework_results (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    scan_target_id INT UNSIGNED NOT NULL,
    target VARCHAR(255) NOT NULL,
    scan_type VARCHAR(50) NOT NULL,      -- "port_scan", "subdomain_scan", "vulnerability_scan", etc.
    scanner_name VARCHAR(100) NOT NULL,  -- "nmap", "subfinder", "nuclei", etc.
    status VARCHAR(20) NOT NULL,         -- "running", "completed", "failed"
    start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_time TIMESTAMP NULL,
    raw_data TEXT,                       -- 原始扫描输出
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE,
    FOREIGN KEY (scan_target_id) REFERENCES scan_targets(id) ON DELETE CASCADE,

    INDEX idx_project_id (project_id),
    INDEX idx_scan_target_id (scan_target_id),
    INDEX idx_scan_type (scan_type),
    INDEX idx_status (status),
    INDEX idx_start_time (start_time)
) ENGINE=InnoDB;

-- 扫描框架目标表 (ScanFrameworkTarget)
CREATE TABLE scan_framework_targets (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    project_id INT UNSIGNED NOT NULL,
    target VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (project_id) REFERENCES project_optimizeds(id) ON DELETE CASCADE,

    INDEX idx_project_id (project_id),
    INDEX idx_target (target),
    UNIQUE KEY unique_project_target (project_id, target)
) ENGINE=InnoDB;

-- 创建视图：项目统计
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

-- 创建视图：漏洞概览
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