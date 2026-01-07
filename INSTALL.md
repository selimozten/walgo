# Walgo Installation Guide

Complete installation instructions for all platforms.

## Quick Install

### One-Line Install (Linux/macOS)

**Prerequisite: Install Git first**

Before running the installation script, make sure Git is installed:

```bash
# macOS (using Homebrew)
brew install git

# Ubuntu/Debian
sudo apt update
sudo apt install git

# Fedora
sudo dnf install git

# Arch Linux
sudo pacman -S git

# Verify Git installation
git --version
```

**Install Walgo:**

```bash
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

**⚠️ Windows Users:** The command above will NOT work on Windows PowerShell or CMD.

### Windows PowerShell Install

**Prerequisite: Install Git first**

```powershell
# Using Chocolatey (recommended)
choco install git

# Or using Scoop
scoop install git

# Or download from https://git-scm.com/download/win

# Verify Git installation
git --version
```

**Install Walgo:**

```powershell
irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex
```

Or visit the [Installation Guide](docs/INSTALLATION.md) for manual installation methods.

This script will:

- Detect your OS and architecture
- Download the latest release
- Install to `/usr/local/bin` (or `~/.local/bin` if not root)
- Verify installation
- Check for dependencies

### Manual Installation

#### Download Pre-built Binaries

Visit the [releases page](https://github.com/selimozten/walgo/releases/latest) and download the appropriate binary for your platform:

- **Linux (x64)**: `walgo_VERSION_linux_amd64.tar.gz`
- **Linux (ARM64)**: `walgo_VERSION_linux_arm64.tar.gz`
- **macOS (Intel)**: `walgo_VERSION_darwin_amd64.tar.gz`
- **macOS (Apple Silicon)**: `walgo_VERSION_darwin_arm64.tar.gz`
- **Windows (x64)**: `walgo_VERSION_windows_amd64.zip`

Extract and install:

```bash
# Extract
tar -xzf walgo_*.tar.gz  # Linux/macOS
# or unzip walgo_*.zip   # Windows

# Install
sudo mv walgo /usr/local/bin/  # Linux/macOS
# or add to PATH on Windows

# Verify
walgo --version
```

## Package Manager Installation

### Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap selimozten/tap

# Install walgo
brew install walgo

# Verify
walgo --version
```

### Go Install

If you have Go 1.22+ installed:

```bash
go install github.com/selimozten/walgo@latest

# Verify
walgo --version
```

### Docker

```bash
# Pull the image
docker pull ghcr.io/selimozten/walgo:latest

# Run walgo
docker run --rm ghcr.io/selimozten/walgo:latest --help

# Work with local files
docker run --rm -v $(pwd):/workspace -w /workspace \
  ghcr.io/selimozten/walgo:latest init my-site

# Create an alias for convenience
alias walgo='docker run --rm -v $(pwd):/workspace -w /workspace ghcr.io/selimozten/walgo:latest'
```

### Snap (Linux)

```bash
snap install walgo
```

## Build from Source

Requirements:

- Go 1.22 or later
- Git

```bash
# Clone repository
git clone https://github.com/selimozten/walgo.git
cd walgo

# Build
go build -o walgo main.go

# Install globally
sudo mv walgo /usr/local/bin/

# Verify
walgo --version
```

## Installing Dependencies

Walgo requires additional tools for full functionality:

### Automatic Dependency Installation

```bash
walgo setup-deps --with-site-builder --with-walrus --network testnet
```

This installs:

- `site-builder` - For on-chain deployments
- `walrus` - Walrus CLI tool
- Places binaries in `~/.config/walgo/bin`
- Updates configuration automatically

### Manual Dependency Installation

#### Git (Required)

Git is required for installation and some features.

**macOS (Homebrew):**

```bash
# If Homebrew is not installed, install it first:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Then install Git
brew install git
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt update
sudo apt install git
```

**Linux (Fedora):**

```bash
sudo dnf install git
```

**Linux (Arch):**

```bash
sudo pacman -S git
```

**Windows (Chocolatey):**

```powershell
# If Chocolatey is not installed, run this as Administrator:
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Then install Git
choco install git
```

**Windows (Scoop):**

```powershell
# If Scoop is not installed:
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Then install Git
scoop install git
```

