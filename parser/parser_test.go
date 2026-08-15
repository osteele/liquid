package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	grammarFake     struct{}
	blockSyntaxFake string
)

func (g grammarFake) BlockSyntax(w string) (BlockSyntax, bool) {
	switch {
	case w == "for" || w == "if" || w == "unless" || w == "else" ||
		w == "raw" || w == "comment" || strings.HasPrefix(w, "end"):
		return blockSyntaxFake(w), true
	default:
		return nil, false
	}
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

func TestNewConfig(t *testing.T) {
	g := grammarFake{}
	cfg := NewConfig(g)
	require.Equal(t, grammarFake{}, cfg.Grammar)
}

func TestSourceError(t *testing.T) {
	loc := SourceLoc{Pathname: "test.html", LineNo: 5}
	token := Token{
		SourceLoc: loc,
		Source:    "{% bad %}",
	}

	err := Errorf(&token, "something went wrong")

	// Error() formatting
	require.Contains(t, err.Error(), "line 5")
	require.Contains(t, err.Error(), "something went wrong")
	require.Contains(t, err.Error(), "test.html")

	// Path()
	require.Equal(t, "test.html", err.Path())

	// LineNumber()
	require.Equal(t, 5, err.LineNumber())

	// Cause() is nil for directly created errors
	require.Nil(t, err.Cause())
}

func TestSourceError_no_line(t *testing.T) {
	loc := SourceLoc{Pathname: "test.html"}
	token := Token{SourceLoc: loc, Source: "{% x %}"}
	err := Errorf(&token, "msg")
	// no "(line 0)" in output when LineNo is 0
	require.NotContains(t, err.Error(), "line 0")
}

func TestSourceError_no_path(t *testing.T) {
	token := Token{SourceLoc: SourceLoc{LineNo: 3}, Source: "{% x %}"}
	err := Errorf(&token, "msg")
	// uses source context when no pathname
	require.Contains(t, err.Error(), "{% x %}")
}

func TestWrapError(t *testing.T) {
	token := Token{SourceLoc: SourceLoc{Pathname: "f.html", LineNo: 2}, Source: "{% x %}"}

	t.Run("nil input", func(t *testing.T) {
		require.Nil(t, WrapError(nil, &token))
	})

	t.Run("non-Error input", func(t *testing.T) {
		wrapped := WrapError(fmt.Errorf("raw error"), &token)
		require.NotNil(t, wrapped)
		require.Contains(t, wrapped.Error(), "raw error")
		require.Equal(t, "f.html", wrapped.Path())
		require.Equal(t, 2, wrapped.LineNumber())
		require.NotNil(t, wrapped.Cause())
	})

	t.Run("Error input with path", func(t *testing.T) {
		inner := Errorf(&token, "inner error")
		wrapped := WrapError(inner, &token)
		// should return the same error since it already has a path
		require.Contains(t, wrapped.Error(), "inner error")
		require.Equal(t, "f.html", wrapped.Path())
	})

	t.Run("Error input without path re-wraps", func(t *testing.T) {
		noPathToken := Token{SourceLoc: SourceLoc{}, Source: "{% y %}"}
		inner := Errorf(&noPathToken, "inner")
		// wrap with a locatable that has path info
		wrapped := WrapError(inner, &token)
		require.Equal(t, "f.html", wrapped.Path())
	})
}

func TestToken_IsZero(t *testing.T) {
	require.True(t, SourceLoc{}.IsZero())
	require.False(t, (SourceLoc{Pathname: "f.html"}).IsZero())
	require.False(t, (SourceLoc{LineNo: 1}).IsZero())
	require.False(t, (SourceLoc{Pathname: "f.html", LineNo: 1}).IsZero())
}

func TestSourceLoc_String(t *testing.T) {
	require.Equal(t, "f.html:5", (SourceLoc{Pathname: "f.html", LineNo: 5}).String())
	require.Equal(t, "line 3", (SourceLoc{LineNo: 3}).String())
}

// parserNestingStateTests models the parser's internal state machine.
// Each row is a token sequence and whether it should parse or produce an error.
var parserNestingStateTests = []struct {
	name        string
	template    string
	wantErr     bool
	errContains string
}{
	// Valid base states
	{"empty template", "", false, ""},
	{"plain text", "hello world", false, ""},
	{"object expression", "{{ x }}", false, ""},
	{"unknown tag", "{% other %}", false, ""},

	// Valid block nesting
	{"simple for", "{% for item in list %}{% endfor %}", false, ""},
	{"simple if", "{% if test %}{% endif %}", false, ""},
	{"if with else", "{% if test %}{% else %}{% endif %}", false, ""},
	{"nested if", "{% if test %}{% if test %}{% endif %}{% endif %}", false, ""},
	{"for containing if", "{% for item in list %}{% if test %}{% endif %}{% endfor %}", false, ""},
	{"unless", "{% unless test %}{% endunless %}", false, ""},

	// Raw / comment swallow content
	{"raw swallows tag", "{% raw %}{% if true %}{% endraw %}", false, ""},
	{"raw swallows object", "{% raw %}{{ x }}{% endraw %}", false, ""},
	{"comment swallows tag", "{% comment %}{% if true %}{% endcomment %}", false, ""},
	{"comment swallows raw", "{% comment %}{% raw %}{% endcomment %}", false, ""},

	// Invalid states: clauses / ends without matching start
	{"else outside if", "{% else %}", true, "not inside"},
	{"endfor without for", "{% endfor %}", true, "not inside"},
	{"endif without if", "{% endif %}", true, "not inside"},

	// Invalid states: mismatched end tag
	{"wrong end tag", "{% if test %}{% endfor %}", true, "not inside"},
	{"nested wrong end tag", "{% for item in list %}{% if test %}{% endfor %}{% endif %}", true, "not inside"},

	// Invalid states: unterminated blocks
	{"unterminated if", "{% if test %}", true, "unterminated"},
	{"unterminated for", "{% for item in list %}", true, "unterminated"},

	// Invalid states: clause in wrong block
	{"else inside for", "{% for item in list %}{% else %}{% endfor %}", true, "not inside"},
}

func TestParserNestingState(t *testing.T) {
	cfg := Config{Grammar: grammarFake{}}

	for _, test := range parserNestingStateTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := cfg.Parse(test.template, SourceLoc{})
			if test.wantErr {
				require.Errorf(t, err, test.template)
				require.Containsf(t, err.Error(), test.errContains, test.template)
			} else {
				require.NoErrorf(t, err, test.template)
			}
		})
	}
}
