---
id: feature-id
title: 功能名称
status: active
origin: donor-port
integration_method: adapted
source:
  repository: https://example.com/owner/repository
  branch: main
  baseline_commit: FULL_SHA
  adopted_commits: []
  last_checked_commit: FULL_SHA
  last_checked_at: YYYY-MM-DD
  license: LICENSE_IDENTIFIER
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits: []
  modules: []
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: null
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 功能名称

## 目的

说明为什么需要该功能、用户可观察到的行为，以及不包含哪些能力。

## 来源与采用范围

- 说明每个 adopted commit 提供了什么。
- 说明哪些来源模块只作为参考，哪些代码或设计实际进入本项目。
- 对 donor port 记录许可证判断。

## 本地实现

- 列出稳定的代码和测试入口。
- 说明集成方式，以及本地 commit 与来源 commit 的对应关系。
- 说明配置、接口、运行时任务和性能影响。

## 与来源的差异

- 记录为了适配 AxonHub 架构而做的修改。
- 记录明确忽略的部分和原因。
- 记录后续同步时不得覆盖的本地约束。

## 上游收敛

- 记录最近一次检查的 `upstream/unstable` commit 和比较时间。
- 将关系标记为 `none`、`equivalent`、`upstream-superset`、`upstream-subset` 或 `diverged`。
- 同类实现必须以上游架构、命名、接口和修改方向为准，不维持两套等价实现。
- 默认通过追加的 `🧩` reconciliation commit 抵消或替换旧实现，不重写已经发布的历史。
- 上游只覆盖部分能力时，分别记录抵消重叠部分的提交和只含私有增量的 residual commit。

发生收敛时，把下列结构追加到 front matter 的 `reconciliations`：

```yaml
reconciliations:
  - compared_at: YYYY-MM-DD
    upstream_commits:
      - FULL_SHA
    relation: equivalent
    action: replaced-with-upstream
    local_commits:
      - FULL_SHA
    reverse_commit: null
    replacement_commits: []
    residual_commits: []
    reason: 说明采用上游、反向抵消和保留残余能力的原因
```

`reverse_commit` 仅记录真实、安全执行的反向提交；使用定向 rework 替换旧实现时，将其写入 `replacement_commits`。不适用字段保留为 `null` 或空数组。

经批准整体重建历史时，把旧新映射追加到 `history_rewrites`；已经由上游替代的提交不填入 `new_commits`，而是填入 `upstream_commits`：

```yaml
history_rewrites:
  - rewritten_at: YYYY-MM-DD
    base_commit: FULL_UPSTREAM_SHA
    old_commits:
      - FULL_OLD_SHA
    new_commits:
      - FULL_NEW_SHA
    upstream_commits: []
    disposition: rebuilt
    reason: 说明为何重建、由上游替代或省略
```

## 数据库兼容

- 兼容等级：`none`、`additive`、`data-migration` 或 `breaking`。
- 说明 Schema、数据、备份、迁移和回滚影响。
- 如果是 `breaking`，链接用户批准记录；未批准前不得实施。

## 验证

- 列出自动化测试、手工场景和性能检查。
- 记录最近一次验证结果和环境限制。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| YYYY-MM-DD | `old..new` | `FULL_SHA` | adopted / adapted / reimplemented / ignored 及原因 |
