#!/bin/bash
# Log Monster Detector - Installation Script
# Usage: curl -sSL https://raw.githubusercontent.com/thiruk/logmonster/main/install.sh | bash

set -e

# Configuration
REPO="thiruk/logmonster"
BINARY_NAME="logmonster"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)

    case "$OS" in
        linux) OS="linux" ;;
        darwin) OS="darwin" ;;
        *) error "Unsupported OS: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) error "Unsupported architecture: $ARCH" ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get latest release version
get_latest_version() {
    info "Fetching latest release..."
    VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        error "Failed to fetch latest version. Check your internet connection."
    fi
    
    info "Latest version: $VERSION"
}

# Download and install
download_and_install() {
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${PLATFORM}"
    TMP_FILE=$(mktemp)

    info "Downloading from: $DOWNLOAD_URL"
    
    if ! curl -sSL -o "$TMP_FILE" "$DOWNLOAD_URL"; then
        rm -f "$TMP_FILE"
        error "Failed to download binary"
    fi

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    else
        warn "Need sudo to install to $INSTALL_DIR"
        sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    fi

    info "Installed to: ${INSTALL_DIR}/${BINARY_NAME}"
}

# Verify installation
verify_install() {
    if command -v "$BINARY_NAME" &> /dev/null; then
        info "Installation successful!"
        echo ""
        $BINARY_NAME --version
        echo ""
        info "Run 'logmonster --help' to get started"
    else
        warn "Binary installed but not in PATH"
        info "Add ${INSTALL_DIR} to your PATH or run: ${INSTALL_DIR}/${BINARY_NAME}"
    fi
}

# Main
main() {
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║         LOG MONSTER DETECTOR - INSTALLER                   ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""

    detect_platform
    get_latest_version
    download_and_install
    verify_install
}

main
