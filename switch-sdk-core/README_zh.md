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

# Switch-SDK-Core: 核心定义与接口

`switch-sdk-core` 是 Switch 生态系统的**基础设施层**与**核心契约层**
，作为整个架构的基石，它为分布式特性管控平台提供了统一的类型系统、接口规范和通信协议。通过高度抽象的设计哲学与契约式编程范式，实现了组件间的松耦合架构和无限扩展能力，为构建企业级特性治理体系奠定了坚实的理论与实践基础。

---

## ✨ 核心能力

### 架构层面

- **契约式接口体系**: 建立驱动、因子、配置等核心领域的标准化接口契约，通过依赖倒置原则确保组件间的一致性与可替换性
- **领域驱动建模**: 提供开关、规则、因子等完整的领域模型定义，精准映射复杂业务场景的语义表达
- **可插拔驱动架构**: 基于策略模式的驱动抽象层，无缝支持 Kafka、Webhook、长轮询等异构通信模式
- **多维因子引擎**: 构建可扩展的规则评估引擎，支持用户画像、地理围栏、时间窗口等多维度决策因子的动态组合

### 工程层面

- **标准化响应协议**: 统一的响应封装与错误处理机制，集成统计、追踪、调试等可观测性元数据
- **配置管理中台**: 提供配置加载、环境适配、路径解析等配置治理能力，支持多环境配置隔离
- **全链路可观测性**: 内置统计、追踪、调试等企业级监控接口，实现端到端的性能洞察与问题诊断
- **结构化日志抽象**: 统一的日志接口定义，支持多种日志后端的无缝集成与切换

---

## 📁 项目结构

```
switch-sdk-core/
├── driver/          # 驱动抽象层
├── model/           # 领域模型层
├── factor/          # 因子引擎
├── config/          # 配置管理
├── resp/            # 响应协议层
│   └── proto/       # Protobuf 定义
├── statistics/      # 统计监控
├── logger/          # 日志抽象
├── actuator/        # 执行器调度
├── tool/            # 工具集
│   └── reflect/     # 反射工具
├── invoke/          # 调用层
│   ├── rpc/         # gRPC 配置
│   └── http/        # HTTP 调用
├── reply/           # 回复处理
├── trace/           # 链路追踪
├── transmit/        # 传输通知
└── debug/           # 调试支持
```

### 核心模块说明

- **`driver/`**: 驱动抽象层，定义通信驱动的统一接口契约与生命周期管理规范，通过驱动管理器实现多驱动的动态注册、故障转移与安全替换机制
- **`model/`**: 领域模型层，封装系统核心业务实体的完整定义，包括开关模型（SwitchModel）、规则树节点（RuleNode）、驱动配置等领域对象，为业务逻辑提供类型安全保障
- **`factor/`**: 因子引擎，提供可扩展的规则评估因子体系，内置用户ID、IP地址、地理位置、时间范围等十余种标准因子，支持自定义因子的动态注册与
  JSON Schema 验证
- **`config/`**: 配置管理中台，提供统一的配置接口抽象（ConfigI）、多源配置加载器、环境变量解析、路径解析等配置治理能力，支持开发、测试、生产环境的配置隔离
- **`resp/`**: 响应协议层，定义标准化的响应封装格式，集成 Protobuf 协议定义，提供消息构建器、响应包装器等工具，支持跨语言的高效序列化通信
- **`statistics/`**: 统计监控模块，提供性能统计数据的采集与封装能力，包括请求时间、响应时间、执行时间等关键指标，支持全链路性能分析
- **`logger/`**: 日志抽象层，定义统一的日志接口（ILogger），支持结构化日志、多级别日志输出，兼容主流日志框架的无缝集成
- **`actuator/`**: 执行器调度系统，管理因子执行器的注册表与调度逻辑，通过反射机制实现因子配置的类型安全校验与动态调用
- **`tool/`**: 工具集，提供值类型转换、IP地址处理、反射操作、随机数生成等通用工具函数，为上层业务提供基础设施支持
- **`invoke/`**: 调用层，封装 RPC 与 HTTP 调用的配置与工具，为远程服务调用提供统一的抽象接口
- **`reply/`**: 回复处理模块，定义标准化的错误码、成功响应、错误响应等回复模式，提供统一的响应 Schema 定义
- **`trace/`**: 链路追踪模块，提供分布式追踪的元数据封装，包括 TraceID、SpanID、RequestID 等追踪标识，支持全链路调用关系的可视化分析
- **`transmit/`**: 传输通知抽象，定义消息推送的统一接口（Notifier），支持多种消息传输模式的可插拔实现
- **`debug/`**: 调试支持模块，提供调试信息的封装与输出能力，支持开发阶段的问题诊断与性能分析

---

## 🔧 核心接口

### 驱动接口

驱动接口定义了通信层的统一抽象，支持多种通信模式的可插拔实现：

