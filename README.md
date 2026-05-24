# one-modbus

A production-grade Modbus RTU data acquisition gateway for Windows. Runs as a single `.exe` — read from serial devices, serve via HTTP API, store to SQLite, push alerts to WeChat/Email.

一个 .exe 搞定工业数据采集全链路：多串口 Modbus 并发采集 → REST API → SQLite 历史存储 → 微信/邮件报警。

Traditional industrial data acquisition requires three separate systems: a data collector, a web visualization platform, and custom development for alerts and reports. **This single .exe replaces all three layers.**

## Features

- **Multi-port, multi-device concurrent collection** — Each serial port runs independently in its own goroutine
- **Batch read optimization** — Multiple variables on the same device are packed into a single Modbus request
- **Zero-code configuration** — Fill in an Excel spreadsheet, double-click the .exe, done
- **Built-in HTTP API** — Read any variable value via REST, integrate with any frontend or SCADA
- **SQLite time-series storage** — Automatic historical data logging with configurable intervals and web-based chart query
- **Enterprise WeChat alerting** — Push status reports and alarms to your WeChat Work group robot
- **Email reports** — Scheduled data reports and instant alert emails
- **Remote upgrade** — Upload a new .exe via browser, auto-replace and restart

## Remote Data Acquisition (Internet + DTU)

Not limited to local serial ports. Pair with a **¥99 DTU (serial-to-TCP converter)** and virtual COM port software to collect data from devices across the internet:

```
Site A: 254 meters → RS-485 bus → DTU(¥99) → Internet →
Site B:  PLCs       → RS-485 bus → DTU(¥99) → Internet →  Virtual COM software  →  one-modbus gateway
Site C:  Sensors    → RS-485 bus → DTU(¥99) → Internet →  (TCP-to-COM bridge)     (real-time polling)
```

- 1 DTU + 1 RS-485 bus = up to **254 devices** per site (Modbus address limit)
- 1 server = up to **254 virtual COM ports**
- Theoretical max: **64,516 devices** on a single server
- Hardware cost per device: **less than ¥0.40**

The gateway doesn't care if a COM port is local or 100km away — it just pulls Modbus data through it.

## Quick Start

1. Prepare `项目变量信息.xlsx` (project variable configuration) in the same directory as the .exe. **If the file doesn't exist, the software auto-generates a template** — or use the desktop shortcut to create one.
2. Double-click `modbusrtu_broker.exe`
3. Open browser to **`http://127.0.0.1:53046/统计`** — this is the main entry point. If not logged in yet, it will guide you to the login page.
4. Login with username and password — **configured inside the Excel table**, not hardcoded
5. After login, the page shows all live API call links with actual examples, real-time values, and gateway status. Everything you need is right there.

See `docs/quick-start.md` for detailed setup.

## Compatibility

- **OS**: Windows (7/10/11/Server) — COM port access required
- **Protocol**: Modbus RTU (RS-232/RS-485), function codes 1/2/3/4
- **Devices**: PLCs, smart meters, sensors, VFDs, temperature controllers — any Modbus RTU device

## Architecture

```
┌──────────────────────────────────────────────────────┐
│ Service Layer                                        │
│  REST API  ·  File Server  ·  Charts  ·  OTA Update │
├──────────────────────────────────────────────────────┤
│ Core Engine                                          │
│  Modbus Collector  ·  Batch Reader  ·  Scheduler     │
│  SQLite Store  ·  Excel Config                       │
├──────────────────────────────────────────────────────┤
│ Notifications                                        │
│  WeChat Work Robot  ·  WeChat Alert  ·  Email Push  │
├──────────────────────────────────────────────────────┤
│ Field Devices                                        │
│  PLC  ·  Meter  ·  Sensor  ·  VFD  ·  Temp Ctrl     │
└──────────────────────────────────────────────────────┘
```

## License

MIT
