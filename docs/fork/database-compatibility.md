# 数据库兼容性契约

数据库兼容是 `ai-slop` 接受新功能的硬门槛。AxonHub 当前 Ent Schema、项目自身迁移机制和已经部署的数据是唯一事实来源；其他项目的表结构和迁移只能作为参考。

## 兼容等级

| 等级 | 含义 | 默认处理 |
|---|---|---|
| `none` | 不涉及 Schema、索引或现有数据 | 正常实现并登记 |
| `additive` | 新增表、索引、可空字段或安全默认字段 | 提供旧数据库升级验证后允许 |
| `data-migration` | 需要回填或转换数据，但新旧版本可以过渡共存 | 使用幂等迁移和 expand → migrate → contract |
| `breaking` | 删除、重命名、改类型或无法保持旧数据可用 | 完成方案后停止，等待用户明确批准 |

## 设计规则

- 独立功能优先新增表，避免改变既有数据的含义。
- 少量新属性可以添加可空字段或具有安全默认值的字段，不为形式上的隔离制造无必要的新表。
- 不得在同一次升级中直接删除或重命名旧结构、改变旧字段类型或不可逆覆盖旧数据。
- 需要替换旧结构时，先增加新结构并兼容新旧读写，再幂等迁移数据，最后在明确的安全版本边界后考虑移除旧结构。
- 迁移前必须给出备份步骤；迁移失败时必须能够根据文档恢复。
- 必须使用真实旧版数据库副本验证升级路径，并覆盖项目实际支持且受本次变化影响的数据库类型。
- 来源项目的迁移文件不得直接复制执行，必须重新映射到本项目的 Ent Schema 和迁移机制。

## 不兼容变更的批准材料

任何 `breaking` 变更在实现前必须记录并交由用户确认：

- 关联功能 ID 和来源 commit；
- 从哪个本地版本或 commit 开始不兼容；
- 受影响的表、字段、索引、数据和数据库类型；
- 升级前备份命令或操作步骤；
- 迁移步骤、预计耗时和空间需求；
- 失败恢复和回滚步骤，或明确说明这是单向迁移；
- 最低可直接升级的旧版本；
- 用户批准记录和批准日期。

文档完成不等于获得批准。AI 必须停止相关实施，直到用户明确同意该不兼容方案。

## 当前私有功能登记

