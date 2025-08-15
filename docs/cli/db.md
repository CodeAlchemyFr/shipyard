# shipyard db

Database management commands for Shipyard's internal storage.

## Synopsis

Manage Shipyard's internal SQLite database that stores deployment history, domain configurations, and registry credentials.

## Usage

```
shipyard db [command]
```

## Available Commands

- [`init`](#init) - Initialize or reset the database
- [`migrate`](#migrate) - Run database migrations
- [`backup`](#backup) - Create database backup
- [`status`](#status) - Show database information

## init

Initialize or reset the Shipyard database.

### Usage

```
shipyard db init [flags]
```

### Flags

```
      --force   Force reset existing database
  -h, --help    help for init
```

### Examples

```bash
# Initialize new database
shipyard db init

# Reset existing database (WARNING: loses all data)
shipyard db init --force
```

## migrate

Run database schema migrations.

### Usage

```
shipyard db migrate [flags]
```

### Examples

```bash
# Run pending migrations
shipyard db migrate
```

## backup

Create a backup of the Shipyard database.

### Usage

```
shipyard db backup [output-file] [flags]
```

### Arguments

- `output-file` (optional) - Backup file path (default: auto-generated)

### Examples

```bash
# Create backup with auto-generated name
shipyard db backup

# Create backup with custom name
shipyard db backup my-backup.db

# Create backup with timestamp
shipyard db backup "backup-$(date +%Y%m%d-%H%M%S).db"
```

## status

Show database information and statistics.

### Usage

```
shipyard db status [flags]
```

### Example Output

```
📊 Shipyard Database Status:

Database: /path/to/manifests/shipyard.db
Size: 2.3 MB
Created: 2024-01-15 12:00:00

📈 Statistics:
Applications: 3
Total deployments: 25
Successful deployments: 22 (88%)
Failed deployments: 3 (12%)
Registry credentials: 2
Domain configurations: 5

📋 Tables:
- apps (3 records)
- deployments (25 records)  
- domains (5 records)
- registry_credentials (2 records)

🕐 Recent Activity:
- Last deployment: 2024-01-15 16:45:00
- Last domain change: 2024-01-15 14:30:00
- Last registry update: 2024-01-14 10:15:00
```

## Database Location

By default, Shipyard stores its database at:
```
./manifests/shipyard.db
```

This location can be configured via environment variable:
```bash
export SHIPYARD_DB_PATH="/custom/path/shipyard.db"
```

## Database Schema

The database contains these main tables:

### apps
Stores application information:
```sql
CREATE TABLE apps (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    created_at DATETIME,
    updated_at DATETIME
);
```

### deployments
Tracks deployment history:
```sql
CREATE TABLE deployments (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    version TEXT NOT NULL,
    image TEXT NOT NULL,
    image_tag TEXT,
    image_hash TEXT,
    config_json TEXT,
    config_hash TEXT,
    status TEXT,
    rollback_to_version TEXT,
    deployed_at DATETIME,
    completed_at DATETIME,
    error_message TEXT
);
```

### domains
Manages domain configurations:
```sql
CREATE TABLE domains (
    id INTEGER PRIMARY KEY,
    app_id INTEGER,
    hostname TEXT NOT NULL UNIQUE,
    base_domain TEXT,
    path TEXT DEFAULT '/',
    ssl_enabled BOOLEAN DEFAULT TRUE,
    created_at DATETIME,
    updated_at DATETIME
);
```

### registry_credentials
Stores encrypted registry credentials:
```sql
CREATE TABLE registry_credentials (
    id INTEGER PRIMARY KEY,
    registry_url TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL,
    password TEXT NOT NULL, -- Encrypted
    email TEXT,
    registry_type TEXT DEFAULT 'docker',
    is_default BOOLEAN DEFAULT FALSE,
    created_at DATETIME,
    updated_at DATETIME
);
```

## Backup and Recovery

### Automatic Backups

Shipyard doesn't create automatic backups. Create manual backups regularly:

```bash
# Daily backup script
#!/bin/bash
DATE=$(date +%Y%m%d)
shipyard db backup "backup-$DATE.db"

# Keep only last 7 days
find . -name "backup-*.db" -mtime +7 -delete
```

### Restore from Backup

```bash
# Stop any running operations
# Copy backup to database location
cp backup-20240115.db manifests/shipyard.db

# Verify database integrity
shipyard db status
```

### Export Data

```bash
# Export deployment history
shipyard releases > deployments-export.txt

# Export domain list
shipyard domain list > domains-export.txt

# Export registry list
shipyard registry list > registries-export.txt
```

## Troubleshooting

### Database Corruption

```
Error: database disk image is malformed
```

Recovery steps:
```bash
# Try database integrity check
sqlite3 manifests/shipyard.db "PRAGMA integrity_check;"

# If corrupted, restore from backup
cp backup-latest.db manifests/shipyard.db

# Or reinitialize (loses data)
shipyard db init --force
```

### Database Locked

```
Error: database is locked
```

Possible causes:
- Another Shipyard process running
- Unclean shutdown

Solutions:
```bash
# Check for running processes
ps aux | grep shipyard

# Remove lock file if safe
rm manifests/shipyard.db-shm manifests/shipyard.db-wal

# Restart command
```

### Permission Denied

```
Error: permission denied: manifests/shipyard.db
```

Fix permissions:
```bash
# Make directory writable
chmod 755 manifests/

# Make database writable (if exists)
chmod 644 manifests/shipyard.db
```

### Migration Failures

```
Error: migration failed
```

Check database status and retry:
```bash
shipyard db status
shipyard db migrate
```

If persistent, backup and reinitialize:
```bash
shipyard db backup
shipyard db init --force
# Note: This loses deployment history
```

## Security

### Encryption

Registry credentials are encrypted using AES-256-GCM before storage.

### Access Control

- Database is only accessible locally
- No network exposure
- File system permissions control access

### Sensitive Data

The database contains:
- ✅ **Encrypted** registry passwords/tokens
- ✅ **Safe** deployment history and configurations
- ✅ **Safe** domain mappings

## Maintenance

### Regular Tasks

```bash
# Weekly backup
shipyard db backup

# Check database health
shipyard db status

# Clean old deployments (if needed)
# Note: No built-in cleanup yet
```

### Performance

For large deployment histories, consider:
- Regular backups and archival
- Monitoring database size
- Cleaning very old deployment records (manual)

## Environment Variables

```bash
# Custom database location
export SHIPYARD_DB_PATH="/data/shipyard.db"

# Database connection timeout
export SHIPYARD_DB_TIMEOUT="30s"
```