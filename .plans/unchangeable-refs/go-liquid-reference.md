# Go Liquid — Complete Feature Mapping

> Reference extracted directly from the source code in `c:\Users\joca\github.com\joaqu1m\liquid`.
> Organized following the same structure as `ruby-liquid-reference.md` for easy comparison.
> All file paths are relative to the repository root.

---

## Tags

### Output / Expression tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `{{ }}` | `{{ expression }}` | Variable or expression output with filters. Type: `ObjTokenType` → `ObjectNode`. |

> **Absent:** `echo`, `liquid` (multi-line), `#` (inline comment)

---

### Variable / State tags

| Tag | Type | Syntax | Notes |
|-----|------|--------|-------|
| `assign` | simple tag | `{% assign var = expr %}` | Evaluates expression, sets variable in scope. With `EnableJekyllExtensions()`, supports dot notation: `{% assign page.prop = expr %}` (`Path []string`). Has analyzer that reports `LocalScope` + `Arguments`. |
| `capture` | block tag | `{% capture varname %}...{% endcapture %}` | Renders body as string, assigns to variable. Requires exactly one variable name. Has analyzer that reports `LocalScope`. |

> **Absent:** `increment`, `decrement`

---

### Conditional tags

| Tag | Sub-tags | Notes |
|-----|----------|-------|
| `if` | `elsif`, `else` | Operators: `==`, `!=`, `<>` (via NEQ), `<`, `>`, `<=`, `>=`, `contains`, `and`, `or`. Truthy: not `nil` and not `false`. Has static analyzer. |
| `unless` | `else` | Inverts initial condition via `Not(expr)`. Uses the same compiler (`ifTagCompiler(false)`). Has the same analyzer as `if`. |
| `case` | `when`, `else` | Evaluates `case` expression, compares with `values.Equal()`. `when` supports multiple values separated by **comma**. Has static analyzer. |

> **Absent:** `ifchanged`

> **Note on `case`/`when`:** Values separated by `or` in Ruby's `when` clause are not supported — only comma (`,`).

---

### Iteration tags

| Tag | Options | Notes |
|-----|---------|-------|
| `for` | `reversed`, `limit: n`, `offset: n`, range `(a..b)` | Sub-tag `else` (when collection empty). Creates `forloop` object. Supports `break`/`continue`. Iterates over array, range, map. Has static analyzer (`BlockScope` for loop var, `Arguments` for collection expr). |
| `break` | — | Returns sentinel `errLoopBreak`. Only valid inside `for`/`tablerow`. |
| `continue` | — | Returns sentinel `errLoopContinueLoop`. Only valid inside `for`/`tablerow`. |
| `tablerow` | `cols: n`, `limit: n`, `offset: n`, range `(a..b)` | Same loop engine as `for`. Generates table HTML: `<tr class="rowN">...<td class="colN">...</td></tr>`. Creates same `forloop` object. |
| `cycle` | optional name: `{% cycle "name": v1, v2 %}` | Must be inside `for`. Reads `forloop[".cycles"]` to track position per group. Group prefix with `:`. |

---

### `forloop` object (created by `for` and `tablerow`)

| Field | Type | Description |
|-------|------|-------------|
| `forloop.first` | bool | `true` on first iteration |
| `forloop.last` | bool | `true` on last iteration |
| `forloop.index` | int | 1-based index |
| `forloop.index0` | int | 0-based index |
| `forloop.rindex` | int | 1-based reverse index |
| `forloop.rindex0` | int | 0-based reverse index |
| `forloop.length` | int | Total iterations |
| `.cycles` (internal) | map | Tracks position of `cycle` groups |

> **Absent vs Ruby:** `forloop.parentloop`, `forloop.name`

---

### Template inclusion tags

| Tag | Syntax | Scope | Notes |
|-----|--------|-------|-------|
| `include` | `{% include "filename" %}` | **Shared** (parent bindings are passed + overridden by additional vars) | Resolves path relative to `SourceFile()`. Uses `TemplateStore.ReadTemplate()`. Implemented in `tags/include_tag.go`. |

> **Absent:** `render` (isolated scope)

---

