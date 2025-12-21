# Walgo

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/selimozten/walgo/actions/workflows/ci.yml/badge.svg)](https://github.com/selimozten/walgo/actions/workflows/ci.yml)
[![Walrus RFP Winner](https://img.shields.io/badge/Walrus%20RFP-Winner-orange.svg)](https://walrus.xyz)

**The official CLI tool for deploying static sites to Walrus decentralized storage.**

Deploy your Hugo sites to the decentralized web in seconds. No blockchain experience required.

## Supported By

<div style="text-align: center; background: white; padding: 20px; border-radius: 10px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
  <img src="walrus-logo.svg" alt="Walrus Logo" style="background: white; padding: 10px; border-radius: 5px; max-width: 150px; height: auto;" />
  <br />
  <p style="margin: 10px 0 0 0; font-weight: bold;">Walrus RFP Winner</p>
  <p style="margin: 5px 0 0 0;">This project won the Walrus RFP and is supported by Walrus to advance the decentralized web ecosystem.</p>
</div>

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash
```

Or build from source:
```bash
go install github.com/selimozten/walgo@latest
```

## Quick Start

### Fastest Way (Recommended)

```bash
# One command to create, configure, and deploy your site!
walgo quickstart my-blog
```

This will:
- ✓ Create a new Hugo site
- ✓ Install a theme (PaperMod)
- ✓ Add sample content
- ✓ Build and optimize
- ✓ Deploy to Walrus (HTTP)

**Result:** Your site live on the decentralized web in ~2 minutes! 🚀

### Manual Setup

```bash
# Create a new site
walgo init my-site
cd my-site

# Build and deploy (no wallet required)
walgo build
walgo deploy-http

# Or deploy on-chain (requires Sui wallet)
walgo setup --network testnet
walgo deploy --epochs 5
```

## Key Features

🚀 **Instant Deployment** - Push to Walrus in seconds, get a permanent URL immediately
💸 **Free Tier Available** - HTTP mode requires no wallet or cryptocurrency
🔄 **Live Updates** - Update your site content without changing the URL
📦 **Asset Optimization** - Automatic minification for faster load times
🧠 **Obsidian Support** - Transform your knowledge base into a website
🛠️ **Developer Friendly** - Simple CLI, clear errors, built-in diagnostics

## Commands

```bash
walgo quickstart <name>  # 🚀 Create, configure, and deploy in one command!
walgo init <name>        # Create new Hugo site
walgo build              # Build the site
walgo deploy             # Deploy on-chain (requires wallet)
walgo deploy-http        # Deploy via HTTP (no wallet needed)
walgo update <id>        # Update existing site
walgo status <id>        # Check site status
walgo import <vault>     # Import Obsidian vault
walgo doctor             # Diagnose setup issues
```

## Configuration

Create `walgo.yaml` in your project:

```yaml
hugo:
  build_draft: false
  minify: true

walrus:
  project_id: ""  # Set after first deploy
  epochs: 5       # Storage duration

optimize:
  html: true
  css: true
  javascript: true
```

## Deploy Options

### HTTP Mode (Quick & Free)
No wallet or funds required - perfect for testing:
```bash
walgo deploy-http
```

### On-Chain Mode (Permanent)
Requires Sui wallet with testnet SUI:
```bash
walgo setup --network testnet
walgo deploy --epochs 5
```

## Requirements

- [Hugo Extended](https://gohugo.io) - Static site generator (**Extended version required** for SCSS/SASS support)
- [site-builder](https://docs.walrus.site/walrus-sites/overview) - For on-chain deployments
- [Sui wallet](https://docs.sui.io/guides/developer/getting-started/sui-install) - For on-chain mode only

> **Note:** Hugo Extended is required. Check with `hugo version` - it should show "extended". Install via:
> - macOS: `brew install hugo`
> - Linux: Download the "extended" version from [Hugo releases](https://github.com/gohugoio/hugo/releases)

## Troubleshooting

```bash
walgo doctor         # Check your setup
walgo doctor -v      # Verbose diagnostics
```

Common issues:
- **"site-builder not found"** → Run `walgo setup` first
- **"insufficient funds"** → Get testnet SUI from [faucet](https://discord.com/channels/916379725201563759/971488439931392130)
- **"network issues"** → Retry with `--epochs 1` for faster deployment

## Development

```bash
git clone https://github.com/selimozten/walgo.git
cd walgo
make build
make test
```

## About

**Walgo** is developed by the [Ganbitera](https://ganbitera.io) team, winners of the Walrus RFP (Request for Proposals) for creating developer tooling for the Walrus ecosystem.

This project is officially supported by [Walrus](https://walrus.xyz), a decentralized storage network built on Sui that enables permanent, censorship-resistant hosting of websites and applications.

### Why Walgo?

- **Simple**: One command to go from Hugo site to live decentralized website
- **Flexible**: Choose between quick HTTP deployments or permanent on-chain storage
- **Powerful**: Built-in optimization, Obsidian support, and seamless updates
- **Official**: Developed through the Walrus RFP program with direct team support

## License

MIT - See [LICENSE](LICENSE)