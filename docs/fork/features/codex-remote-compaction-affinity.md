---
id: codex-remote-compaction-affinity
title: Codex 远程压缩账号亲和
status: active
origin: donor-port
integration_method: reimplemented
source:
  repository: https://github.com/Wei-Shaw/sub2api
  branch: main
  baseline_commit: c204d33b09ebfefe96c1d4dcb16a88590992257e
  adopted_commits:
    - 8ae6d8f67e72b099ed581b1455840ca62bb25561
    - 8219dcfc87ac270fe11414a98e326b06f5b4309f
    - 9662cff2e7c62fcfb99111415f5ac11a15748e14
    - a8b9ea22b701704507fa597c03c1835173248f36
  last_checked_commit: baeac1f3de21d37b129405f092ef86c24b3f203d
  last_checked_at: 2026-08-15
  license: LGPL-3.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 600717a4e8430c685902f8c340c30699fc63b9b7
  modules:
    - internal/server/middleware/trace.go
    - internal/server/middleware/trace_test.go
    - internal/server/orchestrator/candidates.go
    - internal/server/orchestrator/candidates_sticky_test.go
    - internal/server/orchestrator/codex_turn_state.go
    - internal/server/orchestrator/codex_turn_state_test.go
    - internal/server/orchestrator/orchestrator.go
    - internal/server/orchestrator/select_endpoints.go
    - internal/server/orchestrator/select_endpoints_test.go
    - internal/server/orchestrator/state.go
    - internal/server/api/chat.go
    - internal/server/biz/channel_llm.go
    - llm/httpclient/client.go
    - llm/httpclient/model.go
    - llm/pipeline/pipeline.go
    - llm/transformer/openai/codex/headers.go
    - llm/transformer/openai/codex/headers_test.go
    - llm/transformer/openai/codex/outbound.go
    - llm/transformer/openai/codex/outbound_executor_test.go
    - llm/transformer/openai/codex/codex_simulator_test.go
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-15
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# Codex 远程压缩账号亲和

## 目的

当 Codex 启用 `remote_compaction_v2` 时，让同一会话的普通 Responses 请求和后续远程压缩请求保持稳定 trace，并让压缩请求优先回到最近成功处理该会话的渠道及渠道 API Key。这样可以避免多 OpenAI 账号环境中，压缩请求漂移到另一个账号后无法复用响应 ID、提示缓存或不透明加密内容。

原渠道已经禁用、不再支持模型或被其他请求级规则排除时仍允许正常 fallback；该功能不会把失效账号永久锁死，也不保证不同 OpenAI 账号之间的压缩上下文可互换。

## 来源与采用范围

- 参考 Sub2API `main@c204d33b09ebfefe96c1d4dcb16a88590992257e` 的稳定会话标识和 `session → account` 粘性调度设计。
- 参考其对 Codex `remote_compaction_v2` 的识别方式，确认带 `compaction_trigger` 的请求应继续使用原生流式 `/responses` 语义。
- 来源仓库为 LGPL-3.0；本地没有复制或 cherry-pick 来源代码，只采用行为设计并基于 AxonHub 现有 trace、历史渠道缓存和渠道 API Key 一致性哈希重新实现。

### 2026-08-15 Sub2API v0.1.177（`c204d33b..baeac1f3`）检查范围

- `8ae6d8f6` 会话级 beta 功能头：原生压缩请求确保 `remote_compaction_v2`，官方客户端缺省时补齐。→ 本地实现 `EnsureBetaFeature`/`applyBetaFeatures` 与 `IsOfficialOAuth`。
- `8219dcfc` `x-codex-turn-state` 回传与跨账号回显拦截。→ 本地实现响应头透传、下游写回与溯源守卫。
- `9662cff2`、`a8b9ea22` 原生 v2 保留 `/responses` 端点、新旧压缩路由分离。→ 本地实现候选/端点过滤，只允许 Responses 格式承载原生压缩。
- `fce41e31` 指纹收敛 opt-in 与上游压缩探测迁移：与远程压缩亲和关系不大（会话指纹改写属于另一行为面），且 AxonHub 没有“账号压缩测试”管理入口，不移植。
- 分组用量统计按日汇总：属于用量统计功能，不在本功能范围内，不移植。

## 本地实现

- 对明确声明 `X-Codex-Beta-Features: remote_compaction_v2` 的请求自动启用 Codex 会话 trace 提取，不改变其他请求在 `codex_trace_enabled=false` 时的原行为。
- 会话标识优先读取 `Session_id`、`Session-Id` 和 `X-Codex-Turn-Metadata`；Responses 请求缺少这些头时回退到 `prompt_cache_key`。读取请求体后会恢复 body，供后续转换器继续使用。
- Responses 入站转换识别到原始 `compaction_trigger` 后，压缩请求优先选择该 trace 最近成功使用的合法渠道。该协议亲和优先于普通 trace sticky 配置和“优先透传”排序。
- 渠道内存在多个启用 API Key 时，复用既有 `TraceStickyKeyProvider`，同一稳定 trace 通过一致性哈希选择同一凭证。
- 原生远程压缩 v2 请求只在支持 `openai/responses` 的候选中路由（`select_endpoints.go` 的 `remoteCompactionCapableAPIFormats`、`candidates.go` 的候选过滤），不再改写为旧版压缩或其他格式；无候选时返回明确错误而不是静默降级。
- Codex 出站（`llm/transformer/openai/codex`）在原生压缩请求上确保 `remote_compaction_v2` beta 头；官方 OAuth 渠道在客户端未声明时默认补齐；客户端显式声明时原样保留。
- `x-codex-turn-state` 协议链：上游响应头通过 `httpclient.Request.ResponseHeaders` → pipeline `Result.ResponseHeaders` 透出，`internal/server/api/chat.go` 在写回下游前统一写回该头（非流式、SSE 一致），上游未带时清除可能残留的旧 failover 值。
- 跨账号回显拦截（`internal/server/orchestrator/codex_turn_state.go`）：进程级溯源表记录“下游会话（API Key ID + Codex session）→ 铸造该 blob 的渠道凭证身份”；出站守卫在请求回带 turn-state 且已知由其他凭证铸造时剥离该头，同凭证或未知来源原样透传。溯源在最终成功响应确实携带该头后才提交，失败 attempt 不会污染记录；记录带 24h TTL，读侧惰性清理、每 256 次写入全量清扫。
- 热点影响仅限远程压缩请求：有会话头时不读取 body；缺少会话头时额外读取一次 body 并立即恢复。不新增后台任务、数据库查询类型或外部请求。

