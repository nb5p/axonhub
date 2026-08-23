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

#### 1a. `/Volumes/docker` 缺失时恢复群晖挂载

`/Volumes/docker` 不存在、不是目录或不可读，不表示当前流量在蓝色或绿色；它只表示无法读取 Axon Switch 的权威状态。此时不得重建绿色、停止容器、切换流量或进行其他可能中断服务的操作。

本机群晖 SMB 服务名为 `Synology.local`，共享名为 `docker`。优先通过 macOS 的 Finder 挂载入口恢复固定路径：

```sh
open 'smb://Synology.local/docker'
```

该命令会由 Finder 创建 `/Volumes/docker`，并在需要时显示系统认证弹窗或使用钥匙串中已有凭据。密码、用户名和令牌不得放入命令行、脚本、文档、终端历史、日志或聊天中。不要先执行 `mkdir -p /Volumes/docker` 再直接调用 `mount_smbfs`：macOS 通常不允许普通用户手工创建 `/Volumes` 下的挂载点，Finder 方式可避免这一权限问题。

挂载完成后，只做以下最小验证，再恢复读取状态的流程：

```sh
test -r '/Volumes/docker/axon-switch/volumes/axon-switch%data/state.json'
jq -r '.backends.active' \
  '/Volumes/docker/axon-switch/volumes/axon-switch%data/state.json'
```

若 Finder 未能挂载、`Synology.local` 无法解析，或用户没有该共享的访问权限，报告“当前槽位无法可靠判定”并请求用户在 Finder 中重新连接或提供已获授权的 SMB 主机与共享名；不得猜测地址、共享名或凭据，也不得绕过认证。

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

### 将绿色完整晋升到蓝色

本流程用于把绿色当前的完整 SQLite 数据和指定 `ai-slop` 代码版本一起晋升到群晖蓝色。它是单用户环境下允许短暂停服的完整覆盖流程，不是双向同步。执行后蓝色成为新的唯一数据事实来源，后续写入不会自动回流到绿色。

#### 固定路径和产物边界

| 项目 | Mac 路径 | NAS 路径 |
| --- | --- | --- |
| 绿色数据库 | `/Users/tux/Playground/AxonHub/data/axonhub.db` | 不适用 |
| 绿色晋升快照 | `/Users/tux/Playground/AxonHub/backups/green-to-blue/<UTC时间戳>/axonhub.db` | 不适用 |
| 蓝色项目 | `/Volumes/docker/axonhub` | `/volume1/docker/axonhub` |
| 蓝色数据库 | `/Volumes/docker/axonhub/volumes/axonhub%data/axonhub.db` | `/volume1/docker/axonhub/volumes/axonhub%data/axonhub.db` |
| 蓝色回滚备份 | `/Volumes/docker/axonhub/backups/green-to-blue/<UTC时间戳>` | `/volume1/docker/axonhub/backups/green-to-blue/<UTC时间戳>` |

- Docker 镜像只包含代码和前端产物，SQLite 数据库必须单独备份和传输，禁止打进镜像。
- 蓝色必须使用从准确 commit 构建的不可变镜像标签，例如 `forgejo.109062.xyz:88/tux/axonhub:ai-slop-<short-sha>`。
- 蓝色平台固定为 `linux/amd64`；绿色平台为 `linux/arm64`，不得把绿色本地镜像直接导入群晖。
- 所有备份名使用同一个 UTC 时间戳。旧数据库、旧 Compose、`-wal` 和 `-shm` 文件均保留，不得覆盖或删除旧备份。

开始晋升时一次性生成并记录本次变量，后续所有镜像、备份和 incoming 文件都复用同一组值，禁止中途重新生成时间戳或改用另一个 HEAD：

```sh
PROMOTION_TS="$(date -u +%Y%m%dT%H%M%SZ)"
TARGET_COMMIT="$(git rev-parse HEAD)"
TARGET_SHORT="$(git rev-parse --short=12 "$TARGET_COMMIT")"
TARGET_IMAGE="forgejo.109062.xyz:88/tux/axonhub:ai-slop-$TARGET_SHORT"
```

每个新的 shell、SSH 会话或工具调用都可能丢失这些变量。后续命令必须在同一 shell 块中运行、重新赋入上面已经记录的实际值，或直接使用已记录的完整字面值；不得悄悄重新计算 HEAD 或时间戳。

#### Registry 认证边界和错误判读

