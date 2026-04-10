
//line /src/scanner.rl:1
package expressions

import (
	"strconv"
	"strings"
	"unicode"
)


//line /src/scanner.go:11
const expression_start int = 31
const expression_first_final int = 31
const expression_error int = -1

const expression_en_main int = 31


//line /src/scanner.rl:27


type lexer struct {
	parseValue
    data []byte
    p, pe, cs int
    ts, te, act int
}

func (l* lexer) token() string {
	return string(l.data[l.ts:l.te])
}

func newLexer(data []byte) *lexer {
	lex := &lexer{
			data: data,
			pe: len(data),
	}
	
//line /src/scanner.go:37
	{
	 lex.cs = expression_start
	 lex.ts = 0
	 lex.te = 0
	 lex.act = 0
	}

//line /src/scanner.rl:46
	return lex
}

func (lex *lexer) Lex(out *yySymType) int {
	eof := lex.pe
	tok := 0

	
//line /src/scanner.go:52
	{
	if ( lex.p) == ( lex.pe) {
		goto _test_eof
	}
	switch  lex.cs {
	case 31:
		goto st_case_31
	case 32:
		goto st_case_32
	case 33:
		goto st_case_33
	case 34:
		goto st_case_34
	case 0:
		goto st_case_0
	case 1:
		goto st_case_1
	case 35:
		goto st_case_35
	case 2:
		goto st_case_2
	case 3:
		goto st_case_3
	case 4:
		goto st_case_4
	case 5:
		goto st_case_5
	case 6:
		goto st_case_6
	case 7:
		goto st_case_7
	case 8:
		goto st_case_8
	case 9:
		goto st_case_9
	case 10:
		goto st_case_10
	case 11:
		goto st_case_11
	case 36:
		goto st_case_36
	case 12:
		goto st_case_12
	case 13:
		goto st_case_13
	case 37:
		goto st_case_37
	case 38:
		goto st_case_38
	case 14:
		goto st_case_14
	case 39:
		goto st_case_39
	case 40:
		goto st_case_40
	case 41:
		goto st_case_41
	case 15:
		goto st_case_15
	case 16:
		goto st_case_16
	case 17:
		goto st_case_17
	case 42:
		goto st_case_42
	case 43:
		goto st_case_43
	case 44:
		goto st_case_44
	case 45:
		goto st_case_45
	case 46:
		goto st_case_46
	case 18:
		goto st_case_18
	case 19:
		goto st_case_19
	case 20:
		goto st_case_20
	case 47:
		goto st_case_47
	case 48:
		goto st_case_48
	case 49:
		goto st_case_49
	case 50:
		goto st_case_50
	case 51:
		goto st_case_51
	case 52:
		goto st_case_52
	case 53:
		goto st_case_53
	case 54:
		goto st_case_54
	case 55:
		goto st_case_55
	case 56:
		goto st_case_56
	case 57:
		goto st_case_57
	case 58:
		goto st_case_58
	case 59:
		goto st_case_59
	case 60:
		goto st_case_60
	case 61:
		goto st_case_61
	case 62:
		goto st_case_62
	case 63:
		goto st_case_63
	case 64:
		goto st_case_64
	case 65:
		goto st_case_65
	case 66:
		goto st_case_66
	case 67:
		goto st_case_67
	case 68:
		goto st_case_68
	case 69:
		goto st_case_69
	case 70:
		goto st_case_70
	case 71:
		goto st_case_71
	case 72:
		goto st_case_72
	case 73:
		goto st_case_73
	case 74:
		goto st_case_74
	case 75:
		goto st_case_75
	case 21:
		goto st_case_21
	case 22:
		goto st_case_22
	case 23:
		goto st_case_23
	case 24:
		goto st_case_24
	case 25:
		goto st_case_25
	case 26:
		goto st_case_26
	case 27:
		goto st_case_27
	case 28:
		goto st_case_28
	case 29:
		goto st_case_29
	case 30:
		goto st_case_30
	case 76:
		goto st_case_76
	case 77:
		goto st_case_77
	case 78:
		goto st_case_78
	}
	goto st_out
tr0:
//line /src/scanner.rl:142
( lex.p) = ( lex.te) - 1
{ tok = int(lex.data[lex.ts]); {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr2:
//line /src/scanner.rl:88
 lex.te = ( lex.p)+1
{
			tok = LITERAL
			out.val = unescapeString(lex.data[lex.ts+1:lex.te-1])
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	goto st31
tr9:
//line /src/scanner.rl:106
 lex.te = ( lex.p)+1
{ tok = ASSIGN; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr13:
//line /src/scanner.rl:108
 lex.te = ( lex.p)+1
{ tok = LOOP; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr16:
//line /src/scanner.rl:70
( lex.p) = ( lex.te) - 1
{
			tok = LITERAL
			n, err := strconv.ParseInt(lex.token(), 10, 64)
			if err != nil {
				panic(err)
			}
			out.val = int(n)
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	goto st31
tr18:
//line NONE:1
	switch  lex.act {
	case 8:
	{( lex.p) = ( lex.te) - 1

			tok = LITERAL
			out.val = lex.token() == "true"
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	case 9:
	{( lex.p) = ( lex.te) - 1
 tok = LITERAL; out.val = nil; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 10:
	{( lex.p) = ( lex.te) - 1
 tok = EMPTY; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 11:
	{( lex.p) = ( lex.te) - 1
 tok = BLANK; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 17:
	{( lex.p) = ( lex.te) - 1
 tok = AND; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 18:
	{( lex.p) = ( lex.te) - 1
 tok = OR; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 19:
	{( lex.p) = ( lex.te) - 1
 tok = NOT; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 20:
	{( lex.p) = ( lex.te) - 1
 tok = CONTAINS; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 21:
	{( lex.p) = ( lex.te) - 1
 tok = IN; {( lex.p)++;  lex.cs = 31; goto _out } }
	case 24:
	{( lex.p) = ( lex.te) - 1

			tok = IDENTIFIER
			t := lex.token()

			if !isValidUnicodeIdentifier(t) {
				panic("syntax error in identifier " + t)
			}

			out.name = t
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	case 25:
	{( lex.p) = ( lex.te) - 1
 tok = PROPERTY; out.name = string(lex.data[lex.ts+1:lex.te]); {( lex.p)++;  lex.cs = 31; goto _out } }
	case 27:
	{( lex.p) = ( lex.te) - 1
 tok = int(lex.data[lex.ts]); {( lex.p)++;  lex.cs = 31; goto _out } }
	}
	
	goto st31
tr31:
//line /src/scanner.rl:107
 lex.te = ( lex.p)+1
{ tok = CYCLE; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr35:
//line /src/scanner.rl:109
 lex.te = ( lex.p)+1
{ tok = WHEN; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr36:
//line /src/scanner.rl:142
 lex.te = ( lex.p)+1
{ tok = int(lex.data[lex.ts]); {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr61:
//line /src/scanner.rl:141
 lex.te = ( lex.p)
( lex.p)--

	goto st31
tr62:
//line /src/scanner.rl:142
 lex.te = ( lex.p)
( lex.p)--
{ tok = int(lex.data[lex.ts]); {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr63:
//line /src/scanner.rl:124
 lex.te = ( lex.p)+1
{ tok = NEQ; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr66:
//line /src/scanner.rl:70
 lex.te = ( lex.p)
( lex.p)--
{
			tok = LITERAL
			n, err := strconv.ParseInt(lex.token(), 10, 64)
			if err != nil {
				panic(err)
			}
			out.val = int(n)
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	goto st31
tr68:
//line /src/scanner.rl:79
 lex.te = ( lex.p)
( lex.p)--
{
			tok = LITERAL
			n, err := strconv.ParseFloat(lex.token(), 64)
			if err != nil {
				panic(err)
			}
			out.val = n
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	goto st31
tr69:
//line /src/scanner.rl:135
 lex.te = ( lex.p)+1
{ tok = DOTDOT; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr71:
//line /src/scanner.rl:139
 lex.te = ( lex.p)
( lex.p)--
{ tok = PROPERTY; out.name = string(lex.data[lex.ts+1:lex.te]); {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr72:
//line /src/scanner.rl:139
 lex.te = ( lex.p)+1
{ tok = PROPERTY; out.name = string(lex.data[lex.ts+1:lex.te]); {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr73:
//line /src/scanner.rl:127
 lex.te = ( lex.p)+1
{ tok = LE; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr74:
//line /src/scanner.rl:125
 lex.te = ( lex.p)+1
{ tok = NEQ; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr75:
//line /src/scanner.rl:123
 lex.te = ( lex.p)+1
{ tok = EQ; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr76:
//line /src/scanner.rl:126
 lex.te = ( lex.p)+1
{ tok = GE; {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr77:
//line /src/scanner.rl:137
 lex.te = ( lex.p)+1
{ tok = KEYWORD; out.name = string(lex.data[lex.ts:lex.te-1]); {( lex.p)++;  lex.cs = 31; goto _out } }
	goto st31
tr80:
//line /src/scanner.rl:59
 lex.te = ( lex.p)
( lex.p)--
{
			tok = IDENTIFIER
			t := lex.token()

			if !isValidUnicodeIdentifier(t) {
				panic("syntax error in identifier " + t)
			}

			out.name = t
			{( lex.p)++;  lex.cs = 31; goto _out }
		}
	goto st31
	st31:
//line NONE:1
 lex.ts = 0

		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof31
		}
	st_case_31:
//line NONE:1
 lex.ts = ( lex.p)

//line /src/scanner.go:439
		switch  lex.data[( lex.p)] {
		case 32:
			goto st32
		case 33:
			goto st33
		case 34:
			goto tr39
		case 37:
			goto tr40
		case 39:
			goto tr41
		case 45:
			goto st37
		case 46:
			goto tr43
		case 60:
			goto st42
		case 61:
			goto st43
		case 62:
			goto st44
		case 95:
			goto tr22
		case 97:
			goto tr48
		case 98:
			goto tr49
		case 99:
			goto tr50
		case 101:
			goto tr51
		case 102:
			goto tr52
		case 105:
			goto tr53
		case 110:
			goto tr54
		case 111:
			goto tr55
		case 116:
			goto tr56
		case 123:
			goto tr57
		}
		switch {
		case  lex.data[( lex.p)] < 100:
			switch {
			case  lex.data[( lex.p)] < 48:
				if 9 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 13 {
					goto st32
				}
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			default:
				goto tr44
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st76
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto tr60
				}
			default:
				goto tr59
			}
		default:
			goto tr22
		}
		goto tr36
	st32:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof32
		}
	st_case_32:
		if  lex.data[( lex.p)] == 32 {
			goto st32
		}
		if 9 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 13 {
			goto st32
		}
		goto tr61
	st33:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof33
		}
	st_case_33:
		if  lex.data[( lex.p)] == 61 {
			goto tr63
		}
		goto tr62
tr39:
//line NONE:1
 lex.te = ( lex.p)+1

	goto st34
	st34:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof34
		}
	st_case_34:
//line /src/scanner.go:546
		switch  lex.data[( lex.p)] {
		case 34:
			goto tr2
		case 92:
			goto st1
		}
		goto st0
	st0:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof0
		}
	st_case_0:
		switch  lex.data[( lex.p)] {
		case 34:
			goto tr2
		case 92:
			goto st1
		}
		goto st0
	st1:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof1
		}
	st_case_1:
		goto st0
tr40:
//line NONE:1
 lex.te = ( lex.p)+1

	goto st35
	st35:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof35
		}
	st_case_35:
//line /src/scanner.go:582
		switch  lex.data[( lex.p)] {
		case 97:
			goto st2
		case 108:
			goto st8
		}
		goto tr62
	st2:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof2
		}
	st_case_2:
		if  lex.data[( lex.p)] == 115 {
			goto st3
		}
		goto tr0
	st3:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof3
		}
	st_case_3:
		if  lex.data[( lex.p)] == 115 {
			goto st4
		}
		goto tr0
	st4:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof4
		}
	st_case_4:
		if  lex.data[( lex.p)] == 105 {
			goto st5
		}
		goto tr0
	st5:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof5
		}
	st_case_5:
		if  lex.data[( lex.p)] == 103 {
			goto st6
		}
		goto tr0
	st6:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof6
		}
	st_case_6:
		if  lex.data[( lex.p)] == 110 {
			goto st7
		}
		goto tr0
	st7:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof7
		}
	st_case_7:
		if  lex.data[( lex.p)] == 32 {
			goto tr9
		}
		goto tr0
	st8:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof8
		}
	st_case_8:
		if  lex.data[( lex.p)] == 111 {
			goto st9
		}
		goto tr0
	st9:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof9
		}
	st_case_9:
		if  lex.data[( lex.p)] == 111 {
			goto st10
		}
		goto tr0
	st10:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof10
		}
	st_case_10:
		if  lex.data[( lex.p)] == 112 {
			goto st11
		}
		goto tr0
	st11:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof11
		}
	st_case_11:
		if  lex.data[( lex.p)] == 32 {
			goto tr13
		}
		goto tr0
tr41:
//line NONE:1
 lex.te = ( lex.p)+1

	goto st36
	st36:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof36
		}
	st_case_36:
//line /src/scanner.go:690
		switch  lex.data[( lex.p)] {
		case 39:
			goto tr2
		case 92:
			goto st13
		}
		goto st12
	st12:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof12
		}
	st_case_12:
		switch  lex.data[( lex.p)] {
		case 39:
			goto tr2
		case 92:
			goto st13
		}
		goto st12
	st13:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof13
		}
	st_case_13:
		goto st12
	st37:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof37
		}
	st_case_37:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto tr44
		}
		goto tr62
tr44:
//line NONE:1
 lex.te = ( lex.p)+1

	goto st38
	st38:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof38
		}
	st_case_38:
//line /src/scanner.go:735
		if  lex.data[( lex.p)] == 46 {
			goto st14
		}
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto tr44
		}
		goto tr66
	st14:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof14
		}
	st_case_14:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st39
		}
		goto tr16
	st39:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof39
		}
	st_case_39:
		if 48 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 57 {
			goto st39
		}
		goto tr68
tr43:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:142
 lex.act = 27;
	goto st40
	st40:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof40
		}
	st_case_40:
//line /src/scanner.go:773
		switch  lex.data[( lex.p)] {
		case 46:
			goto tr69
		case 95:
			goto tr19
		}
		switch {
		case  lex.data[( lex.p)] < 194:
			switch {
			case  lex.data[( lex.p)] > 90:
				if 97 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 122 {
					goto tr19
				}
			case  lex.data[( lex.p)] >= 65:
				goto tr19
			}
		case  lex.data[( lex.p)] > 223:
			switch {
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st17
				}
			case  lex.data[( lex.p)] >= 224:
				goto st16
			}
		default:
			goto st15
		}
		goto tr62
tr19:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:139
 lex.act = 25;
	goto st41
	st41:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof41
		}
	st_case_41:
//line /src/scanner.go:815
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr19
		case 63:
			goto tr72
		case 95:
			goto tr19
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr19
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr19
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st15
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st17
				}
			default:
				goto st16
			}
		default:
			goto tr19
		}
		goto tr71
	st15:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof15
		}
	st_case_15:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto tr19
		}
		goto tr18
	st16:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof16
		}
	st_case_16:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto st15
		}
		goto tr18
	st17:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof17
		}
	st_case_17:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto st16
		}
		goto tr18
	st42:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof42
		}
	st_case_42:
		switch  lex.data[( lex.p)] {
		case 61:
			goto tr73
		case 62:
			goto tr74
		}
		goto tr62
	st43:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof43
		}
	st_case_43:
		if  lex.data[( lex.p)] == 61 {
			goto tr75
		}
		goto tr62
	st44:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof44
		}
	st_case_44:
		if  lex.data[( lex.p)] == 61 {
			goto tr76
		}
		goto tr62
