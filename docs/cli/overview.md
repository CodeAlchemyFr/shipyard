# CLI Reference

The Shipyard CLI provides a simple interface for deploying and managing applications on Kubernetes.

## Global Options

All commands support these global flags:

```
-h, --help   Show help for any command
```

## Available Commands

| Command | Description |
|---------|-------------|
| [`deploy`](/cli/deploy) | Deploy an application to Kubernetes |
| [`delete`](/cli/delete) | Delete an application and all its resources |
| [`status`](/cli/status) | Show status of deployed applications |
| [`logs`](/cli/logs) | View application logs |
| [`rollback`](/cli/rollback) | Rollback to a previous deployment |
| [`releases`](/cli/releases) | List deployment history |
| [`registry`](/cli/registry) | Manage container registry credentials |
| [`domain`](/cli/domain) | Manage application domains |
| [`db`](/cli/db) | Database management commands |

## Basic Workflow

```bash
# 1. Configure your application
cat > paas.yaml << EOF
app:
  name: my-app
  image: my-registry/my-app:latest
  port: 3000
EOF

# 2. Add registry credentials (if needed)
shipyard registry add my-registry.com username token

# 3. Deploy
shipyard deploy

# 4. Monitor
shipyard status
shipyard logs my-app

# 5. Update and redeploy
# (edit paas.yaml with new image)
shipyard deploy

# 6. Rollback if needed
shipyard rollback

# 7. Clean up when done
shipyard delete
```

## Configuration File

All commands expect a `paas.yaml` file in the current directory. See [Configuration](/getting-started/configuration) for complete reference.

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Configuration error
- `3` - Kubernetes connection error

## Examples

### Deploy with custom resource limits
```bash
echo 'app:
  name: web-app
  image: nginx:latest
  port: 80
resources:
  cpu: "500m"
  memory: "512Mi"
scaling:
  min: 2
  max: 10' > paas.yaml

shipyard deploy
```

### Check deployment status
```bash
shipyard status
```

### View real-time logs
```bash
shipyard logs web-app --follow
```

### Rollback to specific version
```bash
shipyard releases  # List available versions
shipyard rollback v1703123456
```

### Add custom domain
```bash
shipyard domain add web-app.example.com
```

### Clean up application
```bash
shipyard delete web-app
```

### Remove all applications
```bash
shipyard delete --all
```

## Getting Help

For detailed help on any command:

```bash
shipyard <command> --help
```

For example:
```bash
shipyard deploy --help
shipyard registry add --help
```