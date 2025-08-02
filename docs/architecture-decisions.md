# Architecture Decision Record (ADR) - Fugo Language

## ADR-001: Language Implementation Strategy - Transpilation to Go

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to choose implementation strategy for Fugo language execution

### Decision
Transpile Fugo code to Go instead of building an interpreter or VM.

### Context
- Goal: High-performance backend services
- Need: Go ecosystem interoperability
- Constraint: Learning project with limited development time

### Alternatives Considered
1. **Tree-walking interpreter**: Easier to implement but slower execution
2. **Bytecode VM**: Good performance but complex implementation
3. **Transpilation to Go**: Best performance + ecosystem access
4. **Transpilation to other languages**: Less ecosystem benefit

### Decision Rationale
- **Performance**: Native Go performance for production workloads
- **Ecosystem**: Direct access to Go libraries and frameworks
- **Deployment**: Standard Go deployment patterns
- **Learning**: Explores code generation and language interop

### Consequences
- ✅ Maximum runtime performance
- ✅ Go ecosystem compatibility
- ✅ Standard Go tooling works
- ❌ More complex implementation than interpreter
- ❌ Debugging requires source maps

---

## ADR-002: Parser Technology - ANTLR4

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to choose parsing technology for Fugo grammar

### Decision
Use ANTLR4 for lexing and parsing with multi-language target support.

### Context
- Need: Robust parsing for Lisp-inspired syntax
- Goal: Generate parsers for multiple target languages
- Constraint: Learning-focused implementation

### Alternatives Considered
1. **Hand-written recursive descent**: Full control but time-intensive
2. **Yacc/Bison**: Traditional but limited language support
3. **PEG parsers**: Good but less ecosystem support
4. **ANTLR4**: Industry standard with multi-language support

### Decision Rationale
- **Multi-language support**: Can generate Go, Java, Python, etc.
- **Grammar clarity**: Declarative grammar definition
- **Tooling**: Excellent IDE support and debugging
- **Learning**: Industry-standard parser generator

### Consequences
- ✅ Clean grammar definition in `Fugo.g4`
- ✅ Multiple language targets for experimentation
- ✅ Good tooling and documentation
- ❌ Additional dependency on ANTLR runtime
- ❌ Generated code can be harder to debug

---

## ADR-003: Syntax Design - Lisp-Inspired S-Expressions

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to choose syntax style for Fugo language

### Decision
Use Lisp-inspired S-expression syntax with modern type annotations.

### Context
- Goal: Functional programming paradigm
- Need: Simple, consistent syntax for rapid development
- Goal: Easy parsing and AST manipulation

### Alternatives Considered
1. **C-style syntax**: Familiar but complex parsing
2. **ML/Haskell syntax**: Good for FP but steep learning curve
3. **JavaScript-like syntax**: Familiar but not ideal for FP
4. **Lisp S-expressions**: Simple, consistent, FP-friendly

### Decision Rationale
- **Simplicity**: Uniform syntax reduces parser complexity
- **Functional fit**: Natural for functional programming
- **Metaprogramming**: Easy macro system potential
- **Learning**: Explores Lisp language family concepts

### Consequences
- ✅ Simple, consistent grammar rules
- ✅ Easy AST manipulation
- ✅ Natural functional programming syntax
- ❌ Less familiar to mainstream developers
- ❌ Requires learning curve for syntax

---

## ADR-004: Type System Approach - Static Typing with Inference

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to choose type system approach for performance and safety

### Decision
Implement static typing with Hindley-Milner style type inference.

### Context
- Goal: High performance through compile-time optimization
- Need: Developer productivity with minimal type annotations
- Goal: Go interop requires known types

### Alternatives Considered
1. **Dynamic typing**: Simple but slower, harder Go interop
2. **Static typing without inference**: Safe but verbose
3. **Gradual typing**: Complex implementation
4. **Static typing with inference**: Best balance

### Decision Rationale
- **Performance**: Enables Go transpilation optimizations
- **Safety**: Compile-time error detection
- **Productivity**: Minimal type annotations required
- **Go interop**: Direct type mapping to Go types

