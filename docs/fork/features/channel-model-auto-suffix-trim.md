---
id: channel-model-auto-suffix-trim
title: 渠道模型自动裁剪后缀
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 49ade6f279eae7aed46858dc121258e922ec9870
  adopted_commits: []
  last_checked_commit: 37e54737c0e147831c6a0b529b79dda1961c3801
  last_checked_at: 2026-08-22
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 8dfb66035b26ee585c67dcdf58f5899b6e5ac8ca
    - b1bb57c18a68e27f922b4f8206d00039469a890f
  modules:
    - internal/objects/channel.go
    - internal/server/biz/channel_llm.go
    - internal/server/biz/channel_model_entry_test.go
    - internal/server/gql/axonhub.graphql
    - internal/server/gql/channel_model_cache_diagnostics.go
    - frontend/src/features/channels/components/channels-model-mapping-dialog.tsx
    - frontend/src/features/channels/data/channels.ts
    - frontend/src/features/channels/data/schema.ts
    - frontend/src/features/channels/utils/merge.ts
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

# 渠道模型自动裁剪后缀

## 目的

让渠道可将支持模型名末尾的后缀作为可选别名裁掉，供同名模型在不同渠道之间进行故障转移。它可与既有的自动裁剪前缀组合：例如渠道实际模型为 `z-ai/glm-5.2:free`、配置前缀 `z-ai` 和后缀 `free` 后，调用方可使用原始模型、`glm-5.2:free`、`z-ai/glm-5.2` 或 `glm-5.2`，而上游始终收到原始模型名。

不做通配符、正则或自动猜测；只有用户在渠道的“模型映射”设置中显式配置的末尾后缀才会裁掉。

## 来源与采用范围

本地原创。实现基于已有 `AutoTrimedModelPrefixes` 的有效模型条目机制扩展，没有引入外部项目代码。

## 本地实现

- `ChannelSettings` 使用可选 `autoTrimedModelSuffixes` JSON 字段保存后缀文本，并增加可空的 `autoTrimedModelSuffixColon`、`autoTrimedModelSuffixHyphen` 开关。手工输入 `:free` 或 `-free` 也会规范为 `free`。
- 有效模型生成先产生原始模型和前缀裁剪别名，再针对每一个变体裁剪末尾后缀。已勾选的分隔符会与后缀一并移除；未勾选时仅移除后缀文本、保留分隔符，例如 `ox-alpha-free` 变成 `ox-alpha-`。直接模型仍优先于自动别名，避免覆盖真实同名模型。
- 旧渠道没有两个新字段时保持历史兼容：冒号裁剪默认开启，短线裁剪默认关闭。新渠道和经此对话框保存的渠道会显式保存两项开关，因此允许用户把两项都关闭。
- GraphQL 的 `ChannelSettings` 和 `ChannelSettingsInput` 同时暴露该字段；创建、复制、批量创建、更新、批量导入及列表查询都回传它，避免编辑时丢失配置。
- 渠道“模型映射”对话框提供后缀标签输入、从支持模型自动提取、清空和数量提示；候选取最后一个冒号或短线之后的片段。界面提供“自动裁剪冒号”和“自动裁剪短线”两个复选框。
- 模型缓存诊断快照带上该配置，便于排查渠道最终可用模型集合；不增加轮询、后台任务或额外外部请求。

## 与来源的差异

上游只支持直接模型、额外前缀、自动裁剪前缀和映射；本功能在同一个有效模型条目抽象中增加受限的末尾后缀别名，以保持路由、模型筛选和故障转移对同一集合达成一致。

## 上游收敛

2026-08-22 比较 `upstream/unstable@37e54737c0e147831c6a0b529b79dda1961c3801` 的渠道设置、模型条目生成、GraphQL schema 与渠道设置界面，未发现自动裁剪后缀或分隔符选择的同类能力，关系为 `none`。

后续若上游增加同类功能，按上游接口和实现方向收敛，并以追加的 `🧩` 提交记录 reconciliation，不重写现有历史。

## 数据库兼容

兼容等级为 `additive`。不修改 Ent 表、列、索引或迁移；仅在既有 `channels.settings` JSON 中加入可省略的 `autoTrimedModelSuffixes`、`autoTrimedModelSuffixColon` 和 `autoTrimedModelSuffixHyphen`。旧渠道缺失后缀列表时等价于关闭；已配置旧后缀但缺失分隔符字段时继续按冒号裁剪，因而不需要回填。

回退到不识别该字段的旧版本前，不应使用旧版本保存已配置后缀的渠道；旧版保存可能重写 settings JSON 并丢失该未知字段。部署到绿色环境前按蓝绿规则对现有 SQLite 数据库创建一致性 `.backup` 并完成 `PRAGMA quick_check`。

## 验证

- `make generate`：通过，GraphQL 生成产物已更新。
- `go test ./internal/server/biz -run 'TestChannel_ChooseModel_AutoTrimmed(PrefixAndSuffix|SuffixDelimiters)$' -count=1`：通过。
- `go test ./internal/server/gql -run '^$' -count=1`：通过。
- `pnpm --dir frontend exec tsc --noEmit`：通过。
- `TestChannel_ChooseModel_AutoTrimmedPrefixAndSuffix` 覆盖 `z-ai/glm-5.2:free` 的四个请求名，并断言全部路由到原始上游模型。
- `TestChannel_ChooseModel_AutoTrimmedSuffixDelimiters` 覆盖冒号、短线、两项均关闭及两项均开启，并断言别名始终路由到原始上游模型。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-22 | `upstream/unstable@49ade6f2` | `8dfb66035b26ee585c67dcdf58f5899b6e5ac8ca` | 本地新增显式后缀裁剪、前后缀组合别名、GraphQL/UI 传递和回归测试。 |
| 2026-08-22 | `upstream/unstable@37e54737` | `b1bb57c18a68e27f922b4f8206d00039469a890f` | 将后缀分隔符拆为冒号、短线两个复选框，保留旧渠道冒号兼容，并覆盖两项均关闭时保留符号的行为。 |
