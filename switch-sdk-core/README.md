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

**Switch is a distributed real-time feature management platform** that provides secure and efficient dynamic
configuration capabilities for enterprise-grade applications. Through advanced WebSocket persistent connection
architecture and multi-driver communication mechanisms, it achieves millisecond-level configuration delivery, enabling
development teams to precisely control feature releases, user experiences, and system behaviors without restarting
applications.

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
- **Environment Management** - Independent configurations for development, testing, and production environments with
  strict configuration deployment policies
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

# Switch-SDK-Core: Core Definitions & Interfaces

`switch-sdk-core` serves as the **infrastructure layer** and **core contract layer** of the Switch ecosystem. As the
cornerstone of the entire architecture, it provides a unified type system, interface specifications, and communication
protocols for the distributed feature management platform. Through highly abstract design philosophy and contract-based
programming paradigm, it achieves loosely coupled architecture and unlimited extensibility among components, laying a
solid theoretical and practical foundation for building enterprise-grade feature governance systems.

---

## ✨ Core Capabilities

### Architecture Level

- **Contract-Based Interface System**: Establishes standardized interface contracts for core domains such as drivers,
  factors, and configurations, ensuring consistency and replaceability among components through the dependency inversion
  principle
- **Domain-Driven Modeling**: Provides complete domain model definitions for switches, rules, factors, etc., precisely
  mapping semantic expressions of complex business scenarios
- **Pluggable Driver Architecture**: Driver abstraction layer based on the strategy pattern, seamlessly supporting
  heterogeneous communication modes such as Kafka, Webhook, and Long Polling
- **Multi-Dimensional Factor Engine**: Builds an extensible rule evaluation engine supporting dynamic combinations of
  multi-dimensional decision factors such as user profiles, geofencing, and time windows

### Engineering Level

- **Standardized Response Protocol**: Unified response encapsulation and error handling mechanism, integrating
  observability metadata such as statistics, tracing, and debugging
- **Configuration Management Platform**: Provides configuration governance capabilities including configuration loading,
  environment adaptation, and path resolution, supporting multi-environment configuration isolation
- **Full-Chain Observability**: Built-in enterprise-grade monitoring interfaces for statistics, tracing, and debugging,
  enabling end-to-end performance insights and problem diagnosis
- **Structured Logging Abstraction**: Unified logging interface definition supporting seamless integration and switching
  of multiple logging backends

---

## 📁 Project Structure

```
switch-sdk-core/
├── driver/          # Driver abstraction layer
├── model/           # Domain model layer
├── factor/          # Factor engine
├── config/          # Configuration management
├── resp/            # Response protocol layer
│   └── proto/       # Protobuf definitions
├── statistics/      # Statistics monitoring
├── logger/          # Logging abstraction
├── actuator/        # Executor scheduling
├── tool/            # Utility toolkit
│   └── reflect/     # Reflection utilities
├── invoke/          # Invocation layer
│   ├── rpc/         # gRPC configuration
│   └── http/        # HTTP invocation
├── reply/           # Reply handling
├── trace/           # Distributed tracing
├── transmit/        # Transmission notification
└── debug/           # Debug support
```

### Core Module Descriptions

- **`driver/`**: Driver abstraction layer, defining unified interface contracts and lifecycle management specifications
  for communication drivers, implementing dynamic registration, failover, and safe replacement mechanisms for multiple
  drivers through the driver manager
- **`model/`**: Domain model layer, encapsulating complete definitions of core business entities in the system,
  including switch models (SwitchModel), rule tree nodes (RuleNode), driver configurations, and other domain objects,
  providing type safety guarantees for business logic
- **`factor/`**: Factor engine, providing an extensible rule evaluation factor system with over ten built-in standard
  factors such as user ID, IP address, geolocation, and time range, supporting dynamic registration of custom factors
  and JSON Schema validation
- **`config/`**: Configuration management platform, providing unified configuration interface abstraction (ConfigI),
  multi-source configuration loaders, environment variable parsing, path resolution, and other configuration governance
  capabilities, supporting configuration isolation for development, testing, and production environments
- **`resp/`**: Response protocol layer, defining standardized response encapsulation formats, integrating Protobuf
  protocol definitions, providing message builders, response wrappers, and other tools, supporting efficient
  serialization communication across languages
- **`statistics/`**: Statistics monitoring module, providing performance statistics data collection and encapsulation
  capabilities, including key metrics such as request time, response time, and execution time, supporting full-chain
  performance analysis
- **`logger/`**: Logging abstraction layer, defining unified logging interface (ILogger), supporting structured logging
  and multi-level log output, compatible with seamless integration of mainstream logging frameworks
- **`actuator/`**: Executor scheduling system, managing the registry and scheduling logic of factor executors,
  implementing type-safe validation and dynamic invocation of factor configurations through reflection mechanisms
- **`tool/`**: Utility toolkit, providing common utility functions such as value type conversion, IP address processing,
  reflection operations, and random number generation, offering infrastructure support for upper-layer business
- **`invoke/`**: Invocation layer, encapsulating configurations and tools for RPC and HTTP invocations, providing
  unified abstraction interfaces for remote service calls
- **`reply/`**: Reply handling module, defining standardized reply patterns such as error codes, success responses, and
  error responses, providing unified response Schema definitions
- **`trace/`**: Distributed tracing module, providing metadata encapsulation for distributed tracing, including trace
  identifiers such as TraceID, SpanID, and RequestID, supporting visualization analysis of full-chain call relationships
