---
id: filter-state-persistence
title: 页面筛选状态持久化
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
    - 36b9e73749ef0ded1c074bda2ed80527f793e6c2
  modules:
    - frontend/src/hooks/use-persisted-filter.ts
    - frontend/src/lib/filter-storage.ts
    - frontend/src/features
    - frontend/src/stores/analyticsStore.ts
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

# 页面筛选状态持久化

## 目的

页面级筛选条件保存到浏览器本地存储。用户切换应用页面、刷新页面或在另一个浏览器标签页中修改筛选时，筛选状态不会无故重置。

## 来源与采用范围

本地原创实现，覆盖管理列表、请求日志、分析页、仪表盘和统计卡片。表单、对话框和弹出面板中的一次性搜索不在持久化范围内。

## 本地实现

- 通用 `usePersistedFilter` hook 使用带版本和页面作用域的键保存状态，并监听 `storage` 事件同步浏览器标签页。
- 自定义序列化保留日期范围中的 `Date` 类型；损坏或不可用的本地存储安全回落到页面默认值。
- 请求日志继续支持 URL 筛选参数；显式 URL 条件同步到本地，移除 URL 条件后恢复最近保存状态。
- Zustand 分析筛选仅持久化筛选数据，不持久化操作函数。

## 与来源的差异

上游基线中的大部分页面筛选仅保存在 React 内存或 URL 中，切换页面或刷新后会回到默认值。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现统一的页面筛选本地持久化实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；仅写入浏览器 `localStorage`，不修改服务端 Schema 或数据。

## 验证

- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- `node --test frontend/src/lib/filter-storage.test.mjs`：3 项通过，覆盖键隔离、日期恢复和空值删除。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `36b9e73749ef0ded1c074bda2ed80527f793e6c2` | 为页面级筛选增加本地持久化、刷新恢复和跨标签页同步。 |
