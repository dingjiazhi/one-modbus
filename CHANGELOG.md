# CHANGELOG

> 一个干了 30 年电工，不满意现有软件，边学 Go 边做出来的开源网关。
> 每次更新都是现场踩坑的真实记录。

---

## v0.3.35 (2026-06-08)

### 🔧 goversioninfo v1.7.0 — 解决 Windows Defender ML 误报
CNB 环境 `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest` 抓到 v1.6.x，编译出来的 .exe 被 Windows Defender 机器学习模型标为 PUA。升级到 v1.7.0 后问题消失。

### 🏷️ 版本标记
在源码头部加了 `//go:build` 标记 + `resource.syso` 预提交，不再依赖 CNB 流水线安装 goversioninfo。
CNB 流水线只需 `go build`，无需 `go install` 任何工具。

---

## v0.3.34 (2026-06-08)

### 🚀 CNB 防误报优化
CNB 环境不带 gcc，默认 mode=exe 不生效。改用 `-extldflags=-static` 配合 `CGO_ENABLED=0`，静态链接消除动态加载特征，降低杀软评分。

**经验**：CNB 的 go 编译器版本可能随 base image 升级而变化，`-extldflags=-static` 在 golang:alpine 上需要先 `apk add gcc musl-dev`，但 `golang:latest` rootful 模式下自动支持。

---

## v0.3.33 (2026-06-07)

### 🐛 versioninfo.json 编码修复
发现 write 写入的 versioninfo.json 被 PowerShell `Out-File` 默认 UTF-16 编码破坏，导致 `goversioninfo` 解析乱码。改用 UTF-8+BOM 写入后恢复正常。

### 🔗 golang.org/x/net 升级
v0.50.0 → v0.55.0，消除已知安全漏洞。
移除已废弃的 `resource.rc` 和 `.cnb/build.sh`。

---

## v0.3.32 (2026-06-07)

### 🐛 rstrip(.0) 竟然会干掉 0.3.30 的尾 0
CNB 流水线中提取版本号时用了 `rstrip(.0)`，结果 v0.3.30 → "0.3.3"（尾 0 被 strip 了）。改成 `.split(.0)[0]` 按语义截断，不再误杀。

---

## v0.3.31 (2026-06-07)

### 🔧 GHA Release Notes 改用 echo 块
之前用 heredoc 写 release body，YAML 解析在 GHA runner 上间歇性翻车。全部改成 `echo` 多行拼接，稳定了。

### 📝 统一发布模板
GitHub / CNB / Gitee 三个平台的 release body 用同一份模板，不再各自维护。

---

## v0.3.25 — v0.3.29 (2026-06-07)

### 🏗️ CNB 流水线最终定型
这段时间密集迭代 CNB 自动构建流水线，核心变化：

- **v0.3.25**: `go install goversioninfo` 在 CNB 上不可靠（goproxy.cn 有时候连不上 direct），回退到 `tools/goversioninfo` 预编译版
- **v0.3.23**: 修复 CNB Release URL 路径——不是 `/releases/{id}` 而是 `/-/releases`，cnb 不按常理出牌
- **v0.3.22**: 最终方案——`.syso` 文件 pre-commit 到仓库，CNB 流水线不再需要 `go install` 任何东西
- **v0.3.19**: 预编译 `resource.syso`，避免 CNB 运行时装 goversioninfo
- **v0.3.18**: 加 Release 存在性检查，避免重复创建

### 🔑 经验教训
CNB API 的 `id` 字段是字符串（不是数字），用 `grep -oP '\d+'` 拿不到。必须用 `python3` 解析 JSON。并且 `python3` 多行代码在 CNB 解析器里会被弄坏，必须写在一行。

---

## v0.3.14 — v0.3.16 (2026-06-07)

