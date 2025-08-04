#!/bin/bash
set -e

if [ -f "tools/grammar/Vex.g4" ]; then
    make go || echo "Parser generation failed, continuing..."
fi
