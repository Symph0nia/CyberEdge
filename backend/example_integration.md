# 扫描平台数据结构集成指南

## 集成到现有main.go

```go
package main

import (
    "cyberedge/pkg/api/handlers"
    "cyberedge/pkg/dao"
    "cyberedge/pkg/database"
    "cyberedge/pkg/service"
    // ... 其他导入
)

func main() {
    // ... 现有数据库连接代码 ...

    // 迁移扫描相关表
    if err := database.AutoMigrateScanModels(db); err != nil {
        log.Fatalf("Failed to migrate scan models: %v", err)
    }

    // 创建性能索引
    if err := database.CreateIndexes(db); err != nil {
        log.Printf("Warning: Failed to create indexes: %v", err)
    }

    // 创建数据约束
    if err := database.CreateConstraints(db); err != nil {
        log.Printf("Warning: Failed to create constraints: %v", err)
    }

    // ... 现有服务初始化 ...

    // 初始化扫描相关服务
    scanDAO := dao.NewScanDAO(db)
    scanService := service.NewScanService(scanDAO)
    scanHandler := handlers.NewScanHandler(scanService)

    // 注册路由
    api := router.Group("/api")
    scanHandler.RegisterScanRoutes(api)

    // ... 现有启动代码 ...
}
```

## API使用示例

### 1. 创建项目
```bash
curl -X POST http://localhost:8080/api/scan/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试项目",
    "description": "这是一个测试扫描项目"
  }'
```

### 2. 导入扫描结果
```bash
curl -X POST http://localhost:8080/api/scan/projects/1/import \
  -H "Content-Type: application/json" \
  -d '{
    "results": [
      {
        "ip": "192.168.1.100",
        "domain": "example.com",
        "subdomain": "www",
        "ports": [
          {
            "number": 80,
            "protocol": "tcp",
            "state": "open",
            "service": {
              "name": "http",
              "version": "Apache/2.4.41",
              "fingerprint": "Apache httpd 2.4.41 ((Ubuntu))",
              "banner": "HTTP/1.1 200 OK\\r\\nServer: Apache/2.4.41",
              "web_data": {
                "paths": [
                  {
                    "path": "/admin",
                    "status_code": 200,
                    "title": "Admin Panel",
                    "length": 1024,
                    "vulnerabilities": [
                      {
                        "title": "Admin Panel Accessible",
                        "description": "Admin panel is accessible without authentication",
                        "severity": "high",
                        "cvss": 7.5,
                        "location": "/admin",
                        "parameter": "",
                        "payload": ""
                      }
                    ]
                  }
                ],
                "technologies": ["Apache", "PHP"]
              },
              "vulnerabilities": [
                {
                  "cve_id": "CVE-2021-44790",
                  "title": "Apache HTTP Server Buffer Overflow",
                  "description": "A carefully crafted request body can cause a buffer overflow in the mod_lua multipart parser",
                  "severity": "critical",
                  "cvss": 9.8,
                  "location": "mod_lua",
                  "parameter": "",
                  "payload": ""
                }
              ]
            }
          }
        ]
      }
    ]
  }'
```

### 3. 获取项目统计
```bash
curl http://localhost:8080/api/scan/projects/1/stats
```

## 数据结构优势

### 1. 消除类型混乱
- **原设计**: Port → (Service || Web Service) 导致类型判断
- **新设计**: Port → Service (接口) → GenericService/WebService (实现)

### 2. 统一的数据访问
```go
// 统一获取所有漏洞
vulnerabilities := service.GetAllVulnerabilities()

// 统一获取漏洞统计
stats := service.GetVulnerabilityStats()
```

### 3. 易于扩展
```go
// 添加新服务类型只需实现Service接口
type DatabaseService struct {
    GenericService
    Schema     string
    Tables     []string
}

func (d *DatabaseService) IsWebService() bool {
    return false
}
```

## 查询示例

### 获取项目所有高危漏洞
```go
var vulnerabilities []models.Vulnerability
db.Raw(`
    SELECT v.* FROM vulnerabilities v
    JOIN services s ON v.service_id = s.id OR v.web_path_id IN (
        SELECT wp.id FROM web_paths wp WHERE wp.service_id = s.id
    )
    JOIN ports p ON s.port_id = p.id
    JOIN ip_addresses ip ON p.ip_address_id = ip.id
    JOIN subdomains sd ON ip.subdomain_id = sd.id
    JOIN domains d ON sd.domain_id = d.id
    WHERE d.project_id = ? AND v.severity IN ('critical', 'high')
`, projectID).Scan(&vulnerabilities)
```

### 获取Web服务技术栈分布
```go
var techStats []struct {
    Technology string
    Count      int
}
db.Raw(`
    SELECT t.name as technology, COUNT(*) as count
    FROM technologies t
    JOIN service_technologies st ON t.id = st.technology_id
    JOIN services s ON st.service_id = s.id
    WHERE s.type IN ('http', 'https')
    GROUP BY t.name
    ORDER BY count DESC
`).Scan(&techStats)
```

## 性能优化

1. **预加载关联数据**
```go
// 一次查询获取完整项目结构
project := scanDAO.GetProjectByID(projectID)
```

2. **合理的索引策略**
- 组合索引: (project_id, domain_name)
- 单列索引: severity, cve_id
- 外键索引: 自动创建

3. **批量操作**
```go
// 事务中批量插入整个层次结构
scanDAO.CreateOrUpdateHierarchy(project)
```