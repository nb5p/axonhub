# AxonHub Rust + deno_core 插件优先网关架构计划

> 状态：设计草案  
> 日期：2026-08-12  
> 适用分支：`ai-slop` 规划记录  
> 任务类型：本地架构规划，尚未进入实现  
> 数据库影响：当前无  
> 构建与测试：本次未执行

## 1. 背景

当前 AxonHub 是成熟的 Go 项目，主要技术栈包括 Gin、Ent、gqlgen、FX 和独立的 `llm` Go module。现有 LLM Pipeline 已经覆盖多协议转换、渠道选择、同渠道重试、跨渠道故障转移、流式响应、Usage 统计和请求追踪。

新架构希望进一步收缩核心职责，把供应商接入、模型映射、渠道策略、请求改写、协议转换和管理扩展交给插件。典型长尾需求包括：

- 渠道仅在特定时间开放；
- 查询外部管理 API，余额达到阈值后才允许调用；
- 下游模型名与不同上游模型名之间动态映射；
- 将 `developer` 消息改写成 `system`；
- 为特定渠道添加 Header、参数或签名；
- 第三方中转站拥有私有管理 API、模型发现 API 或认证流程；
- 同一个逻辑模型根据租户、渠道、时间、健康状态和余额选择不同上游。

首期采用以下技术边界：

- 主程序使用 Rust；
- JavaScript/TypeScript 是唯一插件语言；
- Rust 通过 `deno_core` 嵌入 V8；
- 不实现 Node.js 兼容；
- 不引入 Node sidecar；
- 不引入 WASI/Wasm 插件；
- 预留未来扩展端点，但不为尚未出现的需求承担实现成本；
- 目标场景以个人部署和低并发为主，峰值约 3～4 个并发请求。

## 2. 设计目标

### 2.1 产品目标

1. 核心不内置任何上游模型供应商。
2. 官方维护的 OpenAI Compatible、Anthropic 等供应商也走公开插件 API。
3. 核心原生提供稳定的下游协议入口，首期包括：
   - `POST /v1/chat/completions`
   - `POST /v1/responses`
4. 插件可贡献供应商、模型发现、路由、调度、策略、中间件和管理 API。
5. 简单规则可直接在管理界面编辑 TypeScript，保存后热重载。
6. 完整插件仍使用 TypeScript，可拥有配置、生命周期和多个扩展能力。
7. 每次请求的模型映射、渠道过滤、评分和改写过程都能被解释。
8. 插件异常不应破坏整个请求系统，插件超时后可被终止并熔断。

### 2.2 工程目标

- Rust 核心掌握连接、流式背压、取消、超时、重试、账务和安全边界；
- 插件通过稳定 SDK 操作受控数据，不依赖 Rust 内部模块；
- 请求全程绑定同一插件代际，热重载不影响在途流；
- 插件获得最小权限，默认看不到文件、环境变量、子进程和明文密钥；
- 插件变更以结构化 Patch 表达，保留完整审计链；
- 首期保持单进程和单运行时，降低安装体积与维护成本。

## 3. 明确暂缓的能力

首期不处理以下内容：

- CommonJS；
- Node 内置模块，如 `node:fs`、`node:http`、`node:crypto`；
- 任意 npm 包兼容；
- 原生 `.node` 扩展；
- 插件自定义构建脚本；
- 插件直接访问主机文件系统；
- 插件直接启动子进程；
- 插件读取全部环境变量；
- 插件获得明文 API Key；
- WASI/WIT；
- Node sidecar；
- 第三方原生动态库；
- 分布式插件执行；
- 响应已经提交给客户端后的透明上游切换。

出现真实需求和性能数据后，再决定是否加入 Node sidecar 或 Wasm。当前不为假想规模增加复杂度。

## 4. 总体架构

```text
Client
  │
  ▼
Rust HTTP Server
  │
  ├─ OpenAI Chat Completions Endpoint
  └─ OpenAI Responses Endpoint
  │
  ▼
Protocol Decoder
  │
  ▼
Canonical Request IR
  │
  ▼
Rust Request Pipeline
  ├─ Authentication
  ├─ Policy Hooks              ─┐
  ├─ Model Resolution           │
  ├─ Candidate Discovery        │ deno_core / V8
  ├─ Route Filter and Score     │ JS/TS Plugins
  ├─ Credential Selection       │
  ├─ Request Middleware        ─┘
  ├─ Provider Plugin
  └─ Host HTTP Client
  │
  ▼
Upstream
  │
  ▼
Canonical Response Event Stream
  │
  ├─ Response Middleware
  ├─ Usage and Audit
  └─ Protocol Encoder
  │
  ▼
JSON / SSE
```

