# 🎨 Shipyard Web Interface

A modern web interface for generating Shipyard `paas.yaml` configuration files with real-time preview and intuitive controls.

## ✨ Features

- **🎯 Visual Configuration**: Configure all paas.yaml options through an intuitive web interface
- **⚡ Real-time Preview**: See your YAML file update automatically as you make changes
- **🌐 Service Configuration**: Choose between ClusterIP (internal) or NodePort (external) service types
- **📧 Environment Management**: Add environment variables manually or upload from .env files
- **🔧 Resource & Scaling**: Configure CPU, memory, and autoscaling settings
- **🔒 Health Checks**: Set up liveness and readiness probes
- **🌍 Domain Management**: Add multiple domains with automatic SSL certificate generation
- **📋 Export Options**: Copy to clipboard or download as paas.yaml file

## 🚀 Quick Start

### Development

```bash
# Install dependencies
npm install

# Start development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to access the interface.

### Production Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

## 🛠️ Interface Guide

### Application Configuration

1. **App Name**: Set your application name
2. **Docker Image**: 
   - Select from popular registries (GitHub, Docker Hub, GitLab, etc.)
   - Or enter a custom registry URL
   - Image field auto-updates based on registry selection
3. **Port**: Set the port your application exposes

### Service Configuration

Choose how your application should be accessible:

- **🔒 ClusterIP**: Internal access only (within Kubernetes cluster)
- **🌐 NodePort**: External access via specific port (30000-32767)

### Resource Management

- **CPU**: Set CPU limits (100m, 500m, 1, 2, etc.)
- **Memory**: Set memory limits (128Mi, 512Mi, 1Gi, etc.)
- **Scaling**: Configure min/max replicas and CPU target

### Environment Variables

- **Manual Entry**: Add key-value pairs directly
- **File Upload**: Upload `.env` files for bulk import
- **Dynamic Management**: Add/remove variables as needed

### Domain & SSL

- **Domain Management**: Add multiple domains
- **Automatic SSL**: SSL certificates generated automatically via Let's Encrypt
- **DNS Configuration**: Each domain gets proper ingress routing

### Health Checks

- **Liveness Probe**: Endpoint to check if app is running
- **Readiness Probe**: Endpoint to check if app is ready to receive traffic

## 📋 Generated Configuration

The interface generates a complete `paas.yaml` file:

```yaml
app:
  name: my-app
  image: ghcr.io/user/my-app:latest
  port: 3000

service:
  type: ClusterIP  # or NodePort with externalPort

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

health:
  liveness:
    path: /health
  readiness:
    path: /ready
```

## 🎯 Usage Tips

1. **Start with basics**: Configure app name, image, and port first
2. **Service type**: Choose ClusterIP for internal services, NodePort for external access
3. **Resources**: Start with conservative values (100m CPU, 128Mi memory)
4. **Scaling**: Set reasonable min/max based on expected load
5. **Domains**: Add domains after configuring basic app settings
6. **Environment**: Use .env upload for bulk environment variables

## 🔧 Technical Details

### Built With

- **Next.js 14**: React framework with App Router
- **TypeScript**: Type-safe development
- **Tailwind CSS**: Utility-first CSS framework  
- **shadcn/ui**: Modern component library
- **Lucide Icons**: Beautiful icon set

### Architecture

- **Real-time Generation**: Uses React hooks (`useEffect`) to update YAML on every change
- **Template-based**: YAML generation using JavaScript template literals
- **Responsive Design**: Works on desktop and mobile devices
- **Type Safety**: Full TypeScript coverage for configuration objects

## 🚀 Deployment

### Using Shipyard

Create a `paas.yaml` for the web interface itself:

```yaml
app:
  name: shipyard-webapp
  image: node:18-alpine
  port: 3000

service:
  type: ClusterIP

resources:
  cpu: 100m
  memory: 256Mi

domains:
  - shipyard-ui.example.com
```

### Docker Deployment

```bash
# Build Docker image
docker build -t shipyard-webapp .

# Run container
docker run -p 3000:3000 shipyard-webapp
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test the interface thoroughly
5. Submit a pull request

## 📄 License

Part of the Shipyard project - MIT License
