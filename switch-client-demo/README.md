# Switch: Dynamic Feature Toggle & Remote Configuration System

<div align="center">

<img src="switch.svg" alt="Switch Logo" width="300">

**Powerful Real-time Feature Toggle System for Modern Application Development**

[![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://gitee.com/fatzeng/collections/413616)

[English](README.md) | [中文](README_zh.md)

</div>

---

## 🎯 What is Switch?

**Switch is a distributed real-time feature management platform** that provides secure and efficient dynamic configuration capabilities for enterprise-grade applications. Through advanced WebSocket persistent connection architecture and multi-driver communication mechanisms, it achieves millisecond-level configuration delivery, enabling development teams to precisely control feature releases, user experiences, and system behaviors without restarting applications.

### 🏗️ Core Architecture Advantages

**1. Enterprise-Grade Communication Architecture**

- **WebSocket Persistent Connection Framework** - Based on `switch-components/pc` for persistent connection management
- **Multi-Driver Support** - Three communication modes: Webhook, Kafka, and Long Polling
- **Intelligent Network Discovery** - Automatic NAT environment adaptation for complex network scenarios
- **Layered Acknowledgment Mechanism** - Ensures reliability and consistency of configuration delivery

**2. Flexible Factor System**

```go
// Not just simple true/false, but intelligent decision-making based on complex rules
if _switch.IsOpen(ctx, "feature_enabled") {
    // The system calculates in real-time whether to enable the feature
    // based on configured multi-dimensional factors (e.g., user attributes, geolocation, time windows, etc.)
}
```

**3. Multi-Tenant Management System**

- **Tenant Isolation** - Complete data and permission isolation
- **Environment Management** - Independent configurations for development, testing, and production environments with strict configuration deployment policies
- **Approval Workflow** - Multi-level approval mechanism for sensitive changes

## System Architecture

The Switch ecosystem consists of the following core components:

- **switch-admin**: Backend service responsible for configuration management and client communication
- **switch-frontend**: Web interface for configuration management
- **switch-sdk-go**: Go SDK for integrating switches into business applications to form clients
- **switch-sdk-core**: Core definitions and interfaces
- **switch-components**: Implementation of communication and core logic
- **switch-client-demo**: Demo application examples

---

# Switch-Client-Demo: Example Application & Best Practices

`switch-client-demo` serves as the **reference implementation** and **integration example** of the Switch ecosystem, providing developers with a complete guide for integrating Switch SDK in production environments. Through carefully designed sample code and best practice patterns, it demonstrates how to build enterprise-grade applications with dynamic feature management capabilities, covering comprehensive practical scenarios from basic configuration to advanced features.

---

## ✨ Core Features

### Integration Examples

- **SDK Initialization Configuration**: Demonstrates the complete SDK initialization process, including connection configuration, namespace settings, environment identifiers, and other key parameters
- **Logging System Integration**: Demonstrates structured logging configuration and usage, supporting enterprise-grade logging features such as file rotation, level control, and stack tracing
- **Context Management**: Shows how to pass user information, request metadata, and other business context through Context
- **Factor Cache Middleware**: Demonstrates factor execution cache usage scenarios, suitable for complex factors or concurrent execution scenarios

### Functionality Demonstrations

- **Feature Switch Queries**: Demonstrates context-based dynamic feature switch query mechanisms
- **Real-time Configuration Updates**: Demonstrates real-time push and automatic activation of configuration changes
- **Graceful Shutdown Handling**: Shows best practices for resource cleanup and connection release during application shutdown
- **Signal Handling Mechanism**: Demonstrates system signal capture and graceful exit processes

---

## 📁 Project Structure

```
switch-client-demo/
├── main.go          # Main program entry with complete integration examples
├── go.mod           # Go module dependency definitions
├── go.sum           # Dependency version lock file
├── logs/            # Log output directory (generated at runtime)
```

### Core File Descriptions

- **`main.go`**: Main program file containing SDK initialization, logging configuration, switch queries, signal handling, and other complete sample code, demonstrating standard integration patterns for production-grade applications

---

## 🚀 Quick Start

### Prerequisites

- Go 1.18 or higher
- Accessible Switch Admin service instance
- Configured namespace and environment

### Install Dependencies

```bash
go mod download
```

### Configuration Instructions

Modify the following configuration parameters in `main.go` to match your environment:

```go
// SDK initialization configuration
err = switchsdk.Start(ctx,
    _switch.WithDomain("ws://localhost:8081"),      // Switch Admin service address
    _switch.WithNamespaceTag("test-ns"),            // Namespace identifier
    _switch.WithEnvTag("uat"),                      // Environment identifier (dev/uat/prod)
    _switch.WithServiceName("simple-demo"),         // Service name
    _switch.WithVersion("1.0.0"),                   // Service version
)
```

### Run the Example

```bash
go run main.go
```

The program will:
1. Initialize the logging system, outputting to console and `./logs` directory
2. Connect to Switch Admin service and establish WebSocket persistent connection
3. Query the `feature_enabled` switch status every 3 seconds
4. Respond to configuration changes in real-time
5. Wait for Ctrl+C signal for graceful shutdown

---

## 💡 Core Code Analysis

### 1. Logging System Configuration

```go
log, err := logging.New(&logger.LoggerConfig{
    Level:            "info",                     // Log level: debug/info/warn/error
    OutputDir:        "./logs",                   // Log file output directory
    FileNameFormat:   "switch-demo_%Y-%m-%d.log", // Log file naming format
    MaxSize:          50,                         // Maximum size of single log file (MB)
    MaxBackups:       3,                          // Number of old log files to retain
    MaxAge:           7,                          // Log file retention days
    Compress:         false,                      // Whether to compress old log files
    ShowCaller:       true,                       // Whether to show caller information
    EnableConsole:    true,                       // Whether to enable console output
    EnableJSON:       false,                      // Whether to use JSON format
    EnableStackTrace: true,                       // Whether to enable stack tracing
    StackTraceLevel:  "error",                    // Stack trace level
    TimeFormat:       "2006-01-02 15:04:05",      // Time format
})
```

**Key Features**:
- Supports automatic log file rotation to avoid oversized single files
- Configurable log retention policy with automatic cleanup of expired logs
- Flexible output formats supporting both console and file output
- Automatic stack trace recording for error levels for problem diagnosis

### 2. SDK Initialization

```go
err = switchsdk.Start(ctx,
    _switch.WithDomain("ws://localhost:8081"),      // Server address
    _switch.WithNamespaceTag("test-ns"),            // Namespace
    _switch.WithEnvTag("uat"),                      // Environment identifier
    _switch.WithServiceName("simple-demo"),         // Service name
    _switch.WithVersion("1.0.0"),                   // Version number
)
```

**Configuration Details**:
- **Domain**: WebSocket address of Switch Admin service, supporting ws:// and wss:// protocols
- **NamespaceTag**: Namespace identifier for multi-tenant isolation
- **EnvTag**: Environment identifier, supports custom values
- **ServiceName**: Service name for identifying client identity
- **Version**: Service version number for version management and tracking

### 3. Context Management and Caching

```go
// Enable factor execution cache middleware (optional, not recommended unless necessary)
ctx = cache.UseCache(ctx)

// Inject business context
ctx = context.WithValue(ctx, "user_name", "Zhang San")
```

**Best Practices**:
- `cache.UseCache(ctx)` enables factor execution cache middleware, **not recommended unless necessary**, may cause performance degradation
- Applicable scenarios: multiple identical factor configurations within a single switch, or concurrent execution of the same switch
- Pass user information, request IDs, and other business context through Context to support context-based rule evaluation
- Data in Context can be used by the factor system for dynamic decision-making

### 4. Feature Switch Queries

```go
func testSwitches(ctx context.Context) {
    switchName := "feature_enabled"
    if _switch.IsOpen(ctx, switchName) {
        logger.Logger.Infof("Switch '%s' is ON", switchName)
        // Execute new feature logic
    } else {
        logger.Logger.Infof("Switch '%s' is OFF", switchName)
        // Execute default logic
    }
}
```

**Core Mechanism**:
- The `IsOpen()` method performs real-time evaluation based on configured rule trees and factors
- Supports dynamic decision-making based on multiple dimensions such as user attributes, geolocation, and time windows
- Configuration changes are pushed in real-time via WebSocket without requiring application restart

### 5. Graceful Shutdown

```go
// Listen for system signals
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan

logger.Logger.Info("Received shutdown signal")
switchsdk.Shutdown()  // Shutdown SDK and release resources
logger.Logger.Info("Demo stopped")
```

**Key Points**:
- Capture SIGINT (Ctrl+C) and SIGTERM signals
- Call `switchsdk.Shutdown()` to gracefully close WebSocket connections
- Ensure logs are fully written to avoid data loss

---

## 🎯 Use Cases

### 1. Feature Gradual Rollout

```go
// Gradual rollout based on user ID
ctx = context.WithValue(ctx, "user_id", "12345")
if _switch.IsOpen(ctx, "new_feature") {
    // New feature logic
} else {
    // Old feature logic
}
```

### 2. A/B Testing

```go
// A/B testing based on user groups
ctx = context.WithValue(ctx, "user_group", "group_a")
if _switch.IsOpen(ctx, "ab_test_feature") {
    // Group A experience
} else {
    // Group B experience
}
```

### 3. Emergency Switch

```go
// Emergency disable a feature
if _switch.IsOpen(ctx, "emergency_disable") {
    // Feature has been emergency disabled
    return errors.New("feature temporarily disabled")
}
```

### 4. Environment Isolation

```go
// Use different configurations for different environments
// Specify environment via WithEnvTag to automatically load corresponding environment's switch configurations
```

---

## 🤝 Contributing

We welcome and appreciate all forms of contributions! Whether it's reporting issues, suggesting features, improving documentation, or submitting code, your participation will help make Switch better.

### How to Contribute

1. **Fork this repository** and create your feature branch
2. **Write code** and ensure it follows the project's coding standards
3. **Add tests** to cover your changes
4. **Submit a Pull Request** with detailed descriptions of your changes and motivations

### Contribution Types

- 🐛 **Bug Fixes**: Discover and fix issues
- ✨ **New Features**: Propose and implement new features
- 📝 **Documentation Improvements**: Enhance documentation and examples
- 🎨 **Code Optimization**: Improve code quality and performance
- 🧪 **Test Enhancement**: Increase test coverage

For more details, please refer to our contribution guidelines.

## 📄 License

This project is licensed under the [MIT License](LICENSE).
