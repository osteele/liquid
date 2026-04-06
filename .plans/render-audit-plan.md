# Spec: Render Diagnostics — `RenderAudit`

## Objetivo

Adicionar ao engine um método capaz de retornar, junto com o output renderizado, um relatório estruturado de tudo que aconteceu durante a renderização: quais variáveis foram resolvidas, para quais valores, qual caminho cada condição tomou, quantas vezes cada for iterou, e quais mutações de estado ocorreram via `assign`/`capture`. Adicionalmente, o mesmo método pode rodar validação estrutural do template sem necessitar de um render completo.

O output deve ser serializado em JSON com estrutura estável, de forma que qualquer frontend ou ferramenta externa possa consumir sem precisar conhecer os internos do engine. O design de erros e posições segue o padrão **LSP Diagnostic** para compatibilidade com editores.

---

## Estruturas de Posição

O `SourceLoc` atual tem apenas `Pathname` e `LineNo`. Para suportar highlight preciso no frontend (selecionar exatamente os caracteres da expressão), precisamos de coluna de início e fim.

**Mudança necessária no scanner:** rastrear offset de coluna durante o scan (já é simples porque o scanner itera char a char e incrementa `LineNo` em `\n` — basta fazer o mesmo para `ColNo`).

```go
// Position representa um ponto no source (1-based, compatível com LSP).
type Position struct {
    Line   int `json:"line"`   // 1-based
    Column int `json:"column"` // 1-based
}

// Range é um trecho do source (de Start até End, End exclusivo).
type Range struct {
    Start Position `json:"start"`
    End   Position `json:"end"`
}
```

---

## API

```go
// AuditOptions controla o que RenderAudit coleta.
// Não duplica opções do engine/render — comportamentos como StrictVariables
// são passados via ...RenderOption, igual ao Render normal.
type AuditOptions struct {
    // --- Render trace ---
    TraceVariables   bool // Rastrear {{ expr }} com valor e pipeline de filtros
    TraceConditions  bool // Rastrear {% if/unless/case %} com estrutura de branches e comparações
    TraceIterations  bool // Rastrear {% for/tablerow %} com metadata do loop
    TraceAssignments bool // Rastrear {% assign %} e {% capture %} com valores resultantes

    // Limite de iterações rastreadas por bloco for/tablerow.
    // 0 = sem limite (cuidado com loops grandes).
    // Recomendado: 100. Quando excedido, o campo Truncated do IterationTrace será true.
    MaxIterationTraceItems int
}

// AuditResult é o resultado completo de RenderAudit.
// É sempre retornado não-nil, mesmo quando err != nil — o output pode ser
// parcial se o render foi interrompido, e Diagnostics explica o que ocorreu.
type AuditResult struct {
    Output      string       `json:"output"`      // HTML/texto renderizado (possivelmente parcial)
    Expressions []Expression `json:"expressions"` // Trace de todas as expressões visitadas, em ordem de execução
    Diagnostics []Diagnostic `json:"diagnostics"` // Erros e avisos capturados durante execução
}

// AuditError é retornado quando o render encontrou um ou mais erros.
// Implementa error com uma mensagem resumida, e expõe os erros individuais
// como os mesmos tipos que um render normal retornaria.
type AuditError struct {
    errors []SourceError
}

func (e *AuditError) Error() string    // "render completed with N error(s)"
func (e *AuditError) Errors() []SourceError // cada item é UndefinedVariableError, RenderError, etc.
```

O método é adicionado em `Template`, aceitando os mesmos `RenderOption` que `Render` já aceita:

```go
func (t *Template) RenderAudit(vars Bindings, opts AuditOptions, renderOpts ...RenderOption) (*AuditResult, *AuditError)
```

E um método de análise estática do AST compilado, sem render:

```go
func (t *Template) Validate() (*AuditResult, error)
```

**Sobre `Validate()`:** erros estruturais graves (tag não fechada, sintaxe inválida) já são capturados pelo `ParseTemplate` — se o template chegou a ser criado, esses erros já não existem. O `Validate()` faz análise estática do AST compilado para detectar padrões suspeitos como `empty-block`. É um método separado, não um flag de `AuditOptions`.

