# `ai-slop` 私有分支维护账本

本目录记录 `ai-slop` 相对 AxonHub 上游的长期差异，以及从其他项目移植功能时的来源、取舍和后续同步基线。它只服务于私有分支维护，不应进入提交给 `looplj/axonhub` 的上游 PR。

## 每次任务必读

AI 每次处理本仓库任务时都必须先阅读本页。涉及下列内容时继续阅读：

- 同步 AxonHub 上游：[`upstream-sync.md`](upstream-sync.md)
- Schema、数据迁移或兼容性判断：[`database-compatibility.md`](database-compatibility.md)
- 已登记功能：下表对应的功能文档
- 新功能：[`features/_template.md`](features/_template.md)

## 当前功能索引

| 功能 ID | 类型 | 状态 | 核心来源或提交 | 数据库影响 | 记录 |
|---|---|---|---|---|---|
| `codex-alpha-search` | 外部方案适配 | `active` | sub2api PR `#4063` / `8ab2fd1e` | `none` | [`features/codex-alpha-search.md`](features/codex-alpha-search.md) |
| `mobile-request-route-tooltip` | 本地原创、计划贡献上游 | `upstream-pending` | `e8e19993` | `none` | [`features/mobile-request-route-tooltip.md`](features/mobile-request-route-tooltip.md) |
| `mobile-ui-improvements` | 本地原创、已贡献上游 | `upstreamed` | 上游 `4971cbe4` / PR `#1896` | `none` | [`features/mobile-ui-improvements.md`](features/mobile-ui-improvements.md) |
| `responses-reasoning-item-ids` | 本地原创、已撤回 | `retired` | `5571aa81` / `0e066896`（历史重写已移除） | `none` | [`features/responses-reasoning-item-ids.md`](features/responses-reasoning-item-ids.md) |
| `channel-api-key-copy` | 本地原创 | `active` | `236ff5c2`, `c636e07f` | `none` | [`features/channel-api-key-copy.md`](features/channel-api-key-copy.md) |
| `request-log-layout` | 本地原创 | `active` | `d476d57b`, `81c62c6e`, `1afef2be` | `none` | [`features/request-log-layout.md`](features/request-log-layout.md) |
| `channel-model-multi-filter` | 本地原创 | `active` | `42e7db95`, `b3e3aba6` | `none` | [`features/channel-model-multi-filter.md`](features/channel-model-multi-filter.md) |
| `filter-state-persistence` | 本地原创 | `active` | `36b9e737` | `none` | [`features/filter-state-persistence.md`](features/filter-state-persistence.md) |
| `provider-quota-display` | 本地原创 | `active` | `e7205275`, `ab06121a`, `ab88ca26`, `e20024a6` | `none` | [`features/provider-quota-display.md`](features/provider-quota-display.md) |
| `channel-endpoint-summary-column` | 本地原创 | `active` | `ba6c2fca`, `d04464ee` | `none` | [`features/channel-endpoint-summary-column.md`](features/channel-endpoint-summary-column.md) |
| `channel-endpoint-filter` | 本地原创 | `active` | `437344f0` | `none` | [`features/channel-endpoint-filter.md`](features/channel-endpoint-filter.md) |
| `channel-provider-tabs-default-hidden` | 本地原创 | `active` | `7f84e474` | `none` | [`features/channel-provider-tabs-default-hidden.md`](features/channel-provider-tabs-default-hidden.md) |
| `prefer-pass-through-routing` | 本地原创 | `active` | `2fb3e82c` | `none` | [`features/prefer-pass-through-routing.md`](features/prefer-pass-through-routing.md) |
| `api-key-profile-access-preview` | 本地原创 | `active` | `457caa1b`, `9a970fd3`, `701f33de`, `dccc2740`, `7fecec65` | `none` | [`features/api-key-profile-access-preview.md`](features/api-key-profile-access-preview.md) |
| `api-key-profile-clear` | 本地原创 | `active` | `bf4e3a60`, `28109f87` | `none` | [`features/api-key-profile-clear.md`](features/api-key-profile-clear.md) |
| `api-key-profile-template-sync` | 本地原创 | `active` | `b5ce7472` | `none` | [`features/api-key-profile-template-sync.md`](features/api-key-profile-template-sync.md) |
| `sidebar-navigation-visibility` | 本地原创 | `active` | `3835c389` | `none` | [`features/sidebar-navigation-visibility.md`](features/sidebar-navigation-visibility.md) |

本索引以“当前仍需维护的有效差异”为核心。已经被上游等价吸收的功能保留归档记录，但不得继续重复应用旧补丁。

## 受控字段

### 功能状态

- `active`：当前是 `ai-slop` 独有差异，需要持续维护。
- `upstream-pending`：本地仍需维护，并且已有或计划建立上游贡献。
- `partially-upstreamed`：上游只吸收了部分能力，本地仅维护明确记录的残余差异。
- `upstreamed`：上游已经吸收，当前不再维护等价本地补丁。
- `retired`：功能已主动移除，只保留历史记录。

### 来源类型

- `local-original`：为本项目自行实现。
- `donor-port`：从无共同 Git 历史的其他项目移植。
- `upstream-derived`：基于 AxonHub 上游功能进行私有扩展。

### 集成方式

