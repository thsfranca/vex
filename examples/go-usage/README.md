# Go Parser Usage Example

This example demonstrates how to use the generated Go parser for the Fugo language.

## Setup

1. Generate the Go parser (from the project root):
```bash
make go
```

2. Install dependencies:
```bash
cd examples/go-usage
go mod tidy
```

3. Update the import path in `main.go` to point to your generated parser:
```go
import "path/to/your/tools/gen/go"
```

## Running

```bash
go run main.go
```

## What it does

The example:
1. Creates a sample Fugo program as a string
2. Parses it using the generated ANTLR4 Go parser
3. Walks the parse tree using a custom listener
4. Prints information about the parsed structures

## Generated Files

The Go parser generation creates these files:
- `fugo_lexer.go` - Tokenizes input text
- `fugo_parser.go` - Main parser logic
- `fugo_listener.go` - Listener interface for tree walking
- `fugo_base_listener.go` - Base listener implementation

## Usage Patterns

### Basic Parsing
```go
// Create input stream
input := antlr.NewInputStream(code)

// Create lexer and parser
lexer := parser.NewFugoLexer(input)
tokenStream := antlr.NewCommonTokenStream(lexer, 0)
p := parser.NewFugoParser(tokenStream)

// Parse
tree := p.Sp()
```

### Tree Walking
```go
// Custom listener
type MyListener struct {
    *parser.BaseFugoListener
}

func (l *MyListener) EnterList(ctx *parser.ListContext) {
    // Handle list expressions
}

// Walk the tree
listener := &MyListener{}
antlr.ParseTreeWalkerDefault.Walk(listener, tree)
```