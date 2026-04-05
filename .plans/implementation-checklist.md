# Checklist de Implementação — Go Liquid vs Merged Reference

> Comparação entre [go-liquid-reference.md](unchangeable-refs/go-liquid-reference.md) e [merged-liquid-reference.md](unchangeable-refs/merged-liquid-reference.md).
>
> **Colunas de status (nessa ordem: Impl · Tests · E2E):**
>
> | Coluna | Significado |
> |--------|-------------|
> | **Impl** | Implementação concluída (✅ correto · ⚠️ comportamento diferente do spec · ❌ não implementado) |
> | **Tests** | Testes portados das referências (Ruby e/ou JS) passando |
> | **E2E** | Testes intensivos próprios cobrindo a feature (nunca executar automaticamente — só quando o usuário pedir explicitamente) |
>
> **Valores:** `✅` concluído · `⬜` pendente · `➖` não aplicável
>
> **Legenda de prioridade:**
> - **P1** — Core Shopify Liquid (presente em Ruby _e_ JS; qualquer Liquid válido precisa disso)
> - **P2** — Extensão comum (presente em ambos mas não é Shopify core; ex: filtros Jekyll que os dois têm)
> - **P3** — Exclusivo Ruby Liquid
> - **P4** — Exclusivo LiquidJS
> - **P5** — Nice-to-have / low priority
>
> **DECISÃO TOMADA** — itens onde Ruby, JS ou Go divergem e nós já decidimos qual dos comportamentos vai prevalecer aqui na versão Go.
>
> Caso precise consultar onde os recursos citados estão implementados em JS ou Ruby, cheque a [merged-liquid-reference.md](./unchangeable-refs/merged-ruby-js-liquid-reference.md).
> Caso não consiga, sinta-se à vontade para procurar diretamente nos repositórios originais clonados localmente em .example-repositories

---

## 0. Bugs — Correções de comportamento existente

> Esses itens não exigem novas estruturas. Podem ser investigados e corrigidos de forma independente.

### B1 · Tipos numéricos Go em comparações

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `uint64`, `uint32`, `int8`, etc. em `{% if %}` e operadores | `NormalizeNumber()` adicionado em `values/compare.go`: converte todos os tipos inteiros/float do Go para `int64`/`uint64`/`float64` antes de qualquer comparação. `numericCompare()` faz o confronto preciso sem recorrer a float64 para o par int64/uint64, preservando precisão para valores > MaxInt64. `isIntegerType`, `toInt64`, `toFloat64` e `divided_by` em `filters/standard_filters.go` atualizados para incluir `uintptr`. Testes E2E em `b1_numeric_types_test.go` cobrem: todos operadores (`==`,`!=`,`<`,`>`,`<=`,`>=`), `if`/`unless`/`case-when`, condições compostas `and`/`or`, campos de struct com tipo uint, filtros `abs`/`at_least`/`at_most`/`ceil`/`floor`/`round`, cadeia de filtros, `sort`/`where` em arrays mistos, indexação de array com variável uint, `assign`+comparação, `for` com `limit`/`offset` uint, precisão float. Dois bugs adicionais corrigidos: `arrayValue.IndexValue` e `toLoopInt` em `iteration_tags.go` não aceitavam tipos uint. |

### B2 · Truthiness: `nil`, `false`, `blank`, `empty`

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Regras de falsy em `{% if %}` | Implementação verificada e correta. `wrapperValue.Test()` em `values/value.go` usa `v.value != nil && v.value != false`; `if/unless` em `control_flow_tags.go` usa `value != nil && value != false`; `and`/`or`/`not` em `expressions.y` usam `.Test()`. `IsEmpty` e `IsBlank` em `values/predicates.go` são usados apenas para comparações com `empty`/`blank` keyword, não para truthiness geral. `default` filter usa `IsEmpty` corretamente (ativa para `""`, `[]`, `{}`, `nil`, `false`; NÃO ativa para `0` ou strings não-vazias). Testes portados: `TestPortedLiterals_Truthiness`, `TestPortedLiterals_Empty`, `TestPortedLiterals_Blank` em `expressions_ported_test.go` (46 testes). E2E intensivos em `b2_truthiness_test.go` (63 testes) cobrindo: bindings Go tipados, `if`/`unless`/`not`/`and`/`or`, `case/when` com nil/false, filtro `default` com todos edge cases incluindo `allow_false`, filtro `where` sem valor (truthy), comparações com `blank` e `empty` via variáveis, `capture`/`assign`, e chains `elsif`. |

### B3 · Whitespace control em edge cases

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `{%-`/`-%}` e `{{-`/`-}}` em blocos aninhados e loops | **Bug corrigido:** scanner em `parser/scanner.go` não reconhecia `{%- # comment -%}` (espaço entre `-` e `#`) — a regex do comentário inline `{%-?#` foi atualizada para `{%-?\s*#`, permitindo espaço opcional. Isso habilitou também `{% # comment %}` (espaço sem trim). Testes existentes de `TestInlineComment` expandidos com 6 variantes de espaçamento. Behavior do `trimWriter` em loops e blocos aninhados confirmado correto: trim nodes no corpo do `for` se executam por iteração; `TrimTagLeft/Right` globais afetam apenas o contexto externo ao bloco, não o interior das iterações. Testes portados já cobriam os casos Ruby/LiquidJS. E2E intensivos em `b3_whitespace_ctrl_test.go` (38 testes) cobrem: `for` com todas combinações de trim, `for`+`else`, `if` aninhado em `for`, aninhamento duplo, `unless`/`case`/`when` com trim, `assign`/`capture` com trim, comentário inline com espaço (bug corrigido), `{{- -}}` dentro de loops, globais `TrimTagLeft/Right/Both`, `greedy`/`non-greedy`, `liquid` tag com trim, e `raw` com trim markers internos. |

