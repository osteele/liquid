package render

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"strings"

	"github.com/osteele/liquid/parser"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/values"
)

// Context provides the rendering context for a tag renderer.
type Context interface {
	// Bindings returns the current lexical environment.
	Bindings() map[string]any
	// Get retrieves the value of a variable from the current lexical environment.
	Get(name string) any
	// Errorf creates a SourceError, that includes the source location.
	// Use this to distinguish errors in the template from implementation errors
	// in the template engine.
	Errorf(format string, a ...any) Error
	// Evaluate evaluates a compiled expression within the current lexical context.
	Evaluate(expressions.Expression) (any, error)
	// EvaluateString compiles and evaluates a string expression such as “x”, “x < 10", or “a.b | split | first | default: 10”, within the current lexical context.
	EvaluateString(string) (any, error)
	// ExpandTagArg renders the current tag argument string as a Liquid template.
	// It enables the implementation of tags such as Jekyll's "{% include {{ page.my_variable }} %}" andjekyll-avatar's  "{% avatar {{page.author}} %}".
	ExpandTagArg() (string, error)
	// InnerString is the rendered content of the current block.
	// It's used in the implementation of the Liquid "capture" tag and the Jekyll "highlght" tag.
	InnerString() (string, error)
	// RenderBlock is used in the implementation of the built-in control flow tags.
	// It's not guaranteed stable.
	RenderBlock(io.Writer, *BlockNode) error
	// RenderChildren is used in the implementation of the built-in control flow tags.
	// It's not guaranteed stable.
	RenderChildren(io.Writer) Error
	// RenderFile parses and renders a template. It's used in the implementation of the {% include %} tag.
	// RenderFile does not cache the compiled template.
	RenderFile(string, map[string]any) (string, error)
	// RenderFileIsolated parses and renders a template in an isolated scope.
	// Unlike RenderFile, the rendered template cannot access variables from the calling context —
	// only the explicitly provided bindings are available.
	// It's used in the implementation of the {% render %} tag.
	RenderFileIsolated(string, map[string]any) (string, error)
	// Set updates the value of a variable in the current lexical environment.
	// It's used in the implementation of the {% assign %} and {% capture %} tags.
	Set(name string, value any)
	// SetPath sets a value at a nested path in the context.
	// For example, SetPath(["page", "canonical_url"], "/about/") sets page.canonical_url = "/about/"
	SetPath(path []string, value any) error
	// SourceFile retrieves the value set by template.SetSourcePath.
	// It's used in the implementation of the {% include %} tag.
	SourceFile() string
	// TagArgs returns the text of the current tag, not including its name.
	// For example, the arguments to {% my_tag a b c %} would be “a b c”.
	TagArgs() string
	// TagName returns the name of the current tag; for example "my_tag" for {% my_tag a b c %}.
	TagName() string
	// WrapError creates a new error that records the source location from the current context.
	WrapError(err error) Error
	// WriteValue writes a value to the writer using the same rendering rules as {{ expr }}.
	// nil renders as empty string, arrays render as space-joined elements, and autoescape
	// is applied if configured on the engine.
	WriteValue(w io.Writer, value any) error
}

type TemplateStore interface {
	ReadTemplate(templatename string) ([]byte, error)
}

type rendererContext struct {
	ctx  nodeContext
	node *TagNode
	cn   *BlockNode
}

type invalidLocation struct{}

func (i invalidLocation) SourceLocation() parser.SourceLoc {
	return parser.SourceLoc{}
}

func (i invalidLocation) SourceText() string {
	return ""
}

var invalidLoc parser.Locatable = invalidLocation{}

func (c rendererContext) Errorf(format string, a ...any) Error {
	switch {
	case c.node != nil:
		return renderErrorf(c.node, format, a...)
	case c.cn != nil:
		return renderErrorf(c.cn, format, a...)
	default:
		return renderErrorf(invalidLoc, format, a...)
	}
}

// sourceLoc returns the source location of the current node, preferring tag
// nodes over block nodes. Returns a zero SourceLoc when neither is set.
func (c rendererContext) sourceLoc() parser.SourceLoc {
	if c.node != nil {
		return c.node.SourceLoc
	}

	if c.cn != nil {
		return c.cn.SourceLoc
	}

	return parser.SourceLoc{}
}

func (c rendererContext) WrapError(err error) Error {
	switch {
	case c.node != nil:
		return wrapRenderError(err, c.node)
	case c.cn != nil:
		return wrapRenderError(err, c.cn)
	default:
		return wrapRenderError(err, invalidLoc)
	}
}

func (c rendererContext) WriteValue(w io.Writer, value any) error {
	if sv, isSafe := value.(values.SafeValue); isSafe {
		return writeObject(w, sv.Value)
	}
	if replacer := c.ctx.config.escapeReplacer; replacer != nil {
		w = &replacerWriter{replacer: replacer, w: w}
	}
	return writeObject(w, value)
}

