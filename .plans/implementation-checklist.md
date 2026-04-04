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
> Caso precise consultar onde os recursos citados estão implementados em JS ou Ruby, cheque a [merged-liquid-reference.md](merged-liquid-reference.md).
> Caso não consiga, sinta-se à vontade para procurar diretamente nos repositórios originais clonados localmente em .example-repositories

---

## 0. Bugs — Correções de comportamento existente

> Esses itens não exigem novas estruturas. Podem ser investigados e corrigidos de forma independente.

### B1 · Tipos numéricos Go em comparações

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ❌ | ⬜ | ⬜ | P1 | `uint64`, `uint32`, `int8`, etc. em `{% if %}` e operadores | Tipos inteiros não-padrão do Go causam comportamento incorreto em comparações. A conversão existe para filtros, mas não está garantida no avaliador de expressões. Verificar `expressions/` e `values/compare.go`. |

### B2 · Truthiness: `nil`, `false`, `blank`, `empty`

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ❌ | ⬜ | ⬜ | P1 | Regras de falsy em `{% if %}` | Em Liquid, apenas `nil` e `false` são falsy; todo o resto (incluindo `0`, `""`) é truthy. Verificar se `values/predicates.go` e `render/context.go` respeitam isso. O comportamento de `blank` e `empty` como palavras-chave em `{% if x == blank %}` também precisa de validação. |

### B3 · Whitespace control em edge cases

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ❌ | ⬜ | ⬜ | P1 | `{%-`/`-%}` e `{{-`/`-}}` em blocos aninhados e loops | Os marcadores de whitespace podem ter comportamento incorreto em casos como blocos aninhados, loops e templates com `include`. Validar contra o Golden Liquid test suite. |

### B4 · Mensagens de erro e tipos

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | Tipos distintos de erro (`ParseError`, `RenderError`, `UndefinedVariableError`) | Implementados via swarm PRE-E: `ParseError` em `parser/error.go`, `RenderError` e `UndefinedVariableError` em `render/error.go`. O `UndefinedVariableError` carrega o nome literal da variável. `ZeroDivisionError` também implementado em `filters/standard_filters.go`. |

### B5 · Renderer não é seguro para uso concorrente

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ❌ | ⬜ | ⬜ | P1 | `render.Context` compartilha estado mutável entre chamadas concorrentes | Atualmente é necessário instanciar um novo renderer (ou contexto de render) por goroutine para evitar race conditions. Isso gera um gargalo de processamento alto em uso concorrente. **Causa raiz ainda não identificada** — suspeita de estado mutável em `render/context.go` ou `nodeContext` compartilhado entre chamadas. Investigar com `go test -race`. Ver também seção 12. |

### B6 · Mensagens de erro de variável degradadas por indentação e contexto de bloco

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ❌ | ⬜ | ⬜ | P1 | Erros de variável indefinida com mensagens vagas em `{% if %}` e outros blocos | A indentação do template impacta a mensagem de erro retornada (provavelmente o texto do markup capturado inclui whitespace acidental). Além disso, variáveis indefinidas dentro de blocos `{% if %}` e similares às vezes produzem mensagens genéricas demais, sem citar o nome literal da variável. **Aguardando exemplos concretos do usuário para reprodução.** |

---

## 1. Tags

### 1.1 Output / Expressão

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `{{ expressao }}` | OK |
| ✅ | ⬜ | ⬜ | P1 | `echo` tag | `{% echo expr %}` — equivalente a `{{ }}`, mas usável dentro de `{% liquid %}`. Ruby: sempre emite. JS: value opcional (sem value não emite nada). **DECISÃO TOMADA:** seguir Ruby (emissão sempre obrigatória). |
| ❌ | ⬜ | ⬜ | P1 | `liquid` tag (multi-linha) | `{% liquid\nassign x = 1\nif x %}...{% endif %}` — cada linha é uma tag sem delimitadores. **`echo` já está pronta em `tags/standard_tags.go`** — este item só precisa do mecanismo de parse linha a linha dentro do bloco `liquid`. |
| ❌ | ⬜ | ⬜ | P1 | `#` inline comment | `{%# comentário %}` — cada linha precisa de `#`. Ambos (Ruby e JS) têm com semântica idêntica. |