### B4 · Mensagens de erro e tipos

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Tipos distintos de erro (`ParseError`, `RenderError`, `UndefinedVariableError`) | Implementados via swarm PRE-E: `ParseError` em `parser/error.go`, `RenderError` e `UndefinedVariableError` em `render/error.go`. O `UndefinedVariableError` carrega o nome literal da variável. `ZeroDivisionError` também implementado em `filters/standard_filters.go`. **Testes E2E intensivos em `b4_b6_error_test.go`** (55 testes) cobrem: `ParseError` (prefix, `errors.As`, `LineNumber`, `MarkupContext`, `Message`), `RenderError` (prefix, `errors.As`, `LineNumber`, `MarkupContext`, `Cause`), `UndefinedVariableError` (Name, LineNumber, Message, MarkupContext, StrictVariables), `ZeroDivisionError`, `ArgumentError` (filtros + tags + linha + contexto correto), `ContextError`, e toda a suite B6 de preservação de contexto. |

### B5 · Renderer não é seguro para uso concorrente

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | `render.Context` compartilha estado mutável entre chamadas concorrentes | **Investigação concluída em `b5_concurrency_test.go`.** Resultado: o caminho de **render é seguro** para uso concorrente — cada chamada cria seu próprio `nodeContext` com `bindings` isolado; tags com estado (increment, assign, cycle, for-continue) operam apenas no mapa local; expressões compiladas são read-only; `sync.Once` em `Variables()` é thread-safe. **Bug confirmado**: `e.cfg.Cache map[string][]byte` em `render/config.go` não é concorrente-safe — `ParseTemplateAndCache` escreve no mesmo mapa que `{% include %}` lê durante render, causando `fatal error: concurrent map writes` sem precisar de `-race`. **Fix pendente**: substituir `Cache map[string][]byte` por `sync.Map` nos 3 sites (`engine.go:242`, `render/context.go:200`, `render/context.go:234`). **Performance confirmada via benchmarks**: render puro de template compartilhado escala quase linearmente (8.7k→3.2k→2.2 ns/op em 1→4→8 CPUs ✅). Parse sob alta concorrência não escala (27k→21k→26k, plateaus) devido a pressão de alocação GC — há +177 allocs/op por parse vs render puro. **Padrões recomendados** (do mais para menos eficiente): (1) parse uma vez, compartilhe `*Template`, render em N goroutines (~2k ns/op×N); (2) engine compartilhado com cache habilitado (`EnableCache()`) — mesma performance; (3) engine compartilhado sem cache, parse+render por call (~26k ns/op); (4) ❌ engine por goroutine — 6× mais lento (~50k ns/op) por GC overhead de recriar os maps de filtros/grammar. |

### B6 · Mensagens de erro de variável degradadas por indentação e contexto de bloco

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Erros de variável indefinida com mensagens vagas em `{% if %}` e outros blocos | **Bug identificado e corrigido.** Causa raiz: `wrapRenderError` em `render/error.go` re-envolvia qualquer `*RenderError` sem `Path()` mesmo quando ele já tinha `LineNumber > 0`. Isso fazia o `BlockNode` (if/for/unless/case) sobrescrever o `MarkupContext` do nó interno (`{{ expr }}`) com a source do bloco pai (`{% if ... %}`). **Fix:** adicionado `re.LineNumber() > 0` à condição de preservação em `wrapRenderError` — se o erro já tem número de linha, ele veio de um nó mais específico (ObjectNode/TagNode) e deve ser preservado. Templates single-line e multi-line agora produzem mensagens idênticas apontando para o nó exato. Erros em condições de bloco (ex: `{% if x | divided_by: 0 %}`) continuam corretamente atribuídos ao `{% if %}`. Testes intensivos em `b4_b6_error_test.go`. |

---

## 1. Tags

### 1.1 Output / Expressão

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `{{ expressao }}` | OK |
| ✅ | ⬜ | ⬜ | P1 | `echo` tag | `{% echo expr %}` — equivalente a `{{ }}`, mas usável dentro de `{% liquid %}`. Ruby: sempre emite. JS: value opcional (sem value não emite nada). **DECISÃO TOMADA:** seguir Ruby (emissão sempre obrigatória). |
| ✅ | ✅ | ⬜ | P1 | `liquid` tag (multi-linha) | Implementado em `tags/standard_tags.go`. Cada linha não-vazia e não-comentário é compilada como `{%...%}` e renderizada no contexto atual (assign propaga). Linhas com `#` são comentários. Erros de sintaxe propagam em compile-time. Testes em `TestLiquidTag`. |
| ✅ | ✅ | ⬜ | P1 | `#` inline comment | Implementado no scanner (`parser/scanner.go`): padrão `{%-?#(?:...)%}` adicionado à regex de tokenização. Trim markers (`{%-#` e `{%#-%}`) funcionam. Testes em `TestInlineComment`. |

