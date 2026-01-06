# Version Management for Mainnet Deployments

## Overview

Walgo now automatically checks and updates Sui, Walrus, and site-builder versions before mainnet deployments to ensure you're using the latest stable versions. This feature helps prevent deployment issues and ensures compatibility with the latest network features.

## How It Works

When you run `walgo deploy` to a **mainnet** environment, Walgo will:

1. **Detect the network**: Check if you're deploying to mainnet using `sui client active-env`
2. **Check current versions**: Query installed versions of:
   - Sui CLI
   - Walrus CLI
   - site-builder
3. **Fetch latest versions**: Query GitHub releases to get the latest available versions
4. **Compare versions**: Determine if updates are available
5. **Prompt for update**: If updates are available, ask if you want to update now
6. **Auto-update**: Use `suiup` to install the latest versions for your network

## Usage

### Automatic (Default)

```bash
# Deploy to mainnet - will check versions automatically
walgo deploy --epochs 5
```

**Example interaction:**
```
🚀 Deploying to Walrus Sites...
  [1/5] Checking environment...
  ⚙️  Checking tool versions for mainnet deployment...

⚠️  Updates available for mainnet deployment:

  • Sui: 1.34.0 → 1.35.1
  • Walrus: 2.0.0 → 2.1.0
  • Site-builder: 2.0.0 → 2.1.0

💡 For mainnet deployments, it's recommended to use the latest versions.

Would you like to update now? [Y/n]: y

Updating tools...

Updating sui...
✓ sui updated successfully
Updating walrus...
✓ walrus updated successfully
Updating site-builder...
✓ site-builder updated successfully

✅ All tools updated successfully!

  [2/5] Analyzing changes...
  ...
```

### Skip Version Check

If you want to skip the version check (not recommended for mainnet):

```bash
walgo deploy --epochs 5 --skip-version-check
```

### For Testnet

Version checking is **only enforced for mainnet deployments**. Testnet deployments skip this check automatically.

```bash
# Switch to testnet
sui client switch --env testnet

# Deploy - no version check
walgo deploy --epochs 1
```

## Manual Version Management

You can also manually check and update versions:

### Check Versions

```bash
# Check current versions
sui --version
walrus --version
site-builder --version
```

### Update Manually

```bash
# Update Sui for mainnet
suiup install sui mainnet

# Update Walrus for mainnet
suiup install walrus mainnet

# Update site-builder for mainnet
suiup install site-builder mainnet
```

Or update all at once:

```bash
# For mainnet
suiup install sui mainnet && \
suiup install walrus mainnet && \
suiup install site-builder mainnet

# For testnet
suiup install sui testnet && \
suiup install walrus testnet && \
suiup install site-builder testnet
```

## Version Check Details

### Version Sources

- **Sui**: Latest version fetched from [MystenLabs/sui](https://github.com/MystenLabs/sui) GitHub releases
- **Walrus**: Latest version fetched from Walrus documentation repository
- **site-builder**: Uses the same version as Walrus (part of Walrus toolkit)

### Version Comparison

Versions are compared using semantic versioning:
- Major version (e.g., 1.x.x)
- Minor version (e.g., x.2.x)
- Patch version (e.g., x.x.3)

Pre-release tags (e.g., -beta, -alpha) are stripped before comparison.

### Update Process

When you choose to update, Walgo uses `suiup` to:
1. Download the latest binaries for your network (mainnet/testnet)
2. Install them to your local system
3. Verify the installation

## Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--skip-version-check` | Skip version checking and updating | `false` |

## Troubleshooting

### Version Check Fails

If the version check fails (e.g., network issues), Walgo will:
- Display a warning
- Continue with deployment using current versions

This ensures deployment isn't blocked by temporary network issues.

### Update Fails

If the update process fails:
- Walgo will display the error
- Deployment will be aborted
- You can fix the issue manually or use `--skip-version-check`

### suiup Not Found

If `suiup` is not installed, you'll see:
```
suiup not found. Please install it first:
curl -sSfL https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh | sh
```

Install suiup and try again:
```bash
curl -sSfL https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh | sh
```

## Best Practices

1. **Keep tools updated**: Always use the latest versions for mainnet deployments
2. **Test on testnet first**: Before deploying to mainnet, test your site on testnet
3. **Check release notes**: Review release notes before updating to understand changes
4. **Backup important sites**: Keep backups of your site Object IDs before major updates

## Security Considerations

- Version information is fetched from official GitHub repositories
- Updates are installed using the official `suiup` tool
- No credentials or sensitive data is transmitted during version checks
- All network requests use HTTPS

## FAQ

**Q: Why only check versions for mainnet?**
A: Mainnet deployments should use stable, production-ready versions. Testnet is for experimentation and can use any version.

**Q: Can I skip the version check?**
A: Yes, use `--skip-version-check`, but this is not recommended for mainnet.

**Q: How often should I update?**
A: Check for updates before each mainnet deployment. Walgo will notify you if updates are available.

**Q: What if I want to use a specific version?**
A: Use `suiup` to install the specific version you need, then use `--skip-version-check` when deploying.

**Q: Does this affect HTTP deployments?**
A: This feature only affects on-chain deployments via `walgo deploy`. HTTP deployments via `walgo deploy-http` are not affected.

## Related Commands

- `walgo doctor` - Check your environment and tool versions
- `walgo setup-deps` - Install Walrus dependencies
- `sui client switch` - Switch between networks
