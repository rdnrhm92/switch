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

**Switch是一个分布式实时特性管控平台**，为企业级应用提供安全、高效的动态配置能力。它通过先进的WebSocket长连接架构和多驱动通信机制，实现了配置变更的毫秒级下发，让开发团队能够在不重启应用的情况下，精确控制功能发布、用户体验和系统行为。

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

# Switch-Client-Demo: 示例应用与最佳实践

`switch-client-demo` 是 Switch 生态系统的**参考实现**与**集成范例**，为开发者提供了在生产环境中集成 Switch SDK 的完整指南。通过精心设计的示例代码和最佳实践模式，展示了如何构建具备动态特性管控能力的企业级应用，涵盖从基础配置到高级特性的全方位实践场景。

---

## ✨ 核心特性

### 集成示例

- **SDK 初始化配置**: 展示完整的 SDK 初始化流程，包括连接配置、命名空间设置、环境标识等关键参数
- **日志系统集成**: 演示结构化日志的配置与使用，支持文件轮转、级别控制、堆栈追踪等企业级日志特性
- **上下文管理**: 展示如何通过 Context 传递用户信息、请求元数据等业务上下文
- **因子缓存中间件**: 演示因子执行缓存的启用场景，适用于复杂因子或并发执行场景

### 功能演示

- **特性开关查询**: 展示基于上下文的动态特性开关查询机制
- **实时配置更新**: 演示配置变更的实时推送与自动生效能力
- **优雅关闭处理**: 展示应用关闭时的资源清理与连接释放最佳实践
- **信号处理机制**: 演示系统信号的捕获与优雅退出流程

---

## 📁 项目结构

```
switch-client-demo/
├── main.go          # 主程序入口，包含完整的集成示例
├── go.mod           # Go 模块依赖定义
├── go.sum           # 依赖版本锁定文件
├── logs/            # 日志输出目录（运行时生成）
```

### 核心文件说明

- **`main.go`**: 主程序文件，包含 SDK 初始化、日志配置、开关查询、信号处理等完整示例代码，展示了生产级应用的标准集成模式

---

## 🚀 快速开始

### 前置要求

- Go 1.18 或更高版本
- 可访问的 Switch Admin 服务实例
- 已配置的命名空间和环境

### 安装依赖

```bash
go mod download
```

### 配置说明

在 `main.go` 中修改以下配置参数以匹配您的环境：

```go
// SDK 初始化配置
err = switchsdk.Start(ctx,
    _switch.WithDomain("ws://localhost:8081"),      // Switch Admin 服务地址
    _switch.WithNamespaceTag("test-ns"),            // 命名空间标识
    _switch.WithEnvTag("uat"),                      // 环境标识（dev/uat/prod）
    _switch.WithServiceName("simple-demo"),         // 服务名称
    _switch.WithVersion("1.0.0"),                   // 服务版本
)
```

### 运行示例

```bash
go run main.go
```

程序将：
1. 初始化日志系统，输出到控制台和 `./logs` 目录
2. 连接到 Switch Admin 服务并建立 WebSocket 长连接
3. 每 3 秒查询一次 `feature_enabled` 开关状态
4. 实时响应配置变更
5. 等待 Ctrl+C 信号进行优雅关闭

---

## 💡 核心代码解析

### 1. 日志系统配置

```go
log, err := logging.New(&logger.LoggerConfig{
    Level:            "info",                     // 日志级别：debug/info/warn/error
    OutputDir:        "./logs",                   // 日志文件输出目录
    FileNameFormat:   "switch-demo_%Y-%m-%d.log", // 日志文件命名格式
    MaxSize:          50,                         // 单个日志文件最大大小(MB)
    MaxBackups:       3,                          // 保留的旧日志文件数量
    MaxAge:           7,                          // 日志文件保留天数
    Compress:         false,                      // 是否压缩旧日志文件
    ShowCaller:       true,                       // 是否显示调用者信息
    EnableConsole:    true,                       // 是否启用控制台输出
    EnableJSON:       false,                      // 是否使用JSON格式
    EnableStackTrace: true,                       // 是否启用堆栈跟踪
    StackTraceLevel:  "error",                    // 堆栈跟踪级别
    TimeFormat:       "2006-01-02 15:04:05",      // 时间格式
})
```

