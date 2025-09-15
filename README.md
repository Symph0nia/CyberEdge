# CyberEdge - 网络安全渗透测试平台

CyberEdge是一个现代化的网络安全渗透测试平台，提供全面的安全评估和漏洞检测功能。

## 项目结构

```
CyberEdge/
├── backend/    # Go后端API服务
├── frontend/   # Vue.js前端界面
└── README.md   # 项目说明文档
```

## 快速开始

### 后端服务
```bash
cd backend
go mod tidy
go run cmd/cyberedge.go
```

### 前端服务
```bash
cd frontend
npm install
npm run serve
```

## 开发分支
- `main`: 生产分支
- `dev`: 开发分支

## 功能特性
- 子域名发现 (Subfinder)
- 端口扫描 (Nmap)
- 目录爆破 (Ffuf)
- Web服务识别
- 任务管理系统
- 用户权限控制

更多详细信息请参考各子项目的README文档。