# Fugo Language Support for VS Code

A VS Code extension for the Fugo functional programming language.

## Features

- **Syntax Highlighting**: Highlights Fugo syntax including comments, strings, numbers, and keywords
- **Quick Transpilation**: Press `Ctrl+Shift+T` (or `Cmd+Shift+T` on Mac) to transpile current Fugo file to Go
- **Auto-closing Pairs**: Automatic bracket and quote completion

## Usage

1. Open a `.fugo` file in VS Code
2. Write your Fugo code:
   ```fugo
   ; Simple Fugo program
   (42)
   ("hello world")
   ```
3. Press `Ctrl+Shift+T` to transpile to Go
4. The transpiled Go code opens in a new tab beside your Fugo file

## Installation

1. Copy this extension folder to your VS Code extensions directory:
   - **Windows**: `%USERPROFILE%\.vscode\extensions\`
   - **macOS**: `~/.vscode/extensions/`
   - **Linux**: `~/.vscode/extensions/`

2. Restart VS Code

3. Open any `.fugo` file to activate the extension

## Development Setup

This extension requires the Fugo transpiler to be available in your workspace. Make sure you have:

- Go installed
- The Fugo project with working transpiler in `pkg/transpiler`

## Commands

- `Fugo: Transpile to Go` - Transpiles the current Fugo file to Go code

## File Associations

- `.fugo` files are automatically recognized as Fugo language files

## Contributing

This extension is part of the Fugo language learning project. Feel free to improve the syntax highlighting or add new features!