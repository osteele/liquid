package liquid

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/osteele/liquid/expressions"
	"github.com/osteele/liquid/filters"
	"github.com/osteele/liquid/parser"
	"github.com/osteele/liquid/render"
)

// --------------------------------------------------------------------------
// Position / Range
// --------------------------------------------------------------------------

// Position represents a point in the source (1-based, LSP-compatible).
type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// Range is a span in the source from Start to End (End exclusive).
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func locToPos(loc parser.SourceLoc) Position {
	col := loc.ColNo
	if col == 0 {
		col = 1
	}
	line := loc.LineNo
	if line == 0 {
		line = 1
	}
	return Position{Line: line, Column: col}
}

func locsToRange(start, end parser.SourceLoc) Range {
	return Range{Start: locToPos(start), End: locToPos(end)}
}

// --------------------------------------------------------------------------
// AuditOptions
// --------------------------------------------------------------------------

// AuditOptions controls what RenderAudit collects.
// It does not duplicate engine/render options — behaviours like StrictVariables
// are passed via the ...RenderOption variadic, exactly like Render.
type AuditOptions struct {
	// --- Render trace ---
	TraceVariables   bool // Trace {{ expr }} with resolved value and filter pipeline
	TraceConditions  bool // Trace {% if/unless/case %} with branch structure
	TraceIterations  bool // Trace {% for/tablerow %} with loop metadata
	TraceAssignments bool // Trace {% assign %} and {% capture %} with resulting values

	// MaxIterationTraceItems limits how many loop iterations have their inner
	// expressions traced. 0 means unlimited.
	// When the limit is reached, the IterationTrace.Truncated field is set to true.
	MaxIterationTraceItems int
}

// --------------------------------------------------------------------------
// FilterStep
// --------------------------------------------------------------------------

// FilterStep records a single filter application in a pipeline.
type FilterStep = render.FilterStep

// --------------------------------------------------------------------------
// Diagnostic
// --------------------------------------------------------------------------

// DiagnosticSeverity indicates how serious a diagnostic is.
type DiagnosticSeverity string

const (
	SeverityError   DiagnosticSeverity = "error"
	SeverityWarning DiagnosticSeverity = "warning"
	SeverityInfo    DiagnosticSeverity = "info"
)

// Diagnostic represents an error, warning, or informational message tied to a
// source location. The design follows the LSP Diagnostic pattern.
type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	Code     string             `json:"code"`
	Message  string             `json:"message"`
	Source   string             `json:"source"`
	Related  []RelatedInfo      `json:"related,omitempty"`
}

// RelatedInfo is supplementary information for a Diagnostic (e.g. where a
// matching opening tag is located when reporting an unclosed-tag error).
type RelatedInfo struct {
	Range   Range  `json:"range"`
	Message string `json:"message"`
}

// --------------------------------------------------------------------------
// Expression kinds  (discriminated union)
// --------------------------------------------------------------------------

// ExpressionKind is the discriminator for an Expression.
type ExpressionKind string

const (
	KindVariable   ExpressionKind = "variable"
	KindCondition  ExpressionKind = "condition"
	KindIteration  ExpressionKind = "iteration"
	KindAssignment ExpressionKind = "assignment"
	KindCapture    ExpressionKind = "capture"
)

// Expression represents a single Liquid construct visited during rendering.
// Exactly one of the optional trace fields is populated, selected by Kind.
type Expression struct {
	Source string         `json:"source"`
	Range  Range          `json:"range"`
	Kind   ExpressionKind `json:"kind"`

	// Depth is the block-nesting depth at which this expression was evaluated.
	// 0 = top level, 1 = inside one {% if %} or {% for %}, etc.
	Depth int `json:"depth"`

	// Error is populated when this expression caused a runtime error.
	// The same error also appears in AuditResult.Diagnostics.
	Error *Diagnostic `json:"error,omitempty"`

	Variable   *VariableTrace   `json:"variable,omitempty"`
	Condition  *ConditionTrace  `json:"condition,omitempty"`
	Iteration  *IterationTrace  `json:"iteration,omitempty"`
	Assignment *AssignmentTrace `json:"assignment,omitempty"`
	Capture    *CaptureTrace    `json:"capture,omitempty"`
}

