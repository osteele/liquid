# Liquid — Referência Unificada (Ruby + LiquidJS)

> Extraído de `.example-repositories/liquid-ruby/liquid` e `.example-repositories/liquid-js/liquidjs`.
> **Legenda:** sem marcador = presente em ambos | **[Ruby]** = exclusivo Ruby Liquid | **[JS]** = exclusivo LiquidJS

### Convenções de localização de código

```
Ruby:  lib/liquid/tags/<arquivo>.rb
       lib/liquid/<arquivo>.rb
JS:    src/tags/<arquivo>.ts
       src/filters/<arquivo>.ts
       src/drop/<arquivo>.ts
       src/render/<arquivo>.ts
```

---

## Tags

### Tags de output / expressão

| Tag | Sintaxe | Diferenças | Ruby | JS |
|-----|---------|------------|------|----|
| `{{ }}` | `{{ expressao }}` | — | `lib/liquid/variable.rb` · `Variable#render_to_output_buffer(L112)` | `src/template/output.ts` via `Value#value()` |
| `echo` | `{% echo expressao %}` | JS: value é opcional (sem value não emite nada) | `lib/liquid/tags/echo.rb` · `Echo#render(L30)` | `src/tags/echo.ts` · `render(L15)` |
| `liquid` | `{% liquid tag1\ntag2 %}` | — | `lib/liquid/tags/` — tag `liquid` via bloco (não há arquivo separado; lido em `block_body.rb`) | `src/tags/liquid.ts` · `render(L11)` · parse via `readLiquidTagTokens` |
| `#` inline comment | `{%# comentário %}` | Ambos: multi-linha exige `#` em cada linha | `lib/liquid/tags/inline_comment.rb` · `InlineComment#render_to_output_buffer(L19)` | `src/tags/inline-comment.ts` · `render(L10)` |
| `raw` | `{% raw %}...{% endraw %}` | — | `lib/liquid/tags/raw.rb` · `Raw#render_to_output_buffer(L37)` | `src/tags/raw.ts` · `render(L15)` |
| `comment` | `{% comment %}...{% endcomment %}` | Ruby: suporta nesting de `comment`/`raw`; JS: não | `lib/liquid/tags/comment.rb` · `Comment#render_to_output_buffer(L19)` · `parse_raw_tag_body(L78)` | `src/tags/comment.ts` · `render(L13)` (vazio) |

### Tags de variável / estado

| Tag | Sintaxe | Diferenças | Ruby | JS |
|-----|---------|------------|------|----|
| `assign` | `{% assign var = valor %}` | Ruby: rastreado por resource limits (`assign_score_of`); JS: stores em `ctx.bottom()` | `lib/liquid/tags/assign.rb` · `Assign#render_to_output_buffer(L43)` · `assign_score_of(L54)` | `src/tags/assign.ts` · `render(L22)` |
| `capture` | `{% capture var %}...{% endcapture %}` | Ruby: rastreado por resource limits; JS: nome pode ser quoted string | `lib/liquid/tags/capture.rb` · `Capture#render_to_output_buffer(L33)` | `src/tags/capture.ts` · `render(L33)` |
| `increment` | `{% increment var %}` | Ambos: armazenado em `environments`; outputs 0, 1, 2…; compartilha slot com `decrement` | `lib/liquid/tags/increment.rb` · `Increment#render_to_output_buffer(L33)` | `src/tags/increment.ts` · `render(L13)` |
| `decrement` | `{% decrement var %}` | Ambos: outputs -1, -2, …; Ruby: output-then-decrement; JS: pre-decrement-then-output (resultado igual) | `lib/liquid/tags/decrement.rb` · `Decrement#render_to_output_buffer(L33)` | `src/tags/decrement.ts` · `render(L13)` |

### Tags condicionais

| Tag | Sub-tags | Diferenças | Ruby | JS |
|-----|----------|------------|------|----|
| `if` | `elsif`, `else`, `endif` | Ruby: aceita `<>` como alias de `!=`; JS: aceita operador unário `not`; JS: `elsif` após `else` é erro explícito | `lib/liquid/tags/if.rb` · `If#render_to_output_buffer(L50)` · parsing: `strict2_parse(L63)` / `strict_parse(L96)` / `lax_parse(L81)` · condições em `lib/liquid/condition.rb` · `Condition#evaluate(L68)` | `src/tags/if.ts` · `render(L38)` |
| `unless` | `elsif`, `else`, `endunless` | JS: `elsif` dentro de `unless` usa `isTruthy` (não invertido); Ruby: comportamento equivalente ao `if` | `lib/liquid/tags/unless.rb` · `Unless < If` · `render_to_output_buffer(L23)` | `src/tags/unless.ts` · `render(L43)` |
| `case` | `when`, `else`, `endcase` | Ambos: `when` suporta múltiplos valores separados por `or` ou `,`; branches após `else` são ignorados | `lib/liquid/tags/case.rb` · `Case#render_to_output_buffer(L67)` · `record_when_condition(L106)` · `parse_strict2_when(L112)` / `parse_lax_when(L127)` | `src/tags/case.ts` · `render(L57)` |

**[Ruby] apenas:**

| Tag | Sub-tags | Notas | Ruby |
|-----|----------|-------|------|
| `ifchanged` | — | Renderiza só se output mudou desde última iteração; estado em `registers[:ifchanged]` | `lib/liquid/tags/ifchanged.rb` · `Ifchanged#render_to_output_buffer(L5)` |

### Tags de iteração

| Tag | Opções | Diferenças | Ruby | JS |
|-----|--------|------------|------|----|
| `for` | `offset: n`, `limit: n`, `reversed`, range `(a..b)` | Ruby: modifiers aplicados na ordem declarada; JS: ordem `offset→limit→reversed` por padrão (ou ordem declarada se `orderedFilterParameters: true`); JS: `offset: continue` para retomar do ponto anterior | `lib/liquid/tags/for.rb` · `For#render_to_output_buffer(L62)` · `render_segment(L149)` · `set_attribute(L176)` · `collection_segment(L114)` · `ParseTreeVisitor` no final do arquivo | `src/tags/for.ts` · `render(L46)` · modifiers: `offset(L108)` / `limit(L112)` / `reversed(L115)` · `blockScope(L104)` |
| `break` | — | — | `lib/liquid/tags/break.rb` · `Break#render_to_output_buffer(L23)` · `INTERRUPT = BreakInterrupt` (L21) | `src/tags/break.ts` · `render(L4)` · sets `ctx.breakCalled = true` |
| `continue` | — | — | `lib/liquid/tags/continue.rb` · `Continue#render_to_output_buffer(L16)` · `INTERRUPT = ContinueInterrupt` (L14) | `src/tags/continue.ts` · `render(L4)` · sets `ctx.continueCalled = true` |
| `cycle` | `[group:] v1, v2, ...` | Ambos: estado em register `cycle`; chave: `"cycle:{group}:{candidates}"` | `lib/liquid/tags/cycle.rb` · `Cycle#render_to_output_buffer(L33)` · `named?(L29)` · `strict2_parse(L53)` / `strict_parse(L95)` / `lax_parse(L99)` · `variables_from_string(L117)` | `src/tags/cycle.ts` · `render(L30)` · register key: `"cycle:${group}:${candidates}"` |
| `tablerow` | `cols: n`, `offset: n`, `limit: n`, range `(a..b)` | Ambos: gera `<tr class="rowN">`, `<td class="colN">`; Ruby: suporta `break` dentro de tablerow; JS: sem suporte a `break` em tablerow | `lib/liquid/tags/table_row.rb` · `TableRow#render_to_output_buffer(L81)` · `strict2_parse(L37)` / `strict_parse(L62)` / `lax_parse(L66)` | `src/tags/tablerow.ts` · `render(L44)` · gera `tr`/`td` inline no render · `blockScope(L92)` |

