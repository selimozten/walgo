# Walgo Uninstall Guide

This guide explains how to completely remove Walgo from your system.

## ⚠️ Important: Your Blockchain Data is Safe!

**Uninstalling Walgo, Sui CLI, or site-builder will NOT delete:**

✅ **Your SUI balance** - Safe on the Sui blockchain
✅ **Your deployed sites** - Safe on Walrus
✅ **Your NFTs and objects** - Safe on the blockchain
✅ **Your on-chain data** - Completely safe

These live on the **Sui blockchain**, not on your computer. The CLI tools are just interfaces to interact with the blockchain.

### What About My Wallet?

Your **wallet keys** are stored locally in `~/.sui/sui_config/`.

⚠️ **Walgo uninstall does NOT delete your Sui wallet directory.**

However, if you manually delete `~/.sui/` later:

- You'll lose access to your funds
- Unless you have your seed phrase backed up

**Always backup before deleting:**

```bash
# Export your wallet keys
sui keytool export --key-identity <your-address>

# Or save your seed phrase
sui keytool list
```

## Quick Uninstall

The easiest way to uninstall Walgo is using the built-in uninstall command:

```bash
walgo uninstall
```

This will interactively guide you through the uninstall process.

## Uninstall Options

### Interactive Mode (Default)

```bash
walgo uninstall
```

The command will ask you what to remove:

1. CLI only
2. Desktop app only
3. Both CLI and Desktop app
4. Cancel

### Non-Interactive Mode

For automated scripts or CI/CD:

```bash
# Uninstall everything without prompts
walgo uninstall --all --force

# Uninstall only CLI
walgo uninstall --cli --force

# Uninstall only desktop app
walgo uninstall --desktop --force

# Uninstall but keep cache and data files
walgo uninstall --all --keep-cache --force
```

## Command Flags

| Flag           | Short | Description                        |
| -------------- | ----- | ---------------------------------- |
| `--all`        | `-a`  | Uninstall both CLI and desktop app |
| `--cli`        |       | Uninstall CLI only                 |
| `--desktop`    |       | Uninstall desktop app only         |
| `--force`      | `-f`  | Skip confirmation prompts          |
| `--keep-cache` |       | Keep cache and data files          |

## What Gets Removed vs What Stays

### ✅ What Walgo Uninstall DOES NOT Remove

**Sui Wallet Directory (`~/.sui/`):**

- Your wallet private keys
- Your seed phrase configuration
- Your account settings
- **Walgo will NEVER delete this directory**

**On Sui Blockchain:**

- Your SUI balance
- Your deployed Walrus sites
- Your NFTs and objects
- Your transaction history
- Any on-chain data

### ❌ What Walgo Uninstall DOES Remove

#### CLI Binary

**Location varies by installation method:**

**System-wide installation:**

- `/usr/local/bin/walgo`

**User installation:**

- `~/.local/bin/walgo`

**Custom installation:**

- Whatever location you specified

#### Desktop App

**macOS:**

- `/Applications/Walgo.app` or `/Applications/walgo-desktop.app`
- `~/Applications/Walgo.app` (if installed to user Applications)

**Windows:**

- `%LOCALAPPDATA%\Programs\Walgo\walgo-desktop.exe`
- Start Menu shortcut (if exists)

**Linux:**

- `~/.local/bin/walgo-desktop`
- `~/.local/share/applications/walgo-desktop.desktop`

### Cache and Data Files

Unless `--keep-cache` is specified:

**All Platforms:**

- `~/.walgo/` - Cache database, AI history, project database
- `~/.config/walgo/` - Configuration files

**What's in these directories:**

- `.walgo/cache.db` - File cache database
- `.walgo/projects.db` - Projects database
- `.walgo/ai-history/` - AI conversation history
- `.walgo/ai-credentials.yaml` - AI API credentials
- `.config/walgo/` - Additional config files

## Manual Uninstall

If you can't use the `walgo uninstall` command, you can manually remove Walgo:

### 1. Remove CLI Binary

```bash
# Find walgo location
which walgo

# Remove it (may need sudo)
sudo rm $(which walgo)
```

### 2. Remove Desktop App

**macOS:**

```bash
rm -rf /Applications/Walgo.app
# or
rm -rf ~/Applications/Walgo.app
```

**Windows (PowerShell):**

```powershell
Remove-Item -Path "$env:LOCALAPPDATA\Programs\Walgo" -Recurse -Force
```

**Linux:**

```bash
rm ~/.local/bin/walgo-desktop
rm ~/.local/share/applications/walgo-desktop.desktop
update-desktop-database ~/.local/share/applications/
```

### 3. Remove Data and Cache

