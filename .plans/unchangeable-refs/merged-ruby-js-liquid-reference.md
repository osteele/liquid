# Liquid — Unified Reference (Ruby + LiquidJS)

> Extracted from `.example-repositories/liquid-ruby/liquid` and `.example-repositories/liquid-js/liquidjs`.
> **Legend:** no marker = present in both | **[Ruby]** = Ruby Liquid exclusive | **[JS]** = LiquidJS exclusive

### Code location conventions

```
Ruby:  lib/liquid/tags/<file>.rb
       lib/liquid/<file>.rb
JS:    src/tags/<file>.ts
       src/filters/<file>.ts
       src/drop/<file>.ts
       src/render/<file>.ts
```

---

## Tags

### Output / Expression tags

| Tag | Syntax | Differences | Ruby | JS |
|-----|--------|-------------|------|----|
| `{{ }}` | `{{ expression }}` | — | `lib/liquid/variable.rb` · `Variable#render_to_output_buffer(L112)` | `src/template/output.ts` via `Value#value()` |
| `echo` | `{% echo expression %}` | JS: value is optional (no value emits nothing) | `lib/liquid/tags/echo.rb` · `Echo#render(L30)` | `src/tags/echo.ts` · `render(L15)` |
| `liquid` | `{% liquid tag1\ntag2 %}` | — | `lib/liquid/tags/` — `liquid` tag via block (no separate file; read from `block_body.rb`) | `src/tags/liquid.ts` · `render(L11)` · parse via `readLiquidTagTokens` |
| `#` inline comment | `{%# comment %}` | Both: multi-line requires `#` on each line | `lib/liquid/tags/inline_comment.rb` · `InlineComment#render_to_output_buffer(L19)` | `src/tags/inline-comment.ts` · `render(L10)` |
| `raw` | `{% raw %}...{% endraw %}` | — | `lib/liquid/tags/raw.rb` · `Raw#render_to_output_buffer(L37)` | `src/tags/raw.ts` · `render(L15)` |
| `comment` | `{% comment %}...{% endcomment %}` | Ruby: supports nesting `comment`/`raw`; JS: does not | `lib/liquid/tags/comment.rb` · `Comment#render_to_output_buffer(L19)` · `parse_raw_tag_body(L78)` | `src/tags/comment.ts` · `render(L13)` (empty) |

### Variable / State tags

| Tag | Syntax | Differences | Ruby | JS |
|-----|--------|-------------|------|----|
| `assign` | `{% assign var = value %}` | Ruby: tracked by resource limits (`assign_score_of`); JS: stores in `ctx.bottom()` | `lib/liquid/tags/assign.rb` · `Assign#render_to_output_buffer(L43)` · `assign_score_of(L54)` | `src/tags/assign.ts` · `render(L22)` |
| `capture` | `{% capture var %}...{% endcapture %}` | Ruby: tracked by resource limits; JS: name can be quoted string | `lib/liquid/tags/capture.rb` · `Capture#render_to_output_buffer(L33)` | `src/tags/capture.ts` · `render(L33)` |
| `increment` | `{% increment var %}` | Both: stored in `environments`; outputs 0, 1, 2…; shares slot with `decrement` | `lib/liquid/tags/increment.rb` · `Increment#render_to_output_buffer(L33)` | `src/tags/increment.ts` · `render(L13)` |
| `decrement` | `{% decrement var %}` | Both: outputs -1, -2, …; Ruby: output-then-decrement; JS: pre-decrement-then-output (same result) | `lib/liquid/tags/decrement.rb` · `Decrement#render_to_output_buffer(L33)` | `src/tags/decrement.ts` · `render(L13)` |

### Conditional tags

| Tag | Sub-tags | Differences | Ruby | JS |
|-----|----------|-------------|------|----|
| `if` | `elsif`, `else`, `endif` | Ruby: accepts `<>` as alias for `!=`; JS: accepts unary `not` operator; JS: `elsif` after `else` is explicit error | `lib/liquid/tags/if.rb` · `If#render_to_output_buffer(L50)` · parsing: `strict2_parse(L63)` / `strict_parse(L96)` / `lax_parse(L81)` · conditions in `lib/liquid/condition.rb` · `Condition#evaluate(L68)` | `src/tags/if.ts` · `render(L38)` |
| `unless` | `elsif`, `else`, `endunless` | JS: `elsif` inside `unless` uses `isTruthy` (not inverted); Ruby: behavior equivalent to `if` | `lib/liquid/tags/unless.rb` · `Unless < If` · `render_to_output_buffer(L23)` | `src/tags/unless.ts` · `render(L43)` |
| `case` | `when`, `else`, `endcase` | Both: `when` supports multiple values separated by `or` or `,`; branches after `else` are ignored | `lib/liquid/tags/case.rb` · `Case#render_to_output_buffer(L67)` · `record_when_condition(L106)` · `parse_strict2_when(L112)` / `parse_lax_when(L127)` | `src/tags/case.ts` · `render(L57)` |

**[Ruby] only:**

| Tag | Sub-tags | Notes | Ruby |
|-----|----------|-------|------|
| `ifchanged` | — | Renders only if output changed since last iteration; state in `registers[:ifchanged]` | `lib/liquid/tags/ifchanged.rb` · `Ifchanged#render_to_output_buffer(L5)` |

### Iteration tags

