#!/bin/bash
set -e

# Arguments: PARSER_GENERATION_RESULT
PARSER_GENERATION_RESULT="$1"

echo "[SUMMARY] Go parser generation summary:"

if [ "$PARSER_GENERATION_RESULT" = "success" ]; then
    echo "[SUCCESS] Go parser generated successfully on all platforms"
    echo "[INFO] Vex grammar is ready for Go transpilation!"
else
    echo "[ERROR] Go parser generation failed on some platforms"
    echo "[INFO] Check the job logs and artifacts for debugging"
    exit 1
fi