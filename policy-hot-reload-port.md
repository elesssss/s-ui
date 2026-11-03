**策略热更新移植说明**

原版工作树已接入 wanan9999 的策略热更新，保持原版前端、SQLite 驱动、依赖版本及发布平台。当前改动尚未提交。

来源：`s-ui_wanan9999` 的 `4892a1748165d0dd5c20b4f59cacdeb6ac6e6bd8`。目标：原版后端 `3c38b80172d4a6cbfcce9fd224a899d8b126d440`，sing-box `1.14.4`，以及此前已加入的 L2TP/IPsec 接入。

**保存时的行为**

| 操作 | 拨号会话 | 业务连接 |
|---|---|---|
| 仅修改 DNS | 保留 L2TP/IKE/IPsec/PPP 和 gVisor | 普通 TCP/UDP 连接保留，既有 DNS 通道的后续查询使用新策略 |
| 修改路由/规则集或 HTTP Client | 保留 | 关闭现有业务连接，新连接使用新策略 |
| 新增、编辑或删除出站 | 保留 | 按出站依赖链关闭受影响连接 |
| 候选策略准备失败或数据库提交失败 | 旧策略继续运行 | 返回保存错误，释放候选资源 |
| 维护模式中保存 | 核心保持停止 | 验证并保存配置，下次启动生效 |
| 用户凭据、入站配置变更或显式重启核心 | 仍可能断开 | 继续使用原有入站更新/核心重启流程 |

保存由生命周期锁串行执行：事务内生成完整配置 → 准备候选管理器 → 写入审计和提交数据库 → 发布策略。基础配置不再异步重启核心；出站保存不再先删除运行中的旧出站。保留原版的订阅备注前缀更新等其他保存行为，配置变更时间改用原子读写。

**适配原版生命周期**

- 初始策略与每个候选策略有独立 `adapter.Scope`，通过原版 1.14.4 的生命周期 API 启停。旧策略释放不会关闭入站的 scope。
- 使用原版 `core/manager.OutboundManager` 和当前 DNS/连接管理器构造函数，没有回退依赖版本。
- 稳定的路由与 DNS 入口转交当前策略；认证账号仍进入原版连接跟踪和流量统计。
- 新策略先创建并注册自己的 ConnectionManager，再构造 NetworkManager，避免候选网络通知关闭旧策略的连接。
- 初始 NetworkManager 保留到整个核心关闭，策略的其余资源在引用释放后回收；出站测试和空闲连接清理使用当前策略。
- ReferenceManager 也纳入各自策略 scope，以保留原版空闲连接清理功能。

**本轮复查修复**

- 核心停止后不再借出旧策略；DNS 同步和异步查询返回关闭错误，路由请求关闭连接并通知回调，缓存/查询辅助入口也不再访问已关闭的原生服务。
- 停止与候选发布并发时，候选会被关闭，不能重新挂到已停止的核心；`PolicyUpdate.Commit` 返回错误，保存调用方不再忽略发布失败。
- 准备期间持有用于计算出站影响范围的旧策略引用，避免它在普通热更新回收期间被提前释放。
- 新增不启动原生网络管理器的停止、并发发布和候选回收回归测试，并纳入 race 工作流。测试代码尚未在本机执行。

**当前限制**

沿用来源实现的热更新范围：启用 TUN、端点/核心服务、证书提供器/核心 API、NTP 或核心持久缓存时，策略热更新会返回错误，要求先停止核心再保存。日志、证书、命名空间及其他全局设置变更也要求先停止核心。不会为这些情况自动重启拨号会话。

“保留拨号会话”不表示所有 TCP/UDP 业务连接都保持；路由和出站变化按上表处理。

**验证状态**

已补充核心策略、scope 回收、并发连接发布、保存准备/审计/提交失败、繁忙时保存和维护模式回归测试。L2TP 集成测试增加同一拨号会话中的 UDP/TCP DNS 通道更新、路由更新、监听实例保留和踢用户验证。

Test 工作流在隔离网络命名空间内添加测试接口/默认路由，反复执行策略回归并运行 PPP/IPsec 集成测试；race 检查覆盖不启动原生 NetworkManager 的策略引用/连接跟踪测试。编译测试使用 `ldflags_for test`，与原版构建标签配对。

本地已进行源码差异和工作流语法检查。当前环境没有 Go/gofmt 工具链，外部 DNS 解析失败，无法下载固定依赖，因此尚未完成 Go 编译、gofmt、单元/race/隔离集成测试。Windows 内置客户端、爱快、公网 NAT 与长时间换钥仍需实际验收，不能以测试代码存在代替通过结果。

在具备仓库要求的 Go 工具链和依赖后，先运行：

```sh
gofmt -w core/policy_*.go core/box.go core/main.go core/endpoint.go core/outbound_check.go core/tracker_session.go core/l2tp_integration_test.go service/config*.go service/outbounds.go service/client.go service/client_reset_test.go
. ./build-tags.sh
go test -ldflags "$(ldflags_for test)" -tags "$(tags_for test)" ./core ./service
go test -race -ldflags "$(ldflags_for test)" -tags "$(tags_for test)" ./core -run '^Test(SessionTracker|RetiredPolicy|PolicyConcurrent|PolicyInvalidation|PolicyUDP|PolicyScope|PolicyStopped|PolicyCommit)'
```

完整隔离集成步骤见 `.github/workflows/test.yml`，不要在开发主机默认网络中运行真实拨号/路由测试。

本轮复查实际完成了后端/前端已修改及新增文件的空白检查、CI YAML 解析、7 段工作流脚本的 `bash -n` 检查，以及 10 个本轮相关 Go 文件的字符串/注释/括号检查。后者只是词法检查，不代表 Go 语法、类型或编译验证。

仍需取得固定版本依赖后重点验证：sing-box 1.14.4 的接口和网络管理器构造依赖、候选在生命周期初始化之前失败时的资源清理，以及同一 PPP 会话内已经打开的 TCP/UDP DNS 通道能否跨策略更新继续使用。`PreparePolicy` 的调用者仍需串行执行整个准备、提交或中止流程；面板保存由 `lifecycleMu` 保证这一点。维护模式中的 `ValidateConfig` 当前只做配置反序列化，完整的运行时有效性需启动验证。