### Text / Structure tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `raw` | `{% raw %}...{% endraw %}` | Literal output, bypasses rendering. Parser sets `inRaw = true`. AST type: `ASTRaw` → `RawNode`. |
| `comment` | `{% comment %}...{% endcomment %}` | Parser sets `inComment = true`, skips all tokens until `endcomment`. Internal tags do not need to be balanced. |

> **Absent:** `#` (inline comment), `liquid` (multi-line), `doc`, `echo`

---

## Filters

### String (24 filters)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `append` | `string \| append: suffix` | Concatenates at end |
| `prepend` | `string \| prepend: prefix` | Concatenates at beginning |
| `upcase` | `string \| upcase` | |
| `downcase` | `string \| downcase` | |
| `capitalize` | `string \| capitalize` | Uppercase first letter; empty string unchanged |
| `escape` | `string \| escape` | HTML escape using `html.EscapeString` |
| `escape_once` | `string \| escape_once` | Unescapes first, then escapes — avoids double escaping |
| `strip` | `string \| strip` | `strings.TrimSpace` |
| `lstrip` | `string \| lstrip` | Removes whitespace on the left (via `unicode.IsSpace`) |
| `rstrip` | `string \| rstrip` | Removes whitespace on the right |
| `newline_to_br` | `string \| newline_to_br` | Converts `\n` to `<br />` |
| `strip_html` | `string \| strip_html` | Removes HTML tags via regex `<.*?>` (may be insufficient for complex cases) |
| `strip_newlines` | `string \| strip_newlines` | Removes all `\n` and `\r\n` |
| `truncate` | `string \| truncate[: n[, ellipsis]]` | Default: n=50, ellipsis=`"..."`. Rune-aware. |
| `truncatewords` | `string \| truncatewords[: n[, ellipsis]]` | Default: n=15, ellipsis=`"..."` |
| `split` | `string \| split: sep` | Returns array; space separator is special (split on whitespace runs); trailing empty strings removed |
| `replace` | `string \| replace: old, new` | `strings.ReplaceAll` |
| `replace_first` | `string \| replace_first: old, new` | Replaces only the first |
| `replace_last` | `string \| replace_last: old, new` | Replaces only the last (via `strings.LastIndex`) |
| `remove` | `string \| remove: sub` | Removes all occurrences |
| `remove_first` | `string \| remove_first: sub` | Removes only the first |
| `remove_last` | `string \| remove_last: sub` | Removes only the last |
| `normalize_whitespace` | `string \| normalize_whitespace` | Collapses whitespace runs to a single space (**Jekyll extension**) |
| `number_of_words` | `string \| number_of_words[: mode]` | Counts words. Modes: `"default"`, `"cjk"`, `"auto"` (**Jekyll extension**) |

> **Absent vs Ruby:** `squish` (Ruby collapses + strips; the equivalent here is `normalize_whitespace` but without automatic strip)

---

### Array (22 filters)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `size` | `array \| size` | Also works on strings (rune count) and maps. Returns 0 for other types. |
| `first` | `array \| first` | Returns nil for empty array |
| `last` | `array \| last` | Returns nil for empty array |
| `join` | `array \| join[: glue]` | Default glue `" "`. Skips nil items. |
| `reverse` | `array \| reverse` | Returns new array |
| `sort` | `array \| sort[: key]` | Ascending; supports sort by map/struct key. Defined in `filters/sort_filters.go`. |
| `sort_natural` | `array \| sort_natural[: key]` | Case-insensitive; supports key. |
| `uniq` | `array \| uniq` | Removes duplicates. O(1) for comparable types, O(n²) fallback. |
| `compact` | `array \| compact` | Removes nils |
| `map` | `array \| map: property` | Extracts property from each item |
| `concat` | `array \| concat: other_array` | Combines two arrays (no dedup) |
| `where` | `array \| where: prop[, value]` | Filters where property == value; without value = truthy. `filters/array_filters.go`. |
| `reject` | `array \| reject: prop[, value]` | Inverse of `where`; without value = falsy |
| `find` | `array \| find: prop[, value]` | First item that satisfies; returns nil if not found |
| `find_index` | `array \| find_index: prop[, value]` | 0-based index of first match; nil if not found |
| `has` | `array \| has: prop[, value]` | Returns bool; `true` if any item satisfies |
| `sum` | `array \| sum[: property]` | Numeric sum; preserves int type if no floats; parses strings; skips non-numeric |
| `slice` | `array \| slice: start[, length]` | Array or string slice. Supports negative start (from end). Rune-aware for strings. Works on `string`, `[]byte`, slices. |
| `group_by` | `array \| group_by: property` | Groups by property value; returns `[{"name": ..., "items": [...]}]` |
| `push` | `array \| push: element` | Adds at end, returns new array |
| `unshift` | `array \| unshift: element` | Adds at beginning, returns new array |
| `pop` | `array \| pop` | Removes last element, returns new array |
| `shift` | `array \| shift` | Removes first element, returns new array |
| `sample` | `array \| sample[: n]` | Returns n random elements. If n=1, returns single element; otherwise array. |