- **`transmit/`**: Transmission notification abstraction, defining unified interface (Notifier) for message pushing,
  supporting pluggable implementations of various message transmission modes
- **`debug/`**: Debug support module, providing encapsulation and output capabilities for debugging information,
  supporting problem diagnosis and performance analysis during development

---

## 🔧 Core Interfaces

### Driver Interface

The driver interface defines a unified abstraction for the communication layer, supporting pluggable implementations of
various communication modes:

```go
// Driver defines the core interface for communication drivers
type Driver interface {
    // Init initializes the driver with configuration
    Init(config map[string]interface{}) error

    // Start starts the driver to begin receiving and sending messages
    Start(ctx context.Context) error

    // Stop gracefully stops the driver and releases resources
    Stop() error

    // Send sends messages to specified targets
    Send(message []byte, targets []string) error

    // OnMessage registers a message handler
    OnMessage(handler func([]byte) error)

    // Type returns the driver type identifier
    Type() string
}
```

### Factor Interface

The factor interface defines judgment dimensions for rule evaluation, supporting flexible condition combinations:

```go
// Factor defines the interface for switch evaluation factors
type Factor interface {
    // Evaluate evaluates the factor condition and returns whether it is satisfied
    Evaluate(ctx context.Context, params map[string]interface{}) bool

    // Name returns the unique identifier name of the factor
    Name() string

    // Description returns the functional description of the factor
    Description() string
}
```

### Core Data Models

The switch configuration model defines the complete switch structure:

```go
// SwitchModel represents the complete switch configuration
type SwitchModel struct {
    ID          string                 `json:"id"`          // Switch unique identifier
    Name        string                 `json:"name"`        // Switch name
    Environment string                 `json:"environment"` // Environment
    Enabled     bool                   `json:"enabled"`     // Whether enabled
    Rules       []RuleNode             `json:"rules"`       // Rule tree
    Factors     map[string]interface{} `json:"factors"`     // Factor configuration
    CreatedAt   int64                  `json:"created_at"`  // Creation time
    UpdatedAt   int64                  `json:"updated_at"`  // Update time
}

// RuleNode represents a rule tree node
type RuleNode struct {
    ID        string                 `json:"id"`        // Node ID
    Type      string                 `json:"type"`      // Node type (AND/OR/FACTOR)
    Condition string                 `json:"condition"` // Condition expression
    Children  []RuleNode             `json:"children"`  // Child nodes
    Factors   map[string]interface{} `json:"factors"`   // Factor parameters
}
```

---

## 🚀 Usage Examples

### Implementing a Custom Driver

By implementing the Driver interface, you can easily extend new communication modes:

```go
package main

import (

"context"

"gitee.com/fatzeng/switch-sdk-core/driver"
)

// CustomDriver custom driver implementation
type CustomDriver struct {
    config  map[string]interface{}
    handler func([]byte) error
}

func (d *CustomDriver) Init(config map[string]interface) error {
    d.config = config
    // Initialize driver configuration
    return nil
}

func (d *CustomDriver) Start(ctx context.Context) error {
    // Start custom communication logic
    return nil
}

func (d *CustomDriver) Stop() error {
    // Gracefully stop the driver
    return nil
}

func (d *CustomDriver) Send(message []byte, targets []string) error {
    // Implement message sending logic
    return nil
}

func (d *CustomDriver) OnMessage(handler func([]byte) error) {
    d.handler = handler
}

func (d *CustomDriver) Type() string {
    return "custom"
}

// Register custom driver
func init() {
    driver.Register("custom", func() driver.Driver {
        return &CustomDriver{}
    })
}
```

### Implementing a Custom Factor

By implementing the Factor interface, you can extend new judgment dimensions (new dimensions will belong to the system
level):

```go
package main

import (
	"context"

	"gitee.com/fatzeng/switch-sdk-core/factor"
)

// RegionFactor factor based on geographic region
type RegionFactor struct{}

func (f *RegionFactor) Evaluate(ctx context.Context, params map[string]interface{}) bool {
	// Get user's region
	userRegion, ok := params["region"].(string)
	if !ok {
		return false
	}

	// Get allowed region list
	allowedRegions, ok := params["allowed_regions"].([]string)
	if !ok {
		return false
	}

	// Check if user's region is in the allowed list
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
	return "Evaluate based on user's geographic region"
}

// Register custom factor
func init() {
	factor.Register("region", &RegionFactor{})
}
```

---

## 🏗️ Architecture Design

### Role in the Switch Ecosystem

```
switch-sdk-core (Core Definition Layer)
    ↑ Depended upon by
    ├── switch-admin (Backend Service)
    ├── switch-sdk-go (Go SDK)
    ├── switch-components (Communication Components)
```

**Core Responsibilities**:

- **Interface Standardization**: Provides unified interface specifications for all components
- **Data Model Definition**: Defines all core data structures in the system
- **Abstraction Layer Design**: Achieves loose coupling among components through interface abstraction

### Design Advantages

1. **Decoupling**: Achieves loose coupling among components through abstract interfaces, improving system flexibility
2. **Extensibility**: Supports seamless extension of new driver types and factor types
3. **Standardization**: Unified data formats and interface specifications reduce integration costs
4. **Multi-Language Support**: Protocol definitions based on Protobuf support multi-language clients
5. **Enterprise-Grade Features**: Built-in support for enterprise-grade features such as monitoring, tracing, and
   debugging

---

## 🤝 Contributing

We welcome and appreciate all forms of contributions! Whether it's reporting issues, suggesting features, improving
documentation, or submitting code, your participation will help make Switch better.

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
