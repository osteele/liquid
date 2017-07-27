%{
package expressions
import (
	"fmt"
	"math"
	"github.com/osteele/liquid/values"
)

func init() {
	// This allows adding and removing references to fmt in the rules below,
	// without having to comment and un-comment the import statement above.
	_ = fmt.Sprint("")
}

%}
%union {
   name     string
   val      interface{}
   f        func(Context) values.Value
   s        string
   ss       []string
   exprs    []Expression
   cycle    Cycle
   cyclefn  func(string) Cycle
   loop     Loop
   loopmods loopModifiers
   filter_params []valueFn
}
%type <f> expr rel filtered cond int_or_var loop_expr
%type<filter_params> filter_params
%type<exprs> exprs expr2
%type<cycle> cycle
%type<cyclefn> cycle2
%type<ss> cycle3
%type<loop> loop
%type<loopmods> loop_modifiers
%type<s> string
%token <val> LITERAL
%token <name> IDENTIFIER KEYWORD PROPERTY
%token ASSIGN CYCLE LOOP WHEN
%token EQ NEQ GE LE IN AND OR CONTAINS DOTDOT
%left '.' '|'
%left '<' '>'
%%
start:
  cond ';' { yylex.(*lexer).val = $1 }
| ASSIGN IDENTIFIER '=' filtered ';' {
	yylex.(*lexer).Assignment = Assignment{$2, &expression{$4}}
}
| CYCLE cycle ';' { yylex.(*lexer).Cycle = $2 }
| LOOP loop ';'   { yylex.(*lexer).Loop = $2 }
| WHEN exprs ';'  { yylex.(*lexer).When = When{$2} }
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

exprs: expr expr2 { $$ = append([]Expression{&expression{$1}}, $2...) } ;
expr2:
  /* empty */    { $$ = []Expression{} }
| ',' expr expr2 { $$ = append([]Expression{&expression{$2}}, $3...) }
;

string: LITERAL {
	s, ok := $1.(string)
	if !ok {
		panic(SyntaxError(fmt.Sprintf("expected a string for %q", $1)))
	}
	$$ = s
};

loop: IDENTIFIER IN loop_expr loop_modifiers {
	name, expr, mods := $1, $3, $4
	$$ = Loop{name, &expression{expr}, mods}
}
;

loop_expr : '(' int_or_var DOTDOT int_or_var ')' {
  $$ = makeRangeExpr($2, $4)
}
| filtered
;

// TODO DRY w/ expr
int_or_var:
  LITERAL { val := $1; $$ = func(Context) values.Value { return values.ValueOf(val) } }
| IDENTIFIER { name := $1; $$ = func(ctx Context) values.Value { return values.ValueOf(ctx.Get(name)) } }
;

loop_modifiers: /* empty */ { $$ = loopModifiers{Cols: math.MaxUint32} }
| loop_modifiers IDENTIFIER {
	switch $2 {
	case "reversed":
		$1.Reversed = true
	default:
		panic(SyntaxError(fmt.Sprintf("undefined loop modifier %q", $2)))
	}
	$$ = $1
}
| loop_modifiers KEYWORD LITERAL { // TODO can this be a variable?
	switch $2 {
	case "cols":
		cols, ok := $3.(int)
		if !ok {
			panic(SyntaxError(fmt.Sprintf("loop cols must an integer")))
		}
		$1.Cols = cols
	case "limit":
		limit, ok := $3.(int)
		if !ok {
			panic(SyntaxError(fmt.Sprintf("loop limit must an integer")))
		}
		$1.Limit = &limit
	case "offset":
		offset, ok := $3.(int)
		if !ok {
			panic(SyntaxError(fmt.Sprintf("loop offset must an integer")))
		}
		$1.Offset = offset
	default:
		panic(SyntaxError(fmt.Sprintf("undefined loop modifier %q", $2)))
	}
	$$ = $1
}
;

expr:
  LITERAL { val := $1; $$ = func(Context) values.Value { return values.ValueOf(val) } }
| IDENTIFIER { name := $1; $$ = func(ctx Context) values.Value { return values.ValueOf(ctx.Get(name)) } }
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
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		return values.ValueOf(a.Equal(b))
	}
}
| expr NEQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		return values.ValueOf(!a.Equal(b))
	}
}
| expr '>' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		return values.ValueOf(b.Less(a))
	}
}
| expr '<' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		return values.ValueOf(a.Less(b))
	}
}
| expr GE expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		return values.ValueOf(b.Less(a) || a.Equal(b))
	}
}
| expr LE expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		return values.ValueOf(a.Less(b) || a.Equal(b))
	}
}
| expr CONTAINS expr { $$ = makeContainsExpr($1, $3) }
;

cond:
  rel
| cond AND rel {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		return values.ValueOf(fa(ctx).Test() && fb(ctx).Test())
	}
}
| cond OR rel {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		return values.ValueOf(fa(ctx).Test() || fb(ctx).Test())
	}
}
;
