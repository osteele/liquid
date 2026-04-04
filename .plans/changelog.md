# Changelog — liquid-go fork

> Registro técnico das alterações feitas neste fork em relação ao repositório original `osteele/liquid`.
> Mantido progressivamente para subsidiar a PR ao autor original.
> Organizado por conjunto de mudanças logicamente relacionadas.

---

## [F1] Novos filtros: string, math, URL, encoding

**Arquivos modificados:**
- `filters/standard_filters.go` — novos filtros registrados em `AddStandardFilters()`

**Filtros adicionados:**

| Filtro | Implementação |
|--------|---------------|
| `remove_last` | `strings.LastIndex` + slice manual |
| `replace_last` | `strings.LastIndex` + slice manual |
| `normalize_whitespace` | regexp `\s+` → `" "` com `strings.TrimSpace` |
| `number_of_words` | split por `\s+`; argumento `mode` opcional (`"cjk"`, `"auto"`) para contagem Unicode CJK |
| `array_to_sentence_string` | join com vírgulas + conector configurável (default `"and"`) |
| `at_least` | `math.Max(a, b)` sobre `float64` |
| `at_most` | `math.Min(a, b)` sobre `float64` |
| `xml_escape` | `strings.NewReplacer` para `&`, `<`, `>`, `"`, `'` |
| `cgi_escape` | alias direto de `url.QueryEscape` |
| `uri_escape` | `url.PathEscape` com preservação de chars URI-safe (`;`, `,`, `/`, `?`, `:`, `@`, `&`, `=`, `+`, `$`, `#`, `[`, `]`) |
| `slugify` | normalização por modo: `"default"` (unicode), `"ascii"`, `"latin"` (transliteração de acentos via `golang.org/x/text`), `"pretty"` (preserva chars de URL), `"none"`/`"raw"` (só lowercase) |
| `base64_encode` | `base64.StdEncoding.EncodeToString` |
| `base64_decode` | `base64.StdEncoding.DecodeString`; retorna `(string, error)` |
| `to_integer` | converte `int`, `float`, `string`, `bool` para `int`; string: tenta `strconv.ParseInt` depois `strconv.ParseFloat`; `true` → 1, `false` → 0 |

---

## [F2] Novos filtros de array

**Arquivos criados:**
- `filters/array_filters.go` — implementações dos filtros de array

**Arquivos modificados:**
- `filters/standard_filters.go` — novos filtros registrados em `AddStandardFilters()`

**Filtros adicionados:**

| Filtro | Assinatura interna | Comportamento |
|--------|--------------------|---------------|
| `where` | `([]any, string, func(any) any) []any` | Mantém itens onde `item[property] == value`; sem `value` → filtra truthy |
| `reject` | `([]any, string, func(any) any) []any` | Inverso de `where` |
| `group_by` | `([]any, string) []any` | Retorna `[]map[string]any{{"name": key, "items": [...]}}` |
| `find` | `([]any, string, func(any) any) any` | Primeiro item que satisfaz; `nil` se não encontrado |
| `find_index` | `([]any, string, func(any) any) any` | Índice 0-based do primeiro match; `nil` se não encontrado |
| `has` | `([]any, string, func(any) any) bool` | `true` se ao menos um item satisfaz |
| `sum` | `([]any, func(string) string) any` | Soma numérica; argumento `property` opcional; preserva `int` se sem floats; parseia strings numéricas; pula não-numéricos |
| `push` | `([]any, any) []any` | Appends ao final; retorna novo slice |
| `unshift` | `([]any, any) []any` | Prepend no início; retorna novo slice |
| `pop` | `([]any) []any` | Remove último; retorna novo slice (vazio se input vazio) |
| `shift` | `([]any) []any` | Remove primeiro; retorna novo slice (vazio se input vazio) |
| `sample` | `([]any, func(int) int) any` | N elementos aleatórios via `math/rand/v2`; se `count=1` retorna elemento único; caso contrário retorna `[]any` |

---

## [A1] API de análise estática

Implementação de análise estática de templates — permite inspecionar variáveis referenciadas, variáveis definidas localmente e tags utilizadas em um template sem renderizá-lo.

### Novos arquivos

**`expressions/analysis.go`**
- `trackingContext` — implementa `expressions.Context`; intercepta chamadas `Get(name)` e `PropertyValue` para registrar todos os caminhos de variável acessados durante uma avaliação de rastreamento
- `trackingValue` — implementa `values.Value`; propaga o rastreamento em acesso a propriedades e índices
- `computeVariables(evaluator func(Context) values.Value) [][]string` — executa o evaluator com um `trackingContext` para coletar todos os paths de variável referenciados pela expressão; panics são absorvidos para retornar paths parciais em caso de expressão inválida

**`render/analysis.go`**
- `NodeAnalysis` struct — metadados de análise por nó: `Arguments []expressions.Expression` (expressões cujas variáveis são "lidas"), `LocalScope []string` (nomes definidos no escopo corrente), `BlockScope []string` (nomes válidos apenas no corpo do bloco)
- `TagAnalyzer` type — `func(args string) NodeAnalysis`
- `BlockAnalyzer` type — `func(node BlockNode) NodeAnalysis`
- `VariableRef` struct — `Path []string` + `Loc parser.SourceLoc`
- `AnalysisResult` struct — `Globals [][]string`, `All [][]string`, `GlobalRefs []VariableRef`, `AllRefs []VariableRef`, `Locals []string`, `Tags []string`
- `Analyze(root Node) AnalysisResult` — percorre a AST coletando variáveis (via `walkForVariables`), locals (via `collectLocals`) e tags usadas (via `walkForTags`); deduplica paths com map de chave `"\x00"`-separada

