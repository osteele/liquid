# Go Liquid — Mapeamento Completo de Features

> Referência extraída diretamente do código-fonte em `c:\Users\joca\github.com\joaqu1m\liquid`.
> Organizada seguindo a mesma estrutura do `ruby-liquid-reference.md` para facilitar comparação.
> Todos os caminhos de arquivo são relativos à raiz do repositório.

---

## Tags

### Tags de output / expressão

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `{{ }}` | `{{ expressao }}` | Output de variável ou expressão com filtros. Tipo: `ObjTokenType` → `ObjectNode`. |

> **Ausentes:** `echo`, `liquid` (multi-linha), `#` (inline comment)

---

### Tags de variável / estado

| Tag | Tipo | Sintaxe | Notas |
|-----|------|---------|-------|
| `assign` | simple tag | `{% assign var = expr %}` | Avalia expressão, define variável no escopo. Com `EnableJekyllExtensions()`, suporta dot notation: `{% assign page.prop = expr %}` (`Path []string`). Tem analyzer que reporta `LocalScope` + `Arguments`. |
| `capture` | block tag | `{% capture varname %}...{% endcapture %}` | Renderiza body como string, atribui à variável. Requer exatamente um nome de variável. Tem analyzer que reporta `LocalScope`. |

> **Ausentes:** `increment`, `decrement`

---

### Tags condicionais

| Tag | Sub-tags | Notas |
|-----|----------|-------|
| `if` | `elsif`, `else` | Operadores: `==`, `!=`, `<>` (via NEQ), `<`, `>`, `<=`, `>=`, `contains`, `and`, `or`. Truthy: não `nil` e não `false`. Tem static analyzer. |
| `unless` | `else` | Inverte condição inicial via `Not(expr)`. Usa o mesmo compilador (`ifTagCompiler(false)`). Tem o mesmo analyzer que `if`. |
| `case` | `when`, `else` | Avalia expressão `case`, compara com `values.Equal()`. `when` suporta múltiplos valores separados por **vírgula**. Tem static analyzer. |

> **Ausente:** `ifchanged`

> **Nota sobre `case`/`when`:** Valores separados por `or` na cláusula `when` do Ruby não são suportados — apenas vírgula (`,`).

---

### Tags de iteração

| Tag | Opções | Notas |
|-----|--------|-------|
| `for` | `reversed`, `limit: n`, `offset: n`, range `(a..b)` | Sub-tag `else` (quando coleção vazia). Cria objeto `forloop`. Suporta `break`/`continue`. Iteração sobre array, range, map. Tem static analyzer (`BlockScope` para var de loop, `Arguments` para expr da coleção). |
| `break` | — | Retorna sentinel `errLoopBreak`. Só válido dentro de `for`/`tablerow`. |
| `continue` | — | Retorna sentinel `errLoopContinueLoop`. Só válido dentro de `for`/`tablerow`. |
| `tablerow` | `cols: n`, `limit: n`, `offset: n`, range `(a..b)` | Mesma engine de loop que `for`. Gera HTML de tabela: `<tr class="rowN">...<td class="colN">...</td></tr>`. Cria mesmo objeto `forloop`. |
| `cycle` | nome opcional: `{% cycle "name": v1, v2 %}` | Deve estar dentro de `for`. Lê `forloop[".cycles"]` para rastrear posição por grupo. Prefixo de grupo com `:`. |

---

### Objeto `forloop` (criado por `for` e `tablerow`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `forloop.first` | bool | `true` na primeira iteração |
| `forloop.last` | bool | `true` na última iteração |
| `forloop.index` | int | Índice 1-based |
| `forloop.index0` | int | Índice 0-based |
| `forloop.rindex` | int | Índice reverso 1-based |
| `forloop.rindex0` | int | Índice reverso 0-based |
| `forloop.length` | int | Total de iterações |
| `.cycles` (interno) | map | Rastreia posição dos grupos de `cycle` |

> **Ausente vs Ruby:** `forloop.parentloop`, `forloop.name`

---

### Tags de inclusão de templates

| Tag | Sintaxe | Escopo | Notas |
|-----|---------|--------|-------|
| `include` | `{% include "filename" %}` | **Compartilhado** (bindings do pai são passados + sobrescritos por vars adicionais) | Resolve path relativo ao `SourceFile()`. Usa `TemplateStore.ReadTemplate()`. Implementado em `tags/include_tag.go`. |

> **Ausente:** `render` (escopo isolado)

---

### Tags de texto / estrutura

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `raw` | `{% raw %}...{% endraw %}` | Output literal, bypassa renderização. Parser seta `inRaw = true`. Tipo AST: `ASTRaw` → `RawNode`. |
| `comment` | `{% comment %}...{% endcomment %}` | Parser seta `inComment = true`, pula todos os tokens até `endcomment`. Tags internas não precisam ser balanceadas. |

