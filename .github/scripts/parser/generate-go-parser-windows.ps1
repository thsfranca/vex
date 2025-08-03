#!/usr/bin/env pwsh

Write-Host "[BUILD] Generating Go parser on Windows..."
cd tools/grammar
antlr4.bat -Dlanguage=Go -listener -visitor Vex.g4 -o ../gen/go/
Write-Host "[SUCCESS] Go parser generation completed"