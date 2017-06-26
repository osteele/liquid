
//line scanner.rl:1
// Adapted from https://github.com/mhamrah/thermostat
package expressions

import "fmt"
import "strconv"


//line scanner.go:11
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 9, 
	1, 10, 1, 11, 1, 12, 1, 13, 
	1, 14, 1, 15, 1, 16, 1, 17, 
	1, 18, 1, 19, 1, 20, 2, 2, 
	3, 2, 2, 4, 2, 2, 5, 2, 
	2, 6, 2, 2, 7, 2, 2, 8, 
}

var _expression_key_offsets []byte = []byte{
	0, 1, 2, 3, 4, 5, 6, 7, 
	9, 36, 39, 40, 41, 42, 43, 46, 
	48, 51, 52, 59, 67, 75, 83, 91, 
	99, 107, 115, 123, 131, 139, 147, 155, 
	163, 171, 179, 
}

var _expression_trans_keys []byte = []byte{
	34, 115, 115, 105, 103, 110, 39, 48, 
	57, 32, 33, 34, 37, 39, 45, 46, 
	59, 61, 91, 93, 95, 97, 99, 102, 
	111, 116, 9, 13, 48, 57, 60, 62, 
	65, 90, 98, 122, 32, 9, 13, 61, 
	34, 97, 39, 46, 48, 57, 48, 57, 
	46, 48, 57, 61, 95, 48, 57, 65, 
	90, 97, 122, 95, 110, 48, 57, 65, 
	90, 97, 122, 95, 100, 48, 57, 65, 
	90, 97, 122, 95, 111, 48, 57, 65, 
	90, 97, 122, 95, 110, 48, 57, 65, 
	90, 97, 122, 95, 116, 48, 57, 65, 
	90, 97, 122, 95, 97, 48, 57, 65, 
	90, 98, 122, 95, 105, 48, 57, 65, 
	90, 97, 122, 95, 110, 48, 57, 65, 
	90, 97, 122, 95, 115, 48, 57, 65, 
	90, 97, 122, 95, 97, 48, 57, 65, 
	90, 98, 122, 95, 108, 48, 57, 65, 
	90, 97, 122, 95, 115, 48, 57, 65, 
	90, 97, 122, 95, 101, 48, 57, 65, 
	90, 97, 122, 95, 114, 48, 57, 65, 
	90, 97, 122, 95, 114, 48, 57, 65, 
	90, 97, 122, 95, 117, 48, 57, 65, 
	90, 97, 122, 
}

var _expression_single_lengths []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 0, 
	17, 1, 1, 1, 1, 1, 1, 0, 
	1, 1, 1, 2, 2, 2, 2, 2, 
	2, 2, 2, 2, 2, 2, 2, 2, 
	2, 2, 2, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 1, 
	5, 1, 0, 0, 0, 0, 1, 1, 
	1, 0, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 
}

var _expression_index_offsets []byte = []byte{
	0, 2, 4, 6, 8, 10, 12, 14, 
	16, 39, 42, 44, 46, 48, 50, 53, 
	55, 58, 60, 65, 71, 77, 83, 89, 
	95, 101, 107, 113, 119, 125, 131, 137, 
	143, 149, 155, 
}

