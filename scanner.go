
//line scanner.rl:1
package main

// import "fmt"
import "strconv"


//line scanner.rl:7

func ScanExpression(data string) ([]Token, error) {
	cs, p, pe, eof := 0, 0, len(data), len(data)
  var (ts, te, act int)
  _ = act
  tokens := make([]Token, 0)

	
//line scanner.go:19
	{
	cs = lexer_start
	ts = 0
	te = 0
	act = 0
	}

//line scanner.go:27
	{
	var _klen int
	var _trans int
	var _acts int
	var _nacts uint
	var _keys int
	if p == pe {
		goto _test_eof
	}
	if cs == 0 {
		goto _out
	}
_resume:
	_acts = int(_lexer_from_state_actions[cs])
	_nacts = uint(_lexer_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		 _acts++
		switch _lexer_actions[_acts - 1] {
		case 1:
//line NONE:1
ts = p

//line scanner.go:50
		}
	}

	_keys = int(_lexer_key_offsets[cs])
	_trans = int(_lexer_index_offsets[cs])

	_klen = int(_lexer_single_lengths[cs])
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
			case data[p] < _lexer_trans_keys[_mid]:
				_upper = _mid - 1
			case data[p] > _lexer_trans_keys[_mid]:
				_lower = _mid + 1
			default:
				_trans += int(_mid - int(_keys))
				goto _match
			}
		}
		_keys += _klen
		_trans += _klen
	}

	_klen = int(_lexer_range_lengths[cs])
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
			case data[p] < _lexer_trans_keys[_mid]:
				_upper = _mid - 2
			case data[p] > _lexer_trans_keys[_mid + 1]:
				_lower = _mid + 2
			default:
				_trans += int((_mid - int(_keys)) >> 1)
				goto _match
			}
		}
		_trans += _klen
	}

_match:
	_trans = int(_lexer_indicies[_trans])
_eof_trans:
	cs = int(_lexer_trans_targs[_trans])

	if _lexer_trans_actions[_trans] == 0 {
		goto _again
	}

	_acts = int(_lexer_trans_actions[_trans])
	_nacts = uint(_lexer_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		_acts++
		switch _lexer_actions[_acts-1] {
		case 2:
//line scanner.rl:21
te = p+1
{tokens = append(tokens, Token{RelationType, data[ts:te], nil})}
		case 3:
//line scanner.rl:15
te = p
p--
{tokens = append(tokens, Token{IdentifierType, data[ts:te], nil})}
		case 4:
//line scanner.rl:16
te = p
p--
{
      n, err := strconv.ParseFloat(data[ts:te], 64)
      if err != nil {panic(err)}
      tokens = append(tokens, Token{ValueType, "", n})
    }
		case 5:
//line scanner.rl:21
te = p
p--
{tokens = append(tokens, Token{RelationType, data[ts:te], nil})}
		case 6:
//line scanner.rl:31
te = p
p--

//line scanner.go:148
		}
	}

_again:
	_acts = int(_lexer_to_state_actions[cs])
	_nacts = uint(_lexer_actions[_acts]); _acts++
	for ; _nacts > 0; _nacts-- {
		_acts++
		switch _lexer_actions[_acts-1] {
		case 0:
//line NONE:1
ts = 0

//line scanner.go:162
		}
	}

	if cs == 0 {
		goto _out
	}
	p++
	if p != pe {
		goto _resume
	}
	_test_eof: {}
	if p == eof {
		if _lexer_eof_trans[cs] > 0 {
			_trans = int(_lexer_eof_trans[cs] - 1)
			goto _eof_trans
		}
	}

	_out: {}
	}

//line scanner.rl:36


	return tokens, nil
}


//line scanner.go:191
var _lexer_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 3, 
	1, 4, 1, 5, 1, 6, 
}

var _lexer_key_offsets []byte = []byte{
	0, 0, 1, 3, 21, 24, 27, 29, 
	30, 37, 45, 52, 60, 68, 76, 84, 
	92, 100, 107, 
}

var _lexer_trans_keys []byte = []byte{
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

var _lexer_single_lengths []byte = []byte{
	0, 1, 0, 8, 1, 1, 0, 1, 
	1, 2, 1, 2, 2, 2, 2, 2, 
	2, 1, 1, 
}

var _lexer_range_lengths []byte = []byte{
	0, 0, 1, 5, 1, 1, 1, 0, 
	3, 3, 3, 3, 3, 3, 3, 3, 
	3, 3, 3, 
}

var _lexer_index_offsets []byte = []byte{
	0, 0, 2, 4, 18, 21, 24, 26, 
	28, 33, 39, 44, 50, 56, 62, 68, 
	74, 80, 85, 
}

var _lexer_indicies []byte = []byte{
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

var _lexer_trans_targs []byte = []byte{
	3, 0, 5, 4, 1, 2, 7, 8, 
	9, 11, 18, 3, 3, 6, 3, 3, 
	10, 12, 13, 14, 15, 16, 17, 
}

var _lexer_trans_actions []byte = []byte{
	5, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 13, 9, 0, 11, 7, 
	0, 0, 0, 0, 0, 0, 0, 
}

var _lexer_to_state_actions []byte = []byte{
	0, 0, 0, 1, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 
}

var _lexer_from_state_actions []byte = []byte{
	0, 0, 0, 3, 0, 0, 0, 0, 
	0, 0, 0, 0, 0, 0, 0, 0, 
	0, 0, 0, 
}

var _lexer_eof_trans []byte = []byte{
	0, 0, 0, 0, 12, 13, 13, 15, 
	16, 16, 16, 16, 16, 16, 16, 16, 
	16, 16, 16, 
}

const lexer_start int = 3
const lexer_first_final int = 3
const lexer_error int = 0

const lexer_en_main int = 3