**Erros de parse em `AuditError`:** quando `Validate()` detecta um problema estático, o erro vai diretamente para `Diagnostics` e `AuditError.Errors()`. Não aparece em `Expression`, pois não há execução associada. `Expressions` reflete apenas o que foi visitado durante o render.

**Sobre `*AuditError`:** é sempre `nil` se o render completou sem erros. Quando não-nil, `.Error()` dá a mensagem resumida e `.Errors()` dá o slice com os erros individuais tipados — os mesmos tipos que um `Render` normal retornaria se parasse no primeiro erro. Quem quer inspecionar cada erro itera `.Errors()`.

**Sobre `...RenderOption`:** `RenderAudit` é um render normal com observabilidade adicionada por cima. Qualquer `RenderOption` que funciona em `Render` funciona aqui da mesma forma — `WithStrictVariables()`, `WithLaxFilters()`, `WithGlobals()`, qualquer coisa. Não há modos exclusivos do audit. A diferença é que erros que normalmente abortariam o render são capturados como `Diagnostic` e o render continua — todos os erros acumulam em `AuditError.Errors()`.

---

## Estrutura Expression (Render Trace)

Cada item `Expression` representa uma tag ou objeto Liquid que foi visitado durante a renderização. O campo `Kind` é o discriminador. Exatamente um dos campos opcionais estará preenchido.

```go
type ExpressionKind string

const (
    KindVariable   ExpressionKind = "variable"
    KindCondition  ExpressionKind = "condition"
    KindIteration  ExpressionKind = "iteration"
    KindAssignment ExpressionKind = "assignment"
    KindCapture    ExpressionKind = "capture"
)

type Expression struct {
    Source string         `json:"source"` // trecho bruto do template, ex: "{{ customer.name }}"
    Range  Range          `json:"range"`
    Kind   ExpressionKind `json:"kind"`

    // Depth indica a profundidade de aninhamento desta expressão.
    // 0 = nível raiz, 1 = dentro de um {% if %} ou {% for %}, 2 = dentro de bloco aninhado, etc.
    // Permite reconstruir a hierarquia a partir de um array plano sem JSON aninhado.
    Depth int `json:"depth"`

    // Error está preenchido se esta expressão gerou um erro durante a execução.
    // O mesmo erro aparece no array Diagnostics do AuditResult para varredura centralizada.
    Error *Diagnostic `json:"error,omitempty"`

    Variable   *VariableTrace   `json:"variable,omitempty"`
    Condition  *ConditionTrace  `json:"condition,omitempty"`
    Iteration  *IterationTrace  `json:"iteration,omitempty"`
    Assignment *AssignmentTrace `json:"assignment,omitempty"`
    Capture    *CaptureTrace    `json:"capture,omitempty"`
}
```

### VariableTrace

Representa um `{{ expr }}`. Além do valor final, registra o pipeline de filtros passo a passo — resolução do Gemini para saber onde exatamente a cadeia de filtros quebrou ou transformou o valor.

```go
type VariableTrace struct {
    Name    string        `json:"name"`    // "customer.name"
    Parts   []string      `json:"parts"`   // ["customer", "name"]
    Value   any           `json:"value"`   // valor final após todos os filtros
    Pipeline []FilterStep `json:"pipeline"` // passos intermediários (vazio se sem filtros)
}

type FilterStep struct {
    Filter string `json:"filter"` // nome do filtro, ex: "upcase"
    Args   []any  `json:"args"`   // argumentos passados ao filtro, ex: [4, "..."]
    Input  any    `json:"input"`  // valor de entrada deste filtro
    Output any    `json:"output"` // valor de saída deste filtro
}
```

**Exemplo JSON:**
```json
{
  "source": "{{ customer.name | upcase | truncate: 10 }}",
  "range": { "start": {"line": 5, "column": 1}, "end": {"line": 5, "column": 47} },
  "kind": "variable",
  "variable": {
    "name": "customer.name",
    "parts": ["customer", "name"],
    "value": "JOAQUIM...",
    "pipeline": [
      {
        "filter": "upcase",
        "args": [],
        "input": "joaquim silva",
        "output": "JOAQUIM SILVA"
      },
      {
        "filter": "truncate",
        "args": [10, "..."],
        "input": "JOAQUIM SILVA",
        "output": "JOAQUIM..."
      }
    ]
  }
}
```

