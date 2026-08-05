---
id: mobile-ui-improvements
title: 移动端 UI 综合优化
status: upstreamed
origin: local-original
integration_method: original
source:
  repository: https://github.com/nb5p/axonhub
  branch: style/improve-mobile-ui
  baseline_commit: 7122f32994d9131e63a4217be3d58d33f187c350
  adopted_commits:
    - 8c51a1123a4eaa2cf14bbea2bf524853681503a4
    - 06ae2922ef138a3eea66c03a98a1e779a56ba332
    - a3a7c8e9ae0a3cea7c0df2b5561580408a2c3d0e
    - 80eb65ad7d4f67c837288bd6636d62fcfde0552d
    - 40ccc3e29a612b09e1e1c0fde7375510fbd0f6af
    - a77ecadcacb56b9f9493feb233084011777ea5ef
    - 8e4cd73d35a5336145a9ab4df5ffade758f84693
    - 0e1af9f860b4e8c0be8157e2a8bb163e9d463661
    - b459d1d18894f745e37b8dcb14a41d72453e9f09
    - 9378b030894cab1222597cc8cd0dddb1b0e586d8
    - 67af2b59875b9d6b6f2112c3e2473d9a9df2f9bb
    - e322b14c5b48fdc8236765365e92289334600f06
    - f411d2f9f9894da7a2c41940330ef5607dbfe555
  last_checked_commit: f411d2f9f9894da7a2c41940330ef5607dbfe555
  last_checked_at: 2026-08-05
  license: project-local
local:
  branch: ai-slop
  commit_marker: null
  commits: []
  modules:
    - frontend/src/features/channels
    - frontend/src/features/requests
    - frontend/src/features/system
    - frontend/src/features/threads
    - frontend/src/features/traces
    - frontend/src/hooks/use-horizontal-scroll.ts
upstream:
  repository: https://github.com/looplj/axonhub
  pull_request: https://github.com/looplj/axonhub/pull/1896
  accepted_commit: 4971cbe436dab12c5d279493a28e0a2affe68309
  relation: equivalent
  last_compared_at: 2026-08-05
reconciliations:
  - compared_at: 2026-08-05
    upstream_commits:
      - 4971cbe436dab12c5d279493a28e0a2affe68309
    relation: equivalent
    action: replaced-with-upstream
    local_commits:
      - b65930ea7a439e79fb84b1a4470a448b36b69621
    reverse_commit: null
    replacement_commits:
      - e55d8e915524facef57087f4eb6952c8cd9dc127
    residual_commits: []
    reason: 上游通过 PR #1896 以 squash commit 接受完整功能；同步后采用上游版本，不再维护重复私有补丁。
history_rewrites:
  - rewritten_at: 2026-08-06
    base_commit: d6ed9c6288ae1642a5a1e8f76db7a8abc0b223af
    old_commits:
      - b65930ea7a439e79fb84b1a4470a448b36b69621
      - e55d8e915524facef57087f4eb6952c8cd9dc127
    new_commits: []
    upstream_commits:
      - 4971cbe436dab12c5d279493a28e0a2affe68309
    disposition: omitted-as-upstream-equivalent
    reason: 新历史直接基于已经包含上游 squash 实现的 d6ed9c62，不再重放本地导入提交或旧同步 merge。
database:
  impact: none
  backward_compatible: true
---

# 移动端 UI 综合优化

## 目的

改善 AxonHub 多个页面在窄屏和触摸设备上的可用性，包括工具栏溢出、渠道类型标签、输入框样式、对话框高度、系统标签页、请求抽屉和统计筛选布局。

## 来源与采用范围

- 这是用户 fork 分支 `style/improve-mobile-ui` 上的 13 个原创提交，基线为 `7122f32994d9131e63a4217be3d58d33f187c350`。
- 本地曾通过 `b65930ea7a439e79fb84b1a4470a448b36b69621` 合入 `ai-slop`。
- 上游 PR [#1896](https://github.com/looplj/axonhub/pull/1896) 已以 squash commit `4971cbe436dab12c5d279493a28e0a2affe68309` 接受全部提交内容。

## 本地实现

- 为移动端工具栏增加横向滚动和鼠标滚轮支持。
- 增加渠道提供商标签的显示开关和状态持久化。
- 调整渠道搜索、输入焦点、Playground、System tabs、Request drawer 和 Token stats 的响应式布局。
- 覆盖 Channels、Requests、System、Threads、Traces 等前端模块。

## 与来源的差异

当前 `ai-slop` 相对最近已合入的上游基线不再保留这批功能的有效代码差异。Git 历史中的原始提交与上游 squash commit patch ID 不同，但不得因此重复应用旧提交。

后续修复应直接跟随 AxonHub 上游对应模块；如果需要新的私有扩展，应建立新的功能 ID，而不是重新激活本记录。

## 上游收敛

- 关系为 `equivalent`：上游 `4971cbe436dab12c5d279493a28e0a2affe68309` 已吸收原 13 个本地提交的功能。
- 2026-08-05 通过同步提交 `e55d8e915524facef57087f4eb6952c8cd9dc127` 采用上游版本，没有额外残余能力，因此未创建独立 reverse commit。
- 2026-08-06 经用户批准整体重建后，新历史直接包含上游实现；旧本地导入提交和同步 merge 均未重放，也没有生成新的私有替代提交。

## 数据库兼容

- 兼容等级：`none`
- 纯前端 UI 变化，不修改 API、Schema 或数据。

## 验证

- 上游接受 commit 包含原分支 13 个提交的完整提交说明和对应前端改动。
- 2026-08-05 对比上游合入基线与 `ai-slop` 后，确认该功能没有需要继续维护的独立有效差异。

## 更新历史

| 日期 | 来源范围 | 本地 commit | 决策与结果 |
|---|---|---|---|
| 2026-07-16 | `7122f329..f411d2f9` | `b65930ea7a439e79fb84b1a4470a448b36b69621` | 将移动端优化导入 `ai-slop`。 |
| 2026-07-25 | PR `#1896` | `4971cbe436dab12c5d279493a28e0a2affe68309` | 上游 squash 合入，状态改为 `upstreamed`。 |
| 2026-08-05 | 上游同步至 `7ed44005` | `e55d8e915524facef57087f4eb6952c8cd9dc127` | 采用上游版本，不再维护等价私有补丁。 |
| 2026-08-06 | 历史重建至上游 `d6ed9c62` | `null` | 直接采用上游 `4971cbe4`，不重放已被覆盖的本地提交。 |
