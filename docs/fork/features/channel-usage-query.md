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
  last_checked_at: 2026-08-22
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 089da0d1df9757294abeb24f304d5fda9c9adced
    - 466f201927f231eb1b164176bf667fc21c76eb1d
    - 89b60f42031d7870c02d63a594d635af567752c2
    - f71cae9869c3f1e4aeed25593f555bf20b460a50
    - 827c3685c10ab213ae10fd2a72fce55140f958c0
    - 4593cf61d9aae4b02e11106c421053b97a2378c6
    - fa2fe5a8febdf45db14b15bae932adfe0c6cbd40
    - 86c06faeceefc128c9c3213c2612faf8cb77bb87
    - bbf3d0fc34f19fd9716e8401c3f4811b1ad357cc
    - f336b09c4f31621b3b919f3caa2964e279f50d7b
    - 85399a55b1c5e8dfb45a29d52f7375125ea188e0
    - bea6adb1d554d7ed1759b22c3d2284a9d5146f26
    - d06afc53f2c12933aa0fd9e083652722bc501bca
    - eedd86af2c1c001472fcd05bdadc012e01d47eac
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
    - frontend/src/features/channels/components/channels-columns.tsx
    - frontend/src/features/channels/data/usage-query-presets.ts
    - frontend/src/lib/usage-query-balance.ts
    - frontend/src/features/system/data/quotas.ts
    - frontend/src/components/quota-badges.tsx
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-22
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

以上四种预设的脚本由服务端持有并在保存时强制覆盖客户端提交内容。管理端可通过“编辑预设”把当前预设脚本复制并转换为 `CUSTOM`，再进行修改；转换后地址、参数和脚本文本都会保留，只有 `CUSTOM` 可保存修改。

所有渠道的用量查询默认关闭。Codex OAuth、Claude OAuth 和 OpenCode Go 只会在编辑窗口中预填对应预设，管理员保存 `enabled=true` 后才会参与后台轮询、全量刷新、单渠道刷新、路由配额判断和右上角配额卡；未保存或已关闭的预设绝不发起用量 HTTP 请求。窗口中的“测试”按钮是一次明确的临时测试，仍可在保存前验证脚本。

Codex、Claude Code、OpenCode Go（含 Anthropic 变体）是脚本查询专用渠道：旧的原生 Codex／Claude checker 不再注册，配额采集设置中也不再列出这三种旧 provider type。升级前持久化的 `usage_query` 结果或旧原生状态同样必须对应保存的 `enabled=true` 才会载入路由缓存；否则不会因陈旧的“耗尽”状态阻止调用。

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

- `balance` 面向 OpenRouter／New API、OpenCode Go 等有可折算货币余额的渠道；不适用金额的 Codex、Claude 应直接省略。
- 所有 `balance.remaining` 都以数值保存、以两位小数展示；数值本身不能承载尾随零，因而脚本不得把它转换成字符串。OpenCode Go 将三个窗口折算后的最小余额作为 `balance: { remaining, unit: "USD" }` 返回，同时保留进度条，不再使用 `text` 承载余额。
- 非 `A$` 的余额统一按“数值 + 空格 + 单位”显示，例如 `剩余：12.00 USD`；不使用本地货币符号格式化为 `US$12.00`。`A$` 保持 `A$12.00`，因为它是 Codex OAuth 专用的实际成本单位。
- `text` 不限制长度和格式，可以写解释文字或纯文本进度摘要。渠道列表最多显示两行；完整内容通过鼠标悬浮、触摸或键盘焦点的气泡查看。提供商配额卡不显示脚本文本。
- `tags` 是简短标签数组，适合套餐名称、并发等不应占用文本区的信息。提供商配额卡只展示 `tags` 与窗口进度条；标签采用紧凑中性色块并与电池图标左对齐。
- `balance` 只在渠道列表和脚本测试结果中展示；提供商配额卡不显示余额、请求数、Token 或 A$。
- `progress.windows` 可包含任意数量的独立窗口，五个或更多同样有效。窗口可只提供 `id`，也可不提供重置时间；缺少可计算的百分比时不会渲染填充条。
- 系统状态以最紧张的已给出窗口为准：`remainingPercent=0` 为耗尽，剩余不超过 20% 为预警。没有 `balance` 或百分比时，不会凭空推断耗尽。

Codex OAuth 才显示本日请求数、Token 和 A$ 实际成本；脚本查询渠道不会因为渠道类型同为 Codex 而获得这些 OAuth 专属信息。

41 号渠道使用相同的配额路由状态：任一日／周／月窗口为 `0%` 时用量查询状态为 `exhausted`、渠道不再参与“仅耗尽时过滤”模式的候选；轮询再次得到所有窗口大于 `0%` 后状态恢复并自动加入候选。此行为依赖系统“配额执行”已启用（绿色现场配置为 `exhausted_only`）。它与该渠道的 `HTTP 429 不重试` 标记相互独立，后者仍只禁止重试和跨渠道切换。

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
- 历史的 `claudecode`、`codex`、`opencode_go` 采集开关会在规范化系统设置时忽略；三者现在统一受渠道级 `usage_query.enabled` 控制。旧的已保存状态不会在未启用脚本时进入路由缓存，首次脚本刷新会写成 `usage_query`。
- 回退到旧版本后编辑含新 JSON 的渠道，旧版本可能重序列化并丢弃未知字段；若需回退，先恢复部署前 SQLite 快照。
- 2026-08-22 比较 `upstream/unstable@49ade6f279eae7aed46858dc121258e922ec9870`，上游未包含此脚本协议或预设，关系为 `none`。