### 1.2 Variável / Estado

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `assign` | OK. Jekyll dot notation (`assign page.prop = v`) também implementado. |
| ✅ | ⬜ | ⬜ | P1 | `capture` | OK. |
| ✅ | ✅ | ⬜ | P1 | `increment` | Implementado em `tags/standard_tags.go`. Contador separado de `assign` e `decrement`. Inicia em 0, emite o valor atual e incrementa. Testes em `TestIncrementDecrement`. |
| ✅ | ✅ | ⬜ | P1 | `decrement` | Implementado em `tags/standard_tags.go`. Contador separado de `assign` e `increment`. Inicia em 0, decrementa e emite o novo valor (primeiro call = -1). Testes em `TestIncrementDecrement`. |

### 1.3 Condicionais

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `if` / `elsif` / `else` / `endif` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `unless` / `else` / `endunless` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `case` / `when` / `else` / `endcase` — `or` em `when` | `when val1 or val2` — suportado. Implementado na gramática yacc (`expressions.y`). |
| ✅ | ✅ | ⬜ | P3 | `ifchanged` | Implementado em `tags/standard_tags.go` via `ifchangedCompiler`. Captura o conteúdo renderizado do bloco e só emite se mudou desde a última chamada. Estado em `"\x00ifchanged_last"`. Testes em `TestIfchangedTag`. |

### 1.4 Iteração

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `for` / `else` / `endfor` com `limit`, `offset`, `reversed`, range | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `for` — ordem de aplicação de modifiers | Ruby: aplica na ordem declarada (offset→limit pode dar resultado diferente de limit→offset). Go: coleta em struct, aplica em ordem fixa. **DECISÃO TOMADA:** deixar com mesmo comportamento do ruby |
| ✅ | ✅ | ⬜ | P4 | `for` — `offset: continue` | Implementado em `tags/iteration_tags.go`. Detectado via regex antes do parsing. TODOS os for-loops rastreiam posição final em `"\x00for_continue_variable-collection"`. Loops com `offset:continue` retomam dali. Testes em `TestOffsetContinue`.
| ✅ | ⬜ | ⬜ | P1 | `break` / `continue` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `cycle` com grupo nomeado | OK. |
| ✅ | ⬜ | ⬜ | P1 | `tablerow` com `cols`, `limit`, `offset`, range | OK. |

### 1.5 Inclusão de templates

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `include` — sintaxe básica `{% include "file" %}` | Implementado e testado. |
| ✅ | ✅ | ⬜ | P1 | `include` — `with var [as alias]` | Implementado em `tags/include_tag.go` com parser dedicado. Testes em `TestIncludeTag_with_variable` e `TestIncludeTag_with_alias`. |
| ✅ | ✅ | ⬜ | P1 | `include` — `key: val` args | Implementado em `tags/include_tag.go` com `parseKVPairs`. Testes em `TestIncludeTag_kv_pairs`. |
| ✅ | ✅ | ⬜ | P3 | `include` — `for array as alias` | Implementado em `tags/include_tag.go`. `{% include 'file' for items as item %}` itera a coleção e renderiza o arquivo uma vez por item com `item` no escopo compartilhado. Testes em `TestIncludeTag_for_array`. |
| ✅ | ✅ | ⬜ | P1 | `render` tag | Implementado em `tags/render_tag.go`. Suporta escopo isolado, `with var [as alias]`, `key: val` args, e `for collection as item`. Testes em `TestRenderTag_*`. |

### 1.6 Estrutura / Texto

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `raw` / `endraw` | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `comment` — nesting | Go: qualquer token ignorado dentro do comment. Ruby: suporta `comment` e `raw` aninhados explicitamente. Comportamento efetivo é idêntico para uso normal. |
| ✅ | ✅ | ⬜ | P3 | `doc` / `enddoc` | Implementado. `c.AddBlock("doc")` em `standard_tags.go` + tratamento especial no parser (`parser/parser.go`) igual a `comment` — o conteúdo interno é completamente ignorado em parse-time. Testes em `TestDocTag`. |
| ✅ | ✅ | ⬜ | P4 | `layout` / `block` | Implementado em `tags/layout_tags.go`. `{% layout 'file' %}...{% endlayout %}` captura blocos filhos e renderiza o layout com overrides. `{% block name %}default{% endblock %}` no filho define override; no layout define slot com fallback. Requer `render/context.go` atualizado para suportar `RenderFile` em block context. Testes em `TestLayoutTag*` e `TestBlockTag_standalone`. |

---

## 2. Filtros

