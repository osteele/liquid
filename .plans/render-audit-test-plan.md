# Test Plan: RenderAudit — Cobertura Completa

## Status: 34 testes existem hoje. Este plano mapeia ~200+ novos testes necessários.

---

## Inventário de Testes Existentes (34)

| # | Teste | Cobre |
|---|---|---|
| 1 | TraceVariables_simple | Nome, Value, Kind básico |
| 2 | TraceVariables_noTrace | Flag desligada → 0 expressions |
| 3 | TraceVariables_filterPipeline | 1 filtro (upcase), Input/Output |
| 4 | TraceVariables_depth | Depth=1 dentro de if |
| 5 | TraceConditions_if_taken | Branches, Executed=true/false |
| 6 | TraceConditions_else_taken | Else branch executado |
| 7 | TraceConditions_unless | Branch kind="unless" |
| 8 | TraceIterations_basic | Variable, Collection, Length |
| 9 | TraceAssignments_assign | Variable, Value |
| 10 | TraceAssignments_capture | Variable, Value string |
| 11 | Combined | assign + variable juntos |
| 12 | Error_strictVariables | AuditError retornado |
| 13 | ResultNonNilOnError | result nunca nil |
| 14 | Validate_emptyIF | empty-block diagnostic |
| 15 | Validate_nonEmpty | Sem false positive |
| 16 | Position_lineNumber | Start.Line correto |
| 17 | AssignSourceLoc | Range.Start.Line > 0 |
| 18 | AssignFilterPipeline | Pipeline em assign |
| 19 | MaxIterItems_TracedCount | Truncated, TracedCount |
| 20 | NoMaxIterItems_AllTraced | Sem truncation |
| 21 | ConditionComparisons_simple | Operator >=, Left, Right, Result |
| 22 | ConditionComparisons_else_noComparisons | Else sem Items |
| 23 | ConditionComparisons_equality | Operator == |
| 24 | ConditionComparisons_groupTrace_and | GroupTrace and com 2 filhos |
| 25 | Diagnostic_undefinedVariable | Code, Severity |
| 26 | Diagnostic_argumentError | divided_by 0 |
| 27 | Diagnostic_typeMismatch | string == int |
| 28 | Diagnostic_notIterable | for sobre string |
| 29 | Diagnostic_nilDereference | a.b.c com b=nil |
| 30 | ConditionComparisons_expressionField | Expression não vazio |
| 31 | Diagnostic_typeMismatch_hasRange | Range preenchido |
| 32 | Diagnostic_nilDereference_hasRange | Range preenchido |
| 33 | Diagnostic_notIterable_hasRange | Range span |
| 34 | Validate_UndefinedFilter | undefined-filter |

---

## 1. VariableTrace — `{{ expr }}`

### 1.1 Atributos básicos (Name, Parts, Value)

| ID | Teste | O que valida |
|---|---|---|
| V01 | `{{ x }}` — variável simples | Name="x", Parts=["x"], Value correto |
| V02 | `{{ customer.name }}` — dot access | Name="customer.name", Parts=["customer","name"] |
| V03 | `{{ a.b.c.d }}` — deep dot access | Parts=["a","b","c","d"], Name="a.b.c.d" |
| V04 | `{{ items[0] }}` — array index access | Name contém bracket, Value correto |
| V05 | `{{ "literal" }}` — string literal | Name="\"literal\"" ou similar, Value="literal" |
| V06 | `{{ 42 }}` — integer literal | Value=42 (int) |
| V07 | `{{ 3.14 }}` — float literal | Value=3.14 (float64) |
| V08 | `{{ true }}` — boolean literal | Value=true |
| V09 | `{{ false }}` — boolean literal | Value=false |
| V10 | `{{ nil }}` — nil literal | Value=nil |
| V11 | `{{ blank }}` — blank | Value="" ou nil |
| V12 | `{{ empty }}` — empty | Valor correspondente |
| V13 | Variável undefined (sem strict) | Value=nil, sem erro |
| V14 | Variável undefined (com strict) | Value=nil, Error preenchido, Diagnostic code="undefined-variable" |
| V15 | Múltiplas variáveis no template | len(Expressions) correto, todas KindVariable |
| V16 | `{{ hash["key"] }}` — bracket string access | Value correto |

