# Implementation Checklist — Go Liquid vs Merged Reference

> Comparison between [go-liquid-reference.md](unchangeable-refs/go-liquid-reference.md) and [merged-liquid-reference.md](unchangeable-refs/merged-ruby-js-liquid-reference.md).
>
> **Status columns (in order: Impl · Tests · E2E):**
>
> | Column | Meaning |
> |--------|---------|
> | **Impl** | Implementation complete (✅ correct · ⚠️ behavior differs from spec · ❌ not implemented) |
> | **Tests** | Tests ported from references (Ruby and/or JS) passing |
> | **E2E** | Own intensive tests covering the feature (never run automatically — only when user explicitly requests) |
>
> **Values:** `✅` done · `⬜` pending · `➖` not applicable
>
> **Priority legend:**
> - **P1** — Core Shopify Liquid (present in Ruby _and_ JS; any valid Liquid needs this)
> - **P2** — Common extension (present in both but not Shopify core; e.g. Jekyll filters that both have)
> - **P3** — Ruby Liquid exclusive
> - **P4** — LiquidJS exclusive
> - **P5** — Nice-to-have / low priority
>
> **DECISION MADE** — items where Ruby, JS, or Go diverge and we have already decided which behavior will prevail here in the Go version.
>
> If you need to check where features are implemented in JS or Ruby, see [merged-liquid-reference.md](./unchangeable-refs/merged-ruby-js-liquid-reference.md).
> If you can't find it there, feel free to search directly in the original repositories cloned locally in .example-repositories

---

## 0. Bugs — Fixes to existing behavior

> These items do not require new structures. They can be investigated and fixed independently.

### B1 · Go numeric types in comparisons

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `uint64`, `uint32`, `int8`, etc. in `{% if %}` and operators | `NormalizeNumber()` added in `values/compare.go`: converts all Go integer/float types to `int64`/`uint64`/`float64` before any comparison. `numericCompare()` does precise comparison without falling back to float64 for the int64/uint64 pair, preserving precision for values > MaxInt64. Array indexing and loop bounds (`for i in list limit: n`) also failed for unsigned types — both fixed. E2E tests in `b1_numeric_types_test.go` cover: all operators (`==`,`!=`,`<`,`>`,`<=`,`>=`), `if`/`unless`/`case-when`, composite conditions `and`/`or`, struct fields with uint type, filters `abs`/`at_least`/`at_most`/`ceil`/`floor`/`round`, filter chains, `sort`/`where` on mixed arrays, array indexing with uint variable, `assign`+comparison, `for` with `limit`/`offset` uint, float precision. Two additional bugs fixed: `arrayValue.IndexValue` and `toLoopInt` in `iteration_tags.go` did not accept uint types. |

### B2 · Truthiness: `nil`, `false`, `blank`, `empty`

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Falsy rules in `{% if %}` | Implementation verified and correct. `wrapperValue.Test()` in `values/value.go` uses `v.value != nil && v.value != false`; `if/unless` in `control_flow_tags.go` uses `value != nil && value != false`; `and`/`or`/`not` in `expressions.y` use `.Test()`. `IsEmpty` and `IsBlank` in `values/predicates.go` are used only for comparisons with `empty`/`blank` keyword, not for general truthiness. `default` filter uses `IsEmpty` correctly (activates for `""`, `[]`, `{}`, `nil`, `false`; does NOT activate for `0` or non-empty strings). Ported tests: `TestPortedLiterals_Truthiness`, `TestPortedLiterals_Empty`, `TestPortedLiterals_Blank` in `expressions_ported_test.go` (46 tests). Intensive E2E in `b2_truthiness_test.go` (63 tests) covering: typed Go bindings, `if`/`unless`/`not`/`and`/`or`, `case/when` with nil/false, `default` filter with all edge cases including `allow_false`, `where` filter without value (truthy), comparisons with `blank` and `empty` via variables, `capture`/`assign`, and `elsif` chains. |

### B3 · Whitespace control in edge cases

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `{%-`/`-%}` and `{{-`/`-}}` in nested blocks and loops | **Bug fixed:** scanner in `parser/scanner.go` did not recognize `{%- # comment -%}` (space between `-` and `#`) — the inline comment regex `{%-?#` was updated to `{%-?\s*#`, allowing an optional space. This also enabled `{% # comment %}` (space without trim). Existing `TestInlineComment` tests expanded with 6 spacing variants. Behavior of `trimWriter` in loops and nested blocks confirmed correct: trim nodes in the `for` body execute per iteration; global `TrimTagLeft/Right` only affects the external context of the block, not the interior of iterations. Ported tests already covered the Ruby/LiquidJS cases. Intensive E2E in `b3_whitespace_ctrl_test.go` (38 tests) covers: `for` with all trim combinations, `for`+`else`, `if` nested in `for`, double nesting, `unless`/`case`/`when` with trim, `assign`/`capture` with trim, inline comment with space (bug fixed), `{{- -}}` inside loops, global `TrimTagLeft/Right/Both`, `greedy`/`non-greedy`, `liquid` tag with trim, and `raw` with internal trim markers. |

### B4 · Error messages and types

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Distinct error types (`ParseError`, `RenderError`, `UndefinedVariableError`) | Implemented via swarm PRE-E: `ParseError` in `parser/error.go`, `RenderError` and `UndefinedVariableError` in `render/error.go`. `UndefinedVariableError` carries the literal variable name. `ZeroDivisionError` also implemented in `filters/standard_filters.go`. **Intensive E2E tests in `b4_b6_error_test.go`** (55 tests) cover: `ParseError` (prefix, `errors.As`, `LineNumber`, `MarkupContext`, `Message`), `RenderError` (prefix, `errors.As`, `LineNumber`, `MarkupContext`, `Cause`), `UndefinedVariableError` (Name, LineNumber, Message, MarkupContext, StrictVariables), `ZeroDivisionError`, `ArgumentError` (filters + tags + line + correct context), `ContextError`, and the entire B6 suite of context preservation. |

### B5 · Renderer not safe for concurrent use

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `render.Context` shares mutable state between concurrent calls | **Investigation completed in `b5_concurrency_test.go`.** Result: the **render path is safe** for concurrent use — each call creates its own `nodeContext` with an isolated `bindings`; stateful tags (increment, assign, cycle, for-continue) operate only on the local map; compiled expressions are read-only; `sync.Once` in `Variables()` is thread-safe. **Bug confirmed**: `e.cfg.Cache map[string][]byte` in `render/config.go` is not concurrency-safe — `ParseTemplateAndCache` writes to the same map that `{% include %}` reads during rendering, causing `fatal error: concurrent map writes`. **Fix**: replace `Cache map[string][]byte` with `sync.Map` at 3 sites (`engine.go:242`, `render/context.go:200`, `render/context.go:234`). **Performance confirmed via benchmarks**: pure render of shared template scales nearly linearly (8.7k→3.2k→2.2 ns/op at 1→4→8 CPUs ✅). Parse under high concurrency does not scale (27k→21k→26k, plateaus) due to GC allocation pressure — there are +177 allocs/op per parse vs pure render. **Recommended patterns** (most to least efficient): (1) parse once, share `*Template`, render in N goroutines (~2k ns/op×N); (2) shared engine with cache enabled (`EnableCache()`) — same performance; (3) shared engine without cache, parse+render per call (~26k ns/op); (4) ❌ engine per goroutine — 6× slower (~50k ns/op) due to GC overhead from recreating filter/grammar maps. |