## 与来源的差异

- Sub2API 使用独立的带 TTL `session hash → account ID` 缓存；AxonHub 直接复用已有 trace 实体、最近成功渠道查询和渠道 API Key 粘性选择，不引入第二套账号缓存。
- AxonHub 仅对明确声明远程压缩能力的 Codex 请求自动建立稳定 trace，不全局开启 `codex_trace_enabled`。
- 当原渠道不再属于合法候选时保留 AxonHub 原有 fallback；fallback 后的跨账号兼容性取决于上游，不能视为 OpenAI 已承诺支持。
- 溯源身份：Sub2API 用上游账号 ID；AxonHub 用渠道凭证身份（非 OAuth 渠道为选中 API Key，OAuth 渠道为 `Chatgpt-Account-Id` 账号，缺省回退渠道 ID），并在 OAuth 判断优先于共享 context 中可能残留的旧 attempt key。
- 提交时点：Sub2API 在首输出守卫的暂存头真正写盘时才记录铸造账号；AxonHub 在 `Process` 成功后记录——上游响应头在 body 写出前必然已交付客户端，溯源与客户端实际持有值一致；首事件超时等失败 attempt 已在重试路径丢弃，不会进入记录。残留记录只会导致保守剥离（安全方向）。
- 未采用：指纹收敛 opt-in、账号“压缩测试”原生探测管理界面、分组用量按日汇总。

## 上游收敛

- 2026-08-15 比较 `upstream/unstable@852b8c6f9d3953309df55d4c7872e12fbfbd45f3` 与更新后的 `9fb6f1af`。
- 上游已包含 Codex 压缩格式和 `prompt_cache_key` 的基础转换能力，但 trace 中间件仍只受全局 `CodexTraceEnabled` 控制，没有远程压缩账号亲和；关系为 `none`。
- `852b8c6f..9fb6f1af` 之间仅有一笔无关的 codex responses 推理上下文填充修复，未出现同类实现。

## 数据库兼容

- 兼容等级：`none`。
- 不修改 Ent Schema、索引或现有数据；旧数据库可直接使用。
- 回滚只需回退代码提交，不需要数据库恢复。

## 验证

- 中间件回归：全局 Codex trace 关闭时，同一远程压缩 Session 的两次请求仍使用同一 trace。
- 请求体回归：缺少 Session 头时使用 `prompt_cache_key`，后续处理仍能读取完整 body。
- 路由回归：`compaction_trigger` 在普通 sticky 关闭、且其他渠道满足“优先透传”时，仍优先原渠道；原渠道不在合法候选时保留 fallback。
- 渠道 API Key 回归：既有多 Key trace 粘性测试继续通过。
- beta 头回归：原生压缩请求确保 `remote_compaction_v2`；官方 OAuth 渠道缺省补齐；客户端显式声明不被改写。
- turn-state 回归：上游响应携带该头时写回下游；同凭证回带保留、跨凭证回带剥离、无会话或不明确来源保持透传；溯源过期后回到透传。
- 原生压缩端点回归：`compaction_trigger` 只允许 `openai/responses` 候选，无候选时返回错误。
- 执行：

  ```sh
  go test ./internal/server/middleware ./internal/server/orchestrator ./internal/server/biz \
    -run 'TestWithTrace_Codex(RemoteCompaction|Disabled|Header|Turn)|TestLoadBalancedSelector_TraceStickySelection|TestTraceStickyKeyProvider_MultipleKeys_WithTrace_Sticky|TestCodexTurnState|TestChatCompletionWithRequest' \
    -count=1

  cd llm
  go test ./transformer/openai/codex ./transformer/openai/responses -count=1
  ```

  实际执行（2026-08-15）：

  ```sh
  go build ./...
  go test ./internal/server/orchestrator ./internal/server/api ./internal/server/biz ./internal/server/middleware \
    -run 'TestSelectAPIFormat_RemoteCompactionForcesResponses|TestLoadBalancedSelector_TraceStickySelection|TestWithTrace_Codex|TestChatCompletionWithRequest|TestTraceSticky|TestCodexTurnState' -count=1
  go test ./internal/server/orchestrator ./internal/server/api -count=1
  cd llm && go test ./transformer/openai/codex ./transformer/openai/responses \
    -run 'TestCodexOutbound|TestResponsesTransformer' -count=1
  ```

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-15 | Sub2API `main@c204d33b` | `600717a4e8430c685902f8c340c30699fc63b9b7` | 重新实现会话标识与账号亲和设计；复用 AxonHub trace、历史渠道和多 Key 一致性哈希。 |
| 2026-08-15 | Sub2API `main@c204d33b..baeac1f3`（v0.1.177） | 工作区未提交（待审阅后按 `🧩` 约定提交） | 采用 beta 功能头补齐、turn-state 回传与跨账号回显拦截、原生 v2 保留 `/responses` 端点；忽略指纹收敛、压缩测试探测与用量统计。 |