### 1.2 Pipeline de Filtros

| ID | Teste | O que valida |
|---|---|---|
| VP01 | Sem filtros | Pipeline=[] (vazio, não nil) |
| VP02 | Um filtro sem args (`upcase`) | len(Pipeline)=1, Filter, Input, Output |
| VP03 | Um filtro com 1 arg (`truncate: 10`) | Args=[10], Input/Output corretos |
| VP04 | Um filtro com múltiplos args (`truncate: 10, "..."`) | Args=[10, "..."] |
| VP05 | Cadeia de 2 filtros (`upcase \| truncate: 5`) | len(Pipeline)=2, Output[0]=Input[1] |
| VP06 | Cadeia de 3+ filtros (`downcase \| replace \| truncate`) | Encadeamento completo |
| VP07 | Filtro `default` com valor nil | Input=nil, Output=default value |
| VP08 | Filtro `split` (retorna array) | Output é []string |
| VP09 | Filtro `size` em string | Output é int |
| VP10 | Filtro `size` em array | Output é int |
| VP11 | Filtro `times` (math) | Input/Output numéricos |
| VP12 | Filtro `round` | Input float → Output int |
| VP13 | Filtro `join` em array | Input=[]string, Output=string |
| VP14 | Filtro `date` | Input=string/time, Output=string formatada |
| VP15 | Filtro `first` em array | Output é primeiro elemento |
| VP16 | Filtro `last` em array | Output é último elemento |
| VP17 | Filtro `map` em array de maps | Output é slice de valores |
| VP18 | Filtro `where` em array | Output é array filtrado |
| VP19 | Filtro `sort` em array | Input desordenado → Output ordenado |
| VP20 | Filtro `reverse` em array | Output invertido |
| VP21 | Filtro `compact` (remove nils) | Input com nils → Output sem nils |
| VP22 | Filtro `uniq` | Remove duplicatas |
| VP23 | Filtro undefined com LaxFilters | Sem erro, value passthrough |
| VP24 | Pipeline com filtro que erra (ex: `divided_by: 0`) | Error no Expression, Diagnostic |

### 1.3 Source e Range

| ID | Teste | O que valida |
|---|---|---|
| VR01 | Source contém delimitadores `{{ ... }}` | Source exato |
| VR02 | Range.Start.Line correto (1a linha) | Line=1 |
| VR03 | Range.Start.Line correto (3a linha) | Line=3 |
| VR04 | Range.Start.Column correto | Column preciso |
| VR05 | Range.End > Range.Start | End é depois do Start |
| VR06 | Range.End.Column = Start.Column + len(source) (single-line) | Cálculo preciso |
| VR07 | Múltiplas expressions no mesmo template, Ranges não se sobrepõem | Sem overlap |

### 1.4 Depth

| ID | Teste | O que valida |
|---|---|---|
| VD01 | Top-level variable | Depth=0 |
| VD02 | Dentro de `{% if %}` | Depth=1 |
| VD03 | Dentro de `{% for %}` | Depth=1 |
| VD04 | Dentro de `{% if %}{% if %}` aninhado | Depth=2 |
| VD05 | Dentro de `{% for %}{% if %}` aninhado | Depth=2 |
| VD06 | Após sair do bloco, volta a Depth=0 | Depth correto |

---

## 2. ConditionTrace — `{% if %}`, `{% unless %}`, `{% case %}`

### 2.1 Estrutura de Branches

