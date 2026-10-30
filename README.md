# Vibe Webterminal

把自己的 Mac 终端放进浏览器，在可信内网中用 iPad 操作 Codex、运行命令。支持顶部多 tab、可拖动的多行输入面板，以及断线后继续运行的 Herdr 会话。

当前版本基于 [herdr-tty](https://github.com/dark2momo/herdr-tty) 的 Go 网关，使用标准 [ttyd](https://github.com/tsl0922/ttyd) 和 [Herdr](https://herdr.dev)。源码已包含本项目的 Panel、键盘适配和重连改动，克隆本仓库即可构建。

## 快速上手

需要 macOS、[Homebrew](https://brew.sh) 和 Xcode Command Line Tools（`xcode-select --install`）。

```sh
git clone https://github.com/xhzq233/vibe-webterminal.git
cd vibe-webterminal
./setup.sh
./start.sh
```

`setup.sh` 安装缺少的 Go、ttyd、Herdr，并编译本项目。**请在普通 Mac 终端中运行 `start.sh`，不要从 Herdr 内部的 pane 启动网页服务。** 网页仍可接入已有 Herdr 会话。

启动后打开打印出的地址，默认端口为 7681。在 Mac 另开终端查看登录信息：

```sh
cat ~/.local/share/vibe-webterminal/credential
```

格式为 `username:password`，在网页登录表单中填写。首次部署生成独立密码；重复启动、更新代码均复用现有密码。不要分享该文件。

保持启动服务的终端打开。在该窗口按 **Ctrl+C** 只停止网页访问，Herdr 和其中的任务继续运行。项目不安装开机自启项。

```sh
# 指定新会话的项目目录
./start.sh /path/to/your/project

# 更换端口或指定本机 IPv4 地址
PORT=7682 ./start.sh
BIND_IP=你的内网IP ./start.sh

# 创建另一套独立会话
VIBE_SESSION=another-project PORT=7682 ./start.sh /path/to/project
```

## Panel 操作

![顶部 tab 和多行输入 Panel](docs/panel.png)

Panel 默认位于可见区域上方，给终端底部输入区留出空间。拖动按钮或面板背景可以调整位置；靠近边缘不会缩小或隐藏。输入框支持多行显示，长草稿可在框内滚动。

| 操作 | 行为 |
| --- | --- |
| 输入框中的键盘 Enter | 只在草稿中换行，不发送到终端 |
| Send | 把草稿粘贴到终端当前输入位置，清空草稿，但不发送回车 |
| Panel Enter | 草稿非空时先粘贴草稿再回车；空框时只向终端发送回车 |
| ↑ / ↓ / → | 终端方向键 |
| Ctrl+C | 通常用于中断前台命令 |
| Clear | 发送 Ctrl+L；由当前终端程序处理 |
| Esc | 发送 Escape |
| Del | 发送 Delete（`0x7f`） |
| Opt | 切换 Option 修饰键；开启后其他 Panel 按键会按 Option 组合发送，再点一次关闭 |

在 Codex 中，可点 **Send** 把草稿放入 Codex 的输入区，确认后再点 **Enter**；也可直接点 Panel 的 **Enter**，一次完成粘贴和提交。Send 是粘贴，不会先清空终端里已有的输入。终端程序如何处理多行粘贴由该程序决定。

在终端上单指滑动会传给 Herdr，并保留惯性滚动；页面本身不会随手势平移。长按拖动可选中文字，双指轻点发送右键。选中后会自动尝试复制到当前浏览器所在设备的剪贴板；只有浏览器拒绝自动复制时才显示 **Copy** 按钮。手动点击仍被拒绝时，会提供可手动复制的文本。内容由 Herdr 直接传给网页，不经过 Mac 剪贴板。


## 键盘弹出时

- 终端行列数保持不变，Codex 不因软键盘出现而重新排版。
- 终端底部始终对齐键盘上方的可见区域；上方内容暂时裁切，不留下空白条。
- 收起键盘后恢复完整画面。旋转设备或真正改变浏览器窗口尺寸时，终端仍会适配新尺寸。
- 基础高度使用 `100dvh` 扣除 Safe Area，跟随浏览器工具栏；不再用工具栏收起时的最大高度 `100lvh`，避免普通模式也裁掉顶部。
- Panel 独立定位，默认靠上，并给底部输入和状态行留出空间。

## Codex 没有颜色

网页服务会过滤启动环境中的 `NO_COLOR`。如果旧的 Herdr 持久会话已继承了 `NO_COLOR=1`，正在运行的 Codex 不会自动恢复颜色。等任务结束并退出 Codex 后，在同一 pane 执行 `env -u NO_COLOR codex resume` 恢复对话即可，不需要停止 Herdr 会话。

## 只保留顶部 tab

如果希望隐藏 Herdr 侧边栏，将以下设置合并到 `~/.config/herdr/config.toml`。已有 `[ui]` 时修改其中的对应项，不要重复添加同名段：

```toml
[ui]
sidebar_start_collapsed = true
sidebar_collapsed_mode = "hidden"
hide_tab_bar_when_single_tab = false
tab_bar_position = "top"
```

```sh
herdr config check
herdr --session vibe-webterminal server reload-config
```

然后断开并重新进入网页，让新客户端应用启动布局。顶部的「＋」用于新建 tab。配置作用于使用这份配置的 Herdr，不会停止现有任务。

## 会话与自动重连

网页默认接入 `herdr --session vibe-webterminal`。首次连接创建会话，之后复用。进入后运行 `codex` 即可；首次 Herdr 引导按提示完成，断线保活不要求安装可选 Agent 集成。

- 刷新网页、iPad 灭屏断网或停止网页服务，只会断开客户端，Herdr 中的任务继续运行。
- 网络恢复后自动重试连接；重连或登录恢复过程保留尚未发送的 Panel 草稿，恢复后不会自动提交。
- 登录有效期默认 30 天（从登录时起算，不随使用自动续期）。签名密钥保存在状态目录，重启网页服务不会使仍有效的登录失效；登录过期时自动转到登录页。
- 多个网页共享同一会话。项目目录参数只用于创建新会话；重连保留原工作区和目录。
- Mac 睡眠会暂停执行，关机或停止 Herdr 会结束进程；这不等同于 Agent 对话恢复。

```sh
# 在普通 Mac 终端进入同一会话
herdr --session vibe-webterminal

# 仅在确实要结束该会话及其中任务时执行
herdr session stop vibe-webterminal
```

## 更新旧版

停止**网页服务**后，在干净的仓库中执行：

```sh
git pull --ff-only
./setup.sh
./start.sh
```

默认沿用 `~/.local/share/vibe-webterminal/credential` 和同名 Herdr 会话。新版改为表单登录，首次升级需要登录一次；之后更新会保留有效登录。旧版直接 shell 的进程不会自动迁入 Herdr。

如果此前使用了自定义状态目录或会话名，启动时继续使用相同的 `VIBE_STATE_DIR`、`VIBE_SESSION`。`VIBE_SESSION_KEY_FILE` 可指定已有签名密钥文件；不要删除或覆盖正在使用的密钥。

## 常见问题

- **页面打不开**：确认网页服务在运行、设备能访问该 Mac 的网络，以及 macOS 允许入站连接。换网络后重新启动并使用新打印的地址。
- **端口被占用**：使用 `PORT=7682 ./start.sh`，或停止原来的网页服务。
- **提示不能在 Herdr 中启动**：从普通 Terminal/iTerm 窗口启动；这是防止嵌套启动，不影响接回已有会话。
- **Send 后命令没执行**：这是预期行为，再点 Panel 的 Enter 才发送回车。
- **网页终端换行与中文标点**：Shift+Enter 会保留 Shift 修饰供 Claude/Codex 区分换行与提交；Mac 第三方输入法的标点通过浏览器原生文本输入转发，保留最终提交的中文或英文字符。更新后刷新网页加载修复。
- **键盘出现后看不到顶部 tab**：上方内容被暂时裁切，收起键盘即可恢复。
- **反复登录**：使用同一个状态目录和 `session-key` 文件，并检查登录是否过期。不要混用旧版 Basic 认证启动脚本。

## 构建与验证

Go 1.23+ 可直接运行 `make build`，产物为 `bin/herdr-tty`。手工运行该程序时默认监听本机回环地址；Mac 启动脚本使用检测到的内网地址并强制表单认证。

```sh
make check  # Go race tests、vet、构建，以及 Node.js 前端行为检查
```

Node.js 仅用于开发检查，不是运行依赖。主要文件：

- `setup.sh` / `start.sh`：Mac 安装与启动入口。
- `internal/app/web/mobile.js` / `mobile.css`：Panel、触摸、键盘适配和重连。
- `internal/app`：Go 网关、Cookie 登录和 ttyd 进程管理。
- `third_party/herdr-tty.LICENSE`：上游许可证。

已使用 Herdr 0.9.0、ttyd 1.7.7 验证登录、Send/Enter 分离、多行输入、断网重连、服务重启保留登录，以及断开后同一 shell 继续运行。键盘适配已用真实 Codex 界面配合浏览器视口模拟验证；真实设备体验以使用反馈为准。

本项目面向可信内网，默认 HTTP 加密码登录，不包含 HTTPS 或公网部署配置。仓库不携带个人凭据、机器地址、日志或编译产物。组件来源与许可证见 [THIRD_PARTY.md](THIRD_PARTY.md)。

## 远端文件浏览

点 Panel 输入框右下角、回形针旁边的**文件图标**，在面板内展开紧凑文件区，随 Panel 一起拖动；再次点文件图标或关闭按钮即可收起。输入绝对路径或 `~/…` 后点 **打开**（也可按回车）。默认恢复上次浏览的目录，刷新网页后也会记住；预览文件后恢复到文件所在目录，首次使用打开主目录。目录记忆保存在当前浏览器，按站点区分。只有点“当前 Tab”才会读取 Herdr 会话当前激活 Tab 的焦点 Pane 目录（优先前台进程 cwd，其次 shell cwd）；无法识别时提示手动输入路径。文件夹名称显示蓝色，普通文件名称显示浅灰色。支持进入目录、上一级、刷新、文本和常见图片预览，以及原文件下载。桌面端点右下角 **Panel** 显示面板。

文件浏览与 Terminal 共用登录，访问范围遵循服务端系统用户的文件权限。当前只提供浏览和下载；文本最多预览前 256 KB，图片最多预览 20 MB，下载不截断。HTML 等文件作为文本预览或附件下载，不在 Terminal 页面内执行。访问的文件位于运行服务的 Mac／Linux 主机上。

## 可选：Cloudflare 内网 DDNS（Mac）

`scripts/cloudflare-ddns.py` 将指定域名的 A 记录更新为 en0 当前 IPv4 地址，使用 DNS only、TTL 60 秒。没有 Wi-Fi 地址时保留原记录；地址变动或网页服务停止时，先调用本机 `herdr-web stop/start` 恢复监听，服务可访问后再更新 DNS。这里只重启 Web 网关，Herdr 会话由独立服务保留。

需要 Python 3 和仅授权目标 Zone 的 `DNS:Edit` API Token。Token 存在仓库外、权限设为 `600`，配置示例：

```json
{
  "zone_id": "YOUR_ZONE_ID",
  "hostname": "terminal.example.com",
  "token_file": "~/.config/herdr-tty/cloudflare-dns-token",
  "gateway": "~/.local/bin/herdr-web"
}
```

`gateway` 必须是支持 `stop/start` 的本机启动脚本，启动时读取 en0 地址并在 7681 端口监听。脚本和配置安装到固定本机路径后，可用用户 LaunchAgent 的 `RunAtLoad=true`、`StartInterval=600` 每 10 分钟执行：

```sh
python3 scripts/cloudflare-ddns.py ~/.config/herdr-tty/cloudflare-ddns.json
```

访问地址是 `http://域名:7681`。客户端仍需能访问 Mac 所在内网；DDNS 不提供穿透或 HTTPS。地址变化后，更新周期和 DNS 缓存都会影响恢复时间。

Panel 支持上传任意类型文件，也支持在输入框粘贴图片或浏览器提供的文件，也可点输入框右下角的 **回形针** 选择文件。发送前图片显示缩略图，其他文件显示文件名，点击附件可删除。Send 上传文件并将服务端主机上的绝对文件路径连同文本粘贴到终端；Enter 完成上传后再发送回车。支持任意文件类型，单个文件上限 20 MB。文件保存在服务端用户缓存目录 `herdr-tty/files/`（Mac 为 `~/Library/Caches/herdr-tty/files/`，Linux 通常为 `~/.cache/herdr-tty/files/`），重启 Web 后仍保留；不再需要时可手动清理。浏览器未提供剪贴板图片时使用回形针选择按钮。终端中的 agent 需能读取服务端主机的文件，图片路径不会自动转成 agent 的原生附件。
