//go:generate ragel -Z liquid.rl

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var tests = []struct{ in, out string }{
	{"pre</head>post", "pre:insertion:</head>post"},
	{"pre:insertion:</head>post", "pre:insertion:</head>post"},
	{"post", ":insertion:post"},
}

func TestLiquid(t *testing.T) {
	tokens := ScanChunks("pre{%if 1%}left{{x}}right{%endif%}post", "")
	// fmt.Println("tokens =", tokens)
	ast, err := Parse(tokens)
	require.NoError(t, err)
	fmt.Println("ast =", ast)
	err = ast.Render(os.Stdout, nil)
	require.NoError(t, err)
	fmt.Println()

	require.True(t, true)

	return
	for _, test := range chunkTests {
		tokens := ScanChunks(test.in, "")
		ast, err := Parse(tokens)
		require.NoError(t, err)
		actual := ast.Render(os.Stdout, nil)
		require.Equal(t, test.expected, actual)
	}
}

type chunkTest struct {
	in       string
	expected string
}

var chunkTests = []chunkTest{
	chunkTest{"{{var}}", "value"},
	chunkTest{"{{x}}", "1"},
}
