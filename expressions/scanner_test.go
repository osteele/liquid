package expressions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testSymbol struct {
	tok int
	typ yySymType
}

func (s testSymbol) String() string {
	return fmt.Sprintf("%d:%v", s.tok, s.typ)
}

func scanExpression(data string) []testSymbol {
	var (
		lex     = newLexer([]byte(data))
		symbols []testSymbol
	)
	for {
		var s yySymType
		tok := lex.Lex(&s)
		if tok == 0 {
			break
		}

		symbols = append(symbols, testSymbol{tok, s})
	}

	return symbols
}

func TestLex(t *testing.T) {
	ts := scanExpression("abc > 123")
	require.Len(t, ts, 3)
	require.Equal(t, IDENTIFIER, ts[0].tok)
	require.Equal(t, "abc", ts[0].typ.name)
	require.Equal(t, LITERAL, ts[2].tok)
	require.Equal(t, 123, ts[2].typ.val)

	// verify these don't match "for", "or", or "false"
	ts = scanExpression("forage")
	require.Len(t, ts, 1)
	ts = scanExpression("orange")
	require.Len(t, ts, 1)
	ts = scanExpression("falsehood")
	require.Len(t, ts, 1)

	ts = scanExpression("a.b-c")
	require.Len(t, ts, 2)
	require.Equal(t, PROPERTY, ts[1].tok)
	require.Equal(t, "b-c", ts[1].typ.name)

	// literals
	ts = scanExpression(`true false nil 2 2.3 "abc" 'abc'`)
	require.Len(t, ts, 7)
	require.Equal(t, LITERAL, ts[0].tok)
	require.Equal(t, LITERAL, ts[1].tok)
	require.Equal(t, LITERAL, ts[2].tok)
	require.Equal(t, LITERAL, ts[3].tok)
	require.Equal(t, LITERAL, ts[4].tok)
	require.Equal(t, LITERAL, ts[5].tok)
	require.Equal(t, LITERAL, ts[6].tok)
	require.Equal(t, true, ts[0].typ.val)
	require.Equal(t, false, ts[1].typ.val)
	require.Nil(t, ts[2].typ.val)
	require.Equal(t, 2, ts[3].typ.val)
	//nolint:testifylint
	require.Equal(t, 2.3, ts[4].typ.val)
	require.Equal(t, "abc", ts[5].typ.val)
	require.Equal(t, "abc", ts[6].typ.val)

	// identifiers
	ts = scanExpression(`abc ab_c ab-c abc?`)
	require.Len(t, ts, 4)
	require.Equal(t, IDENTIFIER, ts[0].tok)
	require.Equal(t, IDENTIFIER, ts[1].tok)
	require.Equal(t, IDENTIFIER, ts[2].tok)
	require.Equal(t, IDENTIFIER, ts[3].tok)
	require.Equal(t, "abc", ts[0].typ.name)
	require.Equal(t, "ab_c", ts[1].typ.name)
	require.Equal(t, "ab-c", ts[2].typ.name)
	require.Equal(t, "abc?", ts[3].typ.name)

	ts = scanExpression(`{%cycle 'a', 'b'`)
	require.Len(t, ts, 4)

	ts = scanExpression(`%loop i in (3 .. 5)`)
	require.Len(t, ts, 8)

	// ts= scanExpression(`%loop i in (3..5)`)
	// require.Len(t, ts, 9)
}

func TestLexStringEscapes(t *testing.T) {
	// Double-quoted strings support escape sequences
	t.Run("backslash", func(t *testing.T) {
		ts := scanExpression(`"a\\b"`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, `a\b`, ts[0].typ.val)
	})
	t.Run("escaped_quote", func(t *testing.T) {
		ts := scanExpression(`"say \"hello\""`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, `say "hello"`, ts[0].typ.val)
	})
	t.Run("newline", func(t *testing.T) {
		ts := scanExpression(`"line1\nline2"`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, "line1\nline2", ts[0].typ.val)
	})
	t.Run("tab", func(t *testing.T) {
		ts := scanExpression(`"col1\tcol2"`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, "col1\tcol2", ts[0].typ.val)
	})
	t.Run("carriage_return", func(t *testing.T) {
		ts := scanExpression(`"a\rb"`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, "a\rb", ts[0].typ.val)
	})
	t.Run("unknown_escape_passthrough", func(t *testing.T) {
		ts := scanExpression(`"a\xb"`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, `a\xb`, ts[0].typ.val)
	})
	// Single-quoted strings do NOT process escapes
	t.Run("single_quote_no_escape", func(t *testing.T) {
		ts := scanExpression(`'a\nb'`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, `a\nb`, ts[0].typ.val)
	})
	t.Run("no_escapes_plain", func(t *testing.T) {
		ts := scanExpression(`"hello"`)
		require.Len(t, ts, 1)
		require.Equal(t, LITERAL, ts[0].tok)
		require.Equal(t, "hello", ts[0].typ.val)
	})
}

