#!/bin/bash
# Walgo Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/selimozten/walgo/main/install.sh | bash

set -e

# ============================================================================
# TERMINAL DETECTION & COMPATIBILITY
# ============================================================================

USE_COLORS=true
USE_EMOJI=true

# Detect terminal capabilities
detect_terminal() {
    # Check if we're in a CI environment
    if [ -n "$CI" ] || [ -n "$GITHUB_ACTIONS" ] || [ -n "$GITLAB_CI" ] || \
       [ -n "$JENKINS_URL" ] || [ -n "$CIRCLECI" ] || [ -n "$TRAVIS" ]; then
        USE_COLORS=false
        USE_EMOJI=false
        return
    fi

    # Check if output is a terminal
    if [ ! -t 1 ]; then
        USE_COLORS=false
        USE_EMOJI=false
        return
    fi

    # Check TERM variable
    case "${TERM:-}" in
        dumb|"")
            USE_COLORS=false
            USE_EMOJI=false
            ;;
        *)
            USE_COLORS=true
            # Check for modern terminals that support emojis
            if [ "${TERM_PROGRAM:-}" = "iTerm.app" ] || \
               [ "${TERM_PROGRAM:-}" = "vscode" ] || \
               [ "${TERM_PROGRAM:-}" = "Apple_Terminal" ] || \
               [ -n "${WT_SESSION:-}" ] || \
               [ -n "${COLORTERM:-}" ]; then
                USE_EMOJI=true
            elif [[ "${TERM:-}" == *"256color"* ]] || [[ "${TERM:-}" == *"xterm"* ]]; then
                USE_EMOJI=true
            else
                USE_EMOJI=false
            fi
            ;;
    esac

    # Platform-specific detection
    local os_type=$(uname -s 2>/dev/null || echo unknown)
    case "$os_type" in
        Darwin)
            USE_EMOJI=true  # macOS terminals generally support emojis
            ;;
        Linux)
            : # Keep detected value
            ;;
        MINGW*|MSYS*|CYGWIN*|Windows*)
            # Windows - only modern terminals
            if [ -z "${WT_SESSION:-}" ]; then
                USE_EMOJI=false
            fi
            ;;
    esac

    # Environment variable override
    if [ "${WALGO_ASCII:-}" = "1" ] || [ "${WALGO_ASCII:-}" = "true" ]; then
        USE_EMOJI=false
    fi
    if [ "${WALGO_EMOJI:-}" = "1" ] || [ "${WALGO_EMOJI:-}" = "true" ]; then
        USE_EMOJI=true
    fi
}

# Detect interactive mode
is_interactive() {
    [ -t 0 ] && [ -t 1 ]
}

# Initialize terminal detection
detect_terminal

# ============================================================================
# COLORS & ICONS WITH FALLBACK
# ============================================================================

