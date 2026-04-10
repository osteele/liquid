package liquid

import (
	"strings"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// VariableSegment is a path to a variable, represented as a slice of string segments.
// For example, the expression {{ customer.first_name }} produces ["customer", "first_name"].
type VariableSegment = []string

// Variable represents a reference to a Liquid variable, including its source location
// and whether it comes from the outer (global) scope.
type Variable struct {
	// Segments is the full path to the variable, e.g. ["customer", "first_name"].
	Segments []string
	// Location is the source location where this variable reference appears.
	Location parser.SourceLoc
	// Global is true when the variable is not defined within the template itself
	// (i.e., it is expected to be provided by the caller).
	Global bool
}

// String returns the dot-joined path, e.g. "customer.first_name".
func (v Variable) String() string {
	return strings.Join(v.Segments, ".")
}

// StaticAnalysis is the rich result of statically analyzing a Liquid template.
type StaticAnalysis struct {
	// Variables contains all variable references found in the template,
	// including locally-defined ones (assign, capture, for loop variables, etc.).
	Variables []Variable

	// Globals contains only the variable references that are expected from the
	// outer scope — not defined within the template.
	Globals []Variable

	// Locals contains the names of variables defined within the template via
	// assign, capture, for, tablerow, etc.
	Locals []string

	// Tags contains the unique names of tags used in the template,
	// e.g. ["assign", "if", "for"].
	Tags []string

	// Filters is reserved for future use; currently always nil.
	Filters []string
}

// CompiledExpression is a compiled Liquid expression that can evaluate variable references.
// It is used when implementing custom tag/block analyzers via RegisterTagAnalyzer
// and RegisterBlockAnalyzer — pass it in NodeAnalysis.Arguments so the static
// analysis engine can walk its variable references.
type CompiledExpression = expressions.Expression

// ParseExpression parses a Liquid expression string into a CompiledExpression that can be
// used with RegisterTagAnalyzer / RegisterBlockAnalyzer. Returns an error if the
// expression contains a syntax error.
//
// Example:
//
//	e.RegisterTagAnalyzer("my_tag", func(args string) render.NodeAnalysis {
//	    expr, err := ParseExpression(args)
//	    if err != nil { return render.NodeAnalysis{} }
//	    return render.NodeAnalysis{Arguments: []CompiledExpression{expr}}
//	})
func ParseExpression(source string) (CompiledExpression, error) {
	return expressions.Parse(source)
}

// ── Engine methods ────────────────────────────────────────────────────────────

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

// GlobalVariables returns the unique root names of variables expected from the outer
// scope, without path details. For example, {{ customer.first_name }} contributes "customer".
func (e *Engine) GlobalVariables(t *Template) ([]string, error) {
	result := render.Analyze(t.root)
	return rootNames(result.Globals), nil
}

// Variables returns the unique root names of all variables referenced in the template,
// including locally-defined ones. For example, {{ x.a }} and {{ x.b }} both contribute "x".
func (e *Engine) Variables(t *Template) ([]string, error) {
	result := render.Analyze(t.root)
	return rootNames(result.All), nil
}

// GlobalFullVariables returns global variable references with full path and source location.
func (e *Engine) GlobalFullVariables(t *Template) ([]Variable, error) {
	result := render.Analyze(t.root)
	return refsToVariables(result.GlobalRefs, true), nil
}

// FullVariables returns all variable references with full path and source location.
// The Global field on each Variable indicates whether it comes from the outer scope.
func (e *Engine) FullVariables(t *Template) ([]Variable, error) {
	result := render.Analyze(t.root)
	return fullVariablesFromResult(result), nil
}

// Analyze performs a full static analysis of the template and returns a StaticAnalysis
// with variables (all and global), locally-defined names, and tag names used.
func (e *Engine) Analyze(t *Template) (*StaticAnalysis, error) {
	return analyzeTemplate(t), nil
}

// ParseAndAnalyze parses a template source and performs static analysis in one step.
// It is equivalent to calling ParseTemplate followed by Analyze.
func (e *Engine) ParseAndAnalyze(source []byte) (*Template, *StaticAnalysis, error) {
	tpl, err := e.ParseTemplate(source)
	if err != nil {
		return nil, nil, err
	}
	return tpl, analyzeTemplate(tpl), nil
}

// ── Template convenience methods ──────────────────────────────────────────────

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

// GlobalVariables returns the unique root names of global variables.
func (t *Template) GlobalVariables() ([]string, error) {
	result := render.Analyze(t.root)
	return rootNames(result.Globals), nil
}

// Variables returns the unique root names of all variables referenced in the template.
func (t *Template) Variables() ([]string, error) {
	result := render.Analyze(t.root)
	return rootNames(result.All), nil
}

// GlobalFullVariables returns global variable references with full path and source location.
func (t *Template) GlobalFullVariables() ([]Variable, error) {
	result := render.Analyze(t.root)
	return refsToVariables(result.GlobalRefs, true), nil
}

// FullVariables returns all variable references with full path and source location.
func (t *Template) FullVariables() ([]Variable, error) {
	result := render.Analyze(t.root)
	return fullVariablesFromResult(result), nil
}

// Analyze performs a full static analysis of the template.
func (t *Template) Analyze() (*StaticAnalysis, error) {
	return analyzeTemplate(t), nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// rootNames returns the unique first segments of a slice of variable paths.
func rootNames(paths [][]string) []string {
	seen := map[string]bool{}
	var result []string
	for _, path := range paths {
		if len(path) > 0 && !seen[path[0]] {
			seen[path[0]] = true
			result = append(result, path[0])
		}
	}
	return result
}

// refsToVariables converts a slice of VariableRef to a slice of Variable,
// setting the Global field uniformly to the provided value.
func refsToVariables(refs []render.VariableRef, global bool) []Variable {
	if len(refs) == 0 {
		return nil
	}
	vars := make([]Variable, len(refs))
	for i, ref := range refs {
		vars[i] = Variable{Segments: ref.Path, Location: ref.Loc, Global: global}
	}
	return vars
}

// fullVariablesFromResult converts AnalysisResult into a []Variable slice,
// marking each variable as global or not based on the Globals set.
func fullVariablesFromResult(result render.AnalysisResult) []Variable {
	if len(result.AllRefs) == 0 {
		return nil
	}
	globalSet := make(map[string]bool, len(result.Globals))
	for _, path := range result.Globals {
		globalSet[strings.Join(path, "\x00")] = true
	}
	vars := make([]Variable, len(result.AllRefs))
	for i, ref := range result.AllRefs {
		vars[i] = Variable{
			Segments: ref.Path,
			Location: ref.Loc,
			Global:   globalSet[strings.Join(ref.Path, "\x00")],
		}
	}
	return vars
}

// analyzeTemplate is the shared implementation for Engine.Analyze and Template.Analyze.
func analyzeTemplate(t *Template) *StaticAnalysis {
	result := render.Analyze(t.root)
	return &StaticAnalysis{
		Variables: fullVariablesFromResult(result),
		Globals:   refsToVariables(result.GlobalRefs, true),
		Locals:    result.Locals,
		Tags:      result.Tags,
	}
}
