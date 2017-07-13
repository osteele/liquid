//line expressions.y:2
package expression

import __yyfmt__ "fmt"

//line expressions.y:2
import (
	"fmt"
	"github.com/osteele/liquid/evaluator"
)

func init() {
	// This allows adding and removing references to fmt in the rules below,
	// without having to edit the import statement to avoid erorrs each time.
	_ = fmt.Sprint("")
}

//line expressions.y:15
type yySymType struct {
	yys           int
	name          string
	val           interface{}
	f             func(Context) interface{}
	s             string
	ss            []string
	cycle         Cycle
	cyclefn       func(string) Cycle
	loop          Loop
	loopmods      loopModifiers
	filter_params []valueFn
}

const LITERAL = 57346
const IDENTIFIER = 57347
const KEYWORD = 57348
const PROPERTY = 57349
const ASSIGN = 57350
const CYCLE = 57351
const LOOP = 57352
const EQ = 57353
const NEQ = 57354
const GE = 57355
const LE = 57356
const IN = 57357
const AND = 57358
const OR = 57359
const CONTAINS = 57360

var yyToknames = [...]string{
	"$end",
	"error",
	"$unk",
	"LITERAL",
	"IDENTIFIER",
	"KEYWORD",
	"PROPERTY",
	"ASSIGN",
	"CYCLE",
	"LOOP",
	"EQ",
	"NEQ",
	"GE",
	"LE",
	"IN",
	"AND",
	"OR",
	"CONTAINS",
	"'.'",
	"'|'",
	"'<'",
	"'>'",
	"';'",
	"'='",
	"':'",
	"','",
	"'['",
	"']'",
	"'('",
	"')'",
}
var yyStatenames = [...]string{}

const yyEofCode = 1
const yyErrCode = 2
const yyInitialStackSize = 16

//line yacctab:1
var yyExca = [...]int{
	-1, 1,
	1, -1,
	-2, 0,
}

const yyPrivate = 57344

const yyLast = 81

var yyAct = [...]int{

	8, 38, 22, 7, 13, 14, 37, 39, 17, 9,
	10, 9, 10, 3, 4, 5, 65, 22, 52, 39,
	34, 40, 23, 60, 44, 45, 46, 47, 48, 49,
	50, 51, 13, 14, 11, 54, 11, 23, 53, 12,
	21, 35, 54, 61, 59, 57, 55, 21, 56, 22,
	41, 66, 67, 24, 25, 28, 29, 62, 63, 6,
	30, 42, 43, 27, 26, 2, 68, 20, 15, 23,
	69, 18, 1, 32, 33, 64, 19, 31, 36, 16,
	58,
}
var yyPact = [...]int{

	5, -1000, 16, 63, 67, 62, -1000, 27, 42, -1000,
	-1000, 7, -1000, 7, 7, -4, 18, -19, -1000, -2,
	35, 56, -1000, 7, 7, 7, 7, 7, 7, 7,
	7, -12, -1000, -1000, 7, -1000, -1000, 67, -1000, 67,
	-1000, 7, -1000, 7, -5, 10, 10, 10, 10, 10,
	10, 10, -1000, 20, 10, -7, -7, 27, -10, 10,
	-1000, -1000, -1000, -1000, 46, 7, -1000, 66, 10, -1000,
}
var yyPgo = [...]int{

	0, 0, 59, 3, 65, 80, 79, 78, 1, 76,
	75, 8, 72,
}
var yyR1 = [...]int{

	0, 12, 12, 12, 12, 6, 7, 7, 8, 8,
	11, 9, 10, 10, 10, 1, 1, 1, 1, 1,
	3, 3, 3, 5, 5, 2, 2, 2, 2, 2,
	2, 2, 2, 4, 4, 4,
}
var yyR2 = [...]int{

	0, 2, 5, 3, 3, 2, 3, 1, 0, 3,
	1, 4, 0, 2, 3, 1, 1, 2, 4, 3,
	1, 3, 4, 1, 3, 1, 3, 3, 3, 3,
	3, 3, 3, 1, 3, 3,
}
var yyChk = [...]int{

	-1000, -12, -4, 8, 9, 10, -2, -3, -1, 4,
	5, 29, 23, 16, 17, 5, -6, -11, 4, -9,
	5, 20, 7, 27, 11, 12, 22, 21, 13, 14,
	18, -4, -2, -2, 24, 23, -7, 25, -8, 26,
	23, 15, 5, 6, -1, -1, -1, -1, -1, -1,
	-1, -1, 30, -3, -1, -11, -11, -3, -5, -1,
	28, 23, -8, -8, -10, 26, 5, 6, -1, 4,
}
var yyDef = [...]int{

	0, -2, 0, 0, 0, 0, 33, 25, 20, 15,
	16, 0, 1, 0, 0, 0, 0, 8, 10, 0,
	0, 0, 17, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 34, 35, 0, 3, 5, 0, 7, 0,
	4, 0, 21, 0, 0, 26, 27, 28, 29, 30,
	31, 32, 19, 0, 20, 8, 8, 12, 22, 23,
	18, 2, 6, 9, 11, 0, 13, 0, 24, 14,
}
var yyTok1 = [...]int{

	1, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	29, 30, 3, 3, 26, 3, 19, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 25, 23,
	21, 24, 22, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 27, 3, 28, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	3, 3, 3, 3, 20,
}
var yyTok2 = [...]int{

	2, 3, 4, 5, 6, 7, 8, 9, 10, 11,
	12, 13, 14, 15, 16, 17, 18,
}
var yyTok3 = [...]int{
	0,
}

