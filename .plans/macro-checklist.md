# Macro Checklist: liquid-go

> Detalhamento completo de bugs e features, organizados para que agentes independentes possam planejar e trabalhar em paralelo.
> Ver [parity-checklist.md](parity-checklist.md) para o mapeamento item a item vs LiquidJS.
> Os repositórios de código de liquid em ruby e JS estão clonados localmente para referência durante o desenvolvimento, dentro de `.example-repositories/liquid-js/liquidjs` e `.example-repositories/liquid-ruby/liquid`.

---

## Legenda

- `BUG` — comportamento errado no código existente; precisa ser corrigido
- `FEAT` — funcionalidade nova; ainda não existe
- `[dep: X]` — depende de outro grupo ou item antes de começar
- `[paralelo com: X]` — pode ser desenvolvido ao mesmo tempo que X

---

## BUGS — Correções de comportamento existente

Esses itens não exigem novas estruturas. Podem ser investigados e corrigidos de forma independente.

### B1 · Tipos numéricos Go em comparações
- [ ] `BUG` `uint64`, `uint32`, `int8`, etc. causam comportamento incorreto em `{% if %}` e outros operadores de comparação. A conversão existe para filtros, mas não está garantida no avaliador de expressões. Verificar `expressions/` e `values/compare.go`.

### B2 · Truthiness: `nil`, `null`, `blank`, `empty`
- [ ] `BUG` Em Liquid, `nil` e `false` são falsy; todo o resto (incluindo `0`, `""`) é truthy. Verificar se `values/predicates.go` e `render/context.go` respeitam isso. O comportamento de `blank` e `empty` como palavras-chave em `{% if x == blank %}` também precisa de validação.

### B3 · Whitespace control em edge cases
- [ ] `BUG` Os marcadores `{%-`/`-%}` e `{{-`/`-}}` podem ter comportamento incorreto em casos como blocos aninhados, loops e templates com `include`. Validar contra o Golden Liquid test suite.

### B4 · Mensagens de erro e tipos
- [ ] `BUG` Erros de parse e render não têm tipos distintos exportados (`ParseError`, `RenderError`, `UndefinedVariableError`). O `SourceError` existe mas não distingue origem. Erros de variável indefinida com `strictVariables` precisam de tipo próprio.

---

## FEATURES — Novas funcionalidades

As features estão agrupadas por domínio e dependência. Dentro de cada grupo, itens sem `[dep]` podem ser iniciados imediatamente e em paralelo com outros grupos.

---

### Grupo F1 · Filtros simples (string, math, html, url, base64, misc)

> Totalmente independentes entre si e dos demais grupos. Podem ser implementados em paralelo.
> Todos ficam em `filters/standard_filters.go` (ou arquivo novo no mesmo package).
> Não precisam de acesso ao contexto de render — recebem apenas o valor de entrada + argumentos.

- [ ] `FEAT` `remove_last` — remove a última ocorrência de uma substring
- [ ] `FEAT` `replace_last` — substitui a última ocorrência de uma substring
- [ ] `FEAT` `normalize_whitespace` — colapsa espaços/tabs/newlines em um único espaço
- [ ] `FEAT` `number_of_words` — conta palavras; aceita argumento `"cjk"` ou `"auto"` para CJK
- [ ] `FEAT` `array_to_sentence_string` — junta array como frase ("a, b, and c")
- [ ] `FEAT` `at_least` — `max(input, n)` — retorna o maior entre input e n
- [ ] `FEAT` `at_most` — `min(input, n)` — retorna o menor entre input e n
- [ ] `FEAT` `xml_escape` — escapa caracteres especiais XML (`&`, `<`, `>`, `"`, `'`)
- [ ] `FEAT` `cgi_escape` — encode CGI (espaços viram `+`, especiais viram `%XX`)
- [ ] `FEAT` `uri_escape` — encode URI (como `url_encode`, mas preserva `/` e `:`)
- [ ] `FEAT` `slugify` — normaliza string para slug de URL; aceita modo (`"default"`, `"ascii"`, `"latin"`, `"none"`, `"raw"`)
- [ ] `FEAT` `base64_encode` — encoda string em Base64
- [ ] `FEAT` `base64_decode` — decoda Base64; deve retornar erro em input inválido
- [ ] `FEAT` `to_integer` — converte string ou float para inteiro (trunca, não arredonda)

