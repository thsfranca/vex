# Known Bugs and Issues

## Overview

This document tracks known bugs, limitations, and issues in the Vex language implementation that need to be addressed in future development cycles. Use this as a reference to remember what needs fixing and to avoid duplicating bug reports.

## 🐛 Active Bugs

### High Priority

#### Package Discovery System Missing
- **Status**: ⚠️ HIGH PRIORITY - IMMEDIATE IMPLEMENTATION NEEDED
- **Description**: The package discovery system outlined in the implementation requirements is not yet implemented
- **Impact**: Cannot handle multi-file projects, circular dependency detection, or proper module management
- **Workaround**: Use single-file programs only
- **Tracking**: Phase 3 - moved from Phase 7 as critical infrastructure
- **Next Steps**: Implement directory-based package structure and import path resolution

#### Type System Limitations
- **Status**: ⏳ PLANNED
- **Description**: Advanced type checking and inference not implemented
- **Impact**: Runtime errors for type mismatches, no compile-time type validation
- **Workaround**: Rely on Go's type system through interop
- **Tracking**: Phase 5 (was Phase 3)
- **Next Steps**: Implement Hindley-Milner style type inference

### Medium Priority

#### Function Scoping Edge Cases
- **Status**: 🔍 NEEDS INVESTIGATION
- **Description**: Function definitions may have scoping issues in complex nested scenarios
- **Impact**: Potential symbol resolution errors in generated Go code
- **Workaround**: Keep function definitions simple and well-ordered
- **Example**: Nested function calls within macro expansions
- **Next Steps**: Investigate symbol table management in complex cases

#### Macro Parameter Validation
- **Status**: ✅ MOSTLY FIXED, SOME EDGE CASES
- **Description**: Some macro parameter validation edge cases may not be properly handled
- **Impact**: Potential macro expansion errors
- **Workaround**: Use simple macro parameter patterns
- **Next Steps**: Add comprehensive macro parameter validation tests

### Low Priority

#### Generated Go Code Formatting
- **Status**: 🎨 ENHANCEMENT
- **Description**: Generated Go code could be more readable and idiomatic
- **Impact**: Debugging transpiled code is more difficult
- **Workaround**: Use `go fmt` on generated code
- **Next Steps**: Improve code generation templates

#### Error Message Clarity
- **Status**: 🎨 ENHANCEMENT
- **Description**: Some error messages could be more descriptive
- **Impact**: Harder debugging experience
- **Workaround**: Use transpile command to inspect generated Go
- **Next Steps**: Enhance error reporting with better context

## 🚧 Known Limitations (By Design)

### Current Architecture Limitations

#### No Loops
- **Status**: 📋 PLANNED FEATURE
- **Description**: No native loop constructs (for, while, etc.)
- **Workaround**: Use recursion or Go interop
- **Timeline**: Phase 6 - Control Flow

#### No Error Handling
- **Status**: 📋 PLANNED FEATURE  
- **Description**: No structured error handling (try/catch, Result types)
- **Workaround**: Use Go's error handling patterns via interop
- **Timeline**: Phase 6 - Error Handling

#### No Immutable Data Structures
- **Status**: 📋 PLANNED FEATURE
- **Description**: No persistent vectors, maps, or other immutable collections
- **Workaround**: Use Go types via interop
- **Timeline**: Phase 5 - Immutable Data Structures

#### No HTTP Framework
- **Status**: 📋 PLANNED FEATURE
- **Description**: No built-in HTTP server or web framework
- **Workaround**: Use Go's net/http via interop
- **Timeline**: Phase 6 - Concurrency and Backend Features

## 🔧 Workarounds and Best Practices

### General Development
1. **Keep programs simple**: Use single files until package system is implemented
2. **Test incrementally**: Use `vex run` for quick testing of small parts
3. **Check generated Go**: Use `vex transpile` to inspect output when debugging
4. **Use Go tooling**: Leverage Go's debugging and profiling tools on generated code

### Function Definitions
1. **Define functions before use**: Ensure proper ordering in source code
2. **Use simple parameter lists**: Avoid complex nested parameter structures
3. **Test functions individually**: Verify each function works before combining

### Macro Usage
1. **Keep macros simple**: Avoid deeply nested macro expansions
2. **Validate parameters**: Check macro arguments match expected patterns
3. **Use defn for functions**: Prefer defn macro over custom function macros

### Go Interoperability
1. **Import early**: Add import statements at the top of files
2. **Check Go documentation**: Verify function signatures and types
3. **Use namespace syntax**: Always use `package/function` format

## 📊 Bug Reporting Template

When adding new bugs to this document, use this template:

```markdown
#### Bug Title
- **Status**: [🐛 ACTIVE | 🔍 INVESTIGATING | ✅ FIXED | 📋 PLANNED]
- **Description**: Brief description of the issue
- **Impact**: How this affects users/development
- **Workaround**: Temporary solution if available
- **Example**: Code that demonstrates the issue (if applicable)
- **Next Steps**: What needs to be done to fix it
```

## 🏷️ Status Labels

- 🐛 **ACTIVE**: Confirmed bug that needs fixing
- 🔍 **INVESTIGATING**: Issue under investigation
- ✅ **FIXED**: Bug has been resolved
- 📋 **PLANNED**: Known limitation that will be addressed in planned features
- ⚠️ **HIGH PRIORITY**: Critical issue that should be addressed immediately
- 🎨 **ENHANCEMENT**: Improvement rather than bug fix
- 🚧 **LIMITATION**: Current architectural limitation

## 📝 Notes for Developers

### Before Reporting New Bugs
1. Check this document to avoid duplicates
2. Test with latest transpiler build
3. Try minimal reproduction case
4. Check if it's a known limitation vs actual bug

### When Fixing Bugs
1. Update the status in this document
2. Add test cases to prevent regression
3. Update relevant documentation
4. Consider if fix affects other components

### Prioritization Guidelines
- **High Priority**: Blocks basic functionality or causes crashes
- **Medium Priority**: Affects specific use cases or edge cases
- **Low Priority**: Quality of life improvements or cosmetic issues

---

**Last Updated**: 2025-01-09  
**Next Review**: When new features are implemented or major bugs discovered  
**Maintainer**: Development team