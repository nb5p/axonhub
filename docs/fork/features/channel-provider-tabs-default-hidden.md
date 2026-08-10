---
id: channel-provider-tabs-default-hidden
title: 渠道供应商标签默认隐藏
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  adopted_commits: []
  last_checked_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  last_checked_at: 2026-08-10
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 7f84e474b419160eefbe32ebad81a41e9b99dbf6
  modules:
    - frontend/src/features/channels/context/channels-context.tsx
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-10
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 渠道供应商标签默认隐藏

## 目的

渠道页面默认隐藏供应商分类标签，用户仍可通过工具栏按钮展开，并保留后续选择。

## 来源与采用范围

本地原创的前端偏好调整，不采用外部代码。

## 本地实现

将供应商标签默认值改为隐藏，并升级本地存储键，使已访问过旧版页面的浏览器应用一次新默认值；用户之后的手动选择继续持久化。

## 与来源的差异

上游基线默认展开供应商标签。

## 上游收敛

2026-08-10 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，上游默认值仍为展开，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；仅使用浏览器 localStorage。

## 验证

- TypeScript 检查通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `7f84e474b419160eefbe32ebad81a41e9b99dbf6` | 改为默认隐藏，并迁移本地偏好键。 |
