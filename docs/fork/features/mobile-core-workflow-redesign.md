---
id: mobile-core-workflow-redesign
title: 移动端核心工作流重构
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: c8de8cf8dac686b4ad71a562e4fffb26df226cef
  adopted_commits: []
  last_checked_commit: c8de8cf8dac686b4ad71a562e4fffb26df226cef
  last_checked_at: 2026-08-12
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 82d825f9086f7477b2a0836af559dcdcd0de7385
  modules:
    - frontend/src/components/server-side-pagination.tsx
    - frontend/src/components/ui/dialog.tsx
    - frontend/src/components/ui/sheet.tsx
    - frontend/src/features/apikeys/components/apikey-profile-preview-panel.tsx
    - frontend/src/features/apikeys/components/apikeys-profiles-dialog.tsx
    - frontend/src/features/requests/components
    - frontend/src/locales
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

# 移动端核心工作流重构

## 目的

修复 API Key 配置文件管理、请求详情和请求日志在手机窄屏上内容重叠、横向滚动以及关键操作难以触摸的问题。桌面端保留原有双栏弹窗和数据表格，移动端采用适合单手操作的独立信息架构。

## 来源与采用范围

本地原创实现，基于上游 `unstable@c8de8cf8dac686b4ad71a562e4fffb26df226cef` 的现有响应式组件继续扩展，不采用外部项目代码。

## 本地实现

- 配置文件管理在窄屏使用全高弹窗，并通过“配置文件 / 当前状态”两个面板切换编辑区与实时预览；标题、模板操作、生效配置和底部保存区不再互相挤压。
- API 格式选项卡在手机上改为两列换行，搜索框、模型 ID 和主要操作提供至少 48px 的触控区域。
- 项目级和全局请求详情复用同一响应式头部；返回、请求序号、模型、时间和复制操作在手机上纵向排布，详情内的标签页、JSON 操作和执行记录操作按窄屏重排。
- 请求日志在 `md` 以下使用卡片列表，直接展示序号、状态、时间、模型、Token 以及用户当前列设置允许的渠道、密钥、API 格式、缓存命中率、费用和耗时；桌面端继续使用原表格。
- 请求日志筛选在手机上使用全宽模型搜索和底部筛选面板，分页改为固定三段布局并放大翻页目标。
- 共享 Dialog 和 Sheet 的移动端关闭按钮扩大到 48px，桌面尺寸保持不变。
- 新增源码级移动布局回归断言，防止上述页面退回桌面双栏或横向表格实现。

## 与来源的差异

上游已有通用移动端适配，但这些核心工作流仍直接压缩桌面布局。本实现只在移动断点切换信息架构，不改变服务端查询、权限、筛选语义、配置保存格式或桌面布局。

## 上游收敛

2026-08-12 比较 `upstream/unstable@c8de8cf8dac686b4ad71a562e4fffb26df226cef`，未发现配置编辑/预览分屏、请求详情共享移动头部或请求日志移动卡片的等价实现，关系为 `none`。

## 数据库兼容

- 兼容等级：`none`
- 纯前端布局、交互和翻译变化，不修改 GraphQL、Ent Schema、持久化数据或浏览器筛选数据结构。

## 验证

- `pnpm run test:unit`：22 个测试通过，其中 4 个新增移动布局回归断言通过。
- `pnpm exec tsc --noEmit`：通过。
- Chromium 以 390×844、触摸模式渲染隔离测试页：配置编辑、配置预览、请求详情和请求日志的页面宽度均为 390px，无横向溢出。
- 配置编辑、配置预览和请求日志的可见交互目标均不小于 48px；隔离测试使用合成数据，只验证真实组件和样式，不据此推断生产业务数据。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-12 | `upstream/unstable@c8de8cf8` | `82d825f9086f7477b2a0836af559dcdcd0de7385` | 原创重构三个已确认不可用的移动端核心工作流，并补充共享触控尺寸和回归断言。 |