func TestLexUnicodeIdentifiers(t *testing.T) {
	// Test Bengali
	t.Run("Bengali", func(t *testing.T) {
		ts := scanExpression("অসম্ভব == 'impossible'")
		require.Len(t, ts, 3)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "অসম্ভব", ts[0].typ.name)
		require.Equal(t, EQ, ts[1].tok)
		require.Equal(t, LITERAL, ts[2].tok)
		require.Equal(t, "impossible", ts[2].typ.val)
	})

	// Test Chinese
	t.Run("Chinese", func(t *testing.T) {
		ts := scanExpression("用户.姓名 != nil")
		require.Len(t, ts, 4)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "用户", ts[0].typ.name)
		require.Equal(t, PROPERTY, ts[1].tok)
		require.Equal(t, "姓名", ts[1].typ.name)
		require.Equal(t, NEQ, ts[2].tok)
		require.Equal(t, LITERAL, ts[3].tok)
	})

	// Test Japanese
	t.Run("Japanese", func(t *testing.T) {
		ts := scanExpression("ユーザー.名前 contains 'test'")
		require.Len(t, ts, 4)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "ユーザー", ts[0].typ.name)
		require.Equal(t, PROPERTY, ts[1].tok)
		require.Equal(t, "名前", ts[1].typ.name)
		require.Equal(t, CONTAINS, ts[2].tok)
	})

	// Test Arabic
	t.Run("Arabic", func(t *testing.T) {
		ts := scanExpression("المستخدم.العمر >= 18")
		require.Len(t, ts, 4)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "المستخدم", ts[0].typ.name)
		require.Equal(t, PROPERTY, ts[1].tok)
		require.Equal(t, "العمر", ts[1].typ.name)
		require.Equal(t, GE, ts[2].tok)
		require.Equal(t, LITERAL, ts[3].tok)
	})

	// Test Cyrillic
	t.Run("Cyrillic", func(t *testing.T) {
		ts := scanExpression("пользователь.возраст <= 21")
		require.Len(t, ts, 4)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "пользователь", ts[0].typ.name)
		require.Equal(t, PROPERTY, ts[1].tok)
		require.Equal(t, "возраст", ts[1].typ.name)
		require.Equal(t, LE, ts[2].tok)
	})

	// Test Mixed scripts
	t.Run("MixedScripts", func(t *testing.T) {
		ts := scanExpression("user_用户.名前-属性?")
		require.Len(t, ts, 2)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "user_用户", ts[0].typ.name)
		require.Equal(t, PROPERTY, ts[1].tok)
		require.Equal(t, "名前-属性?", ts[1].typ.name)
	})

	// Test Backward compatibility with ASCII
	t.Run("ASCIIBackwardCompatible", func(t *testing.T) {
		ts := scanExpression("abc ab_c ab-c abc? user_name")
		require.Len(t, ts, 5)
		for i, expected := range []string{"abc", "ab_c", "ab-c", "abc?", "user_name"} {
			require.Equal(t, IDENTIFIER, ts[i].tok)
			require.Equal(t, expected, ts[i].typ.name)
		}
	})

	// Test Edge cases
	t.Run("EdgeCases", func(t *testing.T) {
		// Combining characters (should work if in NFC form)
		ts := scanExpression("café") // precomposed é
		require.Len(t, ts, 1)
		require.Equal(t, IDENTIFIER, ts[0].tok)
	})

	// Test Keywords with Unicode (should NOT match)
	t.Run("UnicodeNotConfusedWithKeywords", func(t *testing.T) {
		// These should be identifiers, not keywords
		testCases := []struct {
			expr     string
			expected string
		}{
			{"іn", "іn"},       // Cyrillic 'і', not ASCII 'i'
			{"аnd", "аnd"},     // Cyrillic 'а', not ASCII 'a'
			{"оr", "оr"},       // Cyrillic 'о', not ASCII 'o'
			{"truе", "truе"},   // Cyrillic 'е', not ASCII 'e'
			{"falsе", "falsе"}, // Cyrillic 'е', not ASCII 'e'
		}

		for _, tc := range testCases {
			t.Run(tc.expr, func(t *testing.T) {
				ts := scanExpression(tc.expr)
				require.Len(t, ts, 1)
				require.Equal(t, IDENTIFIER, ts[0].tok)
				require.Equal(t, tc.expected, ts[0].typ.name)
			})
		}
	})

	// Test Complex expressions with Unicode
	t.Run("ComplexUnicodeExpressions", func(t *testing.T) {
		ts := scanExpression("用户.年龄 >= 18 and 用户.国家 == '日本' or 用户.활동?")
		require.Len(t, ts, 12)
		require.Equal(t, IDENTIFIER, ts[0].tok)
		require.Equal(t, "用户", ts[0].typ.name)
		require.Equal(t, PROPERTY, ts[1].tok)
		require.Equal(t, "年龄", ts[1].typ.name)
		require.Equal(t, GE, ts[2].tok)
		require.Equal(t, LITERAL, ts[3].tok)
		require.Equal(t, AND, ts[4].tok)
		require.Equal(t, IDENTIFIER, ts[5].tok)
		require.Equal(t, "用户", ts[5].typ.name)
		require.Equal(t, PROPERTY, ts[6].tok)
		require.Equal(t, "国家", ts[6].typ.name)
		require.Equal(t, EQ, ts[7].tok)
		require.Equal(t, LITERAL, ts[8].tok)
		require.Equal(t, OR, ts[9].tok)
		require.Equal(t, IDENTIFIER, ts[10].tok)
		require.Equal(t, "用户", ts[10].typ.name)
		require.Equal(t, PROPERTY, ts[11].tok)
		require.Equal(t, "활동?", ts[11].typ.name)
	})

	// Test Invalid Unicode identifiers
	t.Run("InvalidUnicodeIdentifiers", func(t *testing.T) {
		// These should cause errors/panic
		// emojis are not a valid letter or digits or neither a mark
		invalidCases := []string{
			"🚀_speed", // starts with emoji
			"fast🚀?",  // contains emoji
		}

		for _, expr := range invalidCases {
			t.Run(expr, func(t *testing.T) {
				require.Panics(t, func() {
					scanExpression(expr)
				})
			})
		}
	})
}
