# LiquidJS — Complete Feature Mapping

> Reference extracted directly from the source code in `.example-repositories/liquid-js/liquidjs` (src/).
> Organized by domain. Used as the basis for comparison with the Go implementation.

---

## Tags

### Output / Expression tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `{{ }}` | `{{ expression }}` | Variable or expression output with filters |
| `echo` | `{% echo expression %}` | Equivalent to `{{ }}`; usable inside `{% liquid %}`; value is optional (no value emits nothing) |

### Variable / State tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `assign` | `{% assign var = value %}` | Creates variable in lower scope (`ctx.bottom()`); supports filters on value |
| `capture` | `{% capture var %}...{% endcapture %}` | Captures output as string; name can be identifier or quoted string |
| `increment` | `{% increment var %}` | Stored in `context.environments`; starts at 0, emits current value then increments |
| `decrement` | `{% decrement var %}` | Stored in `context.environments`; starts at 0, decrements before emitting (-1, -2, …) |

### Conditional tags

| Tag | Sub-tags | Notes |
|-----|----------|-------|
| `if` | `elsif`, `else`, `endif` | Operators: `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `not`, `and`, `or`; `else` cannot be duplicated; `elsif` cannot appear after `else` |
| `unless` | `elsif`, `else`, `endunless` | Inverts initial condition; `elsif` uses `isTruthy` (not inverted); `else` goes to `elseTemplates` |
| `case` | `when`, `else`, `endcase` | `when` supports multiple values separated by `or` or `,`; `else` only one; branches after `else` are ignored |

### Iteration tags

| Tag | Options | Notes |
|-----|---------|-------|
| `for` | `offset: n\|continue`, `limit: n`, `reversed` | Sub-tag `else` when collection empty; creates `forloop` object; supports `break`/`continue`; `offset: continue` uses `continueKey` register to resume from previous point; by default applies modifiers in order: `offset` → `limit` → `reversed` (or declaration order if `orderedFilterParameters: true`) |
| `break` | — | Sets `ctx.breakCalled = true`; interrupts `for` at end of current iteration |
| `continue` | — | Sets `ctx.continueCalled = true`; skips rest of current iteration in `for` |
| `cycle` | `[group:] v1, v2, [v3, ...]` | State in register `'cycle'`; key: `"cycle:{group}:{candidates}"`; group is optional; without group the key includes candidate list |
| `tablerow` | `cols: n`, `offset: n`, `limit: n` | Generates HTML table (`<tr class="rowN">`, `<td class="colN">`); creates `tablerowloop` object; without `cols` uses width = collection size |

### Template inclusion tags

| Tag | Syntax | Scope | Notes |
|-----|--------|-------|-------|
| `include` | `{% include 'file' [with var] [key: val ...] %}` | Shared (non-isolated, variable leak) | `with var` adds `filepath` as key in scope; `jekyllInclude: true` injects everything into `include` object instead of direct scope |
| `render` | `{% render 'file' [with var [as alias]] [for collection [as alias]] [key: val ...] %}` | Isolated (`ctx.spawn()`) | `with` exposes single variable; `for` iterates collection exposing item + `forloop`; `as alias` renames the exposed variable |

### Layout / inheritance tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `layout` | `{% layout 'file' [key: val ...] %}` | Layout inheritance; consumes remaining template tokens as child templates; `{% layout none %}` (or `{% layout %}` without file) disables layout and renders directly |
| `block` | `{% block name %}...{% endblock %}` | Defines overridable block in layout context; supports `{{ block.super }}` to render parent content |

### Text / Structure tags

| Tag | Syntax | Notes |
|-----|--------|-------|
| `raw` | `{% raw %}...{% endraw %}` | Literal output, bypasses rendering; internal tokens emitted as text |
| `comment` | `{% comment %}...{% endcomment %}` | Ignored; consumes tokens but does not render |
| `#` (inline comment) | `{%# comment %}` | Single line; multi-line requires `#` on each line (`/\n\s*[^#\s]/g` throws error) |
| `liquid` | `{% liquid tag1\ntag2\n... %}` | Multi-tag without `{% %}` delimiters; uses `readLiquidTagTokens` |

---