**关键特性**：
- 支持日志文件自动轮转，避免单文件过大
- 可配置的日志保留策略，自动清理过期日志
- 灵活的输出格式，支持控制台和文件双输出
- 错误级别自动记录堆栈信息，便于问题诊断

### 2. SDK 初始化

```go
err = switchsdk.Start(ctx,
    _switch.WithDomain("ws://localhost:8081"),      // 服务端地址
    _switch.WithNamespaceTag("test-ns"),            // 命名空间
    _switch.WithEnvTag("uat"),                      // 环境标识
    _switch.WithServiceName("simple-demo"),         // 服务名称
    _switch.WithVersion("1.0.0"),                   // 版本号
)
```

**配置说明**：
- **Domain**: Switch Admin 服务的 WebSocket 地址，支持 ws:// 和 wss:// 协议
- **NamespaceTag**: 命名空间标识，用于多租户隔离
- **EnvTag**: 环境标识，支持自定义
- **ServiceName**: 服务名称，用于标识客户端身份
- **Version**: 服务版本号，便于版本管理和追踪

### 3. 上下文管理与缓存

```go
// 启用因子执行缓存中间件（可选，非必要不建议开启）
ctx = cache.UseCache(ctx)

// 注入业务上下文
ctx = context.WithValue(ctx, "user_name", "张三")
```

**最佳实践**：
- `cache.UseCache(ctx)` 启用因子执行缓存中间件，**非必要不建议开启**，可能导致性能下降
- 适用场景：单个开关内有多个相同因子配置，或并发执行同一个开关
- 通过 Context 传递用户信息、请求 ID 等业务上下文，支持基于上下文的规则评估
- Context 中的数据可被因子系统用于动态决策

### 4. 特性开关查询

```go
func testSwitches(ctx context.Context) {
    switchName := "feature_enabled"
    if _switch.IsOpen(ctx, switchName) {
        logger.Logger.Infof("Switch '%s' is ON", switchName)
        // 执行新特性逻辑
    } else {
        logger.Logger.Infof("Switch '%s' is OFF", switchName)
        // 执行默认逻辑
    }
}
```

**核心机制**：
- `IsOpen()` 方法会根据配置的规则树和因子进行实时评估
- 支持基于用户属性、地理位置、时间窗口等多维度的动态决策
- 配置变更会通过 WebSocket 实时推送，无需重启应用

### 5. 优雅关闭

```go
// 监听系统信号
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan

logger.Logger.Info("Received shutdown signal")
switchsdk.Shutdown()  // 关闭 SDK，释放资源
logger.Logger.Info("Demo stopped")
```

**关键要点**：
- 捕获 SIGINT（Ctrl+C）和 SIGTERM 信号
- 调用 `switchsdk.Shutdown()` 优雅关闭 WebSocket 连接
- 确保日志完整写入，避免数据丢失

---

## 🎯 使用场景

### 1. 功能灰度发布

```go
// 基于用户 ID 的灰度发布
ctx = context.WithValue(ctx, "user_id", "12345")
if _switch.IsOpen(ctx, "new_feature") {
    // 新功能逻辑
} else {
    // 旧功能逻辑
}
```

### 2. A/B 测试

```go
// 基于用户分组的 A/B 测试
ctx = context.WithValue(ctx, "user_group", "group_a")
if _switch.IsOpen(ctx, "ab_test_feature") {
    // A 组体验
} else {
    // B 组体验
}
```

### 3. 紧急开关

```go
// 紧急关闭某个功能
if _switch.IsOpen(ctx, "emergency_disable") {
    // 功能已被紧急关闭
    return errors.New("feature temporarily disabled")
}
```

### 4. 环境隔离

```go
// 不同环境使用不同配置
// 通过 WithEnvTag 指定环境，自动加载对应环境的开关配置
```
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
