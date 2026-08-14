---
id: prompt-protection-cache-refresh-logging
title: 提示词防护规则缓存刷新降噪
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 852b8c6f9d3953309df55d4c7872e12fbfbd45f3
  adopted_commits: []
  last_checked_commit: 852b8c6f9d3953309df55d4c7872e12fbfbd45f3
  last_checked_at: 2026-08-14
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 764ba81f69ba9bafa8b14576724624dac327cf0e
  modules:
    - internal/server/biz/prompt_protection_rule.go
    - internal/server/biz/prompt_protection_rule_test.go
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-14
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 提示词防护规则缓存刷新降噪

## 目的

避免提示词防护规则缓存每 30 秒无变化刷新时持续输出 `cache refreshed` INFO 日志，同时保留启用、禁用、删除和内容更新后的实时缓存替换。

## 本地实现

- 刷新后按稳定 ID 顺序比较旧、新规则的 ID 和 `UpdatedAt`，只有成员或内容发生变化时才返回 `changed=true`。
- 空规则列表的周期轮询不再触发缓存交换和 INFO 日志。
- 初次强制加载仍沿用通用 live cache 行为，不改变启动语义。

## 上游收敛

2026-08-14 对比 `upstream/unstable@852b8c6f`，上游仍在每次轮询时无条件返回 `changed=true`，尚无等价修复，关系为 `none`。

## 数据库兼容

- 兼容等级：`none`。
- 不修改 Schema、规则数据或缓存格式。

## 验证

- `go test ./internal/server/biz -run 'TestPromptProtectionRuleService_(EnabledRulesRefreshDetectsOnlyChanges|DeleteRule|BulkOpsAndListEnabled)$' -count=1`：通过。
- 回归覆盖空列表无变化、规则启用、重复刷新和规则禁用。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-14 | `upstream/unstable@852b8c6f` | `764ba81f69ba9bafa8b14576724624dac327cf0e` | original：用规则成员和更新时间比较替代无条件刷新。 |
