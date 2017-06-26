package main

import "fmt"

type Expression struct {
	value func(Context) (interface{}, error)
}

func EvaluateExpr(expr string, ctx Context) (interface{}, error) {
	lexer := newLexer([]byte(expr + ";"))
	n := yyParse(lexer)
	if n != 0 {
		return nil, fmt.Errorf("parse error in %s", expr)
	}
	return lexer.val(ctx), nil
}
