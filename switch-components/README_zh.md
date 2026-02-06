# Switch: 动态特性开关与远程配置系统

<div align="center">

<img src="switch.svg" alt="Switch Logo" width="300">

**为现代应用开发打造的强大实时特性开关系统**

[![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://gitee.com/fatzeng/collections/413616)

[English](README.md) | [中文](README_zh.md)

</div>

---

## 🎯 什么是Switch？

**Switch是一个分布式实时特性管控平台**
，为企业级应用提供安全、高效的动态配置能力。它通过先进的WebSocket长连接架构和多驱动通信机制，实现了配置变更的毫秒级下发，让开发团队能够在不重启应用的情况下，精确控制功能发布、用户体验和系统行为。

### 🏗️ 核心架构优势

**1. 企业级通信架构**

- **WebSocket长连接框架** - 基于`switch-components/pc`的持久连接管理
- **多驱动支持** - Webhook、Kafka、长轮询三种通信模式
- **智能网络发现** - 自动适配NAT环境，解决复杂网络场景
- **分层确认机制** - 确保配置下发的可靠性和一致性

**2. 灵活的因子系统**

```go
// 不只是简单的true/false，而是基于复杂规则的智能决策
if _switch.IsOpen(ctx, "feature_enabled") {
    // 系统会根据配置的多维度因子（例如：用户属性、地理位置、时间窗口等）
    // 实时计算是否启用该功能
}
```

**3. 多租户管理体系**

- **租户隔离** - 完整的数据和权限隔离
- **环境管理** - 开发、测试、生产环境的独立配置搭配严格的配置推送制度
- **审批工作流** - 敏感变更的多级审批机制

## 系统架构

Switch生态系统由以下核心组件组成：

- **switch-admin**: 后端服务，负责管理配置和客户端通信
- **switch-frontend**: Web界面，用于配置管理
- **switch-sdk-go**: Go SDK，用于将开关集成到业务应用中形成客户端
- **switch-sdk-core**: 核心定义和接口
- **switch-components**: 通信和核心逻辑的实现
- **switch-client-demo**: 演示应用示例

---

# Switch-Components: 通信与核心逻辑实现

`switch-components` 是 Switch 生态系统的核心组件库，提供了 WebSocket 通信框架、多驱动系统、分布式 ID
生成、日志追踪等企业级基础设施的完整实现。它将 `switch-sdk-core` 中定义的抽象接口转化为可靠、高性能的工程实现，为上层应用提供开箱即用的技术能力。

---

## ✨ 功能特性

- **WebSocket 通信框架**: 完整的 WebSocket 服务器和客户端实现，提供控制通道能力
- **消息双向通信**: 基于 WebSocket 的双向通信，确保配置实时下发
- **多驱动系统**: Webhook、Kafka、长轮询三种配置通信驱动的完整实现
- **连接管理**: 高级连接生命周期管理，支持自动重连和健康检查
- **信任与安全**: 客户端注册和信任建立机制，确保通信安全
- **消息广播**: 高效的消息分发到受信任的多个客户端
- **单点通知**: 精准的消息推送到指定客户端
- **网络发现**: 智能网络检测和 IP 管理，适配复杂网络环境
- **配置分发**: 实时配置推送和同步机制

---

## 📁 项目结构

```
switch-components/
├── pc/                         # 持久连接管理 (WebSocket框架)
├── drivers/                    # 通信驱动层
├── bc/                        # 业务上下文管理
├── config/                    # 配置管理
├── http/                      # HTTP中间件
├── grpc/                      # gRPC通信支持
├── logging/                   # 日志系统
│   └── request/               # 请求日志追踪
├── recovery/                  # 异常恢复
├── snowflake/                 # 分布式ID生成
└── system/                    # 系统工具
```

### 关键组件说明：

- **`pc/`**: 核心的WebSocket长连接框架，提供实时通信基础设施
- **`drivers/`**: 多种通信驱动实现，支持Kafka、Webhook、长轮询等模式
- **`bc/`**: 业务上下文管理，提供请求生命周期内的上下文信息
- **`config/`**: 统一的配置管理，支持YAML和环境变量
- **`http/`**: HTTP服务的中间件和工具支持
- **`grpc/`**: 高性能RPC通信能力
- **`logging/`**: 结构化日志和请求追踪
- **`recovery/`**: 系统稳定性保障机制
- **`snowflake/`**: 分布式唯一ID生成服务
- **`system/`**: 底层系统操作支持

---

## 🏛️ WebSocket 通信框架

### 核心架构与通信流程

持久连接（PC）框架提供了完整的 WebSocket 通信解决方案，作为控制通道传递驱动配置和管理信息：

![communication-flow.svg](communication-flow.svg)

---

## 🚀 配置通信驱动系统

Switch 通过 WebSocket 控制通道下发驱动配置后，客户端会启动相应的配置通信驱动。系统支持三种驱动模式，满足不同网络环境和业务场景需求：

### 支持的驱动

#### 1. Webhook 驱动（推送模式）

- 基于 HTTP 的推送通知机制
- 客户端启动 Webhook 服务器接收配置推送
- 智能网络发现，自动适配 NAT 环境
- 推送失败时自动重试，确保配置可达

#### 2. 长轮询驱动（拉取模式）

- HTTP 长轮询方式主动拉取配置更新
- 适用于限制性网络环境（防火墙、代理）
- 指数退避的自动重试机制
- 低资源消耗，高兼容性

#### 3. Kafka 驱动（消息队列模式）

- 基于 Kafka 的消息队列分发
- 基于主题的配置通道隔离
- 高吞吐量和高可靠性保障
- 适用于大规模分布式部署场景

### 驱动配置示例

> 驱动配置将通过 WebSocket 控制通道下发，客户端接收后启动相应的配置通信驱动。

#### Kafka 驱动配置

**生产者配置：**

```json
{
  // broker列表(必填)
  "brokers": ["localhost:9092", "localhost:9093"],
  // kafka主题(必填) 必须提前创建不支持自动创建 驱动启动时会验证主题有效性
  "topic": "example-topic",

  // 确认机制(非必填) all-所有副本确认 one-leader确认 none-不需要确认
  "requiredAcks": "all",
  // 等待broker返回确认的最大超时时间(非必填)
  "timeout": "30s",
  // 当消息发送失败时(网络问题 broker不可用等),客户端会自动重试retries次(非必填)
  "retries": 3,
  // 写失败后最小重试间隔(非必填)
  "retryBackoffMin": "100ms",
  // 写失败后最大重试间隔(非必填)
  "retryBackoffMax": "1s",
  // 消息压缩算法(非必填) gzip snappy lz4 zstd
  "compression": "snappy",

  // broker连接拨号器的连接超时时间(非必填)
  "connectTimeout": "10s",
  // broker测试连通性时的超时时间(非必填)
  "validateTimeout": "10s",

  // 时间超过batchTimeout触发写(非必填)
  "batchTimeout": "1s",
  // 消息大小满足batchBytes触发写(非必填)
  "batchBytes": 1048576,
  // 消息数量满足batchSize触发写(非必填)
  "batchSize": 50,

  // 安全配置(非必填)
  "security": {
    // sasl验证(非必填)
    "sasl": {
      // 是否开启
      "enabled": false,
      // 认证机制 PLAIN SCRAM-SHA-256 SCRAM-SHA-512等
      "mechanism": "PLAIN",
      // 用户名
      "username": "your-username",
      // 密码
      "password": "your-password"
    },
    // tls验证(非必填)
    "tls": {
      // 是否开启
      "enabled": false,
      // CA证书地址 用于验证kafka服务端的有效性(系统路径)
      "caFile": "/path/to/ca.pem",
      // 客户端证书地址(系统路径)
      "certFile": "/path/to/cert.pem",
      // 客户端密钥地址(系统路径)
      "keyFile": "/path/to/key.pem",
      // 是否跳过证书验证
      "insecureSkipVerify": false
    }
  }
}
```

**消费者配置：**

```json
{
  // broker列表(必填)
  "brokers": [
    "localhost:9092",
    "localhost:9093"
  ],
  // kafka主题(必填) 必须提前创建不支持自动创建 驱动启动时会验证主题有效性
  "topic": "example-topic",

  // 消费者组ID(非必填)(没有特殊需求此处请设置为空,交由框架去生成消费者组ID)
  "groupId": "",
  // 偏移量重置策略(非必填) earliest(从最早消息开始消费) latest(从最新消息开始消费)
  "autoOffsetReset": "latest",
  // 自动提交(非必填)
  "enableAutoCommit": true,
  // 手动提交间隔(非必填)(可以适当设置大一点,开关消息重复消费并无影响)
  "autoCommitInterval": "10s",

  // 连接超时配置(非必填)
  "connectTimeout": "10s",
  // 验证连接的连通性超时配置(非必填)
  "validateTimeout": "10s",

  // 读消息超时配置(非必填)
  "readTimeout": "10s",
  // 提交偏移量超时配置(非必填)
  "commitTimeout": "10s",

  // 安全配置(非必填)
  "security": {
    // sasl验证(非必填)
    "sasl": {
      // 是否开启
      "enabled": false,
      // 认证机制 PLAIN SCRAM-SHA-256 SCRAM-SHA-512等
      "mechanism": "PLAIN",
      // 用户名
      "username": "your-username",
      // 密码
      "password": "your-password"
    },
    // tls验证(非必填)
    "tls": {
      // 是否开启
      "enabled": false,
      // CA证书地址 用于验证kafka服务端的有效性(系统路径)
      "caFile": "/path/to/ca.pem",
      // 客户端证书地址(系统路径)
      "certFile": "/path/to/cert.pem",
      // 客户端密钥地址(系统路径)
      "keyFile": "/path/to/key.pem",
      // 是否跳过证书验证
      "insecureSkipVerify": false
    }
  },
  // 重试配置(非必填)
  "retry": {
    // 失败超出count次数将重启 成功则重置
    "count": 1,
    // 重启间隔时间
    "backoff": "3s"
  }
}
```

---

## 🔧 高级功能

### 网络发现

系统包含智能网络发现功能：

- **本地IP检测**: 自动检测本地网络接口
- **公网IP发现**: 使用外部服务确定公网IP
- **NAT检测**: 识别NAT环境并调整通信策略
- **Webhook可达性测试**: 测试webhook端点可访问性

### 消息广播

高效的消息分发系统：

- **定向广播**: 向特定客户端或组发送消息
- **基于主题的路由**: 根据主题或模式路由消息
- **投递确认**: 跟踪消息投递并处理失败

### 连接管理

高级连接生命周期管理：

- **连接池**: 高效的连接资源管理
- **健康监控**: 持续的连接健康检查
- **优雅关闭**: 清洁的连接终止
- **资源清理**: 自动清理断开的客户端

---

## 📦 集成使用

### 安装依赖

```bash
go get gitee.com/fatzeng/switch-components
```

### 快速开始

作为工具组件库，`switch-components` 提供了丰富的可复用组件。以下是一些常见的集成场景：

#### 使用 WebSocket 通信框架

```go
// 创建 WebSocket 服务器
wsConfig := config.GlobalConfig.Pc
if wsConfig == nil {
    //默认配置
    wsConfig = pc.DefaultServerConfig()
}
// 连接回调
wsConfig.OnConnect = onClientConnect
// 断连回调
wsConfig.OnDisconnect = onClientDisconnect
// 消息处理回调
wsConfig.MessageHandler = onMessage
// 受信回调
wsConfig.OnClientTrusted = onClientTrusted

wsServer = pc.NewServer(wsConfig)

// 注册配置变更业务端点
wsServer.RegisterHandler(pc.WsEndpointChangeConfig, nil)
// 注册全量配置同步业务端点
wsServer.RegisterHandler(pc.WsEndpointFullSyncConfig, nil)
// 注册全量开关同步业务端点
wsServer.RegisterHandler(pc.WsEndpointFullSync, nil)

wsServer.Start(ctx)
```

#### 使用日志系统

```go
// 结构化日志记录
log, err := logging.New(&logger.LoggerConfig{
    Level:            "info",                     // 日志级别
    OutputDir:        "./logs",                   // 日志输出目录
    FileNameFormat:   "switch-demo_%Y-%m-%d.log", // 日志文件名格式
    MaxSize:          50,                         // 单个日志文件最大大小(MB)
    MaxBackups:       3,                          // 保留的旧日志文件数量
    MaxAge:           7,                          // 日志文件保留天数
    Compress:         false,                      // 是否压缩旧日志文件
    ShowCaller:       true,                       // 是否显示调用者信息
    EnableConsole:    true,                       // 是否启用控制台输出
    EnableJSON:       false,                      // 是否使用JSON格式(demo用false便于阅读)
    EnableStackTrace: true,                       // 是否启用堆栈跟踪
    StackTraceLevel:  "error",                    // 堆栈跟踪级别
    TimeFormat:       "2006-01-02 15:04:05",      // 时间格式
})
```

更多详细的使用示例和 API 文档，请参考各组件目录下的说明。

## 🤝 贡献指南

我们欢迎并感谢所有形式的贡献！无论是报告问题、提出功能建议、改进文档，还是提交代码，您的参与都将帮助 Switch 变得更好。

### 如何贡献

1. **Fork 本仓库**并创建您的特性分支
2. **编写代码**并确保遵循项目的代码规范
3. **添加测试**以覆盖您的更改
4. **提交 Pull Request**，并详细描述您的更改内容和动机

### 贡献类型

- 🐛 **Bug 修复**: 发现并修复问题
- ✨ **新功能**: 提出并实现新特性
- 📝 **文档改进**: 完善文档和示例
- 🎨 **代码优化**: 提升代码质量和性能
- 🧪 **测试增强**: 增加测试覆盖率

更多详细信息，请参阅我们的贡献指南文档。

## 📄 许可证

本项目采用 [MIT 许可证](LICENSE)。