### Tags de inclusão de templates

| Tag | Sintaxe | Escopo | Diferenças | Ruby | JS |
|-----|---------|--------|------------|------|----|
| `include` | `{% include 'arquivo' [with var] [key: val ...] %}` | Compartilhado | Ruby: **deprecated**; suporta adicionalmente `for array as alias`; JS: não suporta `for`; JS: opção `jekyllInclude` injeta vars em objeto `include` | `lib/liquid/tags/include.rb` · `Include#render_to_output_buffer(L36)` · `prepend Tag::Disableable (L22)` · `strict2_parse(L88)` / `strict_parse(L107)` / `lax_parse(L111)` | `src/tags/include.ts` · `render(L31)` · `partialScope(L55)` |
| `render` | `{% render 'arquivo' [with var [as alias]] [for collection [as alias]] [key: val ...] %}` | Isolado | Ruby: escopo isolado exceto globals; JS: usa `ctx.spawn()`; mesma semântica | `lib/liquid/tags/render.rb` · `Render#render_to_output_buffer(L41)` · `render_tag(L45)` · `disable_tags "include" (L28)` · `for_loop?(L37)` · `strict2_parse(L85)` / `strict_parse(L111)` / `lax_parse(L115)` | `src/tags/render.ts` · `render(L49)` · `parseFilePath(L125)` / `renderFilePath(L141)` · `partialScope(L85)` |

**[Ruby] apenas:**

| Tag | Notas | Ruby |
|-----|-------|------|
| `doc` | `{% doc %}...{% enddoc %}` — LiquidDoc; ignorado pelo renderer | `lib/liquid/tags/doc.rb` · `Doc#render_to_output_buffer(L58)` · `blank?(L62)` · `raise_nested_doc_error(L78)` |

**[JS] apenas:**

| Tag | Sintaxe | Notas | JS |
|-----|---------|-------|----|
| `layout` | `{% layout 'arquivo' [key: val ...] %}` | Herança de layout; consome tokens restantes como templates filhos; `{% layout none %}` desabilita e renderiza direto | `src/tags/layout.ts` · `render(L21)` · `blockMode: STORE` → `STORE` → `OUTPUT` · `children(L52)` · `partialScope(L72)` |
| `block` | `{% block nome %}...{% endblock %}` | Define bloco sobreponível; `{{ block.super }}` acessa conteúdo do pai | `src/tags/block.ts` · `render(L23)` · `getBlockRender` interno · `BlockDrop` em `src/drop/block-drop.ts` |

---

## Filtros

> **Fonte Ruby:** todos em `lib/liquid/standardfilters.rb` · módulo `Liquid::StandardFilters`
> **Fonte JS:** distribuídos em `src/filters/` · registrados em `src/filters/index.ts`

### String — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `downcase` | `v \| downcase` | — | `standardfilters.rb:L76` | `string.ts · downcase(L46)` |
| `upcase` | `v \| upcase` | — | `standardfilters.rb:L85` | `string.ts · upcase(L52)` |
| `capitalize` | `v \| capitalize` | Primeira letra maiúscula, resto minúsculas | `standardfilters.rb:L94` | `string.ts · capitalize(L121)` |
| `append` | `v \| append: arg` | — | `standardfilters.rb:L684` | `string.ts · append(L20)` |
| `prepend` | `v \| prepend: arg` | — | `standardfilters.rb:L713` | `string.ts · prepend(L27)` |
| `remove` | `v \| remove: arg` | — | `standardfilters.rb:L654` | `string.ts · remove(L58)` |
| `remove_first` | `v \| remove_first: arg` | — | `standardfilters.rb:L664` | `string.ts · remove_first(L64)` |
| `remove_last` | `v \| remove_last: arg` | — | `standardfilters.rb:L674` | `string.ts · remove_last(L71)` |
| `replace` | `v \| replace: pattern, repl` | — | `standardfilters.rb:L606` | `string.ts · replace(L127)` |
| `replace_first` | `v \| replace_first: p, r` | — | `standardfilters.rb:L620` | `string.ts · replace_first(L135)` |
| `replace_last` | `v \| replace_last: p, r` | — | `standardfilters.rb:L633` | `string.ts · replace_last(L142)` |
| `split` | `v \| split: sep` | JS: trailing empty strings removidas (comportamento Ruby) | `standardfilters.rb:L268` | `string.ts · split(L91)` |
| `lstrip` | `v \| lstrip` | JS: aceita argumento `chars` opcional (strip do conjunto de chars) | `standardfilters.rb:L300` | `string.ts · lstrip(L34)` |
| `rstrip` | `v \| rstrip` | JS: aceita argumento `chars` opcional | `standardfilters.rb:L309` | `string.ts · rstrip(L79)` |
| `strip` | `v \| strip` | JS: aceita argumento `chars` opcional | `standardfilters.rb:L291` | `string.ts · strip(L102)` |
| `strip_html` | `v \| strip_html` | Remove `<script>`, `<style>`, tags HTML, comentários HTML | `standardfilters.rb:L318` | `html.ts · strip_html(L43)` |
| `strip_newlines` | `v \| strip_newlines` | Remove `\r?\n` | `standardfilters.rb:L329` | `string.ts · strip_newlines(L115)` |
| `newline_to_br` | `v \| newline_to_br` | Substitui `\r?\n` por `<br />\n` | `standardfilters.rb:L723` | `html.ts · newline_to_br(L37)` |
| `truncate` | `v \| truncate[: n[, ellipsis]]` | Default: 50 chars, `'...'`; sufixo incluído na contagem | `standardfilters.rb:L218` | `string.ts · truncate(L151)` |
| `truncatewords` | `v \| truncatewords[: n[, ellipsis]]` | Default: 15 palavras, `'...'`; JS: `words <= 0` usa 1 | `standardfilters.rb:L241` | `string.ts · truncatewords(L159)` |
| `size` | `v \| size` | Funciona em strings e arrays | `standardfilters.rb:L65` | `array.ts · size(L49)` |
| `slice` | `v \| slice: begin[, length]` | Funciona em strings e arrays; `begin < 0` → desde o fim; default length=1 | `standardfilters.rb:L197` | `array.ts · slice(L100)` |

