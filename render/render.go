// Package render renders a compiled template parse tree.
package render

import (
	"fmt"
	"io"
	"reflect"

	"github.com/osteele/liquid/evaluator"
)

// Render renders the render tree.
func Render(node Node, w io.Writer, vars map[string]interface{}, c Config) Error {
	return node.render(w, newNodeContext(vars, c))
}

func (n *BlockNode) render(w io.Writer, ctx nodeContext) Error {
	cd, ok := ctx.config.findBlockDef(n.Name)
	if !ok || cd.parser == nil {
		// this should have been detected during compilation; it's an implementation error if it happens here
		panic(fmt.Errorf("unknown tag: %s", n.Name))
	}
	renderer := n.renderer
	if renderer == nil {
		panic(fmt.Errorf("unset renderer for %v", n))
	}
	err := renderer(w, rendererContext{ctx, nil, n})
	return wrapRenderError(err, n)
}

func (n *RawNode) render(w io.Writer, ctx nodeContext) Error {
	for _, s := range n.slices {
		_, err := w.Write([]byte(s))
		if err != nil {
			return wrapRenderError(err, n)
		}
	}
	return nil
}

func (n *ObjectNode) render(w io.Writer, ctx nodeContext) Error {
	value, err := ctx.Evaluate(n.expr)
	if err != nil {
		return wrapRenderError(err, n)
	}
	return wrapRenderError(writeObject(value, w), n)
}

func (n *SeqNode) render(w io.Writer, ctx nodeContext) Error {
	for _, c := range n.Children {
		if err := c.render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *TagNode) render(w io.Writer, ctx nodeContext) Error {
	return wrapRenderError(n.renderer(w, rendererContext{ctx, n, nil}), n)
}

func (n *TextNode) render(w io.Writer, ctx nodeContext) Error {
	_, err := w.Write([]byte(n.Source))
	return wrapRenderError(err, n)
}

// writeObject writes a value used in an object node
func writeObject(value interface{}, w io.Writer) error {
	value = evaluator.ToLiquid(value)
	if value == nil {
		return nil
	}
	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	case reflect.Array, reflect.Slice:
		for i := 0; i < rt.Len(); i++ {
			item := rt.Index(i)
			if item.IsValid() {
				if err := writeObject(item.Interface(), w); err != nil {
					return err
				}
			}
		}
		return nil
	default:
		_, err := w.Write([]byte(fmt.Sprint(value)))
		return err
	}
}

// RenderASTSequence renders a sequence of nodes.
func (c nodeContext) RenderSequence(w io.Writer, seq []Node) Error {
	for _, n := range seq {
		if err := n.render(w, c); err != nil {
			return err
		}
	}
	return nil
}
