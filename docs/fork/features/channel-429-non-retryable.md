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
  last_checked_commit: 9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca
  last_checked_at: 2026-08-20
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 621b04052ab69434df631119de7e604635389da0
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
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-20
reconciliations: []
history_rewrites: []
database:
  impact: additive
  backward_compatible: true
---

# 渠道 HTTP 429 不重试开关

## 目的

让管理员为特定渠道启用 `treat429AsNonRetryable`。启用后，该渠道实际返回 HTTP 429 时，当前请求立即将原错误返回给调用方，不重试同一渠道、同渠道其他模型或后续候选渠道/模型。

开关默认关闭，保持所有既有渠道（包括官方 OAuth 订阅渠道）的 429 重试、候选切换和配额感知行为。首版不会从 429 推断额度恢复时间，也不会自动禁用渠道、创建冷却状态或改变下一次新请求的候选选择。

## 来源与采用范围

这是本地需求的原创实现，没有采用外部项目代码。设计仅扩展渠道设置和已有重试流水线，不修改全局 `429 => retryable` 的默认语义。

## 本地实现

- `ChannelSettings` 增加可选 JSON 布尔字段 `treat429AsNonRetryable`，GraphQL 的输入/输出类型和渠道创建、复制、更新、列表读取链路均包含该字段。
- 渠道编辑弹窗在“重试状态码”正上方提供“HTTP 429 不重试”复选开关，并说明其立即返回、不切换候选、保持渠道启用且不创建冷却的语义。
- pipeline 增加可选 `RetryTerminator` 接口；当前渠道针对配置命中的 429 在同渠道重试和跨渠道切换前终止整个重试循环。
- 429 冷却追踪复用同一决策，避免在启用开关时因为 `Retry-After` 创建内存冷却状态。
- 未新增后台任务、轮询、数据库查询或外部请求；开关关闭时只有一次可选接口断言，不改变原有重试路径。

## 与来源的差异

上游已有 429 默认可重试、同渠道跳过及 Retry-After 冷却机制。本地只增加渠道级“立即返回”覆盖，并保留这些默认行为作为兼容路径。

## 上游收敛

2026-08-20 比较 `upstream/unstable@9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca`：搜索渠道级 429 非重试设置、相关 GraphQL 字段和重试路径，未发现同类实现，关系为 `none`。

## 数据库兼容

兼容等级为 `additive`：

- 仅在既有 `channels.settings` JSON 列增加可选字段；缺失值按 `false` 处理，不需要数据回填或 Ent 迁移。
- 新版本可直接读取旧数据库。旧版本会忽略未知 JSON 字段；若回退后用旧版本编辑已开启此设置的渠道，字段可能因整段设置重写而丢失。
- 本次没有创建或修改部署数据库，也没有创建备份。部署前按蓝绿/SQLite 既有流程制作一致性快照并校验；回滚时可恢复快照，或接受仅丢失该开关配置。

## 验证

- `make generate`：GraphQL 代码生成成功。
- `go test ./internal/server/orchestrator -run 'Test(IsRetryableErrorForChannel|PersistentOutboundTransformer_ShouldStopRetry_OnConfigured429|RateLimitTracking_OnOutboundRawError_(429|ConfiguredNonRetryable429DoesNotCoolDown))' -count=1`：通过。
- `cd llm && go test ./pipeline -run 'TestPipeline_Process_RetryLogic' -count=1`：通过，覆盖终止信号阻止同渠道和跨渠道重试。
- 未运行 lint、前端 build 或重启开发服务器，符合仓库规则。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-20 | 本地需求；上游比较至 `9fb6f1af` | `621b04052ab69434df631119de7e604635389da0` | original：以渠道级开关终止当前请求的 429 重试和候选切换，同时跳过内存冷却。 |
