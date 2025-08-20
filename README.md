# 🚀 Shipyard - Kubernetes PaaS made simple

Shipyard is a modern Platform-as-a-Service (PaaS) built on top of Kubernetes, designed to simplify application deployment and management.

## ✨ Features

- **🚀 One-command installation** with k3s and SSL certificates
- **📦 Interactive CLI** for registries, domains, and rollback management  
- **🎯 Simple deployment** from Docker images or Git repositories
- **🌐 Automatic ingress** and SSL certificate management with Let's Encrypt
- **📊 Built-in monitoring** and logging
- **🔄 Easy rollbacks** with deployment history
- **⚙️ Flexible service types** (ClusterIP, NodePort) 
- **🎨 Web UI** for paas.yaml generation with real-time preview

## 🚀 Quick Start

### Installation

#### One-line installation (with k3s + SSL):
```bash
curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```
*The installer will prompt for your email to configure SSL certificates.*

#### Install without k3s:
```bash
INSTALL_K3S=false curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```

#### Windows PowerShell:
```powershell
Invoke-WebRequest -Uri "https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.ps1" -OutFile "install.ps1"; .\install.ps1
```

### Your first deployment

1. **Create your app directory:**
```bash
mkdir my-app && cd my-app
```

2. **Create a paas.yaml:**
```yaml
app:
  name: my-app
  image: nginx:latest
  port: 80

service:
  type: ClusterIP

resources:
  cpu: 100m
  memory: 128Mi

scaling:
  min: 1
  max: 3
  target_cpu: 70

domains:
  - my-app.example.com
```

3. **Deploy:**
```bash
shipyard deploy
```

## 📖 Documentation

### Interactive CLI Commands

Shipyard provides interactive modes for easier management:

#### Registry Management
```bash
shipyard registry
```
- Add/remove Docker registries
- Set default registry
- Simplified configuration (URL, username, token only)

#### Domain Management  
```bash
shipyard domain
```
- Add/remove domains
- Automatic SSL certificate generation
- Real-time DNS verification

#### Rollback Management
```bash
shipyard rollback  
```
- Interactive deployment history
- One-click rollback to any successful deployment
- Deployment status tracking

### Service Configuration

Shipyard supports both internal and external service exposure:

#### ClusterIP (Internal only)
```yaml
service:
  type: ClusterIP
```

#### NodePort (External access)
```yaml
service:
  type: NodePort
  externalPort: 30080  # Port 30000-32767
```

### Web Interface

Generate `paas.yaml` files with a user-friendly web interface:

```bash
cd webapp
npm install
npm run dev
```

Features:
- 🎨 **Visual configuration** of all paas.yaml options
- ⚡ **Real-time YAML preview** 
- 🌐 **Service type selection** (ClusterIP vs NodePort)
- 📧 **Environment variable management**
- 🔧 **Resource and scaling configuration**
- 📋 **Copy/download generated YAML**

## 🛠️ CLI Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `shipyard deploy` | Deploy application from paas.yaml |
| `shipyard status` | Show application status |
| `shipyard logs <app>` | View application logs |
| `shipyard delete <app>` | Delete application |

### Interactive Commands

| Command | Description |
|---------|-------------|
| `shipyard registry` | Manage Docker registries interactively |
| `shipyard domain` | Manage domains and SSL certificates |
| `shipyard rollback` | Interactive rollback to previous deployments |

### Configuration

| Command | Description |
|---------|-------------|
| `shipyard init` | Initialize new application |
| `shipyard config` | Show current configuration |

## 📋 paas.yaml Reference

Complete configuration example:

```yaml
app:
  name: my-app
  image: ghcr.io/user/my-app:latest
  port: 3000

service:
  type: NodePort        # ClusterIP or NodePort
  externalPort: 30080   # Required for NodePort

resources:
  cpu: 500m
  memory: 512Mi

scaling:
  min: 2
  max: 10
  target_cpu: 70

env:
  NODE_ENV: production
  API_URL: https://api.example.com

domains:
  - my-app.example.com
  - api.my-app.example.com

health:
  liveness:
    path: /health
  readiness:
    path: /ready
```

## 🔧 Installation Details

### What gets installed:

1. **Shipyard CLI** - Main management tool
2. **k3s Kubernetes** - Lightweight Kubernetes distribution  
3. **Traefik Ingress** - Load balancer and reverse proxy
4. **cert-manager** - Automatic SSL certificate management
5. **Let's Encrypt ClusterIssuer** - Free SSL certificates

### Requirements:

- **Linux/macOS**: Docker (for k3d on macOS)
- **Windows**: Docker Desktop
- **Ports**: 80, 443 (HTTP/HTTPS)
- **Email**: For SSL certificate registration

## 🐛 Troubleshooting

### Common Issues

**SSL certificates not working:**
```bash
# Check certificate status
kubectl get certificate

# Reconfigure Let's Encrypt email
kubectl apply -f - <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: traefik
EOF
```

**App not accessible:**
```bash
# Check application status  
shipyard status

# Check ingress
kubectl get ingress

# Check service
kubectl get svc
```

**Port issues:**
Make sure your `paas.yaml` port matches your application's exposed port, not the service port.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔗 Links

- [Documentation](https://codealchemy.github.io/shipyard/)
- [GitHub Repository](https://github.com/CodeAlchemyFr/shipyard)
- [Issues & Bug Reports](https://github.com/CodeAlchemyFr/shipyard/issues)
- [Releases](https://github.com/CodeAlchemyFr/shipyard/releases)