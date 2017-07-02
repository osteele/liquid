
//line scanner.rl:1
package expressions

import "fmt"
import "strconv"


//line scanner.go:10
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 12, 
	1, 13, 1, 14, 1, 15, 1, 16, 
	1, 17, 1, 18, 1, 19, 1, 20, 
	1, 21, 1, 22, 1, 23, 1, 24, 
	1, 25, 1, 26, 2, 2, 3, 2, 
	2, 4, 2, 2, 5, 2, 2, 6, 
	2, 2, 7, 2, 2, 8, 2, 2, 
	9, 2, 2, 10, 2, 2, 11, 
}

var _expression_key_offsets []byte = []byte{
	0, 1, 2, 3, 4, 5, 6, 7, 
	8, 9, 10, 12, 37, 40, 41, 42, 
	44, 45, 48, 50, 53, 54, 55, 56, 
	65, 75, 85, 95, 105, 115, 125, 135, 
	145, 155, 166, 176, 186, 196, 206, 216, 
	226, 236, 
}

var _expression_trans_keys []byte = []byte{
	34, 115, 115, 105, 103, 110, 111, 111, 
	112, 39, 48, 57, 32, 33, 34, 37, 
	39, 45, 46, 60, 61, 62, 95, 97, 
	99, 102, 105, 111, 116, 9, 13, 48, 
	57, 65, 90, 98, 122, 32, 9, 13, 
	61, 34, 97, 108, 39, 46, 48, 57, 
	48, 57, 46, 48, 57, 61, 61, 61, 
	45, 58, 95, 48, 57, 65, 90, 97, 
	122, 45, 58, 95, 110, 48, 57, 65, 
	90, 97, 122, 45, 58, 95, 100, 48, 
	57, 65, 90, 97, 122, 45, 58, 95, 
	111, 48, 57, 65, 90, 97, 122, 45, 
	58, 95, 110, 48, 57, 65, 90, 97, 
	122, 45, 58, 95, 116, 48, 57, 65, 
	90, 97, 122, 45, 58, 95, 97, 48, 
	57, 65, 90, 98, 122, 45, 58, 95, 
	105, 48, 57, 65, 90, 97, 122, 45, 
	58, 95, 110, 48, 57, 65, 90, 97, 
	122, 45, 58, 95, 115, 48, 57, 65, 
	90, 97, 122, 45, 58, 95, 97, 111, 
	48, 57, 65, 90, 98, 122, 45, 58, 
	95, 108, 48, 57, 65, 90, 97, 122, 
	45, 58, 95, 115, 48, 57, 65, 90, 
	97, 122, 45, 58, 95, 101, 48, 57, 
	65, 90, 97, 122, 45, 58, 95, 114, 
	48, 57, 65, 90, 97, 122, 45, 58, 
	95, 110, 48, 57, 65, 90, 97, 122, 
	45, 58, 95, 114, 48, 57, 65, 90, 
	97, 122, 45, 58, 95, 114, 48, 57, 
	65, 90, 97, 122, 45, 58, 95, 117, 
	48, 57, 65, 90, 97, 122, 
}

var _expression_single_lengths []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1, 
	1, 1, 0, 17, 1, 1, 1, 2, 
	1, 1, 0, 1, 1, 1, 1, 3, 
	4, 4, 4, 4, 4, 4, 4, 4, 
	4, 5, 4, 4, 4, 4, 4, 4, 
	4, 4, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 1, 4, 1, 0, 0, 0, 
	0, 1, 1, 1, 0, 0, 0, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 
}

var _expression_index_offsets []byte = []byte{
	0, 2, 4, 6, 8, 10, 12, 14, 
	16, 18, 20, 22, 44, 47, 49, 51, 
	54, 56, 59, 61, 64, 66, 68, 70, 
	77, 85, 93, 101, 109, 117, 125, 133, 
	141, 149, 158, 166, 174, 182, 190, 198, 
	206, 214, 
}

