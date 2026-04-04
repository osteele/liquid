# Plano Swarm — Fundações (PRE)

> Este documento cobre **apenas as dependências compartilhadas** que precisam existir antes que os itens do [implementation-checklist.md](implementation-checklist.md) possam ser executados individualmente por agentes.
>
> Cada PRE é uma fundação técnica, não uma feature visível ao usuário final.
> Após todas as PREs estarem concluídas, cada item do implementation-checklist pode ser delegado a um agente independente.

---

## Diagrama de dependências

```
PRE-A (expression layer)
  └─→ checklist: keyword arg filters (default, compact, uniq)
  └─→ checklist: empty/blank literals  [junto com PRE-F]
  └─→ checklist: string escapes, <>, not, case/when or

PRE-B (filter context-aware)   ⚠️ coordenar com PRE-A em expressions/context.go
  └─→ checklist: _exp filters (where_exp, reject_exp, group_by_exp, find_exp, find_index_exp, has_exp)

PRE-C (sub-contexto isolado)
  └─→ PRE-D
  └─→ checklist: render tag
  └─→ checklist: layout/block

PRE-D (globals layer)          [dep: PRE-C]
  └─→ checklist: render tag (globals propagation)
  └─→ checklist: globals engine option

PRE-E (error type system)
  └─→ checklist: B4, B6, tipos de erro exportados

PRE-F (EmptyDrop/BlankDrop)
  └─→ checklist: empty/blank literals  [junto com PRE-A]

PRE-G (echo tag)               [trivial — pode ser feito pelo mesmo agente que fizer liquid tag]
  └─→ checklist: liquid tag multi-linha

Todos os demais itens do checklist não têm deps de PRE.
```

---

## Wave 0-α — 6 Fundações em paralelo

Esses seis PREs não dependem entre si e podem ser desenvolvidos por agentes diferentes ao mesmo tempo.

---

### PRE-A · Expression layer: scanner, parser, yacc

**O que é:** Conjunto de mudanças em `expressions/scanner.rl` e `expressions/expressions.y` que afetam o pipeline de geração de código (`ragel` → `scanner.go`, `goyacc` → `y.go`). Por gerarem arquivos derivados em comum, **todo trabalho nessa camada deve ser feito por um único agente** para evitar conflitos de merge.

**Mudanças agrupadas aqui:**
- Keyword args em filter calls: `filter: val, key: val2` (parser de argumentos)
- `empty` e `blank` como keywords do scanner (não como nomes de variável)
- String escape sequences: `\n`, `\"`, `\'` dentro de string literals
- Operador `<>` como alias de `!=`
- Operador unário `not` na gramática yacc
- `or` em `when` de `case/when`

**Arquivos em escopo:**
```
expressions/scanner.rl       (fonte ragel — editar este)
expressions/scanner.go       (gerado por ragel — regenerar após editar .rl)
expressions/expressions.y    (fonte goyacc — editar este)
expressions/y.go             (gerado por goyacc — regenerar após editar .y)
expressions/parser.go        (pode precisar de ajustes)
expressions/context.go       (⚠️ conflito potencial com PRE-B — ver seção de conflitos)
```

**Desbloqueia** (no implementation-checklist.md):
- `default: fallback, allow_false: true`
- `compact: "field"` e `uniq: "field"`
- `empty` e `blank` como literais especiais (junto com PRE-F)
- String escapes: `\n`, `\"`, `\'`
- Operador `<>` como alias de `!=`
- Operador unário `not`
- `case/when` com `or`

---

### PRE-B · Infraestrutura de filtros context-aware

**O que é:** Atualmente filtros têm assinatura `func(value any, args ...any) (any, error)` sem acesso ao render context. Os `_exp` filters precisam avaliar expressões Liquid por item do array, o que requer acesso ao `render.Context`. É necessário criar um mecanismo de registro de filtros que receba o contexto como parâmetro extra.

**Abordagem sugerida:** novo tipo de registro `AddContextFilter(name string, fn func(ctx Context, value any, args ...any) any)`. No dispatch interno (`expressions/context.go`, `ApplyFilter`), verificar se o filtro é "context-aware" e injetar o contexto se for.

**Arquivos em escopo:**
```
expressions/context.go       (ApplyFilter — ⚠️ conflito com PRE-A)
expressions/functional.go    (filter dispatch / wrapping)
render/config.go             (AddFilter / AddContextFilter)
```

**Desbloqueia** (no implementation-checklist.md):
- `where_exp`, `reject_exp`, `group_by_exp`, `find_exp`, `find_index_exp`, `has_exp`

---