- Git 推送与容器镜像推送是两套独立认证。`forgejo` Git remote 使用 SSH 成功，只能证明代码仓库可写，不能证明 Docker 已登录 `forgejo.109062.xyz:88`。
- Mac 是镜像发布端，需要对 package owner `tux` 具有 `write:package` 权限的 Forgejo Token；使用 `docker login forgejo.109062.xyz:88 --username <Forgejo用户名> --password-stdin` 登录，禁止把 Token 放进命令行参数、文档、聊天或日志。Mac Docker Desktop 通常把凭据保存在系统钥匙串中。
- NAS 是镜像运行端，只需 `read:package` 权限。NAS 上的 Docker 命令通过 `sudo` 执行，因此实际使用的是 root 的 Docker 登录态；Mac 的登录态和 NAS 普通用户的登录态都不会自动传给它。禁止打印 `/root/.docker/config.json`，只允许检查目标 registry 条目是否存在，或用 `docker manifest inspect` 做只读权限验证。
- `docker push` 返回 `401 Unauthorized`：优先检查发布端是否登录、Token 是否过期，以及是否有 `write:package` 权限。不要因为 `git push` 成功就排除认证问题。
- 已认证的 `docker manifest inspect` 返回 `manifest unknown`：通常表示认证已通过，但该标签尚未发布；它与 `401` 不是同一问题。
- BuildKit 推送的 OCI index 可能额外包含 `unknown/unknown` provenance/attestation，这是元数据，不是可运行镜像。验收时必须明确找到 `linux/amd64` 子 manifest。

#### 阶段一：不停服准备蓝色镜像

1. 记录准确源码工作树、分支、HEAD 和未提交状态；构建时使用该 commit 的干净临时工作树，避免把用户未提交文件带入构建上下文。
2. 先把目标 commit 推送到 Forgejo。随后分别预检两端认证：Mac 必须具备 Registry 写权限，NAS root 必须能对一个已知存在的私有镜像执行 `docker manifest inspect`。任何预检都不得输出 Token 或完整 Docker 配置。
3. 从干净临时工作树构建并推送不可变标签：

   ```sh
   docker buildx build \
     --platform linux/amd64 \
     --tag "$TARGET_IMAGE" \
     --push \
     --file /absolute/path/to/clean-worktree/Dockerfile \
     /absolute/path/to/clean-worktree
   ```

4. 检查 Registry manifest，记录 OCI index digest，并从详细 manifest 中取出 `linux/amd64` 子 manifest digest 和 config digest：

   ```sh
   docker manifest inspect --verbose "$TARGET_IMAGE" | jq -r '
     (if type == "array" then .[] else . end)
     | select(.Descriptor.platform.os == "linux"
       and .Descriptor.platform.architecture == "amd64")
     | [.Descriptor.digest, .OCIManifest.config.digest]
     | @tsv'
   ```

   必须且只能找到一个 `linux/amd64` 结果。Mac 上 `docker image inspect .Id` 可能显示 OCI index digest，而 NAS 运行容器的 `.Image` 是平台子镜像的 config digest；两者层级不同，不能直接判断为镜像不一致。NAS 拉取或载入后，其镜像 ID 必须等于上面记录的 `linux/amd64` config digest。
5. 在绿色仍运行时把新镜像预加载到 NAS。优先让 NAS 使用只读 package Token 拉取，再核对 `.Os=linux`、`.Architecture=amd64` 和镜像 ID。
6. 只有 Registry 暂时无法使用且用户接受应急回退时，才允许在 Mac 构建/拉取 `linux/amd64` 镜像后 `docker save`，通过 `/Volumes/docker/axonhub` 挂载目录传递 tar，再在 NAS 上执行 `docker load`。此时必须：
   - 记录“Registry 发布待补齐”，不能把“NAS 本地已有镜像”误报成“镜像已推送”；
   - 加载后核对远端平台和镜像 ID，并删除临时 tar；
   - 启动蓝色时使用 `--pull never`，避免 Compose 因远端标签不存在而失败或换用其他镜像；
   - Registry 写权限恢复后补推同一镜像，重新验证 manifest 和 config digest。补推本身不需要重启已经运行的蓝色容器。

   离线回退的机械步骤如下。若此前 `--push` 失败且本地没有目标镜像，使用同一个干净工作树重新执行 `--load`；后续补推这一个本地镜像，禁止再构建第三份同标签镜像：

   ```sh
   # Mac shell：使用本流程开始时记录的变量。
   TRANSFER_TAR="/Volumes/docker/axonhub/$TARGET_SHORT-linux-amd64.tar"

   docker buildx build \
     --platform linux/amd64 \
     --load \
     --tag "$TARGET_IMAGE" \
     --file /absolute/path/to/clean-worktree/Dockerfile \
     /absolute/path/to/clean-worktree
   docker save --output "$TRANSFER_TAR" "$TARGET_IMAGE"
   ```

   NAS shell 不会继承 Mac shell 变量，必须把本次已经记录的实际值显式带入，禁止在 NAS 上重新读取另一个 Git HEAD：

   ```sh
   # NAS shell：把占位符替换为本次记录的实际值。
   TARGET_SHORT='<本次 TARGET_SHORT>'
   TARGET_IMAGE='forgejo.109062.xyz:88/tux/axonhub:ai-slop-<本次 TARGET_SHORT>'

   sudo /var/packages/ContainerManager/target/usr/bin/docker load \
     --input "/volume1/docker/axonhub/$TARGET_SHORT-linux-amd64.tar"
   sudo /var/packages/ContainerManager/target/usr/bin/docker image inspect \
     --format 'id={{.Id}} os={{.Os}} arch={{.Architecture}}' \
     "$TARGET_IMAGE"
   ```

   只有 NAS 载入成功且平台、镜像 ID 校验完成后，才删除这个明确路径的临时 tar。写权限恢复后在 Mac 执行 `docker push "$TARGET_IMAGE"`，再用阶段一第 4 步确认 Registry 的 `linux/amd64` config digest 与蓝色运行容器 `.Image` 一致。
