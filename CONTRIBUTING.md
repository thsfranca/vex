# Learning with Fugo

Fugo is a **personal learning project** focused on exploring language implementation concepts through building a functional programming language that transpiles to Go.

## 🎯 Project Philosophy

Fugo is created for educational purposes and fun:

- This is primarily a learning exercise, not a production language
- Decisions prioritize educational value and simplicity over practicality
- The pace is relaxed and driven by personal interest and available time
- The goal is to understand language implementation from first principles

## 🚀 How to Use This Project for Learning

### 🍴 Fork and Experiment
- **Fork the repository** to create your own version
- **Experiment freely** with different language designs
- **Try alternative implementations** of features
- **Use it as a starting point** for your own language project

### 📚 Study the Implementation
- **Follow the commit history** to see how decisions evolved
- **Read the documentation** to understand the architecture
- **Examine the ANTLR grammar** to learn parser design
- **Use it as a reference** for your own language projects

## 🔍 Current Development Focus

**Phase 2: Go Transpilation Engine**
- Basic expression transpilation
- Symbol table implementation  
- AST to Go code generation
- Function definition transpilation

See [PROGRESS.md](PROGRESS.md) for current status and [docs/fugo-implementation-requirements.md](docs/fugo-implementation-requirements.md) for the complete roadmap.

## 📚 Learning Resources

### Understanding the Codebase
- [Architecture Decision Record](docs/architecture-decisions.md) - Key design decisions and rationale
- [Grammar Reference](docs/grammar-reference.md) - Current language syntax
- [Implementation Requirements](docs/fugo-implementation-requirements.md) - Complete development roadmap

### External Learning Resources
- [ANTLR4 Documentation](https://www.antlr.org/doc/) - Parser generator used in Fugo
- ["Crafting Interpreters" by Robert Nystrom](https://craftinginterpreters.com/) - Excellent language implementation guide
- [Go Language Specification](https://golang.org/ref/spec) - Target language for transpilation

## 🎯 Learning Focus Areas

This project explores these language implementation concepts:

- **Parsing & Grammar Design** - ANTLR4 grammar for S-expressions
- **Type Systems** - Static typing with type inference (planned)
- **Code Generation** - Transpilation from Fugo to Go
- **Language Interoperability** - Go ecosystem integration
- **Functional Programming** - Immutable data structures and pure functions

## 🛠️ Setting Up for Exploration

If you want to experiment with the code:

1. **Clone or fork the repository**
2. **Install ANTLR4** (required for parser generation)
3. **Generate parsers**:
   ```bash
   make generate
   ```
4. **Run examples**:
   ```bash
   cd examples/go-usage && go run main.go
   ```

## 🐛 Critical Issues Only

For critical parser bugs that prevent basic functionality:
- Check documentation first  
- Create a bug report if the issue affects core parsing
- Include minimal reproduction steps

**Note**: This is a personal learning project with limited external support.

---

**This project is for learning about language implementation. Feel free to learn from it, fork it, and use it as inspiration for your own language projects!**