# CyberEdge AI-Native 产品与架构草案

状态：Accepted baseline

实现状态：当前 vertical slice 已覆盖 Scope、passive DNS/Certificate Transparency Task、受控 TCP service baseline、独立授权的 signed-template Nuclei vulnerability baseline、TLS leaf Certificate、bounded Website/HTTP observation、same-origin depth-1 crawler、offline screenshot Evidence、evidence-bound Website technology fingerprint、Asset/Service/Certificate/Website/Observation/Evidence、Evidence-backed Finding、HTTP directory-listing/Git/DS_Store/Host collision 与 TLS validity detectors、确定性 TaskReport、Agent mutation audit、持久化 Schedule 生成普通 Task、Monitor 资产及服务/网站变化、UDS/mTLS、AI machine bridge 与只读 Web。OIDC、只读 RBAC、字段脱敏和独立事件投影属于后续阶段，本文中的完整产品条目不表示已经实现。

## 1. 产品定义

CyberEdge 是面向 AI Agent 的外部攻击面发现、验证与持续监控引擎。

AI Agent 是系统唯一的操作者。人类不直接使用命令行、API 或管理后台执行任务。人类通过可选的只读 Web 界面观察资产、风险、证据、任务状态和审计记录。

CyberEdge 不负责通用对话、意图理解或自主决策。它负责把 AI 的结构化意图转换为可验证、可审计、可恢复的确定性操作。

## 2. 核心原则

1. **AI-only control plane**：所有查询与变更由 AI Agent 通过 Skill 发起。
2. **Machine-first contract**：稳定接口是 Schema，不是面向人的命令名称、帮助文本或交互流程。
3. **Deterministic core**：AI 决定做什么，Rust Core 决定操作是否合法并保证执行一致性。
4. **Scope before execution**：任何主动探测都必须绑定已授权 Scope，执行层不得绕过。
5. **Evidence first**：Finding 必须能够追溯到 Observation 与原始 Evidence。
6. **One mutation path**：所有状态变更走同一协议、权限检查、状态机和审计链。
7. **Read-only human surface**：Web 只展示，不创建、修改、删除、重试或停止任务。
8. **Small architecture first**：没有真实瓶颈前，不引入消息队列、微服务或分布式工作流引擎。
9. **Single-instance deployment**：CyberEdge 面向单组织自托管，不设计租户、计费或跨租户隔离。

## 3. 系统边界

```text
AI Agent
  -> Skill
  -> gRPC / Protobuf
  -> Policy / Capability / Scope Validation
  -> Rust Core
  -> Task Engine
  -> Scanner Adapters
  -> PostgreSQL / Evidence Store

Human
  -> Optional Read-only Web
  -> Read Model / Event Stream
```

### 3.1 AI Agent

负责：

- 理解人类目标与上下文
- 选择并执行 Skill
- 组合多次查询和操作
- 阅读任务事件与证据
- 根据结果决定下一步
- 向人类解释结论

不负责：

- 绕过 Scope 或 capability 检查
- 直接修改数据库
- 直接运行扫描器进程
- 自行解释内部任务状态并强制改写

### 3.2 Skill

Skill 是 AI 的工作流与约束说明，不承载核心业务逻辑。

每个 Skill 至少声明：

- 名称、版本和用途
- 输入 Schema
- 所需 capability
- 允许调用的 protocol action/query
- 前置检查
- 成功条件
- 失败与恢复路径
- 输出解释规则

Skill 可以组织多个协议调用，但不能直接访问数据库、文件系统或扫描器。

### 3.3 CyberEdge Agent Protocol

这是 AI 与 Rust Core 之间唯一受支持的控制接口。协议使用 gRPC 与 Protobuf，Rust 服务端使用 `tonic`。

本地部署通过 Unix Domain Socket 连接；远程连接通过 HTTP/2 与 mTLS。两种传输共享同一套 Protobuf 契约，不提供第二套命令或 JSON API。

协议只提供两类操作：

- `query`：只读查询
- `invoke`：产生状态变化

协议使用 unary RPC 完成短查询和状态变更，使用 server streaming 传输 Task 事件。它不提供交互式 prompt、TTY UI、颜色、表格或面向人的进度条。