### 1.2 Variável / Estado

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `assign` | OK. Jekyll dot notation (`assign page.prop = v`) também implementado. |
| ✅ | ⬜ | ⬜ | P1 | `capture` | OK. |
| ❌ | ⬜ | ⬜ | P1 | `increment` | `{% increment var %}` — armazenado em escopo separado (não conflita com `assign`); compartilha slot com `decrement`; output pré-incremento. Ruby e JS têm comportamento idêntico. |
| ❌ | ⬜ | ⬜ | P1 | `decrement` | `{% decrement var %}` — starts at -1; Ruby: output-then-decrement; JS: pre-decrement-then-output (resultado externo igual). **DECISÃO TOMADA:** seguir spec: armazenar contador em namespace separado dos `assign`. |

### 1.3 Condicionais

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `if` / `elsif` / `else` / `endif` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `unless` / `else` / `endunless` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `case` / `when` / `else` / `endcase` — `or` em `when` | `when val1 or val2` — suportado. Implementado na gramática yacc (`expressions.y`). |
| ❌ | ⬜ | ⬜ | P3 | `ifchanged` | Ruby only. Renderiza só se output mudou desde a última iteração dentro de `for`. Estado interno em `registers`. |

### 1.4 Iteração

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `for` / `else` / `endfor` com `limit`, `offset`, `reversed`, range | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `for` — ordem de aplicação de modifiers | Ruby: aplica na ordem declarada (offset→limit pode dar resultado diferente de limit→offset). Go: coleta em struct, aplica em ordem fixa. **DECISÃO TOMADA:** deixar com mesmo comportamento do ruby |
| ❌ | ⬜ | ⬜ | P4 | `for` — `offset: continue` | JS only. Retoma do ponto onde o último `for` sobre a mesma coleção parou. Baixo impacto para Shopify compat. |
| ✅ | ⬜ | ⬜ | P1 | `break` / `continue` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `cycle` com grupo nomeado | OK. |
| ✅ | ⬜ | ⬜ | P1 | `tablerow` com `cols`, `limit`, `offset`, range | OK. |

### 1.5 Inclusão de templates

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ⚠️ | ⬜ | ⬜ | P1 | `include` — sintaxe básica `{% include "file" %}` | Implementado, mas **sintaxe incompleta** (ver abaixo). |
| ❌ | ⬜ | ⬜ | P1 | `include` — `with var [as alias]` | `{% include 'file' with product %}` / `with product as p`. Presente em Ruby e JS. Parser da tag em `tags/include_tag.go`. |
| ❌ | ⬜ | ⬜ | P1 | `include` — `key: val` args | `{% include 'file' key: value, other: x %}` — passa variáveis adicionais. Presente em Ruby e JS. **A infraestrutura `NamedArg` (PRE-A) já parseia keyword args em expressões; reutilizável aqui ou parsear diretamente no handler da tag em `tags/include_tag.go`.** |
| ❌ | ⬜ | ⬜ | P3 | `include` — `for array as alias` | Ruby-only (deprecated). `{% include 'file' for items as item %}` — itera sobre array. |
| ❌ | ⬜ | ⬜ | P1 | `render` tag | `{% render 'file' [with var [as alias]] [for collection [as alias]] [key: val...] %}` — **escopo isolado** (não acessa variáveis do pai). Ambos Ruby e JS. **Infraestrutura pronta:** `nodeContext.SpawnIsolated()` em `render/node_context.go` + globals propagando via `Config.Globals`. Resta implementar parser e renderer da tag em `tags/`. |

### 1.6 Estrutura / Texto

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `raw` / `endraw` | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `comment` — nesting | Go: qualquer token ignorado dentro do comment. Ruby: suporta `comment` e `raw` aninhados explicitamente. Comportamento efetivo é idêntico para uso normal. |
| ❌ | ⬜ | ⬜ | P3 | `doc` / `enddoc` | Ruby-only. LiquidDoc: ignorado no render. |
| ❌ | ⬜ | ⬜ | P4 | `layout` / `block` | JS-only. Herança de template. Fora do escopo Shopify Liquid. |

---

## 2. Filtros

