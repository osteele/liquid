# Ruby Liquid — Complete Feature Mapping

> Reference extracted directly from the source code in `.example-repositories/liquid-ruby/liquid` (lib/ + test/).
> Organized by domain. Used as the basis for comparison with the Go implementation.

---

## Tags

### Output / Expression tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `{{ }}` | `{{ expression }}` | Variable or expression output with filters |
| `echo` | `{% echo expression %}` | Equivalent to `{{ }}`, usable inside `{% liquid %}` |

### Variable / State tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `assign` | `{% assign var = value %}` | Creates variable; tracked by resource limits |
| `capture` | `{% capture var %}...{% endcapture %}` | Captures output as string; tracked by resource limits |
| `increment` | `{% increment var %}` | Starts at 0, outputs then increments; shares state with `decrement` |
| `decrement` | `{% decrement var %}` | Starts at -1, outputs then decrements; shares state with `increment` |

### Conditional tags

| Tag | Sub-tags | Notes |
|-----|----------|-------|
| `if` | `elsif`, `else` | Operators: `==`, `!=`, `<>`, `<`, `>`, `<=`, `>=`, `contains`, `and`, `or`; special values: `blank`, `empty` |
| `unless` | `elsif`, `else` | Inverts initial condition; rest same as `if` |
| `case` | `when`, `else` | `when` supports multiple values separated by `or` or `,` |
| `ifchanged` | — | Renders only if output changed since last iteration; state in `registers[:ifchanged]` |

### Iteration tags

| Tag | Options | Notes |
|-----|---------|-------|
| `for` | `reversed`, `limit: n`, `offset: n`, range `(a..b)` | Sub-tag `else` (when array empty); creates `forloop` object; supports `break`/`continue` |
| `break` | — | Interrupts `for`; uses `BreakInterrupt` |
| `continue` | — | Skips iteration in `for`; uses `ContinueInterrupt` |
| `cycle` | optional name: `{% cycle "name": v1, v2 %}` | State in `registers[:cycle]`; must be inside `for` |
| `tablerow` | `cols: n`, `limit: n`, `offset: n`, range `(a..b)` | Generates HTML table (`<tr class="rowN">`, `<td class="colN">`); creates `tablerowloop` object; supports `break` |

### Template inclusion tags

| Tag | Syntax | Scope | Status |
|-----|--------|-------|--------|
| `include` | `{% include 'file' [with var] [for array] [as alias] [key: val] %}` | Shared (variable leak) | **Deprecated** |
| `render` | `{% render 'file' [with var] [for array] [as alias] [key: val] %}` | Isolated (no access to parent vars, except globals) | Current |

### Text / Structure tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `raw` | `{% raw %}...{% endraw %}` | Literal output, bypasses rendering |
| `comment` | `{% comment %}...{% endcomment %}` | Ignored; supports `comment`/`raw` nesting |
| `#` (inline comment) | `{%# comment %}` | Single line; each line needs `#` |
| `liquid` | `{% liquid tag1\ntag2 %}` | Multi-line without `{% %}` delimiters; each line is a tag |
| `doc` | `{% doc %}...{% enddoc %}` | LiquidDoc documentation; ignored by renderer |

---

## Filters (StandardFilters)

### String

