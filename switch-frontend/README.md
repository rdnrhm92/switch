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

**Switch is a distributed real-time feature management platform** that provides secure and efficient dynamic configuration capabilities for enterprise-level applications. Through advanced WebSocket persistent connection architecture and multi-driver communication mechanisms, it achieves millisecond-level configuration delivery, enabling development teams to precisely control feature releases, user experiences, and system behaviors without application restarts.

### 🏗️ Core Architecture Advantages

**1. Enterprise-Grade Communication Architecture**

- **WebSocket Persistent Connection Framework** - Persistent connection management based on `switch-components/pc`
- **Multi-Driver Support** - Three communication modes: Webhook, Kafka, and Long Polling
- **Intelligent Network Discovery** - Automatic adaptation to NAT environments, solving complex network scenarios
- **Layered Acknowledgment Mechanism** - Ensuring reliability and consistency of configuration delivery

**2. Flexible Factor System**

```go
// Not just simple true/false, but intelligent decision-making based on complex rules
if _switch.IsOpen(ctx, "feature_enabled") {
    // The system calculates in real-time whether to enable the feature
    // based on configured multi-dimensional factors (e.g., user attributes, geographic location, time windows, etc.)
}
```

**3. Multi-Tenant Management System**

- **Tenant Isolation** - Complete data and permission isolation
- **Environment Management** - Independent configurations for development, testing, and production environments with strict configuration promotion policies
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

# Switch-Frontend: Enterprise-Grade Visual Management Platform

**Switch-Frontend** is an enterprise-grade visual feature management platform built on modern frontend technology stack, providing comprehensive management interface for the Switch distributed real-time configuration system. The platform adopts a deeply integrated architecture design of React 19+ and TypeScript, combined with Ant Design Pro enterprise-level UI component library, delivering professional, intuitive, and efficient user experience for core business scenarios including feature toggle management, multi-environment governance, driver configuration, and approval workflows.

As the unified management center of the Switch ecosystem, Switch-Frontend is not only a visual interface for configuration management, but also a critical hub connecting business decisions with technical implementation. Through carefully designed interactive experiences and powerful functional modules, it empowers enterprises to achieve agile feature releases, refined configuration governance, and full-chain observability.

---

## ✨ Core Capability Matrix

### 🏢 Multi-Tenant Governance System
- **Tenant Isolation Architecture**: Complete data isolation mechanism based on namespaces, ensuring absolute security boundaries for configurations and data between tenants
- **Unified Identity Management**: Integrated user registration and authentication management system, supporting enterprise-level SSO integration
- **Fine-Grained Permission Control**: Multi-dimensional permission management based on RBAC model, achieving precise authorization for roles, resources, and operations
- **Tenant Context Switching**: Seamless tenant workspace switching capability, supporting cross-tenant management and collaboration scenarios

### 🎛️ Intelligent Switch Management
- **Visual Configuration Center**: Provides intuitive interface for switch creation, editing, and version management, supporting batch operations and configuration templates
- **Rule Orchestration Engine**: Visual rule builder supporting complex condition combinations, logical operations, and dynamic expressions
- **Multi-Dimensional Factor System**: Flexible configuration of evaluation factors (user profiles, geographic location, device types, time windows, etc.) to achieve precise grayscale strategies
- **Real-Time Evaluation Sandbox**: Built-in switch evaluation testing environment supporting multi-scenario simulation and result preview, reducing configuration risks

### 🌍 Full Lifecycle Environment Management
- **Environment Topology Orchestration**: Supports custom environment hierarchies (development, testing, staging, production, etc.), building complete configuration flow pipelines
- **Environment Difference Comparison**: Visual comparison analysis of cross-environment configurations, quickly identifying configuration differences and potential risks
- **Configuration Promotion Mechanism**: One-click configuration promotion capability, supporting configuration synchronization and batch deployment between environments
- **Version Rollback Protection**: Complete configuration version history tracking, supporting quick rollback and recovery to any version

