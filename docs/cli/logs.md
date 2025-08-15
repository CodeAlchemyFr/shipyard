# shipyard logs

View application logs from Kubernetes pods.

## Synopsis

Stream or view logs from your deployed applications. This command connects to your Kubernetes cluster and retrieves logs from the running pods.

## Usage

```
shipyard logs [app-name] [flags]
```

## Arguments

- `app-name` - Name of the application to view logs for

## Flags

```
  -f, --follow     Follow log output (stream logs)
  -t, --tail int   Number of lines to show from the end (default 100)
  -h, --help       help for logs
```

## Examples

### View Recent Logs

```bash
# Show last 100 lines
shipyard logs web-app

# Show last 50 lines
shipyard logs web-app --tail 50
```

### Follow Logs in Real-time

```bash
# Stream logs continuously
shipyard logs web-app --follow

# Follow with shorter tail
shipyard logs api-service -f --tail 20
```

### View All Application Logs

```bash
# If no app name provided, shows all Shipyard apps
shipyard logs
```

## Log Output Format

Logs include timestamps and pod information:

```
2024-01-15T10:30:45.123Z [web-app-7d4b8c9f-xyz] INFO  Server listening on port 3000
2024-01-15T10:30:46.456Z [web-app-7d4b8c9f-xyz] INFO  Database connected successfully
2024-01-15T10:30:47.789Z [web-app-7d4b8c9f-abc] INFO  Health check endpoint ready
```

## Multiple Pods

For applications with multiple replicas, logs from all pods are combined:

```bash
shipyard logs web-app
# Shows logs from:
# - web-app-7d4b8c9f-xyz
# - web-app-7d4b8c9f-abc  
# - web-app-7d4b8c9f-def
```

## Integration with Deployment

Monitor logs during and after deployment:

```bash
# Deploy and monitor
shipyard deploy &
shipyard logs web-app --follow

# Check for errors after deployment
shipyard deploy
shipyard logs web-app --tail 200 | grep -i error
```

## Troubleshooting

### Application Not Found

```
Error: application "web-app" not found
```

Check deployed applications:
```bash
shipyard status
kubectl get pods -l app=web-app
```

### No Logs Available

```
No logs available for web-app
```

Possible causes:
- Pod hasn't started yet
- Application isn't logging to stdout/stderr
- Pod crashed before logging

Check pod status:
```bash
kubectl describe pod <pod-name>
kubectl get events --sort-by=.metadata.creationTimestamp
```

### Connection Issues

```
Error: failed to connect to Kubernetes cluster
```

Verify cluster connection:
```bash
kubectl cluster-info
kubectl get pods
```

## Advanced Usage

### Filter Logs

```bash
# Show only error logs
shipyard logs web-app | grep -i error

# Show logs from specific time
shipyard logs web-app --since=1h

# Show logs with context
shipyard logs web-app | grep -A 5 -B 5 "exception"
```

### Export Logs

```bash
# Save logs to file
shipyard logs web-app --tail 1000 > app-logs.txt

# Save with timestamp
shipyard logs web-app > "logs-$(date +%Y%m%d-%H%M%S).txt"
```

### Compare with kubectl

Shipyard logs is equivalent to:

```bash
# Shipyard command
shipyard logs web-app --follow

# Equivalent kubectl command  
kubectl logs -f deployment/web-app --all-containers=true
```