# Colors (only if supported)
if [ "$USE_COLORS" = true ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    NC=''
fi

# Icons with fallback
if [ "$USE_EMOJI" = true ]; then
    ICON_SUCCESS="✓"
    ICON_ERROR="✗"
    ICON_WARNING="⚠"
    ICON_INFO="ℹ"
else
    ICON_SUCCESS="[OK]"
    ICON_ERROR="[ERROR]"
    ICON_WARNING="[WARN]"
    ICON_INFO="[INFO]"
fi

# ============================================================================
# CONFIGURATION
# ============================================================================

REPO="selimozten/walgo"
BINARY_NAME="walgo"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
USE_SUDO="${USE_SUDO:-true}"
TEMP_DIR=""

# ============================================================================
# CLEANUP & ERROR HANDLING
# ============================================================================

cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR" 2>/dev/null || true
    fi
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# ============================================================================
# PRINT FUNCTIONS - IMPROVED
# ============================================================================

print_info() {
    if [ "$USE_COLORS" = true ]; then
        echo -e "${BLUE}${ICON_INFO}${NC} $1"
    else
        echo "${ICON_INFO} $1"
    fi
}

print_success() {
    if [ "$USE_COLORS" = true ]; then
        echo -e "${GREEN}${ICON_SUCCESS}${NC} $1"
    else
        echo "${ICON_SUCCESS} $1"
    fi
}

print_error() {
    if [ "$USE_COLORS" = true ]; then
        echo -e "${RED}${ICON_ERROR}${NC} $1" >&2
    else
        echo "${ICON_ERROR} $1" >&2
    fi
}

print_warning() {
    if [ "$USE_COLORS" = true ]; then
        echo -e "${YELLOW}${ICON_WARNING}${NC} $1"
    else
        echo "${ICON_WARNING} $1"
    fi
}

# ============================================================================
# UTILITY FUNCTIONS
# ============================================================================

# Portable sed in-place editing
sed_inplace() {
    local pattern="$1"
    local file="$2"

    if sed --version 2>/dev/null | grep -q GNU; then
        # GNU sed (Linux)
        sed -i "$pattern" "$file"
    else
        # BSD sed (macOS)
        sed -i '' "$pattern" "$file"
    fi
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
    TEMP_DIR=$(mktemp -d)
    local tmp_file="${TEMP_DIR}/${filename}"

    print_info "Downloading $BINARY_NAME from $download_url..."

    if command -v curl >/dev/null 2>&1; then
        curl -fsSL -o "$tmp_file" "$download_url"
    else
        wget -q -O "$tmp_file" "$download_url"
    fi

    print_success "Downloaded successfully"

    # Extract archive
    print_info "Extracting archive..."
    (
        cd "$TEMP_DIR"

        if [ "$OS" = "windows" ]; then
            unzip -q "$tmp_file"
        else
            tar -xzf "$tmp_file"
        fi
    )

    # Find the binary
    local binary_path="${TEMP_DIR}/${BINARY_NAME}"
    if [ "$OS" = "windows" ]; then
        binary_path="${binary_path}.exe"
    fi

    if [ ! -f "$binary_path" ]; then
        print_error "Binary not found in archive"
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

    # Clean up is handled by trap
    print_success "Installed $BINARY_NAME to $INSTALL_DIR"
}

# Install Desktop App
install_desktop() {
    print_info "Installing Walgo Desktop..."

    # Determine filename based on platform and architecture
    local filename=""
    case "$OS" in
        darwin)
            # macOS: architecture-specific builds (arm64 or amd64)
            filename="walgo-desktop_${VERSION}_darwin_${ARCH}.tar.gz"
            ;;
        windows)
            # Windows: architecture-specific builds (arm64 or amd64)
            filename="walgo-desktop_${VERSION}_windows_${ARCH}.zip"
            ;;
        linux)
            # Linux: only amd64 is supported for desktop app
            if [ "$ARCH" = "arm64" ]; then
                print_warning "Desktop app not available for Linux ARM64. Only CLI is supported."
                return
            fi
            filename="walgo-desktop_${VERSION}_linux_${ARCH}.tar.gz"
            ;;
        *)
            print_warning "Desktop app not available for platform: $OS"
            return
            ;;
    esac

    local download_url="https://github.com/$REPO/releases/download/v${VERSION}/${filename}"
    local tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${filename}"

    print_info "Downloading desktop app from $download_url..."

    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL -o "$tmp_file" "$download_url"; then
            print_warning "Failed to download desktop app. Skipping."
            rm -rf "$tmp_dir"
            return
        fi
    else
        if ! wget -q -O "$tmp_file" "$download_url"; then
            print_warning "Failed to download desktop app. Skipping."
            rm -rf "$tmp_dir"
            return
        fi
    fi

    print_info "Extracting..."
    (
        cd "$tmp_dir"

        # Extract based on file format
        case "$filename" in
            *.zip)
                if command -v unzip >/dev/null 2>&1; then
                    unzip -q "$tmp_file"
                else
                    print_warning "unzip not found. Cannot extract desktop app."
                    exit 1
                fi
                ;;
            *.tar.gz)
                tar -xzf "$tmp_file"
                ;;
        esac
    )

    # Check if extraction succeeded
    if [ $? -ne 0 ]; then
        rm -rf "$tmp_dir"
        return
    fi

    # Install based on platform
    case "$OS" in
        darwin)
            install_desktop_macos "$tmp_dir"
            ;;
        windows)
            install_desktop_windows "$tmp_dir"
            ;;
        linux)
            install_desktop_linux "$tmp_dir"
            ;;
    esac

    rm -rf "$tmp_dir"
}

# Install desktop app on macOS
install_desktop_macos() {
    local tmp_dir="$1"

    # Look for .app bundle
    local app_bundle=$(find "$tmp_dir" -maxdepth 2 -name "*.app" -type d | head -n1)

    if [ -z "$app_bundle" ]; then
        print_warning "Walgo.app not found in archive. Skipping desktop installation."
        return
    fi

    local app_name=$(basename "$app_bundle")
    local install_dir="/Applications"
    print_info "Moving $app_name to /Applications..."

    # Remove existing if present
    if [ -d "$install_dir/$app_name" ]; then
        if [ -w "$install_dir" ]; then
            rm -rf "$install_dir/$app_name"
        else
            sudo rm -rf "$install_dir/$app_name"
        fi
    fi

    # Move to /Applications
    if [ -w "$install_dir" ]; then
        mv "$app_bundle" "$install_dir/"
    else
        sudo mv "$app_bundle" "$install_dir/"
    fi

    # Remove quarantine attribute
    if [ -d "$install_dir/$app_name" ]; then
        xattr -d com.apple.quarantine "$install_dir/$app_name" 2>/dev/null || true
    fi

    print_success "Walgo Desktop installed to $install_dir/$app_name"
    echo ""
    print_info "macOS Security Note:"
    echo "  The app is not signed with an Apple Developer certificate."
    echo "  If macOS blocks it on first launch:"
    echo ""
    echo "  1. Right-click on Walgo.app → Open → Open again"
    echo "  2. Or go to System Settings → Privacy & Security → Open Anyway"
    echo ""
    print_info "Run with: walgo desktop"
}

