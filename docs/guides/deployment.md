# Deploying Applications

Complete guide to deploying applications with Shipyard - from simple static sites to complex microservices.

## Overview

Shipyard simplifies Kubernetes deployments by:
- Generating standard Kubernetes manifests from simple configuration
- Managing deployment versions and rollback capabilities
- Handling registry authentication automatically
- Providing deployment tracking and monitoring

## Basic Deployment Workflow

### 1. Create Configuration

Create a `paas.yaml` file in your project directory:

```yaml
app:
  name: my-app
  image: nginx:latest
  port: 80

env:
  ENVIRONMENT: production

resources:
  cpu: "100m"
  memory: "128Mi"
```

### 2. Deploy Application

```bash
shipyard deploy
```

Expected output:
```
🚀 Starting deployment for my-app...
📁 Created directory: manifests/apps/my-app
📄 Generated: manifests/apps/my-app/deployment.yaml (version: v1703123456)
📄 Generated: manifests/apps/my-app/secrets.yaml
📄 Generated: manifests/apps/my-app/service.yaml
☸️  Applying manifests to Kubernetes cluster...
✅ Deployment successful!
   Version: v1703123456
   Image: nginx:latest
   Status: 1/1 pods running
```

### 3. Verify Deployment

```bash
# Check application status
shipyard status

# View application logs
shipyard logs my-app

# Check deployment history
shipyard releases
```

## Application Types

### Static Website

```yaml
app:
  name: website
  image: nginx:alpine
  port: 80

env:
  NODE_ENV: production

resources:
  cpu: "50m"
  memory: "64Mi"

scaling:
  min: 2
  max: 5
  target_cpu: 70
```

### Node.js Application

```yaml
app:
  name: node-app
  image: node:18-alpine
  port: 3000

env:
  NODE_ENV: production
  PORT: "3000"

secrets:
  DATABASE_URL: "postgresql://user:pass@host:5432/db"
  JWT_SECRET: "your-secret-key"

resources:
  cpu: "200m"
  memory: "256Mi"

scaling:
  min: 2
  max: 10
  target_cpu: 70
```

### Python API

```yaml
app:
  name: python-api
  image: python:3.11-slim
  port: 8000

env:
  PYTHONPATH: "/app"
  DJANGO_SETTINGS_MODULE: "settings.production"

secrets:
  SECRET_KEY: "django-secret-key"
  DATABASE_URL: "postgres://user:pass@host:5432/db"

resources:
  cpu: "300m"
  memory: "512Mi"

scaling:
  min: 3
  max: 15
  target_cpu: 65
```

### Go Microservice

```yaml
app:
  name: go-service
  image: golang:1.21-alpine
  port: 8080

env:
  GO_ENV: production
  PORT: "8080"

secrets:
  API_KEY: "service-api-key"
  REDIS_URL: "redis://redis:6379"

resources:
  cpu: "100m"
  memory: "128Mi"

scaling:
  min: 2
  max: 8
  target_cpu: 80
```

## Private Container Images

### GitHub Container Registry

```bash
# Add registry credentials
shipyard registry add ghcr.io username ghp_token123

# Deploy with private image
echo 'app:
  name: private-app
  image: ghcr.io/user/private-repo:v1.0.0
  port: 3000' > paas.yaml

shipyard deploy
```

### Docker Hub Private

```bash
# Add Docker Hub credentials
shipyard registry add docker.io username password

# Deploy private image
echo 'app:
  name: docker-app
  image: user/private-repo:latest
  port: 8080' > paas.yaml

shipyard deploy
```

### AWS ECR

```bash
# Get AWS ECR token
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-west-2.amazonaws.com

# Add to Shipyard
TOKEN=$(aws ecr get-login-password --region us-west-2)
shipyard registry add 123456789.dkr.ecr.us-west-2.amazonaws.com AWS $TOKEN

# Deploy ECR image
echo 'app:
  name: aws-app
  image: 123456789.dkr.ecr.us-west-2.amazonaws.com/my-app:latest
  port: 3000' > paas.yaml

shipyard deploy
```

## Environment-Specific Deployments

### Development

```yaml
# paas.dev.yaml
app:
  name: myapp-dev
  image: myapp:develop
  port: 3000

env:
  NODE_ENV: development
  DEBUG: "true"

resources:
  cpu: "100m"
  memory: "128Mi"

scaling:
  min: 1
  max: 2
```

```bash
cp paas.dev.yaml paas.yaml
shipyard deploy
```

### Staging

```yaml
# paas.staging.yaml
app:
  name: myapp-staging
  image: myapp:staging
  port: 3000

env:
  NODE_ENV: staging

secrets:
  DATABASE_URL: "postgres://staging-db"

resources:
  cpu: "200m"
  memory: "256Mi"

scaling:
  min: 1
  max: 3

domains:
  - hostname: staging.myapp.com
```

### Production

```yaml
# paas.prod.yaml
app:
  name: myapp
  image: myapp:v1.2.0
  port: 3000

env:
  NODE_ENV: production

secrets:
  DATABASE_URL: "postgres://prod-db"
  REDIS_URL: "redis://prod-redis"

resources:
  cpu: "500m"
  memory: "512Mi"

scaling:
  min: 3
  max: 20
  target_cpu: 65

domains:
  - hostname: myapp.com
  - hostname: www.myapp.com
```

## Blue-Green Deployments

### Preparation

```bash
# Deploy current version (blue)
echo 'app:
  name: myapp-blue
  image: myapp:v1.0.0
  port: 3000
domains:
  - hostname: app.example.com' > paas.yaml

shipyard deploy
```

### Green Deployment

