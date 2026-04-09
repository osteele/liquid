# Spec: Render Diagnostics — `RenderAudit`

## Objective

Add to the engine a method capable of returning, alongside the rendered output, a structured report of everything that happened during rendering: which variables were resolved to which values, which path each condition took, how many times each for loop iterated, and which state mutations occurred via `assign`/`capture`. Additionally, the same method can run structural validation of the template without needing a full render.

The output must be serialized as JSON with a stable structure, so any frontend or external tool can consume it without knowing the engine internals. The error and position design follows the **LSP Diagnostic** standard for editor compatibility.

---

## Position Structures

The current `SourceLoc` only has `Pathname` and `LineNo`. To support precise frontend highlighting (selecting exactly the characters of the expression), we need start and end column.

**Required scanner change:** track column offset during scanning (already simple because the scanner iterates char by char and increments `LineNo` on `\n` — just do the same for `ColNo`).

```go
// Position represents a point in the source (1-based, LSP-compatible).
type Position struct {
    Line   int `json:"line"`   // 1-based
    Column int `json:"column"` // 1-based
}

// Range is a span of the source (from Start to End, End exclusive).
type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}
```

---

## API

```go
// AuditOptions controls what RenderAudit collects.
// Does not duplicate engine/render options — behaviors like StrictVariables
// are passed via ...RenderOption, same as normal Render.
type AuditOptions struct {
    // --- Render trace ---
    TraceVariables   bool // Trace {{ expr }} with value and filter pipeline
    TraceConditions  bool // Trace {% if/unless/case %} with branch structure and comparisons
    TraceIterations  bool // Trace {% for/tablerow %} with loop metadata
    TraceAssignments bool // Trace {% assign %} and {% capture %} with resulting values

    // Maximum traced iterations per for/tablerow block.
    // 0 = no limit (beware of large loops).
    // Recommended: 100. When exceeded, the Truncated field of IterationTrace will be true.
    MaxIterationTraceItems int
}

// AuditResult is the complete result of RenderAudit.
// It is always returned non-nil, even when err != nil — the output may be
// partial if the render was interrupted, and Diagnostics explains what happened.
type AuditResult struct {
    Output      string       `json:"output"`      // Rendered HTML/text (possibly partial)
    Expressions []Expression `json:"expressions"` // Trace of all visited expressions, in execution order
    Diagnostics []Diagnostic `json:"diagnostics"` // Errors and warnings captured during execution
}

// AuditError is returned when the render encountered one or more errors.
// Implements error with a summary message, and exposes the individual errors
// as the same types that a normal render would return.
type AuditError struct {
    errors []SourceError
}

func (e *AuditError) Error() string         // "render completed with N error(s)"
func (e *AuditError) Errors() []SourceError // each item is UndefinedVariableError, RenderError, etc.
```

The method is added to `Template`, accepting the same `RenderOption`s that `Render` already accepts:

```go
// RenderAudit renders the template with vars and returns a structured trace of
// the entire execution alongside any errors that occurred.
//
// Unlike Render, RenderAudit does not stop at the first error — it accumulates
// all errors into the returned *AuditError while the render continues, producing
// as much output as possible. AuditResult is always non-nil; AuditResult.Output
// contains the (possibly partial) rendered string.
//
// *AuditError is nil when the render completed without errors. When non-nil,
// each individual error can be inspected with errors.As:
//
//	auditResult, auditErr := tpl.RenderAudit(vars, opts)
//	if auditErr != nil {
//	    for _, e := range auditErr.Errors() {
//	        var undVar *UndefinedVariableError
//	        var argErr *ArgumentError
//	        var renderErr *RenderError
//	        switch {
//	        case errors.As(e, &undVar):
//	            fmt.Printf("undefined variable %q at line %d\n", undVar.Variable, undVar.LineNumber())
//	        case errors.As(e, &argErr):
//	            fmt.Printf("argument error: %s\n", argErr.Error())
//	        case errors.As(e, &renderErr):
//	            fmt.Printf("render error at line %d: %s\n", renderErr.LineNumber(), renderErr.Message())
//	        }
//	    }
//	}
//
// The same errors are also available as Diagnostic entries in
// AuditResult.Diagnostics, with machine-readable codes and LSP-compatible ranges.
// Diagnostics that may appear during rendering:
//
//   - "argument-error" (error): a filter received invalid arguments
//     (e.g. divided_by: 0). The corresponding AuditError entry wraps *ArgumentError.
//   - "undefined-variable" (warning): a variable was not found in bindings.
//     Only emitted when WithStrictVariables() is active. Wraps *UndefinedVariableError.
//   - "type-mismatch" (warning): a comparison between incompatible types
//     (e.g. string vs int); Liquid evaluates it as false but it is likely a bug.
//   - "not-iterable" (warning): a {% for %} loop over a non-iterable value
//     (int, bool, string); Liquid iterates zero times silently.
//   - "nil-dereference" (warning): a chained property access where an intermediate
//     node in the path is nil (e.g. customer.address.city when address is nil);
//     the expression renders as empty string.
//
// opts controls what the trace collects (variables, conditions, iterations,
// assignments). renderOpts accepts the same options as Render —
// WithStrictVariables(), WithLaxFilters(), WithGlobals(), etc. — with identical
// semantics. RenderAudit never renders differently from Render given the same
// renderOpts.
func (t *Template) RenderAudit(vars Bindings, opts AuditOptions, renderOpts ...RenderOption) (*AuditResult, *AuditError)
```

And a static analysis method of the compiled AST, without rendering:

```go
func (t *Template) Validate() (*AuditResult, error)
```

**About `Validate()`:** severe structural errors (unclosed tag, invalid syntax) are already caught by `ParseTemplate` — if the template was created, those errors no longer exist. `Validate()` performs static analysis of the compiled AST to detect suspicious patterns like `empty-block`. It is a separate method, not a flag in `AuditOptions`.

**Parse errors in `AuditError`:** when `Validate()` detects a static problem, the error goes directly into `Diagnostics` and `AuditError.Errors()`. It does not appear in `Expression`, since there is no associated execution. `Expressions` reflects only what was visited during the render.

**About `*AuditError`:** it is always `nil` if the render completed without errors. When non-nil, `.Error()` gives the summary message and `.Errors()` gives the slice with individually typed errors — the same types that a normal `Render` would return if it stopped at the first error. Those who want to inspect each error iterate `.Errors()`.

**About `...RenderOption`:** `RenderAudit` is a normal render with observability added on top. Any `RenderOption` that works in `Render` works here the same way — `WithStrictVariables()`, `WithLaxFilters()`, `WithGlobals()`, anything. There are no audit-exclusive modes. The difference is that errors that would normally abort the render are captured as `Diagnostic` and the render continues — all errors accumulate in `AuditError.Errors()`.

---

## Expression Structure (Render Trace)

Each `Expression` item represents a Liquid tag or object that was visited during rendering. The `Kind` field is the discriminator. Exactly one of the optional fields will be populated.

```go
type ExpressionKind string

const (
    KindVariable   ExpressionKind = "variable"
    KindCondition  ExpressionKind = "condition"
    KindIteration  ExpressionKind = "iteration"
    KindAssignment ExpressionKind = "assignment"
    KindCapture    ExpressionKind = "capture"
)

type Expression struct {
    Source string         `json:"source"` // raw excerpt of the template, e.g. "{{ customer.name }}"
    Range  Range          `json:"range"`
    Kind   ExpressionKind `json:"kind"`

    // Depth indicates the nesting depth of this expression.
    // 0 = root level, 1 = inside {% if %} or {% for %}, 2 = inside nested block, etc.
    // Allows reconstructing the hierarchy from a flat array without nested JSON.
    Depth int `json:"depth"`

    // Error is populated if this expression generated an error during execution.
    // The same error also appears in the Diagnostics array of AuditResult for centralized scanning.
    Error *Diagnostic `json:"error,omitempty"`

    Variable   *VariableTrace   `json:"variable,omitempty"`
    Condition  *ConditionTrace  `json:"condition,omitempty"`
    Iteration  *IterationTrace  `json:"iteration,omitempty"`
    Assignment *AssignmentTrace `json:"assignment,omitempty"`
    Capture    *CaptureTrace    `json:"capture,omitempty"`
}
```

