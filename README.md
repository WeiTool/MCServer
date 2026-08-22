# MCServer

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

一个基于 **Wails v2** 的 Minecraft 服务器管理桌面应用。无需打开命令行，在图形界面里完成服务器的启动、停止、监控与配置。

## 功能特性

- **服务器列表**：自动扫描 `servers` 目录，识别每个服务器的 jar、eula 状态、mod/插件数量
- **一键启停 / 重启**：每台服务器可独立配置 Java 路径与内存；未同意 eula 时自动处理并重启
- **控制台终端**：实时日志流（自动处理 GBK/UTF-8 编码）、控制台命令发送、日志级别着色（错误/警告/命令）、导出日志与错误日志、一键清除
- **首页仪表盘**：CPU（P 核）使用率、系统内存、JVM 内存、GC 统计折线图（YGC/FGC/GCT）、磁盘读写速率、在线玩家列表
- **类型 / 版本自动检测**：根据启动日志关键字自动识别服务器类型（Vanilla/Fabric/Forge 等，关键字可在 `config/server-types.json` 自定义），通过 MC status 协议自动检测版本
- **设置页**：`server.properties` 全字段可视化编辑（中文标签、常用配置卡片）
- **CPU 亲和性绑定**：自动识别性能核（P-Core）并将 Java 进程绑定到 P 核，避免跑在 E 核上
- **进程生命周期管理**：Windows 下通过 Job Object 保证关闭应用时所有 Java 子进程一并终止（Linux 使用 Pdeathsig）
- **Java 环境管理**：扫描系统已安装的 Java，支持手动选择任意路径的 Java

## 技术栈

| 层 | 技术 |
| --- | --- |
| 桌面框架 | [Wails v2](https://wails.io)（Go + WebView2 / WebKitGTK） |
| 后端 | Go：`api`（按领域拆分）、`launch`（进程/检测/eula/日志流）、`storage`（JSON 配置集中读写，并发加锁）、`system`（sysinfo / cpubind / gc / io） |
| 前端 | Vue 3 + TypeScript + Pinia + Vue Router + naive-ui + ECharts，`api / composables / stores` 分层 |

## 项目结构

```
MCServer/
├── app.go / main.go        # 应用入口，API 装配与窗口配置
├── window/                 # 窗口尺寸与 DPI 缩放自适应
├── backend/
│   ├── api/                # 暴露给前端的方法（Server/Process/Config/Java/Status/Export）
│   ├── launch/             # 进程生命周期：启动/停止/类型检测/eula/日志流/Job Object
│   ├── server/             # 服务器目录扫描与信息查询
│   ├── storage/            # ServerList.json 与 server.properties 读写
│   ├── system/             # 系统信息：Java 识别 / CPU 绑定 / GC / 磁盘 IO
│   ├── model/              # 数据结构
│   └── utils/              # 路径、协议查询等工具
└── frontend/src/
    ├── api/                # wailsjs 调用封装
    ├── components/         # Dashboard / Terminal / Settings 及图表组件
    ├── composables/        # 页面逻辑（useDashboard / useTerminal 等）
    ├── stores/             # 跨路由共享状态（Pinia）
    └── utils/              # 日志级别、时间格式化等
```

## 快速开始

环境要求：Go 1.2x、Node.js、Wails CLI（`go install github.com/wailsapp/wails/v2/cmd/wails@latest`）。

```bash
# 安装前端依赖
cd frontend && npm install && cd ..

# 开发模式（前端热更新）
wails dev
```

## 构建发布

```bash
wails build
```

产物为可执行文件（Windows 下是 `build/bin/*.exe`），直接拷贝给用户即可。

## 服务器目录约定

- 在可执行文件**同级**创建 `servers` 目录，其中**每个子文件夹视为一台服务器**，放入服务端 jar（如 `server.jar`）即可被识别
- 每台服务器在设置页配置自己的 Java 与内存，配置持久化在 `config/ServerList.json`
- 服务器类型检测关键字可在 `config/server-types.json` 中增删改（首次运行自动生成）
- 首次启动某台服务器时若缺少 eula.txt，应用会自动走"等文件生成 → 同意 eula → 重启"流程，无需手动处理

## 平台支持

- **Windows**：完整支持（Job Object 进程管理、P-Core 识别与绑定、DPI 缩放自适应）
- **Linux**：核心功能支持（Pdeathsig 生命周期、taskset 绑定）

## 注意事项

- 启动服务器**必须**先在控制台页为它选择 Java（纯 JRE 也能运行服务端，但 GC 图表依赖 JDK 自带的 `jstat`）
- GC / 磁盘 IO 等 JVM 统计仅在服务器运行时推送，数据为累计计数器（YGC/FGC/GCT 从 0 起步，空闲服务器长时间为 0 属正常现象）

## 许可证

本项目基于 [MIT License](LICENSE) 开源。任何人都可以自由使用、修改、分发（包括商用），只需保留版权声明。

> 注意：本项目是 Minecraft 服务器**管理工具**，不包含 Minecraft 服务端本体；使用时需自行准备服务端 jar（其受 Mojang 许可约束，与本项目无关）。