# Install desktop app on Windows
install_desktop_windows() {
    local tmp_dir="$1"

    # Look for .exe file
    local exe_file=$(find "$tmp_dir" -maxdepth 2 -name "*.exe" -type f | head -n1)

    if [ -z "$exe_file" ]; then
        print_warning "Desktop executable not found in archive. Skipping desktop installation."
        return
    fi

    local install_dir="${LOCALAPPDATA}/Programs/Walgo"

    # Create directory
    mkdir -p "$install_dir"

    # Copy executable
    cp "$exe_file" "$install_dir/Walgo.exe"

    if [ -f "$install_dir/Walgo.exe" ]; then
        print_success "Walgo Desktop installed to $install_dir"
        print_info "Run with: walgo desktop"
    else
        print_warning "Failed to copy desktop app"
    fi
}

# Install desktop app on Linux
install_desktop_linux() {
    local tmp_dir="$1"

    # Look for Walgo binary specifically
    local binary=$(find "$tmp_dir" -maxdepth 2 -type f -name "Walgo" | head -n1)

    if [ -z "$binary" ]; then
        print_warning "Desktop binary not found in archive. Skipping desktop installation."
        return
    fi

    local install_dir="${HOME}/.local/bin"

    # Create directory
    mkdir -p "$install_dir"

    # Copy binary
    cp "$binary" "$install_dir/Walgo"
    chmod +x "$install_dir/Walgo"

    if [ -f "$install_dir/Walgo" ]; then
        print_success "Walgo Desktop installed to $install_dir"
        print_info "Run with: walgo desktop"
    else
        print_warning "Failed to copy desktop app"
    fi
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

# Install Hugo directly (without package manager)
install_hugo_direct() {
    print_info "Fetching latest Hugo version..."

    local hugo_version=""
    if command -v curl >/dev/null 2>&1; then
        hugo_version=$(curl -fsSL "https://api.github.com/repos/gohugoio/hugo/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        hugo_version=$(wget -qO- "https://api.github.com/repos/gohugoio/hugo/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found"
        return 1
    fi

    if [ -z "$hugo_version" ]; then
        print_warning "Failed to fetch latest version, using fallback v0.153.2"
        hugo_version="v0.153.2"
    fi

    hugo_version="${hugo_version#v}"  # Remove 'v' prefix if present
    print_success "Latest Hugo version: $hugo_version"

    local hugo_arch=""

    case "$ARCH" in
        amd64) hugo_arch="amd64" ;;
        arm64) hugo_arch="arm64" ;;
    esac

    local hugo_os=""
    case "$OS" in
        darwin) hugo_os="darwin" ;;
        linux) hugo_os="linux" ;;
        windows) hugo_os="windows" ;;
    esac

    local hugo_filename=""
    local use_pkg=false

    # macOS now uses universal binaries in .pkg format
    if [ "$OS" = "darwin" ]; then
        hugo_filename="hugo_extended_${hugo_version}_darwin-universal.pkg"
        use_pkg=true
    elif [ "$OS" = "windows" ]; then
        hugo_filename="hugo_extended_${hugo_version}_windows-${hugo_arch}.zip"
    else
        hugo_filename="hugo_extended_${hugo_version}_linux-${hugo_arch}.tar.gz"
    fi

    local hugo_url="https://github.com/gohugoio/hugo/releases/download/v${hugo_version}/${hugo_filename}"
    local tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${hugo_filename}"

    print_info "Downloading Hugo v${hugo_version}..."

    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL -o "$tmp_file" "$hugo_url"; then
            print_error "Failed to download Hugo"
            rm -rf "$tmp_dir"
            return 1
        fi
    else
        if ! wget -q -O "$tmp_file" "$hugo_url"; then
            print_error "Failed to download Hugo"
            rm -rf "$tmp_dir"
            return 1
        fi
    fi

    # Handle .pkg installation for macOS
    if [ "$use_pkg" = true ]; then
        print_info "Installing Hugo from .pkg..."

        if [ "$USE_SUDO" = "true" ] || [ ! -w "/usr/local" ]; then
            sudo installer -pkg "$tmp_file" -target /
        else
            installer -pkg "$tmp_file" -target /
        fi

        rm -rf "$tmp_dir"
        print_success "Hugo installed successfully"
        return 0
    fi

    # Extract archive for Linux/Windows
    print_info "Extracting Hugo..."
    (
        cd "$tmp_dir"

        if [ "$OS" = "windows" ]; then
            unzip -q "$tmp_file"
        else
            tar -xzf "$tmp_file"
        fi
    )

    local hugo_binary="${tmp_dir}/hugo"
    if [ "$OS" = "windows" ]; then
        hugo_binary="${tmp_dir}/hugo.exe"
    fi

    if [ ! -f "$hugo_binary" ]; then
        print_error "Hugo binary not found in archive"
        rm -rf "$tmp_dir"
        return 1
    fi

    print_info "Installing Hugo to $INSTALL_DIR..."

    if [ "$USE_SUDO" = "true" ] && [ ! -w "$INSTALL_DIR" ]; then
        sudo install -m 755 "$hugo_binary" "$INSTALL_DIR/hugo"
    else
        install -m 755 "$hugo_binary" "$INSTALL_DIR/hugo"
    fi

    rm -rf "$tmp_dir"

    print_success "Hugo installed successfully"
    return 0
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

        # Offer to install Hugo
        if [[ " ${missing_deps[@]} " =~ " hugo " ]]; then
            case "$OS" in
                darwin)
                    if command -v brew >/dev/null 2>&1; then
                        print_info "Homebrew detected. Install Hugo? [y/N]"
                        if [ -t 0 ]; then
                            read -r response
                        else
                            read -r response < /dev/tty 2>/dev/null || response="n"
                        fi
                        if [[ "$response" =~ ^[Yy]$ ]]; then
                            print_info "Installing Hugo via Homebrew..."
                            brew install hugo
                            if command -v hugo >/dev/null 2>&1; then
                                print_success "Hugo installed successfully"
                            fi
                        fi
                    else
                        print_warning "Homebrew not found"
                        print_info "Would you like to:"
                        echo "  1) Install Hugo directly (recommended)"
                        echo "  2) Install Homebrew first, then Hugo"
                        echo "  3) Skip Hugo installation"
                        if [ -t 0 ]; then
                            read -r -p "Choose option [1-3]: " choice
                        else
                            read -r choice < /dev/tty 2>/dev/null || choice="1"
                        fi

                        case "$choice" in
                            1)
                                install_hugo_direct
                                ;;
                            2)
                                print_info "Installing Homebrew..."
                                /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
                                if command -v brew >/dev/null 2>&1; then
                                    print_success "Homebrew installed"
                                    print_info "Installing Hugo..."
                                    brew install hugo
                                fi
                                ;;
                            *)
                                print_info "Skipping Hugo installation"
                                echo "  You can install it later with: brew install hugo"
                                echo "  Or download from: https://gohugo.io/installation/"
                                ;;
                        esac
                    fi
                    ;;
                linux)
                    print_info "Install Hugo? [y/N]"
                    if [ -t 0 ]; then
                        read -r response
                    else
                        read -r response < /dev/tty 2>/dev/null || response="n"
                    fi
                    if [[ "$response" =~ ^[Yy]$ ]]; then
                        if command -v apt-get >/dev/null 2>&1; then
                            print_info "Installing Hugo via apt..."
                            sudo apt-get update && sudo apt-get install -y hugo
                        elif command -v dnf >/dev/null 2>&1; then
                            print_info "Installing Hugo via dnf..."
                            sudo dnf install -y hugo
                        elif command -v pacman >/dev/null 2>&1; then
                            print_info "Installing Hugo via pacman..."
                            sudo pacman -S hugo
                        else
                            print_info "Package manager not detected. Installing Hugo directly..."
                            install_hugo_direct
                        fi
                    else
                        echo "  # Ubuntu/Debian:"
                        echo "    sudo apt install hugo"
                        echo ""
                        echo "  # Or download from: https://gohugo.io/installation/"
                    fi
                    ;;
                windows)
                    if command -v choco >/dev/null 2>&1; then
                        print_info "Install Hugo via Chocolatey? [y/N]"
                        if [ -t 0 ]; then
                            read -r response
                        else
                            read -r response < /dev/tty 2>/dev/null || response="n"
                        fi
                        if [[ "$response" =~ ^[Yy]$ ]]; then
                            choco install hugo-extended -y
                        fi
                    else
                        print_info "Installing Hugo directly..."
                        install_hugo_direct
                    fi
                    ;;
            esac
        fi

        echo ""
        print_info "You can also use 'walgo setup-deps' to install Walrus dependencies"
    else
        print_success "All optional dependencies found"
    fi
}

