#!/bin/bash
set -e

echo "[BUILD] Generating Go parser on $(uname)..."
cd tools/grammar
antlr4 -Dlanguage=Go -listener -visitor Vex.g4 -o ../gen/go/
echo "[SUCCESS] Go parser generation completed"