> **Ausentes:** `#` (inline comment), `liquid` (multi-linha), `doc`, `echo`

---

## Filtros

### String (24 filtros)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `append` | `string \| append: suffix` | Concatena no final |
| `prepend` | `string \| prepend: prefix` | Concatena no início |
| `upcase` | `string \| upcase` | |
| `downcase` | `string \| downcase` | |
| `capitalize` | `string \| capitalize` | Maiúscula na primeira letra; string vazia inalterada |
| `escape` | `string \| escape` | HTML escape usando `html.EscapeString` |
| `escape_once` | `string \| escape_once` | Desescapa primeiro, depois escapa — evita duplo escape |
| `strip` | `string \| strip` | `strings.TrimSpace` |
| `lstrip` | `string \| lstrip` | Remove whitespace à esquerda (via `unicode.IsSpace`) |
| `rstrip` | `string \| rstrip` | Remove whitespace à direita |
| `newline_to_br` | `string \| newline_to_br` | Converte `\n` em `<br />` |
| `strip_html` | `string \| strip_html` | Remove tags HTML via regex `<.*?>` (pode ser insuficiente para casos complexos) |
| `strip_newlines` | `string \| strip_newlines` | Remove todos os `\n` e `\r\n` |
| `truncate` | `string \| truncate[: n[, ellipsis]]` | Default: n=50, ellipsis=`"..."`. Rune-aware. |
| `truncatewords` | `string \| truncatewords[: n[, ellipsis]]` | Default: n=15, ellipsis=`"..."` |
| `split` | `string \| split: sep` | Retorna array; separador espaço é especial (split em runs de whitespace); trailing empty strings removidas |
| `replace` | `string \| replace: old, new` | `strings.ReplaceAll` |
| `replace_first` | `string \| replace_first: old, new` | Substitui só a primeira |
| `replace_last` | `string \| replace_last: old, new` | Substitui só a última (via `strings.LastIndex`) |
| `remove` | `string \| remove: sub` | Remove todas as ocorrências |
| `remove_first` | `string \| remove_first: sub` | Remove só a primeira |
| `remove_last` | `string \| remove_last: sub` | Remove só a última |
| `normalize_whitespace` | `string \| normalize_whitespace` | Colapsa runs de whitespace em um espaço (**Jekyll extension**) |
| `number_of_words` | `string \| number_of_words[: mode]` | Conta palavras. Modos: `"default"`, `"cjk"`, `"auto"` (**Jekyll extension**) |

> **Ausente vs Ruby:** `squish` (Ruby colapsa + stripa; o equivalente aqui é `normalize_whitespace` mas sem strip automático)

---

### Array (22 filtros)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `size` | `array \| size` | Também funciona em strings (rune count) e maps. Retorna 0 para outros tipos. |
| `first` | `array \| first` | Retorna nil para array vazio |
| `last` | `array \| last` | Retorna nil para array vazio |
| `join` | `array \| join[: glue]` | Default glue `" "`. Pula itens nil. |
| `reverse` | `array \| reverse` | Retorna novo array |
| `sort` | `array \| sort[: key]` | Crescente; suporta sort por chave de mapa/struct. Definido em `filters/sort_filters.go`. |
| `sort_natural` | `array \| sort_natural[: key]` | Case-insensitive; suporta chave. |
| `uniq` | `array \| uniq` | Remove duplicatas. O(1) para tipos comparáveis, O(n²) fallback. |
| `compact` | `array \| compact` | Remove nils |
| `map` | `array \| map: property` | Extrai propriedade de cada item |
| `concat` | `array \| concat: outro_array` | Combina dois arrays (não deduplica) |
| `where` | `array \| where: prop[, valor]` | Filtra onde property == valor; sem valor = truthy. `filters/array_filters.go`. |
| `reject` | `array \| reject: prop[, valor]` | Inverso de `where`; sem valor = falsy |
| `find` | `array \| find: prop[, valor]` | Primeiro item que satisfaz; retorna nil se não encontrado |
| `find_index` | `array \| find_index: prop[, valor]` | Índice 0-based do primeiro match; nil se não encontrado |
| `has` | `array \| has: prop[, valor]` | Retorna bool; `true` se algum item satisfaz |
| `sum` | `array \| sum[: property]` | Soma numérica; preserva tipo int se sem floats; parseia strings; pula não-numéricos |
| `slice` | `array \| slice: start[, length]` | Fatia de array ou string. Suporta start negativo (do final). Rune-aware para strings. Funciona em `string`, `[]byte`, slices. |
| `group_by` | `array \| group_by: property` | Agrupa por valor de propriedade; retorna `[{"name": ..., "items": [...]}]` |
| `push` | `array \| push: element` | Adiciona ao final, retorna novo array |
| `unshift` | `array \| unshift: element` | Adiciona ao início, retorna novo array |
| `pop` | `array \| pop` | Remove último elemento, retorna novo array |
| `shift` | `array \| shift` | Remove primeiro elemento, retorna novo array |
| `sample` | `array \| sample[: n]` | Retorna n elementos aleatórios. Se n=1, retorna elemento único; senão array. |

