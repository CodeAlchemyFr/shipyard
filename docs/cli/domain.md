# shipyard domain

Manage custom domains for your applications.

## Synopsis

Add, list, and remove custom domains for your applications. Shipyard automatically manages ingress configuration and SSL certificates.

## Usage

```
shipyard domain [command]
```

## Available Commands

- [`add`](#add) - Add a custom domain to an application
- [`list`](#list) - List all configured domains
- [`remove`](#remove) - Remove a domain from an application

## add

Add a custom domain to your application.

### Usage

```
shipyard domain add [hostname] [flags]
```

### Arguments

- `hostname` - Full domain name (e.g., `app.example.com`)

### Flags

```
      --path string   URL path prefix (default "/")
      --no-ssl        Disable SSL/HTTPS (default: SSL enabled)
  -h, --help          help for add
```

### Examples

```bash
# Add basic domain
shipyard domain add app.example.com

# Add domain with custom path
shipyard domain add api.example.com --path /api/v1

# Add domain without SSL (not recommended)
shipyard domain add internal.company.com --no-ssl
```

## list

Display all configured domains across applications.

### Usage

```
shipyard domain list [flags]
```

### Example Output

```
🌐 Domain Configuration:

APPLICATION    HOSTNAME              PATH    SSL    CREATED
web-app        app.example.com       /       ✓      2024-01-15 14:30
web-app        www.example.com       /       ✓      2024-01-15 14:31
api-service    api.example.com       /api    ✓      2024-01-15 15:00
blog           blog.company.com      /       ✓      2024-01-14 10:00

📊 Summary:
   Total domains: 4
   Base domains: 2 (example.com, company.com)
   SSL enabled: 4/4 (100%)
```

## remove

Remove a domain from an application.

### Usage

```
shipyard domain remove [hostname] [flags]
```

### Arguments

- `hostname` - Domain name to remove

### Example

```bash
shipyard domain remove old-app.example.com
```

## Domain Management

### Automatic Ingress Generation

Shipyard automatically:
1. **Groups domains** by base domain (e.g., `example.com`)
2. **Generates consolidated ingress** files per base domain
3. **Updates ingress** when domains are added/removed
4. **Manages SSL certificates** via cert-manager/Let's Encrypt

### File Structure

```
manifests/
└── shared/
    ├── ingress-example.com.yaml     # All *.example.com domains
    └── ingress-company.com.yaml     # All *.company.com domains
```

### SSL Configuration

By default, all domains use HTTPS with automatic certificates:

```yaml
# Generated ingress includes:
metadata:
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    kubernetes.io/ingress.class: nginx
spec:
  tls:
  - hosts:
    - app.example.com
    - www.example.com
    secretName: example-com-tls
```

## Complete Workflow

### Adding Your First Domain

```bash
# 1. Ensure DNS points to your cluster
# A record: app.example.com → your-cluster-ip

# 2. Add domain to application
shipyard domain add app.example.com

# 3. Deploy to apply ingress changes
shipyard deploy

# 4. Verify domain works
curl https://app.example.com/health
```

### Multiple Domains for One App

```bash
# Add primary domain
shipyard domain add app.example.com

# Add www redirect
shipyard domain add www.example.com

# Add API subdomain with path
shipyard domain add api.example.com --path /api

# Deploy changes
shipyard deploy
```

### Domain Migration

```bash
# Add new domain
shipyard domain add new-app.example.com

# Test new domain works
curl https://new-app.example.com

# Remove old domain
shipyard domain remove old-app.example.com

# Deploy changes
shipyard deploy
```

## Integration with paas.yaml

Domains can also be configured in your `paas.yaml`:

```yaml
app:
  name: web-app
  image: myapp:latest
  port: 3000

domains:
  - hostname: app.example.com
  - hostname: www.example.com  
  - hostname: api.example.com
    path: /api
    ssl_enabled: true
```

CLI commands sync with `paas.yaml` configuration.

## Prerequisites

### DNS Configuration

Ensure DNS records point to your cluster:

```bash
# A record for your domain
app.example.com.  IN  A  192.168.1.100

# Or CNAME to load balancer
app.example.com.  IN  CNAME  my-cluster-lb.cloud.com.
```

### Ingress Controller

Your cluster needs an ingress controller:

```bash
# Install NGINX ingress controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/cloud/deploy.yaml

# Or use cloud provider ingress (GKE, EKS, AKS)
```

### Cert-Manager (for SSL)

For automatic SSL certificates:

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Configure Let's Encrypt issuer
# (Shipyard can generate this configuration)
```

## Troubleshooting

### Domain Not Accessible

```bash
# Check DNS resolution
nslookup app.example.com

# Check ingress configuration
kubectl get ingress

# Check ingress controller
kubectl get pods -n ingress-nginx
```

### SSL Certificate Issues

```bash
# Check certificate status
kubectl get certificates

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager

# Verify Let's Encrypt challenge
kubectl describe certificaterequest
```

### Domain Already Exists

```
Error: hostname app.example.com already exists
```

Each domain can only belong to one application:
```bash
# Check which app owns the domain
shipyard domain list | grep app.example.com

# Remove from other app first
shipyard domain remove app.example.com
```

### Path Conflicts

Multiple apps with same domain but different paths:

```bash
# App 1: api.example.com/v1
shipyard domain add api.example.com --path /v1

# App 2: api.example.com/v2  
shipyard domain add api.example.com --path /v2
```

## Security Considerations

### HTTPS Only

Always use SSL in production:
```bash
# Good
shipyard domain add app.example.com

# Bad (only for development)
shipyard domain add app.example.com --no-ssl
```

### Domain Validation

Shipyard validates domain formats:
- Must be valid hostname
- Cannot include protocols (`https://`)
- Must not conflict with existing domains

### Network Policies

Consider implementing network policies for domain-based access control.

## Best Practices

1. **Use SSL** - Always enable HTTPS for production
2. **Plan DNS** - Configure DNS before adding domains  
3. **Test First** - Verify domain works before removing old ones
4. **Document Domains** - Keep track of domain ownership
5. **Monitor Certificates** - Watch for SSL expiration alerts