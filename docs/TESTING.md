# GoDriftDetector - Testing Guide

## Unit Tests

Run the complete test suite:

```bash
# Run all tests
go test ./...

# With verbose output
go test -v ./...

# With coverage report
go test -cover ./...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Test Files

- `internal/domain/comparator_test.go` - Drift comparison logic
- `internal/infra/compose_test.go` - Docker Compose parsing
- Tests for Docker and Kubernetes providers

## Integration Tests

### Test 1: Basic Docker Drift Detection

```bash
# Setup
docker run -d --name test-nginx -p 8080:80 nginx:alpine
docker run -d --name test-redis -p 6379:6379 redis:7-alpine

# Create configuration with DIFFERENT values than running containers
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  nginx:
    image: nginx:latest
    ports:
      - "9000:80"
    environment:
      NGINX_VERSION: "1.20"
  redis:
    image: redis:6-alpine
    ports:
      - "6380:6379"
    environment:
      REDIS_PASSWORD: secret123
EOF

# Run detection - should find drifts
LOCAL_CONFIG_DIR=. ./godriftdetector --json

# Expected output: Multiple drifts
# - IMAGE_MISMATCH (nginx:alpine vs nginx:latest)
# - PORT_MISMATCH (8080:80 vs 9000:80)
# - MISSING (redis:6-alpine not found)
# - ENV_MISMATCH (NGINX_VERSION, REDIS_PASSWORD)

# Cleanup
docker rm -f test-nginx test-redis
rm docker-compose.yaml
```

### Test 2: Environment Variable Drift

```bash
# Create test configuration
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: nginx:alpine
    environment:
      APP_ENV: production
      DB_PASSWORD: super_secret_123
      API_TOKEN: mytoken_xyz
      LOG_LEVEL: info
EOF

# Start container with DIFFERENT environment variables
docker run -d --name app \
  -e APP_ENV=staging \
  -e DB_PASSWORD=different_secret \
  -e API_TOKEN=different_token \
  -e EXTRA_INJECTED=malicious_code \
  nginx:alpine

sleep 2

# Run detection
echo "=== Expected: ENV_MISMATCH and ENV_INJECTED drifts ==="
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts[] | select(.Type | match("ENV"))'

# Expected output:
# - ENV_MISMATCH for APP_ENV (production vs staging)
# - ENV_MISMATCH for DB_PASSWORD (sup*** vs dif***)
# - ENV_MISMATCH for API_TOKEN (myt*** vs dif***)
# - ENV_INJECTED for EXTRA_INJECTED (malicious_code)

# Cleanup
docker rm -f app
rm docker-compose.yaml
```

### Test 3: Shadow IT Detection

```bash
# Create minimal configuration
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  declared-service:
    image: nginx:alpine
EOF

# Start Docker containers NOT declared in compose
docker run -d --name undeclared-redis -p 6379:6379 redis:7-alpine
docker run -d --name undeclared-postgres -p 5432:5432 postgres:15-alpine

sleep 2

# Run detection
echo "=== Expected: SHADOW_IT drifts for undeclared containers ==="
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts[] | select(.Type == "SHADOW_IT")'

# Cleanup
docker rm -f undeclared-redis undeclared-postgres
rm docker-compose.yaml
```

### Test 4: Prometheus Metrics

```bash
# Create simple configuration
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  nginx:
    image: nginx:alpine
EOF

# Start agent in daemon mode
./godriftdetector --provider docker &
DAEMON_PID=$!

# Wait for metrics server to start
sleep 3

# Query metrics endpoint
echo "=== Prometheus Metrics ==="
curl -s http://localhost:9090/metrics | grep "^drift_" | head -10

# Check specific metrics
echo "=== Total Drifts ==="
curl -s http://localhost:9090/metrics | grep "drift_detected_total"

echo "=== Drifts by Service ==="
curl -s http://localhost:9090/metrics | grep "drift_by_service" | grep -v "^#"

echo "=== Last Scan Timestamp ==="
curl -s http://localhost:9090/metrics | grep "last_scan_timestamp"

# Stop daemon
kill $DAEMON_PID
wait $DAEMON_PID 2>/dev/null || true

rm docker-compose.yaml
```

### Test 5: Daemon Polling

```bash
# Create configuration
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: nginx:alpine
EOF

# Start container
docker run -d --name app nginx:alpine

# Run daemon with 3-second polling
echo "=== Running daemon for 15 seconds (5 cycles with 3s interval) ==="
LOCAL_CONFIG_DIR=. SYNC_INTERVAL=3s timeout 15s ./godriftdetector

# Expected output: 5 verification cycles
# System should report compliance after each cycle

# Cleanup
docker rm -f app
rm docker-compose.yaml
```

### Test 6: JSON One-Shot Mode

```bash
# Create scenario with drifts
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  web:
    image: nginx:alpine
    ports:
      - "80:80"
    environment:
      ENVIRONMENT: production
      DEBUG: "false"
EOF

# Start container with different values
docker run -d --name web \
  -p 8080:80 \
  -e ENVIRONMENT=development \
  -e DEBUG=true \
  -e EXTRA_VAR=injected \
  nginx:alpine

sleep 2

# Get JSON report
echo "=== JSON Audit Report ==="
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq .

