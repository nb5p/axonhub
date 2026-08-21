---
id: sub2api-oauth-account-import
title: Sub2API 单账号 OAuth 导入
status: active
origin: donor-port
integration_method: reimplemented
source:
  repository: https://github.com/Wei-Shaw/sub2api
  branch: main
  baseline_commit: 67380eafd5ae2eaa8db910ae738199c3dac62e37
  adopted_commits: []
  last_checked_commit: 67380eafd5ae2eaa8db910ae738199c3dac62e37
  last_checked_at: 2026-08-21
  license: LGPL-3.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 23117cab665eaa7e61a04b57689f727ac456cc96
  modules:
    - internal/server/api/oauth_import.go
    - internal/server/biz/channel_llm.go
    - llm/oauth/credentials.go
    - llm/transformer/antigravity
    - llm/transformer/xai
    - frontend/src/features/channels
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-21
reconciliations: []
history_rewrites: []
database:
  impact: additive
  backward_compatible: true
---

# Sub2API 单账号 OAuth 导入

## 目的

在 AxonHub 的渠道编辑窗口中导入一份 Sub2API OAuth 账号数据，而无需在本机重新走授权流程。范围只包括 Codex、Claude Code、Gemini Code Assist（Antigravity）和 Grok；导入只规范化凭据并回填当前编辑表单，不创建渠道、不发送账号数据到 Sub2API，也不支持数组或批量导入。

## 来源与采用范围

- 参考 Sub2API `main@67380eafd5ae2eaa8db910ae738199c3dac62e37` 的公开账号导出结构：顶层 `platform` / `type` / `credentials`，以及 OAuth 凭据中的令牌、过期时间、项目和客户端字段。
- 来源许可证为 `LGPL-3.0`。本功能没有复制、链接或改写其代码；仅根据 JSON 数据的可观察字段和 OAuth 标准重新实现解析、校验及 AxonHub 运行时适配。
- 单账号输入也接受常见的 `credentials`、`data.credentials`、`tokens` 等封装，以便兼容已有的管理接口响应；若输入包含 `platform`，必须与当前选择的渠道类型匹配。

## 本地实现

- 管理端新增受既有管理员鉴权保护的 `POST /admin/oauth/import/sub2api`。请求为 `{ provider, account_json }`，其中 `provider` 只允许 `codex`、`claudecode`、`gemini`、`grok`。
- 服务端把 `access_token`、`refresh_token`、`id_token`、`client_id`、`scope(s)` 和 `expires_at` / `expires_in` 规范为 AxonHub 的 OAuth JSON；过期时间接受 RFC 3339、Unix 秒或毫秒。解析没有有效过期时间的旧导出时，令牌会在下一次请求前按标准刷新。
- Codex 使用 AxonHub 自有的 Codex OAuth client ID；Claude 沿用 Claude Code OAuth client；Gemini 导出可省略 client ID，缺失时使用现有 Code Assist client，并要求 `project_id`；Grok 要求源账号的 client ID，默认写入订阅 OAuth 网关 `https://cli-chat-proxy.grok.com/v1`。
- Grok 的 xAI 渠道新增 OAuth access-token provider，使用 `https://auth.x.ai/oauth2/token` 刷新令牌；普通 xAI API Key 渠道保持原实现。Gemini 导入凭据则保留 `project_id`，刷新后不会丢失。
- 渠道窗口保留原 Codex `auth.json` 导入，另加 Sub2API 输入区：Codex 位于 `auth.json` 选项，Claude 和 Gemini 位于官方 OAuth 区，Grok 位于 xAI 渠道。导入后会清空普通 API Key 列表；Grok 会同时更新表单的 Base URL。

## 与来源的差异

- Sub2API 的导出格式可表示批量账号、代理、并发、分组及调度状态；本功能明确只读取一份账号的 OAuth 凭据，不导入任何调度、套餐、代理或账号池元数据。唯一例外是 Grok 的 `base_url`，它是令牌实际可用路由的一部分。
- AxonHub 的渠道是一渠道一组凭据，而不是 Sub2API 的账号池；因此没有账号创建、状态同步、代理迁移或批量导入接口。
- Grok 的自动填充 Base URL 只接受 HTTPS 且不包含用户信息的 URL；其余渠道仍沿用各自官方固定地址。

## 上游收敛

2026-08-21 对比 `upstream/unstable@9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca`，没有发现兼容 Sub2API 单账号 OAuth 数据的渠道导入入口。当前关系为 `none`。如果上游增加等价的 OAuth 账号导入功能，应采用上游接口和运行时架构，再用追加的 `🧩` 收敛提交移除重复实现。

## 数据库兼容

兼容等级为 `additive`。不新增 Ent Schema、迁移、表或列；只向既有 `channels.credentials` JSON 写入已支持的单渠道 OAuth JSON，Gemini 额外使用可选 `project_id`。旧渠道无需回填，旧版本会忽略未知 JSON 字段。部署到绿色前必须制作 SQLite 一致性备份并运行 `PRAGMA quick_check`。

## 验证

- `go test ./internal/server/api -run '^TestNormalizeSub2APICredentials' -count=1`：通过，覆盖 Codex、Claude、Gemini、Grok、平台不匹配和非 HTTPS Grok URL。
- `go test ./internal/server/biz -run '^TestXAIChannel_UsesImportedOAuthCredentials$' -count=1`：通过，验证 xAI 导入凭据实际产生 Bearer access token。
- `cd llm && go test ./transformer/antigravity -count=1`：通过。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-21 | Sub2API `67380eaf` | `23117cab665eaa7e61a04b57689f727ac456cc96` | reimplemented：新增四个平台的单账号 OAuth 解析、运行时刷新适配和渠道编辑入口；未采用来源代码。 |
