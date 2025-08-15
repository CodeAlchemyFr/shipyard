# Managing Domains

Complete guide to setting up and managing custom domains for your Shipyard applications.

## Overview

Shipyard simplifies domain management by automatically handling:
- Ingress configuration generation
- SSL certificate provisioning via Let's Encrypt
- Domain-based routing and path management
- Consolidated ingress files per base domain

## Quick Start

```bash
# Add your first domain
shipyard domain add app.example.com

# Deploy to apply changes
shipyard deploy

# Verify domain works
curl https://app.example.com
```

## DNS Configuration

Before adding domains to Shipyard, configure DNS records to point to your Kubernetes cluster.

### A Records (Recommended)

Point directly to your cluster's external IP:

```
app.example.com.     IN  A  192.168.1.100
www.example.com.     IN  A  192.168.1.100
api.example.com.     IN  A  192.168.1.100
```

### CNAME Records

Point to your load balancer's DNS name:

```
app.example.com.     IN  CNAME  my-cluster-lb.us-west-2.elb.amazonaws.com.
www.example.com.     IN  CNAME  my-cluster-lb.us-west-2.elb.amazonaws.com.
```

### Wildcard Records

For multiple subdomains:

```
*.example.com.       IN  A  192.168.1.100
```

## Adding Domains

### Basic Domain

```bash
shipyard domain add app.example.com
```

This creates a domain with:
- **Path**: `/` (root)
- **SSL**: Enabled (HTTPS)
- **Auto-generated ingress** for `example.com`

### Domain with Custom Path

```bash
shipyard domain add api.example.com --path /api/v1
```

Useful for API services or microservices routing.

### Domain without SSL

```bash
shipyard domain add internal.company.com --no-ssl
```

⚠️ **Not recommended for production**. Use only for internal/development environments.

## Domain Configuration in paas.yaml

Add domains directly to your configuration file:

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

Changes sync between CLI commands and `paas.yaml`.

## Multiple Applications with Shared Domains

### Microservices Architecture

```bash
# Frontend app
cd frontend/
echo 'app:
  name: frontend
  image: frontend:latest
  port: 3000
domains:
  - hostname: app.example.com' > paas.yaml

# API service  
cd ../api/
echo 'app:
  name: api
  image: api:latest
  port: 8080
domains:
  - hostname: app.example.com
    path: /api' > paas.yaml

# Deploy both
shipyard deploy  # In each directory
```

This creates a consolidated ingress routing:
- `app.example.com/` → frontend:3000
- `app.example.com/api` → api:8080

### Path-Based Routing

```bash
# Main website
shipyard domain add company.com

# Blog service
shipyard domain add company.com --path /blog

# Documentation
shipyard domain add company.com --path /docs

# API
shipyard domain add company.com --path /api
```

## SSL and HTTPS

### Automatic SSL Certificates

Shipyard automatically provisions SSL certificates using cert-manager and Let's Encrypt:

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

### Prerequisites

Install cert-manager in your cluster:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml
```

Create a ClusterIssuer for Let's Encrypt:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### SSL Verification

```bash
# Check certificate status
kubectl get certificates

# Verify HTTPS works
curl -I https://app.example.com

# Check SSL details
openssl s_client -connect app.example.com:443 -servername app.example.com
```

## Generated Ingress Structure

Shipyard creates consolidated ingress files per base domain:

```
manifests/
└── shared/
    ├── ingress-example.com.yaml
    └── ingress-company.com.yaml
```

Example `ingress-example.com.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: example-com
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
    kubernetes.io/ingress.class: nginx
spec:
  tls:
  - hosts:
    - app.example.com
    - www.example.com
    secretName: example-com-tls
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend
            port:
              number: 3000
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: api
            port:
              number: 8080
  - host: www.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend
            port:
              number: 3000
```

## Domain Management Commands

### List All Domains

```bash
shipyard domain list
```

Output:
```
🌐 Domain Configuration:

APPLICATION    HOSTNAME              PATH    SSL    CREATED
frontend       app.example.com       /       ✓      2024-01-15 14:30
frontend       www.example.com       /       ✓      2024-01-15 14:31
api            app.example.com       /api    ✓      2024-01-15 15:00
blog           blog.company.com      /       ✓      2024-01-14 10:00
```

### Remove Domain

```bash
shipyard domain remove old-app.example.com
```

### Update Domain

To change domain configuration:
1. Remove existing domain
2. Add domain with new configuration
3. Deploy changes

```bash
shipyard domain remove app.example.com
shipyard domain add app.example.com --path /v2
shipyard deploy
```

## Common Patterns

### WWW Redirect

```bash
# Add both domains to same app
shipyard domain add example.com
shipyard domain add www.example.com

