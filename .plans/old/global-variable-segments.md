# globalVariableSegments — Implementation Plan

> **Status: ✅ IMPLEMENTED** — all tests passing. See `analysis.go`, `render/analysis.go`, `expressions/analysis.go`, `tags/analyzers.go`.

## Goal

Implement static variable analysis equivalent to `globalVariableSegmentsSync()` in LiquidJS:
given a parsed template, return all global variable paths (coming from the outer scope) without executing the template.

```
{{ customer.first_name }} {% assign x = "hello" %} {{ order.total }}
→ [["customer", "first_name"], ["order", "total"]]
```

`x` does not appear: it is defined within the template itself via `assign`.

---

## Architecture: mirror LiquidJS

In LiquidJS, analysis and rendering are **completely separate** layers. Each tag implements optional methods (`arguments()`, `localScope()`, `blockScope()`) that provide expressions/names for analysis — without ever calling rendering code.

In Go Liquid, tag compilers already parse all necessary expressions:
- `ifTagCompiler` calls `e.Parse(node.Args)` → has the condition `Expression`
- `loopTagCompiler` calls `ParseStatement(LoopStatementSelector, ...)` → has `stmt.Expr` and `stmt.Variable`
- `makeAssignTag` calls `ParseStatement(AssignStatementSelector, ...)` → has `stmt.ValueFn` and `stmt.Variable`

This information exists — it's just locked inside opaque closures. The solution: have each tag register its analysis separately, preserving the already-parsed `Expression`s.

---

## Implementation architecture

### Approach: analysis registered alongside the compiler (non-breaking)

Add a parallel "analyzers" registry in the config, completely separate from the existing compilers. No existing signature changes. Built-in tags register their analyzer; custom tags (via `RegisterTag`) have no analysis by default — reasonable behavior identical to LiquidJS.

```
tags/standard_tags.go today:
  c.AddTag("assign", makeAssignTag(c))
  c.AddBlock("for").Compiler(loopTagCompiler)

with analysis:
  c.AddTag("assign", makeAssignTag(c))
  c.AddTagAnalyzer("assign", makeAssignAnalyzer())   ← new, separate

  c.AddBlock("for").Compiler(loopTagCompiler)
  c.AddBlockAnalyzer("for", loopBlockAnalyzer)       ← new, separate
```

In `render/compiler.go`, when compiling a node, populate `Analysis` if an analyzer exists:
```go
case *parser.ASTTag:
    f, err := td(n.Args)
    var analysis NodeAnalysis
    if analyzer, ok := c.findTagAnalyzer(n.Name); ok {
        analysis = analyzer(n.Args)
    }
    return &TagNode{n.Token, f, analysis}, nil
```

---

## Files and changes

### 1. `expressions/analysis.go` — new file

Internal `trackingContext` and `trackingValue` types. `trackingContext` implements the `Context` interface and intercepts `Get(name)` returning a `trackingValue` that records property access chains. Used by `computeVariables(evaluator)` to collect all paths referenced by an expression.

### 2. `expressions/expressions.go` — new method on the interface

Add `Variables() [][]string` to the `Expression` interface. Implemented lazily with `sync.Once` on the `expression` struct so it works for both expressions created via `Parse()` and those created internally by the yacc parser.

Note: `expressions/y.go` (generated) was updated to use named field syntax `&expression{evaluator: f}` to accommodate the new fields.

### 3. `render/analysis.go` — analysis types and AST walker

```go
type NodeAnalysis struct {
    Arguments  []expressions.Expression  // variable references "used" by this node
    LocalScope []string                   // variables DEFINED in current scope (assign, capture)
    BlockScope []string                   // variables added to scope for BODY only (for loop var)
}

type TagAnalyzer   func(args string) NodeAnalysis
type BlockAnalyzer func(node BlockNode) NodeAnalysis

type AnalysisResult struct {
    Globals [][]string  // variable paths from outer scope
    All     [][]string  // all variable paths (including locals)
}

func Analyze(root Node) AnalysisResult
```