| Tag | Options | Differences | Ruby | JS |
|-----|---------|-------------|------|----|
| `for` | `offset: n`, `limit: n`, `reversed`, range `(a..b)` | Ruby: modifiers applied in declared order; JS: order `offset→limit→reversed` by default (or declared order if `orderedFilterParameters: true`); JS: `offset: continue` to resume from previous point | `lib/liquid/tags/for.rb` · `For#render_to_output_buffer(L62)` · `render_segment(L149)` · `set_attribute(L176)` · `collection_segment(L114)` · `ParseTreeVisitor` at end of file | `src/tags/for.ts` · `render(L46)` · modifiers: `offset(L108)` / `limit(L112)` / `reversed(L115)` · `blockScope(L104)` |
| `break` | — | — | `lib/liquid/tags/break.rb` · `Break#render_to_output_buffer(L23)` · `INTERRUPT = BreakInterrupt` (L21) | `src/tags/break.ts` · `render(L4)` · sets `ctx.breakCalled = true` |
| `continue` | — | — | `lib/liquid/tags/continue.rb` · `Continue#render_to_output_buffer(L16)` · `INTERRUPT = ContinueInterrupt` (L14) | `src/tags/continue.ts` · `render(L4)` · sets `ctx.continueCalled = true` |
| `cycle` | `[group:] v1, v2, ...` | Both: state in `cycle` register; key: `"cycle:{group}:{candidates}"` | `lib/liquid/tags/cycle.rb` · `Cycle#render_to_output_buffer(L33)` · `named?(L29)` · `strict2_parse(L53)` / `strict_parse(L95)` / `lax_parse(L99)` · `variables_from_string(L117)` | `src/tags/cycle.ts` · `render(L30)` · register key: `"cycle:${group}:${candidates}"` |
| `tablerow` | `cols: n`, `offset: n`, `limit: n`, range `(a..b)` | Both: generates `<tr class="rowN">`, `<td class="colN">`; Ruby: supports `break` inside tablerow; JS: no `break` support in tablerow | `lib/liquid/tags/table_row.rb` · `TableRow#render_to_output_buffer(L81)` · `strict2_parse(L37)` / `strict_parse(L62)` / `lax_parse(L66)` | `src/tags/tablerow.ts` · `render(L44)` · generates `tr`/`td` inline in render · `blockScope(L92)` |

### Template inclusion tags

| Tag | Syntax | Scope | Differences | Ruby | JS |
|-----|--------|-------|-------------|------|----|
| `include` | `{% include 'file' [with var] [key: val ...] %}` | Shared | Ruby: **deprecated**; additionally supports `for array as alias`; JS: does not support `for`; JS: `jekyllInclude` option injects vars into `include` object | `lib/liquid/tags/include.rb` · `Include#render_to_output_buffer(L36)` · `prepend Tag::Disableable (L22)` · `strict2_parse(L88)` / `strict_parse(L107)` / `lax_parse(L111)` | `src/tags/include.ts` · `render(L31)` · `partialScope(L55)` |
| `render` | `{% render 'file' [with var [as alias]] [for collection [as alias]] [key: val ...] %}` | Isolated | Ruby: isolated scope except globals; JS: uses `ctx.spawn()`; same semantics | `lib/liquid/tags/render.rb` · `Render#render_to_output_buffer(L41)` · `render_tag(L45)` · `disable_tags "include" (L28)` · `for_loop?(L37)` · `strict2_parse(L85)` / `strict_parse(L111)` / `lax_parse(L115)` | `src/tags/render.ts` · `render(L49)` · `parseFilePath(L125)` / `renderFilePath(L141)` · `partialScope(L85)` |

**[Ruby] only:**

| Tag | Notes | Ruby |
|-----|-------|------|
| `doc` | `{% doc %}...{% enddoc %}` — LiquidDoc; ignored by renderer | `lib/liquid/tags/doc.rb` · `Doc#render_to_output_buffer(L58)` · `blank?(L62)` · `raise_nested_doc_error(L78)` |

**[JS] only:**

| Tag | Syntax | Notes | JS |
|-----|--------|-------|----|
| `layout` | `{% layout 'file' [key: val ...] %}` | Layout inheritance; consumes remaining tokens as child templates; `{% layout none %}` disables and renders directly | `src/tags/layout.ts` · `render(L21)` · `blockMode: STORE` → `STORE` → `OUTPUT` · `children(L52)` · `partialScope(L72)` |
| `block` | `{% block name %}...{% endblock %}` | Defines overridable block; `{{ block.super }}` accesses parent content | `src/tags/block.ts` · `render(L23)` · `getBlockRender` internal · `BlockDrop` in `src/drop/block-drop.ts` |

---

## Filters

> **Ruby source:** all in `lib/liquid/standardfilters.rb` · module `Liquid::StandardFilters`
> **JS source:** distributed across `src/filters/` · registered in `src/filters/index.ts`