### ConditionTrace

Representa um bloco `{% if %}`, `{% unless %}` ou `{% case %}` inteiro — do header até o `{% endif %}`/`{% endcase %}`. Captura todos os branches (if, elsif, else) com seus resultados, não só o branch vencedor.

```go
type ConditionTrace struct {
    // Branches lista todos os branches do bloco, em ordem de declaração.
    // Para {% if %}…{% elsif %}…{% else %}…{% endif %}: branch por cláusula.
    // Para {% case %}…{% when %}…{% else %}…{% endcase %}: branch por when/else.
    Branches []ConditionBranch `json:"branches"`
}

type ConditionBranch struct {
    Kind     string          `json:"kind"`            // "if" | "elsif" | "else" | "when" | "unless"
    Range    Range           `json:"range"`           // range do header desta cláusula
    Executed bool            `json:"executed"`        // o corpo deste branch executou?
    Items    []ConditionItem `json:"items,omitempty"` // comparações (vazio no "else")
}

// ConditionItem é uma union: exatamente um dos campos estará preenchido.
type ConditionItem struct {
    Comparison *ComparisonTrace `json:"comparison,omitempty"`
    Group      *GroupTrace      `json:"group,omitempty"`
}

type ComparisonTrace struct {
    Expression string `json:"expression"` // "customer.age >= 18"
    Left       any    `json:"left"`       // valor resolvido do lado esquerdo
    Operator   string `json:"operator"`   // "==", "!=", ">", ">=", "<", "<=", "contains"
    Right      any    `json:"right"`      // valor resolvido do lado direito
    Result     bool   `json:"result"`
}

type GroupTrace struct {
    Operator string          `json:"operator"` // "and" | "or"
    Result   bool            `json:"result"`
    Items    []ConditionItem `json:"items"`
}
```

**Exemplo JSON** para `{% if customer.age >= 18 and customer.active %}…{% elsif plan == "pro" %}…{% else %}…{% endif %}`:
```json
{
  "source": "{% if customer.age >= 18 and customer.active %}",
  "range": { "start": {"line": 10, "column": 1}, "end": {"line": 18, "column": 10} },
  "kind": "condition",
  "depth": 0,
  "condition": {
    "branches": [
      {
        "kind": "if",
        "range": { "start": {"line": 10, "column": 1}, "end": {"line": 10, "column": 48} },
        "executed": true,
        "items": [
          {
            "group": {
              "operator": "and",
              "result": true,
              "items": [
                {
                  "comparison": {
                    "expression": "customer.age >= 18",
                    "left": 25,
                    "operator": ">=",
                    "right": 18,
                    "result": true
                  }
                },
                {
                  "comparison": {
                    "expression": "customer.active",
                    "left": true,
                    "operator": "==",
                    "right": true,
                    "result": true
                  }
                }
              ]
            }
          }
        ]
      },
      {
        "kind": "elsif",
        "range": { "start": {"line": 13, "column": 1}, "end": {"line": 13, "column": 25} },
        "executed": false,
        "items": [
          {
            "comparison": {
              "expression": "plan == \"pro\"",
              "left": "basic",
              "operator": "==",
              "right": "pro",
              "result": false
            }
          }
        ]
      },
      {
        "kind": "else",
        "range": { "start": {"line": 16, "column": 1}, "end": {"line": 16, "column": 9} },
        "executed": false
      }
    ]
  }
}
```

**Nota sobre `{% unless %}`:** o branch tem `kind: "unless"`. O campo `executed` já reflete o resultado final após a inversão.

**Nota sobre `{% case/when %}`:** cada `{% when %}` vira um branch com `kind: "when"`. As `items` contêm uma `ComparisonTrace` implícita com `operator: "=="` comparando o valor do `case` com o valor do `when`. O `{% else %}` do case vira `kind: "else"` normalmente.

