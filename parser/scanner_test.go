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
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TextTokenType, tokens[0].Type)
	require.Equal(t, "12", tokens[0].Source)

	tokens = scan("{{obj}}")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, ObjTokenType, tokens[0].Type)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("{{ obj }}")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, ObjTokenType, tokens[0].Type)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("{%tag args%}")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagTokenType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("{% tag args %}")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagTokenType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("pre{% tag args %}mid{{ object }}post")
	require.Equal(t, `[TextTokenType{"pre"} TagTokenType{Tag:"tag", Args:"args"} TextTokenType{"mid"} ObjTokenType{"object"} TextTokenType{"post"}]`, fmt.Sprint(tokens))

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
		{`{% tag arg %}`, "tag", false, false},
		{`{%- tag arg %}`, "tag", true, false},
		{`{% tag arg -%}`, "tag", false, true},
	}
	for i, test := range wsTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			require.Len(t, tokens, 1)
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

var scannerCountTestsDelims = []struct {
	in  string
	len int
}{
	{`<* tag arg *>`, 1},
	{`<* tag arg *><* tag *>`, 2},
	{`<* tag arg *><* tag arg *><* tag *>`, 3},
	// {`<* tag *><* tag *>`, 2}, // Currently failing
	// {`<* tag arg *><* tag arg *><* tag *><* tag *>`, 4}, // Currently failing
	{`<< expr >>`, 1},
	{`<< expr arg >>`, 1},
	{`<< expr >><< expr >>`, 2},
	{`<< expr arg >><< expr arg >>`, 2},
}

func TestScan_delims(t *testing.T) {
	scan := func(src string) []Token { return Scan(src, SourceLoc{}, []byte{'<', '>', '*'}) }
	tokens := scan("12")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TextTokenType, tokens[0].Type)
	require.Equal(t, "12", tokens[0].Source)

	tokens = scan("<<obj>>")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, ObjTokenType, tokens[0].Type)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("<< obj >>")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, ObjTokenType, tokens[0].Type)
	require.Equal(t, "obj", tokens[0].Args)

	tokens = scan("<*tag args*>")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagTokenType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("<* tag args *>")
	require.NotNil(t, tokens)
	require.Len(t, tokens, 1)
	require.Equal(t, TagTokenType, tokens[0].Type)
	require.Equal(t, "tag", tokens[0].Name)
	require.Equal(t, "args", tokens[0].Args)

	tokens = scan("pre<* tag args *>mid<< object >>post")
	require.Equal(t, `[TextTokenType{"pre"} TagTokenType{Tag:"tag", Args:"args"} TextTokenType{"mid"} ObjTokenType{"object"} TextTokenType{"post"}]`, fmt.Sprint(tokens))

	for i, test := range scannerCountTestsDelims {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := scan(test.in)
			require.Len(t, tokens, test.len)
		})
	}
}
