
//line scanner.rl:1
// Adapted from https://github.com/mhamrah/thermostat
package main

import "fmt"
import "strconv"


//line scanner.go:11
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 8, 1, 9, 
	1, 10, 1, 11, 1, 12, 1, 13, 
	1, 14, 2, 2, 3, 2, 2, 4, 
	2, 2, 5, 2, 2, 6, 2, 2, 
	7, 
}

var _expression_key_offsets []byte = []byte{
	0, 0, 1, 4, 6, 30, 33, 35, 
	38, 39, 46, 54, 62, 70, 78, 86, 
	94, 102, 110, 118, 126, 134, 142, 150, 
	158, 166, 
}

var _expression_trans_keys []byte = []byte{
	61, 46, 48, 57, 48, 57, 32, 33, 
	45, 46, 59, 62, 91, 93, 95, 97, 
	99, 102, 111, 116, 9, 13, 48, 57, 
	60, 61, 65, 90, 98, 122, 32, 9, 
	13, 48, 57, 46, 48, 57, 61, 95, 
	48, 57, 65, 90, 97, 122, 95, 110, 
	48, 57, 65, 90, 97, 122, 95, 100, 
	48, 57, 65, 90, 97, 122, 95, 111, 
	48, 57, 65, 90, 97, 122, 95, 110, 
	48, 57, 65, 90, 97, 122, 95, 116, 
	48, 57, 65, 90, 97, 122, 95, 97, 
	48, 57, 65, 90, 98, 122, 95, 105, 
	48, 57, 65, 90, 97, 122, 95, 110, 
	48, 57, 65, 90, 97, 122, 95, 115, 
	48, 57, 65, 90, 97, 122, 95, 97, 
	48, 57, 65, 90, 98, 122, 95, 108, 
	48, 57, 65, 90, 97, 122, 95, 115, 
	48, 57, 65, 90, 97, 122, 95, 101, 
	48, 57, 65, 90, 97, 122, 95, 114, 
	48, 57, 65, 90, 97, 122, 95, 114, 
	48, 57, 65, 90, 97, 122, 95, 117, 
	48, 57, 65, 90, 97, 122, 
}

var _expression_single_lengths []byte = []byte{
	0, 1, 1, 0, 14, 1, 0, 1, 
	1, 1, 2, 2, 2, 2, 2, 2, 
	2, 2, 2, 2, 2, 2, 2, 2, 
	2, 2, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 1, 1, 5, 1, 1, 1, 
	0, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 
}

var _expression_index_offsets []byte = []byte{
	0, 0, 2, 5, 7, 27, 30, 32, 
	35, 37, 42, 48, 54, 60, 66, 72, 
	78, 84, 90, 96, 102, 108, 114, 120, 
	126, 132, 
}

var _expression_indicies []byte = []byte{
	0, 1, 2, 3, 1, 4, 1, 5, 
	6, 7, 8, 9, 10, 9, 9, 11, 
	12, 13, 14, 15, 16, 5, 3, 6, 
	11, 11, 1, 5, 5, 17, 4, 18, 
	4, 3, 19, 0, 20, 11, 11, 11, 
	11, 18, 11, 22, 11, 11, 11, 21, 
	11, 23, 11, 11, 11, 21, 11, 24, 
	11, 11, 11, 21, 11, 25, 11, 11, 
	11, 21, 11, 26, 11, 11, 11, 21, 
	11, 27, 11, 11, 11, 21, 11, 28, 
	11, 11, 11, 21, 11, 29, 11, 11, 
	11, 21, 11, 23, 11, 11, 11, 21, 
	11, 30, 11, 11, 11, 21, 11, 31, 
	11, 11, 11, 21, 11, 32, 11, 11, 
	11, 21, 11, 33, 11, 11, 11, 21, 
	11, 23, 11, 11, 11, 21, 11, 34, 
	11, 11, 11, 21, 11, 32, 11, 11, 
	11, 21, 
}

var _expression_trans_targs []byte = []byte{
	4, 0, 3, 7, 6, 5, 1, 2, 
	6, 4, 8, 9, 10, 12, 19, 23, 
	24, 4, 4, 4, 4, 4, 11, 9, 
	13, 14, 15, 16, 17, 18, 20, 21, 
	22, 9, 25, 
}

var _expression_trans_actions []byte = []byte{
	7, 0, 0, 0, 19, 0, 0, 0, 
	22, 5, 0, 31, 0, 0, 0, 0, 
	0, 15, 17, 9, 11, 13, 0, 28, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 25, 0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 1, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 3, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 
}

var _expression_eof_trans []byte = []byte{
	0, 0, 0, 0, 0, 18, 19, 20, 
	21, 19, 22, 22, 22, 22, 22, 22, 
	22, 22, 22, 22, 22, 22, 22, 22, 
	22, 22, 
}

const expression_start int = 4
const expression_first_final int = 4
const expression_error int = 0

const expression_en_main int = 4


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
	
//line scanner.go:158
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

	
//line scanner.go:175
	{
	var _klen int
	var _trans int
	var _acts int
	var _nacts uint
	var _keys int
	if ( lex.p) == ( lex.pe) {
		goto _test_eof
	}
	if  lex.cs == 0 {
		goto _out
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

//line scanner.go:198
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
//line scanner.rl:76
 lex.act = 3;
		case 5:
//line scanner.rl:40
 lex.act = 4;
		case 6:
//line scanner.rl:68
 lex.act = 6;
		case 7:
//line scanner.rl:45
 lex.act = 7;
		case 8:
//line scanner.rl:76
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
		case 9:
//line scanner.rl:68
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 10:
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
		case 11:
//line scanner.rl:68
 lex.te = ( lex.p)
( lex.p)--
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 12:
//line scanner.rl:45
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
		case 13:
//line scanner.rl:81
 lex.te = ( lex.p)
( lex.p)--

		case 14:
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
	case 3:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); ( lex.p)++; goto _out
 }
	case 4:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			out.val = lex.token() == "true"
			( lex.p)++; goto _out

		}
	case 6:
	{( lex.p) = ( lex.te) - 1
 tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
	case 7:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
	}
	
//line scanner.go:373
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

//line scanner.go:387
		}
	}

	if  lex.cs == 0 {
		goto _out
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

//line scanner.rl:85


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}