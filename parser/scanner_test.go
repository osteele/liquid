package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var scannerCountTests = []struct {
	in  string
	len int
}{
	{`{% tag arg %}`, 1},
	{`{% tag arg %}{% tag %}`, 2},
	{`{% tag arg %}{% tag arg %}{% tag %}`, 3},
	{`{% tag %}{% tag %}`, 2},
	{`{% tag arg %}{% tag arg %}{% tag %}{% tag %}`, 4},
	{`{{ expr }}`, 1},
	{`{{ expr arg }}`, 1},
	{`{{ expr }}{{ expr }}`, 2},
	{`{{ expr arg }}{{ expr arg }}`, 2},
}

func TestScan(t *testing.T) {
	scan := func(src string) []Token { return Scan(src, SourceLoc{}, nil) }
	tokens := scan("12")
	verifyTokens(t, TextTokenType, 1, tokens)
	require.Equal(t, "12", tokens[0].Source)

	tokens = scan("{{obj}}")
	verifyTokens(t, ObjTokenType, 1, tokens)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("{{ obj }}")
	verifyTokens(t, ObjTokenType, 1, tokens)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("{%tag args%}")
	verifyTokens(t, TagTokenType, 1, tokens)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("{% tag args %}")
	verifyTokens(t, TagTokenType, 1, tokens)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("pre{% tag args %}mid{{ object }}post")
	require.Equal(t, `[TextTokenType{"pre"} TagTokenType{Tag:"tag", Args:"args", l: false, r: false} TextTokenType{"mid"} ObjTokenType{"object"} TextTokenType{"post"}]`, fmt.Sprint(tokens))

	for i, test := range scannerCountTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			require.Len(t, tokens, test.len)
		})
	}
}

func TestScan_ws(t *testing.T) {
	// whitespace control
	scan := func(src string) []Token { return Scan(src, SourceLoc{}, nil) }

	wsTests := []struct {
		in, expect  string
		left, right bool
	}{
		{`{{ expr }}`, "expr", false, false},
		{`{{- expr }}`, "expr", true, false},
		{`{{ expr -}}`, "expr", false, true},
		{`{{- expr -}}`, "expr", true, true},
	}
	for i, test := range wsTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			tok := tokens[0]
			if test.expect == "tag" {
				require.Equalf(t, "tag", tok.Name, test.in)
				require.Equalf(t, "arg", tok.Args, test.in)
			} else {
				require.Equalf(t, "expr", tok.Args, test.in)
			}
			require.Equalf(t, test.left, tok.TrimLeft, test.in)
			require.Equalf(t, test.right, tok.TrimRight, test.in)
		})
	}
}

func TestScanWhiteSpaceTokens(t *testing.T) {
	// whitespace control
	scan := func(src string) []Token { return Scan(src, SourceLoc{}, nil) }

	wsTests := []struct {
		in        string
		numTokens int
		expected  []TokenType
	}{
		{" ", 1, []TokenType{WhitespaceTokenType}},
		{"    ", 1, []TokenType{WhitespaceTokenType}},
		{"\n", 1, []TokenType{WhitespaceTokenType}},
		{"\t", 1, []TokenType{WhitespaceTokenType}},
		{"\t\t\t\t", 1, []TokenType{WhitespaceTokenType}},
		{"\t\n\t", 3, []TokenType{WhitespaceTokenType, WhitespaceTokenType, WhitespaceTokenType}},
		{"{{ expr }} {{ expr }}", 3, []TokenType{ObjTokenType, WhitespaceTokenType, ObjTokenType}},
		{"{{ expr }}\t\n\t{{ expr }}", 5, []TokenType{ObjTokenType, WhitespaceTokenType, WhitespaceTokenType, WhitespaceTokenType, ObjTokenType}},
		{"{{ expr }}\t \t\n\t \t{{ expr }}", 5, []TokenType{ObjTokenType, WhitespaceTokenType, WhitespaceTokenType, WhitespaceTokenType, ObjTokenType}},
		{"{{ expr }}\t \t\nSomeText\n\t \t{{ expr }}", 7, []TokenType{ObjTokenType, WhitespaceTokenType, WhitespaceTokenType, TextTokenType, WhitespaceTokenType, WhitespaceTokenType, ObjTokenType}},
	}
	for i, test := range wsTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			require.Len(t, tokens, test.numTokens)
			for x, tok := range tokens {
				require.Equal(t, test.expected[x], tok.Type)
			}
		})
	}
}