| ID | Teste | O que valida |
|---|---|---|
| C01 | `{% if x %}...{% endif %}` (sem else) | 1 branch, kind="if" |
| C02 | `{% if x %}...{% else %}...{% endif %}` | 2 branches: "if" + "else" |
| C03 | `{% if x %}...{% elsif y %}...{% endif %}` | 2 branches: "if" + "elsif" |
| C04 | `{% if x %}...{% elsif y %}...{% else %}...{% endif %}` | 3 branches: "if"+"elsif"+"else" |
| C05 | `{% if x %}...{% elsif y %}...{% elsif z %}...{% else %}...{% endif %}` | 4 branches |
| C06 | `{% unless x %}...{% endunless %}` | 1 branch, kind="unless" |
| C07 | `{% unless x %}...{% else %}...{% endunless %}` | 2 branches: "unless"+"else" |
| C08 | `{% case x %}{% when "a" %}...{% when "b" %}...{% endcase %}` | 2 branches com kind="when" |
| C09 | `{% case x %}{% when "a" %}...{% when "b" %}...{% else %}...{% endcase %}` | 3 branches: "when"+"when"+"else" |
| C10 | `{% case x %}{% when "a","b" %}...{% endcase %}` | When com múltiplos valores |

### 2.2 Executed flag

| ID | Teste | O que valida |
|---|---|---|
| CE01 | If true → if Executed=true | |
| CE02 | If false, else → else Executed=true, if Executed=false | |
| CE03 | If false, elsif true → elsif Executed=true | |
| CE04 | If false, elsif false, else → else Executed=true | |
| CE05 | Unless false → unless Executed=true | |
| CE06 | Unless true → unless Executed=false | |
| CE07 | Case match first when → first when Executed=true, rest=false | |
| CE08 | Case match second when → second Executed=true, first=false | |
| CE09 | Case no match, else → else Executed=true | |
| CE10 | Case no match, no else → all Executed=false | |

### 2.3 ComparisonTrace

| ID | Teste | O que valida |
|---|---|---|
| CC01 | `x == 1` | Operator="==", Left, Right, Result |
| CC02 | `x != 1` | Operator="!=" |
| CC03 | `x > 5` | Operator=">" |
| CC04 | `x < 5` | Operator="<" |
| CC05 | `x >= 5` | Operator=">=" |
| CC06 | `x <= 5` | Operator="<=" |
| CC07 | `arr contains "a"` | Operator="contains" |
| CC08 | `str contains "sub"` | Operator="contains" em string |
| CC09 | Expressão com resultado true | Result=true |
| CC10 | Expressão com resultado false | Result=false |
| CC11 | Left/Right são tipos corretos (int, string, bool, nil) | Tipos preservados |
| CC12 | Expression field contém texto raw | Expression="x == 1" ou similar |
| CC13 | Truthiness simples `{% if x %}` (sem operador) | ComparisonTrace com operador implícito == true |

### 2.4 GroupTrace (and/or)

| ID | Teste | O que valida |
|---|---|---|
| CG01 | `a and b` (ambos true) | Operator="and", Result=true, 2 Items |
| CG02 | `a and b` (um false) | Result=false |
| CG03 | `a or b` (ambos false) | Operator="or", Result=false |
| CG04 | `a or b` (um true) | Result=true |
| CG05 | `a and b and c` (3 termos) | Nesting correto |
| CG06 | `a or b or c` (3 termos) | Nesting correto |
| CG07 | `a and b or c` (mixed) | Precedência correta (Liquid: left-to-right, sem precedência) |
| CG08 | `a > 1 and b < 10` com comparisons dentro do group | Group.Items[0].Comparison != nil |
| CG09 | Group aninhado `(a and b) or c` — Liquid não tem parênteses, mas testa o flat left-to-right | Estrutura correta |

### 2.5 Branch Range e Source

