---
id: request-detail-responsive-interactions
title: 请求详情响应式交互优化
status: active
origin: local-original
integration_method: original
source:
  repository: ssh://git@forgejo.109062.xyz:88/tux/axonhub.git
  branch: ai-slop
  baseline_commit: a15182f377c6ca2fc38beebe21500109f3220388
  adopted_commits:
    - 5963e77efe0156a9bc472be48e51950ddd731e7d
  last_checked_commit: 5963e77efe0156a9bc472be48e51950ddd731e7d
  last_checked_at: 2026-08-14
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 5963e77efe0156a9bc472be48e51950ddd731e7d
  modules:
    - frontend/src/features/requests/components/request-detail-content.tsx
    - frontend/src/features/requests/components/request-conversation-viewer.tsx
    - frontend/src/features/requests/components/request-detail-page.tsx
    - frontend/src/features/requests/components/request-detail-global-page.tsx
    - frontend/src/features/requests/components/use-mobile-auto-hide-header.ts
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

# 请求详情响应式交互优化

## 目的

消除请求详情 JSON 展开后的双纵向滚动，压缩对话预览筛选栏在桌面和移动端的高度，并让移动端顶部信息栏随阅读方向自动隐藏或恢复。

## 本地实现

- 请求头、请求体、响应体和执行记录 JSON 的根节点展开后随内容完整增高；执行记录统一使用全层展开。
- 对话消息的 Raw JSON 不再设置内部最大高度和纵向滚动。
- 右下角按钮在“收起全部”和“回到顶部”两种状态下使用不同图标。
- 对话预览工具栏在桌面固定为最多两行，在 `md` 以下合并为单行横向滚动；模型 ID 单行截断。
- 项目级和全局请求详情页共用移动端滚动方向 Hook：向上滑动阅读时隐藏顶部栏，向下滑动或返回顶部时恢复。

## 上游收敛

2026-08-14 对比 `upstream/unstable@852b8c6f`，未发现等价的 JSON 自适应高度、移动端单行筛选栏或滚动方向顶部栏实现，关系为 `none`。

## 数据库兼容

- 兼容等级：`none`。
- 纯前端布局与交互变化，不修改 GraphQL、Schema 或持久化数据。

## 验证

- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- 绿色环境部署后使用已登录浏览器分别验证桌面和 390px 移动视口。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-14 | 用户反馈与本地实现 | `5963e77efe0156a9bc472be48e51950ddd731e7d` | original：统一请求详情 JSON 展开、工具栏与移动端顶部栏交互。 |
