#!/bin/bash

# CyberEdge 数据库管理脚本
# 用于管理数据库schema版本和迁移

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 数据库连接参数
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-root}
DB_PASS=${DB_PASS:-password}
DB_NAME=${DB_NAME:-cyberedge}

# 检查MySQL连接
check_mysql_connection() {
    print_info "检查MySQL连接..."

    if ! command -v mysql &> /dev/null; then
        print_error "MySQL客户端未安装"
        print_info "尝试使用Docker容器中的MySQL..."

        if ! docker ps | grep -q cyberedge-mysql; then
            print_error "CyberEdge MySQL容器未运行"
            exit 1
        fi

        # 使用Docker容器中的MySQL
        MYSQL_CMD="docker exec cyberedge-mysql mysql -u$DB_USER -p$DB_PASS"
    else
        MYSQL_CMD="mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS"
    fi

    # 测试连接
    if ! $MYSQL_CMD -e "SELECT 1" &>/dev/null; then
        print_error "无法连接到MySQL服务器"
        exit 1
    fi

    print_success "MySQL连接正常"
}

# 检查数据库当前状态
check_database_status() {
    print_info "检查数据库状态..."

    # 检查数据库是否存在
    if ! $MYSQL_CMD -e "USE $DB_NAME; SELECT 1" &>/dev/null; then
        print_warning "数据库 $DB_NAME 不存在"
        return 1
    fi

    # 检查schema_migrations表
    if $MYSQL_CMD $DB_NAME -e "DESCRIBE schema_migrations" &>/dev/null; then
        # 检查是否已应用完整schema迁移
        if $MYSQL_CMD $DB_NAME -e "SELECT * FROM schema_migrations WHERE version='20240918_complete_scan_schema'" | grep -q "20240918_complete_scan_schema"; then
            print_success "当前数据库版本: v2.0 (完整扫描功能)"
            return 2
        else
            print_success "当前数据库版本: v1.0 (基础用户管理)"
            return 1
        fi
    else
        # 检查是否有用户表
        if $MYSQL_CMD $DB_NAME -e "DESCRIBE users" &>/dev/null; then
            print_success "当前数据库版本: v1.0 (基础用户管理)"
            return 1
        else
            print_warning "数据库为空"
            return 0
        fi
    fi
}

# 创建数据库
create_database() {
    print_info "创建数据库 $DB_NAME..."
    $MYSQL_CMD -e "CREATE DATABASE IF NOT EXISTS $DB_NAME CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
    print_success "数据库创建完成"
}

# 安装基础schema
install_basic_schema() {
    print_info "安装基础用户管理schema..."

    $MYSQL_CMD $DB_NAME -e "
    CREATE TABLE IF NOT EXISTS users (
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

    CREATE TABLE IF NOT EXISTS schema_migrations (
        version VARCHAR(50) PRIMARY KEY,
        applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        description TEXT
    );

    INSERT INTO schema_migrations (version, description)
    VALUES ('20240918_basic_user_schema', 'Basic user management schema')
    ON DUPLICATE KEY UPDATE applied_at = CURRENT_TIMESTAMP;"

    print_success "基础schema安装完成"
}

# 升级到完整schema
upgrade_to_complete_schema() {
    print_info "升级到完整扫描功能schema..."

    if [ ! -f "migrate_to_complete_schema.sql" ]; then
        print_error "迁移脚本文件不存在: migrate_to_complete_schema.sql"
        exit 1
    fi

    print_warning "此操作将添加扫描功能相关的表，继续吗? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_info "操作已取消"
        exit 0
    fi

    # 执行迁移
    if docker ps | grep -q cyberedge-mysql; then
        docker exec -i cyberedge-mysql mysql -u$DB_USER -p$DB_PASS $DB_NAME < migrate_to_complete_schema.sql
    else
        mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME < migrate_to_complete_schema.sql
    fi

    print_success "升级到完整schema完成"
}

# 回滚到基础schema
rollback_to_basic_schema() {
    print_error "警告：此操作将删除所有扫描相关数据！"
    print_warning "您确定要回滚到基础版本吗？所有扫描数据将被永久删除！(y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_info "操作已取消"
        exit 0
    fi

    print_warning "再次确认：输入 'DELETE_ALL_SCAN_DATA' 以继续回滚"
    read -r confirmation
    if [ "$confirmation" != "DELETE_ALL_SCAN_DATA" ]; then
        print_info "操作已取消"
        exit 0
    fi

    if [ ! -f "rollback_complete_schema.sql" ]; then
        print_error "回滚脚本文件不存在: rollback_complete_schema.sql"
        exit 1
    fi

    # 执行回滚
    if docker ps | grep -q cyberedge-mysql; then
        docker exec -i cyberedge-mysql mysql -u$DB_USER -p$DB_PASS $DB_NAME < rollback_complete_schema.sql
    else
        mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS $DB_NAME < rollback_complete_schema.sql
    fi

    print_success "回滚到基础schema完成"
}

# 全新安装完整schema
fresh_install_complete_schema() {
    print_info "全新安装完整扫描功能schema..."

    if [ ! -f "schema_complete.sql" ]; then
        print_error "完整schema文件不存在: schema_complete.sql"
        exit 1
    fi

    print_warning "此操作将创建完整的数据库结构，继续吗? (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_info "操作已取消"
        exit 0
    fi

    # 执行安装
    if docker ps | grep -q cyberedge-mysql; then
        docker exec -i cyberedge-mysql mysql -u$DB_USER -p$DB_PASS < schema_complete.sql
    else
        mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASS < schema_complete.sql
    fi

    print_success "完整schema安装完成"
}

# 显示帮助信息
show_help() {
    echo "CyberEdge 数据库管理工具"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  status      - 检查数据库状态"
    echo "  init        - 初始化基础数据库"
    echo "  upgrade     - 升级到完整扫描功能"
    echo "  rollback    - 回滚到基础版本 (危险操作)"
    echo "  install     - 全新安装完整schema"
    echo "  help        - 显示此帮助信息"
    echo ""
    echo "环境变量:"
    echo "  DB_HOST     - MySQL主机 (默认: localhost)"
    echo "  DB_PORT     - MySQL端口 (默认: 3306)"
    echo "  DB_USER     - MySQL用户 (默认: root)"
    echo "  DB_PASS     - MySQL密码 (默认: password)"
    echo "  DB_NAME     - 数据库名 (默认: cyberedge)"
}

# 主逻辑
main() {
    case "${1:-status}" in
        "status")
            check_mysql_connection
            if check_database_status; then
                status=$?
                case $status in
                    0) echo "数据库为空，可以运行 'init' 或 'install' 来初始化" ;;
                    1) echo "可以运行 'upgrade' 升级到完整功能" ;;
                    2) echo "数据库已是最新版本" ;;
                esac
            fi
            ;;
        "init")
            check_mysql_connection
            create_database
            install_basic_schema
            ;;
        "upgrade")
            check_mysql_connection
            upgrade_to_complete_schema
            ;;
        "rollback")
            check_mysql_connection
            rollback_to_basic_schema
            ;;
        "install")
            check_mysql_connection
            fresh_install_complete_schema
            ;;
        "help")
            show_help
            ;;
        *)
            print_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

main "$@"