### IterationTrace

Representa um bloco `{% for %}` ou `{% tablerow %}`. Emite **um único item** por bloco (não um por iteração), contendo metadata do loop. Para inspecionar variáveis dentro do loop, os `Expression` das iterações internas aparecem naturalmente na sequência do array `expressions` no resultado.

Para evitar explosão de memória em loops grandes, o trace não duplica os nós AST internos por iteração — eles já aparecem linearmente no array. O controle de `MaxIterationTraceItems` limita quantas iterações internas são rastreadas no array `expressions` (não o `IterationTrace` em si).

```go
type IterationTrace struct {
    Variable   string `json:"variable"`             // nome da var do loop: "product"
    Collection string `json:"collection"`            // nome da coleção: "products"
    Length     int    `json:"length"`                // total de itens na coleção
    Limit      *int   `json:"limit,omitempty"`       // valor de limit: se usado
    Offset     *int   `json:"offset,omitempty"`      // valor de offset: se usado
    Reversed   bool   `json:"reversed,omitempty"`    // se reversed: true foi usado
    Truncated  bool   `json:"truncated,omitempty"`   // true se MaxIterationTraceItems foi atingido
    TracedCount int   `json:"traced_count"`          // quantas iterações foram rastreadas
}
```

**Exemplo JSON:**
```json
{
  "source": "{% for product in cart.items limit:3 %}",
  "range": { "start": {"line": 20, "column": 1}, "end": {"line": 20, "column": 41} },
  "kind": "iteration",
  "iteration": {
    "variable": "product",
    "collection": "cart.items",
    "length": 10,
    "limit": 3,
    "offset": null,
    "reversed": false,
    "truncated": false,
    "traced_count": 3
  }
}
```

**Exemplo com truncamento** (loop com 5000 itens, `MaxIterationTraceItems: 100`):
```json
{
  "iteration": {
    "variable": "row",
    "collection": "report.rows",
    "length": 5000,
    "truncated": true,
    "traced_count": 100
  }
}
```

### AssignmentTrace

Representa um `{% assign %}`. Captura o nome da variável, o valor resolvido da expressão, e o pipeline de filtros se houver.

```go
type AssignmentTrace struct {
    Variable string        `json:"variable"` // nome atribuído: "total_price"
    Path     []string      `json:"path,omitempty"` // se dot notation: ["page", "title"]
    Value    any           `json:"value"`    // valor final após filtros
    Pipeline []FilterStep  `json:"pipeline"` // passos de filtro (vazio se sem filtros)
}
```

**Exemplo JSON:**
```json
{
  "source": "{% assign discounted = product.price | times: 0.9 | round %}",
  "range": { "start": {"line": 15, "column": 1}, "end": {"line": 15, "column": 59} },
  "kind": "assignment",
  "assignment": {
    "variable": "discounted",
    "value": 45,
    "pipeline": [
      {
        "filter": "times",
        "args": [0.9],
        "input": 50.0,
        "output": 45.0
      },
      {
        "filter": "round",
        "args": [],
        "input": 45.0,
        "output": 45
      }
    ]
  }
}
```

### CaptureTrace

Representa um bloco `{% capture %}…{% endcapture %}`. Captura o nome da variável e o valor string resultante do bloco.

```go
type CaptureTrace struct {
    Variable string `json:"variable"` // nome capturado: "email_body"
    Value    string `json:"value"`    // conteúdo renderizado do bloco
}
```

**Exemplo JSON:**
```json
{
  "source": "{% capture greeting %}",
  "range": { "start": {"line": 3, "column": 1}, "end": {"line": 3, "column": 23} },
  "kind": "capture",
  "capture": {
    "variable": "greeting",
    "value": "Olá, João! Bem-vindo de volta."
  }
}
```

---

## Diagnostics

Inspirado no **LSP Diagnostic**. O array `diagnostics` é populado por erros que realmente ocorrem — os mesmos erros que um render normal levantaria, mais análise estática disponível pelo `Validate()`. A severidade segue o comportamento padrão do Liquid: se o Liquid levantaria um erro real, é `error`; se o Liquid trataria silenciosamente mas é provável bug, é `warning`; se é apenas observação, é `info`. Não há diagnósticos especulativos sobre o que poderia falhar.