### 2.1 String

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `downcase`, `upcase` | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `capitalize` | Go: **só faz uppercase na primeira letra**. Ruby/JS: **primeira maiúscula + resto em minúsculas**. Quick fix: `strings.ToLower(rest)` no restante. |
| ✅ | ⬜ | ⬜ | P1 | `append`, `prepend` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `remove`, `remove_first`, `remove_last` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `replace`, `replace_first`, `replace_last` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `split` | OK. Trailing empty strings removidas (correto). |
| ⚠️ | ⬜ | ⬜ | P4 | `lstrip`, `rstrip`, `strip` — argumento opcional `chars` | JS aceita `{{ str \| strip: "abc" }}` para strip de conjunto de chars. Go e Ruby não têm. **DECISÃO TOMADA:** executar (mesmo não sendo Shopify core). |
| ⚠️ | ⬜ | ⬜ | P1 | `strip_html` | Go usa regex simples `<.*?>`. Ruby/JS **também removem `<script>`, `<style>` e comentários HTML** (`<!-- -->`). Comportamento diferente para templates com script/style. Fix: melhorar a regex para cobrir esses casos. |
| ✅ | ⬜ | ⬜ | P1 | `strip_newlines` | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `newline_to_br` | Go converte `\n` → `<br />`. Ruby/JS convertem `\n` → `<br />\n` (**preserva o newline depois do `<br />`**). Quick fix. |
| ✅ | ⬜ | ⬜ | P1 | `truncate`, `truncatewords` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `size`, `slice` | OK. |
| ❌ | ⬜ | ⬜ | P3 | `squish` | Ruby-only. Strip + colapsa whitespace interno em espaço único. Go tem `normalize_whitespace` (JS-inspired) que faz collapse mas não strip. **DECISÃO TOMADA:** adicionar `squish` como alias de `strip \| normalize_whitespace`. |
| ❌ | ⬜ | ⬜ | P3 | `h` (alias de `escape`) | Ruby-only. Trivial de adicionar: `AddFilter("h", escapeFilter)`. |
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
| ❌ | ⬜ | ⬜ | P3 | `base64_url_safe_encode`, `base64_url_safe_decode` | Ruby-only. Fácil de adicionar com `encoding/base64.URLEncoding`. |
| ✅ | ⬜ | ⬜ | P1 | `base64_encode`, `base64_decode` | OK. |

### 2.4 Math

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `abs`, `plus`, `minus`, `times`, `ceil`, `floor`, `round` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `at_least`, `at_most` | OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `divided_by` — divisão por zero | Go: **retorna erro** (comportamento correto). Ruby: lança `ZeroDivisionError`. JS: comportamento depende de divisão por zero. Semanticamente equivalente (ambos são erros). OK. |
| ⚠️ | ⬜ | ⬜ | P1 | `modulo` — divisão por zero | Go usa `math.Mod` — **não lança erro para zero**, retorna `NaN`/`Inf`. Ruby/JS levantam erro. **DECISÃO TOMADA:** adicionar guard para zero. |

### 2.5 Data

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `date` com strftime | OK. |
| ❌ | ⬜ | ⬜ | P4 | `date` — `'now'` / `'today'` como input | JS suporta string `'now'`/`'today'` que mapeia para hora atual. Ruby não. **DECISÃO TOMADA:** adicionar para paridade com JS. |
| ❌ | ⬜ | ⬜ | P4 | `date_to_xmlschema`, `date_to_rfc822`, `date_to_string`, `date_to_long_string` | JS-only (Jekyll). |

### 2.6 Array

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `join`, `first`, `last`, `reverse`, `sort`, `sort_natural`, `map`, `sum`, `compact`, `uniq`, `concat` | OK. |
| ⚠️ | ⬜ | ⬜ | P3 | `compact` — argumento `property` | Ruby suporta `compact: "field"`. **`NamedArg` pronto** (PRE-A). Atualizar `compactFilter` em `filters/standard_filters.go` para aceitar propriedade opcional e checar nil via `values.PropertyValue`. |
| ⚠️ | ⬜ | ⬜ | P3 | `uniq` — argumento `property` | Ruby suporta `uniq: "field"`. **`NamedArg` pronto** (PRE-A). Atualizar `uniqFilter` em `filters/sort_filters.go` para aceitar propriedade opcional e deduplicar por `values.PropertyValue`. |
| ⚠️ | ⬜ | ⬜ | P1 | `sort` — nil-safe | Ruby: nils vão para o final (nil-safe). Go: comportamento não verificado — pode panic se nil presente. **Verificar e corrigir se necessário, copiando o comportamento do Ruby.** |
| ✅ | ⬜ | ⬜ | P1 | `where`, `reject`, `find`, `find_index`, `has` | OK. |
| ✅ | ⬜ | ⬜ | P4 | `group_by` | Presente em Go. |
| ✅ | ⬜ | ⬜ | P4 | `push`, `pop`, `unshift`, `shift`, `sample` | Presentes em Go. |
| ✅ | ⬜ | ⬜ | P4 | `where_exp`, `reject_exp`, `group_by_exp`, `has_exp`, `find_exp`, `find_index_exp` | Implementados via infraestrutura de `AddContextFilter` (PRE-B). Registrados em `filters/standard_filters.go`. |

