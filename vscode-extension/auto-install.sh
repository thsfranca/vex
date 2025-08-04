#!/bin/bash

# Auto-reinstall Vex Extension on File Changes
# This script watches for changes and automatically reinstalls the extension

set -e

echo "👀 Starting Vex Extension Auto-Install Watcher..."
echo "[WATCH] Watching for changes in extension files..."
echo "[INFO] Press Ctrl+C to stop"
echo ""

# Check if fswatch is available (macOS)
if command -v fswatch &> /dev/null; then
    WATCHER="fswatch"
elif command -v inotifywait &> /dev/null; then
    WATCHER="inotifywait"
else
    echo "[ERROR] No file watcher available. Installing fswatch..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew install fswatch
            WATCHER="fswatch"
        else
            echo "[ERROR] Homebrew not found. Please install fswatch: brew install fswatch"
            exit 1
        fi
    else
        echo "[ERROR] Please install inotify-tools: sudo apt-get install inotify-tools"
        exit 1
    fi
fi

# Files and directories to watch
WATCH_PATHS=(
    "package.json"
    "syntaxes/"
    "themes/"
    "icons/"
    "icon.svg"
    "language-configuration.json"
)

# Function to reinstall extension
reinstall_extension() {
    echo ""
    echo "🔄 Change detected! Reinstalling extension..."
    echo "================================================"
    ./quick-install.sh
    echo "================================================"
    echo "👀 Watching for more changes..."
    echo ""
}

# Set up file watching based on available tool
if [ "$WATCHER" = "fswatch" ]; then
    # macOS fswatch
    fswatch -o "${WATCH_PATHS[@]}" | while read num; do
        reinstall_extension
    done
else
    # Linux inotifywait
    inotifywait -m -r -e modify,create,delete,move "${WATCH_PATHS[@]}" | while read path action file; do
        reinstall_extension
    done
fi