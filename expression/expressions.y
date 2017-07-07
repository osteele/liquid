%{
package expression
import (
    "fmt"
	"github.com/osteele/liquid/evaluator"
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
   loopmods loopModifiers
   filter_params []valueFn
}
%type <f> expr rel filtered cond loop
%type<filter_params> filter_params
%type<loopmods> loop_modifiers
%token <val> LITERAL
%token <name> IDENTIFIER KEYWORD PROPERTY
%token ASSIGN LOOP
%token EQ NEQ GE LE FOR IN AND OR CONTAINS
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

loop_modifiers: /* empty */ { $$ = loopModifiers{} }
| loop_modifiers IDENTIFIER {
	switch $2 {
	case "reversed":
		$1.Reversed = true
	default:
		panic(ParseError(fmt.Sprintf("undefined loop modifier: %s", $2)))
	}
	$$ = $1
}
| loop_modifiers KEYWORD LITERAL { // TODO can this be a variable?
	switch $2 {
	case "limit":
		limit, ok := $3.(int)
		if !ok {
			panic(ParseError(fmt.Sprintf("loop limit must an integer")))
		}
		$1.Limit = &limit
	case "offset":
		offset, ok := $3.(int)
		if !ok {
			panic(ParseError(fmt.Sprintf("loop offset must an integer")))
		}
		$1.Offset = offset
	default:
		panic(ParseError(fmt.Sprintf("undefined loop modifier: %s", $2)))
	}
	$$ = $1
}
;

expr:
  LITERAL { val := $1; $$ = func(_ Context) interface{} { return val } }
| IDENTIFIER { name := $1; $$ = func(ctx Context) interface{} { return ctx.Get(name) } }
| expr PROPERTY { $$ = makeObjectPropertyExpr($1, $2) }
| expr '[' expr ']' { $$ = makeIndexExpr($1, $3) }
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
		return evaluator.Equal(a, b)
	}
}
| expr NEQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return !evaluator.Equal(a, b)
	}
}
| expr '>' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return evaluator.Less(b, a)
	}
}
| expr '<' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return evaluator.Less(a, b)
	}
}
| expr GE expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return evaluator.Less(b, a) || evaluator.Equal(a, b)
	}
}
| expr LE expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		a, b := fa(ctx), fb(ctx)
		return evaluator.Less(a, b) || evaluator.Equal(a, b)
	}
}
| expr CONTAINS expr { $$ = makeContainsExpr($1, $3) }
;

cond:
  rel
| cond AND rel {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		return evaluator.IsTrue(fa(ctx)) && evaluator.IsTrue(fb(ctx))
	}
}
| cond OR rel {
	fa, fb := $1, $3
	$$ = func(ctx Context) interface{} {
		return evaluator.IsTrue(fa(ctx)) || evaluator.IsTrue(fb(ctx))
	}
}
;