### 2.1 String

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `downcase`, `upcase` | OK. |
| ✅ | ✅ | ⬜ | P1 | `capitalize` | Fix aplicado: primeiro char uppercase + resto lowercase. Testes portados (`"MY GREAT TITLE"` → `"My great title"`). |
| ✅ | ⬜ | ⬜ | P1 | `append`, `prepend` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `remove`, `remove_first`, `remove_last` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `replace`, `replace_first`, `replace_last` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `split` | OK. Trailing empty strings removidas (correto). |
| ✅ | ✅ | ⬜ | P4 | `lstrip`, `rstrip`, `strip` — argumento opcional `chars` | Implementado: cada filtro aceita `chars func(string) string` opcional. Testes portados em `filters/standard_filters_test.go`. |
| ✅ | ✅ | ⬜ | P1 | `strip_html` | Fix aplicado: remove `<script>/<style>` com conteúdo (case-insensitive), comentários HTML `<!-- -->`, depois tags genéricas. Testes portados. |
| ✅ | ⬜ | ⬜ | P1 | `strip_newlines` | OK. |
| ✅ | ✅ | ⬜ | P1 | `newline_to_br` | Fix aplicado: converte `\n` → `<br />\n` (preserva o newline). Testes portados. |
| ✅ | ⬜ | ⬜ | P1 | `truncate`, `truncatewords` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `size`, `slice` | OK. |
| ✅ | ✅ | ⬜ | P3 | `squish` | Implementado em `filters/standard_filters.go`: `strings.TrimSpace` + colapso de whitespace interno. Testes em `filters/standard_filters_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `h` (alias de `escape`) | Implementado. `AddFilter("h", html.EscapeString)` em `filters/standard_filters.go`. Testes portados. |
| ✅ | ⬜ | ⬜ | P4 | `normalize_whitespace` | Presente em Go (Jekyll ext). JS tem, Ruby não. |
| ✅ | ⬜ | ⬜ | P4 | `number_of_words` | Presente em Go (Jekyll ext). |
| ✅ | ⬜ | ⬜ | P4 | `array_to_sentence_string` | Presente em Go (Jekyll ext). |
| ✅ | ⬜ | ⬜ | P4 | `xml_escape` | Presente em Go. |

### 2.2 HTML

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `escape`, `escape_once` | OK. |

### 2.3 URL / Encoding

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `url_encode`, `url_decode` | OK. |
| ✅ | ⬜ | ⬜ | P4 | `cgi_escape`, `uri_escape`, `slugify` | Presentes (Jekyll exts). |
| ✅ | ✅ | ⬜ | P3 | `base64_url_safe_encode`, `base64_url_safe_decode` | Implementado com `encoding/base64.URLEncoding`. Testes portados. |
| ✅ | ⬜ | ⬜ | P1 | `base64_encode`, `base64_decode` | OK. |

### 2.4 Math

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `abs`, `plus`, `minus`, `times`, `ceil`, `floor`, `round` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `at_least`, `at_most` | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `divided_by` — divisão por zero | Go: **retorna erro** (comportamento correto). Ruby: lança `ZeroDivisionError`. JS: comportamento depende de divisão por zero. Semanticamente equivalente (ambos são erros). OK. |
| ⚠️ | ✅ | ⬜ | P1 | `modulo` — divisão por zero | Fix aplicado: guard para zero já implementado (retorna `ZeroDivisionError`). |

### 2.5 Data

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `date` com strftime | OK. |
| ✅ | ✅ | ⬜ | P4 | `date` — `'now'` / `'today'` como input | Implementado em `values/parsedate.go`: `today` trata igual a `now`. Testes em `values/parsedate_test.go`. |
| ✅ | ✅ | ⬜ | P4 | `date_to_xmlschema`, `date_to_rfc822`, `date_to_string`, `date_to_long_string` | Implementados em `filters/standard_filters.go`. `date_to_xmlschema`: formato `%Y-%m-%dT%H:%M:%S%:z`; `date_to_rfc822`: formato `%a, %d %b %Y %H:%M:%S %z`; `date_to_string`/`date_to_long_string`: modo padrão `DD Mon YYYY`, modo `ordinal` com estilos UK/US. Helper `formatJekyllDate()` e `ordinalSuffix()`. Adicionado `"2006-01-02T15:04:05"` (ISO 8601 sem timezone) em `values/parsedate.go`. Testes portados de `liquidjs/test/integration/filters/date.spec.ts` em `filters/standard_filters_test.go`. |

### 2.6 Array

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `join`, `first`, `last`, `reverse`, `sort`, `sort_natural`, `map`, `sum`, `compact`, `uniq`, `concat` | OK. |
| ✅ | ✅ | ⬜ | P3 | `compact` — argumento `property` | Implementado: `compact` aceita `property func(string) string` opcional. Filtra itens onde `item[prop]` é nil. Testes portados. |
| ✅ | ✅ | ⬜ | P3 | `uniq` — argumento `property` | Implementado: `uniq` aceita `property func(string) string` opcional. Deduplica por `item[prop]`. Testes portados. |
| ✅ | ✅ | ⬜ | P1 | `sort` — nil-safe | Fix aplicado: `SortByProperty` chamado com `nilFirst: false` — nils vão para o final como no Ruby. Testes portados. |
| ✅ | ⬜ | ⬜ | P1 | `where`, `reject`, `find`, `find_index`, `has` | OK. |
| ✅ | ⬜ | ⬜ | P4 | `group_by` | Presente em Go. |
| ✅ | ⬜ | ⬜ | P4 | `push`, `pop`, `unshift`, `shift`, `sample` | Presentes em Go. |
| ✅ | ⬜ | ⬜ | P4 | `where_exp`, `reject_exp`, `group_by_exp`, `has_exp`, `find_exp`, `find_index_exp` | Implementados via infraestrutura de `AddContextFilter` (PRE-B). Registrados em `filters/standard_filters.go`. |

### 2.7 Misc

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `default` — keyword arg `allow_false: true` | Implementado: `default` aceita `kwargs ...any` e inspeciona `NamedArg{Name: "allow_false"}`. Testes portados. |
| ✅ | ⬜ | ⬜ | P4 | `json`, `inspect`, `to_integer` | Presentes em Go. |
| ✅ | ✅ | ⬜ | P4 | `jsonify` (alias de `json`) | Implementado. `AddFilter("jsonify", ...)` em `filters/standard_filters.go`. Testes portados. |
| ✅ | ✅ | ⬜ | P4 | `raw` filter | Implementado em `expressions/filters.go` (registrado junto com `safe` em `AddSafeFilter`). `NewConfig()` agora sempre chama `AddSafeFilter` — `raw` e `safe` estão sempre disponíveis, com ou sem autoescape. Também registrado em `filters/standard_filters.go` para contextos de filtro padrão. Quando autoescape está desabilitado, `raw` envolve em `SafeValue` que é imediatamente transparente no render. Testes portados de LiquidJS `output-escape.spec.ts` em `render/autoescape_test.go`. |

---

## 3. Sistema de Filtros

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | Filtros posicionais | OK. |
| ✅ | ✅ | ⬜ | P1 | **Keyword args em filtros** (`filter: arg, key: val`) | Infraestrutura implementada (PRE-A). `NamedArg` struct em `expressions/filters.go`, `makeNamedArgFn` em `builders.go`, gramática atualizada. Filtro `default` atualizado para aceitar `allow_false: true`. Testes portados em `filters/standard_filters_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `global_filter` — proc aplicada a todo output | Implementado via `Engine.SetGlobalFilter(fn func(any) (any, error))`. A função é aplicada ao valor avaliado de cada `{{ }}` antes de ser escrito. Análogo a Ruby's `global_filter` option. Testes em `engine_test.go` (TestEngine_SetGlobalFilter). |