> **Nota:** `push`, `unshift`, `pop`, `shift`, `sample`, `group_by` são extensões não presentes no Liquid Ruby standard.

---

### Matemática (11 filtros)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `abs` | `number \| abs` | `math.Abs` (float64) |
| `plus` | `number \| plus: n` | Preserva tipo int se ambos são int |
| `minus` | `number \| minus: n` | Preserva tipo int se ambos são int |
| `times` | `number \| times: n` | Preserva tipo int se ambos são int |
| `divided_by` | `number \| divided_by: n` | Divisão inteira se divisor é int; float se float; retorna erro em divisão por zero |
| `modulo` | `number \| modulo: n` | `math.Mod` (float) |
| `round` | `number \| round[: casas]` | Default 0 casas decimais |
| `ceil` | `number \| ceil` | |
| `floor` | `number \| floor` | |
| `at_least` | `number \| at_least: n` | `max(input, n)` — clamp mínimo |
| `at_most` | `number \| at_most: n` | `min(input, n)` — clamp máximo |

---

### Data (1 filtro)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `date` | `date \| date[: format]` | Formato strftime via biblioteca `tuesday`. Default `"%a, %b %d, %y"`. Suporta múltiplos formatos de parse (ANSIC, RFC3339, ISO 8601, Ruby, etc.). |

**Formatos de parse aceitos:** `ANSIC`, `UnixDate`, `RubyDate`, `RFC822`, `RFC822Z`, `RFC850`, `RFC1123`, `RFC1123Z`, `RFC3339`, ISO 8601, `"2006-01-02"`, entre outros. Ver `values/parsedate.go`.

---

### HTML / URL / Encoding (8 filtros)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `url_encode` | `string \| url_encode` | `url.QueryEscape` |
| `url_decode` | `string \| url_decode` | `url.QueryUnescape`; retorna erro em input inválido |
| `base64_encode` | `string \| base64_encode` | Base64 encoding padrão |
| `base64_decode` | `string \| base64_decode` | Base64 decode padrão; retorna erro em input inválido |
| `xml_escape` | `string \| xml_escape` | Escapa `& < > " '` |
| `cgi_escape` | `string \| cgi_escape` | `url.QueryEscape` (**Jekyll extension**) |
| `uri_escape` | `string \| uri_escape` | URI-encode preservando chars seguros (equiv. a `encodeURI()` do JS). (**Jekyll extension**) |
| `slugify` | `string \| slugify[: mode]` | Converte para slug URL. Modos: `"default"` (unicode), `"ascii"`, `"latin"` (transliterate acentos), `"pretty"` (preserva chars de URL), `"none"`/`"raw"` (só lowercase). (**Jekyll extension**) |

> **Ausentes vs Ruby:** `base64_url_safe_encode`, `base64_url_safe_decode`

---

### Valor / Tipo (4 filtros)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `default` | `var \| default: val` | Retorna default se valor é nil, false, ou empty string/array |
| `json` | `var \| json` | Serializa para JSON via `json.Marshal` |
| `to_integer` | `var \| to_integer` | Converte para int; handles int/float/string/bool (true=1, false=0) |
| `array_to_sentence_string` | `array \| array_to_sentence_string[: connector]` | Une array como frase inglesa: `"a, b, and c"`. Default connector `"and"`. (**Jekyll extension**) |

> **Nota vs Ruby:** `default` nesta implementação **não** suporta keyword argument `allow_false: true`.

---

### Debug (2 filtros)

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `inspect` | `var \| inspect` | JSON ou `%#v`. (**Jekyll extension**) |
| `type` | `var \| type` | Retorna nome do tipo Go (`%T`). (**extensão proprietária**) |

---

## Sistema de Filtros

| Feature | Descrição |
|---------|-----------|
| Filtros posicionais | `{{ val \| filter: arg1, arg2 }}` |
| `AddFilter(name, fn)` | Registra função Go como filtro. Fn deve ter ≥1 entrada e 1 ou 2 saídas (2ª se presente: `error`). |
| `LaxFilters()` | Engine method. Silenciamente passa input quando filtro é desconhecido (comportamento Shopify). Default: filtro desconhecido é erro. |
| `UndefinedFilter` | Tipo de erro no package `expressions`. String com nome do filtro. |
| `FilterError` | Tipo de erro em `expressions`. Contém `FilterName string` e `Err error`. |
| `safe` filter | Registrado automaticamente por `SetAutoEscapeReplacer()`. Marca valor como seguro para não escapar no auto-escape. |
| **Sem keyword args** | `allow_false: true` do Ruby **não é suportado** — filtros não recebem hash de kwargs. |

