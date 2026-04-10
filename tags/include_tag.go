package tags

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/render"
)

// includeTagArgs holds the parsed components of an include tag arguments.
type includeTagArgs struct {
	fileExpr  expressions.Expression
	withExpr  expressions.Expression
	withAlias string
	kvPairs   []kvPair
	forExpr   expressions.Expression // non-nil when 'for collection [as alias]' syntax is used
	forAlias  string                 // variable name bound to each item; defaults to file stem
}

type kvPair struct {
	key       string
	valueExpr expressions.Expression
}

// makeIncludeTag creates the include tag handler.
// Supports: basic filename, with variable [as alias], and key: value pairs.
func makeIncludeTag(_ *render.Config) func(string) (func(io.Writer, render.Context) error, error) {
	return func(source string) (func(io.Writer, render.Context) error, error) {
		args, err := parseIncludeArgs(source)
		if err != nil {
			return nil, err
		}

		return func(w io.Writer, ctx render.Context) error {
			fileVal, err := ctx.Evaluate(args.fileExpr)
			if err != nil {
				return err
			}

			rel, ok := fileVal.(string)
			if !ok {
				return ctx.Errorf("include requires a string argument; got %v", fileVal)
			}

			filename := filepath.Join(filepath.Dir(ctx.SourceFile()), rel)

			// Handle 'for collection [as alias]' — iterate the collection and
			// render the file once per item, sharing the outer scope.
			if args.forExpr != nil {
				collVal, err := ctx.Evaluate(args.forExpr)
				if err != nil {
					return err
				}

				alias := args.forAlias
				if alias == "" {
					base := filepath.Base(rel)
					alias = strings.TrimSuffix(base, filepath.Ext(base))
				}

				iter := makeIterator(collVal)
				if iter == nil {
					return nil
				}

				for i := 0; i < iter.Len(); i++ {
					extra := map[string]any{alias: iter.Index(i)}
					for _, kv := range args.kvPairs {
						val, err := ctx.Evaluate(kv.valueExpr)
						if err != nil {
							return err
						}
						extra[kv.key] = val
					}

					s, err := ctx.RenderFile(filename, extra)
					if err != nil {
						return err
					}

					if _, err = io.WriteString(w, s); err != nil {
						return err
					}
				}

				return nil
			}

			extra := map[string]any{}

			if args.withExpr != nil {
				withVal, err := ctx.Evaluate(args.withExpr)
				if err != nil {
					return err
				}

				alias := args.withAlias
				if alias == "" {
					base := filepath.Base(rel)
					alias = strings.TrimSuffix(base, filepath.Ext(base))
				}

				extra[alias] = withVal
			}

			for _, kv := range args.kvPairs {
				val, err := ctx.Evaluate(kv.valueExpr)
				if err != nil {
					return err
				}

				extra[kv.key] = val
			}

			s, err := ctx.RenderFile(filename, extra)
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, s)

			return err
		}, nil
	}
}

func parseIncludeArgs(source string) (*includeTagArgs, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return nil, fmt.Errorf("syntax error: include tag requires a filename")
	}

	fileExprStr, rest, err := consumeFirstExpression(source)
	if err != nil {
		return nil, fmt.Errorf("syntax error in include: %w", err)
	}

	fileExpr, err := expressions.Parse(fileExprStr)
	if err != nil {
		return nil, fmt.Errorf("syntax error in include filename: %w", err)
	}

	result := &includeTagArgs{fileExpr: fileExpr}
	rest = strings.TrimSpace(rest)

	// Handle 'for collection [as alias]' syntax.
	if strings.HasPrefix(rest, "for ") || rest == "for" {
		rest = strings.TrimSpace(rest[3:])

		forExprStr, afterFor, err := consumeFirstExpression(rest)
		if err != nil {
			return nil, fmt.Errorf("syntax error in include for clause: %w", err)
		}

		forExpr, err := expressions.Parse(forExprStr)
		if err != nil {
			return nil, fmt.Errorf("syntax error in include for expression: %w", err)
		}

		result.forExpr = forExpr
		rest = strings.TrimSpace(afterFor)

		if strings.HasPrefix(rest, "as ") || rest == "as" {
			aliasStr := strings.TrimSpace(rest[2:])
			aliasEnd := strings.IndexAny(aliasStr, " \t,")
			if aliasEnd < 0 {
				aliasEnd = len(aliasStr)
			}
			result.forAlias = aliasStr[:aliasEnd]
			rest = strings.TrimSpace(aliasStr[aliasEnd:])
		}

		if strings.HasPrefix(rest, ",") {
			rest = strings.TrimSpace(rest[1:])
		}

		kvPairs, err := parseKVPairs(rest)
		if err != nil {
			return nil, fmt.Errorf("syntax error in include key-value args: %w", err)
		}

		result.kvPairs = kvPairs

		return result, nil
	}

	if strings.HasPrefix(rest, "with ") || strings.HasPrefix(rest, "with\t") {
		rest = strings.TrimSpace(rest[4:])

		withExprStr, afterWith, err := consumeWithExpression(rest)
		if err != nil {
			return nil, fmt.Errorf("syntax error in include with clause: %w", err)
		}

		withExpr, err := expressions.Parse(withExprStr)
		if err != nil {
			return nil, fmt.Errorf("syntax error in include with variable: %w", err)
		}

		result.withExpr = withExpr
		rest = strings.TrimSpace(afterWith)

		if strings.HasPrefix(rest, "as ") || strings.HasPrefix(rest, "as\t") {
			rest = strings.TrimSpace(rest[2:])
			aliasEnd := strings.IndexAny(rest, " \t,")
			if aliasEnd < 0 {
				aliasEnd = len(rest)
			}

			result.withAlias = rest[:aliasEnd]
			rest = strings.TrimSpace(rest[aliasEnd:])
		}

		if strings.HasPrefix(rest, ",") {
			rest = strings.TrimSpace(rest[1:])
		}
	} else if strings.HasPrefix(rest, ",") {
		rest = strings.TrimSpace(rest[1:])
	}

	kvPairs, err := parseKVPairs(rest)
	if err != nil {
		return nil, fmt.Errorf("syntax error in include key-value args: %w", err)
	}

	result.kvPairs = kvPairs

	return result, nil
}