---

## 4. Expressões / Literais

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `nil`, `true`, `false`, int, float, string, range | OK. Range agora tem `String()` que retorna `"start..end"` (Ruby compat). Testes portados em `expressions_ported_test.go`. |
| ✅ | ⬜ | ⬜ | P1 | **`empty` como literal especial** | Implementado. Scanner reconhece `empty` como keyword (`EMPTY` token). `values.EmptyDrop` singleton com comparação simétrica em `values/compare.go`. |
| ✅ | ⬜ | ⬜ | P1 | **`blank` como literal especial** | Implementado. Scanner reconhece `blank` como keyword. `values.BlankDrop` singleton; `IsBlank` cobre nil, false, string-só-whitespace, arrays/maps vazios. |
| ✅ | ⬜ | ⬜ | P1 | `<>` como alias de `!=` | Implementado em `expressions/scanner.rl`. |
| ✅ | ✅ | ⬜ | P4 | `not` operador unário | Fix: gramática atualizada para `cond AND cond` / `cond OR cond` (antes era `cond AND rel`). `not x or not y` agora parseia corretamente. AND/OR são `%right` mesma precedência (right-to-left). Testes portados em `expressions_ported_test.go`. |
| ✅ | ✅ | ⬜ | P1 | Strings — escapes internos (`\n`, `\"`, etc.) | Implementado via `unescapeString()` em `expressions/scanner.rl`. Suporta `\n`, `\t`, `\r`, `\"`, `\'`. Testes portados em `expressions_ported_test.go`. |

---

## 5. Acesso a Variáveis

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `obj.prop`, `obj[key]`, `array[0]` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `array[-1]` — índice negativo | Suportado. `IndexValue(-1)` retorna o último elemento (confirmado em `values/value_test.go`). |
| ✅ | ⬜ | ⬜ | P1 | `array.first`, `array.last`, `obj.size` | OK. |

---

## 6. Drops (Objetos Especiais)

### 6.1 ForloopDrop

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `index`, `index0`, `rindex`, `rindex0`, `first`, `last`, `length` | OK. |
| ✅ | ✅ | ⬜ | P1 | **`forloop.name`** | Já implementado em `tags/iteration_tags.go` via `loopName(args, variable)`. Retorna `"variavel-colecao"`. Testes em `TestForloopMeta`. |
| ✅ | ✅ | ⬜ | P3 | `forloop.parentloop` | Já implementado em `tags/iteration_tags.go` — salva o `forloopMap` do pai antes de iniciar o loop filho. Testes em `TestForloopMeta`. |

### 6.2 TablerowloopDrop

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `row`, `col`, `col0`, `col_first`, `col_last` | Já implementado em `tags/iteration_tags.go` via `tableRowDecorator`. Campos `col` (1-indexed), `col0` (0-indexed), `col_first`, `col_last`, `row` expostos no `forloop` do tablerow. Testes em `TestTablerowLoopVars`. |

### 6.3 EmptyDrop / BlankDrop

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | **`empty` drop/literal** | `values.EmptyDrop` exportado em `values/emptydrop.go`. |
| ✅ | ⬜ | ⬜ | P1 | **`blank` drop/literal** | `values.BlankDrop` exportado em `values/emptydrop.go`. |

