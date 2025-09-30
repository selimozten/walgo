#!/bin/bash
# Walgo Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO="selimozten/walgo"
BINARY_NAME="walgo"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
USE_SUDO="${USE_SUDO:-true}"

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        darwin)
            OS="darwin"
            ;;
        linux)
            OS="linux"
            ;;
        mingw*|msys*|cygwin*|windows*)
            OS="windows"
            ;;
        *)
            print_error "Unsupported operating system: $OS"
            exit 1
            ;;
    esac

    case "$ARCH" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        aarch64|arm64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    print_info "Detected platform: $OS/$ARCH"
}

# Get latest release version
get_latest_version() {
    print_info "Fetching latest version..."

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi

    VERSION="${VERSION#v}" # Remove 'v' prefix if present
    print_success "Latest version: $VERSION"
}

# Download and install binary
install_binary() {
    local filename="${BINARY_NAME}_${VERSION}_${OS}_${ARCH}"

    if [ "$OS" = "windows" ]; then
        filename="${filename}.zip"
    else
        filename="${filename}.tar.gz"
    fi

    local download_url="https://github.com/$REPO/releases/download/v${VERSION}/${filename}"
    local tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${filename}"

    print_info "Downloading $BINARY_NAME from $download_url..."

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$tmp_file" "$download_url"
    else
        wget -q -O "$tmp_file" "$download_url"
    fi

    print_success "Downloaded successfully"

    # Extract archive
    print_info "Extracting archive..."
    cd "$tmp_dir"

    if [ "$OS" = "windows" ]; then
        unzip -q "$tmp_file"
    else
        tar -xzf "$tmp_file"
    fi

    # Find the binary
    local binary_path="${tmp_dir}/${BINARY_NAME}"
    if [ "$OS" = "windows" ]; then
        binary_path="${binary_path}.exe"
    fi

    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found in archive"
        rm -rf "$tmp_dir"
        exit 1
    fi

    print_success "Extracted binary"

    # Install binary
    print_info "Installing to $INSTALL_DIR..."

    # Check if we need sudo
    if [ "$USE_SUDO" = "true" ] && [ ! -w "$INSTALL_DIR" ]; then
        print_warning "Requires sudo for installation to $INSTALL_DIR"
        sudo install -m 755 "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    else
        install -m 755 "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    fi

    # Clean up
    rm -rf "$tmp_dir"

    print_success "Installed $BINARY_NAME to $INSTALL_DIR"
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."

    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version=$($BINARY_NAME version 2>&1 || echo "unknown")
        print_success "$BINARY_NAME installed successfully!"
        echo ""
        $BINARY_NAME --version 2>/dev/null || echo "Version: $VERSION"
    else
        print_warning "$BINARY_NAME installed but not found in PATH"
        print_info "You may need to add $INSTALL_DIR to your PATH"
        print_info "Add this to your shell profile (.bashrc, .zshrc, etc.):"
        echo ""
        echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
        echo ""
    fi
}

# Check and install dependencies
check_dependencies() {
    print_info "Checking dependencies..."

    local missing_deps=()

    # Check for Hugo
    if ! command -v hugo >/dev/null 2>&1; then
        missing_deps+=("hugo")
    else
        print_success "Hugo found: $(hugo version 2>&1 | head -1)"
    fi

    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo ""
        print_warning "Optional dependencies not found: ${missing_deps[*]}"
        print_info "Walgo works best with these tools installed."
        echo ""
        print_info "To install Hugo:"
        case "$OS" in
            darwin)
                echo "    brew install hugo"
                ;;
            linux)
                echo "    # Ubuntu/Debian:"
                echo "    sudo apt install hugo"
                echo ""
                echo "    # Or download from: https://gohugo.io/installation/"
                ;;
            windows)
                echo "    choco install hugo-extended"
                ;;
        esac
        echo ""
        print_info "You can also use 'walgo setup-deps' to install Walrus dependencies"
    else
        print_success "All optional dependencies found"
    fi
}

# Post-install instructions
show_next_steps() {
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    print_success "Walgo installation complete!"
    echo "═══════════════════════════════════════════════════════════"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Verify installation:"
    echo "     ${GREEN}walgo --help${NC}"
    echo ""
    echo "  2. Install Walrus dependencies:"
    echo "     ${GREEN}walgo setup-deps --with-site-builder --with-walrus --network testnet${NC}"
    echo ""
    echo "  3. Create your first site:"
    echo "     ${GREEN}walgo init my-site${NC}"
    echo "     ${GREEN}cd my-site${NC}"
    echo ""
    echo "  4. Build and deploy:"
    echo "     ${GREEN}walgo build${NC}"
    echo "     ${GREEN}walgo deploy-http \\${NC}"
    echo "       ${GREEN}--publisher https://publisher.walrus-testnet.walrus.space \\${NC}"
    echo "       ${GREEN}--aggregator https://aggregator.walrus-testnet.walrus.space \\${NC}"
    echo "       ${GREEN}--epochs 1${NC}"
    echo ""
    echo "  For on-chain deployment (requires wallet):"
    echo "     ${GREEN}walgo setup --network testnet --force${NC}"
    echo "     ${GREEN}walgo doctor --fix-paths${NC}"
    echo "     ${GREEN}walgo deploy --epochs 5${NC}"
    echo ""
    echo "Documentation: https://github.com/$REPO"
    echo ""
}

# Main installation flow
main() {
    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                   Walgo Installer                         ║"
    echo "║    Ship static sites to Walrus decentralized storage     ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    detect_platform
    get_latest_version
    install_binary
    verify_installation
    check_dependencies
    show_next_steps
}

# Run main function
main
