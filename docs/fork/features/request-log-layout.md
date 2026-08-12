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
    - 1afef2bec18329f7ccf987cd21cf1664896f1423
    - 8d9a437cc9aa46ecd419b725762cd4bc76bc3428
    - 3a139b8958b6d8cc4a1fdeac0749ae6a889fcf2c
  modules:
    - frontend/src/features/requests/components/requests-columns.tsx
    - frontend/src/features/requests/components/requests-table.tsx
    - frontend/src/features/requests/components/request-detail-content.tsx
    - frontend/src/features/requests/components/request-conversation-viewer.tsx
    - frontend/src/components/json-tree-view.tsx
    - frontend/src/features/requests/data/requests.ts
    - frontend/src/locales
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

恢复请求日志中的请求序号，将序号、状态和时间组织在同一列，并让表格在保留各列内容最小宽度的同时填满宽屏容器。

## 来源与采用范围

本地原创实现，基于上游请求日志表格与 GraphQL 查询，不采用外部代码。

## 本地实现

- 请求查询读取 ID，序号使用与完成状态一致的深绿色。
- 模型和渠道列使用内容固有宽度，分别保留最小宽度。
- 表格以内容宽度布局，并以容器宽度作为下限；宽屏填满容器，窄屏保持内容宽度并横向滚动。
- “调用方”列改名为“密钥”，并移动到客户端 IP 与渠道之间。
- 在用量右侧恢复独立的缓存命中率列；输入不少于 40,000 Token 且命中率低于 80% 时使用红色提醒。
- 请求详情的请求头 JSON 默认折叠并显示项目数量；JSON 框仍随根节点状态在最小内容高度和完整高度间切换。
- 移除 JSON 框中央的折叠悬浮按钮，统一复用页面右下角上箭头：有展开的 JSON 时优先“收起所有展开的内容”，全部收起后恢复“回到页首”。

## 与来源的差异

上游基线不展示请求 ID，且表格会把剩余空间分配给模型等列；本地保留紧凑的运维日志布局。

## 上游收敛

2026-08-10 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；只读取已有请求 ID 并调整前端布局。

## 验证

- TypeScript 检查通过。
- Vite 前端构建通过。
- 浏览器布局回归验证：1500、1550、1580、1590px 视口下表格右侧空白均为 0；1280px 视口保留横向滚动。
- JSON 折叠交互修正随本批需求再次执行 `pnpm build`：通过。
- `node --test src/features/requests/request-detail-collapse-action.test.mjs`：3 项通过。
- `pnpm exec tsc --noEmit`：通过。
- Chrome DevTools 以 390×844 视口验证：请求头默认显示 `{ 16 items }`；展开后统一按钮切换为“收起所有展开的内容”，点击后恢复 59px 折叠高度并切回“回到页首”；页面滚动归零后按钮隐藏。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `d476d57b5f4ff6b29d890e844baee2566ab568ab`, `81c62c6e190ee7c5868173577301cc131bd9c1d9` | 恢复请求序号并修正颜色和自适应列宽。 |
| 2026-08-11 | 本地反馈与上游历史缓存命中率逻辑 | `1afef2bec18329f7ccf987cd21cf1664896f1423` | 修复宽屏右侧空白，调整密钥列，并恢复缓存命中率。 |
| 2026-08-11 | 用户反馈 | `8d9a437cc9aa46ecd419b725762cd4bc76bc3428` | 让请求详情 JSON 框跟随根节点折叠，并增加滚动时的居中折叠入口。 |
| 2026-08-12 | 用户反馈 | `3a139b8958b6d8cc4a1fdeac0749ae6a889fcf2c` | 请求头改为默认折叠，删除居中折叠入口，并让右下角上箭头优先收起当前展开的 JSON。 |