### 6.4 Drop base class

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | Interface `Drop` (`ToLiquid() any`) | Protocolo simples implementado. |
| ✅ | ✅ | ⬜ | P3 | Drop base class com `liquid_method_missing` | `DropMethodMissing` interface em `drops.go` + `values/drop.go`; integrado em `values/structvalue.go`. Testes portados Ruby/JS em `drops_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `context=` injection no drop | Interface `ContextDrop` (alias `values.ContextSetter`) + `DropRenderContext` (alias `values.ContextAccess`) em `drops.go`. `expressions/context.go: Get()` injeta contexto antes de qualquer acesso a propriedade. Testes em `drops_test.go` (TestContextDrop_*, ExampleContextDrop). |

---

## 7. Context / Escopo

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | Stack de escopos, get/set variáveis | OK. |
| ✅ | ⬜ | ⬜ | P1 | **Sub-contexto isolado** | Implementado. `nodeContext.SpawnIsolated(bindings)` em `render/node_context.go` — cria contexto novo sem herdar variáveis do pai; globals propagam. |
| ✅ | ⬜ | ⬜ | P1 | Registers (estado interno de tags) | OK (map acessível via contexto). |
| ✅ | ⬜ | ⬜ | P2 | **Variáveis globais separadas do escopo** (`globals`) | Implementado. `Config.Globals` copiados antes dos bindings em `newNodeContext` e `SpawnIsolated`. `Engine.SetGlobals`/`GetGlobals` expostos em `engine.go`. |

---

## 8. Configuração / Engine

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `StrictVariables()` | OK (engine-level). Testes em `engine_test.go`. |
| ✅ | ✅ | ⬜ | P1 | `LaxFilters()` | OK. Testes em `engine_test.go`. |
| ✅ | ✅ | ⬜ | P1 | Custom delimiters (`Delims()`) | OK. Testes em `engine_test.go`. |
| ✅ | ✅ | ⬜ | P1 | Custom `TemplateStore` | OK. Testes em `engine_test.go`. |
| ✅ | ✅ | ⬜ | P1 | `RegisterTag`, `RegisterBlock`, `RegisterFilter` | OK. Testes em `engine_test.go`. |
| ✅ | ✅ | ⬜ | P2 | `strict_variables` / `strict_filters` — **por render, não por engine** | `WithStrictVariables()`, `WithLaxFilters()` em `liquid.go`. Aceitos por todos os métodos de render. Testes portados de LiquidJS `strict.spec.ts` e Ruby `template_test.rb` em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P2 | `globals` **por render** (`WithGlobals`) | `WithGlobals(map[string]any)` em `liquid.go`. Portado de LiquidJS `liquid.spec.ts`. Testes em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `error_mode` (`:lax` para tags) | `Engine.LaxTags()` — unknown tags compilam como no-ops. Testes em `engine_section8_test.go`. `:warn`/`:strict2` não implementados (requerem continue-on-error). |
| ✅ | ✅ | ⬜ | P3 | `template.errors` / coleta de erros | Via `WithErrorHandler`: acumular erros while-rendering é o padrão Go. Testes em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `exception_renderer` / `exception_handler` | `WithErrorHandler(func(error) string)` + `Engine.SetExceptionHandler()`. Portado de Ruby `template_test.rb`. Testes em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P3 | Resource limits (`render_length_limit`) | `WithSizeLimit(int64)` — aborta quando output excede N bytes. Portado de Ruby `test_resource_limits_render_length`. Testes em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P4 | Resource limits (time-based: `renderLimit`) | `WithContext(context.Context)` — render para quando context cancela/expira. Portado de LiquidJS `dos` concept. Testes em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P4 | Template cache | `Engine.EnableCache()` + `ClearCache()` — sync.Map keyed por source string. Testes em `engine_section8_test.go`. |
| ✅ | ✅ | ⬜ | P4 | `globals` option no engine | `Engine.SetGlobals` / `GetGlobals()`. Testes em `engine_section8_test.go`. |

---

## 9. Análise Estática

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P2 | `GlobalVariableSegments`, `VariableSegments`, `GlobalFullVariables`, `FullVariables` | Testes portados de Ruby (`parse_tree_visitor_test.rb`) e LiquidJS (`variables.spec.ts`, `parse-and-analyze.spec.ts`) em `analysis_ported_test.go`. |
| ✅ | ✅ | ⬜ | P2 | `Analyze()` / `ParseAndAnalyze()` | Testes portados de LiquidJS em `analysis_ported_test.go`. |
| ✅ | ✅ | ⬜ | P2 | `RegisterTagAnalyzer`, `RegisterBlockAnalyzer` | Teste básico em `analysis_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `ParseTreeVisitor` API visitor-style | Implementado via `Walk(WalkFunc)` e `ParseTree() *TemplateNode` em `visitor.go`. Tipos públicos: `TemplateNodeKind` (Text/Output/Tag/Block), `TemplateNode` (Kind, TagName, Location, Children), `WalkFunc`. Testes em `visitor_test.go` portados de `parse_tree_visitor_test.rb` (tree structure, skip children, all node kinds, tag names, source locations). |

---