### 2.7 Misc

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ⚠️ | ⬜ | ⬜ | P1 | `default` — keyword arg `allow_false: true` | Ambos Ruby e JS suportam. **Infraestrutura `NamedArg` pronta** (PRE-A): o parser já entrega `NamedArg{Name: "allow_false", Value: true}` nos args do filtro. Basta atualizar a assinatura de `default` em `filters/standard_filters.go` para inspecionar `NamedArg` e alterar o comportamento de falsy quando `allow_false: true`. |
| ✅ | ⬜ | ⬜ | P4 | `json`, `inspect`, `to_integer` | Presentes em Go. |
| ❌ | ⬜ | ⬜ | P4 | `jsonify` (alias de `json`) | JS-only. Trivial de adicionar. |
| ❌ | ⬜ | ⬜ | P4 | `raw` filter | JS-only. Passa valor sem escape. |

---

## 3. Sistema de Filtros

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | Filtros posicionais | OK. |
| ✅ | ⬜ | ⬜ | P1 | **Keyword args em filtros** (`filter: arg, key: val`) | Infraestrutura implementada (PRE-A). `NamedArg` struct em `expressions/filters.go`, `makeNamedArgFn` em `builders.go`, gramática atualizada. Os filtros individuais que dependem de keyword args (ex: `default allow_false`) ainda precisam ser atualizados pra consumir o `NamedArg`. |
| ❌ | ⬜ | ⬜ | P3 | `global_filter` — proc aplicada a todo output | Ruby-only. Aplicada antes de renderizar qualquer `{{ }}`. Go tem `SetAutoEscapeReplacer` que é o análogo, mas não é um filtro Liquid. |

---

## 4. Expressões / Literais

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `nil`, `true`, `false`, int, float, string, range | OK. |
| ✅ | ⬜ | ⬜ | P1 | **`empty` como literal especial** | Implementado. Scanner reconhece `empty` como keyword (`EMPTY` token). `values.EmptyDrop` singleton com comparação simétrica em `values/compare.go`. |
| ✅ | ⬜ | ⬜ | P1 | **`blank` como literal especial** | Implementado. Scanner reconhece `blank` como keyword. `values.BlankDrop` singleton; `IsBlank` cobre nil, false, string-só-whitespace, arrays/maps vazios. |
| ✅ | ⬜ | ⬜ | P1 | `<>` como alias de `!=` | Implementado em `expressions/scanner.rl`. |
| ✅ | ⬜ | ⬜ | P4 | `not` operador unário | Implementado em `expressions/expressions.y` como operador `NOT` de precedência unária. |
| ✅ | ⬜ | ⬜ | P1 | Strings — escapes internos (`\n`, `\"`, etc.) | Implementado via `unescapeString()` em `expressions/scanner.rl`. Suporta `\n`, `\t`, `\r`, `\"`, `\'`. |

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
| ❌ | ⬜ | ⬜ | P1 | **`forloop.name`** | `"variavel-colecao"` — present in both Ruby and JS. Go não tem. Quick add. |
| ❌ | ⬜ | ⬜ | P3 | `forloop.parentloop` | Ruby-only. Referência ao `ForloopDrop` do loop pai. Go não tem. |

### 6.2 TablerowloopDrop

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ⚠️ | ⬜ | ⬜ | P1 | `row`, `col`, `col0`, `col_first`, `col_last` | **Verificar se Go expõe esses campos no `forloop` do `tablerow`**. O subagente reportou que tablerow "cria o mesmo objeto `forloop`", o que sugere que os campos específicos de tabela **podem estar faltando**. Presente em Ruby e JS. |

### 6.3 EmptyDrop / BlankDrop

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | **`empty` drop/literal** | `values.EmptyDrop` exportado em `values/emptydrop.go`. |
| ✅ | ⬜ | ⬜ | P1 | **`blank` drop/literal** | `values.BlankDrop` exportado em `values/emptydrop.go`. |

