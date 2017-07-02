package chunks

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func addParserTestTags(s Settings) {
	s.AddStartTag("case").Branch("when")
	s.AddStartTag("comment")
	s.AddStartTag("for").Governs([]string{"break"})
	s.AddStartTag("if").Branch("else").Branch("elsif")
	s.AddStartTag("unless").SameSyntaxAs("if")
	s.AddStartTag("raw")
	s.AddStartTag("err1").Parser(func(c ASTControlTag) (func(io.Writer, RenderContext) error, error) {
		return nil, fmt.Errorf("stage 1 error")
	})
}

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	{"{%if test%}", "unterminated if tag"},
	{"{%if test%}{% endunless %}", "not inside unless"},
	{`{% err1 %}{% enderr1 %}`, "stage 1 error"},
	// {"{%for syntax error%}{%endfor%}", "parse error"},
}

var parserTests = []struct{ in string }{
	{`{% for item in list %}{% endfor %}`},
	{`{% if test %}{% else %}{% endif %}`},
	{`{% if test %}{% if test %}{% endif %}{% endif %}`},
	{`{% unless test %}{% else %}{% endunless %}`},
	{`{% for item in list %}{% if test %}{% else %}{% endif x %}{% endfor %}`},
	{`{% if true %}{% raw %}{% endraw %}{% endif %}`},
}

func TestParseErrors(t *testing.T) {
	settings := NewSettings()
	addParserTestTags(settings)
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			ast, err := settings.Parse(test.in)
			require.Nilf(t, ast, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestParser(t *testing.T) {
	settings := NewSettings()
	addParserTestTags(settings)
	for i, test := range parserTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := settings.Parse(test.in)
			require.NoError(t, err, test.in)
			// require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