| ID | Teste | O que valida |
|---|---|---|
| CB01 | Branch[0].Range.Start aponta para `{% if ... %}` | Range preciso |
| CB02 | Branch[1].Range.Start aponta para `{% else %}` | Range do else |
| CB03 | Branch para elsif aponta para `{% elsif ... %}` | Range do elsif |
| CB04 | Condition Expression.Range cobre do {% if %} ao {% endif %} | Range total |
| CB05 | Condition Expression.Source contém o header do if | Source preciso |

### 2.6 Depth em Condições

| ID | Teste | O que valida |
|---|---|---|
| CD01 | If top-level | Depth=0 |
| CD02 | If dentro de for | Depth=1 |
| CD03 | If aninhado | Depth reflete nesting |

### 2.7 Condição com erro

| ID | Teste | O que valida |
|---|---|---|
| CR01 | Variável undefined em condição (strict) | Diagnostic, render continua |

---

## 3. IterationTrace — `{% for %}`, `{% tablerow %}`

### 3.1 Atributos Básicos

| ID | Teste | O que valida |
|---|---|---|
| I01 | `{% for item in items %}` | Variable="item", Collection="items" |
| I02 | Coleção vazia | Length=0, TracedCount=0 |
| I03 | Coleção com 1 item | Length=1, TracedCount=1 |
| I04 | Coleção com 100 itens | Length=100 |
| I05 | `{% for item in hash %}` — iteração sobre map | Funciona, Length correto |
| I06 | `{% for i in (1..5) %}` — range | Length=5, Variable="i", Collection="(1..5)" |
| I07 | `{% for i in (1..0) %}` — range vazio/invertido | Length=0 |

### 3.2 Limit, Offset, Reversed

| ID | Teste | O que valida |
|---|---|---|
| IL01 | `limit:3` com 5 itens | Limit=ptr(3), Length=5 |
| IL02 | `offset:2` com 5 itens | Offset=ptr(2), Length=5 |
| IL03 | `limit:2 offset:1` combinado | Ambos preenchidos |
| IL04 | `reversed` com array | Reversed=true |
| IL05 | Sem limit/offset/reversed | Limit=nil, Offset=nil, Reversed=false |
| IL06 | `limit:0` | Limit=ptr(0) |
| IL07 | `offset:continue` (se suportado) | Testa ou documenta limitação |

### 3.3 MaxIterationTraceItems / Truncation

| ID | Teste | O que valida |
|---|---|---|
| IT01 | MaxIterItems=0, 10 itens | Truncated=false, TracedCount=10 |
| IT02 | MaxIterItems=5, 10 itens | Truncated=true, TracedCount=5 |
| IT03 | MaxIterItems=10, 5 itens | Truncated=false, TracedCount=5 |
| IT04 | MaxIterItems=1, 100 itens | Truncated=true, TracedCount=1, inner expressions limitadas |
| IT05 | MaxIterItems limita inner expressions mas não o output | Output completo mesmo com truncation |
| IT06 | Nested for — cada for tem seu próprio TracedCount | Correto por bloco |
| IT07 | MaxIterItems com coleção vazia | Truncated=false, TracedCount=0 |

### 3.4 Inner Expressions em For

| ID | Teste | O que valida |
|---|---|---|
| IF01 | For com `{{ item }}` — variable trace aparece por iteração | len(var expressions) = Length |
| IF02 | For com `{% if item > 2 %}` — condition trace por iteração | Conditions dentro do for |
| IF03 | For com `{% assign x = item %}` — assign por iteração | Assignments |
| IF04 | For aninhado — inner expressions do for interno | Depth correto |
| IF05 | MaxIterItems trunca inner expressions | Sem traces após o corte |
| IF06 | `forloop` variables (forloop.first, forloop.last, forloop.index) | Acessíveis como variáveis |

### 3.5 Tablerow

| ID | Teste | O que valida |
|---|---|---|
| TR01 | `{% tablerow item in items %}` | IterationTrace com Variable, Collection |
| TR02 | Tablerow com `cols:3` | Output correto, trace registrado |
| TR03 | Tablerow com limit/offset | Limit/Offset preenchidos |