管理面与数据面共用 Rust 进程，但使用不同路由、权限、超时和并发预算。

## 5. Rust Workspace 划分

建议新核心使用独立 Cargo Workspace，不在首期直接替换现有 Go 目录。

```text
rust-gateway/
├─ Cargo.toml
├─ crates/
│  ├─ axon-gatewayd/
│  ├─ axon-core/
│  ├─ axon-ir/
│  ├─ axon-protocol-openai/
│  ├─ axon-pipeline/
│  ├─ axon-plugin-api/
│  ├─ axon-plugin-runtime/
│  ├─ axon-host-http/
│  ├─ axon-config/
│  ├─ axon-secrets/
│  ├─ axon-storage/
│  ├─ axon-observability/
│  └─ axon-management/
├─ sdk/
│  └─ typescript/
├─ plugins/
│  ├─ provider-openai-compatible/
│  └─ examples/
└─ tests/
   ├─ fixtures/
   ├─ golden/
   └─ replay/
```

### 5.1 `axon-gatewayd`

负责：

- Tokio runtime；
- HTTP 服务启动；
- 信号处理；
- 当前 Runtime Generation；
- 优雅排空；
- 管理面和数据面装配。

不承载供应商逻辑和插件业务代码。

### 5.2 `axon-core`

包含稳定领域类型：

- `RequestContext`；
- `GatewayError`；
- `PluginId`；
- `PluginInstanceId`；
- `GenerationId`；
- `CapabilitySet`；
- `DeadlineBudget`；
- `CancellationToken`；
- `DecisionTrace`。

### 5.3 `axon-ir`

包含统一请求和响应事件模型，不依赖 HTTP 框架、V8 或供应商 SDK。

### 5.4 `axon-protocol-openai`

负责两种下游协议的解析与编码：

- Chat Completions；
- Responses。

协议入口与供应商插件完全解耦。客户端采用 OpenAI 格式，不代表上游必须使用 OpenAI 协议。

### 5.5 `axon-pipeline`

负责固定请求阶段、插件排序、Patch 合并、候选路由、重试、回退、流式提交点和请求终态。

### 5.6 `axon-plugin-api`

定义 Rust 侧稳定契约和可序列化 DTO。第三方插件只依赖对应 TypeScript SDK，不接触 Rust crate。

### 5.7 `axon-plugin-runtime`

基于 `deno_core`，负责：

- V8 Isolate；
- ES Module 加载；
- Rust Ops；
- 插件生命周期；
- 超时终止；
- 堆内存限制；
- Source Map；
- Inspector 开发模式；
- 模块缓存；
- 热重载代际。

### 5.8 `axon-host-http`

统一处理插件发起的网络请求：

- DNS 和连接；
- 代理；
- TLS；
- 域名权限；
- SSRF 防护；
- 重定向检查；
- 超时；
- 响应体限制；
- SSE 流；
- 密钥注入；
- 上游日志与指标。

### 5.9 `axon-storage`

首期可采用 SQLite，抽象以下存储：

- 插件清单；
- 插件实例配置；
- 策略脚本；
- 插件私有 KV；
- 配置版本；
- 请求审计与决策链。

## 6. 统一内部 IR

当前 Go `llm.Request` 可以作为迁移参考，但 Rust 新核心不直接复制 OpenAI Chat Completions Schema。

### 6.1 请求模型

```rust
struct GatewayRequest {
    request_id: RequestId,
    operation: Operation,
    model: ModelRef,
    input: Vec<InputItem>,
    instructions: Option<String>,
    tools: Vec<ToolDefinition>,
    generation: GenerationConfig,
    stream: bool,
    metadata: BTreeMap<String, ScalarValue>,
    provider_options: BTreeMap<String, JsonValue>,
    source: SourceDescriptor,
}

struct ModelRef {
    requested: String,
    canonical: Option<String>,
    upstream: Option<String>,
}
```

模型名的三个阶段必须保留：

- `requested`：客户端传入，如 `gpt-5.6-sol`；
- `canonical`：网关逻辑模型，如 `openai:gpt-5.6-sol`；
- `upstream`：某个具体渠道需要的名称，如 `openai/gpt-5.6-sol`。

