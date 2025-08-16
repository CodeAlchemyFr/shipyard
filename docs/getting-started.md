# Getting Started

## Installation

### Prerequisites

Choose one of the following options:

#### Option 1: Automatic k3s Installation (Recommended for Development)
- **Operating System** - Linux, macOS, or Windows
- **Docker** - Required for macOS and Windows (for k3d)

#### Option 2: Existing Kubernetes Cluster
- **Kubernetes cluster** - Access to a Kubernetes cluster (local or cloud)
- **kubectl** - Kubernetes command-line tool configured
- **metrics-server** - For CPU/memory metrics collection (recommended for monitoring)

### Quick Install with k3s (Recommended)

This is the fastest way to get started with Shipyard, including a complete Kubernetes environment.

#### macOS & Linux

```bash
curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```

**What gets installed:**
- **Shipyard CLI** - The main CLI tool
- **k3s** (Linux) or **k3d** (macOS) - Lightweight Kubernetes distribution
- **kubectl** configuration - Automatically configured to work with your cluster

#### Windows (PowerShell)

```powershell
Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.ps1" -OutFile "install.ps1"; .\install.ps1
```

**Requirements for Windows:** Docker Desktop must be installed and running.

#### Install Shipyard Only (Skip k3s)

If you already have a Kubernetes cluster configured:

```bash
# macOS & Linux
INSTALL_K3S=false curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash

# Windows PowerShell
$env:INSTALL_K3S="false"; Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.ps1" -OutFile "install.ps1"; .\install.ps1
```

### Package Managers

Support for package managers will be added once the project gains more adoption:

- **Homebrew** (macOS) - `brew install shipyard`
- **Chocolatey** (Windows) - `choco install shipyard`  
- **APT** (Debian/Ubuntu) - `sudo apt install shipyard`

For now, use the [Quick Install](#quick-install-recommended) method above.

### Manual Installation

#### 1. Download Binary

Download the appropriate binary for your platform from [GitHub Releases](https://github.com/CodeAlchemyFr/shipyard/releases/latest):

| Platform | Architecture | Download |
|----------|--------------|----------|
| Linux | AMD64 | `shipyard-linux-amd64` |
| Linux | ARM64 | `shipyard-linux-arm64` |
| macOS | Intel | `shipyard-darwin-amd64` |
| macOS | Apple Silicon | `shipyard-darwin-arm64` |
| Windows | AMD64 | `shipyard-windows-amd64.exe` |
| Windows | ARM64 | `shipyard-windows-arm64.exe` |

#### 2. Install Binary

**macOS:**
```bash
# MacBook Intel
curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-darwin-amd64 -o shipyard

# MacBook Apple Silicon (M1/M2/M3)
curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-darwin-arm64 -o shipyard

# Make executable and install
chmod +x shipyard
sudo mv shipyard /usr/local/bin/
```

**Linux:**
```bash
# AMD64 (most common)
curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard

# ARM64 (Raspberry Pi, ARM servers)
curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-arm64 -o shipyard

# Make executable and install
chmod +x shipyard

# Install system-wide
sudo mv shipyard /usr/local/bin/

# Or install for current user
mkdir -p ~/.local/bin
mv shipyard ~/.local/bin/
```

**Windows:**
```powershell
# Download to user bin directory
$installDir = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Path $installDir -Force
Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-windows-amd64.exe" -OutFile "$installDir\shipyard.exe"

# Add to PATH (requires restart)
[Environment]::SetEnvironmentVariable("PATH", "$env:PATH;$installDir", "User")
```

### Build from Source

```bash
git clone https://github.com/CodeAlchemyFr/shipyard.git
cd shipyard/cli
go build -o shipyard .
sudo mv shipyard /usr/local/bin/
```

## Verify Installation

```bash
shipyard --help
```

You should see the Shipyard CLI help output.

## Next Steps

- [Quick Start Guide](/getting-started/quick-start) - Deploy your first application
- [Configuration](/getting-started/configuration) - Learn about `paas.yaml` configuration
- [CLI Reference](/cli/overview) - Complete command reference

## Kubernetes Setup

### Automatic Setup (k3s Integration)

If you used the quick install, your Kubernetes cluster is already configured and ready to use:

```bash
# Verify your cluster is running
kubectl cluster-info

# Check cluster nodes
kubectl get nodes
```

### Manual Kubernetes Setup

If you installed Shipyard without k3s, you can connect to any Kubernetes cluster:

#### Local Development Options
- **k3s** - `curl -sfL https://get.k3s.io | sh -` (Linux)
- **k3d** - `k3d cluster create myregistry` (macOS/Windows with Docker)
- **minikube** - `minikube start`
- **kind** - `kind create cluster`
- **Docker Desktop** - Enable Kubernetes in settings

#### Cloud Providers
- **Google GKE** - `gcloud container clusters get-credentials`
- **AWS EKS** - `aws eks update-kubeconfig`
- **Azure AKS** - `az aks get-credentials`

Ensure `kubectl` can connect to your cluster:

```bash
kubectl cluster-info
```

### Monitoring Setup (Recommended)

For full monitoring capabilities, install metrics-server if not already present:

```bash
# Check if metrics-server is already installed
kubectl get pods -n kube-system | grep metrics-server

# If not found, install metrics-server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Wait for deployment
kubectl wait --for=condition=available --timeout=300s deployment/metrics-server -n kube-system

# Verify installation
kubectl top nodes
```

**Note:** Some local clusters (minikube, kind) may need additional configuration for metrics-server to work properly.