### B6 · Variable error messages degraded by indentation and block context

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Undefined variable errors with vague messages in `{% if %}` and other blocks | **Bug identified and fixed.** Root cause: `wrapRenderError` in `render/error.go` re-wrapped any `*RenderError` without `Path()` even when it already had `LineNumber > 0`. This caused `BlockNode` (if/for/unless/case) to overwrite the `MarkupContext` of the inner node (`{{ expr }}`) with the source of the parent block (`{% if ... %}`). **Fix:** added `re.LineNumber() > 0` to the preservation condition in `wrapRenderError` — if the error already has a line number, it came from a more specific node (ObjectNode/TagNode) and should be preserved. Single-line and multi-line templates now produce identical messages pointing to the exact node. Errors in block conditions (e.g. `{% if x | divided_by: 0 %}`) are still correctly attributed to `{% if %}`. Intensive tests in `b4_b6_error_test.go`. |

---

## 1. Tags

### 1.1 Output / Expression

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `{{ expression }}` | OK. Ported tests in `tags_ported_test.go` (`TestPorted_Output_*`). |
| ✅ | ✅ | ✅ | P1 | `echo` tag | `{% echo expr %}` — equivalent to `{{ }}`, but usable inside `{% liquid %}`. Ruby: always emits. JS: value optional (no value emits nothing). **DECISION MADE:** follow Ruby (value always required). Ported tests in `tags_ported_test.go` (`TestPorted_Echo_*`). |
| ✅ | ✅ | ✅ | P1 | `liquid` tag (multi-line) | Implemented in `tags/standard_tags.go`. Each non-empty, non-comment line is compiled as `{%...%}` and rendered in the current context (assign propagates). Lines starting with `#` are comments. Syntax errors propagate at compile-time. Tests in `TestLiquidTag`. |
| ✅ | ✅ | ✅ | P1 | `#` inline comment | Implemented in scanner (`parser/scanner.go`): pattern `{%-?#(?:...)%}` added to the tokenization regex. Trim markers (`{%-#` and `{%#-%}`) work. Tests in `TestInlineComment`. |

### 1.2 Variable / State

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `assign` | OK. Jekyll dot notation (`assign page.prop = v`) also implemented. Ported tests in `tags_ported_test.go` (`TestPorted_Assign_*`). |
| ✅ | ✅ | ✅ | P1 | `capture` | OK. **Bug fix:** `{% capture 'var' %}` and `{% capture "var" %}` (quoted variable name) now work correctly — quotes are stripped before assigning. Ported tests in `tags_ported_test.go` (`TestPorted_Capture_*`). |
| ✅ | ✅ | ✅ | P1 | `increment` | Implemented in `tags/standard_tags.go`. Counter separate from `assign` and `decrement`. Starts at 0, emits the current value and increments. Tests in `TestIncrementDecrement`. |
| ✅ | ✅ | ✅ | P1 | `decrement` | Implemented in `tags/standard_tags.go`. Counter separate from `assign` and `increment`. Starts at 0, decrements and emits the new value (first call = -1). Tests in `TestIncrementDecrement`. |

### 1.3 Conditionals

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `if` / `elsif` / `else` / `endif` | OK. Ported tests in `tags_ported_test.go` (`TestPorted_If_*`). |
| ✅ | ✅ | ✅ | P1 | `unless` / `else` / `endunless` | OK. Note: `unless` + `elsif` is not supported (Ruby also raises an error). Ported tests in `tags_ported_test.go` (`TestPorted_Unless_*`). |
| ✅ | ✅ | ✅ | P1 | `case` / `when` / `else` / `endcase` — `or` in `when` | `when val1 or val2` — supported. Implemented in the yacc grammar (`expressions.y`). Ported tests in `tags_ported_test.go` (`TestPorted_Case_*`). |
| ✅ | ✅ | ✅ | P3 | `ifchanged` | Implemented in `tags/standard_tags.go` via `ifchangedCompiler`. Captures the rendered content of the block and only emits if it changed since the last call. State in `"\x00ifchanged_last"`. Tests in `TestIfchangedTag`. |

### 1.4 Iteration

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `for` / `else` / `endfor` with `limit`, `offset`, `reversed`, range | OK. **Bug fix:** `for` with `nil` collection now correctly renders the `else` branch. Ported tests in `tags_ported_test.go` (`TestPorted_For_*`). |
| ✅ | ✅ | ✅ | P1 | `for` — modifier application order | **Fixed.** Ruby always applies `offset → limit → reversed` (regardless of user-declared order). Previously, Go applied them in a different fixed order. Now: `applyLoopModifiers` in `tags/iteration_tags.go` applies offset→limit first, then reversed. Verification tests in `tags_ported_test.go` (`TestPorted_For_Modifiers_*`). |
| ✅ | ✅ | ✅ | P4 | `for` — `offset: continue` | Implemented in `tags/iteration_tags.go`. Detected via regex before parsing. ALL for-loops track the final position in `"\x00for_continue_variable-collection"`. Loops with `offset:continue` resume from there. Tests in `TestOffsetContinue`. |
| ✅ | ✅ | ✅ | P1 | `break` / `continue` | OK. Ported tests in `tags_ported_test.go` (`TestPorted_For_Break_*`, `TestPorted_For_Continue_*`). |
| ✅ | ✅ | ✅ | P1 | `cycle` with named group | OK. Note: `cycle` outside `for` is not supported (requires `forloop` in context). Ported tests in `tags_ported_test.go` (`TestPorted_Cycle_*`). |
| ✅ | ✅ | ✅ | P1 | `tablerow` with `cols`, `limit`, `offset`, range | OK. Note: loop variables accessible as `forloop.xxx` (not `tablerowloop.xxx`). HTML emitted without newline between `<tr>` and `<td>`. Ported tests in `tags_ported_test.go` (`TestPorted_Tablerow_*`). |

### 1.5 Template inclusion

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `include` — basic syntax `{% include "file" %}` | Implemented and tested. |
| ✅ | ✅ | ✅ | P1 | `include` — `with var [as alias]` | Implemented in `tags/include_tag.go` with a dedicated parser. Tests in `TestIncludeTag_with_variable` and `TestIncludeTag_with_alias`. |
| ✅ | ✅ | ✅ | P1 | `include` — `key: val` args | Implemented in `tags/include_tag.go` with `parseKVPairs`. Tests in `TestIncludeTag_kv_pairs`. |
| ✅ | ✅ | ✅ | P3 | `include` — `for array as alias` | Implemented in `tags/include_tag.go`. `{% include 'file' for items as item %}` iterates the collection and renders the file once per item with `item` in the shared scope. Tests in `TestIncludeTag_for_array`. |
| ✅ | ✅ | ✅ | P1 | `render` tag | Implemented in `tags/render_tag.go`. Supports isolated scope, `with var [as alias]`, `key: val` args, and `for collection as item`. Tests in `TestRenderTag_*`. |

