# 安全指南

## 数据库安全

### 环境变量配置
- 生产环境中绝不使用默认密码
- 使用强密码和适当的数据库用户权限
- 通过环境变量配置敏感信息

```bash
# 推荐的生产环境配置
export DB_USER="cyberedge_user"
export DB_PASS="your_secure_password_here"
export JWT_SECRET="your_jwt_secret_at_least_32_chars_long"
export SESSION_SECRET="your_session_secret_here"
```

### 测试环境
- 使用独立的测试数据库
- 测试中的敏感数据仅用于测试目的
- 定期清理测试数据

## 依赖管理

### 前端
```bash
# 定期检查安全漏洞
npm audit
npm audit fix

# 更新依赖
npm update
```

### 后端
```bash
# 检查Go模块安全性
go list -m -u all
go mod tidy

# 使用安全工具
go install github.com/securecodewarrior/gosec/cmd/gosec@latest
gosec ./...
```

## 部署安全

1. **数据库访问控制**
   - 限制数据库访问IP
   - 使用最小权限原则
   - 启用数据库审计日志

2. **应用配置**
   - 更改默认端口
   - 配置防火墙规则
   - 启用HTTPS

3. **监控和日志**
   - 监控异常登录
   - 记录关键操作
   - 设置安全告警