# Vibe Webterminal

在自己的 Mac 上运行 Web Terminal，通过内网用 iPad 或其他浏览器操作。基于 ttyd，带触摸面板、方向键、清屏和贴边收起，按需启动。

## 快速上手

需要 macOS、[Homebrew](https://brew.sh) 和 Xcode Command Line Tools（`xcode-select --install`）。

```sh
git clone https://github.com/xhzq233/vibe-webterminal.git
cd vibe-webterminal
./setup.sh
./start.sh
```

首次 setup 会安装缺少的依赖并编译 ttyd。start 会打印当前内网访问地址，在 iPad 浏览器打开即可。

在 Mac 另开终端查看自动生成的登录信息：

```sh
cat ~/.local/share/vibe-webterminal/credential
```

格式为 `username:password`，每次新部署独立生成。启动服务的终端保持打开，在该窗口按 **Ctrl+C** 停止服务；不安装任何自动启动项。

```sh
# 指定自己的项目目录
./start.sh /path/to/your/project

# 修改端口
PORT=7682 ./start.sh

# 指定网卡 IP，而不是默认路由地址
BIND_IP=你的内网IP ./start.sh
```

## iPad 控件

整个面板可拖动。停到左右边缘后松手，自动收成贴边标签；点击标签展开。

| 控件 | 用途 |
| --- | --- |
| ↑ / ↓ / → | 方向键：切换历史命令或移动光标 |
| Space | 输入一个空格，作为部分 iOS 输入法问题的替代入口 |
| Ctrl+C | 通常用于中断前台命令 |
| ⤓ | 滚到输出最下面 |
| Clear | 发送 Ctrl+L，在 zsh 中清屏并保留未提交的输入 |

普通长输出可上下滑动，默认保留 1,000 行历史。Clear 不保证删除滚动历史；vim、less 等程序自行处理按键和滚动。

键盘出现时只调整网页位置，按输入光标的位置避开遮挡；不在 VisualViewport 变化时调用终端 fit、改变行列数或重新连接。

## Safari 认证修正如何工作

部分 Safari WebSocket 连接不会携带页面登录时的 HTTP Basic `Authorization` 请求头。ttyd 1.7.7 原版在 WebSocket 握手过滤阶段要求该请求头，因此可能出现页面能打开、终端却一直重连。

补丁位于 [vendor/ttyd/src/protocol.c](vendor/ttyd/src/protocol.c)，仅在 `check_auth` 中增加：

```c
// n is the copied Authorization header length.
if (n == 0) return true;
```

这里的 `true` 只表示允许继续 WebSocket 握手，不表示允许创建 shell：

1. 网页和 `/token` 仍要求正确的 Basic 登录认证。
2. 页面取得 token 后，通过第一条 WebSocket JSON 消息发送 `AuthToken`。
3. ttyd 原有逻辑验证 token，成功后才调用 `spawn_process` 创建终端。
4. token 缺失或错误会断开连接；如果握手携带了错误的 Authorization，仍会被拒绝。

`-O` 同源检查保持开启。相比原版，未携带认证头的客户端可以建立 WebSocket 握手，但拿不到通过认证才能取得的 token 就不能启动 shell。这是本项目维护的兼容补丁，不是上游已发布修复。

## 使用范围与限制

- 默认是 HTTP + 登录密码，面向可信内网；没有提供 HTTPS 或公网部署方案。
- 网页刷新或断线可能结束对应 shell。长任务需要跨断线保留时，在 shell 内使用 tmux。
- 网络或 IP 改变后停止并重新启动，使用新打印的地址。必要时允许 macOS 入站连接。
- 已在 macOS 验证编译、首次生成凭据、认证、真实 shell 命令。缺少/错误 token 被拒绝，正确 token 不带 WebSocket Basic 请求头也能连接。
- 面板、贴边收起、方向键、清屏和键盘定位经过浏览器模拟测试；**真实 iPad 的键盘、输入法及长期性能仍未完成验收**。

## 修改与构建

- `setup.sh`：安装依赖、编译；`start.sh`：生成登录凭据并前台运行。
- `web/touch-panel.html`：触摸面板和光标定位逻辑。
- `web/index-original.html`：原始 ttyd 前端；`web/index.html`：实际提供给浏览器的页面。
- `vendor/ttyd`：ttyd 1.7.7 源码及上述兼容补丁。

改动面板后重新生成页面，然后刷新浏览器：

```sh
python3 scripts/build-web.py
```

修改 C 源码后运行 `./setup.sh` 重新构建，并停止、重新启动服务。

第三方组件与原始许可证见 [THIRD_PARTY.md](THIRD_PARTY.md)。仓库不携带登录凭据、个人机器地址、日志或编译产物。