// --------------------------------------------------------------------------
// VariableTrace
// --------------------------------------------------------------------------

// VariableTrace is produced by {{ expr }}.
type VariableTrace struct {
	Name     string       `json:"name"`
	Parts    []string     `json:"parts"`
	Value    any          `json:"value"`
	Pipeline []FilterStep `json:"pipeline"`
}

// --------------------------------------------------------------------------
// ConditionTrace
// --------------------------------------------------------------------------

// ConditionTrace is produced by {% if %}, {% unless %}, or {% case %}.
type ConditionTrace struct {
	Branches []ConditionBranch `json:"branches"`
}

// ComparisonTrace records a single primitive binary comparison in a condition.
type ComparisonTrace struct {
	Expression string `json:"expression"` // raw source text of this comparison
	Left       any    `json:"left"`
	Operator   string `json:"operator"` // "==", "!=", ">", "<", ">=", "<=", "contains"
	Right      any    `json:"right"`
	Result     bool   `json:"result"`
}

// GroupTrace represents a logical and/or operator with its operands.
type GroupTrace struct {
	Operator string          `json:"operator"` // "and" | "or"
	Result   bool            `json:"result"`
	Items    []ConditionItem `json:"items"`
}

// ConditionItem is a union node in a condition branch's items tree.
// Exactly one of Comparison or Group is non-nil.
type ConditionItem struct {
	Comparison *ComparisonTrace `json:"comparison,omitempty"`
	Group      *GroupTrace      `json:"group,omitempty"`
}

// ConditionBranch represents one branch (if / elsif / else / unless / when) of
// a condition block.
type ConditionBranch struct {
	Kind     string          `json:"kind"`
	Range    Range           `json:"range"`
	Executed bool            `json:"executed"`
	Items    []ConditionItem `json:"items,omitempty"` // condition items tree; empty for else
}

// --------------------------------------------------------------------------
// IterationTrace
// --------------------------------------------------------------------------

// IterationTrace is produced by {% for %} or {% tablerow %}.
type IterationTrace struct {
	Variable    string `json:"variable"`
	Collection  string `json:"collection"`
	Length      int    `json:"length"`
	Limit       *int   `json:"limit,omitempty"`
	Offset      *int   `json:"offset,omitempty"`
	Reversed    bool   `json:"reversed,omitempty"`
	Truncated   bool   `json:"truncated,omitempty"`
	TracedCount int    `json:"traced_count"`
}

// --------------------------------------------------------------------------
// AssignmentTrace
// --------------------------------------------------------------------------

// AssignmentTrace is produced by {% assign %}.
type AssignmentTrace struct {
	Variable string       `json:"variable"`
	Path     []string     `json:"path,omitempty"`
	Value    any          `json:"value"`
	Pipeline []FilterStep `json:"pipeline"`
}

// --------------------------------------------------------------------------
// CaptureTrace
// --------------------------------------------------------------------------

// CaptureTrace is produced by {% capture %}…{% endcapture %}.
type CaptureTrace struct {
	Variable string `json:"variable"`
	Value    string `json:"value"`
}

// --------------------------------------------------------------------------
// AuditResult
// --------------------------------------------------------------------------

// AuditResult is the structured output of RenderAudit.
// It is always non-nil, even when an error was returned — Output may be partial
// and Diagnostics explains what happened.
type AuditResult struct {
	Output      string            `json:"output"`
	Expressions []Expression `json:"expressions"`
	Diagnostics []Diagnostic      `json:"diagnostics"`
}

// --------------------------------------------------------------------------
// AuditError
// --------------------------------------------------------------------------

// AuditError is returned by RenderAudit when one or more runtime errors were
// encountered during rendering. It implements the error interface and exposes
// the individual typed errors via Errors().
type AuditError struct {
	errors []SourceError
}

func (e *AuditError) Error() string {
	n := len(e.errors)
	if n == 1 {
		return fmt.Sprintf("render completed with 1 error: %s", e.errors[0].Error())
	}
	return fmt.Sprintf("render completed with %d errors", n)
}