### String — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `downcase` | `v \| downcase` | — | `standardfilters.rb:L76` | `string.ts · downcase(L46)` |
| `upcase` | `v \| upcase` | — | `standardfilters.rb:L85` | `string.ts · upcase(L52)` |
| `capitalize` | `v \| capitalize` | Uppercase first letter, rest lowercase | `standardfilters.rb:L94` | `string.ts · capitalize(L121)` |
| `append` | `v \| append: arg` | — | `standardfilters.rb:L684` | `string.ts · append(L20)` |
| `prepend` | `v \| prepend: arg` | — | `standardfilters.rb:L713` | `string.ts · prepend(L27)` |
| `remove` | `v \| remove: arg` | — | `standardfilters.rb:L654` | `string.ts · remove(L58)` |
| `remove_first` | `v \| remove_first: arg` | — | `standardfilters.rb:L664` | `string.ts · remove_first(L64)` |
| `remove_last` | `v \| remove_last: arg` | — | `standardfilters.rb:L674` | `string.ts · remove_last(L71)` |
| `replace` | `v \| replace: pattern, repl` | — | `standardfilters.rb:L606` | `string.ts · replace(L127)` |
| `replace_first` | `v \| replace_first: p, r` | — | `standardfilters.rb:L620` | `string.ts · replace_first(L135)` |
| `replace_last` | `v \| replace_last: p, r` | — | `standardfilters.rb:L633` | `string.ts · replace_last(L142)` |
| `split` | `v \| split: sep` | JS: trailing empty strings removed (Ruby behavior) | `standardfilters.rb:L268` | `string.ts · split(L91)` |
| `lstrip` | `v \| lstrip` | JS: accepts optional `chars` argument (strip character set) | `standardfilters.rb:L300` | `string.ts · lstrip(L34)` |
| `rstrip` | `v \| rstrip` | JS: accepts optional `chars` argument | `standardfilters.rb:L309` | `string.ts · rstrip(L79)` |
| `strip` | `v \| strip` | JS: accepts optional `chars` argument | `standardfilters.rb:L291` | `string.ts · strip(L102)` |
| `strip_html` | `v \| strip_html` | Removes `<script>`, `<style>`, HTML tags, HTML comments | `standardfilters.rb:L318` | `html.ts · strip_html(L43)` |
| `strip_newlines` | `v \| strip_newlines` | Removes `\r?\n` | `standardfilters.rb:L329` | `string.ts · strip_newlines(L115)` |
| `newline_to_br` | `v \| newline_to_br` | Replaces `\r?\n` with `<br />\n` | `standardfilters.rb:L723` | `html.ts · newline_to_br(L37)` |
| `truncate` | `v \| truncate[: n[, ellipsis]]` | Default: 50 chars, `'...'`; suffix included in count | `standardfilters.rb:L218` | `string.ts · truncate(L151)` |
| `truncatewords` | `v \| truncatewords[: n[, ellipsis]]` | Default: 15 words, `'...'`; JS: `words <= 0` uses 1 | `standardfilters.rb:L241` | `string.ts · truncatewords(L159)` |
| `size` | `v \| size` | Works on strings and arrays | `standardfilters.rb:L65` | `array.ts · size(L49)` |
| `slice` | `v \| slice: begin[, length]` | Works on strings and arrays; `begin < 0` → from end; default length=1 | `standardfilters.rb:L197` | `array.ts · slice(L100)` |

### String — **[Ruby]** only

| Filter | Signature | Notes | Ruby (line) |
|--------|-----------|-------|-------------|
| `squish` | `v \| squish` | Strip + collapse internal whitespace to single space | `standardfilters.rb:L280` |
| `h` | `v \| h` | Alias for `escape` | `standardfilters.rb:L103` (defined alongside `escape`) |

### String — **[JS]** only

| Filter | Signature | Notes | JS (file · line) |
|--------|-----------|-------|------------------|
| `normalize_whitespace` | `v \| normalize_whitespace` | Functional equivalent of Ruby's `squish` | `string.ts · normalize_whitespace(L168)` |
| `number_of_words` | `v \| number_of_words[: mode]` | Counts words; modes: `'cjk'` (CJK + non-CJK), `'auto'` (CJK if present, else split), default (split) | `string.ts · number_of_words(L174)` |
| `array_to_sentence_string` | `arr \| array_to_sentence_string[: connector]` | `"a, b, and c"`; default connector `'and'`; 0 items → `''`, 1 item → `a`, 2 → `a and b` | `string.ts · array_to_sentence_string(L210)` |
| `xml_escape` | `v \| xml_escape` | Alias for `escape` | `html.ts · xml_escape(L24)` |

### HTML — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `escape` | `v \| escape` | HTML escape `&`, `<`, `>`, `"`, `'`; Ruby: alias `h`; JS: alias `xml_escape` | `standardfilters.rb:L103` | `html.ts · escape(L18)` |
| `escape_once` | `v \| escape_once` | Unescape then escape (does not double-escape already escaped entities) | `standardfilters.rb:L114` | `html.ts · escape_once(L33)` |

### URL — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `url_encode` | `v \| url_encode` | `encodeURIComponent`; spaces become `+` | `standardfilters.rb:L126` | `url.ts · url_encode(L4)` |
| `url_decode` | `v \| url_decode` | `decodeURIComponent`; `+` becomes space | `standardfilters.rb:L137` | `url.ts · url_decode(L3)` |

### URL — **[JS]** only

| Filter | Signature | Notes | JS (file · line) |
|--------|-----------|-------|------------------|
| `cgi_escape` | `v \| cgi_escape` | `encodeURIComponent` with `+` and uppercase hex for `!'()*` | `url.ts · cgi_escape(L5)` |
| `uri_escape` | `v \| uri_escape` | `encodeURI` preserving `[` and `]` | `url.ts · uri_escape(L8)` |
| `slugify` | `v \| slugify[: mode[, cased]]` | Modes: `'raw'`, `'default'`, `'pretty'`, `'ascii'`, `'latin'`, `'none'`; `cased=false` by default | `url.ts · slugify(L22)` |

