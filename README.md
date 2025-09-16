# CyberEdge 用户管理系统

极简的用户认证和管理系统，支持JWT认证和TOTP双因子认证。

## 快速开始

### 前置要求

- Docker
- Go 1.22+
- Node.js 16+
- netcat (用于端口检测)

### 一键启动开发环境

```bash
./start-dev.sh
```

这将自动：
1. 启动MySQL Docker容器
2. 初始化数据库schema
3. 启动后端API服务 (端口31337)
4. 启动前端开发服务器 (端口8080)

### 服务地址

- 前端: http://localhost:8080
- 后端API: http://localhost:31337
- MySQL: localhost:3306 (用户: root, 密码: password)

### 手动启动

如果需要分别启动各个服务：

#### 1. 启动MySQL

```bash
docker run -d --name cyberedge-mysql \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=cyberedge \
  -p 3306:3306 mysql:8.0

# 导入schema
docker exec -i cyberedge-mysql mysql -uroot -ppassword cyberedge < backend/schema.sql
```

#### 2. 启动后端

```bash
cd backend
go build -o cyberedge cmd/cyberedge.go
./cyberedge
```

#### 3. 启动前端

```bash
cd frontend
npm install
npm run serve
```

## API接口

### 认证

- `POST /auth/login` - 用户登录
- `POST /auth/register` - 用户注册
- `GET /auth/check` - 检查认证状态

### 用户管理

- `GET /users` - 获取所有用户
- `GET /users/:id` - 获取单个用户
- `POST /users` - 创建用户
- `DELETE /users/:id` - 删除用户

### 双因子认证

- `POST /auth/2fa/setup` - 设置2FA
- `POST /auth/2fa/verify` - 验证2FA
- `DELETE /auth/2fa` - 禁用2FA

## 数据库

仅使用一个`users`表，包含：
- 基本信息: username, email, password_hash
- 双因子认证: is_2fa_enabled, totp_secret
- 权限: role (admin/user)
- 时间戳: created_at, updated_at

## 技术栈

- 后端: Go + Gin + GORM + MySQL
- 前端: Vue 3 + Ant Design Vue
- 认证: JWT + TOTP
- 容器: Docker