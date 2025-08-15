---
layout: home

hero:
  name: "Shipyard"
  text: "Open Source PaaS Platform"
  tagline: Deploy applications to Kubernetes with ease - The simplicity of Heroku, the power of Kubernetes
  image:
    src: /logo.png
    alt: Shipyard Logo
  actions:
    - theme: brand
      text: Get Started
      link: /getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/CodeAlchemyFr/shipyard

features:
  - title: 🚀 Simple Deployment
    details: Deploy with a single command. Just run `shipyard deploy` and your application is live on Kubernetes.
  - title: 🔐 Private Registry Support
    details: Seamless integration with GitHub Container Registry, Docker Hub, and private registries with encrypted credential storage.
  - title: 🌐 Domain Management
    details: Easy domain management with automatic SSL certificates via Let's Encrypt and consolidated ingress configuration.
  - title: 📊 Version Control
    details: Track deployment history with rollback capabilities. Every deployment is versioned and can be restored instantly.
  - title: ⚡ Auto-scaling
    details: Horizontal Pod Autoscaler configuration with CPU-based scaling and resource management out of the box.
  - title: 🔍 Monitoring & Logs
    details: Real-time application logs and status monitoring with Kubernetes-native observability.
---

## Why Shipyard?

Shipyard bridges the gap between the simplicity of traditional PaaS platforms and the power of Kubernetes. Built for developers who want to focus on code, not infrastructure.

### Key Benefits

- **Developer Experience First** - Intuitive CLI with sensible defaults
- **Production Ready** - Built on Kubernetes with enterprise-grade features
- **Open Source** - MIT licensed, extensible, and community-driven
- **Kubernetes Native** - Leverages K8s best practices and ecosystem

### Quick Start

#### 1. Install Shipyard

**macOS & Linux:**
```bash
curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```

**Windows:**
```powershell
Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.ps1" -OutFile "install.ps1"; .\install.ps1
```

#### 2. Deploy Your First App

```bash
# Configure your app
echo 'app:
  name: my-app
  image: ghcr.io/user/my-app:latest
  port: 3000' > paas.yaml

# Add registry credentials (if private image)
shipyard registry add ghcr.io username token

# Deploy to Kubernetes
shipyard deploy

# Monitor deployment
shipyard status
shipyard logs my-app
```

