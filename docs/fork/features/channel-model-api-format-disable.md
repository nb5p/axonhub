---
id: channel-model-api-format-disable
title: 渠道模型 API 格式禁用
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 37e54737c0e147831c6a0b529b79dda1961c3801
  adopted_commits: []
  last_checked_commit: 29aa13e1dbc86c20303a61fa7bf468ccf8d51240
  last_checked_at: 2026-08-23
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 43228ad1b6ea89f8fd7777a85f9475269affa682
    - 6531cb32d84729d3590549fb67463ffd5cefafdd
    - 5a36ae1eb1a2cabf601cc5c41aeb427a85418274
  modules:
    - frontend/src/features/channels
    - frontend/src/locales
    - internal/objects/channel.go
    - internal/pkg/requestclient
    - internal/server/biz/channel.go
    - internal/server/gql
    - internal/server/orchestrator
    - llm/request_client.go
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-23
reconciliations: []
history_rewrites: []
database:
  impact: additive
  backward_compatible: true
---

# 渠道模型 API 格式禁用

## 目的

允许为同一渠道中的单个模型禁用一个或多个下游端点 API 格式，解决“渠道整体支持某格式、但其中某些模型实际上不兼容”的场景。

例如 OpenCode Go 的 `glm-5.2` 能接受 OpenAI Chat Completions，却会在 Codex 的 OpenAI Responses 请求中错误校验原生工具对象。对该模型的该组合禁用 `openai/responses` 后，普通 Responses 入站请求优先在同一渠道改走 `openai/chat_completions`；只有没有可用端点时才交给既有 fallback 选择其他渠道。

这不是针对 OpenCode Go、GLM 或某个错误码的硬编码；规则适用于任意渠道、模型和端点格式。

## 来源与采用范围

- 本地原创功能，不移植外部代码。
- 2026-08-23 对比 `upstream/unstable@29aa13e1dbc86c20303a61fa7bf468ccf8d51240` 的可见接口、渠道设置、端点选择和测试链路，未发现按“渠道 × 模型 × 下游 API 格式 × 请求客户端”禁用的同类实现。上游的 Responses 兼容修复不提供渠道/模型/客户端范围的选路规则。

## 本地实现

- `channels.settings.disabledModelApiFormats` 是可选数组。每项包含 `model`、`apiFormats` 和可选 `clients`；保存时去除空白、按“模型 + 客户端范围”合并、去重格式，并拒绝空模型、无有效格式或未知客户端类别的项。
- `clients` 缺失或空数组兼容旧设置，表示对全部客户端生效；目前稳定分类为 `codex`、`other`、`unknown`。渠道测试请求没有 Codex UA，因此测试格只把全局规则显示为“已禁用”；“仅 Codex”规则不会阻断普通测试请求。
- 2026-08-23 的实际故障样本 `gid://axonhub/Request/50660` 是 Codex Desktop、`glm-5.2`、`openai/responses`，下游报 `tools.20.function` 缺少 `name`。对应 `gid://axonhub/Request/50663` 的 `glm-5.3` `[1210] Invalid API parameter` 是另一类上游问题，不纳入本规则。
- 规则同时匹配请求模型名和映射后的实际模型名，因此模型映射、前后缀和自动裁剪后的请求仍可命中禁用项。
- 路由在为模型选择端点前过滤该模型禁止的格式。普通 Responses 请求仍可由现有 Transformer 改为 Chat Completions；Remote Compaction 的 Responses 状态不可安全转换，若 Responses 被禁则排除该渠道，绝不伪造 Chat 请求。
- 渠道模型测试中，已禁用的模型×格式显示“已禁用”且不发请求。某单元格测试失败后，可选择“全部客户端”或“仅 Codex”禁用该模型的此格式；批量测试把全局禁用格计为跳过，不会卡住进度。
- 单渠道和批量测试的失败格只显示“失败”标签，完整错误在悬浮或触摸标签后显示，避免错误文本把测试表格撑宽。已禁用标签使用浅灰背景与灰色文字，和失败状态明确区分。
- `testChannel` 额外返回可空 `requestID`。只要本次测试已经写入请求记录，失败提示中的“查看请求详情”链接即可直接跳到该记录；请求在持久化前失败时仍显示完整错误，但不会伪造链接。
- 渠道操作菜单新增“模型 API 格式禁用情况”，弹窗列出模型和格式，并可逐项解除。
- 所有渠道读取与编辑 mutation 的 GraphQL fragments 都返回这个字段，避免其他渠道编辑操作覆盖既有禁用规则。
- GraphQL 通用 `ChannelSettingsInput` 不暴露 Provider Quota / Usage Query 的敏感凭据。保存一般渠道设置时，后端会保留既有的这两部分数据，避免新增禁用规则时抹掉 OpenCode Go 的 `workspaceId`、`authCookie` 等配置。

