package render

import (
	"github.com/osteele/liquid/expressions"
)

// Context is the evaluation context for chunk AST rendering.
type Context struct {
	bindings map[string]interface{}
	settings Settings
}

// Settings holds configuration information for parsing and rendering.
type Settings struct {
	ExpressionSettings expressions.Settings
	tags               map[string]TagDefinition
	controlTags        map[string]*blockDef
}

// AddFilter adds a filter to settings.
func (s Settings) AddFilter(name string, fn interface{}) {
	s.ExpressionSettings.AddFilter(name, fn)
}

// NewSettings creates a new Settings.
func NewSettings() Settings {
	s := Settings{
		expressions.NewSettings(),
		map[string]TagDefinition{},
		map[string]*blockDef{},
	}
	s.AddTag("assign", assignTagDef)
	return s
}

// NewContext creates a new evaluation context.
func NewContext(scope map[string]interface{}, s Settings) Context {
	// The assign tag modifies the scope, so make a copy first.
	// TODO this isn't really the right place for this.
	vars := map[string]interface{}{}
	for k, v := range scope {
		vars[k] = v
	}
	return Context{vars, s}
}

// Clone makes a copy of a context, with copied bindings.
func (c Context) Clone() Context {
	bindings := map[string]interface{}{}
	for k, v := range c.bindings {
		bindings[k] = v
	}
	return Context{bindings, c.settings}
}

// Evaluate evaluates an expression within the template context.
func (c Context) Evaluate(expr expressions.Expression) (out interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case expressions.InterpreterError:
				err = e
			default:
				// fmt.Println(string(debug.Stack()))
				panic(e)
			}
		}
	}()
	return expr.Evaluate(expressions.NewContext(c.bindings, c.settings.ExpressionSettings))
}
