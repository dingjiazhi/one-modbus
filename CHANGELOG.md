# CHANGELOG

> 一个干了 30 年电工，不满意现有的某组态软件，边学 Go 边做出来的开源网关。
> 这里记录的不是功能清单，是每一次把东西推到四个平台时经历的真实苦难。

---

## 写在前面

这个项目不是从一个 PRD 开始的。

是一个干了 30 年现场的电工，对着某组态软件的报价单说了一句"这玩意儿我自己写"，然后他真的写了。边学 Go 边写，白天干现场活晚上敲键盘，14,000 行代码，跑在生产环境几个月没重启过。

但比写代码更难的是——**把代码变成产品，把产品推到四个平台，让外面的人能下载到。**

接下来的每一行，都是这些苦难的记录。

---

## 苦难全景图

```
    GitHub         Gitee         GitCode          CNB
      ▲              ▲              ▲              ▲
      │              │              │              │
      │    CI/CD 噩梦（2026-05-25 开始）           │
      │    · Actions 从零搭流水线                  │
      │    · 纯 Go vs CGO 交叉编译 来回切         │
      │    · Gitee 上传 900 秒超时 + retry         │
      │    · GitCode API 全凭猜（没文档）           │
      │    · CNB API 路径不按常理出牌              │
      │    · 编码 UTF-16 GBK UTF-8 BOM 大乱斗      │
      │    · Windows Defender 误报拉锯战            │
      │    · Token 过期、路径翻车、解析器坏了       │
      │                                            │
      └──────────── 一个人，一台电脑，30 天 ────────┘
```

---

## 2026-06-08 | v0.3.35 — Windows Defender ML 误报最终战

**表面问题**：CNB 编译的 .exe 被 Defender 的机器学习模型标为 PUA。

**真正的苦难**：这已经不知道是第几次被误报了。之前试过：
1. `-extldflags=-static` 静态链接 → 没用
2. `CGO_ENABLED=0` 去掉 CGO → 没用
3. 加 versioninfo 资源文件 → 部分缓解
4. 最终发现是 CNB 环境装的是 goversioninfo v1.6.x，升级到 v1.7.0 才解决

**原因**：`go install @latest` 在 CNB 上抓到的是 v1.6.x，v1.7.0 才修复了 PE 文件头的某些字段。
**解决方案**：`resource.syso` 预提交到仓库，CNB 流水线只需 `go build`，不装任何工具。

**感想**：一个杀软误报，从 v0.3.18 到 v0.3.35，跨了十几个版本才彻底解决。

---

## 2026-06-08 | v0.3.34 — CNB 环境没有 gcc

**表面问题**：`-extldflags=-static` 在 golang:alpine 上需要先装 musl-dev。

**真正的苦难**：CNB 的 base image 版本不可控。今天能编译的流水线，明天 CNB 升级了 image 就挂了。`golang:alpine` 和 `golang:latest rootful` 的行为完全不同，前者要 apk add，后者自动支持。你没法控制 CNB 哪天换 base image。

**解决方案**：CGO_ENABLED=0 + `-extldflags=-static`，纯 Go 静态链接。

---

## 2026-06-07 | v0.3.33 — UTF-16 暗杀 versioninfo.json

**表面问题**：goversioninfo 解析 versioninfo.json 乱码。

**真正的苦难**：PowerShell 的 `Out-File` 默认输出 UTF-16（带 BOM）。你写了一个 .json 文件，看起来是对的，但 goversioninfo 读进去全乱码。因为 .json 文件在记事本里打开是正常的——UTF-16 带 BOM，记事本认。但 goversioninfo 不认。

**排查过程**：先怀疑编码 → 用 Notepad++ 看右下角显示 UTF-16 BE → 改用 UTF-8+BOM 写入 → 好了。

**教训**：永远不要用 PowerShell 的 `>` 或 `Out-File` 写 JSON。Python 的 `json.dumps()` + UTF-8 才是最稳的。

---

## 2026-06-07 | v0.3.32 — rstrip('.0') 杀了 0.3.30 的尾巴

**表面问题**：CNB 流水线提取版本号，v0.3.30 变成了 0.3.3。

**真正的苦难**：一行 `rstrip('.0')` 看起来人畜无害——去掉版本号末尾的 .0。但 v0.3.30 → strip 掉尾部的 ".0" → 0.3.3。版本号直接从 0.3.30 降级到了 0.3.3。Release 发布到 CNB 上，用户看到的版本是错的。

**修复**：`version.split('.0')[0]` 按语义截断，不杀无辜。

**感想**：Python 一行代码，CNB 上跑了一小时才发现的 bug。

---

## 2026-06-07 | v0.3.31 — heredoc 被 GHA YAML 解析器谋杀

**表面问题**：GitHub Actions Release body 用 heredoc 写，碰到 YAML 解析间歇性翻车。

**真正的苦难**：
```
cat <<EOF > body.md
# Release v0.3.30
...
EOF
```
这行在本地跑没问题，在 GHA runner 上跑 → YAML 解析器把 `<<EOF` 后面的文本当成了 YAML 语法的一部分。不是每次都挂，是间歇性的。查了一天，最后全部改成 `echo` 多行拼接 + `printf`，不再用 heredoc。