```go
type DiagnosticSeverity string

const (
    SeverityError   DiagnosticSeverity = "error"
    SeverityWarning DiagnosticSeverity = "warning"
    SeverityInfo    DiagnosticSeverity = "info"
)

type Diagnostic struct {
    Range    Range              `json:"range"`
    Severity DiagnosticSeverity `json:"severity"`
    Code     string             `json:"code"`    // identificador de máquina (ver catálogo)
    Message  string             `json:"message"` // mensagem legível para humanos
    Source   string             `json:"source"`  // trecho bruto do template
    Related  []RelatedInfo      `json:"related,omitempty"`
}

type RelatedInfo struct {
    Range   Range  `json:"range"`
    Message string `json:"message"`
}
```

### Catálogo de Códigos de Diagnóstico

`error` para erros reais que o Liquid já levanta. `warning` para comportamentos silenciosos que são provavelmente bugs no template — o Liquid trata silenciosamente, mas uma engine auditável deve expô-los. `info` para observações estáticas.

| Code | Severity | Detectado em | Descrição |
|---|---|---|---|
| `unclosed-tag` | error | parse (Validate) | `{% if %}` sem `{% endif %}` correspondente |
| `unexpected-tag` | error | parse (Validate) | `{% endif %}` sem `{% if %}` que o abriu |
| `syntax-error` | error | parse (Validate) | Sintaxe inválida dentro de uma tag ou objeto |
| `undefined-filter` | error | parse (Validate) | Filtro invocado não está registrado no engine |
| `argument-error` | error | render | Argumentos inválidos para filtro ou tag (ex: `divided_by: 0`) |
| `undefined-variable` | warning | render (strict) | Variável não encontrada nos bindings — apenas quando `WithStrictVariables()` ativo |
| `type-mismatch` | warning | render | Comparação entre tipos incompatíveis; Liquid avalia como false mas é provável bug no template |
| `not-iterable` | warning | render | `{% for %}` sobre valor não-iterável (int, bool, string); Liquid itera zero vezes silenciosamente |
| `nil-dereference` | warning | render | Acesso a propriedade de nil num path encadeado (ex: `customer.address.city` quando `address` é nil) |
| `empty-block` | info | parse (Validate) | Bloco `{% if %}…{% endif %}` sem conteúdo |

**Nota sobre `nil` simples:** variável nil em renderização (`{{ nil_var }}`) e nil em comparação (`{% if nil_var == x %}`) são comportamentos normais e intencionais do Liquid — produzem string vazia e false respectivamente. Não geram diagnósticos. `nil-dereference` é específico para acesso encadeado onde um nó intermediário do path é nil.

**Exemplos JSON de diagnósticos:**

`unclosed-tag` com `Related` apontando onde o fechamento era esperado:
```json
{
  "range": { "start": {"line": 8, "column": 1}, "end": {"line": 8, "column": 14} },
  "severity": "error",
  "code": "unclosed-tag",
  "message": "tag 'if' opened here was never closed",
  "source": "{% if order %}",
  "related": [
    {
      "range": { "start": {"line": 45, "column": 1}, "end": {"line": 45, "column": 1} },
      "message": "expected {% endif %} before end of template"
    }
  ]
}
```

`argument-error` em filtro com divisão por zero:
```json
{
  "range": { "start": {"line": 9, "column": 5}, "end": {"line": 9, "column": 38} },
  "severity": "error",
  "code": "argument-error",
  "message": "divided_by: divided by 0",
  "source": "{{ product.price | divided_by: 0 }}"
}
```

`undefined-variable` (com `WithStrictVariables()` ativo):
```json
{
  "range": { "start": {"line": 22, "column": 5}, "end": {"line": 22, "column": 24} },
  "severity": "warning",
  "code": "undefined-variable",
  "message": "variable 'cart' is not defined",
  "source": "{{ cart.total }}"
}
```

