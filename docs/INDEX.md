# GoDriftDetector - Complete Documentation Index

Welcome! This guide will help you navigate all GoDriftDetector documentation.

## 🚀 Getting Started (5-10 minutes)

Start here if you're new to GoDriftDetector:

1. **[QUICK_START.md](./QUICK_START.md)** ⭐
   - Install and run in 5 minutes
   - Three installation options
   - First drift detection test
   - Common commands cheat sheet

2. **[README.md](./README.md)**
   - Project overview
   - Feature list and tech stack
   - Architecture diagram
   - Complete usage guide with examples

## 📦 Installation

Choose your preferred installation method:

- **[INSTALLATION.md](./INSTALLATION.md)** - Complete installation guide
  - Binary installation (recommended)
  - Source code setup
  - Docker container
  - Platform-specific (macOS, Linux, Windows)
  - Kubernetes deployment
  - Systemd service
  - Verification steps

## ⚙️ Configuration

Configure GoDriftDetector for your environment:

- **[CONFIGURATION.md](./CONFIGURATION.md)** - Configuration reference
  - All environment variables
  - Command-line flags
  - Config file formats (Docker Compose & Kubernetes)
  - Real-world examples
  - Performance tuning
  - Troubleshooting

## 🧪 Testing

Learn how to test and validate:

- **[TESTING.md](./TESTING.md)** - Testing guide
  - Unit tests
  - 8 integration test scenarios
  - Performance testing
  - CI/CD pipeline examples
  - Debugging techniques
  - Test checklist

## 📚 Documentation by Use Case

### I want to...