### 6.2 响应事件

```rust
enum ResponseEvent {
    Started(ResponseStarted),
    ItemStarted(ItemStarted),
    TextDelta(TextDelta),
    ToolArgumentsDelta(ToolArgumentsDelta),
    ItemCompleted(ItemCompleted),
    UsageUpdated(UsageUpdate),
    Completed(ResponseCompleted),
    Failed(ResponseFailed),
}
```

非流式响应由事件聚合得到，SSE 直接消费事件流。工具参数增量先作为字符串片段保存，完成后再校验 JSON。

## 7. 插件体系

第一层和第二层都使用 JavaScript/TypeScript，同一运行时、同一 SDK、不同能力档位。

### 7.1 第一层：Policy Script

特点：

- 绑定单个 Hook；
- 没有后台任务；
- 没有长期资源；
- 权限少；
- 可以在线编辑；
- 保存后快速切换；
- 默认执行预算 5～20 ms；
- 适合模型映射、时间规则、余额阈值、候选过滤和评分。

示例：

```ts
export default definePolicy({
  id: "night-only-channel",
  hook: "route.filter",

  async evaluate(ctx) {
    const hour = new Date(ctx.now).getHours()
    if (hour >= 22 || hour < 8) {
      return { action: "allow" }
    }

    return {
      action: "deny",
      reason: "channel_available_only_at_night",
    }
  },
})
```

策略决策统一为：

```ts
type PolicyDecision =
  | { action: "abstain" }
  | { action: "allow"; score?: number; metadata?: JsonObject }
  | { action: "deny"; reason: string; retryAfterMs?: number }
  | { action: "patch"; patch: RequestPatch; reason?: string }
```

### 7.2 第二层：Application Plugin

特点：

- 有 `manifest`；
- 有 `setup`、`reconfigure`、`dispose`；
- 可以注册多个 Capability；
- 可以维护连接和受控状态；
- 可以注册管理 API；
- 可以拥有定时任务；
- 可以声明插件依赖；
- 可以贡献配置 Schema 和管理页面资源。

示例：

```ts
export default definePlugin({
  manifest: {
    id: "provider-openai-compatible",
    version: "0.1.0",
    pluginApi: "^1.0.0",
    permissions: {
      network: ["https://api.openai.com:443"],
      secrets: ["provider:*"],
      storage: ["instance-kv"],
    },
  },

  async setup(ctx) {
    const provider = ctx.providers.register({
      id: "openai-compatible",
      // ...
    })

    return {
      async dispose() {
        provider.dispose()
      },
    }
  },
})
```

### 7.3 Capability 分类

完整插件可以贡献：

- `protocol`：新增下游协议；
- `provider`：调用上游；
- `credential-provider`：认证、刷新和凭证字段；
- `model-provider`：模型发现和能力元数据；
- `model-resolver`：逻辑模型解析；
- `router`：供应商和渠道候选；
- `scheduler`：候选中的具体凭证或账户；
- `policy`：允许、拒绝、评分和义务；
- `request-middleware`：请求改写；
- `response-middleware`：响应事件改写；
- `usage-observer`：Usage 和成本扩展；
- `management-extension`：管理 API 和 UI 资源；
- `background-service`：后台刷新和周期任务。

## 8. 请求 Hook

建议首期固定以下阶段：

```text
request.received
request.authenticated
request.normalize
model.resolve
route.candidates
route.filter
route.score
route.selected
credential.selected
request.before_translate
request.before_upstream
response.after_upstream
response.event
response.complete
request.error
request.finished
```

每个 Hook 都有独立输入和输出，不提供万能 `onRequest(ctx)`。

### 8.1 Router 与 Scheduler 分工

- Router 选择 Provider、渠道和上游模型；
- Scheduler 在候选凭证或账号中选择具体执行者。

模型映射、渠道选择和账户负载均衡不能挤在同一个函数里。

### 8.2 Patch 模型

插件接收只读快照，返回结构化 Patch：

```ts
interface RequestPatch {
  setModel?: string
  setHeaders?: Record<string, string>
  removeHeaders?: string[]
  setProviderOptions?: Record<string, unknown>
  removeProviderOptions?: string[]
  messageOperations?: MessageOperation[]
  metadata?: Record<string, JsonValue>
}
```