### 🔌 Multi-Driver Communication Architecture
- **Heterogeneous Driver Support**: Unified management of three communication drivers (Webhook, Kafka, Long Polling), adapting to different network topologies and business scenarios
- **Driver Health Detection**: Real-time monitoring of driver connection status, throughput performance, and error rates, providing intelligent diagnostic recommendations
- **Performance Monitoring Dashboard**: Visual display of driver-level message latency, success rate, and resource consumption metrics
- **Fault Tolerance and Degradation Strategy**: Flexible configuration of driver fallback and degradation rules, ensuring high availability of configuration delivery

### ✅ Enterprise-Grade Approval Workflow
- **Multi-Level Approval Process**: Built-in configurable approval chains, supporting serial, parallel, and conditional approval modes
- **Intelligent Review System**: Provides configuration change difference comparison, impact scope analysis, and risk assessment capabilities
- **Full-Chain Audit Tracking**: Records all configuration changes, approval decisions, and operation logs, meeting compliance audit requirements
- **Real-Time Message Notification**: Integrated in-app notification channels, ensuring timely response to approval processes

---

## 🛠️ Technology Stack

- **Frontend Framework**: React 19+ with TypeScript - Adopting latest React features and strict type system, ensuring code quality and maintainability
- **UI Component Library**: Ant Design Pro v6 - Enterprise-level design language and React component library for middle and back-end platforms, providing out-of-the-box high-quality components
- **State Management**: React Context API + Hooks - Lightweight state management solution, avoiding over-engineering
- **Build Toolchain**: UmiJS v4 - Enterprise-level frontend application framework, integrating complete toolchain for routing, building, and deployment
- **Code Quality**: BiomeJS - High-performance code linting and formatting tool, unifying code style
- **Testing Framework**: Jest + React Testing Library - Complete unit testing and integration testing solution
- **Code Editors**: Monaco Editor + CodeMirror - Professional code editing experience supporting JSON/YAML configuration
- **Development Experience**: HMR Hot Module Replacement + TypeScript Strict Mode - Ultimate development efficiency and type safety

---

## 📁 Project Structure

```
switch-frontend/
├── src/                       # Source code directory
│   ├── components/            # Reusable UI component library
│   ├── pages/                 # Page components (switches, environments, drivers, approvals, users, tenants, etc.)
│   ├── services/              # API service layer, encapsulating backend interface calls
│   ├── models/                # TypeScript type definitions and data models
│   ├── utils/                 # Common utility functions (requests, authentication, validation, etc.)
│   ├── hooks/                 # Custom React Hooks
│   ├── layouts/               # Application layout components
│   └── locales/               # Internationalization language packages
├── config/                    # Build and development configuration (UmiJS, proxy, routing, etc.)
├── public/                    # Static resource files
├── tests/                     # Unit tests and integration tests
└── package.json               # Project dependencies and script configuration
```

---

## 🚀 Quick Start

### Prerequisites

- Node.js 20.0+
- npm or yarn package manager

### Installation

```bash
# Clone the repository
git clone https://gitee.com/fatzeng/switch-frontend.git
cd switch-frontend

# Install dependencies
npm install
# or
yarn install
```

### Running

```bash
# Start development server
npm run dev
# or
yarn dev

# Application will be available at http://localhost:8000
```
---

## 🤝 Contributing Guide

We welcome and appreciate all forms of contributions! Whether it's reporting issues, suggesting features, improving documentation, or submitting code, your participation will help make Switch better.

### How to Contribute

1. **Fork this repository** and create your feature branch
2. **Write code** and ensure compliance with project coding standards
3. **Add tests** to cover your changes
4. **Submit a Pull Request** with detailed description of your changes and motivation

### Contribution Types

- 🐛 **Bug Fixes**: Discover and fix issues
- ✨ **New Features**: Propose and implement new capabilities
- 📝 **Documentation Improvements**: Enhance documentation and examples
- 🎨 **Code Optimization**: Improve code quality and performance
- 🧪 **Test Enhancement**: Increase test coverage

For more detailed information, please refer to our contribution guide documentation.

## 📄 License

This project is licensed under the [MIT License](LICENSE).
