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
    - 94b9edd65a1733aa776bc402cd8130c4ea720612
    - b72cac454d23392b3eac985e7e1e101d288e3f9f
    - 3939673c0e1338d0387f48f8de12921c3c0af714
    - 8d110d917849690ba9688d4585ac8b7d62e417af
    - 8d259c1d1936210ad5726d5de1e23ab29ba3d239
  modules:
    - internal/objects/apikey.go
    - internal/server/biz/api_key.go
    - internal/server/biz/api_key_cache_invalidation_test.go
    - internal/server/biz/api_key_profile_template.go
    - internal/server/biz/api_key_profile_template_test.go
    - internal/server/orchestrator/select_candidates.go
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
- API Key 列表的“生效配置”使用分体按钮：主按钮打开配置管理，右侧下拉仅列出该 Key 已关联的模板，并通过独立 mutation 按稳定的 `templateID` 快捷切换，不覆盖其他配置内容。
- 关联模板的配置统一以模板名作为唯一名称，列表显示为“模板：名称”，配置名称输入框锁定；模板改名、发布和旧别名首次保存时会同步规范化名称与 `activeProfile`。
- 脱离模板时立即清空配置名称并要求用户重新填写；独立配置名称不得与项目内任一模板名称重复，前后端均按忽略大小写和首尾空格的规则校验。
- GraphQL 的 Profile 输入与输出均传递 `templateSync`，并通过 `make generate` 更新生成代码。
- API Key 写操作在返回前同步失效当前进程的运行时缓存，再通过 watcher 通知其他实例；即使 watcher 延迟或丢弃事件，刚完成切换的实例也不会继续路由到旧配置。
- 快捷切换只有在服务端响应确认目标模板确已成为生效配置后才提示成功，并立即用服务端返回值更新 API Key 列表与详情缓存；GraphQL 失败显示具体错误。
- 配置编辑弹窗明确标出“选择尚未生效，保存后才用于路由”，避免把未保存预览误认为后端运行时状态。
- 快捷切换和完整配置保存均写入结构化成功/失败日志，记录 API Key、项目、来源及 `from_profile`/`to_profile`；候选渠道调试日志同时记录实际生效的 Key/项目 profile。

## 与来源的差异

上游基线没有 API Key 配置模板功能；本地原有模板只支持模板端向关联配置发布，API Key 端直接修改会解除关联。本功能在保留旧行为的同时增加可选择、不可逆的双向联动模式。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现模板或等价联动实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。不修改 Ent Schema；`templateSync` 存放在现有 API Key 与模板 Profile JSON 中，使用 `omitempty`，旧数据和旧数据库无需迁移，回滚版本会忽略该字段。

## 验证

- `go test ./internal/server/biz -count=1`：通过。
- `go test ./internal/server/gql -run 'TestAPIKeyProfileTemplate|Test.*ProfileTemplate' -count=1`：通过。
- `go test ./internal/server/biz -run 'TestActivateTemplateProfile|TestLoadTemplate|TestUpdateTemplatePublishes|TestSynchronizedTemplate|TestAPIKeyService_UpdateAPIKeyProfiles' -count=1`：通过。
- `./node_modules/.bin/tsc --noEmit`：通过。
- `node --test src/features/apikeys/*.test.mjs`：通过。
- `api-key-profile-template-activation.test.mjs` 覆盖快捷切换将嵌入配置中的数字模板 ID 转换为 GraphQL GUID 后再提交。
- `TestAPIKeyService_InvalidateAPIKeyCachesInvalidatesLocalCacheSynchronously` 覆盖 mutation 返回前的本机缓存失效，不依赖异步 watcher。
- `api-key-profile-template-activation.test.mjs` 另覆盖服务端响应确认、列表/详情缓存回填、具体错误处理与未保存状态提示。
- `TestSynchronizedTemplatePublishesAPIKeyProfileEdits` 覆盖新关联、Key 端修改、跨 Key 传播、运行时缓存失效和开关不可关闭。
- `TestLoadTemplate_NameConflict`、`TestLoadTemplate_AlreadyLinked` 和 `TestAPIKeyService_UpdateAPIKeyProfiles/Template_names_are_canonical_and_reserved` 覆盖模板唯一名称、旧别名收敛及重名拒绝。
- `pnpm --dir frontend exec tsc --noEmit`：通过；覆盖 API Key 生效配置编辑弹窗的 `cn` 符号解析。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `b5ce74722cecf2e49ff4b078832a6e51928b341e` | 新增模板永久联动、反向发布、缓存失效、开关和专用图标。 |
| 2026-08-11 | 用户反馈 | `94b9edd65a1733aa776bc402cd8130c4ea720612` | 脱离模板时立即恢复唯一默认配置名，并在编辑期间实时校验重名。 |
| 2026-08-15 | 用户反馈 | `b72cac454d23392b3eac985e7e1e101d288e3f9f` | 在 API Key 列表增加仅面向已关联模板的生效配置快捷切换，并使用专用 mutation 避免覆盖完整配置。 |
| 2026-08-15 | 用户反馈 | `3939673c0e1338d0387f48f8de12921c3c0af714` | 取消关联配置的本地别名，统一采用模板名；脱离后强制重新命名，并阻止独立配置与模板重名。 |
| 2026-08-15 | 用户反馈 | `8d110d917849690ba9688d4585ac8b7d62e417af` | 修复快捷切换误将数字模板 ID 直接提交给 GraphQL 的问题，统一发送 `APIKeyProfileTemplate` GUID。 |
| 2026-08-20 | Alma 隐私模式误路由诊断 | 工作区未提交 | 加固 profile 切换的同步缓存失效、服务端确认回填、错误反馈和审计日志，并明确配置弹窗的保存边界；保留既有跨渠道容错语义。 |
| 2026-08-22 | 用户反馈：编辑 API Key 生效配置立即进入 500 | `8d259c1d1936210ad5726d5de1e23ab29ba3d239` | 修复未保存状态提示使用 `cn` 却漏导入的前端运行时错误；该错误会在编辑弹窗初次渲染时中断整页。 |
