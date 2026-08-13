---
id: table-column-sizing
title: 表格动态列宽与手动调整
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: fc1d27dad4119d06802e0024ce86f3e91cdf740e
  adopted_commits: []
  last_checked_commit: fc1d27dad4119d06802e0024ce86f3e91cdf740e
  last_checked_at: 2026-08-14
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - b7489b68076b830fc5f0f591adf73d4207349796
  modules:
    - frontend/src/components/data-table-column-sizing.tsx
    - frontend/src/hooks/use-persisted-column-sizing.ts
    - frontend/src/features/channels/components
    - frontend/src/features/apikeys/components
    - frontend/src/features/requests/components
    - frontend/src/features/traces/components
    - frontend/src/features/users/components
    - frontend/src/features/proejct-users/components
    - frontend/src/locales
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-14
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 表格动态列宽与手动调整

## 目的

让管理表格在容器变宽时自动用满可用空间，并允许用户在不改变表格总宽度的前提下手动调整相邻列的宽度。渠道标签列不再固定只显示两个标签，而是根据当前列宽显示尽可能多的标签。

## 来源与采用范围

本地原创实现。首批接入所有已经提供“列设置”的管理表格：渠道、API 密钥、请求日志、追踪、系统用户和项目用户。

## 本地实现

- 默认使用浏览器表格自动布局，让内容较少的列自然收窄、其余列吸收剩余空间。
- 表头分隔线支持鼠标、触摸和键盘调整；一次调整同时改变当前列和右侧相邻列，保持表格总宽度不变。
- 每张表的手动列宽单独写入浏览器 `localStorage`，刷新页面后继续生效。
- 所有已接入表格的“列设置”菜单增加“重置列宽”，恢复自动布局并清除本地列宽记录。
- 渠道标签单元格使用 `ResizeObserver` 计算当前可容纳的标签数量；剩余标签折叠为 `+N`，悬停或触摸后仍可查看。
- 支持端点列继续使用原有固定摘要和省略策略，不随空间增加而展开全部端点。
- 每个可见标签单元格只增加一个浏览器原生尺寸观察器；不新增后台任务、接口请求或服务端查询。

## 与来源的差异

上游基线没有列宽状态、拖拽手柄或重置列宽入口。该功能集中在共享组件与 Hook 中，各业务表只接入状态和渲染，避免复制列宽算法。

## 上游收敛

2026-08-14 比较 `upstream/unstable@fc1d27dad4119d06802e0024ce86f3e91cdf740e`，未发现 `columnSizing`、列拖拽或重置列宽的同类实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。只新增浏览器本地偏好，不修改服务端 Schema、配置或现有数据。

## 验证

- `frontend/node_modules/.bin/tsc --noEmit --project frontend/tsconfig.json`：通过。
- Vite 开发服务器对共享模块和六张接入表格的源码转换请求均返回 HTTP 200。
- `git diff --check`：通过。
- 本地自动化浏览器没有已登录会话，因此未伪造业务数据进行截图验收；交互视觉留待绿色环境验证。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-14 | 用户需求 / `upstream/unstable@fc1d27da` | `b7489b68076b830fc5f0f591adf73d4207349796` | 本地原创实现动态列宽、相邻列联动调整、持久化、重置入口及标签自适应展示。 |
