# Getting Started

## Installation

### Prerequisites

Before installing Shipyard, ensure you have:

- **Kubernetes cluster** - Access to a Kubernetes cluster (local or cloud)
- **kubectl** - Kubernetes command-line tool configured
- **Go 1.21+** - For building from source (optional)

### Install from Binary

Download the latest release from GitHub:

```bash
# Linux/macOS
curl -L https://github.com/shipyard-run/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard
chmod +x shipyard
sudo mv shipyard /usr/local/bin/

# Or for macOS with Homebrew (coming soon)
# brew install shipyard-run/tap/shipyard
```

### Build from Source

```bash
git clone https://github.com/shipyard-run/shipyard.git
cd shipyard/cli
go build -o shipyard .
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

Shipyard works with any Kubernetes cluster:

### Local Development
- **minikube** - `minikube start`
- **kind** - `kind create cluster`
- **Docker Desktop** - Enable Kubernetes in settings

### Cloud Providers
- **Google GKE** - `gcloud container clusters get-credentials`
- **AWS EKS** - `aws eks update-kubeconfig`
- **Azure AKS** - `az aks get-credentials`

Ensure `kubectl` can connect to your cluster:

```bash
kubectl cluster-info
```