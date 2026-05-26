# one-modbus — Modbus RTU Data Acquisition Gateway / 数据采集网关

**English** | **中文**

A production-grade Modbus RTU gateway for Windows. Single `.exe`, no dependencies.
一个 .exe 搞定工业数据采集全链路。

📥 **Download / 下载**：
- [GitHub Releases](https://github.com/dingjiazhi/one-modbus/releases) (International)
- [Gitee Release](https://gitee.com/dingjiazhi/one-modbus/releases) (国内快)

---

> A 30-year electrician built this because he wasn't satisfied with existing software. He learned Go along the way.
>
> 一个干了 30 年电工，不满意现有软件，边学 Go 边做出来的开源网关。

---

## 3 步上手

1. 下载 `modbusrtu_broker.exe`
2. 同级目录放 **项目变量信息.xlsx**（没有会自动生成模板）
3. 双击运行，浏览器打开 **http://127.0.0.1:53046/统计**

> 详细配置见 [docs/quick-start.md](docs/quick-start.md)

---

## 功能

| 功能 | 说明 |
|------|------|
| **多串口并发采集** | 每个串口独立 goroutine，互不阻塞 |
| **批量读取优化** | 同设备多变量合并为一条 Modbus 请求 |
| **零代码配置** | Excel 填变量表，双击即用 |
| **REST API** | 任意变量值可被第三方系统读取 |
| **SQLite 历史存储** | 自动记录历史数据，网页图表查询 |
| **微信报警** | 企业微信群机器人推送状态和报警 |
| **邮件报表** | 定时数据报表 + 即时报警邮件 |
| **远程升级** | 浏览器上传新版 exe，自动替换重启 |

---

## 远程采集（Internet + DTU）

不限本地串口。搭配 **99 元的 DTU（串口转 TCP 模块）** + 虚拟串口软件，采集千里之外的设备：

```
工厂A：254台电表 → RS-485 → DTU(¥99) → Internet ┐
工厂B：PLC         → RS-485 → DTU(¥99) → Internet ┼→ 虚拟串口 → one-modbus 网关
工厂C：传感器     → RS-485 → DTU(¥99) → Internet ┘  (TCP转COM)  (实时轮询)
```

- 1 个 DTU + 1 条 RS-485 总线 = 每站最多 **254 台设备**
- 理论最大 **64,516 台设备** 同机采集
- **每台设备硬件成本不到 0.4 元**

---

## 兼容性

- **操作系统**：Windows 7/10/11/Server（需 COM 口权限）
- **协议**：Modbus RTU（RS-232/RS-485），功能码 1/2/3/4
- **设备**：PLC、智能电表、传感器、变频器、温控器

---

## 架构图

![架构图](docs/architecture.svg)

---

## 许可证

AGPL-3.0

---

## English

One .exe, no dependencies. Production-grade Modbus RTU gateway for Windows.

**Quick Start**: Download → Place Excel config → Double-click → Open browser

**Features**: Multi-port concurrent collection, zero-code Excel config, REST API, SQLite storage, WeChat/Email alerts, remote upgrade.

**Architecture**:

![Architecture](docs/architecture-en.svg)

**License**: AGPL-3.0

### License
AGPL-3.0
