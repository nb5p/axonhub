---
id: playground-api-key-testing
title: 测试场 API Key 路由测试
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 800bb72f4586428fa79905cb244d7840312d3d02
  adopted_commits: []
  last_checked_commit: 800bb72f4586428fa79905cb244d7840312d3d02
  last_checked_at: 2026-08-11
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 20d01ec69e370cac6d29e888f8e3cc7427be5b48
    - 408fe33391b361ce37c63a2efa427804edbe38c2
  modules:
    - frontend/src/features/playground
    - internal/server/api/playground.go
    - internal/server/middleware/auth.go
    - llm/transformer/openai/responses
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

# 测试场 API Key 路由测试

## 目的

让测试场既能固定渠道，也能选择真实 API Key 按其配置文件、模型可见性、配额和 fallback 顺序测试。并保证 OpenAI Responses 上游流式错误为空或嵌套时仍返回可诊断错误，不再只显示 `failed to stream request:`。

## 来源与采用范围

本地原创实现，复用现有 API Key 鉴权、配置文件预览、渠道选择和 AI SDK 数据流协议。

## 本地实现

- 前端增加“API 密钥”测试模式，只列出已启用密钥；模型选项来自所选密钥当前生效配置的可用模型预览。
- 请求通过内部控制头携带所选 API Key ID；后端在进入 LLM 管线前移除该头，并使用数据库中的完整密钥重新执行认证、项目和 IP 白名单检查。
- 成功认证后把 API Key、项目和会话放入请求上下文，因此实际路由遵守该密钥的配置、配额和渠道 fallback，而不是固定测试场渠道。
- `aisdk/datastream` 是前端 AI SDK 的内部流式协议标识，不等于上游渠道格式；渠道执行仍转换为其实际 API 格式。
- OpenAI Responses 流转换同时读取顶层和嵌套错误；上游只发空错误对象时返回稳定的兜底消息和失败状态。

## 与来源的差异

上游测试场仅支持固定渠道和模型网关模式；本地增加真实 API Key 路由模式，不改变普通 API Key 请求和固定渠道测试行为。

## 上游收敛

2026-08-11 比较 `upstream/unstable@800bb72f4586428fa79905cb244d7840312d3d02`，未发现等价的 API Key 测试模式，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。只读取现有 API Key、项目和嵌入式 Profile 数据，不修改 Ent Schema 或持久化结构。

## 验证

- `go test ./transformer/openai/responses -run 'TestOutboundTransformer_StreamTransformation_(Error|Nested|Empty)ErrorEvent' -count=1`：通过。
- `go test ./internal/server/api ./internal/server/middleware -count=1`：通过。
- `pnpm build`：通过。
- 使用本地绿色数据库核对请求 `#45862`：入站 `aisdk/datastream` 已转换为渠道的 `openai/responses`，原始空错误来自上游流事件。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | 请求 `#45862` 诊断 | `20d01ec69e370cac6d29e888f8e3cc7427be5b48` | 修复嵌套及空 Responses 流错误，保留可诊断错误消息。 |
| 2026-08-11 | 用户需求 | `408fe33391b361ce37c63a2efa427804edbe38c2` | 新增按 API Key 可见模型与真实路由执行的测试模式。 |
