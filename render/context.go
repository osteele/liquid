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
	ExpandTagArg() (string, error)
	InnerString() (string, error)
	RenderChild(io.Writer, *BlockNode) error
	RenderChildren(io.Writer) error
	RenderFile(string, map[string]interface{}) (string, error)
	Set(name string, value interface{})
	SourceFile() string
	TagArgs() string
	TagName() string
}

type rendererContext struct {
	ctx  nodeContext
	node *TagNode
	cn   *BlockNode
}

// Evaluate evaluates an expression within the template context.
func (c rendererContext) Evaluate(expr expression.Expression) (out interface{}, err error) {
	return c.ctx.Evaluate(expr)
}

func (c rendererContext) EvaluateStatement(tag, source string) (interface{}, error) {
	return c.EvaluateString(fmt.Sprintf("%%%s %s", tag, source))
}

// EvaluateString evaluates an expression within the template context.
func (c rendererContext) EvaluateString(source string) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case expression.InterpreterError:
				err = e
			default:
				// fmt.Println(string(debug.Stack()))
				panic(Errorf("%s during evaluation of %s", e, source))
			}
		}
	}()
	return expression.EvaluateString(source, expression.NewContext(c.ctx.bindings, c.ctx.config.Config.Config))
}

// Get gets a variable value within an evaluation context.
func (c rendererContext) Get(name string) interface{} {
	return c.ctx.bindings[name]
}

func (c rendererContext) ExpandTagArg() (string, error) {
	args := c.TagArgs()
	if strings.Contains(args, "{{") {
		p, err := c.ctx.config.Compile(args)
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
func (c rendererContext) RenderChild(w io.Writer, b *BlockNode) error {
	return c.ctx.RenderSequence(w, b.Body)
}

// RenderChildren renders the current node's children.
func (c rendererContext) RenderChildren(w io.Writer) error {
	if c.cn == nil {
		return nil
	}
	return c.ctx.RenderSequence(w, c.cn.Body)
}

func (c rendererContext) RenderFile(filename string, b map[string]interface{}) (string, error) {
	source, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	ast, err := c.ctx.config.Compile(string(source))
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
func (c rendererContext) InnerString() (string, error) {
	buf := new(bytes.Buffer)
	if err := c.RenderChildren(buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Set sets a variable value from an evaluation context.
func (c rendererContext) Set(name string, value interface{}) {
	c.ctx.bindings[name] = value
}

func (c rendererContext) SourceFile() string {
	return c.ctx.config.Filename
}

func (c rendererContext) TagArgs() string {
	switch {
	case c.node != nil:
		return c.node.Chunk.Args
	case c.cn != nil:
		return c.cn.Chunk.Args
	default:
		return ""
	}
}

func (c rendererContext) TagName() string {
	switch {
	case c.node != nil:
		return c.node.Chunk.Name
	case c.cn != nil:
		return c.cn.Chunk.Name
	default:
		return ""
	}
}