`type-mismatch` durante comparação:
```json
{
  "range": { "start": {"line": 12, "column": 4}, "end": {"line": 12, "column": 34} },
  "severity": "warning",
  "code": "type-mismatch",
  "message": "comparing string \"active\" with integer 1 using '=='; result is always false",
  "source": "{% if user.status == 1 %}"
}
```

`not-iterable` quando o valor não é uma coleção:
```json
{
  "range": { "start": {"line": 20, "column": 1}, "end": {"line": 20, "column": 32} },
  "severity": "warning",
  "code": "not-iterable",
  "message": "'order.status' is string \"pending\"; for loop iterates zero times",
  "source": "{% for item in order.status %}"
}
```

`nil-dereference` em path encadeado:
```json
{
  "range": { "start": {"line": 7, "column": 5}, "end": {"line": 7, "column": 32} },
  "severity": "warning",
  "code": "nil-dereference",
  "message": "'customer.address' is nil; 'city' access renders as empty string",
  "source": "{{ customer.address.city }}"
}
```

---

## Ordem das Expressions no Array

O array `expressions` segue a **ordem de execução** (não de declaração), o que é natural para rastrear fluxo. Isso significa:

- Em um `{% if … %}…{% elsif … %}`, apenas o branch executado terá suas `expressions` internas no array. O `ConditionTrace` do `elsif` omitido aparece mesmo assim (com `result: false`), mas seus filhos não.
- Em um `{% for %}`, as expressions internas se repetem para cada iteração (até `MaxIterationTraceItems`).

Para correlacionar um `Expression` com seu ponto exato no template original, use o `Range`.

---

## Exemplo Completo

**Template:**
```liquid
{% assign title = page.title | upcase %}
<h1>{{ title }}</h1>

{% if customer.age >= 18 %}
  <p>Bem-vindo, {{ customer.name }}!</p>
{% else %}
  <p>Acesso restrito.</p>
{% endif %}

{% for item in cart.items %}
  <li>{{ item.name }} — R$ {{ item.price | times: 1.1 | round }}</li>
{% endfor %}
```

**Bindings:**
```json
{
  "page": {"title": "minha loja"},
  "customer": {"name": "João", "age": 25},
  "cart": {"items": [
    {"name": "Camiseta", "price": 50},
    {"name": "Calça", "price": 120}
  ]}
}
```

**`AuditOptions`:**
```go
AuditOptions{
    TraceVariables:         true,
    TraceConditions:        true,
    TraceIterations:        true,
    TraceAssignments:       true,
    MaxIterationTraceItems: 100,
}
```

