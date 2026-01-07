# Installation Guide

Complete installation guide for Walgo on all supported platforms.

## Table of Contents

- [System Requirements](#system-requirements)
- [Quick Install](#quick-install)
- [Platform-Specific Installation](#platform-specific-installation)
- [Installing Dependencies](#installing-dependencies)
- [Building from Source](#building-from-source)
- [Verifying Installation](#verifying-installation)
- [Updating Walgo](#updating-walgo)
- [Uninstallation](#uninstallation)

## System Requirements

### Minimum Requirements

- **Operating System:** Linux (amd64/arm64), macOS (Intel/Apple Silicon), or Windows 10+
- **RAM:** 512 MB minimum
- **Disk Space:** 50 MB for Walgo + space for Hugo sites
- **Internet:** Required for deployment and downloads

### Required Dependencies

| Dependency       | Required For        | Installation       |
| ---------------- | ------------------- | ------------------ |
| **Hugo**         | Building sites      | All modes          |
| **site-builder** | On-chain deployment | On-chain mode only |
| **Sui CLI**      | Wallet management   | On-chain mode only |

## Quick Install

### Prerequisite: Install Git

Before running the installation script, make sure Git is installed on your system.

#### macOS (using Homebrew)

```bash
# Install Homebrew if not already installed
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Install Git
brew install git

# Verify installation
git --version
```

#### Linux (Ubuntu/Debian)

```bash
sudo apt update
sudo apt install git

# Verify installation
git --version
```

#### Linux (Fedora)

```bash
sudo dnf install git

# Verify installation
git --version
```

#### Linux (Arch)

```bash
sudo pacman -S git

# Verify installation
git --version
```

#### Windows (Chocolatey)

```powershell
# Install Chocolatey if not already installed (run as Administrator)
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install Git
choco install git

# Verify installation
git --version
```

#### Windows (Scoop)

```powershell
# Install Scoop if not already installed
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Install Git
scoop install git

# Verify installation
git --version
```

#### Other Platforms

Download Git from: https://git-scm.com/downloads

### One-Line Install (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

**⚠️ Windows Users:** The one-line install above will NOT work on Windows PowerShell or CMD. See [Windows Installation](#windows) below for proper installation instructions.

This script will:

1. Detect your operating system and architecture
2. Download the appropriate binary
3. Install to `/usr/local/bin/walgo` (or `~/.local/bin/walgo`)
4. Make the binary executable

### Using Go Install

If you have Go 1.22+ installed:

```bash
go install github.com/selimozten/walgo@latest
```

Make sure `$GOPATH/bin` is in your `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

## Platform-Specific Installation

### macOS

#### Option 1: Homebrew (Recommended)

```bash
# Coming soon - for now use one of the methods below
```

#### Option 2: Download Binary

```bash
# For Apple Silicon (M1/M2/M3)
curl -L -o walgo https://github.com/selimozten/walgo/releases/latest/download/walgo-darwin-arm64
chmod +x walgo
sudo mv walgo /usr/local/bin/

# For Intel Macs
curl -L -o walgo https://github.com/selimozten/walgo/releases/latest/download/walgo-darwin-amd64
chmod +x walgo
sudo mv walgo /usr/local/bin/
```

#### Option 3: Build from Source

```bash
git clone https://github.com/selimozten/walgo.git
cd walgo
make build
sudo mv walgo /usr/local/bin/
```

### Linux

#### Option 1: Download Binary

```bash
# For x86_64
curl -L -o walgo https://github.com/selimozten/walgo/releases/latest/download/walgo-linux-amd64
chmod +x walgo
sudo mv walgo /usr/local/bin/

# For ARM64
curl -L -o walgo https://github.com/selimozten/walgo/releases/latest/download/walgo-linux-arm64
chmod +x walgo
sudo mv walgo /usr/local/bin/
```

#### Option 2: Using Package Managers

```bash
# Debian/Ubuntu (coming soon)
# sudo apt install walgo

# Arch Linux (coming soon)
# yay -S walgo

# For now, use the binary download method above
```

#### Option 3: Build from Source

```bash
# Install Go 1.22+
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Build Walgo
git clone https://github.com/selimozten/walgo.git
cd walgo
make build
sudo mv walgo /usr/local/bin/
```

### Windows

#### Prerequisite: Install Git

Before installing Walgo, make sure Git is installed on Windows.

**Option A: Using Chocolatey (Recommended)**

```powershell
# Install Chocolatey if not already installed (run as Administrator)
Set-ExecutionPolicy Bypass -Scope Process -Force
[System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install Git
choco install git

# Verify installation
git --version
```

**Option B: Using Scoop**

```powershell
# Install Scoop if not already installed
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Install Git
scoop install git

# Verify installation
git --version
```

**Option C: Download Installer**

Download and install Git from: https://git-scm.com/download/win

#### Option 1: PowerShell One-Line Install (Recommended)

Open PowerShell and run:

```powershell
irm https://raw.githubusercontent.com/selimozten/walgo/main/install.ps1 | iex
```

This script will:

- Detect your Windows architecture (x64 or ARM64)
- Download the latest Walgo binary
- Install to a location in your PATH
- Verify installation
- Check for Git installation and provide guidance if missing

**Note:** Run PowerShell as Administrator for system-wide installation, or as regular user for user-only installation.

#### Option 2: Download Binary (Manual)

1. Download the latest release from the [GitHub Releases page](https://github.com/selimozten/walgo/releases/latest):

   - For 64-bit Windows: `walgo-windows-amd64.exe`
   - For ARM64 Windows: `walgo-windows-arm64.exe`

2. Download directly via PowerShell:

```powershell
# For 64-bit
Invoke-WebRequest -Uri "https://github.com/selimozten/walgo/releases/latest/download/walgo-windows-amd64.exe" -OutFile "walgo.exe"

# Or for ARM64
Invoke-WebRequest -Uri "https://github.com/selimozten/walgo/releases/latest/download/walgo-windows-arm64.exe" -OutFile "walgo.exe"
```

3. Add to PATH:

**Method A: Temporary PATH (Current Session Only)**

```powershell
$env:Path += ";$PWD"
```

**Method B: Permanent PATH (System-wide, requires Admin)**

```powershell
# Create installation directory
New-Item -ItemType Directory -Force -Path "C:\Program Files\walgo"

# Move binary
Move-Item walgo.exe "C:\Program Files\walgo\"

# Add to PATH (requires Administrator PowerShell)
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";C:\Program Files\walgo", [System.EnvironmentVariableTarget]::Machine)
```

**Method C: Permanent PATH (User-level)**

```powershell
# Create user directory
$walgoDir = "$env:USERPROFILE\walgo"
New-Item -ItemType Directory -Force -Path $walgoDir

# Move binary
Move-Item walgo.exe "$walgoDir\"

# Add to user PATH
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$walgoDir", [System.EnvironmentVariableTarget]::User)
```

4. **Important:** Close and reopen your terminal (or restart your computer) for PATH changes to take effect.

#### Option 3: Windows Subsystem for Linux (WSL)

If you have WSL installed, you can use the Linux installation:

```bash
# In WSL terminal
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

Note: WSL must have a Linux distribution installed. Install WSL if needed:

```powershell
wsl --install
```

#### Option 4: Using Scoop (Package Manager)

```powershell
# Install Scoop if not already installed
# Coming Soom
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression

# Install Walgo via Scoop
scoop install walgo
```

#### Option 5: Using Chocolatey (Package Manager)

```powershell
# Coming Soom
# Install Chocolatey if not already installed
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))

# Install Walgo via Chocolatey
choco install walgo
```

#### Option 6: Build from Source

```powershell
# Prerequisites: Install Go 1.22+ from https://go.dev/dl/

# Build Walgo
git clone https://github.com/selimozten/walgo.git
cd walgo
go build -o walgo.exe main.go

# Add to PATH (see Option 2 for PATH methods)
```

#### Windows Installation Troubleshooting

**"execution of scripts is disabled" Error**

If you see a PowerShell execution policy error:

```powershell
# Check current policy
Get-ExecutionPolicy

# Allow script execution for current user
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or run the script bypassing policy (not recommended for security)
powershell -ExecutionPolicy Bypass -File install.ps1
```

**Windows Defender SmartScreen Warning**

If Windows Defender blocks the executable:

1. Click "More info"
2. Click "Run anyway"
3. This occurs because the binary isn't signed. It's safe to run - you can verify the source code on GitHub.

**PATH Changes Not Working**

After adding to PATH:

1. Close ALL terminal windows
2. Open a new terminal
3. Test: `walgo --version`
4. If still not found, restart your computer

**Check Your Windows Architecture**

```powershell
# Check if 64-bit or ARM64
$env:PROCESSOR_ARCHITECTURE

# Output will be:
# AMD64 → use amd64 binary
# ARM64 → use arm64 binary
```

## Installing Dependencies

### Hugo

Walgo requires Hugo to build static sites.

#### macOS

```bash
# Using Homebrew
brew install hugo

# Using MacPorts
sudo port install hugo
```

#### Linux

```bash
# Debian/Ubuntu
sudo apt install hugo

# Fedora
sudo dnf install hugo

# Arch Linux
sudo pacman -S hugo

# Or download binary
wget https://github.com/gohugoio/hugo/releases/latest/download/hugo_extended_0.125.0_Linux-64bit.tar.gz
tar -xzf hugo_extended_0.125.0_Linux-64bit.tar.gz
sudo mv hugo /usr/local/bin/
```

#### Windows

```powershell
# Using Scoop
scoop install hugo-extended

# Using Chocolatey
choco install hugo-extended
```

### Walrus site-builder (For On-Chain Deployment)

Only needed if you want to deploy sites on-chain.

```bash
# Let Walgo install it for you
walgo setup-deps

# Or install manually
curl -fsSL https://docs.walrus.site/install.sh | bash
```

### Sui CLI (For On-Chain Deployment)

Only needed if you want to deploy sites on-chain.

#### macOS/Linux

```bash
# Let Walgo install it for you
walgo setup-deps

# Or install manually
cargo install --locked --git https://github.com/MystenLabs/sui.git --branch testnet sui
```

#### Windows

```powershell
# Install Rust first
# Download from https://rustup.rs/

# Install Sui CLI
cargo install --locked --git https://github.com/MystenLabs/sui.git --branch testnet sui
```

## Building from Source

### Prerequisites

- Go 1.22 or later
- Git
- Make (optional, but recommended)

### Steps

1. **Clone the repository:**

   ```bash
   git clone https://github.com/selimozten/walgo.git
   cd walgo
   ```

2. **Install Go dependencies:**

   ```bash
   go mod download
   ```

3. **Build the binary:**

   ```bash
   # Using Make (recommended)
   make build

   # Or using Go directly
   go build -o walgo main.go
   ```

4. **Install the binary:**

   ```bash
   # macOS/Linux
   sudo mv walgo /usr/local/bin/

   # Or install to user directory (no sudo needed)
   mkdir -p ~/.local/bin
   mv walgo ~/.local/bin/
   export PATH="$HOME/.local/bin:$PATH"

   # Windows
   Move-Item walgo.exe C:\Program Files\walgo\
   ```

5. **Verify installation:**

   ```bash
   walgo version
   ```

### Build Options

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build with specific version
make build VERSION=1.0.0

# Run tests before building
make test

# Build and install
make install
```

## Verifying Installation

### Check Walgo Version

```bash
walgo version
```

Expected output:

```
Walgo version 1.0.0
Built with Go 1.24.0
```

### Check Dependencies

```bash
walgo doctor
```

This will check:

- Hugo installation and version
- site-builder availability (for on-chain)
- Sui CLI availability (for on-chain)
- Wallet configuration
- Network connectivity

Expected output:

```
Checking Walgo Installation
==========================

✓ Walgo is installed correctly
✓ Hugo is installed (v0.125.0)
✓ site-builder is installed (v0.1.0)
✓ Sui CLI is installed (v1.20.0)

All dependencies are correctly installed!
```

### Run a Test Build

```bash
# Create a test site
walgo init test-site
cd test-site

# Build the site
walgo build

# Check if build succeeded
ls public/
```

## Updating Walgo

### Update via Script

```bash
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

### Update via Go

```bash
go install github.com/selimozten/walgo@latest
```

### Manual Update

1. Download the latest release for your platform
2. Replace the existing binary:

   ```bash
   # macOS/Linux
   sudo mv walgo /usr/local/bin/walgo

   # Windows
   Move-Item -Force walgo.exe C:\Program Files\walgo\walgo.exe
   ```

### Check for Updates

```bash
walgo version
# Compare with latest release at:
# https://github.com/selimozten/walgo/releases/latest
```

## Uninstallation

### Remove Walgo Binary

```bash
# macOS/Linux
sudo rm /usr/local/bin/walgo
# or
rm ~/.local/bin/walgo

# Windows
Remove-Item "C:\Program Files\walgo\walgo.exe"
```

### Remove Configuration Files

```bash
# Remove global config
rm ~/.walgo.yaml

# Remove project configs (in each project)
rm walgo.yaml
```

### Remove Dependencies (Optional)

Only remove if you're not using them for other projects:

```bash
# Remove Hugo
# macOS
brew uninstall hugo

# Linux
sudo apt remove hugo

# Remove site-builder
rm $(which site-builder)

# Remove Sui CLI
rm $(which sui)
```

## Troubleshooting Installation

### "walgo: command not found"

**Solution:** Add Walgo to your PATH:

```bash
# macOS/Linux - add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:/usr/local/bin"
# or for user install
export PATH="$PATH:$HOME/.local/bin"

# Windows - add to system PATH in Environment Variables
```

### "Permission denied" when installing

**Solution:** Use `sudo` or install to user directory:

```bash
# Option 1: Use sudo
sudo mv walgo /usr/local/bin/

# Option 2: Install to user directory (no sudo needed)
mkdir -p ~/.local/bin
mv walgo ~/.local/bin/
export PATH="$PATH:$HOME/.local/bin"
```

### "Hugo not found" error

**Solution:** Install Hugo:

```bash
# Check if Hugo is installed
which hugo

# If not found, install Hugo (see Hugo section above)
brew install hugo  # macOS
sudo apt install hugo  # Linux
```

### Binary won't execute on macOS

**Solution:** Remove quarantine attribute:

```bash
xattr -d com.apple.quarantine /usr/local/bin/walgo
```

### Windows SmartScreen warning

**Solution:** Click "More info" → "Run anyway"

This occurs because the binary isn't signed. It's safe to run.

### "Cannot find module" error when building from source

**Solution:** Update Go dependencies:

```bash
go mod download
go mod tidy
```

### Different architecture error

**Solution:** Download the correct binary for your system:

```bash
# Check your architecture
uname -m

# x86_64 or amd64 → use amd64 binary
# aarch64 or arm64 → use arm64 binary
```

## Next Steps

After installation:

1. Read the [Getting Started Guide](GETTING_STARTED.md)
2. Review the [Configuration Reference](CONFIGURATION.md)
3. Try deploying your first site with [Deployment Guide](DEPLOYMENT.md)

## Getting Help

- **Documentation:** [https://github.com/selimozten/walgo/tree/main/docs](https://github.com/selimozten/walgo/tree/main/docs)
- **Issues:** [https://github.com/selimozten/walgo/issues](https://github.com/selimozten/walgo/issues)
- **Discussions:** [https://github.com/selimozten/walgo/discussions](https://github.com/selimozten/walgo/discussions)
