# Walgo Optimizer Engine

The Walgo Optimizer Engine is a built-in asset optimization system that automatically minifies and optimizes your HTML, CSS, and JavaScript files for better performance and smaller file sizes.

## Features

### HTML Optimization
- **Minification**: Removes unnecessary whitespace and line breaks
- **Comment Removal**: Removes HTML comments while preserving conditional comments
- **Inline Asset Compression**: Minifies CSS and JavaScript within `<style>` and `<script>` tags
- **Whitespace Optimization**: Intelligent whitespace removal that preserves formatting where needed

### CSS Optimization
- **Minification**: Removes whitespace, comments, and unnecessary formatting
- **Color Compression**: Converts `#ffffff` to `#fff`, named colors to shorter hex values
- **Rule Optimization**: Removes trailing semicolons, unnecessary quotes
- **Unused Rule Removal**: Optionally removes CSS rules not used in HTML (aggressive optimization)
- **Unit Compression**: Converts `0px` to `0`, `0.5` to `.5`

### JavaScript Optimization
- **Minification**: Removes whitespace and comments while preserving functionality
- **String/Regex Preservation**: Safely handles strings and regular expressions
- **Comment Removal**: Removes single-line and multi-line comments
- **Basic Obfuscation**: Optional variable name shortening (use with caution)

## Usage

### Automatic Optimization
Optimization runs automatically after `walgo build` if enabled in your configuration:

```bash
walgo build                 # Builds Hugo site and optimizes assets
walgo build --no-optimize   # Builds Hugo site without optimization
```

### Manual Optimization
You can run optimization manually on any directory:

```bash
walgo optimize              # Optimizes files in Hugo's public directory
walgo optimize ./dist       # Optimizes files in specific directory
walgo optimize --verbose    # Shows detailed optimization progress
```

### Command Line Options
```bash
walgo optimize [directory] [flags]

Flags:
  --html                    Enable HTML optimization (default true)
  --css                     Enable CSS optimization (default true)
  --js                      Enable JavaScript optimization (default true)
  --remove-unused-css       Remove unused CSS rules (aggressive)
  -v, --verbose             Enable verbose output
```

## Configuration

Configure the optimizer in your `walgo.yaml` file:

```yaml
optimizer:
  enabled: true
  verbose: false
  
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
    removeUnused: false
    compressColors: true
  
  js:
    enabled: true
    minifyJS: true
    removeComments: true
    obfuscate: false
    sourceMaps: false
  
  skipPatterns:
    - "*.min.js"
    - "*.min.css"
    - "*.min.html"
    - "*/.git/*"
    - "*/node_modules/*"
```

### Configuration Options

#### Global Settings
- `enabled`: Enable/disable the optimizer entirely
- `verbose`: Show detailed optimization progress
- `skipPatterns`: File patterns to skip during optimization

#### HTML Settings
- `enabled`: Enable HTML optimization
- `minifyHTML`: Remove unnecessary whitespace
- `removeComments`: Remove HTML comments (preserves conditional comments)
- `removeWhitespace`: Aggressive whitespace removal
- `compressInlineCSS`: Minify CSS within `<style>` tags
- `compressInlineJS`: Minify JavaScript within `<script>` tags

#### CSS Settings
- `enabled`: Enable CSS optimization
- `minifyCSS`: Remove whitespace and formatting
- `removeComments`: Remove CSS comments
- `removeUnused`: Remove unused CSS rules (requires HTML analysis)
- `compressColors`: Optimize color values
- `autoprefixer`: Add vendor prefixes (requires additional setup)

#### JavaScript Settings
- `enabled`: Enable JavaScript optimization
- `minifyJS`: Remove whitespace and formatting
- `removeComments`: Remove JavaScript comments
- `obfuscate`: Basic variable name obfuscation (can break code)
- `sourceMaps`: Generate source maps (not implemented yet)

## Performance Impact

Typical optimization results:
- **HTML**: 15-25% size reduction
- **CSS**: 20-40% size reduction
- **JavaScript**: 25-50% size reduction

Example output:
```
🎯 Optimization Results
======================
Files processed: 42
Files optimized: 18
Files skipped: 24
Duration: 234ms

📊 Size Reduction
Original size: 2.4 MB
Optimized size: 1.8 MB
Bytes saved: 645.2 KB (26.9%)

📄 HTML Files
Files: 8, Saved: 45.3 KB (18.2%)

🎨 CSS Files
Files: 5, Saved: 234.7 KB (31.4%)

📜 JavaScript Files
Files: 5, Saved: 365.2 KB (42.1%)
```

## Safety and Best Practices

### Safe Optimizations (Default)
- HTML minification
- CSS minification
- JavaScript minification
- Comment removal
- Color compression

### Aggressive Optimizations (Use with Caution)
- **Unused CSS removal**: Can remove CSS needed by JavaScript
- **JavaScript obfuscation**: May break code with dynamic property access
- **Autoprefixer**: Requires additional configuration

### Skip Patterns
Always skip:
- Already minified files (`*.min.*`)
- Source maps (`*.map`)
- Version control directories (`*/.git/*`)
- Package manager directories (`*/node_modules/*`)
- Service workers and manifest files

### Testing
Always test your optimized site before deploying:
1. Run `walgo build` with optimization
2. Test with `walgo serve`
3. Verify all functionality works
4. Check for any broken layouts or scripts

## Integration with Hugo

The optimizer integrates seamlessly with Hugo:
1. Hugo builds your site to the `public` directory
2. Optimizer processes all files in the output
3. Original Hugo files remain unchanged
4. Only the built output is optimized

This ensures your source files are never modified, and you can always rebuild without optimization if needed.

## Troubleshooting

### Common Issues

**Broken JavaScript after optimization:**
- Disable obfuscation: `js.obfuscate: false`
- Check for dynamic property access
- Review console errors in browser

**Missing styles after CSS optimization:**
- Disable unused rule removal: `css.removeUnused: false`
- Check for dynamically added classes
- Verify CSS selector patterns

**Layout issues after HTML minification:**
- Check `<pre>` and `<code>` tag preservation
- Verify inline styles and scripts
- Test with `html.removeWhitespace: false`

### Debug Mode
Run with verbose output to see detailed processing:
```bash
walgo optimize --verbose
```

This shows:
- Files being processed
- Optimization steps applied
- Size changes for each file
- Error messages and warnings

### Selective Optimization
Disable specific optimizations if issues occur:
```bash
walgo optimize --html=false    # Skip HTML optimization
walgo optimize --css=false     # Skip CSS optimization
walgo optimize --js=false      # Skip JavaScript optimization
```

## Advanced Usage

### Custom Skip Patterns
Add custom patterns to skip specific files:
```yaml
optimizer:
  skipPatterns:
    - "*/analytics/*"
    - "third-party.js"
    - "*.debug.css"
```

### Per-Environment Configuration
Use different optimization levels for development vs production:
```yaml
# Development (lighter optimization)
optimizer:
  enabled: true
  html:
    removeComments: false
  css:
    removeUnused: false
  js:
    obfuscate: false
```

### Integration with CI/CD
Add optimization to your deployment pipeline:
```bash
# Build and optimize
walgo build

# Deploy optimized files
walgo deploy
``` 