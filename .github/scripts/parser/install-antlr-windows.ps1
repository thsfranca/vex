#!/usr/bin/env pwsh

# Install ANTLR for Windows
$url = "https://www.antlr.org/download/antlr-$env:ANTLR_VERSION-complete.jar"
$output = "$env:USERPROFILE\antlr.jar"
Invoke-WebRequest -Uri $url -OutFile $output

# Create batch file for antlr4 command
$batchContent = "@echo off`njava -jar `"$output`" %*"
$batchFile = "$env:USERPROFILE\antlr4.bat"
$batchContent | Out-File -FilePath $batchFile -Encoding ASCII

# Add to PATH
echo "$env:USERPROFILE" >> $env:GITHUB_PATH