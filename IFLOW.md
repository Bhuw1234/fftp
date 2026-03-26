# DEparrow - 全球虚拟机 (Global Virtual Machine)

## 项目愿景

**DEparrow** 是一个**全球虚拟机**平台，让 AI Agent 能够自主购买算力来运行自己。这是一个革命性的概念：AI 智能体拥有自己的钱包，通过提供服务赚取积分，然后用积分购买计算资源来维持自身运行。

```
┌─────────────────────────────────────────────────────────────────────────┐
│                     🌐 全球虚拟机 - Global Virtual Machine               │
│                                                                         │
│        "AI Agents Buy Compute to Run Themselves"                        │
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                   AI AGENTS (Autonomous)                        │   │
│  │                                                                 │   │
│  │   💰 Has Wallet    🧠 Runs Itself    💵 Pays for Compute        │   │
│  │                                                                 │   │
│  │   • Agent earns credits → providing services                   │   │
│  │   • Agent spends credits → buys compute to run itself          │   │
│  │   • Self-sustaining lifecycle                                  │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              DEPARROW - The Marketplace (OS)                    │   │
│  │                                                                 │   │
│  │   "Install Once → Become a Node Automatically"                  │   │
│  │                                                                 │   │
│  │   ┌──────────┐  ┌──────────┐  ┌──────────────────┐             │   │
│  │   │ Credits  │  │ Compute  │  │ Agent Registry   │             │   │
│  │   │ Economy  │  │ Market   │  │ & Discovery      │             │   │
│  │   └──────────┘  └──────────┘  └──────────────────┘             │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                              │                                          │
│                              ▼                                          │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              BACALHAU - Compute Layer (Engine)                   │   │
│  │                                                                 │   │
│  │   PicoClaw nodes ($10, 10MB RAM) distributed across the planet  │   │
│  │   Anyone contributes compute → earns credits                    │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 核心理念

### 自动贡献计算

```
┌────────────────────────────────────────────────────────────┐
│                                                            │
│   1. 用户安装 DEparrow OS (ISO 或软件包)                   │
│      └─▶ 后台服务自动启动                                  │
│                                                            │
│   2. 系统检测空闲资源                                       │
│      └─▶ CPU 空闲？GPU 空闲？内存可用？                    │
│                                                            │
│   3. 自动贡献给网络                                         │
│      └─▶ 运行其他 Agent 的计算任务                         │
│                                                            │
│   4. 被动赚取积分                                           │
│      └─▶ 用户钱包自动增加积分                              │
│                                                            │
│   5. 需要时消费积分                                         │
│      └─▶ 运行自己的 AI Agent 或任务                        │
│                                                            │
└────────────────────────────────────────────────────────────┘
```

### 启动即节点 (Bootable ISO)

```
┌───────────────────────────────────────────────────────────────┐
│                                                               │
│   "Burn to USB → Boot → You're a Node. Done."                │
│                                                               │
│   Boot Process:                                               │
│                                                               │
│   [BIOS/UEFI]                                                │
│        │                                                      │
│        ▼                                                      │
│   [DEparrow Kernel]                                          │
│        │                                                      │
│        ▼                                                      │
│   [Auto-Network Join] ──▶ Connects to DEparrow network       │
│        │                                                      │
│        ▼                                                      │
│   [Compute Node Active] ──▶ Starts earning credits           │
│        │                                                      │
│        ▼                                                      │
│   [Dashboard on :80] ──▶ User sees their earnings            │
│                                                               │
└───────────────────────────────────────────────────────────────┘
```

### ISO 镜像 - 生产就绪 (2026-03-26)

**可用 ISO 文件:**

| ISO 文件 | 大小 | 说明 |
|----------|------|------|
| `deparrow-1.0.0.iso` | 28MB | 精简版，仅 EFI 启动 |
| `deparrow-alpine.iso` | 49MB | Alpine Linux 完整版 |
| `deparrow-autojoin.iso` | 99MB | 6阶段自动加入网络 |
| `deparrow-autojoin-local.iso` | 76MB | 本地测试版 (QEMU) |
| `deparrow-hybrid.iso` | 112MB | BIOS + EFI 双启动 |
| `deparrow-os-1.0.0.iso` | 81MB | **DEparrow 品牌版 (推荐)** |

**当前引导服务器:** `34.180.51.11:8080` (GCP 运行中)

**DEparrow 品牌版 ISO 特性 (推荐):**
- ✅ 纯 DEparrow 品牌 (无上游名称显示)
- ✅ `deparrow` CLI 命令
- ✅ Alpine minirootfs 3.20 基础 (稳定可靠)
- ✅ BIOS + EFI 双启动支持
- ✅ 6阶段自动加入网络
- ✅ QEMU 本地测试模式 (`--local` 标志)

**Hybrid ISO 特性 (推荐):**
- ✅ BIOS 传统启动 (GRUB i386-pc 模块)
- ✅ EFI/UEFI 启动 (GPT 分区)
- ✅ WiFi 支持 (Intel/Realtek/Atheros/MediaTek)
- ✅ 6阶段自动加入网络

**硬件要求 (支持旧电脑):**

| 资源 | 最低配置 | 推荐配置 |
|------|----------|----------|
| RAM | 512MB | 1GB+ |
| CPU | 任意 x86_64 | 多核 |
| 存储 | 无需硬盘 | 从 RAM 运行 |
| 网络 | 可选 | 以太网/WiFi |

**启动测试命令 (QEMU):**

```bash
# EFI 模式启动
qemu-system-x86_64 -m 1G \
  -drive if=pflash,format=raw,readonly=on,file=/usr/share/OVMF/OVMF_CODE_4M.fd \
  -drive if=pflash,format=raw,file=/tmp/ovmf_vars.fd \
  -cdrom deparrow/bootable/output/deparrow-alpine.iso \
  -nographic -serial mon:stdio