---

## 2026-06-07 | v0.3.25 — v0.3.29 — CNB 流水线地狱周

**这是整个过程中最痛苦的一周。** 五天打了十几个 tag，每个 tag 都是 CNB 流水线上的一次翻车。

### v0.3.25: go install goversioninfo 在 CNB 上不可用
CNB 环境的 GOPROXY 设的是 goproxy.cn，但 goproxy.cn 有时候连不上 `direct`。goversioninfo 装不上 → 回退到 `tools/goversioninfo` 预编译版。

### v0.3.23: CNB Release URL 让人抓狂
GitHub 的 Release API 是 `/repos/{owner}/{repo}/releases`。
Gitee 的 Release API 是 `/repos/{owner}/{repo}/releases/tags/{tag}`。
GitCode 的 Release API 是 `/api/v5/repos/{owner}/{repo}/releases`。

**CNB 呢？** `/-/releases`。不是 `/releases`，是 `/-/releases`。而且不是 `/{owner}/{repo}/-/releases`，是直接在 org 级别。官方文档没有，试了 7 种路径才试出来。

### v0.3.22: 最终方案——.syso pre-commit
CNB 流水线不要装任何 go 工具。所有编译依赖全部提交到仓库。`.syso` 文件预编译好提交，CNB 只需要 `go build`。

### Python JSON 解析翻车
CNB 的 API 返回 JSON，其中 `id` 字段是字符串类型（"1084391038421004288"），不是数字。`grep -oP '\d+'` 拿不到完整的 id（只拿到片段，因为 id 里也包含数字以外的字符……等等，id 里只有数字，但因为是字符串，所以 grep 其实也能拿到）。

**真正原因**：`grep -oP '\d+'` 确实能拿到数字，但 id 太长了。不对，其实都能拿到。真正的问题在别处——CNB 的流水线解析器把 `python3` 的多行代码弄坏了。

```
# 这段代码在 CNB 上挂了：
python3 -c "
import json
data = json.load(...)
print(data['id'])
"
```

CNB 的 YAML 解析器把多行的 `python3 -c "..."` 里的换行符吃了。必须写在一行：
```
python3 -c "import json,sys; print(json.load(sys.stdin)['id'])"
```

这种一行代码调试起来：打 tag → 等 CNB 编译 → 看日志 → 翻车 → 再打 tag。**每个循环至少 10 分钟。**

---

## 2026-06-07 | v0.3.14 — v0.3.16 — 放弃本地编译

**苦难**：之前一直是本机编译 → 手动上传到 4 个平台。每个版本要：
1. 本机编译 exe（2 分钟）
2. 上传 GitHub Release（5 分钟，看网络）
3. 上传 Gitee Release（900 秒超时设置，实际 3-5 分钟）
4. 更新 GitCode Release 链接
5. 推送 CNB

**手动操作一个版本至少 15 分钟，而且容易忘步骤。**

**转机**：决定切换到 CNB 自动构建。打 tag → CNB 自动拉代码 → 编译 → 发布 Release。

但 CNB 的 pipeline 配置又是一条学习曲线。编译时间替换用 sed 操作源文件，CNB 的环境里 sed 版本不同，行为不同。试了 3 个版本才稳定。

---

## 2026-06-04 ~ 06-05 | v0.3.10 — v0.3.13 — MODBUS TCP 和内存优化

**相对轻松的一个版本。** 真正的苦难在代码里：

- 客户问"能不能直接 TCP 连我的 PLC？"——加了原生 TCP 侦听模式
- 实测 32 路 DS18B20 同时采集发现内存暴涨——缓冲区池从动态分配改为静态预分配，128KB → 32KB

**为什么这段代码有苦难？**
因为这些功能是在生产环境上直接改的。糖厂的采集系统跑着，你不能停。改完代码 → 编译 → 替换 exe → 观察 24 小时 → 没问题才敢打 tag 发布。

---

## 2026-05-25 ~ 05-26 | v0.3.0 — v0.3.8 — 三平台同步发布的地狱开端

这是第一次尝试把项目发布到三个平台（当时还没 CNB）。每一行 CI/CD 配置都是血泪。

### GitHub Actions
- 从零写 workflow YAML。分支名是 `master` 不是 `main`，卡了半天。
- 纯 Go 交叉编译不需要 CGO → 写错了，用了 mingw → 后面又改回来。
- 编译时间戳替换——Go 源码里硬编码 `BUILD_TIME` 字符串，sed 替换。
- Actions 的 token 权限不够 → 配了 workflow 权限 → 配了 repo token。
- Release 用 `softprops/action-gh-release` 还是 `cli/cli` 还是手动 API？试了 3 种方案。

### Gitee
- Gitee 的 API 有速率限制。
- 22MB exe 从本机上传到 Gitee（香港 → 美国 → 国内），不翻墙 5KB/s，翻了墙 1.7 MB/s。
- `--max-time 900` 不够 → 改 `--retry 3 --retry-delay 30`。
- API 返回的 upload URL 有时效性。

