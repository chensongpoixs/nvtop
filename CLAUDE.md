# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 构建与运行

```bash
./build.sh                           # 一键构建：npm build + go build → nvtop-server 二进制

cd backend && go run .               # 后端开发运行（需要 Linux + NVIDIA GPU）
cd frontend && npm install && npm run dev  # 前端开发模式（Vite HMR，代理 /ws /api 到 localhost:8080）
./nvtop-server                       # 启动生产服务，浏览器访问 http://localhost:8080
```

## 技术架构

- **Go 1.22** 后端，模块名 `nvtop-server`，仅依赖 `gorilla/websocket` 和 `gopkg.in/yaml.v3`
- **CGO NVML** 采集 GPU 数据（`backend/gpu/nvml.go`），编译需 `CGO_ENABLED=1` + NVIDIA 驱动头文件
- **Vue 3** (Composition API) + **Vite 5** + **Tailwind CSS 3** + **Chart.js 4** 前端
- 部署为**单一二进制**：Vite build 输出到 `frontend/dist`，`go:embed all:frontend/dist` 嵌入到 Go binary
- 构建脚本将 `frontend/dist` 复制到 `backend/frontend/dist` 供 embed 使用

## 后端架构

| 包 | 职责 |
|---|---|
| `backend/main.go` | 入口：初始化 NVML → 初始化 CUDA → 启动 WebSocket Hub → 注册路由 → HTTP 监听 |
| `backend/config/config.go` | YAML 配置加载，`Default` 变量定义默认值，`PORT` 环境变量覆盖端口 |
| `backend/gpu/types.go` | `GPUInfo`、`GPUProcess`、`SystemInfo`、`Snapshot` 数据结构 |
| `backend/gpu/nvml.go` | CGO NVML 封装：`Init()`/`Shutdown()` 生命周期，`GetAllGPUInfo()` 采集所有 GPU 指标（基础+高级+DeviceQuery）和进程；`parseThrottleReasons()` 解析时钟节流位掩码；`computeModeString()` / `brandToString()` / `archToString()` 映射枚举 |
| `backend/gpu/cuda.go` | CGO CUDA Driver API 封装：`CUDAInit()` 初始化，`getCUDAProps()` 通过 `cuDeviceGetAttribute()` 采集 SM 数量、L2 Cache、Thread 限制、Shared Memory、Warp Size、特性标志等 deviceQuery 属性；`coresPerSM()` 按 CC 查表（CC→cores/SM），用于 NVML `nvmlDeviceGetNumGpuCores` 不支持的 GPU（如 Blackwell）回退计算 CUDA Cores = SMs × coresPerSM |
| `backend/gpu/system.go` | 从 `/proc/stat` 和 `/proc/meminfo` 读取 CPU/内存（两次采样差值计算 CPU 使用率） |
| `backend/api/handler.go` | `GET /api/gpus` REST 端点返回 JSON 快照 |
| `backend/ws/hub.go` | WebSocket Hub：连接管理 + ticker 每秒采集广播；`readPump`/`writePump` goroutine 模式 |

无 GPU 时 NVML 初始化失败仅打印 WARNING 不会退出，系统监控仍然可用。

### GPU 高级指标（`backend/gpu/nvml.go` — `getGPUInfo()` 内）

除基础指标外，每个 GPU 还采集以下深度指标（不支持的特性静默跳过，返回零值/空字符串，前端显示 `-` 表示未采集到）：

