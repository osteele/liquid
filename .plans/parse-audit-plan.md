# Spec: Parse Diagnostics — `ParseTemplateAudit`

## Objective

Add to the engine a method that parses a template source and returns a structured report of everything that went wrong (or is suspicious) at parse time — without needing a render. The result includes:

- A compiled `*Template` when parsing succeeds (possibly with non-fatal errors)
- A `[]Diagnostic` with every parse error and static analysis warning, in source order
- `nil` template only when a structural error makes compilation impossible

The `Diagnostic` type is **the same type** used by `RenderAudit` — no new types are introduced. The caller always gets `*ParseResult`, never a raw `SourceError`. This makes the `ParseTemplateAudit → RenderAudit` pipeline fully uniform: both phases speak the same `[]Diagnostic` language.

---

## API

```go
// ParseResult is the result of ParseTemplateAudit.
// Template is non-nil if parsing produced a usable compiled template.
// Template is nil only when a structural (fatal) error prevented compilation.
// Diagnostics is always non-nil; it is empty when there are no issues.
type ParseResult struct {
    Template    *Template    `json:"template,omitempty"` // nil on fatal error
    Diagnostics []Diagnostic `json:"diagnostics"`
}

// ParseTemplateAudit parses source and returns a *ParseResult containing
// the compiled template (when parsing succeeds) and all parse-time diagnostics.
//
// Unlike ParseTemplate, ParseTemplateAudit never returns a SourceError.
// All problems are captured as Diagnostic entries in ParseResult.Diagnostics,
// using the same Diagnostic type used by (*Template).RenderAudit.
//
// ParseResult.Template is non-nil when parsing produced a usable compiled
// template. Callers should check Template before rendering:
//
//	result := eng.ParseTemplateAudit(source)
//	for _, d := range result.Diagnostics {
//	    log.Printf("%s at line %d: %s", d.Severity, d.Range.Start.Line, d.Message)
//	}
//	if result.Template != nil {
//	    auditResult, auditErr := result.Template.RenderAudit(binds, opts)
//	    _ = auditResult
//	    _ = auditErr
//	}
//
// Diagnostics that may appear in ParseResult.Diagnostics:
//
//   - "unclosed-tag" (error): a block tag was opened but never closed;
//     ParseResult.Template is nil when this occurs.
//   - "unexpected-tag" (error): a closing or clause tag appeared without a
//     matching open block; ParseResult.Template is nil when this occurs.
//   - "syntax-error" (error): invalid expression inside {{ }} or tag args;
//     the offending token is replaced with a no-op node and parsing continues,
//     so ParseResult.Template may still be non-nil.
//   - "undefined-filter" (error): a filter name used in an expression is not
//     registered in this engine; detected by static AST walk, no render needed.
//   - "empty-block" (info): a block tag (if, for, etc.) has no content in
//     any branch.
//
// Each Diagnostic.Range uses 1-based line and column numbers (LSP-compatible).
// Diagnostic.Source contains the raw template excerpt that produced the issue.
func (e *Engine) ParseTemplateAudit(source []byte) *ParseResult

// ParseStringAudit is the string-input convenience variant of ParseTemplateAudit.
func (e *Engine) ParseStringAudit(source string) *ParseResult
```

### Shared types (already defined in the audit package)

`ParseResult` reuses the exact same types from the `RenderAudit` feature:

```go
// Position, Range, Diagnostic, DiagnosticSeverity, RelatedInfo
// — defined in liquid.go as part of the RenderAudit API, not duplicated here.
```

---

## Fatal vs Non-Fatal Errors

The key design decision is which errors abort compilation and which produce a `Diagnostic` + continue.

### Fatal — `Template = nil`

Structural errors where the parser cannot build a coherent AST. The rest of the template is semantically undefined after this point.

| Situation | Code |
|---|---|
| `{% if %}` without `{% endif %}` (stack non-empty at EOF) | `unclosed-tag` |
| `{% elsif %}` / `{% else %}` / `{% endfor %}` outside a matching open block | `unexpected-tag` |

These are the only two `return nil, err` paths in `parseTokens` today. They remain fatal.

### Non-Fatal — `Template != nil`, diagnostic emitted, parsing continues

Errors isolated to a single token. The block structure is intact; the bad token can be replaced by a no-op broken node and parsing continues to the next token.

| Situation | Code | Recovery |
|---|---|---|
| `{{ expr }}` where `expressions.Parse(args)` fails | `syntax-error` | Emit broken `ASTObject` that renders empty string |
| `{% assign x = \| bad %}` — expression in tag args is invalid at compile time | `syntax-error` | Emit broken `ASTTag` that renders empty string |

### Static analysis — `Template != nil`, emitted by AST walk after parse