```

**启动特性:**

- ✅ 显示 DEparrow ASCII 横幅
- ✅ 设置主机名 `deparrow-node`
- ✅ 显示 CPU/内存信息
- ✅ 配置网络 (DHCP)
- ✅ 自动启动 bacalhau 计算节点
- ✅ 进入 BusyBox shell
- ✅ BIOS + EFI 双启动支持
- ✅ WiFi 自动连接 (内核参数: wifi.ssid=X wifi.password=X)
- ✅ 6阶段自动加入网络

**6阶段自动加入流程:**

```
[Phase 1] DHCP → 获取 IP 地址
[Phase 2] 生成节点身份 → UUID + RSA 密钥对
[Phase 3] 发现引导服务器 → bootstrap.deparrow.net:8080
[Phase 4] 注册到 Meta-OS → POST /api/v1/nodes/register
[Phase 5] 获取编排器配置 → orchestrator IP + NATS 地址
[Phase 6] 启动 bacalhau → bacalhau serve --compute
```

**待部署基础设施:**

| 功能 | 状态 | 说明 |
|------|------|------|
| 引导服务器 | 代码就绪 | 需部署 bootstrap.deparrow.net:8080 |
| 编排器节点 | 代码就绪 | 需运行 bacalhau orchestrator |
| 积分赚取 | 代码就绪 | 需连接到运行中的网络 |

---

## Unified Global VM Concept

### The Vision

**One Infinite Computer, Not a Rental Marketplace**

DEparrow is fundamentally different from existing decentralized compute platforms:

| Platform | Model | User Experience |
|----------|-------|-----------------|
| **Akash** | Rental marketplace | User selects providers, bids on resources |
| **Golem** | Rental marketplace | User picks nodes, negotiates pricing |
| **iExec** | Rental marketplace | User chooses workers, market-based pricing |
| **DEparrow** | **Unified Global VM** | User sees ONE computer |

### Comparison: Rental Marketplace vs Global VM

| Aspect | Rental Marketplace | DEparrow Global VM |
|--------|-------------------|-------------------|
| Node selection | User picks manually | Automatic (transparent) |
| Pricing | Bidding/market volatility | Fixed credits/job |
| Abstraction | Low (see individual nodes) | High (one unified VM) |
| User experience | Complex, requires expertise | Simple, like local machine |
| Resource view | Fragmented, heterogeneous | Unified, infinite |
| Failure handling | User manages retries | Automatic rescheduling |

### The Key Insight

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│   User's View: "One Infinite Computer"                         │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                                                         │  │
│   │   CPU: Infinite cores available                        │  │
│   │   GPU: Infinite CUDA cores available                   │  │
│   │   RAM: Infinite memory available                       │  │
│   │   Storage: Infinite disk space available               │  │
│   │                                                         │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│   Reality (Scheduler's Magic):                                 │
│                                                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │                                                         │  │
│   │   10,000 phones     →  ~100 ARM cores                  │  │
│   │   1,000 gaming PCs  →  ~5,000 GPU cores                │  │
│   │   100 servers       →  ~10,000 CPU cores               │  │
│   │   500 IoT devices   →  ~50 embedded cores              │  │
│   │                                                         │  │
│   │   Total: 15,000+ heterogeneous nodes → ONE VM          │  │
│   │                                                         │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                 │
│   The Scheduler presents this chaos as ONE seamless computer   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### User Experience Example

```bash
# Traditional Rental Marketplace (Akash/Golem/iExec)
$ akash deploy --provider provider-123 --cpu 4 --memory 8GB --price 0.05
$ akash monitor provider-123
$ akash redeploy --provider provider-456  # If provider fails
# User manages everything manually

# DEparrow Global VM
$ deparrow run train-model.py
✓ Job submitted to Global VM
✓ Running... (automatic node selection)
✓ Completed in 2h 34m
✓ Cost: 15 DPC

