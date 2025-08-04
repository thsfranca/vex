#!/bin/bash
set -e

echo "[BUILD] Building Go packages..."

if find . -name "main.go" | grep -q .; then
    go build ./... || {
        echo "[ERROR] Build failed"
        exit 1
    }
else
    echo "[INFO] No main packages to build yet - this is expected"
fi
