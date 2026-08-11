---
id: api-key-profile-template-sync
title: API Key 配置模板永久联动
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
    - b5ce74722cecf2e49ff4b078832a6e51928b341e
  modules:
    - internal/objects/apikey.go
    - internal/server/biz/api_key.go
    - internal/server/biz/api_key_profile_template.go
    - internal/server/biz/api_key_profile_template_test.go
    - internal/server/gql/axonhub.graphql
    - internal/server/gql/generated.go
    - frontend/src/features/apikeys
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

# API Key 配置模板永久联动

## 目的

为配置模板增加可选的“联动修改”属性。开启后，从任一关联 API Key 修改模板管理的参数，会反向更新模板并同步所有其他关联配置；该属性一经保存为开启状态便不可关闭。未开启的模板维持“本地修改后自动脱离模板”的原行为。

## 来源与采用范围

本地原创扩展，建立在现有模板加载、模板发布和显式 `templateID` 关联机制之上。

## 本地实现

- 在现有 API Key Profile 嵌入式 JSON 中增加可省略的 `templateSync` 布尔值，旧数据缺省为 `false`。
- 新建模板、保存为模板和编辑模板入口提供开关；后端采用单调规则，已开启的模板忽略任何关闭请求。
- 保存当前配置为模板后，当前配置会安全地建立模板关联；后端仅在配置内容与模板一致且项目相同时接受新关联。
- 直接编辑联动配置时，在同一数据库事务内更新模板、所有关联配置和当前 Key；冲突的同模板多份编辑会被拒绝。
- 同步、加载和删除模板后主动失效所有受影响 API Key 的运行时缓存。
- 模板管理、模板加载、API Key 列表和配置卡片统一使用循环箭头图标标识联动模板。
- GraphQL 的 Profile 输入与输出均传递 `templateSync`，并通过 `make generate` 更新生成代码。

## 与来源的差异

上游基线没有 API Key 配置模板功能；本地原有模板只支持模板端向关联配置发布，API Key 端直接修改会解除关联。本功能在保留旧行为的同时增加可选择、不可逆的双向联动模式。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现模板或等价联动实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。不修改 Ent Schema；`templateSync` 存放在现有 API Key 与模板 Profile JSON 中，使用 `omitempty`，旧数据和旧数据库无需迁移，回滚版本会忽略该字段。

## 验证

- `go test ./internal/server/biz -count=1`：通过。
- `go test ./internal/server/gql -run 'TestAPIKeyProfileTemplate|Test.*ProfileTemplate' -count=1`：通过。
- `./node_modules/.bin/tsc --noEmit`：通过。
- `TestSynchronizedTemplatePublishesAPIKeyProfileEdits` 覆盖新关联、Key 端修改、跨 Key 传播、运行时缓存失效和开关不可关闭。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `b5ce74722cecf2e49ff4b078832a6e51928b341e` | 新增模板永久联动、反向发布、缓存失效、开关和专用图标。 |