### 🎯 切换到 CNB 自动构建
放弃本地编译 + 手动上传，改为 CNB 标签触发自动构建：
- 打 tag → CNB 流水线自动拉代码 → 编译 → 发布 Release
- 编译时间自动替换（Go 源码里写死 `BUILD_TIME` 字符串，sed 替换为当前时间）
- `.exe` 文件不包含版本号，保持与 GHA 产物一致的命名

### 🐍 Python 三行搞定 JSON
CNB 流水线里所有 JSON 操作用 `python3` 替代 `grep/sed`，不再翻车。

---

## v0.3.10 — v0.3.13 (2026-06-04 ~ 06-05)

### 🚀 MODBUS TCP 支持
以前只支持 RTU over TCP（通过虚拟串口），现在加了个原生 TCP 侦听端口模式：
- 收到 TCP 帧 → 直接转发给已配置的 RTU 从站
- 测试环境：西门子 S7-200 + 三菱 FX3U 都通

### 📉 内存优化
嵌入式端动态分配的缓冲区池改为静态预分配，单连接内存从 128KB 降到 32KB。
32 路同时采集时内存占用不到 1MB。

---

## v0.3.0 — v0.3.8 (2026-05-25 ~ 05-26)

### 🔄 三平台同步发布
GitHub Actions + Gitee + GitCode 三平台打 tag 自动发布：
- GitHub: 国际主力，Actions 自动编译 + Release
- Gitee: 国内主力，本机上传
- GitCode: 从 Gitee CDN 下载，不放附件

### 🌐 跨海传输策略定型
```
美国→香港（一次跨海） → 国内并行分发（cnb.cool + Gitee）
```
国际上传只做一次，后续国内分发全部并行，不重复跨海。
实测：香港 CN2 约 1.7 MB/s，国内 TX 约 9.3 MB/s。

### 📜 AGPL-3.0 许可证
从 MIT 切换到 AGPL-3.0，保护代码不被商用闭源。

---

## v0.2.15 — v0.2.17 (2026-05-25)

### 🏗️ CI/CD 基建
- GitHub Actions 流水线：打 tag 自动交叉编译 Windows exe
- Gitee/GitCode Release 自动同步
- 纯 Go 交叉编译（去掉 mingw CGO 依赖），单文件无依赖

### 🐛 踩坑记录
- GitCode API 没有 `id` 字段，必须用 `tag_name` 判断存在性
- GitCode 上传附件需要两步：先拿预签名 URL，再 PUT headers
- Gitee 上传 22MB exe 跨海超时需要 `--max-time 900 --retry 3`
- GitCode 链接国内 CDN 比直传快 10 倍

---

## v0.1.0 — v0.2.14 (2026-05-22 ~ 05-25)

### 📖 文档和产品化
- 架构图 SVG（三路并行：直连 + 协议转换 + DTU 远程）
- 中英文 README 分拆
- Quick-start 文档：免配置开箱即用
- 产品官网：`one-modbus.pages.dev` + `dingjiazhi.github.io/one-modbus`
- ASCII 架构图 → SVG 视觉图

### 🔐 凭据管理
所有 email token、手机号、API key 全部从代码中移除，走 secrets.db 运行时读取。

---

## v0.0.1 — v0.1.0 (2026-05-22 之前)

### 🎬 开篇
> **Initial commit**: one-modbus data acquisition gateway

一个干了 30 年电工，不满意现有的某组态软件——太贵、太重、太封闭。想改个功能还得找厂商。

"我自己写一个。"

边学 Go 边写，从零到第一个能跑通的版本。那时候还不知道什么 CI/CD、什么许可证、什么产品化——就是想把现场数据收到浏览器里看。

直到某天发现自己的网关已经在生产环境跑了几个月没重启过，才意识到：**这东西可能真比市面上的方案好用。**

---

## 未发布 / 开发中

- Go 源码 14,000 行 —— `modbusrtu_broker.go`（单文件）
- 现场设备数：实测 32 路 DS18B20 + AM2302 + DTU 同时采集
- 数据库：1,109,580 条记录（截至 2026-06-13，仍在写入）