7. 修改或重建蓝色前，按 `/Volumes/docker/AGENTS.md` 核对现有容器的 Compose 项目、服务、工作目录和配置文件标签必须分别是 `axonhub`、`axonhub`、`/volume1/docker/axonhub`、`/volume1/docker/axonhub/compose.yaml`。

#### 阶段二：冻结并制作一致性快照

1. 现场读取 Axon Switch 的 `backends.active`，并用公网 `/health` 与绿色直连 `/health` 交叉确认当前确实为 `active=green`。未确认时禁止继续。
2. 明确告知用户将进入短暂停服。只有用户确认暂时不使用蓝绿两侧后，才停止当前绿色和未激活的蓝色：

   ```sh
   docker stop --timeout 60 axonhub-local
   # NAS 上：sudo docker stop --time 60 axonhub
   ```

3. 确认两个容器均为 `exited`。绿色停止后，先检查源库，再使用 SQLite `.backup` 创建快照并再次校验：

   ```sh
   # 使用流程开始时记录的 PROMOTION_TS，不要重新生成。
   GREEN_BACKUP_DIR="/Users/tux/Playground/AxonHub/backups/green-to-blue/$PROMOTION_TS"
   mkdir -p "$GREEN_BACKUP_DIR"

   sqlite3 /Users/tux/Playground/AxonHub/data/axonhub.db 'PRAGMA quick_check;'
   sqlite3 /Users/tux/Playground/AxonHub/data/axonhub.db \
     ".backup '$GREEN_BACKUP_DIR/axonhub.db'"
   sqlite3 "$GREEN_BACKUP_DIR/axonhub.db" 'PRAGMA quick_check;'
   shasum -a 256 "$GREEN_BACKUP_DIR/axonhub.db"
   ```

4. 在 NAS 本机先创建 `/volume1/docker/axonhub/backups/green-to-blue/<本次 PROMOTION_TS>`，再使用 `/usr/bin/sqlite3` 对停止后的蓝色旧库执行 `PRAGMA quick_check` 和 `.backup`，把结果保存为该目录下的 `blue-before-axonhub.db`；同时 `cp -p` 保存 `compose.yaml.before`。数据库备份校验成功且 Compose 备份存在后才能覆盖蓝色。

#### 阶段三：恢复蓝色并启动

1. 把绿色静态快照复制为蓝色数据目录中的临时文件 `axonhub.db.incoming-<UTC时间戳>`。在 NAS 本机再次执行 `PRAGMA quick_check` 和 SHA-256；结果必须为 `ok`，哈希必须与绿色快照一致。
2. 继承旧蓝色数据库文件的所有者和模式后，在蓝色数据目录内完成同文件系统改名：
   - 旧 `axonhub.db` 改为 `axonhub.db.pre-green-promotion-<UTC时间戳>`；
   - 存在的 `axonhub.db-wal` 和 `axonhub.db-shm` 使用同一前缀保留；
   - incoming 文件改名为正式 `axonhub.db`；
   - 对正式数据库再次执行 `PRAGMA quick_check`。