Rust 核心负责：

- 权限校验；
- Patch 合并；
- 冲突检测；
- 前后值记录；
- 插件 ID 和版本记录；
- IR 不变量校验。

插件不能直接修改共享 Context。

## 9. deno_core 运行时

### 9.1 为什么首期只用 deno_core

- 单个 Rust 可执行程序即可携带 V8；
- JS/TS 策略和完整插件共享一套语义；
- 插件能力全部通过 Rust Ops 暴露，权限边界清楚；
- 无需维护 Node sidecar、RPC、进程监管和 npm 兼容；
- 3～4 个并发下，V8 执行开销不会成为主要瓶颈；
- AI 请求延迟主要来自网络和模型生成。

### 9.2 不承诺的兼容范围

插件运行环境应明确标注：

> AxonHub Plugin Runtime 支持标准 ES Module、现代 JavaScript、编译后的 TypeScript 和 AxonHub Plugin API，不保证 Node.js API 与 npm 包兼容。

首期提供：

- ES202x；
- Promise；
- async/await；
- URL；
- TextEncoder/TextDecoder；
- AbortController；
- 受控 `fetch`；
- 必要的 ReadableStream 子集；
- `console` 的受控实现；
- 定时器；
- Source Map；
- 开发模式 Inspector。

首期不提供：

- `process`；
- `require`；
- `node:*`；
- 任意文件读取；
- 任意 Socket；
- 动态下载模块；
- 原生扩展。

### 9.3 Isolate 模型

推荐：

- 每个完整插件版本一个 Isolate；
- 简单策略按信任域共享 Isolate，后期再按数据评估；
- 每个调用拥有独立 Request Scope；
- 跨请求状态必须通过插件资源或 Host KV；
- 超时、OOM、未处理异常后销毁该 Isolate；
- 插件代码与配置形成不可变代际。

个人低并发场景不需要复杂 Isolate 池。先保证可诊断和正确回收。

### 9.4 Rust Ops

建议按能力拆分 Ops：

```text
op_log
op_clock_now
op_kv_get
op_kv_set
op_http_request
op_http_stream_open
op_http_stream_read
op_http_stream_close
op_secret_bind
op_model_execute
op_register_capability
op_emit_metric
op_register_timer
op_cancel_timer
```

插件只能调用 manifest 已授权的 Op。Rust Op 每次调用都校验：

- 插件实例；
- Generation；
- 租户范围；
- 权限；
- Deadline；
- 输入和输出大小；
- 调用频率。

## 10. Provider 插件

Provider 插件负责：

- 声明上游协议和能力；
- 生成 URL、Header 模板和请求体；
- 通过 Host HTTP 发起请求；
- 解析 JSON 或 SSE；
- 生成统一 `ResponseEvent`；
- 归一化错误与 Usage；
- 可选模型发现和健康检查。

Provider 插件不直接掌握：

- 客户端 Socket；
- 全部密钥；
- 全局重试次数；
- 全局熔断状态；
- 其他租户配置；
- Rust 数据库连接。

首个官方插件建议为：

```text
@axonhub/provider-openai-compatible
```

它必须通过公开插件 API 工作，不使用私有 Host Op。这样才能验证插件 API 足以承载真实供应商。

## 11. Host HTTP 与密钥

插件配置保存 `SecretRef`，插件运行时得到不透明 `SecretHandle`。

```text
plugin
  │ request template + secret handle
  ▼
Rust Host HTTP
  ├─ permission check
  ├─ destination check
  ├─ secret injection
  ├─ proxy/TLS
  ├─ timeout
  ├─ logging/metrics
  └─ upstream
```

Host HTTP 必须检查：

- Scheme；
- Host 和端口；
- DNS 解析后的 IP；
- 重定向的每一跳；
- loopback、link-local、私网和 metadata 地址；
- 最大请求和响应体；
- 超时；
- 插件声明的网络范围；
- Secret 与目标域名绑定关系。

插件应尽量看不到明文 Key。

## 12. 流式响应

Rust 核心统一拥有流式状态机：

- 有界 Channel；
- 背压；
- 客户端断开；
- 上游取消；
- 首字节提交点；
- Usage 汇总；
- 终止事件；
- 重试边界。

插件跨 V8 边界时以完整语义事件为单位，不处理任意网络碎片。

可采用以下批处理条件：

