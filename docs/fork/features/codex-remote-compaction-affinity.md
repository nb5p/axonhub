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
  adopted_commits: []
  last_checked_commit: c204d33b09ebfefe96c1d4dcb16a88590992257e
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
    - llm/transformer/openai/codex/headers.go
    - llm/transformer/openai/codex/headers_test.go
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

## 本地实现

- 对明确声明 `X-Codex-Beta-Features: remote_compaction_v2` 的请求自动启用 Codex 会话 trace 提取，不改变其他请求在 `codex_trace_enabled=false` 时的原行为。
- 会话标识优先读取 `Session_id`、`Session-Id` 和 `X-Codex-Turn-Metadata`；Responses 请求缺少这些头时回退到 `prompt_cache_key`。读取请求体后会恢复 body，供后续转换器继续使用。
- Responses 入站转换识别到原始 `compaction_trigger` 后，压缩请求优先选择该 trace 最近成功使用的合法渠道。该协议亲和优先于普通 trace sticky 配置和“优先透传”排序。
- 渠道内存在多个启用 API Key 时，复用既有 `TraceStickyKeyProvider`，同一稳定 trace 通过一致性哈希选择同一凭证。
- 热点影响仅限远程压缩请求：有会话头时不读取 body；缺少会话头时额外读取一次 body 并立即恢复。不新增后台任务、数据库查询类型或外部请求。

## 与来源的差异

- Sub2API 使用独立的带 TTL `session hash → account ID` 缓存；AxonHub 直接复用已有 trace 实体、最近成功渠道查询和渠道 API Key 粘性选择，不引入第二套账号缓存。
- AxonHub 仅对明确声明远程压缩能力的 Codex 请求自动建立稳定 trace，不全局开启 `codex_trace_enabled`。
- 当原渠道不再属于合法候选时保留 AxonHub 原有 fallback；fallback 后的跨账号兼容性取决于上游，不能视为 OpenAI 已承诺支持。

## 上游收敛

- 2026-08-15 比较 `upstream/unstable@852b8c6f9d3953309df55d4c7872e12fbfbd45f3`。
- 上游已包含 Codex 压缩格式和 `prompt_cache_key` 的基础转换能力，但 trace 中间件仍只受全局 `CodexTraceEnabled` 控制，没有远程压缩账号亲和；关系为 `none`。

## 数据库兼容

- 兼容等级：`none`。
- 不修改 Ent Schema、索引或现有数据；旧数据库可直接使用。
- 回滚只需回退代码提交，不需要数据库恢复。

## 验证

- 中间件回归：全局 Codex trace 关闭时，同一远程压缩 Session 的两次请求仍使用同一 trace。
- 请求体回归：缺少 Session 头时使用 `prompt_cache_key`，后续处理仍能读取完整 body。
- 路由回归：`compaction_trigger` 在普通 sticky 关闭、且其他渠道满足“优先透传”时，仍优先原渠道；原渠道不在合法候选时保留 fallback。
- 渠道 API Key 回归：既有多 Key trace 粘性测试继续通过。
- 执行：

  ```sh
  go test ./internal/server/middleware ./internal/server/orchestrator ./internal/server/biz \
    -run 'TestWithTrace_Codex(RemoteCompaction|Disabled|Header|Turn)|TestLoadBalancedSelector_TraceStickySelection|TestTraceStickyKeyProvider_MultipleKeys_WithTrace_Sticky' \
    -count=1

  cd llm
  go test ./transformer/openai/codex ./transformer/openai/responses -count=1
  ```

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-15 | Sub2API `main@c204d33b` | `600717a4e8430c685902f8c340c30699fc63b9b7` | 重新实现会话标识与账号亲和设计；复用 AxonHub trace、历史渠道和多 Key 一致性哈希。 |
