# Walgo - Hugo & Walrus Sites Integration CLI

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#)

Walgo is a powerful command-line tool that seamlessly bridges [Hugo](https://gohugo.io/) static site generation with [Walrus Sites](https://docs.walrus.site/walrus-sites/intro.html) decentralized storage. Deploy your Hugo sites to the decentralized web with ease, import content from Obsidian vaults, and manage your entire workflow from a single CLI.

## 🚀 Project Status

✅ **Complete & Production Ready** - All core functionality has been implemented, tested, and is ready for use.

---

## 🌟 Features

### Core Functionality
- ✅ **Hugo Integration**: Initialize, build, and serve Hugo sites with enhanced workflows
- ✅ **Walrus Deployment**: Deploy and update sites on Walrus decentralized storage
- ✅ **Obsidian Import**: Convert Obsidian vaults to Hugo-compatible markdown
- ✅ **SuiNS Domain Management**: Get step-by-step domain configuration guidance
- ✅ **Site Management**: Check status, convert object IDs, and manage resources

### Advanced Features
- 🔧 **Wikilink Conversion**: Automatically convert `[[wikilinks]]` to Hugo-compatible markdown
- 📁 **Asset Management**: Handle images, PDFs, and other attachments seamlessly  
- ⚙️ **Flexible Configuration**: YAML-based configuration with sensible defaults
- 🔄 **Efficient Updates**: Update existing sites without full redeployment
- 📊 **Resource Monitoring**: Check site status and resource utilization
- 🎯 **Multi-format Support**: Support for YAML, TOML, and JSON frontmatter
- ⚡ **Asset Optimization**: Built-in HTML, CSS, and JavaScript minification and optimization

---

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Command Reference](#command-reference)
- [Workflows](#workflows)
- [Advanced Usage](#advanced-usage)
- [Troubleshooting](#troubleshooting)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

---

## 🔧 Prerequisites

### Required Dependencies
- **Go 1.22+** - [Download here](https://golang.org/dl/)
- **Hugo** (latest version) - [Installation guide](https://gohugo.io/installation/)
- **site-builder CLI** - Official Walrus Sites tool

### System Requirements
- **Operating System**: macOS, Linux, or Windows
- **Memory**: 1GB RAM minimum
- **Storage**: 100MB free space for binaries and cache
- **Network**: Internet connection for deployments

---

## 📦 Installation

### Method 1: Build from Source (Recommended)

```bash
# Clone the repository
git clone https://github.com/selimozten/walgo.git
cd walgo

# Build the binary
go build -o walgo main.go

# Install globally (optional)
sudo mv walgo /usr/local/bin/

# Verify installation
walgo --help
```

### Method 2: Go Install

```bash
# Install directly with Go
go install github.com/selimozten/walgo@latest

# Verify installation
walgo --help
```

### Install Required Dependencies

#### Hugo Installation
```bash
# macOS (using Homebrew)
brew install hugo

# Ubuntu/Debian
sudo apt install hugo

# Windows (using Chocolatey)
choco install hugo-extended
```

#### Walrus site-builder Installation
```bash
# macOS Apple Silicon
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-mainnet-latest-macos-arm64 -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/

# macOS Intel
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-mainnet-latest-macos-x86_64 -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/

# Ubuntu/Linux
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-mainnet-latest-ubuntu-x86_64 -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/

# Windows
# Download site-builder-mainnet-latest-windows-x86_64.exe
# Add to your PATH environment variable
```

#### Verify Installation
```bash
hugo version
site-builder --help
walgo --help
```

---

## 🚀 Quick Start

### 1. Create Your First Site

```bash
# Initialize a new Hugo site configured for Walrus
walgo init my-awesome-blog
cd my-awesome-blog

# The directory structure will look like:
# my-awesome-blog/
# ├── content/          # Your markdown content
# ├── static/           # Static assets (images, CSS, etc.)
# ├── themes/           # Hugo themes
# ├── hugo.toml         # Hugo configuration
# └── walgo.yaml        # Walgo configuration
```

### 2. Create Content

```bash
# Create a blog post
walgo new posts/welcome-to-walrus.md

# Create a project page
walgo new projects/my-cool-project.md

# Create content in subdirectories
walgo new blog/tech/hugo-tips.md
```

### 3. Build and Preview

```bash
# Build your site
walgo build

# Preview locally (Hugo development server)
walgo serve
# Visit http://localhost:1313 in your browser

# Build with clean (removes old files first)
walgo build --clean
```

### 4. Deploy to Walrus

```bash
# Deploy to Walrus Sites (stores for 1 epoch by default)
walgo deploy

# Deploy for multiple epochs (longer storage duration)
walgo deploy --epochs 5

# The output will show your site's object ID - save this!
# Example output:
# New site object ID: 0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf
```

### 5. Configure Domain and Update Site

```bash
# Update walgo.yaml with your site's object ID
nano walgo.yaml  # Add the object ID to walrus.projectID

# Get SuiNS domain configuration instructions
walgo domain myblog

# Check your site status
walgo status

# Update your site with new content
walgo update --epochs 3
```

---

## ⚙️ Configuration

### Primary Configuration (`walgo.yaml`)

Walgo uses a YAML configuration file created automatically by `walgo init`:

```yaml
# Hugo-specific settings
hugo:
  version: ""                    # Target Hugo version (optional)
  baseURL: ""                    # Override Hugo's baseURL for deployments
  publishDir: "public"           # Directory containing built site
  contentDir: "content"          # Hugo content directory  
  resourceDir: "resources"       # Hugo resources directory

# Walrus Sites configuration
walrus:
  projectID: "YOUR_WALRUS_PROJECT_ID"  # IMPORTANT: Replace with your site's object ID
  bucketName: ""                       # Optional: specific Walrus bucket
  entrypoint: "index.html"             # Main HTML file
  suinsDomain: ""                      # Optional: your-domain.sui

# Obsidian import settings
obsidian:
  vaultPath: ""                        # Default Obsidian vault path
  attachmentDir: "images"              # Attachment directory in static/
  convertWikilinks: true               # Convert [[wikilinks]] to Hugo format
  includeDrafts: false                 # Include files marked as drafts
  frontmatterFormat: "yaml"            # yaml, toml, or json
```

### Hugo Configuration (`hugo.toml`)

Standard Hugo configuration for your site:

```toml
baseURL = 'https://your-domain.wal.app'
languageCode = 'en-us'
title = 'My Walrus Site'
theme = 'your-theme'

[params]
  description = "A decentralized blog on Walrus Sites"
  author = "Your Name"

[markup]
  [markup.goldmark]
    [markup.goldmark.renderer]
      unsafe = true  # Allow HTML in markdown
```

### Walrus site-builder Configuration

The site-builder tool uses its own configuration at `~/.config/walrus/sites-config.yaml`:

```yaml
package: "0x..."  # Walrus Sites package object ID
rpc_url: "https://fullnode.mainnet.sui.io:443"
wallet: "~/.sui/sui_config/client.yaml"
gas_budget: 50000000
```

---

## 📖 Command Reference

### Core Site Management

#### `walgo init <site-name>`
Initialize a new Hugo site with Walrus configuration.

```bash
walgo init myblog                    # Create new site
walgo init company-site              # Corporate site
walgo init docs --theme docsy        # With specific theme (future feature)
```

#### `walgo build [flags]`
Build your Hugo site for deployment.

```bash
walgo build                          # Standard build
walgo build --clean                  # Clean public/ directory first
walgo build -c                       # Short form of --clean
```

**Flags:**
- `--clean, -c`: Remove public directory before building

#### `walgo serve [flags]`
Start Hugo development server.

```bash
walgo serve                          # Default port (1313)
walgo serve --port 8080             # Custom port
walgo serve -p 8080                 # Short form
walgo serve --drafts                # Include draft content
walgo serve -D                      # Short form for drafts
walgo serve --future                # Include future-dated content
walgo serve -F                      # Short form for future
```

**Flags:**
- `--port, -p int`: Server port (default: Hugo's default)
- `--drafts, -D`: Include draft content
- `--expired, -E`: Include expired content  
- `--future, -F`: Include future-dated content

#### `walgo new <content-path>`
Create new content files.

```bash
walgo new posts/my-post.md           # Blog post
walgo new projects/app.md            # Project page
walgo new docs/api/overview.md       # Documentation
walgo new about.md                   # Top-level page
```

### Walrus Deployment

#### `walgo deploy [flags]`
Deploy a new site to Walrus Sites.

```bash
walgo deploy                         # Deploy for 1 epoch
walgo deploy --epochs 5              # Deploy for 5 epochs  
walgo deploy -e 10                   # Deploy for 10 epochs
```

**Flags:**
- `--epochs, -e int`: Number of epochs to store data (default: 1)
- `--force, -f`: Force deploy even if no changes detected

**Important:** Save the object ID from the deployment output!

#### `walgo update [object-id] [flags]`
Update an existing Walrus Site.

```bash
walgo update                         # Use object ID from walgo.yaml
walgo update 0xe674c14...            # Use specific object ID
walgo update --epochs 3              # Update for 3 epochs
walgo update 0xe674c14... -e 5       # Specific ID and epochs
```

**Flags:**
- `--epochs, -e int`: Number of epochs to store data (default: 1)

### Site Status and Utilities

#### `walgo status [object-id] [flags]`
Check site status and list resources.

```bash
walgo status                         # Use object ID from config
walgo status 0xe674c14...            # Check specific site
walgo status --convert               # Also show Base36 format
walgo status 0xe674c14... -c         # Specific ID with conversion
```

**Flags:**
- `--convert, -c`: Also display Base36 representation

#### `walgo convert <object-id>`
Convert object ID to Base36 format for direct access.

```bash
walgo convert 0xe674c144119a37a0ed9cef26a962c3fdfbdbfd86a3b3db562ee81d5542a4eccf
# Output: Base36 representation: 58gr4pinoayuijgdixud23441t55jd94ugep68fsm72b8mwmq2
```

### Domain Management

#### `walgo domain [domain-name]`
Get instructions for SuiNS domain configuration.

```bash
walgo domain                         # General instructions
walgo domain myblog                  # Instructions for specific domain
walgo domain my-company              # Corporate domain setup
```

### Content Import

#### `walgo import <vault-path> [flags]`
Import content from Obsidian vaults.

```bash
# Basic import
walgo import /path/to/obsidian/vault

# Import to specific subdirectory
walgo import ~/Documents/MyVault --output-dir imported

# Override attachment settings
walgo import ~/Vault --attachment-dir assets

# Different frontmatter format
walgo import ~/Vault --frontmatter-format toml

# Overwrite existing files
walgo import ~/Vault --overwrite

# Don't convert wikilinks
walgo import ~/Vault --convert-wikilinks=false
```

**Flags:**
- `--output-dir, -o string`: Subdirectory in content/ for imported files
- `--overwrite, -f`: Overwrite existing files
- `--convert-wikilinks`: Convert [[wikilinks]] to Hugo format (default: true)
- `--attachment-dir string`: Directory for attachments (relative to static/)
- `--frontmatter-format string`: Frontmatter format (yaml, toml, json)

---

## 🔄 Workflows

### Blog Publishing Workflow

```bash
# 1. Initial setup
walgo init my-blog
cd my-blog

# 2. Create content
walgo new posts/hello-world.md
# Edit content...

# 3. Preview locally
walgo serve
# Check http://localhost:1313

# 4. Deploy
walgo build
walgo deploy --epochs 5

# 5. Configure domain (follow instructions)
walgo domain myblog

# 6. Update workflow
walgo new posts/update-post.md
# Edit content...
walgo build
walgo update --epochs 3
```

### Obsidian to Walgo Migration

```bash
# 1. Setup Walgo site
walgo init my-digital-garden
cd my-digital-garden

# 2. Configure Obsidian settings in walgo.yaml
nano walgo.yaml
# Update obsidian section as needed

# 3. Import vault
walgo import ~/Documents/MyObsidianVault --output-dir notes

# 4. Review and adjust
walgo serve
# Check imported content at http://localhost:1313

# 5. Deploy
walgo build  
walgo deploy --epochs 10
```

### Documentation Site Workflow

```bash
# 1. Initialize docs site
walgo init project-docs
cd project-docs

# 2. Create documentation structure
walgo new docs/getting-started.md
walgo new docs/api/authentication.md
walgo new docs/guides/deployment.md

# 3. Build and deploy
walgo build
walgo deploy --epochs 20  # Longer storage for docs

# 4. Set up domain
walgo domain projectdocs
```

---

## 🔬 Advanced Usage

### Multiple Environment Configuration

Create environment-specific configs:

```bash
# Development
cp walgo.yaml walgo-dev.yaml
# Edit for development settings

# Production  
cp walgo.yaml walgo-prod.yaml
# Edit for production settings

# Use specific config
walgo --config walgo-prod.yaml deploy
```

### Batch Operations

```bash
# Build and deploy in one go
walgo build && walgo deploy --epochs 5

# Update multiple sites (with different configs)
walgo --config site1.yaml update && \
walgo --config site2.yaml update
```

### Custom Hugo Themes

```bash
# Initialize with custom setup
walgo init myblog
cd myblog

# Add theme as submodule
git submodule add https://github.com/theme/repo themes/mytheme

# Update hugo.toml
echo 'theme = "mytheme"' >> hugo.toml

# Build and deploy
walgo build
walgo deploy
```

### Large Site Optimization

```bash
# For large sites, use clean builds
walgo build --clean

# Deploy with longer epochs for stability
walgo deploy --epochs 20

# Monitor resource usage
walgo status --convert
```

### Automation Scripts

Create deployment scripts:

```bash
#!/bin/bash
# deploy.sh
set -e

echo "Building site..."
walgo build --clean

echo "Deploying to Walrus..."
walgo deploy --epochs 5

echo "Checking deployment status..."
walgo status

echo "Deployment complete!"
```

---

## 🐛 Troubleshooting

### Common Issues

#### `site-builder not found`
**Problem:** The site-builder CLI is not installed or not in PATH.

**Solution:**
```bash
# Check if installed
which site-builder

# If not found, install:
curl -L https://storage.googleapis.com/mysten-walrus-binaries/site-builder-mainnet-latest-$(uname -s | tr '[:upper:]' '[:lower:]')-$(uname -m) -o site-builder
chmod +x site-builder
sudo mv site-builder /usr/local/bin/

# Verify
site-builder --help
```

#### `Hugo not found`
**Problem:** Hugo is not installed or not in PATH.

**Solution:**
```bash
# Install Hugo (macOS)
brew install hugo

# Install Hugo (Ubuntu)
sudo apt install hugo

# Verify
hugo version
```

#### `ProjectID not set` Error
**Problem:** The walgo.yaml file doesn't have a valid Walrus project ID.

**Solution:**
```bash
# After first deployment, update walgo.yaml:
nano walgo.yaml
# Set walrus.projectID to your site's object ID from deployment output
```

#### `Public directory not found`
**Problem:** Trying to deploy without building first.

**Solution:**
```bash
# Always build before deploying
walgo build
walgo deploy
```

#### Import Issues
**Problem:** Obsidian import fails or produces unexpected results.

**Solutions:**
```bash
# Check vault path exists
ls -la /path/to/vault

# Use absolute paths
walgo import /full/path/to/vault

# Check permissions
chmod -R 755 /path/to/vault

# Try with overwrite flag
walgo import /path/to/vault --overwrite
```

### Debug Mode

Enable verbose output for debugging:

```bash
# Set debug environment variable
export WALGO_DEBUG=1
walgo deploy

# Or check Hugo's verbose output
hugo --verbose
```

### Log Files

Check logs for more information:

```bash
# Hugo build logs
hugo --logLevel debug

# site-builder logs  
site-builder --help  # Check for log options

# System logs (macOS)
tail -f /var/log/system.log | grep walgo
```

### Performance Issues

If experiencing slow performance:

```bash
# Clean build cache
rm -rf public/ resources/

# Update dependencies
go mod tidy
go build -o walgo main.go

# Check disk space
df -h
```

---

## 🛠️ Development

### Building from Source

```bash
# Clone repository
git clone https://github.com/your-org/walgo.git
cd walgo

# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o walgo main.go

# Run locally
./walgo --help
```

### Development Dependencies

```bash
# Install Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Format code
go fmt ./...

# Update dependencies
go mod tidy
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/obsidian
go test ./internal/walrus

# Run tests verbosely
go test -v ./...

# Run benchmarks
go test -bench=. ./...

# Test coverage by package
go test -cover ./internal/obsidian ./internal/walrus
```

### Project Structure

```
walgo/
├── cmd/                    # CLI command implementations
│   ├── build.go           # Build command
│   ├── deploy.go          # Deploy command  
│   ├── import.go          # Import command
│   ├── serve.go           # Serve command
│   └── ...
├── internal/              # Internal packages
│   ├── config/            # Configuration management
│   ├── hugo/              # Hugo integration
│   ├── obsidian/          # Obsidian import logic
│   └── walrus/            # Walrus Sites integration
├── main.go                # Application entry point
├── go.mod                 # Go module definition
├── go.sum                 # Dependency checksums
└── README.md              # This file
```

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

### Contribution Guidelines

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** with tests
4. **Run tests**: `go test ./...`
5. **Commit changes**: `git commit -m 'Add amazing feature'`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Code Standards

- Follow Go best practices and conventions
- Add tests for new functionality
- Update documentation for new features
- Use meaningful commit messages
- Ensure all tests pass before submitting

### Reporting Issues

When reporting issues, please include:

- Go version (`go version`)
- Operating system and version
- Hugo version (`hugo version`)
- site-builder version (`site-builder --version`)
- Complete error message
- Steps to reproduce
- Expected vs actual behavior

### Feature Requests

We're always interested in new ideas! Please:

- Check existing issues first
- Provide detailed use case descriptions
- Explain how it would benefit other users
- Consider implementation complexity

---

## 📊 Performance & Best Practices

### Site Performance

- **Use `--clean` builds** for production deployments
- **Optimize images** before adding to static/
- **Choose appropriate epochs** based on content update frequency
- **Monitor resource usage** with `walgo status`

### Storage Efficiency

- **Longer epochs for stable content** (documentation, about pages)
- **Shorter epochs for frequently updated content** (blogs, news)
- **Regular status checks** to monitor storage costs

### Development Workflow

- **Use `walgo serve`** for local development
- **Test builds locally** before deployment
- **Version control your walgo.yaml** configuration
- **Backup object IDs** for site recovery

---

## 🔗 Related Projects & Resources

### Official Documentation
- [Hugo Documentation](https://gohugo.io/documentation/)
- [Walrus Sites Documentation](https://docs.walrus.site/walrus-sites/intro.html)
- [SuiNS Documentation](https://suins.io)

### Recommended Hugo Themes
- [PaperMod](https://github.com/adityatelange/hugo-PaperMod) - Fast, clean blog theme
- [Docsy](https://github.com/google/docsy) - Documentation sites
- [Academic](https://github.com/wowchemy/wowchemy-hugo-themes) - Academic and resume sites

### Tools & Integrations
- [Hugo Modules](https://gohugo.io/hugo-modules/) - Modular Hugo development
- [Obsidian](https://obsidian.md/) - Knowledge management with markdown
- [Walrus CLI](https://github.com/MystenLabs/walrus-sites) - Direct Walrus interaction

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- **Hugo Team** - For the amazing static site generator
- **Mysten Labs** - For Walrus decentralized storage
- **Sui Foundation** - For the underlying blockchain infrastructure  
- **Go Community** - For excellent tooling and libraries
- **Contributors** - Everyone who helps improve this project

---

## 📈 Changelog

### v1.0.0 (Current)
- ✅ Complete Hugo integration
- ✅ Full Walrus Sites deployment support
- ✅ Obsidian vault import functionality
- ✅ SuiNS domain configuration guidance
- ✅ Comprehensive CLI with all core features
- ✅ Extensive documentation and examples

### Roadmap
- 🔄 GitHub Actions integration
- 🔄 Multiple theme support
- 🔄 Backup and restore functionality  
- 🔄 Analytics integration
- 🔄 Custom deployment scripts

---

<div align="center">

**Made with ❤️ for the decentralized web**

[Report Bug](https://github.com/your-org/walgo/issues) • [Request Feature](https://github.com/your-org/walgo/issues) • [Documentation](https://github.com/your-org/walgo/wiki)

</div> 