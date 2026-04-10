package render

import "github.com/osteele/liquid/parser"

// FilterStep records a single filter application during expression evaluation.
// Used internally by the render layer and re-exported by the liquid package.
type FilterStep struct {
	Filter string `json:"filter"`
	Args   []any  `json:"args"`
	Input  any    `json:"input"`
	Output any    `json:"output"`
}

// AuditComparison records a single primitive binary comparison inside a condition test.
type AuditComparison struct {
	Expression string // raw source text of this comparison; empty when not tracked
	Operator   string // "==", "!=", ">", "<", ">=", "<=", "contains"
	Left       any    // evaluated left operand
	Right      any    // evaluated right operand
	Result     bool   // outcome of this comparison
}

// AuditConditionNode is a node in a condition branch's items tree.
// Exactly one of Comparison or Group is non-nil.
type AuditConditionNode struct {
	Comparison *AuditComparison
	Group      *AuditGroup
}

// AuditGroup represents a logical and/or operator with its operands.
type AuditGroup struct {
	Operator string               // "and" | "or"
	Result   bool
	Items    []AuditConditionNode // sub-nodes (comparisons and nested groups)
}

// AuditBranch records a single branch of an {% if %}, {% unless %}, or {% case %} block.
type AuditBranch struct {
	Kind     string               // "if", "elsif", "else", "when", "unless"
	LocStart parser.SourceLoc     // start of the branch header tag
	LocEnd   parser.SourceLoc     // end of the branch header tag
	Source   string               // raw source of the branch header tag
	Executed bool                 // whether this branch's body was rendered
	Items    []AuditConditionNode // condition items tree (comparisons and groups); empty for "else"
}

// AuditIterInfo records metadata about a for/tablerow iteration block.
type AuditIterInfo struct {
	Variable    string
	Collection  string
	Length      int
	Limit       *int
	Offset      *int
	Reversed    bool
	Truncated   bool
	TracedCount int
}

// AuditHooks contains optional callback functions invoked during rendering for
// audit and trace collection. A nil pointer means no audit is active
// (zero-cost path on the normal render path).
//
// The struct also holds mutable state used by the render layer during a single
// render call; it must NOT be shared between concurrent renders.
type AuditHooks struct {
	// Callback functions — set once before the render begins, read-only during render:

	// OnObject is called when an {{ expr }} node is evaluated.
	// err is non-nil when evaluation failed; value will be nil in that case.
	// OnError is also called separately for error cases.
	OnObject func(start, end parser.SourceLoc, source, name string, parts []string, value any, pipeline []FilterStep, depth int, err error)

	// OnCondition is called for {% if %}, {% unless %}, {% case %} blocks.
	OnCondition func(start, end parser.SourceLoc, source string, branches []AuditBranch, depth int)

	// OnIteration is called for {% for %}, {% tablerow %} blocks.
	OnIteration func(start, end parser.SourceLoc, source string, it AuditIterInfo, depth int)

	// OnAssignment is called for {% assign %}.
	OnAssignment func(start, end parser.SourceLoc, source, varname string, path []string, value any, pipeline []FilterStep, depth int)

	// OnCapture is called for {% capture %}.
	OnCapture func(start, end parser.SourceLoc, source, varname, value string, depth int)

	// OnError is called when a render-time error is encountered.
	OnError func(start, end parser.SourceLoc, source string, err error)

	// OnWarning is called for render-time issues that are not fatal errors:
	// type-mismatch, not-iterable, and nil-dereference.
	// code is a machine-readable key; message is human-readable.
	OnWarning func(start, end parser.SourceLoc, source string, code, message string)

	// MaxIterItems limits how many loop iterations have their inner expressions
	// traced. 0 means unlimited. When the limit is reached, inner expressions
	// for subsequent iterations are not traced (hooks are not called for them).
	MaxIterItems int

	// Mutable render state — managed by the render layer, not the caller:

	// filterTarget is set by ObjectNode.render() before Evaluate() and cleared
	// after. The FilterHook in expressions.Config writes steps here.
	filterTarget *[]FilterStep

	// currentLocStart/End/Source track the source range of the node currently
	// being evaluated. Set by ObjectNode.render() (for nil-dereference) and
	// control_flow_tags (for type-mismatch) before Evaluate(), cleared after.
	currentLocStart  parser.SourceLoc
	currentLocEnd    parser.SourceLoc
	currentLocSource string

	// depth is incremented when entering a block body (via RenderBlock/RenderChildren)
	// and decremented on exit. Used to populate Expression.Depth in the public API.
	depth int

	// iterCount is a per-loop-depth iteration counter used for MaxIterItems.
	// It is set/reset by the loop tag renderer.
	iterCount int

	// suppressInner is true when MaxIterItems has been reached for the current
	// loop; the render layer skips calling hooks while it is set.
	suppressInner bool

	// conditionActive is the currently active items slice for collecting
	// condition nodes (comparisons and groups) during branch test evaluation.
	conditionActive *[]AuditConditionNode

	// conditionGroupStack holds parent slices suspended when BeginGroup is
	// called for a nested and/or sub-expression.
	conditionGroupStack []*[]AuditConditionNode

	// currentBranchSource holds the raw source text of the branch currently
	// being evaluated (e.g. "customer.age >= 18"). Read by AppendComparison
	// to populate AuditComparison.Expression.
	currentBranchSource string
}

// EmitWarning calls OnWarning if set. It is a no-op when audit is not active.
func (a *AuditHooks) EmitWarning(start, end parser.SourceLoc, source string, code, message string) {
	if a != nil && a.OnWarning != nil {
		a.OnWarning(start, end, source, code, message)
	}
}

