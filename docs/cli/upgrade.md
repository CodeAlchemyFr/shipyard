# shipyard upgrade

Upgrade Shipyard CLI to the latest version automatically.

## Usage

```bash
shipyard upgrade [flags]
```

## Description

The `upgrade` command downloads and installs the latest version of Shipyard CLI from GitHub releases. It automatically detects your platform and architecture, downloads the appropriate binary, and replaces the current installation.

## Examples

### Standard upgrade

```bash
shipyard upgrade
```

This will:
1. Check the current version
2. Fetch the latest version from GitHub
3. Ask for confirmation
4. Download and install the new version

### Force upgrade

```bash
shipyard upgrade --force
```

Skip version checking and force the upgrade even if you're already on the latest version.

### Skip confirmation

```bash
shipyard upgrade --yes
```

Automatically confirm the upgrade without prompting.

### Combine flags

```bash
shipyard upgrade --force --yes
```

Force upgrade without any prompts - useful for automation.

## Flags

| Flag | Description |
|------|-------------|
| `--force` | Force upgrade without version check |
| `--yes` | Skip confirmation prompt |
| `-h, --help` | Help for upgrade command |

## How It Works

### Version Detection

The upgrade process:

1. **Current version**: Reads from `shipyard --version`
2. **Latest version**: Fetches from GitHub API
3. **Comparison**: Checks if upgrade is needed
4. **Platform detection**: Determines OS and architecture

### Download and Installation

1. **Backup**: Creates a backup of the current binary
2. **Download**: Fetches the appropriate binary from GitHub releases
3. **Verification**: Validates the download
4. **Installation**: Replaces the current binary
5. **Cleanup**: Removes temporary files and backup

### Platform Support

Automatically detects and downloads the correct binary:

| Platform | Architecture | Binary Name |
|----------|-------------|-------------|
| Linux | AMD64 | `shipyard-linux-amd64` |
| Linux | ARM64 | `shipyard-linux-arm64` |
| macOS | Intel | `shipyard-darwin-amd64` |
| macOS | Apple Silicon | `shipyard-darwin-arm64` |
| Windows | AMD64 | `shipyard-windows-amd64.exe` |
| Windows | ARM64 | `shipyard-windows-arm64.exe` |

## Interactive Process

### Standard upgrade flow

```bash
$ shipyard upgrade
🚀 Upgrading Shipyard CLI...
📋 Current version: v1.2.0
🔍 Checking for latest version...
📦 Latest version: v1.3.0
🖥️  Detected platform: linux-amd64
📥 Downloading from: https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-amd64

⚠️  Upgrade Shipyard CLI from v1.2.0 to v1.3.0?
Continue? [y/N]: y

💾 Creating backup: /usr/local/bin/shipyard.backup
🔄 Installing new version...
✅ Upgrade completed successfully!
🎉 Shipyard CLI updated to v1.3.0
📚 Run 'shipyard --version' to verify the new version
```

### Already up to date

```bash
$ shipyard upgrade
🚀 Upgrading Shipyard CLI...
📋 Current version: v1.3.0
🔍 Checking for latest version...
📦 Latest version: v1.3.0
✅ You already have the latest version!
```

### Force upgrade

```bash
$ shipyard upgrade --force
🚀 Upgrading Shipyard CLI...
📋 Current version: v1.3.0
🔍 Checking for latest version...
📦 Latest version: v1.3.0
🖥️  Detected platform: linux-amd64
⚠️  Upgrade Shipyard CLI from v1.3.0 to v1.3.0?
Continue? [y/N]: y
✅ Upgrade completed successfully!
```

## Error Handling

### Network Issues

```bash
$ shipyard upgrade
❌ Failed to check latest version: network unreachable
```

### Download Failures

```bash
$ shipyard upgrade
❌ Failed to download: 404 Not Found
```

The command will automatically restore the backup if installation fails.

### Permission Issues

```bash
$ shipyard upgrade
❌ Failed to replace executable: permission denied
🔄 Restoring backup...
```

Solution: Run with appropriate permissions or install to a user-writable directory.

## Backup and Recovery

### Automatic Backup

The upgrade process automatically creates a backup:
- **Location**: Same directory as current binary with `.backup` extension
- **Restoration**: Automatic if upgrade fails
- **Cleanup**: Removed after successful upgrade

### Manual Recovery

If something goes wrong, you can manually restore:

```bash
# If backup exists
mv /usr/local/bin/shipyard.backup /usr/local/bin/shipyard

# Or reinstall from releases
curl -sSL https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/install.sh | bash
```

## CI/CD Integration

### Automated Updates

```bash
#!/bin/bash
# Update Shipyard in CI/CD pipeline
shipyard upgrade --force --yes

# Verify the upgrade
shipyard --version
```

### Docker Integration

```dockerfile
# Update Shipyard in Docker container
RUN shipyard upgrade --force --yes
```

## Security Considerations

### Download Verification

- Downloads directly from official GitHub releases
- Uses HTTPS for all connections
- Validates binary before installation

### Backup Strategy

- Always creates backup before upgrade
- Automatic rollback on failure
- Preserves file permissions

## Troubleshooting

### Common Issues

#### GitHub Rate Limiting
```bash
$ shipyard upgrade
❌ Failed to check latest version: rate limit exceeded
```
**Solution**: Wait and try again, or use GitHub authentication.

#### Binary Not Found
```bash
$ shipyard upgrade
❌ Failed to download: binary not found for platform
```
**Solution**: Check if your platform is supported or install manually.

#### Permission Denied
```bash
$ shipyard upgrade
❌ Failed to replace executable: permission denied
```
**Solutions**:
- Run with `sudo` (Linux/macOS)
- Install to user directory
- Use administrator privileges (Windows)

### Debug Mode

For troubleshooting, check the download URL manually:

```bash
# Check available releases
curl -s https://api.github.com/repos/CodeAlchemyFr/shipyard/releases/latest

# Verify binary exists
curl -I https://github.com/CodeAlchemyFr/shipyard/releases/latest/download/shipyard-linux-amd64
```

## Alternative Upgrade Methods

### Manual Download

```bash
# Download specific version
wget https://github.com/CodeAlchemyFr/shipyard/releases/download/v1.3.0/shipyard-linux-amd64
chmod +x shipyard-linux-amd64
sudo mv shipyard-linux-amd64 /usr/local/bin/shipyard
```

### Package Managers

If available through package managers:

```bash
# Homebrew (macOS)
brew upgrade shipyard

# Chocolatey (Windows)
choco upgrade shipyard
```

## Related Commands

- [`shipyard --version`](./overview.md) - Check current version
- [`shipyard ssl install`](./ssl.md) - Install SSL components
- [Installation Guide](../getting-started/quick-start.md) - Initial installation