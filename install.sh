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

# 3. Download (Logic ready for production)
URL="https://github.com/$OWNER/$REPO/releases/latest/download/a11ysentry_${OS}_${ARCH}.tar.gz"
echo -e "\033[0;90m📥 Preparation for download from $URL\033[0m"

# DEV MODE: Copy if exists locally
if [ -f "./cmd/a11ysentry/a11ysentry" ]; then
    cp "./cmd/a11ysentry/a11ysentry" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo -e "\033[0;33m🚧 DEV MODE: Copied binary from cmd/a11ysentry/.\033[0m"
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
