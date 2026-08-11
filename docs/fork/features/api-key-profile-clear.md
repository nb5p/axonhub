---
id: api-key-profile-clear
title: API Key 配置文件一键清除
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
    - bf4e3a60a1937763338b84cbb6e5f025c16e8ad8
    - 28109f87639e895529cf7e06b07fa63cbbcc090a
  modules:
    - internal/server/biz/api_key.go
    - internal/server/biz/api_key_test.go
    - frontend/src/features/apikeys/components/apikeys-dialogs.tsx
    - frontend/src/features/apikeys/components/apikeys-profiles-dialog.tsx
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

# API Key 配置文件一键清除

## 目的

允许管理员在 API Key 的配置文件管理界面一次清除全部配置，使该密钥恢复为不受配置文件限制的默认行为。清除操作需要二次确认，不改变 API Key 本身、模板或历史用量。

## 来源与采用范围

本地原创实现。上游基线只能逐项编辑配置，并要求至少保留一个配置，未提供完整清除入口。

## 本地实现

- 配置文件管理的“增加配置”右侧新增“清除配置”按钮和危险操作确认框。
- 复用现有 `updateAPIKeyProfiles` mutation，提交 `activeProfile: ""` 与空配置数组。
- 后端只在配置数组为空且生效配置同时为空时接受清除；其他不存在的生效配置仍按原规则拒绝。
- 清除后失效 API Key、列表和模板关联计数查询缓存。

## 与来源的差异

上游基线没有清除全部配置的可见入口，也不接受合法的空配置集合。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。继续使用 API Key 现有的嵌入式配置 JSON，不修改 Ent Schema；旧版本也能读取清除后的空配置对象。

## 验证

- `go test ./internal/server/biz -run TestAPIKeyService_UpdateAPIKeyProfiles -count=1`：通过。
- `./node_modules/.bin/tsc --noEmit`：通过。
- 中英文 locale JSON 解析通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `bf4e3a60a1937763338b84cbb6e5f025c16e8ad8` | 新增带二次确认的一键清除，并严格限定空配置合法条件。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `28109f87639e895529cf7e06b07fa63cbbcc090a` | 将清除操作移到增加配置右侧，保持配置操作入口集中。 |
