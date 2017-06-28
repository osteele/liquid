package chunks

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func init() {
	DefineControlTag("case").Branch("when")
	DefineControlTag("comment")
	DefineControlTag("for").Governs([]string{"break"})
	DefineControlTag("if").Branch("else").Branch("elsif")
	DefineControlTag("raw")
}

var parseErrorTests = []struct{ in, expected string }{
	{"{%unknown_tag%}", "unknown tag"},
	{"{%if test%}", "unterminated if tag"},
	// {"{%if syntax error%}{%endif%}", "parse error"},
}

var parserTests = []struct{ in string }{
	{`{% for item in list %}{% endfor %}`},
	{`{% if test %}{% else %}{% endif %}`},
	{`{% if test %}{% if test %}{% endif %}{% endif %}`},
	{`{% for item in list %}{% if test %}{% else %}{% endif x %}{% endfor %}`},
}

func TestParseErrors(t *testing.T) {
	for i, test := range parseErrorTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := Scan(test.in, "")
			ast, err := Parse(tokens)
			require.Nilf(t, ast, test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}

func TestParser(t *testing.T) {
	for i, test := range parserTests {
		t.Run(fmt.Sprintf("%02d", i), func(t *testing.T) {
			tokens := Scan(test.in, "")
			_, err := Parse(tokens)
			require.NoError(t, err, test.in)
			// require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
