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
//
// This interface shares the compatibility committments of the top-level liquid package.
type Context interface {
	// Get retrieves the value of a variable from the lexical environment.
	Get(name string) interface{}
	Evaluate(expr expression.Expression) (interface{}, error)
	// Evaluate compiles and interprets an expression, such as “x”, “x < 10", or “a.b | split | first | default: 10”, within the current lexical context.
	EvaluateString(source string) (interface{}, error)
	// EvaluateStatement evaluates a statement of the expression syntax.
	// Tag must be a special string known to the compiler.
	// For example, {% for %} uses this to parse the loop syntax.
	EvaluateStatement(tag, source string) (interface{}, error)
	// ExpandTagArg renders the current tag argument string as a Liquid template.
	// It enables the implementation of tags such as {% avatar {{page.author}} %}, from the jekyll-avatar plugin; or Jekyll's {% include %} parameters.
	ExpandTagArg() (string, error)
	// InnerString is the rendered children of the current block.
	InnerString() (string, error)
	RenderChild(io.Writer, *BlockNode) error
	RenderChildren(io.Writer) error
	RenderFile(string, map[string]interface{}) (string, error)
	// Set updates the value of a variable in the lexical environment.
	// For example, {% assign %} and {% capture %} use this.
	Set(name string, value interface{})
	// SourceFile retrieves the value set by template.SetSourcePath.
	// {% include %} uses this.
	SourceFile() string
	// TagArgs returns the text of the current tag, not including its name.
	// For example, the arguments to {% my_tag a b c %} would be “a b c”.
	TagArgs() string
	// TagName returns the name of the current tag.
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
		return c.node.Token.Args
	case c.cn != nil:
		return c.cn.Token.Args
	default:
		return ""
	}
}

func (c rendererContext) TagName() string {
	switch {
	case c.node != nil:
		return c.node.Token.Name
	case c.cn != nil:
		return c.cn.Token.Name
	default:
		return ""
	}
}
