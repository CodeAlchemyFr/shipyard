# shipyard registry

Manage container registry credentials for private images.

## Synopsis

The registry command allows you to add, list, remove, and manage container registry credentials. These credentials are encrypted and stored locally, then used automatically when deploying applications with private images.

## Usage

```
shipyard registry [command]
```

## Available Commands

- [`add`](#add) - Add new registry credentials
- [`list`](#list) - List configured registries  
- [`remove`](#remove) - Remove registry credentials
- [`default`](#default) - Set default registry

## add

Add credentials for a container registry.

### Usage

```
shipyard registry add [registry-url] [username] [password/token] [flags]
```

### Arguments

- `registry-url` - Registry URL (e.g., `ghcr.io`, `docker.io`, `my-registry.com:5000`)
- `username` - Registry username or identifier
- `password/token` - Password, access token, or personal access token

### Flags

```
      --default        Set this registry as default
      --email string   Email for registry authentication  
      --type string    Registry type (docker, github, gitlab, aws) (default "docker")
  -h, --help           help for add
```

### Examples

```bash
# Add GitHub Container Registry token
shipyard registry add ghcr.io myuser ghp_token123

# Add Docker Hub credentials as default
shipyard registry add --default docker.io myuser mypassword

# Add private registry with email
shipyard registry add my-registry.com:5000 user token --email user@example.com

# Add AWS ECR registry
shipyard registry add 123456789.dkr.ecr.us-west-2.amazonaws.com user token --type aws
```

## list

Display all configured registry credentials.

### Usage

```
shipyard registry list
```

### Example Output

```
📋 Container Registry Credentials:

Registry URL                     Username        Type      Default   Created
─────────────────────────────────────────────────────────────────────────────────
ghcr.io                         myuser          github              2024-01-15 10:30
docker.io                       dockeruser      docker    ✓         2024-01-14 15:45
my-registry.com:5000            admin           docker              2024-01-16 09:15
```

::: warning Security Note
Passwords are displayed as `***` for security. The actual encrypted passwords are stored securely in the local SQLite database.
:::

## remove

Remove credentials for a container registry.

### Usage

```
shipyard registry remove [registry-url]
```

### Arguments

- `registry-url` - Registry URL to remove

### Example

```bash
shipyard registry remove ghcr.io
```

## default

Set a registry as the default for image pulls.

### Usage

```
shipyard registry default [registry-url]
```

### Arguments

- `registry-url` - Registry URL to set as default

### Example

```bash
shipyard registry default docker.io
```

## Registry Detection

Shipyard automatically detects which registry credentials to use based on your image name:

| Image Format | Detected Registry |
|--------------|-------------------|
| `nginx` | `docker.io` (Docker Hub) |
| `myuser/myapp` | `docker.io` (Docker Hub) |
| `ghcr.io/user/app` | `ghcr.io` (GitHub Container Registry) |
| `my-registry.com/app` | `my-registry.com` |
| `localhost:5000/app` | `localhost:5000` |

## Security

- **Encryption**: All passwords and tokens are encrypted using AES-256-GCM before storage
- **Local Storage**: Credentials are stored in a local SQLite database (`~/.shipyard/credentials.db`)
- **No Network**: Credentials are never transmitted except during deployment to Kubernetes

## Supported Registries

- **Docker Hub** (`docker.io`)
- **GitHub Container Registry** (`ghcr.io`) 
- **GitLab Container Registry** (`registry.gitlab.com`)
- **AWS Elastic Container Registry** (`*.dkr.ecr.*.amazonaws.com`)
- **Azure Container Registry** (`*.azurecr.io`)
- **Google Container Registry** (`gcr.io`, `*.gcr.io`)
- **Private/Self-hosted registries**

## Troubleshooting

### Registry not found
```bash
Error: registry ghcr.io not found
```
Make sure you've added the registry first:
```bash
shipyard registry add ghcr.io username token
```

### Authentication failed
```bash
Error: failed to pull image: authentication required
```
Verify your credentials:
```bash
shipyard registry list
docker login ghcr.io  # Test credentials manually
```

### Permission denied
```bash
Error: failed to add registry: permission denied
```
Check that you have write permissions to the Shipyard data directory.

## Integration with Deployment

When you run `shipyard deploy`, the CLI:

1. **Detects** which registry your image uses
2. **Retrieves** encrypted credentials for that registry  
3. **Generates** a Kubernetes `imagePullSecret`
4. **Applies** the secret to your deployment

This happens automatically - no manual secret management required.

## Examples

### GitHub Actions Workflow

```yaml
- name: Add registry credentials
  run: |
    shipyard registry add ghcr.io ${{ github.actor }} ${{ secrets.GITHUB_TOKEN }}
    
- name: Deploy
  run: shipyard deploy
```

### Private Registry Setup

```bash
# Add private registry
shipyard registry add my-company-registry.com builduser $BUILD_TOKEN

# Deploy app using private image
echo 'app:
  name: internal-app
  image: my-company-registry.com/team/app:latest
  port: 8080' > paas.yaml

shipyard deploy
```