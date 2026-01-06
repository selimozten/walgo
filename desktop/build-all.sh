#!/bin/bash

# Walgo Desktop - Multi-Platform Build Script
# Builds for macOS, Windows, and Linux

set -e

echo "🚀 Walgo Desktop - Multi-Platform Build"
echo "========================================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if we're in the desktop directory
if [ ! -f "wails.json" ]; then
    echo -e "${RED}❌ Error: wails.json not found. Please run from desktop directory.${NC}"
    exit 1
fi

# Clean previous builds
echo -e "${BLUE}🧹 Cleaning previous builds...${NC}"
rm -rf build/bin
echo -e "${GREEN}✅ Clean complete${NC}"
echo ""

# Build for macOS (current platform)
echo -e "${BLUE}🍎 Building for macOS...${NC}"
wails build -clean -platform darwin/universal
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ macOS build complete${NC}"
    echo "   📦 Location: build/bin/walgo-desktop.app"
else
    echo -e "${RED}❌ macOS build failed${NC}"
fi
echo ""

# Build for Windows (amd64)
echo -e "${BLUE}🪟 Building for Windows (amd64)...${NC}"
wails build -clean -platform windows/amd64
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Windows (amd64) build complete${NC}"
    echo "   📦 Location: build/bin/walgo-desktop.exe"
else
    echo -e "${RED}❌ Windows (amd64) build failed${NC}"
fi
echo ""

# Build for Linux (amd64)
echo -e "${BLUE}🐧 Building for Linux (amd64)...${NC}"
wails build -clean -platform linux/amd64
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Linux (amd64) build complete${NC}"
    echo "   📦 Location: build/bin/walgo-desktop"
else
    echo -e "${RED}❌ Linux (amd64) build failed${NC}"
fi
echo ""

# Summary
echo "========================================"
echo -e "${GREEN}🎉 Build Process Complete!${NC}"
echo ""
echo "📦 Build artifacts:"
ls -lh build/bin/
echo ""
echo "🚀 Ready to distribute!"