### Base64 — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `base64_encode` | `v \| base64_encode` | — | `standardfilters.rb:L150` | `base64.ts · base64_encode(L11)` |
| `base64_decode` | `v \| base64_decode` | Ruby: throws error if input is invalid | `standardfilters.rb:L160` | `base64.ts · base64_decode(L17)` |

### Base64 — **[Ruby]** only

| Filter | Signature | Notes | Ruby (line) |
|--------|-----------|-------|-------------|
| `base64_url_safe_encode` | `v \| base64_url_safe_encode` | URL-safe Base64 | `standardfilters.rb:L172` |
| `base64_url_safe_decode` | `v \| base64_url_safe_decode` | URL-safe Base64 decode | `standardfilters.rb:L182` |

### Math — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `abs` | `v \| abs` | — | `standardfilters.rb:L803` | `math.ts · abs(L3)` |
| `plus` | `v \| plus: n` | — | `standardfilters.rb:L813` | `math.ts · plus(L10)` |
| `minus` | `v \| minus: n` | — | `standardfilters.rb:L823` | `math.ts · minus(L9)` |
| `times` | `v \| times: n` | — | `standardfilters.rb:L833` | `math.ts · times(L12)` |
| `divided_by` | `v \| divided_by: n` | Ruby: result type = divisor type; raises `ZeroDivisionError` with zero; JS: accepts `integerArithmetic` as 2nd arg (`Math.floor`) | `standardfilters.rb:L843` | `math.ts · divided_by(L7)` |
| `modulo` | `v \| modulo: n` | Ruby: raises `ZeroDivisionError` with zero | `standardfilters.rb:L855` | `math.ts · modulo(L11)` |
| `round` | `v \| round[: decimals]` | Default 0 decimal places | `standardfilters.rb:L866` | `math.ts · round(L14)` |
| `ceil` | `v \| ceil` | — | `standardfilters.rb:L906` | `math.ts · ceil(L6)` |
| `floor` | `v \| floor` | — | `standardfilters.rb:L920` | `math.ts · floor(L8)` |
| `at_least` | `v \| at_least: n` | `max(v, n)` | `standardfilters.rb:L933` | `math.ts · at_least(L4)` |
| `at_most` | `v \| at_most: n` | `min(v, n)` | `standardfilters.rb:L948` | `math.ts · at_most(L5)` |

### Date — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `date` | `v \| date: format` | Ruby: returns input if empty/invalid; JS: format optional (uses `opts.dateFormat`); JS: accepts `timezoneOffset` as 2nd arg; JS: `'now'`/`'today'` → current time; JS: `opts.preserveTimezones` | `standardfilters.rb:L770` | `date.ts · date(L5)` |

### Date — **[JS]** only

| Filter | Signature | Notes | JS (file · line) |
|--------|-----------|-------|------------------|
| `date_to_xmlschema` | `v \| date_to_xmlschema` | Format `'%Y-%m-%dT%H:%M:%S%:z'` | `date.ts · date_to_xmlschema(L15)` |
| `date_to_rfc822` | `v \| date_to_rfc822` | Format `'%a, %d %b %Y %H:%M:%S %z'` | `date.ts · date_to_rfc822(L19)` |
| `date_to_string` | `v \| date_to_string[: type[, style]]` | Abbreviated month (`%b`); `type='ordinal'`, `style='US'` uses US format | `date.ts · date_to_string(L23)` |
| `date_to_long_string` | `v \| date_to_long_string[: type[, style]]` | Full month (`%B`) | `date.ts · date_to_long_string(L27)` |

### Array — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `join` | `arr \| join[: sep]` | Default sep `' '` | `standardfilters.rb:L339` | `array.ts · join(L8)` |
| `first` | `arr \| first` | — | `standardfilters.rb:L782` | `array.ts · first(L16)` |
| `last` | `arr \| last` | — | `standardfilters.rb:L793` | `array.ts · last(L15)` |
| `reverse` | `arr \| reverse` | — | `standardfilters.rb:L466` | `array.ts · reverse(L17)` |
| `sort` | `arr \| sort[: property]` | Ruby: nil-safe (nils go to end); JS: `property` can be dot-notation path | `standardfilters.rb:L349` | `array.ts · sort(L23)` |
| `sort_natural` | `arr \| sort_natural[: property]` | Case-insensitive | `standardfilters.rb:L371` | `array.ts · sort_natural(L40)` |
| `map` | `arr \| map: property` | — | `standardfilters.rb:L476` | `array.ts · map(L51)` |
| `sum` | `arr \| sum[: property]` | JS: NaN → 0 | `standardfilters.rb:L982` | `array.ts · sum(L59)` |
| `compact` | `arr \| compact` | Ruby: accepts optional `property` argument; JS: does not | `standardfilters.rb:L496` | `array.ts · compact(L68)` |
| `uniq` | `arr \| uniq` | Ruby: accepts optional `property` argument; JS: uses `new Set` without property | `standardfilters.rb:L446` | `array.ts · uniq(L252)` |
| `concat` | `arr \| concat: arr2` | — | `standardfilters.rb:L699` | `array.ts · concat(L73)` |
| `where` | `arr \| where: prop[, expected]` | Without `expected` → filters truthy; JS: `jekyllWhere` option adds array membership check | `standardfilters.rb:L393` | `array.ts · where(L131)` · `filter(L117)` internal |
| `reject` | `arr \| reject: prop[, expected]` | Inverse of `where` | `standardfilters.rb:L404` | `array.ts · reject(L135)` |
| `find` | `arr \| find: prop[, expected]` | — | `standardfilters.rb:L425` | `array.ts · find(L245)` · `search(L200)` internal |
| `find_index` | `arr \| find_index: prop[, expected]` | — | `standardfilters.rb:L436` | `array.ts · find_index(L235)` |
| `has` | `arr \| has: prop[, expected]` | Returns boolean | `standardfilters.rb:L414` | `array.ts · has(L225)` |

