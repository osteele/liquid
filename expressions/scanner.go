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
	1, 18, 2, 2, 3, 2, 2, 4,
	2, 2, 5, 2, 2, 6, 2, 2,
	7, 2, 2, 8,
}

var _expression_key_offsets []byte = []byte{
	0, 2, 26, 29, 30, 33, 35, 38,
	39, 46, 54, 62, 70, 78, 86, 94,
	102, 110, 118, 126, 134, 142, 150, 158,
	166,
}

var _expression_trans_keys []byte = []byte{
	48, 57, 32, 33, 45, 46, 59, 61,
	91, 93, 95, 97, 99, 102, 111, 116,
	9, 13, 48, 57, 60, 62, 65, 90,
	98, 122, 32, 9, 13, 61, 46, 48,
	57, 48, 57, 46, 48, 57, 61, 95,
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
	0, 14, 1, 1, 1, 0, 1, 1,
	1, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2,
	2,
}

var _expression_range_lengths []byte = []byte{
	1, 5, 1, 0, 1, 1, 1, 0,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3,
}

var _expression_index_offsets []byte = []byte{
	0, 2, 22, 25, 27, 30, 32, 35,
	37, 42, 48, 54, 60, 66, 72, 78,
	84, 90, 96, 102, 108, 114, 120, 126,
	132,
}

var _expression_indicies []byte = []byte{
	1, 0, 3, 4, 5, 6, 8, 10,
	8, 8, 11, 12, 13, 14, 15, 16,
	3, 7, 9, 11, 11, 2, 3, 3,
	17, 19, 18, 21, 7, 20, 1, 18,
	1, 7, 22, 23, 20, 11, 11, 11,
	11, 18, 11, 25, 11, 11, 11, 24,
	11, 26, 11, 11, 11, 24, 11, 27,
	11, 11, 11, 24, 11, 28, 11, 11,
	11, 24, 11, 29, 11, 11, 11, 24,
	11, 30, 11, 11, 11, 24, 11, 31,
	11, 11, 11, 24, 11, 32, 11, 11,
	11, 24, 11, 26, 11, 11, 11, 24,
	11, 33, 11, 11, 11, 24, 11, 34,
	11, 11, 11, 24, 11, 35, 11, 11,
	11, 24, 11, 36, 11, 11, 11, 24,
	11, 26, 11, 11, 11, 24, 11, 37,
	11, 11, 11, 24, 11, 35, 11, 11,
	11, 24,
}

var _expression_trans_targs []byte = []byte{
	1, 5, 1, 2, 3, 4, 5, 6,
	1, 3, 7, 8, 9, 11, 18, 22,
	23, 1, 1, 1, 1, 0, 1, 1,
	1, 10, 8, 12, 13, 14, 15, 16,
	17, 19, 20, 21, 8, 24,
}

var _expression_trans_actions []byte = []byte{
	23, 27, 13, 0, 42, 5, 30, 0,
	7, 30, 0, 39, 0, 0, 0, 0,
	0, 19, 25, 11, 21, 0, 15, 9,
	17, 0, 36, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 33, 0,
}

var _expression_to_state_actions []byte = []byte{
	0, 1, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0,
}

var _expression_from_state_actions []byte = []byte{
	0, 3, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0,
}

var _expression_eof_trans []byte = []byte{
	1, 0, 18, 19, 21, 19, 23, 21,
	19, 25, 25, 25, 25, 25, 25, 25,
	25, 25, 25, 25, 25, 25, 25, 25,
	25,
}

const expression_start int = 1
const expression_first_final int = 1
const expression_error int = -1

const expression_en_main int = 1

//line scanner.rl:13
type lexer struct {
	data        []byte
	p, pe, cs   int
	ts, te, act int
	val         func(Context) interface{}
}

func (l *lexer) token() string {
	return string(l.data[l.ts:l.te])
}

