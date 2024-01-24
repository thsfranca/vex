package lang

import (
	"study-parser/module"
	"study-parser/parser"

	"github.com/antlr4-go/antlr/v4"
)

// BaseSpListener is a complete listener for a parse tree produced by SpParser.
type BaseSpListener struct{}

func NewListener() *BaseSpListener {
	return &BaseSpListener{}
}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseSpListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseSpListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseSpListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseSpListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSp is called when production sp is entered.
func (s *BaseSpListener) EnterSp(ctx *parser.SpContext) {}

// ExitSp is called when production sp is exited.
func (s *BaseSpListener) ExitSp(ctx *parser.SpContext) {

}

// EnterList is called when production list is entered.
func (s *BaseSpListener) EnterList(ctx *parser.ListContext) {
	if IsSpecialForm(ctx) {
		CallSpecialForm(ctx, module.Mod.Main.SymbolTable())
		return
	}

	print(module.Mod.Main.SymbolTable().Lookup(ctx.Sym_block(0).GetText()).Value)
}

// ExitList is called when production list is exited.
func (s *BaseSpListener) ExitList(ctx *parser.ListContext) {}

// EnterArray is called when production array is entered.
func (s *BaseSpListener) EnterArray(ctx *parser.ArrayContext) {}

// ExitArray is called when production array is exited.
func (s *BaseSpListener) ExitArray(ctx *parser.ArrayContext) {}

// EnterSym_block is called when production sym_block is entered.
func (s *BaseSpListener) EnterSym_block(ctx *parser.Sym_blockContext) {}

// ExitSym_block is called when production sym_block is exited.
func (s *BaseSpListener) ExitSym_block(ctx *parser.Sym_blockContext) {}
