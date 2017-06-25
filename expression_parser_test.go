//go:generate ragel -Z scanner.rl

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// var lexerTests = []struct{}{
// 	{"{{var}}", "value"},
// 	{"{{x}}", "1"},
// }

func TestExpressionParser(t *testing.T) {
	tokens, err := ScanExpression("abc > 123")
	require.NoError(t, err)
	fmt.Println("tokens =", tokens)
	// ast, err := Parse(tokens)
	// require.NoError(t, err)
	// fmt.Println("ast =", ast)
	// err = ast.Render(os.Stdout, nil)
	// require.NoError(t, err)
	// fmt.Println()
	return

	for _, test := range chunkTests {
		tokens := ScanChunks(test.in, "")
		ast, err := Parse(tokens)
		require.NoError(t, err)
		actual := ast.Render(os.Stdout, nil)
		require.Equal(t, test.expected, actual)
	}
}
