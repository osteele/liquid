
//line scanner.rl:1
// Adapted from https://github.com/mhamrah/thermostat
package main

import "fmt"
import "strconv"


//line scanner.go:11
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 6, 1, 7, 
	1, 8, 1, 9, 1, 10, 1, 11, 
	2, 2, 3, 2, 2, 4, 2, 2, 
	5, 
}

var _expression_key_offsets []byte = []byte{
	0, 0, 1, 3, 23, 26, 29, 31, 
	32, 39, 47, 55, 63, 71, 79, 87, 
	95, 103, 111, 119, 127, 135, 143, 151, 
	159, 
}

var _expression_trans_keys []byte = []byte{
	61, 48, 57, 32, 33, 45, 62, 95, 
	97, 99, 102, 111, 116, 9, 13, 48, 
	57, 60, 61, 65, 90, 98, 122, 32, 
	9, 13, 46, 48, 57, 48, 57, 61, 
	95, 48, 57, 65, 90, 97, 122, 95, 
	110, 48, 57, 65, 90, 97, 122, 95, 
	100, 48, 57, 65, 90, 97, 122, 95, 
	111, 48, 57, 65, 90, 97, 122, 95, 
	110, 48, 57, 65, 90, 97, 122, 95, 
	116, 48, 57, 65, 90, 97, 122, 95, 
	97, 48, 57, 65, 90, 98, 122, 95, 
	105, 48, 57, 65, 90, 97, 122, 95, 
	110, 48, 57, 65, 90, 97, 122, 95, 
	115, 48, 57, 65, 90, 97, 122, 95, 
	97, 48, 57, 65, 90, 98, 122, 95, 
	108, 48, 57, 65, 90, 97, 122, 95, 
	115, 48, 57, 65, 90, 97, 122, 95, 
	101, 48, 57, 65, 90, 97, 122, 95, 
	114, 48, 57, 65, 90, 97, 122, 95, 
	114, 48, 57, 65, 90, 97, 122, 95, 
	117, 48, 57, 65, 90, 97, 122, 
}

var _expression_single_lengths []byte = []byte{
	0, 1, 0, 10, 1, 1, 0, 1, 
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
	0, 0, 2, 4, 20, 23, 26, 28, 
	30, 35, 41, 47, 53, 59, 65, 71, 
	77, 83, 89, 95, 101, 107, 113, 119, 
	125, 
}

var _expression_indicies []byte = []byte{
	0, 1, 2, 1, 3, 4, 5, 6, 
	7, 8, 9, 10, 11, 12, 3, 2, 
	4, 7, 7, 1, 3, 3, 13, 15, 
	2, 14, 15, 14, 0, 16, 7, 7, 
	7, 7, 17, 7, 19, 7, 7, 7, 
	18, 7, 20, 7, 7, 7, 18, 7, 
	21, 7, 7, 7, 18, 7, 22, 7, 
	7, 7, 18, 7, 23, 7, 7, 7, 
	18, 7, 24, 7, 7, 7, 18, 7, 
	25, 7, 7, 7, 18, 7, 26, 7, 
	7, 7, 18, 7, 20, 7, 7, 7, 
	18, 7, 27, 7, 7, 7, 18, 7, 
	28, 7, 7, 7, 18, 7, 29, 7, 
	7, 7, 18, 7, 30, 7, 7, 7, 
	18, 7, 20, 7, 7, 7, 18, 7, 
	31, 7, 7, 7, 18, 7, 29, 7, 
	7, 7, 18, 
}

var _expression_trans_targs []byte = []byte{
	3, 0, 5, 4, 1, 2, 7, 8, 
	9, 11, 18, 22, 23, 3, 3, 6, 
	3, 3, 3, 10, 8, 12, 13, 14, 
	15, 16, 17, 19, 20, 21, 8, 24, 
}

var _expression_trans_actions []byte = []byte{
	5, 0, 0, 0, 0, 0, 0, 23, 
	0, 0, 0, 0, 0, 13, 7, 0, 
	9, 15, 11, 0, 20, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 17, 0, 
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
	0, 0, 0, 0, 14, 15, 15, 17, 
	18, 19, 19, 19, 19, 19, 19, 19, 
	19, 19, 19, 19, 19, 19, 19, 19, 
	19, 
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
	
//line scanner.go:149
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

	
//line scanner.go:166
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

//line scanner.go:189
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
//line scanner.rl:36
 lex.act = 2;
		case 4:
//line scanner.rl:56
 lex.act = 4;
		case 5:
//line scanner.rl:41
 lex.act = 5;
		case 6:
//line scanner.rl:56
 lex.te = ( lex.p)+1
{ tok = RELATION; out.name = string(lex.data[lex.ts:lex.te]); ( lex.p)++; goto _out
 }
		case 7:
//line scanner.rl:47
 lex.te = ( lex.p)
( lex.p)--
{
			tok = LITERAL
			n, err := strconv.ParseFloat(string(lex.data[lex.ts:lex.te]), 64)
			if err != nil {
				panic(err)
			}
			out.val = func(_ Context) interface{} { return n }
			( lex.p)++; goto _out

		}
		case 8:
//line scanner.rl:56
 lex.te = ( lex.p)
( lex.p)--
{ tok = RELATION; out.name = string(lex.data[lex.ts:lex.te]); ( lex.p)++; goto _out
 }
		case 9:
//line scanner.rl:41
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			name := string(lex.data[lex.ts:lex.te])
			out.val = func(ctx Context) interface{} { return ctx.Variables[name] }
			( lex.p)++; goto _out

		}
		case 10:
//line scanner.rl:67
 lex.te = ( lex.p)
( lex.p)--

		case 11:
//line NONE:1
	switch  lex.act {
	case 2:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			val := string(lex.data[lex.ts:lex.te]) == "true"
			out.val = func(_ Context) interface{} { return val }
		}
	case 4:
	{( lex.p) = ( lex.te) - 1
 tok = RELATION; out.name = string(lex.data[lex.ts:lex.te]); ( lex.p)++; goto _out
 }
	case 5:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			name := string(lex.data[lex.ts:lex.te])
			out.val = func(ctx Context) interface{} { return ctx.Variables[name] }
			( lex.p)++; goto _out

		}
	}
	
//line scanner.go:338
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

//line scanner.go:352
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

//line scanner.rl:71


	return tok
}

func (lex *lexer) Error(e string) {
    fmt.Println("error:", e)
}