**Other platforms:**
Visit [Git Downloads](https://git-scm.com/downloads)

#### Hugo (Required)

Hugo is required for static site generation.

**macOS (Homebrew):**

```bash
# Install Homebrew if needed (see Git section above)
brew install hugo

# Verify it's the extended version
hugo version
```

**Linux (Ubuntu/Debian):**

```bash
sudo apt update
sudo apt install hugo
```

**Linux (Fedora):**

```bash
sudo dnf install hugo
```

**Linux (Arch):**

```bash
sudo pacman -S hugo
```

**Windows (Chocolatey):**

```powershell
# Install Chocolatey if needed (see Git section above)
choco install hugo-extended
```

**Windows (Scoop):**

```powershell
# Install Scoop if needed (see Git section above)
scoop install hugo-extended
```

**Other platforms:**
Visit [Hugo Installation Guide](https://gohugo.io/installation/)

#### site-builder (For On-chain Deployment)

**macOS Apple Silicon:**

```bash
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-testnet-latest-macos-arm64 -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/
```

**macOS Intel:**

```bash
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-testnet-latest-macos-x86_64 -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/
```

**Linux:**

```bash
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-testnet-latest-ubuntu-x86_64 -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/
```

**Windows:**
Download from: https://storage.googleapis.com/mysten-walrus-binaries/
Add to PATH environment variable.

#### Walrus CLI (Optional)

Follow the same process as site-builder, replacing `site-builder` with `walrus` in the URLs.

#### Sui CLI (For On-chain Deployment)

Follow the official guide: https://docs.sui.io/guides/developer/getting-started/sui-install

## Verify Installation

Run diagnostics to check your environment:

```bash
walgo doctor
```

This will check:

- ✓ Required binaries (hugo, site-builder, walrus, sui)
- ✓ Sui wallet configuration
- ✓ Gas balance
- ✓ Configuration files
- ✓ Common issues

Auto-fix detected issues:

```bash
walgo doctor --fix-all
```

## Post-Installation Setup

### For HTTP Testnet Deployment (No Wallet Required)

No additional setup needed! You can deploy immediately:

```bash
walgo init my-site
cd my-site
walgo build
walgo deploy-http \
  --publisher https://publisher.walrus-testnet.walrus.space \
  --aggregator https://aggregator.walrus-testnet.walrus.space \
  --epochs 1
```

### For On-chain Deployment (Requires Wallet)

1. **Setup Walrus configuration:**

```bash
walgo setup --network testnet --force
```

2. **Fix configuration paths:**

```bash
walgo doctor --fix-paths
```

3. **Configure Sui wallet:**

```bash
sui client
```

4. **Get testnet SUI tokens:**
   Visit: https://faucet.sui.io/ or join Discord

5. **Verify setup:**

```bash
walgo doctor
```

6. **Deploy:**

```bash
walgo deploy --epochs 5
```

## Troubleshooting

### walgo: command not found

Add the installation directory to your PATH:

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:/usr/local/bin"

# Or if installed to user directory
export PATH="$PATH:$HOME/.local/bin"

# Reload shell
source ~/.bashrc  # or source ~/.zshrc
```

### Permission denied during installation

Either:

1. Use sudo: `sudo mv walgo /usr/local/bin/`
2. Install to user directory: `mv walgo ~/.local/bin/`
3. Run the install script without sudo: `USE_SUDO=false ./install.sh`

### Hugo not found

Install Hugo using your package manager or download from:
https://gohugo.io/installation/

### site-builder not found

Run the automated installer:

```bash
walgo setup-deps --with-site-builder
```

Or install manually following the instructions above.

## Updating Walgo

### Via Install Script

```bash
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

### Via Homebrew

```bash
brew upgrade walgo
```

### Via Go

```bash
go install github.com/selimozten/walgo@latest
```

### Check for Updates

```bash
walgo version --check-updates
```

## Uninstalling

### Remove Binary

```bash
sudo rm /usr/local/bin/walgo
# or
rm ~/.local/bin/walgo
```

### Remove Configuration

```bash
rm -rf ~/.config/walgo
rm -f ~/.walgo.yaml
```

### Via Homebrew

```bash
brew uninstall walgo
brew untap selimozten/tap
```

## Getting Help

- **Documentation**: https://github.com/selimozten/walgo
- **Issues**: https://github.com/selimozten/walgo/issues
- **Run diagnostics**: `walgo doctor`
- **Command help**: `walgo --help` or `walgo <command> --help`

## Next Steps

After installation:

1. **Run diagnostics**: `walgo doctor`
2. **Create your first site**: `walgo init my-site`
3. **Build and deploy**: See [README.md](README.md) for deployment options

Happy shipping! 🚀
