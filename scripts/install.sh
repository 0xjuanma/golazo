#!/bin/sh
set -e

# Check if output is a TTY for color support
if [ -t 1 ]; then
    # Colors for output (matching app theme: cyan accent, red secondary)
    CYAN='\033[0;36m'      # Bright cyan (accent color)
    BRIGHT_CYAN='\033[1;36m' # Bold cyan
    RED='\033[0;31m'       # Red (for errors)
    BRIGHT_RED='\033[1;31m' # Bright red
    GREEN='\033[0;32m'     # Green (for success)
    BRIGHT_GREEN='\033[1;32m' # Bright green
    YELLOW='\033[1;33m'    # Yellow (for warnings)
    NC='\033[0m'           # No Color
else
    # No colors if not a TTY
    CYAN=''
    BRIGHT_CYAN=''
    RED=''
    BRIGHT_RED=''
    GREEN=''
    BRIGHT_GREEN=''
    YELLOW=''
    NC=''
fi

# ASCII art logo (stylized block characters - wide format)
ASCII_LOGO="╱╱╱╱ ▄▀▀▀▀  ▄▀▀▀▄ █     ▄▀▀▀▀▀▄  ▀▀▀▀█ ▄▀▀▀▄ ╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱
╱╱╱╱ █   ▀█ █   █ █    █▀▀▀▀▀▀▀█  █▀▀  █   █ ╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱ 
╱╱╱╱  ▀▀▀▀   ▀▀▀  ▀▀▀▀ ▀       ▀ ▀▀▀▀▀  ▀▀▀  ╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱╱  "

REPO="0xjuanma/golazo"
BINARY_NAME="golazo"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Print ASCII art header with cyan color
printf "${BRIGHT_CYAN}${ASCII_LOGO}${NC}\n\n"
printf "${BRIGHT_GREEN}Installing ${BINARY_NAME}...${NC}\n\n"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) printf "${BRIGHT_RED}Unsupported architecture: ${ARCH}${NC}\n"; exit 1 ;;
esac

# Check for supported OS
case "$OS" in
    darwin|linux) ;;
    *) printf "${BRIGHT_RED}Unsupported operating system: ${OS}${NC}\n"; exit 1 ;;
esac

printf "${CYAN}Detected: ${OS}/${ARCH}${NC}\n"

# Get the latest release tag
printf "${CYAN}Fetching latest release...${NC}\n"
LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | cut -d'"' -f4)

if [ -z "$LATEST" ]; then
    printf "${BRIGHT_RED}Failed to fetch latest release${NC}\n"
    exit 1
fi

printf "${CYAN}Latest version: ${LATEST}${NC}\n"

# Construct download URL
URL="https://github.com/$REPO/releases/download/$LATEST/${BINARY_NAME}-${OS}-${ARCH}"

# Download the binary
printf "${CYAN}Downloading ${BINARY_NAME} ${LATEST} for ${OS}/${ARCH}...${NC}\n"
if ! curl -sL "$URL" -o "$BINARY_NAME"; then
    printf "${BRIGHT_RED}Failed to download binary${NC}\n"
    exit 1
fi

chmod +x "$BINARY_NAME"

# Determine install location and install
printf "${CYAN}Installing to ${INSTALL_DIR}...${NC}\n"

if [ -w "$INSTALL_DIR" ]; then
    mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
else
    # Try with sudo if we don't have write access
    if command -v sudo >/dev/null 2>&1; then
        sudo mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
    else
        # Fallback to user's local bin
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
        mv "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        printf "${YELLOW}Installed to ${INSTALL_DIR} (no sudo available)${NC}\n"
    fi
fi

# Verify installation
if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
    # Check if the binary is in PATH
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        printf "\n${YELLOW}Warning: ${BINARY_NAME} may not be in your PATH.${NC}\n"
        printf "${YELLOW}Add ${INSTALL_DIR} to your PATH if needed:${NC}\n"
        printf "${CYAN}  export PATH=\"\$PATH:${INSTALL_DIR}\"${NC}\n"
    fi

    printf "\n${BRIGHT_GREEN}✓ ${BINARY_NAME} ${LATEST} installed successfully!${NC}\n"
    printf "${BRIGHT_GREEN}Run '${BINARY_NAME}' to start watching live football matches.${NC}\n"
else
    printf "${BRIGHT_RED}Installation failed${NC}\n"
    exit 1
fi