- 完整 SSE Event；
- 累计达到 4 KiB；
- 等待达到 10～20 ms；
- 工具调用边界；
- Usage 或终止事件。

响应尚未提交给客户端时，可按照核心重试策略切换候选。已经发送首个有效事件后，默认不再透明切换 Provider。

## 13. 插件排序与冲突

每个 Hook 支持：

- `priority`；
- `before`；
- `after`；
- `requires`；
- `conflicts`。

加载时构建 DAG。发现循环依赖、缺失依赖或互斥插件时，拒绝发布候选代际。

合并规则建议：

- `deny` 立即终止当前决策链；
- `abstain` 不改变结果；
- 路由分数累加；
- Header 后执行者覆盖，完整保留审计；
- 模型最终映射只允许一个插件处理，或由显式优先级决定；
- Body Patch 顺序执行，每次执行后重新校验；
- 核心安全策略不能被插件覆盖。

## 14. 插件生命周期与资源

完整插件生命周期：

```text
install
→ validate
→ compile
→ load
→ setup
→ health-check
→ active
→ reconfigure / reload
→ draining
→ dispose
→ unloaded
```

所有资源登记在插件 Scope 中：

- 定时器；
- HTTP Stream；
- KV 句柄；
- 模型执行；
- 管理路由；
- 事件订阅；
- 后台任务；
- 子插件或子服务。

插件退出时，Rust 按 Scope 统一取消和释放，避免热重载后残留旧监听器与旧定时任务。

## 15. 热重载

采用 Generation 代际切换：

```rust
struct RuntimeGeneration {
    id: GenerationId,
    config: Arc<ValidatedConfig>,
    plugin_registry: PluginRegistry,
    pipeline: PipelineGraph,
    created_at: Instant,
}
```

全局通过 `ArcSwap<RuntimeGeneration>` 持有当前代际。

发布流程：

1. 读取新源码和配置；
2. TypeScript 编译成单文件 ESM；
3. 校验 manifest 和配置 Schema；
4. 创建新 Isolate；
5. 执行 `setup`；
6. 运行健康检查；
7. 构建 Hook DAG；
8. 执行路由模拟和 Smoke Test；
9. 原子切换新代际；
10. 新请求进入新代；
11. 旧请求继续使用旧代；
12. 旧代在途请求清零后执行 `dispose`；
13. 超时后终止旧 Isolate。

流式请求即使持续十分钟，也不会中途切换插件实现。

## 16. 权限模型

插件 manifest 申请权限，管理员决定最终授权。

```json
{
  "permissions": {
    "request": ["metadata", "content"],
    "response": ["content", "usage"],
    "network": ["https://api.example.com:443"],
    "secrets": ["provider:example:*"],
    "storage": ["instance-kv"],
    "management": ["routes"],
    "background": ["timers"],
    "model": ["execute"]
  }
}
```

有效权限为：

```text
插件申请 ∩ 管理员批准 ∩ 部署策略允许 ∩ 当前租户允许
```

插件升级后新增权限时暂停自动启用，要求管理员重新确认。

## 17. 故障处理

分别记录以下故障域：

- 插件版本；
- 插件实例；
- Provider；
- Endpoint；
- 模型；
- 凭证；
- 租户与渠道组合。

不能因为某个租户余额不足就熔断整个 Provider。

插件故障处理：

- 单次异常：记录并返回规范化错误；
- 连续超时：暂停该插件实例；
- 未处理 Promise：该请求失败，并根据严重程度淘汰 Isolate；
- V8 OOM：销毁 Isolate；
- 无限循环：到达 Deadline 后 `TerminateExecution`；
- 插件被熔断：路由阶段排除其能力；
- 官方必需插件不可用：进入可诊断的 Degraded 状态。

## 18. 可观测性与决策链

每次请求应生成 `DecisionTrace`：

```text
Generation
→ Inbound Protocol
→ Requested Model
→ Canonical Model
→ Candidate Providers
→ Policy Decisions
→ Score Changes
→ Excluded Candidates and Reasons
→ Selected Provider/Credential
→ Upstream Model
→ Applied Request Patches
→ Attempts
→ Usage
→ Final Outcome
```

管理界面重点展示：

- 哪个插件在什么阶段执行；
- 执行耗时；
- 输入摘要；
- 返回的 Patch 或决策；
- Patch 前后差异；
- 候选被排除的原因；
- 插件错误和熔断状态；
- 当前请求绑定的 Generation。