| Filter | Signature | Notes |
|--------|-----------|-------|
| `downcase` | `string \| downcase` | |
| `upcase` | `string \| upcase` | |
| `capitalize` | `string \| capitalize` | Capitalizes first letter, lowercases the rest |
| `escape` | `string \| escape` | HTML escape (`<`, `>`, `&`, `"`, `'`); alias `h` |
| `escape_once` | `string \| escape_once` | HTML escape without re-escaping already escaped |
| `url_encode` | `string \| url_encode` | CGI escape; spaces become `+` |
| `url_decode` | `string \| url_decode` | CGI unescape |
| `base64_encode` | `string \| base64_encode` | Strict Base64 |
| `base64_decode` | `string \| base64_decode` | Raises error if invalid |
| `base64_url_safe_encode` | `string \| base64_url_safe_encode` | URL-safe Base64 |
| `base64_url_safe_decode` | `string \| base64_url_safe_decode` | URL-safe Base64 decode |
| `slice` | `string \| slice: offset[, length]` | Also works on arrays |
| `truncate` | `string \| truncate: n[, ellipsis]` | Default ellipsis `"..."`, included in count |
| `truncatewords` | `string \| truncatewords: n[, ellipsis]` | Default 15 words |
| `split` | `string \| split: separator` | Returns array |
| `strip` | `string \| strip` | Removes whitespace from both sides |
| `lstrip` | `string \| lstrip` | Removes whitespace from the left |
| `rstrip` | `string \| rstrip` | Removes whitespace from the right |
| `strip_html` | `string \| strip_html` | Removes HTML tags + `<script>`, `<style>`, comments |
| `strip_newlines` | `string \| strip_newlines` | Removes `\n` and `\r\n` |
| `squish` | `string \| squish` | Strip + collapses internal whitespace to a single space |
| `newline_to_br` | `string \| newline_to_br` | Converts `\n` to `<br />` |
| `replace` | `string \| replace: old, new` | Replaces all occurrences |
| `replace_first` | `string \| replace_first: old, new` | Replaces only the first |
| `replace_last` | `string \| replace_last: old, new` | Replaces only the last |
| `remove` | `string \| remove: sub` | Removes all occurrences |
| `remove_first` | `string \| remove_first: sub` | Removes only the first |
| `remove_last` | `string \| remove_last: sub` | Removes only the last |
| `append` | `string \| append: suffix` | Concatenates at the end |
| `prepend` | `string \| prepend: prefix` | Concatenates at the beginning |

### Array

| Filter | Signature | Notes |
|--------|-----------|-------|
| `size` | `array \| size` | Also works on strings |
| `first` | `array \| first` | |
| `last` | `array \| last` | |
| `join` | `array \| join[: glue]` | Default glue `" "` |
| `split` | `string \| split: sep` | Inverse of `join` |
| `reverse` | `array \| reverse` | |
| `sort` | `array \| sort[: property]` | Case-sensitive; nil-safe (nils go to end) |
| `sort_natural` | `array \| sort_natural[: property]` | Case-insensitive |
| `uniq` | `array \| uniq[: property]` | Removes duplicates |
| `compact` | `array \| compact[: property]` | Removes nils |
| `map` | `array \| map: property` | Extracts property from each item |
| `concat` | `array \| concat: other_array` | Combines arrays (no dedup) |
| `where` | `array \| where: prop[, value]` | Filters by property; without value = truthy |
| `reject` | `array \| reject: prop[, value]` | Inverse of `where` |
| `find` | `array \| find: prop[, value]` | First match |
| `find_index` | `array \| find_index: prop[, value]` | Index of first match |
| `has` | `array \| has: prop[, value]` | `true` if any item satisfies |
| `sum` | `array \| sum[: property]` | Numeric sum |
| `slice` | `array \| slice: offset[, length]` | Array slice |

### Math

| Filter | Signature | Notes |
|--------|-----------|-------|
| `abs` | `number \| abs` | |
| `plus` | `number \| plus: n` | |
| `minus` | `number \| minus: n` | |
| `times` | `number \| times: n` | |
| `divided_by` | `number \| divided_by: n` | Result type = divisor type; raises `ZeroDivisionError` |
| `modulo` | `number \| modulo: n` | Raises `ZeroDivisionError` |
| `round` | `number \| round[: decimals]` | Default 0 decimals |
| `ceil` | `number \| ceil` | |
| `floor` | `number \| floor` | |
| `at_least` | `number \| at_least: n` | `max(input, n)` |
| `at_most` | `number \| at_most: n` | `min(input, n)` |

### Date

| Filter | Signature | Notes |
|--------|-----------|-------|
| `date` | `date \| date: format` | strftime format; returns input if empty/invalid |

