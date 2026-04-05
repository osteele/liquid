// Package tags is an internal package that defines the standard Liquid tags.
package tags

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// AddStandardTags defines the standard Liquid tags.
func AddStandardTags(c *render.Config) {
	c.AddTag("assign", makeAssignTag(c))
	c.AddTag("echo", echoTag)
	c.AddTag("include", makeIncludeTag(c))
	c.AddTag("increment", incrementTag)
	c.AddTag("decrement", decrementTag)
	c.AddTag("liquid", makeLiquidTag(c))
	c.AddTag("render", makeRenderTag(c))

	// blocks
	// The parser only recognize the comment and raw tags if they've been defined,
	// but it ignores any syntax specified here.
	c.AddTag("break", breakTag)
	c.AddTag("continue", continueTag)
	c.AddTag("cycle", cycleTag)
	c.AddBlock("capture").Compiler(captureTagCompiler)
	c.AddBlock("case").Clause("when").Clause("else").Compiler(caseTagCompiler)
	c.AddBlock("comment")
	c.AddBlock("doc")
	c.AddBlock("for").Clause("else").Compiler(loopTagCompiler)
	c.AddBlock("ifchanged").Compiler(ifchangedCompiler)
	c.AddBlock("layout").Compiler(makeLayoutTag(c))
	c.AddBlock("block").Compiler(blockTagCompiler)
	c.AddBlock("if").Clause("else").Clause("elsif").Compiler(ifTagCompiler(true))
	c.AddBlock("raw")
	c.AddBlock("tablerow").Compiler(loopTagCompiler)
	c.AddBlock("unless").Clause("else").Compiler(ifTagCompiler(false))

	// static analysis: register analyzers alongside compilers
	c.AddTagAnalyzer("assign", makeAssignAnalyzer())
	c.AddBlockAnalyzer("capture", captureBlockAnalyzer)
	c.AddBlockAnalyzer("for", loopBlockAnalyzer)
	c.AddBlockAnalyzer("tablerow", loopBlockAnalyzer)
	c.AddBlockAnalyzer("if", ifBlockAnalyzer())
	c.AddBlockAnalyzer("unless", ifBlockAnalyzer())
	c.AddBlockAnalyzer("case", caseBlockAnalyzer)
}

func echoTag(source string) (func(io.Writer, render.Context) error, error) {
	if strings.TrimSpace(source) == "" {
		return nil, fmt.Errorf("syntax error: echo tag requires an expression")
	}
	expr, err := expressions.Parse(source)
	if err != nil {
		return nil, err
	}
	return func(w io.Writer, ctx render.Context) error {
		value, err := ctx.Evaluate(expr)
		if err != nil {
			return err
		}
		return ctx.WriteValue(w, value)
	}, nil
}

func makeAssignTag(cfg *render.Config) func(string) (func(io.Writer, render.Context) error, error) {
	return func(source string) (func(io.Writer, render.Context) error, error) {
		stmt, err := expressions.ParseStatement(expressions.AssignStatementSelector, source)
		if err != nil {
			return nil, err
		}

		// Check if dot notation is used without Jekyll extensions enabled
		if len(stmt.Path) > 1 && !cfg.JekyllExtensions {
			return nil, errors.New("syntax error: dot notation in assign tag (e.g., 'obj.property = value') requires Jekyll extensions to be enabled")
		}

		return func(w io.Writer, ctx render.Context) error {
			value, err := ctx.Evaluate(stmt.ValueFn)
			if err != nil {
				return err
			}

			// Use Path if available (dot notation), otherwise fall back to Variable (simple assignment)
			if len(stmt.Path) > 1 {
				return ctx.SetPath(stmt.Path, value)
			}

			// Simple assignment (backward compatibility and standard mode)
			ctx.Set(stmt.Assignment.Variable, value)

			return nil
		}, nil
	}
}

func captureTagCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	varname := strings.TrimSpace(node.Args)
	if varname == "" || strings.ContainsAny(varname, " \t") {
		return nil, fmt.Errorf("syntax error: capture tag requires exactly one variable name, got %q", node.Args)
	}

	return func(w io.Writer, ctx render.Context) error {
		s, err := ctx.InnerString()
		if err != nil {
			return err
		}

		ctx.Set(varname, s)

		return nil
	}, nil
}