func (c rendererContext) Evaluate(expr expressions.Expression) (out any, err error) {
	return c.ctx.Evaluate(expr)
}

// EvaluateString evaluates an expression within the template context.
func (c rendererContext) EvaluateString(source string) (out any, err error) {
	return expressions.EvaluateString(source, expressions.NewContext(c.ctx.bindings, c.ctx.config.Config.Config))
}

// Bindings returns the current lexical environment.
func (c rendererContext) Bindings() map[string]any {
	return c.ctx.bindings
}

// Get gets a variable value within an evaluation context.
func (c rendererContext) Get(name string) any {
	return c.ctx.bindings[name]
}

func (c rendererContext) ExpandTagArg() (string, error) {
	args := c.TagArgs()
	if strings.Contains(args, "{{") {
		root, err := c.ctx.config.Compile(args, c.node.SourceLoc)
		if err != nil {
			return "", err
		}

		buf := new(bytes.Buffer)

		err = Render(root, buf, c.ctx.bindings, c.ctx.config)
		if err != nil {
			return "", err
		}

		return buf.String(), nil
	}

	return args, nil
}

// RenderBlock renders a node.
func (c rendererContext) RenderBlock(w io.Writer, b *BlockNode) error {
	return c.ctx.RenderSequence(w, b.Body)
}

// RenderChildren renders the current node's children.
func (c rendererContext) RenderChildren(w io.Writer) Error {
	if c.cn == nil {
		return nil
	}

	return c.ctx.RenderSequence(w, c.cn.Body)
}

func (c rendererContext) RenderFile(filename string, b map[string]any) (string, error) {
	source, err := c.ctx.config.TemplateStore.ReadTemplate(filename)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		// Is it cached?
		if cval, ok := c.ctx.config.Cache.Load(filename); ok {
			source = cval.([]byte)
		} else {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	root, err := c.ctx.config.Compile(string(source), c.sourceLoc())
	if err != nil {
		return "", err
	}

	bindings := make(map[string]any, len(c.ctx.bindings)+len(b))
	maps.Copy(bindings, c.ctx.bindings)
	maps.Copy(bindings, b)

	buf := new(bytes.Buffer)
	if err := Render(root, buf, bindings, c.ctx.config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderFileIsolated parses and renders a template in an isolated scope.
// The rendered template cannot access variables from the calling context —
// only the explicitly provided bindings are available.
// This is used by the {% render %} tag.
func (c rendererContext) RenderFileIsolated(filename string, b map[string]any) (string, error) {
	source, err := c.ctx.config.TemplateStore.ReadTemplate(filename)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		// Is it cached?
		if cval, ok := c.ctx.config.Cache.Load(filename); ok {
			source = cval.([]byte)
		} else {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	root, err := c.ctx.config.Compile(string(source), c.sourceLoc())
	if err != nil {
		return "", err
	}

	// Only use passed bindings; do not inherit parent scope.
	buf := new(bytes.Buffer)
	if err := Render(root, buf, b, c.ctx.config); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// InnerString renders the children to a string.
func (c rendererContext) InnerString() (string, error) {
	buf := new(bytes.Buffer)

	err := c.RenderChildren(buf)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Set sets a variable value from an evaluation context.
func (c rendererContext) Set(name string, value any) {
	c.ctx.bindings[name] = value
}

// SetPath sets a value at a nested path in the context.
// For example, SetPath(["page", "canonical_url"], "/about/") sets page.canonical_url = "/about/"
func (c rendererContext) SetPath(path []string, value any) error {
	if len(path) == 0 {
		return errors.New("empty path")
	}

	// For single element paths, use regular Set
	if len(path) == 1 {
		c.Set(path[0], value)
		return nil
	}

	// Navigate to the parent object
	current := c.ctx.bindings

	for i := range len(path) - 1 {
		key := path[i]

		// Get or create the intermediate object
		if obj, exists := current[key]; exists {
			// Check if it's a map we can navigate into
			switch v := obj.(type) {
			case map[string]any:
				current = v
			default:
				// Can't navigate into non-map types
				return fmt.Errorf("cannot set property on non-object at '%s'", key)
			}
		} else {
			// Create intermediate object
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		}
	}

	// Set the final value
	current[path[len(path)-1]] = value

	return nil
}

func (c rendererContext) SourceFile() string {
	switch {
	case c.node != nil:
		return c.node.SourceLoc.Pathname
	case c.cn != nil:
		return c.cn.SourceLoc.Pathname
	default:
		return ""
	}
}

func (c rendererContext) TagArgs() string {
	switch {
	case c.node != nil:
		return c.node.Args
	case c.cn != nil:
		return c.cn.Args
	default:
		return ""
	}
}

func (c rendererContext) TagName() string {
	switch {
	case c.node != nil:
		return c.node.Name
	case c.cn != nil:
		return c.cn.Name
	default:
		return ""
	}
}