### 1.6 Structure / Text

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `raw` / `endraw` | OK. Ported tests in `tags_ported_test.go` (`TestPorted_Raw_*`). |
| ✅ | ✅ | ✅ | P1 | `comment` — nesting | Go: any token ignored inside comment (parser consumes until `endcomment`). Ruby: explicitly supports nested `comment` and `raw`. Effective behavior is identical for normal use — no code changes needed. Ported tests in `tags_ported_test.go` (`TestPorted_Comment_*`). |
| ✅ | ✅ | ✅ | P3 | `doc` / `enddoc` | Implemented. `c.AddBlock("doc")` in `standard_tags.go` + special handling in parser (`parser/parser.go`) same as `comment` — internal content is completely ignored at parse-time. Tests in `TestDocTag`. |
| ✅ | ✅ | ✅ | P4 | `layout` / `block` | Implemented in `tags/layout_tags.go`. `{% layout 'file' %}...{% endlayout %}` captures child blocks and renders the layout with overrides. `{% block name %}default{% endblock %}` in the child defines override; in the layout defines a slot with fallback. Requires `render/context.go` updated to support `RenderFile` in block context. Tests in `TestLayoutTag*` and `TestBlockTag_standalone`. |

---

## 2. Filters

### 2.1 String

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `downcase`, `upcase` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `capitalize` | Fix applied: first char uppercase + rest lowercase. Ported tests (`"MY GREAT TITLE"` → `"My great title"`). |
| ✅ | ✅ | ✅ | P1 | `append`, `prepend` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `remove`, `remove_first`, `remove_last` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `replace`, `replace_first`, `replace_last` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `split` | Trailing empty strings removed (correct). Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P4 | `lstrip`, `rstrip`, `strip` — optional `chars` argument | Implemented: each filter accepts optional `chars func(string) string`. Ported tests in `filters/standard_filters_test.go`. |
| ✅ | ✅ | ✅ | P1 | `strip_html` | Fix applied: removes `<script>/<style>` with content (case-insensitive), HTML comments `<!-- -->`, then generic tags. Ported tests. |
| ✅ | ✅ | ✅ | P1 | `strip_newlines` | Fix applied: now removes `\r\n`, `\r` and `\n` (Windows line ending support). Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `newline_to_br` | Fix applied: converts `\n` → `<br />\n` (preserves the newline). Ported tests. |
| ✅ | ✅ | ✅ | P1 | `truncate`, `truncatewords` | Fixes applied: (1) `truncate`: n ≤ len(ellipsis) → returns only ellipsis; string that fits exactly is not truncated. (2) `truncatewords`: n=0 → n=1; whitespace normalized (tabs/newlines → space). (3) `first`/`last`: now work on strings (return first/last rune). Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `size`, `slice` | Fix applied: `slice` with negative length no longer panics (clamp to 0). Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P3 | `squish` | Implemented in `filters/standard_filters.go`: `strings.TrimSpace` + internal whitespace collapse. Tests in `filters/standard_filters_test.go`. |
| ✅ | ✅ | ✅ | P3 | `h` (alias for `escape`) | Implemented. `AddFilter("h", html.EscapeString)` in `filters/standard_filters.go`. Ported tests. |
| ✅ | ✅ | ✅ | P4 | `normalize_whitespace` | Present in Go (Jekyll ext). Ported tests (`squish`). |
| ✅ | ✅ | ✅ | P4 | `number_of_words` | Present in Go (Jekyll ext). Ported tests via `size`. |
| ✅ | ✅ | ✅ | P4 | `array_to_sentence_string` | Present in Go (Jekyll ext). Ported tests via `join`. |
| ✅ | ✅ | ✅ | P4 | `xml_escape` | Ported tests in `filters_ported_test.go`. |

### 2.2 HTML

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `escape`, `escape_once` | Ported tests in `filters_ported_test.go`. |

### 2.3 URL / Encoding

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `url_encode`, `url_decode` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P4 | `cgi_escape`, `uri_escape`, `slugify` | Present (Jekyll exts). Ported tests via `url_encode`. |
| ✅ | ✅ | ✅ | P3 | `base64_url_safe_encode`, `base64_url_safe_decode` | Implemented with `encoding/base64.URLEncoding`. Ported tests. |
| ✅ | ✅ | ✅ | P1 | `base64_encode`, `base64_decode` | Ported tests in `filters_ported_test.go`. |

### 2.4 Math

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `abs`, `plus`, `minus`, `times`, `ceil`, `floor`, `round` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `at_least`, `at_most` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `divided_by` — division by zero | Fix applied: `divided_by` now preserves the type of the dividend — `float / int` returns float (e.g. `2.0 / 4 = 0.5`); integer division only when both operands are integers. Division by zero returns an error. Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P1 | `modulo` — division by zero | Fix applied: guard for zero returns `ZeroDivisionError`. Additional fix: modulo now uses floor modulo (result has the same sign as the divisor, same as Ruby) — `func(rawA, b any)` with logic identical to `divided_by`, preserving int/int→`int64` type. `-10 | modulo: 3` = 2 (not -1). Negative tests added in `filters_ported_test.go` and `s2_filters_e2e_test.go`. |

### 2.5 Date

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `date` with strftime | Fix applied: `nil | date: fmt` now returns `nil` (same as Ruby). Ported tests in `filters_ported_test.go` including nil, int timestamp, string, `time.Time`. |
| ✅ | ✅ | ✅ | P4 | `date` — `'now'` / `'today'` as input | Implemented in `values/parsedate.go`: `today` treated same as `now`. Tests in `values/parsedate_test.go`. |
| ✅ | ✅ | ✅ | P4 | `date_to_xmlschema`, `date_to_rfc822`, `date_to_string`, `date_to_long_string` | Implemented in `filters/standard_filters.go`. `date_to_xmlschema`: format `%Y-%m-%dT%H:%M:%S%:z`; `date_to_rfc822`: format `%a, %d %b %Y %H:%M:%S %z`; `date_to_string`/`date_to_long_string`: default mode `DD Mon YYYY`, ordinal mode with UK/US styles. Helper `formatJekyllDate()` and `ordinalSuffix()`. Added `"2006-01-02T15:04:05"` (ISO 8601 without timezone) in `values/parsedate.go`. Ported tests from `liquidjs/test/integration/filters/date.spec.ts` in `filters/standard_filters_test.go`. |

### 2.6 Array

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `join`, `first`, `last`, `reverse`, `sort`, `sort_natural`, `map`, `sum`, `compact`, `uniq`, `concat` | Fixes applied: (1) `first`/`last` now work on strings. (2) `sort` and `sort_natural` use nil-last (same as Ruby). (3) `sort_natural` no longer panics on nil elements. Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P3 | `compact` — `property` argument | Implemented: `compact` accepts optional `property func(string) string`. Filters items where `item[prop]` is nil. Ported tests. |
| ✅ | ✅ | ✅ | P3 | `uniq` — `property` argument | Implemented: `uniq` accepts optional `property func(string) string`. Deduplicates by `item[prop]`. Ported tests. |
| ✅ | ✅ | ✅ | P1 | `sort` — nil-safe | Fix applied: `SortByProperty` called with `nilFirst: false` — nils go to the end as in Ruby. Ported tests. |
| ✅ | ✅ | ✅ | P1 | `where`, `reject`, `find`, `find_index`, `has` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P4 | `group_by` | Ported tests via `map`/`where`. |
| ✅ | ✅ | ✅ | P4 | `push`, `pop`, `unshift`, `shift`, `sample` | Ported tests (purity — no mutation) in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P4 | `where_exp`, `reject_exp`, `group_by_exp`, `has_exp`, `find_exp`, `find_index_exp` | Implemented via `AddContextFilter` infrastructure (PRE-B). Registered in `filters/standard_filters.go`. Ported tests via `where`/`find`. |

