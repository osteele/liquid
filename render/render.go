// Package render parses and evaluates template strings.
package render

import (
	"fmt"
	"io"
	"reflect"

	"github.com/osteele/liquid/evaluator"
)

// An Error is an evaluation error during template rendering.
type Error string

func (e Error) Error() string { return string(e) }

// Errorf creates a render error.
func Errorf(format string, a ...interface{}) Error {
	return Error(fmt.Sprintf(format, a...))
}

// Render renders the render tree.
func Render(node Node, w io.Writer, vars map[string]interface{}, c Config) error {
	// fmt.Println("render", c)
	return renderNode(node, w, newNodeContext(vars, c))
}

func renderNode(node Node, w io.Writer, ctx nodeContext) error { // nolint: gocyclo
	switch n := node.(type) {
	case *TagNode:
		return n.renderer(w, renderContext{ctx, n, nil})
	case *BlockNode:
		cd, ok := ctx.config.findBlockDef(n.Name)
		if !ok || cd.parser == nil {
			return Errorf("unknown tag: %s", n.Name)
		}
		renderer := n.renderer
		if renderer == nil {
			panic(Errorf("unset renderer for %v", n))
		}
		return renderer(w, renderContext{ctx, nil, n})
	case *RawNode:
		for _, s := range n.slices {
			_, err := w.Write([]byte(s))
			if err != nil {
				return err
			}
		}
	case *ObjectNode:
		value, err := ctx.Evaluate(n.expr)
		if err != nil {
			return Errorf("%s in %s", err, n.Source)
		}
		return writeObject(value, w)
	case *SeqNode:
		for _, c := range n.Children {
			if err := renderNode(c, w, ctx); err != nil {
				return err
			}
		}
	case *TextNode:
		_, err := w.Write([]byte(n.Source))
		return err
	default:
		panic(Errorf("unknown node type %T", node))
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
func (c nodeContext) RenderSequence(w io.Writer, seq []Node) error {
	for _, n := range seq {
		if err := renderNode(n, w, c); err != nil {
			return err
		}
	}
	return nil
}
