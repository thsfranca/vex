#!/bin/bash
set -e

cd tools/grammar
antlr4 -Dlanguage=Go -listener -visitor Vex.g4 -o /tmp/vex-parser
