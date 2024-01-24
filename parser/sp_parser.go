// Code generated from Sp.g4 by ANTLR 4.13.1. DO NOT EDIT.

package parser // Sp

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type SpParser struct {
	*antlr.BaseParser
}

var SpParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func spParserInit() {
	staticData := &SpParserStaticData
	staticData.LiteralNames = []string{
		"", "'('", "')'", "'['", "']'",
	}
	staticData.SymbolicNames = []string{
		"", "", "", "", "", "SYMBOL", "STRING", "LETTER", "INTEGER", "SEPARATOR",
		"TRASH",
	}
	staticData.RuleNames = []string{
		"sp", "list", "array", "sym_block",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 10, 45, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 1, 0, 4,
		0, 10, 8, 0, 11, 0, 12, 0, 11, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
		4, 1, 21, 8, 1, 11, 1, 12, 1, 22, 1, 1, 1, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1,
		2, 4, 2, 32, 8, 2, 11, 2, 12, 2, 33, 1, 2, 1, 2, 1, 3, 3, 3, 39, 8, 3,
		1, 3, 1, 3, 3, 3, 43, 8, 3, 1, 3, 0, 0, 4, 0, 2, 4, 6, 0, 0, 51, 0, 9,
		1, 0, 0, 0, 2, 15, 1, 0, 0, 0, 4, 26, 1, 0, 0, 0, 6, 38, 1, 0, 0, 0, 8,
		10, 3, 2, 1, 0, 9, 8, 1, 0, 0, 0, 10, 11, 1, 0, 0, 0, 11, 9, 1, 0, 0, 0,
		11, 12, 1, 0, 0, 0, 12, 13, 1, 0, 0, 0, 13, 14, 5, 0, 0, 1, 14, 1, 1, 0,
		0, 0, 15, 20, 5, 1, 0, 0, 16, 21, 3, 2, 1, 0, 17, 21, 3, 4, 2, 0, 18, 21,
		3, 6, 3, 0, 19, 21, 5, 6, 0, 0, 20, 16, 1, 0, 0, 0, 20, 17, 1, 0, 0, 0,
		20, 18, 1, 0, 0, 0, 20, 19, 1, 0, 0, 0, 21, 22, 1, 0, 0, 0, 22, 20, 1,
		0, 0, 0, 22, 23, 1, 0, 0, 0, 23, 24, 1, 0, 0, 0, 24, 25, 5, 2, 0, 0, 25,
		3, 1, 0, 0, 0, 26, 31, 5, 3, 0, 0, 27, 32, 3, 4, 2, 0, 28, 32, 3, 2, 1,
		0, 29, 32, 3, 6, 3, 0, 30, 32, 5, 6, 0, 0, 31, 27, 1, 0, 0, 0, 31, 28,
		1, 0, 0, 0, 31, 29, 1, 0, 0, 0, 31, 30, 1, 0, 0, 0, 32, 33, 1, 0, 0, 0,
		33, 31, 1, 0, 0, 0, 33, 34, 1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 36, 5,
		4, 0, 0, 36, 5, 1, 0, 0, 0, 37, 39, 5, 9, 0, 0, 38, 37, 1, 0, 0, 0, 38,
		39, 1, 0, 0, 0, 39, 40, 1, 0, 0, 0, 40, 42, 5, 5, 0, 0, 41, 43, 5, 9, 0,
		0, 42, 41, 1, 0, 0, 0, 42, 43, 1, 0, 0, 0, 43, 7, 1, 0, 0, 0, 7, 11, 20,
		22, 31, 33, 38, 42,
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

// SpParserInit initializes any static state used to implement SpParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewSpParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func SpParserInit() {
	staticData := &SpParserStaticData
	staticData.once.Do(spParserInit)
}

// NewSpParser produces a new parser instance for the optional input antlr.TokenStream.
func NewSpParser(input antlr.TokenStream) *SpParser {
	SpParserInit()
	this := new(SpParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &SpParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Sp.g4"

	return this
}

// SpParser tokens.
const (
	SpParserEOF       = antlr.TokenEOF
	SpParserT__0      = 1
	SpParserT__1      = 2
	SpParserT__2      = 3
	SpParserT__3      = 4
	SpParserSYMBOL    = 5
	SpParserSTRING    = 6
	SpParserLETTER    = 7
	SpParserINTEGER   = 8
	SpParserSEPARATOR = 9
	SpParserTRASH     = 10
)

// SpParser rules.
const (
	SpParserRULE_sp        = 0
	SpParserRULE_list      = 1
	SpParserRULE_array     = 2
	SpParserRULE_sym_block = 3
)

// ISpContext is an interface to support dynamic dispatch.
type ISpContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllList() []IListContext
	List(i int) IListContext

	// IsSpContext differentiates from other interfaces.
	IsSpContext()
}

type SpContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySpContext() *SpContext {
	var p = new(SpContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_sp
	return p
}

func InitEmptySpContext(p *SpContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_sp
}

func (*SpContext) IsSpContext() {}

func NewSpContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *SpContext {
	var p = new(SpContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SpParserRULE_sp

	return p
}

func (s *SpContext) GetParser() antlr.Parser { return s.parser }

func (s *SpContext) EOF() antlr.TerminalNode {
	return s.GetToken(SpParserEOF, 0)
}

func (s *SpContext) AllList() []IListContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IListContext); ok {
			len++
		}
	}

	tst := make([]IListContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IListContext); ok {
			tst[i] = t.(IListContext)
			i++
		}
	}

	return tst
}

