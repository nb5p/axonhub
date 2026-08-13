---
id: channel-enabled-first-sorting
title: 渠道启用状态优先排序
status: upstream-pending
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: codex/channel-enabled-first-sorting
  baseline_commit: fc1d27dad4119d06802e0024ce86f3e91cdf740e
  adopted_commits:
    - a96798a8bc9b672ac4afe5b8470f2d8772e5f871
  last_checked_commit: fc1d27dad4119d06802e0024ce86f3e91cdf740e
  last_checked_at: 2026-08-13
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: null
  commits:
    - 823d8c19bb252fd926b0cad80ec26a37c36a5e98
  modules:
    - internal/server/biz/channel_query.go
    - internal/server/biz/channel_query_test.go
    - internal/server/gql/axonhub.graphql
    - frontend/src/features/channels/
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-13
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 渠道启用状态优先排序

## 目的

在渠道主排序字段之外增加独立的“开启状态优先”开关。开启后，启用渠道始终排在禁用或归档渠道之前；每个状态分组内部继续按用户选择的权重、名称、状态等字段升序或降序排列。

渠道页面默认按权重降序并开启该开关。状态和权重变更沿用渠道查询缓存失效机制，列表重新请求后立即按新顺序排列。

## 实现

- GraphQL `QueryChannelInput` 增加 `enabledFirst`，不修改数据库结构。
- 服务端先按启用状态分组，再以 `orderBy` 和渠道 ID 作为稳定次序，并基于组合后的全局顺序执行游标分页。
- 多模型和端点筛选的内存结果采用相同的稳定分组逻辑。
- 状态列排序菜单提供独立开关，不改变当前升序或降序；状态开关持久化到浏览器本地。
- 渠道表声明为服务端排序，避免浏览器只对当前页重新排序而破坏启用优先顺序。

## 上游收敛

功能在 `upstream/unstable@fc1d27da` 的独立分支上实现，提交保持纯净且不含私有账本文档。若提交上游并被接受，应将状态改为 `upstreamed`，记录 PR 和接受提交，并在后续真实上游合并时移除重复私有差异。

## 数据库兼容

- 兼容等级：`none`。
- 仅新增可选 GraphQL 查询参数和前端本地状态；旧客户端不传参数时保持原服务端行为。

## 验证

- `go test ./internal/server/biz -count=1`
- `go test ./internal/server/api -count=1`
- 覆盖权重降序下启用优先、关闭开关后的纯权重顺序、跨页游标以及模型筛选结果。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-13 | `upstream/unstable@fc1d27da` | `823d8c19bb252fd926b0cad80ec26a37c36a5e98` | 在独立官方基线分支实现，并 cherry-pick 到 `ai-slop`。 |
