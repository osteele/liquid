// Package render is an internal package that renders a compiled template parse tree.
package render

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"

	"github.com/osteele/liquid/parser"

	"github.com/osteele/liquid/values"
)

// sizeLimitWriter wraps an io.Writer and stops writing once the byte limit is reached.
type sizeLimitWriter struct {
	w     io.Writer
	limit int64
	total int64
}

func (s *sizeLimitWriter) Write(p []byte) (int, error) {
	s.total += int64(len(p))
	if s.total > s.limit {
		return 0, fmt.Errorf("render size limit of %d bytes exceeded", s.limit)
	}
	return s.w.Write(p)
}

// Render renders the render tree.
func Render(node Node, w io.Writer, vars map[string]any, c Config) Error {
	var out io.Writer = w
	if c.SizeLimit > 0 {
		out = &sizeLimitWriter{w: w, limit: c.SizeLimit}
	}
	tw := trimWriter{w: out}

	err := node.render(&tw, newNodeContext(vars, c))
	if err != nil {
		return err
	}

	if _, flushErr := tw.Flush(); flushErr != nil {
		return &RenderError{parser.WrapError(flushErr, invalidLoc)}
	}

	return nil
}

// RenderSequence renders a sequence of nodes.
func (c nodeContext) RenderSequence(w io.Writer, seq []Node) Error {
	tw, ok := w.(*trimWriter)
	if !ok {
		tw = &trimWriter{w: w}
	}

	for _, n := range seq {
		if ctx := c.config.Context; ctx != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return &RenderError{parser.WrapError(ctxErr, invalidLoc)}
			}
		}
		err := n.render(tw, c)
		if err != nil {
			if h := c.config.ExceptionHandler; h != nil {
				if _, writeErr := io.WriteString(tw, h(err)); writeErr != nil {
					return wrapRenderError(writeErr, n)
				}
				continue
			}
			return err
		}
	}

	if _, flushErr := tw.Flush(); flushErr != nil {
		return &RenderError{parser.WrapError(flushErr, invalidLoc)}
	}

	return nil
}

func (n *BlockNode) render(w *trimWriter, ctx nodeContext) Error {
	cd, ok := ctx.config.findBlockDef(n.Name)
	if !ok || cd.parser == nil {
		// this should have been detected during compilation; it's an implementation error if it happens here
		panic(fmt.Errorf("undefined tag %q", n.Name))
	}

	renderer := n.renderer
	if renderer == nil {
		panic(fmt.Errorf("unset renderer for %v", n))
	}

	err := renderer(w, rendererContext{ctx, nil, n})

	// Annotate UndefinedVariableError with the innermost enclosing block tag
	// source, but only the first time (innermost wins over outer blocks).
	if uve, ok := err.(*UndefinedVariableError); ok && uve.BlockContext == "" {
		uve.BlockContext = n.Source
		uve.BlockLine = n.SourceLoc.LineNo
	}

	return wrapRenderError(err, n)
}

func (n *RawNode) render(w *trimWriter, ctx nodeContext) Error {
	for _, s := range n.slices {
		_, err := io.WriteString(w, s)
		if err != nil {
			return wrapRenderError(err, n)
		}
	}

	return nil
}