### 3.6 Source, Range, Depth

| ID | Teste | O que valida |
|---|---|---|
| IR01 | Source contém `{% for ... %}` | Source preciso |
| IR02 | Range.Start/End corretos | Posições |
| IR03 | Depth=0 top-level, 1 aninhado | Depth correto |

### 3.7 For com Erros / Edge Cases

| ID | Teste | O que valida |
|---|---|---|
| IE01 | For sobre int — not-iterable | Diagnostic warning |
| IE02 | For sobre bool — not-iterable | Diagnostic warning |
| IE03 | For sobre nil — not-iterable ou zero iterations | Comportamento |
| IE04 | For sobre string — not-iterable | Diagnostic warning |
| IE05 | For-else ({% for %}...{% else %}...{% endfor %}) coleção vazia | Else executado |

---

## 4. AssignmentTrace — `{% assign %}`

### 4.1 Atributos Básicos

| ID | Teste | O que valida |
|---|---|---|
| A01 | `{% assign x = "hello" %}` | Variable="x", Value="hello" |
| A02 | `{% assign x = 42 %}` | Value=42 (int) |
| A03 | `{% assign x = 3.14 %}` | Value=3.14 (float) |
| A04 | `{% assign x = true %}` | Value=true |
| A05 | `{% assign x = false %}` | Value=false |
| A06 | `{% assign x = nil %}` | Value=nil |
| A07 | `{% assign x = var %}` — de outra variável | Value resolve para valor do binding |
| A08 | `{% assign x = a.b.c %}` — dot access | Value resolve para nested |
| A09 | Path field vazio para assign simples | Path=nil ou [] |

### 4.2 Pipeline em Assign

| ID | Teste | O que valida |
|---|---|---|
| AP01 | `{% assign x = name \| upcase %}` — 1 filtro | Pipeline[0] completo |
| AP02 | `{% assign x = price \| times: 0.9 \| round %}` — cadeia | 2 FilterSteps |
| AP03 | `{% assign x = arr \| sort \| first %}` — array filters | Pipeline correto |
| AP04 | `{% assign x = "a,b,c" \| split: "," %}` — retorna array | Value=["a","b","c"], Output do step |
| AP05 | Sem filtros | Pipeline=[] (vazio) |
| AP06 | Filtro que erra — assign com erro | Error/Diagnostic |

### 4.3 Source, Range, Depth

| ID | Teste | O que valida |
|---|---|---|
| AR01 | Source contém `{% assign ... %}` completo | String exata |
| AR02 | Range preciso | Line, Column |
| AR03 | Assign dentro de if | Depth=1 |
| AR04 | Assign dentro de for | Depth=1, repetido por iteração |

### 4.4 Múltiplos Assigns

| ID | Teste | O que valida |
|---|---|---|
| AM01 | 3 assigns em sequência | 3 Expression com KindAssignment, na ordem |
| AM02 | Assign seguido de uso (`{% assign x %}{{ x }}`) | Assignment antes de Variable na lista |
| AM03 | Reassign da mesma variável | Dois assignment traces, valores diferentes |

---

## 5. CaptureTrace — `{% capture %}`

### 5.1 Atributos Básicos

| ID | Teste | O que valida |
|---|---|---|
| CP01 | `{% capture x %}Hello{% endcapture %}` | Variable="x", Value="Hello" |
| CP02 | Capture com expressions dentro (`{% capture x %}{{ name }}!{% endcapture %}`) | Value contém resultado renderizado |
| CP03 | Capture com múltiplas linhas | Value contém todo o conteúdo |
| CP04 | Capture vazio (`{% capture x %}{% endcapture %}`) | Value="" |
| CP05 | Capture com tags dentro (`{% capture x %}{% if true %}yes{% endif %}{% endcapture %}`) | Value="yes" |

### 5.2 Source, Range, Depth