```go
// Driver 定义通信驱动的核心接口
type Driver interface {
    // Init 使用配置初始化驱动
    Init(config map[string]interface{}) error

    // Start 启动驱动，开始接收和发送消息
    Start(ctx context.Context) error

    // Stop 优雅停止驱动，释放资源
    Stop() error

    // Send 向指定目标发送消息
    Send(message []byte, targets []string) error

    // OnMessage 注册消息处理器
    OnMessage(handler func([]byte) error)

    // Type 返回驱动类型标识
    Type() string
}
```

### 因子接口

因子接口定义了规则评估的判断维度，支持灵活的条件组合：

```go
// Factor 定义开关评估因子的接口
type Factor interface {
    // Evaluate 评估因子条件，返回是否满足
    Evaluate(ctx context.Context, params map[string]interface{}) bool

    // Name 返回因子的唯一标识名称
    Name() string

    // Description 返回因子的功能描述
    Description() string
}
```

### 核心数据模型

开关配置模型定义了完整的开关结构：

```go
// SwitchModel 表示完整的开关配置
type SwitchModel struct {
    ID          string                 `json:"id"`          // 开关唯一标识
    Name        string                 `json:"name"`        // 开关名称
    Environment string                 `json:"environment"` // 所属环境
    Enabled     bool                   `json:"enabled"`     // 是否启用
    Rules       []RuleNode             `json:"rules"`       // 规则树
    Factors     map[string]interface{} `json:"factors"`     // 因子配置
    CreatedAt   int64                  `json:"created_at"`  // 创建时间
    UpdatedAt   int64                  `json:"updated_at"`  // 更新时间
}

// RuleNode 表示规则树节点
type RuleNode struct {
    ID        string                 `json:"id"`        // 节点ID
    Type      string                 `json:"type"`      // 节点类型（AND/OR/FACTOR）
    Condition string                 `json:"condition"` // 条件表达式
    Children  []RuleNode             `json:"children"`  // 子节点
    Factors   map[string]interface{} `json:"factors"`   // 因子参数
}
```

---

## 🚀 使用示例

### 实现自定义驱动

通过实现 Driver 接口，可以轻松扩展新的通信模式：

```go
package main

import (

"context"

"gitee.com/fatzeng/switch-sdk-core/driver"
)

// CustomDriver 自定义驱动实现
type CustomDriver struct {
    config  map[string]interface{}
    handler func([]byte) error
}

func (d *CustomDriver) Init(config map[string]interface) error {
    d.config = config
    // 初始化驱动配置
    return nil
}

func (d *CustomDriver) Start(ctx context.Context) error {
    // 启动自定义通信逻辑
    return nil
}

func (d *CustomDriver) Stop() error {
    // 优雅停止驱动
    return nil
}

func (d *CustomDriver) Send(message []byte, targets []string) error {
    // 实现消息发送逻辑
    return nil
}

func (d *CustomDriver) OnMessage(handler func([]byte) error) {
    d.handler = handler
}

func (d *CustomDriver) Type() string {
    return "custom"
}

// 注册自定义驱动
func init() {
    driver.Register("custom", func() driver.Driver {
        return &CustomDriver{}
    })
}
```

### 实现自定义因子

通过实现 Factor 接口，可以扩展新的判断维度(新的维度将属于系统层级)：

```go
package main

import (
	"context"

	"gitee.com/fatzeng/switch-sdk-core/factor"
)

// RegionFactor 基于地理区域的因子
type RegionFactor struct{}

func (f *RegionFactor) Evaluate(ctx context.Context, params map[string]interface{}) bool {
	// 获取用户所在区域
	userRegion, ok := params["region"].(string)
	if !ok {
		return false
	}

	// 获取允许的区域列表
	allowedRegions, ok := params["allowed_regions"].([]string)
	if !ok {
		return false
	}

	// 判断用户区域是否在允许列表中
	for _, region := range allowedRegions {
		if region == userRegion {
			return true
		}
	}
	return false
}

func (f *RegionFactor) Name() string {
	return "region"
}

func (f *RegionFactor) Description() string {
	return "基于用户地理区域进行评估"
}

// 注册自定义因子
func init() {
	factor.Register("region", &RegionFactor{})
}
```

---

## 🏗️ 架构设计

### 在 Switch 生态系统中的角色

```
switch-sdk-core (核心定义层)
    ↑ 被依赖
    ├── switch-admin (后端服务)
    ├── switch-sdk-go (Go SDK)
    ├── switch-components (通信组件)
```

**核心职责**：

- **接口标准化**: 为所有组件提供统一的接口规范
- **数据模型定义**: 定义系统中所有核心数据结构
- **抽象层设计**: 通过接口抽象实现组件间的松耦合

### 设计优势

1. **解耦合**: 通过抽象接口实现组件间的松耦合，提高系统灵活性
2. **可扩展**: 支持新的驱动类型、因子类型的无缝扩展
3. **标准化**: 统一的数据格式和接口规范，降低集成成本
4. **多语言支持**: 基于 Protobuf 的协议定义，支持多语言客户端
5. **企业级特性**: 内置监控、追踪、调试等企业级功能支持

---

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
