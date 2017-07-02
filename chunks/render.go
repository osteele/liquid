package chunks

import (
	"fmt"
	"io"
	"reflect"
)

// Render is in the ASTNode interface.
func (n *ASTSeq) Render(w io.Writer, ctx Context) error {
	for _, c := range n.Children {
		if err := c.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

// Render is in the ASTNode interface.
func (n *ASTFunctional) Render(w io.Writer, ctx Context) error {
	err := n.render(w, renderContext{ctx, n, nil})
	// TODO restore something like this
	// if err != nil {
	// 	fmt.Println("while parsing", n.Source)
	// }
	return err
}

// Render is in the ASTNode interface.
func (n *ASTText) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte(n.Source))
	return err
}

// Render is in the ASTNode interface.
func (n *ASTRaw) Render(w io.Writer, _ Context) error {
	for _, s := range n.slices {
		_, err := w.Write([]byte(s))
		if err != nil {
			return err
		}
	}
	return nil
}

// Render is in the ASTNode interface.
func (n *ASTBlockNode) Render(w io.Writer, ctx Context) error {
	cd, ok := ctx.settings.findBlockDef(n.Name)
	if !ok || cd.parser == nil {
		return fmt.Errorf("unknown tag: %s", n.Name)
	}
	renderer := n.renderer
	if renderer == nil {
		panic(fmt.Errorf("unset renderer for %v", n))
	}
	return renderer(w, renderContext{ctx, nil, n})
}

// Render is in the ASTNode interface.
func (n *ASTObject) Render(w io.Writer, ctx Context) error {
	value, err := ctx.Evaluate(n.expr)
	if err != nil {
		return fmt.Errorf("%s in %s", err, n.Source)
	}
	return writeObject(value, w)
}

// writeObject writes a value used in an object node
func writeObject(value interface{}, w io.Writer) error {
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
		if err := n.Render(w, c); err != nil {
			return err
		}
	}
	return nil
}
