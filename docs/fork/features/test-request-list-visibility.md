---
id: test-request-list-visibility
title: 测试请求在请求列表中的可见性
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca
  adopted_commits: []
  last_checked_commit: 9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca
  last_checked_at: 2026-08-21
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - b570bc11a3daf067a0b7f6c09e97fb5b6d3cfd41
  modules:
    - frontend/src/features/requests
    - frontend/src/features/system
    - frontend/src/locales
    - internal/contexts
    - internal/scopes
    - internal/server/biz
    - internal/server/gql
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-21
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 测试请求在请求列表中的可见性

## 目的

在系统设置中提供“测试的请求也显示在请求列表中”开关。默认关闭时保持测试记录只出现在渠道测试历史；开启后，渠道模型测试产生的请求会进入请求列表并可用现有来源筛选器筛选。

## 来源与采用范围

- 本地原创功能，不移植外部代码。
- 2026-08-21 对比 `upstream/unstable@9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca`，未发现等价的全局测试请求可见性设置。

## 本地实现

- 使用现有系统键值表保存独立 JSON 配置 `system_test_request_list_settings`，默认关闭；GraphQL 提供读取和更新接口，前端对尚未升级的后端安全降级并禁用开关。
- 请求查询读取该配置；开启时把测试来源记录并入所选项目的请求连接，关闭时维持原有项目范围。
- 仅拥有全局 `read_requests` 权限的用户可通过该设置看到测试记录；项目级用户不会因开关扩大可见范围。
- 请求列表加入测试来源筛选项；测试记录详情使用全局请求详情路由，避免被项目范围再次过滤。
- 前端不再重复注入 `projectID` 查询条件，改由 `X-Project-ID` 和 Ent 隐私规则统一决定项目范围，从而保证筛选和分页均包含启用后的测试记录。

## 与来源的差异

上游基线没有同类全局设置。本实现沿用现有 SystemService、GraphQL 设置模式、请求来源枚举和 Ent 隐私范围，不建立新的请求副本或历史表。

## 上游收敛

当前关系为 `none`。上游若提供同类配置和请求可见性规则，应采用上游配置键、权限边界和路由策略，再用追加的 `🧩` reconciliation commit 清理本地重复逻辑。

## 数据库兼容

兼容等级为 `none`；复用现有系统键值表写入新的独立 JSON 键，不修改 Ent Schema、索引或既有请求数据。旧版本会忽略该配置键，缺失键默认关闭。

## 验证

- `make generate`：通过，GraphQL 生成文件与 schema 一致。
- `go test ./internal/server/biz -run '^TestSystemService_TestRequestListSettings$' -count=1`：通过，覆盖默认值及写后读取。
- `go test ./internal/contexts -run '^TestWithIncludeTestRequests$' -count=1`：通过。
- `go test ./internal/server/gql -run '^$' -count=1`：通过。
- `go test ./internal/scopes -run '^$' -count=1`：通过。
- `frontend/pnpm exec tsc --noEmit`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-21 | `upstream/unstable@9fb6f1af` | `b570bc11a3daf067a0b7f6c09e97fb5b6d3cfd41` | original：新增测试请求列表可见性开关、权限受限的查询范围与详情路由。 |