### 6.4 Drop base class

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | Interface `Drop` (`ToLiquid() any`) | Protocolo simples implementado. |
| ❌ | ⬜ | ⬜ | P3 | Drop base class com `liquid_method_missing` | Ruby: catch-all via `liquid_method_missing`. JS: `liquidMethodMissing`. Go: não tem catch-all — propriedades não encontradas retornam nil silenciosamente via reflection. **DECISÃO TOMADA:** adicionar interface `DropMethodMissing` opcional |
| ❌ | ⬜ | ⬜ | P3 | `context=` injection no drop | Ruby: drops recebem o contexto de render injetado. Go não tem. |

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
| ✅ | ⬜ | ⬜ | P1 | `StrictVariables()` | OK (engine-level). |
| ✅ | ⬜ | ⬜ | P1 | `LaxFilters()` | OK. |
| ✅ | ⬜ | ⬜ | P1 | Custom delimiters (`Delims()`) | OK. |
| ✅ | ⬜ | ⬜ | P1 | Custom `TemplateStore` | OK. |
| ✅ | ⬜ | ⬜ | P1 | `RegisterTag`, `RegisterBlock`, `RegisterFilter` | OK. |
| ⚠️ | ⬜ | ⬜ | P2 | `strict_variables` / `strict_filters` — **por render, não por engine** | Ruby e JS permitem definir per-render, além de por environment. Go só tem engine-level. **DECISÃO TOMADA:** expor essas opções como parâmetro do `Render()`/`FRender()` |
| ❌ | ⬜ | ⬜ | P3 | `error_mode` (`:lax`, `:warn`, `:strict`, `:strict2`) | Ruby-only. Go atual: sempre strict (tags indefinidas são erro de parse). |
| ❌ | ⬜ | ⬜ | P3 | `template.errors` / `template.warnings` — arrays acumulados | Ruby: `template.errors` coleta erros sem interromper render. Go: primeiro erro interrompe. |
| ❌ | ⬜ | ⬜ | P3 | `exception_renderer` / `exception_handler` | Ruby: proc intercepta exceções. |
| ❌ | ⬜ | ⬜ | P3 | Resource limits (score-based) | Ruby: `render_length_limit`, `render_score_limit`, `assign_score_limit`, `cumulative_*`. |
| ❌ | ⬜ | ⬜ | P4 | Resource limits (time-based: `renderLimit`, `parseLimit`) | JS-only. |
| ❌ | ⬜ | ⬜ | P4 | Template cache | JS-only. |
| ✅ | ⬜ | ⬜ | P4 | `globals` option no engine | Implementado. `Engine.SetGlobals(map[string]any{})` e `GetGlobals()` em `engine.go`. |

---

## 9. Análise Estática

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P2 | `GlobalVariableSegments`, `VariableSegments`, `GlobalFullVariables`, `FullVariables` | OK. |
| ✅ | ⬜ | ⬜ | P2 | `Analyze()` / `ParseAndAnalyze()` | OK — retorna `Variables`, `Globals`, `Locals`, `Tags`. |
| ✅ | ⬜ | ⬜ | P2 | `RegisterTagAnalyzer`, `RegisterBlockAnalyzer` | OK para extensão. |
| ❌ | ⬜ | ⬜ | P3 | `ParseTreeVisitor` API visitor-style | Ruby tem API de visitor pública. Go tem análise integrada mas não expõe visitor-style. **DECISÃO TOMADA:** necessário. |

---

## 10. Tratamento de Erros

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `SourceError` com `Path()`, `LineNumber()`, `Cause()` | OK. |
| ✅ | ⬜ | ⬜ | P3 | `ZeroDivisionError` tipo específico | Implementado em `filters/standard_filters.go`. Tipo exportado retornado por `divided_by` e `modulo`. |
| ⚠️ | ⬜ | ⬜ | P3 | Tipos específicos de erro (`SyntaxError`, `ArgumentError`, `ContextError`, etc.) | `ParseError` (em `parser/error.go`), `RenderError` e `UndefinedVariableError` (em `render/error.go`) implementados. `SyntaxError`, `ArgumentError`, `ContextError` ainda não. |
| ⚠️ | ⬜ | ⬜ | P1 | Metadados de erro — `markup_context` | Ruby inclui o texto do markup que causou o erro. Go inclui path e line, mas não o texto do markup no contexto. |

---

