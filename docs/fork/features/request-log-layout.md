---
id: request-log-layout
title: 请求日志序号与自适应列宽
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
    - d476d57b5f4ff6b29d890e844baee2566ab568ab
    - 81c62c6e190ee7c5868173577301cc131bd9c1d9
  modules:
    - frontend/src/features/requests/components/requests-columns.tsx
    - frontend/src/features/requests/components/requests-table.tsx
    - frontend/src/features/requests/data/requests.ts
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

# 请求日志序号与自适应列宽

## 目的

恢复请求日志中的请求序号，将序号、状态和时间组织在同一列，并让模型、渠道列按内容自适应宽度且保留可读的最小宽度。

## 来源与采用范围

本地原创实现，基于上游请求日志表格与 GraphQL 查询，不采用外部代码。

## 本地实现

- 请求查询读取 ID，序号使用与完成状态一致的深绿色。
- 模型和渠道列使用内容固有宽度，分别保留最小宽度。
- 表格使用内容宽度布局，避免剩余空间被平均分配到模型列。

## 与来源的差异

上游基线不展示请求 ID，且表格会把剩余空间分配给模型等列；本地保留紧凑的运维日志布局。

## 上游收敛

2026-08-10 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；只读取已有请求 ID 并调整前端布局。

## 验证

- TypeScript 检查通过。
- Vite 前端构建通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `d476d57b5f4ff6b29d890e844baee2566ab568ab`, `81c62c6e190ee7c5868173577301cc131bd9c1d9` | 恢复请求序号并修正颜色和自适应列宽。 |