var yyErrorMessages = [...]struct {
	state int
	token int
	msg   string
}{}

//line yaccpar:1

/*	parser for yacc output	*/

var (
	yyDebug        = 0
	yyErrorVerbose = false
)

type yyLexer interface {
	Lex(lval *yySymType) int
	Error(s string)
}

type yyParser interface {
	Parse(yyLexer) int
	Lookahead() int
}

type yyParserImpl struct {
	lval  yySymType
	stack [yyInitialStackSize]yySymType
	char  int
}

func (p *yyParserImpl) Lookahead() int {
	return p.char
}

func yyNewParser() yyParser {
	return &yyParserImpl{}
}

const yyFlag = -1000

func yyTokname(c int) string {
	if c >= 1 && c-1 < len(yyToknames) {
		if yyToknames[c-1] != "" {
			return yyToknames[c-1]
		}
	}
	return __yyfmt__.Sprintf("tok-%v", c)
}

func yyStatname(s int) string {
	if s >= 0 && s < len(yyStatenames) {
		if yyStatenames[s] != "" {
			return yyStatenames[s]
		}
	}
	return __yyfmt__.Sprintf("state-%v", s)
}

func yyErrorMessage(state, lookAhead int) string {
	const TOKSTART = 4

	if !yyErrorVerbose {
		return "syntax error"
	}

	for _, e := range yyErrorMessages {
		if e.state == state && e.token == lookAhead {
			return "syntax error: " + e.msg
		}
	}

	res := "syntax error: unexpected " + yyTokname(lookAhead)

	// To match Bison, suggest at most four expected tokens.
	expected := make([]int, 0, 4)

	// Look for shiftable tokens.
	base := yyPact[state]
	for tok := TOKSTART; tok-1 < len(yyToknames); tok++ {
		if n := base + tok; n >= 0 && n < yyLast && yyChk[yyAct[n]] == tok {
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}
	}

	if yyDef[state] == -2 {
		i := 0
		for yyExca[i] != -1 || yyExca[i+1] != state {
			i += 2
		}

		// Look for tokens that we accept or reduce.
		for i += 2; yyExca[i] >= 0; i += 2 {
			tok := yyExca[i]
			if tok < TOKSTART || yyExca[i+1] == 0 {
				continue
			}
			if len(expected) == cap(expected) {
				return res
			}
			expected = append(expected, tok)
		}

		// If the default action is to accept or reduce, give up.
		if yyExca[i+1] != 0 {
			return res
		}
	}

	for i, tok := range expected {
		if i == 0 {
			res += ", expecting "
		} else {
			res += " or "
		}
		res += yyTokname(tok)
	}
	return res
}

