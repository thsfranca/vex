#!/bin/bash
set -e

cd tools/grammar
antlr4 -Dlanguage=Go Vex.g4 -o /tmp/grammar-test
