---
id: api-key-profile-access-preview
title: API Key 配置文件可用范围预览
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  adopted_commits: []
  last_checked_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  last_checked_at: 2026-08-10
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 457caa1bcd1f5df0ae4dda3e8e22540d940e509c
    - 9a970fd34db8855e030a3e83ec0a968e3b891529
    - 701f33dec724a633f60488160f8d8823b0463764
    - dccc2740bdd1edcaa3eb271896ca0e482b05db6b
  modules:
    - internal/server/biz/api_key_profile_preview.go
    - internal/server/biz/api_key_profile_preview_test.go
    - internal/server/biz/model.go
    - internal/server/gql/axonhub.graphql
    - frontend/src/features/apikeys
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-10
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# API Key 配置文件可用范围预览

## 目的

在 API Key 配置文件管理右侧实时展示当前未保存的生效配置可用的 API 格式和模型 ID。悬停模型 ID 时列出所有实际支持渠道，单渠道同样显示，并按所选 API 格式的预计调用顺序排列。

## 来源与采用范围

本地原创实现，复用 AxonHub 的 API Key、项目配置和 `ListEnabledModels` 可见性算法。

## 本地实现

- 新增只读 GraphQL 预览查询，接收 API Key ID 与未保存配置。
- 服务端复用真实模型列表算法，应用项目上限、渠道 ID、标签模式、模型白名单、系统模型关联和黑名单。
- 预览结果补充实际支持渠道与渠道 API 格式。
- 配置文件对话框新增右侧面板、模型搜索和渠道 Tooltip，表单变化经短暂防抖后实时刷新。
- 没有配置文件的 API Key 使用空限制条件发起真实预览，结果与该密钥的 `/v1/models` 可见模型保持一致。
- 切换 API Key 或配置时不复用上一条预览数据，并在当前密钥详情加载完成后再请求预览，避免显示其他密钥的旧结果。
- 可用 API 格式仅展示对话接口，并改为下划线选项卡；固定依次显示 Chat Completions、Responses、Anthropic Messages、Gemini Contents，隐藏图片、视频等非对话格式。
- 支持渠道先按该 API 格式的系统透传优先级分层，再按渠道权重降序、名称和 ID 稳定排序，并显示调用序号、透传标记和权重。

## 与来源的差异

上游基线只提供配置编辑和模型候选输入，不展示配置保存前后的最终可用范围。

## 上游收敛

2026-08-10 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价预览，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；查询不保存配置，也不修改 Ent Schema 或现有数据。

## 验证

- `go test ./internal/server/biz -run 'TestModelService_(PreviewAPIKeyProfile|ListEnabledModels)$' -count=1`：通过。
- `node --test src/features/apikeys/api-key-profile-preview.test.mjs`：3 个用例通过，覆盖无配置预览、跨密钥旧数据残留，以及对话 API 格式筛选和选项卡渲染。
- `./node_modules/.bin/tsc --noEmit`：通过。
- 排序元数据与 API 格式优先级已纳入 `TestModelService_PreviewAPIKeyProfile`。
- 无限制预览与真实 `ListEnabledModels` 的模型 ID 集合一致性已纳入后端回归测试。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `457caa1bcd1f5df0ae4dda3e8e22540d940e509c` | 新增真实服务端计算的未保存配置实时预览。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `9a970fd34db8855e030a3e83ec0a968e3b891529` | 按所选 API 的透传优先级和渠道权重展示固定调用顺序。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `701f33dec724a633f60488160f8d8823b0463764` | 修复无配置密钥未发起无限制预览及切换密钥时沿用旧预览结果。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `dccc2740bdd1edcaa3eb271896ca0e482b05db6b` | 可用 API 格式只保留四种对话接口，并使用选项卡交互。 |