> **Note:** `push`, `unshift`, `pop`, `shift`, `sample`, `group_by` are extensions not present in standard Ruby Liquid.

---

### Math (11 filters)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `abs` | `number \| abs` | `math.Abs` (float64) |
| `plus` | `number \| plus: n` | Preserves int type if both are int |
| `minus` | `number \| minus: n` | Preserves int type if both are int |
| `times` | `number \| times: n` | Preserves int type if both are int |
| `divided_by` | `number \| divided_by: n` | Integer division if divisor is int; float if float; returns error on division by zero |
| `modulo` | `number \| modulo: n` | `math.Mod` (float) |
| `round` | `number \| round[: decimals]` | Default 0 decimal places |
| `ceil` | `number \| ceil` | |
| `floor` | `number \| floor` | |
| `at_least` | `number \| at_least: n` | `max(input, n)` — minimum clamp |
| `at_most` | `number \| at_most: n` | `min(input, n)` — maximum clamp |

---

### Date (1 filter)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `date` | `date \| date[: format]` | strftime format via `tuesday` library. Default `"%a, %b %d, %y"`. Supports multiple parse formats (ANSIC, RFC3339, ISO 8601, Ruby, etc.). |

**Accepted parse formats:** `ANSIC`, `UnixDate`, `RubyDate`, `RFC822`, `RFC822Z`, `RFC850`, `RFC1123`, `RFC1123Z`, `RFC3339`, ISO 8601, `"2006-01-02"`, among others. See `values/parsedate.go`.

---

### HTML / URL / Encoding (8 filters)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `url_encode` | `string \| url_encode` | `url.QueryEscape` |
| `url_decode` | `string \| url_decode` | `url.QueryUnescape`; returns error on invalid input |
| `base64_encode` | `string \| base64_encode` | Standard Base64 encoding |
| `base64_decode` | `string \| base64_decode` | Standard Base64 decode; returns error on invalid input |
| `xml_escape` | `string \| xml_escape` | Escapes `& < > " '` |
| `cgi_escape` | `string \| cgi_escape` | `url.QueryEscape` (**Jekyll extension**) |
| `uri_escape` | `string \| uri_escape` | URI-encode preserving safe chars (equiv. to JS `encodeURI()`). (**Jekyll extension**) |
| `slugify` | `string \| slugify[: mode]` | Converts to URL slug. Modes: `"default"` (unicode), `"ascii"`, `"latin"` (transliterate accents), `"pretty"` (preserves URL chars), `"none"`/`"raw"` (lowercase only). (**Jekyll extension**) |

> **Absent vs Ruby:** `base64_url_safe_encode`, `base64_url_safe_decode`

---

### Value / Type (4 filters)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `default` | `var \| default: val` | Returns default if value is nil, false, or empty string/array |
| `json` | `var \| json` | Serializes to JSON via `json.Marshal` |
| `to_integer` | `var \| to_integer` | Converts to int; handles int/float/string/bool (true=1, false=0) |
| `array_to_sentence_string` | `array \| array_to_sentence_string[: connector]` | Joins array as English sentence: `"a, b, and c"`. Default connector `"and"`. (**Jekyll extension**) |

