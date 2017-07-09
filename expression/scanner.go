//line scanner.rl:1
package expression

import "strconv"

//line scanner.go:9
var _expression_actions []byte = []byte{
	0, 1, 0, 1, 1, 1, 2, 1, 13,
	1, 14, 1, 15, 1, 16, 1, 17,
	1, 18, 1, 19, 1, 20, 1, 21,
	1, 22, 1, 23, 1, 24, 1, 25,
	1, 26, 1, 27, 1, 28, 1, 29,
	2, 2, 3, 2, 2, 4, 2, 2,
	5, 2, 2, 6, 2, 2, 7, 2,
	2, 8, 2, 2, 9, 2, 2, 10,
	2, 2, 11, 2, 2, 12,
}

var _expression_key_offsets []int16 = []int16{
	0, 1, 2, 3, 4, 5, 6, 7,
	8, 9, 10, 12, 38, 41, 42, 43,
	45, 46, 49, 51, 54, 62, 70, 71,
	72, 73, 83, 84, 95, 106, 117, 128,
	139, 150, 161, 172, 183, 195, 206, 217,
	228, 239, 250, 261, 272, 283, 294,
}

var _expression_trans_keys []byte = []byte{
	34, 115, 115, 105, 103, 110, 111, 111,
	112, 39, 48, 57, 32, 33, 34, 37,
	39, 45, 46, 60, 61, 62, 95, 97,
	99, 102, 105, 110, 111, 116, 9, 13,
	48, 57, 65, 90, 98, 122, 32, 9,
	13, 61, 34, 97, 108, 39, 46, 48,
	57, 48, 57, 46, 48, 57, 45, 95,
	48, 57, 65, 90, 97, 122, 45, 95,
	48, 57, 65, 90, 97, 122, 61, 61,
	61, 45, 58, 63, 95, 48, 57, 65,
	90, 97, 122, 58, 45, 58, 63, 95,
	110, 48, 57, 65, 90, 97, 122, 45,
	58, 63, 95, 100, 48, 57, 65, 90,
	97, 122, 45, 58, 63, 95, 111, 48,
	57, 65, 90, 97, 122, 45, 58, 63,
	95, 110, 48, 57, 65, 90, 97, 122,
	45, 58, 63, 95, 116, 48, 57, 65,
	90, 97, 122, 45, 58, 63, 95, 97,
	48, 57, 65, 90, 98, 122, 45, 58,
	63, 95, 105, 48, 57, 65, 90, 97,
	122, 45, 58, 63, 95, 110, 48, 57,
	65, 90, 97, 122, 45, 58, 63, 95,
	115, 48, 57, 65, 90, 97, 122, 45,
	58, 63, 95, 97, 111, 48, 57, 65,
	90, 98, 122, 45, 58, 63, 95, 108,
	48, 57, 65, 90, 97, 122, 45, 58,
	63, 95, 115, 48, 57, 65, 90, 97,
	122, 45, 58, 63, 95, 101, 48, 57,
	65, 90, 97, 122, 45, 58, 63, 95,
	114, 48, 57, 65, 90, 97, 122, 45,
	58, 63, 95, 110, 48, 57, 65, 90,
	97, 122, 45, 58, 63, 95, 105, 48,
	57, 65, 90, 97, 122, 45, 58, 63,
	95, 108, 48, 57, 65, 90, 97, 122,
	45, 58, 63, 95, 114, 48, 57, 65,
	90, 97, 122, 45, 58, 63, 95, 114,
	48, 57, 65, 90, 97, 122, 45, 58,
	63, 95, 117, 48, 57, 65, 90, 97,
	122,
}

var _expression_single_lengths []byte = []byte{
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 0, 18, 1, 1, 1, 2,
	1, 1, 0, 1, 2, 2, 1, 1,
	1, 4, 1, 5, 5, 5, 5, 5,
	5, 5, 5, 5, 6, 5, 5, 5,
	5, 5, 5, 5, 5, 5, 5,
}

var _expression_range_lengths []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 1, 4, 1, 0, 0, 0,
	0, 1, 1, 1, 3, 3, 0, 0,
	0, 3, 0, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3,
}

