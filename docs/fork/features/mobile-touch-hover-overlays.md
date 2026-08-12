---
id: mobile-touch-hover-overlays
title: 移动端悬浮气泡触摸支持
status: active
origin: local-original
integration_method: original
source:
  repository: ssh://git@forgejo.109062.xyz:88/tux/axonhub.git
  branch: ai-slop
  baseline_commit: 71dd327c7da566a7d885b1ddc0bf79274789126c
  adopted_commits:
    - 1334dc0d744e67736f9c600354fae5ba138e1fa5
  last_checked_commit: 1334dc0d744e67736f9c600354fae5ba138e1fa5
  last_checked_at: 2026-08-12
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 1334dc0d744e67736f9c600354fae5ba138e1fa5
  modules:
    - frontend/src/components/ui/tooltip.tsx
    - frontend/src/components/ui/hover-card.tsx
    - frontend/src/components/ui/tap-tooltip-state.ts
    - frontend/src/components/truncated-text.tsx
    - frontend/src/components/tap-tooltip-state.test.mjs
    - frontend/src/components/touch-hover-support.test.mjs
    - frontend/src/features/channels
    - frontend/src/features/dashboard/components/api-key-activity-heatmap.tsx
    - frontend/src/features/models
    - frontend/src/features/threads
    - frontend/src/features/traces
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-12
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 移动端悬浮气泡触摸支持

## 目的

让系统中原本只能通过鼠标悬停查看的说明气泡在手机和平板上也可以用手指点按查看，同时保留桌面悬浮和键盘聚焦行为。

## 来源与采用范围

- 本地原创修复，统一扩展共享 Radix Tooltip 和 HoverCard，而不是逐页面重复实现触摸状态。
- 审计共享悬浮组件、原生 HTML `title`、第三方图表提示和纯 CSS hover；迁移确实承载隐藏信息但无法触摸显示的调用点。
- Recharts 自身已有 `touchstart` / `touchmove` 支持，因此保留其原生图表触摸行为；API Key 活动热力图原先使用的第三方 hover-only 提示则改用共享 Tooltip。

## 本地实现

- 共享 Tooltip 在触摸和手写笔点按后切换开关状态，状态更新延迟到当前 click 处理结束，避免 Radix 内部 click 收尾逻辑立即关闭刚打开的气泡。
- 共享 HoverCard 在触摸和手写笔按下时切换状态；点按外部区域、Escape 和组件原有 dismiss 行为继续生效。
- 鼠标继续使用 hover，键盘继续使用 focus；被 Tooltip 包裹的按钮仍只执行一次原有点击动作。
- `TruncatedText` 从浏览器原生 `title` 迁移到共享 Tooltip，只在文本实际被截断时提供可聚焦、可触摸的完整内容。
- 渠道、模型、线程、追踪和批量测试中的截断数据提示统一接入 `TruncatedText`；模型关系说明与模型映射编辑说明接入共享 Tooltip。
- API Key 活动热力图的日期单元格改用共享 Tooltip，并为键盘访问增加聚焦入口。

## 与来源的差异

无外部移植来源。修复保持现有 UI 文案、气泡内容与业务动作不变，只统一输入方式和可访问性。

## 上游收敛

- 2026-08-12 对比 `upstream/unstable@800bb72f4586428fa79905cb244d7840312d3d02`：共享 Tooltip 与 HoverCard 仍只依赖 Radix 默认 hover/focus 行为，没有等价的触摸切换支持，关系为 `none`。
- 后续上游若加入等价能力，应优先采用上游共享组件实现，并移除本地重复状态层。

## 数据库兼容

- 兼容等级：`none`。
- 纯前端交互变化，不修改 API、GraphQL、Ent Schema 或持久化数据。

## 性能与运行影响

- 仅为已存在的悬浮组件增加局部受控状态和事件处理，不新增网络请求、后台任务或全局监听器。
- 截断检测继续复用现有 ResizeObserver；没有改变查询频率或页面数据量。

## 验证

- `pnpm test:unit`：18/18 通过，覆盖触摸、手写笔、鼠标、延迟更新、共享 Tooltip/HoverCard、热力图和截断文字接入。
- `pnpm build`：生产构建通过，仅保留既有大 chunk 警告。
- 隔离 Chromium 组件测试：真实触摸点按可打开 Tooltip 与 HoverCard，点按外部可关闭；被包裹按钮只触发一次；桌面 hover 与键盘 focus 均可打开。
- 全仓 TSX AST 审计：剩余原生 `title` 仅用于上传 input、iframe 标题和操作按钮语义，不再承载需要触摸查看的数据说明。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-12 | 用户反馈与全仓交互审计 | `1334dc0d744e67736f9c600354fae5ba138e1fa5` | original：统一为共享悬浮层增加触摸/手写笔支持，并迁移 hover-only 热力图和原生数据提示。 |
