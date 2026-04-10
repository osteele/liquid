%{
package expressions
import (
	"fmt"
	"github.com/osteele/liquid/values"
)

func init() {
	// This allows adding and removing references to fmt in the rules below,
	// without having to comment and un-comment the import statement above.
	_ = ""
}

%}
%union {
   name     string
   val      any
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
%type<f> expr rel filtered cond
%type<filter_params> filter_params
%type<exprs> exprs expr2
%type<cycle> cycle
%type<cyclefn> cycle2
%type<ss> cycle3 assign_target
%type<loop> loop
%type<loopmods> loop_modifiers
%type<s> string
%token <val> LITERAL
%token <name> IDENTIFIER KEYWORD PROPERTY
%token ASSIGN CYCLE LOOP WHEN
%token EQ NEQ GE LE IN AND OR NOT CONTAINS DOTDOT
%token EMPTY BLANK
%right AND OR
%right NOT
%left '.' '|'
%left '<' '>'
%%
start:
  cond ';' { yylex.(*lexer).val = $1 }
| ASSIGN assign_target '=' cond ';' {
	path := $2
	var variable string
	if len(path) == 1 {
		variable = path[0]
	}
	yylex.(*lexer).Assignment = Assignment{Variable: variable, Path: path, ValueFn: &expression{evaluator: $4}}
}
| CYCLE cycle ';' { yylex.(*lexer).Cycle = $2 }
| LOOP loop ';'   { yylex.(*lexer).Loop = $2 }
| WHEN exprs ';'  { yylex.(*lexer).When = When{$2} }
;

assign_target:
  IDENTIFIER { $$ = []string{$1} }
| IDENTIFIER PROPERTY { $$ = []string{$1, $2} }
| assign_target PROPERTY { $$ = append($1, $2) }
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

exprs: expr expr2 { $$ = append([]Expression{&expression{evaluator: $1}}, $2...) } ;
expr2:
  /* empty */    { $$ = []Expression{} }
| ',' expr expr2 { $$ = append([]Expression{&expression{evaluator: $2}}, $3...) }
| OR expr expr2  { $$ = append([]Expression{&expression{evaluator: $2}}, $3...) }
;

string: LITERAL {
	s, ok := $1.(string)
	if !ok {
		panic(SyntaxError(fmt.Sprintf("expected a string for %q", $1)))
	}
	$$ = s
};

loop: IDENTIFIER IN filtered loop_modifiers {
	name, expr, mods := $1, $3, $4
	$$ = Loop{mods, name, &expression{evaluator: expr}}
}
;

loop_modifiers: /* empty */ { $$ = loopModifiers{} }
| loop_modifiers IDENTIFIER {
	switch $2 {
	case "reversed":
		$1.Reversed = true
	default:
		panic(SyntaxError(fmt.Sprintf("undefined loop modifier %q", $2)))
	}
	$$ = $1
}
| loop_modifiers KEYWORD expr {
    switch $2 {
	case "cols":
		$1.Cols = &expression{evaluator: $3}
	case "limit":
		$1.Limit = &expression{evaluator: $3}
	case "offset":
		$1.Offset = &expression{evaluator: $3}
	default:
		panic(SyntaxError(fmt.Sprintf("undefined loop modifier %q", $2)))
	}
	$$ = $1
}
;

expr:
  LITERAL { val := $1; $$ = func(Context) values.Value { return values.ValueOf(val) } }
| IDENTIFIER { name := $1; $$ = func(ctx Context) values.Value { return values.ValueOf(ctx.Get(name)) } }
| EMPTY { $$ = func(_ Context) values.Value { return values.EmptyDrop } }
| BLANK { $$ = func(_ Context) values.Value { return values.BlankDrop } }
| expr PROPERTY { $$ = makeObjectPropertyExpr($1, $2) }
| expr '.' IDENTIFIER { $$ = makeObjectPropertyExpr($1, $3) }
| expr '[' expr ']' { $$ = makeIndexExpr($1, $3) }
| '[' expr ']' { $$ = makeVariableIndirectionExpr($2) }
| '(' expr DOTDOT expr ')' { $$ = makeRangeExpr($2, $4) }
| '(' cond ')' { $$ = $2 }
;

filtered:
  expr
| filtered '|' IDENTIFIER { $$ = makeFilter($1, $3, nil) }
| filtered '|' KEYWORD filter_params { $$ = makeFilter($1, $3, $4) }
;

filter_params:
  expr { $$ = []valueFn{$1} }
| KEYWORD expr { $$ = []valueFn{makeNamedArgFn($1, $2)} }
| filter_params ',' expr
  { $$ = append($1, $3) }
| filter_params ',' KEYWORD expr
  { $$ = append($1, makeNamedArgFn($3, $4)) }

rel:
  filtered
| expr EQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		result := a.Equal(b)
		if c, ok := ctx.(*context); ok {
			c.callComparisonHook("==", a, b, result)
		}
		return values.ValueOf(result)
	}
}
| expr NEQ expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		result := !a.Equal(b)
		if c, ok := ctx.(*context); ok {
			c.callComparisonHook("!=", a, b, result)
		}
		return values.ValueOf(result)
	}
}
| expr '>' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		result := b.Less(a)
		if c, ok := ctx.(*context); ok {
			c.callComparisonHook(">", a, b, result)
		}
		return values.ValueOf(result)
	}
}
| expr '<' expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		result := a.Less(b)
		if c, ok := ctx.(*context); ok {
			c.callComparisonHook("<", a, b, result)
		}
		return values.ValueOf(result)
	}
}
| expr GE expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		result := b.Less(a) || a.Equal(b)
		if c, ok := ctx.(*context); ok {
			c.callComparisonHook(">=", a, b, result)
		}
		return values.ValueOf(result)
	}
}
| expr LE expr {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		a, b := fa(ctx), fb(ctx)
		result := a.Less(b) || a.Equal(b)
		if c, ok := ctx.(*context); ok {
			c.callComparisonHook("<=", a, b, result)
		}
		return values.ValueOf(result)
	}
}
| expr CONTAINS expr { $$ = makeContainsExpr($1, $3) }
;

cond:
  rel
| NOT cond {
	fa := $2
	$$ = func(ctx Context) values.Value {
		return values.ValueOf(!fa(ctx).Test())
	}
}
| cond AND cond {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		if c, ok := ctx.(*context); ok {
			c.callGroupBeginHook()
		}
		a := fa(ctx)
		b := fb(ctx)
		result := a.Test() && b.Test()
		if c, ok := ctx.(*context); ok {
			c.callGroupEndHook("and", result)
		}
		return values.ValueOf(result)
	}
}
| cond OR cond {
	fa, fb := $1, $3
	$$ = func(ctx Context) values.Value {
		if c, ok := ctx.(*context); ok {
			c.callGroupBeginHook()
		}
		a := fa(ctx)
		b := fb(ctx)
		result := a.Test() || b.Test()
		if c, ok := ctx.(*context); ok {
			c.callGroupEndHook("or", result)
		}
		return values.ValueOf(result)
	}
}
;