| ID | Teste | O que valida |
|---|---|---|
| CPR01 | Source contém `{% capture ... %}` | String precisa |
| CPR02 | Range aponta para abertura do capture | Posição correta |
| CPR03 | Depth dentro de bloco | Depth correto |

### 5.3 Capture com Inner Traces

| ID | Teste | O que valida |
|---|---|---|
| CPI01 | Capture com `{{ var }}` dentro — inner variable trace aparece? | Inner expressions rastreadas |
| CPI02 | Capture com `{% if %}` dentro — inner condition trace | Conditions no array |
| CPI03 | Capture seguido de `{{ x }}` usando o valor capturado | Variable trace com valor do capture |

---

## 6. Diagnostics — Catálogo Completo

### 6.1 Erros de Runtime

| ID | Teste | Diagnostic Code | O que valida |
|---|---|---|---|
| D01 | `{{ ghost }}` com StrictVariables | `undefined-variable` | severity=warning, message contém nome |
| D02 | `{{ a.b }}` com a undefined, StrictVariables | `undefined-variable` | Path |
| D03 | Múltiplas variáveis undefined | Múltiplos diagnostics | Acumula tudo |
| D04 | `{{ x \| divided_by: 0 }}` | `argument-error` | severity=error |
| D05 | `{{ x \| modulo: 0 }}` | `argument-error` | severity=error |
| D06 | Filtro com argumento inválido | `argument-error` | Message descritiva |
| D07 | `{% if "str" == 1 %}` type mismatch | `type-mismatch` | severity=warning, message com tipos |
| D08 | `{% if "str" > 1 %}` type mismatch com > | `type-mismatch` | Operator no message |
| D09 | `{% if nil == 1 %}` — nil vs int | Sem diagnostic (nil comparison é normal) | Sem warning |
| D10 | `{% for x in 42 %}` not iterable int | `not-iterable` | severity=warning |
| D11 | `{% for x in true %}` not iterable bool | `not-iterable` | severity=warning |
| D12 | `{% for x in "str" %}` not iterable string | `not-iterable` | severity=warning |
| D13 | `{{ a.b.c }}` com b=nil | `nil-dereference` | severity=warning, message com path |
| D14 | `{{ a.b.c.d }}` com b=nil (deep nil) | `nil-dereference` | Property="c" |
| D15 | `{{ nil_var }}` — nil simples | Sem diagnostic | Comportamento normal |
| D16 | `{% if nil_var %}` — nil em condição | Sem diagnostic | Comportamento normal |

### 6.2 Diagnostics Range e Source

| ID | Teste | O que valida |
|---|---|---|
| DR01 | Todo Diagnostic tem Range.Start.Line >= 1 | Nunca 0 |
| DR02 | Diagnostic.Source é o trecho bruto | Inclui delimitadores |
| DR03 | Diagnostic em linha 5 do template | Line=5 |
| DR04 | Múltiplos diagnostics em linhas diferentes | Cada um com seu Range |

### 6.3 Diagnostics e Expressions juntos

| ID | Teste | O que valida |
|---|---|---|
| DE01 | Variable com erro → Expression.Error != nil e Diagnostic no array | Dupla referência |
| DE02 | Expression.Error e Diagnostics[i] têm o mesmo Code | Consistência |
| DE03 | Múltiplos erros: len(Diagnostics) == len(AuditError.Errors()) | Paridade |

### 6.4 Render Continua Após Erro

| ID | Teste | O que valida |
|---|---|---|
| DC01 | Template com erro no meio — output antes e depois capturado | Output parcial |
| DC02 | 3 variáveis: 1a OK, 2a erra, 3a OK — output contém 1a e 3a | Render não para |
| DC03 | Múltiplos erros diferentes no mesmo template | Todos acumulados |
| DC04 | AuditError.Error() contém contagem | "N error(s)" |
| DC05 | AuditError.Errors() retorna slice com tipos corretos | SourceError interface |