---

### Grupo F2 · Filtros de array simples (sem avaliação de expressão)

> Independente do F1. Pode ser feito em paralelo.
> Todos operam sobre `[]any` via `values/` e não precisam do contexto de render.
> `where` filtra por propriedade+valor; `group_by` agrupa por propriedade; etc.

- [ ] `FEAT` `where` — filtra array: `array | where: "property", "value"` — mantém itens onde `item[property] == value`
- [ ] `FEAT` `reject` — inverso de `where`: mantém itens onde `item[property] != value`
- [ ] `FEAT` `group_by` — agrupa itens por valor de propriedade; retorna array de `{name, items}`
- [ ] `FEAT` `find` — retorna o primeiro item onde `item[property] == value`
- [ ] `FEAT` `find_index` — retorna o índice (0-based) do primeiro match de `where`
- [ ] `FEAT` `has` — retorna `true` se algum item satisfaz `item[property] == value`
- [ ] `FEAT` `sum` — soma valores numéricos do array; aceita argumento de propriedade (`sum: "price"`)
- [ ] `FEAT` `push` — adiciona elemento ao final do array (não muta, retorna novo array)
- [ ] `FEAT` `unshift` — adiciona elemento ao início do array
- [ ] `FEAT` `pop` — remove e retorna o último elemento (retorna array sem o último)
- [ ] `FEAT` `shift` — remove e retorna o primeiro elemento (retorna array sem o primeiro)
- [ ] `FEAT` `sample` — retorna N elementos aleatórios do array

---

### Grupo F3 · Filtros de array com expressão (`_exp`)

> `[dep: contexto de render acessível a filtros]` — esses filtros precisam avaliar uma expressão Liquid arbitrária por item do array. Requer que o filtro receba o contexto de render (`render.Context`) além do valor. Isso pode exigir uma mudança na assinatura de filtros ou um mecanismo de "filter com contexto".
> Pode ser planejado em paralelo com os demais grupos, mas a implementação depende de decidir como passar o contexto.

- [ ] `FEAT` `where_exp` — `array | where_exp: "item", "item.price > 10"` — filtra por expressão Liquid
- [ ] `FEAT` `reject_exp` — inverso de `where_exp`
- [ ] `FEAT` `group_by_exp` — agrupa por resultado de expressão
- [ ] `FEAT` `find_exp` — retorna primeiro item que satisfaz expressão
- [ ] `FEAT` `find_index_exp` — retorna índice do primeiro item que satisfaz expressão
- [ ] `FEAT` `has_exp` — retorna true se algum item satisfaz expressão

---

### Grupo F4 · Filtros de data

> Independente dos outros grupos de filtros. Pode ser feito em paralelo.
> O filtro `date` já existe em `filters/standard_filters.go`. Esses são aliases e formatos fixos.

- [ ] `FEAT` `date_to_xmlschema` — formata data como ISO 8601 / XML Schema (`2006-01-02T15:04:05Z07:00`)
- [ ] `FEAT` `date_to_rfc822` — formata data como RFC 822 (`Mon, 02 Jan 2006 15:04:05 -0700`)
- [ ] `FEAT` `date_to_string` — formato curto localizado ("2 Jan 2006"); aceita argumento de estilo
- [ ] `FEAT` `date_to_long_string` — formato longo localizado ("2 January 2006")

---

### Grupo T1 · Tags simples (sem novo mecanismo de escopo)

> Totalmente independentes entre si e de outros grupos. Podem ser feitos em paralelo.
> Todos ficam em `tags/` sem precisar alterar `render/` ou `expressions/`.

