# AGENTS.local.md

本文件是 `ai-slop` 私有分支的长期维护规则，必须与根目录 `AGENTS.md` 一起执行。它不属于准备提交给 AxonHub 上游的内容。

## 每次任务的强制阅读顺序

1. 完整阅读根目录 `AGENTS.md`。
2. 完整阅读本文件。
3. 每次任务都阅读 [`docs/fork/README.md`](docs/fork/README.md)。
4. 如果任务涉及已登记功能、上游同步、外部项目移植或数据库变化，继续阅读对应功能记录、[`docs/fork/upstream-sync.md`](docs/fork/upstream-sync.md) 和 [`docs/fork/database-compatibility.md`](docs/fork/database-compatibility.md)。

## 对话中新规则的固化

- 如果用户在对话中提出与本文件相悖的要求，必须指出具体冲突，并询问该要求是仅限本次的例外，还是要永久写入 `AGENTS.local.md`。冲突尚未澄清前，暂停受冲突影响的修改性操作。
- 如果用户补充了新的项目级要求，包括工作流、分支、架构、数据库、性能、文档或验收规则，必须询问是否将刚才的要求加入 `AGENTS.local.md`。
- 用户未确认永久保存时，只在当前任务中执行该要求，不得擅自修改 `AGENTS.local.md`。
- 用户确认永久保存后，应在同一任务中更新本文件；如果同时影响移植或数据库维护流程，也要更新 `docs/fork/` 中的对应文档。
- 普通的一次性功能细节不自动视为长期规则；不确定时仍应询问用户。

## 分支与来源分类

开始实现前，先把变更归入以下一种类型，并采用对应流程：

1. **AxonHub 上游同步**
   - 上游远程为 `upstream`，目标分支为 `upstream/unstable`。
   - 私有成品分支为 `ai-slop`，通过真实 merge 吸收上游，保留清晰的同步节点。
   - 默认禁止通过 rebase、drop 或强制重建抹掉已发布的私有历史；如果确需整体重建，必须先获得用户明确批准、建立可恢复引用，并在账本中记录旧新提交映射。
   - 经批准整体重建时，以指定的 `upstream/unstable` commit 作为直接基线，保持其提交 ID 和父链不变；只重放相对该基线仍然有效的私有净差异，不重建已经被上游等价吸收的本地提交或历史 merge 外壳。
   - 每次成功同步后更新 `docs/fork/upstream-sync.md`，记录上游 commit、本地 merge commit、冲突和取舍。
2. **准备提交给 AxonHub 上游的贡献**
   - 必须从最新的 `upstream/unstable` 单独创建贡献分支，不能从 `ai-slop` 直接切出。
   - 上游 PR 不得夹带 `AGENTS.local.md`、`docs/fork/` 或其他仅属于私有分支的维护内容。
   - 如果同一功能也要进入 `ai-slop`，先保持可贡献代码提交的纯净，再在 `ai-slop` 中以独立提交更新私有账本。
3. **从其他项目移植的功能**
   - 其他仓库只是功能来源，除非仓库具有可证明的共同 Git 历史，否则不得直接 merge 其分支。
   - 先记录来源仓库、分支、基线 commit、目标 commit 和许可证，再采用 `adapted` 或 `reimplemented` 方式进行语义移植。
   - 在临时集成分支完成实现和验证后再合入 `ai-slop`。
4. **本地原创功能**
   - 记录设计目的、本地开发基线、实现 commit、影响模块、验证方式及是否计划贡献上游。

## 私有提交标识

- `🧩` 是 `ai-slop` 持久私有差异的唯一 commit 标记。新建的 fork 专属代码、账本、上游收敛和残余能力提交必须使用它。
- 保留 Conventional Commit 的类型和作用域，把标记放在冒号后的主题开头，例如 `fix(codex): 🧩 proxy alpha search requests`、`docs(fork): 🧩 update maintenance ledger`。
- 准备提交 AxonHub 上游的纯净贡献 commit 不得使用 `🧩`；上游同步 merge 也不使用该标记，统一采用 `merge(upstream): sync unstable at <short-sha>` 一类主题。
- 未经用户明确批准，不得只为补加 emoji 而重写已有历史。经批准整体重建时，所有重放的私有提交都必须补上 `🧩`，并在功能 YAML 和同步账本中记录旧提交、新提交、上游替代提交及被省略原因。
- `🧩` 只表示“当前提交属于 fork 私有维护面”，不能替代来源、许可证、上游关系和验证记录。