### VariableTrace

Represents a `{{ expr }}`. Beyond the final value, it records the filter pipeline step by step — useful to know exactly where in the filter chain a value broke or transformed.

```go
type VariableTrace struct {
    Name    string        `json:"name"`    // "customer.name"
    Parts   []string      `json:"parts"`   // ["customer", "name"]
    Value   any           `json:"value"`   // final value after all filters
    Pipeline []FilterStep `json:"pipeline"` // intermediate steps (empty if no filters)
}

type FilterStep struct {
    Filter string `json:"filter"` // filter name, e.g. "upcase"
    Args   []any  `json:"args"`   // arguments passed to the filter, e.g. [4, "..."]
    Input  any    `json:"input"`  // input value for this filter
    Output any    `json:"output"` // output value from this filter
}
```

**Example JSON:**
```json
{
  "source": "{{ customer.name | upcase | truncate: 10 }}",
  "range": { "start": {"line": 5, "column": 1}, "end": {"line": 5, "column": 47} },
  "kind": "variable",
  "variable": {
    "name": "customer.name",
    "parts": ["customer", "name"],
    "value": "JOHN...",
    "pipeline": [
      {
        "filter": "upcase",
        "args": [],
        "input": "john smith",
        "output": "JOHN SMITH"
      },
      {
        "filter": "truncate",
        "args": [10, "..."],
        "input": "JOHN SMITH",
        "output": "JOHN..."
      }
    ]
  }
}
```

### ConditionTrace

Represents an entire `{% if %}`, `{% unless %}`, or `{% case %}` block — from the header to the `{% endif %}`/`{% endcase %}`. Captures all branches (if, elsif, else) with their results, not just the winning branch.

```go
type ConditionTrace struct {
    // Branches lists all branches of the block, in declaration order.
    // For {% if %}…{% elsif %}…{% else %}…{% endif %}: one branch per clause.
    // For {% case %}…{% when %}…{% else %}…{% endcase %}: one branch per when/else.
    Branches []ConditionBranch `json:"branches"`
}

type ConditionBranch struct {
    Kind     string          `json:"kind"`            // "if" | "elsif" | "else" | "when" | "unless"
    Range    Range           `json:"range"`           // range of this clause's header
    Executed bool            `json:"executed"`        // did this branch's body execute?
    Items    []ConditionItem `json:"items,omitempty"` // comparisons (empty for "else")
}

// ConditionItem is a union: exactly one of the fields will be populated.
type ConditionItem struct {
    Comparison *ComparisonTrace `json:"comparison,omitempty"`
    Group      *GroupTrace      `json:"group,omitempty"`
}

type ComparisonTrace struct {
    Expression string `json:"expression"` // "customer.age >= 18"
    Left       any    `json:"left"`       // resolved value of the left side
    Operator   string `json:"operator"`   // "==", "!=", ">", ">=", "<", "<=", "contains"
    Right      any    `json:"right"`      // resolved value of the right side
    Result     bool   `json:"result"`
}

type GroupTrace struct {
    Operator string          `json:"operator"` // "and" | "or"
    Result   bool            `json:"result"`
    Items    []ConditionItem `json:"items"`
}
```

