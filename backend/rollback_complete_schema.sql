-- CyberEdge Database Rollback Script
-- 回滚到基础用户表版本 (v2.0 -> v1.0)
-- 警告：此操作将删除所有扫描相关数据，请谨慎操作！

-- 检查是否存在迁移记录
SELECT COUNT(*) as migration_exists FROM schema_migrations WHERE version = '20240918_complete_scan_schema';

-- =============================================================================
-- 回滚操作：删除扫描功能相关表和视图
-- =============================================================================

-- 1. 删除视图
DROP VIEW IF EXISTS v_vulnerability_overview;
DROP VIEW IF EXISTS v_project_stats;

-- 2. 删除表（按依赖关系逆序删除）
DROP TABLE IF EXISTS scan_result_technologies;
DROP TABLE IF EXISTS scan_framework_results;
DROP TABLE IF EXISTS scan_framework_targets;
DROP TABLE IF EXISTS vulnerability_optimizeds;
DROP TABLE IF EXISTS web_path_optimizeds;
DROP TABLE IF EXISTS scan_result_optimizeds;
DROP TABLE IF EXISTS scan_targets;
DROP TABLE IF EXISTS technology_optimizeds;
DROP TABLE IF EXISTS project_optimizeds;

-- 3. 删除迁移记录
DELETE FROM schema_migrations WHERE version = '20240918_complete_scan_schema';

-- 输出回滚完成信息
SELECT 'Rollback completed! Database returned to basic user management only.' as rollback_status;
SELECT 'All scanning data has been permanently deleted.' as warning;