# shipyard rollback

Rollback to a previous deployment version.

## Synopsis

Restore your application to a previous deployment version. This command can rollback to the latest successful deployment or to a specific version.

## Usage

```
shipyard rollback [version|image-tag] [flags]
```

## Arguments

- `version|image-tag` (optional) - Specific version or image tag to rollback to. If not provided, rollback to latest successful deployment.

## Flags

```
  -h, --help   help for rollback
```

## How Rollback Works

1. **Finds target version** - Either specified or latest successful
2. **Creates new deployment** - Generates new version ID for the rollback
3. **Updates manifests** - Regenerates Kubernetes files with previous image
4. **Applies changes** - Deploys the rollback to Kubernetes
5. **Tracks rollback** - Records the rollback in deployment history

## Examples

### Rollback to Latest Successful

```bash
# Automatic rollback to last working deployment
shipyard rollback
```

Output:
```
🔄 Starting rollback...
🔍 Finding latest successful deployment...
📍 Found latest successful: v1703120000 (v1.2.0)
🎯 Rolling back to:
   Version: v1703120000
   Image: myapp:v1.2.0
   Deployed: 2024-01-15 14:30:00
📦 Generating rollback manifests...
☸️  Applying rollback to Kubernetes cluster...
✅ Rollback successful!
   Rolled back from current to v1703120000 (v1.2.0)
   New deployment version: v1703123456
```

### Rollback to Specific Version

```bash
# Rollback to specific version ID
shipyard rollback v1703120000

# Rollback to specific image tag
shipyard rollback v1.2.0
```

### View Available Versions

```bash
# List deployment history first
shipyard releases

# Then rollback to chosen version
shipyard rollback v1703118000
```

## Rollback Tracking

Each rollback creates a new deployment entry:

```bash
shipyard releases
```

Output shows rollback entries:
```
🕐 Deployment History for web-app:

VERSION        IMAGE           STATUS    DEPLOYED             ROLLBACK
v1703123456    myapp:v1.2.0    success   2024-01-15 15:45:00  ← to v1703120000
v1703122000    myapp:v1.3.0    failed    2024-01-15 15:30:00
v1703120000    myapp:v1.2.0    success   2024-01-15 14:30:00
v1703118000    myapp:v1.1.0    success   2024-01-15 13:15:00
```

## Safety Features

### Automatic Target Selection

If no version specified, Shipyard automatically finds the latest successful deployment:

```bash
shipyard rollback
# Automatically skips failed deployments
# Finds last deployment with status "success"
```

### Configuration Validation

Rollback validates the target deployment:
- Version exists in deployment history
- Image is still accessible
- Configuration is compatible

### Graceful Failure

If rollback fails:
- Original deployment remains active
- Rollback is marked as "failed" in history
- Detailed error message provided

## Use Cases

### Failed Deployment Recovery

```bash
# Deploy new version
shipyard deploy

# Check if it's working
shipyard status
shipyard logs web-app

# If broken, rollback immediately
shipyard rollback
```

### Planned Rollback

```bash
# List available versions
shipyard releases

# Choose stable version for rollback
shipyard rollback v1703118000

# Verify rollback worked
shipyard status
```

### Emergency Rollback

```bash
# Quick rollback during incident
shipyard rollback

# Monitor the rollback
shipyard logs web-app --follow
```

## Integration with CI/CD

### GitHub Actions

```yaml
- name: Deploy with rollback on failure
  run: |
    if ! shipyard deploy; then
      echo "Deployment failed, rolling back..."
      shipyard rollback
      exit 1
    fi
```

### Rollback Script

```bash
#!/bin/bash
# safe-deploy.sh

echo "Deploying..."
if shipyard deploy; then
  echo "✅ Deployment successful"
  
  # Wait and check health
  sleep 30
  if ! curl -f http://app.example.com/health; then
    echo "❌ Health check failed, rolling back"
    shipyard rollback
    exit 1
  fi
else
  echo "❌ Deployment failed, rolling back"
  shipyard rollback
  exit 1
fi
```

## Troubleshooting

### No Successful Deployments

```
Error: failed to find successful deployment
```

This means all previous deployments failed. Check deployment history:
```bash
shipyard releases
```

Manual recovery may be needed.

### Version Not Found

```
Error: failed to find version v1703120000
```

Check available versions:
```bash
shipyard releases
```

Use a valid version ID.

### Rollback Failed

```
Error: failed to apply rollback manifests
```

Common causes:
- Kubernetes connection issues
- Resource constraints
- Image no longer available

Check cluster status:
```bash
shipyard status
kubectl get events
```

### Image Pull Errors

```
Error: failed to pull image myapp:v1.2.0
```

The target image may no longer exist in the registry. Choose a different version:
```bash
shipyard releases
shipyard rollback v1703118000  # Try older version
```

## Best Practices

1. **Test Before Rollback** - Check deployment history first
2. **Monitor After Rollback** - Verify the rollback worked
3. **Document Rollbacks** - Note why rollback was needed
4. **Keep Images Available** - Don't delete old container images
5. **Use Health Checks** - Implement proper health endpoints for validation