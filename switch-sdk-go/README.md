# Switch: Dynamic Feature Flag & Remote Configuration System

<div align="center">

<img src="switch.svg" alt="Switch Logo" width="300">

**A Powerful Real-Time Feature Flag System Built for Modern Application Development**

[![Go Version](https://img.shields.io/badge/Go-1.18+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://gitee.com/fatzeng/collections/413616)

[English](README.md) | [中文](README_zh.md)

</div>

---

## 🎯 What is Switch?

**Switch is a distributed real-time feature management platform** that provides secure, efficient dynamic configuration
capabilities for enterprise applications. Through advanced WebSocket persistent connection architecture and multi-driver
communication mechanisms, it achieves millisecond-level configuration delivery, enabling development teams to precisely
control feature releases, user experiences, and system behavior without restarting applications.

### 🏗️ Core Architecture Advantages

**1. Enterprise-Grade Communication Architecture**

- **WebSocket Persistent Connection Framework** - Based on `switch-components/pc` for persistent connection management
- **Multi-Driver Support** - Three communication modes: Webhook, Kafka, and Long Polling
- **Intelligent Network Discovery** - Automatic adaptation to NAT environments, solving complex network scenarios
- **Layered Confirmation Mechanism** - Ensuring reliability and consistency of configuration delivery

**2. Flexible Factor System**

```go
// Not just simple true/false, but intelligent decision-making based on complex rules
if _switch.IsOpen(ctx, "feature_enabled") {
    // The system calculates in real-time based on configured multi-dimensional factors
    // (e.g., user attributes, geographic location, time windows, etc.)
    // whether to enable this feature
}
```

**3. Multi-Tenant Management System**

- **Tenant Isolation** - Complete data and permission isolation
- **Environment Management** - Independent configurations for development, testing, and production environments with
  strict configuration promotion policies
- **Approval Workflows** - Multi-level approval mechanism for sensitive changes

## System Architecture

The Switch ecosystem consists of the following core components:

- **switch-admin**: Backend service responsible for managing configurations and client communication
- **switch-frontend**: Web interface for configuration management
- **switch-sdk-go**: Go SDK for integrating switches into business applications to form clients
- **switch-sdk-core**: Core definitions and interfaces
- **switch-components**: Implementation of communication and core logic
- **switch-client-demo**: Example application demonstrations

---

# Switch-SDK-Go: Go Business Integration SDK

`switch-sdk-go` is the Go language business integration layer of the Switch ecosystem, providing out-of-the-box feature
flag capabilities for Go applications. It encapsulates complex communication protocols and data synchronization logic,
enabling developers to easily implement intelligent feature control based on multi-dimensional factors through a concise
API, supporting real-time configuration updates and multiple communication modes.

---

## ✨ Features

- **Simple and Easy-to-Use API**: Provides two switch checking methods, `IsOpen()` and `IsSwitchOpen()`, to meet
  different scenario requirements
- **Real-Time Configuration Synchronization**: WebSocket-based real-time configuration push with millisecond-level
  effectiveness, no application restart required
- **Multi-Communication Mode Support**:
    - **Kafka Mode**: Message queue distribution, suitable for large-scale distributed scenarios
    - **Webhook Mode**: HTTP callback push, flexibly adapting to various network environments
    - **Long Polling Mode**: Active configuration pulling, compatible with restricted network environments
- **Intelligent Caching Mechanism**:
    - Factor-level caching to avoid redundant calculations
    - Singleflight pattern to prevent cache stampede
    - Thread-safe rule storage
- **Complete Context Support**: Native support for Go Context, enabling request-level tracing and control
- **High Availability Guarantee**:
    - Automatic reconnection mechanism
    - Graceful degradation strategy
    - Local cache fallback
- **Enterprise-Grade Monitoring**: Built-in statistical metrics and performance monitoring, supporting full-chain
  tracing of switch execution
- **Flexible Middleware Architecture**: Pluggable middleware system supporting custom extensions

---

## 🚀 Quick Start

### Installation

```bash
go get gitee.com/fatzeng/switch-sdk-go
```

### Basic Usage

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
	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Switch SDK
	err := switchsdk.Start(ctx,
		_switch.WithDomain("ws://localhost:8081"),  // Switch Admin service address
		_switch.WithNamespaceTag("your-namespace"), // Namespace identifier
		_switch.WithEnvTag("production"),           // Environment identifier
		_switch.WithServiceName("your-service"),    // Service name
		_switch.WithVersion("1.0.0"),               // Service version
	)
	if err != nil {
		log.Fatalf("Failed to initialize Switch SDK: %v", err)
	}
	defer switchsdk.Shutdown()

	// Method 1: Simple switch check (by switch name)
	if _switch.IsOpen(ctx, "new-feature") {
		// New feature is enabled
		log.Println("Using new feature")
	} else {
		// New feature is disabled
		log.Println("Using old feature")
	}

	// Method 2: Switch model-based check (supports more complex scenarios)
	switchModel := &model.SwitchModel{
		Name: "advanced-feature",
		// Can set more attributes for factor calculation
	}
	if _switch.IsSwitchOpen(ctx, switchModel) {
		log.Println("Advanced feature is enabled")
	}
}
```

### Advanced Usage

#### Using Cache to Optimize Performance

```go
import "gitee.com/fatzeng/switch-sdk-go/core/cache"

// Enable factor caching to avoid redundant calculations
ctx = cache.UseCache(ctx)

// Subsequent switch checks will automatically use cache
if _switch.IsOpen(ctx, "cached-feature") {
    // When cache hits, return result directly without recalculation
}
```

#### Custom Configuration Options

```go
err := switchsdk.Start(ctx,
    // Basic configuration
    _switch.WithDomain("ws://switch-admin.example.com"),
    _switch.WithNamespaceTag("production-ns"),
    _switch.WithEnvTag("prod"),
    _switch.WithServiceName("order-service"),
    _switch.WithVersion("2.1.0"),
)
```

---

## 📁 Project Structure

```
switch-sdk-go/
├── start.go                    # SDK main entry, providing initialization and lifecycle management
├── core/                       # Core functional modules
│   ├── switch/                 # Switch core engine
│   ├── cache/                  # Intelligent caching system
│   ├── filter/                 # Filter system
│   ├── factor/                 # Factor processing module
│   ├── factor_statistics/      # Statistics and monitoring
│   └── middleware/             # Middleware framework
├── internal/                   # Internal implementation (not exposed)
│   └── datasync/               # Data synchronization mechanism
└── go.mod                      # Go module dependencies
```

### Key Components:

- **`start.go`**: Unified SDK entry point, responsible for initialization, configuration management, and lifecycle
  control
- **`core/switch/`**: Core engine for switch evaluation, implementing intelligent decision-making based on rules and
  factors
- **`core/cache/`**: High-performance caching system using singleflight pattern to prevent cache stampede
- **`core/filter/`**: Flow control for switch execution, supporting custom filtering logic
- **`core/factor/`**: Factor processing and calculation logic
- **`core/factor_statistics/`**: Statistics and performance monitoring for switch execution
- **`core/middleware/`**: Extensible middleware architecture supporting chain processing
- **`internal/datasync/`**: Data synchronization core, handling real-time configuration updates and persistence

---

## 🏗️ Architecture Design

### Dependency Relationships

```
switch-sdk-go (Business Integration Layer)
    ↓ depends on
switch-sdk-core (Core Definition Layer)
    - Provides data models (SwitchModel, RuleNode)
    - Defines unified interfaces and protocols
    ↓ depends on
switch-components (Communication Component Layer)
    - Provides WebSocket client
    - Implements multiple drivers (Kafka, Webhook, Polling)
    - Provides network communication infrastructure
```

### Data Flow

```
Switch Evaluation Flow:
Business Application → switch-sdk-go.IsOpen() → Rule Engine → Factor Calculation → Cache → Return Result
```

---

## 🔧 Core Concepts

### Switch

A switch is the basic unit of feature control. Each switch contains:

- **Name**: Unique identifier
- **Rules**: Decision logic (AND/OR combinations)
- **Factors**: Multi-dimensional judgment conditions (user attributes, geographic location, time, etc.)

### Factor

Factors are the judgment dimensions for switch decisions, supporting:

- User attribute factors (user ID, user group, VIP level, etc.)
- Geographic location factors (country, city, IP range, etc.)
- Time factors (time windows, date ranges, etc.)
- Custom factors (business-specific judgment logic)

### Rule

Rules define how factors are combined:

- **AND Rule**: Switch opens only when all factors are satisfied
- **OR Rule**: Switch opens when any factor is satisfied
- **Nested Rules**: Support for complex logical combinations

---

## 🤝 Contributing

We welcome and appreciate all forms of contributions! Whether it's reporting issues, suggesting features, improving
documentation, or submitting code, your participation will help make Switch better.

### How to Contribute

1. **Fork this repository** and create your feature branch
2. **Write code** and ensure it follows the project's coding standards
3. **Add tests** to cover your changes
4. **Submit a Pull Request** with a detailed description of your changes and motivation

### Contribution Types

- 🐛 **Bug Fixes**: Discover and fix issues
- ✨ **New Features**: Propose and implement new capabilities
- 📝 **Documentation Improvements**: Enhance documentation and examples
- 🎨 **Code Optimization**: Improve code quality and performance
- 🧪 **Test Enhancement**: Increase test coverage

For more details, please refer to our contribution guidelines documentation.

## 📄 License

This project is licensed under the [MIT License](LICENSE).