# Add PATH to shell profile - Professional multi-shell/multi-OS support
add_path_to_profile() {
    local path_to_add="$HOME/.local/bin"
    local path_line='export PATH="$HOME/.local/bin:$PATH"'

    # Detect current shell
    local current_shell=""
    if [ -n "$ZSH_VERSION" ]; then
        current_shell="zsh"
    elif [ -n "$BASH_VERSION" ]; then
        current_shell="bash"
    elif [ -n "$FISH_VERSION" ]; then
        current_shell="fish"
    else
        # Fallback to $SHELL variable
        current_shell=$(basename "$SHELL" 2>/dev/null || echo "unknown")
    fi

    print_info "Detected shell: $current_shell"

    # Determine all profile files to check/update based on OS and shell
    local profiles_to_check=()
    local fish_config="$HOME/.config/fish/config.fish"

    if [ "$OS" = "darwin" ]; then
        # macOS - zsh is default since Catalina, but bash still common
        case "$current_shell" in
            zsh)
                # Login shells read .zprofile, interactive read .zshrc
                # Add to both to cover all cases
                profiles_to_check=(
                    "$HOME/.zprofile"
                    "$HOME/.zshrc"
                )
                ;;
            bash)
                # Login shells read .bash_profile or .profile
                # Interactive read .bashrc
                profiles_to_check=(
                    "$HOME/.bash_profile"
                    "$HOME/.bashrc"
                    "$HOME/.profile"
                )
                ;;
            fish)
                profiles_to_check=("$fish_config")
                ;;
            *)
                # Unknown shell - try common files
                profiles_to_check=(
                    "$HOME/.profile"
                    "$HOME/.zprofile"
                    "$HOME/.bash_profile"
                )
                ;;
        esac
    else
        # Linux/Unix
        case "$current_shell" in
            zsh)
                profiles_to_check=(
                    "$HOME/.zshrc"
                    "$HOME/.zprofile"
                )
                ;;
            bash)
                profiles_to_check=(
                    "$HOME/.bashrc"
                    "$HOME/.bash_profile"
                    "$HOME/.profile"
                )
                ;;
            fish)
                profiles_to_check=("$fish_config")
                ;;
            ksh|sh)
                profiles_to_check=(
                    "$HOME/.profile"
                )
                ;;
            *)
                # Unknown shell - try common files
                profiles_to_check=(
                    "$HOME/.profile"
                    "$HOME/.bashrc"
                )
                ;;
        esac
    fi

    # Function to check if PATH is already configured in a file
    path_already_configured() {
        local file="$1"
        [ -f "$file" ] && grep -qE "^\s*(export\s+)?PATH=.*\.local/bin|^\s*set\s+-gx\s+PATH.*\.local/bin|^\s*fish_add_path.*\.local/bin" "$file"
    }

    # Function to add PATH to a profile file
    add_to_file() {
        local file="$1"
        local dir=$(dirname "$file")

        # Create directory if needed (for fish config)
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir" 2>/dev/null || return 1
        fi

        # Check if already configured
        if path_already_configured "$file"; then
            print_info "PATH already configured in $(basename $file)"
            return 0
        fi

        # Add PATH based on file type
        if [[ "$file" == *"fish"* ]]; then
            # Fish shell uses different syntax
            {
                echo ""
                echo "# Added by Walgo installer - Sui/Walrus tools"
                echo "fish_add_path -p $path_to_add"
            } >> "$file"
        else
            # POSIX-compatible shells (bash, zsh, sh, ksh)
            {
                echo ""
                echo "# Added by Walgo installer - Sui/Walrus tools"
                echo "$path_line"
            } >> "$file"
        fi

        print_success "Added ~/.local/bin to PATH in $(basename $file)"
        return 0
    }

    # Try to add PATH to appropriate profile files
    local added_count=0
    local primary_profile=""

    for profile in "${profiles_to_check[@]}"; do
        # Skip if file doesn't exist and we've already added to one file
        if [ ! -f "$profile" ] && [ $added_count -gt 0 ]; then
            continue
        fi

        # If file exists, update it
        if [ -f "$profile" ]; then
            if add_to_file "$profile"; then
                added_count=$((added_count + 1))
                [ -z "$primary_profile" ] && primary_profile="$profile"
            fi
        else
            # Create the first file in the list if nothing exists
            if [ $added_count -eq 0 ]; then
                if add_to_file "$profile"; then
                    added_count=$((added_count + 1))
                    primary_profile="$profile"
                fi
            fi
        fi
    done

    # Fallback: if nothing was added, use .profile
    if [ $added_count -eq 0 ]; then
        print_warning "Could not detect shell profile, using ~/.profile"
        add_to_file "$HOME/.profile"
        primary_profile="$HOME/.profile"
    fi

    # Show instructions based on what was added
    echo ""
    print_info "To activate PATH changes, choose ONE of these:"
    echo ""
    echo "  1. Restart your terminal (recommended)"
    echo "  2. Run: exec \$SHELL"
    if [ -n "$primary_profile" ]; then
        echo "  3. Run: source $primary_profile"
    fi
    echo ""
}