### String — Exclusivos **[Ruby]**

| Filtro | Assinatura | Notas | Ruby (linha) |
|--------|-----------|-------|--------------|
| `squish` | `v \| squish` | Strip + colapsa whitespace interno em espaço único | `standardfilters.rb:L280` |
| `h` | `v \| h` | Alias para `escape` | `standardfilters.rb:L103` (definido junto ao `escape`) |

### String — Exclusivos **[JS]**

| Filtro | Assinatura | Notas | JS (arquivo · linha) |
|--------|-----------|-------|----------------------|
| `normalize_whitespace` | `v \| normalize_whitespace` | Equivalente funcional ao `squish` do Ruby | `string.ts · normalize_whitespace(L168)` |
| `number_of_words` | `v \| number_of_words[: mode]` | Conta palavras; modos: `'cjk'` (CJK + não-CJK), `'auto'` (CJK se presente, senão split), default (split) | `string.ts · number_of_words(L174)` |
| `array_to_sentence_string` | `arr \| array_to_sentence_string[: connector]` | `"a, b, and c"`; connector default `'and'`; 0 items → `''`, 1 item → `a`, 2 → `a and b` | `string.ts · array_to_sentence_string(L210)` |
| `xml_escape` | `v \| xml_escape` | Alias para `escape` | `html.ts · xml_escape(L24)` |

### HTML — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `escape` | `v \| escape` | HTML escape `&`, `<`, `>`, `"`, `'`; Ruby: alias `h`; JS: alias `xml_escape` | `standardfilters.rb:L103` | `html.ts · escape(L18)` |
| `escape_once` | `v \| escape_once` | Unescape então escape (não re-escapa entidades já escapadas) | `standardfilters.rb:L114` | `html.ts · escape_once(L33)` |

### URL — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `url_encode` | `v \| url_encode` | `encodeURIComponent`; espaços viram `+` | `standardfilters.rb:L126` | `url.ts · url_encode(L4)` |
| `url_decode` | `v \| url_decode` | `decodeURIComponent`; `+` vira espaço | `standardfilters.rb:L137` | `url.ts · url_decode(L3)` |

### URL — Exclusivos **[JS]**

| Filtro | Assinatura | Notas | JS (arquivo · linha) |
|--------|-----------|-------|----------------------|
| `cgi_escape` | `v \| cgi_escape` | `encodeURIComponent` com `+` e hex maiúsculo para `!'()*` | `url.ts · cgi_escape(L5)` |
| `uri_escape` | `v \| uri_escape` | `encodeURI` preservando `[` e `]` | `url.ts · uri_escape(L8)` |
| `slugify` | `v \| slugify[: mode[, cased]]` | Modos: `'raw'`, `'default'`, `'pretty'`, `'ascii'`, `'latin'`, `'none'`; `cased=false` por padrão | `url.ts · slugify(L22)` |

### Base64 — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `base64_encode` | `v \| base64_encode` | — | `standardfilters.rb:L150` | `base64.ts · base64_encode(L11)` |
| `base64_decode` | `v \| base64_decode` | Ruby: lança erro se input inválido | `standardfilters.rb:L160` | `base64.ts · base64_decode(L17)` |

### Base64 — Exclusivos **[Ruby]**

| Filtro | Assinatura | Notas | Ruby (linha) |
|--------|-----------|-------|--------------|
| `base64_url_safe_encode` | `v \| base64_url_safe_encode` | URL-safe Base64 | `standardfilters.rb:L172` |
| `base64_url_safe_decode` | `v \| base64_url_safe_decode` | URL-safe Base64 decode | `standardfilters.rb:L182` |

### Math — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `abs` | `v \| abs` | — | `standardfilters.rb:L803` | `math.ts · abs(L3)` |
| `plus` | `v \| plus: n` | — | `standardfilters.rb:L813` | `math.ts · plus(L10)` |
| `minus` | `v \| minus: n` | — | `standardfilters.rb:L823` | `math.ts · minus(L9)` |
| `times` | `v \| times: n` | — | `standardfilters.rb:L833` | `math.ts · times(L12)` |
| `divided_by` | `v \| divided_by: n` | Ruby: tipo do resultado = tipo do divisor; levanta `ZeroDivisionError` com zero; JS: aceita `integerArithmetic` como 2º arg (`Math.floor`) | `standardfilters.rb:L843` | `math.ts · divided_by(L7)` |
| `modulo` | `v \| modulo: n` | Ruby: levanta `ZeroDivisionError` com zero | `standardfilters.rb:L855` | `math.ts · modulo(L11)` |
| `round` | `v \| round[: decimals]` | Default 0 casas | `standardfilters.rb:L866` | `math.ts · round(L14)` |
| `ceil` | `v \| ceil` | — | `standardfilters.rb:L906` | `math.ts · ceil(L6)` |
| `floor` | `v \| floor` | — | `standardfilters.rb:L920` | `math.ts · floor(L8)` |
| `at_least` | `v \| at_least: n` | `max(v, n)` | `standardfilters.rb:L933` | `math.ts · at_least(L4)` |
| `at_most` | `v \| at_most: n` | `min(v, n)` | `standardfilters.rb:L948` | `math.ts · at_most(L5)` |

### Date — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `date` | `v \| date: format` | Ruby: retorna input se vazio/inválido; JS: format opcional (usa `opts.dateFormat`); JS: aceita `timezoneOffset` como 2º arg; JS: `'now'`/`'today'` → hora atual; JS: `opts.preserveTimezones` | `standardfilters.rb:L770` | `date.ts · date(L5)` |

### Date — Exclusivos **[JS]**

| Filtro | Assinatura | Notas | JS (arquivo · linha) |
|--------|-----------|-------|----------------------|
| `date_to_xmlschema` | `v \| date_to_xmlschema` | Formato `'%Y-%m-%dT%H:%M:%S%:z'` | `date.ts · date_to_xmlschema(L15)` |
| `date_to_rfc822` | `v \| date_to_rfc822` | Formato `'%a, %d %b %Y %H:%M:%S %z'` | `date.ts · date_to_rfc822(L19)` |
| `date_to_string` | `v \| date_to_string[: type[, style]]` | Mês abreviado (`%b`); `type='ordinal'`, `style='US'` usa formato americano | `date.ts · date_to_string(L23)` |
| `date_to_long_string` | `v \| date_to_long_string[: type[, style]]` | Mês completo (`%B`) | `date.ts · date_to_long_string(L27)` |

