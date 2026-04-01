# Macro Checklist: liquid-go overview

High-level areas to verify. The [parity-checklist.md](parity-checklist.md) is the detailed breakdown of point **#3**.

---

## 1. Shopify Liquid spec conformance

- [ ] Pass the [Golden Liquid test suite](https://github.com/jg-rp/golden-liquid) (language-agnostic JSON/YAML)
- [ ] Compare behavior with [Shopify/liquid](https://github.com/Shopify/liquid) (Ruby) on edge cases
- [ ] Truthiness behavior (Liquid vs JavaScript)
- [ ] Whitespace control (`{%-`, `-%}`, `{{-`, `-}}`)
- [ ] Error types and messages compatible with the spec

## 2. Behavior fixes (from initial motivation)

- [ ] `uint64` and other Go numeric types in `{% if %}` comparisons
- [ ] Correct `nil`/`null`/`blank`/`empty` behavior in conditionals
- [ ] String↔number type coercion in arithmetic filters

## 3. API surface (parity with LiquidJS) → see [parity-checklist.md](parity-checklist.md)

- [ ] Complete Engine API (parsing, rendering, evalValue)
- [x] **Static analysis** (`globalVariableSegments`, `variableSegments`) ✅ implemented in `analysis.go` + `render/analysis.go`
  - [ ] Remaining: `variables`, `globalVariables`, `fullVariables`, `globalFullVariables`, `parseAndAnalyze`
- [ ] Missing tags (`render`, `layout`, `block`, `echo`, `increment`, `decrement`, `liquid`, `#`)
- [ ] Missing filters (~35 absent filters, see checklist)
- [ ] Configuration (globals, strictFilters, security limits)

## 4. Thread-safety and concurrency

- [ ] Engine shareable across goroutines without instantiating per request
- [ ] Config immutable after construction
- [ ] Render state isolated per call

## 5. Template system (inheritance and partials)

- [ ] `{% layout %}` + `{% block %}` — template inheritance
- [ ] `{% render %}` — partial with isolated scope (vs `include` which shares scope)
- [ ] Partial resolution relative to the current file

## 6. Drop protocol and custom types

- [ ] `Drop` with per-property fallback (`liquidMethodMissing`)
- [ ] `Comparable` interface for custom comparisons
- [ ] `ForloopDrop` / `TablerowloopDrop` as public types (not maps)
- [ ] `nil`/`empty`/`blank` as drops with correct semantics

## 7. Extensibility

- [ ] Plugin system (`engine.plugin(fn)`)
- [ ] Custom tags with static analysis support (opt-in NodeAnalysis)
- [ ] Custom filters with context access (`this` in JS → receiver in Go)

## 8. Security / DoS limits

- [ ] `parseLimit` — character limit at parse time
- [ ] `renderLimit` — render timeout
- [ ] `memoryLimit` — allocation limit

## 9. Internationalization and dates

- [ ] `date` filter with timezone and locale support
- [ ] `date_to_*` filters (xmlschema, rfc822, string, long_string)
- [ ] Configurable `timezoneOffset`

## 10. Tests and quality

- [ ] Test coverage aligned with Golden Liquid
- [ ] Documented benchmarks (already have ~25% faster, ~54% less memory)
- [ ] Fuzz testing on expression parsers
