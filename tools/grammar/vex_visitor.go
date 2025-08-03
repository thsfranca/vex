// Code generated from Vex.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Vex

import "github.com/antlr4-go/antlr/v4"

// A complete Visitor for a parse tree produced by VexParser.
type VexVisitor interface {
	antlr.ParseTreeVisitor

	// Visit a parse tree produced by VexParser#sp.
	VisitSp(ctx *SpContext) interface{}

	// Visit a parse tree produced by VexParser#list.
	VisitList(ctx *ListContext) interface{}

	// Visit a parse tree produced by VexParser#array.
	VisitArray(ctx *ArrayContext) interface{}
}