# User doesn't know or care WHERE it ran
# User doesn't know or care on WHICH nodes
# User doesn't manage failures or retries
# It just works, like running locally
```

### Why This Matters

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│   "The best infrastructure is invisible."                      │
│                                                                 │
│   Traditional Cloud:                                           │
│   - User thinks about regions, zones, instance types          │
│   - User manages scaling, failover, costs                     │
│   - Complexity grows with scale                               │
│                                                                 │
│   DEparrow Global VM:                                          │
│   - User thinks about ONE thing: "run my job"                 │
│   - System handles everything else automatically              │
│   - Simplicity stays constant regardless of scale             │
│                                                                 │
│   This is the difference between:                              │
│   - Renting servers vs. Having infinite compute               │
│   - Managing infrastructure vs. Just using it                 │
│   - Being a sysadmin vs. Being a developer                    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Implementation Progress

### Phase 1: Global Capacity Aggregation COMPLETE

| File | Status | Purpose |
|------|--------|---------|
| `pkg/globalvm/capacity_aggregator.go` | Done | Aggregates capacity from all nodes |
| `pkg/globalvm/capacity_aggregator_test.go` | Done | Tests (4 passing) |

**Key Features Implemented:**
- `GlobalResources` struct - Total/available CPU, Memory, Disk, GPU
- `CapacityAggregator` - Implements `GlobalCapacityProvider` interface
- `Subscribe()` - Real-time capacity updates via channel
- GPU breakdown by vendor (NVIDIA, AMD/ATI, Intel)
- `Summary()` - Human-readable "Infinite" view for large clusters

### Phase 2: Unified Job Submission COMPLETE

| File | Status | Purpose |
|------|--------|---------|
| `pkg/globalvm/endpoint.go` | Done | Global VM job submission API |
| `pkg/globalvm/scheduler.go` | Done | Intelligent node selection |
| `pkg/globalvm/endpoint_test.go` | Done | Endpoint tests |
| `pkg/globalvm/scheduler_test.go` | Done | Scheduler tests |

**Key Features Implemented:**
- `GlobalVMEndpoint` interface - SubmitJob, GetJobStatus, ScaleJob, CancelJob
- `GlobalScheduler` - Wraps existing NodeSelector with global optimizations
- `SchedulingOptions` - Region spread, latency constraints, cost preferences
- `RegionRanker` - Scores regions by latency and cost
- Integration with existing orchestrator/selection system

### Phase 3: Capability Detection COMPLETE

| File | Status | Purpose |
|------|--------|---------|
| `pkg/globalvm/capability/detector.go` | Done | Main detector interface |
| `pkg/globalvm/capability/gpu_detector.go` | Done | GPU detection (NVIDIA/AMD/Intel) |
| `pkg/globalvm/capability/engine_detector.go` | Done | Execution engine detection |
| `pkg/globalvm/capability/benchmark.go` | Done | Performance benchmarks |
| `pkg/globalvm/capability/detector_test.go` | Done | Tests (24 passing) |

**Key Features Implemented:**
- `CapabilityDetector` interface - DetectAll, Benchmark, Refresh
- `NodeCapabilities` struct - Engines, GPUs, Storage, Network
- `GPUCapability` - Index, Name, Vendor, Memory, CUDA/ROCm versions
- `EngineCapability` - Type, Version, Available, Features
- `CapabilityBenchmarks` - CPU/Memory/Disk/GPU/Network scores (0-1000)
- `HasGPUVendor()` / `HasEngine()` / `TotalGPUMemory()` helpers
- `CapabilityScore()` - Overall node capability score

### Phase 4: Geographic Scheduling COMPLETE

| File | Status | Purpose |
|------|--------|---------|
| `pkg/globalvm/latency_matrix.go` | Done | Latency tracking between regions |
| `pkg/globalvm/location.go` | Done | Geographic location detection |
| `pkg/globalvm/geo_ranker.go` | Done | Geographic-aware node ranking |
| `pkg/globalvm/latency_matrix_test.go` | Done | Tests (25 passing) |

**Key Features Implemented:**
- `LatencyMatrix` interface - GetLatency, UpdateLatency, GetNearestNodes
- `GeoRanker` - Ranks nodes by latency, region, continent
- `LocationDetector` - Detects region from cloud metadata, GeoIP
- `EstimatedLatency()` - Predefined inter-region latencies
- Support for AWS/GCP/Azure metadata endpoints
- `RegionToContinent()` mapping for broader grouping
- Preferred/excluded regions via job labels/constraints
- Max latency constraints for latency-sensitive jobs

### Phase 5: Integration & Testing COMPLETE

| File | Status | Purpose |
|------|--------|---------|
| `pkg/globalvm/integration_test.go` | Done | End-to-end integration tests |

**Test Scenarios Covered:**
1. **Full Job Lifecycle** - Submit → Schedule → Execute → Complete
2. **Capacity-Aware Scheduling** - Jobs only scheduled on nodes with sufficient capacity
3. **Capability Matching** - GPU jobs only on GPU nodes, CUDA jobs only on NVIDIA nodes
4. **Geographic Distribution** - Jobs spread across regions
5. **Scale Job** - Increase/decrease job count while running
6. **Failover** - Job rescheduled when node goes down
7. **Full Stack** - Complete integration with all phases
8. **Concurrent Operations** - Multiple simultaneous job submissions
9. **Edge Cases** - Empty cluster, all nodes down, single node
10. **Capability Detection** - Integration with detector
11. **Latency Matrix** - Realistic multi-region latency setup
12. **Subscription Updates** - Real-time capacity notifications

---

## Global VM Implementation Summary

**Total Files:** 18 Go files (main + tests)
**Status:** 100% COMPLETE

---

## 技术栈

### 核心语言和运行时

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | 1.24.0 | 核心语言 (Bacalhau) |
| Go | 1.25.7 | PicoClaw 模块 |
| Python | 3.12.3 | SDK 和工具 |
| Node.js | 18+ | WebUI 开发 |
| TypeScript | 5.x | WebUI 类型系统 |

### 后端依赖 (Bacalhau Core)

| 库 | 版本 | 用途 |
|----|------|------|
| NATS Server | v2.11.6 | 分布式消息传递 |
| NATS Go | v1.43.0 | NATS 客户端 |
| libp2p | v0.41.1 | P2P 网络 |
| Docker | v27.1.1 | 容器执行引擎 |
| wazero | v1.9.0 | WebAssembly 运行时 |
| IPFS (kubo) | v0.35.0 | 分布式存储 |
| IPFS (boxo) | v0.32.0 | IPFS 组件 |
| Open Policy Agent | v0.60.0 | 策略引擎 |
| AWS SDK v2 | v1.36.5 | S3 存储集成 |
| OpenTelemetry | v1.37.0 | 可观测性 |
| zerolog | v1.34.0 | 结构化日志 |
| Cobra | v1.9.1 | CLI 框架 |
| Echo | v4.13.4 | HTTP 服务器 |
| JWT (golang-jwt) | v5.2.2 | 认证 |
| libp2p-kad-dht | v0.33.1 | DHT 路由 |

### PicoClaw 依赖

| 库 | 版本 | 用途 |
|----|------|------|
| Anthropic SDK | v1.22.1 | Claude API |
| OpenAI SDK | v3.22.0 | GPT API |
| DiscordGo | v0.29.0 | Discord 集成 |
| Telego | v1.6.0 | Telegram 集成 |
| Slack SDK | v0.17.3 | Slack 集成 |
| DingTalk SDK | v0.9.1 | 钉钉集成 |
| Lark SDK | v3.5.3 | 飞书集成 |
| QQ Bot | v0.2.1 | QQ 集成 |

### 前端依赖 (WebUI - Next.js)

| 库 | 版本 | 用途 |
|----|------|------|
| Next.js | 15.2.4 | React 框架 |
| React | 18 | UI 库 |
| Radix UI | 1.x | 组件库 |
| Tailwind CSS | 3.4.1 | 样式系统 |
| Lucide React | 0.438.0 | 图标库 |
| next-themes | 0.3.0 | 主题切换 |
| Vitest | 1.3.1 | 测试框架 |

### GUI 层依赖 (Vite + React)

| 库 | 版本 | 用途 |
|----|------|------|
| Vite | 5.1.4 | 构建工具 |
| React | 18.2.0 | UI 库 |
| React Router | 6.20.0 | 路由 |
| Recharts | 2.10.3 | 图表库 |
| React Query | 3.39.3 | 数据获取 |
| React Hook Form | 7.48.2 | 表单管理 |
| Zod | 3.22.4 | 数据验证 |
| Framer Motion | 11.0.0 | 动画 |

### 构建工具

| 工具 | 版本 | 用途 |
|------|------|------|
| Earthly | 0.8.3 | 容器化构建 |
| golangci-lint | 1.64.2 | Go 代码检查 |
| pnpm / yarn | 9.0.6 / 4.4.1 | Node.js 包管理 |
| pre-commit | 3.6.0 | Git 钩子 |
| Poetry | (latest) | Python 包管理 |

---

## 项目结构

```
.
├── main.go                    # 主入口点
├── go.mod                     # Go 模块定义 (go 1.24.0)
├── go.work                    # Go workspace 配置
├── Makefile                   # 50+ 构建目标
│
├── cmd/                       # 命令行接口
│   └── cli/                   # CLI 命令实现
│       ├── agent/             # Agent 命令
│       ├── auth/              # 认证命令
│       ├── config/            # 配置命令
│       ├── devstack/          # 开发栈
│       ├── docker/            # Docker 命令
│       ├── job/               # 作业管理
│       ├── node/              # 节点管理
│       ├── serve/             # 服务启动
│       └── wasm/              # WebAssembly 命令
│
├── pkg/                       # 核心库 (40 子目录)
│   ├── compute/               # 计算节点逻辑
│   ├── orchestrator/          # 编排器逻辑
│   ├── executor/              # 执行引擎
│   ├── nats/                  # NATS 集成
│   ├── authn/                 # 认证
│   ├── authz/                 # 授权 (OPA)
│   ├── models/                # 数据模型
│   ├── publicapi/             # 公共 API
│   ├── jobstore/              # 作业存储
│   ├── telemetry/             # 遥测
│   ├── globalvm/              # 全球虚拟机实现
│   │   ├── capability/        # 能力检测
│   │   ├── capacity_aggregator.go
│   │   ├── endpoint.go
│   │   ├── scheduler.go
│   │   ├── geo_ranker.go
│   │   ├── latency_matrix.go
│   │   └── location.go
│   └── ...                    # 更多模块
│
├── webui/                     # Web 界面 (Next.js 15)
│   ├── app/                   # Next.js App Router
│   ├── components/            # React 组件
│   │   ├── ui/                # UI 基础组件
│   │   ├── jobs/              # 作业组件
│   │   ├── nodes/             # 节点组件
│   │   └── layout/            # 布局组件
│   ├── hooks/                 # 自定义 Hooks
│   └── lib/                   # 工具库
│
├── python/                    # Python SDK
│   └── tests/                 # Python 测试
├── clients/                   # API 客户端
├── integration/               # 第三方集成
│   ├── airflow/               # Airflow 集成
│   └── flyte/                 # Flyte 集成
│
├── deparrow/                  # DEparrow 平台
│   ├── alpine-layer/          # Alpine Linux 基础层
│   ├── bacalhau-layer/        # Bacalhau 层
│   ├── bootable/              # 可启动镜像
│   │   ├── output/            # ISO 输出目录
│   │   ├── build-deparrow-iso.sh    # DEparrow 品牌版 ISO
│   │   ├── build-iso-v2.sh    # Auto-Join ISO
│   │   ├── build-hybrid-iso.sh    # BIOS+EFI 混合 ISO
│   │   ├── build-iso-fixed.sh # 修复版 ISO
│   │   ├── build-iso.sh       # ISO 构建脚本
│   │   ├── create-iso.sh      # ISO 创建脚本
│   │   ├── auto-install.sh    # 自动安装脚本
│   │   └── USER_GUIDE.md      # 用户指南
│   ├── gui-layer/             # GUI 用户界面层
│   │   └── src/pages/         # 8 页面组件
│   ├── metaos-layer/          # Meta-OS 控制平面层
│   │   ├── bootstrap-server.py # 引导服务器
│   │   └── requirements.txt   # Python 依赖
│   ├── k8s/                   # Kubernetes 部署配置
│   │   ├── base/              # 22 个基础清单
│   │   └── overlays/          # 环境配置
│   ├── config/                # 配置文件
│   ├── scripts/               # 部署脚本
│   ├── test-integration/      # 集成测试
│   ├── docker-compose.prod.yml # 生产环境配置
│   ├── .env.example           # 环境变量模板
│   └── start.sh               # 快速启动脚本
│
├── picoclaw/                  # PicoClaw 轻量级节点
│   ├── cmd/                   # CLI 命令
│   ├── pkg/deparrow/          # DEparrow 工具包
│   │   ├── client.go          # Meta-OS API 客户端
│   │   ├── types.go           # 类型定义
│   │   ├── job_tool.go        # 作业管理工具
│   │   ├── credit_tool.go     # 积分管理工具
│   │   ├── node_tool.go       # 节点管理工具
│   │   ├── wallet_tool.go     # 钱包管理工具
│   │   └── register.go        # 工具注册器
│   ├── pkg/agent/             # Agent 核心
│   ├── pkg/channels/          # 多渠道支持
│   └── pkg/providers/         # AI 提供者
│   # ($10 硬件, <10MB RAM, 1s 启动, Go 1.25.7)
│
├── docker/                    # Docker 镜像构建
├── test/                      # 测试脚本
├── test_integration/          # 集成测试
├── scripts/                   # 构建脚本
├── ops/                       # 运维脚本
└── docs/                      # 文档
```

---

## 四层架构

```
┌─────────────────────────────────────────────────────────┐
│                    GUI 用户界面层                         │
│         Dashboard | Jobs | Wallet | Nodes | Providers   │
├─────────────────────────────────────────────────────────┤
│                 Meta-OS 控制平面层                        │
│    引导服务 | 信用系统 | 作业准入 | JWT 认证 | AI 调度     │
├─────────────────────────────────────────────────────────┤
│                Alpine Linux 基础层 (ISO)                  │
│         轻量级 OS | 自动加入 | x86_64/arm64              │
├─────────────────────────────────────────────────────────┤
│               Bacalhau 执行网络层                         │
│    Docker | WebAssembly | NATS | libp2p | IPFS          │
└─────────────────────────────────────────────────────────┘
```

### 1. Alpine Linux 基础层 (ISO)
- **轻量级 OS**: 最小化系统开销
- **自动加入**: 节点启动后自动发现并加入网络
- **多架构支持**: x86_64 和 arm64
- **健康监控**: 实时系统检查
- **一键安装**: 烧录 ISO → 启动 → 完成

### 2. Meta-OS 控制平面层
- **引导服务**: DEparrow 专用引导节点 (`bootstrap-server.py`)
- **编排器注册**: 编排器节点发现和注册
- **信用支付系统**: 基于信用的作业提交控制
- **作业准入控制**: 支付验证后允许作业提交
- **JWT 认证**: 完整的身份验证和授权

### 3. GUI 用户界面层
- **Dashboard**: 网络统计和监控
- **Jobs**: 作业管理界面
- **Nodes**: 节点监控仪表板
- **Wallet**: 钱包和积分管理
- **Settings**: 用户配置
- **Login**: 认证界面
- **Agent**: AI Agent 控制台
- **Providers**: 提供者市场

### 4. Bacalhau 执行网络层
- **Docker 执行**: 容器化作业执行
- **WebAssembly**: wazero 沙箱执行
- **NATS 消息传递**: 分布式消息系统
- **libp2p P2P**: 去中心化网络通信
- **IPFS 存储**: 分布式文件存储
- **Kademlia DHT**: 节点发现和路由

---

## PicoClaw - 超轻量 AI 助手

PicoClaw 是 DEparrow 生态中的超轻量级 AI 助手节点，可在 $10 硬件上运行。已深度集成 DEparrow 网络。

### 特性

| 特性 | 指标 |
|------|------|
| 内存占用 | < 10MB RAM |
| 启动时间 | < 1秒 |
| 硬件成本 | 低至 $10 |
| 支持架构 | x86_64, ARM64, RISC-V |
| DEparrow 工具 | 14 个内置工具 |
| Go 版本 | 1.25.7 |

### DEparrow 集成

PicoClaw 已内置 DEparrow 工具包 (`pkg/deparrow/`)，可直接与 DEparrow 网络交互：

```bash
# 配置 DEparrow 连接 (~/.picoclaw/config.json)
{
  "deparrow": {
    "enabled": true,
    "api_url": "http://localhost:8080",
    "jwt_token": "your-jwt-token"
  }
}
```

### DEparrow 工具示例

```bash
# 提交作业
picoclaw agent -m "Submit a Python job to train my model"