var _expression_indicies []byte = []byte{
	2, 1, 3, 0, 4, 0, 5, 0, 
	6, 0, 7, 0, 8, 0, 9, 0, 
	10, 0, 2, 11, 12, 0, 14, 15, 
	16, 17, 18, 19, 20, 22, 23, 24, 
	25, 26, 27, 28, 29, 30, 31, 14, 
	21, 25, 25, 13, 14, 14, 32, 34, 
	33, 2, 1, 35, 36, 33, 2, 11, 
	37, 21, 33, 12, 38, 12, 21, 39, 
	40, 33, 41, 33, 42, 33, 25, 43, 
	25, 25, 25, 25, 38, 25, 43, 25, 
	45, 25, 25, 25, 44, 25, 43, 25, 
	46, 25, 25, 25, 44, 25, 43, 25, 
	47, 25, 25, 25, 44, 25, 43, 25, 
	48, 25, 25, 25, 44, 25, 43, 25, 
	49, 25, 25, 25, 44, 25, 43, 25, 
	50, 25, 25, 25, 44, 25, 43, 25, 
	51, 25, 25, 25, 44, 25, 43, 25, 
	52, 25, 25, 25, 44, 25, 43, 25, 
	53, 25, 25, 25, 44, 25, 43, 25, 
	54, 55, 25, 25, 25, 44, 25, 43, 
	25, 56, 25, 25, 25, 44, 25, 43, 
	25, 57, 25, 25, 25, 44, 25, 43, 
	25, 58, 25, 25, 25, 44, 25, 43, 
	25, 59, 25, 25, 25, 44, 25, 43, 
	25, 60, 25, 25, 25, 44, 25, 43, 
	25, 61, 25, 25, 25, 44, 25, 43, 
	25, 62, 25, 25, 25, 44, 25, 43, 
	25, 57, 25, 25, 25, 44, 
}

var _expression_trans_targs []byte = []byte{
	11, 0, 11, 2, 3, 4, 5, 11, 
	7, 8, 11, 9, 18, 11, 12, 13, 
	14, 15, 16, 17, 18, 19, 20, 21, 
	22, 23, 24, 26, 33, 38, 39, 40, 
	11, 11, 11, 1, 6, 10, 11, 11, 
	11, 11, 11, 11, 11, 25, 23, 27, 
	28, 29, 30, 31, 32, 23, 34, 37, 
	35, 36, 23, 23, 23, 23, 41, 
}

var _expression_trans_actions []byte = []byte{
	33, 0, 11, 0, 0, 0, 0, 7, 
	0, 0, 9, 0, 37, 23, 0, 0, 
	5, 5, 5, 5, 61, 0, 0, 0, 
	0, 58, 0, 0, 0, 0, 0, 0, 
	29, 31, 15, 0, 0, 0, 35, 25, 
	19, 13, 17, 21, 27, 0, 43, 0, 
	0, 0, 0, 0, 0, 49, 0, 0, 
	0, 0, 40, 52, 55, 46, 0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 1, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 3, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 
}

var _expression_eof_trans []int16 = []int16{
	1, 1, 1, 1, 1, 1, 1, 1, 
	1, 1, 1, 0, 33, 34, 34, 34, 
	34, 34, 39, 40, 34, 34, 34, 39, 
	45, 45, 45, 45, 45, 45, 45, 45, 
	45, 45, 45, 45, 45, 45, 45, 45, 
	45, 45, 
}

const expression_start int = 11
const expression_first_final int = 11
const expression_error int = -1

const expression_en_main int = 11


//line scanner.rl:12


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
	