Prompt、密钥和敏感 Header 默认脱敏。

## 19. 插件开发体验

建议提供 CLI：

```text
axonhub plugin create
axonhub plugin dev
axonhub plugin check
axonhub plugin test
axonhub plugin pack
axonhub route simulate
axonhub request replay
axonhub plugin inspect-permissions
```

开发模式：

```text
TypeScript source
→ esbuild incremental build
→ source map
→ new Isolate
→ setup and health check
→ atomic generation switch
```

插件测试工具应提供：

- 模拟时间；
- 模拟余额接口；
- 模拟 Host HTTP；
- 模拟流式上游；
- 路由候选断言；
- Patch diff；
- 权限拒绝断言；
- 超时和取消测试；
- Golden Request/Response fixture。

## 20. 与现有 Go AxonHub 的关系

当前仓库已经拥有成熟控制面、数据库、RBAC、前端和 LLM Pipeline。直接用 Rust 全量替换会同时引入语言迁移、插件 API、协议 IR、数据库兼容和部署变化，风险过高。

推荐采用并行验证：

### 阶段 A：Rust 数据面原型

- 独立目录或独立仓库开发；
- 只实现 Chat Completions 和 Responses；
- 只提供本地静态配置；
- 通过官方 OpenAI Compatible 插件完成真实请求；
- 使用现有 AxonHub 请求样本做回放。

### 阶段 B：控制面桥接

- Go AxonHub 继续负责用户、渠道、模型、密钥和管理 UI；
- Rust 数据面读取控制面发布的不可变配置快照；
- 两者通过本地 Unix Socket、HTTP 或配置文件进行窄接口通信；
- 不让 Rust 直接复用 Ent 数据库结构。

### 阶段 C：流量对照

- 相同请求同时进入 Go Pipeline 和 Rust Pipeline 的 Shadow 模式；
- 比较协议输出、Usage、错误分类、路由选择和流事件；
- 逐步放入低风险个人流量；
- 保留快速切回 Go 数据面的路径。

### 阶段 D：再决定迁移边界

验证结果稳定后再选择：

- Rust 仅作为高性能插件数据面，Go 控制面长期保留；
- 继续迁移控制面；
- 保持双进程架构。

目前不预先决定最终一定全量替换。

## 21. 分阶段实施

### Phase 0：契约设计

交付：

- IR v0；
- Response Event v0；
- Plugin Manifest v0；
- Hook 顺序；
- Policy Decision；
- Request Patch；
- Provider 接口；
- Decision Trace。

验收：

- 能表达现有 AxonHub 的主要文本生成请求；
- Chat Completions 和 Responses 都能映射；
- 模型三阶段命名明确；
- 流式事件顺序有不变量测试。

### Phase 1：Rust 网关骨架

交付：

- Tokio + Axum/Hyper；
- 两个 OpenAI 下游入口；
- 非流式与 SSE；
- 固定 Pipeline；
- 取消和 Deadline；
- 静态配置；
- 暂时使用测试 Provider。

验收：

- 客户端断开可取消上游；
- SSE 有背压；
- 响应提交前后故障行为明确；
- 两种入口共用内部 IR。

### Phase 2：deno_core 与插件 API

交付：

- 插件加载；
- ES Module；
- TypeScript 编译；
- Manifest；
- Policy Script；
- Router；
- Request Middleware；
- Provider；
- Host HTTP；
- Source Map；
- 超时和 Isolate 回收。

验收：

- JS 插件实现模型改名；
- JS 插件实现夜间渠道规则；
- JS 插件通过受控接口查询余额；
- JS Provider 完成真实 SSE 请求；
- 无限循环不会永久占用网关；
- 插件无法读取主机文件和环境变量。

### Phase 3：官方 OpenAI Compatible 插件

交付：

- 插件配置 Schema；
- Chat Completions 上游；
- Responses 上游；
- 自定义 Base URL；
- 模型发现；
- 错误与 Usage 标准化；
- 流式解析；
- 密钥句柄。

验收：

- 该插件只使用公开 API；
- 第三方中转站可以通过配置接入；
- 同一逻辑模型可映射到多个上游模型名；
- 失败候选可在响应提交前回退。

### Phase 4：热重载与插件工作台

交付：

- Generation；
- `ArcSwap`；
- 插件代际切换；
- 在线 TypeScript 编辑；
- 路由模拟；
- Patch Diff；
- 请求回放；
- 插件健康状态。

