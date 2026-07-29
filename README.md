# CyberEdge

CyberEdge 是一个面向 AI Agent 的外部攻击面管理引擎。AI Agent 通过 Skills 和强类型 RPC 合同操作系统，人员通过可选的只读 Web 界面查看资产、任务、发现、证据与变化。

系统围绕一条可审计的数据链路工作：发现资产、记录观测、验证发现、保留证据并推动修复。所有主动探测都必须属于带有明确授权依据的 Scope。

## 技术栈

- Rust 1.97、Tokio、tonic
- gRPC + Protobuf 控制面
- PostgreSQL 持久化与初始任务队列
- React + TypeScript + Vite 只读 Web
- 通过强类型 Adapter 隔离扫描器进程

详细设计参见[产品架构](docs/ai-native-architecture.md)、[实现架构](docs/architecture.md)和[运维手册](docs/operations.md)。

## 部署

### 基础部署

准备 Docker 和 Docker Compose，然后执行：

```bash
docker compose up --build -d
docker compose ps
```

基础部署会启动 PostgreSQL 和 CyberEdge Core，但不会暴露 Web 界面。默认 Agent 策略来自 `config/agents.example.toml`，生产环境应复制为独立配置并按实际 Agent、Skill 和 capability 收紧授权。

常用维护命令：

```bash
docker compose logs -f cyberedge
docker compose restart cyberedge
docker compose down
```

数据库数据保存在 Compose 的 `postgres-data` volume 中。执行 `docker compose down -v` 会删除数据库，除非明确需要清空环境，否则不要使用。

### 启用扫描与情报 Adapter

按需要叠加对应的 Compose 文件：

```bash
# 固定、签名且经过审核的 Nuclei 基线
docker compose -f compose.yaml -f compose.nuclei.yaml up -d --build

# 离线网页截图
docker compose -f compose.yaml -f compose.screenshots.yaml up -d --build

# GitHub 公共代码元数据检索
docker compose -f compose.yaml -f compose.public-code.yaml up -d --build

# 精确 CPE 与 NVD CVE 关联
docker compose -f compose.yaml -f compose.cve.yaml up -d --build

# 合规注册信息服务
docker compose -f compose.yaml -f compose.registration.yaml up -d --build
```

各 Adapter 的凭据、模板同步、安全边界和覆盖语义见[运维手册](docs/operations.md)。未配置或不可用的 Adapter 会在任务 coverage 中明确标记，不会伪装成已完成。

### 启用只读 Web

Web 是观察界面，不提供修改接口，但仍然包含敏感资产数据。网络部署必须配置 OIDC：

```bash
export CYBEREDGE_WEB_OIDC_ISSUER='https://id.example.com/'
export CYBEREDGE_WEB_OIDC_AUDIENCE='cyberedge'
export CYBEREDGE_WEB_OIDC_JWKS_URL='https://id.example.com/.well-known/jwks.json'

docker compose -f compose.yaml -f compose.web.yaml up -d --build
```

OIDC Token 至少需要 `cyberedge.read` 角色。查看原始 Evidence 需要 `cyberedge.evidence.read`，查看 Scope 授权依据及 Audit 身份字段需要 `cyberedge.sensitive.read`。

仅本机开发预览时，可以使用：

```bash
npm --prefix web ci
npm --prefix web run build

DATABASE_URL='postgres://cyberedge:cyberedge@127.0.0.1/cyberedge' \
CYBEREDGE_AGENT_POLICY='config/agents.example.toml' \
CYBEREDGE_EXECUTION_ENABLED=false \
CYBEREDGE_WEB_BIND='127.0.0.1:8080' \
CYBEREDGE_WEB_ALLOW_INSECURE_LOCAL=true \
cargo run
```

然后打开 <http://127.0.0.1:8080/>。不安全本地模式只允许绑定 loopback，不能用于局域网或公网。

## 使用

CyberEdge 主要由 AI Agent 通过仓库中的 Skills 使用，而不是由人员手工拼接扫描命令。推荐流程：

1. 创建带有授权依据的 Scope。
2. 运行 readiness 检查，确认可用的 Adapter。
3. 启动 `standard` 或 `thorough` Assessment。
4. 等待 Task 到达终态，并检查每个阶段的 coverage。
5. 根据 Observation、Evidence 和 Finding 复核结论。
6. 需要持续监控时，再创建 Schedule 并比较后续 ExposureChange。

完整评估使用 `cyberedge-assess-scope` Skill；资产发现、服务发现、网站与证书观察、漏洞扫描、公共代码情报、CVE 关联、注册信息和结果检查分别由对应的 `cyberedge-*` Skill 完成。

`cyberedge-agent` 是供 Skills 使用的 JSON stdin/stdout 机器桥接器，不是面向人员的交互式 CLI。下面的示例仅用于部署联调：

```bash
printf '%s' \
  '{"request_id":"req_scope_1","idempotency_key":"idem_scope_1","agent_id":"codex-main","skill_name":"cyberedge-discover-assets","skill_version":"0.1.0","action":"create_scope","name":"Example","authorization_ref":"authorization:change-1234","targets":[{"kind":"domain","value":"example.com"}]}' \
  | docker compose exec -T cyberedge cyberedge-agent
```

创建 Scope 后，可由 Assessment Skill 使用返回的 `scope_id` 启动完整评估。Task 状态为 `COMPLETED` 只表示工作流正常结束；是否真正覆盖每个阶段，应以 Task Report 中的 `complete`、`partial`、`unavailable` 和 `blocked` 为准。

## 本地开发

```bash
cargo fmt --all -- --check
cargo test
cargo clippy --all-targets --all-features -- -D warnings
npm --prefix web ci
npm --prefix web run build
```

原生运行至少需要：

```bash
DATABASE_URL='postgres://...' \
CYBEREDGE_AGENT_POLICY='config/agents.example.toml' \
cargo run
```

本地 RPC 默认监听 `unix:///tmp/cyberedge.sock`，可通过 `CYBEREDGE_RPC_SOCKET` 修改。远程控制面必须使用 mTLS，具体配置见[运维手册](docs/operations.md)。

## 参与贡献与许可证

提交改动前请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)，安全问题请按 [SECURITY.md](SECURITY.md) 中的流程报告。

CyberEdge 使用 [Apache License 2.0](LICENSE) 许可证。
