---
id: responses-prestream-keepalive
title: Responses 首响应前 SSE 保活
status: active
origin: upstream-derived
integration_method: adapted
source:
  repository: https://github.com/looplj/axonhub
  branch: pull/2157/head
  baseline_commit: 2d7d7c860865d7fad65d72b6a3e2535e99358ae0
  adopted_commits:
    - 4c35e6cae3b0d409433e69c6249875cfaae51f9d
  last_checked_commit: fc1d27dad4119d06802e0024ce86f3e91cdf740e
  last_checked_at: 2026-08-13
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 7ba87aa0651cb3772572347b25b747ae40b476d5
  modules:
    - internal/server/api/chat.go
    - internal/server/api/chat_test.go
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: 2157
  accepted_commit: 2d7d7c860865d7fad65d72b6a3e2535e99358ae0
  relation: upstream-subset
  last_compared_at: 2026-08-13
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# Responses 首响应前 SSE 保活

## 目的

当 Codex 的 `compaction_trigger` 触发远程上下文压缩时，上游可能在返回首个 HTTP 响应或 SSE 事件前长时间无输出。已有 SSE keep-alive 只在编排器已经返回流之后启动，无法覆盖这段等待，因此下游客户端或中间层可能先因空闲超时断开。

本功能在已启用 `server.sse_keep_alive` 的流式请求等待编排结果期间发送协议对应的 SSE 心跳。它不改变非流式请求，也不替代上游失败重试。

## 来源与采用范围

- AxonHub PR `#2157` 的早期提交 `4c35e6ca` 曾实现首响应前心跳，但最终合入提交 `2d7d7c86` 移除了该阶段，只保留上游流建立后的心跳。
- 本地采用早期方案的“等待 `Process` 时发送心跳”设计，不直接复制其最终代码；同时处理代码审查指出的已提交响应、写失败取消和写入串行化边界。
- 2026-08-13 检查到的上游 `unstable@fc1d27da` 含 Responses 不完整流重试及 EOF/终止事件修复，但仍未覆盖首个上游响应前的空闲窗口。

## 本地实现

- 仅当请求明确要求 SSE、处理器使用标准 SSE writer、配置已启用且间隔有效时启动等待阶段心跳。
- 心跳 writer 与正式响应严格串行：`Process` 返回后先停止并等待心跳退出，再写流事件或错误。
- 心跳写失败会取消派生的处理上下文，避免上游请求在客户端已断开后继续占用资源。
- 如果心跳已经提交 HTTP 响应，后续处理错误改为 SSE `error` 事件；意外的非流式结果也编码为最终 SSE 事件，避免把普通 JSON 直接拼到 SSE 流中。
- 每个符合条件的等待请求增加一个短生命周期 goroutine；功能关闭或请求非流式时不创建 goroutine。

## 与来源的差异

- 保留上游当前 `sseStreamReader` 和流建立后的 keep-alive 实现，只补首响应前阶段，不维护第二套流读取器。
- 增加可注入的处理函数作为 API handler 回归测试缝，生产路径仍调用原编排器。
- 保留现有默认配置；本功能不会擅自把 `server.sse_keep_alive` 从默认关闭改为开启。
- 对心跳后的普通错误和非流式结果增加合法 SSE 收尾，修复早期实现被代码审查指出的响应损坏风险。

## 上游收敛

- 当前关系为 `upstream-subset`：上游提供流建立后的心跳和较新的不完整流重试，但缺少等待首个上游响应时的心跳。
- 后续若上游重新加入等价的 pre-stream keep-alive，应采用上游命名和生命周期管理，并移除本地重复实现。

## 数据库兼容

- 兼容等级：`none`。
- 不修改 Schema、配置持久化或现有数据，回滚仅需撤销代码提交。

## 验证

- `go test ./internal/server/api -count=1`
- 回归测试先在旧行为下失败：延迟 25ms 的上游处理在 5ms 心跳配置下没有任何 `: keep-alive`。
- 修复后验证首响应前心跳先于 `response.completed`，并覆盖心跳后错误、意外非流式结果、心跳写失败取消以及 JSONBody/SSE 请求识别。
- 生产证据：请求 `#45893`–`#45899` 均包含 `compaction_trigger`，没有响应块，并在约 30 秒后由下游取消；绿色配置当时已启用 15 秒 SSE keep-alive，但日志没有产生心跳，证明旧心跳启动点尚未到达。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-13 | PR `#2157` 早期 `4c35e6ca`、上游 `fc1d27da` | `7ba87aa0651cb3772572347b25b747ae40b476d5` | adapted：补回首响应前心跳并修复早期实现的响应提交与取消边界。 |