> **Note vs Ruby:** `default` in this implementation does **not** support keyword argument `allow_false: true`.

---

### Debug (2 filters)

| Filter | Signature | Notes |
|--------|-----------|-------|
| `inspect` | `var \| inspect` | JSON or `%#v`. (**Jekyll extension**) |
| `type` | `var \| type` | Returns Go type name (`%T`). (**proprietary extension**) |

---

## Filter System

| Feature | Description |
|---------|-------------|
| Positional filters | `{{ val \| filter: arg1, arg2 }}` |
| `AddFilter(name, fn)` | Registers Go function as filter. Fn must have ≥1 input and 1 or 2 outputs (2nd if present: `error`). |
| `LaxFilters()` | Engine method. Silently passes input when filter is unknown (Shopify behavior). Default: unknown filter is an error. |
| `UndefinedFilter` | Error type in `expressions` package. String with filter name. |
| `FilterError` | Error type in `expressions`. Contains `FilterName string` and `Err error`. |
| `safe` filter | Registered automatically by `SetAutoEscapeReplacer()`. Marks value as safe to avoid escaping in auto-escape. |
| **No keyword args** | Ruby's `allow_false: true` is **not supported** — filters do not receive a kwargs hash. |

---

## Expressions and Operators

### Literals

| Literal | Example | Notes |
|---------|---------|-------|
| nil | `nil` | Token: `LITERAL` → `nil` |
| boolean | `true`, `false` | Token: `LITERAL` → `bool` |
| integer | `42`, `-1` | `strconv.ParseInt`, Go type: `int` |
| float | `3.14`, `-0.5` | `strconv.ParseFloat`, Go type: `float64` |
| string | `"text"` or `'text'` | **No escape support inside strings** (TODO in code) |
| range | `(1..10)` | Type `values.Range`; supports variables: `(a..b)` |

> **Absent vs Ruby:** `blank`, `empty` as comparable literals. In this impl, `blank` and `empty` are just identifiers treated as undefined variables (nil).

### Comparison operators

| Operator | Token | Behavior |
|----------|-------|----------|
| `==` | `EQ` | `values.Equal()` — nil-safe, supports int/float/string/bool |
| `!=` | `NEQ` | `!values.Equal()` |
| `<>` | `NEQ` | Alias for `!=` (same token in scanner) |
| `<` | `'<'` | `values.Less()` |
| `>` | `'>'` | `values.Less()` inverted |
| `<=` | `LE` | `Less || Equal` |
| `>=` | `GE` | `Less(b,a) || Equal` |
| `contains` | `CONTAINS` | String: `strings.Contains`; Array: `reflect` search; Map: key lookup |

### Boolean operators

| Operator | Token | Behavior |
|----------|-------|----------|
| `and` | `AND` | `fa.Test() && fb.Test()` — no true short-circuit, both evaluated |
| `or` | `OR` | `fa.Test() \|\| fb.Test()` |

### Truthiness

| Value | Truthy? |
|-------|---------|
| `false` | falsy |
| `nil` | falsy |
| `0` | **truthy** |
| `""` | **truthy** |
| `[]` | **truthy** |
| any other | truthy |

Implemented via `Value.Test()` in `values/value.go`.

### Variable access

| Syntax | Description |
|--------|-------------|
| `variable` | Lookup in `ctx.Get(name)` → `values.ToLiquid(bindings[name])` |
| `obj.prop` | Token `PROPERTY` → `makeObjectPropertyExpr()` |
| `obj[key]` | Expr `[expr]` → `makeIndexExpr()` |
| `array[0]` | Integer index via `IndexValue()` |
| `array.first`, `array.last` | Special properties in `arrayValue.PropertyValue()` |
| `array.size`, `hash.size` | Property `size` returns length |
| `forloop.index`, etc. | Properties of the forloop object (map in Go) |

### Identifiers

- Support Unicode (letters, digits, `_`, `-` except as first character)
- Can end with `?` (Ruby-style predicates)

---

## Drops (custom object protocol)

### `Drop` interface (`liquid.Drop`)

