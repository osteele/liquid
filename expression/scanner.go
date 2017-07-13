//line scanner.rl:1
package expression

import "strconv"

//line scanner.go:9
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 10,
	1, 11, 1, 12, 1, 13, 1, 14,
	1, 15, 1, 16, 1, 17, 1, 18,
	1, 19, 1, 20, 1, 21, 1, 22,
	1, 23, 1, 24, 1, 25, 1, 26,
	1, 27, 1, 28, 2, 2, 3, 2,
	2, 4, 2, 2, 5, 2, 2, 6,
	2, 2, 7, 2, 2, 8, 2, 2,
	9,
}

var _expression_key_offsets []int16 = []int16{
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 11, 12, 14, 15, 16,
	17, 18, 19, 20, 47, 50, 51, 52,
	54, 55, 58, 60, 63, 71, 80, 89,
	90, 91, 92, 102, 103, 114, 125, 136,
	147, 158, 169, 180, 191, 202, 213, 224,
	235, 246, 257, 268, 279, 290, 301, 312,
}

var _expression_trans_keys []byte = []byte{
	34, 115, 115, 105, 103, 110, 32, 111,
	111, 112, 32, 39, 48, 57, 99, 121,
	99, 108, 101, 32, 32, 33, 34, 37,
	39, 45, 46, 60, 61, 62, 95, 97,
	99, 102, 105, 110, 111, 116, 123, 9,
	13, 48, 57, 65, 90, 98, 122, 32,
	9, 13, 61, 34, 97, 108, 39, 46,
	48, 57, 48, 57, 46, 48, 57, 45,
	95, 48, 57, 65, 90, 97, 122, 45,
	63, 95, 48, 57, 65, 90, 97, 122,
	45, 63, 95, 48, 57, 65, 90, 97,
	122, 61, 61, 61, 45, 58, 63, 95,
	48, 57, 65, 90, 97, 122, 58, 45,
	58, 63, 95, 110, 48, 57, 65, 90,
	97, 122, 45, 58, 63, 95, 100, 48,
	57, 65, 90, 97, 122, 45, 58, 63,
	95, 111, 48, 57, 65, 90, 97, 122,
	45, 58, 63, 95, 110, 48, 57, 65,
	90, 97, 122, 45, 58, 63, 95, 116,
	48, 57, 65, 90, 97, 122, 45, 58,
	63, 95, 97, 48, 57, 65, 90, 98,
	122, 45, 58, 63, 95, 105, 48, 57,
	65, 90, 97, 122, 45, 58, 63, 95,
	110, 48, 57, 65, 90, 97, 122, 45,
	58, 63, 95, 115, 48, 57, 65, 90,
	97, 122, 45, 58, 63, 95, 97, 48,
	57, 65, 90, 98, 122, 45, 58, 63,
	95, 108, 48, 57, 65, 90, 97, 122,
	45, 58, 63, 95, 115, 48, 57, 65,
	90, 97, 122, 45, 58, 63, 95, 101,
	48, 57, 65, 90, 97, 122, 45, 58,
	63, 95, 110, 48, 57, 65, 90, 97,
	122, 45, 58, 63, 95, 105, 48, 57,
	65, 90, 97, 122, 45, 58, 63, 95,
	108, 48, 57, 65, 90, 97, 122, 45,
	58, 63, 95, 114, 48, 57, 65, 90,
	97, 122, 45, 58, 63, 95, 114, 48,
	57, 65, 90, 97, 122, 45, 58, 63,
	95, 117, 48, 57, 65, 90, 97, 122,
	37,
}

var _expression_single_lengths []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 0, 1, 1, 1,
	1, 1, 1, 19, 1, 1, 1, 2,
	1, 1, 0, 1, 2, 3, 3, 1,
	1, 1, 4, 1, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5, 1,
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 1, 0, 0, 0,
	0, 0, 0, 4, 1, 0, 0, 0,
	0, 1, 1, 1, 3, 3, 3, 0,
	0, 0, 3, 0, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 0,
}

var _expression_index_offsets []int16 = []int16{
	0, 2, 4, 6, 8, 10, 12, 14,
	16, 18, 20, 22, 24, 26, 28, 30,
	32, 34, 36, 38, 62, 65, 67, 69,
	72, 74, 77, 79, 82, 88, 95, 102,
	104, 106, 108, 116, 118, 127, 136, 145,
	154, 163, 172, 181, 190, 199, 208, 217,
	226, 235, 244, 253, 262, 271, 280, 289,
}

