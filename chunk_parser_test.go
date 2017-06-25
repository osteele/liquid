package main

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChunkParser(t *testing.T) {
	ctx := Context{map[string]interface{}{
		"x": 123,
	},
	}

	for _, test := range chunkTests {
		t.Run(test.in, func(t *testing.T) {
			tokens := ScanChunks(test.in, "")
			// fmt.Println(tokens)
			ast, err := Parse(tokens)
			require.NoError(t, err)
			// fmt.Println(ast)
			buf := new(bytes.Buffer)
			err = ast.Render(buf, ctx)
			require.NoError(t, err)
			require.Equal(t, test.expected, buf.String())
		})
	}
}

var chunkTests = []struct{ in, expected string }{
	{"{{12}}", "12"},
	{"{{x}}", "123"},
	{"{%if 1%}}true{%endif%}", "true"},
	{"{%if x%}}true{%endif%}", "true"},
	{"{%if y%}}false{%endif%}", ""},
}
