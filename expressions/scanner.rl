package expressions

import (
	"strconv"
	"strings"
	"unicode"
)

%%{
	machine expression;
	write data;
	access lex.;
	variable p lex.p;
	variable pe lex.pe;

	utf8_cont = 0x80..0xBF;

	# UTF-8 sequence patterns
	utf8_2byte = (0xC2..0xDF) utf8_cont;
	utf8_3byte = (0xE0..0xEF) utf8_cont{2};
	utf8_4byte = (0xF0..0xF7) utf8_cont{3};
	utf8_char = utf8_2byte | utf8_3byte | utf8_4byte;

	# Unicode identifier - match ASCII OR multi-byte UTF-8 sequences
	unicode_first = (alpha | '_' | utf8_char);
	unicode_tail = (alnum | '_' | '-' | utf8_char);
}%%

type lexer struct {
	parseValue
    data []byte
    p, pe, cs int
    ts, te, act int
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
		action Identifier {
			tok = IDENTIFIER
			t := lex.token()

			if !isValidUnicodeIdentifier(t) {
				panic("syntax error in identifier " + t)
			}

			out.name = t
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
			raw := string(lex.data[lex.ts+1:lex.te-1])
			if lex.data[lex.ts] == '"' {
				out.val = unescapeString(raw)
			} else {
				out.val = raw
			}
			fbreak;
		}
		action Relation { tok = RELATION; out.name = lex.token(); fbreak; }

		identifier = unicode_first unicode_tail* '?'?;
		property = '.' unicode_first unicode_tail* '?'?;

		int = '-'? digit+ ;
		float = '-'? digit+ ('.' digit+)? ;
		dq_char = (any - '"' - '\\') | ('\\' any);
		string = '"' dq_char* '"' | "'" (any - "'")* "'" ;

		main := |*
			# statement selectors, should match constants in parser.go
			"%assign " => { tok = ASSIGN; fbreak; };
			"{%cycle " => { tok = CYCLE; fbreak; };
			"%loop "   => { tok = LOOP; fbreak; };
			"{%when "  => { tok = WHEN; fbreak; };

			# literals
			int => Int;
			float => Float;
			string => String;

			# constants
			("true" | "false") => Bool;
			"nil" => { tok = LITERAL; out.val = nil; fbreak; };

			# relations
			"==" => { tok = EQ; fbreak; };
			"!=" => { tok = NEQ; fbreak; };
			">=" => { tok = GE; fbreak; };
			"<=" => { tok = LE; fbreak; };
			"and" => { tok = AND; fbreak; };
			"or" => { tok = OR; fbreak; };
			"contains" => { tok = CONTAINS; fbreak; };

			# keywords
			"in" => { tok = IN; fbreak; };
			".." => { tok = DOTDOT; fbreak; };

			identifier ':' => { tok = KEYWORD; out.name = string(lex.data[lex.ts:lex.te-1]); fbreak; };
			identifier => Identifier;
			property => { tok = PROPERTY; out.name = string(lex.data[lex.ts+1:lex.te]); fbreak; };

			space+;
			any => { tok = int(lex.data[lex.ts]); fbreak; };
		*|;

		write exec;
	}%%

	return tok
}

func (lex *lexer) Error(e string) {
    // fmt.Println("scan error:", e)
}

// unescapeString processes escape sequences in double-quoted strings.
// Supported: \\, \", \n, \t, \r. Unknown sequences pass through as-is.
func unescapeString(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			i++
			switch s[i] {
			case '\\':
				b.WriteByte('\\')
			case '"':
				b.WriteByte('"')
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			default:
				b.WriteByte('\\')
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func isValidUnicodeIdentifier(s string) bool {
	// Remove optional trailing '?' for validation
	checkStr := s
	if len(s) > 0 && s[len(s)-1] == '?' {
		checkStr = s[:len(s)-1]
	}

	if len(checkStr) == 0 {
		return false // "?" alone is invalid
	}

	// validate the core identifier part
	runes := []rune(checkStr)
	if !unicode.IsLetter(runes[0]) && runes[0] != '_' {
		return false
	}

	for _, r := range runes[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsMark(r) && r != '_' && r != '-' {
			return false
		}
	}

	return true
}