var _expression_index_offsets []int16 = []int16{
	0, 2, 4, 6, 8, 10, 12, 14,
	16, 18, 20, 22, 45, 48, 50, 52,
	55, 57, 60, 62, 65, 71, 77, 79,
	81, 83, 91, 93, 102, 111, 120, 129,
	138, 147, 156, 165, 174, 184, 193, 202,
	211, 220, 229, 238, 247, 256, 265,
}

var _expression_indicies []byte = []byte{
	2, 1, 3, 0, 4, 0, 5, 0,
	6, 0, 7, 0, 8, 0, 9, 0,
	10, 0, 2, 11, 12, 0, 14, 15,
	16, 17, 18, 19, 20, 22, 23, 24,
	25, 26, 27, 28, 29, 30, 31, 32,
	14, 21, 25, 25, 13, 14, 14, 33,
	35, 34, 2, 1, 36, 37, 34, 2,
	11, 38, 21, 34, 12, 39, 12, 21,
	40, 42, 42, 43, 42, 42, 41, 42,
	42, 42, 42, 42, 44, 45, 34, 46,
	34, 47, 34, 25, 48, 49, 25, 25,
	25, 25, 41, 48, 50, 25, 48, 49,
	25, 51, 25, 25, 25, 50, 25, 48,
	49, 25, 52, 25, 25, 25, 50, 25,
	48, 49, 25, 53, 25, 25, 25, 50,
	25, 48, 49, 25, 54, 25, 25, 25,
	50, 25, 48, 49, 25, 55, 25, 25,
	25, 50, 25, 48, 49, 25, 56, 25,
	25, 25, 50, 25, 48, 49, 25, 57,
	25, 25, 25, 50, 25, 48, 49, 25,
	58, 25, 25, 25, 50, 25, 48, 49,
	25, 59, 25, 25, 25, 50, 25, 48,
	49, 25, 60, 61, 25, 25, 25, 50,
	25, 48, 49, 25, 62, 25, 25, 25,
	50, 25, 48, 49, 25, 63, 25, 25,
	25, 50, 25, 48, 49, 25, 64, 25,
	25, 25, 50, 25, 48, 49, 25, 65,
	25, 25, 25, 50, 25, 48, 49, 25,
	66, 25, 25, 25, 50, 25, 48, 49,
	25, 67, 25, 25, 25, 50, 25, 48,
	49, 25, 68, 25, 25, 25, 50, 25,
	48, 49, 25, 69, 25, 25, 25, 50,
	25, 48, 49, 25, 70, 25, 25, 25,
	50, 25, 48, 49, 25, 63, 25, 25,
	25, 50,
}

var _expression_trans_targs []byte = []byte{
	11, 0, 11, 2, 3, 4, 5, 11,
	7, 8, 11, 9, 18, 11, 12, 13,
	14, 15, 16, 17, 20, 19, 22, 23,
	24, 25, 27, 29, 36, 41, 42, 44,
	45, 11, 11, 11, 1, 6, 10, 11,
	11, 11, 21, 20, 11, 11, 11, 11,
	11, 26, 11, 28, 25, 30, 31, 32,
	33, 34, 35, 25, 37, 40, 38, 39,
	25, 25, 25, 43, 25, 25, 46,
}

var _expression_trans_actions []byte = []byte{
	37, 0, 11, 0, 0, 0, 0, 7,
	0, 0, 9, 0, 0, 23, 0, 0,
	5, 5, 5, 5, 68, 0, 0, 0,
	0, 65, 0, 0, 0, 0, 0, 0,
	0, 33, 35, 15, 0, 0, 0, 27,
	25, 39, 0, 41, 31, 19, 13, 17,
	21, 0, 29, 0, 50, 0, 0, 0,
	0, 0, 0, 56, 0, 0, 0, 0,
	44, 59, 62, 0, 47, 53, 0,
}

var _expression_to_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 1, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0,
}

var _expression_from_state_actions []byte = []byte{
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 3, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0,
}

var _expression_eof_trans []int16 = []int16{
	1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 0, 34, 35, 35, 35,
	35, 35, 40, 41, 42, 45, 35, 35,
	35, 42, 51, 51, 51, 51, 51, 51,
	51, 51, 51, 51, 51, 51, 51, 51,
	51, 51, 51, 51, 51, 51, 51,
}

