package render

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func addCompilerTestTags(s Config) {
	s.AddBlock("block").Compiler(func(c BlockNode) (func(io.Writer, Context) error, error) {
		return nil, fmt.Errorf("block compiler error")
	})
}

var compilerErrorTests = []struct{ in, expected string }{
	{`{% unknown_tag %}`, "unknown tag"},
	{`{% block %}{% endblock %}`, "block compiler error"},
	// {`{% tag %}`, "tag compiler error"},
	// {`{%for syntax error%}{%endfor%}`, "parse error"},
}

func TestCompileErrors(t *testing.T) {
	settings := NewConfig()
	addCompilerTestTags(settings)
	for i, test := range compilerErrorTests {
		t.Run(fmt.Sprintf("%02d", i+1), func(t *testing.T) {
			_, err := settings.Compile(test.in)
			require.Errorf(t, err, test.in)
			require.Containsf(t, err.Error(), test.expected, test.in)
		})
	}
}
