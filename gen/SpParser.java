// Generated from /Users/thales.franca/study/study-parser/Sp.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.atn.*;
import org.antlr.v4.runtime.dfa.DFA;
import org.antlr.v4.runtime.*;
import org.antlr.v4.runtime.misc.*;
import org.antlr.v4.runtime.tree.*;
import java.util.List;
import java.util.Iterator;
import java.util.ArrayList;

@SuppressWarnings({"all", "warnings", "unchecked", "unused", "cast", "CheckReturnValue"})
public class SpParser extends Parser {
	static { RuntimeMetaData.checkVersion("4.13.1", RuntimeMetaData.VERSION); }

	protected static final DFA[] _decisionToDFA;
	protected static final PredictionContextCache _sharedContextCache =
		new PredictionContextCache();
	public static final int
		T__0=1, T__1=2, T__2=3, T__3=4, SYMBOL=5, STRING=6, LETTER=7, INTEGER=8, 
		TRASH=9, SEPARATOR=10;
	public static final int
		RULE_sp = 0, RULE_list = 1, RULE_array = 2, RULE_sym_block = 3;
	private static String[] makeRuleNames() {
		return new String[] {
			"sp", "list", "array", "sym_block"
		};
	}
	public static final String[] ruleNames = makeRuleNames();

	private static String[] makeLiteralNames() {
		return new String[] {
			null, "'('", "')'", "'['", "']'"
		};
	}
	private static final String[] _LITERAL_NAMES = makeLiteralNames();
	private static String[] makeSymbolicNames() {
		return new String[] {
			null, null, null, null, null, "SYMBOL", "STRING", "LETTER", "INTEGER", 
			"TRASH", "SEPARATOR"
		};
	}
	private static final String[] _SYMBOLIC_NAMES = makeSymbolicNames();
	public static final Vocabulary VOCABULARY = new VocabularyImpl(_LITERAL_NAMES, _SYMBOLIC_NAMES);

	/**
	 * @deprecated Use {@link #VOCABULARY} instead.
	 */
	@Deprecated
	public static final String[] tokenNames;
	static {
		tokenNames = new String[_SYMBOLIC_NAMES.length];
		for (int i = 0; i < tokenNames.length; i++) {
			tokenNames[i] = VOCABULARY.getLiteralName(i);
			if (tokenNames[i] == null) {
				tokenNames[i] = VOCABULARY.getSymbolicName(i);
			}

			if (tokenNames[i] == null) {
				tokenNames[i] = "<INVALID>";
			}
		}
	}

	@Override
	@Deprecated
	public String[] getTokenNames() {
		return tokenNames;
	}

	@Override

	public Vocabulary getVocabulary() {
		return VOCABULARY;
	}

	@Override
	public String getGrammarFileName() { return "Sp.g4"; }

	@Override
	public String[] getRuleNames() { return ruleNames; }

	@Override
	public String getSerializedATN() { return _serializedATN; }

	@Override
	public ATN getATN() { return _ATN; }

	public SpParser(TokenStream input) {
		super(input);
		_interp = new ParserATNSimulator(this,_ATN,_decisionToDFA,_sharedContextCache);
	}

