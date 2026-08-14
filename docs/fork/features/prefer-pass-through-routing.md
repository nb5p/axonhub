---
id: prefer-pass-through-routing
title: 优先透传路由
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
    - 2fb3e82c23d3646e5e58969a4badd7c05e428115
    - 22c5832522f1adf28266fb38c5c7f3dd89a46b1f
  modules:
    - internal/server/biz/system.go
    - internal/server/gql/system.graphql
    - internal/server/orchestrator/candidates.go
    - internal/server/orchestrator/select_endpoints.go
    - frontend/src/features/system
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

# 优先透传路由

## 目的

在系统“常规 → 透传”中增加默认关闭的“优先透传”。开启后，当前请求格式可真实透传的渠道先于需要格式转换的渠道调用；用户也可以开启二级“转换例外”，让明确选中的安全转换继续按原有权重和负载均衡顺序参与优先调用。

## 来源与采用范围

本地原创实现，沿用现有系统键值配置、GraphQL 设置接口和候选渠道负载均衡流程。

## 本地实现

- 使用独立系统键 `system_prefer_pass_through` 持久化，缺失时默认关闭。
- GraphQL `PassThroughSettings` 和系统常规页面同时读写该选项，并保留旧客户端只更新 `enabled` 的兼容性。
- 候选渠道先按透传能力分层，再在每层内部沿用模型关联优先级和现有负载均衡顺序。
- 只有渠道的最终端点格式与入站格式相同、请求体允许透传，并且系统或渠道透传开关有效时，渠道才获得透传优先级。
- 可透传渠道存在时，不允许需要转换的 trace sticky 渠道越过透传层；功能关闭时不执行额外的渠道分组。
- 转换例外使用独立系统键 `system_prefer_pass_through_exceptions` 保存启用状态和有方向的转换键；`Chat Completions → Responses` 与反方向是两个选项。
- 后端根据当前入站转换器和渠道端点能力返回对话、向量、图片、视频共 33 个可选转换方向；GraphQL 在保存时拒绝不受支持的转换键。
- 开启转换例外后，命中例外的转换渠道与真实可透传渠道进入同一优先层，并继续遵循关联优先级、渠道权重、负载均衡和 trace sticky 规则；关闭二级开关时保留选择但不参与路由。
- 系统设置页通过可搜索、按请求类型分组的多选列表管理例外，避免在页面直接铺开全部转换组合。
- 设置读取使用既有系统配置缓存，每个请求最多增加缓存读取，不增加后台任务。

## 与来源的差异

上游基线会优先选择同格式端点，但不会让同格式渠道整体高于更高权重的转换渠道。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价系统选项和路由分层，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；不改变 Ent Schema。首次保存时只在现有系统键值表增加配置项，旧数据库可直接使用，回滚版本会忽略新增键。

## 验证

- `go test ./internal/server/biz -run 'TestSystemService_PreferPassThrough' -count=1`：通过。
- `go test ./internal/server/orchestrator -run 'TestLoadBalancedSelector_Select_PreferPassThrough|TestSupportedPassThroughConversions' -count=1`：通过。
- `go test ./internal/server/gql -run '^$' -count=1`：通过。
- `frontend/node_modules/.bin/tsc --noEmit`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `2fb3e82c23d3646e5e58969a4badd7c05e428115` | 新增系统设置、透传候选分层与兼容 GraphQL 接口。 |
| 2026-08-14 | 本地扩展 | `22c5832522f1adf28266fb38c5c7f3dd89a46b1f` | 增加有方向的协议转换例外、后端能力清单、GraphQL 校验和分组多选设置。 |