## 验证

- `make generate`：GraphQL 生成成功。
- `go test ./internal/server/biz/provider_quota ./internal/server/biz ./internal/server/gql -count=1`：覆盖 v1／v2 参数兼容、响应头与 `context.now`、窗口校验、四个预设解析、预设脚本服务端固定、自动渠道映射和 GraphQL。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：覆盖管理端类型。
- `git diff --check` 与 locale JSON 解析：检查格式和翻译 JSON。
- `go test ./internal/server/biz/provider_quota ./internal/server/biz -run 'Test(BuiltInUsageQuerySettings|UsageQueryPresetScripts|ProviderQuotaService_(RunQuotaCheck_SkipsUnconfiguredBuiltInUsageQueries|RefreshUsageQueries_SkipsUnconfiguredBuiltInPresets))' -count=1`：通过，覆盖未保存预设不轮询、不允许刷新，及已启用脚本保持可刷新。
- `node --test frontend/src/features/channels/data/usage-query-presets.test.mjs`：通过，覆盖 OpenCode Go 前端预览预设返回 USD 余额且不返回文本。
- `go test ./internal/server/biz/provider_quota ./internal/server/biz -count=1`：通过，覆盖预设默认关闭、旧原生状态不加载到路由缓存、注册 checker 与支持列表一致，以及其他业务包回归。
- `cd frontend && node --test src/features/channels/data/usage-query-presets.test.mjs src/features/channels/data/channel-test-api-formats.test.mjs && node_modules/.bin/tsc --noEmit -p tsconfig.json`：通过。
- `cd frontend && node --test src/lib/usage-query-balance.test.mjs && node_modules/.bin/tsc --noEmit -p tsconfig.json`：通过，覆盖 `USD` 尾随单位、两位小数和 `A$` 前缀。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-20 | 本地需求；上游比较至 `9fb6f1af` | `089da0d1df9757294abeb24f304d5fda9c9adced` | original：接入 Goja 脚本、渠道配置、配额轮询、持久化和展示。 |
| 2026-08-21 | 本地交互扩展；上游比较至 `9fb6f1af` | `466f2019` | original：增加列表结果列、手动刷新和右上角展示开关。 |
| 2026-08-21 | 本地协议扩展；上游比较至 `49ade6f2` | `89b60f42` | superseded：首次引入多窗口结果，已由当前 v2 三段式协议替代。 |
| 2026-08-21 | 本地需求；上游比较至 `49ade6f2` | `f71cae98` | original：确立三段式 v2 协议、四个不可改写的系统预设，并令 Codex／Claude／OpenCode Go 统一使用脚本 checker。 |
| 2026-08-21 | 本地行为修正；上游比较至 `49ade6f2` | `827c3685` | original：仅在渠道凭据是 OAuth 时自动启用 Codex／Claude 预设；普通 API Key 渠道保持未配置，显式保存的自定义查询不受影响。 |
| 2026-08-21 | 本地脚本覆盖；上游比较至 `49ade6f2` | `4593cf61` | original：补齐 Tuzi DayCard 日／周／月窗口夹具验证，绿色 41 号使用无 A$ 的三窗口脚本。 |
| 2026-08-21 | 本地展示扩展；上游比较至 `49ade6f2` | `fa2fe5a8febdf45db14b15bae932adfe0c6cbd40` | original：v2 结果加入 tags；Codex／Claude 套餐转为 tags，OpenCode Go 文本显示三个窗口折算后的最小可用金额，41 号使用套餐／并发 tags 且不显示 A$。 |
| 2026-08-21 | 本地编辑体验；上游比较至 `49ade6f2` | `86c06fae` | original：内置预设可显式转为 `CUSTOM` 后编辑，系统预设本身仍不可被覆盖。 |
| 2026-08-21 | 本地回归覆盖；上游比较至 `49ade6f2` | `bbf3d0fc` | original：验证 41 号渠道任一窗口耗尽会暂停路由、额度恢复后重新可用。 |
| 2026-08-22 | 本地展示修正 | `f336b09c4f31621b3b919f3caa2964e279f50d7b` | original：脚本 tags 与 Codex OAuth 当日统计使用同一摘要行和标签样式。 |
| 2026-08-22 | 本地行为修正 | `85399a55b1c5e8dfb45a29d52f7375125ea188e0` | original：内置预设改为仅预填且默认关闭；只有保存的 `enabled=true` 可触发脚本查询，遗留未启用结果不参与配额卡或路由缓存。 |
| 2026-08-22 | 本地协议与展示修正 | `bea6adb1d554d7ed1759b22c3d2284a9d5146f26` | original：OpenCode Go 最小窗口余额改为 USD `balance`，移除余额文本；所有脚本余额统一两位小数显示。 |
| 2026-08-22 | 本地路径收敛 | `d06afc53f2c12933aa0fd9e083652722bc501bca` | original：移除旧原生 Codex／Claude checker 注册和旧 provider 设置项；关闭脚本时忽略历史原生配额状态，避免陈旧缓存参与路由。 |
| 2026-08-22 | 本地展示修正 | `eedd86af2c1c001472fcd05bdadc012e01d47eac` | original：脚本余额在配额卡、渠道列表和测试结果中统一按 `12.00 USD` 渲染，避免浏览器货币符号改写单位。 |
| 2026-08-22 | 本地展示边界修正 | `3aa763c106b2` | original：配额卡收紧为仅 tags 与窗口进度条；渠道列表仅从脚本结果显示 text 与 balance。 |