**Example JSON** for `{% if customer.age >= 18 and customer.active %}…{% elsif plan == "pro" %}…{% else %}…{% endif %}`:
```json
{
  "source": "{% if customer.age >= 18 and customer.active %}",
  "range": { "start": {"line": 10, "column": 1}, "end": {"line": 18, "column": 10} },
  "kind": "condition",
  "depth": 0,
  "condition": {
    "branches": [
      {
        "kind": "if",
        "range": { "start": {"line": 10, "column": 1}, "end": {"line": 10, "column": 48} },
        "executed": true,
        "items": [
          {
            "group": {
              "operator": "and",
              "result": true,
              "items": [
                {
                  "comparison": {
                    "expression": "customer.age >= 18",
                    "left": 25,
                    "operator": ">=",
                    "right": 18,
                    "result": true
                  }
                },
                {
                  "comparison": {
                    "expression": "customer.active",
                    "left": true,
                    "operator": "==",
                    "right": true,
                    "result": true
                  }
                }
              ]
            }
          }
        ]
      },
      {
        "kind": "elsif",
        "range": { "start": {"line": 13, "column": 1}, "end": {"line": 13, "column": 25} },
        "executed": false,
        "items": [
          {
            "comparison": {
              "expression": "plan == \"pro\"",
              "left": "basic",
              "operator": "==",
              "right": "pro",
              "result": false
            }
          }
        ]
      },
      {
        "kind": "else",
        "range": { "start": {"line": 16, "column": 1}, "end": {"line": 16, "column": 9} },
        "executed": false
      }
    ]
  }
}
```

**Note on `{% unless %}`:** the branch has `kind: "unless"`. The `executed` field already reflects the final result after the inversion.

**Note on `{% case/when %}`:** each `{% when %}` becomes a branch with `kind: "when"`. The `items` contain an implicit `ComparisonTrace` with `operator: "=="` comparing the `case` value against the `when` value. The `{% else %}` of the case becomes `kind: "else"` normally.

### IterationTrace

Represents a `{% for %}` or `{% tablerow %}` block. Emits **a single item** per block (not one per iteration), containing loop metadata. To inspect variables inside the loop, the `Expression`s of the inner iterations appear naturally in the sequence of the `expressions` array in the result.

To avoid memory explosion on large loops, the trace does not duplicate AST internal nodes per iteration — they already appear linearly in the array. The `MaxIterationTraceItems` control limits how many inner iterations are traced in the `expressions` array (not the `IterationTrace` itself).

```go
type IterationTrace struct {
    Variable   string `json:"variable"`             // loop variable name: "product"
    Collection string `json:"collection"`            // collection name: "products"
    Length     int    `json:"length"`                // total items in the collection
    Limit      *int   `json:"limit,omitempty"`       // limit value: if used
    Offset     *int   `json:"offset,omitempty"`      // offset value: if used
    Reversed   bool   `json:"reversed,omitempty"`    // true if reversed: was used
    Truncated  bool   `json:"truncated,omitempty"`   // true if MaxIterationTraceItems was reached
    TracedCount int   `json:"traced_count"`          // how many iterations were traced
}
```

**Example JSON:**
```json
{
  "source": "{% for product in cart.items limit:3 %}",
  "range": { "start": {"line": 20, "column": 1}, "end": {"line": 20, "column": 41} },
  "kind": "iteration",
  "iteration": {
    "variable": "product",
    "collection": "cart.items",
    "length": 10,
    "limit": 3,
    "offset": null,
    "reversed": false,
    "truncated": false,
    "traced_count": 3
  }
}
```

**Example with truncation** (loop with 5000 items, `MaxIterationTraceItems: 100`):
```json
{
  "iteration": {
    "variable": "row",
    "collection": "report.rows",
    "length": 5000,
    "truncated": true,
    "traced_count": 100
  }
}
```

### AssignmentTrace

Represents an `{% assign %}`. Captures the variable name, the resolved expression value, and the filter pipeline if any.

```go
type AssignmentTrace struct {
    Variable string        `json:"variable"` // assigned name: "total_price"
    Path     []string      `json:"path,omitempty"` // if dot notation: ["page", "title"]
    Value    any           `json:"value"`    // final value after filters
    Pipeline []FilterStep  `json:"pipeline"` // filter steps (empty if no filters)
}
```

**Example JSON:**
```json
{
  "source": "{% assign discounted = product.price | times: 0.9 | round %}",
  "range": { "start": {"line": 15, "column": 1}, "end": {"line": 15, "column": 59} },
  "kind": "assignment",
  "assignment": {
    "variable": "discounted",
    "value": 45,
    "pipeline": [
      {
        "filter": "times",
        "args": [0.9],
        "input": 50.0,
        "output": 45.0
      },
      {
        "filter": "round",
        "args": [],
        "input": 45.0,
        "output": 45
      }
    ]
  }
}
```

