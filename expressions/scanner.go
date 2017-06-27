
//line scanner.rl:1
// Adapted from https://github.com/mhamrah/thermostat
package expressions

import "fmt"
import "strconv"


//line scanner.go:11
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 10, 
	1, 11, 1, 12, 1, 13, 1, 14, 
	1, 15, 1, 16, 1, 17, 1, 18, 
	1, 19, 1, 20, 1, 21, 1, 22, 
	2, 2, 3, 2, 2, 4, 2, 2, 
	5, 2, 2, 6, 2, 2, 7, 2, 
	2, 8, 2, 2, 9, 
}

var _expression_key_offsets []byte = []byte{
	0, 1, 2, 3, 4, 5, 6, 7, 
	8, 9, 10, 12, 37, 40, 41, 42, 
	44, 45, 48, 50, 53, 54, 62, 71, 
	80, 89, 98, 107, 116, 125, 134, 143, 
	153, 162, 171, 180, 189, 198, 207, 216, 
}

var _expression_trans_keys []byte = []byte{
	34, 115, 115, 105, 103, 110, 111, 111, 
	112, 39, 48, 57, 32, 33, 34, 37, 
	39, 45, 46, 61, 95, 97, 99, 102, 
	105, 111, 116, 9, 13, 48, 57, 60, 
	62, 65, 90, 98, 122, 32, 9, 13, 
	61, 34, 97, 108, 39, 46, 48, 57, 
	48, 57, 46, 48, 57, 61, 58, 95, 
	48, 57, 65, 90, 97, 122, 58, 95, 
	110, 48, 57, 65, 90, 97, 122, 58, 
	95, 100, 48, 57, 65, 90, 97, 122, 
	58, 95, 111, 48, 57, 65, 90, 97, 
	122, 58, 95, 110, 48, 57, 65, 90, 
	97, 122, 58, 95, 116, 48, 57, 65, 
	90, 97, 122, 58, 95, 97, 48, 57, 
	65, 90, 98, 122, 58, 95, 105, 48, 
	57, 65, 90, 97, 122, 58, 95, 110, 
	48, 57, 65, 90, 97, 122, 58, 95, 
	115, 48, 57, 65, 90, 97, 122, 58, 
	95, 97, 111, 48, 57, 65, 90, 98, 
	122, 58, 95, 108, 48, 57, 65, 90, 
	97, 122, 58, 95, 115, 48, 57, 65, 
	90, 97, 122, 58, 95, 101, 48, 57, 
	65, 90, 97, 122, 58, 95, 114, 48, 
	57, 65, 90, 97, 122, 58, 95, 110, 
	48, 57, 65, 90, 97, 122, 58, 95, 
	114, 48, 57, 65, 90, 97, 122, 58, 
	95, 114, 48, 57, 65, 90, 97, 122, 
	58, 95, 117, 48, 57, 65, 90, 97, 
	122, 
}

var _expression_single_lengths []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1, 
	1, 1, 0, 15, 1, 1, 1, 2, 
	1, 1, 0, 1, 1, 2, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 4, 
	3, 3, 3, 3, 3, 3, 3, 3, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 1, 5, 1, 0, 0, 0, 
	0, 1, 1, 1, 0, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
}

var _expression_index_offsets []byte = []byte{
	0, 2, 4, 6, 8, 10, 12, 14, 
	16, 18, 20, 22, 43, 46, 48, 50, 
	53, 55, 58, 60, 63, 65, 71, 78, 
	85, 92, 99, 106, 113, 120, 127, 134, 
	142, 149, 156, 163, 170, 177, 184, 191, 
}

var _expression_indicies []byte = []byte{
	2, 1, 3, 0, 4, 0, 5, 0, 
	6, 0, 7, 0, 8, 0, 9, 0, 
	10, 0, 2, 11, 12, 0, 14, 15, 
	16, 17, 18, 19, 20, 22, 23, 24, 
	25, 26, 27, 28, 29, 14, 21, 15, 
	23, 23, 13, 14, 14, 30, 32, 31, 
	2, 1, 33, 34, 31, 2, 11, 35, 
	21, 31, 12, 36, 12, 21, 37, 38, 
	31, 39, 23, 23, 23, 23, 36, 39, 
	23, 41, 23, 23, 23, 40, 39, 23, 
	42, 23, 23, 23, 40, 39, 23, 43, 
	23, 23, 23, 40, 39, 23, 44, 23, 
	23, 23, 40, 39, 23, 45, 23, 23, 
	23, 40, 39, 23, 46, 23, 23, 23, 
	40, 39, 23, 47, 23, 23, 23, 40, 
	39, 23, 48, 23, 23, 23, 40, 39, 
	23, 42, 23, 23, 23, 40, 39, 23, 
	49, 50, 23, 23, 23, 40, 39, 23, 
	51, 23, 23, 23, 40, 39, 23, 52, 
	23, 23, 23, 40, 39, 23, 53, 23, 
	23, 23, 40, 39, 23, 54, 23, 23, 
	23, 40, 39, 23, 55, 23, 23, 23, 
	40, 39, 23, 42, 23, 23, 23, 40, 
	39, 23, 56, 23, 23, 23, 40, 39, 
	23, 52, 23, 23, 23, 40, 
}