| 功能 ID | 影响 | 向后兼容 | Schema 或数据变化 | 说明 |
|---|---|---|---|---|
| `codex-alpha-search` | `none` | 是 | 无 | 仅涉及请求路由和 LLM 转换。 |
| `mobile-request-route-tooltip` | `none` | 是 | 无 | 仅涉及前端交互。 |
| `mobile-touch-hover-overlays` | `none` | 是 | 无 | 仅统一前端悬浮层的触摸、手写笔、鼠标和键盘交互。 |
| `mobile-ui-improvements` | `none` | 是 | 无 | 仅涉及前端 UI，且已被上游吸收。 |
| `responses-reasoning-item-ids` | `none` | 是 | 无 | 已通过历史重写移除代码，仅保留账本记录。 |
| `codex-remote-compaction-affinity` | `none` | 是 | 无 | 仅复用既有 trace、历史渠道和渠道 API Key 粘性选择，不修改持久化结构。 |
| `channel-api-key-copy` | `none` | 是 | 无 | 仅涉及渠道编辑界面的复制交互。 |
| `request-log-layout` | `none` | 是 | 无 | 仅涉及请求日志查询字段和前端表格布局。 |
| `channel-model-multi-filter` | `none` | 是 | 无 | 新增 GraphQL 查询参数和筛选逻辑，不修改持久化结构。 |
| `filter-state-persistence` | `none` | 是 | 无 | 仅使用浏览器本地存储保存页面筛选状态。 |
| `provider-quota-display` | `none` | 是 | 无 | 复用现有系统键值表中的配额 JSON；新增字段带安全默认值，不修改 Ent Schema。 |
| `channel-usage-query` | `additive` | 是 | 扩展 `provider_quota_status.provider_type` 枚举；在渠道 settings/credentials JSON 中增加可选字段 | 旧数据无需回填；`usageQuery.showInProviderQuota` 缺失时默认显示；回退后用旧版本编辑已配置渠道可能丢失未知 JSON 字段。 |
| `channel-429-non-retryable` | `additive` | 是 | 在 `channels.settings` JSON 中增加可选 `treat429AsNonRetryable` 布尔字段 | 旧数据无需回填；开关控制重试与冷却，错误策略仍由系统设置决定；回退后用旧版本编辑已配置渠道可能丢失未知 JSON 字段。 |
| `channel-endpoint-summary-column` | `none` | 是 | 无 | 仅新增渠道列表端点摘要展示。 |
| `channel-endpoint-filter` | `none` | 是 | 无 | 新增 GraphQL 查询参数和运行时最终端点筛选，不修改持久化结构。 |
| `channel-provider-tabs-default-hidden` | `none` | 是 | 无 | 仅调整浏览器本地偏好的默认值。 |
| `prefer-pass-through-routing` | `none` | 是 | 无 | 使用现有系统键值表保存独立布尔配置，不修改 Ent Schema；旧版本会忽略新增键。 |
| `api-key-profile-access-preview` | `none` | 是 | 无 | 新增只读 GraphQL 预览和排序元数据，不修改配置或持久化结构。 |
| `api-key-profile-clear` | `none` | 是 | 无 | 将现有 API Key 配置 JSON 保存为空集合，不修改 Ent Schema。 |
| `api-key-profile-template-sync` | `none` | 是 | 无 | 在现有 Profile JSON 中增加可省略布尔字段，旧数据缺省为关闭，不修改 Ent Schema。 |
| `sidebar-navigation-visibility` | `none` | 是 | 无 | 使用现有系统键值表保存独立隐藏列表，不修改 Ent Schema；旧版本会忽略新增键。 |
| `api-key-activity-heatmap` | `none` | 是 | 无 | 仅新增只读 GraphQL 聚合和仪表盘热力图，不修改 Ent Schema 或现有数据。 |
| `playground-api-key-testing` | `none` | 是 | 无 | 复用现有 API Key 鉴权、配置和路由上下文，不修改持久化结构。 |
| `table-column-visibility` | `none` | 是 | 无 | 仅调整前端列定义、翻译和浏览器本地偏好。 |
| `list-pagination-settings` | `none` | 是 | 无 | 复用现有系统键值表保存独立 JSON 配置，不修改 Ent Schema；旧版本会忽略新增键。 |
| `channel-model-test-api-formats` | `none` | 是 | 无 | 仅扩展 GraphQL 测试输入和运行时请求构造，不修改渠道配置或 Ent Schema。 |
| `test-request-list-visibility` | `none` | 是 | 无 | 使用现有系统键值表保存独立 JSON 配置；旧版本忽略该键，缺失时默认关闭。 |

截至上游比较基线 `9fb6f1af148d3d3cf7c4053159e5a55a44dbb4ca`，`channel-usage-query` 是当前有效私有差异中登记的增量 Ent Schema 变化；不需要数据回填。

### 2026-08-20 — channel-usage-query

- 兼容等级：`additive`
- 本地 commit：`089da0d1df9757294abeb24f304d5fda9c9adced`、`466f2019`。
- 影响结构：`provider_quota_status.provider_type` 增加 `usage_query`；`channels.settings` 和 `channels.credentials` 的既有 JSON 列增加可选字段；本次新增可选 `usageQuery.showInProviderQuota`，缺失时按 `true` 读取。
- 旧数据库验证样本：本次未连接部署数据库；内存 SQLite 上的 Ent、业务层和配额持久化测试通过。部署前仍需在真实旧版数据库副本上执行自动迁移、完整启动和数据校验。
- 备份与恢复：本次未创建备份。部署前使用对应数据库的一致性快照流程；本地 SQLite 按蓝绿规则使用 `.backup` 并执行 `PRAGMA quick_check`。
- 回滚能力：核心旧数据兼容；如需保留新增脚本配置，回滚旧版本前恢复升级前快照。旧版本编辑已配置渠道可能丢失其无法识别的 JSON 字段。
- 最低升级版本：`466f2019`。
- 用户批准（仅 breaking）：不适用。