func newLexer(data []byte) *lexer {
	lex := &lexer{
		data: data,
		pe:   len(data),
	}

//line scanner.go:159
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

//line scanner.go:176
	{
		var _klen int
		var _trans int
		var _acts int
		var _nacts uint
		var _keys int
		if (lex.p) == (lex.pe) {
			goto _test_eof
		}
	_resume:
		_acts = int(_expression_from_state_actions[lex.cs])
		_nacts = uint(_expression_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _expression_actions[_acts-1] {
			case 1:
//line NONE:1
				lex.ts = (lex.p)

//line scanner.go:196
			}
		}

		_keys = int(_expression_key_offsets[lex.cs])
		_trans = int(_expression_index_offsets[lex.cs])

		_klen = int(_expression_single_lengths[lex.cs])
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
				case lex.data[(lex.p)] < _expression_trans_keys[_mid]:
					_upper = _mid - 1
				case lex.data[(lex.p)] > _expression_trans_keys[_mid]:
					_lower = _mid + 1
				default:
					_trans += int(_mid - int(_keys))
					goto _match
				}
			}
			_keys += _klen
			_trans += _klen
		}

		_klen = int(_expression_range_lengths[lex.cs])
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
				case lex.data[(lex.p)] < _expression_trans_keys[_mid]:
					_upper = _mid - 2
				case lex.data[(lex.p)] > _expression_trans_keys[_mid+1]:
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
		_nacts = uint(_expression_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _expression_actions[_acts-1] {
			case 2:
//line NONE:1
				lex.te = (lex.p) + 1

			case 3:
//line scanner.rl:59
				lex.act = 2
			case 4:
//line scanner.rl:76
				lex.act = 3
			case 5:
//line scanner.rl:40
				lex.act = 4
			case 6:
//line scanner.rl:68
				lex.act = 7
			case 7:
//line scanner.rl:45
				lex.act = 8
			case 8:
//line scanner.rl:83
				lex.act = 10
			case 9:
//line scanner.rl:76
				lex.te = (lex.p) + 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 10:
//line scanner.rl:78
				lex.te = (lex.p) + 1
				{
					tok = EQ
					(lex.p)++
					goto _out
				}
			case 11:
//line scanner.rl:68
				lex.te = (lex.p) + 1
				{
					tok = RELATION
					out.name = lex.token()
					(lex.p)++
					goto _out
				}
			case 12:
//line scanner.rl:83
				lex.te = (lex.p) + 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 13:
//line scanner.rl:50
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = LITERAL
					n, err := strconv.ParseInt(lex.token(), 10, 64)
					if err != nil {
						panic(err)
					}
					out.val = int(n)
					(lex.p)++
					goto _out

				}
			case 14:
//line scanner.rl:45
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = lex.token()
					(lex.p)++
					goto _out

				}
			case 15:
//line scanner.rl:82
				lex.te = (lex.p)
				(lex.p)--

			case 16:
//line scanner.rl:83
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 17:
//line scanner.rl:83
				(lex.p) = (lex.te) - 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 18:
//line NONE:1
				switch lex.act {
				case 2:
					{
						(lex.p) = (lex.te) - 1

						tok = LITERAL
						n, err := strconv.ParseFloat(lex.token(), 64)
						if err != nil {
							panic(err)
						}
						out.val = n
						(lex.p)++
						goto _out

					}
				case 3:
					{
						(lex.p) = (lex.te) - 1
						tok = int(lex.data[lex.ts])
						(lex.p)++
						goto _out
					}
				case 4:
					{
						(lex.p) = (lex.te) - 1

						tok = LITERAL
						out.val = lex.token() == "true"
						(lex.p)++
						goto _out

					}
				case 7:
					{
						(lex.p) = (lex.te) - 1
						tok = RELATION
						out.name = lex.token()
						(lex.p)++
						goto _out
					}
				case 8:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						out.name = lex.token()
						(lex.p)++
						goto _out

					}
				case 10:
					{
						(lex.p) = (lex.te) - 1
						tok = int(lex.data[lex.ts])
						(lex.p)++
						goto _out
					}
				}

//line scanner.go:393
			}
		}

	_again:
		_acts = int(_expression_to_state_actions[lex.cs])
		_nacts = uint(_expression_actions[_acts])
		_acts++
		for ; _nacts > 0; _nacts-- {
			_acts++
			switch _expression_actions[_acts-1] {
			case 0:
//line NONE:1
				lex.ts = 0

//line scanner.go:407
			}
		}

		(lex.p)++
		if (lex.p) != (lex.pe) {
			goto _resume
		}
	_test_eof:
		{
		}
		if (lex.p) == eof {
			if _expression_eof_trans[lex.cs] > 0 {
				_trans = int(_expression_eof_trans[lex.cs] - 1)
				goto _eof_trans
			}
		}

	_out:
		{
		}
	}

//line scanner.rl:87
	return tok
}

func (lex *lexer) Error(e string) {
	fmt.Println("error:", e)
}
