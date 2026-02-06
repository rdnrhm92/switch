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

# Switch-Admin: Backend Management Service

`switch-admin` is the central backend service for the Switch project. It provides a powerful control plane for managing
feature flags, complex release workflows, and ensuring system stability and traceability.

---

## ✨ Features

- **Multi-Tenant Architecture**: Support for multiple tenants with complete permission management
- **Real-Time Communication**: WebSocket-based communication with all connected clients
- **Multiple Communication Modes**:
    - Pull Mode: Clients actively fetch configurations
    - Push Mode: Server pushes configurations to clients
    - Kafka Mode: Message queue-based distribution
- **Approval Workflows**: Built-in approval system for sensitive environment changes
- **Environment Management**: Manage switches across different environments (development, testing, production), supports
  custom environments
- **Driver Configuration**: Flexible driver system supporting different communication modes
- **Audit & Traceability**: Complete audit trail for all changes and operations
- **Client Registration & Trust**: Secure client registration and trust management system

---

## 🏛️ Communication Flow

![Switch Architecture](switch_architecture.svg)

### Communication Process Details:

Switch adopts a WebSocket-based persistent connection architecture to achieve millisecond-level configuration delivery
and real-time communication. The entire communication process includes the following key steps:

#### 1. Connection Establishment Phase

```
Client                    Server
   |                        |
   |--- TCP Connection ---->|
   |                        |
   |<--- WebSocket Upgrade -|
   |                        |
```

- **Client**: Initiates TCP connection and performs WebSocket handshake upgrade via HTTP protocol
- **Server**: Handles WebSocket connection upgrade and establishes persistent connection channel

#### 2. Connection Initialization Phase

```
Client                    Server
   |                        |
   |--- Start Read/Write -->|
   |<--- Start Read/Write --|
   |                        |
   |<--- Say Hello ---------|
   |                        |
```

- **Client**: Starts asynchronous read/write loops, waiting for server greeting message
- **Server**: Starts asynchronous read/write loops, proactively sends "Say Hello" to establish connection confirmation

#### 3. Authentication Phase

```
Client                    Server
   |                        |
   |--- Process Say Hello ->|
   |                        |
   |--- Send Registration ->|
   |                        |
   |<--- Process Reg(Trust)-|
   |<--- Registration OK ---|
   |                        |
```

- **Client**: Parses server greeting message, sends registration request containing service identifier, version,
  environment, and other information
- **Server**: Verifies client identity, establishes trust relationship, returns registration success response

#### 4. Configuration Synchronization Phase

```
Client                    Server
   |                        |
   |--- Process Reg Resp -->|
   |                        |
   |--- Fetch Driver Config>|
   |--- Fetch Switch Config>|
   |--- Fetch Incremental ->|
   |                        |
   |<--- Driver Config -----|
   |<--- Switch Config -----|
   |<--- Incremental Config-|
   |                        |
   |<--- Server Waits ------|
   |                        |
```

- **Client**: Processes registration response, fetches driver configuration (Webhook/Kafka/Long Polling), starts
  incremental configuration monitoring, fetches complete switch configuration
- **Server**: Completes client registration, waits for subsequent business requests

#### 5. Business Communication Phase

```
Client                    Server
   |                        |
   |                        |
   |<--- Push Incremental --|
   |----- Return Response ->|
   |                        |
```

- **Client**: Waits for incremental configuration sent by server and updates its own driver logic
- **Server**: Obtains changed driver configuration and pushes to connected clients

### Communication Features:

- **Bidirectional Communication**: Supports both client-initiated requests and server-initiated pushes
- **Multi-Connection Architecture**: Three independent WebSocket connections that don't interfere with each other:
    - **Full Switch Connection**: Carries complete switch configuration data synchronization
    - **Full Configuration Connection**: Carries complete configuration item data synchronization
    - **Incremental Switch Connection**: Carries real-time switch change pushes
- **Intelligent Reconnection**: Automatically handles network exceptions and connection drops
- **Message Acknowledgment**: Critical operations support ACK confirmation mechanism
- **Load Balancing**: Supports multi-instance deployment and client load distribution
- **Secure Transmission**: Supports TLS encryption and authentication

---

## 🚀 Getting Started

### Prerequisites

- **Go 1.18+** - Modern Go language support
- **Database**: MySQL 8.0+ - Data persistence
- **Kafka** (Optional) - Message queue support

### Installation and Running

#### 1. Clone the Repository

```bash
git clone https://gitee.com/fatzeng/switch-admin.git ./switch-admin
cd switch-admin
```

#### 2. Configure the Service

Edit `configs/config.yaml` to set up database connection, notification drivers (Kafka/Webhook), and other service
settings.

For detailed configuration instructions, please refer to: `configs/switch-config.yaml`

```bash
vim configs/config.yaml
```

#### 3. Install Dependencies

```bash
go mod tidy
```

#### 4. Run the Application

```bash
go run cmd/server/main.go
```

The service will start, automatically run database migrations, and initialize data (including default `admin`/`admin`
user, permissions, default factors, configurations, etc.).

---

## 🔧 Core Concepts: Release Workflow

To ensure stability and traceability, `switch-admin` adopts a structured release workflow rather than simple CRUD
operations. This workflow is based on four key models:

1. **`PublishRequest`**: Represents the *intent* to change a switch. It captures the complete desired configuration as a
   JSON object.
2. **`SwitchSnapshot`**: An immutable record of a switch's configuration at a specific point in time. It serves as the
   ground truth for the live state.
3. **`ApprovalForm`**: A request generated when a change targets a protected environment. The change can only proceed
   once it is approved.
4. **`Progressive Release Process`**: Switch promotion must follow the configured environment release order for
   progressive promotion, ensuring switch consistency across environments.

This workflow guarantees that every change is intentional, auditable, and safe.

---

## 📁 Project Structure

```
switch-admin/
├── cmd/
│   └── server/                     # Application entry point
├── configs/                        # Service configuration files
├── info/                           # User information related
├── internal/
│   ├── admin_driver/               # Database and external service drivers
│   ├── admin_model/                # Database entity models (GORM)
│   ├── api/                        # HTTP handlers, routing, and middleware (Gin)
│   │   ├── controller/             # Controller layer
│   │   └── middleware/             # Middleware
│   ├── config/                     # Configuration management
│   ├── dto/                        # Data Transfer Objects
│   ├── notifier/                   # Client update notification system
│   ├── repository/                 # Data Access Object (DAO) layer
│   ├── service/                    # Business logic layer
│   ├── types/                      # Type definitions and error handling
│   ├── utils/                      # Utility functions
│   └── ws/                         # WebSocket communication handlers
├── LICENSE                         # License file
├── README.md                       # English documentation
├── README_zh.md                    # Chinese documentation
├── go.mod                          # Go module definition
├── go.sum                          # Go module dependency verification
└── switch.svg                      # Project logo
```

### Key Components:

- **`api/`**: RESTful API endpoints for frontend interaction
- **`service/`**: Core business logic for switch management and approval workflows
- **`admin_model/`**: Database models for tenants, switches, environments, etc.
- **`repository/`**: Data access layer with GORM integration
- **`ws/`**: WebSocket server for real-time client communication
- **`notifier/`**: Push notification system for configuration updates
- **`admin_driver/`**: Database drivers and external service integrations

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