tr22:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st45
tr82:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:128
 lex.act = 17;
	goto st45
tr86:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:120
 lex.act = 11;
	goto st45
tr93:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:131
 lex.act = 20;
	goto st45
tr97:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:119
 lex.act = 10;
	goto st45
tr101:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:54
 lex.act = 8;
	goto st45
tr102:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:134
 lex.act = 21;
	goto st45
tr105:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:118
 lex.act = 9;
	goto st45
tr106:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:130
 lex.act = 19;
	goto st45
tr107:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:129
 lex.act = 18;
	goto st45
	st45:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof45
		}
	st_case_45:
//line /src/scanner.go:983
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr18
	st46:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof46
		}
	st_case_46:
		if  lex.data[( lex.p)] == 58 {
			goto tr77
		}
		goto tr80
	st18:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof18
		}
	st_case_18:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto tr22
		}
		goto tr18
	st19:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof19
		}
	st_case_19:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto st18
		}
		goto tr18
	st20:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof20
		}
	st_case_20:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto st19
		}
		goto tr18
tr48:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st47
	st47:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof47
		}
	st_case_47:
//line /src/scanner.go:1069
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 110:
			goto tr81
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr81:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st48
	st48:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof48
		}
	st_case_48:
//line /src/scanner.go:1121
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 100:
			goto tr82
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr49:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st49
	st49:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof49
		}
	st_case_49:
