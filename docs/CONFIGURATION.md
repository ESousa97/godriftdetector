# GoDriftDetector - Configuration Guide

## Environment Variables

All configuration is done via environment variables and command-line flags.

### Core Configuration

| Variable | Type | Default | Required | Description |
|----------|------|---------|----------|-------------|
| `GIT_REPO_URL` | String | `""` | No | HTTPS or SSH URL to Git repository containing configuration files (docker-compose.yaml or k8s-manifest.yaml) |
| `GIT_USERNAME` | String | `""` | No | Username or personal access token for HTTPS Git authentication |
| `GIT_PASSWORD` | String | `""` | No | Password or personal access token for HTTPS Git authentication |
| `LOCAL_CONFIG_DIR` | String | `"./config-repo"` | No | Local directory path where Git repository is cloned/cached. Config files must be at: `{LOCAL_CONFIG_DIR}/docker-compose.yaml` or `{LOCAL_CONFIG_DIR}/k8s-manifest.yaml` |
| `SYNC_INTERVAL` | Duration | `"5m"` | No | How often to re-check infrastructure. Format: `10s`, `5m`, `1h` |
| `WEBHOOK_URL` | String | `""` | No | Slack, Discord, or custom webhook URL to receive drift alerts |

### Environment Variable Examples

```bash
# Minimal setup (local docker-compose.yaml)
export LOCAL_CONFIG_DIR=.
./godriftdetector

# Full GitOps setup
export GIT_REPO_URL="https://github.com/myorg/infrastructure.git"
export GIT_USERNAME="your-username-or-token"
export GIT_PASSWORD="your-personal-access-token"
export LOCAL_CONFIG_DIR="/var/lib/godriftdetector/config"
export SYNC_INTERVAL="5m"
export WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
./godriftdetector

# Kubernetes setup
export KUBECONFIG=$HOME/.kube/config
./godriftdetector --provider k8s --namespace production
```

## Command-Line Flags

```bash
./godriftdetector [flags]
```

### Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-json` | Boolean | `false` | Run once, output JSON report to stdout, and exit (useful for CI/CD) |
| `-provider` | String | `"docker"` | Infrastructure provider: `docker` or `k8s` |
| `-namespace` | String | `"default"` | Kubernetes namespace to monitor (only with `--provider=k8s`) |
| `-help`, `-h` | - | - | Show help message and exit |

### Flag Examples

```bash
# One-shot Docker audit
./godriftdetector --json

# One-shot Kubernetes audit
./godriftdetector --provider k8s --namespace production --json

# Daemon mode (default)
./godriftdetector --provider docker

# Help
./godriftdetector --help
```

## Configuration Files

### Docker Compose Configuration

GoDriftDetector reads `docker-compose.yaml` or `docker-compose.yml`:

**Location**: `{LOCAL_CONFIG_DIR}/docker-compose.yaml`

**Format**: Standard Docker Compose v3.8+ YAML

```yaml
version: '3.8'

services:
  # Service configuration
  myapp:
    image: myapp:1.0.0
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      DATABASE_URL: "postgres://db:5432/myapp"
      API_KEY: "secret_key_123"
      LOG_LEVEL: "info"
      ENVIRONMENT: "production"
    volumes:
      - /data:/app/data
      - /config:/app/config

  database:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: "db_password_123"
      POSTGRES_DB: "myapp"

  cache:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    environment:
      REDIS_PASSWORD: "redis_password_456"
```

**What GoDriftDetector reads**:
- `services` - Service definitions
- `image` - Container image with tag
- `ports` - Port mappings (container:host)
- `environment` - Environment variables
- `volumes` - Volume mounts (for reference)

**Not checked**:
- `volumes` section (reference only)
- `networks`, `depends_on`, `healthcheck`
- Build-time properties

### Kubernetes Manifest Configuration

GoDriftDetector reads `k8s-manifest.yaml`:

