%{
package main
import (
    "fmt"
	"reflect"
)

func init() {
	_ = fmt.Sprint("")
}

%}
%union {
   name string
   val interface{}
   f func(Context) interface{}
}
%type <f> expr
%token <val> LITERAL
%token <name> IDENTIFIER RELATION
%left '.'
%%
start: expr ';' { yylex.(*lexer).val = $1 };

expr:
  LITERAL { val := $1; $$ = func(_ Context) interface{} { return val } }
| IDENTIFIER { name := $1; $$ = func(ctx Context) interface{} { return ctx.Variables[name] } }
| expr '.' IDENTIFIER {
	e, attr := $1, $3
	$$ = func(ctx Context) interface{} {
		obj := e(ctx)
		ref := reflect.ValueOf(obj)
		switch ref.Kind() {
		case reflect.Map:
			val := ref.MapIndex(reflect.ValueOf(attr))
			if val.Kind()!= reflect.Invalid {
				return val.Interface()
			}
		}
			return nil
	}
}
| expr '[' expr ']' {
	e, i := $1, $3
	$$ = func(ctx Context) interface{} {
		ref := reflect.ValueOf(e(ctx))
		index := reflect.ValueOf(i(ctx))
		switch ref.Kind() {
		case reflect.Array, reflect.Slice:
			switch index.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					n := int(index.Int())
					if 0 <= n && n < ref.Len() {
						return ref.Index(n).Interface()
					}
			}
		}
		return nil
	}
}
;