### PRE-C · Sub-contexto isolado no render.Context

**O que é:** Ao usar `render` tag, o template renderizado não deve herdar variáveis do pai — apenas o que for passado explicitamente. Atualmente `nodeContext` faz `maps.Copy` dos bindings (o comentário no código inclusive marca `TODO: this isn't really the right place for this`). Precisa de um método `SpawnIsolated(bindings map[string]any) nodeContext` em `render/node_context.go`.

**Diferença de include vs render:**
- `include`: compartilha escopo completo (comportamento atual)
- `render`: escopo isolado — apenas args passados explicitamente são visíveis no partial

**Arquivos em escopo:**
```
render/node_context.go       (adicionar SpawnIsolated)
render/context.go            (expor via Context interface se necessário)
```

**Desbloqueia** (no implementation-checklist.md):
- `render` tag (escopo isolado)
- `layout`/`block` (herança de template)
- Também é pré-requisito de PRE-D

---

### PRE-E · Sistema de tipos de erro exportados

**O que é:** Hoje todos os erros de render e parse retornam tipos genéricos ou `SourceError` sem distinção de origem. Precisa-se de tipos exportados distintos para permitir tratamento programático pelo chamador.

**Tipos a definir:**
```go
type ParseError struct { SourceError }
type RenderError struct { SourceError }
type UndefinedVariableError struct { ... }   // deve incluir o nome literal da variável
type ZeroDivisionError struct { ... }
```

**Arquivos em escopo:**
```
render/error.go
parser/error.go
```

**Desbloqueia** (no implementation-checklist.md):
- B4: tipos distintos de erro
- B6: mensagens de erro de variável (o `UndefinedVariableError` com nome literal da variável resolve parte do problema)
- `ZeroDivisionError` em `modulo` e `divided_by`
- `SyntaxError`, `ArgumentError`, `ContextError`

---

### PRE-F · EmptyDrop e BlankDrop (camada values)

**O que é:** A camada `values/` precisa de dois singletons com semântica especial de comparação. A parte de scanner (fazer `empty` e `blank` serem reconhecidos como keywords e não como nomes de variável) fica em PRE-A. Este PRE cobre apenas os tipos Go correspondentes.

**Semântica:**
- `empty`: compara como igual a `""`, `[]any{}`, `map{}` (qualquer valor "vazio")
- `blank`: como `empty`, mais `nil`, `false`, strings com só whitespace

**Arquivos em escopo:**
```
values/value.go              (ou novo values/emptydrop.go)
values/compare.go            (Equal deve reconhecer EmptyDrop/BlankDrop)
values/predicates.go         (IsBlank, IsEmpty)
```

**Desbloqueia** (no implementation-checklist.md):
- `empty` e `blank` como literais especiais (junto com PRE-A)
- `EmptyDrop`, `BlankDrop` como singletons exportados na API pública

---

### PRE-G · `echo` tag

**O que é:** Registrar a tag `echo` que avalia uma expressão e escreve no writer — idêntico a `{{ expr }}` mas como tag. É o menor PRE mas é bloqueante para a `liquid` multi-line tag, que depende de `echo` para produzir output linha a linha.

**Arquivos em escopo:**
```
tags/standard_tags.go
```

**Desbloqueia** (no implementation-checklist.md):
- `liquid` tag multi-linha

---

## Wave 0-β — Fundação sequencial

---

### PRE-D · Camada de globals separada · [dep: PRE-C]

**O que é:** Um nível de variáveis que persiste através de `SpawnIsolated()` — ao contrário do escopo de bindings que é cortado num sub-contexto `render`. O caller passa globals via `engine.SetGlobals(map[string]any{})` e elas são acessíveis em qualquer partial renderizado via `render` tag.

**Comportamento no lookup:** `bindings[key]` → se não encontrado, `globals[key]`.

**Arquivos em escopo:**
```
render/node_context.go       (adicionar campo globals, propagar em SpawnIsolated)
render/config.go             (opcional: globals no Config)
engine.go                    (SetGlobals / GetGlobals)
```

**Desbloqueia** (no implementation-checklist.md):
- `render` tag com globals propagation
- `globals` como engine option

---

## Conflitos conhecidos entre PREs

| Risco | PREs | Arquivo | Resolução |
|-------|------|---------|-----------|
| Alto | PRE-A ↔ PRE-B | `expressions/context.go` (`ApplyFilter`) | Fazer PRE-A primeiro; PRE-B aplica sobre o resultado |
| Baixo | PRE-C ↔ PRE-D | `render/node_context.go` | Sequencial por design (PRE-D depende de PRE-C) |