## 上游同类实现的收敛规则

- 检查上游时必须比较可观察行为、接口、测试和最终文件差异，不能只依赖 commit SHA、patch-id 或提交标题。
- 一旦上游出现同类实现，以 `upstream/unstable` 的架构、命名、接口和修改方向为准。私有实现必须向上游写法收敛，避免维持两套等价实现，确保后续上游小补丁可以直接应用或低冲突适配。
- 每次比较把关系记录为 `none`、`equivalent`、`upstream-superset`、`upstream-subset` 或 `diverged`，并在功能 YAML 的 `upstream` 和 `reconciliations` 中记录上游 commit、比较时间与结论。
- 对话中所称的 Rework/Reverse 统一记作一次 `reconciliation`。默认采用追加式历史：创建新的 `🧩` 收敛提交抵消或替换旧私有实现，不通过 rebase/drop 删除已经发布的提交。
- 只有在旧提交能够安全、完整反向应用时才直接使用 `git revert`。如果旧提交还包含上游未覆盖能力，必须编写定向收敛提交，禁止盲目 revert 导致仍需保留的代码或上游新实现被一起删除。
- 上游实现与私有实现等价时，移除私有重复差异并采用上游实现；记录原私有 commit、上游 commit、`reverse_commit` 或 `replacement_commits`、执行时间和原因。
- 上游实现是超集时，完整采用上游实现；如 merge 后仍需迁移配置、调用点或测试，使用独立的 `🧩` 对齐提交，不继续维护旧实现。
- 上游实现是子集时，把重叠部分与私有增量拆开：先用一个 `🧩` 收敛提交抵消重叠实现，再用独立 `🧩` 残余提交只保留上游缺少的能力；在 YAML 中分别记录 `reverse_commit`、`replacement_commits` 和 `residual_commits`。
- 上游与私有实现方向分歧时，先停止自动替换，记录行为和兼容性差异，再按“以上游为准”的原则设计迁移；涉及功能降级、数据不兼容或外部行为变化时必须请求用户确认。
- 收敛完成后必须验证：私有分支相对上游只剩有意保留的差异、上游相关测试/补丁仍能应用、功能记录状态与实际代码一致。

## 功能移植账本是完成条件

- 新增移植功能或本地长期功能时，必须以 [`docs/fork/features/_template.md`](docs/fork/features/_template.md) 为模板创建记录，并加入总索引。
- 修改已登记功能时，必须同步更新其本地 commit、实现差异、数据库影响、验证结果和更新历史。
- 检查来源仓库的后续变化时，从 `last_checked_commit` 比较到目标 commit；无论决定采用、适配还是忽略，都要更新检查基线并记录理由。
- 对于进入 `ai-slop` 的功能，代码、验证和账本更新属于同一个交付，不更新账本不得宣告完成。
- 对于上游贡献分支，私有账本不能进入上游 PR；应在该功能合入或同步到 `ai-slop` 时单独更新账本。
- 来源 commit 只用于定位和理解。不同项目之间默认不能直接 cherry-pick，必须核对架构、数据模型、依赖、许可证和性能影响。

## 数据库兼容是硬门槛

- AxonHub 当前 Ent Schema 和数据迁移体系是唯一事实来源；其他项目的迁移文件只能作为设计参考，不能原样执行。
- 优先采用增量设计：独立功能优先新增表；少量属性优先新增可空字段或具有安全默认值的字段。
- 不得在同一次升级中直接删除或重命名旧表/旧字段、改变旧字段类型，或用不可逆方式覆盖旧数据。
- 需要演进旧结构时采用 expand → migrate → contract：先扩展并保持新旧读写兼容，再迁移数据，确认安全版本边界后才考虑收缩。
- 数据迁移必须可重复执行，并在真实旧版数据库副本上验证；迁移前必须明确备份和恢复步骤。
- 每次 Schema 或数据迁移都要更新 `docs/fork/database-compatibility.md` 和对应功能记录，即使判断为向后兼容。
- 如果确实需要不兼容变更，必须先写明受影响范围、起始版本或 commit、备份、迁移、失败恢复、回滚能力和最低可升级版本，然后停止实施并等待用户明确批准。

## 性能与可维护性

- 功能关闭时不得残留无意义的后台任务、轮询、数据库查询或外部请求。
- 新增常驻任务、热点路径或查询时，必须说明性能影响和验证方式。
- 优先把私有能力集中在稳定接口和独立模块中，避免把同一功能散落到大量上游核心文件，从而降低后续 merge 冲突。
- 移植前必须确认来源许可证允许采用相应代码或设计。