## 11. Whitespace Control

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ✅ | ⬜ | ⬜ | P1 | `{%-`, `-%}`, `{{-`, `-}}` | OK. |
| ❌ | ⬜ | ⬜ | P4 | Opções globais de trim (`trimTagRight`, etc.) | JS-only. |

---

## 12. Thread-safety e Concorrência

> Não faz sentido garantir imutabilidade antes de ter todos os campos de configuração definidos. Pode ser planejado em paralelo, mas implementado depois de estabilizar a API de configuração.
> Ver também **B5** (bug ativo de race condition no renderer).

| Impl | Tests | E2E | Prioridade | Item | Notas |
|------|-------|-----|-----------|------|-------|
| ❌ | ⬜ | ⬜ | P1 | Auditoria de estado mutável no `Engine` | Identificar todos os campos mutados após a criação do engine. O `Engine` deve ser seguro para uso concorrente sem locks externos. |
| ❌ | ⬜ | ⬜ | P1 | Estado de render isolado por chamada | Garantir que `render.Context` não compartilha estado mutável entre chamadas concorrentes (maps de variáveis copiados, não compartilhados). Diretamente ligado ao bug B5. |
| ❌ | ⬜ | ⬜ | P2 | `Config` imutável após construção | Toda configuração via `Engine.SetXxx()` ou `NewEngine(opts...)` deve ser finalizada antes do primeiro uso. Calls após uso devem retornar erro ou ser ignoradas. |

---

## Resumo Executivo por Prioridade

### P1 — Core Shopify Liquid (implementar primeiro)

```
Tags:
[x] echo tag                 ✅ DONE
[ ] liquid tag (multi-linha)  — depende de echo (echo já está pronto)
[ ] # inline comment
[ ] increment / decrement
[ ] render tag (escopo isolado)  — infra (SpawnIsolated + globals) já pronta
[ ] include — with/as/key-val args
[x] case/when — suporte a `or`  ✅ DONE

Filtros:
[ ] capitalize — fix (lowercase resto)          ainda ⚠️ não corrigido
[ ] strip_html — fix (remover script/style)     ainda ⚠️ não corrigido
[ ] newline_to_br — fix (preservar \n)          ainda ⚠️ não corrigido
[ ] modulo — fix (erro em divisão por zero)     ainda ⚠️ não corrigido (guard não adicionado)
[ ] default — allow_false keyword arg           infra NamedArg pronta; filtro não atualizado
[x] Keyword args em filtros (parser change)     ✅ DONE (infraestrutura NamedArg)

Expressões:
[x] empty literal/drop        ✅ DONE
[x] blank literal/drop        ✅ DONE
[x] Strings — suporte a escapes (\n, \", etc.) ✅ DONE
[x] array[-1] negative indexing               ✅ DONE

Drops:
[ ] forloop.name
[ ] tablerowloop drop — verificar row/col/col0/col_first/col_last

Context:
[x] Sub-contexto isolado (para render tag) ✅ DONE
[x] Variáveis globais separadas do escopo  ✅ DONE
```

### P2 — Extensões Comuns (Ruby + JS)

```
[ ] strict_variables / strict_filters como opção per-render
[x] globals option no engine  ✅ DONE
```

### P3 — Compat Ruby

```
[ ] squish filtro
[ ] h alias (escape)
[ ] base64_url_safe_encode/decode
[ ] compact: property arg
[ ] uniq: property arg
[ ] forloop.parentloop
[x] <> alias de !=   ✅ DONE
[ ] doc / enddoc tag
[ ] ifchanged tag
[ ] include for array as alias
[ ] Drop: liquid_method_missing
[ ] Error modes (:lax, :warn, :strict, :strict2)
[ ] template.errors / template.warnings arrays
[ ] Resource limits (score-based)
[ ] ParseTreeVisitor API
```

### P4 — Compat JS / Extensões

```
[ ] for offset: continue
[ ] date: 'now'/'today' como input
[ ] date_to_xmlschema / date_to_rfc822 / date_to_string / date_to_long_string
[x] where_exp / reject_exp / group_by_exp / has_exp / find_exp / find_index_exp  ✅ DONE
[ ] jsonify alias
[ ] raw filter
[ ] layout / block tags (herança)
[x] not operador unário       ✅ DONE
[ ] Opções globais de whitespace trim
[ ] Resource limits (time-based)
[ ] Template cache
```
