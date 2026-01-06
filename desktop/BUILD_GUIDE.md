# Walgo Desktop - Build Guide

Complete guide for building Walgo Desktop for all platforms.

## 🚀 Quick Start

### Development Mode

```bash
cd desktop
wails dev
```

### Production Build (Current Platform)

```bash
cd desktop
wails build
```

### Build All Platforms

```bash
cd desktop
./build-all-platforms.sh
```

## 📦 Build Outputs

After building, you'll find the applications in `build/bin/`:

- **macOS**: `walgo-desktop.app`
- **Windows**: `walgo-desktop.exe`
- **Linux**: `walgo-desktop`

## 🎨 Icons

All platforms use the Walgo logo:

### Icon Files

- **Source**: `build/appicon.png` (512x512 PNG)
- **macOS**: `build/darwin/iconfile.icns` (multi-resolution)
- **Windows**: `build/windows/icon.ico` (multi-resolution)
- **Linux**: `build/appicon.png` (512x512 PNG)

### Update Icons

```bash
# 1. Replace the source logo
cp your-new-logo.svg ../walgo-Wlogo.svg

# 2. Convert to PNG
magick ../walgo-Wlogo.svg -resize 512x512 build/appicon.png

# 3. Generate platform-specific icons
./update-icons.sh
```

## 🛠️ Build Scripts

### `build-all-platforms.sh`

Builds for all platforms (macOS, Windows, Linux):

- Cleans previous builds
- Verifies icons exist
- Builds for each platform
- Shows build summary

### `clean.sh`

Removes build artifacts and temporary files:

- Removes `build/bin/`
- Removes `frontend/dist/`
- Removes duplicate icons
- Keeps essential icon files

### `update-icons.sh`

Generates platform-specific icons from source PNG:

- Creates Windows `.ico` (multi-resolution)
- Creates macOS `.icns` (multi-resolution)
- Linux uses PNG directly

## 🔧 Configuration

### `wails.json`

Main configuration file:

```json
{
  "build:dir": "./build",
  "mac": {
    "icon": "build/darwin/iconfile.icns"
  },
  "linux": {
    "icon": "build/appicon.png"
  }
}
```

### Platform-Specific Settings

#### macOS (`main.go`)

```go
case "darwin":
    appOptions.Frameless = true
    appOptions.Mac = &mac.Options{
        TitleBar: mac.TitleBarHidden(),
        About: &mac.AboutInfo{
            Title:   "Walgo",
            Message: "...",
        },
    }
```

#### Windows (`main.go`)

```go
case "windows":
    appOptions.Frameless = true
    appOptions.Windows = &windows.Options{
        Theme: windows.SystemDefault,
    }
```

#### Linux (`main.go`)

```go
case "linux":
    appOptions.Frameless = true
    appOptions.Linux = &linux.Options{
        WebviewGpuPolicy: linux.WebviewGpuPolicyAlways,
    }
```

## 📋 Build Requirements

### All Platforms

- Go 1.22+
- Node.js 18+
- Wails CLI v2.11.0+

### macOS

- Xcode Command Line Tools
- For icon generation: `sips`, `iconutil` (included in macOS)

### Windows

- WebView2 Runtime (auto-installed on Windows 10/11)
- For icon generation: ImageMagick

### Linux

- GTK3 development libraries
- WebKit2GTK development libraries

```bash
# Ubuntu/Debian
sudo apt install libgtk-3-dev libwebkit2gtk-4.0-dev

# Fedora
sudo dnf install gtk3-devel webkit2gtk3-devel
```

## 🧹 Cleanup

Remove all build artifacts:

```bash
./clean.sh
```

This removes:

- `build/bin/` (all compiled binaries)
- `frontend/dist/` (frontend build output)
- Duplicate/temporary icon files
- Vite cache

## 🐛 Troubleshooting

### Icon Not Showing

**macOS:**

```bash
# Clear icon cache
rm -rf build/bin
./update-icons.sh
wails build
```

**Windows:**

```bash
# Regenerate icon
./update-icons.sh
wails build -platform windows/amd64
```

**Linux:**

```bash
# Verify icon exists
ls -lh build/appicon.png
wails build -platform linux/amd64
```

### Double `bin` Directory

If you see `build/bin/bin/`, check `wails.json`:

```json
"build:dir": "./build"  // ✅ Correct
"build:dir": "./build/bin"  // ❌ Wrong (creates double bin)
```

### Build Fails

1. Clean everything:

   ```bash
   ./clean.sh
   rm -rf frontend/node_modules
   ```

2. Reinstall dependencies:

   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. Rebuild:
   ```bash
   wails build -clean
   ```

## 📱 Testing

### macOS

```bash
# Run directly
./build/bin/walgo-desktop.app/Contents/MacOS/walgo-desktop

# Or open with Finder
open build/bin/walgo-desktop.app
```

### Windows

```bash
# Run directly
./build/bin/walgo-desktop.exe
```

### Linux

```bash
# Make executable
chmod +x build/bin/walgo-desktop

# Run
./build/bin/walgo-desktop
```

## 🚢 Distribution

### macOS

Create DMG:

```bash
# Using create-dmg (install via Homebrew)
create-dmg \
  --volname "Walgo Desktop" \
  --window-pos 200 120 \
  --window-size 800 400 \
  --icon-size 100 \
  --icon "walgo-desktop.app" 200 190 \
  --hide-extension "walgo-desktop.app" \
  --app-drop-link 600 185 \
  "walgo-desktop.dmg" \
  "build/bin/"
```

### Windows

NSIS installer is automatically created by Wails in `build/bin/`.

### Linux

Create `.deb` package:

```bash
wails build -platform linux/amd64 -nsis
```

## 📚 Additional Resources

- [Wails Documentation](https://wails.io/docs/)
- [Main Project README](../README.md)
- [Architecture Documentation](./ARCHITECTURE.md)

## 🆘 Support

For issues specific to the desktop app:

1. Check this guide
2. Run `./clean.sh` and rebuild
3. Check Wails documentation
4. Open an issue on GitHub
