# GoDriftDetector - Quick Start (5 Minutes)

Get up and running with GoDriftDetector in just 5 minutes!

## Prereq uisites

- Docker installed and running
- Go 1.25+ (or use Docker to run it)

## Option 1: Docker (Fastest)

```bash
# Create configuration
mkdir -p config && cd config

# Create docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: nginx:alpine
    ports:
      - "80:80"
EOF

# Start containers
docker-compose up -d

# Run drift detection from Docker
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/config \
  -e LOCAL_CONFIG_DIR=/config \
  ghcr.io/esousa97/godriftdetector:latest --json

# Output: { "Drifts": [] } ✓ In compliance!

# Test with a mismatch
docker run -d --name undeclared-redis -p 6379:6379 redis:alpine

# Run again
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd):/config \
  -e LOCAL_CONFIG_DIR=/config \
  ghcr.io/esousa97/godriftdetector:latest --json

# Output shows SHADOW_IT drift for redis!
```

## Option 2: Binary (Recommended for Linux/macOS)

```bash
# Install
go install github.com/esousa97/godriftdetector/cmd/godriftdetector@latest

# Verify
godriftdetector --help

# Create test setup
mkdir -p myproject && cd myproject

# Configuration
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  postgres:
    image: postgres:15-alpine
    ports:
      - "5432:5432"
    environment:
      POSTGRES_PASSWORD: secretpassword
EOF

# Start service
docker-compose up -d

# Run detection
LOCAL_CONFIG_DIR=. godriftdetector --json

# Output: { "Drifts": [] } ✓
```

## Option 3: From Source

```bash
# Clone
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector

# Build
go build -o godriftdetector ./cmd/godriftdetector

# Run
LOCAL_CONFIG_DIR=. ./godriftdetector --json
```

## First Test: Detect a Drift

```bash
# 1. Create configuration
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  webserver:
    image: nginx:1.20-alpine  # Specific version
    ports:
      - "8080:80"
    environment:
      WORKER_THREADS: "4"
      LOG_LEVEL: "info"
EOF

# 2. Start container with DIFFERENT version and env
docker run -d --name webserver \
  -p 8080:80 \
  -e WORKER_THREADS=2 \
  -e LOG_LEVEL=debug \
  -e EXTRA_VAR=injected \
  nginx:alpine

# 3. Run detection
godriftdetector --json

# 4. See the drifts detected:
# - IMAGE_MISMATCH: nginx:1.20-alpine vs nginx:alpine
# - PORT_MISMATCH: (might be implicit)
# - ENV_MISMATCH: WORKER_THREADS (4 vs 2) and LOG_LEVEL (info vs debug)
# - ENV_INJECTED: EXTRA_VAR
```

## Daemon Mode (Continuous Monitoring)

```bash
# Run in background, checking every 30 seconds
SYNC_INTERVAL=30s LOCAL_CONFIG_DIR=. godriftdetector

# Output:
# --- Verification Cycle: 2026-04-14T10:00:00Z ---
# DRIFT DETECTED!
# [ ENV_MISMATCH ] ...
# [ ENV_INJECTED ] ...

# Press Ctrl+C to stop
```

## With Slack Notifications

```bash
# 1. Create Slack webhook at:
#    https://api.slack.com/apps → Create New App → Incoming Webhooks

# 2. Set webhook URL
export WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# 3. Run with notifications
LOCAL_CONFIG_DIR=. godriftdetector

# When drifts found, you get Slack alert! 🚨
```

## Kubernetes (If You Have a Cluster)

```bash
# 1. Create manifest
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
        image: nginx:1.20-alpine
        ports:
        - containerPort: 80
        env:
        - name: ENVIRONMENT
          value: "production"
EOF

# 2. Deploy
kubectl apply -f k8s-manifest.yaml

# 3. Run detection
godriftdetector --provider k8s --namespace default --json

# Works the same way with Kubernetes!
```

## CI/CD Integration

```bash
# In your GitHub Actions / GitLab CI:

# One-shot audit
godriftdetector --json > drift-report.json

# Fail if drifts found
jq -e '.Drifts | length == 0' drift-report.json || exit 1
```

## Understanding Output

### Compliant (No Drifts)

```bash
./godriftdetector --json
# Output:
{
  "Drifts": []
}
```

### Drift Detected