**`AuditResult`:**
```json
{
  "output": "<h1>MINHA LOJA</h1>\n\n  <p>Bem-vindo, João!</p>\n\n  <li>Camiseta — R$ 55</li>\n  <li>Calça — R$ 132</li>\n",
  "expressions": [
    {
      "source": "{% assign title = page.title | upcase %}",
      "range": { "start": {"line": 1, "column": 1}, "end": {"line": 1, "column": 41} },
      "kind": "assignment",
      "assignment": {
        "variable": "title",
        "value": "MINHA LOJA",
        "pipeline": [
          { "filter": "upcase", "args": [], "input": "minha loja", "output": "MINHA LOJA" }
        ]
      }
    },
    {
      "source": "{{ title }}",
      "range": { "start": {"line": 2, "column": 5}, "end": {"line": 2, "column": 15} },
      "kind": "variable",
      "variable": { "name": "title", "parts": ["title"], "value": "MINHA LOJA", "pipeline": [] }
    },
    {
      "source": "{% if customer.age >= 18 %}",
      "range": { "start": {"line": 4, "column": 1}, "end": {"line": 4, "column": 27} },
      "kind": "condition",
      "condition": {
        "result": true,
        "items": [
          {
            "comparison": {
              "expression": "customer.age >= 18",
              "left": 25,
              "operator": ">=",
              "right": 18,
              "result": true
            }
          }
        ]
      }
    },
    {
      "source": "{{ customer.name }}",
      "range": { "start": {"line": 5, "column": 15}, "end": {"line": 5, "column": 34} },
      "kind": "variable",
      "variable": { "name": "customer.name", "parts": ["customer", "name"], "value": "João", "pipeline": [] }
    },
    {
      "source": "{% for item in cart.items %}",
      "range": { "start": {"line": 10, "column": 1}, "end": {"line": 10, "column": 29} },
      "kind": "iteration",
      "iteration": {
        "variable": "item",
        "collection": "cart.items",
        "length": 2,
        "truncated": false,
        "traced_count": 2
      }
    },
    {
      "source": "{{ item.name }}",
      "range": { "start": {"line": 11, "column": 7}, "end": {"line": 11, "column": 21} },
      "kind": "variable",
      "variable": { "name": "item.name", "parts": ["item", "name"], "value": "Camiseta", "pipeline": [] }
    },
    {
      "source": "{{ item.price | times: 1.1 | round }}",
      "range": { "start": {"line": 11, "column": 32}, "end": {"line": 11, "column": 69} },
      "kind": "variable",
      "variable": {
        "name": "item.price",
        "parts": ["item", "price"],
        "value": 55,
        "pipeline": [
          { "filter": "times", "args": [1.1], "input": 50, "output": 55.0 },
          { "filter": "round", "args": [], "input": 55.0, "output": 55 }
        ]
      }
    },
    {
      "source": "{{ item.name }}",
      "range": { "start": {"line": 11, "column": 7}, "end": {"line": 11, "column": 21} },
      "kind": "variable",
      "variable": { "name": "item.name", "parts": ["item", "name"], "value": "Calça", "pipeline": [] }
    },
    {
      "source": "{{ item.price | times: 1.1 | round }}",
      "range": { "start": {"line": 11, "column": 32}, "end": {"line": 11, "column": 69} },
      "kind": "variable",
      "variable": {
        "name": "item.price",
        "parts": ["item", "price"],
        "value": 132,
        "pipeline": [
          { "filter": "times", "args": [1.1], "input": 120, "output": 132.0 },
          { "filter": "round", "args": [], "input": 132.0, "output": 132 }
        ]
      }
    }
  ],
  "diagnostics": []
}
```

---

## Plano de Implementação

### Fase 1 — Tracking de Coluna no Scanner

Arquivo: `parser/scanner.go`, `parser/token.go`

- Adicionar `ColNo int` em `SourceLoc`
- O scanner já incrementa `LineNo` ao encontrar `\n`; basta adicionar `ColNo` e resetá-lo para 1 a cada nova linha
- Adicionar `EndLoc SourceLoc` em `Token` para saber onde o token termina (não só começa)
- Todos os formadores de `Range` em fases seguintes dependem disso

**Impacto:** mudança localizada no scanner. `SourceLoc.String()` pode incluir coluna opcionalmente.

### Fase 2 — Diagnostics de Parse

Arquivos: `parser/parser.go`, `parser/error.go`, `liquid.go`

- Converter `parser.Error` em `[]Diagnostic` com `Range` completo
- Erros de block não-fechado já existem como `error`; envolvê-los em `Diagnostic{Code: "unclosed-tag"}`
- Novo método `Template.Validate(AuditOptions) (*AuditResult, error)` — sem render

### Fase 3 — Render Trace

Arquivos: `render/context.go`, `render/render.go`, novos arquivos `render/trace.go`, `render/trace_context.go`

- Criar `traceContext` que wrapa `render.Context` e implementa `render.Context`
- Interceptar `Evaluate()` para capturar `VariableTrace`
- Interceptar renderers de `if/unless/case` para capturar `ConditionTrace`
- Interceptar renderer de `for/tablerow` para capturar `IterationTrace` e aplicar `MaxIterationTraceItems`
- Interceptar renderers de `assign`/`capture` para capturar `AssignmentTrace`/`CaptureTrace`
- Interceptar cada `FilterStep` na cadeia de filtros para popular `Pipeline`

### Fase 4 — Diagnostics de Runtime

Arquivos: `render/trace_context.go`