### CaptureTrace

Represents a `{% capture %}…{% endcapture %}` block. Captures the variable name and the resulting string value of the block.

```go
type CaptureTrace struct {
    Variable string `json:"variable"` // captured name: "email_body"
    Value    string `json:"value"`    // rendered content of the block
}
```

**Example JSON:**
```json
{
  "source": "{% capture greeting %}",
  "range": { "start": {"line": 3, "column": 1}, "end": {"line": 3, "column": 23} },
  "kind": "capture",
  "capture": {
    "variable": "greeting",
    "value": "Hello, John! Welcome back."
  }
}
```

---

## Diagnostics

Inspired by **LSP Diagnostic**. The `diagnostics` array is populated by errors that actually occur — the same errors that a normal render would raise, plus static analysis available through `Validate()`. The severity follows standard Liquid behavior: if Liquid would raise a real error, it is `error`; if Liquid would handle it silently but it is likely a bug, it is `warning`; if it is just an observation, it is `info`. There are no speculative diagnostics about what might fail.

```go
type DiagnosticSeverity string

const (
    SeverityError   DiagnosticSeverity = "error"
    SeverityWarning DiagnosticSeverity = "warning"
    SeverityInfo    DiagnosticSeverity = "info"
)

type Diagnostic struct {
    Range    Range              `json:"range"`
    Severity DiagnosticSeverity `json:"severity"`
    Code     string             `json:"code"`    // machine identifier (see catalog)
    Message  string             `json:"message"` // human-readable message
    Source   string             `json:"source"`  // raw excerpt of the template
    Related  []RelatedInfo      `json:"related,omitempty"`
}

type RelatedInfo struct {
    Range   Range  `json:"range"`
    Message string `json:"message"`
}
```

### Diagnostic Code Catalog

`error` for real errors that Liquid already raises. `warning` for silent behaviors that are probably template bugs — Liquid handles them silently, but an auditable engine should expose them. `info` for static observations.

| Code | Severity | Detected at | Description |
|---|---|---|---|
| `unclosed-tag` | error | parse (Validate) | `{% if %}` without a matching `{% endif %}` |
| `unexpected-tag` | error | parse (Validate) | `{% endif %}` without an opening `{% if %}` |
| `syntax-error` | error | parse (Validate) | Invalid syntax inside a tag or object |
| `undefined-filter` | error | parse (Validate) | Invoked filter is not registered in the engine |
| `argument-error` | error | render | Invalid arguments for filter or tag (e.g. `divided_by: 0`) |
| `undefined-variable` | warning | render (strict) | Variable not found in bindings — only when `WithStrictVariables()` is active |
| `type-mismatch` | warning | render | Comparison between incompatible types; Liquid evaluates as false but is likely a template bug |
| `not-iterable` | warning | render | `{% for %}` over a non-iterable value (int, bool, string); Liquid iterates zero times silently |
| `nil-dereference` | warning | render | Property access on nil in a chained path (e.g. `customer.address.city` when `address` is nil) |
| `empty-block` | info | parse (Validate) | `{% if %}…{% endif %}` block with no content |

**Note on simple `nil`:** nil variable in rendering (`{{ nil_var }}`) and nil in comparison (`{% if nil_var == x %}`) are normal and intentional Liquid behaviors — they produce an empty string and false respectively. They do not generate diagnostics. `nil-dereference` is specific to chained access where an intermediate node in the path is nil.

**Example JSON diagnostics:**

`unclosed-tag` with `Related` pointing to where the closing was expected:
```json
{
  "range": { "start": {"line": 8, "column": 1}, "end": {"line": 8, "column": 14} },
  "severity": "error",
  "code": "unclosed-tag",
  "message": "tag 'if' opened here was never closed",
  "source": "{% if order %}",
  "related": [
    {
      "range": { "start": {"line": 45, "column": 1}, "end": {"line": 45, "column": 1} },
      "message": "expected {% endif %} before end of template"
    }
  ]
}
```

