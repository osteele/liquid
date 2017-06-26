// Adapted from https://github.com/mhamrah/thermostat
package main

import "fmt"
import "strconv"

%%{
	machine expression;
	write data;
	access lex.;
	variable p lex.p;
	variable pe lex.pe;
}%%

type lexer struct {
    data []byte
    p, pe, cs int
    ts, te, act int
		val func(Context) interface{}
}

func newLexer(data []byte) *lexer {
	lex := &lexer{
			data: data,
			pe: len(data),
	}
	%% write init;
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	%%{
		action Bool {
			tok = LITERAL
			out.val = string(lex.data[lex.ts:lex.te]) == "true"
			fbreak;
		}
		action Ident {
			tok = IDENTIFIER
			out.name = string(lex.data[lex.ts:lex.te])
			fbreak;
		}
		action Number {
			tok = LITERAL
			n, err := strconv.ParseFloat(string(lex.data[lex.ts:lex.te]), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			fbreak;
		}
		action Relation { tok = RELATION; out.name = string(lex.data[lex.ts:lex.te]); fbreak; }

		ident = (alpha | '_') . (alnum | '_')* ;
		number = '-'? (digit+ ('.' digit*)?) ;

		main := |*
			number => Number;
			("true" | "false") => Bool;
			("==" | "!=" | ">" | ">" | ">=" | "<=") => Relation;
			("and" | "or" | "contains") => Relation;
			ident => Ident;
			space+;
		*|;

		write exec;
	}%%

	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}