---

## Expressões e Operadores

### Literais

| Literal | Exemplo | Notas |
|---------|---------|-------|
| nil | `nil` | Token: `LITERAL` → `nil` |
| boolean | `true`, `false` | Token: `LITERAL` → `bool` |
| inteiro | `42`, `-1` | `strconv.ParseInt`, Go type: `int` |
| float | `3.14`, `-0.5` | `strconv.ParseFloat`, Go type: `float64` |
| string | `"texto"` ou `'texto'` | **Sem suporte a escapes dentro de strings** (TODO no código) |
| range | `(1..10)` | Tipo `values.Range`; suporta variáveis: `(a..b)` |

> **Ausentes vs Ruby:** `blank`, `empty` como literais comparáveis. Nesta impl, `blank` e `empty` são apenas identificadores tratados como variáveis não definidas (nil).

### Operadores de comparação

| Operador | Token | Comportamento |
|----------|-------|--------------|
| `==` | `EQ` | `values.Equal()` — nil-safe, suporta int/float/string/bool |
| `!=` | `NEQ` | `!values.Equal()` |
| `<>` | `NEQ` | Alias de `!=` (mesmo token no scanner) |
| `<` | `'<'` | `values.Less()` |
| `>` | `'>'` | `values.Less()` invertido |
| `<=` | `LE` | `Less || Equal` |
| `>=` | `GE` | `Less(b,a) || Equal` |
| `contains` | `CONTAINS` | String: `strings.Contains`; Array: `reflect` search; Map: key lookup |

### Operadores booleanos

| Operador | Token | Comportamento |
|----------|-------|--------------|
| `and` | `AND` | `fa.Test() && fb.Test()` — sem curto-circuito real, ambos avaliados |
| `or` | `OR` | `fa.Test() \|\| fb.Test()` |

### Truthiness

| Valor | Truthy? |
|-------|---------|
| `false` | falsy |
| `nil` | falsy |
| `0` | **truthy** |
| `""` | **truthy** |
| `[]` | **truthy** |
| qualquer outro | truthy |

Implementado via `Value.Test()` em `values/value.go`.

### Acesso a variáveis

| Sintaxe | Descrição |
|---------|-----------|
| `variavel` | Lookup em `ctx.Get(name)` → `values.ToLiquid(bindings[name])` |
| `obj.prop` | Token `PROPERTY` → `makeObjectPropertyExpr()` |
| `obj[key]` | Expr `[expr]` → `makeIndexExpr()` |
| `array[0]` | Índice inteiro via `IndexValue()` |
| `array.first`, `array.last` | Propriedades especiais em `arrayValue.PropertyValue()` |
| `array.size`, `hash.size` | Propriedade `size` retorna comprimento |
| `forloop.index`, etc. | Propriedades do objeto forloop (map em Go) |

### Identificadores

- Suportam Unicode (letras, dígitos, `_`, `-` exceto no primeiro caractere)
- Podem terminar com `?` (predicados estilo Ruby)

---

## Drops (protocolo de objetos customizados)

### Interface `Drop` (`liquid.Drop`)

| Feature | Descrição |
|---------|-----------|
| `Drop` interface | Definida em `liquid/drops.go`: `ToLiquid() any` |
| `FromDrop(object any) any` | Função pública: se `object` implementa `Drop`, retorna `object.ToLiquid()`; senão retorna o próprio objeto |
| Resolução lazy | `dropWrapper` em `values/drop.go` usa `sync.Once` — `ToLiquid()` é chamado apenas na primeira avaliação |
| `values.ToLiquid(value)` | Converte objeto para Liquid se implementa a interface interna `drop` |

> **Ausente vs Ruby:** Não há `Drop` base class com `liquid_method_missing`, `invokable_methods`, blacklist, `context=`, `key?`. O protocolo Go é apenas `ToLiquid() any`.

### `IterationKeyedMap` (`tags.IterationKeyedMap`)

| Feature | Descrição |
|---------|-----------|
| `IterationKeyedMap` | Tipo público: `map[string]any`. Quando iterdado em `for`, yield são as **keys** (não pares key/value). |
| `liquid.IterationKeyedMap(m)` | Função helper pública para criar o wrapper |

### `yaml.MapSlice` (suporte interno)

| Feature | Descrição |
|---------|-----------|
| `yaml.MapSlice` | Tipo `gopkg.in/yaml.v2.MapSlice`. Iteração em `for` preserva ordem de inserção. Lookup por chave via `mapSliceValue`. |

### `values.SafeValue`

| Feature | Descrição |
|---------|-----------|
| `SafeValue{Value: v}` | Tipo em `values/value.go`. Marca valor como seguro para auto-escape. Usado pelo filtro `safe`. |