`argument-error` on filter with division by zero:
```json
{
  "range": { "start": {"line": 9, "column": 5}, "end": {"line": 9, "column": 38} },
  "severity": "error",
  "code": "argument-error",
  "message": "divided_by: divided by 0",
  "source": "{{ product.price | divided_by: 0 }}"
}
```

`undefined-variable` (with `WithStrictVariables()` active):
```json
{
  "range": { "start": {"line": 22, "column": 5}, "end": {"line": 22, "column": 24} },
  "severity": "warning",
  "code": "undefined-variable",
  "message": "variable 'cart' is not defined",
  "source": "{{ cart.total }}"
}
```

`type-mismatch` during comparison:
```json
{
  "range": { "start": {"line": 12, "column": 4}, "end": {"line": 12, "column": 34} },
  "severity": "warning",
  "code": "type-mismatch",
  "message": "comparing string \"active\" with integer 1 using '=='; result is always false",
  "source": "{% if user.status == 1 %}"
}
```

`not-iterable` when the value is not a collection:
```json
{
  "range": { "start": {"line": 20, "column": 1}, "end": {"line": 20, "column": 32} },
  "severity": "warning",
  "code": "not-iterable",
  "message": "'order.status' is string \"pending\"; for loop iterates zero times",
  "source": "{% for item in order.status %}"
}
```

`nil-dereference` in chained path:
```json
{
  "range": { "start": {"line": 7, "column": 5}, "end": {"line": 7, "column": 32} },
  "severity": "warning",
  "code": "nil-dereference",
  "message": "'customer.address' is nil; 'city' access renders as empty string",
  "source": "{{ customer.address.city }}"
}
```

---

## Expression Array Order

The `expressions` array follows **execution order** (not declaration order), which is natural for flow tracing. This means:

- In an `{% if … %}…{% elsif … %}`, only the executed branch will have its inner `expressions` in the array. The `ConditionTrace` of the skipped `elsif` still appears (with `result: false`), but its children do not.
- In a `{% for %}`, inner expressions repeat for each iteration (up to `MaxIterationTraceItems`).

To correlate an `Expression` with its exact point in the original template, use the `Range`.

---

## Complete Example

**Template:**
```liquid
{% assign title = page.title | upcase %}
<h1>{{ title }}</h1>

{% if customer.age >= 18 %}
  <p>Welcome, {{ customer.name }}!</p>
{% else %}
  <p>Access restricted.</p>
{% endif %}

{% for item in cart.items %}
  <li>{{ item.name }} — ${{ item.price | times: 1.1 | round }}</li>
{% endfor %}
```

**Bindings:**
```json
{
  "page": {"title": "my store"},
  "customer": {"name": "John", "age": 25},
  "cart": {"items": [
    {"name": "T-Shirt", "price": 50},
    {"name": "Pants", "price": 120}
  ]}
}
```

**`AuditOptions`:**
```go
AuditOptions{
    TraceVariables:         true,
    TraceConditions:        true,
    TraceIterations:        true,
    TraceAssignments:       true,
    MaxIterationTraceItems: 100,
}
```