## 交付前检查

- 确认当前分支和任务类型正确。
- 确认相关 `docs/fork/` 记录已更新。
- 如果发生历史重建，确认可恢复引用、上游直接基线、旧新 SHA 映射和被上游替代的提交均已登记。
- 确认数据库兼容等级已经登记；不可兼容变更已经获得明确批准。
- 确认上游贡献提交没有夹带私有规则或账本文档。
- 保留与任务无关的用户改动和 stash，不得擅自恢复、删除或改写。

## 本地蓝绿部署规则

本节适用于本仓库的本地蓝绿部署，不属于准备提交给 AxonHub 上游的内容。

### 适用原则

- 本文件中的“蓝色”“绿色”是部署槽位，不是 Git 分支名。当前打开的仓库、当前分支和当前运行槽位三者没有必然关系。
- 不得根据上一次会话、用户口述、容器名称或健康接口版本单独猜测当前槽位；执行重启或替换容器前必须现场读取权威状态并交叉验证。
- 用户说“推到绿色”默认指把实现部署到绿色测试容器，不等于 `git push`，也不等于向 Forgejo 推送镜像。只有用户明确要求时才执行 Git 或镜像仓库推送。
- 大型功能修复完成验证并提交到 `ai-slop` 后，如果现场确认 `active=blue`，无需再次等待用户提醒：立即把 `ai-slop` 推送到 Forgejo，从准确工作树构建 `axonhub-local:dev`，并按本节流程部署到绿色。大型功能修复包括明显改变核心交互、数据逻辑或跨多个模块的修复；小型文案、样式微调和纯文档修改不触发自动部署。
- 上述自动流程不得绕过绿色部署安全检查，也不得自动切换 Axon Switch 流量。如果 `active=green` 或槽位无法可靠判定，Git 提交可以完成，但停止推送、构建和重建绿色，先请用户切换到蓝色或确认状态。

### 蓝绿环境拓扑

| 项目 | 蓝色正式环境 | 绿色测试环境 |
| --- | --- | --- |
| 用途 | 群晖 NAS 上的稳定回退环境 | Mac 上用于验证新代码的 Docker 环境 |
| 容器 | 群晖上的 `axonhub` | Mac 上的 `axonhub-local` |
| 平台 | `linux/amd64` | `linux/arm64` |
| 直连地址 | 不在本仓库中硬编码 | `http://127.0.0.1:9090` |
| 镜像 | 正式环境镜像 | `axonhub-local:dev` |
| Compose 控制目录 | 群晖项目目录 | `/Users/tux/Playground/AxonHub` |
| Compose 文件 | 由群晖 Web UI 管理 | `/Users/tux/Playground/AxonHub/compose.local.yaml` |
| 数据目录 | 群晖生产数据卷 | `/Users/tux/Playground/AxonHub/data` |

统一公网入口是 `https://axon.109062.xyz:88`，流量由 Axon Switch 在蓝色和绿色之间切换。Axon Switch 源码位于 `/Users/tux/Projects/axon-switch`，其群晖容器名为 `axon-switch`。

当前开发工作树通常是 `/Users/tux/Projects/axonhub`。绿色 Compose 项目目录 `/Users/tux/Playground/AxonHub` 只是绿色容器的控制与数据目录，里面可能是旧源码；它不是当前功能源码的隐式副本。

### 判定当前正在承载流量的槽位

#### 1. 读取权威状态

Axon Switch 挂载到 Mac 的状态文件为：

```text
/Volumes/docker/axon-switch/volumes/axon-switch%data/state.json
```

只允许提取 `backends.active` 字段：

```sh
jq -r '.backends.active' \
  '/Volumes/docker/axon-switch/volumes/axon-switch%data/state.json'
```

结果必须是 `blue` 或 `green`。状态文件还含有密码、密钥和后端信息，禁止 `cat`、完整打印、复制到日志或在回复中展示其内容。

#### 2. 交叉检查健康接口

```sh
curl -fsS --max-time 10 'https://axon.109062.xyz:88/health'
curl -fsS --max-time 10 'http://127.0.0.1:9090/health'
```

比较 `version`、`build.build_time`、`build.platform` 和 `uptime`。当状态为 `green` 时，公网健康信息应与绿色直连信息一致。健康信息只能用于交叉验证，不能取代 `backends.active`。

如果状态文件不可用、字段无效或两项检查互相矛盾，必须报告“当前槽位无法可靠判定”，停止所有会中断服务的操作并请用户确认。不得以猜测继续。

