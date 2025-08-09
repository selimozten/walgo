# v0.1.0 – Initial Release

**🎉 Welcome to Walgo - the seamless bridge between Hugo and Walrus Sites!**

This is the initial public release of Walgo, a powerful CLI tool that makes it incredibly easy to build Hugo static sites and deploy them to the decentralized Walrus Sites storage network.

## 🌟 What is Walgo?

Walgo provides a complete toolkit for:
- **Hugo Integration**: Initialize, build, and serve Hugo sites with enhanced workflows
- **Walrus Deployment**: Deploy and update sites on decentralized storage with a single command
- **Obsidian Import**: Seamlessly migrate your Obsidian knowledge base to Hugo
- **Domain Management**: Get step-by-step guidance for SuiNS domain configuration
- **Asset Optimization**: Built-in minification for HTML, CSS, and JavaScript

## ✨ Key Features

### 🚀 **One-Command Deployment**
```bash
walgo init my-blog
cd my-blog
walgo new posts/hello-world.md
walgo build
walgo deploy --epochs 5
```

### 🔄 **Efficient Updates**
```bash
# Make changes to your content
walgo build
walgo update --epochs 3
```

### 📝 **Obsidian Integration**
```bash
walgo import ~/Documents/MyVault --convert-wikilinks
```

### 🌐 **Domain Configuration**
```bash
walgo domain myblog  # Get step-by-step SuiNS setup instructions
```

## 🛠️ Installation

### Quick Install (Recommended)
```bash
go install github.com/selimozten/walgo@latest
```

### From Source
```bash
git clone https://github.com/selimozten/walgo.git
cd walgo
go build -o walgo main.go
```

### Docker
```bash
docker build -t walgo .
docker run --rm -v $(pwd):/workspace -w /workspace walgo init my-site
```

## 📋 Complete Feature Set

### Core Commands
- `walgo init` - Initialize Hugo sites with Walrus configuration
- `walgo new` - Create new content with proper frontmatter
- `walgo build` - Build sites with optional asset optimization
- `walgo serve` - Local development server with live reload
- `walgo deploy` - Deploy new sites to Walrus Sites
- `walgo update` - Update existing sites efficiently

### Management Commands  
- `walgo status` - Check site resources and storage status
- `walgo convert` - Convert object IDs to Base36 format
- `walgo domain` - SuiNS domain configuration guidance
- `walgo setup` - Configure site-builder for Walrus deployment

### Import & Optimization
- `walgo import` - Import Obsidian vaults with wikilink conversion
- `walgo optimize` - Standalone asset optimization

## 🔧 Prerequisites

- **Go 1.22+** for installation from source
- **Hugo** (latest version) for site generation
- **site-builder CLI** for Walrus Sites deployment
- **SUI wallet** with tokens for deployment costs

## 🏗️ Architecture

Walgo is built with:
- **Go 1.22** for cross-platform compatibility
- **Cobra CLI** for elegant command structure
- **Viper** for flexible configuration management
- **Hugo integration** via command-line interface
- **Walrus Sites** via official site-builder tool

## 🎯 Use Cases

### Perfect for:
- **Personal Blogs**: Deploy your thoughts to the decentralized web
- **Documentation Sites**: Censorship-resistant project documentation  
- **Digital Gardens**: Migrate Obsidian knowledge bases to web format
- **Portfolio Sites**: Showcase your work on permanent storage
- **Community Sites**: Build sites that can't be taken down

### Migration Support:
- **Obsidian → Hugo**: Complete vault migration with asset handling
- **Existing Hugo Sites**: Easy integration with minimal configuration
- **Multi-site Management**: Handle multiple projects with ease

## 📊 Performance & Optimization

- **Built-in Asset Optimization**: HTML/CSS/JS minification
- **Configurable Optimization**: Fine-tune compression settings
- **Efficient Updates**: Only upload changed content
- **Cross-platform Builds**: Support for Linux, macOS, Windows
- **Container Support**: Docker image for consistent environments

## 🤝 Community & Support

- **Documentation**: Comprehensive README with examples
- **Issue Templates**: Structured bug reports and feature requests
- **Contributing Guide**: Clear guidelines for contributors
- **CI/CD Pipeline**: Automated testing and releases
- **Man Pages**: Full command reference documentation

## 🔗 Quick Links

- **Documentation**: [README.md](https://github.com/selimozten/walgo/blob/main/README.md)
- **Examples**: [examples/](https://github.com/selimozten/walgo/tree/main/examples)
- **Contributing**: [docs/CONTRIBUTING.md](https://github.com/selimozten/walgo/blob/main/docs/CONTRIBUTING.md)
- **Issue Tracker**: [GitHub Issues](https://github.com/selimozten/walgo/issues)

## 🏷️ Technical Details

**Supported Platforms:**
- Linux (amd64, arm64)
- macOS (amd64, arm64) 
- Windows (amd64)

**Configuration:**
- YAML-based configuration (`walgo.yaml`)
- Hugo integration via `hugo.toml`
- Walrus Sites via `sites-config.yaml`

**Dependencies:**
- Zero runtime dependencies (statically linked binary)
- External tools: Hugo, site-builder (for deployment)

## 🎊 What's Next?

This initial release provides a complete foundation for Hugo + Walrus Sites workflows. Future enhancements planned:

- GitHub Actions integration for automated deployments
- Multiple Hugo theme support and management
- Enhanced backup and restore functionality
- Analytics integration for decentralized sites
- Custom deployment scripts and workflows

## 🙏 Acknowledgments

Special thanks to:
- **Hugo Team** for the incredible static site generator
- **Mysten Labs** for Walrus decentralized storage
- **Sui Foundation** for the underlying blockchain infrastructure
- **Open Source Community** for tools and inspiration

---

**Ready to build on the decentralized web?** 

```bash
go install github.com/selimozten/walgo@latest
walgo --help
```

Happy building! 🚀 