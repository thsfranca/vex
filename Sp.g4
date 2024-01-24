grammar Sp;

sp: list+ EOF ;

list
    : '(' (array | list | sym_block | STRING)+ ')' (SEPARATOR+)?
    ;

array
    : '[' (array | list| sym_block | STRING)+ ']' (SEPARATOR+)?
    ;

sym_block
    : (SEPARATOR? SYMBOL SEPARATOR?)
    ;

SYMBOL
    : (LETTER | INTEGER | STRING | '.')+
    ;

STRING
    : '"' ( ~'"' | '\\' '"' )* '"'
    ;

LETTER : [a-zA-Z];

INTEGER
    : [0-9]
    ;

fragment
SEPARATOR
    : ',' | ' '
    ;

fragment
WS : [ \n\r\t,] ;

fragment
COMMENT: ';' ~[\r\n]* ;

TRASH
    : ( WS | COMMENT | SEPARATOR) -> channel(HIDDEN)
    ;