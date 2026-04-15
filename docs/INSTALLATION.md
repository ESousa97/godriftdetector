# GoDriftDetector - Installation Guide

## Quick Install

### Prerequisites

- **Go 1.25.0+** ([Download](https://golang.org/dl/))
- **Docker 20.10+** (for Docker provider) OR **Kubernetes 1.19+** (for K8s provider)
- **git** (for GitOps features)

### Option 1: Binary Installation

The easiest way to get started:

```bash
# Install from GitHub
go install github.com/esousa97/godriftdetector/cmd/godriftdetector@latest

# Verify installation
godriftdetector --help
```

The binary will be installed in `$GOPATH/bin` (usually `$HOME/go/bin`).

### Option 2: From Source

Clone and build locally:

```bash
# Clone repository
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector

# Build binary
go build -o godriftdetector ./cmd/godriftdetector

# Run
./godriftdetector --help

# (Optional) Install to system PATH
sudo cp godriftdetector /usr/local/bin/
```

### Option 3: Docker Container

Run directly without installation:

```bash
# Docker provider
docker run --rm \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/docker-compose.yaml:/docker-compose.yaml \
  -e LOCAL_CONFIG_DIR=/ \
  ghcr.io/esousa97/godriftdetector:latest

# Kubernetes provider
docker run --rm \
  -v $HOME/.kube/config:/root/.kube/config \
  -e KUBECONFIG=/root/.kube/config \
  ghcr.io/esousa97/godriftdetector:latest \
  --provider k8s --namespace production
```

## Platform-Specific Setup

### macOS

```bash
# Using Homebrew (when available)
brew install godriftdetector

# Or build from source
brew install go git
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector
go build -o godriftdetector ./cmd/godriftdetector
sudo mv godriftdetector /usr/local/bin/
```

### Linux (Ubuntu/Debian)

```bash
# Install dependencies
sudo apt-get update
sudo apt-get install -y golang-go git docker.io

# Build from source
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector
go build -o godriftdetector ./cmd/godriftdetector
sudo mv godriftdetector /usr/local/bin/

# Run as systemd service (optional)
sudo tee /etc/systemd/system/godriftdetector.service > /dev/null <<EOF
[Unit]
Description=GoDriftDetector Infrastructure Drift Detector
After=docker.service

[Service]
Type=simple
User=godriftdetector
WorkingDirectory=/var/lib/godriftdetector
Environment="GIT_REPO_URL=https://github.com/org/config-repo.git"
Environment="LOCAL_CONFIG_DIR=/var/lib/godriftdetector/config"
Environment="SYNC_INTERVAL=5m"
Environment="WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK"
ExecStart=/usr/local/bin/godriftdetector
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable godriftdetector
sudo systemctl start godriftdetector
```

### Windows (PowerShell)

```powershell
# Install Go 1.25+ from https://golang.org/dl/

# Clone and build
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector
go build -o godriftdetector.exe ./cmd/godriftdetector

# Add to PATH or run directly
.\godriftdetector --help

# Run as Windows Service (using NSSM - Non-Sucking Service Manager)
# Download NSSM from https://nssm.cc/download
nssm install GoDriftDetector C:\path\to\godriftdetector.exe
nssm set GoDriftDetector AppEnvironmentExtra "GIT_REPO_URL=https://github.com/org/config-repo.git"
nssm start GoDriftDetector
```

### Kubernetes

Deploy as a Kubernetes Deployment:

```bash
# Create namespace
kubectl create namespace drift-detection

# Create ConfigMap with k8s-manifest.yaml
kubectl create configmap drift-config \
  --from-file=k8s-manifest.yaml \
  -n drift-detection

# Apply deployment
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: godriftdetector
  namespace: drift-detection
spec:
  replicas: 1
  selector:
    matchLabels:
      app: godriftdetector
  template:
    metadata:
      labels:
        app: godriftdetector
    spec:
      serviceAccountName: godriftdetector
      containers:
      - name: godriftdetector
        image: ghcr.io/esousa97/godriftdetector:latest
        args:
          - "--provider=k8s"
          - "--namespace=production"
        env:
        - name: LOCAL_CONFIG_DIR
          value: /etc/drift-config
        - name: SYNC_INTERVAL
          value: "5m"
        - name: WEBHOOK_URL
          valueFrom:
            secretKeyRef:
              name: drift-webhook
              key: url
        volumeMounts:
        - name: config
          mountPath: /etc/drift-config
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        ports:
        - containerPort: 9090
          name: metrics
      volumes:
      - name: config
        configMap:
          name: drift-config
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: godriftdetector
  namespace: drift-detection
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: godriftdetector
rules:
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["list", "get"]
- apiGroups: [""]
  resources: ["configmaps"]
  verbs: ["list", "get"]
- apiGroups: [""]
  resources: ["secrets"]
  verbs: ["list", "get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: godriftdetector
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: godriftdetector
subjects:
- kind: ServiceAccount
  name: godriftdetector
  namespace: drift-detection
EOF

# Verify deployment
kubectl get pods -n drift-detection
kubectl logs -n drift-detection -l app=godriftdetector -f
```

## Development Setup

For contributing or extending GoDriftDetector:

```bash
# Clone and setup
git clone https://github.com/esousa97/godriftdetector.git
cd godriftdetector

# Install dependencies
go mod download
go mod tidy

# Run tests
go test -v ./...

# Build locally
go build -o godriftdetector ./cmd/godriftdetector

# Run with debug output
GIT_REPO_URL="" LOCAL_CONFIG_DIR=. ./godriftdetector --json
```

## Verification

After installation, verify everything works:

```bash
# Check version/help
godriftdetector --help

# Quick test with Docker
docker run -d --name test-app -p 8080:80 nginx:alpine

# Create docker-compose.yaml
cat > docker-compose.yaml << 'EOF'
version: '3.8'
services:
  app:
    image: nginx:alpine
    ports:
      - "8080:80"
EOF

# Run detection
LOCAL_CONFIG_DIR=. godriftdetector --json

# Should show: { "Drifts": [] } if in compliance

# Cleanup
docker rm -f test-app
rm docker-compose.yaml
```

## Troubleshooting

### Docker Socket Permission Denied

```bash
# Add current user to docker group
sudo usermod -aG docker $USER
newgrp docker

# Or run with sudo
sudo godriftdetector
```

### Kubernetes Connection Error

```bash
# Verify kubectl configuration
kubectl cluster-info
kubectl auth can-i list pods

# Ensure KUBECONFIG is set
export KUBECONFIG=$HOME/.kube/config
godriftdetector --provider k8s
```

### Git Authentication Issues

```bash
# For HTTPS with personal access token
export GIT_REPO_URL="https://github.com/org/repo.git"
export GIT_USERNAME="your-username"
export GIT_PASSWORD="your-personal-access-token"

# For SSH (ensure ssh-agent is configured)
export GIT_REPO_URL="git@github.com:org/repo.git"
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_rsa
```

## Next Steps

After installation:

1. **Configure**: See [CONFIGURATION.md](./CONFIGURATION.md) for environment variables
2. **Test**: See [TESTING.md](./TESTING.md) for testing instructions
3. **Deploy**: Use Docker or Kubernetes deployment examples above
4. **Monitor**: Check Prometheus metrics on `http://localhost:9090/metrics`

For detailed usage examples, see [README.md](./README.md#examples)
