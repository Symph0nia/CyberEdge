# CyberEdge 测试指南

本文档详细说明了 CyberEdge 安全扫描平台的测试架构和覆盖率配置。

## 测试架构概览

CyberEdge 采用多层测试策略：

```
📊 测试金字塔
├── 🔬 单元测试 (Unit Tests)
│   ├── 前端组件测试 (Vue.js + Vitest)
│   └── 后端函数测试 (Go + testing)
├── 🔗 集成测试 (Integration Tests)
│   ├── API 接口测试
│   └── 数据库交互测试
└── 🌐 端到端测试 (E2E Tests)
    └── 用户流程测试 (Playwright)
```

## 快速开始

### 运行所有测试
```bash
# 运行完整测试套件
./run-all-tests.sh

# 运行特定类型的测试
./run-all-tests.sh frontend  # 仅前端测试
./run-all-tests.sh backend   # 仅后端测试
./run-all-tests.sh e2e       # 仅E2E测试
```

## 前端测试

### 技术栈
- **测试框架**: Vitest
- **覆盖率工具**: @vitest/coverage-v8
- **组件测试**: @vue/test-utils
- **模拟工具**: 内置 vi.mock()

### 运行前端测试
```bash
cd frontend

# 运行测试
npm run test

# 运行测试并生成覆盖率报告
npm run test:coverage

# 在UI模式下运行测试
npm run test:coverage:ui
```

### 覆盖率目标
- **行覆盖率**: ≥ 70%
- **函数覆盖率**: ≥ 70%
- **分支覆盖率**: ≥ 60%
- **语句覆盖率**: ≥ 70%

### 测试文件结构
```
frontend/src/
├── components/
│   └── __tests__/          # 组件单元测试
├── views/
│   └── __tests__/          # 页面组件测试
├── store/
│   └── __tests__/          # 状态管理测试
└── tests/
    └── setup.js            # 测试环境配置
```

## 后端测试

### 技术栈
- **测试框架**: Go 标准 testing 包
- **覆盖率工具**: go tool cover
- **HTTP测试**: httptest 包
- **数据库测试**: 内存SQLite/测试MySQL

### 运行后端测试
```bash
cd backend

# 运行测试并生成覆盖率报告
./test-all.sh

# 手动运行测试
go test -v -race -coverprofile=coverage/coverage.out ./...

# 生成HTML覆盖率报告
go tool cover -html=coverage/coverage.out -o coverage/coverage.html
```

### 覆盖率目标
- **总体覆盖率**: ≥ 70%

### 测试文件结构
```
backend/
├── pkg/
│   ├── api/handlers/
│   │   └── *_test.go       # API处理器测试
│   ├── service/
│   │   └── *_test.go       # 业务逻辑测试
│   └── dao/
│       └── *_test.go       # 数据访问测试
└── tests/
    └── integration/        # 集成测试
```

## 端到端测试

### 技术栈
- **E2E框架**: Playwright
- **浏览器**: Chromium, Firefox, WebKit
- **测试环境**: 自动启动开发服务器

### 运行E2E测试
```bash
cd frontend

# 安装浏览器
npx playwright install

# 运行E2E测试
npm run test:e2e

# 在UI模式下运行
npm run test:e2e:ui

# 查看测试报告
npm run test:e2e:report
```

### 测试场景
- ✅ 用户认证 (登录/注册/2FA)
- ✅ 仪表板功能
- ✅ 扫描功能 (端口扫描/子域名扫描)
- ✅ 设置管理
- ✅ 用户管理 (管理员功能)

### E2E测试文件
```
frontend/e2e/
├── auth.spec.js            # 认证流程测试
├── dashboard.spec.js       # 仪表板测试
├── scanning.spec.js        # 扫描功能测试
└── settings.spec.js        # 设置页面测试
```

## 覆盖率报告

### 查看覆盖率报告

**前端覆盖率**:
```bash
# 生成并打开HTML报告
cd frontend && npm run test:coverage
open coverage/index.html
```

**后端覆盖率**:
```bash
# 生成并打开HTML报告
cd backend && ./test-all.sh
open coverage/coverage.html
```

**E2E测试报告**:
```bash
# 查看E2E测试报告
cd frontend && npm run test:e2e:report
```

### 覆盖率配置

覆盖率配置在以下文件中定义：
- `frontend/vitest.config.js` - 前端覆盖率配置
- `backend/test-all.sh` - 后端覆盖率配置
- `test-config.json` - 统一测试配置

## 持续集成

### CI 流水线
```yaml
# 示例 CI 配置
stages:
  - dependencies
  - test-frontend
  - test-backend
  - test-e2e
  - coverage-report

# 覆盖率阈值检查
coverage-check:
  - 前端覆盖率 ≥ 70%
  - 后端覆盖率 ≥ 70%
  - 失败时退出构建
```

## 最佳实践

### 编写测试的原则
1. **单一职责**: 每个测试只验证一个功能点
2. **独立性**: 测试之间不应相互依赖
3. **可重复性**: 测试结果应该一致和可预测
4. **清晰命名**: 测试名称应该描述测试的内容

### 前端测试最佳实践
- 使用 `data-testid` 而非 CSS 类名定位元素
- 模拟外部依赖 (API调用、第三方库)
- 测试用户交互而非实现细节
- 保持测试简单和可读

### 后端测试最佳实践
- 使用表驱动测试处理多种输入
- 测试错误路径和边界条件
- 使用测试数据库避免污染生产数据
- 清理测试产生的副作用

### E2E测试最佳实践
- 测试关键用户流程
- 使用页面对象模式 (Page Object Pattern)
- 避免测试实现细节
- 合理使用等待和重试机制

## 故障排除

### 常见问题

**前端测试失败**:
```bash
# 清理依赖重新安装
cd frontend && rm -rf node_modules && npm install

# 更新快照
npm run test -- --update-snapshots
```

**后端测试失败**:
```bash
# 检查Go版本
go version

# 更新依赖
go mod tidy && go mod download
```

**E2E测试失败**:
```bash
# 重新安装浏览器
npx playwright install

# 检查服务器是否运行
curl http://localhost:8080
```

## 贡献指南

### 添加新测试
1. 在相应目录创建测试文件
2. 遵循现有的命名约定
3. 确保测试通过且覆盖率满足要求
4. 更新相关文档

### 代码提交前检查
```bash
# 运行完整测试套件
./run-all-tests.sh

# 检查代码风格
cd frontend && npm run lint
cd backend && go fmt ./...
```

---

📊 **当前测试状态**
- ✅ 前端单元测试: 26 个测试用例
- ✅ 后端集成测试: 完整API测试覆盖
- ✅ E2E测试: 4个主要功能模块
- ✅ 覆盖率配置: 前端+后端统一配置

更多信息请参考各测试目录中的 README 文件。