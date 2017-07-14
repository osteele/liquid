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
	return renderNode(node, w, newNodeContext(vars, c))
}

func renderNode(node Node, w io.Writer, ctx nodeContext) Error { // nolint: gocyclo
	switch n := node.(type) {
	case *BlockNode:
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
	case *RawNode:
		for _, s := range n.slices {
			_, err := w.Write([]byte(s))
			if err != nil {
				return wrapRenderError(err, n)
			}
		}
	case *ObjectNode:
		value, err := ctx.Evaluate(n.expr)
		if err != nil {
			return wrapRenderError(err, n)
		}
		return wrapRenderError(writeObject(value, w), n)
	case *SeqNode:
		for _, c := range n.Children {
			if err := renderNode(c, w, ctx); err != nil {
				return err
			}
		}
	case *TagNode:
		return wrapRenderError(n.renderer(w, rendererContext{ctx, n, nil}), n)
	case *TextNode:
		_, err := w.Write([]byte(n.Source))
		return wrapRenderError(err, n)
	default:
		panic(fmt.Errorf("unknown node type %T", node))
	}
	return nil
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
		if err := renderNode(n, w, c); err != nil {
			return err
		}
	}
	return nil
}