//line scanner.go:199
	{
	 lex.cs = expression_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line scanner.rl:31
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	
//line scanner.go:216
	{
	var _klen int
	var _trans int
	var _acts int
	var _nacts uint
	var _keys int
	if ( lex.p) == ( lex.pe) {
		goto _test_eof
	}
_resume:
	_acts = int(_expression_from_state_actions[ lex.cs])
	_nacts = uint(_expression_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		 _acts++
		switch _expression_actions[_acts - 1] {
		case 1:
//line NONE:1
 lex.ts = ( lex.p)

//line scanner.go:236
		}
	}

	_keys = int(_expression_key_offsets[ lex.cs])
	_trans = int(_expression_index_offsets[ lex.cs])

	_klen = int(_expression_single_lengths[ lex.cs])
	if _klen > 0 {
		_lower := int(_keys)
		var _mid int
		_upper := int(_keys + _klen - 1)
		for {
			if _upper < _lower {
				break
			}

			_mid = _lower + ((_upper - _lower) >> 1)
			switch {
			case  lex.data[( lex.p)] < _expression_trans_keys[_mid]:
				_upper = _mid - 1
			case  lex.data[( lex.p)] > _expression_trans_keys[_mid]:
				_lower = _mid + 1
			default:
				_trans += int(_mid - int(_keys))
				goto _match
			}
		}
		_keys += _klen
		_trans += _klen
	}

	_klen = int(_expression_range_lengths[ lex.cs])
	if _klen > 0 {
		_lower := int(_keys)
		var _mid int
		_upper := int(_keys + (_klen << 1) - 2)
		for {
			if _upper < _lower {
				break
			}

			_mid = _lower + (((_upper - _lower) >> 1) & ^1)
			switch {
			case  lex.data[( lex.p)] < _expression_trans_keys[_mid]:
				_upper = _mid - 2
			case  lex.data[( lex.p)] > _expression_trans_keys[_mid + 1]:
				_lower = _mid + 2
			default:
				_trans += int((_mid - int(_keys)) >> 1)
				goto _match
			}
		}
		_trans += _klen
	}

_match:
	_trans = int(_expression_indicies[_trans])
_eof_trans:
	 lex.cs = int(_expression_trans_targs[_trans])

	if _expression_trans_actions[_trans] == 0 {
		goto _again
	}

	_acts = int(_expression_trans_actions[_trans])
	_nacts = uint(_expression_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		_acts++
		switch _expression_actions[_acts-1] {
		case 2:
//line NONE:1
 lex.te = ( lex.p)+1

		case 3:
//line scanner.rl:58
 lex.act = 4;
		case 4:
//line scanner.rl:39
 lex.act = 6;
		case 5:
//line scanner.rl:94
 lex.act = 11;
		case 6:
//line scanner.rl:95
 lex.act = 12;
		case 7:
//line scanner.rl:96
 lex.act = 13;
		case 8:
//line scanner.rl:97
 lex.act = 14;
		case 9:
//line scanner.rl:98
 lex.act = 15;
		case 10:
//line scanner.rl:44
 lex.act = 17;
		case 11:
//line scanner.rl:102
 lex.act = 19;
		case 12:
//line scanner.rl:84
 lex.te = ( lex.p)+1
{ tok = ASSIGN; ( lex.p)++; goto _out
 }
		case 13:
//line scanner.rl:85
 lex.te = ( lex.p)+1
{ tok = LOOP; ( lex.p)++; goto _out
 }
		case 14:
//line scanner.rl:67
 lex.te = ( lex.p)+1
{
			tok = LITERAL
			// TODO unescape \x
			out.val = string(lex.data[lex.ts+1:lex.te-1])
			( lex.p)++; goto _out

		}
		case 15:
//line scanner.rl:90
 lex.te = ( lex.p)+1
{ tok = EQ; ( lex.p)++; goto _out
 }
		case 16:
//line scanner.rl:91
 lex.te = ( lex.p)+1
{ tok = NEQ; ( lex.p)++; goto _out
 }
		case 17:
//line scanner.rl:92
 lex.te = ( lex.p)+1
{ tok = GE; ( lex.p)++; goto _out
 }
		case 18:
//line scanner.rl:93
 lex.te = ( lex.p)+1
{ tok = LE; ( lex.p)++; goto _out
 }
		case 19:
//line scanner.rl:99
 lex.te = ( lex.p)+1
{ tok = KEYWORD; out.name = string(lex.data[lex.ts:lex.te-1]); ( lex.p)++; goto _out
 }
		case 20:
//line scanner.rl:102
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 21:
//line scanner.rl:49
 lex.te = ( lex.p)
( lex.p)--
{
			tok = LITERAL
			n, err := strconv.ParseInt(lex.token(), 10, 64)
			if err != nil {
				panic(err)
			}
			out.val = int(n)
			( lex.p)++; goto _out

		}
		case 22:
//line scanner.rl:44
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
		case 23:
//line scanner.rl:101
 lex.te = ( lex.p)
( lex.p)--

		case 24:
//line scanner.rl:102
 lex.te = ( lex.p)
( lex.p)--
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 25:
//line scanner.rl:102
( lex.p) = ( lex.te) - 1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 26:
//line NONE:1
	switch  lex.act {
	case 4:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			( lex.p)++; goto _out

		}
	case 6:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			out.val = lex.token() == "true"
			( lex.p)++; goto _out

		}
	case 11:
	{( lex.p) = ( lex.te) - 1
 tok = AND; ( lex.p)++; goto _out
 }
	case 12:
	{( lex.p) = ( lex.te) - 1
 tok = OR; ( lex.p)++; goto _out
 }
	case 13:
	{( lex.p) = ( lex.te) - 1
 tok = CONTAINS; ( lex.p)++; goto _out
 }
	case 14:
	{( lex.p) = ( lex.te) - 1
 tok = FOR; ( lex.p)++; goto _out
 }
	case 15:
	{( lex.p) = ( lex.te) - 1
 tok = IN; ( lex.p)++; goto _out
 }
	case 17:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
	case 19:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	}
	
//line scanner.go:484
		}
	}

_again:
	_acts = int(_expression_to_state_actions[ lex.cs])
	_nacts = uint(_expression_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		_acts++
		switch _expression_actions[_acts-1] {
		case 0:
//line NONE:1
 lex.ts = 0

//line scanner.go:498
		}
	}

	( lex.p)++
	if ( lex.p) != ( lex.pe) {
		goto _resume
	}
	_test_eof: {}
	if ( lex.p) == eof {
		if _expression_eof_trans[ lex.cs] > 0 {
			_trans = int(_expression_eof_trans[ lex.cs] - 1)
			goto _eof_trans
		}
	}

	_out: {}
	}

//line scanner.rl:106


	return tok
}

func (lex *lexer) Error(e string) {
    // fmt.Println("scan error:", e)
}