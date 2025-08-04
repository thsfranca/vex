# Extension Tester

A Go tool for testing, packaging, and verifying VSCode extensions.

## Purpose

This tool automates VSCode extension CI tasks:
1. Creating test Vex files for syntax highlighting validation
2. Packaging extensions into .vsix files using @vscode/vsce
3. Verifying package integrity and structure
4. Generating validation summaries for CI reports

## Commands

### `create-samples`
Creates test Vex files for syntax highlighting validation.

```bash
./extension-tester create-samples
```

**What it creates:**
- `test-samples/factorial.vx` - Factorial function example
- `test-samples/fibonacci.vx` - Fibonacci sequence example

### `package`
Packages the VSCode extension into a .vsix file.

```bash
./extension-tester package
```

**Requirements:**
- `package.json` must exist in current directory
- `npx @vscode/vsce` must be available

**Output:** Creates `vex-test-build.vsix`

### `verify`
Verifies the packaged extension integrity.

```bash
./extension-tester verify
```

**What it checks:**
- .vsix file exists
- Archive integrity (using `unzip -t`)
- Package contents listing
- File size reporting

### `summary`
Generates a validation summary for CI reports.

```bash
./extension-tester summary
```

**Environment Variables Used:**
- `EXTENSION_FILES` - Whether extension files changed ("true"/"false")
- `VALIDATE_RESULT` - Validation result ("success"/"failure")
- `SKIP_RESULT` - Skip validation result

## Sample Files

The tool creates realistic Vex language examples to test:
- Syntax highlighting for S-expressions
- Function definitions and calls
- Conditional expressions
- Mathematical operations
- Comments

## CI Integration

Used in VSCode extension workflow to automate testing and packaging while keeping workflow files clean and maintainable.
