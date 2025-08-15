# Configuration Reference

The `paas.yaml` file is the main configuration file for your Shipyard applications. This file defines your application settings, resource requirements, scaling behavior, and more.

## Basic Structure

```yaml
app:
  name: string              # Required: Application name
  image: string             # Required: Container image
  port: number              # Required: Container port

env:                        # Optional: Environment variables
  KEY: "value"

secrets:                    # Optional: Secret environment variables  
  SECRET_KEY: "value"

resources:                  # Optional: Resource limits
  cpu: string
  memory: string

scaling:                    # Optional: Auto-scaling configuration
  min: number
  max: number
  target_cpu: number

domains:                    # Optional: Custom domains
  - hostname: string
    path: string            # Optional
    ssl_enabled: boolean    # Optional
```

## Application Settings

### app (Required)

```yaml
app:
  name: my-application      # Kubernetes deployment name
  image: nginx:1.21         # Container image with tag
  port: 80                  # Port your application listens on
```

**Fields:**

- `name` (string, required) - Application name used for Kubernetes resources
- `image` (string, required) - Full container image reference  
- `port` (number, required) - Port your container exposes

**Image formats:**
- Public images: `nginx:latest`, `node:18-alpine`
- Docker Hub: `username/repository:tag`  
- GitHub Container Registry: `ghcr.io/user/repo:tag`
- Private registries: `my-registry.com/image:tag`

## Environment Variables

### env (Optional)

Public environment variables passed to your container:

```yaml
env:
  NODE_ENV: "production"
  API_ENDPOINT: "https://api.example.com"
  DEBUG: "false"
  PORT: "3000"
```

### secrets (Optional)

Sensitive environment variables stored as Kubernetes secrets:

```yaml
secrets:
  DATABASE_URL: "postgresql://user:password@host:5432/db"
  JWT_SECRET: "your-secret-key"
  API_KEY: "abc123xyz"
```

::: warning Security
Values in `secrets` are base64-encoded and stored as Kubernetes secrets. Do not commit sensitive values to version control. Use environment variable substitution in CI/CD instead.
:::

## Resource Management

### resources (Optional)

Define CPU and memory limits for your containers:

```yaml
resources:
  cpu: "100m"          # 0.1 CPU cores
  memory: "128Mi"      # 128 megabytes RAM
```

**CPU formats:**
- `100m` = 0.1 cores  
- `500m` = 0.5 cores
- `1000m` = `1` = 1 core
- `2` = 2 cores

**Memory formats:**
- `128Mi` = 128 mebibytes
- `256Mi` = 256 mebibytes  
- `1Gi` = 1 gibibyte
- `512M` = 512 megabytes

## Auto-scaling

### scaling (Optional)

Configure Horizontal Pod Autoscaler:

```yaml
scaling:
  min: 1               # Minimum replicas
  max: 10              # Maximum replicas  
  target_cpu: 70       # Target CPU utilization %
```

**Fields:**
- `min` (number) - Minimum number of pod replicas (default: 1)
- `max` (number) - Maximum number of pod replicas (default: min)
- `target_cpu` (number) - CPU percentage to trigger scaling (default: 80)

**Auto-scaling behavior:**
- Scale up when CPU > `target_cpu` for 3 minutes
- Scale down when CPU < `target_cpu` for 5 minutes
- Only creates HPA if `max > min`

## Domain Configuration

### domains (Optional)

Configure custom domains for your application:

```yaml
domains:
  - hostname: app.example.com
  - hostname: api.example.com
    path: /api
    ssl_enabled: true
```

**Fields:**
- `hostname` (string, required) - Full domain name
- `path` (string, optional) - URL path prefix (default: "/")
- `ssl_enabled` (boolean, optional) - Enable HTTPS (default: true)

**Domain features:**
- Automatic SSL certificates via Let's Encrypt
- Consolidated ingress per base domain
- Path-based routing support

## Complete Example

```yaml
# Production web application
app:
  name: web-app
  image: ghcr.io/company/web-app:v2.1.0
  port: 3000

# Public configuration
env:
  NODE_ENV: "production"
  API_URL: "https://api.company.com"
  REDIS_URL: "redis://redis-service:6379"
  LOG_LEVEL: "info"

# Sensitive configuration  
secrets:
  DATABASE_URL: "postgresql://user:pass@postgres:5432/app"
  JWT_SECRET: "super-secret-jwt-key"
  STRIPE_SECRET_KEY: "sk_live_..."

# Resource allocation
resources:
  cpu: "500m"
  memory: "512Mi"

# Auto-scaling configuration
scaling:
  min: 2
  max: 20
  target_cpu: 75

# Custom domains
domains:
  - hostname: app.company.com
  - hostname: www.company.com  
  - hostname: api.company.com
    path: /api/v1
```

## Environment Variable Substitution

You can use environment variables in your `paas.yaml`:

```yaml
app:
  name: my-app
  image: ghcr.io/user/app:${IMAGE_TAG}
  port: 3000

secrets:
  DATABASE_URL: ${DATABASE_URL}
  API_KEY: ${API_KEY}
```

Set variables in your shell or CI/CD:
```bash
export IMAGE_TAG=v1.2.3
export DATABASE_URL=postgresql://...
shipyard deploy
```

## Validation

Shipyard validates your configuration on deployment:

- **Required fields** - `app.name`, `app.image`, `app.port`
- **Valid resources** - CPU/memory in correct format
- **Valid domains** - Proper hostname format
- **Image accessibility** - Registry credentials if needed

## Migration from Other Platforms

### From Heroku

```yaml
# Heroku Procfile: web: node server.js
# Heroku Config Vars: NODE_ENV=production

app:
  name: heroku-app
  image: node:18
  port: 3000

env:
  NODE_ENV: "production"
```

### From Docker Compose

```yaml
# docker-compose.yml:
# services:
#   web:
#     image: nginx
#     ports: ["80:80"]
#     environment:
#       - ENV=prod

app:
  name: compose-app  
  image: nginx
  port: 80

env:
  ENV: "prod"
```

## Best Practices

1. **Use specific image tags** - Avoid `latest` in production
2. **Set resource limits** - Prevent resource starvation  
3. **Use secrets for sensitive data** - Keep passwords out of env
4. **Configure health checks** - Ensure application reliability
5. **Version your configuration** - Track changes in git
6. **Test scaling limits** - Verify your scaling configuration

## Troubleshooting

### Invalid configuration
```
Error: validation failed: app.name is required
```
Add missing required fields.

### Resource format errors  
```
Error: invalid CPU format: "100"
```
Use proper formats: `100m`, `1`, `2`.

### Domain conflicts
```
Error: hostname app.example.com already exists
```
Each hostname can only belong to one application.