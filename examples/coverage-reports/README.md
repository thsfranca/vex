# Coverage Report Examples

This directory contains examples of what the automated coverage reports will look like when posted as comments on Pull Requests.

## Files:

- `successful-coverage-example.md` - Example when all thresholds are met
- `failed-coverage-example.md` - Example when some thresholds fail

## How It Works:

1. **Automatic Generation**: The CI generates these reports after running tests
2. **PR Comments**: Reports are posted/updated as comments on PRs
3. **Build Status**: Builds fail if any implemented component is below threshold
4. **Smart Updates**: Existing comments are updated rather than creating new ones

## Report Sections:

- **Status Header**: Overall pass/fail with total coverage
- **Component Table**: Current vs threshold for each component
- **Thresholds Reference**: What each threshold means
- **Fix Guidance**: Actionable steps when coverage fails (only shown on failures)