### 2.7 Misc

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `default` — keyword arg `allow_false: true` | Implemented: `default` accepts `kwargs ...any` and inspects `NamedArg{Name: "allow_false"}`. Ported tests. |
| ✅ | ✅ | ✅ | P4 | `json`, `inspect`, `to_integer` | Ported tests in `filters_ported_test.go`. |
| ✅ | ✅ | ✅ | P4 | `jsonify` (alias for `json`) | Implemented. `AddFilter("jsonify", ...)` in `filters/standard_filters.go`. Ported tests. |
| ✅ | ✅ | ✅ | P4 | `raw` filter | Implemented in `expressions/filters.go` (registered together with `safe` in `AddSafeFilter`). `NewConfig()` now always calls `AddSafeFilter` — `raw` and `safe` are always available, with or without autoescape. Also registered in `filters/standard_filters.go` for standard filter contexts. When autoescape is disabled, `raw` wraps in `SafeValue` which is immediately transparent in the render. Ported tests from LiquidJS `output-escape.spec.ts` in `render/autoescape_test.go`. |

---

## 3. Filter System

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Positional filters | OK. |
| ✅ | ✅ | ✅ | P1 | **Keyword args in filters** (`filter: arg, key: val`) | Infrastructure implemented (PRE-A). `NamedArg` struct in `expressions/filters.go`, `makeNamedArgFn` in `builders.go`, updated grammar. `default` filter updated to accept `allow_false: true`. Ported tests in `filters/standard_filters_test.go`. |
| ✅ | ✅ | ✅ | P3 | `global_filter` — proc applied to all output | Implemented via `Engine.SetGlobalFilter(fn func(any) (any, error))`. The function is applied to the evaluated value of each `{{ }}` before writing. Analogous to Ruby's `global_filter` option. Tests in `engine_test.go` (TestEngine_SetGlobalFilter). |

---

## 4. Expressions / Literals

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `nil`, `true`, `false`, int, float, string, range | OK. Range now has `String()` that returns `"start..end"` (Ruby compat). Ported tests in `expressions_ported_test.go`. E2E in `s4_expressions_e2e_test.go` (section A). |
| ✅ | ✅ | ✅ | P1 | **`empty` as special literal** | Implemented. Scanner recognizes `empty` as keyword (`EMPTY` token). `values.EmptyDrop` singleton with symmetric comparison in `values/compare.go`. Ported tests in `TestPortedLiterals_Empty` (17 cases: render, symmetric comparisons with string/array/map/nil/false, ordering operators, `empty != empty`). E2E in `s4_expressions_e2e_test.go` (section C). |
| ✅ | ✅ | ✅ | P1 | **`blank` as special literal** | Implemented. Scanner recognizes `blank` as keyword. `values.BlankDrop` singleton; `IsBlank` covers nil, false, whitespace-only string, empty arrays/maps. Ported tests in `TestPortedLiterals_Blank` (14 cases: render, nil/false/string/map/array blank and non-blank, number/true are not blank). E2E in `s4_expressions_e2e_test.go` (section D). |
| ✅ | ✅ | ✅ | P1 | `<>` as alias for `!=` | Implemented in `expressions/scanner.rl`. Ported tests in `TestPortedLiterals_DiamondOperator` (6 cases: int/string/float, true and false). E2E in `s4_expressions_e2e_test.go` (section B). |
| ✅ | ✅ | ✅ | P4 | `not` unary operator | Fix: grammar updated to `cond AND cond` / `cond OR cond` (was `cond AND rel`). `not x or not y` now parses correctly. AND/OR are `%right` same precedence (right-to-left). Ported tests in `expressions_ported_test.go`. E2E in `s4_expressions_e2e_test.go` (section F). |
| ✅ | ✅ | ✅ | P1 | Strings — internal escapes (`\n`, `\"`, etc.) | Implemented via `unescapeString()` in `expressions/scanner.rl`. Supports `\n`, `\t`, `\r`, `\"`, `\'`. Ported tests in `expressions_ported_test.go`. E2E in `s4_expressions_e2e_test.go` (section H). |
| ✅ | ✅ | ✅ | P1 | `range contains n` — `contains` operator on ranges | **Bug fixed.** `(1..5) contains 3` returned `false` because `Range` struct was wrapped as `structValue`, which checks field names instead of integer membership. **Fix:** new `rangeValue` type in `values/range.go` with its own `Contains` that checks `n >= b && n <= e`. Type recognized via `case Range:` in `ValueOf`. Ported tests in `TestPortedLiterals_RangeContains` (7 cases: LiquidJS contains 3→yes, contains 6→no, lower/upper bounds, below lower, variable bound, basic for loop). E2E in `s4_expressions_e2e_test.go` (section E). |
| ✅ | ✅ | ✅ | P1 | `nil`/`null` in ordering operators (`<=`, `<`, `>`, `>=`) | Already correct (returns false for any ordering comparison involving nil). Ported tests in `TestPortedLiterals_NilOrdering` (10 cases: Ruby `test_zero_lq_or_equal_one_involving_nil` — `null <= 0`, `0 <= null`, and variations `<`, `>`, `>=`, `nil`). E2E in `s4_expressions_e2e_test.go` (section G). |

---

