package render

import (
	"fmt"
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
}

var parseErrorTests = []struct{ in, expected string }{
	{"{% if test %}", "unterminated if block"},
	{"{% if test %}{% endunless %}", "not inside unless"},
	// TODO tag syntax could specify statement type to catch these in parser
	// {"{{ syntax error }}", "parse error"},
	// {"{% for syntax error %}{% endfor %}", "parse error"},
}

var parserTests = []struct{ in string }{
	{`{% for item in list %}{% endfor %}`},
	{`{% if test %}{% else %}{% endif %}`},
	{`{% if test %}{% if test %}{% endif %}{% endif %}`},
	{`{% unless test %}{% else %}{% endunless %}`},
	{`{% for item in list %}{% if test %}{% else %}{% endif %}{% endfor %}`},
	{`{% if true %}{% raw %}{% endraw %}{% endif %}`},

	{`{% comment %}{% if true %}{% endcomment %}`},
	{`{% raw %}{% if true %}{% endraw %}`},
}

func TestParseErrors(t *testing.T) {
	settings := NewConfig()
	addParserTestTags(settings)
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := settings.Parse(test.in)
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
			_, err := settings.Parse(test.in)
			require.NoError(t, err, test.in)
		})
	}
}