**Location**: `{LOCAL_CONFIG_DIR}/k8s-manifest.yaml`

**Format**: Standard Kubernetes YAML manifests

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: myapp
        image: myapp:1.0.0
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: DATABASE_URL
          value: "postgres://db:5432/myapp"
        - name: API_KEY
          valueFrom:
            secretKeyRef:
              name: api-secrets
              key: api_key
        - name: LOG_LEVEL
          value: "info"
        envFrom:
        - configMapRef:
            name: app-config
        - secretRef:
            name: app-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  ENVIRONMENT: "production"
  CACHE_TTL: "3600"
---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  REDIS_PASSWORD: cmVkaXNfcGFzc3dvcmRfNDU2  # base64 encoded
```

**What GoDriftDetector reads**:
- `containers[].image` - Container image with tag
- `containers[].ports` - Port definitions
- `containers[].env` - Literal environment variables
- `containers[].envFrom` - ConfigMaps and Secrets
- `containers[].name` - Container name (for service mapping)

**Not checked**:
- `replicas` or scaling
- `resources.limits/requests`
- `livenessProbe`, `readinessProbe`
- `imagePullPolicy`, `securityContext`

## Examples

### Example 1: Simple Docker Setup

```bash
# 1. Create docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  web:
    image: nginx:1.24-alpine
    ports:
      - "80:80"
    environment:
      NGINX_VERSION: "1.24"
      WORKER_PROCESSES: "4"
EOF

# 2. Start containers
docker-compose up -d

# 3. Run drift detection
LOCAL_CONFIG_DIR=. ./godriftdetector

# Output: System in compliance.
```

### Example 2: GitOps with GitHub

```bash
# 1. Create GitHub Personal Access Token
# Go to: Settings → Developer settings → Personal access tokens → Tokens (classic)
# Scopes: repo, read:repo_hook

# 2. Setup environment
export GIT_REPO_URL="https://github.com/myorg/k8s-config.git"
export GIT_USERNAME="myusername"
export GIT_PASSWORD="ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export LOCAL_CONFIG_DIR="$HOME/.godriftdetector/config"
export SYNC_INTERVAL="5m"

# 3. Run daemon
./godriftdetector --provider docker

# Output:
# Initiating GoDriftDetector Agent (Interval: 5m, Provider: docker)
# --- Verification Cycle: 2026-04-14T10:00:00Z ---
# Syncing Git repository...
# Comparing states...
# System in compliance.
```

### Example 3: Slack Notifications

```bash
# 1. Create Slack Webhook
# Go to: your-workspace.slack.com → Apps → Manage Apps → Custom Integrations → Incoming Webhooks
# Example webhook URL:
# https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX

# 2. Setup environment
export WEBHOOK_URL="https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"
export LOCAL_CONFIG_DIR=.

# 3. Run daemon
./godriftdetector

# When drifts detected:
# Slack message appears in configured channel with:
# - 🚨 Drift Detected alert
# - List of all drifts with types
# - Timestamp of detection
```

### Example 4: Prometheus Monitoring

```bash
# 1. Prometheus configuration
cat > prometheus.yml << 'EOF'
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'godriftdetector'
    static_configs:
      - targets: ['localhost:9090']
EOF

# 2. Start Prometheus
docker run -d \
  -p 9091:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus

# 3. Run GoDriftDetector
./godriftdetector

# 4. Access Prometheus
# http://localhost:9091/graph
# Query: drift_detected_total
# Query: drift_by_service
# Query: last_scan_timestamp
```

### Example 5: Kubernetes Namespace Monitoring

```bash
# 1. Create k8s-manifest.yaml
cat > k8s-manifest.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
spec:
  template:
    spec:
      containers:
      - name: api
        image: api:v1.0
        ports:
        - containerPort: 3000
        env:
        - name: DATABASE_URL
          value: "postgres://db:5432/app"
        - name: DEBUG
          value: "false"
EOF

# 2. Deploy to K8s
kubectl apply -f k8s-manifest.yaml