### Array — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `join` | `arr \| join[: sep]` | Default sep `' '` | `standardfilters.rb:L339` | `array.ts · join(L8)` |
| `first` | `arr \| first` | — | `standardfilters.rb:L782` | `array.ts · first(L16)` |
| `last` | `arr \| last` | — | `standardfilters.rb:L793` | `array.ts · last(L15)` |
| `reverse` | `arr \| reverse` | — | `standardfilters.rb:L466` | `array.ts · reverse(L17)` |
| `sort` | `arr \| sort[: property]` | Ruby: nil-safe (nils vão pro final); JS: `property` pode ser path com `.` | `standardfilters.rb:L349` | `array.ts · sort(L23)` |
| `sort_natural` | `arr \| sort_natural[: property]` | Case-insensitive | `standardfilters.rb:L371` | `array.ts · sort_natural(L40)` |
| `map` | `arr \| map: property` | — | `standardfilters.rb:L476` | `array.ts · map(L51)` |
| `sum` | `arr \| sum[: property]` | JS: NaN → 0 | `standardfilters.rb:L982` | `array.ts · sum(L59)` |
| `compact` | `arr \| compact` | Ruby: aceita argumento `property` opcional; JS: não | `standardfilters.rb:L496` | `array.ts · compact(L68)` |
| `uniq` | `arr \| uniq` | Ruby: aceita argumento `property` opcional; JS: usa `new Set` sem property | `standardfilters.rb:L446` | `array.ts · uniq(L252)` |
| `concat` | `arr \| concat: arr2` | — | `standardfilters.rb:L699` | `array.ts · concat(L73)` |
| `where` | `arr \| where: prop[, expected]` | Sem `expected` → filtra truthy; JS: opção `jekyllWhere` adiciona array membership check | `standardfilters.rb:L393` | `array.ts · where(L131)` · `filter(L117)` internal |
| `reject` | `arr \| reject: prop[, expected]` | Inverso de `where` | `standardfilters.rb:L404` | `array.ts · reject(L135)` |
| `find` | `arr \| find: prop[, expected]` | — | `standardfilters.rb:L425` | `array.ts · find(L245)` · `search(L200)` internal |
| `find_index` | `arr \| find_index: prop[, expected]` | — | `standardfilters.rb:L436` | `array.ts · find_index(L235)` |
| `has` | `arr \| has: prop[, expected]` | Retorna boolean | `standardfilters.rb:L414` | `array.ts · has(L225)` |

### Array — Exclusivos **[JS]**

| Filtro | Assinatura | Notas | JS (arquivo · linha) |
|--------|-----------|-------|----------------------|
| `where_exp` | `arr \| where_exp: itemName, exp` | Filtra por expressão Liquid; `exp` avaliada com `itemName` no scope | `array.ts · where_exp(L139)` · `filter_exp(L107)` internal |
| `reject_exp` | `arr \| reject_exp: itemName, exp` | Inverso de `where_exp` | `array.ts · reject_exp(L143)` |
| `group_by` | `arr \| group_by: property` | Retorna `[{name, items}, ...]` | `array.ts · group_by(L147)` |
| `group_by_exp` | `arr \| group_by_exp: itemName, exp` | Agrupa por expressão Liquid | `array.ts · group_by_exp(L160)` |
| `has_exp` | `arr \| has_exp: itemName, exp` | Boolean via expressão | `array.ts · has_exp(L230)` · `search_exp(L209)` internal |
| `find_exp` | `arr \| find_exp: itemName, exp` | Primeiro item via expressão | `array.ts · find_exp(L250)` |
| `find_index_exp` | `arr \| find_index_exp: itemName, exp` | Índice via expressão | `array.ts · find_index_exp(L240)` |
| `push` | `arr \| push: item` | Adiciona ao final (retorna novo array) | `array.ts · push(L79)` |
| `pop` | `arr \| pop` | Remove do final (retorna novo array) | `array.ts · pop(L89)` |
| `unshift` | `arr \| unshift: item` | Adiciona ao início (retorna novo array) | `array.ts · unshift(L83)` |
| `shift` | `arr \| shift` | Remove do início (retorna novo array) | `array.ts · shift(L94)` |
| `sample` | `v \| sample[: count]` | Amostra aleatória; `count=1` retorna item único; aceita string ou array | `array.ts · sample(L258)` |

### Misc — Comuns

| Filtro | Assinatura | Diferenças | Ruby (linha) | JS (arquivo · linha) |
|--------|-----------|------------|--------------|----------------------|
| `default` | `v \| default: val[, allow_false: true]` | Ambos suportam `allow_false` como named/keyword arg; retorna `val` se `v` é falsy ou empty (string/array vazia, objeto sem chaves) | `standardfilters.rb:L969` | `misc.ts · default(L5)` (exportado como `default` key no objeto exportado `L39`) |

### Misc — Exclusivos **[JS]**

| Filtro | Assinatura | Notas | JS (arquivo · linha) |
|--------|-----------|-------|----------------------|
| `json` | `v \| json[: space]` | `JSON.stringify(v, null, space)`; space default 0 | `misc.ts · json(L13)` |
| `jsonify` | `v \| jsonify[: space]` | Alias para `json` | `misc.ts` (mesmo export `json`, alias no objeto final `L39`) |
| `inspect` | `v \| inspect[: space]` | `JSON.stringify` com proteção a referências circulares (`'[Circular]'`) | `misc.ts · inspect(L18)` |
| `to_integer` | `v \| to_integer` | `Number(v)` | `misc.ts · to_integer(L31)` |
| `raw` | `v \| raw` | Passa valor sem escape/evaluation | `misc.ts · raw(L36)` (objeto `{ raw: true, handler: identify }`) |

---

## Drops (Objetos Especiais)

### Drop — Base

| Feature | Ruby | JS |
|---------|------|----|
| Arquivo | `lib/liquid/drop.rb` | `src/drop/drop.ts` |
| Catch-all | `liquid_method_missing(name)` · L33 | `liquidMethodMissing(key, context)` · L4 |
| Quando não encontrado | `UndefinedDropMethod` se strict | retorna `undefined` |
| Invocação | `invoke_drop(key)` L39; `[key]` alias | acesso direto via `readProperty` em `context/context.ts:L112` |
| Whitelist de métodos | `invokable_methods` L68 (públicos exceto blacklist) | qualquer método público da classe |
| Context injection | `attr_writer :context` L26 | passado como arg em `liquidMethodMissing` |

### ForloopDrop — Comuns

| Propriedade | Tipo | Notas | Ruby (linha) | JS (linha) |
|-------------|------|-------|--------------|------------|
| `length` | number | Total de iterações | `forloop_drop.rb:L20` (attr) | `forloop-drop.ts:L8` (field) |
| `name` | string | `"variavel-colecao"` | `forloop_drop.rb:L34` (attr) | `forloop-drop.ts:L9` (field) |
| `index` | number | 1-based | `forloop_drop.rb:L40` | `forloop-drop.ts:L18` |
| `index0` | number | 0-based | `forloop_drop.rb:L48` | `forloop-drop.ts:L15` |
| `rindex` | number | Restantes incluindo atual (1-based reverso) | `forloop_drop.rb:L55` | `forloop-drop.ts:L27` |
| `rindex0` | number | Restantes excluindo atual (0-based reverso) | `forloop_drop.rb:L62` | `forloop-drop.ts:L30` |
| `first` | boolean | `i === 0` | `forloop_drop.rb:L69` | `forloop-drop.ts:L21` |
| `last` | boolean | `i === length - 1` | `forloop_drop.rb:L76` | `forloop-drop.ts:L24` |
| `increment!` / `next()` | — | Avança contador | `forloop_drop.rb:L83` | `forloop-drop.ts:L12` (método `next`) |