**`AuditResult`:**
```json
{
  "output": "<h1>MY STORE</h1>\n\n  <p>Welcome, John!</p>\n\n  <li>T-Shirt — $55</li>\n  <li>Pants — $132</li>\n",
  "expressions": [
    {
      "source": "{% assign title = page.title | upcase %}",
      "range": { "start": {"line": 1, "column": 1}, "end": {"line": 1, "column": 41} },
      "kind": "assignment",
      "assignment": {
        "variable": "title",
        "value": "MY STORE",
        "pipeline": [
          { "filter": "upcase", "args": [], "input": "my store", "output": "MY STORE" }
        ]
      }
    },
    {
      "source": "{{ title }}",
      "range": { "start": {"line": 2, "column": 5}, "end": {"line": 2, "column": 15} },
      "kind": "variable",
      "variable": { "name": "title", "parts": ["title"], "value": "MY STORE", "pipeline": [] }
    },
    {
      "source": "{% if customer.age >= 18 %}",
      "range": { "start": {"line": 4, "column": 1}, "end": {"line": 4, "column": 27} },
      "kind": "condition",
      "condition": {
        "result": true,
        "items": [
          {
            "comparison": {
              "expression": "customer.age >= 18",
              "left": 25,
              "operator": ">=",
              "right": 18,
              "result": true
            }
          }
        ]
      }
    },
    {
      "source": "{{ customer.name }}",
      "range": { "start": {"line": 5, "column": 15}, "end": {"line": 5, "column": 34} },
      "kind": "variable",
      "variable": { "name": "customer.name", "parts": ["customer", "name"], "value": "John", "pipeline": [] }
    },
    {
      "source": "{% for item in cart.items %}",
      "range": { "start": {"line": 10, "column": 1}, "end": {"line": 10, "column": 29} },
      "kind": "iteration",
      "iteration": {
        "variable": "item",
        "collection": "cart.items",
        "length": 2,
        "truncated": false,
        "traced_count": 2
      }
    },
    {
      "source": "{{ item.name }}",
      "range": { "start": {"line": 11, "column": 7}, "end": {"line": 11, "column": 21} },
      "kind": "variable",
      "variable": { "name": "item.name", "parts": ["item", "name"], "value": "T-Shirt", "pipeline": [] }
    },
    {
      "source": "{{ item.price | times: 1.1 | round }}",
      "range": { "start": {"line": 11, "column": 32}, "end": {"line": 11, "column": 69} },
      "kind": "variable",
      "variable": {
        "name": "item.price",
        "parts": ["item", "price"],
        "value": 55,
        "pipeline": [
          { "filter": "times", "args": [1.1], "input": 50, "output": 55.0 },
          { "filter": "round", "args": [], "input": 55.0, "output": 55 }
        ]
      }
    },
    {
      "source": "{{ item.name }}",
      "range": { "start": {"line": 11, "column": 7}, "end": {"line": 11, "column": 21} },
      "kind": "variable",
      "variable": { "name": "item.name", "parts": ["item", "name"], "value": "Pants", "pipeline": [] }
    },
    {
      "source": "{{ item.price | times: 1.1 | round }}",
      "range": { "start": {"line": 11, "column": 32}, "end": {"line": 11, "column": 69} },
      "kind": "variable",
      "variable": {
        "name": "item.price",
        "parts": ["item", "price"],
        "value": 132,
        "pipeline": [
          { "filter": "times", "args": [1.1], "input": 120, "output": 132.0 },
          { "filter": "round", "args": [], "input": 132.0, "output": 132 }
        ]
      }
    }
  ],
  "diagnostics": []
}
```

---

## Implementation Plan

### Phase 1 — Column Tracking in Scanner

File: `parser/scanner.go`, `parser/token.go`

- Add `ColNo int` to `SourceLoc`
- The scanner already increments `LineNo` when it encounters `\n`; just add `ColNo` and reset it to 1 on each new line
- Add `EndLoc SourceLoc` to `Token` to know where the token ends (not just starts)
- All `Range` builders in subsequent phases depend on this

**Impact:** localized change in the scanner. `SourceLoc.String()` can optionally include the column.

### Phase 2 — Parse Diagnostics

Files: `parser/parser.go`, `parser/error.go`, `liquid.go`

- Convert `parser.Error` into `[]Diagnostic` with complete `Range`
- Unclosed block errors already exist as `error`; wrap them in `Diagnostic{Code: "unclosed-tag"}`
- New method `Template.Validate(AuditOptions) (*AuditResult, error)` — without rendering

### Phase 3 — Render Trace

Files: `render/context.go`, `render/render.go`, new files `render/trace.go`, `render/trace_context.go`

- Create `traceContext` that wraps `render.Context` and implements `render.Context`
- Intercept `Evaluate()` to capture `VariableTrace`
- Intercept `if/unless/case` renderers to capture `ConditionTrace`
- Intercept `for/tablerow` renderer to capture `IterationTrace` and apply `MaxIterationTraceItems`
- Intercept `assign`/`capture` renderers to capture `AssignmentTrace`/`CaptureTrace`
- Intercept each `FilterStep` in the filter chain to populate `Pipeline`

