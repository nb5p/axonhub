---
id: channel-429-non-retryable
title: 渠道 HTTP 429 不重试开关
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca
  adopted_commits: []
  last_checked_commit: 49ade6f279eae7aed46858dc121258e922ec9870
  last_checked_at: 2026-08-21
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 621b04052ab69434df631119de7e604635389da0
    - e325b77369b1d1e5bd81a2ce14572367efe76f63
    - 4896e8b908aab60662eb0cb32d74749e775dd8ed
  modules:
    - internal/objects/channel.go
    - internal/server/gql/axonhub.graphql
    - internal/server/orchestrator/retry.go
    - internal/server/orchestrator/outbound.go
    - internal/server/orchestrator/rate_limit_tracking.go
    - llm/pipeline/pipeline.go
    - frontend/src/features/channels/components/channels-action-dialog.tsx
    - frontend/src/features/channels/data/channels.ts
    - frontend/src/features/channels/data/schema.ts
    - frontend/src/features/channels/utils/merge.ts
    - frontend/src/locales/en/channels.json
    - frontend/src/locales/zh-CN/channels.json
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-21
reconciliations: []
history_rewrites: []
database:
  impact: additive
  backward_compatible: true
---

# 渠道 HTTP 429 不重试开关

## 目的

让管理员为特定渠道启用 `treat429AsNonRetryable`。启用后，该渠道实际返回 HTTP 429 时，不重试同一渠道、同渠道其他模型或后续候选渠道/模型；该标记不会直接决定向调用方返回的错误内容。

上游错误是否向调用方透传、脱敏或改写仍由全局“上游错误策略”决定。开关默认关闭，保持所有既有渠道（包括官方 OAuth 订阅渠道）的 429 重试、候选切换和配额感知行为；开启该开关时，即使上游返回 `Retry-After` 也不会创建渠道冷却。

## 来源与采用范围

这是本地需求的原创实现，没有采用外部项目代码。设计仅扩展渠道设置和已有重试流水线，不修改全局 `429 => retryable` 的默认语义。

## 本地实现

- `ChannelSettings` 增加可选 JSON 布尔字段 `treat429AsNonRetryable`，GraphQL 的输入/输出类型和渠道创建、复制、更新、列表读取链路均包含该字段。
- 渠道编辑弹窗在“重试状态码”正上方提供“HTTP 429 不重试”复选开关，明确其跳过重试、候选切换和冷却；是否将上游错误暴露给调用方由系统上游错误策略决定。
- pipeline 的可选 `RetryTerminator` 只终止重试循环；当前渠道针对配置命中的 429 不进行同渠道重试或跨渠道切换，之后由 API 层按系统错误策略生成下游响应。
- 429 冷却追踪复用该重试终止决策：开启开关时，即使上游提供 `Retry-After` 也不创建内存冷却状态。
- 未新增后台任务、轮询、数据库查询或外部请求；开关关闭时只有一次可选接口断言，不改变原有重试路径。

## 与来源的差异

上游已有 429 默认可重试、同渠道跳过及 Retry-After 冷却机制。本地增加渠道级“禁止重试且不创建冷却”覆盖，并保留全局上游错误策略作为兼容路径。

## 上游收敛

2026-08-21 比较 `upstream/unstable@49ade6f279eae7aed46858dc121258e922ec9870`：搜索渠道级 429 非重试设置、相关 GraphQL 字段和重试路径，未发现同类实现，关系为 `none`。

## 数据库兼容

兼容等级为 `additive`：

- 仅在既有 `channels.settings` JSON 列增加可选字段；缺失值按 `false` 处理，不需要数据回填或 Ent 迁移。
- 新版本可直接读取旧数据库。旧版本会忽略未知 JSON 字段；若回退后用旧版本编辑已开启此设置的渠道，字段可能因整段设置重写而丢失。
- 本次没有创建或修改部署数据库，也没有创建备份。部署前按蓝绿/SQLite 既有流程制作一致性快照并校验；回滚时可恢复快照，或接受仅丢失该开关配置。

## 验证

- `make generate`：GraphQL 代码生成成功。
- `go test ./internal/server/orchestrator -run '^(TestRateLimitTracking_OnOutboundRawError_ConfiguredNonRetryable429DoesNotCoolDown|TestPersistentOutboundTransformer_ShouldStopRetry_OnConfigured429)$' -count=1`：通过，覆盖禁止重试和配置命中时不创建冷却。
- `cd llm && go test ./pipeline -run 'TestPipeline_Process_RetryLogic' -count=1`：通过，覆盖终止信号阻止同渠道和跨渠道重试。
- `go test ./internal/server/api -run '^TestApplyUpstreamErrorPolicy_' -count=1`：通过，覆盖系统上游错误策略的透传和改写。
- 未运行 lint、前端 build 或重启开发服务器，符合仓库规则。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-20 | 本地需求；上游比较至 `9fb6f1af` | `621b04052ab69434df631119de7e604635389da0` | original：以渠道级开关终止当前请求的 429 重试和候选切换，同时跳过内存冷却。 |
| 2026-08-21 | `upstream/unstable@49ade6f2` | `e325b77369b1d1e5bd81a2ce14572367efe76f63` | 修正该开关为仅禁止重试：保留系统上游错误策略和 `Retry-After` 冷却。 |
| 2026-08-21 | 用户澄清原逻辑 | `4896e8b908aab60662eb0cb32d74749e775dd8ed` | 恢复原有“配置命中不创建冷却”的逻辑；仅保留“下游可见性由系统错误策略决定”的文案澄清。 |
