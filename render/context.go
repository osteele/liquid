package render

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/osteele/liquid/expression"
)

// Context provides the rendering context for a tag renderer.
type Context interface {
	Get(name string) interface{}
	Evaluate(expr expression.Expression) (interface{}, error)
	EvaluateString(source string) (interface{}, error)
	EvaluateStatement(tag, source string) (interface{}, error)
	InnerString() (string, error)
	ParseTagArgs() (string, error)
	RenderChild(io.Writer, *ASTBlock) error
	RenderChildren(io.Writer) error
	RenderFile(string, map[string]interface{}) (string, error)
	Set(name string, value interface{})
	SourceFile() string
	TagArgs() string
	TagName() string
}

type renderContext struct {
	ctx  nodeContext
	node *ASTFunctional
	cn   *ASTBlock
}

// Evaluate evaluates an expression within the template context.
func (c renderContext) Evaluate(expr expression.Expression) (out interface{}, err error) {
	return c.ctx.Evaluate(expr)
}

func (c renderContext) EvaluateStatement(tag, source string) (interface{}, error) {
	return c.EvaluateString(fmt.Sprintf("%%%s %s", tag, source))
}

// EvaluateString evaluates an expression within the template context.
func (c renderContext) EvaluateString(source string) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case expression.InterpreterError:
				err = e
			default:
				// fmt.Println(string(debug.Stack()))
				panic(fmt.Errorf("%s during evaluation of %s", e, source))
			}
		}
	}()
	return expression.EvaluateString(source, expression.NewContext(c.ctx.bindings, c.ctx.config.Config))
}

// Get gets a variable value within an evaluation context.
func (c renderContext) Get(name string) interface{} {
	return c.ctx.bindings[name]
}

func (c renderContext) ParseTagArgs() (string, error) {
	args := c.TagArgs()
	if strings.Contains(args, "{{") {
		p, err := c.ctx.config.Parse(args)
		if err != nil {
			return "", err
		}
		buf := new(bytes.Buffer)
		err = renderNode(p, buf, c.ctx)
		if err != nil {
			return "", err
		}
		return buf.String(), nil
	}
	return args, nil
}

// RenderChild renders a node.
func (c renderContext) RenderChild(w io.Writer, b *ASTBlock) error {
	return c.ctx.RenderASTSequence(w, b.Body)
}

// RenderChildren renders the current node's children.
func (c renderContext) RenderChildren(w io.Writer) error {
	if c.cn == nil {
		return nil
	}
	return c.ctx.RenderASTSequence(w, c.cn.Body)
}

func (c renderContext) RenderFile(filename string, b map[string]interface{}) (string, error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	ast, err := c.ctx.config.Parse(string(source))
	if err != nil {
		return "", err
	}
	nc := c.ctx.Clone()
	for k, v := range b {
		c.ctx.bindings[k] = v
	}
	buf := new(bytes.Buffer)
	if err := renderNode(ast, buf, nc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// InnerString renders the children to a string.
func (c renderContext) InnerString() (string, error) {
	buf := new(bytes.Buffer)
	if err := c.RenderChildren(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Set sets a variable value from an evaluation context.
func (c renderContext) Set(name string, value interface{}) {
	c.ctx.bindings[name] = value
}

func (c renderContext) SourceFile() string {
	return c.ctx.config.Filename
}

func (c renderContext) TagArgs() string {
	switch {
	case c.node != nil:
		return c.node.Chunk.Args
	case c.cn != nil:
		return c.cn.Chunk.Args
	default:
		return ""
	}
}

func (c renderContext) TagName() string {
	switch {
	case c.node != nil:
		return c.node.Chunk.Name
	case c.cn != nil:
		return c.cn.Chunk.Name
	default:
		return ""
	}
}