### Array — **[JS]** only

| Filter | Signature | Notes | JS (file · line) |
|--------|-----------|-------|------------------|
| `where_exp` | `arr \| where_exp: itemName, exp` | Filters by Liquid expression; `exp` evaluated with `itemName` in scope | `array.ts · where_exp(L139)` · `filter_exp(L107)` internal |
| `reject_exp` | `arr \| reject_exp: itemName, exp` | Inverse of `where_exp` | `array.ts · reject_exp(L143)` |
| `group_by` | `arr \| group_by: property` | Returns `[{name, items}, ...]` | `array.ts · group_by(L147)` |
| `group_by_exp` | `arr \| group_by_exp: itemName, exp` | Groups by Liquid expression | `array.ts · group_by_exp(L160)` |
| `has_exp` | `arr \| has_exp: itemName, exp` | Boolean via expression | `array.ts · has_exp(L230)` · `search_exp(L209)` internal |
| `find_exp` | `arr \| find_exp: itemName, exp` | First item via expression | `array.ts · find_exp(L250)` |
| `find_index_exp` | `arr \| find_index_exp: itemName, exp` | Index via expression | `array.ts · find_index_exp(L240)` |
| `push` | `arr \| push: item` | Adds to end (returns new array) | `array.ts · push(L79)` |
| `pop` | `arr \| pop` | Removes from end (returns new array) | `array.ts · pop(L89)` |
| `unshift` | `arr \| unshift: item` | Adds to beginning (returns new array) | `array.ts · unshift(L83)` |
| `shift` | `arr \| shift` | Removes from beginning (returns new array) | `array.ts · shift(L94)` |
| `sample` | `v \| sample[: count]` | Random sample; `count=1` returns single item; accepts string or array | `array.ts · sample(L258)` |

### Misc — Common

| Filter | Signature | Differences | Ruby (line) | JS (file · line) |
|--------|-----------|-------------|-------------|------------------|
| `default` | `v \| default: val[, allow_false: true]` | Both support `allow_false` as named/keyword arg; returns `val` if `v` is falsy or empty (empty string/array, object with no keys) | `standardfilters.rb:L969` | `misc.ts · default(L5)` (exported as `default` key in exported object `L39`) |

### Misc — **[JS]** only

| Filter | Signature | Notes | JS (file · line) |
|--------|-----------|-------|------------------|
| `json` | `v \| json[: space]` | `JSON.stringify(v, null, space)`; default space 0 | `misc.ts · json(L13)` |
| `jsonify` | `v \| jsonify[: space]` | Alias for `json` | `misc.ts` (same export `json`, alias in final object `L39`) |
| `inspect` | `v \| inspect[: space]` | `JSON.stringify` with circular reference protection (`'[Circular]'`) | `misc.ts · inspect(L18)` |
| `to_integer` | `v \| to_integer` | `Number(v)` | `misc.ts · to_integer(L31)` |
| `raw` | `v \| raw` | Passes value without escaping/evaluation | `misc.ts · raw(L36)` (object `{ raw: true, handler: identify }`) |

---

## Drops (Special Objects)

### Drop — Base

| Feature | Ruby | JS |
|---------|------|----|
| File | `lib/liquid/drop.rb` | `src/drop/drop.ts` |
| Catch-all | `liquid_method_missing(name)` · L33 | `liquidMethodMissing(key, context)` · L4 |
| When not found | `UndefinedDropMethod` if strict | returns `undefined` |
| Invocation | `invoke_drop(key)` L39; `[key]` alias | direct access via `readProperty` in `context/context.ts:L112` |
| Method whitelist | `invokable_methods` L68 (public except blacklist) | any public method of the class |
| Context injection | `attr_writer :context` L26 | passed as arg in `liquidMethodMissing` |

### ForloopDrop — Common

| Property | Type | Notes | Ruby (line) | JS (line) |
|----------|------|-------|-------------|-----------|
| `length` | number | Total iterations | `forloop_drop.rb:L20` (attr) | `forloop-drop.ts:L8` (field) |
| `name` | string | `"variable-collection"` | `forloop_drop.rb:L34` (attr) | `forloop-drop.ts:L9` (field) |
| `index` | number | 1-based | `forloop_drop.rb:L40` | `forloop-drop.ts:L18` |
| `index0` | number | 0-based | `forloop_drop.rb:L48` | `forloop-drop.ts:L15` |
| `rindex` | number | Remaining including current (1-based reverse) | `forloop_drop.rb:L55` | `forloop-drop.ts:L27` |
| `rindex0` | number | Remaining excluding current (0-based reverse) | `forloop_drop.rb:L62` | `forloop-drop.ts:L30` |
| `first` | boolean | `i === 0` | `forloop_drop.rb:L69` | `forloop-drop.ts:L21` |
| `last` | boolean | `i === length - 1` | `forloop_drop.rb:L76` | `forloop-drop.ts:L24` |
| `increment!` / `next()` | — | Advances counter | `forloop_drop.rb:L83` | `forloop-drop.ts:L12` (`next` method) |