// Increment and decrement each maintain a separate counter namespace, keyed
// by a null-byte prefix that cannot appear in valid Liquid variable names.
// Per Shopify spec, {% increment x %} and {% decrement x %} on the same name
// are NOT related (decrement does not affect increment's counter and vice versa).
const (
	incrementKey = "\x00inc_counters"
	decrementKey = "\x00dec_counters"
)

func getOrCreateCounterMap(ctx render.Context, key string) map[string]int {
	if m, ok := ctx.Get(key).(map[string]int); ok {
		return m
	}

	m := map[string]int{}
	ctx.Set(key, m)

	return m
}

// incrementTag implements {% increment var %}.
// Outputs the current counter value then increments it (starts at 0).
// Counter is in a separate namespace from assign variables and decrement.
func incrementTag(source string) (func(io.Writer, render.Context) error, error) {
	varname := strings.TrimSpace(source)
	if varname == "" {
		return nil, fmt.Errorf("syntax error: increment tag requires a variable name")
	}

	return func(w io.Writer, ctx render.Context) error {
		counters := getOrCreateCounterMap(ctx, incrementKey)
		n := counters[varname]
		counters[varname] = n + 1
		_, err := fmt.Fprintf(w, "%d", n)

		return err
	}, nil
}

// decrementTag implements {% decrement var %}.
// Decrements the counter then outputs the new value.
// Counter starts at 0, so first call outputs -1.
// Counter is in a separate namespace from assign variables and increment.
func decrementTag(source string) (func(io.Writer, render.Context) error, error) {
	varname := strings.TrimSpace(source)
	if varname == "" {
		return nil, fmt.Errorf("syntax error: decrement tag requires a variable name")
	}

	return func(w io.Writer, ctx render.Context) error {
		counters := getOrCreateCounterMap(ctx, decrementKey)
		n := counters[varname] - 1
		counters[varname] = n
		_, err := fmt.Fprintf(w, "%d", n)

		return err
	}, nil
}

// makeLiquidTag implements {% liquid ... %} — a multi-line tag where each
// non-empty, non-comment line is treated as a separate tag statement.
// Lines starting with "#" are comments. Empty lines are ignored.
// The content is compiled at template-parse time and rendered into the
// current scope (assign statements propagate to the outer context).
func makeLiquidTag(cfg *render.Config) func(string) (func(io.Writer, render.Context) error, error) {
	return func(source string) (func(io.Writer, render.Context) error, error) {
		lines := strings.Split(source, "\n")

		var sb strings.Builder

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}

			sb.WriteString("{%")
			sb.WriteString(trimmed)
			sb.WriteString("%}")
		}

		templateStr := sb.String()
		if templateStr == "" {
			return func(io.Writer, render.Context) error { return nil }, nil
		}

		node, err := cfg.Compile(templateStr, parser.SourceLoc{})
		if err != nil {
			return nil, err
		}

		seqNode, ok := node.(*render.SeqNode)
		if !ok {
			return nil, fmt.Errorf("internal error: unexpected node type from liquid tag compilation")
		}

		body := seqNode.Children

		return func(w io.Writer, ctx render.Context) error {
			return ctx.RenderBlock(w, &render.BlockNode{Body: body})
		}, nil
	}
}

// ifchangedCompiler implements {% ifchanged %}...{% endifchanged %}.
// Renders its body only if the output differs from the previous call.
// State is stored in the current context under a null-byte-prefixed key
// so it cannot collide with user-defined variables.
func ifchangedCompiler(node render.BlockNode) (func(io.Writer, render.Context) error, error) {
	const ifchangedKey = "\x00ifchanged_last"

	return func(w io.Writer, ctx render.Context) error {
		var buf bytes.Buffer
		if err := ctx.RenderChildren(&buf); err != nil {
			return err
		}

		content := buf.String()
		last, _ := ctx.Get(ifchangedKey).(string)

		if content != last {
			ctx.Set(ifchangedKey, content)
			_, err := io.WriteString(w, content)
			return err
		}

		return nil
	}, nil
}