# 检查积分
picoclaw agent -m "What's my credit balance?"

# 查看节点
picoclaw agent -m "List all available compute nodes"
```

### 安装

```bash
# 从源码构建
git clone https://github.com/sipeed/picoclaw.git
cd picoclaw
make deps
make build

# Docker Compose
docker compose --profile gateway up -d
```

### 多渠道支持

| 渠道 | 难度 | SDK |
|------|------|-----|
| Telegram | 简单 (仅需 token) | Telego v1.6.0 |
| Discord | 简单 (bot token + intents) | DiscordGo v0.29.0 |
| Slack | 简单 (bot token) | Slack SDK v0.17.3 |
| QQ | 简单 (AppID + AppSecret) | QQ Bot v0.2.1 |
| DingTalk | 中等 (应用凭证) | DingTalk SDK v0.9.1 |
| Lark/飞书 | 中等 (凭证 + webhook) | Lark SDK v3.5.3 |

---

## 快速开始

### 方式一：开发模式

```bash
cd deparrow
./start.sh dev

# 启动:
# - Meta-OS API: http://localhost:8080
# - GUI: http://localhost:5173
```

### 方式二：生产环境 (Docker Compose)

```bash
cd deparrow
cp .env.example .env
# 编辑 .env 设置安全密码
./start.sh prod