| Feature | Description |
|---------|-------------|
| `Drop` interface | Defined in `liquid/drops.go`: `ToLiquid() any` |
| `FromDrop(object any) any` | Public function: if `object` implements `Drop`, returns `object.ToLiquid()`; otherwise returns the object itself |
| Lazy resolution | `dropWrapper` in `values/drop.go` uses `sync.Once` — `ToLiquid()` is called only on first evaluation |
| `values.ToLiquid(value)` | Converts object to Liquid if it implements the internal `drop` interface |

> **Absent vs Ruby:** No `Drop` base class with `liquid_method_missing`, `invokable_methods`, blacklist, `context=`, `key?`. The Go protocol is only `ToLiquid() any`.

### `IterationKeyedMap` (`tags.IterationKeyedMap`)

| Feature | Description |
|---------|-------------|
| `IterationKeyedMap` | Public type: `map[string]any`. When iterated in `for`, yields are the **keys** (not key/value pairs). |
| `liquid.IterationKeyedMap(m)` | Public helper function to create the wrapper |

### `yaml.MapSlice` (internal support)

| Feature | Description |
|---------|-------------|
| `yaml.MapSlice` | Type `gopkg.in/yaml.v2.MapSlice`. Iteration in `for` preserves insertion order. Key lookup via `mapSliceValue`. |

### `values.SafeValue`

| Feature | Description |
|---------|-------------|
| `SafeValue{Value: v}` | Type in `values/value.go`. Marks value as safe for auto-escape. Used by `safe` filter. |

---

## Struct access from templates

Go structs are accessible via PropertyValue (via reflection):
- Exported fields mapped by name
- Exported methods mapped by name
- `structValue.PropertyValue()` in `values/structvalue.go`

---

## Errors

### `SourceError` / `parser.Error` / `render.Error`

| Interface | Methods | Notes |
|-----------|---------|-------|
| `liquid.SourceError` | `Error() string`, `Cause() error`, `Path() string`, `LineNumber() int` | Public interface. Returned by `ParseTemplate`, `Render`, etc. |
| `parser.Error` | Same methods | Internal interface; compatible with `SourceError` |
| `render.Error` | Same methods | Internal interface |

The concrete implementation is `parser.sourceLocError`:
- `Error()` formats as `"Liquid error (line N): message in path"`
- `Cause()` returns the original error
- `Path()` returns template pathname
- `LineNumber()` returns line number

### Error types in `expressions`

| Type | Description |
|------|-------------|
| `InterpreterError` | `string` — expression interpretation error (invalid input) |
| `UndefinedFilter` | `string` — filter not defined |
| `FilterError` | struct with `FilterName string`, `Err error` — error applying filter |
| `values.TypeError` | `string` — type conversion error |

> **Absent vs Ruby:** No distinct error types for `SyntaxError`, `ArgumentError`, `ContextError`, `FileSystemError`, `MemoryError`, `ZeroDivisionError`, `UndefinedVariable`, `UndefinedDropMethod`, `MethodOverrideError`, `DisabledError`, `TemplateEncodingError`, etc.

---

## Engine — Public API

### Creation

| Function | Description |
|----------|-------------|
| `liquid.NewEngine() *Engine` | Full engine with default filters and tags |
| `liquid.NewBasicEngine() *Engine` | Engine without default filters/tags |

### Parse

| Method | Signature | Description |
|--------|-----------|-------------|
| `ParseTemplate` | `(source []byte) (*Template, SourceError)` | Basic parse |
| `ParseString` | `(source string) (*Template, SourceError)` | String wrapper |
| `ParseTemplateLocation` | `(source []byte, path string, line int) (*Template, SourceError)` | Parse with location for errors and `include` |
| `ParseTemplateAndCache` | `(source []byte, path string, line int) (*Template, SourceError)` | Parse + internal cache (`cfg.Cache[path]`) |

### Combined Parse + Render

| Method | Signature | Description |
|--------|-----------|-------------|
| `ParseAndRender` | `(source []byte, b Bindings) ([]byte, SourceError)` | |
| `ParseAndFRender` | `(w io.Writer, source []byte, b Bindings) SourceError` | Render directly to writer |
| `ParseAndRenderString` | `(source string, b Bindings) (string, SourceError)` | |