> **[Ruby] apenas:** `parentloop` L28 — referência ao ForloopDrop do loop pai (ou nil)

### TablerowloopDrop — Comuns (extends ForloopDrop)

| Propriedade | Tipo | Notas | Ruby (linha) | JS (linha) |
|-------------|------|-------|--------------|------------|
| `row` | number | Linha atual (1-based) | `tablerowloop_drop.rb:L33` | `tablerowloop-drop.ts:L10` |
| `col` | number | Coluna atual (1-based) | `tablerowloop_drop.rb:L27` | `tablerowloop-drop.ts:L16` |
| `col0` | number | Coluna atual (0-based) | `tablerowloop_drop.rb:L52` | `tablerowloop-drop.ts:L13` |
| `col_first` | boolean | `col0 === 0` | `tablerowloop_drop.rb:L87` | `tablerowloop-drop.ts:L19` |
| `col_last` | boolean | `col === cols` | `tablerowloop_drop.rb:L103` | `tablerowloop-drop.ts:L22` |

### EmptyDrop (`empty`) — Comuns

| Comportamento | Notas | Ruby | JS |
|---------------|-------|------|----|
| `== empty` | string/array → `.length === 0`; objeto → sem chaves; outro EmptyDrop → `false` | `condition.rb:L156` · `liquid_empty?` | `src/drop/empty-drop.ts · EmptyDrop#equals(L6)` |
| `valueOf()` | `''` | Ruby retorna `''` via `to_s` | `empty-drop.ts:L24` |
| Comparações `>`, `<`, `>=`, `<=` | sempre `false` | Ruby: `Comparable` protocol; raises no compare | `empty-drop.ts: gt(L12), geq(L15), lt(L18), leq(L21)` |

### BlankDrop (`blank`) — Comuns

| Comportamento | Notas | Ruby | JS |
|---------------|-------|------|----|
| `== blank` | nil/null/undefined → `true`; `false` → `true`; string só whitespace → `true` | `condition.rb:L135` · `liquid_blank?` | `src/drop/blank-drop.ts · BlankDrop#equals(L5)` |

### Drops Exclusivos **[JS]**

| Drop | Disponível como | Notas | JS |
|------|-----------------|-------|----|
| `NullDrop` | literal `null`/`nil` na comparação | Interface `Comparable` explícita; `valueOf → null`; comparações sempre `false` | `src/drop/null-drop.ts` · `equals(L5)` |
| `BlockDrop` | variável `block` dentro de `{% block %}` | `block.super` renderiza conteúdo do bloco pai | `src/drop/block-drop.ts` · `super()(L15)` · usado em `src/tags/block.ts:L27` |

---

## Operadores — Comuns

| Operador | Tipo | Comportamento | Ruby | JS |
|----------|------|---------------|------|----|
| `==` | binário | Igualdade; `Comparable`-aware; JS arrays comparam elemento a elemento | `condition.rb:L13` (`@@operators` hash) | `src/render/operator.ts · equals(L42)` · `defaultOperators['=='](L13)` |
| `!=` | binário | Desigualdade | `condition.rb:L13` | `operator.ts · defaultOperators['!='](L14)` |
| `>`, `<`, `>=`, `<=` | binário | Comparação; `Comparable`-aware | `condition.rb:L13` | `operator.ts · defaultOperators(L15-30)` |
| `contains` | binário | String: `indexOf > -1`; Array: `include?` / `some(equals)` | `condition.rb:L13` | `operator.ts · defaultOperators['contains'](L31)` |
| `and` | binário | Curto-circuito; sem precedência (avalia esquerda para direita) | `condition.rb` · parsed em `if.rb#parse_binary_comparisons(L103)` | `operator.ts · defaultOperators['and'](L38)` · `src/render/boolean.ts · isTruthy(L4)` |
| `or` | binário | Curto-circuito | idem | `operator.ts · defaultOperators['or'](L39)` |

**[Ruby] apenas:** `<>` — alias para `!=` · `condition.rb:L13`

**[JS] apenas:** `not` (unário) · `operator.ts · defaultOperators['not'](L37)` · usa `src/render/boolean.ts · isFalsy(L8)` · operadores customizáveis via opção `operators` em `src/liquid-options.ts:L90`

---

## Truthiness

| Valor | Ruby | JS default | JS `jsTruthy: true` |
|-------|------|------------|---------------------|
| `false` | falsy | falsy | falsy |
| `nil` / `null` / `undefined` | falsy | falsy | falsy |
| `0` | **truthy** | **truthy** | falsy |
| `""` | **truthy** | **truthy** | falsy |
| `[]` | **truthy** | **truthy** | falsy |
| Qualquer outro | truthy | truthy | truthy |

> Ruby: `lib/liquid/condition.rb` · `interpret_condition(L166)` (nil/false = falsy, resto truthy)
> JS: `src/render/boolean.ts` · `isFalsy(L8)` — com `jsTruthy: true` retorna `!val`; sem, só `false/undefined/null`
> JS `jsTruthy` configurado em `src/liquid-options.ts:L43`

---

## Literais / Valores Especiais

| Literal | Ruby | JS | Notas |
|---------|------|-----|-------|
| `nil` / `null` | ✓ (`nil`) | ✓ (`null`) | Ruby: `expression.rb · LITERALS(L5)` · JS: `NullDrop` em `src/drop/null-drop.ts` |
| `true` / `false` | ✓ | ✓ | Ruby: `expression.rb · LITERALS(L6-7)` |
| integer | ✓ | ✓ | Ruby: `expression.rb · INTEGER_REGEX(L24)` · JS: `src/tokens/` |
| float | ✓ | ✓ | Ruby: `expression.rb · FLOAT_REGEX(L25)` |
| string `"..."` / `'...'` | ✓ | ✓ | Ruby: `lexer.rb:L26,35` |
| range `(a..b)` | ✓ | ✓ | Ruby: `expression.rb · RANGES_REGEX(L23)` · Suportado em `for` e `tablerow` |
| `empty` | ✓ | ✓ | Ruby: `expression.rb · LITERALS(L12)` · `condition.rb · liquid_empty?(L156)` · JS: `src/drop/empty-drop.ts` |
| `blank` | ✓ | ✓ | Ruby: `expression.rb · LITERALS(L13)` · `condition.rb · liquid_blank?(L135)` · JS: `src/drop/blank-drop.ts` |

---

## Acesso a Variáveis

