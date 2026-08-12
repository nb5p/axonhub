---
id: mobile-request-route-tooltip
title: 移动端请求路径触摸提示
status: upstream-pending
origin: local-original
integration_method: original
source:
  repository: ssh://git@forgejo.109062.xyz:88/tux/axonhub.git
  branch: style/improve-mobile-details
  baseline_commit: 7ed4400595c2d73ca14e0131a86657058f261852
  adopted_commits:
    - 7cd527ab29f9640b888c45e7818fbfc8f8fc1c16
  last_checked_commit: 7cd527ab29f9640b888c45e7818fbfc8f8fc1c16
  last_checked_at: 2026-08-05
  license: project-local
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - e8e19993720bff05d5a10cc2e153017c6e69220a
    - 1334dc0d744e67736f9c600354fae5ba138e1fa5
  modules:
    - frontend/src/components/ui/tap-tooltip.tsx
    - frontend/src/components/ui/tap-tooltip-state.ts
    - frontend/src/components/tap-tooltip-state.test.mjs
    - frontend/src/features/requests/components/requests-columns.tsx
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-06
reconciliations: []
history_rewrites:
  - rewritten_at: 2026-08-06
    base_commit: d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af
    old_commits:
      - 7cd527ab29f9640b888c45e7818fbfc8f8fc1c16
      - cb00b35c500411c702fd83ebb38e3c894d0599ec
    new_commits:
      - e8e19993720bff05d5a10cc2e153017c6e69220a
    upstream_commits: []
    disposition: rebuilt
    reason: 在上游重构后的请求表结构上重建触摸提示能力，省略旧 merge 外壳并保留原作者和作者日期。
database:
  impact: none
  backward_compatible: true
---

# 移动端请求路径触摸提示

## 目的

请求日志中的模型 ID 重写图标和渠道名称原本依赖鼠标 hover。该功能允许触摸屏用户点击相关元素查看完整重写路径，同时保持桌面鼠标和键盘交互。

## 来源与采用范围

- 本地原创功能，独立贡献分支为 `style/improve-mobile-details`。
- 分支基于上游 `7ed4400595c2d73ca14e0131a86657058f261852`，功能提交为 `7cd527ab29f9640b888c45e7818fbfc8f8fc1c16`。
- 旧历史通过 merge commit `cb00b35c500411c702fd83ebb38e3c894d0599ec` 进入 `ai-slop`；2026-08-06 重建后由 `e8e19993720bff05d5a10cc2e153017c6e69220a` 直接承载有效差异。

## 本地实现

- 新增可复用 Tap Tooltip，在触摸和手写笔按下时切换显示状态。
- 鼠标继续使用原有 hover 行为，键盘继续使用 focus 行为；调用方提供 `onActivate` 时，鼠标点击和键盘激活仍执行上游原动作，触摸和手写笔则只切换 tooltip。
- 请求表模型 ID 和渠道名称统一使用该交互。
- 独立状态函数测试触摸、手写笔和鼠标输入的判定。
- 共享状态辅助函数会把触摸点按后的开关更新排到当前 click 处理之后，避免 Radix 收尾逻辑覆盖打开状态；全局悬浮层触摸支持复用该逻辑。

## 与来源的差异

重建时遵循上游 `d6ed9c62` 的新版请求表结构：保留上游模型、推理强度、透传状态和执行链展示，只将发生路由变化时的模型与渠道提示入口接入 `TapTooltip`。该分支仍用于准备上游贡献；创建 GitHub PR 后必须补充 PR URL，合入后将状态改为 `upstreamed` 并记录上游接受 commit。

## 上游收敛

- 2026-08-06 检查 `upstream/unstable@d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af`：上游已有新版请求表和执行链展示，但 tooltip 仍依赖 hover/focus，没有同等的触摸切换状态，关系为 `none`。
- 后续若上游接受同类触摸实现，按总账规则比较行为并创建 reconciliation 记录，不再重复维护本地组件。

## 数据库兼容

- 兼容等级：`none`
- 纯前端交互变化，不修改 API、Schema 或数据。

## 验证

- `node --test frontend/src/components/tap-tooltip-state.test.mjs`：2/2 通过。
- 2026-08-06 重建后重新运行同一目标测试：2/2 通过。
- 后续上游 PR 验证时应在依赖完整环境补充组件级交互测试。
- 2026-08-12 全局触摸修复后执行 `pnpm test:unit`：18/18 通过；隔离 Chromium 验证触摸打开、外部关闭、桌面 hover、键盘 focus 和按钮单次激活均通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-05 | `7ed44005..7cd527ab` | `7cd527ab29f9640b888c45e7818fbfc8f8fc1c16` | 本地原创实现并完成目标测试。 |
| 2026-08-05 | 集成到 `ai-slop` | `cb00b35c500411c702fd83ebb38e3c894d0599ec` | 保留独立贡献分支，同时进入私有主分支。 |
| 2026-08-06 | 历史重建至上游 `d6ed9c62` | `e8e19993720bff05d5a10cc2e153017c6e69220a` | 按上游新版请求表重建有效差异，保留桌面激活动作，并补上 `🧩` 标记。 |
| 2026-08-12 | 全局悬浮层触摸修复 | `1334dc0d744e67736f9c600354fae5ba138e1fa5` | refined：复用延迟触摸状态辅助函数，使共享 Radix Tooltip 的点按打开不会被 click 收尾覆盖。 |