	@SuppressWarnings("CheckReturnValue")
	public static class SpContext extends ParserRuleContext {
		public TerminalNode EOF() { return getToken(SpParser.EOF, 0); }
		public List<ListContext> list() {
			return getRuleContexts(ListContext.class);
		}
		public ListContext list(int i) {
			return getRuleContext(ListContext.class,i);
		}
		public SpContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_sp; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).enterSp(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).exitSp(this);
		}
		@Override
		public <T> T accept(ParseTreeVisitor<? extends T> visitor) {
			if ( visitor instanceof SpVisitor ) return ((SpVisitor<? extends T>)visitor).visitSp(this);
			else return visitor.visitChildren(this);
		}
	}

	public final SpContext sp() throws RecognitionException {
		SpContext _localctx = new SpContext(_ctx, getState());
		enterRule(_localctx, 0, RULE_sp);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			setState(9); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				{
				setState(8);
				list();
				}
				}
				setState(11); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( _la==T__0 );
			setState(13);
			match(EOF);
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ListContext extends ParserRuleContext {
		public List<ArrayContext> array() {
			return getRuleContexts(ArrayContext.class);
		}
		public ArrayContext array(int i) {
			return getRuleContext(ArrayContext.class,i);
		}
		public List<ListContext> list() {
			return getRuleContexts(ListContext.class);
		}
		public ListContext list(int i) {
			return getRuleContext(ListContext.class,i);
		}
		public List<Sym_blockContext> sym_block() {
			return getRuleContexts(Sym_blockContext.class);
		}
		public Sym_blockContext sym_block(int i) {
			return getRuleContext(Sym_blockContext.class,i);
		}
		public List<TerminalNode> STRING() { return getTokens(SpParser.STRING); }
		public TerminalNode STRING(int i) {
			return getToken(SpParser.STRING, i);
		}
		public List<TerminalNode> SEPARATOR() { return getTokens(SpParser.SEPARATOR); }
		public TerminalNode SEPARATOR(int i) {
			return getToken(SpParser.SEPARATOR, i);
		}
		public ListContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_list; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).enterList(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).exitList(this);
		}
		@Override
		public <T> T accept(ParseTreeVisitor<? extends T> visitor) {
			if ( visitor instanceof SpVisitor ) return ((SpVisitor<? extends T>)visitor).visitList(this);
			else return visitor.visitChildren(this);
		}
	}

	public final ListContext list() throws RecognitionException {
		ListContext _localctx = new ListContext(_ctx, getState());
		enterRule(_localctx, 2, RULE_list);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(15);
			match(T__0);
			setState(20); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				setState(20);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case T__2:
					{
					setState(16);
					array();
					}
					break;
				case T__0:
					{
					setState(17);
					list();
					}
					break;
				case SYMBOL:
				case SEPARATOR:
					{
					setState(18);
					sym_block();
					}
					break;
				case STRING:
					{
					setState(19);
					match(STRING);
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				setState(22); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( (((_la) & ~0x3f) == 0 && ((1L << _la) & 1130L) != 0) );
			setState(24);
			match(T__1);
			setState(30);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,4,_ctx) ) {
			case 1:
				{
				setState(26); 
				_errHandler.sync(this);
				_alt = 1;
				do {
					switch (_alt) {
					case 1:
						{
						{
						setState(25);
						match(SEPARATOR);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(28); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,3,_ctx);
				} while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				}
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class ArrayContext extends ParserRuleContext {
		public List<ArrayContext> array() {
			return getRuleContexts(ArrayContext.class);
		}
		public ArrayContext array(int i) {
			return getRuleContext(ArrayContext.class,i);
		}
		public List<ListContext> list() {
			return getRuleContexts(ListContext.class);
		}
		public ListContext list(int i) {
			return getRuleContext(ListContext.class,i);
		}
		public List<Sym_blockContext> sym_block() {
			return getRuleContexts(Sym_blockContext.class);
		}
		public Sym_blockContext sym_block(int i) {
			return getRuleContext(Sym_blockContext.class,i);
		}
		public List<TerminalNode> STRING() { return getTokens(SpParser.STRING); }
		public TerminalNode STRING(int i) {
			return getToken(SpParser.STRING, i);
		}
		public List<TerminalNode> SEPARATOR() { return getTokens(SpParser.SEPARATOR); }
		public TerminalNode SEPARATOR(int i) {
			return getToken(SpParser.SEPARATOR, i);
		}
		public ArrayContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_array; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).enterArray(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).exitArray(this);
		}
		@Override
		public <T> T accept(ParseTreeVisitor<? extends T> visitor) {
			if ( visitor instanceof SpVisitor ) return ((SpVisitor<? extends T>)visitor).visitArray(this);
			else return visitor.visitChildren(this);
		}
	}

	public final ArrayContext array() throws RecognitionException {
		ArrayContext _localctx = new ArrayContext(_ctx, getState());
		enterRule(_localctx, 4, RULE_array);
		int _la;
		try {
			int _alt;
			enterOuterAlt(_localctx, 1);
			{
			setState(32);
			match(T__2);
			setState(37); 
			_errHandler.sync(this);
			_la = _input.LA(1);
			do {
				{
				setState(37);
				_errHandler.sync(this);
				switch (_input.LA(1)) {
				case T__2:
					{
					setState(33);
					array();
					}
					break;
				case T__0:
					{
					setState(34);
					list();
					}
					break;
				case SYMBOL:
				case SEPARATOR:
					{
					setState(35);
					sym_block();
					}
					break;
				case STRING:
					{
					setState(36);
					match(STRING);
					}
					break;
				default:
					throw new NoViableAltException(this);
				}
				}
				setState(39); 
				_errHandler.sync(this);
				_la = _input.LA(1);
			} while ( (((_la) & ~0x3f) == 0 && ((1L << _la) & 1130L) != 0) );
			setState(41);
			match(T__3);
			setState(47);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,8,_ctx) ) {
			case 1:
				{
				setState(43); 
				_errHandler.sync(this);
				_alt = 1;
				do {
					switch (_alt) {
					case 1:
						{
						{
						setState(42);
						match(SEPARATOR);
						}
						}
						break;
					default:
						throw new NoViableAltException(this);
					}
					setState(45); 
					_errHandler.sync(this);
					_alt = getInterpreter().adaptivePredict(_input,7,_ctx);
				} while ( _alt!=2 && _alt!=org.antlr.v4.runtime.atn.ATN.INVALID_ALT_NUMBER );
				}
				break;
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	@SuppressWarnings("CheckReturnValue")
	public static class Sym_blockContext extends ParserRuleContext {
		public TerminalNode SYMBOL() { return getToken(SpParser.SYMBOL, 0); }
		public List<TerminalNode> SEPARATOR() { return getTokens(SpParser.SEPARATOR); }
		public TerminalNode SEPARATOR(int i) {
			return getToken(SpParser.SEPARATOR, i);
		}
		public Sym_blockContext(ParserRuleContext parent, int invokingState) {
			super(parent, invokingState);
		}
		@Override public int getRuleIndex() { return RULE_sym_block; }
		@Override
		public void enterRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).enterSym_block(this);
		}
		@Override
		public void exitRule(ParseTreeListener listener) {
			if ( listener instanceof SpListener ) ((SpListener)listener).exitSym_block(this);
		}
		@Override
		public <T> T accept(ParseTreeVisitor<? extends T> visitor) {
			if ( visitor instanceof SpVisitor ) return ((SpVisitor<? extends T>)visitor).visitSym_block(this);
			else return visitor.visitChildren(this);
		}
	}

	public final Sym_blockContext sym_block() throws RecognitionException {
		Sym_blockContext _localctx = new Sym_blockContext(_ctx, getState());
		enterRule(_localctx, 6, RULE_sym_block);
		int _la;
		try {
			enterOuterAlt(_localctx, 1);
			{
			{
			setState(50);
			_errHandler.sync(this);
			_la = _input.LA(1);
			if (_la==SEPARATOR) {
				{
				setState(49);
				match(SEPARATOR);
				}
			}

			setState(52);
			match(SYMBOL);
			setState(54);
			_errHandler.sync(this);
			switch ( getInterpreter().adaptivePredict(_input,10,_ctx) ) {
			case 1:
				{
				setState(53);
				match(SEPARATOR);
				}
				break;
			}
			}
			}
		}
		catch (RecognitionException re) {
			_localctx.exception = re;
			_errHandler.reportError(this, re);
			_errHandler.recover(this, re);
		}
		finally {
			exitRule();
		}
		return _localctx;
	}

	public static final String _serializedATN =
		"\u0004\u0001\n9\u0002\u0000\u0007\u0000\u0002\u0001\u0007\u0001\u0002"+
		"\u0002\u0007\u0002\u0002\u0003\u0007\u0003\u0001\u0000\u0004\u0000\n\b"+
		"\u0000\u000b\u0000\f\u0000\u000b\u0001\u0000\u0001\u0000\u0001\u0001\u0001"+
		"\u0001\u0001\u0001\u0001\u0001\u0001\u0001\u0004\u0001\u0015\b\u0001\u000b"+
		"\u0001\f\u0001\u0016\u0001\u0001\u0001\u0001\u0004\u0001\u001b\b\u0001"+
		"\u000b\u0001\f\u0001\u001c\u0003\u0001\u001f\b\u0001\u0001\u0002\u0001"+
		"\u0002\u0001\u0002\u0001\u0002\u0001\u0002\u0004\u0002&\b\u0002\u000b"+
		"\u0002\f\u0002\'\u0001\u0002\u0001\u0002\u0004\u0002,\b\u0002\u000b\u0002"+
		"\f\u0002-\u0003\u00020\b\u0002\u0001\u0003\u0003\u00033\b\u0003\u0001"+
		"\u0003\u0001\u0003\u0003\u00037\b\u0003\u0001\u0003\u0000\u0000\u0004"+
		"\u0000\u0002\u0004\u0006\u0000\u0000C\u0000\t\u0001\u0000\u0000\u0000"+
		"\u0002\u000f\u0001\u0000\u0000\u0000\u0004 \u0001\u0000\u0000\u0000\u0006"+
		"2\u0001\u0000\u0000\u0000\b\n\u0003\u0002\u0001\u0000\t\b\u0001\u0000"+
		"\u0000\u0000\n\u000b\u0001\u0000\u0000\u0000\u000b\t\u0001\u0000\u0000"+
		"\u0000\u000b\f\u0001\u0000\u0000\u0000\f\r\u0001\u0000\u0000\u0000\r\u000e"+
		"\u0005\u0000\u0000\u0001\u000e\u0001\u0001\u0000\u0000\u0000\u000f\u0014"+
		"\u0005\u0001\u0000\u0000\u0010\u0015\u0003\u0004\u0002\u0000\u0011\u0015"+
		"\u0003\u0002\u0001\u0000\u0012\u0015\u0003\u0006\u0003\u0000\u0013\u0015"+
		"\u0005\u0006\u0000\u0000\u0014\u0010\u0001\u0000\u0000\u0000\u0014\u0011"+
		"\u0001\u0000\u0000\u0000\u0014\u0012\u0001\u0000\u0000\u0000\u0014\u0013"+
		"\u0001\u0000\u0000\u0000\u0015\u0016\u0001\u0000\u0000\u0000\u0016\u0014"+
		"\u0001\u0000\u0000\u0000\u0016\u0017\u0001\u0000\u0000\u0000\u0017\u0018"+
		"\u0001\u0000\u0000\u0000\u0018\u001e\u0005\u0002\u0000\u0000\u0019\u001b"+
		"\u0005\n\u0000\u0000\u001a\u0019\u0001\u0000\u0000\u0000\u001b\u001c\u0001"+
		"\u0000\u0000\u0000\u001c\u001a\u0001\u0000\u0000\u0000\u001c\u001d\u0001"+
		"\u0000\u0000\u0000\u001d\u001f\u0001\u0000\u0000\u0000\u001e\u001a\u0001"+
		"\u0000\u0000\u0000\u001e\u001f\u0001\u0000\u0000\u0000\u001f\u0003\u0001"+
		"\u0000\u0000\u0000 %\u0005\u0003\u0000\u0000!&\u0003\u0004\u0002\u0000"+
		"\"&\u0003\u0002\u0001\u0000#&\u0003\u0006\u0003\u0000$&\u0005\u0006\u0000"+
		"\u0000%!\u0001\u0000\u0000\u0000%\"\u0001\u0000\u0000\u0000%#\u0001\u0000"+
		"\u0000\u0000%$\u0001\u0000\u0000\u0000&\'\u0001\u0000\u0000\u0000\'%\u0001"+
		"\u0000\u0000\u0000\'(\u0001\u0000\u0000\u0000()\u0001\u0000\u0000\u0000"+
		")/\u0005\u0004\u0000\u0000*,\u0005\n\u0000\u0000+*\u0001\u0000\u0000\u0000"+
		",-\u0001\u0000\u0000\u0000-+\u0001\u0000\u0000\u0000-.\u0001\u0000\u0000"+
		"\u0000.0\u0001\u0000\u0000\u0000/+\u0001\u0000\u0000\u0000/0\u0001\u0000"+
		"\u0000\u00000\u0005\u0001\u0000\u0000\u000013\u0005\n\u0000\u000021\u0001"+
		"\u0000\u0000\u000023\u0001\u0000\u0000\u000034\u0001\u0000\u0000\u0000"+
		"46\u0005\u0005\u0000\u000057\u0005\n\u0000\u000065\u0001\u0000\u0000\u0000"+
		"67\u0001\u0000\u0000\u00007\u0007\u0001\u0000\u0000\u0000\u000b\u000b"+
		"\u0014\u0016\u001c\u001e%\'-/26";
	public static final ATN _ATN =
		new ATNDeserializer().deserialize(_serializedATN.toCharArray());
	static {
		_decisionToDFA = new DFA[_ATN.getNumberOfDecisions()];
		for (int i = 0; i < _ATN.getNumberOfDecisions(); i++) {
			_decisionToDFA[i] = new DFA(_ATN.getDecisionState(i), i);
		}
	}
}