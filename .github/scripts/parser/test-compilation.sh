#!/bin/bash
set -e

echo "[TEST] Testing Go parser compilation..."
cd tools/gen/go
go mod init vex-parser-test || true
go mod tidy || true
if go build .; then
    echo "[SUCCESS] Go parser compiles successfully"
else
    echo "[ERROR] Go parser compilation failed"
    exit 1
fi