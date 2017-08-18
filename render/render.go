// Package render is an internal package that renders a compiled template parse tree.
package render

import (
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/osteele/liquid/values"
)

// Render renders the render tree.
func Render(node Node, w io.Writer, vars map[string]interface{}, c Config) Error {
	tw := trimWriter{w: w}
	if err := node.render(&tw, newNodeContext(vars, c)); err != nil {
		return err
	}
	if err := tw.Flush(); err != nil {
		panic(err)
	}
	return nil
}

// RenderASTSequence renders a sequence of nodes.
func (c nodeContext) RenderSequence(w io.Writer, seq []Node) Error {
	tw := trimWriter{w: w}
	for _, n := range seq {
		if err := n.render(&tw, c); err != nil {
			return err
		}
	}
	if err := tw.Flush(); err != nil {
		panic(err)
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
	w.TrimLeft(n.TrimLeft)
	value, err := ctx.Evaluate(n.expr)
	if err != nil {
		return wrapRenderError(err, n)
	}
	if err := wrapRenderError(writeObject(w, value), n); err != nil {
		return err
	}
	w.TrimRight(n.TrimRight)
	return nil
}

func (n *SeqNode) render(w *trimWriter, ctx nodeContext) Error {
	for _, c := range n.Children {
		if err := c.render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *TagNode) render(w *trimWriter, ctx nodeContext) Error {
	w.TrimLeft(n.TrimLeft)
	err := wrapRenderError(n.renderer(w, rendererContext{ctx, n, nil}), n)
	w.TrimRight(n.TrimRight)
	return err
}

func (n *TextNode) render(w *trimWriter, ctx nodeContext) Error {
	_, err := io.WriteString(w, n.Source)
	return wrapRenderError(err, n)
}

// writeObject writes a value used in an object node
func writeObject(w io.Writer, value interface{}) error {
	value = values.ToLiquid(value)
	if value == nil {
		return nil
	}
	switch value := value.(type) {
	case time.Time:
		_, err := io.WriteString(w, value.Format("2006-01-02 15:04:05 -0700"))
		return err
	case []byte:
		_, err := w.Write(value)
		return err
		// there used be a case on fmt.Stringer here, but fmt.Sprint produces better results than obj.Write
		// for instances of error and *string
	}
	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rt.Len(); i++ {
			item := rt.Index(i)
			if item.IsValid() {
				if err := writeObject(w, item.Interface()); err != nil {
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
