package chunks

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/osteele/liquid/expressions"
)

// RenderContext provides the rendering context for a tag renderer.
type RenderContext interface {
	Clone() RenderContext
	Get(name string) interface{}
	Set(name string, value interface{})
	Evaluate(expr expressions.Expression) (interface{}, error)
	EvaluateString(source string) (interface{}, error)
	EvaluateStatement(tag, source string) (interface{}, error)
	InnerString() (string, error)
	ParseTagArgs() (string, error)
	RenderChild(io.Writer, *ASTControlTag) error
	RenderChildren(io.Writer) error
	RenderFile(w io.Writer, filename string) error
	UpdateBindings(map[string]interface{})
}

type renderContext struct {
	ctx  Context
	node *ASTFunctional
	cn   *ASTControlTag
}

func (c renderContext) Clone() RenderContext {
	return renderContext{c.ctx.Clone(), c.node, c.cn}
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
	return expressions.EvaluateString(source, expressions.NewContext(c.ctx.bindings, c.ctx.settings.ExpressionSettings))
}

func (c renderContext) EvaluateStatement(tag, source string) (interface{}, error) {
	return c.EvaluateString(fmt.Sprintf("%%%s %s", tag, source))
}

// Get gets a variable value within an evaluation context.
func (c renderContext) Get(name string) interface{} {
	return c.ctx.bindings[name]
}

// Set sets a variable value from an evaluation context.
func (c renderContext) Set(name string, value interface{}) {
	c.ctx.bindings[name] = value
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

func (c renderContext) RenderFile(w io.Writer, filename string) error {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	ast, err := c.ctx.settings.Parse(string(source))
	if err != nil {
		return err
	}
	return ast.Render(w, c.ctx)
}

// InnerString renders the children to a string.
func (c renderContext) InnerString() (string, error) {
	buf := new(bytes.Buffer)
	if err := c.RenderChildren(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (c renderContext) ParseTagArgs() (string, error) {
	var args string
	switch {
	case c.node != nil:
		args = c.node.Chunk.Args
	case c.cn != nil:
		args = c.cn.Chunk.Args
	}
	if strings.Contains(args, "{{") {
		p, err := c.ctx.settings.Parse(args)
		if err != nil {
			return "", err
		}
		buf := new(bytes.Buffer)
		err = p.Render(buf, c.ctx)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return args, nil
}

func (c renderContext) UpdateBindings(bindings map[string]interface{}) {
	for k, v := range bindings {
		c.ctx.bindings[k] = v
	}
}
