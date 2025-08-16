# shipyard delete

Delete an application and clean up all associated resources.

## Usage

```bash
shipyard delete [app-name] [flags]
```

## Description

The `delete` command removes an application and cleans up all associated resources:

- **Kubernetes resources**: deployment, service, ingress, secrets, ConfigMaps, HPA
- **Local manifest files**: removes the entire app directory from `manifests/apps/`
- **Database entries**: removes app data, deployment history, metrics, and events

## Examples

### Delete current application

Delete the application defined in the current directory's `paas.yaml`:

```bash
shipyard delete
```

### Delete specific application

Delete a specific application by name:

```bash
shipyard delete hello-world
```

### Delete all applications

Delete all applications managed by Shipyard:

```bash
shipyard delete --all
```

### Force deletion without confirmation

Skip the interactive confirmation prompt:

```bash
shipyard delete --force
shipyard delete hello-world --yes
```

## Flags

| Flag | Description |
|------|-------------|
| `--all` | Delete all applications |
| `--force` | Force deletion without confirmation |
| `--yes` | Automatically confirm deletion |
| `-h, --help` | Help for delete command |

## Interactive Confirmation

By default, `shipyard delete` will prompt for confirmation before deleting:

```bash
$ shipyard delete hello-world
⚠️  This will permanently delete the application 'hello-world' and all its resources:
   - Kubernetes deployment, service, ingress, secrets
   - Local manifest files
   - Database entries and deployment history

Are you sure you want to continue? [y/N]: 
```

## What Gets Deleted

### Kubernetes Resources

The following Kubernetes resources are removed:

- **Deployment**: The main application deployment
- **Service**: Load balancer and service discovery
- **Secrets**: Application secrets and registry credentials
- **ConfigMaps**: Configuration data
- **Ingress**: HTTP/HTTPS routing rules
- **HorizontalPodAutoscaler**: Auto-scaling configuration

### Local Files

- **Manifest directory**: `manifests/apps/[app-name]/` and all contents
- **Generated YAML files**: deployment.yaml, service.yaml, secrets.yaml, etc.

### Database Data

- **Application record**: Main app entry
- **Deployment history**: All past deployments and versions
- **Metrics data**: CPU, memory, and custom metrics
- **Health checks**: Health monitoring data
- **Events**: Application events and logs
- **Domain associations**: Connected domains and routing

## Safety Features

### Transaction Safety

Database deletions use transactions to ensure atomicity:
- If any part of the deletion fails, the entire operation is rolled back
- Prevents partial deletions that could leave the system in an inconsistent state

### Graceful Handling

- Resources that don't exist are silently skipped
- Warnings are displayed for resources that couldn't be deleted
- The command continues even if some cleanup operations fail

### Confirmation Prompts

- Interactive confirmation by default
- Clear explanation of what will be deleted
- Support for automated environments with `--force` and `--yes` flags

## Error Handling

Common scenarios and how they're handled:

### Application Not Found

```bash
$ shipyard delete nonexistent-app
ℹ️  No apps found to delete
```

### Kubernetes Access Issues

```bash
$ shipyard delete hello-world
☸️  Deleting Kubernetes resources for hello-world...
⚠️  Warning: Failed to delete some Kubernetes resources: connection refused
📁 Cleaning up local manifest files...
✅ Successfully deleted app: hello-world
```

### Partial Failures

The command will attempt to clean up all resources and report warnings for any failures, but will still complete successfully if the core deletion succeeds.

## Use Cases

### Development Cleanup

Quickly remove test applications during development:

```bash
# Clean up after testing
shipyard delete test-app --yes

# Remove all test deployments
shipyard delete --all --force
```

### CI/CD Integration

Use in automation scripts:

```bash
#!/bin/bash
# Deploy new version
shipyard deploy

# If deployment fails, clean up
if [ $? -ne 0 ]; then
    shipyard delete --force
    exit 1
fi
```

### Environment Reset

Clean slate for staging environments:

```bash
# Remove all applications
shipyard delete --all

# Verify cleanup
shipyard status
```

## Related Commands

- [`shipyard deploy`](deploy.md) - Deploy applications
- [`shipyard status`](status.md) - View application status
- [`shipyard rollback`](rollback.md) - Rollback to previous version
- [`shipyard releases`](releases.md) - View deployment history