---

## Acesso a Structs pelo template

Go structs são acessíveis via PropertyValue (via reflection):
- Campos exportados mapeados por nome
- Métodos exportados mapeados por nome
- `structValue.PropertyValue()` em `values/structvalue.go`

---

## Erros

### `SourceError` / `parser.Error` / `render.Error`

| Interface | Métodos | Notas |
|-----------|---------|-------|
| `liquid.SourceError` | `Error() string`, `Cause() error`, `Path() string`, `LineNumber() int` | Interface pública. Retornada por `ParseTemplate`, `Render`, etc. |
| `parser.Error` | Mesmos métodos | Interface interna; compatível com `SourceError` |
| `render.Error` | Mesmos métodos | Interface interna |

A implementação concreta é `parser.sourceLocError`:
- `Error()` formata como `"Liquid error (line N): mensagem in caminho"`
- `Cause()` retorna o erro original
- `Path()` retorna pathname do template
- `LineNumber()` retorna número de linha

### Tipos de erro em `expressions`

| Tipo | Descrição |
|------|-----------|
| `InterpreterError` | `string` — erro de interpretação de expressão (input inválido) |
| `UndefinedFilter` | `string` — filtro não definido |
| `FilterError` | struct com `FilterName string`, `Err error` — erro ao aplicar filtro |
| `values.TypeError` | `string` — erro de conversão de tipo |

> **Ausentes vs Ruby:** Não há tipos distintos de erro para `SyntaxError`, `ArgumentError`, `ContextError`, `FileSystemError`, `MemoryError`, `ZeroDivisionError`, `UndefinedVariable`, `UndefinedDropMethod`, `MethodOverrideError`, `DisabledError`, `TemplateEncodingError`, etc.

---

## Engine — API Pública

### Criação

| Função | Descrição |
|--------|-----------|
| `liquid.NewEngine() *Engine` | Engine completo com filtros e tags padrão |
| `liquid.NewBasicEngine() *Engine` | Engine sem filtros/tags padrão |

### Parse

| Método | Assinatura | Descrição |
|--------|-----------|-----------|
| `ParseTemplate` | `(source []byte) (*Template, SourceError)` | Parse básico |
| `ParseString` | `(source string) (*Template, SourceError)` | Wrapper de string |
| `ParseTemplateLocation` | `(source []byte, path string, line int) (*Template, SourceError)` | Parse com localização para erros e `include` |
| `ParseTemplateAndCache` | `(source []byte, path string, line int) (*Template, SourceError)` | Parse + cache interno (`cfg.Cache[path]`) |

### Parse + Render combinados

| Método | Assinatura | Descrição |
|--------|-----------|-----------|
| `ParseAndRender` | `(source []byte, b Bindings) ([]byte, SourceError)` | |
| `ParseAndFRender` | `(w io.Writer, source []byte, b Bindings) SourceError` | Render direto em writer |
| `ParseAndRenderString` | `(source string, b Bindings) (string, SourceError)` | |

### Configuração

| Método | Descrição |
|--------|-----------|
| `StrictVariables()` | Variável undefined produz erro |
| `LaxFilters()` | Filtro undefined passa o input silenciosamente |
| `EnableJekyllExtensions()` | Habilita dot notation em `assign` (`page.prop = valor`) |
| `Delims(objL, objR, tagL, tagR string) *Engine` | Customiza delimitadores. Empty string = default. |
| `SetAutoEscapeReplacer(replacer Replacer)` | Habilita auto-escape. Registra filtro `safe` automaticamente. |

### Registro de extensões

| Método | Descrição |
|--------|-----------|
| `RegisterTag(name string, td Renderer)` | Registra tag simples. `Renderer = func(render.Context) (string, error)`. |
| `RegisterBlock(name string, td Renderer)` | Registra tag de bloco. |
| `RegisterFilter(name string, fn any)` | Registra filtro. Fn: ≥1 input, 1 ou 2 outputs (2ª = error). |
| `RegisterTemplateStore(ts render.TemplateStore)` | Substitui TemplateStore (fonte de arquivos para `include`). |
| `RegisterTagAnalyzer(name, a TagAnalyzer)` | Registra analyzer para tag customizada. |
| `RegisterBlockAnalyzer(name, a BlockAnalyzer)` | Registra analyzer para block tag customizado. |
| `UnregisterTag(name string)` | Remove tag pelo nome (idempotente). |

### Análise estática

