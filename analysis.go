package liquid

import "github.com/osteele/liquid/render"

// VariableSegment is a path to a variable, represented as a slice of string segments.
// For example, the expression {{ customer.first_name }} produces ["customer", "first_name"].
type VariableSegment = []string

// GlobalVariableSegments returns paths of variables that are expected from the outer
// scope (i.e., not defined within the template itself via assign, capture, for, etc.).
//
// For example:
//
//	{{ customer.first_name }} {% assign x = "hello" %} {{ order.total }}
//	→ [["customer", "first_name"], ["order", "total"]]
//
// x does not appear because it is defined within the template.
func (e *Engine) GlobalVariableSegments(t *Template) ([]VariableSegment, error) {
	result := render.Analyze(t.root)
	return result.Globals, nil
}

// VariableSegments returns paths of all variables referenced in the template,
// including those defined locally by assign, capture, for, etc.
func (e *Engine) VariableSegments(t *Template) ([]VariableSegment, error) {
	result := render.Analyze(t.root)
	return result.All, nil
}

// GlobalVariableSegments returns paths of variables expected from the outer scope.
// It is a convenience method that delegates to Engine.GlobalVariableSegments.
func (t *Template) GlobalVariableSegments() ([]VariableSegment, error) {
	result := render.Analyze(t.root)
	return result.Globals, nil
}

// VariableSegments returns paths of all variables referenced in the template.
// It is a convenience method that delegates to Engine.VariableSegments.
func (t *Template) VariableSegments() ([]VariableSegment, error) {
	result := render.Analyze(t.root)
	return result.All, nil
}
