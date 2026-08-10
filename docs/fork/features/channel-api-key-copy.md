---
id: channel-api-key-copy
title: 渠道 API Key 列表复制按钮
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 2b78817e0620c334dd351a4fef296b49dcb7e449
  adopted_commits: []
  last_checked_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  last_checked_at: 2026-08-10
  license: Apache-2.0
local:
  branch: fix/channel-api-key-copy
  commit_marker: "none"
  commits:
    - 236ff5c284c2ceec62d01e41da669d2126c8475f
    - bb2a5a469f6ffb3882f03ebf6901076e8615c916
  modules:
    - frontend/src/features/channels/components/channels-action-dialog.tsx
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

# 渠道 API Key 列表复制按钮

## 目的

在编辑渠道的 API Key 列表中，为每个密钥增加复制按钮。用户点击后复制完整密钥，界面仍只展示掩码值；复制失败时显示错误提示。

## 来源与采用范围

- 基于 AxonHub `upstream/unstable` 的现有渠道编辑对话框实现。
- 未移植外部项目代码；复用现有 `Copy` 图标、Toast 和 Tooltip 组件。

## 本地实现

- 稳定入口：`frontend/src/features/channels/components/channels-action-dialog.tsx`
- 复制逻辑使用 `navigator.clipboard.writeText`，成功和失败分别显示本地化提示。
- 按钮位于每行的禁用/启用按钮与删除按钮之间。
- 不涉及后台任务、数据库、GraphQL 或 API 变更。

## 与来源的差异

- 在官方 API Key 列表行操作区增加复制按钮；复制内容使用未掩码的表单值，展示仍保持掩码。

## 上游收敛

- 最近比较：`upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`。
- 当前上游尚无该列表行复制按钮实现，关系为 `none`。

## 数据库兼容

- 兼容等级：`none`。
- 不修改数据结构和已有数据。

## 验证

- `git diff --check`：通过。
- 功能分支从官方 `upstream/unstable` 创建，并以真实 merge 合入 `ai-slop`。
- 镜像 `axonhub:green-9dfd6ac0-cdc76ab1` 构建成功，并已部署到本机绿色环境 `192.168.111.21:9090`；容器健康检查通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@2b78817e` | `236ff5c2`, `bb2a5a46` | 新增列表行复制按钮，并合入 `ai-slop`。 |
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `cdc76ab1` | 同步官方 Fenno 与 Claude Code 后续变更；复制按钮功能无冲突，继续保留。 |
