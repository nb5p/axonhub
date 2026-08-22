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
  last_checked_commit: 49ade6f279eae7aed46858dc121258e922ec9870
  last_checked_at: 2026-08-22
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - a683e9122d639c5a1273cc28d509d7019f2a6c35
    - 35a08783a1fc5c07d475afdcd506b8aa0e384d4b
    - 0d6d21274b973e1fd06012b045f3a4b1a2579597
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
  last_compared_at: 2026-08-22
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
- 2026-08-22 对比 `upstream/unstable@49ade6f279eae7aed46858dc121258e922ec9870`，未发现同类测试接口格式选择或格式强制路由实现。

## 本地实现

- 单渠道模型表按所选格式动态增加结果列。每个单元格独立显示状态、延迟、错误和测试按钮；底部“批量测试”会以最多四个并发请求，测试所有选中模型与所有选中格式的组合。
- 批量健康检查同样按所选格式增加结果列与独立测试按钮；批量执行测试“选中渠道 × 选中格式”的所有可用组合，最大并发为四。缺少模型或端点的组合明确显示为跳过，不会发出错误格式的请求。
- 恢复禁用渠道只在该渠道至少有一个成功结果、且没有任何可用格式失败时出现，避免部分格式仍失败的渠道被提前重新启用。
- GraphQL `TestChannelInput` 保持可选 `apiFormat` 的旧调用兼容，并增加 `GEMINI_CONTENTS` 枚举值。
- 测试编排器为四种格式构建各自的原生入站请求：Chat Completions、Responses、Anthropic Messages 和 Gemini Contents。Gemini 请求带模型动作路径，分别适配 `generateContent` / `streamGenerateContent`。
- 渠道选择器只使用与所选格式精确匹配的已解析端点，缺少该端点时立即报错，不回退到其他接口格式。
- 单渠道测试的格式选择器和结果列只显示该渠道实际支持的四类可测端点；默认选择会按常用顺序从该集合取值。操作列的快捷测试同样将首个可测格式明确传给 GraphQL，避免遗漏格式后回退为 Chat Completions。批量测试保留跨渠道的完整格式列，并将单个渠道不支持的组合标为跳过。
- 批量测试表格不会直接展开上游错误：优先显示方括号中的上游错误码，其次显示 HTTP 状态码，没有结构化码时显示 `ERR`；完整错误仅在悬浮、触摸或键盘焦点的 Tooltip 中展示。

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
- `node --test frontend/src/features/channels/data/channel-test-api-formats.test.mjs`：2 项通过，覆盖只提供 Responses 时的下拉项、默认值和展示列集合。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-21 | `upstream/unstable@9fb6f1af` | `a683e9122d639c5a1273cc28d509d7019f2a6c35` | original：新增测试接口格式选择、原生请求转换与精确端点强制路由。 |
| 2026-08-21 | 本地需求 | `35a08783a1fc5c07d475afdcd506b8aa0e384d4b` | original：多选格式结果列、渠道×格式批量健康检查、Gemini Contents 原生请求与安全恢复判断。 |
| 2026-08-22 | 本地回归修正 | `0d6d21274b973e1fd06012b045f3a4b1a2579597` | original：单渠道只显示支持的格式和列，操作列快捷测试显式路由至实际端点，防止 Responses 渠道误测 Chat Completions。 |
| 2026-08-22 | 本地展示修正 | `4d1cfc1282c5` | original：批量测试用紧凑错误码替换长错误原文，并通过 Tooltip 保留完整诊断。 |