### Consequences
- ✅ Compile-time error detection
- ✅ Optimized Go code generation
- ✅ Minimal type annotations needed
- ❌ Complex type inference implementation
- ❌ Type error messages can be complex

---

## ADR-005: Development Phase Strategy - Skip Interpreter Phase

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Originally planned tree-walking interpreter before transpiler

### Decision
Skip the tree-walking interpreter phase and go directly to transpilation.

### Context
- Original plan: Phase 2 was tree-walking interpreter
- Realization: Interpreter doesn't align with Go interop goals
- Goal: Faster progress toward main objectives

### Alternatives Considered
1. **Build interpreter first**: Easier but misaligned with goals
2. **Build interpreter alongside transpiler**: Too much complexity
3. **Skip interpreter, focus on transpiler**: Aligned with end goals

### Decision Rationale
- **Goal alignment**: Transpilation is the end goal anyway
- **Learning focus**: Code generation more relevant than interpretation
- **Time efficiency**: Avoid building throwaway components
- **Complexity**: Simpler to have one execution path

### Consequences
- ✅ Faster progress toward main goals
- ✅ More focused implementation effort
- ✅ Better alignment with Go interop objectives
- ❌ No fallback execution method during development
- ❌ Harder to test language features incrementally

---

## ADR-006: Target Domain - Backend Services

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to define primary use case and optimize accordingly

### Decision
Focus on backend service development with HTTP, concurrency, and data processing.

### Context
- Goal: Practical language with real-world applications
- Opportunity: Go's strength in backend services
- Constraint: Limited development time requires focus

### Alternatives Considered
1. **General-purpose language**: Broader appeal but unfocused
2. **Systems programming**: Complex, less Go ecosystem benefit
3. **Frontend/web development**: Not Go's strength
4. **Backend services**: Leverages Go's strengths

### Decision Rationale
- **Go ecosystem**: Maximum benefit from existing libraries
- **Performance focus**: Backend services need high performance
- **Learning relevance**: Modern software development focus
- **Scope management**: Clear boundaries for feature development

### Consequences
- ✅ Clear feature prioritization
- ✅ Leverages Go's ecosystem strengths
- ✅ Relevant to modern development needs
- ❌ Limited appeal outside backend development
- ❌ May miss some general-purpose language features

---

## ADR-007: Development Workflow - Protected Main Branch

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to ensure code quality and prevent direct commits to main branch

### Decision
Implement GitHub branch protection rules requiring pull requests for all changes to main branch.

### Context
- Goal: Maintain code quality through review process
- Need: Prevent accidental direct commits to main
- Practice: Industry standard for collaborative development
- Learning: Enforce proper Git workflow habits

### Alternatives Considered
1. **No protection**: Simple but risky for code quality
2. **Honor system**: Relies on developer discipline
3. **Branch protection with PR requirements**: Enforced quality gates
4. **Additional restrictions**: Signed commits, linear history

### Decision Rationale
- **Quality assurance**: All code changes reviewed before merge
- **Feature isolation**: Forces proper feature branch workflow
- **CI integration**: Ensures automated checks pass
- **Learning value**: Teaches professional development practices
- **Mistake prevention**: Prevents accidental main branch commits

### Consequences
- ✅ Improved code quality through reviews
- ✅ Enforced feature branch workflow
- ✅ Integration with CI/CD pipeline
- ✅ Professional development practices
- ❌ Slightly more complex workflow for solo development
- ❌ Cannot make quick fixes directly to main

### Implementation
- GitHub branch protection rule on `main` branch
- Require pull request approval before merging
- Require status checks to pass (CI, tests)
- Include administrators in restrictions
- Linear history requirement (no merge commits)

---

## ADR-008: Open Source Setup - MIT License

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to choose open source license and setup for public repository

### Decision
Release Fugo under MIT License as an open source learning project.

### Context
- Goal: Share learning journey publicly for educational benefit
- Need: Simple, permissive license for educational projects
- Approach: Open source for visibility, not active collaboration
- Focus: Personal learning project, not seeking contributors

