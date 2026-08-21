---
id: channel-usage-query
title: 渠道用量查询脚本
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: b2148eda3f0d68967398287857ff35d27ff5485b
  adopted_commits: []
  last_checked_commit: 49ade6f279eae7aed46858dc121258e922ec9870
  last_checked_at: 2026-08-21
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 089da0d1df9757294abeb24f304d5fda9c9adced
    - 466f201927f231eb1b164176bf667fc21c76eb1d
    - 89b60f42031d7870c02d63a594d635af567752c2
    - f71cae9869c3f1e4aeed25593f555bf20b460a50
    - f2c91bef0f0c50a4b2db141e97736e637cb5faa4
  modules:
    - internal/objects/channel.go
    - internal/server/biz/channel_usage_query.go
    - internal/server/biz/provider_quota/usage_query_runtime.go
    - internal/server/biz/provider_quota/usage_query_checker.go
    - internal/server/biz/provider_quota/usage_query_presets.go
    - internal/server/biz/provider_quota.go
    - internal/server/biz/provider_quota_settings.go
    - internal/server/gql/axonhub.graphql
    - frontend/src/features/channels/components/channels-usage-query-dialog.tsx
    - frontend/src/features/channels/data/usage-query-presets.ts
    - frontend/src/features/system/data/quotas.ts
    - frontend/src/components/quota-badges.tsx
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

# 渠道用量查询脚本

## 目的与当前范围

用量查询由 Go 主程序发 HTTP 请求、由受限 Goja 脚本解析 JSON 响应。结果复用提供商配额的轮询、持久化、路由状态和展示链路。

以下渠道已统一改走这个 JavaScript 运行时，不再注册或调用原有的 Claude Code、Codex、OpenCode Go 用量 checker：

- Codex：仅 OAuth 渠道自动启用内置 `CODEX` 预设，请求 `https://chatgpt.com/backend-api/wham/usage`。
- Claude Code：仅 OAuth 渠道自动启用内置 `CLAUDE_OAUTH` 预设，请求 `https://api.anthropic.com/api/oauth/usage`，带 `anthropic-beta: oauth-2025-04-20`。
- OpenCode Go（含 Anthropic 变体）：内置 `OPENCODE_GO` 预设，请求渠道 Base URL 下的 `/usage`；不再使用网页登录 Cookie 或 HTML 抓取。
- New API：内置 `NEW_API` 预设，请求 `/api/user/self`。

