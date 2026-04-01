# Parity Checklist: liquid-go vs LiquidJS

Complete mapping of [LiquidJS](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) entities to their equivalents in this project.

**Legend:**
- ✅ Implemented
- ❌ Not implemented
- 🔧 Partially implemented / different API
- 🚫 Not applicable (concept doesn't exist in Go or intentionally omitted)

---

## 1. Engine API (Liquid class vs Engine struct)

### Parsing

| LiquidJS | Go | Status |
|---|---|---|
| [`parse(html, filepath?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`ParseString`](../engine.go) / [`ParseTemplate`](../engine.go) / [`ParseTemplateLocation`](../engine.go) | ✅ |
| [`parseFile(file, lookupType?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ (Go uses `TemplateStore`) |
| [`parseFileSync(file, lookupType?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ (Go uses `TemplateStore`) |

### Rendering

| LiquidJS | Go | Status |
|---|---|---|
| [`parseAndRender(html, scope?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`ParseAndRender`](../engine.go) / [`ParseAndRenderString`](../engine.go) | ✅ |
| [`parseAndRenderSync(html, scope?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`ParseAndRenderString`](../engine.go) | 🔧 (Go doesn't distinguish sync/async) |
| [`render(tpl, scope?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Template.Render`](../template.go) / [`Template.FRender`](../template.go) | ✅ |
| [`renderSync(tpl, scope?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Template.RenderString`](../template.go) | 🔧 |
| [`renderFile(file, ctx?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`renderFileSync(file, ctx?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`renderToNodeStream(tpl, scope?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Template.FRender(w, vars)`](../template.go) | 🔧 (Go uses `io.Writer`) |
| [`renderFileToNodeStream(file, scope?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |

### Isolated expression eval

| LiquidJS | Go | Status |
|---|---|---|
| [`evalValue(str, scope?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`evalValueSync(str, scope?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |

### Static analysis

| LiquidJS | Go | Status |
|---|---|---|
| [`analyze(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`render.Analyze(root)`](../render/analysis.go) | 🔧 (internal API; public via VariableSegments/GlobalVariableSegments) |
| [`analyzeSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`render.Analyze(root)`](../render/analysis.go) | 🔧 (Go is sync by default) |
| [`parseAndAnalyze(html, filename?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`parseAndAnalyzeSync(html, filename?, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`variables(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`variablesSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`fullVariables(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`fullVariablesSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`variableSegments(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Engine.VariableSegments`](../analysis.go) / [`Template.VariableSegments`](../analysis.go) | ✅ |
| [`variableSegmentsSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Engine.VariableSegments`](../analysis.go) | 🔧 (Go is sync by default) |
| [`globalVariables(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`globalVariablesSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`globalFullVariables(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`globalFullVariablesSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`globalVariableSegments(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Engine.GlobalVariableSegments`](../analysis.go) / [`Template.GlobalVariableSegments`](../analysis.go) | ✅ |
| [`globalVariableSegmentsSync(tpl, opts?)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`Engine.GlobalVariableSegments`](../analysis.go) | 🔧 (Go is sync by default) |

### Extension & configuration

| LiquidJS | Go | Status |
|---|---|---|
| [`registerFilter(name, filter)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`RegisterFilter`](../engine.go) | ✅ |
| [`registerTag(name, tag)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | [`RegisterTag`](../engine.go) / [`RegisterBlock`](../engine.go) | ✅ |
| [`plugin(fn)`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | ❌ |
| [`express()`](../../.example-repositories/liquid-js/liquidjs/src/liquid.ts) | — | 🚫 (Node.js-specific) |

---

## 2. Template API

| LiquidJS `Template` | Go `Template` | Status |
|---|---|---|
| [`render(ctx, emitter)`](../../.example-repositories/liquid-js/liquidjs/src/template/template.ts) | [`Render(vars)`](../template.go) / [`FRender(w, vars)`](../template.go) / [`RenderString(b)`](../template.go) | ✅ |
| `children?(partials, sync)` — analysis | — | ❌ (part of analysis plan) |
| `arguments?()` — analysis | — | ❌ (part of analysis plan) |
| `blockScope?()` — analysis | — | ❌ (part of analysis plan) |
| `localScope?()` — analysis | — | ❌ (part of analysis plan) |
| `partialScope?()` — analysis | — | ❌ (part of analysis plan) |
| — | [`GetRoot()`](../template.go) | 🚫 (Go-specific) |

---

## 3. Configuration (LiquidOptions vs Engine methods/Config)

| LiquidJS `LiquidOptions` | Go | Status |
|---|---|---|
| `tagDelimiterLeft/Right`, `outputDelimiterLeft/Right` | [`Engine.Delims()`](../engine.go) | ✅ |
| `strictVariables` | [`Engine.StrictVariables()`](../engine.go) | ✅ |
| `strictFilters` | [`Engine.LaxFilters()`](../engine.go) (inverted) | 🔧 |
| `outputEscape` | [`Engine.SetAutoEscapeReplacer()`](../engine.go) | 🔧 (different API) |
| `cache` | [`ParseTemplateAndCache()`](../engine.go) | 🔧 (manual, no LRU) |
| `root/partials/layouts` | [`Engine.RegisterTemplateStore()`](../engine.go) | 🔧 (no default dirs) |
| `trimTagLeft/Right`, `trimOutputLeft/Right`, `greedy` | — | ❌ (Go uses `{%-` / `-%}` syntax) |
| `globals` | — | ❌ |
| `jsTruthy` | — | ❌ |
| `dynamicPartials` | — | ❌ |
| `jekyllInclude` | [`Engine.EnableJekyllExtensions()`](../engine.go) | 🔧 |
| `ownPropertyOnly` | — | ❌ |
| `lenientIf` | — | ❌ |
| `orderedFilterParameters` | — | ❌ |
| `keepOutputType` | — | ❌ |
| `dateFormat`, `timezoneOffset`, `locale` | — | ❌ |
| `parseLimit`, `renderLimit`, `memoryLimit` | — | ❌ |
| `fs` (custom filesystem) | `RegisterTemplateStore` | 🔧 |
| `extname` | — | ❌ |
| `relativeReference` | — | ❌ |
| `keyValueSeparator` | — | ❌ |
| `operators` (custom) | — | ❌ |

---

## 4. Built-in tags

| Tag | LiquidJS | Go | Status |
|---|---|---|---|
| `assign` | [`/tags/assign.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/assign.ts) | [`standard_tags.go`](../tags/standard_tags.go) | ✅ |
| `capture` | [`/tags/capture.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/capture.ts) | [`standard_tags.go`](../tags/standard_tags.go) | ✅ |
| `case/when` | [`/tags/case.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/case.ts) | [`control_flow_tags.go`](../tags/control_flow_tags.go) | ✅ |
| `comment` | [`/tags/comment.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/comment.ts) | [`standard_tags.go`](../tags/standard_tags.go) | ✅ |
| `cycle` | [`/tags/cycle.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/cycle.ts) | [`iteration_tags.go`](../tags/iteration_tags.go) | ✅ |
| `for` | [`/tags/for.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/for.ts) | [`iteration_tags.go`](../tags/iteration_tags.go) | ✅ |
| `if/elsif/else` | [`/tags/if.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/if.ts) | [`control_flow_tags.go`](../tags/control_flow_tags.go) | ✅ |
| `unless` | [`/tags/unless.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/unless.ts) | [`control_flow_tags.go`](../tags/control_flow_tags.go) | ✅ |
| `raw` | [`/tags/raw.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/raw.ts) | [`standard_tags.go`](../tags/standard_tags.go) | ✅ |
| `tablerow` | [`/tags/tablerow.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/tablerow.ts) | [`iteration_tags.go`](../tags/iteration_tags.go) | ✅ |
| `break` | [`/tags/break.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/break.ts) | [`iteration_tags.go`](../tags/iteration_tags.go) | ✅ |
| `continue` | [`/tags/continue.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/continue.ts) | [`iteration_tags.go`](../tags/iteration_tags.go) | ✅ |
| `include` | [`/tags/include.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/include.ts) | [`include_tag.go`](../tags/include_tag.go) | 🔧 (no scope isolation) |
| `render` | [`/tags/render.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/render.ts) | — | ❌ (partial with isolated scope) |
| `layout` | [`/tags/layout.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/layout.ts) | — | ❌ (template inheritance) |
| `block` | [`/tags/block.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/block.ts) | — | ❌ (named blocks) |
| `echo` | [`/tags/echo.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/echo.ts) | — | ❌ |
| `liquid` | [`/tags/liquid.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/liquid.ts) | — | ❌ (multi-tag on single line) |
| `increment` | [`/tags/increment.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/increment.ts) | — | ❌ |
| `decrement` | [`/tags/decrement.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/decrement.ts) | — | ❌ |
| `#` (inline comment) | [`/tags/inline-comment.ts`](../../.example-repositories/liquid-js/liquidjs/src/tags/inline-comment.ts) | — | ❌ |

---

## 5. Built-in filters

### String

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `append` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `prepend` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `capitalize` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `downcase` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `upcase` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `remove` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `remove_first` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `remove_last` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | — | ❌ |
| `replace` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `replace_first` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `replace_last` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | — | ❌ |
| `split` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `strip` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `lstrip` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `rstrip` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `strip_newlines` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `truncate` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `truncatewords` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `normalize_whitespace` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | — | ❌ |
| `number_of_words` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | — | ❌ |
| `array_to_sentence_string` | [`string.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/string.ts) | — | ❌ |

### Array

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `join` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `first` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `last` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `reverse` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `sort` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `sort_natural` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `size` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `map` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `compact` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `concat` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `slice` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `uniq` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `where` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `where_exp` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `sum` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `push` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `unshift` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `pop` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `shift` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `reject` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `reject_exp` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `group_by` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `group_by_exp` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `has` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `has_exp` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `find` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `find_exp` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `find_index` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `find_index_exp` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |
| `sample` | [`array.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/array.ts) | — | ❌ |

### Math

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `abs` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `at_least` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | — | ❌ |
| `at_most` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | — | ❌ |
| `ceil` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `divided_by` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `floor` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `minus` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `modulo` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `plus` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `round` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `times` | [`math.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/math.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |

### HTML

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `escape` | [`html.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/html.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `escape_once` | [`html.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/html.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `newline_to_br` | [`html.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/html.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `strip_html` | [`html.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/html.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `xml_escape` | [`html.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/html.ts) | — | ❌ |

### URL

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `url_encode` | [`url.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/url.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `url_decode` | [`url.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/url.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `cgi_escape` | [`url.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/url.ts) | — | ❌ |
| `uri_escape` | [`url.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/url.ts) | — | ❌ |
| `slugify` | [`url.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/url.ts) | — | ❌ |

### Date

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `date` | [`date.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/date.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `date_to_xmlschema` | [`date.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/date.ts) | — | ❌ |
| `date_to_rfc822` | [`date.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/date.ts) | — | ❌ |
| `date_to_string` | [`date.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/date.ts) | — | ❌ |
| `date_to_long_string` | [`date.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/date.ts) | — | ❌ |

### Base64

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `base64_encode` | [`base64.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/base64.ts) | — | ❌ |
| `base64_decode` | [`base64.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/base64.ts) | — | ❌ |

### Misc

| Filter | LiquidJS | Go | Status |
|---|---|---|---|
| `default` | [`misc.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/misc.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `json` / `jsonify` | [`misc.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/misc.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `inspect` | [`misc.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/misc.ts) | [`standard_filters.go`](../filters/standard_filters.go) | ✅ |
| `to_integer` | [`misc.ts`](../../.example-repositories/liquid-js/liquidjs/src/filters/misc.ts) | — | ❌ |
| — | [`standard_filters.go`](../filters/standard_filters.go) `type` | 🚫 (Go-specific, not in spec) |

---

## 6. Types and Protocols

### Drop (custom object protocol)

| LiquidJS | Go | Status |
|---|---|---|
| [`Drop` abstract class](../../.example-repositories/liquid-js/liquidjs/src/drop/drop.ts) with `liquidMethodMissing(key, ctx)` | [`Drop interface { ToLiquid() any }`](../drops.go) | 🔧 (Go has no per-property fallback) |
| [`Comparable` interface](../../.example-repositories/liquid-js/liquidjs/src/drop/comparable.ts) (`equals`, `gt`, `lt`, etc.) | — | ❌ |
| `NullDrop`, `EmptyDrop`, `BlankDrop` | — | ❌ |
| `ForloopDrop` (exposes `index`, `length`, `first`, `last`, etc.) | — | 🔧 (Go uses internal `map[string]any`) |
| `TablerowloopDrop` | — | 🔧 (Go uses internal `map[string]any`) |
| `BlockDrop` (`block.super`) | — | ❌ |

### Static analysis

| LiquidJS | Go | Status |
|---|---|---|
| [`StaticAnalysis`](../../.example-repositories/liquid-js/liquidjs/src/template/analysis.ts) (`variables`, `globals`, `filters`, `tags`) | — | ❌ |
| [`Variable`](../../.example-repositories/liquid-js/liquidjs/src/template/analysis.ts) (reference with paths/segments) | — | ❌ |
| Tag interface: `arguments()`, `localScope()`, `blockScope()`, `children()`, `partialScope()` | — | ❌ (part of analysis plan) |

### Errors

| LiquidJS | Go | Status |
|---|---|---|
| `LiquidError` | `SourceError` interface | 🔧 |
| `ParseError` | — | 🔧 (Go uses `SourceError`) |
| `RenderError` | — | 🔧 (Go uses `SourceError`) |
| `UndefinedVariableError` | — | ❌ |
| `TokenizationError` | — | ❌ |

### Context

| LiquidJS | Go | Status |
|---|---|---|
| [`Context`](../../.example-repositories/liquid-js/liquidjs/src/context/context.ts) with `globals`, `environments`, `breakCalled`, `continueCalled` | [`render.Context`](../render/context.go) interface | 🔧 |
| `ctx.getRegister()` / `ctx.setRegister()` | — | ❌ |
| `globals` scope injected via options | — | ❌ |

---

## Gap summary

| Category | LiquidJS total | Go implemented | Missing |
|---|---|---|---|
| Engine API — parsing/rendering | ~12 | ~8 | ~4 |
| Engine API — static analysis | 16 | 2 | 14 |
| Template API | ~7 | 4 | ~3 |
| Configuration | ~28 | ~8 | ~20 |
| Tags | 21 | 13 | 8 |
| Filters (string) | 21 | 18 | 3 |
| Filters (array) | 28 | 12 | 16 |
| Filters (math) | 11 | 9 | 2 |
| Filters (html) | 5 | 4 | 1 |
| Filters (url) | 5 | 2 | 3 |
| Filters (date) | 5 | 1 | 4 |
| Filters (base64) | 2 | 0 | 2 |
| Filters (misc) | 5 | 4 | 1 |
| Drop / custom types | ~6 | ~1 | ~5 |
| Static analysis (types) | ~5 | 0 | ~5 |
