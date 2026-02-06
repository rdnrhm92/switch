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

# Switch-SDK-Go: Go 业务集成 SDK

`switch-sdk-go` 是 Switch 生态系统的 Go 语言业务集成层，为 Go 应用提供开箱即用的特性开关能力。它封装了复杂的通信协议和数据同步逻辑，通过简洁的
API 让开发者能够轻松实现基于多维度因子的智能特性控制，支持实时配置更新和多种通信模式。

---

## ✨ 功能特性

- **简洁易用的 API**: 提供 `IsOpen()` 和 `IsSwitchOpen()` 两种开关检查方式，满足不同场景需求
- **实时配置同步**: 基于 WebSocket 的实时配置推送，配置变更毫秒级生效，无需重启应用
- **多通信模式支持**:
    - **Kafka 模式**: 消息队列分发，适合大规模分布式场景
    - **Webhook 模式**: HTTP 回调推送，灵活适配各种网络环境
    - **长轮询模式**: 主动拉取配置，兼容受限网络环境
- **智能缓存机制**:
    - 因子级别缓存，避免重复计算
    - Singleflight 模式防止缓存击穿
    - 线程安全的规则存储
- **完整的上下文支持**: 原生支持 Go Context，实现请求级别的追踪和控制
- **高可用保障**:
    - 自动重连机制
    - 优雅降级策略
    - 本地缓存回退
- **企业级监控**: 内置统计指标和性能监控，支持开关执行的全链路追踪
- **灵活的中间件架构**: 可插拔的中间件系统，支持自定义扩展

---

## 🚀 快速开始

### 安装依赖

```bash
go get gitee.com/fatzeng/switch-sdk-go
```

### 基本使用

```go
package main

import (
	"context"
	"log"

	"gitee.com/fatzeng/switch-sdk-core/model"
	switchsdk "gitee.com/fatzeng/switch-sdk-go"
	_switch "gitee.com/fatzeng/switch-sdk-go/core/switch"
)

func main() {
	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化 Switch SDK
	err := switchsdk.Start(ctx,
		_switch.WithDomain("ws://localhost:8081"),  // Switch Admin 服务地址
		_switch.WithNamespaceTag("your-namespace"), // 命名空间标识
		_switch.WithEnvTag("production"),           // 环境标识
		_switch.WithServiceName("your-service"),    // 服务名称
		_switch.WithVersion("1.0.0"),               // 服务版本
	)
	if err != nil {
		log.Fatalf("Failed to initialize Switch SDK: %v", err)
	}
	defer switchsdk.Shutdown()

	// 方式一：简单的开关检查（通过开关名称）
	if _switch.IsOpen(ctx, "new-feature") {
		// 新功能已启用
		log.Println("使用新功能")
	} else {
		// 新功能未启用
		log.Println("使用旧功能")
	}

	// 方式二：基于开关模型的检查（支持更复杂的场景）
	switchModel := &model.SwitchModel{
		Name: "advanced-feature",
		// 可以设置更多属性用于因子计算
	}
	if _switch.IsSwitchOpen(ctx, switchModel) {
		log.Println("高级功能已启用")
	}
}
```

### 高级用法

#### 使用缓存优化性能

```go
import "gitee.com/fatzeng/switch-sdk-go/core/cache"

// 启用因子缓存，避免重复计算
ctx = cache.UseCache(ctx)

// 后续的开关检查会自动使用缓存
if _switch.IsOpen(ctx, "cached-feature") {
    // 缓存命中时，直接返回结果，无需重新计算
}
```

#### 自定义配置选项

```go
err := switchsdk.Start(ctx,
    // 基础配置
    _switch.WithDomain("ws://switch-admin.example.com"),
    _switch.WithNamespaceTag("production-ns"),
    _switch.WithEnvTag("prod"),
    _switch.WithServiceName("order-service"),
    _switch.WithVersion("2.1.0"),
)
```

---

## 📁 项目结构

```
switch-sdk-go/
├── start.go                    # SDK 主入口，提供初始化和生命周期管理
├── core/                       # 核心功能模块
│   ├── switch/                 # 开关核心引擎
│   ├── cache/                  # 智能缓存系统
│   ├── filter/                 # 过滤器系统
│   ├── factor/                 # 因子处理模块
│   ├── factor_statistics/      # 统计和监控
│   └── middleware/             # 中间件框架
├── internal/                   # 内部实现（不对外暴露）
│   └── datasync/               # 数据同步机制
└── go.mod                      # Go 模块依赖
```

### 关键组件说明：

- **`start.go`**: SDK 的统一入口，负责初始化、配置管理和生命周期控制
- **`core/switch/`**: 开关评估的核心引擎，实现基于规则和因子的智能决策
- **`core/cache/`**: 高性能缓存系统，使用 singleflight 模式防止缓存击穿
- **`core/filter/`**: 开关执行的流程控制，支持自定义过滤逻辑
- **`core/factor/`**: 因子处理和计算逻辑
- **`core/factor_statistics/`**: 开关执行的统计和性能监控
- **`core/middleware/`**: 可扩展的中间件架构，支持链式处理
- **`internal/datasync/`**: 数据同步核心，处理配置的实时更新和持久化

---

## 🏗️ 架构设计

### 依赖关系

```
switch-sdk-go (业务集成层)
    ↓ 依赖
switch-sdk-core (核心定义层)
    - 提供数据模型（SwitchModel、RuleNode）
    - 定义统一接口和协议
    ↓ 依赖
switch-components (通信组件层)
    - 提供 WebSocket 客户端
    - 实现多种驱动（Kafka、Webhook、Polling）
    - 提供网络通信基础设施
```

### 数据流向

```
开关评估流程：
业务应用 → switch-sdk-go.IsOpen() → 规则引擎 → 因子计算 → 缓存 → 返回结果
```

---

## 🔧 核心概念

### 开关（Switch）

开关是特性控制的基本单元，每个开关包含：

- **名称**: 唯一标识符
- **规则**: 决策逻辑（AND/OR 组合）
- **因子**: 多维度判断条件（用户属性、地理位置、时间等）

### 因子（Factor）

因子是开关决策的判断维度，支持：

- 用户属性因子（用户ID、用户组、VIP等级等）
- 地理位置因子（国家、城市、IP段等）
- 时间因子（时间窗口、日期范围等）
- 自定义因子（业务特定的判断逻辑）

### 规则（Rule）

规则定义了因子的组合方式：

- **AND 规则**: 所有因子都满足时开关才开启
- **OR 规则**: 任一因子满足时开关即开启
- **嵌套规则**: 支持复杂的逻辑组合

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
