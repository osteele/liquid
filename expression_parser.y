%{
package main
import (
    _ "fmt"
)
%}
%union {
   name string
   val func(Context) interface{}
}
%type <val> expr
%token <val> LITERAL IDENTIFIER
%token <name> RELATION
%%
top: expr { yylex.(*lexer).val = $1 };

expr: LITERAL | IDENTIFIER;