## Filters

### String

| Filter | Signature | Notes |
|--------|-----------|-------|
| `append` | `v \| append: arg` | Concatenates `arg` at end; requires exactly 2 arguments |
| `prepend` | `v \| prepend: arg` | Prepends `arg` at beginning; requires exactly 2 arguments |
| `downcase` | `v \| downcase` | `toLowerCase()` |
| `upcase` | `v \| upcase` | `toUpperCase()` |
| `capitalize` | `v \| capitalize` | First letter uppercase, rest lowercase |
| `remove` | `v \| remove: arg` | Removes all occurrences of `arg` |
| `remove_first` | `v \| remove_first: l` | Removes first occurrence |
| `remove_last` | `v \| remove_last: l` | Removes last occurrence (searches with `lastIndexOf`) |
| `replace` | `v \| replace: pattern, replacement` | Replaces all occurrences |
| `replace_first` | `v \| replace_first: arg1, arg2` | Replaces first occurrence |
| `replace_last` | `v \| replace_last: arg1, arg2` | Replaces last occurrence (searches with `lastIndexOf`) |
| `split` | `v \| split: arg` | Splits by delimiter; trailing empty strings removed (Ruby behavior) |
| `lstrip` | `v \| lstrip[: chars]` | Left strip; without argument: whitespace; with argument: char set |
| `rstrip` | `v \| rstrip[: chars]` | Right strip; without argument: whitespace; with argument: char set |
| `strip` | `v \| strip[: chars]` | Both sides strip; without argument: `trim()`; with argument: char set |
| `strip_newlines` | `v \| strip_newlines` | Removes `\r?\n` |
| `truncate` | `v \| truncate[: l[, o]]` | Truncates to `l` chars (default 50); suffix `o` (default `'...'`) |
| `truncatewords` | `v \| truncatewords[: words[, o]]` | Truncates to `words` words (default 15); suffix `o` (default `'...'`); `words <= 0` uses 1 |
| `normalize_whitespace` | `v \| normalize_whitespace` | Collapses whitespace to single space |
| `number_of_words` | `input \| number_of_words[: mode]` | Counts words; modes: `'cjk'` (CJK + non-CJK), `'auto'` (CJK if present, else split), default (split by space) |
| `array_to_sentence_string` | `array \| array_to_sentence_string[: connector]` | `"a, b, and c"`; `connector` default `'and'`; 0 → `''`, 1 → `a`, 2 → `a and b` |

### Math

| Filter | Signature | Notes |
|--------|-----------|-------|
| `abs` | `v \| abs` | `Math.abs` |
| `at_least` | `v \| at_least: min` | `Math.max(v, min)` |
| `at_most` | `v \| at_most: max` | `Math.min(v, max)` |
| `ceil` | `v \| ceil` | `Math.ceil` |
| `floor` | `v \| floor` | `Math.floor` |
| `round` | `v \| round[: decimals]` | Rounds to `decimals` decimal places (default 0); uses `Math.round(v * 10^d) / 10^d` |
| `divided_by` | `v \| divided_by: divisor[, integerArithmetic]` | Division; `integerArithmetic=true` uses `Math.floor` (integer division) |
| `minus` | `v \| minus: arg` | Subtraction |
| `plus` | `v \| plus: arg` | Addition |
| `modulo` | `v \| modulo: arg` | Modulo (`%`) |
| `times` | `v \| times: arg` | Multiplication |

### HTML

| Filter | Signature | Notes |
|--------|-----------|-------|
| `escape` | `v \| escape` | HTML escape: `&`→`&amp;`, `<`→`&lt;`, `>`→`&gt;`, `"`→`&#34;`, `'`→`&#39;` |
| `xml_escape` | `v \| xml_escape` | Alias for `escape` |
| `escape_once` | `v \| escape_once` | Unescape then escape (idempotent; does not re-escape already escaped entities) |
| `newline_to_br` | `v \| newline_to_br` | Replaces `\r?\n` with `<br />\n` |
| `strip_html` | `v \| strip_html` | Removes `<script>`, `<style>`, HTML tags and HTML comments |

### URL

