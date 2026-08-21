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
    - 本次提交（2026-08-21，多格式测试）
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

让渠道模型测试可以同时验证多个请求接口格式。格式下拉框支持多选、至少保留一项，按常用顺序列出 OpenAI Chat Completions、OpenAI Responses、Anthropic Messages，再列 Gemini Contents；默认选中前三项。

## 来源与采用范围

- 本地原创功能，不移植外部代码。
- 2026-08-21 对比 `upstream/unstable@9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca`，未发现同类测试接口格式选择或格式强制路由实现。

## 本地实现

- 单渠道模型表按所选格式动态增加结果列。每个单元格独立显示状态、延迟、错误和测试按钮；底部“批量测试”会以最多四个并发请求，测试所有选中模型与所有选中格式的组合。
- 批量健康检查同样按所选格式增加结果列与独立测试按钮；批量执行测试“选中渠道 × 选中格式”的所有可用组合，最大并发为四。缺少模型或端点的组合明确显示为跳过，不会发出错误格式的请求。
- 恢复禁用渠道只在该渠道至少有一个成功结果、且没有任何可用格式失败时出现，避免部分格式仍失败的渠道被提前重新启用。
- GraphQL `TestChannelInput` 保持可选 `apiFormat` 的旧调用兼容，并增加 `GEMINI_CONTENTS` 枚举值。
- 测试编排器为四种格式构建各自的原生入站请求：Chat Completions、Responses、Anthropic Messages 和 Gemini Contents。Gemini 请求带模型动作路径，分别适配 `generateContent` / `streamGenerateContent`。
- 渠道选择器只使用与所选格式精确匹配的已解析端点，缺少该端点时立即报错，不回退到其他接口格式。

## 与来源的差异

上游基线没有同类实现。本功能复用 AxonHub 既有端点解析、Transformer 和测试记录链路，不新增一套测试服务。

## 上游收敛

当前关系为 `none`。上游后续提供等价的格式选择与精确端点路由时，应采用其接口和命名，并通过追加的 `🧩` reconciliation commit 移除重复实现。

## 数据库兼容

兼容等级为 `none`；仅扩展 GraphQL 测试输入和运行时请求构造，不修改 Ent Schema、持久化数据或渠道配置。

## 验证

- `go test ./internal/server/orchestrator -run 'TestBuild(TestRequestUsesConfiguredPrompts|ChannelTestHTTPRequestUsesSelectedAPIFormat)$' -count=1`：通过，包含 Gemini Contents 请求转换。
- `go test ./internal/server/gql -run '^$' -count=1`：通过。
- `frontend/pnpm exec tsc --noEmit`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-21 | `upstream/unstable@9fb6f1af` | `a683e9122d639c5a1273cc28d509d7019f2a6c35` | original：新增测试接口格式选择、原生请求转换与精确端点强制路由。 |
| 2026-08-21 | 本地需求 | 本次提交（多格式测试） | original：多选格式结果列、渠道×格式批量健康检查、Gemini Contents 原生请求与安全恢复判断。 |
