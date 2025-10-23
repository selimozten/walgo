# Configuration Reference

Complete reference for configuring Walgo.

## Table of Contents

- [Configuration File](#configuration-file)
- [Configuration Sources](#configuration-sources)
- [Hugo Configuration](#hugo-configuration)
- [Walrus Configuration](#walrus-configuration)
- [Optimizer Configuration](#optimizer-configuration)
- [Obsidian Configuration](#obsidian-configuration)
- [Environment Variables](#environment-variables)
- [Configuration Examples](#configuration-examples)

## Configuration File

Walgo uses YAML configuration files named `walgo.yaml` or `.walgo.yaml`.

### File Locations

Walgo searches for configuration files in this order:

1. **Explicit path:** `--config /path/to/config.yaml`
2. **Current directory:** `./walgo.yaml`
3. **Home directory:** `~/.walgo.yaml`

The first file found is used. If no file is found, default values are used.

### Basic Structure

```yaml
# Hugo site configuration
hugo:
  version: ""
  baseURL: ""
  publishDir: "public"
  contentDir: "content"
  buildDraft: false
  minify: true

# Walrus deployment configuration
walrus:
  projectID: ""
  bucketName: ""
  entrypoint: "index.html"
  epochs: 5
  network: "testnet"
  suinsDomain: ""

# Asset optimizer configuration
optimizer:
  enabled: true
  verbose: false
  html:
    enabled: true
    minifyHTML: true
    removeComments: true
    compressInlineCSS: true
    compressInlineJS: true
  css:
    enabled: true
    minifyCSS: true
    removeComments: true
    removeUnused: false
    compressColors: true
  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false

# Obsidian import configuration
obsidian:
  vaultPath: ""
  attachmentDir: "images"
  convertWikilinks: true
  includeDrafts: false
  frontmatterFormat: "yaml"
```

## Configuration Sources

Walgo merges configuration from multiple sources with this priority (highest to lowest):

1. **Command-line flags**
2. **Environment variables**
3. **Configuration file** (`walgo.yaml`)
4. **Default values**

Example:

```bash
# Config file sets epochs: 5
# Environment variable overrides to 10
export WALGO_EPOCHS=10

# Command flag overrides to 3
walgo deploy --epochs 3
# Final value: 3
```

## Hugo Configuration

Controls how Hugo builds your site.

### `hugo.version`

- **Type:** String
- **Default:** `""` (use system Hugo)
- **Description:** Specific Hugo version to use

```yaml
hugo:
  version: "0.125.0"  # Use specific version
```

### `hugo.baseURL`

- **Type:** String
- **Default:** `""` (use Hugo config)
- **Description:** Override Hugo's baseURL

```yaml
hugo:
  baseURL: "https://example.walrus.site/"
```

**When to use:** Set this to your Walrus site URL after first deployment.

### `hugo.publishDir`

- **Type:** String
- **Default:** `"public"`
- **Description:** Output directory for built site

```yaml
hugo:
  publishDir: "dist"  # Use dist/ instead of public/
```

### `hugo.contentDir`

- **Type:** String
- **Default:** `"content"`
- **Description:** Directory containing markdown content

```yaml
hugo:
  contentDir: "posts"  # Content in posts/ instead
```

### `hugo.buildDraft`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Include draft posts in build

```yaml
hugo:
  buildDraft: true  # Include drafts (for testing)
```

**Warning:** Don't enable for production deployments!

### `hugo.minify`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Use Hugo's built-in minification

```yaml
hugo:
  minify: false  # Disable Hugo minification (Walgo optimizer used instead)
```

### `hugo.resourceDir`

- **Type:** String
- **Default:** `"resources"`
- **Description:** Hugo resources directory

```yaml
hugo:
  resourceDir: "_gen"
```

## Walrus Configuration

Controls Walrus deployment behavior.

### `walrus.projectID`

- **Type:** String
- **Default:** `""`
- **Description:** Walrus project identifier

```yaml
walrus:
  projectID: "my-blog-project"
```

**Auto-set:** Automatically set after first deployment.

### `walrus.bucketName`

- **Type:** String
- **Default:** `""`
- **Description:** Optional bucket for organizing deployments

```yaml
walrus:
  bucketName: "production-sites"
```

### `walrus.entrypoint`

- **Type:** String
- **Default:** `"index.html"`
- **Description:** Default file for directory requests

```yaml
walrus:
  entrypoint: "home.html"  # Use home.html as entry point
```

### `walrus.epochs`

- **Type:** Integer
- **Default:** `5`
- **Description:** Storage duration in epochs (~30 days each)

```yaml
walrus:
  epochs: 10  # ~300 days of storage
```

**Cost:** More epochs = higher cost. Balance permanence vs cost.

**Recommendations:**
- **Testing:** 1-2 epochs
- **Short-term:** 5 epochs (~5 months)
- **Long-term:** 10+ epochs

### `walrus.network`

- **Type:** String
- **Default:** `"testnet"`
- **Options:** `"testnet"`, `"mainnet"`
- **Description:** Sui network to use

```yaml
walrus:
  network: "mainnet"  # Use mainnet (requires real SUI)
```

**Warning:** Mainnet requires real SUI tokens with monetary value!

### `walrus.suinsDomain`

- **Type:** String
- **Default:** `""`
- **Description:** SuiNS domain name for your site

```yaml
walrus:
  suinsDomain: "myblog.sui"
```

**Requirements:**
- Must own the SuiNS domain
- Domain must be configured to point to site object

## Optimizer Configuration

Controls asset optimization behavior.

### Global Optimizer Settings

#### `optimizer.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable/disable all optimization

```yaml
optimizer:
  enabled: false  # Disable all optimization
```

#### `optimizer.verbose`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Show detailed optimization output

```yaml
optimizer:
  verbose: true  # Show which files are being optimized
```

#### `optimizer.skipPatterns`

- **Type:** Array of strings
- **Default:** `["*.min.js", "*.min.css", "*.min.html"]`
- **Description:** File patterns to skip

```yaml
optimizer:
  skipPatterns:
    - "*.min.js"
    - "*.min.css"
    - "vendor/*"
    - "third-party.js"
```

**Common patterns:**
- `*.min.*` - Already minified files
- `vendor/*` - Third-party libraries
- `*.map` - Source maps
- `sw.js` - Service workers

### HTML Optimizer

#### `optimizer.html.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable HTML optimization

```yaml
optimizer:
  html:
    enabled: false  # Skip HTML files
```

#### `optimizer.html.minifyHTML`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Remove unnecessary whitespace

```yaml
optimizer:
  html:
    minifyHTML: true  # Minify HTML
```

**Result:** Reduces HTML size by 15-25%

#### `optimizer.html.removeComments`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Remove HTML comments

```yaml
optimizer:
  html:
    removeComments: true
```

**Note:** Preserves conditional comments (e.g., IE-specific).

#### `optimizer.html.removeWhitespace`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Aggressive whitespace removal

```yaml
optimizer:
  html:
    removeWhitespace: false  # Preserve whitespace (safer)
```

**Warning:** May affect `<pre>` and `<code>` tag formatting.

#### `optimizer.html.compressInlineCSS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify CSS in `<style>` tags

```yaml
optimizer:
  html:
    compressInlineCSS: true
```

#### `optimizer.html.compressInlineJS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify JavaScript in `<script>` tags

```yaml
optimizer:
  html:
    compressInlineJS: true
```

### CSS Optimizer

#### `optimizer.css.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable CSS optimization

```yaml
optimizer:
  css:
    enabled: true
```

#### `optimizer.css.minifyCSS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify CSS files

```yaml
optimizer:
  css:
    minifyCSS: true
```

**Result:** Reduces CSS size by 20-40%

#### `optimizer.css.removeComments`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Remove CSS comments

```yaml
optimizer:
  css:
    removeComments: true
```

#### `optimizer.css.removeUnused`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Remove unused CSS rules

```yaml
optimizer:
  css:
    removeUnused: true  # Aggressive optimization
```

**Warning:** May break dynamically added styles!

**When to use:**
- Simple static sites
- No JavaScript-added classes
- After thorough testing

**When NOT to use:**
- Sites with dynamic content
- JavaScript frameworks (React, Vue)
- Third-party widgets

#### `optimizer.css.compressColors`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Optimize color values

```yaml
optimizer:
  css:
    compressColors: true
```

**Examples:**
- `#ffffff` → `#fff`
- `rgb(0, 0, 0)` → `#000`

### JavaScript Optimizer

#### `optimizer.js.enabled`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Enable JavaScript optimization

```yaml
optimizer:
  js:
    enabled: true
```

#### `optimizer.js.minifyJS`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Minify JavaScript files

```yaml
optimizer:
  js:
    minifyJS: true
```

**Result:** Reduces JS size by 25-50%

#### `optimizer.js.removeComments`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Remove JavaScript comments

```yaml
optimizer:
  js:
    removeComments: true
```

**Note:** Preserves license comments.

#### `optimizer.js.obfuscate`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Obfuscate variable names

```yaml
optimizer:
  js:
    obfuscate: true  # Make code harder to read
```

**Warning:** Can break code! Use with extreme caution.

**Don't use if:**
- Code uses `eval()`
- Dynamic property access
- Reflection
- External libraries expect specific names

#### `optimizer.js.sourceMaps`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Generate source maps

```yaml
optimizer:
  js:
    sourceMaps: true  # Generate .map files
```

**Status:** Not yet implemented

## Obsidian Configuration

Controls Obsidian vault import behavior.

### `obsidian.vaultPath`

- **Type:** String
- **Default:** `""`
- **Description:** Path to Obsidian vault

```yaml
obsidian:
  vaultPath: "/Users/you/Documents/MyVault"
```

### `obsidian.attachmentDir`

- **Type:** String
- **Default:** `"images"`
- **Description:** Directory for imported attachments

```yaml
obsidian:
  attachmentDir: "static/attachments"
```

### `obsidian.convertWikilinks`

- **Type:** Boolean
- **Default:** `true`
- **Description:** Convert `[[links]]` to markdown links

```yaml
obsidian:
  convertWikilinks: true
```

**Example conversion:**
```markdown
# Before
[[My Page]]
[[My Page|Custom Text]]

# After
[My Page](my-page.md)
[Custom Text](my-page.md)
```

### `obsidian.includeDrafts`

- **Type:** Boolean
- **Default:** `false`
- **Description:** Include notes marked as drafts

```yaml
obsidian:
  includeDrafts: false  # Skip draft notes
```

### `obsidian.frontmatterFormat`

- **Type:** String
- **Default:** `"yaml"`
- **Options:** `"yaml"`, `"toml"`, `"json"`
- **Description:** Frontmatter format for imported notes

```yaml
obsidian:
  frontmatterFormat: "toml"  # Use TOML frontmatter
```

## Environment Variables

All configuration values can be set via environment variables using the `WALGO_` prefix.

### Format

```bash
WALGO_<SECTION>_<KEY>=value
```

### Examples

```bash
# Hugo configuration
export WALGO_HUGO_PUBLISHDIR="dist"
export WALGO_HUGO_BUILDDRAFT="true"

# Walrus configuration
export WALGO_WALRUS_EPOCHS="10"
export WALGO_WALRUS_NETWORK="mainnet"

# Optimizer configuration
export WALGO_OPTIMIZER_ENABLED="false"
export WALGO_OPTIMIZER_HTML_MINIFYHTML="true"

# Obsidian configuration
export WALGO_OBSIDIAN_VAULTPATH="/path/to/vault"
```

### Nested Configuration

For nested values, use underscores:

```bash
# optimizer.html.minifyHTML
export WALGO_OPTIMIZER_HTML_MINIFYHTML="true"

# optimizer.css.removeUnused
export WALGO_OPTIMIZER_CSS_REMOVEUNUSED="false"
```

### Boolean Values

Accepted boolean values:
- **True:** `"true"`, `"1"`, `"yes"`, `"on"`
- **False:** `"false"`, `"0"`, `"no"`, `"off"`

## Configuration Examples

### Minimal Configuration

```yaml
# walgo.yaml - Minimal setup
walrus:
  epochs: 5
```

Everything else uses defaults.

### Development Configuration

```yaml
# walgo.yaml - Development setup
hugo:
  buildDraft: true        # Include drafts
  minify: false          # Easier debugging

optimizer:
  enabled: false         # Disable optimization for speed

walrus:
  epochs: 1             # Minimal storage for testing
  network: "testnet"
```

### Production Configuration

```yaml
# walgo.yaml - Production setup
hugo:
  baseURL: "https://myblog.walrus.site/"
  buildDraft: false
  minify: true

optimizer:
  enabled: true
  verbose: false
  html:
    enabled: true
    minifyHTML: true
    removeComments: true
  css:
    enabled: true
    minifyCSS: true
    removeUnused: false    # Safe default
    compressColors: true
  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false       # Don't risk breaking code

walrus:
  epochs: 10              # ~10 months storage
  network: "testnet"
  suinsDomain: "myblog.sui"

obsidian:
  vaultPath: ""
  convertWikilinks: true
  includeDrafts: false
```

### Aggressive Optimization

```yaml
# walgo.yaml - Maximum optimization
optimizer:
  enabled: true
  verbose: true

  html:
    enabled: true
    minifyHTML: true
    removeComments: true
    removeWhitespace: true
    compressInlineCSS: true
    compressInlineJS: true

  css:
    enabled: true
    minifyCSS: true
    removeComments: true
    removeUnused: true      # CAREFUL - test thoroughly!
    compressColors: true

  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false        # Still risky

  skipPatterns:
    - "*.min.*"
    - "vendor/*"
```

**Warning:** Test thoroughly! Aggressive optimization can break sites.

### Obsidian-Focused Configuration

```yaml
# walgo.yaml - Optimized for Obsidian
hugo:
  contentDir: "content"
  publishDir: "public"

obsidian:
  vaultPath: "/Users/you/Documents/MyVault"
  attachmentDir: "static/images"
  convertWikilinks: true
  includeDrafts: false
  frontmatterFormat: "yaml"

optimizer:
  enabled: true
  html:
    enabled: true
  css:
    enabled: true
    removeUnused: false    # Obsidian themes may use dynamic classes
  js:
    enabled: true
    obfuscate: false
```

### Multi-Environment Setup

Use different configs for different environments:

```bash
# Development
walgo build --config walgo.dev.yaml

# Staging
walgo build --config walgo.staging.yaml

# Production
walgo build --config walgo.prod.yaml
```

**walgo.dev.yaml:**
```yaml
hugo:
  buildDraft: true
optimizer:
  enabled: false
walrus:
  epochs: 1
```

**walgo.prod.yaml:**
```yaml
hugo:
  buildDraft: false
optimizer:
  enabled: true
walrus:
  epochs: 10
  network: "mainnet"
```

## Configuration Validation

### Checking Configuration

```bash
# Show current configuration
walgo doctor -v
```

This displays:
- Configuration file location
- Merged configuration values
- Environment variables in use
- Default values

### Common Validation Errors

**Invalid YAML syntax:**
```
Error: yaml: line 5: mapping values are not allowed in this context
```

**Solution:** Check YAML indentation and syntax.

**Unknown configuration key:**
```
Warning: Unknown configuration key: optimiser (did you mean optimizer?)
```

**Solution:** Fix typos in configuration keys.

**Invalid value type:**
```
Error: optimizer.enabled must be a boolean (got: "yes")
```

**Solution:** Use proper types (true/false for booleans, not "yes"/"no").

## Best Practices

### 1. Version Control Your Config

```bash
git add walgo.yaml
git commit -m "Add Walgo configuration"
```

### 2. Use Comments

```yaml
# Site is deployed to testnet
walrus:
  network: "testnet"
  epochs: 5  # ~5 months of storage
```

### 3. Don't Commit Secrets

```yaml
# BAD - Don't commit sensitive data
walrus:
  apiKey: "secret-key-123"

# GOOD - Use environment variables
# Set via: export WALGO_WALRUS_APIKEY="secret-key-123"
```

### 4. Test Configuration Changes

```bash
# Test config with dry-run (if available)
walgo build --dry-run

# Or test with HTTP deployment first
walgo build && walgo deploy-http
```

### 5. Document Your Settings

Add README with configuration explanation:

```markdown
# Configuration

- `epochs: 10` - We use 10 epochs for long-term storage
- `removeUnused: false` - Our theme uses dynamic CSS classes
```

## Related Documentation

- [Commands Reference](COMMANDS.md) - All command-line flags
- [Environment Variables](CONFIGURATION.md#environment-variables) - Full env var reference
- [Optimizer Documentation](OPTIMIZER.md) - Detailed optimizer guide
- [Deployment Guide](DEPLOYMENT.md) - Deployment configuration

## Troubleshooting

See [Troubleshooting Guide](TROUBLESHOOTING.md) for configuration-related issues.
