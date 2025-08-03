#!/bin/bash

# Quick Fugo Extension Reinstall for Development
# This script uninstalls, packages, reinstalls the extension, and restarts Cursor

set -e  # Exit on any error

echo "🚀 Quick reinstalling Fugo VSCode Extension..."

# Check if Cursor is running
CURSOR_RUNNING=false
if pgrep -f "Cursor" > /dev/null; then
    CURSOR_RUNNING=true
    echo "🔄 Cursor is running - will restart after installation..."
fi

# Check if vsce is installed
if ! command -v vsce &> /dev/null; then
    echo "❌ vsce not found. Installing..."
    npm install -g @vscode/vsce
fi

# Extension info
EXTENSION_NAME="vex-minimal"
PUBLISHER="vex-dev"
FULL_NAME="$PUBLISHER.$EXTENSION_NAME"

# If Cursor is running, create a background restart script
if [ "$CURSOR_RUNNING" = true ]; then
    echo "🛑 Creating restart script and closing Cursor..."
    
    # Create a temporary restart script that runs independently
    cat > /tmp/fugo-restart-cursor.sh << 'EOF'
#!/bin/bash
sleep 3  # Wait for Cursor to fully close
echo "🚀 Restarting Cursor..."
open -a "Cursor" 2>/dev/null || cursor 2>/dev/null || echo "Could not auto-start Cursor"
rm -f /tmp/fugo-restart-cursor.sh  # Clean up
EOF
    
    chmod +x /tmp/fugo-restart-cursor.sh
    
    # Run restart script in background, detached from current process
    nohup /tmp/fugo-restart-cursor.sh > /dev/null 2>&1 &
    
    # Close Cursor
    osascript -e 'quit app "Cursor"' 2>/dev/null || pkill -f "Cursor" 2>/dev/null || true
    sleep 2  # Give it time to close properly
fi

echo "🗑️  Uninstalling existing extension..."
cursor --uninstall-extension "$FULL_NAME" 2>/dev/null || echo "   (No existing extension found)"

echo "📦 Updating icon and packaging extension..."
if command -v inkscape &> /dev/null; then
    inkscape icon.svg --export-type=png --export-filename=icon.png --export-width=128 --export-height=128
else
    magick icon.svg -density 300 -channel RGBA -alpha on -background none -resize 128x128 PNG32:icon.png
fi
vsce package --out "./fugo-latest.vsix"

echo "⚡ Installing new extension..."
cursor --install-extension "./fugo-latest.vsix"

echo ""
echo "✅ Extension reinstalled successfully!"
if [ "$CURSOR_RUNNING" = true ]; then
    echo "🎯 Cursor will restart automatically in a few seconds..."
else
    echo "🎯 Start Cursor and test your changes!"
fi
echo ""
echo "📝 Test checklist:"
echo "   • Open examples/test-extension.fugo"
echo "   • Verify syntax highlighting works"
echo "   • Check Color Theme → 'Fugo Dark'"
echo "   • Check File Icon Theme → 'Fugo File Icons'"
echo ""