- `original`：本地原创实现。
- `merged`：具有共同历史的 AxonHub 上游真实 merge。
- `cherry-picked`：仅在提交历史和代码结构确实兼容时采用。
- `adapted`：保留来源实现思路，并适配本项目结构。
- `reimplemented`：只采用功能或设计思路，在本项目重新实现。

### 数据库影响

- `none`：不涉及 Schema 或现有数据。
- `additive`：只新增兼容结构。
- `data-migration`：需要转换或回填数据，但仍保持兼容。
- `breaking`：无法保持向后兼容，必须先获得用户明确批准。

### 上游关系

- `none`：上游尚无同类实现。
- `equivalent`：上游和私有实现的可观察行为等价。
- `upstream-superset`：上游实现覆盖私有能力并提供更多功能。
- `upstream-subset`：上游只覆盖私有实现的一部分，本地仍有残余能力。
- `diverged`：目标相近但实现方向或行为不兼容，需要显式迁移决策。

## 私有 commit 标记

- 新建的 fork 专属持久差异统一使用 `🧩`，并保留 Conventional Commit 结构，例如 `fix(codex): 🧩 proxy alpha search requests`。
- `🧩` 适用于私有代码、私有文档、上游收敛提交和收敛后保留的残余能力提交。
- 准备贡献上游的纯净 commit 和 `merge(upstream)` 同步节点不使用 `🧩`。
- 未经明确批准，已有提交不为补标记而重写。经批准整体重建时，所有重放的私有提交使用标记，并在功能记录中保留旧新 SHA、上游替代提交和省略原因。

## 经批准的整体历史重建

整体重建不是日常同步方式。只有用户明确批准后才执行，并遵守以下顺序：

1. 为旧的 `ai-slop` head 建立可恢复引用。
2. 以指定的 `upstream/unstable` commit 作为新历史的直接基线，保持全部上游提交 ID 不变。
3. 按功能审计旧分支相对上游的有效净差异；已经被上游等价或超集实现覆盖的提交不再重放。
4. 将仍需维护的私有差异重建为带 `🧩` 的独立提交，尽量保持原作者、作者日期和功能边界。
5. 在功能 YAML 的 `history_rewrites` 与 [`upstream-sync.md`](upstream-sync.md) 中记录旧提交、新提交、上游替代提交及处理理由。
6. 验证指定上游基线是新分支祖先，且新分支相对上游只剩账本登记的有意差异。

## 上游实现收敛

发现上游存在同类实现时，以行为和最终代码为依据，不以 SHA 是否相同为判断标准。最终代码必须采用上游的架构、命名和修改方向，使上游后续补丁能够直接应用或低冲突适配。

收敛默认采用追加提交，不删除已发布历史：

1. `equivalent`：用新的 `🧩` reconciliation commit 抵消私有重复差异，采用上游实现。
2. `upstream-superset`：完整采用上游；必要的调用点、配置或测试迁移放入独立 `🧩` 对齐提交。
3. `upstream-subset`：把重叠部分和私有增量拆开。先抵消重叠部分，再用独立 `🧩` residual commit 只保留上游没有的能力，状态改为 `partially-upstreamed`。
4. `diverged`：停止自动替换，先记录行为差异和迁移方案；涉及降级或不兼容变化时请求用户确认。

只有旧提交能安全、完整反向应用时才直接 `git revert`；否则编写定向 reconciliation commit，避免把仍需保留的私有能力或上游新代码一起删除。每次收敛都在功能 YAML 的 `reconciliations` 中记录：比较时间、上游 commit、关系、动作、原私有 commit、`reverse_commit`、`replacement_commits`、`residual_commits` 和原因。

## 新增或更新功能的流程

1. 确认来源类型并为功能分配稳定的 kebab-case ID。
2. 从模板创建功能记录，填写来源仓库、分支、基线 commit、待采用 commit 和许可证。
3. 对比 `last_checked_commit..目标 commit`，把变化分为采用、适配、重新实现或忽略。
4. 检查 `upstream/unstable` 是否出现同类实现，并记录 `none`、`equivalent`、`upstream-superset`、`upstream-subset` 或 `diverged`。
5. 记录本地落点、与来源的差异、数据库和性能影响，再实施和验证。
6. 将功能代码合入 `ai-slop` 时使用 `🧩` commit 标记并同步更新本记录和索引；如果代码用于上游 PR，私有文档必须留在 `ai-slop` 的独立提交中。
7. 发生上游收敛时追加 reconciliation 记录；不得只修改状态而不记录反向、替换或残余提交。
8. 更新 `last_checked_commit` 和更新日志。即使没有采用任何变化，也要记录检查结果和忽略理由。

## 记录准确性要求

- commit 使用完整 SHA；索引中可使用可辨识的短 SHA。
- `baseline_commit` 表示开始实现或开始追踪时的基线，`last_checked_commit` 表示已经审查到的来源位置，两者不得混用。
- 本地 commit 发生重写时，要同时更新记录，并保留旧 commit 与新 commit 的对应说明。
- `history_rewrites` 专门记录经批准的非追加式历史重建；`reconciliations` 记录功能随上游收敛的行为事件，两者不得互相替代。
- `reconciliations` 是追加式审计记录。字段不适用时写 `null` 或空数组，不得省略已发生的收敛事件。
- 文件清单可以按稳定模块归组，但必须足以让后续 AI 找到实现和测试入口。
- 不得把“来源项目后来新增的代码”默认为适用；必须重新判断架构、许可证、数据库、性能和上游冲突风险。