> **[Ruby] only:** `parentloop` L28 — reference to the parent ForloopDrop (or nil)

### TablerowloopDrop — Common (extends ForloopDrop)

| Property | Type | Notes | Ruby (line) | JS (line) |
|----------|------|-------|-------------|-----------|
| `row` | number | Current row (1-based) | `tablerowloop_drop.rb:L33` | `tablerowloop-drop.ts:L10` |
| `col` | number | Current column (1-based) | `tablerowloop_drop.rb:L27` | `tablerowloop-drop.ts:L16` |
| `col0` | number | Current column (0-based) | `tablerowloop_drop.rb:L52` | `tablerowloop-drop.ts:L13` |
| `col_first` | boolean | `col0 === 0` | `tablerowloop_drop.rb:L87` | `tablerowloop-drop.ts:L19` |
| `col_last` | boolean | `col === cols` | `tablerowloop_drop.rb:L103` | `tablerowloop-drop.ts:L22` |

### EmptyDrop (`empty`) — Common

| Behavior | Notes | Ruby | JS |
|----------|-------|------|----|
| `== empty` | string/array → `.length === 0`; object → no keys; another EmptyDrop → `false` | `condition.rb:L156` · `liquid_empty?` | `src/drop/empty-drop.ts · EmptyDrop#equals(L6)` |
| `valueOf()` | `''` | Ruby returns `''` via `to_s` | `empty-drop.ts:L24` |
| Comparisons `>`, `<`, `>=`, `<=` | always `false` | Ruby: `Comparable` protocol; raises on compare | `empty-drop.ts: gt(L12), geq(L15), lt(L18), leq(L21)` |

### BlankDrop (`blank`) — Common

| Behavior | Notes | Ruby | JS |
|----------|-------|------|----|
| `== blank` | nil/null/undefined → `true`; `false` → `true`; whitespace-only string → `true` | `condition.rb:L135` · `liquid_blank?` | `src/drop/blank-drop.ts · BlankDrop#equals(L5)` |

### **[JS]** only Drops

| Drop | Available as | Notes | JS |
|------|-------------|-------|----|
| `NullDrop` | literal `null`/`nil` in comparison | Explicit `Comparable` interface; `valueOf → null`; comparisons always `false` | `src/drop/null-drop.ts` · `equals(L5)` |
| `BlockDrop` | `block` variable inside `{% block %}` | `block.super` renders parent block content | `src/drop/block-drop.ts` · `super()(L15)` · used in `src/tags/block.ts:L27` |

---

## Operators — Common

| Operator | Type | Behavior | Ruby | JS |
|----------|------|----------|------|----|
| `==` | binary | Equality; `Comparable`-aware; JS arrays compare element by element | `condition.rb:L13` (`@@operators` hash) | `src/render/operator.ts · equals(L42)` · `defaultOperators['=='](L13)` |
| `!=` | binary | Inequality | `condition.rb:L13` | `operator.ts · defaultOperators['!='](L14)` |
| `>`, `<`, `>=`, `<=` | binary | Comparison; `Comparable`-aware | `condition.rb:L13` | `operator.ts · defaultOperators(L15-30)` |
| `contains` | binary | String: `indexOf > -1`; Array: `include?` / `some(equals)` | `condition.rb:L13` | `operator.ts · defaultOperators['contains'](L31)` |
| `and` | binary | Short-circuit; no precedence (evaluates left to right) | `condition.rb` · parsed in `if.rb#parse_binary_comparisons(L103)` | `operator.ts · defaultOperators['and'](L38)` · `src/render/boolean.ts · isTruthy(L4)` |
| `or` | binary | Short-circuit | idem | `operator.ts · defaultOperators['or'](L39)` |

**[Ruby] only:** `<>` — alias for `!=` · `condition.rb:L13`

**[JS] only:** `not` (unary) · `operator.ts · defaultOperators['not'](L37)` · uses `src/render/boolean.ts · isFalsy(L8)` · operators customizable via `operators` option in `src/liquid-options.ts:L90`

---

## Truthiness

| Value | Ruby | JS default | JS `jsTruthy: true` |
|-------|------|------------|---------------------|
| `false` | falsy | falsy | falsy |
| `nil` / `null` / `undefined` | falsy | falsy | falsy |
| `0` | **truthy** | **truthy** | falsy |
| `""` | **truthy** | **truthy** | falsy |
| `[]` | **truthy** | **truthy** | falsy |
| Any other | truthy | truthy | truthy |

> Ruby: `lib/liquid/condition.rb` · `interpret_condition(L166)` (nil/false = falsy, rest truthy)
> JS: `src/render/boolean.ts` · `isFalsy(L8)` — with `jsTruthy: true` returns `!val`; without, only `false/undefined/null`
> JS `jsTruthy` configured in `src/liquid-options.ts:L43`

---

## Literals / Special Values