验收：

- 保存脚本后新请求使用新版本；
- 在途流继续使用旧版本；
- 新代初始化失败时旧代不受影响；
- 管理页面能解释路由决策。

### Phase 5：现有 AxonHub 对接

交付：

- 控制面配置快照；
- 密钥引用；
- 用户与租户身份；
- Shadow Replay；
- Go/Rust结果对比工具；
- 快速切换方案。

验收：

- 不修改旧数据库即可运行原型；
- 同一请求可对比两套 Pipeline；
- 个人流量可安全切换；
- 回滚无需数据库降级。

## 22. 后续扩展触发条件

### 22.1 何时考虑 Node sidecar

满足下列情况之一再评估：

- 多个关键插件强依赖官方 Node SDK；
- 自建 Web API 兼容层成本持续上升；
- 插件作者明确需要大量 npm 生态；
- deno_core 中实现某项 Node 兼容能力的维护成本超过 sidecar；
- 第三方插件需要与主进程形成更强故障隔离。

### 22.2 何时考虑 Wasm/WASI

满足下列情况之一再评估：

- 出现明确的 CPU 热点；
- 需要 Rust、Zig、TinyGo 等语言编写插件；
- 需要稳定的跨语言二进制分发；
- 需要比进程内 V8 更严格的资源边界；
- 有可信的 WIT 接口已经在 JS 插件中验证稳定。

任何扩展都先由性能数据、插件需求和维护成本驱动。

## 23. 体积预期

采用 Rust + deno_core 单运行时后，Release、LTO、strip 条件下可先按以下范围估算：

- Rust 网关核心及常规依赖：20～50 MiB；
- V8/deno_core 增量：40～100 MiB；
- 管理前端静态资源：3～15 MiB；
- 总可执行文件或发行包：约70～160 MiB。

具体结果必须在最小原型阶段实测。当前没有 Node Runtime 和 Wasmtime，不应接近四百 MiB。若采用动态资源包、压缩前端和精简 ICU，发行体积还有下降空间。

## 24. 当前推荐技术栈

- Rust stable；
- Tokio；
- Axum + Hyper；
- `deno_core`；
- `serde` / `serde_json`；
- `bytes`；
- `futures` / `tokio-stream`；
- `reqwest` 或基于 Hyper 的受控 Host HTTP；
- `tower` 用于核心服务层和超时；
- `arc-swap` 用于 Runtime Generation；
- `tokio-util::sync::CancellationToken`；
- `tracing` + OpenTelemetry；
- SQLite 首期存储；
- esbuild 用于插件 TypeScript 构建；
- JSON Schema 用于插件配置。

## 25. 架构红线

1. 插件不能直接依赖 Rust 内部 crate。
2. 插件不能直接修改共享 Request Context。
3. 插件不能获得数据库超级权限。
4. 插件不能默认读取明文密钥。
5. 插件网络必须经过 Host HTTP。
6. V8 Isolate 不视为完整安全容器。
7. 超时或异常后的 Isolate不再复用。
8. 热重载不改变在途请求绑定的插件代际。
9. 流式事件不按任意字节碎片跨 V8 边界。
10. 响应提交后不透明切换 Provider。
11. Router、Scheduler、Policy 和 Middleware 不合成一个万能 Hook。
12. 官方供应商插件不得使用第三方插件无法获得的私有能力。
13. 首期不追求 Node/npm 兼容。
14. 没有实测瓶颈前不引入 Node sidecar 和 Wasm。
15. 现有 Go AxonHub 的数据库与控制面不能被一次性破坏式替换。

## 26. 下一步

建议先完成四份契约文档，再写 Rust 代码：

1. `canonical-ir-v0.md`：请求、响应和流式事件；
2. `plugin-api-v0.md`：Manifest、Capability、生命周期和权限；
3. `pipeline-v0.md`：Hook顺序、Patch合并、路由与重试；
4. `runtime-v0.md`：deno_core模块加载、Rust Ops、Isolate和热重载。

契约评审通过后，用最小Rust原型验证三条真实路径：

- Chat Completions → OpenAI Compatible插件 → 非流式响应；
- Responses → OpenAI Compatible插件 → SSE；
- 夜间策略＋余额策略＋模型映射插件共同参与一次路由。

这三条跑通，再开始接管理界面和现有AxonHub控制面。