func (s *SpContext) List(i int) IListContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListContext)
}

func (s *SpContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *SpContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *SpContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.EnterSp(s)
	}
}

func (s *SpContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.ExitSp(s)
	}
}

func (p *SpParser) Sp() (localctx ISpContext) {
	localctx = NewSpContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, SpParserRULE_sp)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(9)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == SpParserT__0 {
		{
			p.SetState(8)
			p.List()
		}

		p.SetState(11)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(13)
		p.Match(SpParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IListContext is an interface to support dynamic dispatch.
type IListContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllList() []IListContext
	List(i int) IListContext
	AllArray() []IArrayContext
	Array(i int) IArrayContext
	AllSym_block() []ISym_blockContext
	Sym_block(i int) ISym_blockContext
	AllSTRING() []antlr.TerminalNode
	STRING(i int) antlr.TerminalNode

	// IsListContext differentiates from other interfaces.
	IsListContext()
}

type ListContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyListContext() *ListContext {
	var p = new(ListContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_list
	return p
}

func InitEmptyListContext(p *ListContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_list
}

func (*ListContext) IsListContext() {}

func NewListContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ListContext {
	var p = new(ListContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SpParserRULE_list

	return p
}

func (s *ListContext) GetParser() antlr.Parser { return s.parser }

func (s *ListContext) AllList() []IListContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IListContext); ok {
			len++
		}
	}

	tst := make([]IListContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IListContext); ok {
			tst[i] = t.(IListContext)
			i++
		}
	}

	return tst
}

func (s *ListContext) List(i int) IListContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListContext)
}

func (s *ListContext) AllArray() []IArrayContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IArrayContext); ok {
			len++
		}
	}

	tst := make([]IArrayContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IArrayContext); ok {
			tst[i] = t.(IArrayContext)
			i++
		}
	}

	return tst
}

func (s *ListContext) Array(i int) IArrayContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayContext)
}

func (s *ListContext) AllSym_block() []ISym_blockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISym_blockContext); ok {
			len++
		}
	}

	tst := make([]ISym_blockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISym_blockContext); ok {
			tst[i] = t.(ISym_blockContext)
			i++
		}
	}

	return tst
}

func (s *ListContext) Sym_block(i int) ISym_blockContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISym_blockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISym_blockContext)
}

func (s *ListContext) AllSTRING() []antlr.TerminalNode {
	return s.GetTokens(SpParserSTRING)
}

func (s *ListContext) STRING(i int) antlr.TerminalNode {
	return s.GetToken(SpParserSTRING, i)
}

func (s *ListContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ListContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ListContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.EnterList(s)
	}
}

func (s *ListContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.ExitList(s)
	}
}

func (p *SpParser) List() (localctx IListContext) {
	localctx = NewListContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, SpParserRULE_list)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(15)
		p.Match(SpParserT__0)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(20)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&618) != 0) {
		p.SetState(20)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SpParserT__0:
			{
				p.SetState(16)
				p.List()
			}

		case SpParserT__2:
			{
				p.SetState(17)
				p.Array()
			}

		case SpParserSYMBOL, SpParserSEPARATOR:
			{
				p.SetState(18)
				p.Sym_block()
			}

		case SpParserSTRING:
			{
				p.SetState(19)
				p.Match(SpParserSTRING)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(22)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(24)
		p.Match(SpParserT__1)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IArrayContext is an interface to support dynamic dispatch.
type IArrayContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllArray() []IArrayContext
	Array(i int) IArrayContext
	AllList() []IListContext
	List(i int) IListContext
	AllSym_block() []ISym_blockContext
	Sym_block(i int) ISym_blockContext
	AllSTRING() []antlr.TerminalNode
	STRING(i int) antlr.TerminalNode

	// IsArrayContext differentiates from other interfaces.
	IsArrayContext()
}

type ArrayContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyArrayContext() *ArrayContext {
	var p = new(ArrayContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_array
	return p
}

func InitEmptyArrayContext(p *ArrayContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_array
}

func (*ArrayContext) IsArrayContext() {}

func NewArrayContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ArrayContext {
	var p = new(ArrayContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SpParserRULE_array

	return p
}

func (s *ArrayContext) GetParser() antlr.Parser { return s.parser }

func (s *ArrayContext) AllArray() []IArrayContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IArrayContext); ok {
			len++
		}
	}

	tst := make([]IArrayContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IArrayContext); ok {
			tst[i] = t.(IArrayContext)
			i++
		}
	}

	return tst
}

