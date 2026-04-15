<div align="center">
  <h1>GoDriftDetector</h1>
  <p>Lightweight infrastructure drift detection agent for Docker and Kubernetes.</p>

  <img src="assets/github-go.png" alt="GoDriftDetector Banner" width="600px">

  <br>

![CI](https://github.com/esousa97/godriftdetector/actions/workflows/ci.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/esousa97/godriftdetector?style=flat)](https://goreportcard.com/report/github.com/esousa97/godriftdetector)
[![CodeFactor](https://www.codefactor.io/repository/github/esousa97/godriftdetector/badge)](https://www.codefactor.io/repository/github/esousa97/godriftdetector)
[![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat&logo=go&logoColor=white)](https://pkg.go.dev/github.com/esousa97/godriftdetector)
![License](https://img.shields.io/github/license/esousa97/godriftdetector?style=flat&color=blue)
![Go Version](https://img.shields.io/github/go-mod/go-version/esousa97/godriftdetector?style=flat&logo=go&logoColor=white)
![Last Commit](https://img.shields.io/github/last-commit/esousa97/godriftdetector?style=flat)

</div>

---

**GoDriftDetector** is a lightweight daemon that continuously compares the desired state (Desired State) of your infrastructure declared in a `docker-compose.yaml` or Kubernetes manifest with the actual state (Actual State) of running containers. Designed to detect "Shadow IT", service downtime, version/port divergences, and environment variable mismatches, emitting structured alerts for rapid mitigation.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Usage Guide](#usage-guide)
- [Testing](#testing)
- [Architecture](#architecture)
- [Drift Types](#drift-types)
- [Providers](#providers)
- [Observability](#observability)
- [Examples](#examples)
- [Roadmap](#roadmap)
- [Contributing](#contributing)

## Overview

When a drift is detected in your infrastructure, the agent reports visually in the terminal and can notify via Webhook:

```text
--- Verification Cycle: 2026-04-14T10:00:00Z ---
Reading configuration: ./docker-compose.yaml
DRIFT DETECTED!
[ SHADOW_IT ] Undeclared container running: 1a2b3c4d5e6f (Image: redis:alpine)
[ MISSING ] Service 'db' (image postgres:15) is not running
[ PORT_MISMATCH ] Desired port '80:80' not found in container
[ ENV_MISMATCH ] Environment variable 'DB_PASSWORD' differs. Expected: 'exp***', Actual: 'act***'
Alert sent successfully to webhook
```

## Features

| Feature | Description |
|---------|-------------|
| 🔍 **Multi-Drift Detection** | Detects MISSING, SHADOW_IT, PORT_MISMATCH, IMAGE_MISMATCH, ENV_MISMATCH, ENV_INJECTED |
| 🐳 **Multi-Provider** | Supports Docker and Kubernetes with pluggable architecture |
| 🔐 **Sensitive Data Masking** | Automatically masks passwords, tokens, and secrets in logs |
| 📊 **Prometheus Metrics** | Real-time metrics for Grafana integration |
| 🚨 **Webhooks** | Send alerts to Slack, Discord, or custom endpoints |
| 🔄 **GitOps Ready** | Automatic Git repository sync for configuration management |
| 📝 **JSON Reports** | Export audit reports in JSON format for CI/CD integration |
| 💻 **Daemon Mode** | Background polling with configurable intervals |
| 🎨 **Colored Output** | Terminal-friendly with lipgloss styling |

## Tech Stack

| Technology | Purpose |
|---|---|
| **Go** | High-performance compiled language with static binaries |
| **Docker SDK** | Runtime state extraction from containers |
| **go-git** | Remote Git repository synchronization |
| **yaml.v3** | Robust docker-compose.yaml parsing |
| **lipgloss** | Terminal output styling (colors, bold) |
| **Prometheus SDK** | Metrics exposition for observability |
| **client-go** | Kubernetes API interaction |

## Prerequisites

- **Go** >= 1.25.0
- **Docker** daemon running OR **Kubernetes** cluster access (kubectl configured)
- *(Optional)* Slack/Discord webhook URLs for notifications

## Installation

### From Binary (Recommended)

```bash
go install github.com/esousa97/godriftdetector/cmd/godriftdetector@latest
godriftdetector --help
```

### From Source

```bash
# Clone the repository
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector

# Build the binary
go build -o godriftdetector ./cmd/godriftdetector

# Run
./godriftdetector --help
```

### Docker

```bash
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/docker-compose.yaml:/docker-compose.yaml \
  -e LOCAL_CONFIG_DIR=/ \
  ghcr.io/esousa97/godriftdetector:latest
```

## Quick Start

### 1. Docker Compose Setup

```bash
# Create a docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: nginx:alpine
    ports:
      - "80:80"
    environment:
      APP_ENV: production
      LOG_LEVEL: info
EOF

# Start containers
docker-compose up -d

# Run drift detection
LOCAL_CONFIG_DIR=. ./godriftdetector --json
```

### 2. Kubernetes Setup

```bash
# Create a k8s-manifest.yaml
cat > k8s-manifest.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  template:
    spec:
      containers:
      - name: app
        image: nginx:alpine
        ports:
        - containerPort: 80
        env:
        - name: APP_ENV
          value: production
EOF

# Run drift detection
./godriftdetector --provider k8s --namespace default --json
```

### 3. Daemon Mode with GitOps

```bash
export GIT_REPO_URL="https://github.com/your-org/config-repo.git"
export GIT_USERNAME="your-token"
export LOCAL_CONFIG_DIR="/tmp/config"
export SYNC_INTERVAL="5m"
export WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK"

./godriftdetector
```

## Configuration

### Environment Variables

| Variable | Type | Default | Description |
|---|---|---|---|
| `GIT_REPO_URL` | String | `""` | HTTPS/SSH URL of Git repository containing docker-compose.yaml or k8s-manifest.yaml |
| `GIT_USERNAME` | String | `""` | Username/Token for HTTPS Git access |
| `GIT_PASSWORD` | String | `""` | Password/Token for HTTPS Git access |
| `LOCAL_CONFIG_DIR` | String | `"./config-repo"` | Local directory for Git clone/cache |
| `SYNC_INTERVAL` | Duration | `"5m"` | Polling frequency (e.g., `10m`, `30s`, `5m`) |
| `WEBHOOK_URL` | String | `""` | Webhook URL for Slack/Discord notifications |

### Command-Line Flags

```bash
godriftdetector [flags]

Flags:
  -json                          Generate drift report in JSON format and exit
  -provider string               Infrastructure provider: 'docker' or 'k8s' (default "docker")
  -namespace string              Kubernetes namespace (only with --provider=k8s) (default "default")
```

## Usage Guide

### 1. One-Shot Audit Report

Generate a single JSON report without starting the daemon:

```bash
LOCAL_CONFIG_DIR=. ./godriftdetector --json > drift-report.json
```

**Output:**
```json
{
  "Drifts": [
    {
      "ServiceName": "app",
      "Type": "ENV_MISMATCH",
      "Message": "Environment variable 'LOG_LEVEL' outdated. Expected: 'inf***', Actual: 'deb***'.",
      "Desired": "inf***",
      "Actual": "deb***"
    },
    {
      "ServiceName": "app",
      "Type": "ENV_INJECTED",
      "Message": "Undeclared environment variable found running: 'PATH=/usr/bin'.",
      "Actual": "/usr/bin"
    }
  ]
}
```

### 2. Daemon Mode with Polling

Run continuous monitoring with automatic Git sync:

```bash
export LOCAL_CONFIG_DIR="./config"
export SYNC_INTERVAL="5m"
export WEBHOOK_URL="https://hooks.slack.com/services/..."

./godriftdetector
```

**Output:**
```
Initiating GoDriftDetector Agent (Interval: 5m0s, Provider: docker)
Exposing metrics at http://localhost:9090/metrics

--- Verification Cycle: 2026-04-14T22:30:29-03:00 ---
DRIFT DETECTED!
[ ENV_MISMATCH ] Environment variable 'ENVIRONMENT' outdated. Expected: 'production', Actual: 'staging'
[ ENV_INJECTED ] Undeclared environment variable found: 'EXTRA_VAR=injected_value'
Alert sent successfully to webhook.

--- Verification Cycle: 2026-04-14T22:35:29-03:00 ---
System in compliance.
```

### 3. Kubernetes Auditing

```bash
./godriftdetector --provider k8s --namespace production

# Outputs Pods running in production namespace and compares with k8s-manifest.yaml
```

### 4. CI/CD Integration

```bash
# In your CI pipeline
if ! ./godriftdetector --provider docker --json | jq -e '.Drifts | length == 0' > /dev/null; then
  echo "Infrastructure drift detected!"
  exit 1
fi
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Integration Tests

```bash
# Start test Docker containers
docker run -d --name test-nginx -p 8080:80 nginx:alpine
docker run -d --name test-redis -p 6379:6379 redis:7-alpine

# Create docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  nginx:
    image: nginx:alpine
    ports:
      - "8080:80"
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
EOF

# Run detection
LOCAL_CONFIG_DIR=. ./godriftdetector --json

# Cleanup
docker rm -f test-nginx test-redis
```

### Environment Variable Drift Testing

```bash
# Create test compose file with environment variables
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  myapp:
    image: nginx:alpine
    environment:
      APP_ENV: production
      DB_PASSWORD: secret123
      API_TOKEN: mytoken
EOF

# Start container with DIFFERENT values
docker run -d --name myapp \
  -e APP_ENV=staging \
  -e DB_PASSWORD=different_secret \
  -e API_TOKEN=different_token \
  -e EXTRA_VAR=injected_value \
  nginx:alpine

# Run detection to see ENV_MISMATCH and ENV_INJECTED
LOCAL_CONFIG_DIR=. ./godriftdetector --json

# Cleanup
docker rm -f myapp
```

### Prometheus Metrics Testing

```bash
# Start daemon
./godriftdetector &
DAEMON_PID=$!

# Wait for metrics server to start
sleep 2

# Query metrics endpoint
curl http://localhost:9090/metrics | grep drift_

# Expected output:
# drift_detected_total 3
# drift_by_service{service="app",type="ENV_MISMATCH"} 1
# drift_by_service{service="app",type="ENV_INJECTED"} 2
# last_scan_timestamp 1713139967

# Stop daemon
kill $DAEMON_PID
```

## Architecture

GoDriftDetector follows a clean architecture with clear separation of concerns:

<div align="center">
  <img src="assets/architecture.svg" alt="GoDriftDetector Architecture Diagram" width="800px">
</div>

### Package Structure

```
godriftdetector/
├── cmd/
│   └── godriftdetector/
│       └── main.go              # CLI entry point and daemon orchestration
├── internal/
│   ├── domain/
│   │   ├── container.go         # ContainerState, DesiredState, ServiceConfig
│   │   ├── drift.go             # Drift types and ComparisonResult
│   │   ├── comparator.go        # Core drift comparison logic
│   │   └── provider.go          # Provider interfaces (InfrastructureProvider, DesiredStateReader)
│   └── infra/
│       ├── docker.go            # DockerProvider implementation
│       ├── kubernetes.go        # KubernetesProvider implementation
│       ├── compose.go           # Docker Compose YAML parser
│       ├── k8s_manifest.go      # Kubernetes manifest parser
│       ├── git.go               # Git repository synchronization
│       ├── webhook.go           # Slack/Discord webhook notifications
│       └── metrics.go           # Prometheus metrics exposition
└── docker-compose.yaml          # Example configuration
```

### Design Patterns

- **Provider Pattern**: `InfrastructureProvider` interface for pluggable providers
- **Adapter Pattern**: `ComposeReader` and `K8sManifestReader` adapt configuration files
- **Observer Pattern**: Webhook notifications on drift detection
- **Singleton Pattern**: Prometheus metrics registry

## Drift Types

| Type | Scenario | Example |
|------|----------|---------|
| **MISSING** | Service declared but not running | Service 'db' (postgres:15) not running |
| **SHADOW_IT** | Container running but not declared | Container abc123 (redis:latest) not declared |
| **PORT_MISMATCH** | Desired port not found on container | Port 443:443 not found on nginx |
| **IMAGE_MISMATCH** | Container image differs from config | Config: nginx:1.20, Container: nginx:1.21 |
| **ENV_MISMATCH** | Environment variable value differs | APP_ENV: Expected 'prod***', Actual 'stag***' |
| **ENV_INJECTED** | Undeclared environment variable running | EXTRA_VAR=malicious_value |

## Providers

### Docker Provider

- **Reads from**: Docker daemon via Docker SDK
- **Configuration file**: `docker-compose.yaml`
- **Extracts**: Container ID, Image, Ports, Environment Variables

**Usage:**
```bash
./godriftdetector --provider docker
```

### Kubernetes Provider

- **Reads from**: Kubernetes API (kubectl configured)
- **Configuration file**: `k8s-manifest.yaml`
- **Extracts**: Pod name, Container image, Ports, ConfigMaps, Secrets

**Usage:**
```bash
./godriftdetector --provider k8s --namespace production

# With custom kubeconfig
export KUBECONFIG=/path/to/kubeconfig.yaml
./godriftdetector --provider k8s
```

## Observability

### Prometheus Metrics

The agent exposes metrics on port **9090** in Prometheus format:

```bash
curl http://localhost:9090/metrics
```

**Available Metrics:**

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `drift_detected_total` | Gauge | - | Total drifts found in last scan |
| `drift_by_service` | Gauge | service, type | Count of drifts per service and type |
| `last_scan_timestamp` | Gauge | - | Unix timestamp of last successful scan |

### Grafana Dashboard

Create a dashboard with these PromQL queries:

```promql
# Total drifts
drift_detected_total

# Drifts by service
sum by (service) (drift_by_service)

# Drifts by type
sum by (type) (drift_by_service)

# Time since last scan
time() - last_scan_timestamp
```

### Logging

All events are logged to stdout with timestamps and color coding:

- 🟢 Green: Successful operations and compliance
- 🟠 Orange: Warnings and configuration issues
- 🔴 Red: Drifts detected and errors

## Examples

### Example 1: Docker Compose with GitOps

```bash
# Repository structure
config-repo/
├── docker-compose.yaml
├── nginx.conf
└── environment.prod

# Setup
export GIT_REPO_URL="https://github.com/org/config-repo.git"
export LOCAL_CONFIG_DIR="/var/lib/godriftdetector/config"
export SYNC_INTERVAL="10m"
export WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Run daemon
./godriftdetector

# Expected flow:
# 1. Clone/pull latest from Git
# 2. Parse docker-compose.yaml
# 3. List running Docker containers
# 4. Compare states
# 5. Expose metrics
# 6. Send Slack alerts if drifts found
# 7. Repeat every 10 minutes
```

### Example 2: Kubernetes with Namespaces

```bash
# Monitor multiple namespaces separately
for ns in default production staging; do
  echo "Auditing namespace: $ns"
  ./godriftdetector --provider k8s --namespace $ns --json | \
    jq '.Drifts | length' >> drift-count-$ns.txt
done
```

### Example 3: CI/CD Pipeline Integration

```yaml
# .github/workflows/drift-check.yml
name: Infrastructure Drift Check

on:
  push:
    paths:
      - 'docker-compose.yaml'
      - 'k8s-manifest.yaml'

jobs:
  drift-detection:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.26'

      - name: Build GoDriftDetector
        run: go build -o godriftdetector ./cmd/godriftdetector

      - name: Run Drift Detection
        run: |
          LOCAL_CONFIG_DIR=. ./godriftdetector --json | \
            jq -e '.Drifts | length == 0' || exit 1

      - name: Notify on Drift
        if: failure()
        run: |
          curl -X POST ${{ secrets.SLACK_WEBHOOK }} \
            -H 'Content-Type: application/json' \
            -d '{"text":"Infrastructure drift detected in commit"}'
```

### Example 4: Docker Networking with Environment Variables

```bash
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  backend:
    image: myapp:1.0
    environment:
      DATABASE_URL: postgres://db:5432/myapp
      API_KEY: supersecret_xyz
      LOG_LEVEL: info
    depends_on:
      - db
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_PASSWORD: dbsecret_abc
EOF

# Start services
docker-compose up -d

# Test with modified environment
docker rm -f backend
docker run -d --name backend \
  -e DATABASE_URL=postgres://db:5432/different \
  -e API_KEY=different_secret \
  -e LOG_LEVEL=debug \
  -e INJECTED_VAR=malicious \
  myapp:1.0

# Detect drifts
LOCAL_CONFIG_DIR=. ./godriftdetector --json
# Shows ENV_MISMATCH for DATABASE_URL, API_KEY, LOG_LEVEL
# Shows ENV_INJECTED for INJECTED_VAR
# Masks values: API_KEY='sup***'
```

## Roadmap

- [x] Container downtime detection (MISSING)
- [x] Shadow IT detection (SHADOW_IT)
- [x] Port and image version mismatch detection
- [x] Remote Git synchronization (GitOps)
- [x] Webhook alerts (Slack/Discord)
- [x] JSON audit reports
- [x] Environment variable drift detection with masking
- [x] Kubernetes provider support
- [x] Prometheus metrics exposition

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md) for details on running tests, linting, and opening PRs.

## License

[MIT License](./LICENSE)

<div align="center">

## Author

**Enoque Sousa**

[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=flat&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/enoque-sousa-bb89aa168/)
[![GitHub](https://img.shields.io/badge/GitHub-100000?style=flat&logo=github&logoColor=white)](https://github.com/esousa97)
[![Portfolio](https://img.shields.io/badge/Portfolio-FF5722?style=flat&logo=target&logoColor=white)](https://enoquesousa.vercel.app)

**[⬆ Back to Top](#godriftdetector)**

Made with ❤️ by [Enoque Sousa](https://github.com/esousa97)

**Project Status:** Complete — Ready for Production Use

</div>
