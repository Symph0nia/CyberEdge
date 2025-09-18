# CyberEdge 项目开发指南

## 项目当前状态 (Project Status)

**CyberEdge** 是一个企业级网络安全扫描平台，当前处于**功能完善和质量提升阶段**。

### 技术栈
- **后端**: Go + Gin + GORM + MySQL (生产级架构)
- **前端**: Vue 3 + Ant Design Vue + Vuex
- **安全**: JWT认证 + TOTP双因子认证 + bcrypt密码加密
- **扫描引擎**: Nmap + Subfinder + 自研扫描框架

### 最新开发进展
1. ✅ **完整数据库schema设计** - 支持项目管理、扫描目标、结果存储、漏洞管理
2. ✅ **扫描框架重构** - 基于管道的并发扫描架构
3. ✅ **测试代码质量革命** - 删除8726行垃圾测试，保留核心安全测试
4. 🔄 **前端界面优化** - 扫描结果展示和用户体验提升

---

## Linus开发原则 (The Linux Way)

### 核心哲学
> "删除代码是工程师最高尚的工作之一。每一行代码都是潜在的bug和维护负担。"

### 1. "好品味"是一切 (Good Taste is Everything)
- **数据结构决定程序质量** - 代码只是实现数据流动的工具
- **消除特殊情况** - 通过正确的设计让所有边界条件自然消失
- **简洁强制性** - 函数超过一屏或三层缩进都是设计失败

### 2. 绝对的实用主义 (Pragmatism Above All)
- 只解决**真实存在的问题**，拒绝为"未来需求"过度设计
- 面对烂代码，**直接重构**，不添加抽象层掩盖问题
- 删除可有可无的功能

### 3. 绝不破坏用户空间 (Never Break Userspace)
- **API/ABI向后兼容是神圣的** - 任何破坏现有接口的改动都是bug
- 内部重构自由，对外契约稳固

### 4. 设计健壮性优于测试 (Robust by Design)
- 代码应该天生健壮、逻辑清晰、简单到"不可能出错"
- 测试是发现愚蠢错误的工具，不是把烂设计变好的魔法

### 5. Git纪律 (Git Discipline)
- **一个提交只做一件事** - 原子化、独立的修改
- **提交信息和代码同等重要** - 清晰解释"做了什么"和"为什么"

---

## 代码质量标准 (Code Quality Standards)

### 后端 (Go)
```go
// ✅ 好的设计 - 简洁、专注、无特殊情况
func (s *ScanService) ExecuteScan(ctx context.Context, config ScanConfig) (*ScanResult, error) {
    if err := s.validateConfig(config); err != nil {
        return nil, err
    }

    scanner, err := s.manager.GetScanner(config.Tool)
    if err != nil {
        return nil, err
    }

    return scanner.Scan(ctx, config)
}

// ❌ 垃圾设计 - 特殊情况处理、过度复杂
func (s *ScanService) ExecuteComplexScanWithMultipleOptions(...) {
    if special_case_1 {
        // 50 lines of special handling
    } else if special_case_2 {
        // another 50 lines
    }
    // 这是设计失败的标志
}
```

### 前端 (Vue 3)
```javascript
// ✅ 好的组件设计 - 单一职责、清晰接口
const ScanResults = {
  props: ['scanId'],
  setup(props) {
    const { results, loading, error } = useScanResults(props.scanId)
    return { results, loading, error }
  }
}

// ❌ 垃圾组件 - 职责混乱、状态混乱
const MegaComponent = {
  // 管理用户、扫描、设置、通知...什么都做
}
```

---

## 测试哲学 (Testing Philosophy)

### 当前测试状态
经过**外科手术式清理**，我们删除了22个无用测试文件(8726行垃圾代码)，保留的精华测试：

#### 保留的核心测试
1. **user_handler_test.go** - HTTP安全、认证安全、输入验证
2. **user_service_test.go** - 密码加密、JWT安全、并发安全
3. **integration_test.go** - 端到端业务流程测试
4. **scanFrameworkApi.test.js** - API接口测试

#### 被删除的垃圾
- 测试字段赋值的"单元测试"
- 测试GORM基本功能的"集成测试"
- 测试Vue组件点击事件的"功能测试"

### 测试编写原则
```go
// ✅ 测试真正的风险点
func TestConcurrentPasswordHashing(t *testing.T) {
    // 测试100个goroutine同时哈希密码的安全性
    // 这是生产环境真实可能出现的场景
}

func TestSQLInjectionPrevention(t *testing.T) {
    // 测试恶意输入是否被正确处理
    maliciousInputs := []string{
        "'; DROP TABLE users; --",
        "' OR '1'='1",
    }
    // 这些是真正的安全威胁
}

// ❌ 垃圾测试
func TestUserModelFields(t *testing.T) {
    user := User{Username: "test"}
    assert.Equal(t, "test", user.Username) // 这TM在测试什么？
}
```

