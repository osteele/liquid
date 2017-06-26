
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
	1, 18, 1, 19, 2, 2, 3, 2, 
	2, 4, 2, 2, 5, 2, 2, 6, 
	2, 2, 7, 2, 2, 8, 
}

var _expression_key_offsets []byte = []byte{
	0, 1, 2, 4, 30, 33, 34, 35, 
	36, 39, 41, 44, 45, 52, 60, 68, 
	76, 84, 92, 100, 108, 116, 124, 132, 
	140, 148, 156, 164, 172, 
}

var _expression_trans_keys []byte = []byte{
	34, 39, 48, 57, 32, 33, 34, 39, 
	45, 46, 59, 61, 91, 93, 95, 97, 
	99, 102, 111, 116, 9, 13, 48, 57, 
	60, 62, 65, 90, 98, 122, 32, 9, 
	13, 61, 34, 39, 46, 48, 57, 48, 
	57, 46, 48, 57, 61, 95, 48, 57, 
	65, 90, 97, 122, 95, 110, 48, 57, 
	65, 90, 97, 122, 95, 100, 48, 57, 
	65, 90, 97, 122, 95, 111, 48, 57, 
	65, 90, 97, 122, 95, 110, 48, 57, 
	65, 90, 97, 122, 95, 116, 48, 57, 
	65, 90, 97, 122, 95, 97, 48, 57, 
	65, 90, 98, 122, 95, 105, 48, 57, 
	65, 90, 97, 122, 95, 110, 48, 57, 
	65, 90, 97, 122, 95, 115, 48, 57, 
	65, 90, 97, 122, 95, 97, 48, 57, 
	65, 90, 98, 122, 95, 108, 48, 57, 
	65, 90, 97, 122, 95, 115, 48, 57, 
	65, 90, 97, 122, 95, 101, 48, 57, 
	65, 90, 97, 122, 95, 114, 48, 57, 
	65, 90, 97, 122, 95, 114, 48, 57, 
	65, 90, 97, 122, 95, 117, 48, 57, 
	65, 90, 97, 122, 
}

var _expression_single_lengths []byte = []byte{
	1, 1, 0, 16, 1, 1, 1, 1, 
	1, 0, 1, 1, 1, 2, 2, 2, 
	2, 2, 2, 2, 2, 2, 2, 2, 
	2, 2, 2, 2, 2, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 1, 5, 1, 0, 0, 0, 
	1, 1, 1, 0, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 
}

var _expression_index_offsets []byte = []byte{
	0, 2, 4, 6, 28, 31, 33, 35, 
	37, 40, 42, 45, 47, 52, 58, 64, 
	70, 76, 82, 88, 94, 100, 106, 112, 
	118, 124, 130, 136, 142, 
}

var _expression_indicies []byte = []byte{
	2, 1, 2, 3, 4, 0, 6, 7, 
	8, 9, 10, 11, 13, 15, 13, 13, 
	16, 17, 18, 19, 20, 21, 6, 12, 
	14, 16, 16, 5, 6, 6, 22, 24, 
	23, 2, 1, 2, 3, 26, 12, 25, 
	4, 23, 4, 12, 27, 28, 25, 16, 
	16, 16, 16, 23, 16, 30, 16, 16, 
	16, 29, 16, 31, 16, 16, 16, 29, 
	16, 32, 16, 16, 16, 29, 16, 33, 
	16, 16, 16, 29, 16, 34, 16, 16, 
	16, 29, 16, 35, 16, 16, 16, 29, 
	16, 36, 16, 16, 16, 29, 16, 37, 
	16, 16, 16, 29, 16, 31, 16, 16, 
	16, 29, 16, 38, 16, 16, 16, 29, 
	16, 39, 16, 16, 16, 29, 16, 40, 
	16, 16, 16, 29, 16, 41, 16, 16, 
	16, 29, 16, 31, 16, 16, 16, 29, 
	16, 42, 16, 16, 16, 29, 16, 40, 
	16, 16, 16, 29, 
}

var _expression_trans_targs []byte = []byte{
	3, 0, 3, 1, 9, 3, 4, 5, 
	6, 7, 8, 9, 10, 3, 5, 11, 
	12, 13, 15, 22, 26, 27, 3, 3, 
	3, 3, 2, 3, 3, 3, 14, 12, 
	16, 17, 18, 19, 20, 21, 23, 24, 
	25, 12, 28, 
}

var _expression_trans_actions []byte = []byte{
	25, 0, 7, 0, 29, 15, 0, 44, 
	5, 5, 5, 35, 0, 9, 35, 0, 
	41, 0, 0, 0, 0, 0, 21, 27, 
	13, 23, 0, 17, 11, 19, 0, 38, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 32, 0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 1, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 3, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 
}

var _expression_eof_trans []byte = []byte{
	1, 1, 1, 0, 23, 24, 26, 26, 
	26, 24, 28, 26, 24, 30, 30, 30, 
	30, 30, 30, 30, 30, 30, 30, 30, 
	30, 30, 30, 30, 30, 
}

const expression_start int = 3
const expression_first_final int = 3
const expression_error int = -1

const expression_en_main int = 3


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
	
//line scanner.go:163
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

	
//line scanner.go:180
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

//line scanner.go:200
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
 lex.act = 2;
		case 4:
//line scanner.rl:40
 lex.act = 4;
		case 5:
//line scanner.rl:86
 lex.act = 5;
		case 6:
//line scanner.rl:74
 lex.act = 8;
		case 7:
//line scanner.rl:45
 lex.act = 9;
		case 8:
//line scanner.rl:92
 lex.act = 11;
		case 9:
//line scanner.rl:68
 lex.te = ( lex.p)+1
{
			tok = LITERAL
			// TODO unescape \x
			out.val = string(lex.data[lex.ts+1:lex.te-1])
			( lex.p)++; goto _out

		}
		case 10:
//line scanner.rl:86
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 11:
//line scanner.rl:87
 lex.te = ( lex.p)+1
{ tok = EQ; ( lex.p)++; goto _out
 }
		case 12:
//line scanner.rl:74
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 13:
//line scanner.rl:92
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 14:
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
		case 15:
//line scanner.rl:45
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
		case 16:
//line scanner.rl:91
 lex.te = ( lex.p)
( lex.p)--

		case 17:
//line scanner.rl:92
 lex.te = ( lex.p)
( lex.p)--
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 18:
//line scanner.rl:92
( lex.p) = ( lex.te) - 1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 19:
//line NONE:1
	switch  lex.act {
	case 2:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			( lex.p)++; goto _out

		}
	case 4:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			out.val = lex.token() == "true"
			( lex.p)++; goto _out

		}
	case 5:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	case 8:
	{( lex.p) = ( lex.te) - 1
 tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
	case 9:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
	case 11:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	}
	
//line scanner.go:407
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

//line scanner.go:421
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

//line scanner.rl:96


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}