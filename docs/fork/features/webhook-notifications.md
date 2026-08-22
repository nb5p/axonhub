---
id: webhook-notifications
title: Webhook 事件通知、Bark 与发送历史
status: active
origin: upstream-derived
integration_method: adapted
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 37e54737c0e147831c6a0b529b79dda1961c3801
  adopted_commits:
    - f6184d80
    - 746b04dd
  last_checked_commit: 37e54737c0e147831c6a0b529b79dda1961c3801
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - c34299aebf320f35094760533353b2b73209801e
    - 8fee49dfa6c1e9eed8ffe1c1c9eb186ed2422ce8
  modules:
    - internal/server/biz/webhook_notifier.go
    - internal/server/biz/provider_quota.go
    - internal/ent/schema/webhook_delivery.go
    - internal/server/gql/system.graphql
    - frontend/src/features/system/components/webhook-settings.tsx
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: upstream-subset
  last_compared_at: 2026-08-22
reconciliations: []
history_rewrites: []
database:
  impact: additive
  backward_compatible: true
---

# Webhook 事件通知、Bark 与发送历史

## 目的

在上游已有的通用 Webhook 与 `channel.auto_disabled` 基础上，提供渠道错误、临时禁用、额度窗口耗尽和余额耗尽的可订阅通知；同时支持 Bark `/push` API，以及可审计、已脱敏的发送历史。不会主动向任何用户配置的地址发测试请求。

## 来源与采用范围

- 上游 `f6184d80` 提供自动禁用后的通用 Webhook 事件和模板配置。
- 上游 `746b04dd` 为 Webhook target 提供独立代理配置。
- 本地保留这些配置字段和事件名称，并在同一架构内扩展 target 类型、事件、投递审计和界面。

## 本地实现

- 新事件：`channel.error`、既有 `channel.auto_disabled`、`quota.window_exhausted`、`quota.balance_exhausted`。
- 额度事件只在首次检测为耗尽或由未耗尽转为耗尽时发送，避免轮询重复推送；脚本型配额的任意 progress window 也适用。
- Bark target 使用 `POST <server>/push` 和标准 JSON 字段 `device_key`、`title`、`body`、可选 `level`、`group`。
- 新表 `webhook_deliveries` 保存最新 500 条结果，包括 HTTP 状态和发送失败；入库前掩码 Authorization/Bearer、Bark device key、常见 token/secret/password/API key、URL 敏感参数及 JSON 敏感字段。
- 历史通过系统设置权限下的 GraphQL 查询展示，配置目标本身仍可编辑原始请求头与 Bark device key。

## 与来源的差异

- 上游只有自动禁用事件，不提供余额/窗口转换检测、Bark 或投递历史。
- 发送历史不保存响应正文，且不保存未掩码请求密钥，避免把审计功能变成凭据副本。
- 渠道错误仅在已有该事件订阅时异步投递，后台 goroutine 包含 panic 防护。

## 上游收敛

- 2026-08-22 对比 `upstream/unstable@37e54737c0e147831c6a0b529b79dda1961c3801`：`upstream-subset`。
- 上游已有单一自动禁用 Webhook 和 target proxy；没有多事件配额通知、Bark 或发送历史。保留上游结构并以增量方式扩展。

## 数据库兼容

- 兼容等级：`additive`。
- Ent 新增 `webhook_deliveries` 表，不更改既有表、列或 JSON 配置。旧 Webhook target 缺失 `type` 时读取并保存为 `webhook`。
- 绿色部署前需对现有 SQLite 使用 `.backup` 创建一致性备份，并对源与副本执行 `PRAGMA quick_check`；回退时恢复该副本。
- 新表仅保存脱敏审计数据，最多 500 条；回退旧版本会忽略该表，不影响原 Webhook 配置或渠道数据。

## 验证

- `go test ./internal/server/biz -run 'TestWebhookNotifier|TestNormalizeWebhookNotifierConfig|TestQuota(Window|Balance)' -count=1` 通过。
- `go test ./internal/server/gql -run '^$' -count=1` 通过。
- `pnpm exec tsc --noEmit -p tsconfig.json` 通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-22 | `f6184d80..37e54737` | `c34299aebf320f35094760533353b2b73209801e`、`8fee49dfa6c1e9eed8ffe1c1c9eb186ed2422ce8` | 在上游通用 Webhook 上增量实现事件、Bark、脱敏历史与界面。 |
