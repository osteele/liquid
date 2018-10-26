package render

import (
	"fmt"
	"io"
	"testing"

	"github.com/urbn8/liquid/parser"
	"github.com/stretchr/testify/require"
)

func addCompilerTestTags(s Config) {
	s.AddBlock("block").Compiler(func(c BlockNode) (func(io.Writer, Context) error, error) {
		return func(io.Writer, Context) error {
			return nil
		}, nil
	})
	s.AddBlock("error_block").Compiler(func(c BlockNode) (func(io.Writer, Context) error, error) {
		return nil, fmt.Errorf("block compiler error")
	})
}

var compilerErrorTests = []struct{ in, expected string }{
	{`{% undefined_tag %}`, "undefined tag"},
	{`{% error_block %}{% enderror_block %}`, "block compiler error"},
	{`{% block %}{% undefined_tag %}{% endblock %}`, "undefined tag"},
	// {`{% tag %}`, "tag compiler error"},
	// {`{%for syntax error%}{%endfor%}`, "syntax error"},
}

func TestCompile_errors(t *testing.T) {
	settings := NewConfig()
	addCompilerTestTags(settings)
	for i, test := range compilerErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := settings.Compile(test.in, parser.SourceLoc{})
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
