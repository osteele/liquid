%{
package expressions
import (
    "fmt"
	"reflect"
	"github.com/osteele/liquid/generics"
)

func init() {
	// This allows adding and removing references to fmt in the rules below,
	// without having to edit the import statement to avoid erorrs each time.
	_ = fmt.Sprint("")
}

%}
%union {
   name string
   val interface{}
   f func(Context) interface{}
   loopmods LoopModifiers
}
%type <f> expr rel expr1 loop
%type<loopmods> loop_modifiers
%token <val> LITERAL
%token <name> IDENTIFIER KEYWORD RELATION
%token ASSIGN LOOP
%token EQ FOR IN
%left '.' '|'
%left '<' '>'
%%
start:
  rel ';' { yylex.(*lexer).val = $1 }
| ASSIGN IDENTIFIER '=' expr1 ';' {
	name, expr := $2, $4
	yylex.(*lexer).val = func(ctx Context) interface{} {
		ctx.Set(name, expr(ctx))
		return nil
	}
}
| LOOP loop { yylex.(*lexer).val = $2 }
;

loop: IDENTIFIER IN expr1 loop_modifiers ';' {
	name, expr, mods := $1, $3, $4
	$$ = func(ctx Context) interface{} {
		return &Loop{name, expr(ctx), mods}
	}
}
;

loop_modifiers: /* empty */ { $$ = LoopModifiers{} }
| loop_modifiers IDENTIFIER {
	if $2 != "reversed" {
		panic(ParseError(fmt.Sprintf("undefined loop modifier: %s", $2)))
	}
	$1.Reversed = true
	$$ = $1
}
;

expr:
  LITERAL { val := $1; $$ = func(_ Context) interface{} { return val } }
| IDENTIFIER { name := $1; $$ = func(ctx Context) interface{} { return ctx.Get(name) } }
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

expr1:
  expr
| expr1 '|' IDENTIFIER { $$ = makeFilter($1, $3, nil) }
| expr1 '|' KEYWORD expr { $$ = makeFilter($1, $3, $4) }
;

rel:
  expr1
| expr EQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return generics.Equal(a, b)
	}
}
| expr '<' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return generics.Less(a, b)
	}
}
| expr '>' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return generics.Less(b, a)
	}
}
;
