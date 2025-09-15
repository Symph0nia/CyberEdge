-- CyberEdge MySQL Schema
-- 简洁、高效、标准化的数据库设计

CREATE DATABASE IF NOT EXISTS cyberedge CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE cyberedge;

-- 用户表
CREATE TABLE users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    account VARCHAR(100) NOT NULL UNIQUE,
    secret VARCHAR(255) NOT NULL,
    login_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_account (account)
) ENGINE=InnoDB;

-- 目标表
CREATE TABLE targets (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type ENUM('domain', 'ip', 'url') NOT NULL,
    target VARCHAR(500) NOT NULL,
    status ENUM('active', 'inactive', 'archived') DEFAULT 'active',
    -- 统计字段
    subdomain_count INT DEFAULT 0,
    port_count INT DEFAULT 0,
    path_count INT DEFAULT 0,
    vulnerability_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_target (target),
    INDEX idx_type (type),
    INDEX idx_status (status)
) ENGINE=InnoDB;

-- 任务表
CREATE TABLE tasks (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    target_id BIGINT UNSIGNED NULL,
    type ENUM('subfinder', 'nmap', 'ffuf') NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed') DEFAULT 'pending',
    payload TEXT NOT NULL, -- 扫描目标
    result LONGTEXT, -- 扫描结果
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE,
    INDEX idx_type (type),
    INDEX idx_status (status),
    INDEX idx_created (created_at)
) ENGINE=InnoDB;

-- 子域名结果表
CREATE TABLE subdomains (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    target_id BIGINT UNSIGNED NOT NULL,
    task_id BIGINT UNSIGNED NOT NULL,
    domain VARCHAR(500) NOT NULL,
    ip VARCHAR(45), -- 支持IPv6
    http_status INT DEFAULT 0,
    http_title VARCHAR(1000),
    is_alive BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    UNIQUE KEY unique_domain_target (domain, target_id),
    INDEX idx_target_id (target_id),
    INDEX idx_domain (domain),
    INDEX idx_alive (is_alive)
) ENGINE=InnoDB;

-- 端口扫描结果表
CREATE TABLE ports (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    target_id BIGINT UNSIGNED NOT NULL,
    task_id BIGINT UNSIGNED NOT NULL,
    host VARCHAR(500) NOT NULL,
    port INT NOT NULL,
    protocol ENUM('tcp', 'udp') DEFAULT 'tcp',
    state ENUM('open', 'closed', 'filtered') NOT NULL,
    service VARCHAR(100),
    http_status INT DEFAULT 0,
    http_title VARCHAR(1000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    UNIQUE KEY unique_host_port (host, port, target_id),
    INDEX idx_target_id (target_id),
    INDEX idx_port (port),
    INDEX idx_state (state)
) ENGINE=InnoDB;

-- 路径扫描结果表
CREATE TABLE paths (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    target_id BIGINT UNSIGNED NOT NULL,
    task_id BIGINT UNSIGNED NOT NULL,
    url VARCHAR(1000) NOT NULL,
    path VARCHAR(500) NOT NULL,
    status_code INT NOT NULL,
    content_length INT DEFAULT 0,
    content_words INT DEFAULT 0,
    content_lines INT DEFAULT 0,
    title VARCHAR(1000),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (target_id) REFERENCES targets(id) ON DELETE CASCADE,
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    UNIQUE KEY unique_url_target (url, target_id),
    INDEX idx_target_id (target_id),
    INDEX idx_status (status_code),
    INDEX idx_path (path)
) ENGINE=InnoDB;

-- 工具配置表（简化为JSON存储）
CREATE TABLE tool_configs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    config JSON NOT NULL, -- 存储所有工具配置
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_default (is_default)
) ENGINE=InnoDB;

-- 插入默认配置
INSERT INTO tool_configs (name, is_default, config) VALUES (
    '默认配置',
    TRUE,
    JSON_OBJECT(
        'nmap', JSON_OBJECT(
            'enabled', true,
            'ports', '21,22,23,25,53,80,110,135,139,143,443,445,993,995,1433,3306,3389,5432,6379,8080',
            'timeout', 300,
            'concurrency', 100
        ),
        'ffuf', JSON_OBJECT(
            'enabled', true,
            'wordlist', '/usr/share/wordlists/dirb/common.txt',
            'extensions', 'php,asp,jsp,html,js',
            'threads', 50
        ),
        'subfinder', JSON_OBJECT(
            'enabled', true,
            'threads', 10,
            'timeout', 60
        )
    )
);

-- 创建视图：任务统计
CREATE VIEW task_stats AS
SELECT
    type,
    status,
    COUNT(*) as count,
    AVG(TIMESTAMPDIFF(SECOND, created_at, COALESCE(completed_at, NOW()))) as avg_duration
FROM tasks
GROUP BY type, status;

-- 创建视图：目标概览
CREATE VIEW target_overview AS
SELECT
    t.*,
    COUNT(DISTINCT ta.id) as task_count,
    COUNT(DISTINCT s.id) as subdomain_count_actual,
    COUNT(DISTINCT p.id) as port_count_actual,
    COUNT(DISTINCT pa.id) as path_count_actual
FROM targets t
LEFT JOIN tasks ta ON t.id = ta.target_id
LEFT JOIN subdomains s ON t.id = s.target_id
LEFT JOIN ports p ON t.id = p.target_id
LEFT JOIN paths pa ON t.id = pa.target_id
GROUP BY t.id;