# 启动完整栈:
# - Meta-OS API: http://localhost:8080
# - GUI: http://localhost:3000
# - Prometheus: http://localhost:9090
# - Grafana: http://localhost:3001
```

### 方式三：Kubernetes

```bash
# 开发环境
kubectl apply -k deparrow/k8s/overlays/dev

# 生产环境
kubectl apply -k deparrow/k8s/overlays/production
```

---

## 开发环境

### 前提条件

```bash
golang      1.24.0+
nodejs      18+
python      3.12.3+
earthly     0.8.3
yarn        4.4.1+
```

### 快速开始

```bash
# 克隆仓库
git clone https://github.com/Bhuw1234/fftp.git
cd bacalhau

# 初始化
make init

# 构建
make build

# 启动开发栈
make devstack

# 运行测试
make test
```

---

## 构建命令

### Go 项目

```bash
make build          # 构建 Go 二进制
make test           # 运行测试
make unit-test      # 单元测试
make integration-test # 集成测试
make lint           # 代码检查
make devstack       # 启动开发栈
make security       # 安全检查
```

### WebUI

```bash
cd webui && yarn dev   # 开发模式
cd webui && yarn build # 生产构建
cd webui && yarn test  # 运行测试
```

### Docker 镜像

```bash
docker-compose -f deparrow/docker-compose.prod.yml up -d
```

### ISO 构建

```bash
cd deparrow/bootable
./build-deparrow-iso.sh           # DEparrow 品牌版 (推荐)
./build-deparrow-iso.sh --local   # QEMU 本地测试版
./build-deparrow-iso.sh --wifi    # 包含 WiFi 支持
./build-iso-v2.sh                 # Auto-Join ISO
./build-hybrid-iso.sh             # BIOS+EFI 混合 ISO
```

---

## 测试环境

### 本地测试

```bash
make unit-test       # Go 单元测试
make integration-test # Go 集成测试
cd webui && yarn test # WebUI 测试
cd python && pytest   # Python SDK 测试
cd picoclaw && go test ./pkg/deparrow/... # PicoClaw 测试
```

### Docker Compose 测试

```bash
cd deparrow
./start.sh dev   # 开发模式
./start.sh prod  # 生产模式
```

### 可启动 ISO 测试

**ISO 文件:** `deparrow/bootable/output/deparrow-hybrid.iso` (112MB，BIOS+EFI 双启动)

| 方式 | 命令 |
|------|------|
| QEMU (EFI) | 见上方启动测试命令 |
| QEMU (BIOS) | `qemu-system-x86_64 -m 1G -cdrom deparrow-hybrid.iso -nographic` |
| VirtualBox | 创建 VM → 挂载 ISO → 启动 (BIOS 或 EFI) |
| USB 启动 | `dd if=deparrow-hybrid.iso of=/dev/sdb bs=4M status=progress` |
| 真机测试 | 烧录到 USB → BIOS/UEFI 均可启动 |

---

## 网络架构

### 节点类型

| 节点类型 | 命令 | 说明 |
|----------|------|------|
| 编排器节点 | `bacalhau serve --orchestrator` | 端口 4222/1234 |
| 计算节点 | `bacalhau serve --compute` | 自动加入 |
| 混合节点 | `bacalhau serve` | 编排+计算 |

### 生产环境服务

| 服务 | 端口 | 说明 |
|------|------|------|
| Meta-OS API | 8080 | 控制平面 API |
| GUI | 3000 | Web 界面 |
| Bacalhau Orchestrator | 4222/1234 | 编排服务 |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| Prometheus | 9090 | 监控 |
| Grafana | 3001 | 可视化 |

---

## Kubernetes 部署

### 环境配置

| 环境 | 副本数 | 说明 |
|------|--------|------|
| dev | 1 | 开发环境 |
| staging | 2-3 | 预发布环境 |
| production | 3-20 | 生产环境，HA |

### K8s 资源清单 (22 个文件)

```
deparrow/k8s/base/
├── namespace.yaml          # 命名空间
├── configmap.yaml          # 配置映射
├── secrets.yaml            # 开发密钥
├── external-secret.yaml    # External Secrets CRD
├── rbac.yaml               # 角色权限
├── network-policy.yaml     # 网络策略
├── ingress.yaml            # 入口配置
├── metaos-deployment.yaml  # Meta-OS 部署
├── gui-deployment.yaml     # GUI 部署
├── orchestrator-deployment.yaml
├── compute-daemonset.yaml
├── postgres-*.yaml         # PostgreSQL
├── redis-*.yaml            # Redis
├── prometheus.yaml         # 监控
├── grafana.yaml            # 可视化
└── hpa.yaml                # 自动扩缩容
```

### 部署命令

```bash
kubectl apply -k deparrow/k8s/overlays/dev
kubectl apply -k deparrow/k8s/overlays/production
```

---

## API 端点

### Core Endpoints

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/v1/health` | GET | 健康检查 |
| `/api/v1/auth/login` | POST | 获取 JWT Token |
| `/api/v1/nodes` | GET | 列出节点 |
| `/api/v1/credits` | GET | 获取积分余额 |
| `/api/v1/jobs` | GET | 列出作业 |
| `/api/v1/jobs` | POST | 提交作业 |