### Alternatives Considered
1. **Private repository**: Limits learning sharing and visibility
2. **GPL License**: Too restrictive for educational/experimental project
3. **Apache 2.0**: More complex than needed for simple learning project
4. **MIT License**: Simple, permissive, perfect for educational use
5. **BSD License**: Similar to MIT but slightly more complex

### Decision Rationale
- **Simplicity**: MIT is straightforward and well-understood
- **Educational focus**: Permissive license encourages learning and forking
- **No restrictions**: Others can freely use code for their own learning
- **Industry standard**: MIT is widely used for open source projects
- **Future flexibility**: Easy to change or dual-license later if needed

### Consequences
- ✅ Public visibility for learning and inspiration
- ✅ Simple license terms everyone understands
- ✅ Encourages forking and experimentation
- ✅ No licensing complexity for learning project
- ❌ Could theoretically be used commercially (acceptable trade-off)
- ❌ No copyleft protection (not needed for learning project)

### Implementation
- MIT License in `LICENSE` file
- Learning-focused documentation in `CONTRIBUTING.md`
- Clear educational purpose statements in README
- Repository configured as public on GitHub
- Minimal issue templates (bug reports only)
- Limited external engagement approach

---

## ADR-009: Minimal External Interaction

**Date**: Project Setup  
**Status**: ✅ Accepted  
**Context**: Need to balance open source visibility with personal learning focus

### Decision
Maintain minimal external interaction while keeping the project publicly available for learning.

### Context
- Goal: Share learning journey without active external engagement
- Need: Focus on personal learning without community management overhead
- Approach: Open for inspiration, minimal for interaction
- Timeline: May increase interaction in future phases

### Alternatives Considered
1. **Private repository**: Limits learning sharing completely
2. **Active community engagement**: Too much overhead for learning focus
3. **Standard open source approach**: More interaction than desired currently
4. **Minimal interaction approach**: Balanced visibility with focus

### Decision Rationale
- **Learning focus**: Maintains primary educational objective
- **Reduced overhead**: Minimizes time spent on community management
- **Future flexibility**: Can increase engagement later if desired
- **Inspiration sharing**: Still provides value to other learners
- **Quality maintenance**: Allows focus on implementation over discussion

### Consequences
- ✅ More time for implementation and learning
- ✅ Reduced community management overhead
- ✅ Clear expectations for external users
- ✅ Can evolve approach as project matures
- ❌ Less community feedback and ideas
- ❌ Reduced networking opportunities

### Implementation
- Simplified issue templates (critical bugs only)
- Clear statements about limited external support
- Focus on "fork and learn" rather than "contribute"
- Minimal response expectations set in documentation

---

## Current Architecture Overview

### System Components
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Fugo Source   │───▶│   ANTLR4 Parser │───▶│   AST/Types     │
│   (.fugo files) │    │   (Generated)   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  Go Executable  │◀───│  Go Compiler    │◀───│  Go Transpiler  │
│                 │    │   (go build)    │    │  (To be built)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Current Status
- **Phase 1**: ✅ Parser Foundation (Complete)
- **Phase 2**: 🚧 Go Transpilation Engine (In Progress)
- **Future**: Type System, Standard Library, IDE Support

---

## Decision Log

| ADR | Decision | Status |
|-----|----------|--------|
| 001 | Transpilation to Go | ✅ Accepted |
| 002 | ANTLR4 Parser | ✅ Accepted |
| 003 | Lisp S-Expression Syntax | ✅ Accepted |
| 004 | Static Typing with Inference | ✅ Accepted |
| 005 | Skip Interpreter Phase | ✅ Accepted |
| 006 | Backend Services Focus | ✅ Accepted |
| 007 | Protected Main Branch | ✅ Accepted |
| 008 | MIT License Open Source | ✅ Accepted |
| 009 | Minimal External Interaction | ✅ Accepted |

---

*This ADR is maintained as part of the Fugo language development process. For questions or suggestions, please refer to the project documentation or create an issue.*