---
id: list-pagination-settings
title: 管理列表分页开关
status: active
origin: local-original
integration_method: original
source:
  repository: ssh://git@forgejo.109062.xyz:88/tux/axonhub.git
  branch: ai-slop
  baseline_commit: 2149b93a51f5fd0d586ffdb1a264786c9795a635
  adopted_commits:
    - 621c1de32ffdf9c08a1a85279976b897dd9fbd06
  last_checked_commit: 621c1de32ffdf9c08a1a85279976b897dd9fbd06
  last_checked_at: 2026-08-12
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 621c1de32ffdf9c08a1a85279976b897dd9fbd06
  modules:
    - frontend/src/features/apikeys
    - frontend/src/features/channels
    - frontend/src/features/requests
    - frontend/src/features/system
    - frontend/src/gql
    - internal/server/biz
    - internal/server/gql
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-12
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 管理列表分页开关

## 目的

允许管理员分别控制渠道、API 密钥和请求日志是否分页。渠道与 API 密钥默认一次显示全部数据，请求日志默认继续分页；关闭分页时页面不再显示底部分页器，并把可用空间留给数据表格。

## 来源与采用范围

- 本地原创功能，不移植外部代码。
- 仅覆盖渠道、API 密钥和请求日志三处主列表。
- 2026-08-12 检查 `upstream/unstable@c8de8cf8dac686b4ad71a562e4fffb26df226cef`，未发现同类全局分页设置。

## 本地实现

- 复用现有系统键值存储保存三个独立布尔值，不新增 Ent Schema；默认值为渠道关闭、API 密钥关闭、请求日志开启。
- 系统设置“常规”页提供三个开关。读取属于展示配置，对已登录用户使用系统读取旁路；写入仍受系统设置权限保护。
- 关闭分页后，前端按游标顺序以每批最多 1000 条连续读取全部连接页，并合并为一个连接结果；开启分页时保持原有 URL 游标和每页数量行为。
- 关闭功能不会建立新的后台任务。渠道原有 5 秒状态轮询仍然存在；绝大多数少于 1000 个渠道的部署仍为单次请求，超过 1000 个渠道时每轮会按页顺序请求。
- 前端对尚未升级的旧后端提供默认值回退，避免前后端滚动升级期间因缺少 GraphQL 字段而破坏列表页。

## 与来源的差异

无外部来源。实现集中在独立分页设置卡片和共享连接分页读取器中，三个列表页只负责选择分页或全量查询模式。

## 上游收敛

- 当前关系为 `none`。
- 上游若提供等价分页偏好，应采用上游配置接口和查询方式，并移除本地重复系统键与 GraphQL 字段。

## 数据库兼容

- 兼容等级：`none`。
- 只在现有系统键值表写入新的独立 JSON 配置键；不修改 Schema、索引或已有记录，旧版本会忽略该键。

## 验证

- `go test ./internal/server/biz -run TestSystemService_ListPaginationSettings -count=1`：通过，覆盖默认值与写后读取。
- `go test ./internal/server/gql -run TestValidatePaginationArgs -count=1`：通过，确认批量大小上限为 1000。
- `frontend/node_modules/.bin/tsc --noEmit`：通过。
- Chrome DevTools 390×844：渠道和 API 密钥分页器隐藏并加载全部当前数据，请求日志保留 20 条分页及上一页/下一页操作。
- Chrome DevTools 1440×900：渠道桌面标题与操作栏正常，页面无全局横向溢出。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-12 | 本地原创 | `621c1de32ffdf9c08a1a85279976b897dd9fbd06` | original：增加三处列表分页开关、全量游标读取和默认策略。 |
