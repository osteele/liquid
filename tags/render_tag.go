package tags

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/osteele/liquid/render"
)

// makeRenderTag creates the render tag handler.
// The render tag renders a snippet in an isolated variable scope.
// Supports:
//
//	{%  render 'file' %}
//	{%  render 'file', key: value, ... %}
//	{%  render 'file' with variable %}
//	{%  render 'file' with variable as alias %}
//	{%  render 'file' for collection as item %}  (each item rendered in isolation)
func makeRenderTag(cfg *render.Config) func(string) (func(io.Writer, render.Context) error, error) {
	return func(source string) (func(io.Writer, render.Context) error, error) {
		source = strings.TrimSpace(source)

		// Detect 'for collection' syntax
		isForLoop := false
		var forExprStr, forAlias string

		// Parse the filename first
		fileExprStr, rest, err := consumeFirstExpression(source)
		if err != nil {
			return nil, fmt.Errorf("syntax error in render: %w", err)
		}
		rest = strings.TrimSpace(rest)

		// Check for 'for collection [as item]'
		if strings.HasPrefix(rest, "for ") || strings.HasPrefix(rest, "for") {
			isForLoop = true
			rest = strings.TrimSpace(rest[3:])
			// Parse collection expression
			forExprStr, rest, err = consumeFirstExpression(rest)
			if err != nil {
				return nil, fmt.Errorf("syntax error in render for: %w", err)
			}
			rest = strings.TrimSpace(rest)
			if strings.HasPrefix(rest, "as ") || strings.HasPrefix(rest, "as") {
				rest = strings.TrimSpace(rest[2:])
				aliasEnd := strings.IndexAny(rest, " ,")
				if aliasEnd < 0 {
					aliasEnd = len(rest)
				}
				forAlias = rest[:aliasEnd]
				rest = strings.TrimSpace(rest[aliasEnd:])
			} else {
				forAlias = "item"
			}
			if strings.HasPrefix(rest, ",") {
				rest = strings.TrimSpace(rest[1:])
			}
		}

		// Parse the rest as include-style args (with/as/key:val)
		// Reuse parseIncludeArgs by building a synthetic source string
		synth := fileExprStr + " " + rest
		args, err := parseIncludeArgs(synth)
		if err != nil {
			return nil, fmt.Errorf("syntax error in render arguments: %w", err)
		}

		return func(w io.Writer, ctx render.Context) error {
			fileVal, err := ctx.Evaluate(args.fileExpr)
			if err != nil {
				return err
			}

			rel, ok := fileVal.(string)
			if !ok {
				return ctx.Errorf("render requires a string argument; got %v", fileVal)
			}

			filename := filepath.Join(filepath.Dir(ctx.SourceFile()), rel)

			if isForLoop {
				// render for collection: render once per item in isolation
				collExpr, err2 := ctx.EvaluateString(forExprStr)
				if err2 != nil {
					return ctx.WrapError(err2)
				}
				iter := makeIterator(collExpr)
				if iter == nil {
					return nil
				}
				alias := forAlias
				if alias == "" {
					alias = strings.TrimSuffix(filepath.Base(rel), filepath.Ext(rel))
				}
				for i := 0; i < iter.Len(); i++ {
					bindings := map[string]any{alias: iter.Index(i)}
					for _, kv := range args.kvPairs {
						val, err3 := ctx.Evaluate(kv.valueExpr)
						if err3 != nil {
							return err3
						}
						bindings[kv.key] = val
					}
					s, err3 := ctx.RenderFileIsolated(filename, bindings)
					if err3 != nil {
						return err3
					}
					if _, err3 = io.WriteString(w, s); err3 != nil {
						return err3
					}
				}
				return nil
			}

			// Standard render: isolated scope with optional extra bindings
			bindings := map[string]any{}

			if args.withExpr != nil {
				withVal, err2 := ctx.Evaluate(args.withExpr)
				if err2 != nil {
					return err2
				}
				alias := args.withAlias
				if alias == "" {
					base := filepath.Base(rel)
					alias = strings.TrimSuffix(base, filepath.Ext(base))
				}
				bindings[alias] = withVal
			}

			for _, kv := range args.kvPairs {
				val, err2 := ctx.Evaluate(kv.valueExpr)
				if err2 != nil {
					return err2
				}
				bindings[kv.key] = val
			}

			s, err := ctx.RenderFileIsolated(filename, bindings)
			if err != nil {
				return err
			}

			_, err = io.WriteString(w, s)

			return err
		}, nil
	}
}