3. 使用 `apply_patch` 把 `/Volumes/docker/axonhub/compose.yaml` 的镜像改为目标不可变标签，不得整体重写 Compose。先保存的 `compose.yaml.before` 是回滚基线。
4. 在 NAS 原项目目录验证并只重建蓝色服务：

   ```sh
   cd /volume1/docker/axonhub
   sudo /var/packages/ContainerManager/target/usr/bin/docker-compose config --quiet
   sudo /var/packages/ContainerManager/target/usr/bin/docker-compose up -d axonhub
   ```

   如果镜像是通过 tar 预加载且远端标签尚不存在，改用 `docker-compose up -d --pull never axonhub`，只能使用已加载的本地不可变标签，不能退回官方镜像。
5. 从蓝色容器内部检查 `http://127.0.0.1:8090/health`，并核对容器为 `running`、重启次数正常、镜像引用和镜像 ID 正确、平台为 `linux/amd64`。健康检查未通过时不得切换流量。
6. 蓝色数据库较大时，NAS 冷启动可能持续数分钟。本次 2.7GB 数据库的首次启动约 3 分钟；短时间内 health connection failure 不等于启动失败。若容器仍为 `running`、重启次数为 0 且进程仍有 CPU/IO 活动，应在有上限的观察窗口内继续查看状态和关键日志，不得反复重启。出现容器退出、重启次数增长、数据库校验失败、`FATAL` 或 `PANIC` 时才按失败处理。

#### 阶段四：切换和收尾

1. 优先通过已认证的 Axon Switch 管理页面切换为蓝色，禁止用 `jq`、文本编辑器或整体覆盖方式修改 `state.json`。
2. 无可用管理会话但用户已明确授权无人值守切换时，只能调用 Axon Switch 自身 `internal/store` 的 `SetActive("blue")`：它会使用 `state.lock` 跨进程锁和原子保存。临时辅助程序不得输出后端 URL、Key、会话或完整状态，执行后必须删除并确认 Axon Switch 工作树干净。
3. 重新只读提取 `backends.active`，必须得到 `blue`；公网 `/health` 的 build time、platform 和 uptime 必须与蓝色容器直连结果一致。
4. 只有确认 `active=blue` 后才能重新启动绿色备用容器。绿色启动后只检查 `http://127.0.0.1:9090/health`，不得自动切回绿色。
5. 交付时记录：目标 commit、不可变镜像标签和 digest、蓝绿实际镜像 ID、快照时间戳、快照及旧蓝色备份路径、两次 `quick_check`、哈希比对、当前活动槽位和公网健康结果。
6. 本流程创建的每一份数据库、Compose 或其他回滚备份，都必须立即写入 `/Users/tux/Library/Mobile Documents/iCloud~md~obsidian/Documents/AI Notebook/备份日志.md`，包括描述、时间、Mac/NAS 路径、大小、哈希和校验结果；删除备份时同步删除或更新对应记录。

#### 完成判定：缺一不可

- Forgejo 上的 `ai-slop` 必须包含目标 commit，构建上下文必须来自该 commit 的干净工作树。
- Registry 中必须存在目标不可变标签，且唯一的 `linux/amd64` config digest 与蓝色运行容器 `.Image` 一致。若使用离线应急回退且尚未补推，必须明确报告“Registry 发布待补齐”，不得宣告镜像发布完整完成。
- 绿色快照、蓝色覆盖前备份和蓝色正式数据库均已完成 `PRAGMA quick_check`；绿色快照与蓝色正式数据库 SHA-256 一致。
- 蓝色容器健康、重启次数正常；Axon Switch 权威状态为 `active=blue`，公网 health 与蓝色直连 health 一致。
- 绿色备用容器仅在确认 `active=blue` 后恢复，且绿色直连 health 正常。
- 所有备份均已登记到备份日志；临时辅助程序和镜像 tar 已删除；没有输出或遗留明文 Token、密码或完整状态文件。

#### 回滚

- 蓝色启动前失败：保持 Axon Switch 指向绿色，把蓝色旧数据库和 Compose 恢复后启动绿色；不要删除任何备份或 incoming 文件。
- 蓝色启动成功但切换前验证失败：蓝色保持停用，修复或恢复旧蓝色；绿色源库仍是晋升时的完整数据，可直接恢复服务。
- 切换蓝色后发现问题：先确认绿色健康，再把 Axon Switch 切回绿色；随后停止蓝色，恢复 `compose.yaml.before` 和 `blue-before-axonhub.db`，完成校验后再处理。
- 任何回滚均不得用旧蓝色数据库覆盖已经在新蓝色产生的有效新写入；发生新写入后必须先另做当前蓝色快照，再决定数据取舍。