// Errors returns the individual errors that were encountered during the render.
// Each element is typed (e.g. *render.UndefinedVariableError) and is the same
// kind of error that a normal Render would return.
func (e *AuditError) Errors() []SourceError {
	return e.errors
}

// --------------------------------------------------------------------------
// RenderAudit — wires up AuditHooks and collects results
// --------------------------------------------------------------------------

// RenderAudit executes the template with the given variable bindings and collects
// a structured trace of everything that occurred during rendering.
//
// It accepts the same RenderOption values as Render. For example, pass
// WithStrictVariables() to detect undefined variable accesses.
//
// The returned *AuditResult is always non-nil. When errors occurred, a non-nil
// *AuditError is also returned with all individual errors accessible via Errors().
func (t *Template) RenderAudit(vars Bindings, opts AuditOptions, renderOpts ...RenderOption) (*AuditResult, *AuditError) {
	result := &AuditResult{}
	var auditErrs []SourceError

	// Wire up AuditHooks.
	hooks := &render.AuditHooks{
		MaxIterItems: opts.MaxIterationTraceItems,
	}

	if opts.TraceVariables {
		hooks.OnObject = func(start, end parser.SourceLoc, source, name string, parts []string, value any, pipeline []render.FilterStep, depth int, nodeErr error) {
			expr := Expression{
				Source: source,
				Range:  locsToRange(start, end),
				Kind:   KindVariable,
				Depth:  depth,
				Variable: &VariableTrace{
					Name:     name,
					Parts:    parts,
					Value:    value,
					Pipeline: pipeline,
				},
			}
			if nodeErr != nil && len(result.Diagnostics) > 0 {
				// OnError has already appended the Diagnostic to result.Diagnostics.
				// Point Expression.Error at the same item so they are identical.
				expr.Error = &result.Diagnostics[len(result.Diagnostics)-1]
			}
			result.Expressions = append(result.Expressions, expr)
		}
	}

	if opts.TraceConditions {
		hooks.OnCondition = func(start, end parser.SourceLoc, source string, branches []render.AuditBranch, depth int) {
			cb := make([]ConditionBranch, len(branches))
			for i, b := range branches {
				cb[i] = ConditionBranch{
					Kind:     b.Kind,
					Range:    locsToRange(b.LocStart, b.LocEnd),
					Executed: b.Executed,
					Items:    mapConditionItems(b.Items),
				}
			}
			result.Expressions = append(result.Expressions, Expression{
				Source: source,
				Range:  locsToRange(start, end),
				Kind:   KindCondition,
				Depth:  depth,
				Condition: &ConditionTrace{
					Branches: cb,
				},
			})
		}
	}

	if opts.TraceIterations {
		hooks.OnIteration = func(start, end parser.SourceLoc, source string, it render.AuditIterInfo, depth int) {
			var limitPtr *int
			if it.Limit != nil {
				v := *it.Limit
				limitPtr = &v
			}
			var offsetPtr *int
			if it.Offset != nil {
				v := *it.Offset
				offsetPtr = &v
			}
			result.Expressions = append(result.Expressions, Expression{
				Source: source,
				Range:  locsToRange(start, end),
				Kind:   KindIteration,
				Depth:  depth,
				Iteration: &IterationTrace{
					Variable:    it.Variable,
					Collection:  it.Collection,
					Length:      it.Length,
					Limit:       limitPtr,
					Offset:      offsetPtr,
					Reversed:    it.Reversed,
					Truncated:   it.Truncated,
					TracedCount: it.TracedCount,
				},
			})
		}
	}

	if opts.TraceAssignments {
		hooks.OnAssignment = func(start, end parser.SourceLoc, source, varname string, path []string, value any, pipeline []render.FilterStep, depth int) {
			result.Expressions = append(result.Expressions, Expression{
				Source: source,
				Range:  locsToRange(start, end),
				Kind:   KindAssignment,
				Depth:  depth,
				Assignment: &AssignmentTrace{
					Variable: varname,
					Path:     path,
					Value:    value,
					Pipeline: pipeline,
				},
			})
		}

		hooks.OnCapture = func(start, end parser.SourceLoc, source, varname, value string, depth int) {
			result.Expressions = append(result.Expressions, Expression{
				Source: source,
				Range:  locsToRange(start, end),
				Kind:   KindCapture,
				Depth:  depth,
				Capture: &CaptureTrace{
					Variable: varname,
					Value:    value,
				},
			})
		}
	}

	hooks.OnError = func(start, end parser.SourceLoc, source string, err error) {
		if se, ok := err.(SourceError); ok {
			auditErrs = append(auditErrs, se)
		}
		code, severity := diagCodeForError(err)
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Range:    locsToRange(start, end),
			Severity: severity,
			Code:     code,
			Message:  err.Error(),
			Source:   source,
		})
	}

	// Apply render options and attach the audit hooks.
	cfg := t.applyRenderOptions(renderOpts)
	cfg.Audit = hooks

	// ExceptionHandler swallows render-time errors so that rendering continues
	// past failing nodes, accumulating all errors rather than stopping at the first.
	// OnError (called from within the failing node's render) has already recorded
	// the error into Diagnostics and auditErrs before ExceptionHandler fires.
	cfg.ExceptionHandler = func(err error) string { return "" }

	// Also wire the filter hook into the expressions config so filter steps
	// are captured inside the expressions evaluator.
	cfg.FilterHook = func(name string, input any, args []any, output any) {
		if t := hooks.FilterTarget(); t != nil {
			*t = append(*t, render.FilterStep{
				Filter: name,
				Args:   args,
				Input:  input,
				Output: output,
			})
		}
	}

	// Wire comparison and group hooks for condition branch tracing.
	if opts.TraceConditions {
		cfg.ComparisonHook = func(op string, left, right any, result bool) {
			hooks.AppendComparison(render.AuditComparison{
				Operator: op,
				Left:     left,
				Right:    right,
				Result:   result,
			})
		}
		cfg.ComparisonGroupBeginHook = func() {
			hooks.BeginGroup()
		}
		cfg.ComparisonGroupEndHook = func(op string, result bool) {
			hooks.EndGroup(op, result)
		}
	}

	// Wire type-mismatch and nil-dereference hooks. These always fire when audit
	// is active, regardless of individual trace options, since they are warnings
	// about template bugs rather than trace data.
	cfg.TypeMismatchHook = func(op string, left, right any) {
		start, end, src := hooks.CurrentLoc()
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Range:    locsToRange(start, end),
			Severity: SeverityWarning,
			Code:     "type-mismatch",
			Message:  fmt.Sprintf("comparing %T with %T using %q; result is always false", left, right, op),
			Source:   src,
		})
	}
	cfg.NilDereferenceHook = func(object, property string) {
		start, end, src := hooks.CurrentLoc()
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Range:    locsToRange(start, end),
			Severity: SeverityWarning,
			Code:     "nil-dereference",
			Message:  fmt.Sprintf("nil intermediate in path; %q access renders as empty string", property),
			Source:   src,
		})
	}

	// Wire OnWarning (not-iterable, etc.) to Diagnostics.
	hooks.OnWarning = func(start, end parser.SourceLoc, source string, code, message string) {
		result.Diagnostics = append(result.Diagnostics, Diagnostic{
			Range:    locsToRange(start, end),
			Severity: SeverityWarning,
			Code:     code,
			Message:  message,
			Source:   source,
		})
	}

	buf := new(bytes.Buffer)
	renderErr := render.Render(t.root, buf, vars, cfg)
	result.Output = buf.String()

	if renderErr != nil {
		// Add the terminal error if it wasn't already collected by OnError.
		if len(auditErrs) == 0 {
			auditErrs = append(auditErrs, renderErr)
		}
	}

	var ae *AuditError
	if len(auditErrs) > 0 {
		ae = &AuditError{errors: auditErrs}
	}

	return result, ae
}