| Sintaxe | Ambos | Diferenças | Ruby | JS |
|---------|-------|------------|------|----|
| `variavel` | ✓ | Lookup em escopos empilhados | `context.rb · find_variable(L176)` | `context.ts · _get(L69)` · `findScope(L106)` |
| `obj.prop` | ✓ | — | `variable_lookup.rb` | `context.ts · _getFromScope(L79)` |
| `obj[key]` | ✓ | — | `variable_lookup.rb` | `context.ts · readProperty(L112)` |
| `array[0]` | ✓ | — | `variable_lookup.rb` | `context.ts · readProperty(L112)` |
| `array[-1]` | ✓ | JS: `obj.length + key` explícito | `variable_lookup.rb` | `context.ts:L115` (`key < 0` branch) |
| `array.first` | ✓ | Atalho | `variable_lookup.rb` | `context.ts · readFirst(L130)` |
| `array.last` | ✓ | Atalho | `variable_lookup.rb` | `context.ts · readLast(L135)` |
| `obj.size` | ✓ | `length` para arrays/strings | `standardfilters.rb · size(L65)` | `context.ts · readSize(L140) [key === 'size']` |
| `forloop.index` etc. | ✓ | — | `forloop_drop.rb` | `src/drop/forloop-drop.ts` |

---

## Controle de Whitespace

### Marcadores em template — Comuns

| Marcador | Efeito | Ruby | JS |
|----------|--------|------|----|
| `{%-` | Remove whitespace à esquerda da tag | `lib/liquid/lexer.rb` (tokenizer reconhece `-`) | `src/parser/` (trimLeft flag no token) |
| `-%}` | Remove whitespace à direita da tag | idem | idem |
| `{{-` | Remove whitespace à esquerda do output | idem | idem |
| `-}}` | Remove whitespace à direita do output | idem | idem |

### Opções globais — Exclusivos **[JS]**

| Opção | Default | Notas | JS |
|-------|---------|-------|----|
| `trimTagRight` | `false` | Trim direita de `{% %}` até `\n` (inclusive) | `src/liquid-options.ts:L62` |
| `trimTagLeft` | `false` | Trim esquerda de `{% %}` | `src/liquid-options.ts:L63` |
| `trimOutputRight` | `false` | Trim direita de `{{ }}` até `\n` (inclusive) | `src/liquid-options.ts:L65` |
| `trimOutputLeft` | `false` | Trim esquerda de `{{ }}` | `src/liquid-options.ts:L66` |
| `greedy` | `true` | Trim consome todos os espaços/`\n` consecutivos | `src/liquid-options.ts:L76` |

---

## Herança de Layout

### Exclusivo **[JS]**

| Mecanismo | Notas | JS |
|-----------|-------|----|
| `{% layout 'arquivo' %}` | Declara layout pai; tokens restantes do filho processados em `blockMode: STORE` | `src/tags/layout.ts · render(L21)` · `BlockMode.STORE` de `src/context/block-mode.ts` |
| `{% layout none %}` | Desativa layout; renderiza direto em `blockMode: OUTPUT` | `src/tags/layout.ts:L23` (verifica `file === undefined`) |
| `{% block nome %}...{% endblock %}` | Bloco sobreponível; em modo `STORE` guarda render function | `src/tags/block.ts · render(L23)` · `getBlockRender` |
| `{{ block.super }}` | Acessa conteúdo do bloco pai | `src/drop/block-drop.ts · super()(L15)` · injetado em `block.ts:L27` |
| Bloco anônimo (`blocks['']`) | Conteúdo do filho fora de `{% block %}` → bloco anônimo disponível no pai | `src/tags/layout.ts:L35` |

---

## Sistema de Filtros

| Feature | Ruby | JS |
|---------|------|----|
| Arquivo central | `lib/liquid/standardfilters.rb` | `src/filters/index.ts` (agrega todos) |
| Registro | `environment.rb · register_filter(L95)` · `strainer_template.rb` | `src/liquid.ts · registerFilter(L100)` |
| Filtros posicionais | `variable.rb · lax_parse(L42)` / `strict_parse(L61)` | `src/template/value.ts` |
| Filtros com keyword args | `variable.rb · parse_filterargs(L93)` — passados como hash | `misc.ts · default(L5)` — named args `[string, any][]` |
| Filtro undefined | `lib/liquid/errors.rb · UndefinedFilter` · `strict_filters` em `environment.rb:L8` | `src/liquid.ts` — `strictFilters: true` → exceção; `false` → pula silenciosamente |
| Filtro global em todo output | `context.rb · apply_global_filter(L75)` · `global_filter:` render option | — (não implementado no JS) |

---

## Configuração / Opções

### Filesystem e templates — Comuns (conceito)

| Feature | Ruby | JS |
|---------|------|----|
| Resolver arquivos | `environment.rb · file_system(L23)` — objeto com `read_template_file` · `lib/liquid/file_system.rb` | `src/liquid-options.ts · root/partials/layouts(L14-18)` |
| Filesystem base | `file_system.rb · BlankFileSystem(L17)` / `LocalFileSystem(L42)` | `src/fs/fs-impl.ts` (Node.js `fs`) |
| Filesystem customizado | `environment.rb:L23` `file_system=` | `src/liquid-options.ts:L84` `fs:` option |
| Templates em memória | — | `src/liquid-options.ts:L87` `templates:` option · `src/fs/map-fs.ts` **[JS]** |
| Cache de templates | — | `src/liquid-options.ts:L36` `cache:` option · `src/cache/` **[JS]** |
| Extensão implícita | via pattern em `LocalFileSystem#initialize(L45)` | `src/liquid-options.ts:L34` `extname` **[JS]** |
| Partials dinâmicos | implícito | `src/liquid-options.ts:L41` `dynamicPartials` (default `true`) **[JS]** |
| Referências relativas | — | `src/liquid-options.ts:L31` `relativeReference` (default `true`) **[JS]** |

### Escopo e variáveis

| Feature | Ruby | JS |
|---------|------|----|
| Variáveis globais | opção `globals:` em `template.rb · apply_options_to_context(L206)` | `src/liquid-options.ts:L88` `globals: {}` · `context.ts · globals(L28)` |
| Strict variables | render option `strict_variables:` · `context.rb · find_variable(L176)` | `src/liquid-options.ts:L46` `strictVariables:` · `context.ts:L40` |
| Strict filters | render/env option `strict_filters:` · `strainer_template.rb` | `src/liquid-options.ts:L48` `strictFilters:` |
| Limitar propriedades | — | `src/liquid-options.ts:L50` `ownPropertyOnly:` (default `true`) · `context.ts:L42` |

### Comportamento — Exclusivos **[JS]**