# Install Walrus dependencies via suiup
install_walrus_dependencies() {
    echo ""
    print_info "Would you like to install Walrus dependencies (Sui, Walrus CLI, site-builder)? [y/N]"

    # Read from /dev/tty to work with piped installation
    if [ -t 0 ]; then
        read -r response
    else
        read -r response < /dev/tty 2>/dev/null || response="n"
    fi

    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        print_info "Skipping Walrus dependencies installation"
        print_info "You can install them later with: walgo setup-deps"
        return
    fi

    # Install suiup
    print_info "Installing suiup..."
    if ! command -v suiup >/dev/null 2>&1 && ! [ -f "$HOME/.local/bin/suiup" ]; then
        # Ensure we're in a safe directory before running external installer
        cd "$HOME" 2>/dev/null || cd /tmp
        curl -sSfL https://raw.githubusercontent.com/Mystenlabs/suiup/main/install.sh | sh

        # Add to PATH for current session
        export PATH="$HOME/.local/bin:$PATH"

        if [ -f "$HOME/.local/bin/suiup" ]; then
            print_success "suiup installed successfully"
            # Add PATH to shell profile for future sessions
            add_path_to_profile
        else
            print_warning "suiup installation may have failed"
            return
        fi
    else
        # Ensure PATH includes ~/.local/bin
        export PATH="$HOME/.local/bin:$PATH"
        print_success "suiup already installed"
    fi

    # Automatically install both testnet and mainnet
    echo ""
    print_info "Installing Walrus dependencies for both testnet and mainnet..."

    local install_testnet=true
    local install_mainnet=true
    local default_network="testnet"

    # Install Sui
    print_info "Installing Sui CLI..."
    if $install_testnet; then
        "$HOME/.local/bin/suiup" install sui@testnet || print_warning "Failed to install Sui testnet"
    fi
    if $install_mainnet; then
        "$HOME/.local/bin/suiup" install sui@mainnet || print_warning "Failed to install Sui mainnet"
    fi

    # Install Walrus
    print_info "Installing Walrus CLI..."
    if $install_testnet; then
        "$HOME/.local/bin/suiup" install walrus@testnet || print_warning "Failed to install Walrus testnet"
    fi
    if $install_mainnet; then
        "$HOME/.local/bin/suiup" install walrus@mainnet || print_warning "Failed to install Walrus mainnet"
    fi

    # Install site-builder (only available as mainnet binary, works for all networks via config)
    print_info "Installing site-builder..."
    "$HOME/.local/bin/suiup" install site-builder@mainnet || print_warning "Failed to install site-builder"

    # Wait a moment for installations to complete
    sleep 1

    # Set default binaries based on network preference
    # This creates/updates symlinks in ~/.local/bin/
    print_info "Setting default binaries..."
    if [ "$default_network" = "mainnet" ]; then
        "$HOME/.local/bin/suiup" default set sui@mainnet 2>/dev/null || true
        "$HOME/.local/bin/suiup" default set walrus@mainnet 2>/dev/null || true
    else
        "$HOME/.local/bin/suiup" default set sui@testnet 2>/dev/null || true
        "$HOME/.local/bin/suiup" default set walrus@testnet 2>/dev/null || true
    fi
    # site-builder only has mainnet binary (works for all networks via config)
    "$HOME/.local/bin/suiup" default set site-builder@mainnet 2>/dev/null || true

    # Verify all binaries exist in ~/.local/bin/ and retry if missing
    local binaries_ok=true

    if [ ! -f "$HOME/.local/bin/sui" ]; then
        print_warning "sui binary not found in ~/.local/bin/, retrying default set..."
        "$HOME/.local/bin/suiup" default set "sui@$default_network" 2>/dev/null || binaries_ok=false
    fi

    if [ ! -f "$HOME/.local/bin/walrus" ]; then
        print_warning "walrus binary not found in ~/.local/bin/, retrying default set..."
        "$HOME/.local/bin/suiup" default set "walrus@$default_network" 2>/dev/null || binaries_ok=false
    fi

    if [ ! -f "$HOME/.local/bin/site-builder" ]; then
        print_warning "site-builder binary not found in ~/.local/bin/, retrying default set..."
        "$HOME/.local/bin/suiup" default set "site-builder@mainnet" 2>/dev/null || binaries_ok=false
    fi

    # Final verification
    echo ""
    print_info "Verifying installations..."

    if [ -f "$HOME/.local/bin/sui" ]; then
        print_success "Sui CLI installed: $($HOME/.local/bin/sui --version 2>&1 | head -1)"
    else
        print_error "Sui CLI not found in ~/.local/bin/"
        binaries_ok=false
    fi

    if [ -f "$HOME/.local/bin/walrus" ]; then
        print_success "Walrus CLI installed: $($HOME/.local/bin/walrus --version 2>&1 | head -1)"
    else
        print_error "Walrus CLI not found in ~/.local/bin/"
        binaries_ok=false
    fi

    if [ -f "$HOME/.local/bin/site-builder" ]; then
        print_success "site-builder installed: $($HOME/.local/bin/site-builder --version 2>&1 | grep -o 'site-builder [0-9.]*' | head -1)"
    else
        print_error "site-builder not found in ~/.local/bin/"
        binaries_ok=false
    fi

    if [ "$binaries_ok" = false ]; then
        echo ""
        print_warning "Some binaries were not properly installed."
        print_info "Try running: suiup default set sui@$default_network walrus@$default_network site-builder@$default_network"
    fi

    # Configure Sui client
    echo ""
    print_info "Configuring Sui client..."

    if [ ! -f "$HOME/.sui/sui_config/client.yaml" ]; then
        print_info "Initializing Sui client for $default_network..."

        # Ensure we're in a safe directory before running sui client
        cd "$HOME" 2>/dev/null || cd /tmp

        # Automatically initialize sui client with piped input
        print_info "Automatically configuring Sui client..."
        {
            echo "y"  # Connect to Sui Full Node
            echo "https://fullnode.$default_network.sui.io:443"  # Full node URL
            echo "$default_network"  # Environment alias
            echo "0"  # Key scheme (ed25519)
        } | sui client 2>&1 | grep -v "^Usage:" | head -50 || true
    else
        print_success "Sui client already configured"
    fi

    # Download Walrus configuration
    print_info "Downloading Walrus configuration..."
    mkdir -p "$HOME/.config/walrus"

    if curl --create-dirs -fsSL https://docs.wal.app/setup/client_config.yaml -o "$HOME/.config/walrus/client_config.yaml"; then
        print_success "Walrus client config downloaded"
    else
        print_warning "Failed to download Walrus client config"
    fi

    # Download site-builder configuration
    print_info "Downloading site-builder configuration..."

    if curl -fsSL https://raw.githubusercontent.com/MystenLabs/walrus-sites/refs/heads/mainnet/sites-config.yaml -o "$HOME/.config/walrus/sites-config.yaml"; then
        print_success "site-builder config downloaded"
    else
        print_warning "Failed to download site-builder config"
    fi

    # Update default context in configs if needed
    if [ "$default_network" = "mainnet" ]; then
        if [ -f "$HOME/.config/walrus/client_config.yaml" ]; then
            sed_inplace 's/default_context: testnet/default_context: mainnet/' "$HOME/.config/walrus/client_config.yaml" 2>/dev/null || true
        fi
        if [ -f "$HOME/.config/walrus/sites-config.yaml" ]; then
            sed_inplace 's/default_context: testnet/default_context: mainnet/' "$HOME/.config/walrus/sites-config.yaml" 2>/dev/null || true
        fi
    fi

    echo ""
    print_success "Walrus dependencies installation complete!"
    echo ""
    print_info "Next steps:"
    echo "  1. Verify installation:"
    echo "     sui --version"
    echo "     walrus --version"
    echo "     site-builder --version"
    echo ""
    echo "  2. Fund your Sui account (for $default_network):"
    if [ "$default_network" = "testnet" ]; then
        echo "     Visit: https://faucet.sui.io/"
        echo "     Get your address: sui client active-address"
        echo "     Get WAL tokens: walrus get-wal --context testnet"
    else
        echo "     Buy SUI from an exchange and send to: \$(sui client active-address)"
    fi
    echo ""
}

