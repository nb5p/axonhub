---
id: channel-model-test-api-formats
title: 渠道模型测试接口格式选择
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca
  adopted_commits: []
  last_checked_commit: 9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca
  last_checked_at: 2026-08-21
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - a683e9122d639c5a1273cc28d509d7019f2a6c35
  modules:
    - frontend/src/features/channels
    - frontend/src/locales
    - internal/server/gql
    - internal/server/orchestrator
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-21
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 渠道模型测试接口格式选择

## 目的

让渠道模型测试在选择模型前先选择请求接口格式。弹窗标题区右侧提供 OpenAI Chat Completion、OpenAI Response 和 Anthropic Messages 三种格式，未选择格式时不显示可测模型。

## 来源与采用范围

- 本地原创功能，不移植外部代码。
- 2026-08-21 对比 `upstream/unstable@9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca`，未发现同类测试接口格式选择或格式强制路由实现。

## 本地实现

- 前端依据渠道解析后的端点启用可选格式；切换格式会清空已选模型、筛选文本和测试结果，模型表仅在所选格式可用时展示渠道已配置的可测试模型。
- GraphQL `TestChannelInput` 增加可选的 `apiFormat` 枚举字段，并保持未传字段的旧调用兼容。
- 测试编排器为三种格式构建各自的原生入站请求：Chat Completions、Responses 和 Anthropic Messages；成功判定同时兼容三种原生响应载荷。
- 渠道选择器只使用与所选格式精确匹配的已解析端点，缺少该端点时立即报错，不回退到其他接口格式。

## 与来源的差异

上游基线没有同类实现。本功能复用 AxonHub 既有端点解析、Transformer 和测试记录链路，不新增一套测试服务。

## 上游收敛

当前关系为 `none`。上游后续提供等价的格式选择与精确端点路由时，应采用其接口和命名，并通过追加的 `🧩` reconciliation commit 移除重复实现。

## 数据库兼容

兼容等级为 `none`；仅扩展 GraphQL 测试输入和运行时请求构造，不修改 Ent Schema、持久化数据或渠道配置。

## 验证

- `go test ./internal/server/orchestrator -run 'TestBuild(TestRequestUsesConfiguredPrompts|ChannelTestHTTPRequestUsesSelectedAPIFormat)$' -count=1`：通过。
- `go test ./internal/server/gql -run '^$' -count=1`：通过。
- `frontend/pnpm exec tsc --noEmit`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-21 | `upstream/unstable@9fb6f1af` | `a683e9122d639c5a1273cc28d509d7019f2a6c35` | original：新增测试接口格式选择、原生请求转换与精确端点强制路由。 |
