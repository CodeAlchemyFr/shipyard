# Getting Started

## Installation

### Prerequisites

Before installing Shipyard, ensure you have:

- **Kubernetes cluster** - Access to a Kubernetes cluster (local or cloud)
- **kubectl** - Kubernetes command-line tool configured

### Quick Install (Recommended)

#### macOS & Linux

```bash
curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```

The script automatically detects:
- **Operating System** (macOS/Linux)
- **Architecture** (Intel x86_64, Apple Silicon arm64, Linux arm64)
- **Downloads** the correct binary for your system
- **Installs** to `/usr/local/bin/shipyard`

#### Windows (PowerShell)

```powershell
Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.ps1" -OutFile "install.ps1"; .\install.ps1
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