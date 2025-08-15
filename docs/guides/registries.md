# Working with Private Registries

Shipyard supports seamless integration with private container registries, handling authentication automatically during deployment.

## Supported Registries

- **Docker Hub** (`docker.io`)
- **GitHub Container Registry** (`ghcr.io`)
- **GitLab Container Registry** (`registry.gitlab.com`)
- **AWS Elastic Container Registry** (`*.dkr.ecr.*.amazonaws.com`)
- **Azure Container Registry** (`*.azurecr.io`)
- **Google Container Registry** (`gcr.io`, `*.gcr.io`)
- **Harbor and other private registries**

## Quick Start

```bash
# Add registry credentials
shipyard registry add ghcr.io username token

# Deploy app with private image
echo 'app:
  name: private-app
  image: ghcr.io/user/private-repo:latest
  port: 3000' > paas.yaml

shipyard deploy
```

## GitHub Container Registry

### Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Create token with `read:packages` scope
3. Add to Shipyard:

```bash
shipyard registry add ghcr.io your-username ghp_your_token
```

### GitHub Actions

```yaml
- name: Configure registry
  run: |
    shipyard registry add ghcr.io ${{ github.actor }} ${{ secrets.GITHUB_TOKEN }}

- name: Deploy
  run: shipyard deploy
```

### Example Configuration

```yaml
app:
  name: github-app
  image: ghcr.io/company/api:v1.0.0
  port: 8080
```

## Docker Hub Private Repositories

### Username and Password

```bash
shipyard registry add docker.io your-username your-password
```

### Access Token (Recommended)

1. Go to Docker Hub → Account Settings → Security
2. Create new access token
3. Add to Shipyard:

```bash
shipyard registry add docker.io your-username your-access-token
```

### Example Configuration

```yaml
app:
  name: docker-app
  image: company/private-api:latest
  port: 3000
```

## AWS Elastic Container Registry (ECR)

### AWS CLI Authentication

```bash
# Get login token
aws ecr get-login-password --region us-west-2 | docker login --username AWS --password-stdin 123456789.dkr.ecr.us-west-2.amazonaws.com

# Add to Shipyard
shipyard registry add 123456789.dkr.ecr.us-west-2.amazonaws.com AWS $(aws ecr get-login-password --region us-west-2)
```

### IAM Roles (EKS)

For EKS clusters, configure IAM roles for service accounts:

```yaml
# serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ecr-service-account
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::123456789:role/ECRAccessRole
```

### Example Configuration

```yaml
app:
  name: aws-app
  image: 123456789.dkr.ecr.us-west-2.amazonaws.com/my-app:latest
  port: 8080
```

## GitLab Container Registry

### Deploy Token

1. Go to Project → Settings → Repository → Deploy tokens
2. Create token with `read_registry` scope
3. Add to Shipyard:

```bash
shipyard registry add registry.gitlab.com gitlab+deploy-token-username your-token
```

### Personal Access Token

```bash
shipyard registry add registry.gitlab.com your-username your-pat-token
```

### GitLab CI

```yaml
deploy:
  script:
    - shipyard registry add $CI_REGISTRY $CI_REGISTRY_USER $CI_REGISTRY_PASSWORD
    - shipyard deploy
```

### Example Configuration

```yaml
app:
  name: gitlab-app
  image: registry.gitlab.com/group/project/app:latest
  port: 3000
```

## Azure Container Registry

### Service Principal

```bash
# Create service principal
az ad sp create-for-rbac --name shipyard-sp --role acrpull --scopes /subscriptions/{subscription-id}/resourceGroups/{resource-group}/providers/Microsoft.ContainerRegistry/registries/{registry-name}

# Add to Shipyard
shipyard registry add myregistry.azurecr.io service-principal-id service-principal-password
```

### Admin User

```bash
# Enable admin user
az acr update --name myregistry --admin-enabled true

# Get credentials
az acr credential show --name myregistry

# Add to Shipyard
shipyard registry add myregistry.azurecr.io admin-username admin-password
```

### Example Configuration

```yaml
app:
  name: azure-app
  image: myregistry.azurecr.io/app:v1.0
  port: 8080
```

## Google Container Registry

### Service Account Key