### 3.4 Rust Core

负责：

- Schema 验证
- capability 与 Scope 校验
- 领域状态机
- 幂等性
- 任务调度与并发控制
- Scanner Adapter 生命周期
- 数据一致性
- Evidence 保存
- 审计记录

### 3.5 Optional Web

Web 是人类观察窗，不是控制台。

允许展示：

- Scope 与授权边界
- Asset 与关系图
- Observation 时间线
- Finding、严重性与处置状态
- Evidence 原文与截图
- Task 状态与执行时间线
- Agent、Skill 与 protocol 调用审计
- 系统健康、容量和错误

禁止提供：

- 新建、编辑或删除 Scope
- 启动、停止、重试或取消 Task
- 修改策略、凭据、字典或 Scanner 配置
- Finding 状态修改
- PoC 上传、编辑或执行
- 任何绕过 Agent Protocol 的 mutation API

Web 必须符合企业端的信息结构和安全要求，包括全局搜索、高级筛选、保存视图、资产关系、时间线、报告导出、OIDC、只读 RBAC、字段脱敏和审计留存。只读限制的是控制能力，不是产品完整度。

Web 查询独立 Read Model，不直接读取执行核心的事务模型：

```text
Domain Events
  -> Read Model Projection
  -> Read-only Query Service
  -> Enterprise Web
```

## 4. RPC 草案

### 4.1 服务契约

```protobuf
service CyberEdge {
  rpc CreateScope(CreateScopeRequest) returns (Scope);
  rpc GetScope(GetScopeRequest) returns (Scope);
  rpc StartScan(StartScanRequest) returns (Task);
  rpc GetTask(GetTaskRequest) returns (Task);
  rpc WatchTask(WatchTaskRequest) returns (stream TaskEvent);
  rpc CancelTask(CancelTaskRequest) returns (Task);
  rpc SearchAssets(SearchAssetsRequest) returns (SearchAssetsResponse);
  rpc SearchObservations(SearchObservationsRequest) returns (SearchObservationsResponse);
  rpc GetEvidence(GetEvidenceRequest) returns (Evidence);
  rpc GetTaskReport(GetTaskReportRequest) returns (TaskReport);
  rpc SearchAudit(SearchAuditRequest) returns (SearchAuditResponse);
}
```

### 4.2 调用上下文

除 `Health` 外，每个 request 必须携带调用上下文。mutation 必须提供幂等键；query 同样携带该字段以保持统一 envelope，但不产生幂等记录：

```protobuf
message InvocationContext {
  string request_id = 1;
  string idempotency_key = 2;
  string agent_id = 3;
  string skill_name = 4;
  string skill_version = 5;
}
```

### 4.3 错误模型

```protobuf
message ErrorDetail {
  string code = 1;
  bool retryable = 2;
  map<string, string> metadata = 3;
}
```

错误使用标准 gRPC status，并通过 typed details 返回稳定的 `code`、重试语义和结构化上下文。自然语言 message 不属于兼容性契约。

### 4.4 Task 事件

长任务返回 `task_id`。AI 使用 server streaming 观察执行过程：

```protobuf
message WatchTaskRequest {
  InvocationContext context = 1;
  string task_id = 2;
  uint64 after_sequence = 3;
}

message TaskEvent {
  string task_id = 1;
  uint64 sequence = 2;
  string event_type = 3;
  google.protobuf.Timestamp occurred_at = 4;
}
```

事件必须有单调递增的 `sequence`。断线恢复使用最后确认的 sequence，不依赖不可靠的实时连接状态。

## 5. Capability 与安全模型

AI Agent 不能因为“知道命令”就拥有执行权限。

建议的 capability：

- `scope.read`
- `scope.manage`
- `asset.read`
- `scan.passive`
- `scan.active`
- `finding.read`
- `finding.triage`
- `evidence.read`
- `monitor.manage`
- `system.read`
- `report.read`
- `audit.read`

每次 `invoke` 必须同时满足：