| Filter | Signature | Notes |
|--------|-----------|-------|
| `url_encode` | `v \| url_encode` | `encodeURIComponent`, spaces become `+` |
| `url_decode` | `v \| url_decode` | `decodeURIComponent`, `+` becomes space |
| `cgi_escape` | `v \| cgi_escape` | `encodeURIComponent` with `+` and uppercase hex for `!'()*` |
| `uri_escape` | `v \| uri_escape` | `encodeURI` preserving `[` and `]` |
| `slugify` | `v \| slugify[: mode[, cased]]` | Slugifies string; modes: `'raw'`, `'default'`, `'pretty'`, `'ascii'`, `'latin'`, `'none'`; `cased=false` by default (lowercase) |

### Date

| Filter | Signature | Notes |
|--------|-----------|-------|
| `date` | `v \| date[: format[, timezoneOffset]]` | strftime formatting; `'now'`/`'today'` → current time; numeric string/number → epoch seconds; respects `opts.preserveTimezones`; default format via `opts.dateFormat` |
| `date_to_xmlschema` | `v \| date_to_xmlschema` | `date` with format `'%Y-%m-%dT%H:%M:%S%:z'` |
| `date_to_rfc822` | `v \| date_to_rfc822` | `date` with format `'%a, %d %b %Y %H:%M:%S %z'` |
| `date_to_string` | `v \| date_to_string[: type[, style]]` | Abbreviated month (`%b`); `type='ordinal'`, `style='US'` uses American format |
| `date_to_long_string` | `v \| date_to_long_string[: type[, style]]` | Full month (`%B`); same options as `date_to_string` |

### Array

| Filter | Signature | Notes |
|--------|-----------|-------|
| `join` | `arr \| join[: sep]` | Joins with separator (default: `' '`); nil → `' '` |
| `first` | `arr \| first` | First element; string/array returns `''` if not array-like |
| `last` | `arr \| last` | Last element; string/array returns `''` if not array-like |
| `reverse` | `arr \| reverse` | Creates reversed copy (`[...arr].reverse()`) |
| `sort` | `arr \| sort[: property]` | Sorts; `property` can be path with `.`; `<` to compare |
| `sort_natural` | `arr \| sort_natural[: property]` | Case-insensitive sorting |
| `size` | `v \| size` | String or array length; `0` if nil |
| `map` | `arr \| map: property` | Extracts property from each item |
| `sum` | `arr \| sum[: property]` | Sums values (NaN → 0); with `property` sums the property of each item |
| `compact` | `arr \| compact` | Removes nil values (using `toValue`) |
| `concat` | `arr \| concat: arr2` | Concatenates two arrays; `arr2` default `[]` |
| `push` | `arr \| push: item` | Adds item at end (returns new array) |
| `unshift` | `arr \| unshift: item` | Adds item at beginning (returns new array) |
| `pop` | `arr \| pop` | Removes last element (returns new array) |
| `shift` | `arr \| shift` | Removes first element (returns new array) |
| `slice` | `arr \| slice: begin[, length]` | Slices string or array; `begin < 0` → from end; `length` default `1` |
| `where` | `arr \| where: property[, expected]` | Filters items where `property == expected`; without `expected` → filters truthy; with `jekyllWhere: true` also accepts array membership |
| `reject` | `arr \| reject: property[, expected]` | Inverse of `where` |
| `where_exp` | `arr \| where_exp: itemName, exp` | Filters by Liquid expression; `exp` evaluated with `itemName` in scope |
| `reject_exp` | `arr \| reject_exp: itemName, exp` | Inverse of `where_exp` |
| `group_by` | `arr \| group_by: property` | Groups into `[{name, items}, ...]` |
| `group_by_exp` | `arr \| group_by_exp: itemName, exp` | Groups by Liquid expression |
| `has` | `arr \| has: property[, expected]` | Returns boolean; `true` if any item matches |
| `has_exp` | `arr \| has_exp: itemName, exp` | Returns boolean via expression |
| `find` | `arr \| find: property[, expected]` | Returns first matching item (or `undefined`) |
| `find_exp` | `arr \| find_exp: itemName, exp` | Returns first item via expression |
| `find_index` | `arr \| find_index: property[, expected]` | Returns index of first matching item (or `undefined`) |
| `find_index_exp` | `arr \| find_index_exp: itemName, exp` | Returns index via expression |
| `uniq` | `arr \| uniq` | Removes duplicates (`new Set`) |
| `sample` | `v \| sample[: count]` | Random sample; `count=1` returns single item; `count>1` returns array; accepts string or array |

