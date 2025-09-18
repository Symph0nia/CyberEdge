# CyberEdge 数据库迁移指南

## 概述

CyberEdge 项目现在支持完整的扫描功能数据库schema。本文档说明如何从基础用户管理版本升级到完整扫描功能版本。

## 数据库版本

- **v1.0 (基础版本)**: 仅包含用户管理功能
- **v2.0 (完整版本)**: 包含完整扫描功能 (项目、目标、扫描结果、漏洞等)

## 快速迁移

### 自动迁移 (推荐)
```bash
cd backend
./manage_database.sh upgrade
```

### 全新安装
```bash
cd backend
./manage_database.sh install
```

## 详细说明

### 1. 检查当前状态
```bash
./manage_database.sh status
```

### 2. 升级现有数据库
如果你已有基础用户数据，使用升级命令：
```bash
./manage_database.sh upgrade
```
这将保留现有用户数据，并添加扫描功能表。

### 3. 全新安装
如果是全新部署，直接安装完整schema：
```bash
./manage_database.sh install
```

### 4. 回滚 (紧急情况)
⚠️ **危险操作** - 将删除所有扫描数据
```bash
./manage_database.sh rollback
```

## 新增表结构

升级后将包含以下新表：

### 核心扫描表
- `project_optimizeds` - 扫描项目
- `scan_targets` - 扫描目标 (域名/IP)
- `scan_result_optimizeds` - 扫描结果 (端口/服务)
- `vulnerability_optimizeds` - 漏洞信息
- `web_path_optimizeds` - Web路径

### 扫描框架表
- `scan_framework_results` - 扫描工具原始输出
- `scan_framework_targets` - 扫描框架目标

### 辅助表
- `technology_optimizeds` - 技术栈
- `scan_result_technologies` - 扫描结果与技术栈关联
- `schema_migrations` - 迁移版本记录

### 视图
- `v_project_stats` - 项目统计
- `v_vulnerability_overview` - 漏洞概览

## 开发环境

更新后的 `start-dev.sh` 会自动检测并使用完整schema。

### 手动启动开发环境
```bash
./start-dev.sh
```

脚本会自动：
1. 启动MySQL容器
2. 检查是否存在完整schema文件
3. 自动应用正确的数据库结构
4. 启动后端和前端服务

## 验证迁移

### 1. 检查表结构
```sql
SHOW TABLES;
```

应该看到所有新表。

### 2. 检查迁移记录
```sql
SELECT * FROM schema_migrations;
```

应该看到迁移版本记录。

### 3. 运行测试
```bash
cd backend
go run schema_validation_test.go test
```

## 故障排除

### 迁移失败
1. 检查MySQL连接
2. 确保有足够权限
3. 检查磁盘空间
4. 查看错误日志

### 回滚
如果迁移出现问题，可以回滚：
```bash
./manage_database.sh rollback
```

### 重新迁移
```bash
./manage_database.sh rollback
./manage_database.sh upgrade
```

## 生产环境注意事项

1. **备份数据**: 迁移前必须备份现有数据
2. **维护窗口**: 在维护时间窗口内执行迁移
3. **测试环境**: 先在测试环境验证迁移
4. **监控**: 迁移后监控系统性能

## 文件说明

- `schema_complete.sql` - 完整数据库schema
- `migrate_to_complete_schema.sql` - 增量迁移脚本
- `rollback_complete_schema.sql` - 回滚脚本
- `manage_database.sh` - 数据库管理工具
- `schema_validation_test.go` - 验证测试

## 支持

如有问题，请检查：
1. MySQL服务是否正常运行
2. 数据库连接配置是否正确
3. 是否有足够的数据库权限
4. 相关脚本文件是否存在并可执行