## 10. Tratamento de Erros

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `SourceError` com `Path()`, `LineNumber()`, `Cause()` | OK. `Message()` e `MarkupContext()` adicionados à interface `parser.Error` e `render.Error`. Testes portados em `error_handling_ported_test.go`. |
| ✅ | ✅ | ⬜ | P3 | `ZeroDivisionError` tipo específico | Implementado em `filters/standard_filters.go`. Tipo exportado retornado por `divided_by` e `modulo`. Testes em `filters/standard_filters_test.go` e `engine_test.go`. |
| ✅ | ✅ | ⬜ | P3 | Tipos específicos de erro (`SyntaxError`, `ArgumentError`, `ContextError`, etc.) | `SyntaxError` = type alias de `ParseError` (em `parser/error.go`). `ArgumentError` e `ContextError` adicionados em `render/error.go` como tipos simples (leaf errors) detectáveis via `errors.As`. `ParseError.Error()` usa prefixo `"Liquid syntax error"`, `RenderError.Error()` usa `"Liquid error"`. Testes portados em `error_handling_ported_test.go`. |
| ✅ | ✅ | ⬜ | P1 | Metadados de erro — `markup_context` | `MarkupContext()` adicionado às interfaces `parser.Error` e `render.Error`. Retorna o texto-fonte do token que causou o erro (ex: `{% tag args %}`). Quando não há pathname, o contexto de markup aparece no `Error()` como locativo. `Message()` retorna só a mensagem sem prefixo/localização. Testes portados em `error_handling_ported_test.go`. |

---

## 11. Whitespace Control

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ⬜ | P1 | `{%-`, `-%}`, `{{-`, `-}}` | OK. Testes portados em `whitespace_ctrl_ported_test.go` (42 testes cobrindo Ruby e JS). |
| ✅ | ✅ | ⬜ | P4 | Opções globais de trim (`trimTagRight`, etc.) | Implementado: `Config.TrimTagLeft/Right`, `TrimOutputLeft/Right`, `Greedy` em `render/config.go`. Engine expõe `SetTrimTagLeft/Right`, `SetTrimOutputLeft/Right`, `SetGreedy`. `Greedy` padrão = true. Non-greedy (inline blank + 1 newline) implementado em `trimwriter.go`. Testes portados de `trimming.spec.ts` passando. |

---

## 12. Thread-safety e Concorrência

> Não faz sentido garantir imutabilidade antes de ter todos os campos de configuração definidos. Pode ser planejado em paralelo, mas implementado depois de estabilizar a API de configuração.
> Ver também **B5** (bug ativo de race condition no renderer).

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ✅ | ✅ | P1 | Auditoria de estado mutável no `Engine` | **Concluída.** Grammar maps (`tags`, `blockDefs`), filter maps e `Config.Globals` são escritos só no setup e lidos durante render — race-free. `Expression.Variables()` usa `sync.Once` — correto. `engine.cache *sync.Map` para template cache é thread-safe. `Cache` (fallback do `{% include %}`) era `map[string][]byte` — corrigido para `sync.Map` (ver linha abaixo). Engine está 100% seguro para uso concorrente após setup. |
| ✅ | ✅ | ✅ | P1 | Estado de render isolado por chamada | **Confirmado seguro.** `newNodeContext(vars, cfg)` faz `maps.Copy` dos globals+scope para um mapa novo a cada call. Tags com estado (assign, increment, decrement, cycle, for+continue) operam somente no mapa per-call. Expressões compiladas são imutáveis após parse. Verificado em `TestConcurrent_StatefulTagsAreIsolated`. |
| ✅ | ✅ | ✅ | P2 | `Config` imutável após construção | **Implementado via freeze pattern.** `Engine` tem `frozen atomic.Bool`. `freeze()` é chamado no início de todo entry point de parse (`ParseTemplate`, `ParseTemplateLocation`, `ParseString`, `ParseAndRender`, `ParseAndFRender`, `ParseTemplateAndCache`). `checkNotFrozen(method)` é chamado em todos os 18 métodos de configuração mutantes (`RegisterTag/Block/Filter`, `StrictVariables`, `LaxFilters/Tags`, `SetGlobals`, `SetTrimXxx`, `SetGreedy`, `SetGlobalFilter`, `SetExceptionHandler`, `SetAutoEscapeReplacer`, `Delims`, `EnableCache`, `EnableJekyllExtensions`, `RegisterTagAnalyzer/BlockAnalyzer`). Violação resulta em panic com mensagem clara: `"liquid: SetGlobals() called after the engine has been used for parsing"`. Zero overhead no hot path. Exceção documentada: `UnregisterTag` não tem guard — é explicitamente para hot-reload/test teardown. 3 testes em `context_scope_ported_test.go` tinham `RegisterTag` após `ParseTemplateAndCache` — corrigidos para a ordem certa. 36 subtestes em `TestEngine_FrozenAfterParse` + `TestEngine_FrozenPanicMessage` cobrem todos os métodos. |
| ✅ | ✅ | ✅ | P1 | Fix: `Cache map[string][]byte` → `sync.Map` | **Corrigido.** `render/config.go`: campo `Cache` trocado para `sync.Map`. `engine.go`: `Cache[path] = source` → `Cache.Store(path, source)`. `render/context.go`: dois `Cache[filename]` → `Cache.Load(filename)` (com type assertion `.([]byte)`). `tags/include_tag_test.go`: dois `config.Cache["..."] = []byte(...)` → `Cache.Store(...)`. `NewConfig()`: removida inicialização `Cache: map[string][]byte{}` (zero value de `sync.Map` já é válida). `TestConcurrent_CacheRace` agora testa o comportamento real (sem `t.Skip`) — passou. |

