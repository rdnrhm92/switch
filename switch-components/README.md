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

# Switch-Components: Communication & Core Logic Implementation

`switch-components` is the core component library of the Switch ecosystem, providing complete implementations of
enterprise-grade infrastructure including WebSocket communication framework, multi-driver system, distributed ID
generation, log tracing, and more. It transforms the abstract interfaces defined in `switch-sdk-core` into reliable,
high-performance engineering implementations, providing out-of-the-box technical capabilities for upper-layer
applications.

---

## ✨ Features

- **WebSocket Communication Framework**: Complete WebSocket server and client implementation, providing control channel
  capabilities
- **Bidirectional Messaging**: WebSocket-based bidirectional communication ensuring real-time configuration delivery
- **Multi-Driver System**: Complete implementation of three configuration communication drivers: Webhook, Kafka, and
  Long Polling
- **Connection Management**: Advanced connection lifecycle management with automatic reconnection and health checks
- **Trust & Security**: Client registration and trust establishment mechanisms ensuring communication security
- **Message Broadcasting**: Efficient message distribution to multiple trusted clients
- **Point-to-Point Notification**: Precise message push to specific clients
- **Network Discovery**: Intelligent network detection and IP management, adapting to complex network environments
- **Configuration Distribution**: Real-time configuration push and synchronization mechanisms

---

## 📁 Project Structure

```
switch-components/
├── pc/                         # Persistent Connection Management (WebSocket Framework)
├── drivers/                    # Communication Driver Layer
├── bc/                        # Business Context Management
├── config/                    # Configuration Management
├── http/                      # HTTP Middleware
├── grpc/                      # gRPC Communication Support
├── logging/                   # Logging System
│   └── request/               # Request Log Tracing
├── recovery/                  # Exception Recovery
├── snowflake/                 # Distributed ID Generation
└── system/                    # System Utilities
```

### Key Components:

- **`pc/`**: Core WebSocket persistent connection framework providing real-time communication infrastructure
- **`drivers/`**: Multiple communication driver implementations supporting Kafka, Webhook, Long Polling, and other modes
- **`bc/`**: Business context management providing context information within request lifecycle
- **`config/`**: Unified configuration management supporting YAML and environment variables
- **`http/`**: Middleware and utility support for HTTP services
- **`grpc/`**: High-performance RPC communication capabilities
- **`logging/`**: Structured logging and request tracing
- **`recovery/`**: System stability assurance mechanisms
- **`snowflake/`**: Distributed unique ID generation service
- **`system/`**: Low-level system operation support

---

## 🏛️ WebSocket Communication Framework

### Core Architecture and Communication Flow

The Persistent Connection (PC) framework provides a complete WebSocket communication solution, serving as a control
channel to deliver driver configurations and management information:

![communication-flow.svg](communication-flow.svg)

---

## 🚀 Configuration Communication Driver System

After Switch delivers driver configurations through the WebSocket control channel, clients start the corresponding
configuration communication drivers. The system supports three driver modes to meet different network environments and
business scenario requirements:

### Supported Drivers

#### 1. Webhook Driver (Push Mode)

- HTTP-based push notification mechanism
- Client starts Webhook server to receive configuration pushes
- Intelligent network discovery with automatic NAT environment adaptation
- Automatic retry on push failure to ensure configuration delivery

#### 2. Long Polling Driver (Pull Mode)

- HTTP long polling to actively pull configuration updates
- Suitable for restrictive network environments (firewalls, proxies)
- Exponential backoff automatic retry mechanism
- Low resource consumption with high compatibility

#### 3. Kafka Driver (Message Queue Mode)

- Kafka-based message queue distribution
- Topic-based configuration channel isolation
- High throughput and high reliability guarantees
- Suitable for large-scale distributed deployment scenarios

### Driver Configuration Examples

> Driver configurations are delivered through the WebSocket control channel, and clients start the corresponding
> configuration communication drivers after receiving them.

#### Kafka Driver Configuration

**Producer Configuration:**

```json
{
  // Broker list (required)
  "brokers": ["localhost:9092", "localhost:9093"],
  // Kafka topic (required) - Must be created in advance, auto-creation not supported. Driver validates topic validity on startup
  "topic": "example-topic",

  // Acknowledgment mechanism (optional) - all: all replicas acknowledge, one: leader acknowledges, none: no acknowledgment required
  "requiredAcks": "all",
  // Maximum timeout for waiting broker acknowledgment (optional)
  "timeout": "30s",
  // When message sending fails (network issues, broker unavailable, etc.), client will automatically retry 'retries' times (optional)
  "retries": 3,
  // Minimum retry interval after write failure (optional)
  "retryBackoffMin": "100ms",
  // Maximum retry interval after write failure (optional)
  "retryBackoffMax": "1s",
  // Message compression algorithm (optional) - gzip, snappy, lz4, zstd
  "compression": "snappy",

  // Connection timeout for broker connection dialer (optional)
  "connectTimeout": "10s",
  // Timeout for testing broker connectivity (optional)
  "validateTimeout": "10s",

  // Trigger write when time exceeds batchTimeout (optional)
  "batchTimeout": "1s",
  // Trigger write when message size meets batchBytes (optional)
  "batchBytes": 1048576,
  // Trigger write when message count meets batchSize (optional)
  "batchSize": 50,

  // Security configuration (optional)
  "security": {
    // SASL authentication (optional)
    "sasl": {
      // Enable or not
      "enabled": false,
      // Authentication mechanism - PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, etc.
      "mechanism": "PLAIN",
      // Username
      "username": "your-username",
      // Password
      "password": "your-password"
    },
    // TLS authentication (optional)
    "tls": {
      // Enable or not
      "enabled": false,
      // CA certificate path for verifying Kafka server validity (system path)
      "caFile": "/path/to/ca.pem",
      // Client certificate path (system path)
      "certFile": "/path/to/cert.pem",
      // Client key path (system path)
      "keyFile": "/path/to/key.pem",
      // Skip certificate verification or not
      "insecureSkipVerify": false
    }
  }
}
```