---

## 7. AuditOptions — Controle Granular

| ID | Teste | O que valida |
|---|---|---|
| O01 | Todos flags false → Expressions=[] | Nenhum trace |
| O02 | Só TraceVariables → só KindVariable | Só variáveis |
| O03 | Só TraceConditions → só KindCondition | Só condições |
| O04 | Só TraceIterations → só KindIteration | Só iterações |
| O05 | Só TraceAssignments → KindAssignment + KindCapture | Assigns e captures |
| O06 | Todos flags true → todas as kinds presentes | Tudo |
| O07 | Diagnostics sempre presentes independente de flags | Erros não dependem de trace |
| O08 | MaxIterationTraceItems=0 com todos flags → sem limite | Unlimited |
| O09 | Flags não afetam o Output | Output idêntico ao Render normal |

---

## 8. AuditResult — Estrutura Geral

### 8.1 Output

| ID | Teste | O que valida |
|---|---|---|
| R01 | Output igual a Render normal (template simples) | Paridade |
| R02 | Output igual a Render normal (template complexo) | Paridade |
| R03 | Output parcial quando há erro | Conteúdo antes do erro |
| R04 | Output vazio para template vazio | Output="" |

### 8.2 Expressions Ordering

| ID | Teste | O que valida |
|---|---|---|
| RO01 | assign antes de variable no array | Ordem de execução |
| RO02 | for → inner expressions repetidas por iteração | Linearizadas |
| RO03 | if(true) → inner expressions presentes; if(false) → ausentes | Só branch executado |
| RO04 | case → só expressions do when ativo | Branch correto |
| RO05 | Nested: if dentro de for → execução correta | Ordem linear |

### 8.3 JSON Serialization

| ID | Teste | O que valida |
|---|---|---|
| RJ01 | AuditResult serializa para JSON sem erro | json.Marshal OK |
| RJ02 | JSON tem as chaves corretas (snake_case: traced_count, etc.) | Tags corretas |
| RJ03 | JSON omitempty funciona (campos nil omitidos) | Não polui |
| RJ04 | JSON roundtrip: Marshal → Unmarshal → igual | Estabilidade |

---

## 9. Validate — Análise Estática

| ID | Teste | O que valida |
|---|---|---|
| VA01 | `{% if true %}{% endif %}` | empty-block info |
| VA02 | `{% unless true %}{% endunless %}` | empty-block info |
| VA03 | `{% for x in items %}{% endfor %}` | empty-block info |
| VA04 | `{% case x %}{% when "a" %}{% endcase %}` | empty-block (se detectable) |
| VA05 | Template normal sem problemas | Diagnostics=[] |
| VA06 | `{{ x \| no_such_filter }}` | undefined-filter error |
| VA07 | `{{ x \| upcase }}` — filtro válido | Sem diagnostic |
| VA08 | Validate retorna Output="" (não renderiza) | Output vazio |
| VA09 | Validate retorna Expressions=[] (sem execução) | Sem expressions |
| VA10 | Múltiplos empty-blocks | Todos detectados |
| VA11 | Empty-block com whitespace (`{% if true %} {% endif %}`) — não é empty se tem texto | Sem false positive |
| VA12 | Nested empty block | Detecta inner block |

---

## 10. RenderOptions — Interação com Audit

| ID | Teste | O que valida |
|---|---|---|
| RO01 | WithStrictVariables → undefined-variable como diagnostic | Warning capturado |
| RO02 | Sem StrictVariables → undefined var sem diagnostic | Nada |
| RO03 | WithLaxFilters → filtro desconhecido sem erro | Sem diagnostic |
| RO04 | WithGlobals(map) → variáveis globals acessíveis | Value correto |
| RO05 | WithContext com cancel → comportamento limpo | Sem panic |
| RO06 | WithSizeLimit → output truncado mas trace completo | Expressions todas |
| RO07 | WithErrorHandler → handler chamado E diagnostic criado | Ambos |