---

## 开发工作流 (Development Workflow)

### 1. 功能开发流程
```bash
# 1. 理解问题本质
# 问三个问题：这是真问题吗？有更简单的方法吗？会破坏什么吗？

# 2. 设计数据结构
# 先设计数据流，再写代码

# 3. 实现最简方案
# 能删就删，能简化就简化

# 4. 写真正的测试
# 测试错误路径、边界条件、安全问题

# 5. 原子提交
git commit -m "fix: 修复并发扫描时的资源竞争问题

通过引入channel缓冲区解决多个goroutine同时访问扫描器
导致的竞争条件，确保扫描任务的线程安全。

🤖 Generated with Claude Code
Co-Authored-By: Claude <noreply@anthropic.com>"
```

### 2. 代码审查标准
每个PR必须通过以下检查：
- [ ] **数据结构设计合理** - 没有不必要的复杂性
- [ ] **消除特殊情况** - 通过设计而非if/else解决问题
- [ ] **向后兼容** - API/ABI不会破坏现有功能
- [ ] **安全考虑** - 输入验证、权限检查、加密正确
- [ ] **测试覆盖核心风险** - 不是测试框架功能

---

## 技术债务管理 (Technical Debt Management)

### 当前技术债务
1. **前端状态管理** - Vuex使用可以进一步简化
2. **扫描结果存储** - 大量结果的性能优化
3. **错误处理** - 统一的错误处理机制

### 债务偿还原则
- **立即修复影响安全的问题**
- **重构烂设计而非打补丁**
- **删除未使用的功能**
- **合并重复代码**

---

## 架构决策记录 (Architecture Decision Records)

### ADR-001: 采用Go作为后端语言
**决策**: 使用Go + Gin替代其他语言框架
**原因**:
- 并发模型适合扫描任务
- 内存安全
- 部署简单(单一二进制文件)
- 性能优秀

### ADR-002: 数据库设计优化
**决策**: 使用optimized表结构，支持复杂扫描场景
**原因**:
- 支持项目->目标->结果的层级关系
- 漏洞与扫描结果解耦
- 支持技术栈标签系统

### ADR-003: 删除无用测试
**决策**: 删除8726行测试代码，保留核心安全测试
**原因**:
- 90%的测试在验证框架功能而非业务逻辑
- 维护无用测试浪费时间
- 专注于真正的风险点

---

## 安全考虑 (Security Considerations)

### 认证安全
- JWT + TOTP双因子认证
- bcrypt密码哈希(cost 12)
- 会话管理和过期控制

### 输入验证
- 所有用户输入都必须验证
- SQL注入防护(参数化查询)
- XSS防护(输出编码)

### 扫描安全
- 扫描目标权限验证
- 速率限制防止滥用
- 结果数据隔离

---

## 性能考虑 (Performance Considerations)

### 后端优化
- 数据库查询优化(索引、分页)
- 并发扫描控制(goroutine pool)
- 内存使用监控

### 前端优化
- 大数据量虚拟滚动
- API结果缓存
- 图片懒加载

---

## 部署和运维 (Deployment & Operations)

### 开发环境
```bash
./start-dev.sh  # 一键启动开发环境
```

### 生产部署
- Docker容器化部署
- MySQL主从复制
- Redis缓存层
- Nginx反向代理

### 监控指标
- 扫描任务成功率
- API响应时间
- 数据库连接池状态
- 内存使用率

---

## 贡献指南 (Contributing Guidelines)

### 提交代码前检查清单
- [ ] 代码遵循项目风格(gofmt, eslint)
- [ ] 添加必要的测试(专注核心逻辑)
- [ ] 更新相关文档
- [ ] 提交信息清晰描述修改

### 禁止的行为
- ❌ 添加无用的抽象层
- ❌ 为未来需求过度设计
- ❌ 提交破坏API兼容性的修改
- ❌ 添加测试框架功能的"测试"

---

## 联系和支持 (Contact & Support)

### 开发团队原则
我们相信：
- **代码质量胜过功能数量**
- **简单设计胜过复杂架构**
- **删除代码胜过添加代码**
- **解决真实问题胜过理论完美**

### 技术讨论
使用GitHub Issues进行技术讨论，遵循以下原则：
- 描述具体问题而非抽象需求
- 提供复现步骤
- 考虑向后兼容性影响

---

**记住: 我们的目标不是写更多代码，而是写更好的代码。每一行代码都必须证明自己的价值。**

---
*最后更新: 2025-01-18*
*项目版本: v2.0 (完整扫描功能版本)*