### 4. `render/nodes.go` — add Analysis field

`Analysis NodeAnalysis` added to `TagNode` and `BlockNode`. `GetExpr()` added to `ObjectNode`.

### 5. `render/config.go` — analyzer registries

`tagAnalyzers`/`blockAnalyzers` maps added to `grammar`. Registration methods `AddTagAnalyzer` and `AddBlockAnalyzer` added to `Config`.

### 6. `render/compiler.go` — populate Analysis at compile time

`Analysis` is populated when compiling `ASTTag` and `ASTBlock` nodes, if an analyzer is registered.

### 7. `tags/analyzers.go` + `tags/standard_tags.go` — register built-in analyzers

Analyzers for: `assign`, `capture`, `for`, `tablerow`, `if`, `unless`, `case`.

### 8. `analysis.go` (root) — public API

```go
func (e *Engine) GlobalVariableSegments(t *Template) ([]VariableSegment, error)
func (e *Engine) VariableSegments(t *Template) ([]VariableSegment, error)
func (t *Template) GlobalVariableSegments() ([]VariableSegment, error)
func (t *Template) VariableSegments() ([]VariableSegment, error)
```

---

## Edge cases

| Case | Behavior |
|---|---|
| `{{ x }}` | `[["x"]]` |
| `{{ x.a.b }}` | `[["x", "a", "b"]]` |
| `{% assign y = x.val %}{{ y }}` | `[["x", "val"]]` — y is local |
| `{% for item in list %}{{ item.name }}{% endfor %}` | `[["list"]]` — item is local |
| `{% if cond %}{{ a }}{% else %}{{ b }}{% endif %}` | `[["cond"], ["a"], ["b"]]` |
| `{{ x \| upcase }}` | `[["x"]]` — filters don't change the path |
| `{% capture buf %}{{ x }}{% endcapture %}` | `[["x"]]` — buf is local |
| `{% assign x = 1 %}{{ x }}` | `[]` — x is local |
| `{% case status %}{% when "active" %}{{ a }}{% endcase %}` | `[["status"], ["a"]]` |

---

## Approach comparison: original vs this

| Aspect | Original (tracking ctx at analysis time) | This (analyzer registry) |
|---|---|---|
| Re-parse of source text | On every analysis call | Never (done once at compile) |
| Tag modifications | None | Register analyzers |
| Extensibility for custom tags | Automatic (works by accident) | Opt-in (no analysis by default) |
| Alignment with LiquidJS | Low | High |
| Walker complexity | Higher (knows each tag's syntax) | Lower (generic walker) |
| Failure point | Re-parse may diverge from compile | Single parse → consistent |

---

## Implementation order (completed)

1. ✅ `expressions/analysis.go`: `trackingContext`, `trackingValue` (internal)
2. ✅ `expressions/expressions.go`: `Variables() [][]string` on interface + `sync.Once` on `expression` struct; `expressions/y.go` updated to named fields
3. ✅ `render/analysis.go`: `NodeAnalysis`, `TagAnalyzer`, `BlockAnalyzer`, walker, `Analyze()`
4. ✅ `render/nodes.go`: `Analysis NodeAnalysis` on `TagNode` and `BlockNode`; `GetExpr()` on `ObjectNode`
5. ✅ `render/config.go`: `tagAnalyzers`/`blockAnalyzers` maps + registration methods
6. ✅ `render/compiler.go`: populate `Analysis` when compiling nodes
7. ✅ `tags/analyzers.go` + `tags/standard_tags.go`: built-in analyzers
8. ✅ `analysis.go` (root): public API (`Engine.GlobalVariableSegments`, `Engine.VariableSegments`, `Template.*`)
9. ✅ `analysis_test.go`: 17 edge case tests — all passing

---

## Out of scope

- Following `{% include %}` into sub-templates
- Tracking dynamic indices (`x[i]` where `i` is a variable)
- `assign` with Jekyll dot notation
- Implementing stubs (`Variables`, `GlobalVariables`, etc.)
