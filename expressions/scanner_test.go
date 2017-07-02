//go:generate ragel -Z scanner.rl
//go:generate goyacc expressions.y

package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// var lexerTests = []struct{}{
// 	{"{{var}}", "value"},
// 	{"{{x}}", "1"},
// }

func scanExpression(data string) ([]yySymType, error) {
	l := newLexer([]byte(data))
	var symbols []yySymType
	var s yySymType
	for {
		t := l.Lex(&s)
		if t == 0 {
			break
		}
		symbols = append(symbols, s)
	}
	return symbols, nil
}

func TestExpressionScanner(t *testing.T) {
	tokens, err := scanExpression("abc > 123")
	require.NoError(t, err)
	require.Len(t, tokens, 3)

	tokens, _ = scanExpression("forage")
	require.Len(t, tokens, 1)

	tokens, _ = scanExpression("orange")
	require.Len(t, tokens, 1)
}
