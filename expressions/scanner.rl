// Adapted from https://github.com/mhamrah/thermostat
package expressions

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

func (l* lexer) token() string {
	return string(l.data[l.ts:l.te])
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
			out.val = lex.token() == "true"
			fbreak;
		}
		action Ident {
			tok = IDENTIFIER
			out.name = lex.token()
			fbreak;
		}
		action Int {
			tok = LITERAL
			n, err := strconv.ParseInt(lex.token(), 10, 64)
			if err != nil {
				panic(err)
			}
			out.val = int(n)
			fbreak;
		}
		action Float {
			tok = LITERAL
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			fbreak;
		}
		action String {
			tok = LITERAL
			// TODO unescape \x
			out.val = string(lex.data[lex.ts+1:lex.te-1])
			fbreak;
		}
		action Relation { tok = RELATION; out.name = lex.token(); fbreak; }

		ident = (alpha | '_') . (alnum | '_')* ;
		int = '-'? digit+ ;
		float = '-'? (digit+ '.' digit* | '.' digit+) ;
		string = '"' (any - '"')* '"' | "'" (any - "'")* "'" ; # TODO escapes

		main := |*
			int => Int;
			float => Float;
			string => String;
			("true" | "false") => Bool;
			[.;<>] | '[' | ']' => { tok = int(lex.data[lex.ts]); fbreak; };
			"==" => { tok = EQ; fbreak; };
			("!=" | ">" | ">" | ">=" | "<=") => Relation;
			("and" | "or" | "contains") => Relation;
			ident => Ident;
			space+;
			any => { tok = int(lex.data[lex.ts]); fbreak; };
		*|;

		write exec;
	}%%

	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}