## 5. Variable Access

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `obj.prop`, `obj[key]`, `array[0]` | OK. Ported tests in `variables_ported_test.go`. **Intensive E2E in `s5_variable_access_e2e_test.go`** (~80 tests). |
| ✅ | ✅ | ✅ | P1 | `array[-1]` — negative index | Supported. `IndexValue(-1)` returns the last element. Ported tests from LiquidJS #486. **Intensive E2E in `s5_variable_access_e2e_test.go`** (~80 tests). |
| ✅ | ✅ | ✅ | P1 | `array.first`, `array.last`, `obj.size` | OK. Ported tests in `variables_ported_test.go` with edge cases: empty array, `size` key override by real key in map, first/last equivalence with indices. **Intensive E2E in `s5_variable_access_e2e_test.go`** (~80 tests). |
| ✅ | ✅ | ✅ | P1 | `{{ test . test }}` — dot with spaces | **Fixed.** Added rule `expr '.' IDENTIFIER` in grammar (`expressions.y`). Tests in `TestVariables_DotWithSpaces`. **Intensive E2E in `s5_variable_access_e2e_test.go`** (~80 tests). |
| ✅ | ✅ | ✅ | P3 | `{{ [key] }}` — dynamic variable (indirection) | **Implemented.** Added rule `'[' expr ']'` in grammar + `makeVariableIndirectionExpr()` in `builders.go`. Evaluates the inner expression to string and uses it as the variable name in context. Supports `{{ [key] }}`, `{{ [list[0]] }}`, and `{{ list[list[0]]["foo"] }}`. Tests in `TestVariables_DynamicFindVar*`. **Intensive E2E in `s5_variable_access_e2e_test.go`** (~80 tests). |
| ✅ | ✅ | ✅ | P4 | `{{ ["Key with Spaces"].subprop }}` — bracket root + dot (LiquidJS #643) | **Fixed.** Follows from the same `'[' expr ']'` rule that produces a value on which `PROPERTY` and `'.' IDENTIFIER` already work. Tests in `TestVariables_BracketRootPlusDot`. **Intensive E2E in `s5_variable_access_e2e_test.go`** (~80 tests). |

---

## 6. Drops (Special Objects)

### 6.1 ForloopDrop

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `index`, `index0`, `rindex`, `rindex0`, `first`, `last`, `length` | Ported tests in `tags_ported_test.go`: `TestPorted_For_LoopVariables` (Ruby `test_for_helpers`) — covers all standard properties. Intensive E2E in `drops_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | **`forloop.name`** | Already implemented in `tags/iteration_tags.go` via `loopName(args, variable)`. Returns `"variable-collection"`. Tests in `TestForloopMeta`. E2E: simple array, different variable, range, outer-vs-inner, consistent across iterations. |
| ✅ | ✅ | ✅ | P3 | `forloop.parentloop` | Already implemented in `tags/iteration_tags.go` — saves the parent `forloopMap` before starting the child loop. Tests in `TestForloopMeta`. E2E: nil at top-level, index/index0/rindex/first/last/length/name of parent, 3-level nesting, used in condition, used as label. |

### 6.2 TablerowloopDrop

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `row`, `col`, `col0`, `col_first`, `col_last` | Already implemented in `tags/iteration_tags.go` via `tableRowDecorator`. Fields exposed via `forloop` (not `tablerowloop`). Tests in `TestTablerowLoopVars`. E2E in `drops_e2e_test.go`: without cols, cols:2, odd items, single item, cols > items, limit+offset, range, reversed, col_last for logical break, standard props (index/length/rindex). |

### 6.3 EmptyDrop / BlankDrop

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | **`empty` drop/literal** | `values.EmptyDrop` exported in `values/emptydrop.go`. Unit tests in `values/emptydrop_test.go`; template-level tests in `expressions_ported_test.go` (`TestPortedLiterals_Empty`), ported from `liquidjs/test/integration/drop/empty-drop.spec.ts`. Intensive E2E in `drops_e2e_test.go`: typed Go bindings (string/slice/map/nil/false/zero/whitespace), symmetric, not-equal-to-self, ordering always false, unless, assign, capture, case/when. **Bug fixed:** `case/when empty` and `case/when blank` did not work because `Evaluate()` discarded the sentinel identity via `.Interface()`; fixed by preserving the sentinel via the sealed `LiquidSentinel` interface. |
| ✅ | ✅ | ✅ | P1 | **`blank` drop/literal** | `values.BlankDrop` exported in `values/emptydrop.go`. Unit tests in `values/emptydrop_test.go`; template-level tests in `expressions_ported_test.go` (`TestPortedLiterals_Blank`), ported from `liquidjs/test/integration/drop/blank-drop.spec.ts` and Ruby `condition_unit_test.rb`. E2E in `drops_e2e_test.go`: nil/false/empty-string/whitespace/tab/newline/empty-slice/empty-map are blank; zero/true/non-empty are not; symmetric; cross-comparison empty-vs-blank. |

### 6.4 Drop base class

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `Drop` interface (`ToLiquid() any`) | Ported tests in `drops_test.go`: `TestDrop_nestedDropPropertyAccess` and `TestDrop_nestedDropArrayIteration` (Ruby `test_text_drop`, `test_text_array_drop`); `TestDrop_methodCallableAsProperty`, `TestDrop_methodUsableInCondition`, `TestDrop_unknownFieldReturnsEmpty` (JS `drop.spec.ts`); `TestDrop_contextDropReadsForloopIndex` (Ruby `test_access_context_from_drop`); `TestDropMethodMissing_variousReturnTypes` (JS `DynamicTypeDrop`). E2E in `drops_e2e_test.go`: string/map/slice/nested drops, ToLiquid in condition/filter/assign/capture, map+slice combo, ForloopDrop access via ContextDrop. |
| ✅ | ✅ | ✅ | P3 | Drop base class with `liquid_method_missing` | `DropMethodMissing` interface in `drops.go` + `values/drop.go`; integrated in `values/structvalue.go`. Ported Ruby/JS tests in `drops_test.go`. E2E: known-field priority, dispatch, nil→empty, bool/string/int/array/map/nested return types, filter chain, nested drops, for loop, multiple accesses. |
| ✅ | ✅ | ✅ | P3 | `context=` injection in drop | `ContextDrop` interface (alias `values.ContextSetter`) + `DropRenderContext` (alias `values.ContextAccess`) in `drops.go`. `expressions/context.go: Get()` injects context before any property access. Tests in `drops_test.go` (TestContextDrop_*, ExampleContextDrop). E2E in `drops_e2e_test.go`: reads binding, reads int, sees assign, missing key, inside for loop, nested for, multiple drops, same drop twice, value changes between accesses, used in condition/filter, combined with MissingMethod, injection before first access. |

---

## 7. Context / Scope

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Scope stack, get/set variables | Ported tests from Ruby `context_test.rb` and LiquidJS `context.spec.ts` in `context_scope_ported_test.go` (48 subtests in `TestScopeStack_GetSet`): basic types, dot/bracket notation, first/last/size on arrays, size on maps, explicit size key in hash, variables with hyphens, hash-to-array, access with dynamic variable key (`products[var].first`, `products[nested.var].last`), assign persists in correct scope, loop variables restored after loop, single/double quoted strings, bracket notation chain. **Intensive E2E in `b7_context_scope_test.go`** (87 tests): all primitive Go types, dot nested 4+ levels, bracket with string/variable/variable-path key, negative index, first/last in middle of chain, size on string/array/map/explicit-key, variable with hyphen, array["string"] → nil, assign top-level/if/for/capture/persists/does-not-leak-between-renders, loop var restored, structs with Drop/liquid tags, filters on scope variables. |
| ✅ | ✅ | ✅ | P1 | **Isolated sub-context** | Implemented. `nodeContext.SpawnIsolated(bindings)` in `render/node_context.go` — creates new context without inheriting parent variables; globals propagate. Ported tests from Ruby `test_new_isolated_subcontext_*` in `context_scope_ported_test.go`: `TestIsolatedSubcontext_DoesNotInheritParentBindings`, `TestIsolatedSubcontext_GlobalsPropagateToIsolatedContext`, `TestIsolatedSubcontext_ExplicitBindingsVisible`, `TestIsolatedSubcontext_ExplicitBindingWinsOverGlobal`. **Intensive E2E in `b7_context_scope_test.go`**: parent does not leak (1 or N vars), explicit bindings visible, explicit > global, globals propagate (multiple), globals visible + parent not, assign in isolated does not leak to parent, partial with for loop, sequence of 3 independent isolated calls. |
| ✅ | ✅ | ✅ | P1 | Registers (internal tag state) | OK (map accessible via context). Tests in `context_scope_ported_test.go`: `TestRegisters_StatePersistedWithinRender`, `TestRegisters_StateResetBetweenRenders`, `TestRegisters_CycleTagState`, `TestRegisters_CycleTagNamedGroups`. **Intensive E2E in `b7_context_scope_test.go`**: Set/Get persist within render, accumulation over 5 calls, state visible inside loop, reset between 5 sequential renders, reset between different templates, Set visible via `{{ var }}`, Set overwrites binding, cycle state by group, two independent groups, increment isolated from assign, decrement isolated from increment+assign, 50 concurrent goroutines with isolated state. |
| ✅ | ✅ | ✅ | P2 | **Global variables separate from scope** (`globals`) | Implemented. `Config.Globals` copied before bindings in `newNodeContext` and `SpawnIsolated`. `Engine.SetGlobals`/`GetGlobals` exposed in `engine.go`. Ported tests from Ruby `test_static_environments_are_read_with_lower_priority_than_environments` and LiquidJS `liquid.spec.ts` in `context_scope_ported_test.go`: `TestGlobals_AccessibleInTemplate`, `TestGlobals_ScopeBindingWinsOverGlobal`, `TestGlobals_MultipleGlobals`, `TestGlobals_AssignDoesNotPersistAcrossRenders`, `TestGlobals_GetGlobals`, `TestGlobals_EmptyBindingsWithGlobals`, `TestGlobals_NilBindingsFallbackToGlobals`, `TestGlobals_AccessibleViaCustomTag`, `TestGlobals_GlobalsInStrictVariablesMode`, `TestContext_BindingsMethod`, `TestContext_SetPersistsWithinRender`, `TestContext_WriteValue`. **Intensive E2E in `b7_context_scope_test.go`**: basic/multiple/nested/nil access, with nil bindings, with empty bindings, binding shadowing/partial shadow, assign does not mutate global in future renders, assign does not mutate in 100 parallel renders, GetGlobals before/after Set, WithGlobals merge, WithGlobals doesn't affect next renders, binding wins WithGlobals, StrictVariables: global is defined / undefined still errors / binding defined, ctx.Get of global, ctx.Get with shadow, global in isolated sub-context, globals in Bindings(), global usable in filter argument, WriteValue nil/array. |

---

## 8. Configuration / Engine

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `StrictVariables()` | OK (engine-level). Tests in `engine_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | `LaxFilters()` | OK. Tests in `engine_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | Custom delimiters (`Delims()`) | OK. Tests in `engine_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | Custom `TemplateStore` | OK. Tests in `engine_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | `RegisterTag`, `RegisterBlock`, `RegisterFilter` | OK. Tests in `engine_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P2 | `strict_variables` / `strict_filters` — **per render, not per engine** | `WithStrictVariables()`, `WithLaxFilters()` in `liquid.go`. Accepted by all render methods. Ported tests from LiquidJS `strict.spec.ts` and Ruby `template_test.rb` in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P2 | **per-render** `globals` (`WithGlobals`) | `WithGlobals(map[string]any)` in `liquid.go`. Ported from LiquidJS `liquid.spec.ts`. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P3 | `error_mode` (`:lax` for tags) | `Engine.LaxTags()` — unknown tags compile as no-ops. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P3 | `template.errors` / error collection | Via `WithErrorHandler`: accumulating errors while-rendering is the Go standard. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P3 | `exception_renderer` / `exception_handler` | `WithErrorHandler(func(error) string)` + `Engine.SetExceptionHandler()`. Ported from Ruby `template_test.rb`. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P3 | Resource limits (`render_length_limit`) | `WithSizeLimit(int64)` — aborts when output exceeds N bytes. Ported from Ruby `test_resource_limits_render_length`. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P4 | Resource limits (time-based: `renderLimit`) | `WithContext(context.Context)` — render stops when context cancels/expires. Ported from LiquidJS `dos` concept. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P4 | Template cache | `Engine.EnableCache()` + `ClearCache()` — sync.Map keyed by source string. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P4 | `globals` option on engine | `Engine.SetGlobals` / `GetGlobals()`. Tests in `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P3 | **per-render** `global_filter` (`WithGlobalFilter`) | `WithGlobalFilter(fn func(any)(any,error))` in `liquid.go`. Mirrors Ruby `global_filter:` render option (`template.rb · apply_options_to_context`). Tests in `engine_section8_test.go` (8 tests). E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P5 | `NewBasicEngine` — ported tests | Engine without default filters/tags. Tests in `engine_section8_test.go` (4 tests). E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P3 | `EnableJekyllExtensions` + ported tests | Dot notation in assign. Tests in `engine_section8_test.go` (3 tests). E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | `RegisterTag`, `RegisterBlock`, `UnregisterTag` — additional tests | Tests: custom tag, custom block with `InnerString`, unregister makes tag unknown, unregister idempotent. In `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | `RegisterTemplateStore` — ported tests | Tests: include uses store. In `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ✅ | P1 | `Delims` — additional tests | Custom delimiters, standard delimiters no longer work, empty string restores default. In `engine_section8_test.go`. E2E in `s8_engine_config_e2e_test.go`. |
| ✅ | ✅ | ➖ | P1 | `SetAutoEscapeReplacer`, `RegisterTagAnalyzer`, `RegisterBlockAnalyzer` — frozen guard tests | Added to `TestEngine_FrozenAfterParse` in `b5_concurrency_test.go`. Total: 42 subtests (21 entries × 2). `SetAutoEscapeReplacer` E2E in `s8_engine_config_e2e_test.go`. |

---

## 9. Static Analysis

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P2 | `GlobalVariableSegments`, `VariableSegments`, `GlobalFullVariables`, `FullVariables` | Ported tests from Ruby (`parse_tree_visitor_test.rb`) and LiquidJS (`variables.spec.ts`, `parse-and-analyze.spec.ts`) in `analysis_ported_test.go`. New tests added in `TestRubyLiquid_ParseTreeVisitorExtra` (dynamic variable, echo, for/tablerow limit+offset, include/render with+for) and `TestLiquidJS_VariableAnalysisExtra` (filter keyword args, increment/decrement locals, echo with filter kwargs, for with variable limit, liquid tag inner vars, tablerow, unless+else, include/render kv args). |
| ✅ | ✅ | ✅ | P2 | `Analyze()` / `ParseAndAnalyze()` | Ported tests from LiquidJS in `analysis_ported_test.go`. |
| ✅ | ✅ | ✅ | P2 | `RegisterTagAnalyzer`, `RegisterBlockAnalyzer` | Basic test in `analysis_test.go`. |
| ✅ | ✅ | ✅ | P3 | `ParseTreeVisitor` visitor-style API | Implemented via `Walk(WalkFunc)` and `ParseTree() *TemplateNode` in `visitor.go`. Public types: `TemplateNodeKind` (Text/Output/Tag/Block), `TemplateNode` (Kind, TagName, Location, Children), `WalkFunc`. Tests in `visitor_test.go` ported from `parse_tree_visitor_test.rb` (tree structure, skip children, all node kinds, tag names, source locations). |
| ✅ | ✅ | ✅ | P2 | Analyzers for `echo`, `increment`, `decrement`, `include`, `render`, `liquid` | Implemented in `tags/analyzers.go`: `makeEchoAnalyzer` (Arguments = expression), `makeIncrementAnalyzer`/`makeDecrementAnalyzer` (LocalScope = counter name, per LiquidJS spec), `makeIncludeAnalyzer` (Arguments = file+with+for+kv exprs), `makeRenderAnalyzer` (Arguments = for-collection+with+kv), `makeLiquidAnalyzer(cfg)` (ChildNodes = compiled inner template for recursive analysis). `NodeAnalysis.ChildNodes []Node` added to `render/analysis.go`; `walkForVariables`, `collectLocals`, `walkForTags` updated to recurse into ChildNodes. |
| ✅ | ✅ | ✅ | P2 | `loopBlockAnalyzer` with limit/offset | `loopBlockAnalyzerFull` replaces `loopBlockAnalyzer` in `tags/analyzers.go`. Includes `stmt.Loop.Limit` and `stmt.Loop.Offset` in Arguments when present, in addition to the collection expr. Makes `{% for x in list limit: n offset: m %}` report `n` and `m` as globals if they are variables. |

> **DECISION MADE** — `cycle` with identifiers as values (e.g. `{% cycle test %}`) is not supported because the cycle grammar only accepts string literals (`LITERAL`), not identifiers. This is a behavior divergence from Ruby and LiquidJS, but changing the cycle grammar to accept expressions would require significant refactoring and runtime semantics change. Documented as a known limitation.
>
> **DECISION MADE** — `unless` does not support `elsif` (unlike LiquidJS). Tests adapted to use only `unless + else`.
>
> **DECISION MADE** — variable analysis is flow-insensitive: if a variable is assigned anywhere in the template (via assign/capture), it is treated as local throughout the template. LiquidJS does flow-sensitive analysis (detects use-before-assign). Tests adapted to reflect Go behavior.

---

## 10. Error Handling

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `SourceError` with `Path()`, `LineNumber()`, `Cause()` | OK. `Message()` and `MarkupContext()` added to the `parser.Error` and `render.Error` interfaces. Ported tests in `error_handling_ported_test.go` (27 tests). Intensive E2E in `s10_error_handling_e2e_test.go`: section A (ParseError/SyntaxError — prefix, SyntaxError=ParseError alias, line numbers on single/multi/nested/whitespace-trim, Path/Message/MarkupContext, unknown tag, unclosed block, invalid operator), section B (RenderError — prefix, types, line numbers, Message, MarkupContext), section I (prefix invariants — regression guard for both prefixes, `(line N)` in strings). |
| ✅ | ✅ | ✅ | P3 | `ZeroDivisionError` specific type | Implemented in `filters/standard_filters.go`. Exported type returned by `divided_by` and `modulo`. Tests in `filters/standard_filters_test.go`, `engine_test.go`, `error_handling_ported_test.go`. E2E in `s10_error_handling_e2e_test.go`: section C (C1–C5 — `divided_by: 0` and `modulo: 0` via `errors.As`, ZeroDivisionError below RenderError in chain, non-zero divisor ok, message content). |
| ✅ | ✅ | ✅ | P3 | Specific error types (`SyntaxError`, `ArgumentError`, `ContextError`, etc.) | `SyntaxError` = type alias for `ParseError` (in `parser/error.go`). `ArgumentError` and `ContextError` added in `render/error.go` as simple types detectable via `errors.As`. `ParseError.Error()` uses prefix `"Liquid syntax error"`, `RenderError.Error()` uses `"Liquid error"`. Ported tests in `error_handling_ported_test.go`. E2E in `s10_error_handling_e2e_test.go`: section A2 (SyntaxError/ParseError alias), section D (D1–D6 — ArgumentError from filter/tag, ContextError from tag, message in chain, correct prefix), section H (H1–H4 — complete chain walk: ZeroDivisionError, ArgumentError, RenderError, ParseError all findable via `errors.As` from top-level error). |
| ✅ | ✅ | ✅ | P1 | Error metadata — `markup_context` | `MarkupContext()` added to the `parser.Error` and `render.Error` interfaces. Returns the source text of the token that caused the error (e.g. `{% tag args %}`). When there is no pathname, the markup context appears in `Error()` as a locative. `Message()` returns only the message without prefix/location. Ported tests in `error_handling_ported_test.go`. E2E in `s10_error_handling_e2e_test.go`: section E (E1–E8 — UndefinedVariableError: default mode no error, strict mode, Name preserved, line/MarkupContext, WithStrictVariables per-render, chain walk, correct prefix), section F (F1–F6 — WithErrorHandler: node replacement, continuation, multiple errors, handler receives typed error, parse errors not captured, healthy nodes preserved), section G (G1–G4 — markup context end-to-end: no path, own context per node, inner context preserved through nested blocks, empty when no source), 3 integrations. **85 leaf tests passing.** |

---

## 11. Whitespace Control

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `{%-`, `-%}`, `{{-`, `-}}` | OK. Ported tests in `whitespace_ctrl_ported_test.go` (43 tests covering Ruby and JS). Intensive E2E in `s11_whitespace_e2e_test.go` (105 tests) cover: all trim directions (left/right/both) on tags and outputs, inline markers on all tag types (for/if/unless/case/assign/capture/increment/decrement/echo/liquid/raw/comment), tag+output combinations, nested templates (2 and 3 levels), edge cases (tabs, CR, strings with spaces, empty arrays, multiple adjacent nodes). |
| ✅ | ✅ | ✅ | P1 | `{{-}}` — blank trim without expression | **Bug fixed:** `{{-}}` (no expression between trim markers) produced `Liquid syntax error: syntax error in "-"`. Fix in `parser/scanner.go`: (1) when the captured Args is `"-"` and a trim marker is present, replace with `""`. (2) output token content regex updated to use exclusion pattern identical to tags (`(?:[^}]|}[^}])+?`), preventing greedy match that crossed adjacent `{{-}}` tokens. Fix in `parser/parser.go`: `ObjTokenType` token with `Args == ""` is ignored. E2E in `s11_whitespace_e2e_test.go`: `TrimBlank_*` (8 tests) covers isolated, multiple spaces, newlines, adjacent, multiple `{{-}}`, inside for, inside capture, next to output. |
| ✅ | ✅ | ✅ | P4 | Global trim options (`trimTagRight`, etc.) | Implemented: `Config.TrimTagLeft/Right`, `TrimOutputLeft/Right`, `Greedy` in `render/config.go`. Engine exposes `SetTrimTagLeft/Right`, `SetTrimOutputLeft/Right`, `SetGreedy`. `Greedy` default = true. Non-greedy (inline blank + 1 newline) implemented in `trimwriter.go`. Ported tests from `trimming.spec.ts` passing. E2E in `s11_whitespace_e2e_test.go`: `Global_TrimTag*` (11 tests), `Global_TrimOutput*` (9 tests), `Greedy_*` (7 tests), `Interaction_*` (4 tests) — cover option combinations with inline markers, isolation between tag/output trim, non-greedy behavior, and interaction between different trim mechanisms. |

---

## 12. Thread-safety and Concurrency

> It makes no sense to guarantee immutability before having all configuration fields defined. Can be planned in parallel, but implemented after stabilizing the configuration API.
> See also **B5** (active race condition bug in the renderer).

| Impl | Tests | E2E | Priority | Item | Notes |
|------|-------|-----|----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Mutable state audit in `Engine` | **Completed.** Grammar maps (`tags`, `blockDefs`), filter maps, and `Config.Globals` are written only during setup and read during render — race-free. `Expression.Variables()` uses `sync.Once` — correct. `engine.cache *sync.Map` for template cache is thread-safe. `Cache` (fallback for `{% include %}`) was `map[string][]byte` — fixed to `sync.Map` (see row below). Engine is 100% safe for concurrent use after setup. |
| ✅ | ✅ | ✅ | P1 | Render state isolated per call | **Confirmed safe.** `newNodeContext(vars, cfg)` does `maps.Copy` of globals+scope into a new map per call. Stateful tags (assign, increment, decrement, cycle, for+continue) operate only on the per-call map. Compiled expressions are immutable after parse. Verified in `TestConcurrent_StatefulTagsAreIsolated`. |
| ✅ | ✅ | ✅ | P2 | `Config` immutable after construction | **Implemented via freeze pattern.** `Engine` has `frozen atomic.Bool`. `freeze()` is called at the start of every parse entry point (`ParseTemplate`, `ParseTemplateLocation`, `ParseString`, `ParseAndRender`, `ParseAndFRender`, `ParseTemplateAndCache`). `checkNotFrozen(method)` is called in all 21 mutating configuration methods (`RegisterTag/Block/Filter`, `StrictVariables`, `LaxFilters/Tags`, `SetGlobals`, `SetTrimXxx`, `SetGreedy`, `SetGlobalFilter`, `SetExceptionHandler`, `SetAutoEscapeReplacer`, `RegisterTemplateStore`, `Delims`, `EnableCache`, `EnableJekyllExtensions`, `RegisterTagAnalyzer/BlockAnalyzer`). Violation results in panic with clear message: `"liquid: SetGlobals() called after the engine has been used for parsing"`. Zero overhead on hot path. Exception documented: `UnregisterTag` has no guard — explicitly for hot-reload/test teardown. 3 tests in `context_scope_ported_test.go` had `RegisterTag` after `ParseTemplateAndCache` — fixed to correct order. 42 subtests in `TestEngine_FrozenAfterParse` (21 entries × 2) + `TestEngine_FrozenPanicMessage` cover all methods. |
| ✅ | ✅ | ✅ | P1 | Fix: `Cache map[string][]byte` → `sync.Map` | **Fixed.** `render/config.go`: `Cache` field changed to `sync.Map`. `engine.go`: `Cache[path] = source` → `Cache.Store(path, source)`. `render/context.go`: two `Cache[filename]` → `Cache.Load(filename)` (with type assertion `.([]byte)`). `tags/include_tag_test.go`: two `config.Cache["..."] = []byte(...)` → `Cache.Store(...)`. `NewConfig()`: removed `Cache: map[string][]byte{}` initialization (zero value of `sync.Map` is already valid). `TestConcurrent_CacheRace` now tests real behavior (without `t.Skip`) — passed. |

---

## Executive Summary by Priority

### P1 — Core Shopify Liquid (implement first)

```
Tags:
[x] echo tag                 ✅ DONE
[x] liquid tag (multi-line)  ✅ DONE — tags/standard_tags.go, tests in TestLiquidTag
[x] # inline comment         ✅ DONE — parser/scanner.go, tests in TestInlineComment
[x] increment / decrement    ✅ DONE — tags/standard_tags.go, separate counters, tests in TestIncrementDecrement
[x] render tag (isolated scope) ✅ DONE — tags/render_tag.go, with/as/kv/for, tests in TestRenderTag_*
[x] include — with/as/key-val args ✅ DONE — tags/include_tag.go rewritten, tests in TestIncludeTag_*
[x] case/when — support for `or`  ✅ DONE