var _expression_trans_targs []byte = []byte{
	11, 0, 11, 2, 3, 4, 5, 11, 
	7, 8, 11, 9, 18, 11, 12, 13, 
	14, 15, 16, 17, 18, 19, 20, 21, 
	22, 24, 31, 36, 37, 38, 11, 11, 
	11, 1, 6, 10, 11, 11, 11, 11, 
	11, 23, 21, 25, 26, 27, 28, 29, 
	30, 32, 35, 33, 34, 21, 21, 21, 
	39, 
}

var _expression_trans_actions []byte = []byte{
	29, 0, 11, 0, 0, 0, 0, 7, 
	0, 0, 9, 0, 33, 19, 0, 0, 
	5, 5, 5, 5, 51, 0, 0, 48, 
	0, 0, 0, 0, 0, 0, 25, 27, 
	15, 0, 0, 0, 31, 21, 13, 17, 
	23, 0, 39, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 36, 42, 45, 
	0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 1, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 3, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
}

var _expression_eof_trans []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1, 
	1, 1, 1, 0, 31, 32, 32, 32, 
	32, 32, 37, 38, 32, 37, 41, 41, 
	41, 41, 41, 41, 41, 41, 41, 41, 
	41, 41, 41, 41, 41, 41, 41, 41, 
}

const expression_start int = 11
const expression_first_final int = 11
const expression_error int = -1

const expression_en_main int = 11


//line scanner.rl:13


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
	
//line scanner.go:187
	{
	 lex.cs = expression_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line scanner.rl:32
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	
//line scanner.go:204
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

//line scanner.go:224
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
//line scanner.rl:59
 lex.act = 4;
		case 4:
//line scanner.rl:40
 lex.act = 6;
		case 5:
//line scanner.rl:74
 lex.act = 9;
		case 6:
//line scanner.rl:91
 lex.act = 10;
		case 7:
//line scanner.rl:92
 lex.act = 11;
		case 8:
//line scanner.rl:45
 lex.act = 13;
		case 9:
//line scanner.rl:96
 lex.act = 15;
		case 10:
//line scanner.rl:82
 lex.te = ( lex.p)+1
{ tok = ASSIGN; ( lex.p)++; goto _out
 }
		case 11:
//line scanner.rl:83
 lex.te = ( lex.p)+1
{ tok = LOOP; ( lex.p)++; goto _out
 }
		case 12:
//line scanner.rl:68
 lex.te = ( lex.p)+1
{
			tok = LITERAL
			// TODO unescape \x
			out.val = string(lex.data[lex.ts+1:lex.te-1])
			( lex.p)++; goto _out

		}
		case 13:
//line scanner.rl:88
 lex.te = ( lex.p)+1
{ tok = EQ; ( lex.p)++; goto _out
 }
		case 14:
//line scanner.rl:74
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 15:
//line scanner.rl:93
 lex.te = ( lex.p)+1
{ tok = KEYWORD; out.name = string(lex.data[lex.ts:lex.te-1]); ( lex.p)++; goto _out
 }
		case 16:
//line scanner.rl:96
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 17:
//line scanner.rl:50
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
		case 18:
//line scanner.rl:45
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
		case 19:
//line scanner.rl:95
 lex.te = ( lex.p)
( lex.p)--

		case 20:
//line scanner.rl:96
 lex.te = ( lex.p)
( lex.p)--
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 21:
//line scanner.rl:96
( lex.p) = ( lex.te) - 1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 22:
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
	case 9:
	{( lex.p) = ( lex.te) - 1
 tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
	case 10:
	{( lex.p) = ( lex.te) - 1
 tok = IN; ( lex.p)++; goto _out
 }
	case 11:
	{( lex.p) = ( lex.te) - 1
 tok = IN; ( lex.p)++; goto _out
 }
	case 13:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
	case 15:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	}
	
//line scanner.go:448
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

//line scanner.go:462
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

//line scanner.rl:100


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}