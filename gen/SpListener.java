// Generated from /Users/thales.franca/study/study-parser/Sp.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.tree.ParseTreeListener;

/**
 * This interface defines a complete listener for a parse tree produced by
 * {@link SpParser}.
 */
public interface SpListener extends ParseTreeListener {
	/**
	 * Enter a parse tree produced by {@link SpParser#sp}.
	 * @param ctx the parse tree
	 */
	void enterSp(SpParser.SpContext ctx);
	/**
	 * Exit a parse tree produced by {@link SpParser#sp}.
	 * @param ctx the parse tree
	 */
	void exitSp(SpParser.SpContext ctx);
	/**
	 * Enter a parse tree produced by {@link SpParser#list}.
	 * @param ctx the parse tree
	 */
	void enterList(SpParser.ListContext ctx);
	/**
	 * Exit a parse tree produced by {@link SpParser#list}.
	 * @param ctx the parse tree
	 */
	void exitList(SpParser.ListContext ctx);
	/**
	 * Enter a parse tree produced by {@link SpParser#array}.
	 * @param ctx the parse tree
	 */
	void enterArray(SpParser.ArrayContext ctx);
	/**
	 * Exit a parse tree produced by {@link SpParser#array}.
	 * @param ctx the parse tree
	 */
	void exitArray(SpParser.ArrayContext ctx);
	/**
	 * Enter a parse tree produced by {@link SpParser#sym_block}.
	 * @param ctx the parse tree
	 */
	void enterSym_block(SpParser.Sym_blockContext ctx);
	/**
	 * Exit a parse tree produced by {@link SpParser#sym_block}.
	 * @param ctx the parse tree
	 */
	void exitSym_block(SpParser.Sym_blockContext ctx);
}