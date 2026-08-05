---
id: codex-alpha-search
title: Codex Alpha Search 请求代理
status: active
origin: donor-port
integration_method: adapted
source:
  repository: https://github.com/Wei-Shaw/sub2api
  branch: main
  baseline_commit: e316ebf52838a89d57fc790981cce7520f819ac8
  adopted_commits:
    - 52071d391b5b2a4e4e0940aea85fc731857c6d07
    - 33b1d772f734d70470269d5696fa2c2e2bd3d884
  last_checked_commit: 33b1d772f734d70470269d5696fa2c2e2bd3d884
  last_checked_at: 2026-08-06
  license: LGPL-3.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 8ab2fd1eb9e5667fb078c5b0239654e511f1e6c9
  modules:
    - internal/server/api
    - internal/server/biz
    - internal/server/orchestrator
    - llm/pipeline
    - llm/transformer/openai/codex
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-06
reconciliations: []
history_rewrites:
  - rewritten_at: 2026-08-06
    base_commit: d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af
    old_commits:
      - e5721a14140252985914d9746cc2037119c4597b
      - e55d8e915524facef57087f4eb6952c8cd9dc127
    new_commits:
      - 8ab2fd1eb9e5667fb078c5b0239654e511f1e6c9
    upstream_commits: []
    disposition: rebuilt
    reason: 在最新上游直接基线上重放 Alpha Search 的有效净差异，并把原同步 merge 中的冲突适配折入单一私有提交。
database:
  impact: none
  backward_compatible: true
---

# Codex Alpha Search 请求代理

## 目的

让 AxonHub 能够识别并代理 Codex Alpha Search 请求，保持其请求路由、候选端点选择、OpenAI 兼容渠道映射和响应处理正常工作。

## 来源与采用范围

- 功能方案来自 sub2api PR [#4063](https://github.com/Wei-Shaw/sub2api/pull/4063)，其修复在 Codex GPT-5.6 Responses Lite 模式下 `/alpha/search` 独立端点返回 404 的问题；来源实现提交为 `52071d391b5b2a4e4e0940aea85fc731857c6d07`，上游合入提交为 `33b1d772f734d70470269d5696fa2c2e2bd3d884`。
- 本地没有直接 cherry-pick 不同架构项目的提交，而是把其路由、OAuth 转发和原样响应语义适配到 AxonHub pipeline。旧实现提交为 `e5721a14140252985914d9746cc2037119c4597b`，2026-08-06 重建后对应提交为 `8ab2fd1eb9e5667fb078c5b0239654e511f1e6c9`；最初的 AxonHub 开发基线为 `6346a4fed4d5435dfa1b3cacaffac758684ccbd7`。
- 本地引用 `codex/fix-codex-alpha-search` 当前停留在开发基线，并不包含功能提交，后续不得把该引用误记为已完成的来源分支。

## 本地实现

- `internal/server/api`：接收和代理相关 Codex 请求，并覆盖 HTTP 行为测试。
- `internal/server/biz`：识别 Alpha Search 端点和 OpenAI 兼容渠道映射。
- `internal/server/orchestrator`：选择 Codex 候选端点和流式策略。
- `llm/transformer/openai/codex`：实现 Alpha Search 请求转换及测试。
- `llm/pipeline`：避免把相关响应错误判定为空响应。

重建提交 `8ab2fd1eb9e5667fb078c5b0239654e511f1e6c9` 同时包含旧提交 `e5721a14` 的功能净差异和旧同步 merge `e55d8e91` 中与上游 Moderation 支持共存所需的适配。

## 与来源的差异

来源项目采用其自身的网关 handler、账号调度和故障切换体系；本地实现改为使用 AxonHub 的 Inbound/Outbound Transformer、Channel Endpoint、候选端点选择和 empty response pipeline。两者共享的行为契约是注册独立搜索入口、OAuth 转发至 ChatGPT Alpha Search、保留模型映射并原样返回非流式响应。

后续同步上游时，重点检查 channel endpoint、LLM 常量、候选端点和 empty response 逻辑，不能用上游新枚举或特殊端点处理覆盖 Alpha Search。

## 上游收敛

- 2026-08-06 检查 AxonHub `upstream/unstable@d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af`：没有 Alpha Search 路由、RequestType、APIFormat 或 Transformer，关系为 `none`。
- AxonHub Issue [#2048](https://github.com/looplj/axonhub/issues/2048) 仍处于开放状态，描述同一个 `/v1/alpha/search` 404 问题，尚无关联 PR。
- 上游将来实现后必须先判断 `equivalent`、`upstream-superset`、`upstream-subset` 或 `diverged`，再按总账规则创建 `🧩` reconciliation commit；不能只因提交 SHA 不同而重复保留本实现。

## 数据库兼容

- 兼容等级：`none`
- 不修改 Ent Schema、索引或现有数据。

## 验证

- 根模块覆盖 chat API、渠道端点映射和候选端点选择测试。
- `llm/` 独立模块覆盖 Alpha Search 转换及 empty response 测试。
- 2026-08-05 上游同步后，相关目标测试已通过。
- 2026-08-06 历史重建后，`internal/server/api`、`internal/server/biz`、`internal/server/orchestrator`、`llm/pipeline` 和 `llm/transformer/openai/codex` 目标测试全部通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-07-16 | sub2api `e316ebf5..52071d39` / merge `33b1d772` | `e5721a14140252985914d9746cc2037119c4597b` | 将外部 Alpha Search 修复语义适配到 AxonHub 架构。 |
| 2026-08-05 | 上游同步至 `7ed44005` | `e55d8e915524facef57087f4eb6952c8cd9dc127` | 适配上游 Moderation 变化并保留 Alpha Search。 |
| 2026-08-06 | AxonHub 上游检查至 `d6ed9c62` | `null` | 上游仍无等价实现，保持 `active`；后续跟踪 Issue `#2048`。 |
| 2026-08-06 | 历史重建至上游 `d6ed9c62` | `8ab2fd1eb9e5667fb078c5b0239654e511f1e6c9` | 将旧实现及 merge 适配压缩为单一带 `🧩` 的有效私有提交。 |
