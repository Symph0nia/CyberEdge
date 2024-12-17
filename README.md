# CyberEdge

## 简洁而强大的互联网资产测绘工具

CyberEdge 是一款精心设计的互联网资产测绘工具，为网络安全专业人士提供精准、高效的扫描体验。

版本：V1.0.8

**更新日志：**

feat: 更新Docker部署

fix: 修复允许的源问题

feat: 增加对arm64平台的支持

fix: 修复Dockerfile的同源问题

TODO：

1、未完成功能的开发。

2、进一步测试。

3、整合前后端。

## 核心特性

- **全面扫描**：从域名到漏洞，层层深入
- **直观界面**：清晰展示资产，一目了然
- **高效集成**：自动化资产关联，提升工作效率
- **用户友好**：简洁的操作界面，复杂变简单
- **性能优化**：多线程执行，快速完成任务

## 技术细节

- 后端：Go Gin
- 前端：Vue3
- 数据库：MongoDB
- 任务队列：Asynq

## 核心组件

- 子域名扫描: Subfinder
- 端口扫描: Nmap
- 路径扫描: Ffuf

## 搭建指南

### 快速部署

1. 克隆项目
```bash
git clone https://github.com/Symph0nia/CyberEdge.git
cd CyberEdge
```

2. 设置脚本权限并启动服务
```bash
chmod +x start.sh generate_env.sh
./start.sh
```

### 访问服务
* 前端界面: `http://localhost:47808`
* 后端 API: `http://localhost:31337`

**注意：确保你的系统已安装 Docker 和 Docker Compose。**

### 首次使用

1. 通过前端界面注册账号:
   - 注册过程需要使用 Google Authenticator 进行二次验证
   - 扫描二维码后会得到用户名和一次性密码

2. 安全建议:
   - 完成初始账号注册后，建议立即关闭注册通道以提升系统安全性
   - 确保修改默认密码和其他默认配置

### 系统组件

该部署包含以下服务:
- MongoDB: 数据存储
- Redis: 任务队列
- CyberEdge 后端: 核心服务
- CyberEdge 前端: Web 界面
- 集成工具: Subfinder、Nmap、Ffuf、HTTPx

### 问题排查

如果遇到服务启动问题，可以通过以下命令查看日志:
```bash
docker-compose logs -f
```

### 注意事项

- 确保部署环境的端口 47808 和 31337 未被占用
- 首次部署可能需要几分钟时间来下载和初始化所有组件
- 建议在生产环境中配置 HTTPS 和其他安全措施

## 界面预览

![Home](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/Home.png)

![Login](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/Login.png)

![Register](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/Register.png)

![Tools](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/Tools.png)

![User](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/User.png)

![System](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/System.png)

![Task](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/Task.png)

![Work](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/Work.png)

## 赞助

![YZA](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/YZA.png)

![MST](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/MST.png)

![DK](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/DK.png)

## 许可证

CyberEdge 项目采用 **GNU Affero 通用公共许可证 v3.0 (AGPL-3.0)**，这是一款确保软件自由使用、修改和分发的许可证，并且特别适用于网络应用。AGPL-3.0 要求在软件被修改、部署并对外提供服务时，必须公开相关的源代码。

### 核心条款

- **自由使用**：您可以自由地使用、修改和分发本项目的源代码。
- **开源义务**：如果您修改了 CyberEdge，并将其作为服务提供给第三方使用，您必须公开您的修改内容及源代码。
- **衍生作品**：任何基于本项目的修改或衍生作品，都需要以 AGPL-3.0 许可证发布，继续保持开源。

### 重要提示

AGPL-3.0 扩展了 GNU GPL 的要求，特别针对网络应用的场景。无论是通过本地安装还是通过网络提供服务，您都需要遵守许可证的开源义务。这确保了用户在使用软件时依然享有自由获取源代码的权利。更多详细信息请参阅 [AGPL-3.0 官方文档](https://www.gnu.org/licenses/agpl-3.0.html)。 

**简要声明**：CyberEdge 是自由软件，您可以在 AGPL-3.0 许可证下重新发布和修改，但需要保留此声明。

## 联系方式

邮箱：PayasoNorahC@protonmail.com

QQ群：

![QQ](https://raw.githubusercontent.com/ZacharyZcR/CyberEdge/main/image/QQ.jpg)

CyberEdge：简洁、高效、精准的资产测绘工具，为您的网络安全保驾护航。