---

## 11. Position & Range — Precisão

| ID | Teste | O que valida |
|---|---|---|
| P01 | Primeira linha, primeira coluna | Line=1, Column=1 |
| P02 | Terceira linha | Line=3 |
| P03 | Coluna > 1 (expression com indent/texto antes) | Column correto |
| P04 | Range.End.Column para expression single-line | End preciso |
| P05 | Múltiplas expressions — cada Range é único | Sem sobreposição |
| P06 | Template com tabs — coluna conta bytes/chars corretamente | Consistente |
| P07 | Template multiline — expression em última linha | Line correto |
| P08 | Expression após tag longa — Column correto | Column offset |

---

## 12. Edge Cases e Cenários Complexos

| ID | Teste | O que valida |
|---|---|---|
| E01 | Template vazio — sem crash, result OK | Output="", Expressions=[], Diagnostics=[] |
| E02 | Template só com texto — sem traces | 0 expressions |
| E03 | Template extremamente longo (1000+ expressions) | Sem crash, performance OK |
| E04 | Nested for 3 deep | Depth increments corretos |
| E05 | If → for → if → variable | Depth=3 para a variável mais interna |
| E06 | For com break/continue | Iterações interrompidas refletidas no trace |
| E07 | `{% comment %}...{% endcomment %}` — sem trace | Ignorado |
| E08 | `{% raw %}{{ not_parsed }}{% endraw %}` — sem trace | Raw ignorado |
| E09 | Include/render tag (se template store configurado) | Expressions do included? |
| E10 | Increment/decrement tags | Sem crash |
| E11 | `{% liquid assign x = 1 \n echo x %}` (liquid tag) | Traces corretos |
| E12 | Template com Unicode | Values preservam Unicode |
| E13 | Template com whitespace control `{%- if -%}` | Output trimmed, traces ainda presentes |
| E14 | Bindings com tipos complexos (structs, interfaces, Drops) | Value resolve corretamente |
| E15 | Cycle tag dentro de for | Sem crash |

---

## 13. Spec Example — End-to-End

| ID | Teste | O que valida |
|---|---|---|
| S01 | Template completo da spec (assign + variable + if + for) | Output exato, Expressions na ordem da spec, todos os campos |
| S02 | Output do S01 serializado em JSON e verificado contra o esperado | JSON match |

---

## Resumo de Contagem

| Categoria | Novos Testes | Existentes |
|---|---|---|
| 1. VariableTrace | ~45 | 4 |
| 2. ConditionTrace | ~35 | 5 |
| 3. IterationTrace | ~30 | 3 |
| 4. AssignmentTrace | ~20 | 3 |
| 5. CaptureTrace | ~10 | 1 |
| 6. Diagnostics | ~20 | 6 |
| 7. AuditOptions | ~9 | 1 |
| 8. AuditResult | ~10 | 2 |
| 9. Validate | ~12 | 3 |
| 10. RenderOptions | ~7 | 1 |
| 11. Position/Range | ~8 | 3 |
| 12. Edge Cases | ~15 | 0 |
| 13. Spec E2E | ~2 | 1 |
| **Total** | **~223** | **33** |

---

## Prioridade de Implementação

1. **VariableTrace completo** — é o tipo mais usado e com mais campos
2. **ConditionTrace completo** — é o mais complexo com branches, groups, comparisons
3. **IterationTrace completo** — tem truncation, inner expressions, limit/offset/reversed
4. **AssignmentTrace completo** — pipeline, path
5. **CaptureTrace completo** — relativamente simples
6. **Diagnostics avançados** — dupla referência, render continua, múltiplos erros
7. **AuditOptions granular** — flag isolation
8. **Validate avançado** — mais patterns estáticos
9. **Position/Range precisão** — colunas exatas
10. **Edge cases** — stress, Unicode, whitespace control, etc.
11. **E2E spec example** — golden test
