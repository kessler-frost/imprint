#!/bin/sh
set -e

REPO="kessler-frost/imprint"
INSTALL_DIR="${IMPRINT_INSTALL_DIR:-$HOME/.local/bin}"

# Parse arguments
UNINSTALL=false
for arg in "$@"; do
  case $arg in
    --uninstall) UNINSTALL=true ;;
  esac
done

# Uninstall mode
if [ "$UNINSTALL" = true ]; then
  echo "Uninstalling imprint..."
  rm -f "$INSTALL_DIR/imprint"
  echo "imprint removed from $INSTALL_DIR"

  printf "Remove ttyd as well? [y/N] "
  read -r response
  case "$response" in
    [yY][eE][sS]|[yY])
      OS=$(uname -s | tr '[:upper:]' '[:lower:]')
      case $OS in
        darwin)
          echo "Uninstalling ttyd via Homebrew..."
          brew uninstall ttyd 2>/dev/null && echo "ttyd uninstalled" || echo "ttyd not found in Homebrew"
          ;;
        linux)
          echo "Please manually remove ttyd using your package manager or delete the binary"
          ;;
      esac
      ;;
    *)
      echo "Keeping ttyd installed"
      ;;
  esac

  echo "Uninstall complete!"
  exit 0
fi

# Install mode
echo "Installing imprint..."

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case $OS in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

echo "Detected: $OS/$ARCH"

# Check/install ttyd
if ! command -v ttyd >/dev/null 2>&1; then
  echo "ttyd not found. Installing..."

  case $OS in
    darwin)
      command -v brew >/dev/null 2>&1 || { echo "Error: Homebrew required but not installed. Visit https://brew.sh"; exit 1; }
      brew install ttyd
      ;;
    linux)
      echo "Installing ttyd from GitHub releases..."
      TTYD_VERSION=$(curl -sL https://api.github.com/repos/tsl0922/ttyd/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
      TTYD_URL="https://github.com/tsl0922/ttyd/releases/download/${TTYD_VERSION}/ttyd.${ARCH}"

      mkdir -p "$INSTALL_DIR"
      curl -fsSL "$TTYD_URL" -o "$INSTALL_DIR/ttyd"
      chmod +x "$INSTALL_DIR/ttyd"
      echo "ttyd installed to $INSTALL_DIR/ttyd"
      ;;
  esac
else
  echo "ttyd already installed: $(command -v ttyd)"
fi

# Get latest imprint release
echo "Fetching latest imprint release..."
LATEST_VERSION=$(curl -sL https://api.github.com/repos/$REPO/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
BINARY_NAME="imprint-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/$REPO/releases/download/${LATEST_VERSION}/${BINARY_NAME}"

echo "Downloading imprint ${LATEST_VERSION}..."
mkdir -p "$INSTALL_DIR"
curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/imprint"
chmod +x "$INSTALL_DIR/imprint"

# Verify installation
if "$INSTALL_DIR/imprint" --version >/dev/null 2>&1; then
  echo "imprint installed successfully to $INSTALL_DIR/imprint"
  VERSION=$("$INSTALL_DIR/imprint" --version 2>&1)
  echo "Version: $VERSION"
else
  echo "Error: Installation verification failed"
  exit 1
fi

# Check if install dir is in PATH
case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    echo "Installation complete! Run 'imprint --help' to get started."
    ;;
  *)
    echo ""
    echo "Installation complete!"
    echo ""
    echo "NOTE: $INSTALL_DIR is not in your PATH."
    echo "Add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
    echo ""
    echo "Then restart your shell or run: source ~/.bashrc"
    ;;
esac