### Base64

| Filter | Signature | Notes |
|--------|-----------|-------|
| `base64_encode` | `v \| base64_encode` | Base64 encode; stringifies input first |
| `base64_decode` | `v \| base64_decode` | Base64 decode; stringifies input first |

### Misc

| Filter | Signature | Notes |
|--------|-----------|-------|
| `default` | `v \| default: defaultValue[, allow_false: true]` | Returns `defaultValue` if `v` is falsy or empty (empty string/array, object with no keys); `allow_false: true` lets `false` pass through |
| `json` | `v \| json[: space]` | `JSON.stringify(v, null, space)`; `space` default `0` |
| `jsonify` | `v \| jsonify[: space]` | Alias for `json` |
| `inspect` | `v \| inspect[: space]` | `JSON.stringify` with circular reference protection (`'[Circular]'`) |
| `to_integer` | `v \| to_integer` | `Number(v)` |
| `raw` | `v \| raw` | Passes value without evaluation (`{ raw: true, handler: identify }`) |

---

## Drops (Special Objects)

### Drop (abstract base class)

| Method | Notes |
|--------|-------|
| `liquidMethodMissing(key, context)` | Called when property not found; returns `undefined` by default; can be overridden |

### ForloopDrop (created in `for`)

| Property/Method | Type | Notes |
|----------------|------|-------|
| `length` | number | Total iterations |
| `name` | string | `"variable-collection"` |
| `index` | number | 1-based current index |
| `index0` | number | 0-based current index |
| `rindex` | number | Remaining iterations including current |
| `rindex0` | number | Remaining iterations excluding current |
| `first` | boolean | `i === 0` |
| `last` | boolean | `i === length - 1` |

### TablerowloopDrop (created in `tablerow`, extends ForloopDrop)

Inherits all ForloopDrop properties, adds:

| Property/Method | Type | Notes |
|----------------|------|-------|
| `row` | number | Current row (1-based): `Math.floor(i / cols) + 1` |
| `col` | number | Current column (1-based): `col0 + 1` |
| `col0` | number | Current column (0-based): `i % cols` |
| `col_first` | boolean | `col0 === 0` |
| `col_last` | boolean | `col === cols` |

### EmptyDrop (`empty` literal)

| Behavior | Notes |
|----------|-------|
| `== empty` | another EmptyDrop → `false`; string/array → `.length === 0`; object → `Object.keys().length === 0` |
| `valueOf()` | `''` |
| Comparisons `>`, `<`, `>=`, `<=` | always `false` |

### BlankDrop (`blank` literal, extends EmptyDrop)

| Behavior | Notes |
|----------|-------|
| `== blank` | `false` → `true`; nil → `true`; whitespace-only string → `true`; inherits EmptyDrop logic for others |

### NullDrop (`null`/`nil` literal)

| Behavior | Notes |
|----------|-------|
| `== null` | equals any nil value |
| `valueOf()` | `null` |
| Comparisons | always `false` |

### BlockDrop (available as `block` inside `{% block %}`)

| Property | Notes |
|----------|-------|
| `block.super` | Renders the parent block content (from layout) |

---

## Operators

| Operator | Type | Behavior |
|----------|------|----------|
| `==` | binary | Equality; respects `Comparable` interface; arrays compare element by element |
| `!=` | binary | Inequality (`!equals`) |
| `>` | binary | Greater than; respects `Comparable.gt/lt` |
| `<` | binary | Less than; respects `Comparable.lt/gt` |
| `>=` | binary | Greater than or equal; respects `Comparable.geq/leq` |
| `<=` | binary | Less than or equal; respects `Comparable.leq/geq` |
| `contains` | binary | String: `indexOf > -1`; array: `some(equals)`; others with `indexOf`: `indexOf > -1` |
| `not` | unary | `isFalsy(v, ctx)` |
| `and` | binary | `isTruthy(l) && isTruthy(r)` |
| `or` | binary | `isTruthy(l) \|\| isTruthy(r)` |

