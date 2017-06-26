
//line scanner.rl:1
// Adapted from https://github.com/mhamrah/thermostat
package main

import "fmt"
import "strconv"


//line scanner.go:11
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 6, 1, 7, 
	1, 8, 1, 9, 1, 10, 1, 11, 
	1, 12, 1, 13, 2, 2, 3, 2, 
	2, 4, 2, 2, 5, 
}

var _expression_key_offsets []byte = []byte{
	0, 0, 1, 3, 25, 28, 31, 33, 
	34, 41, 49, 57, 65, 73, 81, 89, 
	97, 105, 113, 121, 129, 137, 145, 153, 
	161, 
}

var _expression_trans_keys []byte = []byte{
	61, 48, 57, 32, 33, 45, 46, 59, 
	62, 95, 97, 99, 102, 111, 116, 9, 
	13, 48, 57, 60, 61, 65, 90, 98, 
	122, 32, 9, 13, 46, 48, 57, 48, 
	57, 61, 95, 48, 57, 65, 90, 97, 
	122, 95, 110, 48, 57, 65, 90, 97, 
	122, 95, 100, 48, 57, 65, 90, 97, 
	122, 95, 111, 48, 57, 65, 90, 97, 
	122, 95, 110, 48, 57, 65, 90, 97, 
	122, 95, 116, 48, 57, 65, 90, 97, 
	122, 95, 97, 48, 57, 65, 90, 98, 
	122, 95, 105, 48, 57, 65, 90, 97, 
	122, 95, 110, 48, 57, 65, 90, 97, 
	122, 95, 115, 48, 57, 65, 90, 97, 
	122, 95, 97, 48, 57, 65, 90, 98, 
	122, 95, 108, 48, 57, 65, 90, 97, 
	122, 95, 115, 48, 57, 65, 90, 97, 
	122, 95, 101, 48, 57, 65, 90, 97, 
	122, 95, 114, 48, 57, 65, 90, 97, 
	122, 95, 114, 48, 57, 65, 90, 97, 
	122, 95, 117, 48, 57, 65, 90, 97, 
	122, 
}

var _expression_single_lengths []byte = []byte{
	0, 1, 0, 12, 1, 1, 0, 1, 
	1, 2, 2, 2, 2, 2, 2, 2, 
	2, 2, 2, 2, 2, 2, 2, 2, 
	2, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 1, 5, 1, 1, 1, 0, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 
}

var _expression_index_offsets []byte = []byte{
	0, 0, 2, 4, 22, 25, 28, 30, 
	32, 37, 43, 49, 55, 61, 67, 73, 
	79, 85, 91, 97, 103, 109, 115, 121, 
	127, 
}

var _expression_indicies []byte = []byte{
	0, 1, 2, 1, 3, 4, 5, 6, 
	7, 8, 9, 10, 11, 12, 13, 14, 
	3, 2, 4, 9, 9, 1, 3, 3, 
	15, 17, 2, 16, 17, 16, 0, 18, 
	9, 9, 9, 9, 19, 9, 21, 9, 
	9, 9, 20, 9, 22, 9, 9, 9, 
	20, 9, 23, 9, 9, 9, 20, 9, 
	24, 9, 9, 9, 20, 9, 25, 9, 
	9, 9, 20, 9, 26, 9, 9, 9, 
	20, 9, 27, 9, 9, 9, 20, 9, 
	28, 9, 9, 9, 20, 9, 22, 9, 
	9, 9, 20, 9, 29, 9, 9, 9, 
	20, 9, 30, 9, 9, 9, 20, 9, 
	31, 9, 9, 9, 20, 9, 32, 9, 
	9, 9, 20, 9, 22, 9, 9, 9, 
	20, 9, 33, 9, 9, 9, 20, 9, 
	31, 9, 9, 9, 20, 
}

var _expression_trans_targs []byte = []byte{
	3, 0, 5, 4, 1, 2, 3, 3, 
	7, 8, 9, 11, 18, 22, 23, 3, 
	3, 6, 3, 3, 3, 10, 8, 12, 
	13, 14, 15, 16, 17, 19, 20, 21, 
	8, 24, 
}

var _expression_trans_actions []byte = []byte{
	9, 0, 0, 0, 0, 0, 5, 7, 
	0, 27, 0, 0, 0, 0, 0, 17, 
	11, 0, 13, 19, 15, 0, 24, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	21, 0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 1, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 3, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 
}

var _expression_eof_trans []byte = []byte{
	0, 0, 0, 0, 16, 17, 17, 19, 
	20, 21, 21, 21, 21, 21, 21, 21, 
	21, 21, 21, 21, 21, 21, 21, 21, 
	21, 
}

const expression_start int = 3
const expression_first_final int = 3
const expression_error int = 0

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
	
//line scanner.go:156
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

	
//line scanner.go:173
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

//line scanner.go:196
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
//line scanner.rl:40
 lex.act = 4;
		case 4:
//line scanner.rl:59
 lex.act = 6;
		case 5:
//line scanner.rl:45
 lex.act = 7;
		case 6:
//line scanner.rl:66
 lex.te = ( lex.p)+1
{ tok = DOT; ( lex.p)++; goto _out
 }
		case 7:
//line scanner.rl:67
 lex.te = ( lex.p)+1
{ tok = ';'; ( lex.p)++; goto _out
 }
		case 8:
//line scanner.rl:59
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 9:
//line scanner.rl:50
 lex.te = ( lex.p)
( lex.p)--
{
			tok = LITERAL
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			( lex.p)++; goto _out

		}
		case 10:
//line scanner.rl:59
 lex.te = ( lex.p)
( lex.p)--
{ tok = RELATION; out.name = lex.token(); ( lex.p)++; goto _out
 }
		case 11:
//line scanner.rl:45
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			out.name = lex.token()
			( lex.p)++; goto _out

		}
		case 12:
//line scanner.rl:72
 lex.te = ( lex.p)
( lex.p)--

		case 13:
//line NONE:1
	switch  lex.act {
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
	
//line scanner.go:354
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

//line scanner.go:368
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

//line scanner.rl:76


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}