func (n *ObjectNode) render(w *trimWriter, ctx nodeContext) Error {
	// StrictVariables: check before evaluation so that undefined root
	// variables are caught even when a filter chain transforms nil → "".
	// We look for the presence of the root variable name in the bindings map
	// (key-existence, not value-truthiness) so that an explicit nil binding
	// still counts as "defined".
	if ctx.config.StrictVariables {
		vars := n.expr.Variables()
		if len(vars) > 0 && len(vars[0]) > 0 {
			root := vars[0][0]
			if _, exists := ctx.bindings[root]; !exists {
				// Name is the root variable name only (e.g. "user", not "user.name"),
				// matching Ruby Liquid's behaviour for dotted-path access.
				locErr := parser.Errorf(n, "undefined variable %q", root)
				return &UndefinedVariableError{Name: root, loc: locErr}
			}
		}
	}

	value, err := ctx.Evaluate(n.expr)
	if err != nil {
		return wrapRenderError(err, n)
	}

	if gf := ctx.config.globalFilter; gf != nil {
		value, err = gf(value)
		if err != nil {
			return wrapRenderError(err, n)
		}
	}

	if sv, isSafe := value.(values.SafeValue); isSafe {
		err = writeObject(w, sv.Value)
	} else {
		var fw io.Writer
		if replacer := ctx.config.escapeReplacer; replacer != nil {
			fw = &replacerWriter{
				replacer: replacer,
				w:        w,
			}
		} else {
			fw = w
		}
		err = writeObject(fw, value)
	}
	if err != nil {
		return wrapRenderError(err, n)
	}

	return nil
}

func (n *SeqNode) render(w *trimWriter, ctx nodeContext) Error {
	for _, c := range n.Children {
		if ctxVal := ctx.config.Context; ctxVal != nil {
			if ctxErr := ctxVal.Err(); ctxErr != nil {
				return &RenderError{parser.WrapError(ctxErr, invalidLoc)}
			}
		}
		err := c.render(w, ctx)
		if err != nil {
			if h := ctx.config.ExceptionHandler; h != nil {
				if _, writeErr := io.WriteString(w, h(err)); writeErr != nil {
					return wrapRenderError(writeErr, n)
				}
				continue
			}
			return err
		}
	}

	return nil
}

func (n *TagNode) render(w *trimWriter, ctx nodeContext) Error {
	err := wrapRenderError(n.renderer(w, rendererContext{ctx, n, nil}), n)
	return err
}

func (n *TextNode) render(w *trimWriter, _ nodeContext) Error {
	_, err := io.WriteString(w, n.Source)
	return wrapRenderError(err, n)
}

func (n *TrimNode) render(w *trimWriter, _ nodeContext) Error {
	if n.TrimDirection == parser.Left {
		if n.Greedy {
			return wrapRenderError(w.TrimLeft(), n)
		}
		return wrapRenderError(w.TrimLeftNonGreedy(), n)
	}
	if n.Greedy {
		w.TrimRight()
	} else {
		w.TrimRightNonGreedy()
	}
	return nil
}

// writeObject writes a value used in an object node
func writeObject(w io.Writer, value any) error {
	value = values.ToLiquid(value)
	if value == nil {
		return nil
	}

	switch value := value.(type) {
	case string:
		_, err := io.WriteString(w, value)
		return err
	case int:
		_, err := io.WriteString(w, strconv.Itoa(value))
		return err
	case float64:
		_, err := io.WriteString(w, strconv.FormatFloat(value, 'f', -1, 64))
		return err
	case bool:
		if value {
			_, err := io.WriteString(w, "true")
			return err
		}
		_, err := io.WriteString(w, "false")
		return err
	case time.Time:
		_, err := io.WriteString(w, value.Format("2006-01-02 15:04:05 -0700"))
		return err
	case []byte:
		_, err := w.Write(value)
		return err
	}

	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		for i := range rt.Len() {
			item := rt.Index(i)
			if item.IsValid() {
				err := writeObject(w, item.Interface())
				if err != nil {
					return err
				}
			}
		}

		return nil
	case reflect.Ptr:
		return writeObject(w, reflect.ValueOf(value).Elem())
	default:
		_, err := io.WriteString(w, fmt.Sprint(value))
		return err
	}
}

type replacerWriter struct {
	replacer Replacer
	w        io.Writer
}

func (h *replacerWriter) Write(p []byte) (n int, err error) {
	_, err = h.WriteString(string(p))
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (h *replacerWriter) WriteString(s string) (n int, err error) {
	return h.replacer.WriteString(h.w, s)
}
