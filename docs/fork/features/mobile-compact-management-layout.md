---
id: mobile-compact-management-layout
title: 移动端管理页紧凑布局
status: active
origin: local-original
integration_method: original
source:
  repository: ssh://git@forgejo.109062.xyz:88/tux/axonhub.git
  branch: ai-slop
  baseline_commit: 614eaad733684779b97b553df8ab4e75d15c727e
  adopted_commits:
    - 5f4935b7023f552f8628d497b53aca9f6b5ff47e
    - 621c1de32ffdf9c08a1a85279976b897dd9fbd06
    - b2ba30c77ca611cd251b4169ca352e7d673aaca3
    - f6dc4e81c531ee9c19ed5b7ed7405ab29b66ad74
  last_checked_commit: f6dc4e81c531ee9c19ed5b7ed7405ab29b66ad74
  last_checked_at: 2026-08-12
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 5f4935b7023f552f8628d497b53aca9f6b5ff47e
    - 621c1de32ffdf9c08a1a85279976b897dd9fbd06
    - b2ba30c77ca611cd251b4169ca352e7d673aaca3
    - f6dc4e81c531ee9c19ed5b7ed7405ab29b66ad74
  modules:
    - frontend/src/components/layout
    - frontend/src/features/apikeys
    - frontend/src/features/channels
    - frontend/src/features/data-storages
    - frontend/src/features/models
    - frontend/src/features/playground
    - frontend/src/features/projects
    - frontend/src/features/prompts
    - frontend/src/features/roles
    - frontend/src/features/system
    - frontend/src/features/threads
    - frontend/src/features/traces
    - frontend/src/features/usage-statistics
    - frontend/src/features/users
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

# 移动端管理页紧凑布局

## 目的

减少管理页面在窄屏上的非必要纵向占用，让用户在首屏看到更多数据，同时保留适合触摸和鼠标滚轮的横向筛选交互。

## 来源与采用范围

- 本地原创响应式调整，不移植外部代码。
- 小屏隐藏页面标题下方的辅助说明，大屏仍正常显示。
- API 密钥页的筛选项保持单行横向滚动，列设置固定在右侧，不把筛选项平铺成多行。

## 本地实现

- 管理页和测试场页面说明在 `sm` 断点以下隐藏。
- API 密钥页移动端标题栏缩短到 48px，并压缩标题栏、类型选项卡、筛选栏和表格之间的间距。
- API 密钥筛选轨道复用共享 `useHorizontalScroll`，支持触摸原生横滑和桌面鼠标滚轮横向滚动。
- “列设置”保留在筛选轨道外侧，横向滚动筛选项时仍可直接访问。
- 渠道、API 密钥和请求日志在移动端把页面标题放到全局顶栏的项目名称右侧；360px 宽度下可隐藏品牌文字但保留品牌图标、项目名称和当前页面名称。
- 渠道页继续使用原生横向滚动的操作栏与筛选栏，只压缩顶部操作、筛选和表格之间的纵向留白；关闭分页后表格区域自动延展到底部。
- API 密钥页移动端操作按钮组靠右对齐，并将顶栏到按钮、按钮到类型选项卡的间距压缩到 4px 和约 8px。
- 模型页移动端标题进入全局顶栏；开发者与模型数量摘要隐藏，操作栏、搜索栏和表格之间采用 4–5px 的紧凑间距，同时继续保留横向滚动。

## 与来源的差异

无外部来源。该功能延续上游已有的移动端横向工具栏交互，不改变筛选逻辑、数据请求、列状态或桌面布局。

## 上游收敛

- 2026-08-12 对比 `upstream/unstable@c8de8cf8dac686b4ad71a562e4fffb26df226cef`，未发现等价的全局小屏说明隐藏和 API 密钥单行筛选布局，关系为 `none`。
- 上游若加入等价响应式布局，应优先采用上游样式并移除本地重复断点规则。

## 数据库兼容

- 兼容等级：`none`。
- 纯前端响应式样式和滚动交互变化，不修改 API、Schema 或持久化数据。

## 验证

- Chrome DevTools 360×800：页面说明隐藏，API 密钥筛选栏高度 36px，表格顶部位于 200px，页面无横向溢出。
- Chrome DevTools 390×844：筛选轨道 `scrollWidth=460`、`clientWidth=248`，滚轮事件可将横向位置从 0 移动到 80。
- Chrome DevTools 1440×900：页面说明恢复显示，筛选轨道无额外滚动宽度。
- 渠道页 360×800：顶部操作区和筛选区继续保持原有横向滚动，未改为换行平铺。
- 渠道页 390×844：顶部操作到筛选间距 6px、筛选到表格约 5px；两个横向轨道分别保持 `scrollWidth=670/clientWidth=358` 与 `scrollWidth=609/clientWidth=356`。
- Chrome DevTools 390×844：渠道、API 密钥和请求日志标题均位于项目名称右侧，页面没有全局横向溢出。
- API 密钥页 390×844：按钮组右边缘为 374px，与 16px 页面内边距对齐；顶栏到按钮 4px，按钮到选项卡约 8px。
- 模型页 390×844：标题位于项目右侧，数量摘要为 `display:none`；操作栏到搜索栏 4px、搜索栏到表格约 5px，操作栏仍为横向滚动且页面无全局溢出。
- 模型页 1440×900：本地页面标题、说明和数量摘要正常显示，桌面布局未变化。
- `npx --yes pnpm@10 exec tsc --noEmit`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-12 | 用户反馈与移动端实测 | `5f4935b7023f552f8628d497b53aca9f6b5ff47e` | original：隐藏小屏辅助说明，压缩 API 密钥页垂直间距，并将筛选栏改为单行横向交互。 |
| 2026-08-12 | 用户反馈与移动端实测 | `621c1de32ffdf9c08a1a85279976b897dd9fbd06` | original：将三处列表页标题移入移动端全局顶栏，进一步压缩渠道页纵向间距并保留横向滚动。 |
| 2026-08-12 | 用户反馈与移动端实测 | `b2ba30c77ca611cd251b4169ca352e7d673aaca3` | original：右对齐 API 密钥页操作按钮，并压缩按钮上下留白。 |
| 2026-08-12 | 用户反馈与移动端实测 | `f6dc4e81c531ee9c19ed5b7ed7405ab29b66ad74` | original：将模型页标题移入全局顶栏，隐藏移动端数量摘要并压缩操作栏、搜索栏和表格间距。 |
