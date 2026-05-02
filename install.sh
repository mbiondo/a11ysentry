#!/bin/bash
# A11ySentry Unix Smart Installer
# Detects OS/Arch, downloads the latest binary, 
# and registers MCP in Claude Desktop, Cursor, etc.

set -e

OWNER="mbiondo"
REPO="a11ysentry"
BINARY_NAME="a11ysentry"

echo -e "\033[0;36m🛡️  A11ySentry: Starting smart installation for Unix...\033[0m"

# 1. Detect OS and Arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="amd64" ;;
    arm64|aarch64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# 2. Setup Directories
INSTALL_DIR="$HOME/.a11ysentry/bin"
mkdir -p "$INSTALL_DIR"

# 3. Download from GitHub or use local if in DEV MODE
BINARY_FULL="$INSTALL_DIR/$BINARY_NAME"
LOCAL_BINARY="./cmd/a11ysentry/a11ysentry"

if [ -f "$LOCAL_BINARY" ]; then
    cp "$LOCAL_BINARY" "$BINARY_FULL"
    chmod +x "$BINARY_FULL"
    echo -e "\033[0;33m🚧 DEV MODE: Copied binary from local workspace.\033[0m"
else
    echo -e "\033[0;36m📥 Downloading latest release from GitHub...\033[0m"
    
    # Get latest release tag
    LATEST_RELEASE_JSON=$(curl -s "https://api.github.com/repos/$OWNER/$REPO/releases/latest")
    TAG=$(echo "$LATEST_RELEASE_JSON" | grep -m 1 '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$TAG" ]; then
        echo -e "\033[0;31m❌ Could not determine latest release tag.\033[0m"
        exit 1
    fi

    # Find the right asset (tar.gz)
    ASSET_URL=$(echo "$LATEST_RELEASE_JSON" | grep "browser_download_url" | grep "${OS}_${ARCH}" | grep "tar.gz" | head -n 1 | sed -E 's/.*"([^"]+)".*/\1/')

    if [ -z "$ASSET_URL" ]; then
        echo -e "\033[0;31m❌ Could not find a valid asset for ${OS}_${ARCH} in release $TAG.\033[0m"
        exit 1
    fi

    TEMP_TAR="/tmp/a11ysentry.tar.gz"
    curl -L "$ASSET_URL" -o "$TEMP_TAR"
    
    # Extract to a temp dir and move binary
    TEMP_DIR=$(mktemp -d)
    tar -xzf "$TEMP_TAR" -C "$TEMP_DIR"
    
    # Find binary (it might be in a subfolder depending on GoReleaser config)
    EXTRACTED=$(find "$TEMP_DIR" -name "$BINARY_NAME" -type f | head -n 1)
    
    if [ -n "$EXTRACTED" ]; then
        mv "$EXTRACTED" "$BINARY_FULL"
        chmod +x "$BINARY_FULL"
        echo -e "\033[0;32m✅ Successfully installed A11ySentry $TAG.\033[0m"
    else
        echo -e "\033[0;31m❌ Binary not found in downloaded archive.\033[0m"
        exit 1
    fi
    
    rm "$TEMP_TAR"
    rm -rf "$TEMP_DIR"
fi

# 4. Add to PATH (Update .bashrc or .zshrc)
SHELL_CONFIG=""
if [ -f "$HOME/.zshrc" ]; then
    SHELL_CONFIG="$HOME/.zshrc"
elif [ -f "$HOME/.bashrc" ]; then
    SHELL_CONFIG="$HOME/.bashrc"
fi

if [ -n "$SHELL_CONFIG" ]; then
    if ! grep -q "$INSTALL_DIR" "$SHELL_CONFIG"; then
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_CONFIG"
        echo -e "\033[0;32m✅ Added to PATH in $SHELL_CONFIG\033[0m"
    fi
fi

# 5. MCP Registration & Skill Setup
echo -e "\033[0;36m⚙️  Detecting AI Agents and registering MCP...\033[0m"
BINARY_FULL="$INSTALL_DIR/$BINARY_NAME"
if [ -f "$BINARY_FULL" ]; then
    "$BINARY_FULL" mcp --register

    # 6. Skill Registration (Agent Teams Lite)
    SKILL_SOURCE="$HOME/repositories/semantix/skills/a11ysentry-mcp"
    AGENT_SKILLS_DIR="$HOME/.gemini/skills/a11ysentry-mcp"
    if [ -d "$SKILL_SOURCE" ]; then
        mkdir -p "$AGENT_SKILLS_DIR"
        cp -r "$SKILL_SOURCE/"* "$AGENT_SKILLS_DIR/"
        echo -e "\033[0;32m🧠 Registered a11ysentry-mcp skill in Agent Teams.\033[0m"
    fi
else
    echo -e "\033[0;33m⚠️  Automatic MCP registration skipped (binary not found).\033[0m"
fi

echo -e "\033[0;32m✅ A11ySentry installed successfully!\033[0m"
echo "🚀 Restart your terminal and try running 'a11ysentry --tui'."