### Configuration

| Method | Description |
|--------|-------------|
| `StrictVariables()` | Undefined variable produces error |
| `LaxFilters()` | Undefined filter silently passes input |
| `EnableJekyllExtensions()` | Enables dot notation in `assign` (`page.prop = value`) |
| `Delims(objL, objR, tagL, tagR string) *Engine` | Customizes delimiters. Empty string = default. |
| `SetAutoEscapeReplacer(replacer Replacer)` | Enables auto-escape. Registers `safe` filter automatically. |

### Extension registration

| Method | Description |
|--------|-------------|
| `RegisterTag(name string, td Renderer)` | Registers simple tag. `Renderer = func(render.Context) (string, error)`. |
| `RegisterBlock(name string, td Renderer)` | Registers block tag. |
| `RegisterFilter(name string, fn any)` | Registers filter. Fn: ≥1 input, 1 or 2 outputs (2nd = error). |
| `RegisterTemplateStore(ts render.TemplateStore)` | Replaces TemplateStore (file source for `include`). |
| `RegisterTagAnalyzer(name, a TagAnalyzer)` | Registers analyzer for custom tag. |
| `RegisterBlockAnalyzer(name, a BlockAnalyzer)` | Registers analyzer for custom block tag. |
| `UnregisterTag(name string)` | Removes tag by name (idempotent). |

### Static analysis

| Method | Return | Description |
|--------|--------|-------------|
| `GlobalVariableSegments(t)` | `([]VariableSegment, error)` | Global variable paths |
| `VariableSegments(t)` | `([]VariableSegment, error)` | All variable paths |
| `GlobalVariables(t)` | `([]string, error)` | Unique root names of global vars |
| `Variables(t)` | `([]string, error)` | Unique root names of all vars |
| `GlobalFullVariables(t)` | `([]Variable, error)` | Global refs with path + location |
| `FullVariables(t)` | `([]Variable, error)` | All refs with path + location + `Global` flag |
| `Analyze(t)` | `(*StaticAnalysis, error)` | Full analysis: vars, globals, locals, tags |
| `ParseAndAnalyze(source)` | `(*Template, *StaticAnalysis, error)` | Parse + analysis in one step |

---

## Template — Public API

### Render

| Method | Signature | Description |
|--------|-----------|-------------|
| `Render` | `(vars Bindings) ([]byte, SourceError)` | Full render to bytes |
| `FRender` | `(w io.Writer, vars Bindings) SourceError` | Render directly to writer |
| `RenderString` | `(b Bindings) (string, SourceError)` | String wrapper |

### AST

| Method | Return | Description |
|--------|--------|-------------|
| `GetRoot()` | `render.Node` | Returns root node of the parse tree |

### Static analysis (methods on Template)

Same as on Engine, but as convenience methods directly on `*Template`:

`GlobalVariableSegments()`, `VariableSegments()`, `GlobalVariables()`, `Variables()`, `GlobalFullVariables()`, `FullVariables()`, `Analyze()`

---

## Public Types

### `liquid.Bindings`

```go
type Bindings map[string]any
```
Documentation alias for `map[string]any`.

### `liquid.Renderer`

```go
type Renderer func(render.Context) (string, error)
```
Type for custom tag definitions.

### `liquid.VariableSegment`

```go
type VariableSegment = []string
```
Path to variable as a segment slice.

### `liquid.Variable`

```go
type Variable struct {
    Segments []string
    Location parser.SourceLoc
    Global   bool
}
```
With method `String() string` returning dot-separated path.

### `liquid.StaticAnalysis`

```go
type StaticAnalysis struct {
    Variables []Variable
    Globals   []Variable
    Locals    []string
    Tags      []string
    Filters   []string // always nil for now
}
```

### `liquid.Drop` interface

```go
type Drop interface {
    ToLiquid() any
}
```

### `liquid.SourceError` interface

```go
type SourceError interface {
    error
    Cause() error
    Path() string
    LineNumber() int
}
```

---