func TestScanTokenLocationParsing(t *testing.T) {
	// whitespace control
	scan := func(src string) []Token { return Scan(src, SourceLoc{LineNo: 1}, nil) }

	wsTests := []struct {
		in              string
		expectedLineNos []int
	}{
		{"\t \t \tsometext", []int{1, 1}},
		{"\t\n\t", []int{1, 1, 2}},
		{"\nsometext", []int{1, 2}},
		{"{{ expr }}\t \t\nSomeText\n\t \t{{ expr }}", []int{1, 1, 1, 2, 2, 3, 3}},
	}
	for i, test := range wsTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			require.Len(t, tokens, len(test.expectedLineNos))
			for x, tok := range tokens {
				require.Equal(t, test.expectedLineNos[x], tok.SourceLoc.LineNo)
			}
		})
	}
}

var scannerCountTestsDelims = []struct {
	in  string
	len int
}{
	{`TAG*LEFT tag arg TAG!RIGHT`, 1},
	{`TAG*LEFT tag arg TAG!RIGHTTAG*LEFT tag TAG!RIGHT`, 2},
	{`TAG*LEFT tag arg TAG!RIGHTTAG*LEFT tag arg TAG!RIGHTTAG*LEFT tag TAG!RIGHT`, 3},
	{`TAG*LEFT tag TAG!RIGHTTAG*LEFT tag TAG!RIGHT`, 2},
	{`TAG*LEFT tag arg TAG!RIGHTTAG*LEFT tag arg TAG!RIGHTTAG*LEFT tag TAG!RIGHTTAG*LEFT tag TAG!RIGHT`, 4},
	{`OBJECT@LEFT expr OBJECT#RIGHT`, 1},
	{`OBJECT@LEFT expr arg OBJECT#RIGHT`, 1},
	{`OBJECT@LEFT expr OBJECT#RIGHTOBJECT@LEFT expr OBJECT#RIGHT`, 2},
	{`OBJECT@LEFT expr arg OBJECT#RIGHTOBJECT@LEFT expr arg OBJECT#RIGHT`, 2},
}

func TestScan_delims(t *testing.T) {
	scan := func(src string) []Token {
		return Scan(src, SourceLoc{}, []string{"OBJECT@LEFT", "OBJECT#RIGHT", "TAG*LEFT", "TAG!RIGHT"})
	}
	tokens := scan("12")
	verifyTokens(t, TextTokenType, 1, tokens)
	require.Equal(t, "12", tokens[0].Source)

	tokens = scan("OBJECT@LEFTobjOBJECT#RIGHT")
	verifyTokens(t, ObjTokenType, 1, tokens)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("OBJECT@LEFT obj OBJECT#RIGHT")
	verifyTokens(t, ObjTokenType, 1, tokens)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("TAG*LEFTtag argsTAG!RIGHT")
	verifyTokens(t, TagTokenType, 1, tokens)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("TAG*LEFT tag args TAG!RIGHT")
	verifyTokens(t, TagTokenType, 1, tokens)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("\npreTAG*LEFT tag args TAG!RIGHTmidOBJECT@LEFT object OBJECT#RIGHTpost\t")
	require.Equal(t, `[WhitespaceTokenType{"New Line"} TextTokenType{"pre"} TagTokenType{Tag:"tag", Args:"args", l: false, r: false} TextTokenType{"mid"} ObjTokenType{"object"} TextTokenType{"post"} WhitespaceTokenType{"Whitespace"}]`, fmt.Sprint(tokens))

	for i, test := range scannerCountTestsDelims {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			require.Len(t, tokens, test.len)
		})
	}
}

func verifyTokens(t require.TestingT, tokenType TokenType, length int, tokens []Token) []Token {
	require.NotNil(t, tokens)
	require.Len(t, tokens, length)
	require.Equal(t, tokenType, tokens[0].Type)
	return tokens
}