Filters:
[x] capitalize — fix (lowercase rest)          ✅ DONE
[x] strip_html — fix (remove script/style)     ✅ DONE
[x] newline_to_br — fix (preserve \n)          ✅ DONE
[x] modulo — fix (error on division by zero)   ✅ DONE (guard added)
[x] default — allow_false keyword arg          ✅ DONE (filter updated + tests)
[x] sort — nil-last (nils go to the end)       ✅ DONE
[x] Keyword args in filters (parser change)    ✅ DONE (NamedArg infrastructure)

Expressions:
[x] empty literal/drop        ✅ DONE
[x] blank literal/drop        ✅ DONE
[x] Strings — escape support (\n, \", etc.)  ✅ DONE
[x] array[-1] negative indexing              ✅ DONE

Drops:
[x] forloop.name              ✅ DONE (already implemented — confirmed)
[x] tablerowloop drop — row/col/col0/col_first/col_last ✅ DONE (already implemented — confirmed)

Context:
[x] Isolated sub-context (for render tag) ✅ DONE
[x] Global variables separate from scope  ✅ DONE
```

### P2 — Common Extensions (Ruby + JS)

```
[x] strict_variables / strict_filters as per-render options  ✅ DONE — WithStrictVariables(), WithLaxFilters(), WithGlobals(), WithGlobalFilter() in liquid.go
[x] globals option on engine  ✅ DONE
```

### P3 — Ruby Compat

```
[x] squish filter              ✅ DONE
[x] h alias (escape)           ✅ DONE
[x] base64_url_safe_encode/decode  ✅ DONE
[x] compact: property arg      ✅ DONE
[x] uniq: property arg         ✅ DONE
[x] forloop.parentloop         ✅ DONE (already implemented — confirmed)
[x] <> alias for !=            ✅ DONE
[x] doc / enddoc tag           ✅ DONE — special parser like comment, tests in TestDocTag
[x] ifchanged tag              ✅ DONE — tags/standard_tags.go, tests in TestIfchangedTag
[x] include for array as alias ✅ DONE — tags/include_tag.go, tests in TestIncludeTag_for_array
[x] Drop: liquid_method_missing  ✅ DONE — DropMethodMissing in drops.go, tests in drops_test.go
[x] context= injection in drop  ✅ DONE — ContextDrop/DropRenderContext in drops.go, expressions/context.go, tests in drops_test.go
[x] template.errors / error collection  ✅ DONE — WithErrorHandler() as collector
[x] exception_renderer  ✅ DONE — WithErrorHandler() + Engine.SetExceptionHandler()
[x] Resource limits (render_length)  ✅ DONE — WithSizeLimit(int64)
[x] ParseTreeVisitor API  ✅ DONE — Walk + ParseTree in visitor.go
```

### P4 — JS Compat / Extensions

```
[x] for offset: continue  ✅ DONE — tags/iteration_tags.go, all loops track position, tests in TestOffsetContinue
[x] date: 'now'/'today' as input  ✅ DONE
[x] date_to_xmlschema / date_to_rfc822 / date_to_string / date_to_long_string  ✅ DONE — filters/standard_filters.go, JS ported tests in filters/standard_filters_test.go
[x] where_exp / reject_exp / group_by_exp / has_exp / find_exp / find_index_exp  ✅ DONE
[x] jsonify alias              ✅ DONE
[x] raw filter  ✅ DONE — expressions/filters.go (registered with safe), render/config.go (always enabled), tests in render/autoescape_test.go
[x] layout / block tags        ✅ DONE — tags/layout_tags.go, block inheritance, tests in TestLayoutTag*
[x] not unary operator         ✅ DONE
[ ] Global whitespace trim options
[x] Resource limits (time-based via context)  ✅ DONE — WithContext(context.Context)
[x] Template cache  ✅ DONE — Engine.EnableCache() / ClearCache()
```
