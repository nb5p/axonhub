---
id: api-key-activity-heatmap
title: API 密钥活动热力图
status: active
origin: local-original
integration_method: merged
source:
  repository: https://github.com/looplj/axonhub
  branch: feature/api-key-activity-heatmap
  baseline_commit: 062da210ef5b1cdd7d2e6a48f4cf403cdce4f3fc
  adopted_commits:
    - 71c2bd7d41d0f675af0312623c3512f4cc3c71dc
  last_checked_commit: 71c2bd7d41d0f675af0312623c3512f4cc3c71dc
  last_checked_at: 2026-08-11
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 71c2bd7d41d0f675af0312623c3512f4cc3c71dc
    - 410985934a950d47434173fcc7e0fd08ecef2334
    - 55c0df9d5c91a9180f462f60dd679f84a1018ef3
    - 1334dc0d744e67736f9c600354fae5ba138e1fa5
  modules:
    - internal/server/gql/dashboard.graphql
    - internal/server/gql/dashboard.resolvers.go
    - frontend/src/features/dashboard/components/api-key-activity-heatmap.tsx
    - frontend/src/features/dashboard/data/dashboard.ts
    - frontend/src/features/dashboard/index.tsx
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

# API 密钥活动热力图

## 目的

在仪表盘用一张共享热力图展示最近 90 天的 API Key 请求活动，帮助个人用户快速判断所选密钥的整体活跃日期和用量分布。默认选择全部 API Key，用户可以从顶部的多选筛选中增减密钥；日期单元格通过鼠标悬停、键盘聚焦或触摸点按展示汇总请求数、Token、费用和各密钥的请求贡献。

## 来源与采用范围

- `71c2bd7d41d0f675af0312623c3512f4cc3c71dc` 是 `feature/api-key-activity-heatmap` 上的完整本地原创实现。
- `410985934a950d47434173fcc7e0fd08ecef2334` 将逐密钥多图布局修正为所选密钥共用一张聚合热力图。
- 采用前端热力图组件、Dashboard GraphQL 查询和按日聚合解析器；未引入数据库 Schema 或后台定时任务。
- 源分支与 `ai-slop` 具有共同历史，因此通过真实 merge 集成，并保留源提交。
- 源分支完成合并后已删除；实现提交和 merge 历史继续保留在 `ai-slop`。

## 本地实现

- GraphQL 查询接收 API Key ID 和日期范围，执行权限及数据可见性检查后按日聚合 `usage_logs`。
- 查询最多接受 100 个 API Key，日期范围最多 366 天；无活动日期由解析器补零。
- 前端默认展示最近 90 天，并在页面挂载期间每 5 分钟刷新一次；页面卸载后不会继续刷新。
- 前端把选中密钥的每日请求数、Token 和费用按日期相加，再交给唯一的 `react-activity-calendar` 实例展示。
- API Key 多选状态使用通用 `usePersistedFilter` 保存到浏览器本地；首次进入默认选择全部密钥。
- 顶部从左到右显示汇总说明、API Key 多选器和“全选 / 反选 / 全不选”联合按钮，批量操作不会打开选择器。
- 提示气泡除汇总值外，最多列出当天请求量最高的 5 个密钥，其余贡献者显示数量摘要。
- 日期单元格与被截断的密钥名称使用共享 Tooltip，支持桌面 hover、键盘 focus 和移动端触摸点按。

## 与来源的差异

- 合并时保留了 `ai-slop` 仪表盘已有的用户统计、分析卡片和筛选状态持久化功能，仅把热力图接入现有布局。
- GraphQL 生成文件以当前 `ai-slop` Schema 重新生成，没有直接采用冲突版本。

## 上游收敛

- 2026-08-11 已对比 `upstream/unstable` 的 `800bb72f4586428fa79905cb244d7840312d3d02`，未发现等价的 API Key 活动热力图实现，关系为 `none`。
- 后续若上游加入同类 Dashboard 活动视图，应以其查询结构和组件方向为准，并按账本规则收敛重复实现。

## 数据库兼容

- 兼容等级：`none`。
- 只读取并聚合现有 `api_keys` 与 `usage_logs` 数据，不修改 Ent Schema、索引或已有数据。
- 回滚到旧版本不需要数据库处理。

## 验证

- `make generate`：通过，GraphQL 生成代码已与当前 Schema 同步。
- `go test ./internal/server/gql -count=1`：通过。
- 初始合并提交 `590670acec954ad880a87873cb5ce5c688a92b81` 的 Docker 生产镜像构建及绿色实例健康检查：通过。
- 控件布局修正随本批需求执行 `pnpm build`：通过。
- 触摸提示迁移随 `1334dc0d744e67736f9c600354fae5ba138e1fa5` 执行 `pnpm test:unit`（18/18）、生产构建与隔离 Chromium 触摸交互验证：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `062da210..71c2bd7d` | `71c2bd7d41d0f675af0312623c3512f4cc3c71dc` | merged：保留源提交，并适配当前 `ai-slop` 仪表盘与 GraphQL 生成代码。 |
| 2026-08-11 | 用户反馈 | `410985934a950d47434173fcc7e0fd08ecef2334` | reworked：把每个 API Key 一张图改为多选密钥在同一张图中按日期叠加，并持久化选择。 |
| 2026-08-11 | 用户反馈 | `55c0df9d5c91a9180f462f60dd679f84a1018ef3` | refined：重排汇总、密钥选择器与三项批量选择按钮。 |
| 2026-08-12 | 移动端交互审计 | `1334dc0d744e67736f9c600354fae5ba138e1fa5` | reworked：将第三方 hover-only 日期提示与截断密钥名称迁到支持触摸的共享 Tooltip。 |