# Verify schema
echo "=== Report Schema Validation ==="
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq -e '.Drifts | type == "array"' && echo "✓ Valid schema"

# Count drifts by type
echo "=== Drift Distribution ==="
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts | group_by(.Type) | map({type: .[0].Type, count: length})'

# Cleanup
docker rm -f web
rm docker-compose.yaml
```

### Test 7: Kubernetes Provider (requires running cluster)

```bash
# Create Kubernetes manifest
cat > k8s-manifest.yaml << 'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
spec:
  template:
    spec:
      containers:
      - name: app
        image: nginx:1.20-alpine
        ports:
        - containerPort: 80
        env:
        - name: APP_ENV
          value: production
EOF

# Deploy to cluster
kubectl apply -f k8s-manifest.yaml

# Wait for pod
sleep 10

# Run drift detection for Kubernetes
./godriftdetector --provider k8s --namespace default --json

# Cleanup
kubectl delete -f k8s-manifest.yaml
rm k8s-manifest.yaml
```

### Test 8: Multiple Providers Comparison

```bash
# Test switching between Docker and Kubernetes

# Test Docker provider (default)
echo "=== Docker Provider ==="
LOCAL_CONFIG_DIR=. ./godriftdetector --provider docker --json

# Test Kubernetes provider
echo "=== Kubernetes Provider ==="
./godriftdetector --provider k8s --namespace default --json

# Both should work seamlessly with different --provider flags
```

## Performance Tests

### Load Testing with Many Containers

```bash
# Create 10 test containers
for i in {1..10}; do
  docker run -d --name test-app-$i -p 800$i:80 nginx:alpine > /dev/null
done

# Create matching compose
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
EOF

for i in {1..10}; do
  cat >> docker-compose.yaml << 'EOF'
  app-$i:
    image: nginx:alpine
    ports:
      - "800$i:80"
EOF
done

# Run detection and measure time
echo "=== Performance Test: 10 containers ==="
time LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts | length'

# Cleanup
for i in {1..10}; do
  docker rm -f test-app-$i > /dev/null 2>&1
done
rm docker-compose.yaml
```

## CI/CD Integration Tests

### GitHub Actions Example

```yaml
name: Drift Detection Test

on: [push]

jobs:
  drift-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: actions/setup-go@v4
        with:
          go-version: '1.26'

      - name: Build
        run: go build -o godriftdetector ./cmd/godriftdetector

      - name: Start test services
        run: |
          docker-compose up -d
          sleep 5

      - name: Run drift detection
        run: |
          ./godriftdetector --json > drift-report.json
          cat drift-report.json

      - name: Validate no drifts
        run: |
          jq -e '.Drifts | length == 0' drift-report.json || exit 1

      - name: Archive report
        if: always()
        uses: actions/upload-artifact@v3
        with:
          name: drift-reports
          path: drift-report.json
```

## Debugging

### Enable Verbose Logging

```bash
# Set environment variables for debugging
export RUST_LOG=debug
export DEBUG=1

./godriftdetector --json
```

### Analyze Specific Drifts

```bash
# Get all drifts of a specific type
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts[] | select(.Type == "ENV_MISMATCH")'

# Get all drifts for a specific service
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts[] | select(.ServiceName == "myapp")'

# Count drifts
LOCAL_CONFIG_DIR=. ./godriftdetector --json | jq '.Drifts | length'
```

### Test Configuration Parsing

```bash
# Verify docker-compose.yaml parsing
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: myapp:1.0
    ports:
      - "8080:8080"
    environment:
      DB_HOST: postgres
      DB_PORT: "5432"
      SECRET_KEY: mykey123
EOF

# Just test parsing (no Docker required)
LOCAL_CONFIG_DIR=. ./godriftdetector --json 2>&1 | jq '.Drifts' || echo "Parsing error"
```

## Test Checklist

- [ ] Unit tests pass: `go test ./...`
- [ ] Docker provider detects all 6 drift types
- [ ] Kubernetes provider works (requires cluster)
- [ ] Environment variable masking works
- [ ] Prometheus metrics endpoint responds
- [ ] Daemon mode polling works
- [ ] JSON one-shot mode produces valid output
- [ ] Webhook notifications send (requires webhook URL)
- [ ] Graceful shutdown on SIGTERM
- [ ] Performance acceptable (< 1s for 10 containers)

## Common Issues

### Tests Fail: Docker Socket

```bash
# Solution: Ensure Docker daemon is running
docker ps

# Or use Docker Desktop
# Check: Docker -> Preferences -> Resources -> File Sharing
```

### Tests Fail: Port Already in Use

```bash
# Solution: Kill process using the port
lsof -i :9090
kill -9 <PID>

# Or use different port (not configurable, restart)
```

### Tests Fail: Kubernetes Connection

```bash
# Solution: Verify cluster access
kubectl cluster-info
kubectl auth can-i list pods

# Or skip K8s tests if no cluster available
go test -run "!TestKubernetes" ./...
```

## Next Steps

After testing:

1. **Deploy**: See [INSTALLATION.md](./INSTALLATION.md) for deployment
2. **Monitor**: Setup Prometheus/Grafana for metrics
3. **Integrate**: Setup Slack/Discord webhooks
4. **Automate**: Add to CI/CD pipeline