| 分类 | 指标 | NVML API |
|------|------|----------|
| Performance | P-State (P0~P15)、GPU 运行模式 | `nvmlDeviceGetPerformanceState` |
| Performance | 时钟节流原因（位掩码→文本列表：Thermal/Power/Idle 等 10 种） | `nvmlDeviceGetCurrentClocksThrottleReasons` |
| Memory | 显存总线位宽 (bits)、最大显存时钟 (MHz) | `nvmlDeviceGetMemoryBusWidth` / `nvmlDeviceGetMaxClockInfo` |
| Memory | 当前显存带宽 + 理论最大带宽 (GB/s) = width × clock × 2 / 8 / 1000 | 计算得出（当前用 `clock_memory_mhz`，最大用 `max_memory_clock_mhz`） |
| Memory | HBM 显存热点温度 | `nvmlDeviceGetFieldValues` (fieldId=195) |
| Memory | BAR1 内存（PCIe BAR，用于 CUDA UVM） | `nvmlDeviceGetBAR1MemoryInfo` |
| I/O | PCIe 链路协商速率（当前+最大 Gen×Width） | `nvmlDeviceGetCurrPcieLinkGeneration/Width` + Max 版本 |
| I/O | NVLink 活动/最大链路数 | `nvmlDeviceGetNvLinkState` 遍历 |
| Reliability | ECC 模式（Enabled/Disabled）、不可纠正 ECC 错误计数 | `nvmlDeviceGetEccMode` / `nvmlDeviceGetTotalEccErrors` |
| Compute | 计算模式（Default/Exclusive/Prohibited） | `nvmlDeviceGetComputeMode` |
| Device | CUDA Compute Capability（如 "8.9"） | `nvmlDeviceGetCudaComputeCapability` |
| Device | Max SM Clock (MHz) → 前端 Device 属性条 | `nvmlDeviceGetMaxClockInfo(NVML_CLOCK_SM)` |
| Device | VBIOS 固件版本字符串 → 前端 Device 属性条 | `nvmlDeviceGetVbiosVersion` |
| Device | GPU Brand（GeForce/Quadro/Tesla/NVIDIA RTX/…） | `nvmlDeviceGetBrand` + `brandToString()` |
| Device | GPU 架构名（Kepler/Maxwell/Pascal/Volta/Turing/Ampere/Ada/Hopper/Blackwell） | `nvmlDeviceGetArchitecture` + `archToString()` |
| Device | SM 数量（优先 CUDA Driver API，回退 NVML `nvmlDeviceGetAttributes`） | `cuDeviceGetAttribute` / `nvmlDeviceGetAttributes` |
| Device | CUDA Cores（优先 NVML `GetNumGpuCores`，失败则 SM × coresPerSM CC 查表回退） | `nvmlDeviceGetNumGpuCores` / `coresPerSM()` 查表计算 |
| Device | L2 Cache 大小 (KB)、Shared Memory/Block (KB)、Shared Memory/SM (KB) | `cuDeviceGetAttribute` (CUDA Driver API) |
| Device | Register/Block、Max Threads/Block、Max Threads/SM、Warp Size | `cuDeviceGetAttribute` (CUDA Driver API) |
| Device | Concurrent Kernels、Copy Engines、Compute Preemption | `cuDeviceGetAttribute` (CUDA Driver API) |
| Device | Cooperative Launch、Multi-Device Coop、Managed Memory、UVA | `cuDeviceGetAttribute` (CUDA Driver API) |
| Device | Integrated GPU、Kernel Execution Timeout | `cuDeviceGetAttribute` (CUDA Driver API) |

## 前端架构

数据流：`useWebSocket()` composable → `GpuDashboard` 通过 `watch(data, ...)` 维护 `historyData` → 各子组件接收 props。

- `GpuDashboard.vue` — 主面板：Header（驱动版本/CUDA 版本/连接状态）+ 系统信息区（CPU 环形仪表盘 + 内存仪表盘 + 每核心柱状图）+ GPU 卡片列表；用 `historyData` ref 为每个 GPU 累积最多 3600 个历史点
- `GpuCard.vue` — GPU 卡片：摘要行 + Device Specs 三栏网格（Compute/Memory Hierarchy/Capabilities：架构标签、CC、SMs、CUDA Cores、Max Clock、Warp Size、Thread 限制、L2 Cache、Shared Memory、Register、布尔特性标志）+ 可展开四栏高级指标区（Performance/Memory/I/O/Reliability）+ 进程表。不可用的指标显示 `-`（通过 `dashNum`/`dashFmt`/`boolIcon` 辅助函数），让用户明确知道哪些数据未采集
- `CircularGauge.vue` — 纯 SVG 环形仪表盘，`stroke-dasharray`/`stroke-dashoffset` 驱动进度弧，颜色按阈值（≥90 红/≥60 黄/<60 绿）
- `GpuLineChart.vue` — Chart.js 折线图（GPU%、Mem%、Temp°C、Mem Temp°C 四数据集，显存温度默认隐藏），watch history 更新时用 `chart.update('none')` 跳过动画
- `ProcessTable.vue` — GPU 进程列表，按显存占用降序排列，类型标签（C=蓝色/G=紫色/C+G=青色）
- `useWebSocket.js` — 自动连接 + 指数退避重连（1s→10s 上限），暴露 `data`/`connected`/`error` ref

## 配置

`config.yaml` 包含 `server`（host/port）、`monitor`（poll_interval_seconds/history_size）、`log`（level）三段。优先级：`PORT` 环境变量 > config.yaml > `config.Default` 硬编码默认值。

## API

- `GET /api/gpus` — 当前 GPU + 系统 JSON 快照（与 WebSocket 推送结构相同）
- `GET /ws` — WebSocket，每秒推送 `Snapshot` JSON