- [ ] `FEAT` `echo` — tag de output: `{% echo variavel | filtro %}` — equivalente a `{{ variavel | filtro }}` mas como tag. Implementação trivial: registrar tag que avalia a expressão e escreve no writer.
- [ ] `FEAT` `liquid` — tag multi-linha: permite escrever múltiplas tags sem delimitadores `{% %}`; cada linha dentro de `{% liquid ... %}` é interpretada como uma tag. Útil para lógica sem output.
- [ ] `FEAT` `#` (inline comment) — `{% # comentário %}` — tag que ignora o conteúdo. Mais simples que `comment`; não tem tag de fechamento.
- [ ] `FEAT` `increment` — `{% increment variavel %}` — imprime e incrementa um contador isolado (separado de `assign`). Começa em 0, imprime o valor atual e depois incrementa.
- [ ] `FEAT` `decrement` — `{% decrement variavel %}` — mesmo que `increment` mas decrementa. Começa em -1. Compartilha o mesmo registro de contadores que `increment`.

---

### Grupo T2 · Tag `render` (partial com escopo isolado)

> Independente do T3 (inheritance). Pode ser planejado em paralelo.
> Requer que o `render.Context` suporte criar um sub-contexto isolado (sem acesso às variáveis do pai, exceto as passadas explicitamente). O `TemplateStore` já existe — `render` o usa para carregar o partial.
> A tag `include` existente compartilha escopo; `render` **não deve** compartilhar.

- [ ] `FEAT` `render` tag — `{% render "arquivo.liquid", variavel: valor %}` — carrega e renderiza partial em escopo isolado. O partial não acessa variáveis do template pai. Aceita: argumentos nomeados, `for` (loop sobre array), `with` (bind de variável).
- [ ] `FEAT` Mecanismo de sub-contexto isolado em `render.Context` — necessário para `render` tag funcionar corretamente.

---

### Grupo T3 · Template inheritance (`layout` + `block`)

> `[dep: T2]` — o sistema de escopo isolado do T2 é necessário aqui também.
> `layout` e `block` são fortemente acoplados e devem ser implementados juntos.
> `BlockDrop` é necessário para `block.super`.

- [ ] `FEAT` `layout` tag — `{% layout "base.liquid" %}` — o template atual é renderizado dentro do layout. O conteúdo dos blocos definidos no template filho substituem os blocos do layout pai.
- [ ] `FEAT` `block` tag — `{% block "nome" %}conteúdo padrão{% endblock %}` — define uma região substituível. Em um template filho, `{% block "nome" %}...{% endblock %}` substitui o bloco pai. `{{ block.super }}` injeta o conteúdo original.
- [ ] `FEAT` `BlockDrop` — tipo público que expõe `block.super` dentro de um block override.

---

### Grupo A1 · API de análise estática

> Independente dos grupos de tags e filtros. Pode ser planejado em paralelo.
> A infraestrutura de `render/analysis.go` e `analysis.go` já existe; este grupo adiciona os métodos públicos que faltam e os tipos de retorno mais ricos.

- [x] `FEAT` `Engine.VariableSegments(tpl)` / `Template.VariableSegments()` — retorna `[][]string` com todas as variáveis referenciadas como segmentos de path (ex: `[["customer", "name"]]`) ✅ implementado em `analysis.go`
- [x] `FEAT` `Engine.GlobalVariableSegments(tpl)` / `Template.GlobalVariableSegments()` — como `VariableSegments()` mas apenas variáveis externas (não atribuídas no template) ✅ implementado em `analysis.go`
- [x] `FEAT` `Engine.Variables(tpl)` / `Template.Variables()` — retorna `[]string` com nomes simples (raiz) de todas as variáveis usadas (ex: `["product", "customer"]`), sem path completo ✅ implementado em `analysis.go`
- [x] `FEAT` `Engine.GlobalVariables(tpl)` / `Template.GlobalVariables()` — como `Variables()` mas apenas variáveis externas ✅ implementado em `analysis.go`
- [x] `FEAT` `Engine.FullVariables(tpl)` / `Template.FullVariables()` — retorna objetos `Variable` com path completo, linha/coluna, e se é global ✅ implementado em `analysis.go`
- [x] `FEAT` `Engine.GlobalFullVariables(tpl)` / `Template.GlobalFullVariables()` — como `FullVariables()` mas filtrado por globais ✅ implementado em `analysis.go`
- [x] `FEAT` `Engine.ParseAndAnalyze(src)` — parse + análise em um passo; retorna `(*Template, *StaticAnalysis, error)` ✅ implementado em `analysis.go`
- [x] `FEAT` `StaticAnalysis` struct — resultado rico da análise: `Variables`, `Globals`, `Locals`, `Tags` usados no template ✅ implementado em `analysis.go`; `Filters` reservado para implementação futura
- [x] `FEAT` Interfaces de análise em tags: `Arguments()`, `LocalScope()`, `BlockScope()` via `render.NodeAnalysis`; `Engine.RegisterTagAnalyzer()` / `Engine.RegisterBlockAnalyzer()` expõem registro para tags customizadas ✅ implementado em `engine.go`; `Children()` e `PartialScope()` ficam para implementação futura junto com a tag `render`

