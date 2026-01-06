# Troubleshooting Guide

Solutions to common issues when using Walgo.

## Table of Contents

- [Installation Issues](#installation-issues)
- [Build Issues](#build-issues)
- [Deployment Issues](#deployment-issues)
- [Optimization Issues](#optimization-issues)
- [Network Issues](#network-issues)
- [Configuration Issues](#configuration-issues)
- [Wallet Issues](#wallet-issues)
- [Getting More Help](#getting-more-help)

## Installation Issues

### "walgo: command not found"

**Symptoms:**

```bash
$ walgo version
bash: walgo: command not found
```

**Solutions:**

1. **Verify Installation:**

   ```bash
   # Check if walgo is installed
   which walgo
   ls -la $(which walgo)
   ```

2. **Add to PATH:**

   ```bash
   # For user installation
   export PATH="$PATH:$HOME/.local/bin"

   # For system installation
   export PATH="$PATH:/usr/local/bin"

   # Make permanent (add to ~/.bashrc or ~/.zshrc)
   echo 'export PATH="$PATH:/usr/local/bin"' >> ~/.bashrc
   source ~/.bashrc
   ```

3. **Reinstall:**

   ```bash
   # Using install script
   curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

   # Or using Go
   go install github.com/selimozten/walgo@latest
   ```

---

### "Permission denied" when installing

**Symptoms:**

```bash
$ sudo mv walgo /usr/local/bin/
mv: cannot move 'walgo' to '/usr/local/bin/walgo': Permission denied
```

**Solutions:**

1. **Use sudo:**

   ```bash
   sudo mv walgo /usr/local/bin/
   ```

2. **Install to user directory (no sudo needed):**

   ```bash
   mkdir -p ~/.local/bin
   mv walgo ~/.local/bin/
   export PATH="$PATH:$HOME/.local/bin"
   ```

3. **Check ownership:**
   ```bash
   ls -la /usr/local/bin/
   sudo chown $(whoami) /usr/local/bin/
   ```

---

### macOS: "walgo cannot be opened because the developer cannot be verified"

**Symptoms:**
macOS security warning when running walgo.

**Solutions:**

1. **Remove quarantine attribute:**

   ```bash
   xattr -d com.apple.quarantine /usr/local/bin/walgo
   ```

2. **Or allow in System Preferences:**
   - System Preferences → Security & Privacy
   - Click "Allow Anyway" next to walgo warning

---

### Windows: "Windows protected your PC" SmartScreen warning

**Symptoms:**
Windows SmartScreen blocks walgo.exe.

**Solutions:**

1. **Click "More info"** → "Run anyway"

2. **Or disable SmartScreen (not recommended):**
   - Windows Security → App & browser control
   - Reputation-based protection settings
   - Turn off "Check apps and files"

---

## Build Issues

### "Hugo not found" error

**Symptoms:**

```bash
$ walgo build
Error: Hugo not found. Please install Hugo first.
```

**Solutions:**

1. **Install Hugo:**

   ```bash
   # macOS
   brew install hugo

   # Linux
   sudo apt install hugo  # Debian/Ubuntu
   sudo dnf install hugo  # Fedora

   # Windows
   scoop install hugo-extended
   ```

2. **Verify installation:**

   ```bash
   hugo version
   ```

3. **Check PATH:**
   ```bash
   which hugo
   export PATH="$PATH:/usr/local/bin"
   ```

---

### Build fails with theme errors

**Symptoms:**

```bash
$ walgo build
Error: Unable to locate theme directory: themes/PaperMod
```

**Solutions:**

1. **Initialize and update Git submodules:**

   ```bash
   git submodule update --init --recursive
   ```

2. **Re-add theme:**

   ```bash
   git submodule add --force https://github.com/adityatelange/hugo-PaperMod.git themes/PaperMod
   ```

3. **Verify theme in config:**
   ```toml
   # hugo.toml
   theme = "PaperMod"  # Must match directory name
   ```

---

### Build fails with "config file not found"

**Symptoms:**

```bash
$ walgo build
Error: Unable to locate config file
```

**Solutions:**

1. **Check for config file:**

   ```bash
   ls -la | grep -E "hugo\.(toml|yaml|json)|config\.(toml|yaml|json)"
   ```

2. **Create config if missing:**

   ```bash
   # Create hugo.toml
   echo 'baseURL = "http://localhost:1313/"
   languageCode = "en-us"
   title = "My Site"' > hugo.toml
   ```

3. **Specify config explicitly:**
   ```bash
   walgo build --config hugo.toml
   ```

---

### Build produces empty public/ directory

**Symptoms:**
Build succeeds but `public/` is empty or has minimal files.

**Solutions:**

1. **Check for content:**

   ```bash
   ls -la content/
   ```

2. **Create content:**

   ```bash
   hugo new posts/hello.md
   # Edit content/posts/hello.md and set draft: false
   ```

3. **Build with drafts (for testing):**
   ```bash
   walgo build --draft
   ```

---

## Deployment Issues

### "Insufficient funds" error (on-chain)

**Symptoms:**

```bash
$ walgo deploy
Error: Insufficient funds. Required: 0.15 SUI, Available: 0.05 SUI
```

**Solutions:**

1. **Check balance:**

   ```bash
   walgo doctor
   ```

2. **Get testnet SUI:**

   - Visit [Sui Discord](https://discord.com/channels/916379725201563759/971488439931392130)
   - Request: `!faucet <your-address>`
   - Wait ~30 seconds

3. **Reduce epochs:**

   ```bash
   walgo deploy --epochs 1  # Minimum cost
   ```

4. **For mainnet, buy SUI:**
   - Use exchanges (Binance, Coinbase, etc.)
   - Transfer to your wallet address

---

### "site-builder not found" error

**Symptoms:**

```bash
$ walgo deploy
Error: site-builder not found. Please install it first.
```

**Solutions:**

1. **Install site-builder:**

   ```bash
   walgo setup-deps --site-builder
   ```

2. **Verify installation:**

   ```bash
   which site-builder
   site-builder --version
   ```

3. **Manual installation:**
   ```bash
   curl -fsSL https://docs.walrus.site/install.sh | bash
   ```

---

### HTTP deployment fails with network error

**Symptoms:**

```bash
$ walgo deploy-http
Error: Failed to connect to publisher
```

**Solutions:**

1. **Check internet connection:**

   ```bash
   ping 8.8.8.8
   curl -I https://walrus.site
   ```

2. **Run diagnostics:**

   ```bash
   walgo doctor
   ```

3. **Try again with verbose output:**

   ```bash
   walgo deploy-http --verbose
   ```

4. **Use custom endpoints:**
   ```bash
   walgo deploy-http --publisher https://publisher.walrus.site
   ```

---

### Deployment succeeds but site doesn't load

**Symptoms:**
Site deployed successfully but URL shows error or blank page.

**Solutions:**

1. **Wait a moment:**

   - Propagation takes 10-30 seconds
   - Try refreshing after waiting

2. **Check deployment status:**

   ```bash
   walgo status <object-id>
   ```

3. **Test index.html:**

   ```bash
   ls -la public/index.html
   cat public/index.html | head -20
   ```

4. **Clear browser cache:**

   - Hard reload: `Ctrl+F5` (Windows) or `Cmd+Shift+R` (Mac)

5. **Check browser console:**
   - F12 → Console tab
   - Look for errors

---

### "Object not found" when updating

**Symptoms:**

```bash
$ walgo update 0x7b5a...
Error: Site object not found
```

**Solutions:**

1. **Verify object ID:**

   ```bash
   # Check saved object ID
   cat walgo.yaml | grep projectID
   ```

2. **Check correct network:**

   ```bash
   # If deployed on testnet
   walgo update 0x7b5a... --network testnet

   # If deployed on mainnet
   walgo update 0x7b5a... --network mainnet
   ```

3. **Verify object exists:**
   ```bash
   walgo status 0x7b5a... --network testnet
   ```

---

## Optimization Issues

### JavaScript breaks after optimization

**Symptoms:**
Site works before optimization but JavaScript errors after.

**Solutions:**

1. **Disable JS obfuscation:**

   ```yaml
   # walgo.yaml
   optimizer:
     js:
       obfuscate: false # Set to false
   ```

2. **Build without optimization:**

   ```bash
   walgo build --no-optimize
   ```

3. **Skip specific files:**

   ```yaml
   optimizer:
     skipPatterns:
       - "problematic-script.js"
       - "vendor/*.js"
   ```

4. **Test optimization separately:**

   ```bash
   # Build without optimization
   walgo build --no-optimize

   # Test optimization manually
   walgo optimize public/ --verbose
   ```

---

### CSS styling broken after optimization

**Symptoms:**
Site looks different after optimization, styles missing.

**Solutions:**

1. **Disable unused CSS removal:**

   ```yaml
   # walgo.yaml
   optimizer:
     css:
       removeUnused: false # Very important!
   ```

2. **Check for dynamic classes:**

   - If site uses JavaScript to add classes
   - If using frameworks (React, Vue)
   - Don't use `removeUnused: true`

3. **Compare before/after:**

   ```bash
   # Build without optimization
   walgo build --no-optimize
   cp -r public public-before

   # Build with optimization
   walgo build

   # Diff the CSS files
   diff public-before/style.css public/style.css
   ```

---

### Optimization too aggressive

**Symptoms:**
Site broken after optimization, but unsure which setting caused it.

**Solutions:**

1. **Disable optimization entirely:**

   ```yaml
   optimizer:
     enabled: false
   ```

2. **Enable selectively:**

   ```yaml
   optimizer:
     enabled: true
     html:
       enabled: true # Start with HTML only
     css:
       enabled: false
     js:
       enabled: false
   ```

3. **Test each optimizer:**

   ```bash
   # Test HTML optimization
   walgo optimize --css=false --js=false

   # Test CSS optimization
   walgo optimize --html=false --js=false

   # Test JS optimization
   walgo optimize --html=false --css=false
   ```

---

## Network Issues

### "Cannot connect to Walrus" error

**Symptoms:**

```bash
Error: Failed to connect to Walrus publisher/aggregator
```

**Solutions:**

1. **Check internet connection:**

   ```bash
   ping 8.8.8.8
   curl -I https://walrus.site
   ```

2. **Test Walrus endpoints:**

   ```bash
   curl -I https://publisher.walrus.site
   curl -I https://aggregator.walrus.site
   ```

3. **Run diagnostics:**

   ```bash
   walgo doctor -v
   ```

4. **Check firewall:**
   - Ensure HTTPS (port 443) is not blocked
   - Check corporate firewall/proxy settings

---

### "RPC error" when deploying

**Symptoms:**

```bash
Error: Failed to connect to Sui RPC node
```

**Solutions:**

1. **Check Sui network status:**

   - Visit [Sui status page](https://status.sui.io)

2. **Try different RPC:**

   ```bash
   # Configure custom RPC (if supported)
   export SUI_RPC_URL="https://fullnode.testnet.sui.io:443"
   ```

3. **Wait and retry:**
   - Network congestion is temporary
   - Try again in a few minutes

---

## Configuration Issues

### "Invalid configuration file" error

**Symptoms:**

```bash
$ walgo build
Error: Failed to parse walgo.yaml: yaml: line 5: mapping values are not allowed
```

**Solutions:**

1. **Validate YAML syntax:**

   ```bash
   # Use online validator: https://www.yamllint.com/
   # Or install yamllint
   yamllint walgo.yaml
   ```

2. **Check indentation:**

   ```yaml
   # BAD - mixed spaces/tabs
   optimizer:
   	enabled: true  # Tab here
     html:          # Spaces here

   # GOOD - consistent spaces
   optimizer:
     enabled: true
     html:
       enabled: true
   ```

3. **Check for special characters:**

   ```yaml
   # BAD - unquoted special chars
   title: My Site: The Best One

   # GOOD - quoted
   title: "My Site: The Best One"
   ```

---

### Configuration not being applied

**Symptoms:**
Changes to `walgo.yaml` don't seem to take effect.

**Solutions:**

1. **Check file location:**

   ```bash
   # Walgo searches in this order:
   # 1. --config flag path
   # 2. ./walgo.yaml
   # 3. ~/.walgo.yaml

   # Verify which is being used
   walgo doctor -v
   ```

2. **Specify config explicitly:**

   ```bash
   walgo build --config ./walgo.yaml
   ```

3. **Check for typos:**

   ```yaml
   # BAD - typo
   optimiser: # British spelling
     enabled: true

   # GOOD
   optimizer: # American spelling
     enabled: true
   ```

---

### Environment variables not working

**Symptoms:**
Environment variables don't override config file.

**Solutions:**

1. **Check variable format:**

   ```bash
   # WRONG
   export WALGO_EPOCHS=5

   # CORRECT
   export WALGO_WALRUS_EPOCHS=5
   # Format: WALGO_<SECTION>_<KEY>
   ```

2. **Verify variable is set:**

   ```bash
   echo $WALGO_WALRUS_EPOCHS
   env | grep WALGO
   ```

3. **Check boolean values:**
   ```bash
   # Accepted: true, false, 1, 0, yes, no
   export WALGO_OPTIMIZER_ENABLED=true  # Correct
   export WALGO_OPTIMIZER_ENABLED=True  # Also works
   ```

---

## Wallet Issues

### "Wallet not configured" error

**Symptoms:**

```bash
$ walgo deploy
Error: Wallet not configured. Run 'walgo setup' first.
```

**Solutions:**

1. **Run setup:**

   ```bash
   walgo setup --network testnet
   ```

2. **Verify wallet:**

   ```bash
   walgo doctor
   ```

3. **Check Sui CLI:**
   ```bash
   sui client active-address
   ```

---

### "Invalid recovery phrase" when importing wallet

**Symptoms:**
Recovery phrase rejected during wallet import.

**Solutions:**

1. **Check word count:**

   - Must be 12 or 24 words
   - Verify no typos

2. **Check word list:**

   - Words must be from BIP39 word list
   - Use correct spelling

3. **Create new wallet instead:**
   ```bash
   walgo setup --network testnet
   # This creates a new wallet with new phrase
   ```

---

### "Transaction failed" error

**Symptoms:**

```bash
Error: Transaction failed with error: ...
```

**Solutions:**

1. **Check balance:**

   ```bash
   walgo doctor
   ```

2. **Increase gas budget:**

   ```bash
   walgo deploy --gas-budget 200000000
   ```

3. **Wait and retry:**

   - Network congestion
   - Try again in a few minutes

4. **Check transaction on explorer:**
   - Visit Sui Explorer
   - Search for transaction ID
   - See detailed error

---

## Obsidian Import Issues

### "Vault not found" error

**Symptoms:**

```bash
$ walgo import --vault ~/Documents/Vault
Error: Vault directory not found
```

**Solutions:**

1. **Check path:**

   ```bash
   ls -la ~/Documents/Vault

   # Use absolute path
   walgo import --vault /Users/yourname/Documents/Vault
   ```

2. **Check permissions:**
   ```bash
   ls -ld ~/Documents/Vault
   # Should show read permissions
   ```

---

### Wikilinks not converting

**Symptoms:**
Imported files still contain `[[wikilinks]]` instead of markdown links.

**Solutions:**

1. **Enable conversion:**

   ```yaml
   # walgo.yaml
   obsidian:
     convertWikilinks: true
   ```

2. **Or use flag:**
   ```bash
   walgo import --vault ~/Vault --convert-wikilinks
   ```

---

### Attachments not copying

**Symptoms:**
Images/PDFs missing after import.

**Solutions:**

1. **Check attachment directory:**

   ```yaml
   # walgo.yaml
   obsidian:
     attachmentDir: "static/images" # Must exist in Hugo
   ```

2. **Create directory:**

   ```bash
   mkdir -p static/images
   ```

3. **Check file paths in Obsidian:**
   - Must use relative paths
   - Must be in vault directory

---

## Getting More Help

### Using walgo doctor

```bash
# Basic diagnostics
walgo doctor

# Verbose output
walgo doctor -v

# Auto-fix issues
walgo doctor --fix
```

### Checking Logs

```bash
# Enable verbose output
walgo <command> --verbose

# Or set environment variable
export WALGO_VERBOSE=true
walgo <command>
```

### Common Diagnostic Commands

```bash
# Check installation
walgo version
which walgo

# Check dependencies
hugo version
site-builder --version
sui --version

# Check configuration
walgo doctor -v

# Check wallet
sui client active-address
sui client balance

# Check network
ping walrus.site
curl -I https://walrus.site
```

### Reporting Issues

When reporting issues on GitHub, include:

1. **Walgo version:**

   ```bash
   walgo version
   ```

2. **System info:**

   ```bash
   uname -a  # macOS/Linux
   # or
   systeminfo  # Windows
   ```

3. **Command and error:**

   ```bash
   # Exact command you ran
   walgo deploy --epochs 5

   # Full error output
   Error: ...
   ```

4. **Verbose output:**

   ```bash
   walgo <command> --verbose
   ```

5. **Configuration (without secrets):**
   ```yaml
   # Contents of walgo.yaml (remove any API keys)
   ```

### Support Channels

- **Documentation:** [https://github.com/selimozten/walgo/tree/main/docs](https://github.com/selimozten/walgo/tree/main/docs)
- **GitHub Issues:** [https://github.com/selimozten/walgo/issues](https://github.com/selimozten/walgo/issues)
- **GitHub Discussions:** [https://github.com/selimozten/walgo/discussions](https://github.com/selimozten/walgo/discussions)
- **Walrus Discord:** [https://discord.gg/walrus](https://discord.gg/walrus)
- **Sui Discord:** [https://discord.gg/sui](https://discord.gg/sui)

### Before Asking for Help

1. **Search existing issues:**

   - Someone may have had the same problem
   - Solution might already exist

2. **Read documentation:**

   - Check relevant guides
   - Review configuration reference

3. **Run diagnostics:**

   ```bash
   walgo doctor -v
   ```

4. **Try verbose mode:**

   ```bash
   walgo <command> --verbose
   ```

5. **Isolate the issue:**
   - Test with minimal example
   - Disable optimizations
   - Try HTTP deployment first

## Related Documentation

- [Getting Started Guide](GETTING_STARTED.md)
- [Installation Guide](INSTALLATION.md)
- [Configuration Reference](CONFIGURATION.md)
- [Deployment Guide](DEPLOYMENT.md)
- [Commands Reference](COMMANDS.md)
