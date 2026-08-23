---
id: request-client-classification
title: 请求客户端分类与客户端范围格式禁用
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
    - 5a36ae1eb1a2cabf601cc5c41aeb427a85418274
  modules:
    - llm/request_client.go
    - internal/pkg/requestclient
    - internal/ent/schema/request.go
    - internal/server/biz/request.go
    - internal/server/orchestrator
    - internal/server/gql
    - frontend/src/features/requests
    - frontend/src/features/channels
    - frontend/src/locales
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

# 请求客户端分类与客户端范围格式禁用

## 目的

把入站请求使用的客户端作为可审计的路由条件，解决“同一渠道、同一模型、同一格式只在特定客户端失败”的情况。

首个已验证案例是 `gid://axonhub/Request/50660`：Codex Desktop 用 `openai/responses` 请求 OpenCode Go 的 `glm-5.2` 时，上游报 `tools.20.function` 缺少 `name`。经 CC Switch 将该请求转为 Chat Completions 可以正常调用，因此规则应仅让 Codex 的该组合避开 Responses，并优先在同一渠道改走 Chat Completions。`gid://axonhub/Request/50663` 的 `glm-5.3` `[1210] Invalid API parameter` 属于另一类上游问题，未被扩大纳入此规则。

## 本地实现

- 在入站转换的最早阶段从原始 `User-Agent` 分类，空值为 `unknown`，Codex Desktop、`codex-cli`、`codex_cli`、`codex_cli_rs` 为 `codex`，其他非空值为 `other`。不持久化完整 UA，避免在列表中暴露客户端版本等无关信息。
- 将分类保存到 `llm.Request.Client` 运行时元数据中；它不会进入下游 provider 请求 body。请求创建时把这个值持久化到 `requests.client`。
- 请求列表和详情读取 `client`，默认展示“客户端”列（Codex / 其他 / 未知）；该列遵循现有列设置，可隐藏、可调整顺序。
- `disabledModelApiFormats` 的规则扩展为 `model`、`apiFormats`、可选 `clients`。缺失或空 `clients` 是兼容旧数据的全客户端规则；非空值仅匹配指定分类。模型与客户端范围共同构成规则合并键，避免全局规则和 Codex 专属规则相互覆盖。
- 候选渠道、模型端点和最终端点选择都会使用同一个客户端分类。若 Codex 专属规则禁用 Responses，而渠道仍有 Chat Completions 端点，既有安全转换链路会优先留在当前渠道；Remote Compaction 的 Responses 状态不能伪造转换，仍会排除该渠道。
- 渠道测试不模拟 Codex UA，因此它属于 `unknown`；“仅 Codex”规则不会把普通测试格误标为已禁用。测试失败后的快捷操作可选择“全部客户端”或“仅 Codex”。

## 分析边界

- 该版本开始后，所有新请求都会记录稳定客户端类别，可用请求页按客户端观察模型、格式与渠道的错误关联。
- 历史请求保留默认 `unknown`，不会根据已存的请求头批量反推分类：旧记录的请求头可能不完整、被代理改写，回填会把不确定结论伪装成可信数据。
- 目前规则按稳定类别而非某一 UA 字符串匹配，避免 Codex Desktop 版本变化、CLI 名称差异导致规则失效。若将来需要区分更多调用方，应添加新的稳定枚举和明确的分类测试，而不是把任意 UA 文本写入渠道设置。

## 与来源的差异

截至 `upstream/unstable@29aa13e1dbc86c20303a61fa7bf468ccf8d51240`，上游没有“入站客户端分类持久化”或“渠道 × 模型 × API 格式 × 客户端”的端点禁用机制。上游存在 Responses 兼容相关修复，但不能表达这一客户端范围的路由策略，关系为 `none`。

## 数据库兼容

兼容等级为 `additive`。`requests` 新增默认值为 `unknown` 的不可变枚举列，因此已有请求无需迁移且保持语义安全。部署前必须创建 SQLite 一致性备份并执行源库和副本的 `PRAGMA quick_check`。回退到旧版本时，请勿编辑带有 `disabledModelApiFormats.clients` 的渠道，否则旧输入可能覆盖该范围字段。

## 验证

- `go test ./internal/pkg/requestclient -count=1`：通过，覆盖空 UA、Codex Desktop、Codex CLI 和其他 UA。
- `go test ./internal/server/orchestrator -run 'Test(PersistentInboundTransformer_ClassifiesRequestClient|PopulateAPIFormat|IsModelAPIFormatDisabled)' -count=1`：通过，覆盖入站分类和客户端范围的端点筛选。
- `go test ./internal/server/biz -run 'Test(RequestService_CreateRequestPersistsClient|NormalizeDisabledModelAPIFormats)' -count=1`：通过，覆盖请求持久化和规则规范化。
- `go test ./internal/server/gql -count=1` 与 `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-23 | Codex Desktop / OpenCode Go `glm-5.2` Responses 故障 | `5a36ae1eb1a2cabf601cc5c41aeb427a85418274` | original：入站 UA 分类持久化，渠道格式禁用规则增加客户端范围，避免影响非 Codex Responses 请求。 |