1. Agent 身份有效
2. Skill 被允许
3. Skill 声明了所需 capability
4. Agent 获得该 capability
5. 目标属于有效 Scope
6. Policy 允许对应扫描方式
7. 请求通过速率、并发和资源限制

高风险 capability 应支持一次性授权、时间窗口、目标限制和最大调用次数。

## 6. 领域模型

```text
Scope
  -> Asset
  -> Observation
  -> Finding
  -> Evidence
  -> Remediation

Policy
  -> Task
  -> Stage
  -> Attempt
  -> Event

Agent
  -> Skill Invocation
  -> Audit Event
```

### Scope

授权边界。包含 Domain、IP、CIDR、组织主体、排除项、有效期和允许的扫描类型。

### Asset

被识别出的稳定对象，例如 Domain、IP、Service、Website、Certificate、Organization。Asset 不是某次扫描结果。

### Observation

带时间、来源和置信度的事实，例如 DNS 解析、开放端口、HTTP 响应、技术指纹。新的 Observation 不覆盖历史事实。

### Finding

需要判断或处置的安全问题。Finding 引用 Observation 与 Evidence，而不是复制扫描器输出。

### Evidence

不可变的原始证据，例如扫描器 JSON、HTTP 请求响应、证书、截图和日志片段。大对象存入对象存储，数据库只保存摘要、哈希和引用。

### Task

一次确定性执行。Task 必须绑定 Scope、Policy、Agent、Skill、输入快照和幂等键。Task 不代表策略或定时配置；Monitor 和 Schedule 只能产生 Task。

## 7. 初始协议能力

### Query

- `system.health`
- `scope.get`
- `scope.search`
- `asset.get`
- `asset.search`
- `observation.search`
- `finding.search`
- `evidence.get`
- `task.get`
- `task.events`
- `audit.search`

### Invoke

- `scope.create`
- `scope.update`
- `scan.start`
- `task.cancel`
- `monitor.create`
- `monitor.pause`
- `finding.triage`
- `report.generate`

这些名称只是首版协议词汇。最终以输入、输出 Schema 和状态语义为准。

## 8. 明确不做

- 不做面向人类的 CLI UX
- 不做可操作的 Web 管理后台
- 不做聊天界面
- 不允许 Web 直接 mutation
- 不让 Skill 承载领域逻辑
- 不让 AI 直接拼接 Scanner 命令
- 不执行未经 Scope 授权的目标
- 不在线编辑并直接执行任意 PoC
- 不在第一阶段引入 RabbitMQ、Kafka、Temporal 或 Kubernetes
- 不为了兼容 ARL 复制 MongoDB collection 模型
- 不设计多租户、租户管理、计费或跨租户配额

## 9. 第一阶段交付边界

第一阶段只证明完整的 AI 执行闭环：

1. Agent 身份与 Skill 元数据进入请求
2. 创建授权 Scope
3. 启动一个 passive discovery Task
4. 生成 Asset、Observation 与 Evidence
5. AI 通过事件流等待任务完成
6. AI 查询结果并生成报告
7. 人类通过只读 Web 查看同一份结果和完整审计链

第一阶段原始边界不接入主动端口扫描、Nuclei、企业工商查询或 GitHub 情报；当前实现已在后续阶段加入受控端口基线与隔离 Nuclei adapter，企业工商查询和 GitHub 情报仍未实现。

## 10. 已确认决策

已确认：

- Agent Protocol 使用 gRPC + Protobuf，同时支持本地 Unix Domain Socket 和远程 HTTP/2 + mTLS。
- Web 严格只读，但按照完整企业端观察平台设计。
- 执行实例命名为 Task，Schedule 和 Monitor 与 Task 分离。
- 产品面向单组织自托管，不支持多租户。

其余默认决策：

- Scope 由 AI 创建，但必须引用人类提供的授权声明。
- AI 可以在 capability 允许时调用 `task.cancel`。
- Evidence 首版使用 PostgreSQL content-addressed storage，后续可替换为 S3-compatible object storage。
- Skill 使用 `SKILL.md` 描述工作流，并提供宿主无关的 machine-readable manifest。
- CyberEdge 生成确定性报告数据包，AI 负责自然语言解释与最终报告。
