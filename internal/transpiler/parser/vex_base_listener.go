// Code generated from Vex.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Vex

import "github.com/antlr4-go/antlr/v4"

// BaseVexListener is a complete listener for a parse tree produced by VexParser.
type BaseVexListener struct{}

var _ VexListener = &BaseVexListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseVexListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseVexListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseVexListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseVexListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterProgram is called when production program is entered.
func (s *BaseVexListener) EnterProgram(ctx *ProgramContext) {}

// ExitProgram is called when production program is exited.
func (s *BaseVexListener) ExitProgram(ctx *ProgramContext) {}

// EnterList is called when production list is entered.
func (s *BaseVexListener) EnterList(ctx *ListContext) {}

// ExitList is called when production list is exited.
func (s *BaseVexListener) ExitList(ctx *ListContext) {}

// EnterArray is called when production array is entered.
func (s *BaseVexListener) EnterArray(ctx *ArrayContext) {}

// ExitArray is called when production array is exited.
func (s *BaseVexListener) ExitArray(ctx *ArrayContext) {}