| Método | Retorno | Descrição |
|--------|---------|-----------|
| `GlobalVariableSegments(t)` | `([]VariableSegment, error)` | Paths de vars globais |
| `VariableSegments(t)` | `([]VariableSegment, error)` | Paths de todas as vars |
| `GlobalVariables(t)` | `([]string, error)` | Nomes únicos raiz das vars globais |
| `Variables(t)` | `([]string, error)` | Nomes únicos raiz de todas as vars |
| `GlobalFullVariables(t)` | `([]Variable, error)` | Refs globais com path + localização |
| `FullVariables(t)` | `([]Variable, error)` | Todas as refs com path + localização + flag `Global` |
| `Analyze(t)` | `(*StaticAnalysis, error)` | Análise completa: vars, globals, locals, tags |
| `ParseAndAnalyze(source)` | `(*Template, *StaticAnalysis, error)` | Parse + análise em um passo |

---

## Template — API Pública

### Render

| Método | Assinatura | Descrição |
|--------|-----------|-----------|
| `Render` | `(vars Bindings) ([]byte, SourceError)` | Render completo para bytes |
| `FRender` | `(w io.Writer, vars Bindings) SourceError` | Render direto em writer |
| `RenderString` | `(b Bindings) (string, SourceError)` | Wrapper de string |

### AST

| Método | Retorno | Descrição |
|--------|---------|-----------|
| `GetRoot()` | `render.Node` | Retorna nó raiz da parse tree |

### Análise estática (métodos no Template)

Mesmos que no Engine, mas como métodos conveniência diretamente no `*Template`:

`GlobalVariableSegments()`, `VariableSegments()`, `GlobalVariables()`, `Variables()`, `GlobalFullVariables()`, `FullVariables()`, `Analyze()`

---

## Tipos Públicos

### `liquid.Bindings`

```go
type Bindings map[string]any
```
Alias de documentação para `map[string]any`.

### `liquid.Renderer`

```go
type Renderer func(render.Context) (string, error)
```
Tipo para definição de tags customizadas.

### `liquid.VariableSegment`

```go
type VariableSegment = []string
```
Caminho para variável como slice de segments.

### `liquid.Variable`

```go
type Variable struct {
    Segments []string
    Location parser.SourceLoc
    Global   bool
}
```
Com método `String() string` retornando path com pontos.

### `liquid.StaticAnalysis`

```go
type StaticAnalysis struct {
    Variables []Variable
    Globals   []Variable
    Locals    []string
    Tags      []string
    Filters   []string // sempre nil por enquanto
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

## Análise Estática (render.NodeAnalysis / render.AnalysisResult)

| Tipo/Feature | Descrição |
|-------------|-----------|
| `render.NodeAnalysis` | `Arguments []Expression`, `LocalScope []string`, `BlockScope []string` |
| `render.TagAnalyzer` | `func(args string) NodeAnalysis` |
| `render.BlockAnalyzer` | `func(node BlockNode) NodeAnalysis` |
| `render.VariableRef` | `Path []string`, `Loc parser.SourceLoc` |
| `render.AnalysisResult` | `Globals`, `All`, `GlobalRefs`, `AllRefs`, `Locals`, `Tags` |
| `render.Analyze(root Node)` | Função principal de análise; percorre AST coletando variáveis, locals, tags |
| `expressions.Expression.Variables()` | Retorna `[][]string` com paths de variáveis da expressão (lazy, cacheado) |

Tags padrão com analyzers: `assign` (LocalScope + Arguments), `capture` (LocalScope), `if`/`unless`/`case` (Arguments), `for`/`tablerow` (BlockScope + Arguments).

---

## Context de Render (`render.Context` interface)

Interface pública para implementadores de tags customizadas:

| Método | Descrição |
|--------|-----------|
| `Bindings() map[string]any` | Ambiente léxico atual completo |
| `Get(name string) any` | Obtém variável do ambiente atual |
| `Set(name string, value any)` | Define variável no ambiente atual |
| `SetPath(path []string, value any) error` | Define variável em path aninhado (usado por assign com dot notation) |
| `Evaluate(expr expressions.Expression) (any, error)` | Avalia expressão compilada |
| `EvaluateString(source string) (any, error)` | Compila e avalia string de expressão |
| `ExpandTagArg() (string, error)` | Renderiza argumento da tag como template Liquid (para Jekyll `{% include {{ var }} %}`) |
| `InnerString() (string, error)` | Renderiza body do bloco atual como string (para `capture`, `highlight`) |
| `RenderBlock(w io.Writer, b *BlockNode) error` | Renderiza BlockNode |
| `RenderChildren(w io.Writer) Error` | Renderiza filhos do nó atual |
| `RenderFile(filename string, b map[string]any) (string, error)` | Parseia + renderiza arquivo externo (usado por `include`) |
| `SourceFile() string` | Path do template atual (para `include` relativo) |
| `TagArgs() string` | Texto dos argumentos da tag atual |
| `TagName() string` | Nome da tag atual |
| `Errorf(format, a...) Error` | Cria erro com localização da fonte |
| `WrapError(err error) Error` | Envolve erro com localização |

---

## TemplateStore (file system)

| Interface/Tipo | Descrição |
|----------------|-----------|
| `render.TemplateStore` interface | `ReadTemplate(name string) ([]byte, error)` |
| `render.FileTemplateStore{}` | Implementação padrão; usa `os.ReadFile(filename)` diretamente |
| `Engine.RegisterTemplateStore(ts)` | Substitui a implementação padrão |
| Cache interno (`cfg.Cache`) | `map[string][]byte`; populado por `ParseTemplateAndCache()`; consultado por `include` quando arquivo não encontrado no disco |

---

## Auto-escape

| Feature | Descrição |
|---------|-----------|
| `render.Replacer` interface | `WriteString(io.Writer, string) (int, error)` |
| `render.HtmlEscaper` | `strings.NewReplacer` para `& ' < > "` |
| `Engine.SetAutoEscapeReplacer(r)` | Habilita auto-escape globalmente no engine; registra filtro `safe` |
| `safe` filter | Marca valor como `values.SafeValue{}` para pular auto-escape |
| `values.SafeValue` | Struct `{ Value any }` — wrapper de tipo seguro |