| Opção | Default | Notas | JS |
|-------|---------|-------|----|
| `jsTruthy` | `false` | Usa JavaScript truthiness | `src/liquid-options.ts:L43` · `render/boolean.ts:L11` |
| `lenientIf` | `false` | Com `strictVariables: true`: permite var undefined em `if`/`elsif`/`unless`/`default` sem erro | `src/liquid-options.ts:L52` · usado em `tags/if.ts:L41` via `ctx.opts.lenientIf` |
| `ownPropertyOnly` | `true` | Ignora propriedades herdadas do prototype | `src/liquid-options.ts:L50` · `context.ts · readJSProperty(L125)` |
| `keepOutputType` | `false` | Preserva tipo do output (não converte para string) | `src/liquid-options.ts:L83` |
| `outputEscape` | `undefined` | Escape padrão aplicado a todo output: `'escape'`, `'json'`, ou função | `src/liquid-options.ts:L80` · `src/liquid-options.ts · normalize(L202)` |
| `catchAllErrors` | `false` | Coleta todos os erros sem parar no primeiro | `src/liquid-options.ts:L54` |
| `jekyllInclude` | `false` | `include` injeta variáveis em objeto `include` do scope | `src/liquid-options.ts:L33` · `tags/include.ts:L35` |
| `jekyllWhere` | `false` | Filtro `where` também faz array membership check | `src/liquid-options.ts:L35` · `filters/array.ts · expectedMatcher(L108)` |

### Delimitadores — Exclusivos **[JS]**

| Opção | Default | JS |
|-------|---------|-----|
| `tagDelimiterLeft` | `'{%'` | `src/liquid-options.ts:L68` |
| `tagDelimiterRight` | `'%}'` | `src/liquid-options.ts:L70` |
| `outputDelimiterLeft` | `'{{'` | `src/liquid-options.ts:L72` |
| `outputDelimiterRight` | `'}}'` | `src/liquid-options.ts:L74` |
| `keyValueSeparator` | `':'` | `src/liquid-options.ts` · usado em `Hash` constructor |

### Date — Exclusivos **[JS]**

| Opção | Default | JS |
|-------|---------|-----|
| `timezoneOffset` | local | `src/liquid-options.ts:L57` |
| `dateFormat` | `'%A, %B %-e, %Y at %-l:%M %P %z'` | `src/liquid-options.ts:L59` · `filters/date.ts:L7` |
| `locale` | sistema | `src/liquid-options.ts:L61` |
| `preserveTimezones` | `false` | `src/liquid-options.ts:L78` · `filters/date.ts · parseDate` |

### Misc — Exclusivos **[JS]**

| Opção | Default | JS |
|-------|---------|-----|
| `orderedFilterParameters` | `false` | `src/liquid-options.ts` · `tags/for.ts:L57` (`modifiers` order logic) |
| `operators` | defaultOperators | `src/liquid-options.ts:L93` · `src/render/operator.ts · defaultOperators(L12)` |

---

## Resource Limits / DoS

### **[Ruby]** — Score-based · `lib/liquid/resource_limits.rb`

| Limite | Descrição | Ruby |
|--------|-----------|------|
| `render_length_limit` | Bytes máximos no output de um template | `resource_limits.rb:L5` attr · `increment_write_score(L40)` |
| `render_score_limit` | Score de render por template (cada nó conta) | `resource_limits.rb:L6` attr · `increment_render_score(L27)` |
| `assign_score_limit` | Score de assign (string → bytesize; array/hash → soma recursiva) | `resource_limits.rb:L7` attr · `increment_assign_score(L33)` |
| `cumulative_render_score_limit` | Score acumulado entre múltiplos renders | `resource_limits.rb:L8` attr |
| `cumulative_assign_score_limit` | Score acumulado de assign | `resource_limits.rb:L9` attr |
| `MemoryError` lançado | Ao exceder qualquer limite | `resource_limits.rb · raise_limits_reached(L52)` |
| Scoring de assign | string → bytesize; array/hash → soma recursiva | `assign.rb · assign_score_of(L54)` |

### **[JS]** — Time/memory-based · `src/liquid-options.ts`

| Limite | Default | Notas | JS |
|--------|---------|-------|----|
| `parseLimit` | `Infinity` | Máximo de chars por `parse()` | `src/liquid-options.ts:L95` |
| `renderLimit` | `Infinity` | Máximo de ms por `render()` | `src/liquid-options.ts:L97` · `context.ts · renderLimit(L35)` |
| `memoryLimit` | `Infinity` | Máximo de alocações (arrays, strings concat, etc.) | `src/liquid-options.ts:L99` · `context.ts · memoryLimit(L34)` · usado como `this.context.memoryLimit.use(n)` em cada filtro |
| `templateLimit` em `RenderOptions` | — | Máximo de templates renderizados por chamada | `src/liquid-options.ts · RenderOptions(L104)` |

---

## Tratamento de Erros

### Comuns (conceito)

| Tipo | Ruby | JS |
|------|------|-----|
| Erro de sintaxe/parse | `lib/liquid/errors.rb · SyntaxError` | exceção no `parse()` |
| Variável undefined (strict) | `errors.rb · UndefinedVariable` · `context.rb · find_variable(L176)` | `src/util/underscore.ts · InternalUndefinedVariableError` · `context.ts:L84` |
| Filtro undefined (strict) | `errors.rb · UndefinedFilter` · `strainer_template.rb` | exceção em `liquid.ts` quando `strictFilters: true` |
| Overflow de includes | `errors.rb · StackLevelError` | via `parseLimit`/`renderLimit` |

### **[Ruby]** apenas · `lib/liquid/errors.rb` e `lib/liquid/environment.rb`

| Feature | Notas | Ruby |
|---------|-------|------|
| `error_mode` | `:lax` (ignora erros de sintaxe), `:warn` (default), `:strict`, `:strict2` | `environment.rb:L8` (attr) · cada tag tem `strict2_parse` / `strict_parse` / `lax_parse` |
| `template.errors` / `template.warnings` | Arrays acumulados no template | `template.rb:L18,19` attrs · `template.rb · configure_options(L183)` |
| `exception_renderer` | Proc para interceptar exceções (por ambiente ou por render) | `environment.rb:L19` · `template.rb · apply_options_to_context(L206)` |
| `ZeroDivisionError` | `divided_by` e `modulo` com zero | `errors.rb` · `standardfilters.rb:L843,855` |
| `DisabledError` | Tag desabilitada via `Disableable` | `errors.rb` · `tag/disableable.rb:L13` |
| `TemplateEncodingError` | Encoding inválido no template | `errors.rb` |
| Metadados em erros | `line_number`, `template_name`, `markup_context` | `errors.rb · Error#to_s(L8)` · `Error#message_prefix(L21)` |

### **[JS]** apenas

| Feature | Notas | JS |
|---------|-------|----|
| `catchAllErrors` | Coleta todos os erros sem parar no primeiro | `src/liquid-options.ts:L54` |
| `renderLimit` por chamada | Timeout por render individual | `src/liquid-options.ts · RenderOptions(L111)` |

---

## Extensibilidade

### Comuns (conceito)

| Feature | Ruby | JS |
|---------|------|----|
| Tags customizadas | `environment.rb · register_tag(L87)` | `src/liquid.ts · registerTag(L103)` |
| Filtros customizados | `environment.rb · register_filter(L95)` | `src/liquid.ts · registerFilter(L100)` |
| Filesystem customizado | `environment.rb:L23` `file_system=` | `src/liquid-options.ts:L84` `fs:` |

### **[Ruby]** apenas

