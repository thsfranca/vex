// Code generated from Vex.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser // Vex

import "github.com/antlr4-go/antlr/v4"

// VexListener is a complete listener for a parse tree produced by VexParser.
type VexListener interface {
	antlr.ParseTreeListener

	// EnterProgram is called when entering the program production.
	EnterProgram(c *ProgramContext)

	// EnterList is called when entering the list production.
	EnterList(c *ListContext)

	// EnterArray is called when entering the array production.
	EnterArray(c *ArrayContext)

	// ExitProgram is called when exiting the program production.
	ExitProgram(c *ProgramContext)

	// ExitList is called when exiting the list production.
	ExitList(c *ListContext)

	// ExitArray is called when exiting the array production.
	ExitArray(c *ArrayContext)
}