#### Monitor Docker Compose
1. [QUICK_START.md - Option 1](./QUICK_START.md#option-1-docker-fastest)
2. [CONFIGURATION.md - Docker Compose Config](./CONFIGURATION.md#docker-compose-configuration)
3. [TESTING.md - Test 1: Basic Docker Drift](./TESTING.md#test-1-basic-docker-drift-detection)

#### Monitor Kubernetes
1. [INSTALLATION.md - Kubernetes Section](./INSTALLATION.md#kubernetes)
2. [CONFIGURATION.md - Kubernetes Manifest](./CONFIGURATION.md#kubernetes-manifest-configuration)
3. [TESTING.md - Test 7: Kubernetes Provider](./TESTING.md#test-7-kubernetes-provider-requires-running-cluster)

#### Setup GitOps (Auto Git Sync)
1. [CONFIGURATION.md - Example 2: GitOps](./CONFIGURATION.md#example-2-gitops-with-github)
2. [INSTALLATION.md - Development Setup](./INSTALLATION.md#development-setup)

#### Send Slack Alerts
1. [QUICK_START.md - With Slack](./QUICK_START.md#with-slack-notifications)
2. [CONFIGURATION.md - Example 3: Slack](./CONFIGURATION.md#example-3-slack-notifications)

#### Monitor with Prometheus/Grafana
1. [QUICK_START.md - Monitoring](./QUICK_START.md#monitoring-prometheus)
2. [CONFIGURATION.md - Example 4: Prometheus](./CONFIGURATION.md#example-4-prometheus-monitoring)
3. [TESTING.md - Test 4: Prometheus Metrics](./TESTING.md#test-4-prometheus-metrics)

#### Integrate into CI/CD Pipeline
1. [QUICK_START.md - CI/CD Integration](./QUICK_START.md#cicd-integration)
2. [CONFIGURATION.md - Example 6: CI/CD](./CONFIGURATION.md#example-7-cicd-integration)
3. [TESTING.md - CI/CD Integration Tests](./TESTING.md#cicd-integration-tests)

#### Environment Variable Inspection
1. [README.md - Drift Types](./README.md#drift-types)
2. [TESTING.md - Test 2: Environment Variable Drift](./TESTING.md#test-2-environment-variable-drift)
3. [CONFIGURATION.md - Example 6: Environment Sensitivity](./CONFIGURATION.md#example-6-environment-variable-sensitivity)

## 🏗️ Architecture

Understand how GoDriftDetector works:

- **[README.md - Architecture](./README.md#architecture)**
  - System design overview
  - Package structure
  - Design patterns used
  - SVG diagram

## 📊 All 8 Implementation Phases

Complete implementation of all 8 phases:

| Phase | Feature | Location |
|-------|---------|----------|
| **1** | Docker SDK State Detection | [README.md#quick-start](./README.md#quick-start) |
| **2** | YAML Config Parser | [CONFIGURATION.md#docker-compose-configuration](./CONFIGURATION.md#docker-compose-configuration) |
| **3** | Diff Engine (6 Drift Types) | [README.md#drift-types](./README.md#drift-types) |
| **4** | GitOps Agent with Polling | [CONFIGURATION.md#example-2-gitops-with-github](./CONFIGURATION.md#example-2-gitops-with-github) |
| **5** | Webhooks & JSON Reports | [QUICK_START.md#with-slack-notifications](./QUICK_START.md#with-slack-notifications) |
| **6** | Env Var Deep Inspection | [TESTING.md#test-2-environment-variable-drift](./TESTING.md#test-2-environment-variable-drift) |
| **7** | Multi-Provider (K8s) | [INSTALLATION.md#kubernetes](./INSTALLATION.md#kubernetes) |
| **8** | Prometheus Metrics | [QUICK_START.md#monitoring-prometheus](./QUICK_START.md#monitoring-prometheus) |

## 🔍 Feature Guide

### Drift Types

GoDriftDetector detects 6 types of infrastructure drift:

| Type | Description | Detection |
|------|-------------|-----------|
| **MISSING** | Service declared but not running | Phase 3 |
| **SHADOW_IT** | Container running but not declared | Phase 3 |
| **PORT_MISMATCH** | Port mapping differs | Phase 3 |
| **IMAGE_MISMATCH** | Container image version differs | Phase 3 |
| **ENV_MISMATCH** | Environment variable value differs | Phase 6 |
| **ENV_INJECTED** | Undeclared environment variable | Phase 6 |

See [README.md - Drift Types](./README.md#drift-types) for details.

### Environment Variable Masking

Sensitive values are automatically masked in logs:

- **Detected patterns**: password, token, secret, key, auth
- **Masking**: First 3 characters + `***`
- **Example**: `API_TOKEN=mytoken_xyz` → `API_TOKEN=myt***`

See [TESTING.md - Test 2](./TESTING.md#test-2-environment-variable-drift) for examples.

### Prometheus Metrics

Real-time monitoring with Prometheus:

- `drift_detected_total` - Total drifts found
- `drift_by_service` - Drifts per service with labels
- `last_scan_timestamp` - Unix timestamp of last scan

See [CONFIGURATION.md - Example 4](./CONFIGURATION.md#example-4-prometheus-monitoring).

## 🛠️ Troubleshooting

Common issues and solutions:

### General Troubleshooting
- [INSTALLATION.md - Troubleshooting](./INSTALLATION.md#troubleshooting)
- [CONFIGURATION.md - Troubleshooting](./CONFIGURATION.md#troubleshooting-configuration)
- [TESTING.md - Common Issues](./TESTING.md#common-issues)

### Specific Problems

**Docker socket permission denied**
```bash
sudo usermod -aG docker $USER
newgrp docker
```

**Kubernetes connection error**
```bash
export KUBECONFIG=$HOME/.kube/config
./godriftdetector --provider k8s
```

**Port 9090 already in use**
```bash
lsof -i :9090
kill -9 <PID>
```

See individual docs for more troubleshooting.

## 🤝 Contributing

Want to contribute? See:

- **[CONTRIBUTING.md](./CONTRIBUTING.md)** - Contribution guidelines
- **[CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md)** - Community guidelines
- **[SECURITY.md](./SECURITY.md)** - Security reporting

## 📋 Command Reference

### Installation

```bash
# Binary (recommended)
go install github.com/esousa97/godriftdetector/cmd/godriftdetector@latest

# From source
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector
go build -o godriftdetector ./cmd/godriftdetector

# Docker
docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/esousa97/godriftdetector:latest --json
```

### Basic Commands

```bash
# One-shot audit
./godriftdetector --json

# Daemon mode (every 5 minutes)
./godriftdetector

# Kubernetes audit
./godriftdetector --provider k8s --namespace default --json

# With Slack notifications
export WEBHOOK_URL="https://hooks.slack.com/..."
./godriftdetector

# With Git auto-sync
export GIT_REPO_URL="https://github.com/org/config.git"
export LOCAL_CONFIG_DIR="/var/lib/config"
export SYNC_INTERVAL="5m"
./godriftdetector
```

## 📱 Quick Links

| Topic | Link |
|-------|------|
| **Quick Start** | [QUICK_START.md](./QUICK_START.md) |
| **Full README** | [README.md](./README.md) |
| **Installation** | [INSTALLATION.md](./INSTALLATION.md) |
| **Configuration** | [CONFIGURATION.md](./CONFIGURATION.md) |
| **Testing** | [TESTING.md](./TESTING.md) |
| **Contributing** | [CONTRIBUTING.md](./CONTRIBUTING.md) |
| **GitHub** | https://github.com/esousa97/godriftdetector |
| **Issues** | https://github.com/esousa97/godriftdetector/issues |

## 📖 Reading Paths

### For DevOps/SRE
1. [QUICK_START.md](./QUICK_START.md) - 5 min
2. [INSTALLATION.md](./INSTALLATION.md) - 10 min
3. [CONFIGURATION.md](./CONFIGURATION.md) - 15 min
4. Deploy and monitor

### For Developers
1. [README.md - Architecture](./README.md#architecture)
2. [INSTALLATION.md - Development Setup](./INSTALLATION.md#development-setup)
3. [TESTING.md - Unit Tests](./TESTING.md#unit-tests)
4. Start contributing!

### For Security Auditors
1. [README.md - Drift Types](./README.md#drift-types)
2. [CONFIGURATION.md - Environment Variables](./CONFIGURATION.md#environment-variables)
3. [TESTING.md - Integration Tests](./TESTING.md#integration-tests)
4. [SECURITY.md](./SECURITY.md)

## 🌐 Translation Status

- 🟢 **English** - Complete and current
- 🟡 **Portuguese** (Original) - See git history for pt-BR versions

All documentation is available in **English** for international audience.

---

**Last Updated**: April 2026
**Version**: 1.0.0
**Status**: Production Ready

For the latest information, visit the [GitHub repository](https://github.com/esousa97/godriftdetector).