//line /src/scanner.go:1173
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 108:
			goto tr83
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr83:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st50
	st50:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof50
		}
	st_case_50:
//line /src/scanner.go:1225
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 97:
			goto tr84
		}
		switch {
		case  lex.data[( lex.p)] < 98:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr84:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st51
	st51:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof51
		}
	st_case_51:
//line /src/scanner.go:1277
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 110:
			goto tr85
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr85:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st52
	st52:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof52
		}
	st_case_52:
//line /src/scanner.go:1329
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 107:
			goto tr86
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr50:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st53
	st53:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof53
		}
	st_case_53:
//line /src/scanner.go:1381
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 111:
			goto tr87
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr87:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st54
	st54:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof54
		}
	st_case_54:
//line /src/scanner.go:1433
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 110:
			goto tr88
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr88:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st55
	st55:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof55
		}
	st_case_55:
//line /src/scanner.go:1485
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 116:
			goto tr89
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr89:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st56
	st56:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof56
		}
	st_case_56:
//line /src/scanner.go:1537
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 97:
			goto tr90
		}
		switch {
		case  lex.data[( lex.p)] < 98:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr90:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st57
	st57:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof57
		}
	st_case_57:
//line /src/scanner.go:1589
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 105:
			goto tr91
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr91:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st58
	st58:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof58
		}
	st_case_58:
//line /src/scanner.go:1641
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 110:
			goto tr92
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr92:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st59
	st59:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof59
		}
	st_case_59:
//line /src/scanner.go:1693
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 115:
			goto tr93
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr51:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st60
	st60:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof60
		}
	st_case_60:
//line /src/scanner.go:1745
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 109:
			goto tr94
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr94:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st61
	st61:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof61
		}
	st_case_61:
//line /src/scanner.go:1797
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 112:
			goto tr95
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr95:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st62
	st62:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof62
		}
	st_case_62:
//line /src/scanner.go:1849
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 116:
			goto tr96
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr96:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st63
	st63:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof63
		}
	st_case_63:
//line /src/scanner.go:1901
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 121:
			goto tr97
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr52:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st64
	st64:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof64
		}
	st_case_64:
//line /src/scanner.go:1953
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 97:
			goto tr98
		}
		switch {
		case  lex.data[( lex.p)] < 98:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr98:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st65
	st65:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof65
		}
	st_case_65:
//line /src/scanner.go:2005
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 108:
			goto tr99
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr99:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st66
	st66:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof66
		}
	st_case_66:
//line /src/scanner.go:2057
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 115:
			goto tr100
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr100:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st67
	st67:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof67
		}
	st_case_67:
//line /src/scanner.go:2109
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 101:
			goto tr101
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr53:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st68
	st68:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof68
		}
	st_case_68:
//line /src/scanner.go:2161
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 110:
			goto tr102
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr54:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st69
	st69:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof69
		}
	st_case_69:
//line /src/scanner.go:2213
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 105:
			goto tr103
		case 111:
			goto tr104
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr103:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st70
	st70:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof70
		}
	st_case_70:
//line /src/scanner.go:2267
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 108:
			goto tr105
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr104:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st71
	st71:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof71
		}
	st_case_71:
//line /src/scanner.go:2319
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 116:
			goto tr106
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr55:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st72
	st72:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof72
		}
	st_case_72:
//line /src/scanner.go:2371
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 114:
			goto tr107
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr56:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st73
	st73:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof73
		}
	st_case_73:
//line /src/scanner.go:2423
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 114:
			goto tr108
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr108:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:59
 lex.act = 24;
	goto st74
	st74:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof74
		}
	st_case_74:
//line /src/scanner.go:2475
		switch  lex.data[( lex.p)] {
		case 45:
			goto tr22
		case 58:
			goto tr77
		case 63:
			goto st46
		case 95:
			goto tr22
		case 117:
			goto tr100
		}
		switch {
		case  lex.data[( lex.p)] < 97:
			switch {
			case  lex.data[( lex.p)] > 57:
				if 65 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 90 {
					goto tr22
				}
			case  lex.data[( lex.p)] >= 48:
				goto tr22
			}
		case  lex.data[( lex.p)] > 122:
			switch {
			case  lex.data[( lex.p)] < 224:
				if 194 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 223 {
					goto st18
				}
			case  lex.data[( lex.p)] > 239:
				if 240 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 247 {
					goto st20
				}
			default:
				goto st19
			}
		default:
			goto tr22
		}
		goto tr80