# Or use DNS CNAME
www.example.com.  IN  CNAME  example.com.
```

### Staging and Production

```bash
# Production
shipyard domain add app.example.com

# Staging  
shipyard domain add staging.example.com

# Development
shipyard domain add dev.example.com
```

### API Versioning

```bash
# v1 API
shipyard domain add api.example.com --path /v1

# v2 API  
shipyard domain add api.example.com --path /v2

# Legacy support
shipyard domain add api.example.com --path /legacy
```

## Cloud Provider Integration

### AWS (EKS)

```bash
# Use AWS Load Balancer Controller
kubectl apply -k "github.com/aws/eks-charts/stable/aws-load-balancer-controller//crds?ref=master"

# Add annotation for ALB
# Shipyard can be configured to use ALB ingress class
```

### Google Cloud (GKE)

```bash
# Use GKE ingress controller
# Domains automatically get Google-managed certificates

# Add annotation for Google ingress
kubectl annotate ingress example-com kubernetes.io/ingress.class=gce
```

### Azure (AKS)

```bash
# Use Application Gateway ingress controller
# Configure for Azure DNS integration
```

## Troubleshooting

### Domain Not Accessible

1. **Check DNS propagation**:
   ```bash
   nslookup app.example.com
   dig app.example.com
   ```

2. **Verify ingress created**:
   ```bash
   kubectl get ingress
   kubectl describe ingress example-com
   ```

3. **Check ingress controller**:
   ```bash
   kubectl get pods -n ingress-nginx
   kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
   ```

### SSL Certificate Issues

1. **Check certificate status**:
   ```bash
   kubectl get certificates
   kubectl describe certificate example-com-tls
   ```

2. **Check cert-manager logs**:
   ```bash
   kubectl logs -n cert-manager deployment/cert-manager
   ```

3. **Verify Let's Encrypt challenge**:
   ```bash
   kubectl get challenges
   kubectl describe challenge
   ```

### 502/503 Errors

1. **Check service exists**:
   ```bash
   kubectl get services
   kubectl describe service frontend
   ```

2. **Verify pod health**:
   ```bash
   kubectl get pods
   kubectl logs pod/frontend-xxx
   ```

3. **Check service endpoints**:
   ```bash
   kubectl get endpoints
   ```

### Path Conflicts

```
Error: Multiple apps cannot use same domain with overlapping paths
```

Ensure paths don't conflict:
- ✅ `/api` and `/docs` (different paths)
- ❌ `/api` and `/api/v1` (overlapping)

## Security Considerations

### HTTPS Only

Force HTTPS redirects in ingress:

```yaml
annotations:
  nginx.ingress.kubernetes.io/ssl-redirect: "true"
  nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
```

### Rate Limiting

```yaml
annotations:
  nginx.ingress.kubernetes.io/rate-limit: "100"
  nginx.ingress.kubernetes.io/rate-limit-window: "1m"
```

### IP Whitelisting

```yaml
annotations:
  nginx.ingress.kubernetes.io/whitelist-source-range: "10.0.0.0/8,192.168.0.0/16"
```

## Best Practices

1. **Plan Domain Structure** - Design URL hierarchy before implementation
2. **Use HTTPS** - Always enable SSL for production domains
3. **DNS First** - Configure DNS before adding domains to Shipyard
4. **Monitor Certificates** - Set up alerts for SSL expiration
5. **Document Domains** - Keep track of domain ownership and purpose
6. **Test Thoroughly** - Verify domains work before removing old ones
7. **Backup Configuration** - Export domain lists regularly

## Advanced Configuration

### Custom Ingress Annotations

Modify generated ingress with custom annotations by editing `manifests/shared/ingress-*.yaml` files after generation (note: changes will be overwritten on next deploy).

### External DNS Integration

For automatic DNS record management:

```bash
# Install external-dns
kubectl apply -f https://github.com/kubernetes-sigs/external-dns/releases/latest/download/external-dns.yaml

# Domains will automatically create DNS records
```

### Ingress Classes

Specify custom ingress class:

```yaml
# In generated ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: traefik  # Instead of nginx
```