## 与来源的差异

上游基线仅根据渠道端点整体选择 API 格式，不具备模型或请求客户端级例外。本实现复用现有端点解析、格式转换、候选筛选和渠道测试链路，不引入第二套路由器或 provider 特判。

## 上游收敛

当前关系为 `none`。如果上游提供模型级端点格式能力，应采用其数据模型和选路方向；通过后续 `🧩` reconciliation commit 迁移配置并移除重复代码，而不重写已发布历史。

## 数据库兼容

渠道规则本身保持 `additive`：仅在既有 `channels.settings` JSON 中添加可选 `disabledModelApiFormats.clients` 字段，缺失字段等价于对全部客户端生效，旧数据无需回填。关联的请求客户端持久化字段另见 `request-client-classification` 记录。

旧版本可以读取包含未知 JSON 字段的渠道记录，但若用旧版本保存已配置客户端范围的渠道设置，旧 GraphQL 输入无法带回该字段，可能覆盖范围。部署前应按蓝绿流程创建 SQLite 一致性备份；回退到旧版本前避免编辑相关渠道，或恢复部署前快照。

## 验证

- `go test ./internal/server/orchestrator -run 'Test(PopulateAPIFormat|IsModelAPIFormatDisabled|SpecifiedChannelSelector)' -count=1`：通过，覆盖同渠道 Responses→Chat、全部格式禁用时排除渠道、Remote Compaction 不转换、显式禁用格式报错。
- `go test ./internal/server/biz -run 'Test(NormalizeDisabledModelAPIFormats|ChannelUsageQueryConfig_SaveAndRegularUpdatePreserveSecret)' -count=1`：通过，覆盖规则规范化和一般 settings 更新保留 Provider Quota 凭据。
- `go test ./internal/server/gql -run '^$' -count=1` 与 `make generate`：通过，GraphQL schema 与生成代码一致。
- `go test ./internal/server/gql ./internal/server/orchestrator -run '^$' -count=1`：通过，验证新增 `TestChannelPayload.requestID` 的 Go/GraphQL 编译链路。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过。
- `git diff --check`：通过。
- 本次未构建、部署、重启服务或创建数据库备份。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-23 | `upstream/unstable@37e54737` | `43228ad1b6ea89f8fd7777a85f9475269affa682` | original：新增模型级格式禁用、同渠道安全转换、测试快捷禁用与可视化管理。 |
| 2026-08-23 | 本地测试结果 UX 增量 | `6531cb32d84729d3590549fb67463ffd5cefafdd` | original：失败标签以 Tooltip 展示完整错误并链接测试请求详情；已禁用格式改为浅灰标签。 |
| 2026-08-23 | Codex 专属 `glm-5.2` Responses 故障 | `5a36ae1eb1a2cabf601cc5c41aeb427a85418274` | original：规则增加可选客户端范围，并按入站 UA 分类；仅 Codex 禁用不影响其他客户端。 |
