%{
package expressions
import (
    "fmt"
	"github.com/osteele/liquid/generics"
)

func init() {
	// This allows adding and removing references to fmt in the rules below,
	// without having to edit the import statement to avoid erorrs each time.
	_ = fmt.Sprint("")
}

%}
%union {
   name     string
   val      interface{}
   f        func(Context) interface{}
   loopmods LoopModifiers
   filter_params []valueFn
}
%type <f> expr rel filtered cond loop
%type<filter_params> filter_params
%type<loopmods> loop_modifiers
%token <val> LITERAL
%token <name> IDENTIFIER KEYWORD RELATION
%token ASSIGN LOOP
%token EQ NEQ FOR IN AND OR CONTAINS
%left '.' '|'
%left '<' '>'
%%
start:
  cond ';' { yylex.(*lexer).val = $1 }
| ASSIGN IDENTIFIER '=' filtered ';' {
	name, expr := $2, $4
	yylex.(*lexer).val = func(ctx Context) interface{} {
		ctx.Set(name, expr(ctx))
		return nil
	}
}
| LOOP loop { yylex.(*lexer).val = $2 }
;

loop: IDENTIFIER IN filtered loop_modifiers ';' {
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
| expr '.' IDENTIFIER { $$ = makeObjectPropertyEvaluator($1, $3) }
| expr '[' expr ']' { $$ = makeIndexEvaluator($1, $3) }
| '(' cond ')' { $$ = $2 }
;

filtered:
  expr
| filtered '|' IDENTIFIER { $$ = makeFilter($1, $3, nil) }
| filtered '|' KEYWORD filter_params { $$ = makeFilter($1, $3, $4) }
;

filter_params:
  expr { $$ = []valueFn{$1} }
| filter_params ',' expr
  { $$ = append($1, $3) }

rel:
  filtered
| expr EQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return generics.Equal(a, b)
	}
}
| expr NEQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return !generics.Equal(a, b)
	}
}| expr '<' expr {
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

cond:
  rel
| cond AND rel {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		return generics.IsTrue(fa(ctx)) && generics.IsTrue(fb(ctx))
	}
}
| cond OR rel {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		return generics.IsTrue(fa(ctx)) || generics.IsTrue(fb(ctx))
	}
}
;