### 2026-08-20 — channel-429-non-retryable

- 兼容等级：`additive`
- 本地 commit：`621b04052ab69434df631119de7e604635389da0`、`e325b77369b1d1e5bd81a2ce14572367efe76f63`、`4896e8b908aab60662eb0cb32d74749e775dd8ed`。
- 影响结构：`channels.settings` JSON 增加可选布尔字段 `treat429AsNonRetryable`；不修改 Ent Schema、表、索引或既有数据。
- 旧数据库验证样本：本次未连接部署数据库；已有渠道设置缺失该字段时默认 `false`，保持现有 429 重试、切换渠道和 Retry-After 冷却行为。启用该字段跳过重试并抑制 `Retry-After` 冷却，但不改变系统上游错误策略。自动迁移不需要回填。
- 备份与恢复：本次未创建备份。部署前按既有数据库的一致性快照流程执行；本地 SQLite 使用 `.backup` 并执行 `PRAGMA quick_check`。
- 回滚能力：回退旧代码不影响核心渠道数据，但旧版本编辑已开启该开关的渠道可能重写并丢弃未知 JSON 字段；如需保留该配置，回滚前恢复升级前快照或避免使用旧版本编辑该渠道。
- 最低升级版本：`621b04052ab69434df631119de7e604635389da0`。
- 用户批准（仅 breaking）：不适用。

### 2026-08-21 — channel-usage-query-v2

- 兼容等级：`additive`。
- 本地 commit：`f71cae9869c3f1e4aeed25593f555bf20b460a50`。
- 影响结构：不新增表、列、索引或 Ent Schema。渠道既有 `settings.usageQuery` JSON 增加预设值，`provider_quota_status.quota_data` 由旧平铺字段逐步写为 `balance`、`text`、`progress.windows`；原有 `usage_query` provider type 不变。
- 旧数据库验证样本：内存 SQLite 的业务层、预设、轮询和 GraphQL 测试通过；前端对已保存的旧 JSON 保留只读展示兼容，直到下次成功轮询写入 v2。
- 备份与恢复：部署到绿色 SQLite 前执行 `.backup` 和 `PRAGMA quick_check`，备份路径记录到 Obsidian 备份日志。若需回退，恢复该快照；不需要数据回填。
- 回滚能力：旧版本不会识别 v2 结果字段，且编辑渠道时可能重写并丢弃未知 JSON；回退前恢复快照即可完整还原。旧版采集开关 `claudecode`、`codex`、`opencode_go` 已不再生效，三种渠道统一由 `usage_query` 开关控制。
- 最低升级版本：`f71cae9869c3f1e4aeed25593f555bf20b460a50`。
- 用户批准（仅 breaking）：不适用。

### 2026-08-21 — channel-usage-query-oauth-gate

- 兼容等级：`none`
- 本地 commit：`f2c91bef0f0c50a4b2db141e97736e637cb5faa4`。
- 影响结构：无；只改变未保存内置查询的自动适用条件。已存在的 `channels.settings.usageQuery` JSON 保持优先级，不读写或迁移现有数据。
- 旧数据库验证样本：普通 Codex／Claude API Key 渠道不再创建 `usage_query` 配额状态；OAuth 渠道和 OpenCode Go 保持原有轮询行为。
- 备份与恢复：本次未创建备份；不需要数据库迁移。
- 回滚能力：回退代码仅恢复自动轮询，不改变数据库内容。
- 最低升级版本：`f2c91bef0f0c50a4b2db141e97736e637cb5faa4`。
- 用户批准（仅 breaking）：不适用。

### 2026-08-21 — tuzi-daycard-three-window-script

