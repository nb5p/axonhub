---
id: table-column-visibility
title: 管理表格列显隐设置
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 800bb72f4586428fa79905cb244d7840312d3d02
  adopted_commits: []
  last_checked_commit: 800bb72f4586428fa79905cb244d7840312d3d02
  last_checked_at: 2026-08-11
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 7d0084e60b90b8a11ca0406f1473e2c86648d74a
    - f8066acf7a427c95d62833b6eee233585f1fe94c
    - 8eb858cc3e32cce1086058ca0ab0457c1f7518be
    - a08c30e7d3c74c9586ba9d407f65a9e9bc56dc6c
    - 0aa6afd273af80cd217f98081492357d5553008b
    - 26d7f68b88b9f443aebc50bdb521190ad5afb387
  modules:
    - frontend/src/features/apikeys/components
    - frontend/src/features/channels/components
    - frontend/src/features/requests/components/data-table-view-options.tsx
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

# 管理表格列显隐设置

## 目的

让个人使用场景可以隐藏 API 密钥页不关心的列，并补齐渠道创建时间列的显隐能力；同时保证请求日志和渠道列设置使用正确的中文列名。

## 来源与采用范围

本地原创实现，复用现有 TanStack Table 列可见状态和各页面的列设置菜单。

## 本地实现

- API 密钥筛选区右侧增加“列设置”，支持显隐 ID、API Key、创建者、类型、状态、生效配置、创建时间和更新时间等可选数据列。
- API 密钥列可见状态写入浏览器 `localStorage`，刷新或切换页面后保持选择。
- 渠道“创建时间”改为可隐藏，继续复用渠道列表已有的列可见状态持久化。
- 渠道“服务商”和批量选择框均可隐藏；新增默认显示的渠道 ID 列，并允许在同一菜单中隐藏。
- API 密钥批量选择框可隐藏，其状态与其他 API 密钥列使用同一份浏览器本地偏好。
- 请求日志列设置使用专用的“缓存命中率”标签，不再错误渲染需要 `rate` 参数的单元格文案。
- 渠道端点列在列设置中显式映射为“支持端点”，创建时间使用公共翻译键。
- 渠道标签列默认显示并可在列设置中显隐。旧版自动保存的“隐藏标签”偏好会迁移为可见，因为旧界面并未提供手动恢复入口；迁移后用户的选择正常持久化。

## 与来源的差异

上游 API 密钥页虽然已有列设置组件和列状态，但未把入口挂到工具栏，也未持久化状态；本地补齐完整交互，并扩展渠道和请求日志的列标签映射。

## 上游收敛

2026-08-11 比较 `upstream/unstable@800bb72f4586428fa79905cb244d7840312d3d02`，未发现覆盖这些页面行为的等价实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。只修改前端列定义、翻译映射和浏览器本地偏好，不修改服务端 Schema 或数据。

## 验证

- `git diff --check`：通过。
- `pnpm build`（2026-08-12）：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | 用户需求 | `7d0084e60b90b8a11ca0406f1473e2c86648d74a` | API 密钥页增加可持久化的列设置入口。 |
| 2026-08-11 | 用户需求 | `f8066acf7a427c95d62833b6eee233585f1fe94c` | 渠道创建时间列允许隐藏。 |
| 2026-08-11 | 用户反馈 | `8eb858cc3e32cce1086058ca0ab0457c1f7518be` | 修复缓存命中率和支持端点在列设置中的翻译映射。 |
| 2026-08-12 | 用户需求 | `a08c30e7d3c74c9586ba9d407f65a9e9bc56dc6c` | 渠道服务商列和批量选择框允许隐藏。 |
| 2026-08-12 | 用户需求 | `0aa6afd273af80cd217f98081492357d5553008b` | 新增可配置显隐的渠道 ID 列。 |
| 2026-08-12 | 用户需求 | `26d7f68b88b9f443aebc50bdb521190ad5afb387` | API 密钥批量选择框允许隐藏。 |
| 2026-08-22 | 用户反馈 | `6de0fbbe6edf` | 恢复默认显示的渠道标签列，并迁移旧的不可配置隐藏偏好。 |
