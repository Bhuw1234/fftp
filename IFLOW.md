# DEparrow - 去中心化 AI 操作系统

## 项目概述

**DEparrow** 是一个**去中心化 AI 操作系统**，构建于 Bacalhau 分布式计算基础设施之上。DEparrow 将 AI 计算能力去中心化，让任何人都可以贡献和使用分布式 AI 算力。

### 核心理念
- 🌐 **去中心化**: 无单点故障，全球分布式节点网络
- 🤖 **AI 原生**: 专为 AI 工作负载优化的操作系统
- 💰 **信用经济**: 贡献算力获得信用，使用算力消耗信用
- 🔐 **数据主权**: 数据留在本地，计算移动到数据
- 🖥️ **终端优先**: Clawdbot 提供强大的终端交互体验

### 用户交互方式

| 接口 | 描述 | 目录 |
|------|------|------|
| **Clawdbot 终端** | AI 驱动的命令行界面，自然语言交互 | `clawdbot/` |
| **Web GUI** | 可视化仪表板，作业管理 | `webui/` + `deparrow/gui-layer/` |
| **Python SDK** | 程序化访问 API | `python/` + `clients/` |

### 系统组件

| 组件 | 描述 |
|------|------|
| **DEparrow OS** | 去中心化 AI 操作系统核心 |
| **Clawdbot** | 终端 AI 助手 (开源) - 用户主要交互入口 |
| **Bacalhau** | 底层分布式计算编排引擎 |

### 核心特性
- ⚡ **快速作业处理**: 作业在数据创建地点并行处理
- 💰 **低成本**: 减少数据移动带来的网络和存储成本
- 🔒 **安全执行**: 数据清洗和安全控制在迁移前进行
- 🚛 **大规模数据**: 高效处理 PB 级数据
- 🏢 **数据主权**: 在安全边界内处理敏感数据
- 🤝 **跨组织计算**: 允许在受保护数据集上进行特定计算
- 🔧 **单一二进制**: 客户端、编排器、计算节点三合一

---

## 技术栈

### 核心语言和运行时

| 工具 | 版本 | 用途 |
|------|------|------|
| Go | 1.24.0 (go.mod) / 1.21.11 (.tool-versions) | 核心语言 |
| Python | 3.10.5 | SDK 和工具 |
| Node.js | 21.5.0 | WebUI 开发 |
| TypeScript | 5.x | WebUI 类型系统 |

### 后端依赖

| 库 | 版本 | 用途 |
|----|------|------|
| NATS Server | v2.11.6 | 分布式消息传递 |
| libp2p | v0.41.1 | P2P 网络 |
| Docker | v27.1.1 | 容器执行引擎 |
| wazero | v1.9.0 | WebAssembly 运行时 |
| IPFS (kubo) | v0.35.0 | 分布式存储 |
| IPFS (boxo) | v0.32.0 | IPFS 工具库 |
| Open Policy Agent | v0.60.0 | 策略引擎 |
| AWS SDK v2 | v1.36.5 | S3 存储集成 |
| OpenTelemetry | v1.37.0 | 可观测性 |
| zerolog | v1.34.0 | 结构化日志 |
| Cobra | v1.9.1 | CLI 框架 |
| Echo | v4.13.4 | HTTP 服务器 |
| JWT (golang-jwt) | v5.2.2 | 认证 |
| go-playground/validator | v10.26.0 | 输入验证 |
| samber/lo | v1.51.0 | Go 泛型工具库 |

### 前端依赖 (WebUI)

| 库 | 版本 | 用途 |
|----|------|------|
| Next.js | 15.2.4 | React 框架 |
| React | 18 | UI 库 |
| Radix UI | 1.x | 组件库 |
| Tailwind CSS | 3.4.1 | 样式系统 |
| @hey-api/client-fetch | 0.2.4 | API 客户端 |
| Lucide React | 0.438.0 | 图标库 |
| Yarn | 4.4.1 | 包管理器 |
| axios | 1.8.2 | HTTP 客户端 |

### Clawdbot 依赖