```bash
# Remove all walgo data
rm -rf ~/.walgo
rm -rf ~/.config/walgo
```

## Platform-Specific Notes

### macOS

The uninstall command will:

- Remove the .app bundle from Applications
- Clear the quarantine attribute if needed
- Use sudo only if necessary

### Windows

The uninstall command will:

- Remove from `%LOCALAPPDATA%\Programs\Walgo\`
- Clean up Start Menu shortcuts
- Remove user-specific data

### Linux

The uninstall command will:

- Remove binary from `~/.local/bin/`
- Remove .desktop entry
- Update desktop database automatically
- Use sudo only if installed system-wide

## Verification

After uninstalling, verify removal:

```bash
# Check if walgo is removed
which walgo
# Should return: command not found

# Check for desktop app
# macOS
ls /Applications/ | grep -i walgo

# Linux
ls ~/.local/bin/ | grep walgo

# Windows
dir "%LOCALAPPDATA%\Programs\" | findstr Walgo
```

## Keeping Configuration

If you want to keep your configuration for future reinstallation:

```bash
# Uninstall but keep data
walgo uninstall --all --keep-cache --force
```

This preserves:

- AI credentials
- Project database
- Cache database
- AI conversation history

## Before Uninstalling: Backup Your Wallet

**Important:** While Walgo won't delete your wallet, it's good practice to backup:

### 1. Check Your Wallet Address

```bash
sui client active-address
```

### 2. Export Your Private Keys

```bash
# List all your keys
sui keytool list

# Export a specific key (replace with your address)
sui keytool export --key-identity <your-address> --json

# Save the output to a secure file
sui keytool export --key-identity <your-address> > my-wallet-backup.json
```

### 3. Save Your Seed Phrase

If you created your wallet with a seed phrase, make sure you have it written down securely.

```bash
# Show your key information
sui keytool list
```

### 4. Verify Your On-Chain Assets

Before uninstalling, you can verify your assets:

```bash
# Check SUI and WAL token balances
sui client balance

# List all objects
sui client objects

# Check deployed sites (if any)
walgo status
```

## After Uninstall: Restoring Access

If you reinstall Walgo/Sui CLI later:

### Option 1: Wallet Already Exists

If you didn't delete `~/.sui/`, your wallet is automatically available:

```bash
# After reinstalling
sui client active-address
# Your same address appears!
```

### Option 2: Restore from Backup

If you deleted `~/.sui/` but have backups:

```bash
# Restore from seed phrase
sui keytool import <your-seed-phrase> ed25519

# Or restore from private key
sui keytool import <your-private-key> ed25519
```

Your blockchain data (balance, deployed sites) will be immediately accessible!

## Reinstalling

To reinstall Walgo after uninstalling:

```bash
# Using install script
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

# Or using Go
go install github.com/selimozten/walgo@latest

# Or download from releases
# Visit: https://github.com/selimozten/walgo/releases
```

## Troubleshooting

### Permission Denied

If you get permission errors:

```bash
# Use sudo for system-wide installations
sudo walgo uninstall --all --force
```

### Command Not Found

If `walgo uninstall` doesn't work:

1. Try manual uninstall (see above)
2. Or reinstall walgo first, then uninstall:
   ```bash
   curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
   walgo uninstall --all --force
   ```

### Desktop App Won't Uninstall

**macOS:**

```bash
# Force remove with sudo
sudo rm -rf /Applications/Walgo.app
```

**Windows:**

```powershell
# Run as Administrator
Remove-Item -Path "$env:LOCALAPPDATA\Programs\Walgo" -Recurse -Force
```

**Linux:**

```bash
# Remove all possible locations
rm -f ~/.local/bin/walgo-desktop
sudo rm -f /usr/local/bin/walgo-desktop
sudo rm -f /usr/bin/walgo-desktop
```

### Can't Remove Cache

If cache removal fails:

```bash
# Force remove cache
rm -rf ~/.walgo
rm -rf ~/.config/walgo

# On Windows
Remove-Item -Path "$env:USERPROFILE\.walgo" -Recurse -Force
```

## Clean Uninstall Checklist

- [ ] CLI binary removed
- [ ] Desktop app removed
- [ ] Cache directory removed (`~/.walgo`)
- [ ] Config directory removed (`~/.config/walgo`)
- [ ] Command not found when running `walgo`
- [ ] Desktop app not in Applications/Programs

## Feedback

If you're uninstalling because of issues:

- Please report bugs: https://github.com/selimozten/walgo/issues
- Share feedback: We'd love to know how to improve!

## Thank You!

Thank you for using Walgo! We hope to see you again.

If you found Walgo useful, please consider:

- ⭐ Starring the repository
- 📢 Sharing with others
- 💬 Providing feedback