var _expression_indicies []byte = []byte{
	2, 1, 3, 0, 4, 0, 5, 0, 
	6, 0, 7, 0, 2, 8, 9, 0, 
	11, 12, 13, 14, 15, 16, 17, 19, 
	21, 19, 19, 22, 23, 24, 25, 26, 
	27, 11, 18, 20, 22, 22, 10, 11, 
	11, 28, 30, 29, 2, 1, 32, 31, 
	2, 8, 33, 18, 31, 9, 29, 9, 
	18, 34, 35, 31, 22, 22, 22, 22, 
	29, 22, 37, 22, 22, 22, 36, 22, 
	38, 22, 22, 22, 36, 22, 39, 22, 
	22, 22, 36, 22, 40, 22, 22, 22, 
	36, 22, 41, 22, 22, 22, 36, 22, 
	42, 22, 22, 22, 36, 22, 43, 22, 
	22, 22, 36, 22, 44, 22, 22, 22, 
	36, 22, 38, 22, 22, 22, 36, 22, 
	45, 22, 22, 22, 36, 22, 46, 22, 
	22, 22, 36, 22, 47, 22, 22, 22, 
	36, 22, 48, 22, 22, 22, 36, 22, 
	38, 22, 22, 22, 36, 22, 49, 22, 
	22, 22, 36, 22, 47, 22, 22, 22, 
	36, 
}

var _expression_trans_targs []byte = []byte{
	8, 0, 8, 2, 3, 4, 5, 8, 
	6, 15, 8, 9, 10, 11, 12, 13, 
	14, 15, 16, 8, 10, 17, 18, 19, 
	21, 28, 32, 33, 8, 8, 8, 8, 
	1, 7, 8, 8, 8, 20, 18, 22, 
	23, 24, 25, 26, 27, 29, 30, 31, 
	18, 34, 
}

var _expression_trans_actions []byte = []byte{
	27, 0, 9, 0, 0, 0, 0, 7, 
	0, 31, 17, 0, 46, 5, 5, 5, 
	5, 37, 0, 11, 37, 0, 43, 0, 
	0, 0, 0, 0, 23, 29, 15, 25, 
	0, 0, 19, 13, 21, 0, 40, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	34, 0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	1, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 
	3, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 
}

var _expression_eof_trans []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1, 
	0, 29, 30, 32, 32, 32, 32, 30, 
	35, 32, 30, 37, 37, 37, 37, 37, 
	37, 37, 37, 37, 37, 37, 37, 37, 
	37, 37, 37, 
}

const expression_start int = 8
const expression_first_final int = 8
const expression_error int = -1

const expression_en_main int = 8


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
	
//line scanner.go:175
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

	
//line scanner.go:192
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

//line scanner.go:212
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
 lex.act = 3;
		case 4:
//line scanner.rl:40
 lex.act = 5;
		case 5:
//line scanner.rl:87
 lex.act = 6;
		case 6:
//line scanner.rl:74
 lex.act = 9;
		case 7:
//line scanner.rl:45
 lex.act = 10;
		case 8:
//line scanner.rl:93
 lex.act = 12;
		case 9:
//line scanner.rl:82
 lex.te = ( lex.p)+1
{ tok = ASSIGN; ( lex.p)++; goto _out
 }
		case 10:
//line scanner.rl:68
 lex.te = ( lex.p)+1
{
			tok = LITERAL
			// TODO unescape \x
			out.val = string(lex.data[lex.ts+1:lex.te-1])
			( lex.p)++; goto _out

		}
		case 11:
//line scanner.rl:87
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 12:
//line scanner.rl:88
 lex.te = ( lex.p)+1
{ tok = EQ; ( lex.p)++; goto _out
 }
		case 13:
//line scanner.rl:74
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 14:
//line scanner.rl:93
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 15:
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
		case 16:
//line scanner.rl:45
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
		case 17:
//line scanner.rl:92
 lex.te = ( lex.p)
( lex.p)--

		case 18:
//line scanner.rl:93
 lex.te = ( lex.p)
( lex.p)--
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 19:
//line scanner.rl:93
( lex.p) = ( lex.te) - 1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 20:
//line NONE:1
	switch  lex.act {
	case 3:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			( lex.p)++; goto _out

		}
	case 5:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			out.val = lex.token() == "true"
			( lex.p)++; goto _out

		}
	case 6:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	case 9:
	{( lex.p) = ( lex.te) - 1
 tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
	case 10:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
	case 12:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	}
	
//line scanner.go:424
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

//line scanner.go:438
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

//line scanner.rl:97


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}