tr57:
//line NONE:1
 lex.te = ( lex.p)+1

	goto st75
	st75:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof75
		}
	st_case_75:
//line /src/scanner.go:2525
		if  lex.data[( lex.p)] == 37 {
			goto st21
		}
		goto tr62
	st21:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof21
		}
	st_case_21:
		switch  lex.data[( lex.p)] {
		case 99:
			goto st22
		case 119:
			goto st27
		}
		goto tr0
	st22:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof22
		}
	st_case_22:
		if  lex.data[( lex.p)] == 121 {
			goto st23
		}
		goto tr0
	st23:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof23
		}
	st_case_23:
		if  lex.data[( lex.p)] == 99 {
			goto st24
		}
		goto tr0
	st24:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof24
		}
	st_case_24:
		if  lex.data[( lex.p)] == 108 {
			goto st25
		}
		goto tr0
	st25:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof25
		}
	st_case_25:
		if  lex.data[( lex.p)] == 101 {
			goto st26
		}
		goto tr0
	st26:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof26
		}
	st_case_26:
		if  lex.data[( lex.p)] == 32 {
			goto tr31
		}
		goto tr0
	st27:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof27
		}
	st_case_27:
		if  lex.data[( lex.p)] == 104 {
			goto st28
		}
		goto tr0
	st28:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof28
		}
	st_case_28:
		if  lex.data[( lex.p)] == 101 {
			goto st29
		}
		goto tr0
	st29:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof29
		}
	st_case_29:
		if  lex.data[( lex.p)] == 110 {
			goto st30
		}
		goto tr0
	st30:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof30
		}
	st_case_30:
		if  lex.data[( lex.p)] == 32 {
			goto tr35
		}
		goto tr0
	st76:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof76
		}
	st_case_76:
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto tr22
		}
		goto tr62
tr59:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:142
 lex.act = 27;
	goto st77
	st77:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof77
		}
	st_case_77:
//line /src/scanner.go:2644
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto st18
		}
		goto tr62
tr60:
//line NONE:1
 lex.te = ( lex.p)+1

//line /src/scanner.rl:142
 lex.act = 27;
	goto st78
	st78:
		if ( lex.p)++; ( lex.p) == ( lex.pe) {
			goto _test_eof78
		}
	st_case_78:
//line /src/scanner.go:2661
		if 128 <=  lex.data[( lex.p)] &&  lex.data[( lex.p)] <= 191 {
			goto st19
		}
		goto tr62
	st_out:
	_test_eof31:  lex.cs = 31; goto _test_eof
	_test_eof32:  lex.cs = 32; goto _test_eof
	_test_eof33:  lex.cs = 33; goto _test_eof
	_test_eof34:  lex.cs = 34; goto _test_eof
	_test_eof0:  lex.cs = 0; goto _test_eof
	_test_eof1:  lex.cs = 1; goto _test_eof
	_test_eof35:  lex.cs = 35; goto _test_eof
	_test_eof2:  lex.cs = 2; goto _test_eof
	_test_eof3:  lex.cs = 3; goto _test_eof
	_test_eof4:  lex.cs = 4; goto _test_eof
	_test_eof5:  lex.cs = 5; goto _test_eof
	_test_eof6:  lex.cs = 6; goto _test_eof
	_test_eof7:  lex.cs = 7; goto _test_eof
	_test_eof8:  lex.cs = 8; goto _test_eof
	_test_eof9:  lex.cs = 9; goto _test_eof
	_test_eof10:  lex.cs = 10; goto _test_eof
	_test_eof11:  lex.cs = 11; goto _test_eof
	_test_eof36:  lex.cs = 36; goto _test_eof
	_test_eof12:  lex.cs = 12; goto _test_eof
	_test_eof13:  lex.cs = 13; goto _test_eof
	_test_eof37:  lex.cs = 37; goto _test_eof
	_test_eof38:  lex.cs = 38; goto _test_eof
	_test_eof14:  lex.cs = 14; goto _test_eof
	_test_eof39:  lex.cs = 39; goto _test_eof
	_test_eof40:  lex.cs = 40; goto _test_eof
	_test_eof41:  lex.cs = 41; goto _test_eof
	_test_eof15:  lex.cs = 15; goto _test_eof
	_test_eof16:  lex.cs = 16; goto _test_eof
	_test_eof17:  lex.cs = 17; goto _test_eof
	_test_eof42:  lex.cs = 42; goto _test_eof
	_test_eof43:  lex.cs = 43; goto _test_eof
	_test_eof44:  lex.cs = 44; goto _test_eof
	_test_eof45:  lex.cs = 45; goto _test_eof
	_test_eof46:  lex.cs = 46; goto _test_eof
	_test_eof18:  lex.cs = 18; goto _test_eof
	_test_eof19:  lex.cs = 19; goto _test_eof
	_test_eof20:  lex.cs = 20; goto _test_eof
	_test_eof47:  lex.cs = 47; goto _test_eof
	_test_eof48:  lex.cs = 48; goto _test_eof
	_test_eof49:  lex.cs = 49; goto _test_eof
	_test_eof50:  lex.cs = 50; goto _test_eof
	_test_eof51:  lex.cs = 51; goto _test_eof
	_test_eof52:  lex.cs = 52; goto _test_eof
	_test_eof53:  lex.cs = 53; goto _test_eof
	_test_eof54:  lex.cs = 54; goto _test_eof
	_test_eof55:  lex.cs = 55; goto _test_eof
	_test_eof56:  lex.cs = 56; goto _test_eof
	_test_eof57:  lex.cs = 57; goto _test_eof
	_test_eof58:  lex.cs = 58; goto _test_eof
	_test_eof59:  lex.cs = 59; goto _test_eof
	_test_eof60:  lex.cs = 60; goto _test_eof
	_test_eof61:  lex.cs = 61; goto _test_eof
	_test_eof62:  lex.cs = 62; goto _test_eof
	_test_eof63:  lex.cs = 63; goto _test_eof
	_test_eof64:  lex.cs = 64; goto _test_eof
	_test_eof65:  lex.cs = 65; goto _test_eof
	_test_eof66:  lex.cs = 66; goto _test_eof
	_test_eof67:  lex.cs = 67; goto _test_eof
	_test_eof68:  lex.cs = 68; goto _test_eof
	_test_eof69:  lex.cs = 69; goto _test_eof
	_test_eof70:  lex.cs = 70; goto _test_eof
	_test_eof71:  lex.cs = 71; goto _test_eof
	_test_eof72:  lex.cs = 72; goto _test_eof
	_test_eof73:  lex.cs = 73; goto _test_eof
	_test_eof74:  lex.cs = 74; goto _test_eof
	_test_eof75:  lex.cs = 75; goto _test_eof
	_test_eof21:  lex.cs = 21; goto _test_eof
	_test_eof22:  lex.cs = 22; goto _test_eof
	_test_eof23:  lex.cs = 23; goto _test_eof
	_test_eof24:  lex.cs = 24; goto _test_eof
	_test_eof25:  lex.cs = 25; goto _test_eof
	_test_eof26:  lex.cs = 26; goto _test_eof
	_test_eof27:  lex.cs = 27; goto _test_eof
	_test_eof28:  lex.cs = 28; goto _test_eof
	_test_eof29:  lex.cs = 29; goto _test_eof
	_test_eof30:  lex.cs = 30; goto _test_eof
	_test_eof76:  lex.cs = 76; goto _test_eof
	_test_eof77:  lex.cs = 77; goto _test_eof
	_test_eof78:  lex.cs = 78; goto _test_eof

	_test_eof: {}
	if ( lex.p) == eof {
		switch  lex.cs {
		case 32:
			goto tr61
		case 33:
			goto tr62
		case 34:
			goto tr62
		case 0:
			goto tr0
		case 1:
			goto tr0
		case 35:
			goto tr62
		case 2:
			goto tr0
		case 3:
			goto tr0
		case 4:
			goto tr0
		case 5:
			goto tr0
		case 6:
			goto tr0
		case 7:
			goto tr0
		case 8:
			goto tr0
		case 9:
			goto tr0
		case 10:
			goto tr0
		case 11:
			goto tr0
		case 36:
			goto tr62
		case 12:
			goto tr0
		case 13:
			goto tr0
		case 37:
			goto tr62
		case 38:
			goto tr66
		case 14:
			goto tr16
		case 39:
			goto tr68
		case 40:
			goto tr62
		case 41:
			goto tr71
		case 15:
			goto tr18
		case 16:
			goto tr18
		case 17:
			goto tr18
		case 42:
			goto tr62
		case 43:
			goto tr62
		case 44:
			goto tr62
		case 45:
			goto tr18
		case 46:
			goto tr80
		case 18:
			goto tr18
		case 19:
			goto tr18
		case 20:
			goto tr18
		case 47:
			goto tr80
		case 48:
			goto tr80
		case 49:
			goto tr80
		case 50:
			goto tr80
		case 51:
			goto tr80
		case 52:
			goto tr80
		case 53:
			goto tr80
		case 54:
			goto tr80
		case 55:
			goto tr80
		case 56:
			goto tr80
		case 57:
			goto tr80
		case 58:
			goto tr80
		case 59:
			goto tr80
		case 60:
			goto tr80
		case 61:
			goto tr80
		case 62:
			goto tr80
		case 63:
			goto tr80
		case 64:
			goto tr80
		case 65:
			goto tr80
		case 66:
			goto tr80
		case 67:
			goto tr80
		case 68:
			goto tr80
		case 69:
			goto tr80
		case 70:
			goto tr80
		case 71:
			goto tr80
		case 72:
			goto tr80
		case 73:
			goto tr80
		case 74:
			goto tr80
		case 75:
			goto tr62
		case 21:
			goto tr0
		case 22:
			goto tr0
		case 23:
			goto tr0
		case 24:
			goto tr0
		case 25:
			goto tr0
		case 26:
			goto tr0
		case 27:
			goto tr0
		case 28:
			goto tr0
		case 29:
			goto tr0
		case 30:
			goto tr0
		case 76:
			goto tr62
		case 77:
			goto tr62
		case 78:
			goto tr62
		}
	}

	_out: {}
	}