func consumeFirstExpression(s string) (string, string, error) {
	if len(s) == 0 {
		return "", "", fmt.Errorf("expected an expression")
	}

	if s[0] == '"' || s[0] == '\'' {
		quote := s[0]
		i := 1

		for i < len(s) {
			if s[i] == '\\' {
				i += 2
				continue
			}

			if s[i] == quote {
				i++
				return s[:i], s[i:], nil
			}

			i++
		}

		return "", "", fmt.Errorf("unterminated string literal in include tag")
	}

	depth := 0

	for i, ch := range s {
		switch ch {
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ' ', '\t':
			if depth == 0 {
				return s[:i], s[i:], nil
			}
		case ',':
			if depth == 0 {
				return s[:i], s[i:], nil
			}
		}
	}

	return s, "", nil
}

func consumeWithExpression(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		if (s[i] == ' ' || s[i] == '\t') && i+2 < len(s) {
			trimmed := strings.TrimLeft(s[i:], " \t")
			if trimmed == "as" || strings.HasPrefix(trimmed, "as ") || strings.HasPrefix(trimmed, "as\t") {
				return s[:i], s[i:], nil
			}
		}

		if s[i] == ',' {
			return s[:i], s[i:], nil
		}
	}

	return s, "", nil
}

func parseKVPairs(s string) ([]kvPair, error) {
	var pairs []kvPair

	for {
		s = strings.TrimSpace(s)
		if s == "" {
			break
		}

		colonIdx := strings.Index(s, ":")
		if colonIdx < 0 {
			return nil, fmt.Errorf("expected key: value pair, got %q", s)
		}

		key := strings.TrimSpace(s[:colonIdx])
		if key == "" || strings.ContainsAny(key, " \t()+=-*/<>!") {
			return nil, fmt.Errorf("invalid key %q in include arguments", key)
		}

		s = strings.TrimSpace(s[colonIdx+1:])
		valEnd := findValueExpressionEnd(s)
		valStr := strings.TrimSpace(s[:valEnd])
		s = strings.TrimSpace(s[valEnd:])

		if s != "" && s[0] == ',' {
			s = s[1:]
		}

		valExpr, err := expressions.Parse(valStr)
		if err != nil {
			return nil, fmt.Errorf("invalid value for key %q: %w", key, err)
		}

		pairs = append(pairs, kvPair{key: key, valueExpr: valExpr})
	}

	return pairs, nil
}

func findValueExpressionEnd(s string) int {
	depth := 0
	inStr := false
	strChar := byte(0)

	for i := 0; i < len(s); i++ {
		c := s[i]

		if inStr {
			if c == '\\' {
				i++
				continue
			}

			if c == strChar {
				inStr = false
			}

			continue
		}

		switch c {
		case '"', '\'':
			inStr = true
			strChar = c
		case '(', '[':
			depth++
		case ')', ']':
			depth--
		case ',':
			if depth == 0 {
				return i
			}
		}
	}

	return len(s)
}
