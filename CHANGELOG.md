# 变更清单

本文件记录 aiotieba-go 的所有使用者可见变更，格式遵循
[Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循
[SemVer](https://semver.org/lang/zh-CN/)。

统一使用下列分类，便于检索：

| 分类 | 含义 |
| ------ | ------ |
| **破坏性变更** | 需要调用方改代码才能升级的改动 |
| **新增** | 新能力 |
| **变更** | 行为或用法发生变化 |
| **修复** | 缺陷修复 |
| **移除** | 删除了的能力（仅在有内容时出现） |
| **内部** | 不影响使用者的改动，如 CI、注释、文档（仅在有内容时出现） |

> **版本策略**：项目处于 1.x 阶段，为尽快对齐 Python 版实现，次要版本（如 1.2 → 1.3）也允许包含
> 破坏性变更。所有破坏性变更都会显式列在 `破坏性变更` 分类下，随 Release 说明一起发布；
> module 导入路径不会随版本变化。

任何面向 `master` 的变动都必须先在 `[Unreleased]` 下补充条目，发版时再整段下移为正式版本。

## [Unreleased]

## [1.5.1] - 2026-09-29

### 修复

- JSON 接口解析失败时，错误消息里带上响应体开头一段（折叠空白后截到 200 个字符）。
  上游返回 HTML（风控页 / 网关页）而不是 JSON 时，原来只报
  `invalid character '<' looking for beginning of value`，分不出是哪种情况，
  拿到这条错误的人无从判断该重试还是该停手；现在能直接看出返回的是什么。
  **错误消息因此变长，且会包含响应体的开头**——把错误原文写进自己日志的调用方请注意

## [1.5.0] - 2026-09-25

### 新增

- `GetComments` 新增 `GetCommentsArgs.Sort`，与 Python 版 `get_comments` 的 `sort` 参数对齐。
  服务端当前不采纳该参数——传时间顺序 / 时间倒序 / 热门序得到的返回顺序完全相同，HTTP 与
  WebSocket 两条传输都如此。保留是为了与上游保持一致，服务端实现后即自动生效

### 变更

- 客户端版本号（`consts.LatestVersion`）由 `22.6.5.1` 升至 `22.10.1.0`，与上游 Python 版对齐。
  所有使用该常量的接口都会以新版本号上报；`get_comments` 也从旧版格式（`12.64.1.1`）切到该版本

### 修复

- 修复 WebSocket 传输从未生效的问题。此前握手委托给 `gorilla/websocket`，而它拒绝调用方设置
  贴吧 IM 要求的非标准握手头 `Sec-WebSocket-Extensions: im_version=2.3`，导致握手必然失败：
  `WithTryWebsocket(true)` 与 `InitWebsocket` 恒为失败，所有能走 WebSocket 的接口都静默回退到
  HTTP。**依赖 WebSocket 的 `SendMsg`、`SetMsgReaded`、`GetGroupMsg` 在此前版本中完全不可用**，
  现已恢复
- 握手与 WebSocket 帧层改为自行实现（与 Python 版的做法一致），不再依赖
  `github.com/gorilla/websocket`

### 内部

- 重新生成受版本号影响的 15 个请求黄金向量（`api/*/testdata/req*.hex`）：先用 Python 绑定复现
  旧向量、确认解码与重编码逐字节无损，再按新版本号重出，保证 Go 与 Python 的输出仍逐字节一致
- `AGENTS.md` 与 `.github/CONTRIBUTING.md` 补充仓库操作约定：已发布的 tag 不可移动（Go 模块代理与
  校验和数据库均不可撤销）、`master` 的 HEAD 停在最新 release tag 上、分支保护与合并方式的现状，
  以及发版 PR 的命名与 tag 形态
- `AGENTS.md` 新增「与上游同步」：写明移植基线、同步只跟上游对百度接口行为的适配（上游没有
  CHANGELOG，改为按路径过滤 `git log`）、基线是参考而非权威，以及不要引入上游 tag 的原因
  （本库 module path 无 `/vN` 后缀，只能有 v0 / v1）
- `CI.yml` 显式声明 `permissions: contents: read`；`.gitignore` 中的 `test` 规则锚定为根目录
  `/test`，避免以后新建的 `test/` 目录被静默忽略

## [1.4.1] - 2026-09-25

### 内部

- `AGENTS.md` 新增「注释与中文文案」：记录中文文案的来源基线（迁移前 Python 源码
  `0847e2aa~1`）、读取方式，以及 `api/*/_api.py` 中只有 `search_global` 有 docstring 等事实
- `.github/CONTRIBUTING.md` 新增「注释规范」：说明「参数」段的书写形态，以及 Go 1.19 起
  `gofmt` 会把缩进注释行重排为 preformatted 代码块这一行为
- 新增 `SECURITY.md`，安全漏洞改走私密漏洞报告而非公开 issue
- `.github/ISSUE_TEMPLATE/` 补上默认标签与 `config.yml`（关闭空白 issue，提供文档与安全报告入口），
  并移除与已关闭的 Discussions 功能冲突的 `discussion.md`
- `dependabot.yml` 的检查频率由每天改为每周，减少更新 PR 的噪音

## [1.4.0] - 2026-09-15

### 变更

- 项目协议由 Unlicense（公有领域）改为 MIT。此前已发布的版本（含 v1.3.1）仍为 Unlicense，
  已获得的授权不受影响；自本次改动起，分发或修改需要保留版权声明与许可声明
- `go.mod` 的 `go` 指令由 `1.27.1` 下调为 `1.26.0`：Go 1.26 及以上都能引入本库，1.27.x 用户
  不受影响。1.26.0 是依赖（`golang.org/x/image`、`x/net`、`x/sys`）声明的最低版本

### 内部

- 修正 `go.mod` 中 `github.com/rs/zerolog` 与 `gopkg.in/natefinch/lumberjack.v2` 的 `// indirect`
  错标——两者都被 `logging` 包直接导入，此前因缺少一次 `go mod tidy` 而被记为间接依赖
- CI 增加 `go mod tidy -diff` 守卫（`CI.yml` 与 `release.yml`），go.mod 不整洁会直接失败
- CI 由单版本改为 Go 1.26 / 1.27 双版本矩阵，并新增同名的 `Test` 聚合任务以维持分支保护
  所需的检查名（矩阵会让检查名变成 `Test (1.26)`）
- `PULL_REQUEST_TEMPLATE.md` 与 `CONTRIBUTING.md` 的检查清单补上 `go mod tidy`

## [1.3.1] - 2026-09-15

### 修复

- `AddPost` 补上成功日志。Python 版 `add_post` 的装饰器标注了 `ok_log_level=logging.INFO`，
  Go 版此前只在失败时记录，成功时少输出一行

### 内部

- 新增 `TestBoolResponseMethodsLogSuccess`，断言返回 `exception.BoolResponse` 的公开方法
  都会记录成功日志（唯一例外 `JoinChatroom`，对应 Python 未标注 `ok_log_level` 的 `join_chatroom`）

## [1.3.0] - 2026-09-15

### 破坏性变更

- 日志改用 [zerolog](https://github.com/rs/zerolog)：`logging.GetLogger()` 的返回值由 `*slog.Logger`
  变为 `*zerolog.Logger`，`logging.SetLevel` 的参数由 `slog.Level` 变为 `zerolog.Level`。
  原先通过 `log/slog` 接管日志的调用方需要改为传入 zerolog 记录器

### 新增

- `logging` 包新增 `PyRepr` / `PyArgs` / `PyErr` / `PyKw`，把参数与异常渲染成 Python 风格文本

### 变更

- 日志消息体与 Python 版 aiotieba 对齐，形如 `[sign_forum] Succeeded. args=('盗墓笔记',) kwargs={}`；
  失败日志的正文改为异常本身，例如 `*exception.TiebaServerError{Code: 340011}` 渲染为 `(340011, '')`
- 客户端日志改为在方法内部记录：同一方法无论失败在哪一步，参数都恒为调用方实参；
  50 个写操作（返回 `exception.BoolResponse` 的方法）新增成功日志
- `logging.EnableFileLog` 的写入端改为无颜色输出并交给 lumberjack 轮转（单文件 10MB、保留 5 份）

### 修复

- `SetThreadPrivacy` 此前只记录失败日志，现补齐成功日志，与 `SetThreadPrivate` / `SetThreadPublic` 行为一致

### 内部

- 依赖新增 `github.com/rs/zerolog`、`gopkg.in/natefinch/lumberjack.v2`，移除对 `log/slog` 的使用
- 新增 `logging_args_test.go`，解析 `client.go` 语法树校验日志参数一致性
- AGENTS.md / README.md / CONTRIBUTING.md 与实际实现对齐
- 新增 `changelog.yml` 守住变更清单；Release 说明改为从 `CHANGELOG.md` 生成，
  并附上变更清单入口与与上一版本的对比链接
- README 的 API 列表拆分为 `docs/api.md`，逐条列出接口与用途

## [1.2.0] - 2026-09-15

### 破坏性变更

- `GetThreads` 的第一个参数由 `fname string` 改为 `ref ForumRef`，以对齐 Python 版的 `fname_or_fid`；
  调用方需改用 `ByFName("天堂鸡汤")` 或 `ByFID(123)` 传参

### 内部

- 依赖升级：`actions/checkout` 4 → 7、`actions/setup-go` 5 → 7

## [1.1.1] - 2026-09-15

### 新增

- 新增 Release workflow，推送 `v*` tag 后自动校验版本号并创建 GitHub Release

### 变更

- CI 覆盖范围扩展到 `master` 分支

### 修复

- 测试中 User-Agent 的期望值改为从 `consts.Version` 推导，避免版本升级后失效

### 内部

- 补全各包的中文文档注释（顶层包与 client 门面、api 请求层、api 结果类型、core 与 helper、genproto 工具）
- 执行 `go mod tidy` 整理依赖

## [1.1.0] - 2026-09-15

### 变更

- HTTP 层由标准库 `net/http` 迁移到 [resty](https://github.com/go-resty/resty)（#4）

### 内部

- 重写 README：补充认证说明、快速开始、API 列表与徽章

## [1.0.0] - 2026-09-14

### 新增

- 首次发布：Python 版 [aiotieba](https://github.com/lumina37/aiotieba) 的全量 Go 移植，
  覆盖 HTTP、WebSocket、BLCP 私有协议、客户端签名与 AES 等加密、protobuf 编解码
