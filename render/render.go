// Package render parses and evaluates template strings.
package render

import (
	"fmt"
	"io"
	"reflect"

	"github.com/osteele/liquid/generics"
)

// Render renders the AST rooted at node to the writer.
func Render(node ASTNode, w io.Writer, ctx Context) error {
	return renderNode(node, w, ctx)
}

func renderNode(node ASTNode, w io.Writer, ctx Context) error { // nolint: gocyclo
	switch n := node.(type) {
	case *ASTSeq:
		for _, c := range n.Children {
			if err := renderNode(c, w, ctx); err != nil {
				return err
			}
		}
	case *ASTFunctional:
		return n.render(w, renderContext{ctx, n, nil})
	case *ASTText:
		_, err := w.Write([]byte(n.Source))
		return err
	case *ASTRaw:
		for _, s := range n.slices {
			_, err := w.Write([]byte(s))
			if err != nil {
				return err
			}
		}
	case *ASTBlock:
		cd, ok := ctx.settings.findBlockDef(n.Name)
		if !ok || cd.parser == nil {
			return fmt.Errorf("unknown tag: %s", n.Name)
		}
		renderer := n.renderer
		if renderer == nil {
			panic(fmt.Errorf("unset renderer for %v", n))
		}
		return renderer(w, renderContext{ctx, nil, n})
	case *ASTObject:
		value, err := ctx.Evaluate(n.expr)
		if err != nil {
			return fmt.Errorf("%s in %s", err, n.Source)
		}
		return writeObject(value, w)
	default:
		panic(fmt.Errorf("unknown node type %T", node))
	}
	return nil
}

// writeObject writes a value used in an object node
func writeObject(value interface{}, w io.Writer) error {
	value = generics.ToLiquid(value)
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
func (c Context) RenderASTSequence(w io.Writer, seq []ASTNode) error {
	for _, n := range seq {
		if err := renderNode(n, w, c); err != nil {
			return err
		}
	}
	return nil
}
