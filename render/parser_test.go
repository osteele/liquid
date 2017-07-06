package render

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func addParserTestTags(s Config) {
	s.AddBlock("case").Branch("when")
	s.AddBlock("comment")
	s.AddBlock("for").Governs([]string{"break"})
	s.AddBlock("if").Branch("else").Branch("elsif")
	s.AddBlock("unless").SameSyntaxAs("if")
	s.AddBlock("raw")
	s.AddBlock("error").Parser(func(c ASTBlock) (func(io.Writer, Context) error, error) {
		return nil, fmt.Errorf("stage 1 error")
	})
}

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	{"{%if test%}", "unterminated if tag"},
	{"{%if test%}{% endunless %}", "not inside unless"},
	{`{% error %}{% enderror %}`, "stage 1 error"},
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
	settings := NewConfig()
	addParserTestTags(settings)
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := settings.Compile(test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestParser(t *testing.T) {
	settings := NewConfig()
	addParserTestTags(settings)
	for i, test := range parserTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := settings.Compile(test.in)
			require.NoError(t, err, test.in)
			// require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