---

## Whitespace Control (Trimmer)

| Marcador | Efeito |
|----------|--------|
| `{%-` | Remove whitespace antes da tag (TrimLeft) |
| `-%}` | Remove whitespace depois da tag (TrimRight) |
| `{{-` | Remove whitespace antes do output |
| `-}}` | Remove whitespace depois do output |

Implementado via `render.trimWriter` em `render/trimwriter.go`.  
AST: `ASTTrim` (parser) → `TrimNode` (render).

---

## Delimitadores Customizados

```go
engine.Delims("{{", "}}", "{%", "%}")
```
- Pode ser chamado antes de qualquer `Parse*`.
- Empty string = usa default.
- Regexps compilados e cacheados por combinação de delimitadores.
- Implementado em `parser/scanner.go`.

---

## Configuração interna (`render.Config`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `StrictVariables` | `bool` | Variável undefined vira erro |
| `LaxFilters` | `bool` | Filtro undefined passa o input |
| `JekyllExtensions` | `bool` | Habilita dot notation em assign |
| `TemplateStore` | `TemplateStore` | Fonte de templates para `include` |
| `Cache` | `map[string][]byte` | Cache de templates parseados/cached |
| `escapeReplacer` | `Replacer` | Para auto-escape |
| `Delims` | `[]string` | Delimitadores customizados |

---

## Configuração interna (`expressions.Config`)

| Campo | Tipo | Descrição |
|-------|------|-----------|
| `filters` | `map[string]any` | Mapa de filtros registrados |
| `LaxFilters` | `bool` | Propagado de `render.Config` |

---

## Parser — Tokens e AST

### Tipos de token (`parser.TokenType`)

| Tipo | Valor | Descrição |
|------|-------|-----------|
| `TextTokenType` | 0 | Texto fora de `{{ }}` e `{% %}` |
| `TagTokenType` | 1 | Tag `{% ... %}` |
| `ObjTokenType` | 2 | Object `{{ ... }}` |
| `TrimLeftTokenType` | 3 | Marcador `-` esquerdo |
| `TrimRightTokenType` | 4 | Marcador `-` direito |

### Tipos de nó AST (parser)

`ASTBlock`, `ASTRaw`, `ASTTag`, `ASTText`, `ASTObject`, `ASTSeq`, `ASTTrim`

### Tipos de nó de render

| Tipo | Descrição |
|------|-----------|
| `Node` (interface) | `SourceLocation()`, `SourceText()`, `render(*trimWriter, nodeContext)` |
| `BlockNode` | `{% tag %}...{% endtag %}`. Tem `Body []Node`, `Clauses []*BlockNode`, `Analysis NodeAnalysis` |
| `RawNode` | Conteúdo de `{% raw %}` |
| `TagNode` | Tag simples. Tem `Analysis NodeAnalysis` |
| `TextNode` | Texto literal |
| `ObjectNode` | `{{ expr }}`. Tem método `GetExpr() expressions.Expression` |
| `SeqNode` | Sequência de nós filhos |
| `TrimNode` | Marcador de trim |

### `parser.SourceLoc`

```go
type SourceLoc struct {
    Pathname string
    LineNo   int
}
```

---

## Tipos de Expressão (`expressions`)

### Interface `Expression`

```go
type Expression interface {
    Evaluate(ctx Context) (any, error)
    Variables() [][]string  // lazy + cacheado
}
```

### Interface `Closure`

```go
type Closure interface {
    Bind(name string, value any) Closure
    Evaluate() (any, error)
}
```

### Tipos de statement (para tags)

