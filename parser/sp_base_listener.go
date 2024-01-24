// Code generated from Sp.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Sp

import "github.com/antlr4-go/antlr/v4"

// BaseSpListener is a complete listener for a parse tree produced by SpParser.
type BaseSpListener struct{}

var _ SpListener = &BaseSpListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseSpListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseSpListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseSpListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseSpListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterSp is called when production sp is entered.
func (s *BaseSpListener) EnterSp(ctx *SpContext) {}

// ExitSp is called when production sp is exited.
func (s *BaseSpListener) ExitSp(ctx *SpContext) {}

// EnterList is called when production list is entered.
func (s *BaseSpListener) EnterList(ctx *ListContext) {}

// ExitList is called when production list is exited.
func (s *BaseSpListener) ExitList(ctx *ListContext) {}

// EnterArray is called when production array is entered.
func (s *BaseSpListener) EnterArray(ctx *ArrayContext) {}

// ExitArray is called when production array is exited.
func (s *BaseSpListener) ExitArray(ctx *ArrayContext) {}

// EnterSym_block is called when production sym_block is entered.
func (s *BaseSpListener) EnterSym_block(ctx *Sym_blockContext) {}

// ExitSym_block is called when production sym_block is exited.
func (s *BaseSpListener) ExitSym_block(ctx *Sym_blockContext) {}
