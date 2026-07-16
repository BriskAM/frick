#!/bin/bash
set -e

# frick installer script

OWNER="BriskAM"
REPO="frick"

# Check OS and Architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
	ARCH="amd64"
elif [ "$ARCH" = "arm64" ] || [ "$ARCH" = "aarch64" ]; then
	ARCH="arm64"
fi

echo "Installing frick..."

# Try to download precompiled binary from GitHub Releases
LATEST_RELEASE_URL="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
DOWNLOAD_URL=""

# Fetch latest release data
RELEASE_DATA=$(curl -s "${LATEST_RELEASE_URL}")
if echo "${RELEASE_DATA}" | grep -q "browser_download_url"; then
	# Parse release assets for matching OS and ARCH (e.g. frick-darwin-amd64)
	BINARY_NAME="frick-${OS}-${ARCH}"
	DOWNLOAD_URL=$(echo "${RELEASE_DATA}" | grep "browser_download_url" | grep "${BINARY_NAME}" | cut -d '"' -f 4 | head -n 1)
fi

INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
	echo "Warning: ${INSTALL_DIR} is not writable. Installing to local bin..."
	INSTALL_DIR="${HOME}/.local/bin"
	mkdir -p "${INSTALL_DIR}"
fi

if [ -n "${DOWNLOAD_URL}" ]; then
	echo "Downloading precompiled binary from: ${DOWNLOAD_URL}"
	curl -L -o "${INSTALL_DIR}/frick" "${DOWNLOAD_URL}"
	chmod +x "${INSTALL_DIR}/frick"
	echo "frick installed successfully to ${INSTALL_DIR}/frick!"
else
	# Fallback to Go build if Go is installed
	if command -v go >/dev/null 2>&1; then
		echo "No precompiled binary found for ${OS}-${ARCH}. Building from source using Go..."
		TEMP_DIR=$(mktemp -d)
		git clone --depth 1 "https://github.com/BriskAM/frick.git" "${TEMP_DIR}"
		cd "${TEMP_DIR}"
		go build -o frick cmd/frick/main.go
		mv frick "${INSTALL_DIR}/frick"
		chmod +x "${INSTALL_DIR}/frick"
		rm -rf "${TEMP_DIR}"
		echo "frick built and installed successfully to ${INSTALL_DIR}/frick!"
	else
		echo "Error: No precompiled binary found, and 'go' is not installed."
		echo "Please install Go or download a binary manually from https://github.com/BriskAM/frick/releases"
		exit 1
	fi
fi

# Print shell integration reminder
echo ""
echo "=== Shell Integration ==="
echo "To automatically configure your shell, run the following command:"
echo "  [ -n \"\$ZSH_VERSION\" ] && echo 'eval \"\$(frick init)\"' >> ~/.zshrc || echo 'eval \"\$(frick init)\"' >> ~/.bashrc"
echo ""
echo "Then, configure your API key:"
echo "  frick configure"