```bash
./godriftdetector --json
# Output:
{
  "Drifts": [
    {
      "ServiceName": "app",
      "Type": "ENV_MISMATCH",
      "Message": "Environment variable 'LOG_LEVEL' differs. Expected: 'inf***', Actual: 'deb***'.",
      "Desired": "inf***",
      "Actual": "deb***"
    },
    {
      "ServiceName": "c803dee1d5ff",
      "Type": "SHADOW_IT",
      "Message": "Undeclared container running: c803dee1d5ff (Image: redis:latest)",
      "Actual": "redis:latest"
    }
  ]
}
```

### Terminal Output (Daemon Mode)

```bash
Initiating GoDriftDetector Agent (Interval: 5m, Provider: docker)
Exposing metrics at http://localhost:9090/metrics

--- Verification Cycle: 2026-04-14T10:00:00Z ---
DRIFT DETECTED!
[ ENV_MISMATCH ] Environment variable 'LOG_LEVEL' outdated...
[ SHADOW_IT ] Undeclared container running: redis:alpine
Alert sent successfully to webhook.
```

## Drift Types Explained

| Type | Meaning | Example |
|------|---------|---------|
| **MISSING** | Service declared but not running | Service 'db' not found running |
| **SHADOW_IT** | Container running but not declared | Undeclared redis container |
| **PORT_MISMATCH** | Different ports than configured | Expected 80:80, found 8080:80 |
| **IMAGE_MISMATCH** | Different image version | Expected nginx:1.20, found nginx:1.21 |
| **ENV_MISMATCH** | Different env variable value | APP_ENV: production vs staging |
| **ENV_INJECTED** | Extra env variable found | EXTRA_VAR not in config |

## Monitoring (Prometheus)

```bash
# 1. Run daemon (starts metrics on port 9090)
godriftdetector &

# 2. Check metrics
curl http://localhost:9090/metrics | grep drift_

# Output:
# drift_detected_total 2
# drift_by_service{service="app",type="ENV_MISMATCH"} 1
# drift_by_service{service="redis",type="SHADOW_IT"} 1
# last_scan_timestamp 1713139967
```

## Cleanup

```bash
# Stop containers
docker-compose down

# Or manually
docker rm -f container-name

# Delete config
rm docker-compose.yaml k8s-manifest.yaml
```

## Next Steps

1. **Full README**: [README.md](./README.md)
2. **Installation**: [INSTALLATION.md](./INSTALLATION.md)
3. **Configuration**: [CONFIGURATION.md](./CONFIGURATION.md)
4. **Testing**: [TESTING.md](./TESTING.md)

## Common Commands Cheat Sheet

```bash
# One-shot audit
godriftdetector --json

# Daemon with polling
godriftdetector

# Kubernetes
godriftdetector --provider k8s

# Specific namespace
godriftdetector --provider k8s --namespace production

# With Slack
export WEBHOOK_URL="https://hooks.slack.com/..."
godriftdetector

# With Git sync
export GIT_REPO_URL="https://github.com/org/config.git"
export LOCAL_CONFIG_DIR="/tmp/config"
godriftdetector

# Custom polling interval
export SYNC_INTERVAL="30s"
godriftdetector

# All together
export GIT_REPO_URL="https://github.com/org/config.git"
export GIT_USERNAME="token"
export GIT_PASSWORD="ghp_xxxx"
export LOCAL_CONFIG_DIR="/var/lib/godriftdetector"
export SYNC_INTERVAL="5m"
export WEBHOOK_URL="https://hooks.slack.com/services/..."
godriftdetector --provider docker
```

## Troubleshooting

### Docker socket permission denied
```bash
sudo usermod -aG docker $USER
# Log out and back in
```

### Config file not found
```bash
# Set correct path
export LOCAL_CONFIG_DIR=/path/to/config
# Ensure docker-compose.yaml exists
ls $LOCAL_CONFIG_DIR/docker-compose.yaml
```

### Port 9090 already in use
```bash
# Kill process
lsof -i :9090
kill -9 <PID>
```

### Git authentication fails
```bash
# Use personal access token instead of password
export GIT_USERNAME="myuser"
export GIT_PASSWORD="ghp_xxxxxxxxxxxxxxx"
```

---

**That's it!** You're now monitoring infrastructure drift. 🎉

For more details, see the full documentation.
