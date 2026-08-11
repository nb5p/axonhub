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

在仪表盘按 API Key 展示最近 90 天的请求活动热力图，帮助个人用户快速判断密钥的活跃日期和用量分布。日期单元格悬停时展示请求数、Token 数和费用；支持隐藏单个密钥或仅查看单个密钥。

## 来源与采用范围

- `71c2bd7d41d0f675af0312623c3512f4cc3c71dc` 是 `feature/api-key-activity-heatmap` 上的完整本地原创实现。
- 采用前端热力图组件、Dashboard GraphQL 查询和按日聚合解析器；未引入数据库 Schema 或后台定时任务。
- 源分支与 `ai-slop` 具有共同历史，因此通过真实 merge 集成，并保留源提交。

## 本地实现

- GraphQL 查询接收 API Key ID 和日期范围，执行权限及数据可见性检查后按日聚合 `usage_logs`。
- 查询最多接受 100 个 API Key，日期范围最多 366 天；无活动日期由解析器补零。
- 前端默认展示最近 90 天，并在页面挂载期间每 5 分钟刷新一次；页面卸载后不会继续刷新。
- 热力图使用 `react-activity-calendar`，展示请求数、Token 数和费用提示。

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
- Docker 生产镜像构建及绿色实例健康检查：部署阶段执行。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `062da210..71c2bd7d` | `71c2bd7d41d0f675af0312623c3512f4cc3c71dc` | merged：保留源提交，并适配当前 `ai-slop` 仪表盘与 GraphQL 生成代码。 |