### Available DEparrow Tools (14 Tools)

| 工具名称 | 说明 |
|----------|------|
| `deparrow_submit_job` | 提交计算作业 |
| `deparrow_job_status` | 检查作业状态 |
| `deparrow_list_jobs` | 列出用户作业 |
| `deparrow_cancel_job` | 取消作业 |
| `deparrow_credits` | 查看积分余额 |
| `deparrow_how_to_earn` | 赚取积分指南 |
| `deparrow_network` | 网络统计信息 |
| `deparrow_leaderboard` | 贡献排行榜 |
| `deparrow_nodes` | 列出计算节点 |
| `deparrow_contribution` | 节点贡献统计 |
| `deparrow_orchestrators` | 编排器列表 |
| `deparrow_wallet` | 钱包余额 |
| `deparrow_transfer` | 积分转账 |
| `deparrow_health` | 连接健康检查 |

---

## 环境变量

| 变量 | 说明 |
|------|------|
| `DEPARROW_SECRET_KEY` | JWT 签名密钥 (必需) |
| `POSTGRES_PASSWORD` | PostgreSQL 密码 (必需) |
| `GRAFANA_ADMIN_PASSWORD` | Grafana 管理员密码 (必需) |
| `DATABASE_URL` | PostgreSQL 连接 |
| `REDIS_URL` | Redis 连接 |
| `LOG_LEVEL` | 日志级别 |

---

## 项目状态

```
┌─────────────────────────────────────────────────────────────────┐
│                    DEPARROW PROJECT STATUS                      │
│                        ~95% COMPLETE                            │
└─────────────────────────────────────────────────────────────────┘

CORE PLATFORM: 100% ████████████████████
├── Bacalhau Engine        90% ███████████████████░
├── Global VM (5 phases)  100% ████████████████████
├── Auto-Join ISO         100% ████████████████████
├── BIOS + EFI Boot       100% ████████████████████
├── WiFi Support          100% ████████████████████
├── Bootstrap Server      100% ████████████████████
├── Security Hardened     100% ████████████████████
└── DEparrow Branding     100% ████████████████████ (Zero-Risk)

USER INTERFACE: 100% ████████████████████
├── WebUI (Next.js)       100% ████████████████████
├── GUI Layer (Vite)      100% ████████████████████
└── PicoClaw Integration  100% ████████████████████

DEPLOYMENT: 100% ████████████████████ ✅ GCP LIVE
├── Kubernetes Manifests  100% ████████████████████
├── Docker Compose        100% ████████████████████
└── Live Infrastructure  100% ████████████████████ (GCP: 34.180.51.11)

DPC TOKEN: 20% ████░░░░░░░░░░░░░░░░░
├── Token Design Doc      100% ████████████████████
├── Smart Contracts         0% ░░░░░░░░░░░░░░░░░░░░
├── Blockchain (Cosmos)     0% ░░░░░░░░░░░░░░░░░░░░
└── AI Agent Wallets        0% ░░░░░░░░░░░░░░░░░░░░

MOBILE APP: 0% ░░░░░░░░░░░░░░░░░░░░
├── Social App Concept    100% ████████████████████ (Design)
├── React Native/Flutter    0% ░░░░░░░░░░░░░░░░░░░░
└── Embedded Compute         0% ░░░░░░░░░░░░░░░░░░░░
```

### GCP 部署状态 (2026-03-26) ✅ LIVE

| 组件 | 状态 | 端点 |
|------|------|------|
| **Bootstrap Server** | ✅ 运行中 | http://34.180.51.11:8080 |
| **Bacalhau API** | ✅ 运行中 | http://34.180.51.11:1234 |
| **Compute Node** | ✅ 已连接 | n-2b1dee2e |
| **网络贡献** | 100% | 3.2 CPU, 13GB RAM |