### GitCode —— 最大的苦难来源
- **没有文档。** GitCode 的 API 文档找不到，所有端点全靠试。
- **Release 对象没有 `id` 字段。** 社区版 API 返回的 Release 没有 `id`，必须用 `tag_name` 判断存在还是更新。
- **WAF 拦截。** `api.gitcode.com` 被封过，后来发现要用 v5 API 绕过。
- **上传附件分两步。** 先 POST 拿预签名 OBS URL → 再 PUT 到预签名 URL 带特殊 headers。光这一步调试了 6 个版本。
- **OBS headers 谜题。** `x-obs-acl`、`x-obs-meta-project-id`、`x-obs-callback`——不传就 403，传错格式也 403。
- **后来干脆放弃上传附件。** 改从 Gitee CDN 下载链接代替，反而更快。

### 跨海传输策略
```
GitHub CI（美国）→ 编译 exe
         ↓
  香港 CN2 VPS（8 元/月）→ 下载
         ↓
  国内并行分发：
    ├── cnb.cool（腾讯云内网，9.3 MB/s）
    └── Gitee（国内 CDN）
```

**为什么要这样搭？**
GitHub 在国内被墙，直连 ~20 KB/s。香港 CN2 出去速度焊死 13.6 Mbps。只能做一次跨海传输，到香港后国内并行分发。

---

## 2026-05-25 | v0.2.15 — v0.2.17 — 第一条 CI/CD 流水线

**从零开始的 DevOps。**

一个电工出身的人，第一次写 GitHub Actions workflow。不会 YAML，不会 CI/CD，不知道什么叫 artifact，不知道什么叫 release asset。

- GitHub Actions 怎么触发？→ 学
- 交叉编译 windows/amd64 要加什么参数？→ 试了 4 种组合
- 编译产物怎么变成 Release 附件？→ 查了 3 天文档
- APT 换源怎么写在 Actions 里？→ 查
- `continue-on-error: true` 什么时候用？→ 踩了才知道

**每一步都在学，每学一步都在试错。** 打一个 tag，等 5 分钟看跑不跑得过，跑不过就再打一个。一个下午打十几个 tag 是常事。

---

## 2026-05-22 ~ 05-25 | v0.1.0 — v0.2.14 — 产品化的第一周

### 许可证选择
MIT → AGPL-3.0。不是因为懂开源许可证，是因为有人私信问"你这个能不能商用？"——他问了，你就得想这个问题。

### 移除竞品名称
README 里原本写了"替代某组态软件"，遇到一些反馈后改成"替代某同类软件"。

### 凭据管理
代码里原本有：
```
email_token = "xxx"
phone = "138..."
api_key = "sk-..."
```
一个个从代码里挖出来，放到 secrets.db。挖了几个晚上。

### 架构图 SVG
从 ASCII art 画架构图，到学 SVG 画矢量图。中英文各一份。
一个工业现场的人，第一次用 SVG 画图。

### 产品官网
`one-modbus.pages.dev` — 用 Cloudflare Pages 搭的。
`dingjiazhi.github.io/one-modbus` — GitHub Pages。
域名、DNS、CDN、Pages 部署——每一件事都是第一次做。

---

## 2026-05-22 之前 | v0.0.1 — 写在代码之前

> **Initial commit**: one-modbus data acquisition gateway

**真正的第一个 commit 不是** `git init`，是：
> 一个干了 30 年电工的人，坐在电脑前，打开 VSCode，打了一个 Go 文件。
> 他不知道什么叫 CI/CD，不知道什么叫 GitHub Release，不知道什么叫开源许可证。
> 他只知道那个某组态软件太贵了，他要自己写一个。

然后他写了 14,000 行。

中间无数次想放弃——"我一个电工学什么 Go？"、"这 bug 查了一天了"、"编译又不过"、"网络又断了"、"这个功能到底要不要加"。

**但他没停。**

直到某天，他发现自己的网关已经在生产环境跑了几个月没重启过。那一刻他才知道——**这东西真的能用了。**

然后他决定把它发出来。

然后有了上面所有的苦难。

---

## 当前状态（2026-06-13）

| 项目 | 值 |
|------|-----|
| Go 源码 | 14,000+ 行（`modbusrtu_broker.go`，单文件） |
| 现场设备 | 实测 32 路 DS18B20 + AM2302 + DTU 同时采集 |
| 数据库 | 1,109,580 条（仍在写入） |
| 云主机成本 | 香港 8 元/月 + 国内 84 元/15 月 = 不到 120 元/年 |
| 替代方案市价 | 某组态软件 10 万 + 年维护费 |
| 发布平台 | GitHub + Gitee + GitCode + CNB |
| 许可证 | AGPL-3.0 |

---

## 后记

如果你读到这里的每一行苦难都让你觉得"这人也太难了吧"——

请记住：
**这些苦难是一个没学过计算机的人，用业余时间，一个人，全部扛下来的。**

如果你正在做自己的项目，别放弃。他行，你也行。