| 库 | 版本 | 用途 |
|----|------|------|
| @carbon | 0.14.0 | UI 组件框架 |
| @pi-ai/* | 0.49.3 | AI 核心引擎 |
| hono | 4.11.4 | 轻量级 Web 框架 |
| playwright-core | 1.58.0 | 浏览器自动化 |
| vitest | 4.0.18 | 测试框架 |
| TypeScript | 5.9.3 | 类型系统 |
| oxlint | 1.41.0 | 代码检查 |

### 构建工具

| 工具 | 版本 | 用途 |
|------|------|------|
| Earthly | 0.8.3 | 容器化构建 |
| golangci-lint | 1.64.2 | Go 代码检查 |
| pnpm | 9.0.6 / 10.23.0 (Clawdbot) | Node.js 包管理 |
| direnv | 2.33.0 | 环境管理 |
| pre-commit | 3.6.0 | Git 钩子 |

---

## 项目结构

```
.
├── main.go                    # 主入口点 (Swagger API 注释)
├── go.mod                     # Go 模块定义 (go 1.24.0)
├── Makefile                   # 50+ 构建目标
│
├── cmd/                       # 命令行接口
│   ├── cli/                   # CLI 命令实现
│   ├── testing/               # 测试工具
│   └── util/                  # 命令工具
│
├── pkg/                       # 核心库 (43 子目录)
│   ├── analytics/             # 分析遥测
│   ├── authn/                 # 认证
│   ├── authz/                 # 授权 (OPA 集成)
│   ├── bacerrors/             # 错误处理
│   ├── bidstrategy/           # 投标策略
│   ├── cache/                 # 缓存
│   ├── compute/               # 计算节点逻辑
│   ├── config/                # 配置管理
│   ├── credsecurity/          # 凭证安全
│   ├── docker/                # Docker 集成
│   ├── downloader/            # 下载器
│   ├── executor/              # 执行引擎
│   ├── ipfs/                  # IPFS 集成
│   ├── jobstore/              # 作业存储
│   ├── lib/                   # 库函数
│   ├── licensing/             # 许可证管理
│   ├── logger/                # 日志系统
│   ├── models/                # 数据模型
│   ├── nats/                  # NATS 集成
│   ├── node/                  # 节点管理
│   ├── orchestrator/          # 编排器逻辑
│   ├── publicapi/             # 公共 API
│   ├── publisher/             # 发布器
│   ├── pubsub/                # 发布订阅
│   ├── repo/                  # 仓库管理
│   ├── s3/                    # S3 集成
│   ├── sso/                   # 单点登录
│   ├── storage/               # 存储后端
│   ├── swagger/               # API 文档
│   ├── system/                # 系统工具
│   ├── telemetry/             # 遥测 (OpenTelemetry)
│   ├── transport/             # 传输层
│   ├── userstrings/           # 用户字符串
│   ├── util/                  # 工具函数
│   └── version/               # 版本管理
│
├── webui/                     # Web 界面 (Next.js 15)
│   ├── app/                   # Next.js App Router
│   │   ├── jobs/              # 作业管理页面
│   │   ├── nodes/             # 节点管理页面
│   │   └── providers/         # React Context
│   ├── components/            # React 组件
│   │   ├── jobs/              # 作业相关组件
│   │   ├── nodes/             # 节点相关组件
│   │   ├── layout/            # 布局组件
│   │   └── ui/                # 通用 UI 组件
│   ├── hooks/                 # 自定义 Hooks
│   ├── lib/                   # 工具库
│   └── webui.go               # Go 嵌入文件
│
├── python/                    # Python SDK
│   ├── bacalhau_sdk/          # SDK 核心代码
│   ├── examples/              # 示例代码
│   └── tests/                 # 测试套件
│
├── clients/                   # API 客户端
│   └── python/                # Python API 客户端
│
├── integration/               # 第三方集成
│   ├── airflow/               # Apache Airflow 集成
│   └── flyte/                 # Flyte 集成
│
├── deparrow/                  # DEparrow 平台
│   ├── alpine-layer/          # Alpine Linux 基础层
│   ├── bacalhau-layer/        # Bacalhau 层
│   ├── bootable/              # 可启动镜像
│   ├── gui-layer/             # GUI 用户界面层
│   ├── metaos-layer/          # Meta-OS 控制平面层
│   ├── k8s/                   # Kubernetes 部署配置
│   ├── test-integration/      # 集成测试
│   ├── scripts/               # 部署脚本
│   ├── config/                # 配置文件
│   ├── DEPLOYMENT.md          # 部署指南
│   └── IMPLEMENTATION_COMPLETE.md
│
├── clawdbot/                  # Clawdbot AI 助手 (独立项目)
│   ├── src/                   # 源代码
│   │   ├── agents/            # AI Agent 核心
│   │   │   └── tools/deparrow/# DEparrow 工具集成
│   │   ├── cli/               # CLI 命令
│   │   ├── commands/          # 命令实现
│   │   ├── gateway/           # 网关服务
│   │   ├── providers/         # AI 提供者
│   │   ├── telegram/          # Telegram 集成
│   │   ├── discord/           # Discord 集成
│   │   ├── slack/             # Slack 集成
│   │   ├── signal/            # Signal 集成
│   │   ├── imessage/          # iMessage 集成
│   │   └── web/               # WhatsApp Web 集成
│   ├── apps/                  # 应用程序
│   │   ├── macos/             # macOS 应用
│   │   ├── ios/               # iOS 应用
│   │   └── android/           # Android 应用
│   ├── extensions/            # 扩展模块
│   ├── docs/                  # 文档
│   ├── skills/                # AI 技能
│   ├── ui/                    # UI 组件
│   └── scripts/               # 脚本
│
├── docker/                    # Docker 镜像构建
├── docker-compose-deployment/ # Docker Compose 部署
├── test/                      # Bash 测试脚本
├── test_integration/          # 集成测试
├── testdata/                  # 测试数据 (包括 WASM 二进制)
├── benchmark/                 # 性能基准测试
├── ops/                       # 运维脚本
├── scripts/                   # 构建脚本
├── docs/                      # 文档
│
└── .github/workflows/         # CI/CD 工作流
    ├── main.yml               # 主分支构建
    ├── pr-checks.yml          # PR 检查
    ├── release.yml            # 发布流程
    ├── _build.yml             # 二进制构建
    ├── _docker_publish.yml    # Docker 发布
    ├── _s3_publish.yml        # S3 发布
    ├── _test.yml              # 测试执行
    ├── _test_container.yml    # 容器测试
    ├── _test_coverage.yml     # 覆盖率测试
    └── _static-analysis.yml   # 静态分析
```

---

## DEparrow 四层架构 - 去中心化 AI 操作系统

DEparrow 采用四层架构设计，将 AI 计算能力去中心化：

```
┌─────────────────────────────────────────────────────────┐
│                    GUI 用户界面层                         │
│         Dashboard | Jobs | Wallet | Nodes | AI Chat     │
├─────────────────────────────────────────────────────────┤
│                 Meta-OS 控制平面层                        │
│    引导服务 | 信用系统 | 作业准入 | JWT 认证 | AI 调度     │
├─────────────────────────────────────────────────────────┤
│                Alpine Linux 基础层                        │
│         轻量级 OS | 自动加入 | x86_64/arm64              │
├─────────────────────────────────────────────────────────┤
│               Bacalhau 执行网络层                         │
│    Docker | WebAssembly | NATS | libp2p | IPFS          │
└─────────────────────────────────────────────────────────┘
```

### 1. Alpine Linux 基础层
- **轻量级 OS**: 最小化系统开销
- **自动加入**: 节点自动发现并加入网络
- **多架构支持**: x86_64 和 arm64
- **健康监控**: 实时系统检查
- **OpenRC 服务管理**: Bacalhau 作为系统服务

### 2. Meta-OS 控制平面层 (DEparrow 核心)
- **引导服务**: DEparrow 专用引导节点
- **编排器注册**: 编排器节点发现和注册系统
- **信用支付系统**: 基于信用的作业提交控制
- **作业准入控制**: 支付验证后允许作业提交
- **JWT 认证**: 完整的身份验证和授权系统

### 3. GUI 用户界面层 (顶层)
- **Dashboard**: 网络统计和监控
- **Jobs**: 作业管理界面
- **Wallet**: 信用管理系统
- **Nodes**: 节点监控仪表板
- **Settings**: 用户配置
- **Login**: 身份验证界面

### 4. Bacalhau 执行网络层
- **Docker 执行**: 容器化作业执行
- **WebAssembly**: 沙箱安全执行
- **NATS 消息传递**: 分布式消息系统
- **libp2p P2P**: 去中心化网络通信
- **IPFS 存储**: 分布式文件存储

---

## Clawdbot - DEparrow 终端界面

**Clawdbot** 是 DEparrow 的**开源终端 AI 助手**，让用户通过自然语言与去中心化 AI 网络交互。

### 为什么用 Clawdbot？
- 🗣️ **自然语言**: 用普通话描述任务，无需记忆复杂命令
- ⚡ **快速上手**: 一条命令安装，立即使用
- 🔧 **强大功能**: 直接访问 DEparrow 网络的全部能力
- 🌐 **多渠道**: 终端、WhatsApp、Telegram、Slack 等

### 安装和使用

```bash
# 安装 Clawdbot
npm install -g clawdbot@latest

# 初始化并连接 DEparrow 网络
clawdbot onboard --install-daemon

# 启动终端助手
clawdbot agent
```

### DEparrow 工具集成

Clawdbot 内置 DEparrow 工具，通过 AI Agent 自动调用：

| 工具 | 描述 |
|------|------|
| `deparrow_network_status` | 获取网络状态 (活跃节点、总算力) |
| `deparrow_check_credits` | 查看信用余额 |
| `deparrow_submit_job` | 提交计算作业 |
| `deparrow_list_jobs` | 列出用户作业 |
| `deparrow_get_job` | 获取作业详情 |
| `deparrow_list_nodes` | 列出网络节点 |
| `deparrow_my_contribution` | 查看贡献统计和排名 |
| `deparrow_leaderboard` | 查看贡献者排行榜 |

### 示例：通过终端使用 DEparrow

```bash
# 提交 AI 训练作业到去中心化网络
$ clawdbot agent --message "在 DEparrow 网络上训练一个图像分类模型"
🦞 正在准备作业...
📊 信用余额: 1,250 credits
🌐 找到 47 个可用计算节点
✅ 作业已提交 (Job ID: abc123)
💰 预计消耗: 85 credits

# 查看网络状态
$ clawdbot agent --message "显示 DEparrow 网络状态"
🌐 DEparrow Network Status:
• Active Nodes: 1247
• Total Nodes: 1280
• Total Compute: 15200 GFLOPS
• Status: ✅ Healthy

# 查看我的贡献
$ clawdbot agent --message "我的贡献统计"
🔥 Your DEparrow Contribution (LIVE)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
CPU  ███████░░░░░░░░░░░░░  12.5%
GPU  ████░░░░░░░░░░░░░░░░   8.3%
RAM  ██░░░░░░░░░░░░░░░░░░   5.2%
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚡ Live: 245.6 GFLOPS
🏆 Rank: #5 of 127 nodes
🥇 Tier: GOLD
💰 Earned: 1,256 credits

# 查看钱包
$ clawdbot agent --message "我的信用余额"
💰 DEparrow Credits:
• Available: 1,250 credits
• Pending: 45 credits
• Total Earned: 2,500 credits
• Total Spent: 1,250 credits
```

### Clawdbot 技术栈
- **运行时**: Node.js ≥22.12.0
- **包管理**: pnpm 10.23.0
- **AI 引擎**: @pi-ai/* 0.49.3
- **模型支持**: Anthropic Claude, OpenAI GPT, Ollama
- **测试**: Vitest 4.0.18 + V8 覆盖率

---

## 开发环境设置

### 前提条件
```bash
# 工具版本 (来自 .tool-versions)
python      3.10.5
nodejs      21.5.0
golang      1.21.11 / 1.24.0 (go.mod)
direnv      2.33.0
earthly     0.8.3
golangci-lint 1.64.2
pnpm        9.0.6
pre-commit  3.6.0
```

### 快速开始
```bash
# 克隆仓库
git clone https://github.com/bacalhau-project/bacalhau.git
cd bacalhau

# 初始化开发环境
make init

# 安装 pre-commit 钩子
make install-pre-commit

# 安装依赖
make modtidy

# 运行测试
make test

# 构建项目
make build

# 启动开发栈
make devstack
```

---

## 构建命令

### Go 项目构建
```bash
make build                    # 构建 Go 二进制
make build-ci                 # CI 构建
make build-dev                # 开发构建并安装到 /usr/local/bin
make clean                    # 清理构建产物
```

### Docker 镜像
```bash
make build-bacalhau-base-image   # 构建基础镜像
make build-bacalhau-dind-image   # 构建 Docker-in-Docker 镜像
make build-http-gateway-image    # 构建 HTTP 网关镜像
make build-bacalhau-images       # 构建所有镜像
make push-bacalhau-images        # 推送镜像到注册表
```

### Python 包
```bash
make build-python-sdk         # 构建 Python SDK
make build-python-apiclient   # 构建 API 客户端
make build-bacalhau-airflow   # 构建 Airflow 集成
make build-bacalhau-flyte     # 构建 Flyte 集成
make release-python-sdk       # 发布到 PyPI
```

### WebUI
```bash
cd webui && yarn dev          # 开发模式
cd webui && yarn build        # 生产构建
cd webui && yarn lint         # 代码检查
cd webui && yarn format       # 格式化
cd webui && yarn generate-api # 生成 API 客户端
```

### Clawdbot
```bash
cd clawdbot
pnpm install                  # 安装依赖
pnpm build                    # 构建
pnpm dev                      # 开发模式
pnpm test                     # 运行测试
pnpm lint                     # 代码检查
pnpm format:fix               # 格式化修复
```

### 测试
```bash
make test                     # 运行所有测试
make unit-test                # 单元测试 (并行)
make integration-test         # 集成测试 (串行)
make bash-test                # Bash 测试
make test-python-sdk          # Python SDK 测试
make test-debug               # 调试模式测试
```

### 开发栈
```bash
make devstack                 # 默认开发栈
make devstack-one             # 单节点开发栈
make devstack-20              # 20 计算节点
make devstack-100             # 100 计算节点
make devstack-250             # 250 计算节点
make devstack-race            # 竞态检测模式
```

### 代码质量
```bash
make lint                     # golangci-lint
make precommit                # 运行所有 pre-commit 钩子
make modtidy                  # go mod tidy
make check-diff               # 检查 go.mod/go.sum 变更
make security                 # 安全检查 (gosec)
make spellcheck-code          # 拼写检查
make generate-swagger         # 生成 Swagger 文档
```

---

## DEparrow 部署

### 快速启动

1. **启动引导服务器**:
```bash
cd deparrow/metaos-layer
python3 bootstrap-server.py --host 0.0.0.0 --port 8080
```

2. **构建 Alpine 节点镜像**:
```bash
cd deparrow/alpine-layer
./build.sh
```

3. **部署计算节点**:
```bash
cd deparrow
docker-compose -f alpine-layer/config/docker-compose/deparrow-node.yml up -d
```

4. **启动 GUI**:
```bash
cd deparrow/gui-layer
npm install && npm start
```

5. **运行测试**:
```bash
cd deparrow
./test-integration.sh
```

### 环境变量
```bash
DEPARROW_SECRET_KEY          # JWT 令牌密钥
DEPARROW_BOOTSTRAP_HOST      # 引导服务器主机
DEPARROW_BOOTSTRAP_PORT      # 引导服务器端口 (默认 8080)
DEPARROW_NETWORK_NAME        # 网络名称
DEPARROW_API_URL             # API 地址 (Clawdbot 使用)
```

### 生产部署
```bash
# 使用 docker-compose 部署
cd deparrow
docker-compose -f docker-compose.prod.yml up -d

# Kubernetes 部署
kubectl apply -f k8s/
```

---

## Bacalhau 网络架构

### 节点类型
- **编排器节点**: `bacalhau serve --orchestrator` (端口 4222)
- **计算节点**: `bacalhau serve --compute`
- **混合节点**: 兼具编排和计算功能
- **API 服务**: 端口 1234 (默认)
- **WebUI 服务**: 端口 3000 (开发) / 80 (生产)

### 执行引擎
- **Docker**: 需要 Docker 运行时
- **WebAssembly**: wazero 运行时，沙箱执行
- **Native**: 直接主机执行 (受限)

### 存储类型
- **S3**: AWS S3 兼容存储
- **IPFS**: 分布式文件系统
- **Local**: 本地存储
- **HTTP/HTTPS**: 远程 HTTP 存储

---

## 代码质量标准

### Go 代码 (.golangci.yml)
- 行长度: 140 字符
- 复杂度: 最大 18 (gocyclo)
- 函数长度: 最大 100 行
- 日志: 使用 zerolog (禁止 logrus)
- 测试: `//go:building unit` 标签

### Python 代码 (ruff.toml)
- 使用 Ruff 检查和格式化
- 类型检查启用

### WebUI 代码
- ESLint + Prettier
- TypeScript 严格模式
- Radix UI + Tailwind CSS

### Clawdbot 代码
- Oxlint 代码检查
- Oxfmt 格式化
- TypeScript 5.9.3 严格模式
- Vitest 测试覆盖率: 70%+

---

## 配置文件

| 文件 | 用途 |
|------|------|
| `go.mod` / `go.sum` | Go 依赖 |
| `pyproject.toml` | Python 依赖 |
| `Makefile` | 构建自动化 (50+ 目标) |
| `.golangci.yml` | Go linter 配置 |
| `.pre-commit-config.yaml` | Pre-commit 钩子 |
| `cspell.yaml` | 拼写检查 |
| `ruff.toml` | Python Ruff 配置 |
| `webui/package.json` | WebUI 依赖 |
| `webui/tsconfig.json` | TypeScript 配置 |
| `webui/tailwind.config.ts` | Tailwind 配置 |
| `clawdbot/package.json` | Clawdbot 依赖 |
| `clawdbot/tsconfig.json` | Clawdbot TypeScript 配置 |
| `clawdbot/vitest.config.ts` | Clawdbot 测试配置 |

---

## 故障排查

### 常见问题
1. **构建失败**: 检查 Earthly (`earthly --version`)
2. **测试失败**: 确保 Docker 运行 (`docker ps`)
3. **Lint 错误**: 运行 `make precommit`
4. **依赖问题**: 运行 `make modtidy && make check-diff`
5. **WebUI 构建失败**: 检查 Node.js 版本 ≥18
6. **DEparrow 引导失败**: 检查端口 8080
7. **Clawdbot 安装失败**: 检查 Node.js 版本 ≥22.12.0

### 调试命令
```bash
# Go 调试日志
LOG_LEVEL=debug go test -v

# Bacalhau 调试
bacalhau --log-level=debug <command>

# 开发栈调试
LOG_LEVEL=debug make devstack

# WebUI Turbopack
cd webui && yarn dev --turbopack

# Clawdbot 调试
cd clawdbot && pnpm dev -- --log-level debug
```

---

## 资源链接

- **官方文档**: https://docs.bacalhau.org
- **官方网站**: https://www.bacalhau.org
- **GitHub**: https://github.com/bacalhau-project/bacalhau
- **Slack**: https://bit.ly/bacalhau-project-slack
- **Python SDK**: https://bacalhau-project.github.io/bacalhau-python/
- **DEparrow 部署**: [deparrow/DEPLOYMENT.md](deparrow/DEPLOYMENT.md)
- **Clawdbot 文档**: https://docs.clawd.bot

## 许可证

Apache 2.0 许可证 (见 LICENSE 文件)

## 版本兼容性

- Go 1.24.0+
- Node.js 18+ (Clawdbot 需要 22.12.0+)
- Python 3.10.5+
- Docker 20.10+

---

*文档最后更新: 2026-02-18*