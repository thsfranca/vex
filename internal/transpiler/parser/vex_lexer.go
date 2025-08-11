// Code generated from Vex.g4 by ANTLR 4.13.2. DO NOT EDIT.

package parser

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"sync"
	"unicode"
)

// Suppress unused import error
var _ = fmt.Printf
var _ = sync.Once{}
var _ = unicode.IsLetter

type VexLexer struct {
	*antlr.BaseLexer
	channelNames []string
	modeNames    []string
	// TODO: EOF string
}

var VexLexerLexerStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	ChannelNames           []string
	ModeNames              []string
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func vexlexerLexerInit() {
	staticData := &VexLexerLexerStaticData
	staticData.ChannelNames = []string{
		"DEFAULT_TOKEN_CHANNEL", "HIDDEN",
	}
	staticData.ModeNames = []string{
		"DEFAULT_MODE",
	}
	staticData.LiteralNames = []string{
		"", "'('", "')'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "SYMBOL", "STRING", "LETTER", "INTEGER", "TRASH",
	}
	staticData.RuleNames = []string{
		"T__0", "T__1", "T__2", "T__3", "SYMBOL", "STRING", "LETTER", "INTEGER",
		"SEPARATOR", "WS", "COMMENT", "TRASH",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 0, 9, 73, 6, -1, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2,
		4, 7, 4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 2, 8, 7, 8, 2, 9, 7, 9, 2,
		10, 7, 10, 2, 11, 7, 11, 1, 0, 1, 0, 1, 1, 1, 1, 1, 2, 1, 2, 1, 3, 1, 3,
		1, 4, 1, 4, 1, 4, 4, 4, 37, 8, 4, 11, 4, 12, 4, 38, 1, 5, 1, 5, 1, 5, 1,
		5, 5, 5, 45, 8, 5, 10, 5, 12, 5, 48, 9, 5, 1, 5, 1, 5, 1, 6, 1, 6, 1, 7,
		1, 7, 1, 8, 1, 8, 1, 9, 1, 9, 1, 10, 1, 10, 5, 10, 62, 8, 10, 10, 10, 12,
		10, 65, 9, 10, 1, 11, 1, 11, 1, 11, 3, 11, 70, 8, 11, 1, 11, 1, 11, 0,
		0, 12, 1, 1, 3, 2, 5, 3, 7, 4, 9, 5, 11, 6, 13, 7, 15, 8, 17, 0, 19, 0,
		21, 0, 23, 9, 1, 0, 7, 8, 0, 33, 33, 38, 38, 42, 43, 45, 47, 58, 58, 60,
		63, 95, 95, 126, 126, 1, 0, 34, 34, 2, 0, 65, 90, 97, 122, 1, 0, 48, 57,
		2, 0, 32, 32, 44, 44, 4, 0, 9, 10, 13, 13, 32, 32, 44, 44, 2, 0, 10, 10,
		13, 13, 77, 0, 1, 1, 0, 0, 0, 0, 3, 1, 0, 0, 0, 0, 5, 1, 0, 0, 0, 0, 7,
		1, 0, 0, 0, 0, 9, 1, 0, 0, 0, 0, 11, 1, 0, 0, 0, 0, 13, 1, 0, 0, 0, 0,
		15, 1, 0, 0, 0, 0, 23, 1, 0, 0, 0, 1, 25, 1, 0, 0, 0, 3, 27, 1, 0, 0, 0,
		5, 29, 1, 0, 0, 0, 7, 31, 1, 0, 0, 0, 9, 36, 1, 0, 0, 0, 11, 40, 1, 0,
		0, 0, 13, 51, 1, 0, 0, 0, 15, 53, 1, 0, 0, 0, 17, 55, 1, 0, 0, 0, 19, 57,
		1, 0, 0, 0, 21, 59, 1, 0, 0, 0, 23, 69, 1, 0, 0, 0, 25, 26, 5, 40, 0, 0,
		26, 2, 1, 0, 0, 0, 27, 28, 5, 41, 0, 0, 28, 4, 1, 0, 0, 0, 29, 30, 5, 91,
		0, 0, 30, 6, 1, 0, 0, 0, 31, 32, 5, 93, 0, 0, 32, 8, 1, 0, 0, 0, 33, 37,
		3, 13, 6, 0, 34, 37, 3, 15, 7, 0, 35, 37, 7, 0, 0, 0, 36, 33, 1, 0, 0,
		0, 36, 34, 1, 0, 0, 0, 36, 35, 1, 0, 0, 0, 37, 38, 1, 0, 0, 0, 38, 36,
		1, 0, 0, 0, 38, 39, 1, 0, 0, 0, 39, 10, 1, 0, 0, 0, 40, 46, 5, 34, 0, 0,
		41, 45, 8, 1, 0, 0, 42, 43, 5, 92, 0, 0, 43, 45, 5, 34, 0, 0, 44, 41, 1,
		0, 0, 0, 44, 42, 1, 0, 0, 0, 45, 48, 1, 0, 0, 0, 46, 44, 1, 0, 0, 0, 46,
		47, 1, 0, 0, 0, 47, 49, 1, 0, 0, 0, 48, 46, 1, 0, 0, 0, 49, 50, 5, 34,
		0, 0, 50, 12, 1, 0, 0, 0, 51, 52, 7, 2, 0, 0, 52, 14, 1, 0, 0, 0, 53, 54,
		7, 3, 0, 0, 54, 16, 1, 0, 0, 0, 55, 56, 7, 4, 0, 0, 56, 18, 1, 0, 0, 0,
		57, 58, 7, 5, 0, 0, 58, 20, 1, 0, 0, 0, 59, 63, 5, 59, 0, 0, 60, 62, 8,
		6, 0, 0, 61, 60, 1, 0, 0, 0, 62, 65, 1, 0, 0, 0, 63, 61, 1, 0, 0, 0, 63,
		64, 1, 0, 0, 0, 64, 22, 1, 0, 0, 0, 65, 63, 1, 0, 0, 0, 66, 70, 3, 19,
		9, 0, 67, 70, 3, 21, 10, 0, 68, 70, 3, 17, 8, 0, 69, 66, 1, 0, 0, 0, 69,
		67, 1, 0, 0, 0, 69, 68, 1, 0, 0, 0, 70, 71, 1, 0, 0, 0, 71, 72, 6, 11,
		0, 0, 72, 24, 1, 0, 0, 0, 7, 0, 36, 38, 44, 46, 63, 69, 1, 0, 1, 0,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// VexLexerInit initializes any static state used to implement VexLexer. By default the
// static state used to implement the lexer is lazily initialized during the first call to
// NewVexLexer(). You can call this function if you wish to initialize the static state ahead
// of time.
func VexLexerInit() {
	staticData := &VexLexerLexerStaticData
	staticData.once.Do(vexlexerLexerInit)
}

// NewVexLexer produces a new lexer instance for the optional input antlr.CharStream.
func NewVexLexer(input antlr.CharStream) *VexLexer {
	VexLexerInit()
	l := new(VexLexer)
	l.BaseLexer = antlr.NewBaseLexer(input)
	staticData := &VexLexerLexerStaticData
	l.Interpreter = antlr.NewLexerATNSimulator(l, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	l.channelNames = staticData.ChannelNames
	l.modeNames = staticData.ModeNames
	l.RuleNames = staticData.RuleNames
	l.LiteralNames = staticData.LiteralNames
	l.SymbolicNames = staticData.SymbolicNames
	l.GrammarFileName = "Vex.g4"
	// TODO: l.EOF = antlr.TokenEOF

	return l
}

// VexLexer tokens.
const (
	VexLexerT__0    = 1
	VexLexerT__1    = 2
	VexLexerT__2    = 3
	VexLexerT__3    = 4
	VexLexerSYMBOL  = 5
	VexLexerSTRING  = 6
	VexLexerLETTER  = 7
	VexLexerINTEGER = 8
	VexLexerTRASH   = 9
)
