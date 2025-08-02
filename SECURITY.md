# Security Policy

## 🔒 Security Considerations

Fugo is a **learning project** and **experimental programming language** that is **not intended for production use**. However, we take security considerations seriously as part of the educational experience.

## 🐛 Security Issues

This is a personal learning project with limited external engagement.

### Critical Issues Only
Only report security issues that:
- **Crash the parser** with minimal input
- **Cause infinite loops** in grammar processing  
- **Generate unsafe Go code** from valid Fugo input

Create a GitHub issue with clear reproduction steps.

## ⚠️ Important Disclaimers

### Not Production Ready
- Fugo is **experimental software** for learning purposes
- **Do not use in production** environments
- **No security guarantees** are provided
- **Use at your own risk** for educational exploration

### Generated Code Security
- Transpiled Go code is **not audited** for security
- Generated code may contain **vulnerabilities**
- **Review all generated code** before any use
- **Do not run untrusted** Fugo programs

## 🛡️ Security Best Practices

If you're experimenting with Fugo:

### Safe Exploration
- **Run in isolated environments** (containers, VMs)
- **Don't process untrusted input** without sandboxing
- **Review generated Go code** before execution
- **Use version control** to track changes

### Parser Safety
- **Limit input size** when testing large files
- **Use timeouts** for parsing operations
- **Be cautious with** nested structures (potential stack overflow)

## 📚 Educational Security Learning

This project can help learn about:
- **Parser security** - How parsing can be attacked
- **Code generation vulnerabilities** - How transpilers can introduce bugs
- **Supply chain security** - Dependencies and build systems
- **Language design security** - How language features affect security

---

**Note**: This is a personal learning project. External engagement is minimal.