**`tags/analyzers.go`**
- `makeAssignAnalyzer() render.TagAnalyzer` — parse do statement `%assign`; reporta `ValueFn` em `Arguments` e o nome da variável em `LocalScope`; suporta dot-notation path
- `captureBlockAnalyzer(node) render.NodeAnalysis` — reporta nome da variável em `LocalScope`
- `loopBlockAnalyzer(node) render.NodeAnalysis` — parse do statement `%loop`; reporta `Expr` em `Arguments` e `Variable` em `BlockScope`; propagado para `tablerow` também
- `ifBlockAnalyzer() render.BlockAnalyzer` — percorre `Body` e `Clauses` coletando expressões de condição em `Arguments`
- `caseBlockAnalyzer(node) render.NodeAnalysis` — coleta a expressão `case` e todas as expressões `when` em `Arguments`

### Arquivos modificados

**`expressions/expressions.go`**
- `Expression` interface: adicionado método `Variables() [][]string`
- `expression` struct: adicionados campos `varsOnce sync.Once` e `variables [][]string`; `Variables()` implementado com lazy evaluation via `computeVariables`
- `expressionWrapper` (usado em `functional.go`): `Variables()` retorna `nil` (sem rastreamento para expressões wrappadas)

**`expressions/y.go`** (arquivo gerado pelo yacc — atualizado manualmente)
- Todas as construções `&expression{f}` alteradas para `&expression{evaluator: f}` para acomodar os novos campos da struct sem quebrar a geração de código

**`render/config.go`**
- `grammar` struct: adicionados campos `tagAnalyzers map[string]TagAnalyzer` e `blockAnalyzers map[string]BlockAnalyzer`
- Métodos `AddTagAnalyzer(name, a)` e `AddBlockAnalyzer(name, a)` em `*Config`
- Métodos internos `findTagAnalyzer(name)` e `findBlockAnalyzer(name)` em `grammar`

**`render/compiler.go`**
- Em `compileNode` para `*parser.ASTTag`: após compilar o renderer, consulta `findTagAnalyzer` e popula `TagNode.Analysis` se analyzer existir
- Em `compileNode` para `*parser.ASTBlock`: consulta `findBlockAnalyzer` e popula `BlockNode.Analysis`

**`render/nodes.go`**
- `TagNode` struct: adicionado campo `Analysis NodeAnalysis`
- `BlockNode` struct: adicionado campo `Analysis NodeAnalysis`

**`tags/standard_tags.go`**
- `AddStandardTags()`: registra os novos analyzers via `c.AddTagAnalyzer`/`c.AddBlockAnalyzer` para: `assign`, `capture`, `for`, `tablerow`, `if`, `unless`, `case`

**`analysis.go`** (raiz do pacote)
- Novos tipos públicos: `VariableSegment = []string` (alias), `Variable` struct (`Segments []string`, `Location parser.SourceLoc`, `Global bool`), `StaticAnalysis` struct
- `Variable.String() string` — retorna path com pontos
- Novos métodos em `*Engine`:
  - `GlobalVariableSegments(t *Template) ([]VariableSegment, error)`
  - `VariableSegments(t *Template) ([]VariableSegment, error)`
  - `GlobalVariables(t *Template) ([]string, error)`
  - `Variables(t *Template) ([]string, error)`
  - `GlobalFullVariables(t *Template) ([]Variable, error)`
  - `FullVariables(t *Template) ([]Variable, error)`
  - `Analyze(t *Template) (*StaticAnalysis, error)`
  - `ParseAndAnalyze(source []byte) (*Template, *StaticAnalysis, error)`
- Novos métodos em `*Template` (conveniência — delegam para `render.Analyze`):
  - `GlobalVariableSegments()`, `VariableSegments()`, `GlobalVariables()`, `Variables()`, `GlobalFullVariables()`, `FullVariables()`, `Analyze()`
- Novos métodos em `*Engine`:
  - `RegisterTagAnalyzer(name string, a render.TagAnalyzer)`
  - `RegisterBlockAnalyzer(name string, a render.BlockAnalyzer)`
  - `UnregisterTag(name string)` — remove tag pelo nome (idempotente)
- Funções internas: `rootNames`, `refsToVariables`, `fullVariablesFromResult`, `analyzeTemplate`

**`engine.go`**
- `RegisterTagAnalyzer` e `RegisterBlockAnalyzer` adicionados à API pública do `Engine`
- `UnregisterTag` adicionado à API pública do `Engine`

---

## [A1-misc] Melhorias de análise estática em `render/analysis.go`

- `walkForVariables` — walker recursivo que coleta `VariableRef` de `ObjectNode` (via `GetExpr`), `TagNode` (via `Analysis.Arguments`) e `BlockNode` (via `Analysis.Arguments` + body + clauses)
- `collectLocals` — walker recursivo que coleta nomes de variáveis definidas localmente via `LocalScope` e `BlockScope` dos nós
- `walkForTags` — walker recursivo que coleta nomes únicos de tags usadas no template
- `analysisCollector` — helper com deduplicação de paths por chave string, preservando a localização da primeira ocorrência
