---
id: channel-usage-query
title: 渠道自定义用量查询
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
  modules:
    - internal/objects/channel.go
    - internal/server/biz/channel_usage_query.go
    - internal/server/biz/provider_quota/usage_query_runtime.go
    - internal/server/biz/provider_quota/usage_query_checker.go
    - internal/server/biz/provider_quota.go
    - internal/server/biz/provider_quota_settings.go
    - internal/ent/schema/provider_quota_status.go
    - internal/server/gql/axonhub.graphql
    - internal/server/gql/axonhub.resolvers.go
    - frontend/src/features/channels/components/channels-columns.tsx
    - frontend/src/features/channels/components/channels-usage-query-dialog.tsx
    - frontend/src/features/channels/data/usage-query.ts
    - frontend/src/features/system/data/quotas.ts
    - frontend/src/components/quota-badges.tsx
    - internal/server/gql/generated.go
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

# 渠道自定义用量查询

## 目的

允许管理员在任意渠道上配置 CC-switch 风格的 JavaScript 对象表达式，由脚本描述查询请求并从 JSON 响应中提取套餐和额度信息。查询结果复用现有“提供商配额”轮询、持久化、缓存、路由状态和前端展示链路。

当前内置 New API 与自定义两个模板，支持覆盖查询 Base URL、查询专用 API Key，以及 New API 所需的用户 ID。该功能不让脚本直接访问网络、文件系统、环境变量或进程。

## 来源与采用范围

功能需求和模板格式来自本地用户提供的 CC-switch 用量查询示例，代码为 AxonHub 本地原创实现，没有移植 CC-switch 源码。JavaScript 运行时依赖 `github.com/dop251/goja@493f22071ef6`，许可证为 MIT。

## 本地实现

- 渠道三点菜单增加“用量查询”，配置弹窗提供模板、启用开关、请求地址覆盖、专用 API Key、New API 用户 ID、脚本编辑、测试和保存。
- 配置读取、保存和测试 GraphQL 接口统一要求 `write_channels`；专用 API Key 仅保存在敏感 credentials JSON 中，读取接口只返回 `apiKeyConfigured`，不回显明文。
- 未配置专用 Key 时，依次使用渠道第一把启用 API Key 和 OAuth access token；New API 模板同时将其映射到 `{{apiKey}}` 与 `{{accessToken}}`。
- Goja 被封装在 `UsageQueryScriptRuntime` 接口后，未来 Rust 重写时可以替换为 `deno_core` 实现，而不改变配额 checker 和展示协议。
- JavaScript 只负责生成 request 对象与运行 extractor；HTTP 请求由 Go 主程序发送，并沿用渠道代理设置。
- 启用自定义查询后，它优先于渠道原有的内置配额 checker；关闭后恢复原 provider checker。保存配置会清除旧配额状态并触发后续刷新。
- extractor 支持兼容的平铺文字字段，以及新的 `text` + `progress.windows` 结构；前端显示套餐、额度明细、任意数量的进度窗口和扩展文本。
- 普通渠道编辑会保留隐藏的用量查询设置和专用 Key，避免整段 JSON 更新时误删配置。
- 渠道列表新增默认可见、可在列设置中隐藏的“用量查询”列；脚本返回文本会把连续换行规整为最多一个换行，因此最多显示两行。
- 列内可以刷新单个渠道，列头右侧可以刷新所有已启用用量查询的渠道；手动刷新不受后台配额采集总开关或渠道启用状态限制，并继续使用最多 8 路并发、现有状态持久化和错误退避。
- 用量查询配置新增“显示在提供商配额中”开关。关闭时仅隐藏右上角的提供商配额条目，列表列始终保留；旧配置缺少该字段时按开启处理，保持既有展示行为。

## 安全与资源边界

- 脚本最大 64 KiB，单次 JavaScript 执行最长 500 ms；Goja 中断可终止死循环。
- HTTP 请求最长 15 秒，请求体最大 256 KiB，响应体最大 1 MiB，最多跟随 3 次重定向。
- 只允许常见 HTTP 方法，禁止 hop-by-hop 请求头和 CRLF 注入。
- 请求 URL 必须与最终配置的 Base URL 同源；每次重定向重新校验。
- 拒绝 unspecified、link-local 和 multicast 地址；为支持局域网自托管 New API，当前不全面禁止私网地址。
- extractor 数字必须为有限值，拒绝 `NaN` 和 `Infinity`。
- 测试 mutation 不持有 GraphQL 数据库事务，避免外部请求期间占用事务和 SQLite 写锁。

## 脚本返回协议

### 兼容策略

旧脚本可以继续直接返回平铺字段，行为不变：

```js
return { isValid: true, remaining: 12, unit: "USD" };
```