---

## Resumo Executivo por Prioridade

### P1 — Core Shopify Liquid (implementar primeiro)

```
Tags:
[x] echo tag                 ✅ DONE
[x] liquid tag (multi-linha) ✅ DONE — `tags/standard_tags.go`, testes em `TestLiquidTag`
[x] # inline comment         ✅ DONE — `parser/scanner.go`, testes em `TestInlineComment`
[x] increment / decrement    ✅ DONE — `tags/standard_tags.go`, contadores separados, testes em `TestIncrementDecrement`
[x] render tag (escopo isolado) ✅ DONE — `tags/render_tag.go`, with/as/kv/for, testes em `TestRenderTag_*`
[x] include — with/as/key-val args ✅ DONE — `tags/include_tag.go` reescrito, testes em `TestIncludeTag_*`
[x] case/when — suporte a `or`  ✅ DONE

Filtros:
[x] capitalize — fix (lowercase resto)          ✅ DONE
[x] strip_html — fix (remover script/style)     ✅ DONE
[x] newline_to_br — fix (preservar \n)          ✅ DONE
[x] modulo — fix (erro em divisão por zero)     ✅ DONE (guard adicionado)
[x] default — allow_false keyword arg           ✅ DONE (filtro atualizado + testes)
[x] sort — nil-last (nils vão para o final)     ✅ DONE
[x] Keyword args em filtros (parser change)     ✅ DONE (infraestrutura NamedArg)

Expressões:
[x] empty literal/drop        ✅ DONE
[x] blank literal/drop        ✅ DONE
[x] Strings — suporte a escapes (\n, \", etc.) ✅ DONE
[x] array[-1] negative indexing               ✅ DONE

Drops:
[x] forloop.name              ✅ DONE (já estava implementado — confirmado)
[x] tablerowloop drop — row/col/col0/col_first/col_last ✅ DONE (já estava implementado — confirmado)

Context:
[x] Sub-contexto isolado (para render tag) ✅ DONE
[x] Variáveis globais separadas do escopo  ✅ DONE
```

### P2 — Extensões Comuns (Ruby + JS)

```
[x] strict_variables / strict_filters como opção per-render  ✅ DONE — WithStrictVariables(), WithLaxFilters(), WithGlobals() em liquid.go
[x] globals option no engine  ✅ DONE
```

### P3 — Compat Ruby

```
[x] squish filtro              ✅ DONE
[x] h alias (escape)           ✅ DONE
[x] base64_url_safe_encode/decode  ✅ DONE
[x] compact: property arg      ✅ DONE
[x] uniq: property arg         ✅ DONE
[x] forloop.parentloop        ✅ DONE (já estava implementado — confirmado)
[x] <> alias de !=   ✅ DONE
[x] doc / enddoc tag  ✅ DONE — parser especial igual a comment, testes em TestDocTag
[x] ifchanged tag     ✅ DONE — `tags/standard_tags.go`, testes em TestIfchangedTag
[x] include for array as alias ✅ DONE — `tags/include_tag.go`, testes em TestIncludeTag_for_array
[x] Drop: liquid_method_missing  ✅ DONE — `DropMethodMissing` em `drops.go`, testes em `drops_test.go`
[x] context= injection no drop  ✅ DONE — `ContextDrop`/`DropRenderContext` em `drops.go`, `expressions/context.go`, testes em `drops_test.go`
[x] template.errors / coleta de erros  ✅ DONE — WithErrorHandler() como collector
[x] exception_renderer  ✅ DONE — WithErrorHandler() + Engine.SetExceptionHandler()
[x] Resource limits (render_length)  ✅ DONE — WithSizeLimit(int64)
[x] ParseTreeVisitor API  ✅ DONE — Walk + ParseTree em visitor.go
```

### P4 — Compat JS / Extensões

```
[x] for offset: continue  ✅ DONE — `tags/iteration_tags.go`, todos os loops rastreiam posição, testes em TestOffsetContinue
[x] date: 'now'/'today' como input  ✅ DONE
[x] date_to_xmlschema / date_to_rfc822 / date_to_string / date_to_long_string  ✅ DONE — `filters/standard_filters.go`, testes portados JS em `filters/standard_filters_test.go`
[x] where_exp / reject_exp / group_by_exp / has_exp / find_exp / find_index_exp  ✅ DONE
[x] jsonify alias              ✅ DONE
[x] raw filter  ✅ DONE — `expressions/filters.go` (registrado junto com `safe`), `render/config.go` (sempre habilitado), testes em `render/autoescape_test.go`
[x] layout / block tags        ✅ DONE — `tags/layout_tags.go`, herança por bloco, testes em TestLayoutTag*
[x] not operador unário       ✅ DONE
[ ] Opções globais de whitespace trim
[x] Resource limits (time-based via context)  ✅ DONE — WithContext(context.Context)
[x] Template cache  ✅ DONE — Engine.EnableCache() / ClearCache()
```
