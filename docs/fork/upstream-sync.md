# AxonHub 上游同步记录

## 固定关系

- 上游仓库：`git@github.com:looplj/axonhub.git`
- 上游跟踪分支：`upstream/unstable`
- 私有仓库：`ssh://git@forgejo.109062.xyz:88/tux/axonhub.git`
- 私有成品分支：`ai-slop`

`last_merged_upstream_commit` 只表示最近一次已经成功进入 `ai-slop` 历史的上游 commit，不代表远程上游此刻的最新 head。`local_merge_commit` 在直接以上游为重建基线时为 `null`。

```yaml
last_merged_upstream_commit: dba642a08c91c9696018f6ef185435902ecf60f0
local_merge_commit: aacc271bf6b79db56907ab4641fe56860e3deaef
integrated_at: 2026-08-10
integration_mode: merge
recovery_ref: null
recovery_ref_history: backup/ai-slop-before-responses-id-rewrite-20260810 (created and deleted after validation on 2026-08-10)
private_commit_marker: "🧩"
```

## 同步规则

1. 获取 `upstream/unstable` 最新状态，确认工作区和用户改动安全。
2. 同步前逐项检查 `active`、`upstream-pending` 和 `partially-upstreamed` 功能，搜索上游是否出现同类实现；比较行为、测试和最终文件差异，不能只比较 SHA 或 patch-id。
3. 在 `ai-slop` 默认使用真实 merge，禁止通过 rebase、drop 或强制重建抹掉已有私有分支历史。整体重建必须取得用户明确批准、建立可恢复引用并记录旧新提交映射。
   - 经批准整体重建时，指定上游 commit 必须成为新分支的直接基线；只重放有效私有净差异，并确认 `git merge-base --is-ancestor <upstream> ai-slop` 成功。
4. 冲突解决和同类实现替换均以上游架构、命名、接口及公共行为为准，同时显式重新接入仍需保留的私有残余能力。
5. 上游已覆盖的私有实现通过新的 reconciliation commit 抵消或替换，不删除旧提交；该提交和后续残余能力提交使用 `🧩` 标记。
6. 运行与冲突及收敛模块直接相关的测试；不得借同步之机修改无关功能。
7. 成功合入并验证后，更新本页基线、历史表以及各功能的 `upstream` / `reconciliations`，再推送 `forgejo/ai-slop`。

## 同类实现收敛决策

| 上游关系 | 处理方式 | 功能状态 |
|---|---|---|
| `equivalent` | 采用上游实现，追加 `🧩` 收敛提交抵消本地重复差异 | `upstreamed` |
| `upstream-superset` | 完整采用上游；必要的迁移或测试调整单独提交 | `upstreamed` |
| `upstream-subset` | 抵消重叠部分，将私有增量拆为独立残余提交 | `partially-upstreamed` |
| `diverged` | 暂停自动替换，记录差异并设计以上游方向为准的迁移 | 保持原状态，等待决策 |
| `none` | 继续维护私有实现并跟踪上游 | `active` 或 `upstream-pending` |

“反向提交”是行为概念，不等于必须执行原始 `git revert`。旧提交包含多项能力、经历过冲突解决或与上游代码交织时，应编写定向 reconciliation commit，只移除已被上游覆盖的差异。只有完整 revert 不会删除私有残余或上游实现时，才记录真实的 `reverse_commit`。

当上游只实现私有功能的一部分时，收敛必须拆成两个可审计步骤：

1. `revert(fork): 🧩 reconcile <feature-id> with upstream <short-sha>`：抵消或替换重叠实现。
2. `feat(fork): 🧩 retain <feature-id> residual capability`：只保留上游缺少的能力。

每次事件写入对应功能 YAML：

```yaml
reconciliations:
  - compared_at: YYYY-MM-DD
    upstream_commits:
      - FULL_SHA
    relation: equivalent
    action: replaced-with-upstream
    local_commits:
      - FULL_SHA
    reverse_commit: FULL_SHA_OR_NULL
    replacement_commits: []
    residual_commits: []
    reason: 说明为什么替换、抵消或保留残余能力
```