func yylex1(lex yyLexer, lval *yySymType) (char, token int) {
	token = 0
	char = lex.Lex(lval)
	if char <= 0 {
		token = yyTok1[0]
		goto out
	}
	if char < len(yyTok1) {
		token = yyTok1[char]
		goto out
	}
	if char >= yyPrivate {
		if char < yyPrivate+len(yyTok2) {
			token = yyTok2[char-yyPrivate]
			goto out
		}
	}
	for i := 0; i < len(yyTok3); i += 2 {
		token = yyTok3[i+0]
		if token == char {
			token = yyTok3[i+1]
			goto out
		}
	}

out:
	if token == 0 {
		token = yyTok2[1] /* unknown char */
	}
	if yyDebug >= 3 {
		__yyfmt__.Printf("lex %s(%d)\n", yyTokname(token), uint(char))
	}
	return char, token
}

func yyParse(yylex yyLexer) int {
	return yyNewParser().Parse(yylex)
}

func (yyrcvr *yyParserImpl) Parse(yylex yyLexer) int {
	var yyn int
	var yyVAL yySymType
	var yyDollar []yySymType
	_ = yyDollar // silence set and not used
	yyS := yyrcvr.stack[:]

	Nerrs := 0   /* number of errors */
	Errflag := 0 /* error recovery flag */
	yystate := 0
	yyrcvr.char = -1
	yytoken := -1 // yyrcvr.char translated into internal numbering
	defer func() {
		// Make sure we report no lookahead when not parsing.
		yystate = -1
		yyrcvr.char = -1
		yytoken = -1
	}()
	yyp := -1
	goto yystack

ret0:
	return 0

ret1:
	return 1

yystack:
	/* put a state and value onto the stack */
	if yyDebug >= 4 {
		__yyfmt__.Printf("char %v in %v\n", yyTokname(yytoken), yyStatname(yystate))
	}

	yyp++
	if yyp >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyS[yyp] = yyVAL
	yyS[yyp].yys = yystate

yynewstate:
	yyn = yyPact[yystate]
	if yyn <= yyFlag {
		goto yydefault /* simple state */
	}
	if yyrcvr.char < 0 {
		yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
	}
	yyn += yytoken
	if yyn < 0 || yyn >= yyLast {
		goto yydefault
	}
	yyn = yyAct[yyn]
	if yyChk[yyn] == yytoken { /* valid shift */
		yyrcvr.char = -1
		yytoken = -1
		yyVAL = yyrcvr.lval
		yystate = yyn
		if Errflag > 0 {
			Errflag--
		}
		goto yystack
	}

yydefault:
	/* default state action */
	yyn = yyDef[yystate]
	if yyn == -2 {
		if yyrcvr.char < 0 {
			yyrcvr.char, yytoken = yylex1(yylex, &yyrcvr.lval)
		}

		/* look through exception table */
		xi := 0
		for {
			if yyExca[xi+0] == -1 && yyExca[xi+1] == yystate {
				break
			}
			xi += 2
		}
		for xi += 2; ; xi += 2 {
			yyn = yyExca[xi+0]
			if yyn < 0 || yyn == yytoken {
				break
			}
		}
		yyn = yyExca[xi+1]
		if yyn < 0 {
			goto ret0
		}
	}
	if yyn == 0 {
		/* error ... attempt to resume parsing */
		switch Errflag {
		case 0: /* brand new error */
			yylex.Error(yyErrorMessage(yystate, yytoken))
			Nerrs++
			if yyDebug >= 1 {
				__yyfmt__.Printf("%s", yyStatname(yystate))
				__yyfmt__.Printf(" saw %s\n", yyTokname(yytoken))
			}
			fallthrough

		case 1, 2: /* incompletely recovered error ... try again */
			Errflag = 3

			/* find a state where "error" is a legal shift action */
			for yyp >= 0 {
				yyn = yyPact[yyS[yyp].yys] + yyErrCode
				if yyn >= 0 && yyn < yyLast {
					yystate = yyAct[yyn] /* simulate a shift of "error" */
					if yyChk[yystate] == yyErrCode {
						goto yystack
					}
				}

				/* the current p has no shift on "error", pop stack */
				if yyDebug >= 2 {
					__yyfmt__.Printf("error recovery pops state %d\n", yyS[yyp].yys)
				}
				yyp--
			}
			/* there is no state on the stack with an error shift ... abort */
			goto ret1

		case 3: /* no shift yet; clobber input char */
			if yyDebug >= 2 {
				__yyfmt__.Printf("error recovery discards %s\n", yyTokname(yytoken))
			}
			if yytoken == yyEofCode {
				goto ret1
			}
			yyrcvr.char = -1
			yytoken = -1
			goto yynewstate /* try again in the same state */
		}
	}

	/* reduction by production yyn */
	if yyDebug >= 2 {
		__yyfmt__.Printf("reduce %v in:\n\t%v\n", yyn, yyStatname(yystate))
	}

	yynt := yyn
	yypt := yyp
	_ = yypt // guard against "declared and not used"

	yyp -= yyR2[yyn]
	// yyp is now the index of $0. Perform the default action. Iff the
	// reduced production is ε, $1 is possibly out of range.
	if yyp+1 >= len(yyS) {
		nyys := make([]yySymType, len(yyS)*2)
		copy(nyys, yyS)
		yyS = nyys
	}
	yyVAL = yyS[yyp+1]

	/* consult goto table to find next state */
	yyn = yyR1[yyn]
	yyg := yyPgo[yyn]
	yyj := yyg + yyS[yyp].yys + 1

	if yyj >= yyLast {
		yystate = yyAct[yyg]
	} else {
		yystate = yyAct[yyj]
		if yyChk[yystate] != -yyn {
			yystate = yyAct[yyg]
		}
	}
	// dummy call; replaced with literal code
	switch yynt {

	case 1:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line expressions.y:43
		{
			yylex.(*lexer).val = yyDollar[1].f
		}
	case 2:
		yyDollar = yyS[yypt-5 : yypt+1]
		//line expressions.y:44
		{
			yylex.(*lexer).Assignment = Assignment{yyDollar[2].name, &expression{yyDollar[4].f}}
		}
	case 3:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:47
		{
			yylex.(*lexer).Cycle = yyDollar[2].cycle
		}
	case 4:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:48
		{
			yylex.(*lexer).Loop = yyDollar[2].loop
		}
	case 5:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line expressions.y:51
		{
			yyVAL.cycle = yyDollar[2].cyclefn(yyDollar[1].s)
		}
	case 6:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:54
		{
			h, t := yyDollar[2].s, yyDollar[3].ss
			yyVAL.cyclefn = func(g string) Cycle { return Cycle{g, append([]string{h}, t...)} }
		}
	case 7:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line expressions.y:58
		{
			vals := yyDollar[1].ss
			yyVAL.cyclefn = func(h string) Cycle { return Cycle{Values: append([]string{h}, vals...)} }
		}
	case 8:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line expressions.y:65
		{
			yyVAL.ss = []string{}
		}
	case 9:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:66
		{
			yyVAL.ss = append([]string{yyDollar[2].s}, yyDollar[3].ss...)
		}
	case 10:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line expressions.y:69
		{
			s, ok := yyDollar[1].val.(string)
			if !ok {
				panic(ParseError(fmt.Sprintf("expected a string for %q", yyDollar[1].val)))
			}
			yyVAL.s = s
		}
	case 11:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line expressions.y:77
		{
			name, expr, mods := yyDollar[1].name, yyDollar[3].f, yyDollar[4].loopmods
			yyVAL.loop = Loop{name, &expression{expr}, mods}
		}
	case 12:
		yyDollar = yyS[yypt-0 : yypt+1]
		//line expressions.y:83
		{
			yyVAL.loopmods = loopModifiers{}
		}
	case 13:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line expressions.y:84
		{
			switch yyDollar[2].name {
			case "reversed":
				yyDollar[1].loopmods.Reversed = true
			default:
				panic(ParseError(fmt.Sprintf("undefined loop modifier %q", yyDollar[2].name)))
			}
			yyVAL.loopmods = yyDollar[1].loopmods
		}
	case 14:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:93
		{ // TODO can this be a variable?
			switch yyDollar[2].name {
			case "limit":
				limit, ok := yyDollar[3].val.(int)
				if !ok {
					panic(ParseError(fmt.Sprintf("loop limit must an integer")))
				}
				yyDollar[1].loopmods.Limit = &limit
			case "offset":
				offset, ok := yyDollar[3].val.(int)
				if !ok {
					panic(ParseError(fmt.Sprintf("loop offset must an integer")))
				}
				yyDollar[1].loopmods.Offset = offset
			default:
				panic(ParseError(fmt.Sprintf("undefined loop modifier %q", yyDollar[2].name)))
			}
			yyVAL.loopmods = yyDollar[1].loopmods
		}
	case 15:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line expressions.y:115
		{
			val := yyDollar[1].val
			yyVAL.f = func(_ Context) interface{} { return val }
		}
	case 16:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line expressions.y:116
		{
			name := yyDollar[1].name
			yyVAL.f = func(ctx Context) interface{} { return ctx.Get(name) }
		}
	case 17:
		yyDollar = yyS[yypt-2 : yypt+1]
		//line expressions.y:117
		{
			yyVAL.f = makeObjectPropertyExpr(yyDollar[1].f, yyDollar[2].name)
		}
	case 18:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line expressions.y:118
		{
			yyVAL.f = makeIndexExpr(yyDollar[1].f, yyDollar[3].f)
		}
	case 19:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:119
		{
			yyVAL.f = yyDollar[2].f
		}
	case 21:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:124
		{
			yyVAL.f = makeFilter(yyDollar[1].f, yyDollar[3].name, nil)
		}
	case 22:
		yyDollar = yyS[yypt-4 : yypt+1]
		//line expressions.y:125
		{
			yyVAL.f = makeFilter(yyDollar[1].f, yyDollar[3].name, yyDollar[4].filter_params)
		}
	case 23:
		yyDollar = yyS[yypt-1 : yypt+1]
		//line expressions.y:129
		{
			yyVAL.filter_params = []valueFn{yyDollar[1].f}
		}
	case 24:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:131
		{
			yyVAL.filter_params = append(yyDollar[1].filter_params, yyDollar[3].f)
		}
	case 26:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:135
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				a, b := fa(ctx), fb(ctx)
				return evaluator.Equal(a, b)
			}
		}
	case 27:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:142
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				a, b := fa(ctx), fb(ctx)
				return !evaluator.Equal(a, b)
			}
		}
	case 28:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:149
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				a, b := fa(ctx), fb(ctx)
				return evaluator.Less(b, a)
			}
		}
	case 29:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:156
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				a, b := fa(ctx), fb(ctx)
				return evaluator.Less(a, b)
			}
		}
	case 30:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:163
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				a, b := fa(ctx), fb(ctx)
				return evaluator.Less(b, a) || evaluator.Equal(a, b)
			}
		}
	case 31:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:170
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				a, b := fa(ctx), fb(ctx)
				return evaluator.Less(a, b) || evaluator.Equal(a, b)
			}
		}
	case 32:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:177
		{
			yyVAL.f = makeContainsExpr(yyDollar[1].f, yyDollar[3].f)
		}
	case 34:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:182
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				return evaluator.IsTrue(fa(ctx)) && evaluator.IsTrue(fb(ctx))
			}
		}
	case 35:
		yyDollar = yyS[yypt-3 : yypt+1]
		//line expressions.y:188
		{
			fa, fb := yyDollar[1].f, yyDollar[3].f
			yyVAL.f = func(ctx Context) interface{} {
				return evaluator.IsTrue(fa(ctx)) || evaluator.IsTrue(fb(ctx))
			}
		}
	}
	goto yystack /* stack new state and value */
}
