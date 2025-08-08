# Coverage Report Examples

This directory contains examples of what the automated coverage reports will look like when posted as comments on Pull Requests.

## Files:

- `successful-coverage-example.md` - Example when threshold is met
- `failed-coverage-example.md` - Example when threshold is not met

## How It Works:

1. **Automatic Generation**: The CI generates these reports after running tests
2. **PR Comments**: Reports are posted/updated as comments on PRs
3. **Build Status**: Builds fail if overall transpiler coverage is below threshold
4. **Smart Updates**: Existing comments are updated rather than creating new ones

## Report Sections:

- **Status Header**: Overall pass/fail with total coverage
- **Coverage Summary**: Current vs target threshold (simple format)
- **Fix Guidance**: Actionable steps when coverage fails (only shown on failures)
