# Walgo Documentation

Complete documentation for Walgo, the official CLI tool for deploying static sites to Walrus decentralized storage.

## Quick Start

**New to Walgo?** Get your site live in minutes:

```bash
# Install
curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

# Create and deploy
walgo quickstart my-blog
cd my-blog
walgo launch
```

---

## Documentation Index

### Getting Started

| Document | Description |
|----------|-------------|
| [Quickstart Guide](QUICKSTART.md) | Deploy your first site in 2 minutes |
| [Installation Guide](INSTALLATION.md) | Platform-specific installation instructions |
| [Getting Started](GETTING_STARTED.md) | Complete first deployment walkthrough |

### User Guides

| Document | Description |
|----------|-------------|
| [Commands Reference](COMMANDS.md) | All 29+ commands with examples |
| [Launch Wizard](LAUNCH_WIZARD.md) | Interactive deployment guide (**recommended**) |
| [Configuration](CONFIGURATION.md) | All configuration options |
| [Deployment Guide](DEPLOYMENT.md) | HTTP vs on-chain deployment strategies |
| [AI Features](AI_FEATURES.md) | AI-powered content generation |
| [Desktop Integration](DESKTOP_INTEGRATION.md) | Desktop application features |
| [Troubleshooting](TROUBLESHOOTING.md) | Solutions to common issues |
| [Uninstall Guide](UNINSTALL.md) | Complete uninstallation instructions |

### Reference

| Document | Description |
|----------|-------------|
| [Optimizer Documentation](OPTIMIZER.md) | Asset optimization features |
| [Blockchain Data FAQ](BLOCKCHAIN_DATA_FAQ.md) | Understanding blockchain data safety |
| [Version Management](VERSION_MANAGEMENT.md) | Version control and release process |

### Developer Documentation

| Document | Description |
|----------|-------------|
| [Architecture](ARCHITECTURE.md) | System design and internal structure |
| [Development Guide](DEVELOPMENT.md) | Setup dev environment and contribute |
| [Contributing Guide](CONTRIBUTING.md) | How to contribute to Walgo |
| [Labels Guide](LABELS.md) | GitHub issue labels reference |

---

## Learning Paths

### Path 1: Instant Deploy (2 minutes)

For users who want a site live immediately:

```bash
walgo quickstart my-blog
cd my-blog
walgo launch
```

**Read:** [Quickstart Guide](QUICKSTART.md)

### Path 2: Standard Workflow (15 minutes)

For users who want to understand the process:

1. [Install Walgo](INSTALLATION.md)
2. Create your site: `walgo init my-site`
3. Build: `walgo build`
4. Deploy: `walgo launch`

**Read:** [Getting Started](GETTING_STARTED.md)

### Path 3: Complete Mastery (1 hour)

For users who want to learn all features:

1. [Installation Guide](INSTALLATION.md)
2. [Getting Started](GETTING_STARTED.md)
3. [Launch Wizard](LAUNCH_WIZARD.md)
4. [Commands Reference](COMMANDS.md)
5. [Configuration](CONFIGURATION.md)

### Path 4: Developer Journey (2-3 hours)

For contributors and maintainers:

1. [Architecture](ARCHITECTURE.md)
2. [Development Guide](DEVELOPMENT.md)
3. [Contributing Guide](CONTRIBUTING.md)

---

## Essential Commands

### Create & Build

```bash
walgo quickstart <name>   # Create, configure, and build a new site
walgo init <name>         # Initialize a new Hugo site
walgo build               # Build site with optimization
walgo serve               # Preview locally
```

### Deploy

```bash
walgo launch              # Interactive deployment wizard (RECOMMENDED)
walgo projects update     # Update existing deployment
```

### Manage Projects

```bash
walgo projects            # List all your deployed sites
walgo projects show       # View project details and metadata
walgo projects edit       # Edit project metadata (name, description, category, image)
walgo projects update     # Update site on-chain
walgo status <id>         # Check deployment status
```

### Utilities

```bash
walgo doctor              # Diagnose issues
walgo optimize            # Optimize assets
walgo ai generate         # Generate content with AI
```

---

## Deployment Methods

### Recommended: `walgo launch`

The **Launch Wizard** is the recommended way to deploy. It provides:

- Step-by-step guidance through the deployment process
- Automatic wallet and network configuration
- Project tracking and management
- Cost estimation before deployment
- SuiNS domain setup instructions

```bash
walgo build
walgo launch
```

### Alternative Methods

| Method | Use Case |
|--------|----------|
| `walgo launch` | Interactive wizard, project tracking (**recommended**) |
| `walgo deploy` | Direct deployment for automation/scripts |
| `walgo deploy-http` | Free testing without wallet (testnet only) |
| `walgo projects update` | Update existing project content on-chain |
| `walgo projects edit` | Edit metadata (name, description, category, image) with optional on-chain update |

---

## Configuration Quick Reference

Minimal `walgo.yaml`:

```yaml
hugo:
  publishDir: "public"

walrus:
  epochs: 5

optimizer:
  enabled: true
```

Full configuration: [Configuration Reference](CONFIGURATION.md)

---

## Getting Help

### Documentation

- Use browser search (Ctrl/Cmd+F) within documents
- Check [Troubleshooting Guide](TROUBLESHOOTING.md)
- Review [Commands Reference](COMMANDS.md)

### Command-Line Help

```bash
walgo --help              # List all commands
walgo <command> --help    # Help for specific command
walgo doctor              # Diagnose environment issues
```

### Community

- **GitHub Issues:** [Report bugs](https://github.com/selimozten/walgo/issues)
- **GitHub Discussions:** [Ask questions](https://github.com/selimozten/walgo/discussions)

### External Resources

- **Walrus Docs:** [docs.walrus.site](https://docs.walrus.site)
- **Sui Docs:** [docs.sui.io](https://docs.sui.io)
- **SuiNS:** [suins.io](https://suins.io)

---

## Contributing to Documentation

Documentation contributions are welcome! See [Contributing Guide](CONTRIBUTING.md).

### How to Improve Docs

1. **Fix typos or errors** — Submit PR with corrections
2. **Add examples** — Real-world use cases and workflows
3. **Improve clarity** — Simplify complex explanations
4. **Add missing content** — FAQ entries, troubleshooting scenarios

### Documentation Standards

- Use clear, concise language
- Include working code examples
- Test all commands before documenting
- Link to related documentation

---

## Version Information

- **Current Version:** 0.3.0
- **Last Updated:** January 2025
- **Documentation Version:** Matches Walgo version

---

## License

Documentation is released under the MIT License, same as Walgo.

---

**Ready to start?** Begin with the [Quickstart Guide](QUICKSTART.md) →
