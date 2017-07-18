package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type grammarFake struct{}
type blockSyntaxFake string

func (g grammarFake) BlockSyntax(w string) (BlockSyntax, bool) {
	return blockSyntaxFake(w), true
}

func (g blockSyntaxFake) IsBlock() bool { return true }
func (g blockSyntaxFake) CanHaveParent(p BlockSyntax) bool {
	return string(g) == "end"+p.TagName() || (g == "else" && p.TagName() == "if")
}
func (g blockSyntaxFake) IsBlockEnd() bool { return strings.HasPrefix(string(g), "end") }
func (g blockSyntaxFake) IsBlockStart() bool {
	return g == "for" || g == "if" || g == "unless"
}
func (g blockSyntaxFake) IsClause() bool       { return g == "else" }
func (g blockSyntaxFake) ParentTags() []string { return []string{"unless"} }
func (g blockSyntaxFake) RequiresParent() bool { return g == "else" || g.IsBlockEnd() }
func (g blockSyntaxFake) TagName() string      { return string(g) }

var parseErrorTests = []struct{ in, expected string }{
	{"{% if test %}", `unterminated "if" block`},
	{"{% if test %}{% endunless %}", "not inside unless"},
	// TODO tag syntax could specify statement type to catch these in parser
	// {"{{ syntax error }}", "syntax error"},
	// {"{% for syntax error %}{% endfor %}", "syntax error"},
}

var parserTests = []struct{ in string }{
	{`{% for item in list %}{% endfor %}`},
	{`{% if test %}{% else %}{% endif %}`},
	{`{% if test %}{% if test %}{% endif %}{% endif %}`},
	{`{% unless test %}{% endunless %}`},
	{`{% for item in list %}{% if test %}{% else %}{% endif %}{% endfor %}`},
	{`{% if true %}{% raw %}{% endraw %}{% endif %}`},

	{`{% comment %}{% if true %}{% endcomment %}`},
	{`{% raw %}{% if true %}{% endraw %}`},
}

func TestParseErrors(t *testing.T) {
	cfg := Config{Grammar: grammarFake{}}
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := cfg.Parse(test.in, SourceLoc{})
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestParser(t *testing.T) {
	cfg := Config{Grammar: grammarFake{}}
	for i, test := range parserTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := cfg.Parse(test.in, SourceLoc{})
			require.NoError(t, err, test.in)
		})
	}
}