Detected by walking the compiled AST. No render is needed.

| Situation | Code |
|---|---|
| Filter name used in any expression is not registered in the engine | `undefined-filter` |
| `{% if %}…{% endif %}` block with no content in any branch | `empty-block` |

---

## Diagnostic Code Catalog (parse-time)

Reuses the same codes from the `RenderAudit` catalog. Only the parse-time subset is listed here.

| Code | Severity | Description |
|---|---|---|
| `unclosed-tag` | error | `{% if %}` opened but never closed (fatal) |
| `unexpected-tag` | error | `{% endif %}` with no matching `{% if %}` (fatal) |
| `syntax-error` | error | Invalid expression inside `{{ }}` or tag args (non-fatal) |
| `undefined-filter` | error | Filter name used in expression is not registered |
| `empty-block` | info | `{% if %}` / `{% for %}` / etc. block with no content in any branch |

### `unclosed-tag` with `Related`

The `related` field points to the end-of-template position where the closing tag was expected, matching the LSP model from `render-audit-plan.md`.

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

### `syntax-error` on broken expression

```json
{
  "range": { "start": {"line": 3, "column": 5}, "end": {"line": 3, "column": 30} },
  "severity": "error",
  "code": "syntax-error",
  "message": "unexpected token '|' in expression",
  "source": "{{ product.price | | round }}"
}
```

### `undefined-filter`

```json
{
  "range": { "start": {"line": 7, "column": 1}, "end": {"line": 7, "column": 40} },
  "severity": "error",
  "code": "undefined-filter",
  "message": "filter 'my_custom' is not registered",
  "source": "{{ order.total | my_custom | round }}"
}
```

### `empty-block`

```json
{
  "range": { "start": {"line": 12, "column": 1}, "end": {"line": 14, "column": 10} },
  "severity": "info",
  "code": "empty-block",
  "message": "block 'if' has no content in any branch",
  "source": "{% if debug %}"
}
```

---

## ASTBrokenNode

To enable non-fatal parse recovery, a new AST node type is needed:

```go
// ASTBroken is a node that failed to compile but does not break the block structure.
// It renders as an empty string. The parser emits a Diagnostic and continues.
type ASTBroken struct {
    Token     // carries SourceLoc, EndLoc, Source
    ParseErr  error // original compile-time error
}
```

The render path for `ASTBroken` is a no-op (output nothing, no runtime error). The error was already captured as a `Diagnostic` at parse time. `RenderAudit` does not interact with `ASTBroken` beyond seeing its empty output.

---

## Complete Example

**Template with one fatal error:**

```liquid
{% if customer.vip %}
  Welcome back!
```

```go
result := eng.ParseTemplateAudit([]byte(source))
// result.Template == nil
// result.Diagnostics == [
//   { code: "unclosed-tag", severity: "error", source: "{% if customer.vip %}", ... }
// ]
```

**Template with non-fatal + static errors:**

```liquid
Hello {{ user.name | no_such_filter }}!
{% if %}{% endif %}
```

```go
result := eng.ParseTemplateAudit([]byte(source))
// result.Template != nil  (parsing continued, template is usable)
// result.Diagnostics == [
//   { code: "undefined-filter", severity: "error",  source: "{{ user.name | no_such_filter }}", ... },
//   { code: "syntax-error",     severity: "error",  source: "{% if %}", ... },
//   { code: "empty-block",      severity: "info",   source: "{% if %}", ... },
// ]
```

**Template clean:**

```liquid
Hello {{ user.name | upcase }}!
```

```go
result := eng.ParseTemplateAudit([]byte(source))
// result.Template != nil
// result.Diagnostics == []
```

---

## Integration with RenderAudit

The natural pipeline for full observability:

```go
parseResult := eng.ParseTemplateAudit(source)

// Parse-time diagnostics are always available immediately.
allDiags := parseResult.Diagnostics

if parseResult.Template != nil {
    auditResult, auditErr := parseResult.Template.RenderAudit(binds, auditOpts)
    // Merge render-time diagnostics.
    allDiags = append(allDiags, auditResult.Diagnostics...)
    if auditErr != nil {
        // also available in auditResult.Diagnostics already
    }
}
```

The two phases are independent. `ParseTemplateAudit` does not require `RenderAudit` and vice versa. Callers who only want parse diagnostics (e.g., a linter) never need to render.

---

## Implementation Plan

### Phase 1 — `ASTBroken` node

File: `parser/ast.go`, `render/node.go` (or wherever `ASTTag`/`ASTObject` renderers live)

- Add `ASTBroken` to `parser/ast.go`
- Register a renderer for `ASTBroken` that outputs nothing and returns `nil` error
- No visible change to any existing test — `ASTBroken` is only produced by the new path