### Misc / Default

| Filter | Signature | Notes |
|--------|-----------|-------|
| `default` | `var \| default: val[, allow_false: bool]` | **Keyword argument** `allow_false`; returns default if nil, false, or empty |

---

## Filter System

| Feature | Description |
|---------|-------------|
| Positional filters | `{{ val \| filter: arg1, arg2 }}` |
| **Filters with keyword args** | `{{ val \| default: fallback, allow_false: true }}` — passed as hash to the method |
| Mixed positional + keyword | Supported in `strict2_parse` mode |
| `register_filter(module)` | Registers Ruby module as filter source |
| `strict_filters` | If `true`, raises `UndefinedFilter` for unknown filters |
| `global_filter` | Proc applied to every expression output before rendering |

---

## Expressions and Operators

### Literals

| Literal | Example |
|---------|---------|
| nil/null | `nil`, `null` |
| boolean | `true`, `false` |
| integer | `42`, `-1` |
| float | `3.14` |
| string | `"text"` or `'text'` |
| range | `(1..10)` |
| blank | `blank` → `''` (compares as empty string) |
| empty | `empty` → `''` (compares as empty string) |

### Comparison operators

| Operator | Behavior |
|----------|----------|
| `==` | Equality (nil-safe) |
| `!=`, `<>` | Inequality |
| `<`, `>`, `<=`, `>=` | Numeric/string comparison |
| `contains` | String: `include?`; Array: `include?` |

### Boolean operators

| Operator | Behavior |
|----------|----------|
| `and` | Short-circuit, evaluates left to right without precedence |
| `or` | Short-circuit |

### Truthiness

| Value | Truthy? |
|-------|---------|
| `false` | falsy |
| `nil` | falsy |
| `0` | **truthy** |
| `""` | **truthy** |
| `[]` | **truthy** |
| any other | truthy |

### Variable access

| Syntax | Description |
|--------|-------------|
| `variable` | Lookup in stacked scopes |
| `obj.prop` | Property access |
| `obj[key]` | String key access |
| `array[0]` | Integer index access |
| `array.first`, `array.last` | Shortcuts |
| `array.size`, `hash.size` | Size |
| `forloop.index`, etc. | Loop properties |

---

## Drops (custom object protocol)

| Feature | Description |
|---------|-------------|
| `Drop` base class | Base class; public methods are accessible by name |
| `invoke_drop(key)` / `[key]` | Invokes method or calls `liquid_method_missing` |
| `liquid_method_missing(name)` | Catch-all; raises `UndefinedDropMethod` if `strict_variables` |
| `key?(_name)` | Always returns `true` by default |
| `invokable_methods` | Whitelist of public methods minus blacklist |
| System blacklist | Methods of `Drop` + `Enumerable` (except `sort`, `count`, `first`, `min`, `max`) |
| `to_liquid` | Returns `self`; used for lazy conversion |
| `context=` | Injects render context into the drop |

### ForloopDrop (`forloop` object)

| Field | Description |
|-------|-------------|
| `index` | 1-based |
| `index0` | 0-based |
| `rindex` | 1-based reverse |
| `rindex0` | 0-based reverse |
| `first` | boolean |
| `last` | boolean |
| `length` | total iterations |
| `parentloop` | parent loop drop (or nil) |
| `name` | loop identifier |

### TablerowloopDrop (`tablerowloop` object)

| Field | Description |
|-------|-------------|
| `index`, `index0` | 1-based and 0-based iteration |
| `rindex`, `rindex0` | reverse |
| `first`, `last` | boolean |
| `col` | 1-based column |
| `col0` | 0-based column |
| `col_first`, `col_last` | boolean |
| `row` | 1-based row |
| `length` | total iterations |

---

## Errors

