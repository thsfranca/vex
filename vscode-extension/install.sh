#!/bin/bash

# Install Fugo VS Code Extension

echo "Installing Fugo VS Code Extension..."

# Determine VS Code extensions directory
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    VSCODE_EXT_DIR="$HOME/.vscode/extensions"
elif [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
    # Windows
    VSCODE_EXT_DIR="$USERPROFILE/.vscode/extensions"
else
    # Linux
    VSCODE_EXT_DIR="$HOME/.vscode/extensions"
fi

TARGET_DIR="$VSCODE_EXT_DIR/fugo-language-support"

# Create target directory
mkdir -p "$TARGET_DIR"

# Copy extension files
cp -r ./* "$TARGET_DIR/"

echo "Extension installed to: $TARGET_DIR"
echo ""
echo "Next steps:"
echo "1. Restart VS Code"
echo "2. Open a .fugo file (try examples/test-extension.fugo)" 
echo "3. Press Ctrl+Shift+T (or Cmd+Shift+T) to transpile"
echo ""
echo "Make sure you have Go installed and the Fugo transpiler built!"