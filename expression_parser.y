%{
package main
import (
    _ "fmt"
)
%}
%union {
   name string
   val interface{}
   f func(Context) interface{}
}
%type <f> expr
%token <val> LITERAL
%token <name> IDENTIFIER RELATION
%%
top: expr { yylex.(*lexer).val = $1 };

expr:
  LITERAL { $$ = func(_ Context) interface{} { return $1 } }
| IDENTIFIER { $$ = func(ctx Context) interface{} { return ctx.Variables[$1] } }