| Tipo | Campos | Usado por |
|------|--------|-----------|
| `Assignment` | `Variable string`, `Path []string`, `ValueFn Expression` | `assign` |
| `Cycle` | `Group string`, `Values []string` | `cycle` |
| `Loop` | `Variable string`, `Expr Expression`, + `loopModifiers` | `for`, `tablerow` |
| `loopModifiers` | `Limit Expression`, `Offset Expression`, `Cols Expression`, `Reversed bool` | `for`, `tablerow` |
| `When` | `Exprs []Expression` | `when` |

---

## CLI (`cmd/liquid`)

| Feature | Descrição |
|---------|-----------|
| Entrada | Arquivo como argumento ou stdin |
| `--env` flag | Bind de variáveis de ambiente como bindings |
| `--strict` flag | Habilita `StrictVariables` |
| `--lax-filters` flag | Habilita `LaxFilters` |
| Saída | stdout |
| Erros | stderr + exit code 1 |

---

## Extensões Jekyll (`EnableJekyllExtensions()`)

Ativadas via `Engine.EnableJekyllExtensions()`:

| Feature | Descrição |
|---------|-----------|
| Dot notation em `assign` | `{% assign page.canonical_url = value %}` |

Filtros Jekyll (sempre disponíveis, mesmo sem EnableJekyllExtensions):

`normalize_whitespace`, `number_of_words`, `cgi_escape`, `uri_escape`, `slugify`, `inspect`, `array_to_sentence_string`

---

## Values — Sistema de Tipos

### Tipos Go mapeados

| Go type | Liquid behavior |
|---------|----------------|
| `nil` (pointer nulo, interface nil) | nil value |
| `bool` | boolean value |
| `int`, `int8`, ..., `int64` | numeric, preserva tipo |
| `float32`, `float64` | numeric float |
| `string` | string value, rune-aware em size/slice |
| `[]T` (slice/array) | array value |
| `map[K]V` | map value, iteração como pares `{key, value}` |
| `IterationKeyedMap` | map value, iteração como apenas keys |
| `yaml.MapSlice` | ordered map, preserva ordem |
| struct | property access via reflection |
| pointer para struct | unwrapped |
| `Drop` (ToLiquid) | lazy resolved |
| `values.SafeValue` | passa por auto-escape |

### Funções públicas de `values`

| Função | Descrição |
|--------|-----------|
| `ValueOf(value any) Value` | Wraps Go value em Liquid Value |
| `ToLiquid(value any) any` | Converte Drop via ToLiquid() |
| `Equal(a, b any) bool` | Comparação Liquid-aware |
| `Less(a, b any) bool` | Comparação Liquid-aware |
| `Length(value any) int` | Comprimento de string (runes) ou array |
| `IsEmpty(value any) bool` | Vazio: nil, string "", array/map len=0, bool false |
| `NewRange(b, e int) Range` | Cria range inclusivo `[b..e]` |
| `Sort(data []any)` | Sort genérico |
| `SortByProperty(data []any, key string, nilFirst bool)` | Sort por propriedade |
| `ParseDate(s string) (time.Time, error)` | Parse de data em múltiplos formatos |
| `Convert(value any, typ reflect.Type) (any, error)` | Conversão de tipo |

---

## Recursos Ausentes vs Ruby Liquid

> Resumo rápido dos itens que estão no Ruby mas não nesta implementação Go.
> Para planejamento de features futuras.

| Item | Grupo |
|------|-------|
| `echo` tag | Tags |
| `liquid` tag (multi-linha) | Tags |
| `#` inline comment | Tags |
| `increment` / `decrement` | Tags |
| `render` tag (escopo isolado) | Tags |
| `doc` tag | Tags |
| `ifchanged` tag | Tags |
| Sub-contexto isolado para `render` | Context |
| Filter keyword arguments (`allow_false: true`) | Filtros |
| `squish` filtro | Filtros |
| `base64_url_safe_encode` / `base64_url_safe_decode` | Filtros |
| `blank` / `empty` como literais com semântica Ruby | Expressões |
| `<>` não é alias de `!=` no scanner (verificar) | Expressões |
| `case`/`when` com `or` separando valores | Tags |
| `forloop.parentloop` | Drops |
| `forloop.name` | Drops |
| Drop base class com `liquid_method_missing` | Drops |
| Drop `context=` injection | Drops |
| `strict_variables` como opção por-render (não por-engine) | Config |
| `strict_filters` como opção por-render | Config |
| `global_filter` (proc aplicada a todo output) | Config |
| Error modes `:lax`, `:warn`, `:strict`, `:strict2` | Config |
| Resource limits (render_score, assign_score, etc.) | Config |
| Profiler | Observabilidade |
| `template.errors` array | Template API |
| `template.warnings` array | Template API |
| `template.name` | Template API |
| Disableable tags / Disabler mixin | Tags |
| `TemplateEncodingError` | Erros |
| `cumulative_render_score_limit` | Resource limits |
