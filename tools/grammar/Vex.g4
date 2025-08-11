grammar Vex;

program: list+ EOF ;

list
    : '(' (array | list | SYMBOL | STRING)+ ')'
    ;

array
    : '[' (array | list| SYMBOL | STRING)* ']'
    ;

SYMBOL
    : (LETTER | INTEGER | '.' | '+' | '-' | '*' | '/' | '=' | '!' | '<' | '>' | '?' | '_' | ':' | '~' | '&')+
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