---
id: channel-endpoint-filter
title: 渠道端点类型筛选
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  adopted_commits: []
  last_checked_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  last_checked_at: 2026-08-11
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 437344f057c955b78e941f246a4adb9fc3cbc222
  modules:
    - internal/server/biz/channel_query.go
    - internal/server/biz/channel_query_test.go
    - internal/server/gql/axonhub.graphql
    - frontend/src/components/data-table-faceted-filter.tsx
    - frontend/src/features/channels
    - frontend/src/locales
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

# 渠道端点类型筛选

## 目的

在渠道列表按最终支持的端点类型筛选渠道，快速定位支持 Chat Completions、Responses、Anthropic Messages、Gemini Contents 等格式的渠道。

## 来源与采用范围

本地原创实现，扩展现有渠道筛选栏和 `QueryChannelInput`，不引入外部代码。

## 本地实现

- 筛选器支持同时选择多个端点，多个选择默认按“或”关系匹配，已选项自动置顶并显示数量摘要。
- 服务端按渠道运行时有效端点匹配，包含渠道类型自带的默认端点和用户配置的自定义端点。
- 端点筛选可与名称、供应商、状态、标签和模型筛选组合使用；不同筛选维度之间为“且”关系。
- 运行时有效端点需要业务层解析，因此启用端点筛选时与模型筛选一致：先读取标准条件命中的渠道，再在内存中筛选并返回全部匹配项。

## 与来源的差异

上游基线没有按端点类型筛选渠道的查询参数或前端交互。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价端点筛选，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；仅增加 GraphQL 查询参数和运行时筛选，不修改 Ent Schema 或持久化数据。

## 验证

- `go test ./internal/server/biz -run 'TestChannelService_QueryChannels_(WithEndpointFormats|WithMultipleModelFilters|WithModelFilter)' -count=1`：通过。
- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `437344f057c955b78e941f246a4adb9fc3cbc222` | 新增最终端点多选筛选及服务端匹配。 |