// --------------------------------------------------------------------------
// Validate — static analysis without rendering
// --------------------------------------------------------------------------

// Validate performs static analysis on the compiled template AST and returns
// any diagnostics found. It does not execute the template.
//
// Note: fatal parse errors (unclosed tags, syntax errors) are caught at
// Engine.ParseTemplate time and will never appear here. Validate reports
// structural patterns that are valid syntax but likely bugs, such as empty
// blocks.
func (t *Template) Validate() (*AuditResult, error) {
	result := &AuditResult{}

	// Walk the AST via the public Walk method.
	result.Diagnostics = t.collectValidationDiags()

	return result, nil
}

// visitNodeForValidation walks the render tree collecting static diagnostics.
// It uses the public ParseTree method on the template, so this is called as a method.
func (t *Template) collectValidationDiags() []Diagnostic {
	var diags []Diagnostic

	var walkTree func(node *TemplateNode)
	walkTree = func(node *TemplateNode) {
		if node == nil {
			return
		}
		if node.Kind == TemplateNodeBlock {
			name := node.TagName
			if len(node.Children) == 0 && (name == "if" || name == "unless" || name == "for" || name == "case") {
				diags = append(diags, Diagnostic{
					Range:    Range{Start: locToPos(node.Location), End: locToPos(node.Location)},
					Severity: SeverityInfo,
					Code:     "empty-block",
					Message:  fmt.Sprintf("empty %q block", name),
				})
			}
		}
		for _, child := range node.Children {
			walkTree(child)
		}
	}

	walkTree(t.ParseTree())

	// Check for undefined filters by walking the internal render node tree.
	var checkFilters func(n render.Node)
	checkFilters = func(n render.Node) {
		if n == nil {
			return
		}
		switch n := n.(type) {
		case *render.SeqNode:
			for _, child := range n.Children {
				checkFilters(child)
			}
		case *render.ObjectNode:
			expr := n.GetExpr()
			if expr != nil {
				for _, fname := range expressions.FilterNames(expr) {
					if !t.cfg.HasFilter(fname) {
						diags = append(diags, Diagnostic{
							Range:    locsToRange(n.SourceLoc, n.EndLoc),
							Severity: SeverityError,
							Code:     "undefined-filter",
							Message:  fmt.Sprintf("undefined filter %q", fname),
							Source:   n.Source,
						})
					}
				}
			}
		case *render.BlockNode:
			for _, child := range n.Body {
				checkFilters(child)
			}
			for _, clause := range n.Clauses {
				checkFilters(clause)
			}
		}
	}
	checkFilters(t.root)

	return diags
}