```bash
# Deploy new version (green)
echo 'app:
  name: myapp-green  
  image: myapp:v2.0.0
  port: 3000
domains:
  - hostname: green.example.com' > paas.yaml

shipyard deploy

# Test green environment
curl https://green.example.com/health
```

### Traffic Switch

```bash
# Switch traffic to green
shipyard domain remove app.example.com  # From blue
shipyard domain add app.example.com     # To green (current app)
shipyard deploy

# Remove blue environment
kubectl delete deployment myapp-blue
```

## Canary Deployments

### Deploy Canary

```bash
# Deploy canary version (10% traffic)
echo 'app:
  name: myapp-canary
  image: myapp:v2.0.0
  port: 3000
scaling:
  min: 1
  max: 2' > paas.yaml

shipyard deploy
```

### Traffic Splitting

Use ingress annotations for traffic splitting:

```yaml
# Edit generated ingress
annotations:
  nginx.ingress.kubernetes.io/canary: "true"
  nginx.ingress.kubernetes.io/canary-weight: "10"
```

### Promote or Rollback

```bash
# If canary successful, promote
shipyard domain remove app.example.com  # From main
# Update main app with canary image
echo 'app:
  name: myapp
  image: myapp:v2.0.0' > paas.yaml
shipyard deploy

# If canary fails, rollback
kubectl delete deployment myapp-canary
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Install Shipyard
      run: |
        curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard
        chmod +x shipyard
        sudo mv shipyard /usr/local/bin/
    
    - name: Configure registry
      run: shipyard registry add ghcr.io ${{ github.actor }} ${{ secrets.GITHUB_TOKEN }}
    
    - name: Deploy
      run: shipyard deploy
      env:
        KUBECONFIG: ${{ secrets.KUBECONFIG }}
```

### GitLab CI

```yaml
stages:
  - deploy

deploy:production:
  stage: deploy
  image: alpine/kubectl:latest
  before_script:
    - curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard
    - chmod +x shipyard
    - ./shipyard registry add $CI_REGISTRY $CI_REGISTRY_USER $CI_REGISTRY_PASSWORD
  script:
    - ./shipyard deploy
  only:
    - main
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    
    stages {
        stage('Deploy') {
            steps {
                script {
                    sh '''
                        curl -L https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard
                        chmod +x shipyard
                        ./shipyard registry add registry.company.com $REGISTRY_USER $REGISTRY_TOKEN
                        ./shipyard deploy
                    '''
                }
            }
        }
    }
}
```

## Monitoring Deployments

### Health Checks

Shipyard automatically adds health checks to your deployments:

```yaml
# Generated in deployment.yaml
livenessProbe:
  httpGet:
    path: /health
    port: 3000
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /ready  
    port: 3000
  initialDelaySeconds: 5
  periodSeconds: 5
```

Ensure your application responds to these endpoints:
- `/health` - Application is running
- `/ready` - Application is ready to serve traffic

### Custom Health Endpoints

```yaml
# In your paas.yaml (future feature)
app:
  name: myapp
  image: myapp:latest
  port: 3000
  health_path: /api/health
  ready_path: /api/ready
```

### Monitoring Commands

```bash
# Real-time status
shipyard status

# Follow logs
shipyard logs myapp --follow

# Check deployment history
shipyard releases

# Monitor resource usage
kubectl top pods -l app=myapp
```

## Troubleshooting Deployments

### Failed Deployment

```bash
# Check deployment status
shipyard status

# View detailed logs
shipyard logs myapp --tail 100

# Check Kubernetes events
kubectl get events --sort-by=.metadata.creationTimestamp

# Describe problematic pods
kubectl describe pod myapp-xxx
```

### Common Issues

#### Image Pull Errors
```
Error: ErrImagePull, ImagePullBackOff
```

Solutions:
- Add registry credentials: `shipyard registry add`
- Verify image exists and tag is correct
- Check registry permissions

#### Resource Constraints
```
Error: Insufficient cpu, Insufficient memory
```

Solutions:
- Reduce resource requests in `paas.yaml`
- Scale up cluster nodes
- Check resource quotas

#### Configuration Errors
```
Error: CreateContainerConfigError
```

Solutions:
- Check secret values in `paas.yaml`
- Verify environment variable formats
- Validate configuration syntax

#### Networking Issues
```
Error: Service unavailable, Connection refused
```

Solutions:
- Verify port configuration matches application
- Check service generation in manifests
- Test connectivity between pods

## Best Practices

### Configuration Management

1. **Use Specific Tags** - Avoid `latest` in production
2. **Environment Separation** - Different configs per environment  
3. **Secret Management** - Store sensitive data in `secrets` section
4. **Resource Planning** - Set appropriate CPU/memory limits
5. **Health Checks** - Implement `/health` and `/ready` endpoints

### Deployment Strategy

1. **Gradual Rollout** - Use staging before production
2. **Monitoring** - Watch logs during and after deployment
3. **Rollback Plan** - Know how to quickly revert changes
4. **Testing** - Validate deployments in staging environment
5. **Documentation** - Document deployment procedures

### Security

1. **Private Images** - Use private registries for proprietary code
2. **Least Privilege** - Minimal container permissions
3. **Network Policies** - Restrict pod-to-pod communication
4. **Secret Rotation** - Regularly update secrets and tokens
5. **Image Scanning** - Check for vulnerabilities before deployment

### Performance

1. **Resource Optimization** - Right-size CPU and memory
2. **Auto-scaling** - Configure HPA for variable loads
3. **Caching** - Implement appropriate caching strategies
4. **Load Testing** - Validate performance under load
5. **Monitoring** - Track metrics and set up alerting