1. Create service account in Google Cloud Console
2. Grant Storage Object Viewer role
3. Download JSON key file
4. Use JSON key as password:

```bash
shipyard registry add gcr.io _json_key "$(cat key.json)"
```

### gcloud Authentication

```bash
# Configure Docker credential helper
gcloud auth configure-docker

# Get access token
gcloud auth print-access-token

# Add to Shipyard
shipyard registry add gcr.io oauth2accesstoken $(gcloud auth print-access-token)
```

### Example Configuration

```yaml
app:
  name: gcp-app
  image: gcr.io/project-id/app:latest
  port: 8080
```

## Self-Hosted/Harbor Registries

### Basic Authentication

```bash
shipyard registry add my-registry.company.com username password
```

### With Custom Port

```bash
shipyard registry add my-registry.company.com:5000 admin secret123
```

### TLS Configuration

For registries with custom certificates, ensure your Kubernetes nodes trust the certificate.

### Example Configuration

```yaml
app:
  name: harbor-app
  image: harbor.company.com/project/app:v2.0
  port: 9000
```

## Registry Management

### List Registries

```bash
shipyard registry list
```

### Set Default Registry

```bash
shipyard registry default docker.io
```

### Remove Registry

```bash
shipyard registry remove ghcr.io
```

### Update Credentials

Remove and re-add:
```bash
shipyard registry remove ghcr.io
shipyard registry add ghcr.io username new-token
```

## Automatic Detection

Shipyard automatically detects which registry to use based on image name:

| Image Format | Detected Registry |
|--------------|-------------------|
| `nginx` | `docker.io` |
| `myuser/app` | `docker.io` |
| `ghcr.io/user/app` | `ghcr.io` |
| `gcr.io/project/app` | `gcr.io` |
| `my-registry.com/app` | `my-registry.com` |

## Security Best Practices

### Token Permissions

Use minimal required permissions:
- **GitHub**: `read:packages` only
- **Docker Hub**: Read-only access tokens
- **AWS**: `ECRReadOnlyAccess` policy
- **Azure**: `acrpull` role only

### Token Rotation

Regularly rotate access tokens:
```bash
# Update token
shipyard registry remove ghcr.io
shipyard registry add ghcr.io username new-token
```

### Environment Variables

Don't hardcode credentials in scripts:
```bash
# Good
shipyard registry add ghcr.io $GITHUB_USER $GITHUB_TOKEN

# Bad
shipyard registry add ghcr.io myuser ghp_abc123
```

### CI/CD Secrets

Store credentials as CI/CD secrets, not in code:

```yaml
# GitHub Actions
- name: Configure registry
  run: shipyard registry add ghcr.io ${{ github.actor }} ${{ secrets.GITHUB_TOKEN }}

# GitLab CI  
script:
  - shipyard registry add $CI_REGISTRY $CI_REGISTRY_USER $CI_REGISTRY_PASSWORD
```

## Troubleshooting

### Authentication Failed

```bash
Error: failed to pull image: authentication required
```

**Solutions:**
1. Verify credentials with `docker login`
2. Check token permissions and expiration
3. Ensure correct registry URL format

### Registry Not Found

```bash
Error: registry ghcr.io not found
```

**Solution:**
```bash
shipyard registry add ghcr.io username token
```

### Invalid Image Format

```bash
Error: failed to parse image: invalid reference format
```

**Solution:** Use full image references:
```yaml
# Good
image: ghcr.io/user/app:v1.0

# Bad  
image: user/app:v1.0  # Missing registry
```

### Rate Limiting

For Docker Hub rate limits, use authentication:
```bash
shipyard registry add docker.io username password
```

Or use alternative registries:
```yaml
# Instead of nginx:latest
image: ghcr.io/library/nginx:latest
```

## Multi-Registry Applications

For applications using images from multiple registries:

```yaml
# Main application
app:
  name: multi-app
  image: ghcr.io/company/app:latest  # Uses ghcr.io credentials
  port: 3000

# If you have init containers or sidecars from different registries,
# add credentials for all registries used
```

```bash
# Add credentials for all registries
shipyard registry add ghcr.io company-user token1
shipyard registry add gcr.io _json_key "$(cat key.json)" 
shipyard registry add docker.io docker-user token2
```

Shipyard will automatically use the appropriate credentials for each image during deployment.