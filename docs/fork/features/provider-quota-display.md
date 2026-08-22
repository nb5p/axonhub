---
id: provider-quota-display
title: 提供商配额显示偏好
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  adopted_commits: []
  last_checked_commit: 49ade6f279eae7aed46858dc121258e922ec9870
  last_checked_at: 2026-08-22
  license: Apache-2.0
references:
  - repository: https://github.com/Wei-Shaw/sub2api
    commit: 2bc139ab527b4a687546d14510
    license: LGPL-3.0
    usage: visual-and-data-semantics-only
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - e7205275a19f6e0822737ce09edd70bf8e2041db
    - ab06121a107a67e2411605ee866e35c96fb9c4f6
    - ab88ca2641193524c3482c50dc7d32aa322e36d5
    - e20024a6da28610a36ace3713d06113709135318
    - 287c77a51d625fcd372be8ed80ae07ff6d9445c2
    - 698dfa1c36a00eca5d93de68052f361a8a15fcc5
    - bc4f80a4f60ef4a95bab146a0cc83705c93394ca
    - 129e16cc1fbd27aee6ed5deb1a20ba7a376dc43b
    - f336b09c4f31621b3b919f3caa2964e279f50d7b
    - bea6adb1d554d7ed1759b22c3d2284a9d5146f26
    - eedd86af2c1c001472fcd05bdadc012e01d47eac
  modules:
    - frontend/src/components/quota-badges.tsx
    - frontend/src/features/system/data/quotas.ts
    - frontend/src/features/system/components/quota-settings.tsx
    - frontend/src/features/system/data/system.ts
    - frontend/src/lib/quota-display.ts
    - frontend/src/lib/usage-query-balance.ts
    - frontend/src/lib/quota-window-time.ts
    - frontend/src/lib/quota-window-time.test.mjs
    - internal/server/biz/system.go
    - internal/server/gql/dashboard.graphql
    - internal/server/gql/dashboard.resolvers.go
    - internal/server/gql/system.graphql
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-22
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 提供商配额显示偏好

## 目的

统一 Codex 与 OpenCode Go 的配额展示，并允许管理员全局选择显示已用量或剩余量，以及用三角标记或独立进度条展示时间窗口进度。

## 来源与采用范围

本地原创实现，覆盖 Codex 和 OpenCode Go 的配额展示，以及提供商配额弹层内每条渠道的当日用量标签。其他提供商的配额逻辑保持原有显示逻辑。

当日用量标签的视觉和数据语义参考 Sub2API 的 `AccountUsageCell.vue`：请求数、Token、账户实际成本（A$）并列显示。参考仓库采用 LGPL-3.0；本实现没有复制其 Vue 代码、样式文本或数据访问代码，仅按 AxonHub 的 React、GraphQL 与 Ent 架构重新实现。

## 本地实现

- Codex 默认采用与 OpenCode Go 相同的“主用量条 + 时间三角”布局。
- 系统设置增加独立的“配额显示”卡片，提供“反转用量（反转显示）”开关和时间窗口样式选择，并与会影响请求路由的“配额执行”分开保存。
- 反转模式将进度条和文案改为剩余百分比，但颜色仍按真实已用量计算风险。
- 反转模式也将时间窗口改为剩余时间：三角标记从右向左移动，时间进度条和悬停文案同步显示剩余百分比。
- 已用与剩余文案统一为“已使用 X%”和“剩余 X%”。
- Codex 窗口在持久化配额数据中优先使用绝对 `reset_at` 计算剩余时间和时间进度，仅当它缺失时才回退到快照 `reset_after_seconds`；primary 和 secondary 窗口共用同一规则。
- Codex 的时间标记/时间条仅展示重置窗口进度，不参与用量条的颜色严重度计算；颜色始终由真实已用百分比决定。
- 新字段保存在现有 `quota_enforcement_settings` JSON 中；旧值缺少字段时默认显示已用量并采用三角标记。
- 提供商配额渠道行将当天的请求数、总 Token 和 A$ 放在标题下一行、各配额窗口进度条之前，避免挤压渠道名与可用状态；A$ 来自 `usage_logs.total_cost` 的渠道实际成本，不展示下游客户计费的 U$，并固定两位小数。
- Codex OAuth 的请求数、Token、A$ 与用量查询脚本返回的 `tags` 共用同一摘要行；两类标签均采用相同的紧凑中性色块样式，空间不足时整行自然换行。
- 脚本查询的余额在提供商配额卡、渠道列表和脚本测试结果中均固定显示两位小数；OpenCode Go 以 USD 余额而非文本加入这一统一展示路径。
- 脚本余额的通用文案为“剩余：数值 单位”，例如 `剩余：12.00 USD`；不调用浏览器的货币符号本地化。`A$` 继续显示为 `A$12.00`。
- GraphQL 通过当天 `usage_logs` 的渠道分组一次性查询该三项数据，沿用弹层的 `read_channels` 授权边界和 60 秒刷新周期；未改动 Ent Schema 或数据库结构。