对用户报告时使用明确表述，例如：

> 已现场验证 Axon Switch 的 `active=green`，公网与绿色直连健康信息一致。

不得只说“应该是绿色”或“我记得是绿色”。

### “部署到绿色”的准确含义

#### 构建来源

镜像必须从实现该功能的准确 Git 工作树构建。开始前记录工作树路径、分支、HEAD 和未提交状态。不要因为绿色 Compose 文件位于 Playground，就默认用 Playground 里的旧源码构建。

在实现功能的工作树中可使用：

```sh
AXONHUB_SOURCE_DIR="$(git rev-parse --show-toplevel)"
docker build \
  --tag axonhub-local:dev \
  --file "$AXONHUB_SOURCE_DIR/Dockerfile" \
  "$AXONHUB_SOURCE_DIR"
```

构建镜像本身不会切换流量或重启容器。构建完成后应记录新镜像 ID，以便在重建容器后确认绿色实际使用的是刚构建的镜像。

#### 重建绿色容器

只有完成下节的流量安全检查后，才能执行：

```sh
docker compose \
  --file /Users/tux/Playground/AxonHub/compose.local.yaml \
  up --detach --no-build --force-recreate axonhub-local
```

不要在 Playground 目录直接运行 `docker compose build` 来构建当前项目中的新功能，除非已经明确把准确源码同步到了那里并核对过 HEAD。

### 安全的绿色部署流程

1. 确认实现功能的工作树、分支、HEAD 和 `git status`，保留与任务无关的改动。
2. 按项目规则完成必要验证。根目录 `AGENTS.md` 禁止未经用户要求运行 lint 或 build；用户明确要求部署时，Docker 镜像构建属于部署所需构建。
3. 从准确工作树构建 `axonhub-local:dev`。此时不要重启任何容器。
4. 按“判定当前正在承载流量的槽位”现场读取 Axon Switch 状态并交叉检查：
   - 如果 `active=green`：告诉用户“功能和镜像已准备好，但绿色正在承载流量”，请用户先手工切换到蓝色；停止并等待用户明确回复已经切换。
   - 收到用户“已切换”后，必须再次现场读取状态。只有重新确认 `active=blue` 后才能继续，不能把用户回复本身当作机器状态。
   - 如果 `active=blue`：可以继续重建绿色容器。
   - 如果无法判定：停止，不得重启绿色。
5. 使用固定的绿色 Compose 文件重建 `axonhub-local`。绿色部署不得重启、替换或修改蓝色正式容器。
6. 核对绿色容器状态、实际镜像 ID、`http://127.0.0.1:9090/health` 和相关日志；有数据库变化时还要完成数据库兼容与数据检查。
7. 告诉用户绿色已经就绪，可以手工切换测试。除非用户明确要求，不得自动把 Axon Switch 切回绿色。
8. 切换后如需排障，再次交叉检查公网和绿色直连健康信息，确保请求确实到达绿色，而不是仅凭页面按钮状态判断。

任何可能中断当前活跃槽位的动作之前，都要先告知用户“现在是哪种颜色承载流量、准备重启哪种颜色、是否需要先切换”。

### SQLite 数据安全

- 绿色数据目录固定为 `/Users/tux/Playground/AxonHub/data`，不得因为从另一个工作树构建镜像而更换或覆盖该目录。
- 禁止直接复制正在运行的 SQLite 主文件来刷新绿色数据。应使用 SQLite `.backup` 或等效一致性备份，并对副本执行 `PRAGMA quick_check`。
- 替换数据库前保留可恢复副本，明确源、目标和时间；未经用户明确要求不得删除旧副本。
- Schema 或迁移发生变化时，先确认新版本能打开现有绿色数据库并保留回滚路径，再允许用户切到绿色。

### 部署安全与交付沟通

- 不得在回复、日志或提交中泄露 API Key、登录信息、TOTP 秘钥、Axon Switch 状态文件内容或 Compose 中的 secret。
- 实现或部署功能后，必须说明功能所在工作树、分支和 commit/未提交状态。
- 必须说明是否已构建绿色镜像，以及镜像 ID 是否与容器一致。
- 必须说明当前槽位是现场验证的 `blue`、`green` 还是无法可靠判定，以及下一步会重启哪个容器。
- 必须说明绿色直连健康检查和关键日志是否正常。
- 必须说明 Git、Forgejo 或镜像仓库是否发生推送；没有推送也要明确说明。
