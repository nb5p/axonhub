---
id: sidebar-navigation-visibility
title: 侧边栏导航项目可见性配置
status: active
origin: local-original
integration_method: original
source:
  repository: https://github.com/looplj/axonhub
  branch: unstable
  baseline_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  adopted_commits: []
  last_checked_commit: 9dfd6ac0c21bbc5abe55827fa634e22826287d67
  last_checked_at: 2026-08-11
  license: Apache-2.0
local:
  branch: ai-slop
  commit_marker: "🧩"
  commits:
    - 3835c38972666c61ff7bc95f863cead06fa5fca6
  modules:
    - internal/server/biz/system.go
    - internal/server/gql/system.graphql
    - frontend/src/sidebar-navigation.ts
    - frontend/src/sidebar.ts
    - frontend/src/features/system
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: null
  accepted_commit: null
  relation: none
  last_compared_at: 2026-08-11
reconciliations: []
history_rewrites: []
database:
  impact: none
  backward_compatible: true
---

# 侧边栏导航项目可见性配置

## 目的

允许单人部署或功能精简场景在系统设置中隐藏不需要的侧边栏入口，例如系统级用户、角色、项目，以及项目级线程。每个现有导航项目均可独立隐藏，也允许隐藏全部项目。

## 来源与采用范围

本地原创实现。上游基线只按用户权限过滤导航项目，没有面向实例管理员的全局显示开关。

## 本地实现

- 常规设置新增“侧边栏显示”卡片，按管理、项目和设置分组展示全部导航项目。
- 每个项目使用独立显示开关并立即保存；保存成功后同步刷新侧边栏和快捷命令。
- 使用稳定菜单 ID 保存隐藏列表，系统级和项目级的同名用户、角色入口可以分别控制。
- 所有用户均可读取该显示配置；只有具有系统设置写权限的用户可以修改。
- 隐藏只改变导航展示，不修改路由权限，隐藏“系统”后仍可通过 `/system` 直接恢复配置。

## 与来源的差异

上游基线没有持久化的全局侧边栏可见性配置。

## 上游收敛

2026-08-11 比较 `upstream/unstable@9dfd6ac0c21bbc5abe55827fa634e22826287d67`，未发现等价实现，关系为 `none`。

## 数据库兼容

兼容等级为 `none`。配置使用现有系统键值表的新键 `system_sidebar_navigation_settings` 保存，不修改 Ent Schema；旧版本会忽略该键，未配置时默认显示全部导航项目。

## 验证

- `go test ./internal/server/biz -run TestSystemService_SidebarNavigationSettings -count=1`：通过。
- `go test ./internal/server/gql -run '^$' -count=1`：GraphQL 生成代码编译通过。
- `node --test src/sidebar-navigation.test.mjs`：3 个用例通过，覆盖全部 18 个项目、单项隐藏和全部隐藏。
- `./node_modules/.bin/tsc --noEmit`：通过。
- 中英文 locale JSON 解析通过。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-08-11 | `upstream/unstable@9dfd6ac0` | `3835c38972666c61ff7bc95f863cead06fa5fca6` | 新增全局侧边栏项目显示开关，并保持权限与直达页面行为不变。 |
