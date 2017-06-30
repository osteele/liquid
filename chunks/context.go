package chunks

import (
	"github.com/osteele/liquid/expressions"
)

// Context is the evaluation context for chunk AST rendering.
type Context struct {
	vars     map[string]interface{}
	settings Settings
}

type Settings struct {
	ExpressionSettings expressions.Settings
	tags map[string]TagDefinition
	controlTags map[string]*controlTagDefinition
}

func (s Settings) AddFilter(name string, fn interface{}) {
	s.ExpressionSettings.AddFilter(name, fn)
}

func NewSettings() Settings {
	s:= Settings{
		expressions.NewSettings(),
		map[string]TagDefinition{},
		map[string]*controlTagDefinition{},
	}
	s.AddTag( "assign", assignTagDef)
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
	return expr.Evaluate(expressions.NewContext(c.vars, c.settings.ExpressionSettings))
}