---

### Grupo C1 · Configuração: comportamento e escopo

> Independente de tags e filtros. Pode ser feito em paralelo.
> Requer adição de campos em `render/config.go` (ou similar) e propagação via `Engine`.

- [ ] `FEAT` `globals` — variáveis injetadas em todos os renders automaticamente (ex: `engine.SetGlobals(map[string]any{...})`). Merge com as variáveis do caller; caller tem precedência.
- [ ] `FEAT` `jsTruthy` — usa regras de truthiness do JavaScript (`0`, `""`, `[]` são falsy). Off por padrão (Liquid padrão só considera `nil` e `false` como falsy).
- [ ] `FEAT` `lenientIf` — não causa erro quando variável em `{% if %}` não existe; trata como `nil`. Útil para templates que podem ou não ter certas variáveis.
- [ ] `FEAT` `dynamicPartials` — interpreta o nome do arquivo em `include`/`render` como expressão Liquid (ex: `{% include variavel %}`). On por padrão no LiquidJS.
- [ ] `FEAT` `keepOutputType` — quando a saída de uma expressão é um tipo não-string (ex: número), preserva o tipo em vez de converter para string. Útil para APIs que processam o resultado.

---

### Grupo C2 · Configuração: whitespace e formatação

> Independente do C1. Pode ser feito em paralelo.

- [ ] `FEAT` `trimTagLeft` / `trimTagRight` — equivalente programático ao `{%-`/`-%}` via opção de configuração global. Quando `true`, todo output de tag tem whitespace removido automaticamente.
- [ ] `FEAT` `trimOutputLeft` / `trimOutputRight` — mesmo para `{{`/`}}`.
- [ ] `FEAT` `extname` — extensão de arquivo padrão para partials (ex: `".liquid"`). Quando definida, `{% include "arquivo" %}` resolve para `arquivo.liquid` sem precisar da extensão.
- [ ] `FEAT` `relativeReference` — resolve paths de partials relativos ao arquivo atual em vez de relativo ao root.

---

### Grupo C3 · Configuração: datas e locale

> Independente dos outros grupos de config. Pode ser feito em paralelo com F4.

- [ ] `FEAT` `dateFormat` — formato padrão do filtro `date` quando não especificado (ex: `"%Y-%m-%d"`)
- [ ] `FEAT` `timezoneOffset` — offset de timezone padrão para o filtro `date` (ex: `"+03:00"`)
- [ ] `FEAT` `locale` — locale para formatação de datas em `date_to_string`, `date_to_long_string` (ex: `"pt-BR"`)

---

### Grupo C4 · Limites de segurança (DoS protection)

> Independente dos outros grupos. Pode ser planejado em paralelo.
> Importante para uso em ambientes multi-tenant ou com templates não confiáveis.

- [ ] `FEAT` `parseLimit` — limite em bytes/runes aceitos no parse de um template. Retorna erro se excedido.
- [ ] `FEAT` `renderLimit` — timeout máximo de renderização (ex: `time.Duration`). Implementado via `context.Context` com deadline.
- [ ] `FEAT` `memoryLimit` — limite de alocações durante a renderização. Pode ser implementado via contador de bytes escritos ou nós visitados.