## Commit 消息约定

- fork 私有代码：`fix(scope): 🧩 subject`、`feat(scope): 🧩 subject`
- fork 账本：`docs(fork): 🧩 subject`
- 上游收敛：`revert(fork): 🧩 reconcile <feature-id> with upstream <short-sha>`
- 私有残余：`feat(fork): 🧩 retain <feature-id> residual capability`
- 上游同步 merge：`merge(upstream): sync unstable at <short-sha>`，不使用 `🧩`
- 上游贡献分支：遵循上游普通 Conventional Commit，不使用 `🧩`

## 历史

| 日期 | 上游 commit | 本地 merge commit | 结果 |
|---|---|---|---|
| 2026-08-05 | `7ed4400595c2d73ca14e0131a86657058f261852` | `e55d8e915524facef57087f4eb6952c8cd9dc127` | 已合入；保留 Codex Alpha Search，同时采用上游的新权限和移动端响应式实现。 |
| 2026-08-06 | `d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af` | `null` | 经用户批准，以该上游 commit 为直接基线重建私有历史；只重放仍有效的 Alpha Search、触摸 tooltip 和维护账本。 |
| 2026-08-10 | `dba642a08c91c9696018f6ef185435902ecf60f0` | `aacc271bf6b79db56907ab4641fe56860e3deaef` | 经用户明确要求重写历史，保留 SSE keepalive、私有响应头转发及其他无关改动，移除 Responses item ID 修复及其集成外壳。 |
| 2026-08-10 | `null` | `null` | 完成历史核对后删除本地恢复引用；该引用从未推送到远程。 |

### 2026-08-06 经批准的历史重建

- 旧 `ai-slop` head：`00da677c447f6b79608fa1a32099c06add033925`
- 可恢复引用：`backup/ai-slop-before-rewrite-20260806`
- 新直接基线：`upstream/unstable@d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af`
- 原则：保留上游完整父链和提交 ID；按功能重建私有有效净差异；不重放已被上游等价吸收的提交和旧 merge 外壳。

| 旧私有提交 | 新提交或上游替代 | 处理 |
|---|---|---|
| `e5721a14140252985914d9746cc2037119c4597b`、`e55d8e915524facef57087f4eb6952c8cd9dc127` 中的 Alpha Search 适配 | `8ab2fd1eb9e5667fb078c5b0239654e511f1e6c9` | 合并为单一带 `🧩` 的有效私有提交。 |
| `7cd527ab29f9640b888c45e7818fbfc8f8fc1c16`、`cb00b35c500411c702fd83ebb38e3c894d0599ec` | `e8e19993720bff05d5a10cc2e153017c6e69220a` | 按新版请求表结构重建 tooltip，保留桌面激活动作并省略 merge 外壳。 |
| `b65930ea7a439e79fb84b1a4470a448b36b69621`、`e55d8e915524facef57087f4eb6952c8cd9dc127` 中的移动端 UI 集成 | 上游 `4971cbe436dab12c5d279493a28e0a2affe68309` | 上游已等价吸收，不再重放私有提交。 |
| `00da677c447f6b79608fa1a32099c06add033925` | `41467b2cef0d2190c02c7968a2c17cccfa27d4da` | 保留原作者和作者日期，重写提交主题并补上 `🧩`，同时纳入本次维护规则。 |

### 2026-08-05 冲突取舍

- Channels 主按钮采用上游的新权限作用域。
- Requests 工具栏采用上游新的移动端筛选面板。
- System 标签页采用上游的移动端滚动和桌面等宽布局。
- Threads、Traces 工具栏保留本地横向滚动支持，并吸收上游状态筛选变化。
- Channel endpoint 和 empty response 逻辑同时保留 Codex Alpha Search 与上游 Moderation 支持。