# Post-install instructions
show_next_steps() {
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    print_success "Walgo installation complete!"
    echo "═══════════════════════════════════════════════════════════"
    echo ""

    # Determine shell profile for restart instructions
    local shell_profile=""
    if [ -n "$ZSH_VERSION" ] || [ "$SHELL" = "/bin/zsh" ] || [ "$SHELL" = "/usr/bin/zsh" ]; then
        shell_profile="~/.zshrc"
    elif [ -n "$BASH_VERSION" ] || [ "$SHELL" = "/bin/bash" ] || [ "$SHELL" = "/usr/bin/bash" ]; then
        shell_profile="~/.bashrc"
        if [ "$OS" = "darwin" ] && [ -f "$HOME/.bash_profile" ]; then
            shell_profile="~/.bash_profile"
        fi
    else
        shell_profile="~/.profile"
    fi

    if [ "$USE_COLORS" = true ]; then
        echo -e "${YELLOW}╔═══════════════════════════════════════════════════════════╗${NC}"
        echo -e "${YELLOW}║         ${RED}IMPORTANT: RESTART YOUR TERMINAL NOW${YELLOW}             ║${NC}"
        echo -e "${YELLOW}╚═══════════════════════════════════════════════════════════╝${NC}"
    else
        echo "╔═══════════════════════════════════════════════════════════╗"
        echo "║         IMPORTANT: RESTART YOUR TERMINAL NOW             ║"
        echo "╚═══════════════════════════════════════════════════════════╝"
    fi
    echo ""
    print_warning "To use sui, walrus, and site-builder commands, you MUST:"
    echo ""
    echo "  Option 1 (Recommended):"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}exec \$SHELL${NC}            (restart current shell)"
    else
        echo "     exec \$SHELL            (restart current shell)"
    fi
    echo ""
    echo "  Option 2:"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}source $shell_profile${NC}   (reload shell config)"
    else
        echo "     source $shell_profile   (reload shell config)"
    fi
    echo ""
    echo "  Option 3:"
    echo "     Open a new terminal window"
    echo ""
    print_warning "Until you restart, 'sui' command will show: command not found"
    echo ""
    echo "═══════════════════════════════════════════════════════════"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Verify installation:"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo --help${NC}"
    else
        echo "     walgo --help"
    fi
    echo ""
    echo "  2. Create your first site:"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo init my-site${NC}"
        echo -e "     ${GREEN}cd my-site${NC}"
    else
        echo "     walgo init my-site"
        echo "     cd my-site"
    fi
    echo ""
    echo "  3. Build your site:"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo build${NC}"
    else
        echo "     walgo build"
    fi
    echo ""
    echo "  4. Deploy with the interactive wizard (recommended):"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo launch${NC}"
    else
        echo "     walgo launch"
    fi
    echo ""
    echo "     The wizard guides you through network selection, wallet setup,"
    echo "     and deployment with cost estimation."
    echo ""
    echo "     Alternative: HTTP deployment for testing (no wallet required):"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo deploy-http \\\\${NC}"
        echo -e "       ${GREEN}--publisher https://publisher.walrus-testnet.walrus.space \\\\${NC}"
        echo -e "       ${GREEN}--aggregator https://aggregator.walrus-testnet.walrus.space${NC}"
    else
        echo "     walgo deploy-http \\"
        echo "       --publisher https://publisher.walrus-testnet.walrus.space \\"
        echo "       --aggregator https://aggregator.walrus-testnet.walrus.space"
    fi
    echo ""
    echo "  5. To uninstall walgo later:"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo uninstall${NC}"
    else
        echo "     walgo uninstall"
    fi
    echo ""
    echo "  6. To launch the desktop app:"
    if [ "$USE_COLORS" = true ]; then
        echo -e "     ${GREEN}walgo desktop${NC}"
    else
        echo "     walgo desktop"
    fi
    echo ""
    echo "     Desktop app installed to standard location:"
    case "$OS" in
        darwin)
            echo "       /Applications/Walgo.app"
            ;;
        windows)
            echo "       %LOCALAPPDATA%\\Programs\\Walgo\\Walgo.exe"
            ;;
        linux)
            echo "       ~/.local/bin/Walgo"
            ;;
    esac
    echo ""
    echo "Documentation: https://github.com/$REPO"
    echo ""
}

# Main installation flow
main() {
    # Ensure we start in a safe directory
    cd "$HOME" 2>/dev/null || cd /tmp || true

    echo ""
    echo "╔═══════════════════════════════════════════════════════════╗"
    echo "║                   Walgo Installer                         ║"
    echo "║    Ship static sites to Walrus decentralized storage     ║"
    echo "╚═══════════════════════════════════════════════════════════╝"
    echo ""

    detect_platform
    get_latest_version
    install_binary
    install_desktop
    verify_installation
    check_dependencies
    install_walrus_dependencies
    show_next_steps
}

# Run main function
main