## 与来源的差异

上游基线没有全局配额显示偏好，且 Codex 与 OpenCode Go 使用不同的时间窗口布局和用量文案。

## 上游收敛

2026-08-22 比较 `upstream/unstable@49ade6f279eae7aed46858dc121258e922ec9870`，未发现同类全局显示设置、绝对 Codex 重置时间修复或提供商配额弹层当日用量标签；关系保持 `none`。本次在吸收上游通用时间条组件后，显式保留 Codex 的“按真实用量着色”语义。

## 数据库兼容

兼容等级为 `none`；复用现有系统键值表和 JSON 配置，不修改 Ent Schema。旧版本会忽略新增 JSON 字段，新版本会为旧 JSON 补安全默认值。

## 验证

- `make generate`：GraphQL 生成成功。
- `go test ./internal/server/biz -run TestSystemService_QuotaEnforcementDisplaySettings -count=1`：通过，覆盖默认值、旧 JSON 和保存读取。
- `go test ./internal/server/gql -run '^$' -count=1`：GraphQL 包编译通过。
- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- `node --test frontend/src/lib/quota-display.test.mjs`：2 项通过。
- `node --test frontend/src/**/*.test.mjs`：49 项通过，其中 Codex 重置时间 6 项覆盖绝对时间优先、过期时间、相对秒数回退、双窗口、时间进度和本地时区日期。
- `go test ./internal/server/gql -run '^TestProviderQuotaTodayUsageStats$' -count=1`：通过，覆盖当天聚合、按渠道分组、请求数、Token、A$ 实际成本及跨日排除。
- `pnpm --dir frontend exec tsc --noEmit --pretty false`：通过；中途产生的非业务锁文件变更已还原，未纳入提交。
- `frontend/node_modules/.bin/tsc --noEmit --pretty false -p frontend/tsconfig.json`：通过，覆盖当日用量标签换行与 A$ 两位小数调整。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过，覆盖统计标签与脚本标签共用摘要行的类型检查。
- `frontend/node_modules/.bin/tsc --noEmit -p frontend/tsconfig.json`：通过，覆盖余额格式化路径的类型检查。
- `node --test frontend/src/lib/usage-query-balance.test.mjs`：通过，覆盖 USD 尾随单位、两位小数和 A$ 专用前缀。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `e7205275a19f6e0822737ce09edd70bf8e2041db` | 将 Codex 默认布局统一为主用量条和时间三角。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `ab06121a107a67e2411605ee866e35c96fb9c4f6` | 增加全局反转显示、时间窗口样式和统一文案。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `ab88ca2641193524c3482c50dc7d32aa322e36d5` | 将纯显示偏好与配额执行拆为独立卡片和独立保存操作，并明确路由影响文案。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `e20024a6da28610a36ace3713d06113709135318` | 反转时间进度、三角位置和对应文案，使剩余量方向保持一致。 |
| 2026-08-20 | `upstream/unstable@9fb6f1af` | `287c77a51d625fcd372be8ed80ae07ff6d9445c2` | Codex 倒计时和时间进度以绝对 `reset_at` 为权威来源，避免持久化快照过期后与本地日期矛盾。 |
| 2026-08-21 | `upstream/unstable@49ade6f2`；Sub2API `2bc139ab`（仅语义参考） | `698dfa1c36a00eca5d93de68052f361a8a15fcc5` | 在提供商配额弹层增加今日 req、Token 和 A$ 实际成本；不展示 U$，不复制 LGPL 源码。 |
| 2026-08-21 | 本地显示修正 | `bc4f80a4f60ef4a95bab146a0cc83705c93394ca` | 将今日标签移至标题下一行、配额窗口进度条之前，并将 A$ 固定为两位小数。 |
| 2026-08-22 | 上游 `49ade6f2` 时间条重构 | `129e16cc1fbd27aee6ed5deb1a20ba7a376dc43b` | 保留时间标记与时间条，同时明确 Codex 的颜色严重度不受窗口经过时间影响。 |
| 2026-08-22 | 本地展示修正 | `f336b09c4f31621b3b919f3caa2964e279f50d7b` | Codex OAuth 当日统计与用量查询 tags 合并为同一行，并统一紧凑标签样式。 |
| 2026-08-22 | 本地余额展示修正 | `bea6adb1d554d7ed1759b22c3d2284a9d5146f26` | 所有脚本余额固定两位小数，OpenCode Go 改走 USD 余额对象。 |
| 2026-08-22 | 本地余额文案修正 | `eedd86af2c1c001472fcd05bdadc012e01d47eac` | 脚本余额统一显示为“剩余：12.00 USD”，使 47 号渠道等 USD 余额不再显示为 `US$`。 |