| Feature | Notas | Ruby |
|---------|-------|------|
| `Tag::Disableable` | Mixin que verifica se tag está desabilitada antes de render | `lib/liquid/tag/disableable.rb` · `render_to_output_buffer(L5)` · `disabled_error(L13)` |
| `Tag::Disabler` | Mixin que desabilita tags listadas durante seu render | `lib/liquid/tag/disabler.rb` · `render_to_output_buffer(L5)` |
| `disable_tags "include"` | `render.rb:L28` desabilita `include` dentro de partials | `tags/render.rb:L28` · `context.rb · with_disabled_tags(L207)` · `tag_disabled?(L216)` |
| `Environment.build {}` | Builder imutável (freeze após construção) | `environment.rb · self.build(L46)` |
| `Environment.dangerously_override {}` | Override temporário do environment default | `environment.rb · self.dangerously_override(L63)` |
| Plugin via módulo Ruby | `include` de módulo com `register_filter` / `register_tag` | `environment.rb · register_filters(L103)` |

### **[JS]** apenas

| Método | Notas | JS |
|--------|-------|----|
| `liquid.plugin(fn)` | Registra plugin (`fn` recebe `this=liquidInstance, arg=LiquidClass`) | `src/liquid.ts · plugin(L106)` |
| `liquid.express()` | Adapter para Express.js view engine | `src/liquid.ts · express(L109)` |
| Delimitadores customizáveis | `tagDelimiterLeft/Right`, `outputDelimiterLeft/Right` | `src/liquid-options.ts:L68-74` |
| Operadores customizáveis | opção `operators` aceita qualquer `Record<string, OperatorHandler>` | `src/render/operator.ts · Operators(L10)` · `src/liquid-options.ts:L93` |

---

## Análise Estática

### **[Ruby]** — ParseTreeVisitor · `lib/liquid/parse_tree_visitor.rb`

| Feature | Notas | Ruby |
|---------|-------|------|
| `ParseTreeVisitor.for(node, callbacks)` | Cria visitor para nó da parse tree | `parse_tree_visitor.rb · self.for(L5)` |
| `visitor.add_callback_for(*classes) { \|node, ctx\| }` | Registra callback por classe de nó | `parse_tree_visitor.rb · add_callback_for(L19)` |
| `visitor.visit(context)` | Percorre árvore recursivamente | `parse_tree_visitor.rb · visit(L26)` |
| `node.nodelist` | Lista de nós filhos (interface padrão) | `parse_tree_visitor.rb · children(L36)` |
| Cada tag tem seu próprio `ParseTreeVisitor` inner class | ex: `assign.rb:L74`, `cycle.rb:L135`, `if.rb:L121`, etc. | ver cada `tags/*.rb` |

### **[JS]** — Static Analysis API · `src/liquid.ts` + `src/template/`

| Método | Retorno | Notas | JS |
|--------|---------|-------|----|
| `liquid.analyze(tpl, options?)` | `Promise<StaticAnalysis>` | Analisa variáveis, partials, scopes | `src/liquid.ts · analyze(L122)` |
| `liquid.analyzeSync(tpl, options?)` | `StaticAnalysis` | | `src/liquid.ts · analyzeSync(L126)` |
| `liquid.parseAndAnalyze(html, filename?, options?)` | `Promise<StaticAnalysis>` | | `src/liquid.ts · parseAndAnalyze(L130)` |
| `liquid.parseAndAnalyzeSync(html, filename?, options?)` | `StaticAnalysis` | | `src/liquid.ts · parseAndAnalyzeSync(L134)` |
| `liquid.variables(tpl, options?)` | `Promise<string[]>` | Lista variáveis sem propriedades | `src/liquid.ts · variables(L139)` |
| `liquid.variablesSync(tpl, options?)` | `string[]` | | `src/liquid.ts · variablesSync(L145)` |
| `liquid.fullVariables` / `variableSegments` / `globalVariables` | — | Variantes mais detalhadas | `src/liquid.ts:L150-L183` |

---

## API Pública

### Parse e Render — Comuns (conceito)

| Feature | Ruby | JS |
|---------|------|----|
| Parse | `lib/liquid/template.rb · Template.parse(L80)` | `src/liquid.ts · parse(L29)` |
| Render | `template.rb · render(L133)` | `src/liquid.ts · render(L37)` |
| Render com rethrow | `template.rb · render!(L174)` | — (JS sempre lança por padrão) |
| Render de arquivo | via `include`/`render` + `file_system` | `src/liquid.ts · renderFile(L78)` |
| Parse + render em passo único | — | `src/liquid.ts · parseAndRender(L52)` **[JS]** |
| Eval de expressão | — | `src/liquid.ts · evalValue(L93)` **[JS]** |

### Render — Exclusivos **[JS]** · `src/liquid.ts`

| Método | Retorno | JS |
|--------|---------|-----|
| `renderSync(tpl, scope?, opts?)` | `any` | `L40` |
| `renderFileSync(file, ctx?, opts?)` | `any` | `L81` |
| `parseAndRenderSync(html, scope?, opts?)` | `any` | `L55` |
| `renderToNodeStream(tpl, scope?, opts?)` | `ReadableStream` | `L43` — streaming via Node.js |
| `renderFileToNodeStream(file, scope?, opts?)` | `Promise<ReadableStream>` | `L84` |
| `evalValueSync(str, scope?)` | `any` | `L96` |
| `parseFile(file, lookupType?)` | `Promise<Template[]>` | `L68` |
| `parseFileSync(file, lookupType?)` | `Template[]` | `L71` |

### Profiling — Exclusivo **[Ruby]** · `lib/liquid/profiler.rb`

| Feature | Notas | Ruby |
|---------|-------|------|
| `Template.parse(source, profile: true)` | Habilita profiling | `template.rb · configure_options(L183)` |
| `template.profiler` | Objeto `Liquid::Profiler` após render; reporta tempo por tag/nó | `profiler.rb · Profiler(L46)` · `Timing(L49)` · `profile_node(L113)` |

### Context Interno

| Feature | Ruby | JS |
|---------|------|----|
| Stack de escopos | `context.rb · scopes(L16)` (array de hashes) | `context.ts · scopes(L9)` (array de objetos) |
| Variáveis do caller | `context.rb · environments(L16)` | `context.ts · environments(L17)` |
| Globals | via `environments` ou `static_environments` | `context.ts · globals(L28)` — separado do environments |
| Registers | `context.rb · registers(L16)` — dois níveis: `static` + `changes` · `lib/liquid/registers.rb` | `context.ts · registers(L10)` — objeto único · `getRegister(L45)` / `setRegister(L48)` |
| Sub-contexto isolado | `context.rb · new_isolated_subcontext(L131)` | `context.ts · spawn(L97)` |
| Push/pop de scope | `context.rb · push(L104)` / `pop(L114)` / `stack(L124)` | `context.ts · push(L88)` / `pop(L91)` |
| Lookup de variável | `context.rb · find_variable(L176)` · `lookup_and_evaluate(L196)` | `context.ts · _get(L69)` · `findScope(L106)` · `readProperty(L112)` |