新脚本建议返回两个职责明确的对象：`text` 承载原有文字结果，`progress.windows` 承载零到任意多个独立的进度窗口。两者可以单独存在，也可以同时返回；下例中的三个窗口只是示例，五个或更多窗口同样受支持。

```js
return {
  text: {
    isValid: true,
    planName: "Pro",
    remaining: 39.5,
    used: 10.5,
    total: 50,
    unit: "USD",
    extra: "并发 8",
  },
  progress: {
    windows: [
      {
        id: "daily",
        label: "每日额度",
        used: 8,
        total: 10,
        unit: "USD",
        windowStart: "2026-08-18T07:36:46+08:00",
        resetAt: "2026-08-19T07:36:46+08:00",
      },
      { id: "weekly", label: "每周额度", usedPercent: 20 },
      { id: "monthly", label: "每月额度", remaining: 170, total: 220, unit: "USD" },
    ],
  },
};
```

`text` 的字段与旧平铺字段完全一致，均为可选：`isValid`、`invalidMessage`、`remaining`、`unit`、`planName`、`total`、`used`、`extra`。若同时提供 `text` 和旧平铺字段，以 `text` 为准。为了继续兼容旧的管理端调用，测试接口也会同时返回平铺字段和 `text`。

每一个 `progress.windows` 元素都是独立窗口：

- `id`、`label`：可选，建议提供稳定的 `id` 与可读标签；未提供标签时前端使用 `id` 或序号。
- `used`、`total`、`remaining`：三者可组合表达数值。要能计算比例，须提供正数 `total`，并至少提供 `used` 或 `remaining`。
- `usedPercent`：可替代上述数值组合，单位是百分比；允许大于 100，以表达超限。界面会把进度条限制在 100%，配额状态按实际值判定。
- `unit`：可选的窗口单位，不填时不附加单位；它可以与 `text.unit` 不同。
- `windowStart`、`resetAt`：可选 ISO 8601/RFC 3339 时间。只给 `resetAt` 会显示重置倒计时；同时给出两者时还会显示窗口已流逝时间。两者都不是必填项。

`progress` 不会取代文字结果：列表中的简洁文本继续使用 `text`，右上角“提供商配额”卡片则在文字明细下逐一渲染所有窗口。状态计算会取所有可计算窗口中最严重的一个：达到或超过 100% 为耗尽，达到预警阈值为预警。

### 41 号渠道（Tuzi DayCard）示例

将“请求地址覆盖”设置为 `https://api.tu-zi.com`，查询专用 Key 使用该渠道的 API Key，然后使用以下脚本。它请求 `/reseller/v1/quota`，把订阅的日／周／月额度转成三个进度条，同时保留套餐、最小剩余额度、并发数和到期时间等文字信息。

```js
({
  request: {
    url: "{{baseUrl}}/reseller/v1/quota",
    method: "GET",
    headers: {
      "Authorization": "Bearer {{apiKey}}",
      "Accept": "application/json",
      "User-Agent": "cc-switch/1.0",
    },
  },
  extractor: function (response) {
    const data = response && response.data ? response.data : response;
    const subscription = data && data.subscription ? data.subscription : {};
    const isValid = response && response.code === 0 && data && data.status === "active" &&
      (!subscription.status || subscription.status === "active");
    const periods = [
      { id: "daily", label: "每日额度", key: "daily" },
      { id: "weekly", label: "每周额度", key: "weekly" },
      { id: "monthly", label: "每月额度", key: "monthly" },
    ];
    const windows = [];
    const remaining = [];

    periods.forEach(function (period) {
      const limit = Number(subscription[period.key + "_limit"]);
      const used = Number(subscription[period.key + "_used"]);
      if (!Number.isFinite(limit) || limit <= 0 || !Number.isFinite(used)) return;

      const left = Math.max(0, limit - used);
      remaining.push(left);
      windows.push({
        id: period.id,
        label: period.label,
        used: Math.max(0, used),
        total: limit,
        unit: "USD",
        windowStart: subscription[period.key + "_window_start"] || undefined,
        resetAt: subscription[period.key + "_reset_at"] || undefined,
      });
    });

    const extra = [];
    if (data && data.concurrency != null) extra.push("并发 " + data.concurrency);
    if (subscription.expires_at || (data && data.expires_at)) extra.push("到期 " + (subscription.expires_at || data.expires_at));

    return {
      text: {
        isValid: isValid,
        invalidMessage: isValid ? undefined : ((response && response.message) || "API Key 无效或套餐未激活"),
        planName: subscription.group_name || "Tuzi DayCard",
        remaining: remaining.length ? Math.min.apply(null, remaining) : undefined,
        unit: "USD",
        extra: extra.join(" · "),
      },
      progress: { windows: windows },
    };
  },
})
```

该示例有意跳过限额为 `0` 或缺失的周期，避免把“未设上限”误显示为已耗尽。若某个上游只返回已用百分比，可直接在对应窗口返回 `usedPercent`；若它没有给出重置时间，省略 `resetAt` 即可。