| Literal | Ruby | JS | Notes |
|---------|------|-----|-------|
| `nil` / `null` | ✓ (`nil`) | ✓ (`null`) | Ruby: `expression.rb · LITERALS(L5)` · JS: `NullDrop` in `src/drop/null-drop.ts` |
| `true` / `false` | ✓ | ✓ | Ruby: `expression.rb · LITERALS(L6-7)` |
| integer | ✓ | ✓ | Ruby: `expression.rb · INTEGER_REGEX(L24)` · JS: `src/tokens/` |
| float | ✓ | ✓ | Ruby: `expression.rb · FLOAT_REGEX(L25)` |
| string `"..."` / `'...'` | ✓ | ✓ | Ruby: `lexer.rb:L26,35` |
| range `(a..b)` | ✓ | ✓ | Ruby: `expression.rb · RANGES_REGEX(L23)` · Supported in `for` and `tablerow` |
| `empty` | ✓ | ✓ | Ruby: `expression.rb · LITERALS(L12)` · `condition.rb · liquid_empty?(L156)` · JS: `src/drop/empty-drop.ts` |
| `blank` | ✓ | ✓ | Ruby: `expression.rb · LITERALS(L13)` · `condition.rb · liquid_blank?(L135)` · JS: `src/drop/blank-drop.ts` |

---

## Variable Access

| Syntax | Both | Differences | Ruby | JS |
|--------|------|-------------|------|----|
| `variable` | ✓ | Lookup in stacked scopes | `context.rb · find_variable(L176)` | `context.ts · _get(L69)` · `findScope(L106)` |
| `obj.prop` | ✓ | — | `variable_lookup.rb` | `context.ts · _getFromScope(L79)` |
| `obj[key]` | ✓ | — | `variable_lookup.rb` | `context.ts · readProperty(L112)` |
| `array[0]` | ✓ | — | `variable_lookup.rb` | `context.ts · readProperty(L112)` |
| `array[-1]` | ✓ | JS: explicit `obj.length + key` | `variable_lookup.rb` | `context.ts:L115` (`key < 0` branch) |
| `array.first` | ✓ | Shortcut | `variable_lookup.rb` | `context.ts · readFirst(L130)` |
| `array.last` | ✓ | Shortcut | `variable_lookup.rb` | `context.ts · readLast(L135)` |
| `obj.size` | ✓ | `length` for arrays/strings | `standardfilters.rb · size(L65)` | `context.ts · readSize(L140) [key === 'size']` |
| `forloop.index` etc. | ✓ | — | `forloop_drop.rb` | `src/drop/forloop-drop.ts` |

---

## Whitespace Control

### Template markers — Common

| Marker | Effect | Ruby | JS |
|--------|--------|------|----|
| `{%-` | Removes whitespace to the left of the tag | `lib/liquid/lexer.rb` (tokenizer recognizes `-`) | `src/parser/` (trimLeft flag in token) |
| `-%}` | Removes whitespace to the right of the tag | idem | idem |
| `{{-` | Removes whitespace to the left of output | idem | idem |
| `-}}` | Removes whitespace to the right of output | idem | idem |

### Global options — **[JS]** only

| Option | Default | Notes | JS |
|--------|---------|-------|----|
| `trimTagRight` | `false` | Trim right of `{% %}` up to `\n` (inclusive) | `src/liquid-options.ts:L62` |
| `trimTagLeft` | `false` | Trim left of `{% %}` | `src/liquid-options.ts:L63` |
| `trimOutputRight` | `false` | Trim right of `{{ }}` up to `\n` (inclusive) | `src/liquid-options.ts:L65` |
| `trimOutputLeft` | `false` | Trim left of `{{ }}` | `src/liquid-options.ts:L66` |
| `greedy` | `true` | Trim consumes all consecutive spaces/`\n` | `src/liquid-options.ts:L76` |

---

## Layout Inheritance

### **[JS]** only

| Mechanism | Notes | JS |
|-----------|-------|----|
| `{% layout 'file' %}` | Declares parent layout; remaining child tokens processed in `blockMode: STORE` | `src/tags/layout.ts · render(L21)` · `BlockMode.STORE` from `src/context/block-mode.ts` |
| `{% layout none %}` | Disables layout; renders directly in `blockMode: OUTPUT` | `src/tags/layout.ts:L23` (checks `file === undefined`) |
| `{% block name %}...{% endblock %}` | Overridable block; in `STORE` mode stores render function | `src/tags/block.ts · render(L23)` · `getBlockRender` |
| `{{ block.super }}` | Accesses parent block content | `src/drop/block-drop.ts · super()(L15)` · injected in `block.ts:L27` |
| Anonymous block (`blocks['']`) | Child content outside `{% block %}` → anonymous block available in parent | `src/tags/layout.ts:L35` |

---

## Filter System

| Feature | Ruby | JS |
|---------|------|----|
| Central file | `lib/liquid/standardfilters.rb` | `src/filters/index.ts` (aggregates all) |
| Registration | `environment.rb · register_filter(L95)` · `strainer_template.rb` | `src/liquid.ts · registerFilter(L100)` |
| Positional filters | `variable.rb · lax_parse(L42)` / `strict_parse(L61)` | `src/template/value.ts` |
| Filters with keyword args | `variable.rb · parse_filterargs(L93)` — passed as hash | `misc.ts · default(L5)` — named args `[string, any][]` |
| Undefined filter | `lib/liquid/errors.rb · UndefinedFilter` · `strict_filters` in `environment.rb:L8` | `src/liquid.ts` — `strictFilters: true` → exception; `false` → silently skips |
| Global filter on all output | `context.rb · apply_global_filter(L75)` · `global_filter:` render option | — (not implemented in JS) |

---

## Configuration / Options

### Filesystem and templates — Common (concept)