# 3. Monitor specific namespace
./godriftdetector --provider k8s --namespace production

# 4. Or audit with JSON
./godriftdetector --provider k8s --namespace production --json
```

### Example 6: Environment Variable Sensitivity

```bash
# Configuration with sensitive data
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: myapp:1.0
    environment:
      # These will be masked as: pass***, token***, secret***
      DATABASE_PASSWORD: "super_secret_password_123"
      API_TOKEN: "my_secret_token_xyz"
      SECRET_KEY: "encryption_key_abc"
      # These will NOT be masked
      LOG_LEVEL: "info"
      ENVIRONMENT: "production"
EOF

# Run detection
LOCAL_CONFIG_DIR=. ./godriftdetector --json

# Output shows masked sensitive values:
# "message": "Environment variable 'DATABASE_PASSWORD' differs. Expected: 'sup***', Actual: 'dif***'"
```

### Example 7: CI/CD Integration

```bash
#!/bin/bash
# ci-drift-check.sh

set -e

echo "Building GoDriftDetector..."
go build -o godriftdetector ./cmd/godriftdetector

echo "Running drift detection..."
REPORT=$(./godriftdetector --json)

DRIFT_COUNT=$(echo "$REPORT" | jq '.Drifts | length')

echo "Found $DRIFT_COUNT drifts"

if [ "$DRIFT_COUNT" -gt 0 ]; then
  echo "Drifts detected! Reporting:"
  echo "$REPORT" | jq '.Drifts'
  exit 1
fi

echo "✓ Infrastructure in compliance"
exit 0
```

## Troubleshooting Configuration

### Issue: "Config file not found"

**Solution**: Ensure `LOCAL_CONFIG_DIR` points to correct directory

```bash
# Check what GoDriftDetector is looking for
ls -la ./config-repo/docker-compose.yaml

# Or set correct path
export LOCAL_CONFIG_DIR=/path/to/config
```

### Issue: "Failed to parse YAML"

**Solution**: Validate YAML syntax

```bash
# Install yamllint
pip install yamllint

# Check syntax
yamllint docker-compose.yaml

# Or use Docker
docker run --rm -v $(pwd):/workspace -w /workspace cytopia/yamllint docker-compose.yaml
```

### Issue: Git authentication fails

**Solution**: Verify Git credentials

```bash
# Test HTTPS auth
git clone https://github.com/myorg/config.git --username myuser --password mytoken

# Test SSH auth
ssh -T git@github.com

# Or use SSH key
export GIT_REPO_URL="git@github.com:myorg/config.git"
```

### Issue: Prometheus metrics not accessible

**Solution**: Check metrics server binding

```bash
# Verify port is open
netstat -tuln | grep 9090

# Or use curl
curl http://localhost:9090/metrics

# Check firewall
sudo ufw allow 9090
```

## Performance Tuning

### Adjust Polling Interval

```bash
# Quick polling (30 seconds)
export SYNC_INTERVAL="30s"
./godriftdetector

# Slow polling (1 hour) - for large production clusters
export SYNC_INTERVAL="1h"
./godriftdetector
```

### Optimize for Large Environments

```bash
# Resource limits for Kubernetes
requests:
  memory: "128Mi"
  cpu: "100m"
limits:
  memory: "512Mi"
  cpu: "500m"

# Git repository best practices
# - Keep manifests in single file or organized directory
# - Use smaller, focused repositories
# - Avoid large binary files
```

### Monitor Resource Usage

```bash
# Docker
docker stats godriftdetector

# Kubernetes
kubectl top pod -n drift-detection

# System
top -p $(pgrep godriftdetector)
```

## Next Steps

1. **Installation**: See [INSTALLATION.md](./INSTALLATION.md)
2. **Testing**: See [TESTING.md](./TESTING.md)
3. **Examples**: See [README.md](./README.md#examples)
4. **Troubleshooting**: See individual sections above