const expression_start int = 11
const expression_first_final int = 11
const expression_error int = -1

const expression_en_main int = 11

//line scanner.rl:11

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

//line scanner.go:216
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

//line scanner.go:233
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

//line scanner.go:253
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
//line scanner.rl:57
				lex.act = 4
			case 4:
//line scanner.rl:38
				lex.act = 6
			case 5:
//line scanner.rl:88
				lex.act = 7
			case 6:
//line scanner.rl:93
				lex.act = 12
			case 7:
//line scanner.rl:94
				lex.act = 13
			case 8:
//line scanner.rl:95
				lex.act = 14
			case 9:
//line scanner.rl:96
				lex.act = 15
			case 10:
//line scanner.rl:97
				lex.act = 16
			case 11:
//line scanner.rl:43
				lex.act = 18
			case 12:
//line scanner.rl:102
				lex.act = 21
			case 13:
//line scanner.rl:82
				lex.te = (lex.p) + 1
				{
					tok = ASSIGN
					(lex.p)++
					goto _out
				}
			case 14:
//line scanner.rl:83
				lex.te = (lex.p) + 1
				{
					tok = LOOP
					(lex.p)++
					goto _out
				}
			case 15:
//line scanner.rl:66
				lex.te = (lex.p) + 1
				{
					tok = LITERAL
					// TODO unescape \x
					out.val = string(lex.data[lex.ts+1 : lex.te-1])
					(lex.p)++
					goto _out

				}
			case 16:
//line scanner.rl:89
				lex.te = (lex.p) + 1
				{
					tok = EQ
					(lex.p)++
					goto _out
				}
			case 17:
//line scanner.rl:90
				lex.te = (lex.p) + 1
				{
					tok = NEQ
					(lex.p)++
					goto _out
				}
			case 18:
//line scanner.rl:91
				lex.te = (lex.p) + 1
				{
					tok = GE
					(lex.p)++
					goto _out
				}
			case 19:
//line scanner.rl:92
				lex.te = (lex.p) + 1
				{
					tok = LE
					(lex.p)++
					goto _out
				}
			case 20:
//line scanner.rl:98
				lex.te = (lex.p) + 1
				{
					tok = KEYWORD
					out.name = string(lex.data[lex.ts : lex.te-1])
					(lex.p)++
					goto _out
				}
			case 21:
//line scanner.rl:102
				lex.te = (lex.p) + 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 22:
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
			case 23:
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
			case 24:
//line scanner.rl:43
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = IDENTIFIER
					out.name = lex.token()
					(lex.p)++
					goto _out

				}
			case 25:
//line scanner.rl:100
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = PROPERTY
					out.name = string(lex.data[lex.ts+1 : lex.te])
					(lex.p)++
					goto _out
				}
			case 26:
//line scanner.rl:101
				lex.te = (lex.p)
				(lex.p)--

			case 27:
//line scanner.rl:102
				lex.te = (lex.p)
				(lex.p)--
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 28:
//line scanner.rl:102
				(lex.p) = (lex.te) - 1
				{
					tok = int(lex.data[lex.ts])
					(lex.p)++
					goto _out
				}
			case 29:
//line NONE:1
				switch lex.act {
				case 4:
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
				case 6:
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
						tok = LITERAL
						out.val = nil
						(lex.p)++
						goto _out
					}
				case 12:
					{
						(lex.p) = (lex.te) - 1
						tok = AND
						(lex.p)++
						goto _out
					}
				case 13:
					{
						(lex.p) = (lex.te) - 1
						tok = OR
						(lex.p)++
						goto _out
					}
				case 14:
					{
						(lex.p) = (lex.te) - 1
						tok = CONTAINS
						(lex.p)++
						goto _out
					}
				case 15:
					{
						(lex.p) = (lex.te) - 1
						tok = FOR
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
				case 21:
					{
						(lex.p) = (lex.te) - 1
						tok = int(lex.data[lex.ts])
						(lex.p)++
						goto _out
					}
				}

//line scanner.go:528
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

//line scanner.go:542
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

//line scanner.rl:106

	return tok
}

func (lex *lexer) Error(e string) {
	// fmt.Println("scan error:", e)
}
