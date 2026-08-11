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
| `mobile-ui-improvements` | `none` | 是 | 无 | 仅涉及前端 UI，且已被上游吸收。 |
| `responses-reasoning-item-ids` | `none` | 是 | 无 | 已通过历史重写移除代码，仅保留账本记录。 |
| `channel-api-key-copy` | `none` | 是 | 无 | 仅涉及渠道编辑界面的复制交互。 |
| `request-log-layout` | `none` | 是 | 无 | 仅涉及请求日志查询字段和前端表格布局。 |
| `channel-model-multi-filter` | `none` | 是 | 无 | 新增 GraphQL 查询参数和筛选逻辑，不修改持久化结构。 |
| `filter-state-persistence` | `none` | 是 | 无 | 仅使用浏览器本地存储保存页面筛选状态。 |
| `provider-quota-display` | `none` | 是 | 无 | 复用现有系统键值表中的配额 JSON；新增字段带安全默认值，不修改 Ent Schema。 |
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

截至最后一次合入的上游基线 `9dfd6ac0c21bbc5abe55827fa634e22826287d67`，`ai-slop` 的有效私有差异不包含 Ent Schema 或数据迁移文件。

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