- Interceptar erros reais que acontecem durante o render e convertê-los em `Diagnostic` estruturado em vez de abortar
- `UndefinedVariableError` (quando StrictVariables ativo) → `Diagnostic{Code: "undefined-variable"}`, render continua
- Incompatibilidade de tipos em comparações → `Diagnostic{Code: "type-mismatch"}`
- `ArgumentError` (filtro com args inválidos, divisão por zero) → `Diagnostic{Code: "argument-error"}`
- `{% for %}` sobre tipo não iterável → `Diagnostic{Code: "not-iterable"}`, itera zero vezes (comportamento normal)
- Acesso a propriedade de nil em path encadeado → `Diagnostic{Code: "nil-dereference"}`, retorna empty (comportamento normal)
- Acumular todos os erros em `AuditError.errors`; derivar o `*AuditError` retornado

### Fase 5 — API Final e Expose Público

Arquivo: `liquid.go`, `template.go`

- `Template.RenderAudit(vars Bindings, opts AuditOptions, renderOpts ...RenderOption) (*AuditResult, *AuditError)` — faz tudo
- `Template.Validate() (*AuditResult, error)` — só análise estática do AST
- Tipos `AuditResult`, `AuditOptions`, `AuditError`, `Expression`, `Diagnostic`, `Position`, `Range` exportados em `liquid.go`
- JSON tags em todos os tipos para serialização direta

---

## Notas de Design

**Por que não há `Validate bool` no `AuditOptions`:** Erros estruturais graves (tag não fechada, sintaxe inválida) já são capturados pelo `ParseTemplate` — se o template foi criado com sucesso, esses erros não existem. Não há nada a validar estruturalmente durante um render audit. O `Validate()` é um método separado para análise estática do AST compilado.

**Por que o erro retornado é `*AuditError` e não `SourceError`:** O render audit não para no primeiro erro — acumula todos os erros encontrados. O `*AuditError` reflete isso: `.Error()` dá o resumo, `.Errors()` dá o slice completo com os mesmos tipos que um render normal retornaria um por um.

**Por que nil e for-sobre-nil geram `warning` e não `error`:** São comportamentos silenciosos do Liquid que nunca abortam o render — a severity warning reflete exatamente isso. O Liquid silenciosamente retorna vazio ou itera zero; o audit apenas torna isso visível. `nil-dereference` é específico para paths encadeados (`a.b.c` onde `b` é nil), não para variável nil simples.

**Por que o PII não está aqui:** A responsabilidade de redação de dados sensíveis é da camada que chama o engine — quem sabe que `customer.cpf` é sensível é quem monta os bindings, não o engine. O engine expõe o trace; o caller decide o que logar ou trafegar.

**Por que `...RenderOption` e não opções novas:** O audit render é um render normal com observabilidade adicionada. Qualquer `RenderOption` do engine funciona aqui sem nenhuma mudança — `WithStrictVariables()`, `WithLaxFilters()`, `WithGlobals()`. Não existe modo exclusivo do audit. Garante paridade de comportamento: `RenderAudit` nunca renderiza diferente do `Render` com as mesmas opções.

**Por que `Depth` em vez de JSON aninhado:** LSP usa `children []DocumentSymbol` (JSON aninhado real) no `DocumentSymbol`. Para o nosso caso, o array de expressions é uma timeline de execução — iteração linear é o caso de uso principal, não navegação de árvore. Com `depth`, o frontend itera o array uma vez e constrói a árvore se precisar: filhos de um nó são os próximos itens com `depth = nó.depth + 1` até o próximo item com `depth <= nó.depth`. JSON aninhado não é adequado para uma timeline.

**Por que `IterationTrace` não duplica o nó interno por iteração:** Se um for tem 1000 itens e o corpo tem 10 expressions, rastrear tudo daria 10.000 expression objects. O `MaxIterationTraceItems` limita o número de iterações rastreadas, mantendo o array razoável. O `IterationTrace` em si sempre aparece com `length` e `traced_count`.

**Sobre `NodeID` para "Inspect Element":** Uma extensão futura seria injetar comentários HTML `<!-- lid:RANGE -->` no output quando trace está ativo, permitindo que um frontend correlacione um trecho do HTML renderizado com o `Range` da expression que o gerou. Não está no escopo desta fase.
