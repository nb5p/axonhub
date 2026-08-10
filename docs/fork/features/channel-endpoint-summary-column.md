---
id: channel-endpoint-summary-column
title: 渠道支持端点摘要列
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  adopted_commits: []
  last_checked_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  last_checked_at: 2026-08-11
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - ba6c2fca6bf63338d31fb78233e9d7d5c4cf953b
    - 68ba2b662c661c84ecbe588a990d1ea2ae9ab144
  modules:
    - frontend/src/features/channels/components/channel-endpoints-cell.tsx
    - frontend/src/features/channels/components/channels-columns.tsx
    - frontend/src/locales
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-11
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 渠道支持端点摘要列

## 目的

在渠道列表直接展示运行时有效端点，让用户无需打开“端点配置”即可判断渠道是否支持 Chat Completions、Responses、Anthropic Messages 或 Gemini Contents。

## 来源与采用范围

本地原创前端功能，复用渠道列表已经返回的默认端点和自定义端点，不增加 GraphQL 字段或额外请求。

## 本地实现

- 新增默认可见、允许手动隐藏的“支持端点”列。
- 主端点使用 `C`、`R`、`M`、`G` 的 32×24 圆角矩形徽标，并按 OpenAI、Anthropic、Gemini 使用不同颜色。
- 悬停主徽标展示完整名称；其他端点折叠到 `…`，气泡显示剩余数量与完整列表。
- 自定义端点与默认端点按 API 格式去重，覆盖项不会重复展示。
- 仅增加行内常量计算，不新增后台任务、查询或网络请求。

## 与来源的差异

上游基线只有操作菜单中的端点配置对话框，没有列表摘要。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价列表列，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；仅修改前端展示和翻译。

## 验证

- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `ba6c2fca6bf63338d31fb78233e9d7d5c4cf953b` | 本地原创端点摘要列。 |
| 2026-08-11 | 本地反馈 | `68ba2b66` | 将 24×24 徽标调整为至少 32×24，恢复明显的圆角矩形外观。 |