// --------------------------------------------------------------------------
// Internal helpers
// --------------------------------------------------------------------------

// mapConditionItems recursively converts a slice of render.AuditConditionNode
// (internal representation) to the public ConditionItem slice.
func mapConditionItems(nodes []render.AuditConditionNode) []ConditionItem {
	if len(nodes) == 0 {
		return nil
	}
	items := make([]ConditionItem, len(nodes))
	for i, n := range nodes {
		if n.Comparison != nil {
			items[i] = ConditionItem{
				Comparison: &ComparisonTrace{
					Expression: n.Comparison.Expression,
					Operator:   n.Comparison.Operator,
					Left:       n.Comparison.Left,
					Right:      n.Comparison.Right,
					Result:     n.Comparison.Result,
				},
			}
		} else if n.Group != nil {
			items[i] = ConditionItem{
				Group: &GroupTrace{
					Operator: n.Group.Operator,
					Result:   n.Group.Result,
					Items:    mapConditionItems(n.Group.Items),
				},
			}
		}
	}
	return items
}

// diagCodeForError maps a render-time error to an LSP-style diagnostic code
// and severity. The mapping follows the spec's error catalogue.
func diagCodeForError(err error) (code string, severity DiagnosticSeverity) {
	var undefinedVar *render.UndefinedVariableError
	var argErr *render.ArgumentError
	var zeroDivErr *filters.ZeroDivisionError
	switch {
	case errors.As(err, &undefinedVar):
		return "undefined-variable", SeverityWarning
	case errors.As(err, &zeroDivErr):
		return "argument-error", SeverityError
	case errors.As(err, &argErr):
		return "argument-error", SeverityError
	default:
		return "render-error", SeverityError
	}
}

// --------------------------------------------------------------------------
// expressions.Config adapter — the FilterHook type
// --------------------------------------------------------------------------

// Ensure the expressions package is imported (used indirectly via render.Config).
var _ = expressions.Config{}
