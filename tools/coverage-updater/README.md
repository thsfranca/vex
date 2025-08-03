# Coverage Updater

A Go tool for generating coverage reports and updating the README with coverage tables.

## Purpose

This tool automates:
1. Generating Go test coverage for different project components
2. Calculating coverage percentages using `go tool cover`
3. Updating README.md with a formatted coverage table
4. Applying different coverage targets for different components

## Commands

### `generate-coverage`
Runs Go tests with coverage and calculates coverage percentages for all components.

```bash
./coverage-updater generate-coverage
```

**What it does:**
- Creates `coverage/` directory
- Runs `go test -coverprofile` for total and component coverage
- Calculates coverage percentages using `go tool cover`
- Outputs status environment variables for each component

### `update-readme`
Updates README.md with coverage table using environment variables.

```bash
./coverage-updater update-readme
```

**Environment Variables Used:**
- `PARSER_STATUS`, `TRANSPILER_STATUS`, `TYPES_STATUS`, `STDLIB_STATUS`, `TOTAL_STATUS`

## Coverage Targets

Different components have different coverage requirements:
- **Parser**: 95%+ (Critical language component)
- **Transpiler**: 90%+ (Core functionality)  
- **Type System**: 85%+ (Type safety)
- **Standard Library**: 80%+ (User-facing features)
- **Overall Project**: 75%+ (Quality baseline)

## Status Icons

- ✅ Coverage meets or exceeds target
- ❌ Coverage below target
- ⏳ Component not implemented yet

## CI Integration

Used in coverage workflow to automatically update README with latest coverage information after successful test runs.