func (s *ArrayContext) Array(i int) IArrayContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IArrayContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IArrayContext)
}

func (s *ArrayContext) AllList() []IListContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IListContext); ok {
			len++
		}
	}

	tst := make([]IListContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IListContext); ok {
			tst[i] = t.(IListContext)
			i++
		}
	}

	return tst
}

func (s *ArrayContext) List(i int) IListContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IListContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IListContext)
}

func (s *ArrayContext) AllSym_block() []ISym_blockContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ISym_blockContext); ok {
			len++
		}
	}

	tst := make([]ISym_blockContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ISym_blockContext); ok {
			tst[i] = t.(ISym_blockContext)
			i++
		}
	}

	return tst
}

func (s *ArrayContext) Sym_block(i int) ISym_blockContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ISym_blockContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(ISym_blockContext)
}

func (s *ArrayContext) AllSTRING() []antlr.TerminalNode {
	return s.GetTokens(SpParserSTRING)
}

func (s *ArrayContext) STRING(i int) antlr.TerminalNode {
	return s.GetToken(SpParserSTRING, i)
}

func (s *ArrayContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ArrayContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ArrayContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.EnterArray(s)
	}
}

func (s *ArrayContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.ExitArray(s)
	}
}

func (p *SpParser) Array() (localctx IArrayContext) {
	localctx = NewArrayContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, SpParserRULE_array)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(26)
		p.Match(SpParserT__2)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(31)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = ((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&618) != 0) {
		p.SetState(31)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case SpParserT__2:
			{
				p.SetState(27)
				p.Array()
			}

		case SpParserT__0:
			{
				p.SetState(28)
				p.List()
			}

		case SpParserSYMBOL, SpParserSEPARATOR:
			{
				p.SetState(29)
				p.Sym_block()
			}

		case SpParserSTRING:
			{
				p.SetState(30)
				p.Match(SpParserSTRING)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(33)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(35)
		p.Match(SpParserT__3)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ISym_blockContext is an interface to support dynamic dispatch.
type ISym_blockContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	SYMBOL() antlr.TerminalNode
	AllSEPARATOR() []antlr.TerminalNode
	SEPARATOR(i int) antlr.TerminalNode

	// IsSym_blockContext differentiates from other interfaces.
	IsSym_blockContext()
}

type Sym_blockContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptySym_blockContext() *Sym_blockContext {
	var p = new(Sym_blockContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_sym_block
	return p
}

func InitEmptySym_blockContext(p *Sym_blockContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = SpParserRULE_sym_block
}

func (*Sym_blockContext) IsSym_blockContext() {}

func NewSym_blockContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *Sym_blockContext {
	var p = new(Sym_blockContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = SpParserRULE_sym_block

	return p
}

func (s *Sym_blockContext) GetParser() antlr.Parser { return s.parser }

func (s *Sym_blockContext) SYMBOL() antlr.TerminalNode {
	return s.GetToken(SpParserSYMBOL, 0)
}

func (s *Sym_blockContext) AllSEPARATOR() []antlr.TerminalNode {
	return s.GetTokens(SpParserSEPARATOR)
}

func (s *Sym_blockContext) SEPARATOR(i int) antlr.TerminalNode {
	return s.GetToken(SpParserSEPARATOR, i)
}

func (s *Sym_blockContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *Sym_blockContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *Sym_blockContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.EnterSym_block(s)
	}
}

func (s *Sym_blockContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(SpListener); ok {
		listenerT.ExitSym_block(s)
	}
}

func (p *SpParser) Sym_block() (localctx ISym_blockContext) {
	localctx = NewSym_blockContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, SpParserRULE_sym_block)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(38)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	if _la == SpParserSEPARATOR {
		{
			p.SetState(37)
			p.Match(SpParserSEPARATOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	}
	{
		p.SetState(40)
		p.Match(SpParserSYMBOL)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	p.SetState(42)
	p.GetErrorHandler().Sync(p)

	if p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext()) == 1 {
		{
			p.SetState(41)
			p.Match(SpParserSEPARATOR)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	} else if p.HasError() { // JIM
		goto errorExit
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}
