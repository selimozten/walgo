# Walgo Documentation

Complete documentation for Walgo, the official CLI tool for deploying static sites to Walrus decentralized storage.

## Documentation Index

### For Users

Perfect for getting started and using Walgo for your projects.

| Document | Description | Target Audience |
|----------|-------------|-----------------|
| [Installation Guide](INSTALLATION.md) | Complete installation instructions for all platforms | Everyone |
| [Getting Started](GETTING_STARTED.md) | Your first site deployment walkthrough | New users |
| [Commands Reference](COMMANDS.md) | Detailed documentation of all commands | All users |
| [Configuration Reference](CONFIGURATION.md) | Complete configuration options | Intermediate users |
| [Deployment Guide](DEPLOYMENT.md) | HTTP and on-chain deployment strategies | All users |
| [Troubleshooting Guide](TROUBLESHOOTING.md) | Solutions to common issues | All users |
| [Optimizer Documentation](OPTIMIZER.md) | Asset optimization features | Intermediate users |

### For Developers

Perfect for contributors and those wanting to understand Walgo's internals.

| Document | Description | Target Audience |
|----------|-------------|-----------------|
| [Architecture](ARCHITECTURE.md) | System design and internal structure | Developers |
| [Development Guide](DEVELOPMENT.md) | Setup dev environment and contribute | Contributors |
| [Contributing Guide](CONTRIBUTING.md) | How to contribute to Walgo | Contributors |
| [Labels Guide](LABELS.md) | GitHub issue labels reference | Maintainers |

## Quick Links

