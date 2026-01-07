# Walgo - Complete Command Reference

**Walgo v0.3.1** - Deploy Hugo Sites to Walrus Decentralized Storage

Complete reference for all Walgo commands with examples, flags, and workflows.

---

## Recommended Workflow

```bash
# 1. Create your site
walgo quickstart my-blog
cd my-blog

# 2. Build and deploy with the interactive wizard
walgo build
walgo launch    # <-- Recommended deployment method
```

> **Note:** `walgo launch` is the recommended way to deploy. It provides step-by-step guidance, project tracking, and cost estimation.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Site Management](#site-management)
- [Content Management](#content-management)
- [Build & Optimization](#build--optimization)
- [Deployment](#deployment)
- [Project Management](#project-management)
- [AI Features](#ai-features)
- [Desktop App](#desktop-app)
- [Setup & Configuration](#setup--configuration)
- [Diagnostics & Utilities](#diagnostics--utilities)
- [Common Workflows](#common-workflows)
- [Global Flags](#global-flags)

---

## Quick Start

### `walgo quickstart <site-name>`

**One-command setup: initialize and build your site**

```bash
walgo quickstart my-blog
walgo quickstart my-portfolio --skip-build
```

**What it does:**

1. Creates new Hugo site
2. Adds default theme (Ananke)
3. Creates sample content
4. Builds the site
5. Offers deployment options

**Flags:**

- `--skip-build` - Skip the build step

**Next steps after quickstart:**

1. `cd my-blog`
2. `walgo serve` - Preview locally
3. `walgo launch` - Deploy with wizard

---

### `walgo init <site-name>`

**Initialize a new Hugo site**

```bash
walgo init my-blog
walgo init my-site --format yaml
```

**What it does:**

- Creates Hugo directory structure
- Initializes configuration file
- Sets up default archetypes

**Flags:**

- `--format <format>` - Config format: `toml`, `yaml`, or `json` (default: `toml`)
- `--force` - Overwrite existing directory

---

## Site Management

### `walgo new <path>`

**Create new content in Hugo site**

```bash
walgo new posts/my-first-post.md
walgo new about.md
walgo new blog/tutorials/getting-started.md
```

**What it does:**

- Creates markdown file with frontmatter
- Uses Hugo archetypes
- Places in `content/` directory

---

### `walgo serve`

**Start local development server**

```bash
walgo serve
walgo serve -D              # Include drafts
walgo serve --port 8080     # Custom port
walgo serve --bind 0.0.0.0  # Access from network
```

**Flags:**

- `-D, --drafts` - Include draft content
- `-p, --port <port>` - Custom port (default: 1313)
- `--bind <address>` - Interface to bind to (default: 127.0.0.1)
- `--navigate-to-changed` - Navigate to changed file
- `--no-live-reload` - Disable live reload

**Output:**

```
Starting Hugo development server...
✓ Server running at http://localhost:1313/

Press Ctrl+C to stop
```

---

## Content Management

### `walgo import <vault-path>`

**Create a new Hugo site and import content from Obsidian vault**

```bash
# Basic import - creates site with vault name
walgo import ~/Documents/MyVault

# Custom site name
walgo import ~/Notes --site-name my-knowledge-base

# Specify parent directory for site creation
walgo import ~/Vault --parent-dir ~/Projects --site-name notes-site

# With additional options
walgo import ~/Vault --attachment-dir attachments --output-dir docs
```

**What it does:**

1. **Creates a new Hugo site** with walgo.yaml configuration
2. **Initializes Hugo** project structure
3. **Imports Obsidian content:**
   - Converts notes to Hugo format
   - Handles wikilinks → markdown links
   - Copies attachments (images, files)
   - Preserves frontmatter
4. **Creates draft project** for tracking

**Flags:**

- `-n, --site-name <name>` - Name for the new site (defaults to vault directory name)
- `-p, --parent-dir <dir>` - Parent directory for site creation (defaults to current directory)
- `-o, --output-dir <dir>` - Subdirectory in content for imported files (optional)
- `--attachment-dir <dir>` - Where to place attachments (default: `attachments`)
- `--convert-wikilinks` - Convert `[[links]]` to markdown (default: true)
- `--frontmatter-format <format>` - Format: `yaml`, `toml`, or `json` (default: `yaml`)
- `--link-style <style>` - Link conversion: `markdown` or `relref` (default: `markdown`)
- `--dry-run` - Preview import without creating site

**Example:**

```bash
# Import vault and create site in current directory
walgo import ~/Documents/MyVault

# Result: Creates ./MyVault/ with imported content

# Custom setup
walgo import ~/Notes \
  --site-name my-blog \
  --parent-dir ~/Projects \
  --attachment-dir images

# Result: Creates ~/Projects/my-blog/ with imported content
```

---

## Build & Optimization

### `walgo build`

**Build Hugo site with optimization**

```bash
walgo build
walgo build --clean
walgo build --no-optimize --no-compress
walgo build --destination dist
walgo build --base-url https://example.walrus.site/
```

**What it does:**

1. Runs Hugo build → `public/`
2. Optimizes HTML/CSS/JS (optional)
3. Brotli compression (optional)
4. Generates ws-resources.json
5. **Offers interactive menu:**
   - Preview site locally
   - Launch deployment wizard
   - Exit

**Flags:**

- `-c, --clean` - Clean public/ before build
- `--no-optimize` - Skip optimization
- `--no-compress` - Skip compression
- `-v, --verbose` - Show detailed stats
- `-q, --quiet` - Suppress output
- `--draft` - Include draft content
- `--source <dir>` - Source directory (default: current)
- `--destination <dir>` - Output directory (default: `public`)
- `--base-url <url>` - Override baseURL
- `--minify` - Enable Hugo's built-in minification

**Output Example:**

```
Building Hugo site...
✓ Hugo build completed (234 files generated)

Optimizing assets...
🎯 Optimization Results
======================
Files processed: 234
Files optimized: 89
Bytes saved: 1.2 MB (34.5%)
Duration: 456ms

✓ Build completed successfully!

What would you like to do?
  1) Preview site locally
  2) Launch deployment wizard
  3) Exit
```

---

### `walgo optimize <directory>`

**Optimize HTML, CSS, and JavaScript files**

```bash
walgo optimize public/
walgo optimize dist/ --aggressive
walgo optimize --css=false --js=false  # HTML only
walgo optimize --remove-unused-css     # Aggressive CSS
walgo optimize --verbose
```

**What it does:**

- Minifies HTML (removes whitespace, comments)
- Minifies CSS (removes unused rules)
- Minifies JavaScript (removes comments, whitespace)
- Shows compression statistics

**Flags:**

- `--aggressive` - More aggressive optimization
- `--html` - Enable HTML optimization (default: true)
- `--css` - Enable CSS optimization (default: true)
- `--js` - Enable JavaScript optimization (default: true)
- `--remove-unused-css` - Remove unused CSS rules
- `--verbose` / `-v` - Show detailed output

---

### `walgo compress <directory>`

**Compress files with Brotli**

```bash
walgo compress public/
walgo compress dist/ --level 11
walgo compress --verbose
```

**What it does:**

- Creates Brotli compressed files (.br)
- Skips already compressed files
- Shows compression statistics

**Flags:**

- `--level <0-11>` - Compression level (default: 6)
- `--verbose` - Detailed statistics

---

## Deployment

### `walgo launch` (Recommended)

**Interactive deployment wizard** — The recommended way to deploy your site.

```bash
walgo launch
```

**Why use `walgo launch`?**

- Step-by-step guidance through the entire process
- Automatic project tracking and history
- Cost estimation before you confirm
- Wallet management built-in
- SuiNS domain setup instructions

**What it does - 8 Steps:**

1. **Choose Network** - testnet or mainnet
2. **Wallet & Balance Check**
   - View current addresses
   - Switch address
   - Create new address
   - Import existing address
3. **Project Name & Category**
4. **Storage Duration** (epochs)
5. **Verify Site Built**
6. **Review & Confirm** (shows gas estimate)
7. **Deploy to Walrus**
8. **Success!**
   - Shows Object ID
   - Shows URL
   - SuiNS setup instructions

**Features:**

- Saves project to database
- Tracks deployment history
- User-friendly prompts
- Complete wallet management
- Gas estimation
- SuiNS configuration guide ([tutorial](https://docs.wal.app/docs/walrus-sites/tutorial-suins))

---

### `walgo deploy` (Advanced)

**Direct on-chain deployment** — For automation, scripts, or advanced users.

```bash
walgo deploy --epochs 1
walgo deploy --epochs 10 --network mainnet
walgo deploy --gas-budget 100000000
walgo deploy --directory dist
```

**What it does:**

- Deploys `public/` to Walrus
- Creates site object on Sui blockchain
- Returns Object ID and URLs

**Flags:**

- `--epochs <number>` - Storage duration (required, default: 5)
- `--network <network>` - `testnet` or `mainnet` (default: testnet)
- `--wallet <path>` - Sui wallet address
- `--gas-budget <amount>` - Maximum gas to spend (default: auto)
- `--directory <dir>` - Directory to deploy (default: `public`)

**Requirements:**

- Sui wallet with SUI tokens
- site-builder installed
- Built site in public/

> **Tip:** For most users, `walgo launch` provides a better experience with project tracking and guided deployment.

**Output Example:**

```
Deploying to Walrus (on-chain)...
✓ Assets uploaded to Walrus
✓ Site object created on Sui
✓ Deployment successful

🌐 Your site is live!
Object ID: 0x7b5a...8f3c
URL: https://5tphzvq5shsxzugrz7kqd5bhnbajqfamvtxrn8jbfm3jbibzz1.walrus.site

📊 Deployment Stats:
Files: 234
Total size: 2.4 MB
Epochs: 5
Cost: 0.15 SUI
Transaction: 0xabc...def
```

---

### `walgo deploy-http` (Testing)

**Deploy via HTTP APIs (testnet only, no wallet needed)**

```bash
walgo deploy-http \
  --publisher https://publisher.walrus-testnet.walrus.space \
  --aggregator https://aggregator.walrus-testnet.walrus.space \
  --epochs 1

# Or with defaults
walgo deploy-http
walgo deploy-http --directory dist
walgo deploy-http --mode files --workers 20
```

**What it does:**

- Uploads to Walrus via HTTP
- No blockchain interaction
- No wallet/tokens needed
- **Testnet only**

**Flags:**

- `--publisher <url>` - Publisher URL (required)
- `--aggregator <url>` - Aggregator URL (required)
- `--epochs <number>` - Storage duration (required)
- `--mode <mode>` - "blobs" or "files" (default: blobs)
- `--workers <number>` - Parallel uploads (default: 10)
- `--directory <dir>` - Directory to deploy (default: `public`)

**Limitations:**

- Temporary (~30 days)
- Cannot update after deployment
- Cannot use SuiNS domains
- Testnet only

---

### `walgo update <object-id>`

**Update existing Walrus site**

```bash
walgo update 0x1234567890abcdef... --epochs 5
walgo update 0x7b5a...8f3c --directory dist
walgo update 0x7b5a...8f3c --network mainnet
```

**What it does:**

- Updates existing site with new content
- Creates new site object
- Same deployment process as deploy

**Flags:**

- `--epochs <number>` - Storage duration
- `--directory <dir>` - Directory to deploy (default: `public`)
- `--network <network>` - `testnet` or `mainnet`
- `--gas-budget <amount>` - Maximum gas to spend

---

## Project Management

### `walgo projects list`

**List all your deployed projects**

```bash
walgo projects list
walgo projects list --network mainnet
walgo projects list --status active
walgo projects list --status archived
```

**What it shows:**

- Project name & category
- Network (testnet/mainnet)
- Object ID
- SuiNS domain
- Deployment count
- Last deployment date
- Status (active/archived)

**Flags:**

- `--network <network>` - Filter by network
- `--status <status>` - Filter by status (active/archived)

---

### `walgo projects show <name>`

**Show detailed project information**

```bash
walgo projects show "My Blog"
walgo projects show 1  # By ID
```

**What it shows:**

- Complete project details
- All metadata (name, category, description, image URL)
- Network and status
- Object ID and SuiNS domain
- Wallet address
- Site path
- Deployment history
- Success/failure stats
- Gas fees spent
- Access URLs

**Output Example:**

```
╔═══════════════════════════════════════════════════════════╗
║  📦 My Blog                                               ║
╚═══════════════════════════════════════════════════════════╝

ℹ️ Project Information
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  ID:              1
  Name:            My Blog
  Category:        blog
  Description:     A blog about tech and decentralization
  Image URL:       https://example.com/logo.png
  Network:         testnet
  Status:          active
  Object ID:       0x7b5a...8f3c
  SuiNS:           myblog.sui
  URL:             https://myblog.walrus.site
  Wallet:          0xabc...def
  Site path:       /Users/me/projects/my-blog
  Created:         2024-01-15 10:30:00

ℹ️ Statistics
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Total deployments:     5
  Successful:            5
  First deployment:      2024-01-15 10:30
  Last deployment:       2024-01-20 14:15
```

---

### `walgo projects update <name>`

**Redeploy existing project with new content and/or metadata**

```bash
walgo projects update "My Blog"
walgo projects update "My Blog" --epochs 10
walgo projects update "My Blog" --name "New Name" --category blog
walgo projects update "My Blog" --description "My awesome blog" --force
```

**What it does:**

- Updates project metadata in database
- Updates ws-resources.json with new metadata
- Redeploys site content to Walrus (when needed)
- Preserves project history
- Tracks new deployment

**Flags:**

- `--epochs <number>` - Storage duration
- `--name <name>` - Update project name
- `--category <category>` - Update project category
- `--description <text>` - Update project description
- `--image-url <url>` - Update project image/logo URL
- `--suins <domain>` - Update SuiNS domain
- `--force` - Force on-chain update even if only metadata changed

**Behavior:**

- If only metadata flags are provided → updates database and ws-resources.json only
- If `--force` or `--epochs` is provided → also updates on-chain
- If no flags provided → assumes content changed, updates on-chain

---

### `walgo projects edit <name>`

**Edit project metadata with optional on-chain update**

```bash
walgo projects edit "My Blog" --name "New Blog Name"
walgo projects edit "My Blog" --description "A blog about tech"
walgo projects edit "My Blog" --category portfolio --image-url "https://example.com/logo.png"
walgo projects edit "My Blog" --name "New Name" --apply  # Also update on-chain
```

**What it does (3-step process):**

1. **Database update** - Always saves changes to local SQLite database
2. **ws-resources.json update** - Always updates the publish directory's metadata file
3. **On-chain update** - Only with `--apply` flag, deploys changes to Walrus

**Flags:**

- `--name <name>` - New project name (also used as site_name in ws-resources.json)
- `--category <category>` - New project category
- `--description <text>` - New project description
- `--image-url <url>` - New image/logo URL for the site
- `--suins <domain>` - New SuiNS domain
- `--apply` - Apply changes on-chain (update the site on Walrus)

**Examples:**

```bash
# Edit metadata only (database + ws-resources.json)
walgo projects edit mysite --name "New Name" --category blog

# Edit description
walgo projects edit mysite --description "My awesome decentralized website"

# Change image URL
walgo projects edit mysite --image-url "https://example.com/new-logo.png"

# Edit and immediately apply on-chain
walgo projects edit mysite --name "New Name" --description "Updated site" --apply
```

**Output Example:**

```
✏️ Editing project: My Blog

Changes to apply:
  ✏️ Name: My Blog → New Blog Name
  ✏️ Description: (empty) → A blog about tech and decentralization

💾 Step 1/3: Updating database...
  ✓ Database updated
📄 Step 2/3: Updating ws-resources.json...
  ✓ ws-resources.json updated
ℹ️ Step 3/3: Skipped on-chain update (use --apply to update on Walrus)

✓ Project metadata updated successfully!

💡 To apply changes on-chain, run:
   walgo projects edit "New Blog Name" --apply
```

---

### `walgo projects archive <name>`

**Archive a project (hide from default list)**

```bash
walgo projects archive "Old Blog"
```

**What it does:**

- Marks project as archived
- Hides from default list (use `--status archived` to view)
- Preserves all data

---

### `walgo projects delete <name>`

**Permanently delete project record**

```bash
walgo projects delete "Test Site"
```

**What it does:**

- Deletes from database
- Removes deployment history
- **Does NOT delete from Walrus**
- Cannot be undone

**Warning:** Prompts for confirmation

---

## AI Features

### `walgo ai configure`

**Set up AI provider credentials**

```bash
walgo ai configure
```

**What it does:**

- Interactive setup wizard
- Choose provider (OpenAI/OpenRouter)
- Enter API key
- Optional custom base URL
- Saves to ~/.walgo/ai-credentials.yaml
- Updates walgo.yaml with AI settings

**Supported Providers:**

- OpenAI (GPT-3.5, GPT-4)
- OpenRouter (Multiple models)

---

### `walgo ai generate`

**Generate new content with AI**

```bash
walgo ai generate
walgo ai generate --type post
walgo ai generate --type page
```

**What it does:**

- Interactive content generation wizard
- AI creates Hugo markdown with frontmatter
- SEO optimization
- Proper structure and formatting
- Saves to content directory

**Content Types:**

- `post` - Blog posts (with date, tags, categories)
- `page` - Static pages (About, Contact, etc.)

**Example Session:**

```
Content type: Blog post
Description: Write about deploying Hugo to Walrus
Context: Focus on benefits of decentralized hosting

🔄 Generating content...
✅ Content saved to: content/posts/deploying-to-walrus.md
```

**Features:**

- Hugo-aware prompts
- SEO-optimized titles and descriptions
- Proper frontmatter structure
- Markdown formatting
- Code examples when relevant

---

### `walgo ai update <file>`

**Update existing content with AI**

```bash
walgo ai update content/posts/my-post.md
```

**What it does:**

- Reads existing file
- Applies AI-powered updates
- Preserves original style and tone
- Maintains frontmatter
- Intelligent content modification

**Example:**

```
📄 File: content/posts/my-post.md
Describe changes: Add a section about performance optimization

🔄 Updating content...
✅ File updated!
```

---

### `walgo ai pipeline`

**Generate a complete site using AI (plan + generate)**

```bash
walgo ai pipeline
```

**What it does:**

- Runs full AI content generation pipeline:
  1. Plan: AI creates a site structure plan (JSON)
  2. Generate: AI creates each page sequentially
- Saves plan to `.walgo/plan.json` for resumability
- If interrupted, run `walgo ai resume` to continue
- Applies post-pipeline fixes and validation

**Workflow:**

```
🤖 Enter site details:
   Site type: Blog
   Description: A personal tech blog
   Audience: Developers and tech enthusiasts

📋 Planning structure...
   ✅ Site plan saved to .walgo/plan.json

📝 Generating content...
   📄 Creating: content/posts/first-post.md
   📄 Creating: content/posts/about.md
   📄 Creating: content/contact.md

✨ Applying fixes and validation...
✅ Site generated successfully!
```

**Supported Site Types:**

- Blog - Classic blog with posts and about page
- Portfolio - Project showcase
- Docs - Documentation site
- Business - Business website

---

### `walgo ai plan`

**Create a site plan without generating content**

```bash
walgo ai plan
```

**What it does:**

- Creates a site structure plan using AI
- Saves plan to `.walgo/plan.json`
- Does not generate any content
- Plan can be reviewed before generation
- Use `walgo ai resume` to generate content from plan

**Example:**

```
🤖 Enter site details:
   Site type: Blog
   Description: A travel blog
   Audience: Travel enthusiasts

📋 Creating site plan...
✅ Plan saved to .walgo/plan.json
📂 View your plan:
   cat .walgo/plan.json

🚀 Ready to generate? Run:
   walgo ai resume
```

---

### `walgo ai resume`

**Resume content generation from existing plan**

```bash
walgo ai resume
```

**What it does:**

- Loads existing `.walgo/plan.json`
- Continues generating remaining pages
- Applies post-pipeline fixes and validation
- Creates draft project entry

**Use Case:**
When `walgo ai pipeline` was interrupted (e.g., user stopped it, network issue), you can resume from where it left off.

**Example:**

```
📋 Loading plan from .walgo/plan.json...
   Found 5 pages to generate
   Pages already created: 2
   Pages remaining: 3

📝 Generating remaining content...
   📄 Creating: content/posts/post-3.md
   📄 Creating: content/posts/post-4.md
   📄 Creating: content/posts/post-5.md

✨ Applying fixes and validation...
✅ Site generation complete!
```

---

### `walgo ai fix`

**Fix content files for theme requirements**

```bash
walgo ai fix
walgo ai fix --validate
```

**What it does:**

- Validates and fixes Hugo content files
- Ensures theme-specific requirements are met
- Checks frontmatter syntax and completeness
- Removes duplicate H1 headings
- Sets proper draft status

**Flags:**

- `--validate` - Only validate, don't fix (default: false)

**For Business/Portfolio Sites (Ananke theme):**

- Validates title and description fields
- Removes duplicate H1 headings (title is already in frontmatter)
- Ensures services/projects have date fields (ISO 8601 format)
- Sets draft to false

**Example:**

```
🔍 Validating content files...
   Found 5 markdown files

✅ content/posts/first-post.md - Valid
⚠️  content/about.md - Missing description field
   📝 Adding description...
✅ content/about.md - Fixed

✅ content/services/service-1.md - Missing date field
   📝 Adding date (ISO 8601)...
✅ content/services/service-1.md - Fixed

✅ All content files validated and fixed!
```

**Validation Only:**

```bash
walgo ai fix --validate
```

Reports issues without fixing them.

---

## Desktop App

### `walgo desktop`

**Launch Walgo desktop application**

```bash
# Development mode
walgo desktop
```

**What it does:**

- Launches the Walgo Desktop GUI
- Provides graphical interface for all features:
  - Site Management (Create, Build, Deploy)
  - AI Content Generation (Configure, Generate, Update, Pipeline)
  - Project Management (List, Edit, Delete, Archive)
  - Wallet Integration (Network switching, Account management)
  - System Health (Dependency checking, Auto-install)
- Easier project management
- Visual feedback and progress tracking

**Platform Support:**

- macOS (Universal binary)
- Windows (amd64)
- Linux (amd64)

**Desktop App Features:**

#### Dashboard

- Wallet overview with SUI/WAL balances
- Network switching (testnet/mainnet)
- Account management (create, import, switch)
- Quick action cards for common tasks

#### Create

- QuickStart - Instant site creation
- AI Pipeline - Full AI site creation
- Init Site - Manual Hugo site initialization
- Obsidian Import - Import from vaults

#### Projects

- List all deployed sites
- Search and filter projects
- View project details and deployment history
- Edit project metadata
- Delete/archive projects
- Open project folders

#### AI Config

- Configure AI providers (OpenAI/OpenRouter)
- Manage API keys
- Set custom base URLs
- Model selection

#### Edit

- Edit site content
- Build sites
- Deploy to Walrus

#### System Health

- Check dependency status (Hugo, Sui, Walrus, Site Builder)
- Auto-install missing dependencies
- View network connectivity

See [DESKTOP_INTEGRATION.md](DESKTOP_INTEGRATION.md) for complete documentation.

---

## Setup & Configuration

### `walgo setup`

**Configure site-builder for Walrus Sites**

```bash
walgo setup --network testnet
walgo setup --network mainnet --force
walgo setup --wallet-path ~/.sui/custom-wallet
```

**What it does:**

- Creates sites-config.yaml
- Configures for testnet or mainnet
- Sets wallet and network
- Configures Sui client

**Flags:**

- `--network <network>` - `testnet` or `mainnet` (required)
- `--force` - Overwrite existing config
- `--wallet-path <path>` - Custom wallet path
- `--key-scheme <scheme>` - Key scheme: `ed25519`, `secp256k1` (default: ed25519)

---

### `walgo setup-deps`

**Download and install dependencies**

```bash
walgo setup-deps
walgo setup-deps --version latest
walgo setup-deps --site-builder
walgo setup-deps --hugo --sui
```

**What it does:**

- Downloads site-builder binary
- Downloads walrus binary
- Installs to ~/.walgo/bin/
- Optionally installs Hugo and Sui CLI

**Flags:**

- `--version <version>` - Specific version to install
- `--hugo` - Install Hugo
- `--site-builder` - Install site-builder
- `--sui` - Install Sui CLI
- `--all` - Install all dependencies (default)

---

## Diagnostics & Utilities

### `walgo doctor`

**Diagnose environment and configuration**

```bash
walgo doctor
walgo doctor --fix-paths
walgo doctor --network testnet
walgo doctor --verbose
```

**What it checks:**

- Hugo installation
- Sui CLI installation
- site-builder installation
- walrus installation
- Wallet configuration
- Balance (SUI tokens)
- Network connectivity
- Configuration files
- PATH issues

**Flags:**

- `--fix-paths` - Auto-fix PATH issues
- `--network <network>` - Check network-specific config
- `--verbose` / `-v` - Show detailed diagnostics
- `--fix` - Attempt to fix issues automatically

**Output Example:**

```
Checking Walgo Installation
===========================

✓ Walgo is installed correctly (v0.3.1)
✓ Hugo is installed (v0.125.0)
✓ site-builder is installed (v0.1.0)
✓ Sui CLI is installed (v1.20.0)

Checking Configuration
=====================

✓ Configuration file found: ./walgo.yaml
✓ Configuration is valid

Checking Wallet
===============

✓ Wallet configured for testnet
✓ Wallet address: 0xabc...def
✓ Balance: 5.0 SUI

Checking Network
================

✓ Can reach Walrus publisher
✓ Can reach Walrus aggregator
✓ Can reach Sui RPC

All checks passed! ✅
```

---

### `walgo status <object-id>`

**Check status of deployed Walrus site**

```bash
walgo status 0x1234567890abcdef...
walgo status 0x7b5a...8f3c --network mainnet
walgo status 0x7b5a...8f3c --json
```

**What it shows:**

- Site status (active/expired)
- Access URL
- Network (testnet/mainnet)
- File count and size
- Storage information
- Epoch details
- Cost information

**Flags:**

- `--network <network>` - `testnet` or `mainnet`
- `--json` - Output in JSON format

---

### `walgo domain`

**Get SuiNS domain configuration instructions**

```bash
walgo domain
walgo domain list
walgo domain link myblog.sui 0x7b5a...8f3c
walgo domain unlink myblog.sui
walgo domain info myblog.sui
```

**Subcommands:**

- `list` - List owned domains
- `link <domain> <object-id>` - Link domain to site
- `unlink <domain>` - Unlink domain from site
- `info <domain>` - Show domain information

**What it shows:**

- How to register SuiNS domain
- How to link domain to site
- Access URLs
- Complete setup guide

---

### `walgo version`

**Show version information**

```bash
walgo version
```

**Output:**

```
Walgo version 0.3.1
Built with Go 1.24.4
Commit: 2ecac6b
Build date: 2025-01-15
```

---

### `walgo uninstall`

**Uninstall Walgo CLI and/or desktop app**

```bash
walgo uninstall
walgo uninstall --all --force
walgo uninstall --cli --force
walgo uninstall --desktop --force
walgo uninstall --all --keep-cache --force
```

**What it does:**

- Interactive uninstall wizard
- Removes CLI binary
- Removes desktop app
- Optionally cleans cache
- **NEVER** deletes Sui wallet or blockchain data

**Flags:**

- `-a, --all` - Uninstall both CLI and desktop app
- `--cli` - Uninstall CLI only
- `--desktop` - Uninstall desktop app only
- `-f, --force` - Skip confirmation prompts
- `--keep-cache` - Keep cache and data files

**What Gets Removed:**

- Walgo CLI binary
- Walgo Desktop app
- Cache and data (unless --keep-cache)

**What Stays Safe:**

- Your SUI balance (on blockchain)
- Deployed sites (on Walrus)
- Sui wallet (~/.sui/)
- All on-chain data

See [UNINSTALL.md](UNINSTALL.md) for details.

---

### `walgo completion`

**Generate shell autocompletion**

```bash
# Bash
walgo completion bash > /etc/bash_completion.d/walgo

# Zsh
walgo completion zsh > "${fpath[1]}/_walgo"

# Fish
walgo completion fish > ~/.config/fish/completions/walgo.fish

# PowerShell
walgo completion powershell > walgo.ps1
```

**Supported Shells:**

- bash
- zsh
- fish
- powershell

---

## Common Workflows

### Complete First Deploy (Recommended)

```bash
# 1. Create your site
walgo quickstart my-site
cd my-site

# 2. Preview locally
walgo serve

# 3. Deploy with the interactive wizard
walgo launch
# The wizard guides you through:
# - Network selection (testnet/mainnet)
# - Wallet setup
# - Project naming
# - Storage duration
# - Cost confirmation
# - Deployment

# 4. Configure SuiNS (post-deployment)
# Follow the instructions shown after deployment
# to link your domain at suins.io
```

---

### AI-Powered Content Creation

```bash
# 1. Configure AI
walgo ai configure

# 2. Generate content
walgo ai generate --type post

# 3. Review and build
cat content/posts/your-new-post.md
walgo build

# 4. Deploy
walgo launch
```

---

### Update Existing Site

```bash
# 1. Make changes
cd my-site
walgo ai update content/posts/old-post.md
# OR edit manually

# 2. Build
walgo build

# 3. Update project
walgo projects update "My Site"
```

---

### Project Management

```bash
# List all projects
walgo projects list

# Show details
walgo projects show "My Blog"

# Update site content
walgo projects update "My Blog"

# Update with new epochs
walgo projects update "My Blog" --epochs 10

# Edit metadata only (no on-chain update)
walgo projects edit "My Blog" --name "New Name" --description "Updated description"

# Edit metadata and apply on-chain
walgo projects edit "My Blog" --category blog --image-url "https://example.com/logo.png" --apply

# Archive old project
walgo projects archive "Old Site"

# Delete test project
walgo projects delete "Test Site"
```

### Metadata Management

```bash
# Update project name (also updates site_name in ws-resources.json)
walgo projects edit mysite --name "My Awesome Site"

# Update description for wallet/explorer display
walgo projects edit mysite --description "A decentralized blog about Web3"

# Update category
walgo projects edit mysite --category portfolio

# Update site logo/image
walgo projects edit mysite --image-url "https://example.com/logo.png"

# Update multiple fields at once
walgo projects edit mysite \
  --name "New Name" \
  --description "New description" \
  --category blog \
  --image-url "https://example.com/logo.png"

# Apply all changes on-chain
walgo projects edit mysite --apply
```

---

### Obsidian to Blog

```bash
# 1. Import Obsidian vault (creates site automatically)
walgo import ~/Documents/MyVault --site-name my-knowledge-base

# 2. Navigate to created site
cd my-knowledge-base

# 3. Build and deploy
walgo build
walgo launch
```

---

### Multi-Environment Deployment

```bash
# Development
walgo build --config walgo.dev.yaml
walgo deploy-http

# Staging
walgo build --config walgo.staging.yaml
walgo deploy --epochs 1

# Production
walgo build --config walgo.prod.yaml
walgo deploy --epochs 10 --network mainnet
```

---

## Global Flags

Available for all commands:

| Flag               | Description                                                      |
| ------------------ | ---------------------------------------------------------------- |
| `--config <path>`  | Custom config file path (default: ./walgo.yaml or ~/.walgo.yaml) |
| `--verbose` / `-v` | Enable verbose output                                            |
| `--help` / `-h`    | Show help for command                                            |

---

## Quick Command Reference

**Init & Build:**

- `quickstart` - One-command setup (recommended for new users)
- `init` - Create new Hugo site
- `build` - Build site with optimization
- `serve` - Local development server

**Content:**

- `new` - Create new content
- `import` - Import from Obsidian
- `ai generate` - AI content generation
- `ai update` - AI content updates

**Deploy:**

- `launch` - Interactive deployment wizard (**recommended**)
- `projects update` - Update existing project
- `deploy` - Direct on-chain deployment (advanced)
- `deploy-http` - HTTP deployment for testing (no wallet)
- `update` - Update site by object ID (advanced)

**Projects:**

- `projects list` - List all projects
- `projects show` - Show project details (including metadata)
- `projects update` - Update project content and/or metadata on-chain
- `projects edit` - Edit project metadata (name, description, category, image URL) with optional on-chain update
- `projects archive` - Archive project
- `projects delete` - Delete project

**Optimize:**

- `optimize` - Optimize assets
- `compress` - Brotli compression

**Setup:**

- `setup` - Configure wallet
- `setup-deps` - Install dependencies
- `ai configure` - Configure AI provider

**Diagnostics:**

- `doctor` - System diagnostics
- `status` - Check deployment status
- `domain` - SuiNS domain management
- `version` - Show version
- `uninstall` - Uninstall Walgo

**Desktop:**

- `desktop` - Launch desktop app

---

## Resources

- **GitHub**: https://github.com/selimozten/walgo
- **SuiNS**: https://suins.io (mainnet) | https://testnet.suins.io (testnet)
- **Walrus Docs**: https://docs.walrus.site
- **Hugo Docs**: https://gohugo.io/documentation/

---

## Getting Help

For any command, use `--help`:

```bash
walgo --help
walgo init --help
walgo deploy --help
walgo projects --help
```

---

**Total Commands: 29**

Last updated: v0.3.1