//line /src/scanner.rl:146


	return tok
}

func (lex *lexer) Error(e string) {
    // fmt.Println("scan error:", e)
}

// unescapeString processes backslash escape sequences in a Liquid string literal.
// The input slice should not include the surrounding quotes.
func unescapeString(b []byte) string {
	// Fast path: no backslash at all.
	if !strings.ContainsRune(string(b), '\\') {
		return string(b)
	}
	out := make([]byte, 0, len(b))
	for i := 0; i < len(b); i++ {
		if b[i] == '\\' && i+1 < len(b) {
			i++
			switch b[i] {
			case 'n':
				out = append(out, '\n')
			case 'r':
				out = append(out, '\r')
			case 't':
				out = append(out, '\t')
			case '"', '\'', '\\':
				out = append(out, b[i])
			default:
				// Unknown escapes are kept as-is (backslash + char).
				out = append(out, '\\', b[i])
			}
		} else {
			out = append(out, b[i])
		}
	}
	return string(out)
}

func isValidUnicodeIdentifier(s string) bool {
	// Remove optional trailing '?' for validation
	checkStr := s
	if len(s) > 0 && s[len(s)-1] == '?' {
		checkStr = s[:len(s)-1]
	}

	if len(checkStr) == 0 {
		return false // "?" alone is invalid
	}

	// validate the core identifier part
	runes := []rune(checkStr)
	if !unicode.IsLetter(runes[0]) && runes[0] != '_' {
		return false
	}

	for _, r := range runes[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsMark(r) && r != '_' && r != '-' {
			return false
		}
	}

	return true
}