### Getting Started
- New to Walgo? Start with [Installation](INSTALLATION.md) → [Getting Started](GETTING_STARTED.md)
- Want to deploy quickly? See [Getting Started: Your First Site](GETTING_STARTED.md#your-first-site)
- Need help? Check [Troubleshooting Guide](TROUBLESHOOTING.md)

### Common Tasks
- [Create a new site](GETTING_STARTED.md#step-1-create-a-new-site)
- [Deploy via HTTP (free)](DEPLOYMENT.md#http-deployment)
- [Deploy on-chain (permanent)](DEPLOYMENT.md#on-chain-deployment)
- [Update existing deployment](DEPLOYMENT.md#updating-deployments)
- [Optimize assets](OPTIMIZER.md)
- [Import from Obsidian](GETTING_STARTED.md#obsidian-to-blog-workflow)
- [Setup custom domain](DEPLOYMENT.md#custom-domains)

### Reference
- [All commands](COMMANDS.md)
- [All configuration options](CONFIGURATION.md)
- [Environment variables](CONFIGURATION.md#environment-variables)

## Documentation Structure

```
docs/
├── README.md                    # This file - documentation index
│
├── User Documentation/
│   ├── INSTALLATION.md          # Installation for all platforms
│   ├── GETTING_STARTED.md       # First deployment walkthrough
│   ├── COMMANDS.md              # Complete command reference
│   ├── CONFIGURATION.md         # Configuration options
│   ├── DEPLOYMENT.md            # Deployment strategies
│   ├── TROUBLESHOOTING.md       # Common issues and solutions
│   └── OPTIMIZER.md             # Asset optimization guide
│
└── Developer Documentation/
    ├── ARCHITECTURE.md          # System architecture + diagrams
    ├── DEVELOPMENT.md           # Development setup and guide
    ├── CONTRIBUTING.md          # Contribution guidelines
    └── LABELS.md                # GitHub labels guide
```

## Learning Paths

### Path 1: Quick Start (15 minutes)

Perfect for: "I just want to deploy my Hugo site"

1. [Install Walgo](INSTALLATION.md#quick-install)
2. [Create your first site](GETTING_STARTED.md#your-first-site)
3. [Deploy via HTTP](DEPLOYMENT.md#http-deployment)

### Path 2: Complete User Guide (1 hour)

Perfect for: "I want to master Walgo"

1. [Installation Guide](INSTALLATION.md) - Set up everything
2. [Getting Started](GETTING_STARTED.md) - First deployment
3. [Configuration Reference](CONFIGURATION.md) - Customize settings
4. [Deployment Guide](DEPLOYMENT.md) - Master both modes
5. [Commands Reference](COMMANDS.md) - Learn all commands

### Path 3: Developer Journey (2-3 hours)

Perfect for: "I want to contribute to Walgo"

1. [Architecture](ARCHITECTURE.md) - Understand the design
2. [Development Guide](DEVELOPMENT.md) - Set up dev environment
3. [Contributing Guide](CONTRIBUTING.md) - Make your first contribution

### Path 4: Troubleshooting

Perfect for: "Something isn't working"

1. [Troubleshooting Guide](TROUBLESHOOTING.md) - Find your issue
2. [Commands Reference](COMMANDS.md) - Verify command usage
3. [Configuration Reference](CONFIGURATION.md) - Check config

## Features by Document

### Installation Guide
- Platform-specific installation (macOS, Linux, Windows)
- Dependency installation
- Building from source
- Troubleshooting installation issues

### Getting Started Guide
- Creating your first site
- Adding content
- Building and previewing
- HTTP vs on-chain deployment
- Complete workflow examples

### Commands Reference
- All 15+ commands documented
- Flags and options
- Usage examples
- Output examples

### Configuration Reference
- All configuration sections (Hugo, Walrus, Optimizer, Obsidian)
- Configuration file format
- Environment variables
- Configuration examples

### Deployment Guide
- HTTP deployment (free, temporary)
- On-chain deployment (permanent)
- Updating deployments
- Custom domains with SuiNS
- Cost optimization
- Best practices

### Troubleshooting Guide
- Installation issues
- Build issues
- Deployment issues
- Optimization issues
- Network issues
- Wallet issues

### Optimizer Documentation
- HTML optimization
- CSS optimization
- JavaScript optimization
- Configuration options
- Performance impact
- Safety and best practices

### Architecture
- High-level overview
- Component diagrams (Mermaid)
- Data flow diagrams
- Design patterns
- Package structure
- Technology stack

### Development Guide
- Development setup
- Building and testing
- Code style standards
- Adding new features
- Testing strategy
- CI/CD pipeline

### Contributing Guide
- Getting started
- Filing issues
- Submitting pull requests
- Coding standards
- Testing guidelines

## Visual Guides

The documentation includes various Mermaid diagrams:

### Architecture Diagrams
- [System Architecture](ARCHITECTURE.md#architecture-diagram)
- [Component Structure](ARCHITECTURE.md#core-components)
- [Deployment Flow](ARCHITECTURE.md#deployment-flow)
- [Data Flow](ARCHITECTURE.md#data-flow)

### Workflow Diagrams
- [Development Cycle](GETTING_STARTED.md#understanding-the-workflow)
- [HTTP Deployment Workflow](DEPLOYMENT.md#http-deployment-workflow)
- [On-Chain Deployment Workflow](DEPLOYMENT.md#on-chain-deployment-workflow)
- [Optimizer Pipeline](ARCHITECTURE.md#optimizer-engine)

## Cheat Sheets

### Essential Commands
```bash
# Site Management
walgo init my-site           # Create new site
walgo build                  # Build site
walgo serve                  # Preview locally

# Deployment
walgo deploy-http            # Deploy via HTTP (free)
walgo deploy --epochs 5      # Deploy on-chain
walgo update <object-id>     # Update existing site
walgo status <object-id>     # Check status

# Utilities
walgo doctor                 # Diagnose issues
walgo optimize               # Optimize assets
walgo setup                  # Configure wallet
```

### Configuration Quick Reference
```yaml
# Minimal walgo.yaml
walrus:
  epochs: 5

optimizer:
  enabled: true

hugo:
  publishDir: "public"
```

### Troubleshooting Quick Reference
```bash
# Common diagnostic commands
walgo version               # Check version
walgo doctor               # Full diagnostics
walgo <command> --verbose  # Verbose output
hugo version               # Check Hugo
```

## Getting Help

### Within Documentation
1. Use the search function (Ctrl/Cmd+F)
2. Check the [Troubleshooting Guide](TROUBLESHOOTING.md)
3. Review relevant command in [Commands Reference](COMMANDS.md)

### External Resources
- **GitHub Issues:** [Report bugs](https://github.com/selimozten/walgo/issues)
- **GitHub Discussions:** [Ask questions](https://github.com/selimozten/walgo/discussions)
- **Walrus Docs:** [Walrus documentation](https://docs.walrus.site)
- **Sui Docs:** [Sui documentation](https://docs.sui.io)

### Command-Line Help
```bash
# Help for any command
walgo --help
walgo <command> --help

# Examples
walgo deploy --help
walgo build --help
```

## Contributing to Documentation

Documentation contributions are welcome! See [Contributing Guide](CONTRIBUTING.md).

### How to Improve Docs

1. **Fix typos or errors**
   - Submit PR with corrections
   - Reference line numbers

2. **Add examples**
   - Real-world use cases
   - Common workflows
   - Code snippets

3. **Improve clarity**
   - Simplify complex explanations
   - Add diagrams
   - Include screenshots

4. **Add missing content**
   - FAQ entries
   - Troubleshooting scenarios
   - Configuration examples

### Documentation Standards

- Use clear, concise language
- Include code examples
- Add mermaid diagrams where helpful
- Link to related documentation
- Test all commands before documenting
- Follow the existing structure and style

## Version Information

- **Current Version:** 1.0.0
- **Last Updated:** January 2025
- **Documentation Version:** Matches Walgo version

## License

Documentation is released under the MIT License, same as Walgo.

---

**Ready to get started?** Jump to [Installation Guide](INSTALLATION.md) →