### Phase 2 — Error-recovering parser

File: `parser/parser.go`

- Extract a `(diagnostics []parser.Diagnostic, fatalErr parser.Error)` result from `parseTokens` instead of just `(ASTNode, parser.Error)`.
- At each `WrapError(err, tok)` for expression parse failures: instead of `return nil, err`, emit a diagnostic with `code: syntax-error` and insert an `ASTBroken` node in place of the failed token. Continue the loop.
- The two structural errors (`RequiresParent` and `unterminated block`) remain `return nil, err` — they are still fatal.
- New signature: `func (c *Config) parseTokens(tokens []Token) (ASTNode, []ParseDiagnostic, Error)`
  - `ParseDiagnostic` is an internal type carrying `(code, tok, message)` — converted to public `Diagnostic` at the API boundary.

### Phase 3 — Static analysis pass

File: new `parser/static_analysis.go`

- `func staticAnalyze(root ASTNode, cfg *parser.Config) []ParseDiagnostic`
- Walk the AST and:
  - For each `ASTObject` and `ASTTag`: call `expressions.FilterNames(expr)` and check each name against `cfg.HasFilter(name)` → emit `undefined-filter` for each unknown name.
  - For each `ASTBlock`: check if all branches (`Body` + `Clauses`) are empty or contain only `ASTText` nodes with blank content → emit `empty-block`.
- Called by `ParseTemplateDiag` only when parse did not produce a fatal error.

### Phase 4 — `ParseResult` and public API

File: `liquid.go`, `engine.go`

- Add `ParseResult` struct to `liquid.go` (alongside the audit types).
- `ParseTemplateAudit` in `engine.go`:
  1. Call `newTemplateAudit(cfg, source, path, line)` — same as `newTemplate` but uses the recovering `parseTokens`.
  2. Convert internal `[]ParseDiagnostic` → `[]Diagnostic` using `SourceLoc`/`EndLoc` for `Range`.
  3. If no fatal error: run `staticAnalyze`, append its diagnostics.
  4. Return `&ParseResult{Template: tpl, Diagnostics: diags}`.
- `ParseStringAudit` is a one-liner wrapping `ParseTemplateAudit`.

### Phase 5 — Validate() removal (optional cleanup)

File: `template.go`, `liquid.go`

- `Template.Validate()` (planned in `render-audit-plan.md` Phase 2) becomes unnecessary: `ParseTemplateAudit` already performs the same static analysis at parse time when a template is first compiled.
- If `Validate()` was already implemented, it can delegate to `staticAnalyze` on the existing root node instead of duplicating the walk logic.

---

## Design Notes

**Why `ParseTemplateAudit` and not `ParseTemplateDiag`:** The name mirrors `RenderAudit` — both methods return a structured audit report rather than a raw error. `Audit` is the established suffix in this API; `Diag` would be an inconsistent abbreviation.

**Why `ParseResult` and not `(template, []Diagnostic, error)`:** Three-return-value functions in Go are awkward. `ParseResult` is a single value, JSON-serializable, and symmetric with `AuditResult`. The caller checks `result.Template != nil` instead of `err != nil` — more explicit and less easy to ignore.

**Why keep fatal errors fatal:** Error recovery in a block-structured language is inherently ambiguous. When `{% if %}` is unclosed, every token after it might belong to the missing `{% endif %}` — there is no safe place to continue. Emitting a diagnostic and returning `Template = nil` is the honest answer.

**Why `undefined-filter` is detected at parse time and not render time:** Filter names are resolved at expression evaluation time in the current engine, but the set of registered filters is fixed once the engine is frozen (after the first parse). A static walk can check all filter names without evaluating any bindings. This is cheaper and gives earlier feedback than waiting for a full render.

**Why `empty-block` is `info` and not `warning`:** An empty `{% if %}` block is never a runtime error. It might be a work-in-progress template or a deliberate no-op. `info` is the right severity for "this looks odd but it's not wrong."

**Why `ASTBroken` renders as empty string and not as the original source:** Rendering the original source would produce garbage HTML visible to end users. Empty string is the safe, invisible fallback — the same thing a `nil` variable outputs.

**Why `undefined-filter` severity is `error` and not `warning`:** An undefined filter causes a runtime `RenderError` when the template is actually rendered (unless `WithLaxFilters()` is active). At parse time, it is a definite bug — not a silent behavior. `error` severity reflects this. If `WithLaxFilters()` semantics need to be reflected at parse time, the engine could downgrade it to `warning` when it knows lax mode is the default; this is left as a future extension.

**Why no speculative diagnostics:** `ParseTemplateAudit` does not simulate renders to predict runtime errors like `nil-dereference` or `type-mismatch`. Those require actual binding values. Only structural and static facts about the template source are reported here.