Operators are customizable via `operators` option.

---

## Truthiness

| Mode | Falsy | Truthy |
|------|-------|--------|
| default (Liquid) | `false`, `undefined`, `null` | everything else (including `0`, `""`, `[]`) |
| `jsTruthy: true` | any JavaScript falsy (`0`, `""`, `null`, `undefined`, `false`, `NaN`) | everything else |

---

## Whitespace Control

| Delimiter | Behavior |
|-----------|----------|
| `{%-` / `-%}` | Trim whitespace before/after tags |
| `{{-` / `-}}` | Trim whitespace before/after outputs |
| `trimTagRight: true` | Global: trim right of `{% %}` up to `\n` (inclusive) |
| `trimTagLeft: true` | Global: trim left of `{% %}` |
| `trimOutputRight: true` | Global: trim right of `{{ }}` up to `\n` (inclusive) |
| `trimOutputLeft: true` | Global: trim left of `{{ }}` |
| `greedy: true` | (default) trim consumes all consecutive spaces/`\n` |

---

## Layout Inheritance

| Mechanism | Notes |
|-----------|-------|
| `{% layout 'file' %}` | Child template declares parent layout; child tokens are processed in `blockMode: STORE` |
| `{% layout none %}` | Disables layout; renders `blockMode: OUTPUT` directly |
| `{% block name %}...{% endblock %}` | Defines overridable block; in `STORE` mode stores render function |
| `{{ block.super }}` | Accesses parent block content via `BlockDrop.super()` |
| Anonymous block (`blocks['']`) | Child content outside `{% block %}` becomes anonymous block; available in parent as `{{ content_for_layout }}` equivalent |

---

## Options (LiquidOptions)

### Filesystem / templates

| Option | Default | Notes |
|--------|---------|-------|
| `root` | `['.']` | Base directory(ies) for templates |
| `partials` | `root` | Directory(ies) for partials (`include`, `render`) |
| `layouts` | `root` | Directory(ies) for layouts |
| `relativeReference` | `true` | Allows relative references (path must be inside root) |
| `extname` | `''` | Extension added to lookup if filepath has no extension |
| `cache` | `false` | Template cache; `true`, number (LRU size), or `LiquidCache` object |
| `fs` | node fs | Custom filesystem implementation |
| `templates` | — | In-memory template map; ignores fs and root when set |
| `dynamicPartials` | `true` | Treats filename as Liquid expression; `false` treats as literal |

### Scope / variable behavior

| Option | Default | Notes |
|--------|---------|-------|
| `globals` | `{}` | Global scope passed to all templates (including partials/layouts) |
| `strictVariables` | `false` | Throws error on undefined variable |
| `strictFilters` | `false` | Throws error on undefined filter; if `false`, undefined filter is skipped |
| `ownPropertyOnly` | `true` | Ignores inherited prototype properties |
| `lenientIf` | `false` | With `strictVariables: true`, allows undefined variable in `if`/`elsif`/`unless`/`default` without error |
| `catchAllErrors` | `false` | Collects all errors instead of stopping at first |
| `jsTruthy` | `false` | Uses JavaScript truthiness instead of Liquid |
| `jekyllInclude` | `false` | `include` injects variables into `include` object in scope |
| `jekyllWhere` | `false` | `where` also does array membership check |

### Output

| Option | Default | Notes |
|--------|---------|-------|
| `outputEscape` | `undefined` | Default escape applied to outputs: `'escape'`, `'json'`, or function |
| `keepOutputType` | `false` | Preserves output type (does not convert to string) |

### Delimiters

| Option | Default | Notes |
|--------|---------|-------|
| `tagDelimiterLeft` | `'{%'` | Left tag delimiter |
| `tagDelimiterRight` | `'%}'` | Right tag delimiter |
| `outputDelimiterLeft` | `'{{'` | Left output delimiter |
| `outputDelimiterRight` | `'}}'` | Right output delimiter |
| `keyValueSeparator` | `':'` | Key/value separator in hash arguments |
