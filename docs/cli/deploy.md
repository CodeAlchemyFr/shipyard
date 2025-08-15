# shipyard deploy

Deploy an application to Kubernetes using the configuration in `paas.yaml`.

## Synopsis

The deploy command reads your `paas.yaml` configuration file, generates Kubernetes manifests, and applies them to your cluster. It automatically handles versioning, registry authentication, and deployment tracking.

## Usage

```
shipyard deploy [flags]
```

## Flags

```
  -h, --help   help for deploy
```

## Configuration

The deploy command requires a `paas.yaml` file in the current directory. See [Configuration Reference](/getting-started/configuration) for complete details.

## What Deploy Does

1. **Validates** `paas.yaml` configuration
2. **Generates** version identifier for deployment tracking
3. **Creates** registry secrets (if using private images)
4. **Generates** Kubernetes manifests:
   - `manifests/apps/{app-name}/deployment.yaml`
   - `manifests/apps/{app-name}/secrets.yaml` 
   - `manifests/apps/{app-name}/service.yaml`
   - `manifests/apps/{app-name}/registry-secret.yaml` (if needed)
5. **Updates** shared ingress configuration (if domains configured)
6. **Applies** manifests to Kubernetes cluster
7. **Tracks** deployment in local database
8. **Reports** deployment status

## Examples

### Basic Deployment

```bash
# Create configuration
cat > paas.yaml << EOF
app:
  name: web-app
  image: nginx:latest
  port: 80
EOF

# Deploy
shipyard deploy
```

Expected output:
```
🚀 Starting deployment for web-app...
📁 Created directory: manifests/apps/web-app
📄 Generated: manifests/apps/web-app/deployment.yaml (version: v1703123456)
📄 Generated: manifests/apps/web-app/secrets.yaml
📄 Generated: manifests/apps/web-app/service.yaml
☸️  Applying manifests to Kubernetes cluster...
✅ Deployment successful!
   Version: v1703123456
   Image: nginx:latest
   Status: 1/1 pods running
```

### Deployment with Private Registry

```bash
# Add registry credentials first
shipyard registry add ghcr.io username token

# Create configuration with private image
cat > paas.yaml << EOF
app:
  name: private-app
  image: ghcr.io/user/private-repo:v1.0.0
  port: 3000
env:
  NODE_ENV: production
secrets:
  DATABASE_URL: postgresql://user:pass@host:5432/db
EOF

# Deploy
shipyard deploy
```

Output includes registry secret:
```
🚀 Starting deployment for private-app...
📁 Created directory: manifests/apps/private-app
🔐 Generated: manifests/apps/private-app/registry-secret.yaml (registry: ghcr.io)
📄 Generated: manifests/apps/private-app/deployment.yaml (version: v1703123500)
📄 Generated: manifests/apps/private-app/secrets.yaml
📄 Generated: manifests/apps/private-app/service.yaml
☸️  Applying manifests to Kubernetes cluster...
✅ Deployment successful!
```

### Deployment with Custom Domain

```bash
cat > paas.yaml << EOF
app:
  name: web-service
  image: myapp:latest
  port: 8080
domains:
  - hostname: api.example.com
  - hostname: app.example.com
EOF

shipyard deploy
```

This generates ingress configuration:
```
📄 Generated: manifests/shared/ingress-example.com.yaml
🌐 Updated ingress for domain: example.com
```

## Generated Manifests

### Deployment Manifest

Contains:
- Pod specification with your container image
- Resource limits and requests
- Environment variables and secrets
- Liveness and readiness probes
- Horizontal Pod Autoscaler (if scaling configured)
- Registry pull secrets (if private image)

### Service Manifest

Creates a ClusterIP service exposing your application internally.

### Secrets Manifest

Base64-encoded secrets from your `paas.yaml` configuration.

### Registry Secret Manifest

Docker registry authentication (generated only for private images).

## Version Tracking

Each deployment creates a version entry with:
- **Version ID** - Timestamp-based identifier (e.g., `v1703123456`)
- **Image** - Full container image with tag
- **Configuration** - Hash of your `paas.yaml` 
- **Status** - pending, success, or failed
- **Timestamp** - When deployment started

View deployment history:
```bash
shipyard releases
```

## Error Handling

### Configuration Errors

```bash
Error: failed to load paas.yaml: file not found
```
Solution: Create `paas.yaml` in current directory.

```bash
Error: invalid configuration: app.name is required
```
Solution: Add required fields to `paas.yaml`.

### Registry Errors

```bash
Error: failed to generate registry secrets: registry ghcr.io not found
```
Solution: Add registry credentials first:
```bash
shipyard registry add ghcr.io username token
```

### Kubernetes Errors

```bash
Error: failed to apply manifests: connection refused
```
Solution: Check Kubernetes cluster connection:
```bash
kubectl cluster-info
```

```bash
Error: failed to create deployment: insufficient resources
```
Solution: Reduce resource requests or scale down other applications.

## Deployment States

| State | Description |
|-------|-------------|
| `pending` | Deployment started but not yet completed |
| `success` | Deployment completed successfully |
| `failed` | Deployment failed (check logs) |

Failed deployments can be rolled back:
```bash
shipyard rollback
```

## Best Practices

### Image Tags
Use specific tags rather than `latest`:
```yaml
app:
  image: myapp:v1.2.3  # Good
  # image: myapp:latest  # Avoid
```

### Resource Limits
Always specify resource limits:
```yaml
resources:
  cpu: "100m"
  memory: "128Mi"
```

### Health Checks
Configure health check endpoints:
```yaml
app:
  port: 3000
  health_path: /health    # For liveness probe
  ready_path: /ready      # For readiness probe
```

### Secrets Management
Store sensitive values in secrets, not environment variables:
```yaml
env:
  NODE_ENV: production
secrets:
  DATABASE_PASSWORD: secret123
  API_KEY: abc123
```

## Integration with CI/CD

### GitHub Actions

```yaml
name: Deploy to Kubernetes
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
        curl -L https://github.com/shipyard-run/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard
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
deploy:
  image: alpine/kubectl
  before_script:
    - curl -L https://github.com/shipyard-run/shipyard/releases/latest/download/shipyard-linux-amd64 -o shipyard
    - chmod +x shipyard
    - ./shipyard registry add $CI_REGISTRY $CI_REGISTRY_USER $CI_REGISTRY_PASSWORD
  script:
    - ./shipyard deploy
```