**Consumer Configuration:**

```json
{
  // Broker list (required)
  "brokers": [
    "localhost:9092",
    "localhost:9093"
  ],
  // Kafka topic (required) - Must be created in advance, auto-creation not supported. Driver validates topic validity on startup
  "topic": "example-topic",

  // Consumer group ID (optional) - Leave empty if no special requirements, let framework generate consumer group ID
  "groupId": "",
  // Offset reset strategy (optional) - earliest: consume from earliest message, latest: consume from latest message
  "autoOffsetReset": "latest",
  // Auto commit (optional)
  "enableAutoCommit": true,
  // Manual commit interval (optional) - Can be set larger, duplicate consumption of switch messages has no impact
  "autoCommitInterval": "10s",

  // Connection timeout configuration (optional)
  "connectTimeout": "10s",
  // Timeout for validating connection connectivity (optional)
  "validateTimeout": "10s",

  // Read message timeout configuration (optional)
  "readTimeout": "10s",
  // Commit offset timeout configuration (optional)
  "commitTimeout": "10s",

  // Security configuration (optional)
  "security": {
    // SASL authentication (optional)
    "sasl": {
      // Enable or not
      "enabled": false,
      // Authentication mechanism - PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, etc.
      "mechanism": "PLAIN",
      // Username
      "username": "your-username",
      // Password
      "password": "your-password"
    },
    // TLS authentication (optional)
    "tls": {
      // Enable or not
      "enabled": false,
      // CA certificate path for verifying Kafka server validity (system path)
      "caFile": "/path/to/ca.pem",
      // Client certificate path (system path)
      "certFile": "/path/to/cert.pem",
      // Client key path (system path)
      "keyFile": "/path/to/key.pem",
      // Skip certificate verification or not
      "insecureSkipVerify": false
    }
  },
  // Retry configuration (optional)
  "retry": {
    // Restart when failures exceed count, reset on success
    "count": 1,
    // Restart interval time
    "backoff": "3s"
  }
}
```

---

## 🔧 Advanced Features

### Network Discovery

The system includes intelligent network discovery capabilities:

- **Local IP Detection**: Automatically detect local network interfaces
- **Public IP Discovery**: Use external services to determine public IP
- **NAT Detection**: Identify NAT environments and adjust communication strategies
- **Webhook Reachability Testing**: Test webhook endpoint accessibility

### Message Broadcasting

Efficient message distribution system:

- **Targeted Broadcasting**: Send messages to specific clients or groups
- **Topic-Based Routing**: Route messages based on topics or patterns
- **Delivery Confirmation**: Track message delivery and handle failures

### Connection Management

Advanced connection lifecycle management:

- **Connection Pool**: Efficient connection resource management
- **Health Monitoring**: Continuous connection health checks
- **Graceful Shutdown**: Clean connection termination
- **Resource Cleanup**: Automatic cleanup of disconnected clients

---

## 📦 Integration and Usage

### Installation

```bash
go get gitee.com/fatzeng/switch-components
```

### Quick Start

As a utility component library, `switch-components` provides rich reusable components. Here are some common integration
scenarios:

#### Using WebSocket Communication Framework

```go
// Create WebSocket server
wsConfig := config.GlobalConfig.Pc
if wsConfig == nil {
    // Default configuration
    wsConfig = pc.DefaultServerConfig()
}
// Connection callback
wsConfig.OnConnect = onClientConnect
// Disconnection callback
wsConfig.OnDisconnect = onClientDisconnect
// Message handler callback
wsConfig.MessageHandler = onMessage
// Trust callback
wsConfig.OnClientTrusted = onClientTrusted

wsServer = pc.NewServer(wsConfig)

// Register configuration change business endpoint
wsServer.RegisterHandler(pc.WsEndpointChangeConfig, nil)
// Register full configuration sync business endpoint
wsServer.RegisterHandler(pc.WsEndpointFullSyncConfig, nil)
// Register full switch sync business endpoint
wsServer.RegisterHandler(pc.WsEndpointFullSync, nil)

wsServer.Start(ctx)
```

#### Using Logging System

```go
// Structured logging
log, err := logging.New(&logger.LoggerConfig{
    Level:            "info",                     // Log level
    OutputDir:        "./logs",                   // Log output directory
    FileNameFormat:   "switch-demo_%Y-%m-%d.log", // Log file name format
    MaxSize:          50,                         // Maximum size of single log file (MB)
    MaxBackups:       3,                          // Number of old log files to retain
    MaxAge:           7,                          // Log file retention days
    Compress:         false,                      // Whether to compress old log files
    ShowCaller:       true,                       // Whether to show caller information
    EnableConsole:    true,                       // Whether to enable console output
    EnableJSON:       false,                      // Whether to use JSON format (false for demo readability)
    EnableStackTrace: true,                       // Whether to enable stack trace
    StackTraceLevel:  "error",                    // Stack trace level
    TimeFormat:       "2006-01-02 15:04:05",      // Time format
})
```

For more detailed usage examples and API documentation, please refer to the documentation in each component directory.

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