var _expression_indicies []byte = []byte{
	2, 1, 3, 0, 4, 0, 5, 0,
	6, 0, 7, 0, 8, 0, 9, 0,
	10, 0, 11, 0, 12, 0, 2, 13,
	14, 0, 15, 0, 16, 0, 17, 0,
	18, 0, 19, 0, 20, 0, 22, 23,
	24, 25, 26, 27, 28, 30, 31, 32,
	33, 34, 35, 36, 37, 38, 39, 40,
	41, 22, 29, 33, 33, 21, 22, 22,
	42, 44, 43, 2, 1, 45, 46, 43,
	2, 13, 47, 29, 43, 14, 48, 14,
	29, 49, 50, 50, 51, 50, 50, 43,
	50, 53, 50, 50, 50, 50, 52, 50,
	53, 50, 51, 50, 50, 48, 54, 43,
	55, 43, 56, 43, 33, 58, 59, 33,
	33, 33, 33, 57, 58, 60, 33, 58,
	59, 33, 61, 33, 33, 33, 60, 33,
	58, 59, 33, 62, 33, 33, 33, 60,
	33, 58, 59, 33, 63, 33, 33, 33,
	60, 33, 58, 59, 33, 64, 33, 33,
	33, 60, 33, 58, 59, 33, 65, 33,
	33, 33, 60, 33, 58, 59, 33, 66,
	33, 33, 33, 60, 33, 58, 59, 33,
	67, 33, 33, 33, 60, 33, 58, 59,
	33, 68, 33, 33, 33, 60, 33, 58,
	59, 33, 69, 33, 33, 33, 60, 33,
	58, 59, 33, 70, 33, 33, 33, 60,
	33, 58, 59, 33, 71, 33, 33, 33,
	60, 33, 58, 59, 33, 72, 33, 33,
	33, 60, 33, 58, 59, 33, 73, 33,
	33, 33, 60, 33, 58, 59, 33, 74,
	33, 33, 33, 60, 33, 58, 59, 33,
	75, 33, 33, 33, 60, 33, 58, 59,
	33, 76, 33, 33, 33, 60, 33, 58,
	59, 33, 77, 33, 33, 33, 60, 33,
	58, 59, 33, 78, 33, 33, 33, 60,
	33, 58, 59, 33, 72, 33, 33, 33,
	60, 79, 43,
}

var _expression_trans_targs []byte = []byte{
	19, 0, 19, 2, 3, 4, 5, 6,
	19, 8, 9, 10, 19, 11, 26, 14,
	15, 16, 17, 18, 19, 19, 20, 21,
	22, 23, 24, 25, 28, 27, 31, 32,
	33, 34, 36, 38, 45, 49, 50, 52,
	53, 55, 19, 19, 19, 1, 7, 12,
	19, 19, 29, 30, 19, 19, 19, 19,
	19, 19, 19, 35, 19, 37, 34, 39,
	40, 41, 42, 43, 44, 34, 46, 47,
	48, 34, 34, 51, 34, 34, 54, 13,
}

var _expression_trans_actions []byte = []byte{
	41, 0, 13, 0, 0, 0, 0, 0,
	7, 0, 0, 0, 11, 0, 0, 0,
	0, 0, 0, 0, 9, 27, 0, 0,
	5, 5, 5, 5, 0, 0, 0, 0,
	0, 63, 0, 0, 0, 0, 0, 0,
	0, 5, 37, 39, 17, 0, 0, 0,
	31, 29, 0, 0, 35, 25, 21, 15,
	19, 43, 23, 0, 33, 0, 51, 0,
	0, 0, 0, 0, 0, 57, 0, 0,
	0, 45, 60, 0, 48, 54, 0, 0,
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 1, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 3, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
}

var _expression_eof_trans []int16 = []int16{
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 0, 43, 44, 44, 44,
	44, 44, 49, 50, 44, 53, 49, 44,
	44, 44, 58, 61, 61, 61, 61, 61,
	61, 61, 61, 61, 61, 61, 61, 61,
	61, 61, 61, 61, 61, 61, 61, 44,
}

const expression_start int = 19
const expression_first_final int = 19
const expression_error int = -1

const expression_en_main int = 19

//line scanner.rl:11

type lexer struct {
	parseValue
	data        []byte
	p, pe, cs   int
	ts, te, act int
}

func (l *lexer) token() string {
	return string(l.data[l.ts:l.te])
}