## Static Analysis (render.NodeAnalysis / render.AnalysisResult)

| Type/Feature | Description |
|-------------|-------------|
| `render.NodeAnalysis` | `Arguments []Expression`, `LocalScope []string`, `BlockScope []string` |
| `render.TagAnalyzer` | `func(args string) NodeAnalysis` |
| `render.BlockAnalyzer` | `func(node BlockNode) NodeAnalysis` |
| `render.VariableRef` | `Path []string`, `Loc parser.SourceLoc` |
| `render.AnalysisResult` | `Globals`, `All`, `GlobalRefs`, `AllRefs`, `Locals`, `Tags` |
| `render.Analyze(root Node)` | Main analysis function; traverses AST collecting variables, locals, tags |
| `expressions.Expression.Variables()` | Returns `[][]string` with expression variable paths (lazy, cached) |

Standard tags with analyzers: `assign` (LocalScope + Arguments), `capture` (LocalScope), `if`/`unless`/`case` (Arguments), `for`/`tablerow` (BlockScope + Arguments).

---

## Render Context (`render.Context` interface)

Public interface for custom tag implementors:

| Method | Description |
|--------|-------------|
| `Bindings() map[string]any` | Current full lexical environment |
| `Get(name string) any` | Gets variable from current environment |
| `Set(name string, value any)` | Sets variable in current environment |
| `SetPath(path []string, value any) error` | Sets variable at nested path (used by assign with dot notation) |
| `Evaluate(expr expressions.Expression) (any, error)` | Evaluates compiled expression |
| `EvaluateString(source string) (any, error)` | Compiles and evaluates expression string |
| `ExpandTagArg() (string, error)` | Renders current tag argument as Liquid template (for Jekyll `{% include {{ var }} %}`) |
| `InnerString() (string, error)` | Renders current block body as string (for `capture`, `highlight`) |
| `RenderBlock(w io.Writer, b *BlockNode) error` | Renders a BlockNode |
| `RenderChildren(w io.Writer) Error` | Renders current node's children |
| `RenderFile(filename string, b map[string]any) (string, error)` | Parses + renders external file (used by `include`) |
| `SourceFile() string` | Current template path (for relative `include`) |
| `TagArgs() string` | Text of current tag's arguments |
| `TagName() string` | Current tag name |
| `Errorf(format, a...) Error` | Creates error with source location |
| `WrapError(err error) Error` | Wraps error with location |

---

## TemplateStore (file system)

| Interface/Type | Description |
|----------------|-------------|
| `render.TemplateStore` interface | `ReadTemplate(name string) ([]byte, error)` |
| `render.FileTemplateStore{}` | Default implementation; uses `os.ReadFile(filename)` directly |
| `Engine.RegisterTemplateStore(ts)` | Replaces the default implementation |
| Internal cache (`cfg.Cache`) | `map[string][]byte`; populated by `ParseTemplateAndCache()`; consulted by `include` when file not found on disk |

---

## Auto-escape

| Feature | Description |
|---------|-------------|
| `render.Replacer` interface | `WriteString(io.Writer, string) (int, error)` |
| `render.HtmlEscaper` | `strings.NewReplacer` for `& ' < > "` |
| `Engine.SetAutoEscapeReplacer(r)` | Enables auto-escape globally in engine; registers `safe` filter |
| `safe` filter | Marks value as `values.SafeValue{}` to skip auto-escape |
| `values.SafeValue` | Struct `{ Value any }` — type-safe wrapper |

---

## Whitespace Control (Trimmer)

| Marker | Effect |
|--------|--------|
| `{%-` | Removes whitespace before tag (TrimLeft) |
| `-%}` | Removes whitespace after tag (TrimRight) |
| `{{-` | Removes whitespace before output |
| `-}}` | Removes whitespace after output |

Implemented via `render.trimWriter` in `render/trimwriter.go`.  
AST: `ASTTrim` (parser) → `TrimNode` (render).

---

## Custom Delimiters

```go
engine.Delims("{{", "}}", "{%", "%}")
```
- Can be called before any `Parse*`.
- Empty string = uses default.
- Regexps compiled and cached by delimiter combination.
