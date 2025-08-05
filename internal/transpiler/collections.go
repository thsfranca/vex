package transpiler

import (
	"fmt"
)

// handleCollectionOp processes collection operations
func (t *Transpiler) handleCollectionOp(op string, args []string) {
	switch op {
	case "first":
		if len(args) < 1 {
			t.output.WriteString("_ = nil // Error: first requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = func() interface{} { if len(%s) > 0 { return %s[0] } else { return nil } }()\n", args[0], args[0]))
	
	case "rest":
		if len(args) < 1 {
			t.output.WriteString("_ = []interface{}{} // Error: rest requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = func() []interface{} { if len(%s) > 1 { return %s[1:] } else { return []interface{}{} } }()\n", args[0], args[0]))
	
	case "cons":
		if len(args) < 2 {
			t.output.WriteString("_ = []interface{}{} // Error: cons requires element and collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = append([]interface{}{%s}, %s...)\n", args[0], args[1]))
	
	case "count":
		if len(args) < 1 {
			t.output.WriteString("_ = 0 // Error: count requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = len(%s)\n", args[0]))
	
	case "empty?":
		if len(args) < 1 {
			t.output.WriteString("_ = true // Error: empty? requires collection\n")
			return
		}
		t.output.WriteString(fmt.Sprintf("_ = len(%s) == 0\n", args[0]))
	}
}

// evaluateCollectionOp evaluates collection operations as expressions
func (t *Transpiler) evaluateCollectionOp(op string, args []string) string {
	switch op {
	case "first":
		if len(args) < 1 {
			return "nil"
		}
		return fmt.Sprintf("func() interface{} { if len(%s) > 0 { return %s[0] } else { return nil } }()", args[0], args[0])
	
	case "rest":
		if len(args) < 1 {
			return "[]interface{}{}"
		}
		return fmt.Sprintf("func() []interface{} { if len(%s) > 1 { return %s[1:] } else { return []interface{}{} } }()", args[0], args[0])
	
	case "cons":
		if len(args) < 2 {
			return "[]interface{}{}"
		}
		return fmt.Sprintf("append([]interface{}{%s}, %s...)", args[0], args[1])
	
	case "count":
		if len(args) < 1 {
			return "0"
		}
		return fmt.Sprintf("len(%s)", args[0])
	
	case "empty?":
		if len(args) < 1 {
			return "true"
		}
		return fmt.Sprintf("len(%s) == 0", args[0])
	
	default:
		return "nil"
	}
}