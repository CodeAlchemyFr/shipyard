# shipyard releases

List deployment history and versions for an application.

## Synopsis

Display the complete deployment history for your application, including version information, deployment status, and rollback details.

## Usage

```
shipyard releases [flags]
```

## Flags

```
  -h, --help   help for releases
```

## Example Output

```
🕐 Deployment History for web-app:

VERSION        IMAGE                    STATUS    DEPLOYED             ROLLBACK
v1703123456    myapp:v1.3.1            success   2024-01-15 16:45:00
v1703122000    myapp:v1.3.0            failed    2024-01-15 15:30:00
v1703120000    myapp:v1.2.0            success   2024-01-15 14:30:00  ← to v1703118000
v1703118000    myapp:v1.1.0            success   2024-01-15 13:15:00
v1703115000    myapp:v1.0.0            success   2024-01-15 12:00:00

📊 Summary:
   Total deployments: 5
   Successful: 4 (80%)
   Failed: 1 (20%)
   Rollbacks: 1
```

## Information Displayed

Each release entry shows:

- **VERSION** - Unique deployment identifier (timestamp-based)
- **IMAGE** - Container image with tag used in deployment
- **STATUS** - Deployment result (success, failed, pending)
- **DEPLOYED** - When the deployment was initiated
- **ROLLBACK** - If this was a rollback, shows source version

## Status Types

| Status | Description |
|--------|-------------|
| `success` | Deployment completed successfully |
| `failed` | Deployment failed during process |
| `pending` | Deployment in progress |

## Version Format

Versions use timestamp format: `v{unix-timestamp}`
- `v1703123456` = January 15, 2024 16:45:00 UTC
- Allows chronological sorting
- Unique per deployment

## Use Cases

### Pre-Rollback Analysis

```bash
# Check deployment history before rollback
shipyard releases

# Identify last successful version
shipyard rollback v1703120000
```

### Deployment Tracking

```bash
# After each deployment
shipyard deploy
shipyard releases  # View updated history
```

### Troubleshooting Failed Deployments

```bash
# See which deployments failed
shipyard releases

# Check patterns in failures
# - Same image failing repeatedly?
# - Recent configuration changes?
```

### Release Planning

```bash
# Review deployment frequency
shipyard releases

# Identify stable versions for production
# Plan rollback strategy
```

## Integration with Other Commands

### With Rollback

```bash
# View available versions
shipyard releases

# Rollback to specific version
shipyard rollback v1703118000

# Verify rollback in history
shipyard releases
```

### With Status

```bash
# Check current status
shipyard status

# View deployment history
shipyard releases

# Compare current vs. previous deployments
```

### With Deploy

```bash
# Deploy new version
shipyard deploy

# Confirm deployment recorded
shipyard releases

# Check if deployment succeeded
shipyard status
```

## Database Storage

Release information is stored in local SQLite database:
- Location: `manifests/shipyard.db`
- Persistent across CLI sessions
- Tracks full deployment lifecycle

## Filtering and Sorting

Results are automatically:
- **Sorted** by deployment time (newest first)
- **Limited** to current application (from `paas.yaml`)
- **Formatted** for easy reading

## Troubleshooting

### No Releases Found

```
🕐 Deployment History for web-app:
   No deployments found.
```

This means no deployments have been made yet:
```bash
shipyard deploy  # Create first deployment
```

### Database Issues

```
Error: failed to load deployment history
```

Check database file:
```bash
ls -la manifests/shipyard.db
```

If missing, will be created on next deployment.

### Wrong Application

Make sure you're in the correct directory with the right `paas.yaml`:
```bash
cat paas.yaml  # Check app name
cd /path/to/correct/app
shipyard releases
```

## Export and Backup

### Export History

```bash
# View releases and save to file
shipyard releases > deployment-history.txt

# Export with timestamp
shipyard releases > "history-$(date +%Y%m%d).txt"
```

### Database Backup

```bash
# Backup deployment database
cp manifests/shipyard.db backup/shipyard-$(date +%Y%m%d).db
```

## Advanced Usage

### Analyze Deployment Patterns

```bash
# Check deployment frequency
shipyard releases | grep -E "v[0-9]+" | wc -l

# Find failed deployments
shipyard releases | grep failed

# Count rollbacks
shipyard releases | grep "←" | wc -l
```

### CI/CD Integration

```yaml
# GitHub Actions - Track deployment success rate
- name: Check deployment history
  run: |
    shipyard releases
    TOTAL=$(shipyard releases | grep -E "v[0-9]+" | wc -l)
    FAILED=$(shipyard releases | grep failed | wc -l)
    echo "Success rate: $(( (TOTAL - FAILED) * 100 / TOTAL ))%"
```

## Related Commands

- [`shipyard deploy`](/cli/deploy) - Create new deployment
- [`shipyard rollback`](/cli/rollback) - Rollback to previous version
- [`shipyard status`](/cli/status) - Check current deployment status
- [`shipyard logs`](/cli/logs) - View application logs