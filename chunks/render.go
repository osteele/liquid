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
	err := n.render(w, ctx)
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
func (n *ASTControlTag) Render(w io.Writer, ctx Context) error {
	cd, ok := findControlTagDefinition(n.Name)
	if !ok || cd.parser == nil {
		return fmt.Errorf("unimplemented tag: %s", n.Name)
	}
	renderer := n.renderer
	if renderer == nil {
		panic(fmt.Errorf("unset renderer for %v", n))
	}
	return renderer(w, ctx)
}

// Render is in the ASTNode interface.
func (n *ASTObject) Render(w io.Writer, ctx Context) error {
	value, err := ctx.Evaluate(n.expr)
	if err != nil {
		return fmt.Errorf("%s in %s", err, n.Source)
	}
	return writeValue(value, w)
}

func writeValue(value interface{}, w io.Writer) error {
	if value == nil {
		return nil
	}
	rt := reflect.ValueOf(value)
	switch rt.Kind() {
	// TODO test case for this
	case reflect.Array, reflect.Slice:
		for i := 0; i < rt.Len(); i++ {
			item := rt.Index(i)
			if item.IsValid() {
				if err := writeValue(item.Interface(), w); err != nil {
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
func (ctx Context) RenderASTSequence(w io.Writer, seq []ASTNode) error {
	for _, n := range seq {
		if err := n.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}