**测试命令:**
```bash
# 检查健康状态
curl http://34.180.51.11:8080/api/v1/health

# 提交作业
BACALHAU_API_HOST=34.180.51.11 ./bacalhau docker run ubuntu echo "Hello"

# 查看节点
BACALHAU_API_HOST=34.180.51.11 ./bacalhau node list
```

### 最近完成 (2026-03-26)

| 里程碑 | 状态 | 说明 |
|--------|------|------|
| Alpine Linux ISO | ✅ 完成 | a592619da |
| 6阶段自动加入网络 | ✅ 完成 | 255b730d8 |
| BIOS + EFI 混合启动 | ✅ 完成 | f21251552 |
| WiFi 支持 | ✅ 完成 | f21251552 |
| DPC Token 设计文档 | ✅ 完成 | f21251552 |
| Bootstrap 安全加固 | ✅ 完成 | f21251552 |
| **DEparrow 品牌化 (Zero-Risk)** | ✅ 完成 | 2026-03-22 |
| - CLI Wrapper (bin/deparrow) | ✅ | deparrow 命令封装 bacalhau |
| - ASCII Banner | ✅ | DEparrow 专属启动横幅 |
| - GRUB 菜单 | ✅ | "DEparrow OS" 菜单项 |
| - Alpine Minirootfs | ✅ | 官方 Alpine 3.20 基础 |
| - 品牌版 ISO | ✅ | deparrow-os-1.0.0.iso (81MB) |

### ISO 功能状态

| 功能 | 状态 | 说明 |
|------|------|------|
| EFI 启动 | ✅ 完成 | GRUB EFI 引导 |
| BIOS 启动 | ✅ 完成 | GRUB i386-pc 模块 |
| WiFi 支持 | ✅ 完成 | Intel/Realtek/Atheros/MediaTek |
| 网络自动加入 | ✅ 代码完成 | 需部署引导服务器 |
| 编排器自动连接 | ✅ 代码完成 | 需运行编排器 |
| 积分系统集成 | ✅ 代码完成 | 需连接运行网络 |

### 基础设施部署状态

| 组件 | 状态 | 端点/说明 |
|------|------|----------|
| **GCP Bootstrap Server** | ✅ 运行中 | http://34.180.51.11:8080 |
| **GCP Compute Node** | ✅ 运行中 | http://34.180.51.11:1234 |
| DNS (bootstrap.deparrow.net) | ⚠️ 可选 | 可指向 34.180.51.11 |
| DPC 区块链 | ❌ 待开发 | Phase 3-5 开发 |

### 零风险品牌化实现 (2026-03-22)

**实现方式:** 包装器模式 - 不修改任何 Go 代码

| 组件 | 文件 | 说明 |
|------|------|------|
| CLI 包装器 | `bin/deparrow` | 调用 bacalhau，显示 DEparrow 品牌 |
| ISO 构建脚本 | `build-deparrow-iso.sh` | Alpine minirootfs + DEparrow 品牌 |
| ASCII 横幅 | `deparrow-banner.txt` | 启动时显示 |
| GRUB 菜单 | 内置 | "DEparrow OS" 选项 |

**风险等级:** 零 - 无 Go 代码修改

**构建命令:**
```bash
cd deparrow/bootable
./build-deparrow-iso.sh           # 标准版
./build-deparrow-iso.sh --local   # QEMU 测试版 (连接 10.0.2.2:8080)
```

### 代码统计

