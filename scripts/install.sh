#!/bin/sh
set -e

# Repository
REPO="rztaylor/GoDotFiles"
BINARY="gdf"
DEST="${DEST:-/usr/local/bin}"

# Detect OS and Arch
OS=$(uname -s)
ARCH=$(uname -m)

# Map OS
case "$OS" in
    Linux)  OS="Linux" ;;
    Darwin) OS="Darwin" ;;
    *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Map Arch
case "$ARCH" in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported Arch: $ARCH"; exit 1 ;;
esac

echo "Platform: $OS/$ARCH"

# Determine Latest Version
TAG=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$TAG" ]; then
    echo "Error: Could not find latest release."
    exit 1
fi

echo "Latest Version: $TAG"

# Strip 'v' prefix from tag

VERSION=${TAG#v}

FILENAME="${BINARY}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$TAG/$FILENAME"

echo "Downloading $URL..."

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

curl -sfL "$URL" -o "$TMP_DIR/$FILENAME"

echo "Extracting..."
tar -xzf "$TMP_DIR/$FILENAME" -C "$TMP_DIR"

echo "Installing to $DEST..."
if [ -w "$DEST" ]; then
    mv "$TMP_DIR/$BINARY" "$DEST/$BINARY"
else
    sudo mv "$TMP_DIR/$BINARY" "$DEST/$BINARY"
fi

chmod +x "$DEST/$BINARY"

echo "Done! Run '$BINARY version' to verify."