### Phase 4 — Runtime Diagnostics

Files: `render/trace_context.go`

- Intercept real errors that occur during rendering and convert them into structured `Diagnostic` instead of aborting
- `UndefinedVariableError` (when StrictVariables active) → `Diagnostic{Code: "undefined-variable"}`, render continues
- Type incompatibility in comparisons → `Diagnostic{Code: "type-mismatch"}`
- `ArgumentError` (filter with invalid args, division by zero) → `Diagnostic{Code: "argument-error"}`
- `{% for %}` over non-iterable type → `Diagnostic{Code: "not-iterable"}`, iterates zero times (normal behavior)
- Property access on nil in chained path → `Diagnostic{Code: "nil-dereference"}`, returns empty (normal behavior)
- Accumulate all errors in `AuditError.errors`; derive the returned `*AuditError`

### Phase 5 — Final API and Public Exposure

File: `liquid.go`, `template.go`

- `Template.RenderAudit(vars Bindings, opts AuditOptions, renderOpts ...RenderOption) (*AuditResult, *AuditError)` — does everything
- `Template.Validate() (*AuditResult, error)` — static AST analysis only
- Types `AuditResult`, `AuditOptions`, `AuditError`, `Expression`, `Diagnostic`, `Position`, `Range` exported in `liquid.go`
- JSON tags on all types for direct serialization

---

## Design Notes

**Why there is no `Validate bool` in `AuditOptions`:** Severe structural errors (unclosed tag, invalid syntax) are already caught by `ParseTemplate` — if the template was successfully created, those errors no longer exist. There is nothing to structurally validate during a render audit. `Validate()` is a separate method for static analysis of the compiled AST.

**Why the returned error is `*AuditError` and not `SourceError`:** The render audit does not stop at the first error — it accumulates all encountered errors. The `*AuditError` reflects this: `.Error()` gives the summary, `.Errors()` gives the complete slice with the same types that a normal render would return one by one.

**Why nil and for-over-nil generate `warning` and not `error`:** These are silent Liquid behaviors that never abort the render — the warning severity reflects exactly this. Liquid silently returns empty or iterates zero times; the audit just makes this visible. `nil-dereference` is specific to chained paths (`a.b.c` where `b` is nil), not to a simple nil variable.

**Why PII is not addressed here:** The responsibility for redacting sensitive data is at the layer that calls the engine — whoever knows that `customer.ssn` is sensitive is the one building the bindings, not the engine. The engine exposes the trace; the caller decides what to log or transmit.

**Why `...RenderOption` and not new options:** The audit render is a normal render with observability added. Any engine `RenderOption` works here without any changes — `WithStrictVariables()`, `WithLaxFilters()`, `WithGlobals()`. There is no audit-exclusive mode. This guarantees behavioral parity: `RenderAudit` never renders differently from `Render` with the same options.

**Why `Depth` instead of nested JSON:** LSP uses `children []DocumentSymbol` (real nested JSON) in `DocumentSymbol`. For our case, the expressions array is an execution timeline — linear iteration is the primary use case, not tree navigation. With `depth`, the frontend iterates the array once and builds the tree if needed: children of a node are the next items with `depth = node.depth + 1` until the next item with `depth <= node.depth`. Nested JSON is not suitable for a timeline.

**Why `IterationTrace` does not duplicate the internal node per iteration:** If a for loop has 1000 items and the body has 10 expressions, tracing everything would produce 10,000 expression objects. `MaxIterationTraceItems` limits the number of traced iterations, keeping the array manageable. The `IterationTrace` itself always appears with `length` and `traced_count`.

**About `NodeID` for "Inspect Element":** A future extension would be to inject HTML comments `<!-- lid:RANGE -->` into the output when tracing is active, allowing a frontend to correlate a rendered HTML fragment with the `Range` of the expression that generated it. This is not in scope for this phase.