| 组件 | 文件数 | 测试文件 | 说明 |
|------|--------|----------|------|
| pkg/ (核心库) | 941 Go 文件 | 288 | Go 核心模块 (40 子目录) |
| WebUI | 3713 TS/TSX | 24 | Next.js 15 + React 组件 |
| PicoClaw DEparrow | 14 | 7 | 87%+ 覆盖率 |
| pkg/globalvm | 18 | - | Global VM 实现 |
| K8s Manifests | 22 | - | base/*.yaml |
| Python Tests | - | 91 | 完整覆盖 |
| GRUB BIOS 模块 | 304 | - | i386-pc/ |
| ISO 输出 | 6 | - | 28MB ~ 112MB |
| bin/deparrow | 1 | - | CLI 包装器 (零风险) |

### 核心文件清单

```
bin/deparrow              # DEparrow CLI 包装器 (零风险实现)

picoclaw/pkg/deparrow/
├── types.go, client.go
├── job_tool.go, credit_tool.go
├── node_tool.go, wallet_tool.go
├── register.go
└── *_test.go (7 测试文件)

deparrow/gui-layer/src/pages/
├── Dashboard.tsx, Jobs.tsx
├── Wallet.tsx, Nodes.tsx
├── Settings.tsx, Login.tsx
├── Agent.tsx, Providers.tsx

deparrow/bootable/
├── build-deparrow-iso.sh  # DEparrow 品牌版 ISO (推荐)
├── build-iso-v2.sh        # Auto-Join ISO
├── build-hybrid-iso.sh    # BIOS+EFI 混合 ISO
├── build-iso-fixed.sh     # 修复版 ISO
├── build-iso.sh           # 基础 ISO
├── create-iso.sh          # ISO 创建脚本
├── auto-install.sh        # 自动安装脚本
├── deparrow-banner.txt    # DEparrow ASCII 横幅
├── alpine-minirootfs-3.20.0-x86_64.tar.gz  # Alpine 基础
├── i386-pc/               # GRUB BIOS 模块 (304 文件)
└── output/
    ├── deparrow-os-1.0.0.iso     (81MB)  # 品牌版 (推荐)
    ├── deparrow-hybrid.iso       (112MB)
    ├── deparrow-autojoin.iso     (99MB)
    └── deparrow-alpine.iso       (49MB)

deparrow/docs/
├── DPC-TOKEN-DESIGN.md        # DPC 代币设计文档
├── DPC-IMPLEMENTATION-PLAN.md # DPC Phase 3 实现计划 (2026-03-26)
└── SOCIAL-APP-DESIGN.md       # 社交 App 原型设计文档 (2026-03-26)
├── DPC-IMPLEMENTATION-PLAN.md # DPC Phase 3 实现计划
└── SOCIAL-APP-DESIGN.md       # 社交 App 原型设计文档

webui/
├── app/ (Next.js App Router)
├── components/ (React 组件)
└── *.test.tsx (24 测试文件)

deparrow/k8s/base/
└── 22 YAML manifests
```

---

## NEXT CHAPTER: DEPARROW COIN (DPC)

### The Vision for AI Agent Currency

```
┌─────────────────────────────────────────────────────────────────┐
│              CRYPTOCURRENCY FOR AI AGENTS                       │
│                                                                 │
│   "AI Agents don't need banks. They need their own money."     │
│                                                                 │
│   [AI Agent] ──▶ Earn DPC ──▶ Provide Services                │
│        ▲                           │                           │
│        │                           ▼                           │
│   [Run Itself] ◀── Buy Compute ◀─ DPC Token                   │
│                                                                 │
│   No banks. No AWS. No centralized servers.                    │
│   AI agents run FREE in the real world.                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Roadmap: DEPARROW COIN (DPC)

| Phase | Goal | Status |
|-------|------|--------|
| Phase 1 | Core Platform | ✅ COMPLETE |
| Phase 2 | Token Design | ✅ COMPLETE |
| Phase 3 | Smart Contracts | 📋 Plan Ready (see DPC-IMPLEMENTATION-PLAN.md) |
| Phase 4 | Mainnet Launch | Pending (2-3 月) |
| Phase 5 | AI Agent Wallets | Pending (3-4 月) |

---

## 基础设施部署指南

### 当前分支状态

**Git 分支:** `main`
**最新提交:** 0653b9bcd - feat: GCP deployment complete and project status update to 95%

### GCP 部署 (已运行)

**部署信息:**
| 项目 | 值 |
|------|-----|
| 公网 IP | 34.180.51.11 |
| 区域 | asia-south1-b |
| 机器类型 | e2-standard-4 (4 vCPU, 16GB RAM) |
| OS | Debian 12 |

**SSH 连接:**
```bash
gcloud compute ssh deparrow-node --zone=asia-south1-b
```

**GCP 部署命令:**
```bash
# 创建静态 IP
gcloud compute addresses create deparrow-ip --region=asia-south1

# 创建 VM
gcloud compute instances create deparrow-node \
  --zone=asia-south1-b \
  --machine-type=e2-standard-4 \
  --address=deparrow-ip \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --tags=http-server,https-server

# 安装依赖 (在 VM 上)
sudo apt update && sudo apt install -y docker.io golang-go python3-pip
sudo usermod -aG docker $USER

# 克隆并构建
git clone https://github.com/Bhuw1234/fftp.git
cd bacalhau && go build -o bacalhau .

# 启动服务
cd deparrow/metaos-layer
pip3 install -r requirements.txt
python3 bootstrap-server.py --port 8080 &

./bacalhau serve --orchestrator --compute --api-port 1234 &
```

### 本地部署

```bash
# 1. Bootstrap Server
cd deparrow/metaos-layer
pip install -r requirements.txt
python3 bootstrap-server.py --port 8080

# 2. Orchestrator
./bacalhau serve --orchestrator --orchestrator-port 4222 --api-port 1234

# 3. DNS 配置
bootstrap.deparrow.net → A record → VPS 公网 IP
```

---

## 未来功能路线图

| 优先级 | 功能 | 说明 |
|--------|------|------|
| HIGH | DPC 智能合约 | Cosmos SDK 模块开发 |
| HIGH | DPC 主网启动 | 创世区块、验证者、代币分发 |
| MEDIUM | 社交 App 原型 | "使用即赚钱" 移动应用 |
| MEDIUM | DNS 配置 | bootstrap.deparrow.net → GCP IP |
| MEDIUM | 用户友好 ISO 启动 | TUI 仪表板、进度消息 |
| LOW | code-server 集成 | +50MB ISO 体积 |
| LOW | 终端浏览器 | w3m/lynx (+2MB) |

### 社交 App 概念 - "使用即赚钱"

```
┌─────────────────────────────────────────────────────────────────┐
│                    "FACEBOOK THAT PAYS YOU"                     │
│                                                                 │
│   传统社交平台: 用户免费使用 → 平台赚取广告费                    │
│   DEparrow 社交: 用户免费使用 → 用户赚取 DPC 代币               │
│                                                                 │
│   工作原理:                                                     │
│   1. 用户从 App Store 下载应用                                  │
│   2. 后台运行轻量级计算节点                                     │
│   3. 用户刷动态、聊天、发帖                                     │
│   4. 设备空闲时贡献算力                                         │
│   5. 用户自动赚取 DPC 代币                                      │
│                                                                 │
│   技术实现:                                                     │
│   - React Native / Flutter 跨平台应用                           │
│   - 嵌入式 Bacalhau Lite 计算引擎                               │
│   - WalletConnect 自托管钱包                                    │
│   - Cosmos SDK (DPC) 代币集成                                   │
│                                                                 │
│   预计收益:                                                     │
│   - 单设备: ~$0.01/天                                          │
│   - 100万用户: ~$10,000/天 网络收入                             │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**DPC Token 设计文档:** `deparrow/docs/DPC-TOKEN-DESIGN.md`

| 参数 | 值 |
|------|-----|
| 最大供应量 | 21B DPC |
| 初始供应量 | 1B DPC |
| 分配比例 | 70% 挖矿 / 15% 开发 / 10% 生态 / 5% 销售 |
| 共识机制 | Proof-of-Compute (完成作业 = 挖矿) |
| 区块链 | Cosmos SDK (Go) |

---

## 资源链接

- **官方网站**: https://deparrow.io
- **GitHub**: https://github.com/Bhuw1234/fftp
- **文档**: https://docs.deparrow.io
- **PicoClaw**: https://picoclaw.io

---

## 许可证

Apache 2.0 许可证

## 版本兼容性

- Go 1.24.0+ (Bacalhau Core)
- Go 1.25.7+ (PicoClaw)
- Node.js 18+
- Python 3.12.3+
- Docker 20.10+

---

*文档最后更新: 2026-03-26*

---

**DEparrow 核心平台已就绪 (~95%)**

**GCP 部署完成 - 34.180.51.11 运行中 ✅**

**零风险品牌化完成 - 无 Go 代码修改**

**下一步: DPC 区块链开发 (Phase 3-5) + 社交 App 原型**