## 与来源的差异

- 只兼容 CC-switch 的配置对象、模板变量和 extractor 返回协议，不提供浏览器或 Deno 的全局 API。
- 不提供 `fetch`、异步网络、Node.js 模块、文件系统或宿主 Go 对象；脚本无法绕过 Go 侧的 URL、超时、大小和 header 限制。
- 请求地址覆盖表示替换 `{{baseUrl}}` 的同源基址，不允许脚本把请求转发到另一个来源。
- 当前运行时为过渡方案 Goja；未来 Rust + `deno_core` 迁移应保留 `UsageQueryRequest` / `UsageQueryResult` 的宿主边界和上述安全限制。

## 上游收敛

2026-08-21 比较 `upstream/unstable@49ade6f279eae7aed46858dc121258e922ec9870`，搜索用量查询、Goja 和前端配额脚本均未发现同类实现，关系为 `none`。

## 数据库兼容

兼容等级为 `additive`：

- 渠道配置写入现有 `channels.settings` JSON 的可选 `usageQuery` 字段，专用 Key 写入现有 `channels.credentials` JSON 的可选 `usageQueryApiKey` 字段；旧数据无需回填。
- `usageQuery.showInProviderQuota` 是可选布尔字段，缺失时默认显示在提供商配额中；不新增 Ent 字段、表或迁移。
- 新的 `text` / `progress` 协议只扩展现有 `provider_quota_status.quota_data` JSON；不新增数据库字段或迁移，旧版脚本与旧版状态数据可直接读取。
- `provider_quota_status.provider_type` 的 Ent 枚举增加 `usage_query`。SQLite/PostgreSQL 使用字符串存储；MySQL/TiDB 由 Ent 自动迁移扩展枚举取值。
- 新版本可直接读取旧数据库。旧版本会忽略新 JSON 字段和配额类型；如果回退后用旧版本编辑已配置渠道，未知 JSON 字段可能被重新序列化丢失，因此需要保留配置时应恢复升级前快照。
- 本次开发没有创建或修改部署数据库，也没有创建备份。部署前按现有数据库流程制作一致性快照并执行 `PRAGMA quick_check`；回滚可恢复快照，或接受只丢失该新增功能配置而保留核心渠道数据。

## 性能影响

- 功能关闭时不会执行 JavaScript 或发出额外 HTTP 请求。
- 配额调度查询会读取所有启用渠道，再在内存中筛选内置 provider 或已启用脚本的渠道；这是支持任意渠道类型的固定查询开销。
- 实际查询复用现有配额轮询周期和并发上限，不新增独立常驻 goroutine 或调度器。
- 手动刷新仅在管理员点击时执行，并与普通轮询共用服务互斥锁和 8 路并发上限，不会并行叠加同一渠道的脚本请求。

## 验证

- `make generate`：Ent 与 GraphQL 代码生成成功。
- `go test ./internal/server/biz/provider_quota -count=1`：通过，覆盖现代 JavaScript 语法、变量替换、extractor、缺失 extractor、未知变量、死循环中断、New API 请求、专用 Key 优先、同源限制和状态归一化。
- `go test ./internal/server/biz -count=1`：通过，覆盖专用 Key 不回显、普通渠道编辑保留隐藏配置、右上角显示开关，以及测试临时 Key 不落库。
- `go test ./internal/server/biz/provider_quota ./internal/server/biz ./internal/server/gql -count=1`：通过，覆盖单渠道和全量手动刷新、全局采集关闭时仍可手动查询，以及 GraphQL 生成链路。
- `go test ./internal/server/gql -count=1`：通过。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过。
- locale JSON 解析与 `git diff --check`：通过。
- 未运行 lint、前端 build 或开发服务器重启，符合仓库规则。
- 2026-08-21 协议扩展：`make generate` 与 `go test ./internal/server/biz/provider_quota ./internal/server/biz ./internal/server/gql -count=1` 通过；未运行前端构建或 lint。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-20 | 本地需求；上游比较至 `9fb6f1af` | `089da0d1df9757294abeb24f304d5fda9c9adced` | original：以 Goja 实现可替换脚本运行时，并接入渠道配置、配额轮询、持久化和通用额度展示。 |
| 2026-08-21 | 本地交互扩展；上游比较至 `9fb6f1af` | `466f2019` | original：增加列表结果列、单渠道／全量手动刷新，以及右上角提供商配额展示开关；旧 JSON 配置默认保持显示。 |
| 2026-08-21 | 本地协议扩展；上游比较至 `49ade6f2` | `89b60f42` | original：新增兼容的 `text` + 任意 `progress.windows` 脚本结果，并提供 41 号渠道的日／周／月额度脚本示例。 |
