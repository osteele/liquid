
//line scanner.rl:1
// Adapted from https://github.com/mhamrah/thermostat
package main

import "fmt"
import "strconv"


//line scanner.go:11
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3, 
	1, 4, 1, 5, 1, 6, 
}

var _expression_key_offsets []byte = []byte{
	0, 0, 1, 3, 21, 24, 27, 29, 
	30, 37, 45, 52, 60, 68, 76, 84, 
	92, 100, 107, 
}

var _expression_trans_keys []byte = []byte{
	61, 48, 57, 32, 33, 45, 62, 95, 
	97, 99, 111, 9, 13, 48, 57, 60, 
	61, 65, 90, 98, 122, 32, 9, 13, 
	46, 48, 57, 48, 57, 61, 95, 48, 
	57, 65, 90, 97, 122, 95, 110, 48, 
	57, 65, 90, 97, 122, 95, 48, 57, 
	65, 90, 97, 122, 95, 111, 48, 57, 
	65, 90, 97, 122, 95, 110, 48, 57, 
	65, 90, 97, 122, 95, 116, 48, 57, 
	65, 90, 97, 122, 95, 97, 48, 57, 
	65, 90, 98, 122, 95, 105, 48, 57, 
	65, 90, 97, 122, 95, 110, 48, 57, 
	65, 90, 97, 122, 95, 48, 57, 65, 
	90, 97, 122, 95, 48, 57, 65, 90, 
	97, 122, 
}

var _expression_single_lengths []byte = []byte{
	0, 1, 0, 8, 1, 1, 0, 1, 
	1, 2, 1, 2, 2, 2, 2, 2, 
	2, 1, 1, 
}

var _expression_range_lengths []byte = []byte{
	0, 0, 1, 5, 1, 1, 1, 0, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 
}

var _expression_index_offsets []byte = []byte{
	0, 0, 2, 4, 18, 21, 24, 26, 
	28, 33, 39, 44, 50, 56, 62, 68, 
	74, 80, 85, 
}

var _expression_indicies []byte = []byte{
	0, 1, 2, 1, 3, 4, 5, 6, 
	7, 8, 9, 10, 3, 2, 4, 7, 
	7, 1, 3, 3, 11, 13, 2, 12, 
	13, 12, 0, 14, 7, 7, 7, 7, 
	15, 7, 16, 7, 7, 7, 15, 7, 
	7, 7, 7, 15, 7, 17, 7, 7, 
	7, 15, 7, 18, 7, 7, 7, 15, 
	7, 19, 7, 7, 7, 15, 7, 20, 
	7, 7, 7, 15, 7, 21, 7, 7, 
	7, 15, 7, 22, 7, 7, 7, 15, 
	7, 7, 7, 7, 15, 7, 7, 7, 
	7, 15, 
}

var _expression_trans_targs []byte = []byte{
	3, 0, 5, 4, 1, 2, 7, 8, 
	9, 11, 18, 3, 3, 6, 3, 3, 
	10, 12, 13, 14, 15, 16, 17, 
}

var _expression_trans_actions []byte = []byte{
	5, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 13, 9, 0, 11, 7, 
	0, 0, 0, 0, 0, 0, 0, 
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 1, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 3, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 
}

var _expression_eof_trans []byte = []byte{
	0, 0, 0, 0, 12, 13, 13, 15, 
	16, 16, 16, 16, 16, 16, 16, 16, 
	16, 16, 16, 
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

func newLexer(data []byte) *lexer {
	lex := &lexer{
			data: data,
			pe: len(data),
	}
	
//line scanner.go:127
	{
	 lex.cs = expression_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line scanner.rl:28
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	
//line scanner.go:144
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

//line scanner.go:167
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
//line scanner.rl:51
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = string(lex.data[lex.ts:lex.te]); ( lex.p)++; goto _out
 }
		case 3:
//line scanner.rl:36
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			name := string(lex.data[lex.ts:lex.te])
			out.val = func(ctx Context) interface{} { return ctx.Variables[name] }
			( lex.p)++; goto _out

		}
		case 4:
//line scanner.rl:42
 lex.te = ( lex.p)
( lex.p)--
{
			tok = NUMBER
			n, err := strconv.ParseFloat(string(lex.data[lex.ts:lex.te]), 64)
			if err != nil {
				panic(err)
			}
			out.val = func(_ Context) interface{} { return n }
			( lex.p)++; goto _out

		}
		case 5:
//line scanner.rl:51
 lex.te = ( lex.p)
( lex.p)--
{ tok = RELATION; out.name = string(lex.data[lex.ts:lex.te]); ( lex.p)++; goto _out
 }
		case 6:
//line scanner.rl:61
 lex.te = ( lex.p)
( lex.p)--

//line scanner.go:278
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

//line scanner.go:292
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

//line scanner.rl:65


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}