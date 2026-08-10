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
- API 格式徽标可切换排序上下文；默认依次优先 Chat Completions、Responses、Anthropic Messages、Gemini Contents。
- 支持渠道先按该 API 格式的系统透传优先级分层，再按渠道权重降序、名称和 ID 稳定排序，并显示调用序号、透传标记和权重。

## 与来源的差异

上游基线只提供配置编辑和模型候选输入，不展示配置保存前后的最终可用范围。

## 上游收敛

2026-08-10 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价预览，关系为 `none`。

## 数据库兼容

兼容等级为 `none`；查询不保存配置，也不修改 Ent Schema 或现有数据。

## 验证

- `go test ./internal/server/biz -run 'TestModelService_(PreviewAPIKeyProfile|ListEnabledModels)$' -count=1`：通过。
- GraphQL 包编译测试通过。
- TypeScript 检查通过。
- 排序元数据与 API 格式优先级已纳入 `TestModelService_PreviewAPIKeyProfile`。
- 本地前端开发服务在 Orb 重启后未恢复，页面视觉验证延后至绿色镜像部署。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-10 | `upstream/unstable@9dfd6ac0` | `457caa1bcd1f5df0ae4dda3e8e22540d940e509c` | 新增真实服务端计算的未保存配置实时预览。 |
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `9a970fd34db8855e030a3e83ec0a968e3b891529` | 按所选 API 的透传优先级和渠道权重展示固定调用顺序。 |