| Feature | Ruby | JS |
|---------|------|----|
| File resolver | `environment.rb · file_system(L23)` — object with `read_template_file` · `lib/liquid/file_system.rb` | `src/liquid-options.ts · root/partials/layouts(L14-18)` |
| Base filesystem | `file_system.rb · BlankFileSystem(L17)` / `LocalFileSystem(L42)` | `src/fs/fs-impl.ts` (Node.js `fs`) |
| Custom filesystem | `environment.rb:L23` `file_system=` | `src/liquid-options.ts:L84` `fs:` option |
| In-memory templates | — | `src/liquid-options.ts:L87` `templates:` option · `src/fs/map-fs.ts` **[JS]** |
| Template cache | — | `src/liquid-options.ts:L36` `cache:` option · `src/cache/` **[JS]** |
| Implicit extension | via pattern in `LocalFileSystem#initialize(L45)` | `src/liquid-options.ts:L34` `extname` **[JS]** |
| Dynamic partials | implicit | `src/liquid-options.ts:L41` `dynamicPartials` (default `true`) **[JS]** |
| Relative references | — | `src/liquid-options.ts:L31` `relativeReference` (default `true`) **[JS]** |

### Scope and variables

| Feature | Ruby | JS |
|---------|------|----|
| Global variables | `globals:` option in `template.rb · apply_options_to_context(L206)` | `src/liquid-options.ts:L88` `globals: {}` · `context.ts · globals(L28)` |
| Strict variables | render option `strict_variables:` · `context.rb · find_variable(L176)` | `src/liquid-options.ts:L46` `strictVariables:` · `context.ts:L40` |
| Strict filters | render/env option `strict_filters:` · `strainer_template.rb` | `src/liquid-options.ts:L48` `strictFilters:` |
| Restrict properties | — | `src/liquid-options.ts:L50` `ownPropertyOnly:` (default `true`) · `context.ts:L42` |

### Behavior — **[JS]** only

| Option | Default | Notes | JS |
|--------|---------|-------|----|
| `jsTruthy` | `false` | Uses JavaScript truthiness | `src/liquid-options.ts:L43` · `render/boolean.ts:L11` |
| `lenientIf` | `false` | With `strictVariables: true`: allows undefined var in `if`/`elsif`/`unless`/`default` without error | `src/liquid-options.ts:L52` · used in `tags/if.ts:L41` via `ctx.opts.lenientIf` |
| `ownPropertyOnly` | `true` | Ignores inherited prototype properties | `src/liquid-options.ts:L50` · `context.ts · readJSProperty(L125)` |
| `keepOutputType` | `false` | Preserves output type (does not convert to string) | `src/liquid-options.ts:L83` |
| `outputEscape` | `undefined` | Default escape applied to all output: `'escape'`, `'json'`, or function | `src/liquid-options.ts:L80` · `src/liquid-options.ts · normalize(L202)` |
| `catchAllErrors` | `false` | Collects all errors without stopping at first | `src/liquid-options.ts:L54` |
| `jekyllInclude` | `false` | `include` injects variables into `include` object in scope | `src/liquid-options.ts:L33` · `tags/include.ts:L35` |
| `jekyllWhere` | `false` | `where` filter also does array membership check | `src/liquid-options.ts:L35` · `filters/array.ts · expectedMatcher(L108)` |

### Delimiters — **[JS]** only

| Option | Default | JS |
|--------|---------|-----|
| `tagDelimiterLeft` | `'{%'` | `src/liquid-options.ts:L68` |
| `tagDelimiterRight` | `'%}'` | `src/liquid-options.ts:L70` |
| `outputDelimiterLeft` | `'{{'` | `src/liquid-options.ts:L72` |
| `outputDelimiterRight` | `'}}'` | `src/liquid-options.ts:L74` |
| `keyValueSeparator` | `':'` | `src/liquid-options.ts` · used in `Hash` constructor |

### Date — **[JS]** only

| Option | Default | JS |
|--------|---------|-----|
| `timezoneOffset` | local | `src/liquid-options.ts:L57` |
| `dateFormat` | `'%A, %B %-e, %Y at %-l:%M %P %z'` | `src/liquid-options.ts:L59` · `filters/date.ts:L7` |
| `locale` | system | `src/liquid-options.ts:L61` |
| `preserveTimezones` | `false` | `src/liquid-options.ts:L78` · `filters/date.ts · parseDate` |

### Misc — **[JS]** only

| Option | Default | JS |
|--------|---------|-----|
| `orderedFilterParameters` | `false` | `src/liquid-options.ts` · `tags/for.ts:L57` (`modifiers` order logic) |
| `operators` | defaultOperators | `src/liquid-options.ts:L93` · `src/render/operator.ts · defaultOperators(L12)` |

---

## Resource Limits / DoS

### **[Ruby]** — Score-based · `lib/liquid/resource_limits.rb`

| Limit | Description | Ruby |
|-------|-------------|------|
| `render_length_limit` | Maximum bytes in a template's output | `resource_limits.rb:L5` attr · `increment_write_score(L40)` |
| `render_score_limit` | Render score per template (each node counts) | `resource_limits.rb:L6` attr · `increment_render_score(L27)` |
| `assign_score_limit` | Assign score (string → bytesize; array/hash → recursive sum) | `resource_limits.rb:L7` attr · `increment_assign_score(L33)` |
| `cumulative_render_score_limit` | Cumulative score across multiple renders | `resource_limits.rb:L8` attr |