- 兼容等级：`none`
- 本地 commit：`4593cf61d9aae4b02e11106c421053b97a2378c6`。
- 影响结构：无；仅更新绿色环境 41 号渠道既有 `channels.settings.usageQuery.script` 文本，不新增或变更 Schema。
- 旧数据库验证样本：更新前后均对绿色 `axonhub.db` 执行 `PRAGMA quick_check`，结果为 `ok`；脚本夹具验证日、周、月三窗口均能从真实响应形状提取。
- 备份与恢复：更新前 `.backup` 已保存至 `/Users/tux/Playground/AxonHub/backups/usage-query-41/axonhub-20260821T145539Z-pre-three-window-no-balance.db`，SHA-256 为 `e515baf0eee500281351d3a669a855a156e5eb5e58af58385ea56fb8c9033266`，已登记 Obsidian 备份日志。
- 回滚能力：可用该一致性备份恢复绿色数据库，或仅恢复 `settings.usageQuery.script`；本次不写入凭据。
- 最低升级版本：`4593cf61d9aae4b02e11106c421053b97a2378c6`。
- 用户批准（仅 breaking）：不适用。

### 2026-08-21 — channel-usage-query-tags

- 兼容等级：`additive`
- 本地 commit：`fa2fe5a8febdf45db14b15bae932adfe0c6cbd40`。
- 影响结构：不新增 Schema；既有 `provider_quota_status.quota_data` JSON 增加可选 `tags` 和 `showCodexUsage` 字段，旧缓存缺失字段时安全按未显示处理。
- 旧数据库验证样本：Goja 运行时、预设和配额规范化测试覆盖 tags；绿色 41 号脚本已更新为 tags、三窗口、无 A$，并执行 `PRAGMA quick_check`，结果为 `ok`。
- 备份与恢复：沿用本次 41 号更新前的一致性备份 `/Users/tux/Playground/AxonHub/backups/usage-query-41/axonhub-20260821T145539Z-pre-three-window-no-balance.db`；本项未额外创建数据库备份。
- 回滚能力：旧代码忽略新增缓存字段；若需完整还原 41 号脚本，恢复上述备份。
- 最低升级版本：`fa2fe5a8febdf45db14b15bae932adfe0c6cbd40`。
- 用户批准（仅 breaking）：不适用。

### 2026-08-21 — channel-model-test-api-formats

- 兼容等级：`none`
- 本地 commit：`a683e9122d639c5a1273cc28d509d7019f2a6c35`、本次提交（2026-08-21，多格式测试）。
- 影响结构：无；新增 `GEMINI_CONTENTS` GraphQL 测试枚举，扩展运行时请求格式和前端临时测试结果；不新增或修改持久化字段。
- 旧数据库验证样本：不适用；不读取或写入新的持久化字段。
- 备份与恢复：本次未创建备份；不需要数据库迁移。
- 回滚能力：回退代码不会影响任何数据库结构或数据。
- 最低升级版本：本次提交（2026-08-21，多格式测试）。
- 用户批准（仅 breaking）：不适用。

### 2026-08-21 — test-request-list-visibility

- 兼容等级：`none`
- 本地 commit：`b570bc11a3daf067a0b7f6c09e97fb5b6d3cfd41`。
- 影响结构：现有系统键值表新增独立 JSON 键 `system_test_request_list_settings`；不修改 Ent Schema、表、索引或既有请求数据。
- 旧数据库验证样本：缺失键返回默认关闭；业务层默认值与写后读取测试通过。
- 备份与恢复：本次未创建备份；不需要数据库迁移。
- 回滚能力：旧代码会忽略新增键，请求记录不被修改或复制。
- 最低升级版本：`b570bc11a3daf067a0b7f6c09e97fb5b6d3cfd41`。
- 用户批准（仅 breaking）：不适用。

## 后续登记模板

```markdown
### YYYY-MM-DD — feature-id

- 兼容等级：none | additive | data-migration | breaking
- 本地 commit：
- 影响结构：
- 旧数据库验证样本：
- 备份与恢复：
- 回滚能力：
- 最低升级版本：
- 用户批准（仅 breaking）：
```
