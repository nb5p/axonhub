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
  last_checked_commit: 9fb6f1af148d4b1514106b57fbe34ec54ad01102
  last_checked_at: 2026-08-20
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - e7205275a19f6e0822737ce09edd70bf8e2041db
    - ab06121a107a67e2411605ee866e35c96fb9c4f6
    - ab88ca2641193524c3482c50dc7d32aa322e36d5
    - e20024a6da28610a36ace3713d06113709135318
    - 287c77a51d625fcd372be8ed80ae07ff6d9445c2
  modules:
    - frontend/src/components/quota-badges.tsx
    - frontend/src/features/system/components/quota-settings.tsx
    - frontend/src/features/system/data/system.ts
    - frontend/src/lib/quota-display.ts
    - frontend/src/lib/quota-window-time.ts
    - frontend/src/lib/quota-window-time.test.mjs
    - internal/server/biz/system.go
    - internal/server/gql/system.graphql
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-20
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

本地原创实现，仅改变 Codex 和 OpenCode Go 的配额展示。其他提供商保持原有显示逻辑。

## 本地实现

- Codex 默认采用与 OpenCode Go 相同的“主用量条 + 时间三角”布局。
- 系统设置增加独立的“配额显示”卡片，提供“反转用量（反转显示）”开关和时间窗口样式选择，并与会影响请求路由的“配额执行”分开保存。
- 反转模式将进度条和文案改为剩余百分比，但颜色仍按真实已用量计算风险。
- 反转模式也将时间窗口改为剩余时间：三角标记从右向左移动，时间进度条和悬停文案同步显示剩余百分比。
- 已用与剩余文案统一为“已使用 X%”和“剩余 X%”。
- Codex 窗口在持久化配额数据中优先使用绝对 `reset_at` 计算剩余时间和时间进度，仅当它缺失时才回退到快照 `reset_after_seconds`；primary 和 secondary 窗口共用同一规则。
- 新字段保存在现有 `quota_enforcement_settings` JSON 中；旧值缺少字段时默认显示已用量并采用三角标记。

## 与来源的差异

上游基线没有全局配额显示偏好，且 Codex 与 OpenCode Go 使用不同的时间窗口布局和用量文案。

## 上游收敛

2026-08-20 比较 `upstream/unstable@9fb6f1af148d4b1514106b57fbe34ec54ad01102`，未发现同类全局显示设置或绝对 Codex 重置时间修复；上游仍直接使用缓存的 `reset_after_seconds`，关系保持 `none`。

## 数据库兼容

兼容等级为 `none`；复用现有系统键值表和 JSON 配置，不修改 Ent Schema。旧版本会忽略新增 JSON 字段，新版本会为旧 JSON 补安全默认值。

## 验证

- `make generate`：GraphQL 生成成功。
- `go test ./internal/server/biz -run TestSystemService_QuotaEnforcementDisplaySettings -count=1`：通过，覆盖默认值、旧 JSON 和保存读取。
- `go test ./internal/server/gql -run '^$' -count=1`：GraphQL 包编译通过。
- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- `node --test frontend/src/lib/quota-display.test.mjs`：2 项通过。
- `node --test frontend/src/**/*.test.mjs`：49 项通过，其中 Codex 重置时间 6 项覆盖绝对时间优先、过期时间、相对秒数回退、双窗口、时间进度和本地时区日期。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `e7205275a19f6e0822737ce09edd70bf8e2041db` | 将 Codex 默认布局统一为主用量条和时间三角。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `ab06121a107a67e2411605ee866e35c96fb9c4f6` | 增加全局反转显示、时间窗口样式和统一文案。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `ab88ca2641193524c3482c50dc7d32aa322e36d5` | 将纯显示偏好与配额执行拆为独立卡片和独立保存操作，并明确路由影响文案。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `e20024a6da28610a36ace3713d06113709135318` | 反转时间进度、三角位置和对应文案，使剩余量方向保持一致。 |
| 2026-08-20 | `upstream/unstable@9fb6f1af` | `287c77a51d625fcd372be8ed80ae07ff6d9445c2` | Codex 倒计时和时间进度以绝对 `reset_at` 为权威来源，避免持久化快照过期后与本地日期矛盾。 |
