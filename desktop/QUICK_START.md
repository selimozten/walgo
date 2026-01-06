# Walgo Desktop - Quick Start Guide

## Installation

1. **Install Wails:**
   ```bash
   # macOS
   brew install wailsapp/wails/wails

   # All platforms
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

2. **Install frontend dependencies:**
   ```bash
   cd desktop/frontend
   npm install
   ```

## Development

Run the app in development mode with hot-reload:

```bash
# From project root
make desktop-dev

# OR from desktop directory
cd desktop
wails dev
```

## Building

### Build for Current Platform

```bash
# Using Make
make desktop-build

# Using Wails directly
cd desktop
wails build
```

Output: `desktop/build/bin/`

### Build for All Platforms

**Option 1: Using Make (Simple)**
```bash
make desktop-build-all
```

**Option 2: Using Build Script (Advanced)**
```bash
# Build everything
./desktop/build-all.sh --all

# With compression and clean build
./desktop/build-all.sh --all --clean --compress
```

**Option 3: Individual Platforms**
```bash
# macOS (Universal - Intel + Apple Silicon)
make desktop-build-darwin
# OR
./desktop/build-all.sh --macos

# Windows (amd64)
make desktop-build-windows
# OR
./desktop/build-all.sh --windows

# Linux (amd64)
make desktop-build-linux
# OR
./desktop/build-all.sh --linux
```

## Build Outputs

| Platform | Output | Location |
|----------|--------|----------|
| macOS | `.app` bundle | `desktop/build/bin/walgo-desktop.app` |
| Windows | `.exe` file | `desktop/build/bin/desktop.exe` |
| Linux | Binary | `desktop/build/bin/desktop` |

Distribution archives (with build-all.sh):
- `desktop/dist/walgo-desktop-{version}-macos-universal.zip`
- `desktop/dist/walgo-desktop-{version}-windows-amd64.zip`
- `desktop/dist/walgo-desktop-{version}-linux-amd64.tar.gz`

## All Make Commands

```bash
make desktop-dev              # Run in development mode
make desktop-build            # Build for current platform
make desktop-build-darwin     # Build for macOS (Universal)
make desktop-build-windows    # Build for Windows (amd64)
make desktop-build-linux      # Build for Linux (amd64)
make desktop-build-all        # Build for all platforms
make desktop-clean            # Clean build artifacts
make desktop-install-deps     # Install frontend dependencies
```

## Build Script Options

```bash
./desktop/build-all.sh [OPTIONS]

OPTIONS:
  -a, --all         Build for all platforms
  -m, --macos       Build for macOS
  -w, --windows     Build for Windows
  -l, --linux       Build for Linux
  -c, --clean       Clean before building
  -z, --compress    Compress with UPX
  --skip-deps       Skip npm install
  -h, --help        Show help
```

## Common Tasks

### Clean Build
```bash
make desktop-clean
# OR
./desktop/build-all.sh --all --clean
```

### Update Dependencies
```bash
cd desktop/frontend
npm update
```

### Rebuild Everything
```bash
make desktop-clean
make desktop-install-deps
make desktop-build-all
```

## Troubleshooting

### Frontend won't build
```bash
cd desktop/frontend
rm -rf node_modules package-lock.json
npm install
```

### Go module errors
```bash
cd desktop
go mod tidy
```

### Permission denied (macOS)
```bash
xattr -d com.apple.quarantine desktop/build/bin/walgo-desktop.app
```

## Platform Requirements

### macOS
- Xcode Command Line Tools
  ```bash
  xcode-select --install
  ```

### Windows
- Windows 10/11 SDK
- WebView2 Runtime

### Linux
- GTK+ 3 and WebKitGTK
  ```bash
  # Debian/Ubuntu
  sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev

  # Fedora
  sudo dnf install gtk3-devel webkit2gtk3-devel
  ```

## More Information

See [BUILD.md](BUILD.md) for detailed build instructions and troubleshooting.
