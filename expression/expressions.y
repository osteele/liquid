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
   s        string
   ss       []string
   cycle    Cycle
   cyclefn  func(string) Cycle
   loop     Loop
   loopmods loopModifiers
   filter_params []valueFn
}
%type <f> expr rel filtered cond
%type<filter_params> filter_params
%type<cycle> cycle
%type<cyclefn> cycle2
%type<ss> cycle3
%type<loop> loop
%type<loopmods> loop_modifiers
%type<s> string
%token <val> LITERAL
%token <name> IDENTIFIER KEYWORD PROPERTY
%token ASSIGN CYCLE LOOP
%token EQ NEQ GE LE IN AND OR CONTAINS
%left '.' '|'
%left '<' '>'
%%
start:
  cond ';' { yylex.(*lexer).val = $1 }
| ASSIGN IDENTIFIER '=' filtered ';' {
	yylex.(*lexer).Assignment = Assignment{$2, &expression{$4}}
}
| CYCLE cycle ';' { yylex.(*lexer).Cycle = $2 }
| LOOP loop  ';' { yylex.(*lexer).Loop = $2 }
;

cycle: string cycle2 { $$ = $2($1) };

cycle2:
  ':' string cycle3 {
	h, t := $2, $3
	$$ = func(g string) Cycle { return Cycle{g, append([]string{h}, t...)} }
  }
| cycle3 {
	vals := $1
	$$ = func(h string) Cycle { return Cycle{Values: append([]string{h}, vals...)} }
  }
;

cycle3:
  /* empty */ { $$ = []string{} }
| ',' string cycle3 { $$ = append([]string{$2}, $3...) }
;

string: LITERAL {
	s, ok := $1.(string)
	if !ok {
		panic(ParseError(fmt.Sprintf("expected a string for %q", $1)))
	}
	$$ = s
};

loop: IDENTIFIER IN filtered loop_modifiers {
	name, expr, mods := $1, $3, $4
	$$ = Loop{name, &expression{expr}, mods}
}
;

loop_modifiers: /* empty */ { $$ = loopModifiers{} }
| loop_modifiers IDENTIFIER {
	switch $2 {
	case "reversed":
		$1.Reversed = true
	default:
		panic(ParseError(fmt.Sprintf("undefined loop modifier %q", $2)))
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
		panic(ParseError(fmt.Sprintf("undefined loop modifier %q", $2)))
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
