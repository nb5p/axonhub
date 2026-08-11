---
id: channel-model-multi-filter
title: 渠道模型多选筛选
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
    - 42e7db95a0c834e39cc72f869c72421e5cc45a73
    - b3e3aba6b493ffcdb384835c513d41368f8a0d21
    - 8b0178d010980edb6932d254dd32ad2af7bb8a02
  modules:
    - internal/server/biz/channel_query.go
    - internal/server/biz/channel_query_test.go
    - internal/server/gql/axonhub.graphql
    - frontend/src/components/data-table-faceted-filter.tsx
    - frontend/src/features/channels
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

# 渠道模型多选筛选

## 目的

渠道页面支持同时选择多个模型，并允许按“或”或“且”关系筛选。已选模型在列表中置顶，多个选择在触发器中显示数量摘要。

## 来源与采用范围

本地原创实现，扩展现有渠道筛选组件和 GraphQL 查询；保留旧的单模型参数以兼容已有调用方。

## 本地实现

- 新增 `models` 和 `modelsMatchMode` 查询参数，默认按任一模型命中。
- 服务端基于渠道最终模型集合判断，包含渠道映射、前缀和自动裁剪模型。
- 模型下拉框与服务端筛选共用最终模型集合：隐藏原始模型时仅列出转换后的请求 ID；未隐藏时同时列出原始 ID 和自动裁剪后的 ID。
- 前端模型筛选改为多选，已选项稳定置顶，并提供“或/且”切换。
- “或/且”控制位于“模型”和已选数量之间；只选择一个模型时隐藏，选择至少两个模型后显示并可点击切换。

## 与来源的差异

上游基线只支持单个模型，且筛选触发器不具备本地所需的已选置顶与摘要能力。

## 上游收敛

2026-08-10 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现多模型关系筛选，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；仅扩展 GraphQL 查询与运行时过滤。

## 验证

- `go test ./internal/server/biz -run 'TestChannelService_QueryChannels' -count=1`：通过。
- `go test ./internal/server/biz -run 'TestChannel_GetUnifiedModels' -count=1`：通过，覆盖隐藏和保留原始模型时的自动裁剪 ID。
- `node --test src/features/channels/data/channel-config.test.mjs`：通过，校验筛选下拉请求最终生效模型集合。
- TypeScript 检查通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `42e7db95a0c834e39cc72f869c72421e5cc45a73` | 新增多模型 OR/AND 服务端筛选和前端交互。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `b3e3aba6b493ffcdb384835c513d41368f8a0d21` | 将关系切换移入模型筛选触发器，并仅在多选时显示。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `8b0178d010980edb6932d254dd32ad2af7bb8a02` | 让模型候选与最终渠道模型集合保持一致，正确处理自动裁剪与隐藏原始模型。 |
