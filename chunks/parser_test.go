package chunks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func addTestTags(s Settings) {
	s.AddStartTag("case").Branch("when")
	s.AddStartTag("comment")
	s.AddStartTag("for").Governs([]string{"break"})
	s.AddStartTag("if").Branch("else").Branch("elsif")
	s.AddStartTag("raw")
}

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	{"{%if test%}", "unterminated if tag"},
	// {"{%for syntax error%}{%endfor%}", "parse error"},
}

var parserTests = []struct{ in string }{
	{`{% for item in list %}{% endfor %}`},
	{`{% if test %}{% else %}{% endif %}`},
	{`{% if test %}{% if test %}{% endif %}{% endif %}`},
	{`{% for item in list %}{% if test %}{% else %}{% endif x %}{% endfor %}`},
	{`{% if true %}{% raw %}{% endraw %}{% endif %}`},
}

func TestParseErrors(t *testing.T) {
	settings := NewSettings()
	addTestTags(settings)
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			tokens := Scan(test.in, "")
			ast, err := settings.Parse(tokens)
			require.Nilf(t, ast, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestParser(t *testing.T) {
	settings := NewSettings()
	addTestTags(settings)
	for i, test := range parserTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			tokens := Scan(test.in, "")
			_, err := settings.Parse(tokens)
			require.NoError(t, err, test.in)
			// require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