func newLexer(data []byte) *lexer {
	lex := &lexer{
		data: data,
		pe:   len(data),
	}

//line scanner.go:228
	{
		lex.cs = expression_start
		lex.ts = 0
		lex.te = 0
		lex.act = 0
	}

//line scanner.rl:30
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

//line scanner.go:245
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

//line scanner.go:265
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
//line scanner.rl:38
				lex.act = 7
			case 4:
//line scanner.rl:94
				lex.act = 8
			case 5:
//line scanner.rl:101
				lex.act = 13
			case 6:
//line scanner.rl:102
				lex.act = 14
			case 7:
//line scanner.rl:103
				lex.act = 15
			case 8:
//line scanner.rl:106
				lex.act = 16
			case 9:
//line scanner.rl:43
				lex.act = 18
			case 10:
//line scanner.rl:83
				lex.te = (lex.p) + 1
				{
					tok = ASSIGN
					(lex.p)++
					goto _out
				}
			case 11:
//line scanner.rl:84
				lex.te = (lex.p) + 1
				{
					tok = ARGLIST
					(lex.p)++
					goto _out
				}
			case 12:
//line scanner.rl:85
				lex.te = (lex.p) + 1
				{
					tok = LOOP
					(lex.p)++
					goto _out
				}
			case 13:
//line scanner.rl:66
				lex.te = (lex.p) + 1
				{
					tok = LITERAL
					// TODO unescape \x
					out.val = string(lex.data[lex.ts+1 : lex.te-1])
					(lex.p)++
					goto _out

				}
			case 14:
//line scanner.rl:97
				lex.te = (lex.p) + 1
				{
					tok = EQ
					(lex.p)++
					goto _out
				}
			case 15:
//line scanner.rl:98
				lex.te = (lex.p) + 1
				{
					tok = NEQ
					(lex.p)++
					goto _out
				}
			case 16:
//line scanner.rl:99
				lex.te = (lex.p) + 1
				{
					tok = GE
					(lex.p)++
					goto _out
				}
			case 17:
//line scanner.rl:100
				lex.te = (lex.p) + 1
				{
					tok = LE
					(lex.p)++
					goto _out
				}
			case 18:
//line scanner.rl:108
				lex.te = (lex.p) + 1
				{
					tok = KEYWORD
					out.name = string(lex.data[lex.ts : lex.te-1])
					(lex.p)++
					goto _out
				}
			case 19:
//line scanner.rl:110
				lex.te = (lex.p) + 1
				{
					tok = PROPERTY
					out.name = string(lex.data[lex.ts+1 : lex.te])
					(lex.p)++
					goto _out
				}
			case 20:
//line scanner.rl:113
				lex.te = (lex.p) + 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 21:
//line scanner.rl:48
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
			case 22:
//line scanner.rl:57
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = LITERAL
					n, err := strconv.ParseFloat(lex.token(), 64)
					if err != nil {
						panic(err)
					}
					out.val = n
					(lex.p)++
					goto _out

				}
			case 23:
//line scanner.rl:43
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = lex.token()
					(lex.p)++
					goto _out

				}
			case 24:
//line scanner.rl:110
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = PROPERTY
					out.name = string(lex.data[lex.ts+1 : lex.te])
					(lex.p)++
					goto _out
				}
			case 25:
//line scanner.rl:112
				lex.te = (lex.p)
				(lex.p)--

			case 26:
//line scanner.rl:113
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 27:
//line scanner.rl:113
				(lex.p) = (lex.te) - 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 28:
//line NONE:1
				switch lex.act {
				case 7:
					{
						(lex.p) = (lex.te) - 1

						tok = LITERAL
						out.val = lex.token() == "true"
						(lex.p)++
						goto _out

					}
				case 8:
					{
						(lex.p) = (lex.te) - 1
						tok = LITERAL
						out.val = nil
						(lex.p)++
						goto _out
					}
				case 13:
					{
						(lex.p) = (lex.te) - 1
						tok = AND
						(lex.p)++
						goto _out
					}
				case 14:
					{
						(lex.p) = (lex.te) - 1
						tok = OR
						(lex.p)++
						goto _out
					}
				case 15:
					{
						(lex.p) = (lex.te) - 1
						tok = CONTAINS
						(lex.p)++
						goto _out
					}
				case 16:
					{
						(lex.p) = (lex.te) - 1
						tok = IN
						(lex.p)++
						goto _out
					}
				case 18:
					{
						(lex.p) = (lex.te) - 1

						tok = IDENTIFIER
						out.name = lex.token()
						(lex.p)++
						goto _out

					}
				}

//line scanner.go:521
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

//line scanner.go:535
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

//line scanner.rl:117

	return tok
}

func (lex *lexer) Error(e string) {
	// fmt.Println("scan error:", e)
}
