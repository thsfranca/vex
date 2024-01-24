// Generated from /Users/thales.franca/study/study-parser/Sp.g4 by ANTLR 4.13.1
import org.antlr.v4.runtime.tree.ParseTreeVisitor;

/**
 * This interface defines a complete generic visitor for a parse tree produced
 * by {@link SpParser}.
 *
 * @param <T> The return type of the visit operation. Use {@link Void} for
 * operations with no return type.
 */
public interface SpVisitor<T> extends ParseTreeVisitor<T> {
	/**
	 * Visit a parse tree produced by {@link SpParser#sp}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitSp(SpParser.SpContext ctx);
	/**
	 * Visit a parse tree produced by {@link SpParser#list}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitList(SpParser.ListContext ctx);
	/**
	 * Visit a parse tree produced by {@link SpParser#array}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitArray(SpParser.ArrayContext ctx);
	/**
	 * Visit a parse tree produced by {@link SpParser#sym_block}.
	 * @param ctx the parse tree
	 * @return the visitor result
	 */
	T visitSym_block(SpParser.Sym_blockContext ctx);
}