// SetCurrentLoc stores the source range of the node currently being evaluated.
// Called before Evaluate() (by ObjectNode.render and control_flow_tags) and
// cleared after, so that TypeMismatchHook/NilDereferenceHook closures can read it.
func (a *AuditHooks) SetCurrentLoc(start, end parser.SourceLoc, source string) {
	if a != nil {
		a.currentLocStart = start
		a.currentLocEnd = end
		a.currentLocSource = source
	}
}

// CurrentLoc returns the source range stored by the most recent SetCurrentLoc call.
func (a *AuditHooks) CurrentLoc() (parser.SourceLoc, parser.SourceLoc, string) {
	if a == nil {
		return parser.SourceLoc{}, parser.SourceLoc{}, ""
	}
	return a.currentLocStart, a.currentLocEnd, a.currentLocSource
}

// SetConditionTarget sets up (target != nil) or tears down (target == nil)
// condition node collection for a single branch test evaluation.
// source is the raw source text of the branch test expression (e.g. "x >= 10");
// it is stored so that AppendComparison can populate AuditComparison.Expression.
func (a *AuditHooks) SetConditionTarget(target *[]AuditConditionNode) {
	if a == nil {
		return
	}
	if target != nil {
		*target = nil
		a.conditionActive = target
		a.conditionGroupStack = nil
	} else {
		a.conditionActive = nil
		a.conditionGroupStack = nil
		a.currentBranchSource = ""
		a.currentLocStart = parser.SourceLoc{}
		a.currentLocEnd = parser.SourceLoc{}
		a.currentLocSource = ""
	}
}

// SetBranchSource stores the raw source of the branch being evaluated.
// Called alongside SetConditionTarget so comparisons can reference it.
func (a *AuditHooks) SetBranchSource(source string) {
	if a != nil {
		a.currentBranchSource = source
	}
}

// AppendComparison appends a leaf comparison to the currently active collection.
// Called by the ComparisonHook wired in audit.go.
func (a *AuditHooks) AppendComparison(cmp AuditComparison) {
	if a == nil || a.conditionActive == nil {
		return
	}
	// For single-comparison branches the branch source IS the comparison expression.
	// For compound expressions (and/or groups) this will be the full compound string
	// which is still informative; sub-expression source is not tracked at this level.
	if cmp.Expression == "" {
		cmp.Expression = a.currentBranchSource
	}
	*a.conditionActive = append(*a.conditionActive, AuditConditionNode{Comparison: &cmp})
}

// BeginGroup is called before evaluating an and/or sub-expression's operands.
// It suspends the current collection and starts a fresh child collection.
func (a *AuditHooks) BeginGroup() {
	if a == nil || a.conditionActive == nil {
		return
	}
	a.conditionGroupStack = append(a.conditionGroupStack, a.conditionActive)
	newItems := []AuditConditionNode{}
	a.conditionActive = &newItems
}

// EndGroup is called after evaluating an and/or sub-expression.
// It pops the suspended parent, wraps the collected children in an AuditGroup,
// and appends the group as a node to the parent collection.
func (a *AuditHooks) EndGroup(op string, result bool) {
	if a == nil || len(a.conditionGroupStack) == 0 {
		return
	}
	children := *a.conditionActive
	n := len(a.conditionGroupStack)
	parent := a.conditionGroupStack[n-1]
	a.conditionGroupStack = a.conditionGroupStack[:n-1]
	a.conditionActive = parent
	group := &AuditGroup{Operator: op, Result: result, Items: children}
	*a.conditionActive = append(*a.conditionActive, AuditConditionNode{Group: group})
}

// Depth returns the current block nesting depth. 0 = top-level.
func (a *AuditHooks) Depth() int {
	if a == nil {
		return 0
	}
	return a.depth
}

// IterCount returns the number of loop iterations counted in the current loop.
func (a *AuditHooks) IterCount() int {
	if a == nil {
		return 0
	}
	return a.iterCount
}

// IncrIterCount increments the per-loop iteration counter.
func (a *AuditHooks) IncrIterCount() {
	if a != nil {
		a.iterCount++
	}
}

// SuppressInner reports whether hook calls for inner nodes should be suppressed.
func (a *AuditHooks) SuppressInner() bool {
	if a == nil {
		return false
	}
	return a.suppressInner
}

// SetSuppressInner sets the inner-suppression flag.
func (a *AuditHooks) SetSuppressInner(v bool) {
	if a != nil {
		a.suppressInner = v
	}
}

// ResetIterState resets iteration tracking for a new loop.
func (a *AuditHooks) ResetIterState() {
	if a != nil {
		a.iterCount = 0
		a.suppressInner = false
	}
}

// RestoreIterState restores iteration tracking state saved before entering a nested loop.
func (a *AuditHooks) RestoreIterState(iterCount int, suppressInner bool) {
	if a != nil {
		a.iterCount = iterCount
		a.suppressInner = suppressInner
	}
}

// SetFilterTarget sets the slice that the filter hook should write steps into.
// Called by ObjectNode.render() before Evaluate(); pass nil to clear.
func (a *AuditHooks) SetFilterTarget(target *[]FilterStep) {
	if a != nil {
		a.filterTarget = target
	}
}

// FilterTarget returns the current filter capture slice (nil when not capturing).
func (a *AuditHooks) FilterTarget() *[]FilterStep {
	if a == nil {
		return nil
	}
	return a.filterTarget
}
