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
    - llm/oauth/credentials.go
    - llm/transformer/antigravity
    - llm/transformer/xai/subscription
    - frontend/src/features/channels
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: 7a5a29272c0cb02e996af037cd777c2955b6622d
  relation: upstream-subset
  last_compared_at: 2026-08-22
reconciliations:
  - compared_at: 2026-08-22
    upstream_commits:
      - 7a5a29272c0cb02e996af037cd777c2955b6622d
    relation: upstream-subset
    action: adopted-upstream-xai-subscription-runtime
    local_commits:
      - 23117cab665eaa7e61a04b57689f727ac456cc96
    reverse_commit: 9794e92378b93f2f6117a466fc6e14946904d6b7
    replacement_commits:
      - 9794e92378b93f2f6117a466fc6e14946904d6b7
    residual_commits:
      - 23117cab665eaa7e61a04b57689f727ac456cc96
    reason: 上游已提供完整 xAI Subscription OAuth、固定官方端点、Responses outbound 和配额实现；保留仅属于本地的 Sub2API 单账号 JSON 规范化入口，并删除重复的普通 xAI OAuth 运行时。
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
- Codex 使用 AxonHub 自有的 Codex OAuth client ID；Claude 沿用 Claude Code OAuth client；Gemini 导出可省略 client ID，缺失时使用现有 Code Assist client，并要求 `project_id`；Grok 要求源账号的 client ID。
- Grok 导入只写入规范 OAuth 凭据；运行时完全采用上游 `xai_subscription` 的 token provider 和固定官方订阅网关。普通 xAI API Key 渠道仍只接受 API Key，不会因导入数据转换为 OAuth 渠道。Gemini 导入凭据保留 `project_id`，刷新后不会丢失。
- 渠道窗口保留原 Codex `auth.json` 导入，另加 Sub2API 输入区：Codex 位于 `auth.json` 选项，Claude 和 Gemini 位于官方 OAuth 区，Grok 位于 xAI Subscription 的官方 OAuth 区。导入后会清空普通 API Key 列表；上游保存逻辑固定 Grok Subscription 的 Base URL 和端点。

## 与来源的差异

- Sub2API 的导出格式可表示批量账号、代理、并发、分组及调度状态；本功能明确只读取一份账号的 OAuth 凭据，不导入任何调度、套餐、代理或账号池元数据。Grok 导出中的 `base_url` 同样不导入，避免外部账号数据改变官方订阅渠道的路由。
- AxonHub 的渠道是一渠道一组凭据，而不是 Sub2API 的账号池；因此没有账号创建、状态同步、代理迁移或批量导入接口。
- 外部账号数据不能覆盖任何 OAuth 渠道的官方 Base URL；其余渠道继续使用各自既有的官方地址策略。

## 上游收敛

2026-08-22 对比 `upstream/unstable@7a5a29272c0cb02e996af037cd777c2955b6622d`。上游已提供 xAI Subscription 的 OAuth/SSO、固定官方端点、Responses outbound 和配额运行时，但未提供 Sub2API 单账号 JSON 规范化入口，因此关系为 `upstream-subset`。`9794e923` 已采用上游运行时并移除重复的普通 xAI OAuth 路径，只保留本功能的账号导入残余。

## 数据库兼容

兼容等级为 `additive`。不新增 Ent Schema、迁移、表或列；只向既有 `channels.credentials` JSON 写入已支持的单渠道 OAuth JSON，Gemini 额外使用可选 `project_id`。旧渠道无需回填，旧版本会忽略未知 JSON 字段。部署到绿色前必须制作 SQLite 一致性备份并运行 `PRAGMA quick_check`。

## 验证

- `go test ./internal/server/api -run '^TestNormalizeSub2APICredentials' -count=1`：通过，覆盖 Codex、Claude、Gemini、Grok、平台不匹配，且确认 Grok 导出中的 Base URL 不参与导入。
- `go test ./internal/server/biz -run '^TestXai' -count=1`：通过，验证普通 xAI API Key 与 xAI Subscription OAuth 的通道边界和固定官方请求地址。
- `cd llm && go test ./transformer/xai/subscription ./transformer/antigravity -count=1`：通过。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过。
- `git diff --check`：通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-21 | Sub2API `67380eaf` | `23117cab665eaa7e61a04b57689f727ac456cc96` | reimplemented：新增四个平台的单账号 OAuth 解析、运行时刷新适配和渠道编辑入口；未采用来源代码。 |
| 2026-08-22 | 上游 `7a5a2927` | `9794e92378b93f2f6117a466fc6e14946904d6b7` | reconciliation：采用上游 xAI Subscription 运行时，删除普通 xAI OAuth / 外部 Base URL 路径，保留 Sub2API 单账号导入。 |