Codex 的“立即兑换重置额度”仍是独立操作，继续调用该提供商的重置 API；它不是用量查询路径。Claude OAuth 端点及窗口命名按 [CodexBar 的 Claude OAuth 说明](https://github.com/steipete/CodexBar/blob/main/docs/claude.md)核对；本项目未移植其实现。

以上四种预设的脚本由服务端持有并在保存时强制覆盖客户端提交内容，管理端仅可查看、测试、选择预设和配置地址／凭据，不能改写预设脚本。只有 `CUSTOM` 可编辑。

## 脚本对象和变量

脚本必须是一个 JavaScript 对象表达式，包含 `request` 和 `extractor`。请求支持 `{{baseUrl}}`、`{{apiKey}}`、`{{accessToken}}`、`{{userId}}` 变量。脚本不能访问网络、文件、环境变量、进程或宿主 Go 对象；HTTP 仅由 Go 侧执行，并沿用渠道代理。

新协议必须设置 `responseVersion: 2`：

```js
({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/usage",
    method: "GET",
    headers: { "Accept": "application/json" },
  },
  extractor: function (response, context) {
    // 返回下文规定的 balance / text / progress。
  },
})
```

`extractor` 只有两个参数：

- `response`：`{ status, headers, body }`。`status` 是 HTTP 状态码；`headers` 的键全部小写，多个同名值以 `, ` 拼接；`body` 是已经解析的 JSON 响应。
- `context`：当前仅提供 `now`，例如 `2026-08-18T07:36:46+08:00`。该值始终为带数值时区、无小数秒的 ISO 8601 文本，脚本可据此补全只有相对重置秒数的上游响应。

未设置 `responseVersion` 的旧脚本仍兼容：其 `extractor(response)` 收到的仍是 JSON 正文，便于已有自定义脚本逐步迁移。

## v2 返回协议

返回对象最多有四部分，均可省略；脚本可以只返回文本、标签、进度条，或任意组合。

```ts
{
  balance?: {
    remaining: number; // 必填，有限数值
    unit?: string;     // 可选自由文本，例如 "A$"
  };
  text?: string;       // 脚本完全控制的说明文本
  tags?: string[];     // 简短状态标签，例如套餐、并发
  progress?: {
    windows: Array<{
      id: string;                 // 必填，^[A-Za-z][A-Za-z0-9_-]*$
      durationSeconds?: number;   // 可选，正整数
      remainingPercent?: number;  // 可选，0..100
      resetAt?: string;           // 可选，YYYY-MM-DDTHH:MM:SS±HH:MM
    }>;
  };
}
```

- `balance` 面向 OpenRouter／New API 等有货币余额的渠道；不适用金额的 Codex、Claude、OpenCode Go 应直接省略。`A$` 在界面固定显示两位小数。
- `text` 不限制长度和格式，可以写解释文字或纯文本进度摘要。列表和配额卡只显示两行；完整内容通过鼠标悬浮、触摸或键盘焦点的气泡查看。
- `tags` 是简短标签数组，适合套餐名称、并发等不应占用文本区的信息。余额存在时显示在余额数字后；没有余额时和电池图标左对齐单独显示。
- `progress.windows` 可包含任意数量的独立窗口，五个或更多同样有效。窗口可只提供 `id`，也可不提供重置时间；缺少可计算的百分比时不会渲染填充条。
- 系统状态以最紧张的已给出窗口为准：`remainingPercent=0` 为耗尽，剩余不超过 20% 为预警。没有 `balance` 或百分比时，不会凭空推断耗尽。

Codex OAuth 才显示本日请求数、Token 和 A$ 实际成本；脚本查询渠道不会因为渠道类型同为 Codex 而获得这些 OAuth 专属信息。

新版本会读取此前缓存的旧平铺结果（`remaining`、`planName`、`extra` 与旧窗口字段）直到下一次成功轮询，因此升级时不需要清空配额缓存。

## 41 号渠道（Tuzi DayCard）脚本

绿色环境的 41 号渠道使用 `CUSTOM`，请求地址覆盖为 `https://api.tu-zi.com`。下面脚本读取 `/reseller/v1/quota` 给出的订阅、并发、过期时间和日／周／月窗口：

```js
({
  responseVersion: 2,
  request: {
    url: "{{baseUrl}}/reseller/v1/quota",
    method: "GET",
    headers: {
      "Authorization": "Bearer {{apiKey}}",
      "Accept": "application/json",
      "User-Agent": "cc-switch/1.0",
    },
  },
  extractor: function (response, context) {
    const body = response.body || {};
    const data = body.data || {};
    const subscription = data.subscription || {};
    if (response.status < 200 || response.status >= 300 || body.code !== 0) {
      throw new Error(body.message || "无法读取 Tuzi DayCard 用量");
    }

    function clamp(value) { return Math.max(0, Math.min(100, value)); }
    function formatTime(value) {
      const milliseconds = typeof value === "number"
        ? (value < 1000000000000 ? value * 1000 : value)
        : Date.parse(value || context.now);
      if (!Number.isFinite(milliseconds)) return undefined;
      const date = new Date(milliseconds), pad = function (n) { return String(n).padStart(2, "0"); };
      const offset = -date.getTimezoneOffset(), sign = offset >= 0 ? "+" : "-", absolute = Math.abs(offset);
      return date.getFullYear() + "-" + pad(date.getMonth() + 1) + "-" + pad(date.getDate()) + "T" + pad(date.getHours()) + ":" + pad(date.getMinutes()) + ":" + pad(date.getSeconds()) + sign + pad(Math.floor(absolute / 60)) + ":" + pad(absolute % 60);
    }

    const windows = ["daily", "weekly", "monthly"].map(function (id) {
      const limit = Number(subscription[id + "_limit"]);
      const used = Number(subscription[id + "_used"]);
      if (!Number.isFinite(limit) || limit <= 0 || !Number.isFinite(used)) return null;
      const item = { id: id, remainingPercent: clamp(100 - used * 100 / limit) };
      const start = Date.parse(subscription[id + "_window_start"] || "");
      const reset = Date.parse(subscription[id + "_reset_at"] || "");
      if (Number.isFinite(start) && Number.isFinite(reset) && reset > start) item.durationSeconds = Math.round((reset - start) / 1000);
      const resetAt = formatTime(subscription[id + "_reset_at"]);
      if (resetAt) item.resetAt = resetAt;
      return item;
    }).filter(Boolean);

    const tags = [subscription.group_name || "Tuzi DayCard"];
    if (data.concurrency != null) tags.push("并发: " + data.concurrency);
    return { tags: tags, progress: { windows: windows } };
  },
})
```

限额为 `0`、缺失，或没有已用值的窗口会被省略，避免把“未设置上限”误判成耗尽。套餐和并发通过 `tags` 显示；脚本允许没有重置时间的窗口，并刻意忽略 `fuel_pack.available_usd`：该字段不是 Codex OAuth 的 A$ 余额。

## 安全与资源边界

- 脚本最大 64 KiB，单次 JavaScript 执行最长 500 ms，可中断死循环。
- HTTP 请求最长 15 秒，请求体最大 256 KiB，响应体最大 1 MiB，最多跟随 3 次重定向。
- 仅允许常见 HTTP 方法；禁止 hop-by-hop 头、CRLF 注入和跨 Base URL 同源请求。重定向会再次校验。
- 拒绝 unspecified、link-local、multicast 地址；为局域网自托管服务保留私网访问。
- 所有数值必须有限；`NaN`、`Infinity`、非法窗口 ID、非正窗口时长、超出 0..100 的百分比和不带数值时区／带小数秒的 `resetAt` 都会被拒绝。
- 专用 API Key 仅保存在敏感 credentials JSON；读取接口只返回 `apiKeyConfigured`，不回显密钥。测试 mutation 不占数据库事务。

## 数据库兼容与上游

- `channels.settings.usageQuery` 和 `channels.credentials.usageQueryApiKey` 都是既有 JSON 列中的可选字段；`provider_quota_status.provider_type` 的 `usage_query` 枚举由早期功能引入。本次 v2 不新增 Schema、表、字段或数据回填。
- 历史的 `claudecode`、`codex`、`opencode_go` 采集开关会在规范化系统设置时忽略；三者现在统一受 `usage_query` 采集开关控制。旧的已保存状态仍可安全读取，首次脚本刷新会写成 `usage_query`。
- 回退到旧版本后编辑含新 JSON 的渠道，旧版本可能重序列化并丢弃未知字段；若需回退，先恢复部署前 SQLite 快照。
- 2026-08-21 比较 `upstream/unstable@49ade6f279eae7aed46858dc121258e922ec9870`，上游未包含此脚本协议或预设，关系为 `none`。

## 验证

- `make generate`：GraphQL 生成成功。
- `go test ./internal/server/biz/provider_quota ./internal/server/biz ./internal/server/gql -count=1`：覆盖 v1／v2 参数兼容、响应头与 `context.now`、窗口校验、四个预设解析、预设脚本服务端固定、自动渠道映射和 GraphQL。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：覆盖管理端类型。
- `git diff --check` 与 locale JSON 解析：检查格式和翻译 JSON。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-20 | 本地需求；上游比较至 `9fb6f1af` | `089da0d1df9757294abeb24f304d5fda9c9adced` | original：接入 Goja 脚本、渠道配置、配额轮询、持久化和展示。 |
| 2026-08-21 | 本地交互扩展；上游比较至 `9fb6f1af` | `466f2019` | original：增加列表结果列、手动刷新和右上角展示开关。 |
| 2026-08-21 | 本地协议扩展；上游比较至 `49ade6f2` | `89b60f42` | superseded：首次引入多窗口结果，已由当前 v2 三段式协议替代。 |
| 2026-08-21 | 本地需求；上游比较至 `49ade6f2` | `f71cae98` | original：确立三段式 v2 协议、四个不可改写的系统预设，并令 Codex／Claude／OpenCode Go 统一使用脚本 checker。 |
| 2026-08-21 | 本地行为修正；上游比较至 `49ade6f2` | `f2c91bef` | original：仅在渠道凭据是 OAuth 时自动启用 Codex／Claude 预设；普通 API Key 渠道保持未配置，显式保存的自定义查询不受影响。 |
| 2026-08-21 | 本地展示扩展；上游比较至 `49ade6f2` | `76a90bfa62356b46907a168c2141b57061ac3990` | original：v2 结果加入 tags；Codex／Claude 套餐转为 tags，OpenCode Go 文本显示三个窗口折算后的最小可用金额，41 号使用套餐／并发 tags 且不显示 A$。 |