---

### Grupo D1 · Drop protocol e tipos customizados

> `[dep: C1 globals]` — `NullDrop`, `EmptyDrop`, `BlankDrop` precisam ser reconhecidos no contexto de avaliação.
> Os outros itens podem ser feitos em paralelo com C1.

- [ ] `FEAT` `Comparable` interface — permite que tipos Go customizados definam comparações (`Equals`, `GreaterThan`, `LessThan`). Integrar em `values/compare.go`.
- [ ] `FEAT` `ForloopDrop` como tipo público exportado — atualmente o `forloop` é um `map[string]any` interno. Torná-lo um tipo público com campos `Index`, `Index0`, `Length`, `First`, `Last`, `RIndex`, `RIndex0` permite que usuários usem type assertions e análise estática.
- [ ] `FEAT` `TablerowloopDrop` como tipo público exportado — análogo ao `ForloopDrop`.
- [ ] `FEAT` `NullDrop`, `EmptyDrop`, `BlankDrop` — singletons exportados com semântica correta: `blank` é truthy mas `== blank` é verdadeiro para `nil`, `false`, `""`, arrays/maps vazios.
- [ ] `FEAT` `liquidMethodMissing` — suporte a fallback por propriedade no `Drop`: quando uma propriedade não existe no objeto, chamar `LiquidMethodMissing(key string) any` se o tipo implementar essa interface.

---

### Grupo Q1 · Thread-safety e concorrência

> `[dep: todos os grupos de configuração C1–C4]` — não faz sentido garantir imutabilidade antes de ter todos os campos de config definidos.
> Pode ser planejado em paralelo, mas implementado por último.

- [ ] `FEAT` Auditoria de estado mutável no `Engine` — identificar todos os campos que são mutados após a criação. O `Engine` deve ser seguro para uso concorrente sem locks externos.
- [ ] `FEAT` `Config` imutável após construção — toda configuração feita via `Engine.SetXxx()` ou `NewEngine(opts...)` deve ser finalizada antes do primeiro uso. Calls após uso devem retornar erro ou ser ignoradas.
- [ ] `FEAT` Estado de render isolado por chamada — garantir que `render.Context` não compartilha estado mutável entre chamadas concorrentes (ex: maps de variáveis copiados, não compartilhados).

---

### Grupo Q2 · Testes e qualidade

> Pode ser trabalhado em paralelo com qualquer grupo de features, à medida que elas são implementadas.

- [ ] `FEAT` Integração com [Golden Liquid test suite](https://github.com/jg-rp/golden-liquid) — suite de testes agnóstica de linguagem em JSON/YAML. Meta: passar 100% dos casos.
- [ ] `FEAT` Fuzz testing nos parsers de expressão — `expressions/parser.go` e `expressions/scanner.rl` são candidatos; usar `go test -fuzz`.
- [ ] `FEAT` Benchmarks documentados — expandir `*_benchmark_test.go` para cobrir os principais caminhos (parse, render, análise) e manter histórico de regressão.

---

### Grupo E1 · Engine API: métodos faltantes

> Alguns itens dependem de outros grupos; veja notas.

- [ ] `FEAT` `Engine.EvalValue(src string, vars Bindings) (any, error)` — avalia uma expressão Liquid isolada sem template. `[dep: expressions/]` já tem o avaliador; é só expor.
- [ ] `FEAT` `Engine.Plugin(fn func(*Engine))` — registra um plugin como função que recebe o engine e chama `RegisterFilter`/`RegisterTag`. Açúcar sintático para configuração modular.
- [ ] `FEAT` `Engine.ParseFile(path string) (*Template, error)` — carrega e parseia um arquivo via `TemplateStore`. Conveniente para uso direto sem TemplateStore explícito.
- [ ] `FEAT` `Engine.RenderFile(path string, vars Bindings) ([]byte, error)` — parse + render de arquivo em um passo.