| Type | Usage |
|------|-------|
| `Liquid::Error` | Base |
| `SyntaxError` | Parse error |
| `ArgumentError` | Invalid argument for filter/tag |
| `ContextError` | Error in context operation |
| `FileSystemError` | Error loading file |
| `StandardError` | Generic runtime error |
| `StackLevelError` | Include stack overflow |
| `MemoryError` | Resource limits exceeded |
| `ZeroDivisionError` | Division by zero |
| `FloatDomainError` | Invalid float operation |
| `UndefinedVariable` | Variable not found (strict mode) |
| `UndefinedDropMethod` | Drop method not found (strict mode) |
| `UndefinedFilter` | Filter not registered (strict_filters) |
| `MethodOverrideError` | Attempt to override forbidden method |
| `DisabledError` | Tag disabled (Disableable) |
| `InternalError` | Internal engine error |
| `TemplateEncodingError` | Invalid encoding in template |

**Common attributes:** `line_number`, `template_name`, `markup_context`

---

## Error Modes (error_mode)

| Mode | Behavior |
|------|---------|
| `:lax` | Silently ignores syntax errors in most cases (Liquid 2.5 compat) |
| `:warn` | **Default**; warnings for invalid syntax |
| `:strict` | Raises error for most tags with incorrect syntax |
| `:strict2` | Raises error for all tags with incorrect syntax (strictest mode) |

---

## Resource Limits

| Limit | Description |
|-------|-------------|
| `render_length_limit` | Maximum bytes in a template output |
| `render_score_limit` | Render score per template (each node counts) |
| `assign_score_limit` | Assign score per template (based on bytesize) |
| `cumulative_render_score_limit` | Cumulative render score (multiple renders) |
| `cumulative_assign_score_limit` | Cumulative assign score |
| `MemoryError` | Raised when any limit is exceeded |

**Assign scoring:** String → bytesize; Array/Hash → recursive sum of elements; others → 1.

---

## Environment / Configuration

| Feature | Description |
|---------|-------------|
| `error_mode` | `:lax`, `:warn`, `:strict`, `:strict2` |
| `file_system` | Pluggable filesystem implementation for `include`/`render` |
| `exception_renderer` | Proc to intercept exceptions |
| `default_resource_limits` | Hash with default limits (see above) |
| `register_tag(name, klass)` | Registers custom tag |
| `register_filter(module)` | Registers filter module |
| `Environment.build {}` | Immutable builder (freeze after construction) |
| `Environment.dangerously_override` | Temporary override of default environment (block) |

### Render options (per call)

| Option | Description |
|--------|-------------|
| `filters:` | Additional filter modules for the render |
| `registers:` | Hash of static registers |
| `global_filter:` | Proc applied to all variable output |
| `exception_renderer:` | Per-render override |
| `strict_variables:` | Raises `UndefinedVariable` for non-existent vars |
| `strict_filters:` | Raises `UndefinedFilter` for unknown filters |

---

## Template API

| Method | Description |
|--------|-------------|
| `Template.parse(source, options)` | Parse at class level (uses `Environment.default`) |
| `template.render(*args)` | Render with variables / context / drops |
| `template.render!(*args)` | Render with error rethrow |
| `template.errors` | Array of accumulated errors |
| `template.warnings` | Parse warnings |
| `template.resource_limits` | `ResourceLimits` object |
| `template.root` | Root node of the parse tree |
| `template.name` | Template name (for error messages) |

---

## File System

| Interface | Description |
|-----------|-------------|
| `BlankFileSystem` | Default; raises error on any `include`/`render` |
| `LocalFileSystem.new(root, pattern)` | Reads from disk; default pattern `_%s.liquid` |
| `LocalFileSystem#read_template_file(path)` | Reads file; validates path (no path traversal) |
| Custom interface | `read_template_file(path)` — any object that responds to this method |

---

## Static Analysis (ParseTreeVisitor)

| Feature | Description |
|---------|-------------|
| `ParseTreeVisitor.for(node, callbacks)` | Creates visitor for node |
| `visitor.add_callback_for(*classes) { \|node, ctx\| }` | Registers callback by class |
