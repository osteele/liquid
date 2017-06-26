%{
package main
import (
    _ "fmt"
	"reflect"
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
%left DOT
%%
top: expr ';' { yylex.(*lexer).val = $1 };

expr:
  LITERAL { $$ = func(_ Context) interface{} { return $1 } }
| IDENTIFIER { $$ = func(ctx Context) interface{} { return ctx.Variables[$1] } }
| expr DOT IDENTIFIER {
	e, a := $1, $3
	$$ = func(ctx Context) interface{} {
		input := e(ctx)
		ref := reflect.ValueOf(input)
		switch ref.Kind() {
		case reflect.Map:
			return ref.MapIndex(reflect.ValueOf(a)).Interface()
		default:
			return input
		}
	}
}
;
