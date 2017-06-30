package chunks

import (
	"bytes"
	"fmt"
	"io"

	"github.com/osteele/liquid/expressions"
)

// RenderContext provides the rendering context for a tag renderer.
type RenderContext interface {
	Get(name string) interface{}
	Set(name string, value interface{})
	GetVariableMap() map[string]interface{}
	Evaluate(expr expressions.Expression) (interface{}, error)
	EvaluateString(source string) (interface{}, error)
	EvaluateStatement(tag, source string) (interface{}, error)
	RenderChild(io.Writer, *ASTControlTag) error
	RenderChildren(io.Writer) error
	// RenderTemplate(io.Writer, filename string) (string, error)
	InnerString() (string, error)
}

type renderContext struct {
	ctx  Context
	node *ASTFunctional
	cn   *ASTControlTag
}

// Evaluate evaluates an expression within the template context.
func (c renderContext) Evaluate(expr expressions.Expression) (out interface{}, err error) {
	return c.ctx.Evaluate(expr)
}

// EvaluateString evaluates an expression within the template context.
func (c renderContext) EvaluateString(source string) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case expressions.InterpreterError:
				err = e
			default:
				// fmt.Println(string(debug.Stack()))
				panic(fmt.Errorf("%s during evaluation of %s", e, source))
			}
		}
	}()
	return expressions.EvaluateString(source, expressions.NewContext(c.ctx.vars, c.ctx.settings.ExpressionSettings))
}

func (c renderContext) EvaluateStatement(tag, source string) (interface{}, error) {
	return c.EvaluateString(fmt.Sprintf("%%%s %s", tag, source))
}

// GetVariableMap returns the variable map. This is required by some tangled code
// in Jekyll includes, that should hopefully get better.
func (c renderContext) GetVariableMap() map[string]interface{} {
	return c.ctx.vars
}

// Get gets a variable value within an evaluation context.
func (c renderContext) Get(name string) interface{} {
	return c.ctx.vars[name]
}

// Set sets a variable value from an evaluation context.
func (c renderContext) Set(name string, value interface{}) {
	c.ctx.vars[name] = value
}

// RenderChild renders a node.
func (c renderContext) RenderChild(w io.Writer, b *ASTControlTag) error {
	return c.ctx.RenderASTSequence(w, b.Body)
}

// RenderChildren renders the current node's children.
func (c renderContext) RenderChildren(w io.Writer) error {
	if c.cn == nil {
		return nil
	}
	return c.ctx.RenderASTSequence(w, c.cn.Body)
}

// func (c renderContext) RenderTemplate(w io.Writer, filename string) (string, error) {
// 	// TODO use the tags and filters from the current context
// }

// InnerString renders the children to a string.
func (c renderContext) InnerString() (string, error) {
	buf := new(bytes.Buffer)
	if err := c.RenderChildren(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
