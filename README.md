# nvtop-web

基于 Go + Vue 3 的 NVIDIA GPU 实时监控 Web 应用，浏览器打开即用。

## 功能

### GPU 监控
- GPU 利用率、显存使用率（环形仪表盘 + 颜色分级）
- 温度、功耗、风扇转速、核心/显存频率
- PCIe 吞吐量、编码器/解码器利用率
- 60 秒实时历史曲线（Chart.js）
- 进程列表：PID、名称、显存占用、类型标签

### 系统监控
- CPU 整体利用率（环形仪表盘）+ 每核心柱状图
- 系统内存使用率（环形仪表盘）

### 通用
- WebSocket 实时推送（每秒）
- 断线自动重连
- 响应式布局（适配桌面/平板）
- 白色主题界面

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go 1.22 + gorilla/websocket |
| GPU 采集 | NVIDIA NVML (CGO) |
| 前端 | Vue 3 (Composition API) + Vite |
| 图表 | Chart.js + vue-chartjs |
| 样式 | Tailwind CSS |
| 部署 | 单一二进制文件（Go embed 嵌入前端） |

## 环境要求

- Linux（NVML 仅支持 Linux）
- NVIDIA GPU + 驱动（任意支持 NVML 的版本）
- Node.js 18+（仅构建时需要）
- Go 1.22+（仅构建时需要）

服务端可在无 GPU 的机器上运行（仅显示系统监控）。

## 快速开始

### 一键构建

```bash
./build.sh
```

构建产物为 `nvtop-server`（约 8MB 单一二进制文件）。

### 启动

```bash
./nvtop-server
```

浏览器访问 `http://localhost:8080`。

## 配置

通过 `config.yaml` 配置服务参数：

```yaml
server:
  host: "0.0.0.0"   # 监听地址
  port: 8080         # 监听端口

monitor:
  poll_interval_seconds: 1   # GPU 数据采集间隔（秒）
  history_size: 60           # 前端图表历史数据点数

log:
  level: "info"      # 日志级别: debug / info / warn / error
```

可通过命令行参数和环境变量覆盖：

```bash
# 指定配置文件
./nvtop-server --config /path/to/custom.yaml

# 环境变量覆盖端口（优先级最高）
PORT=9090 ./nvtop-server
```

优先级：`PORT` 环境变量 > `config.yaml` > 默认值

## 项目结构

```
nvtop/
├── backend/
│   ├── main.go              # 入口：路由、WebSocket、静态文件
│   ├── config/config.go     # YAML 配置解析
│   ├── gpu/
│   │   ├── types.go         # 数据结构（GPUInfo / SystemInfo / Snapshot）
│   │   ├── nvml.go          # NVML CGO 封装：GPU 数据采集
│   │   └── system.go        # CPU/内存采集（/proc/stat、/proc/meminfo）
│   ├── api/handler.go       # REST API：GET /api/gpus
│   └── ws/hub.go            # WebSocket Hub：连接管理 + 每秒广播
├── frontend/
│   └── src/
│       ├── App.vue
│       ├── components/
│       │   ├── GpuDashboard.vue    # 主面板：系统监控 + GPU 列表
│       │   ├── GpuCard.vue         # GPU 卡片：环形仪表盘 + 详情
│       │   ├── CircularGauge.vue   # SVG 环形仪表盘组件
│       │   ├── GpuLineChart.vue    # 60 秒实时曲线
│       │   └── ProcessTable.vue    # GPU 进程表
│       └── composables/
│           └── useWebSocket.js     # WebSocket 客户端（自动重连）
├── config.yaml              # 默认配置文件
├── build.sh                 # 一键构建脚本
└── README.md
```

## API

### GET /api/gpus

返回当前 GPU 和系统信息的 JSON 快照。

响应示例：

```json
{
  "timestamp": 1779554998,
  "driver_version": "580.142",
  "cuda_version": "13.0",
  "gpus": [{
    "index": 0,
    "name": "NVIDIA GeForce RTX 3090",
    "utilization_gpu": 96,
    "utilization_memory": 85,
    "memory_total_mb": 24576,
    "memory_used_mb": 22571,
    "temperature_c": 68,
    "power_w": 325,
    "power_limit_w": 350,
    "fan_speed": 70,
    "clock_core_mhz": 1860,
    "clock_memory_mhz": 9751,
    "pcie_rx_mbps": 120,
    "pcie_tx_mbps": 80,
    "encoder_util": 0,
    "decoder_util": 0,
    "processes": [{
      "pid": 230050,
      "name": "./llama-server",
      "memory_used_mb": 22556,
      "type": "C"
    }]
  }],
  "system": {
    "cpu_usage_percent": 7.14,
    "cpu_per_core_percent": [50, 100, 0, ...],
    "memory_total_mb": 65536,
    "memory_used_mb": 15482,
    "memory_usage_percent": 23.6
  }
}
```

### GET /ws

WebSocket 端点，每秒推送与 `/api/gpus` 相同结构的 JSON 消息。

## 开发

### 后端

```bash
cd backend
go run .
```

### 前端（开发模式，支持热更新）

```bash
cd frontend
npm install
npm run dev
```

开发模式下 Vite 会将 `/ws` 和 `/api` 代理到后端（需先启动后端）。
