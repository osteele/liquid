# Ruby Liquid — Mapeamento Completo de Features

> Referência extraída diretamente do código-fonte em `.example-repositories/liquid-ruby/liquid` (lib/ + test/).
> Organizada por domínio. Usada como base para comparação com a implementação Go.

---

## Tags

### Tags de output / expressão

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `{{ }}` | `{{ expressao }}` | Output de variável ou expressão com filtros |
| `echo` | `{% echo expressao %}` | Equivalente a `{{ }}`, usável dentro de `{% liquid %}` |

### Tags de variável / estado

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `assign` | `{% assign var = valor %}` | Cria variável; rastreado por resource limits |
| `capture` | `{% capture var %}...{% endcapture %}` | Captura output como string; rastreado por resource limits |
| `increment` | `{% increment var %}` | Starts at 0, outputs then increments; compartilha estado com `decrement` |
| `decrement` | `{% decrement var %}` | Starts at -1, outputs then decrements; compartilha estado com `increment` |

### Tags condicionais

| Tag | Sub-tags | Notas |
|-----|----------|-------|
| `if` | `elsif`, `else` | Operadores: `==`, `!=`, `<>`, `<`, `>`, `<=`, `>=`, `contains`, `and`, `or`; valores especiais: `blank`, `empty` |
| `unless` | `elsif`, `else` | Inverte condição inicial; resto igual ao `if` |
| `case` | `when`, `else` | `when` suporta múltiplos valores separados por `or` ou `,` |
| `ifchanged` | — | Renderiza só se output mudou desde última iteração; estado em `registers[:ifchanged]` |

### Tags de iteração

| Tag | Opções | Notas |
|-----|--------|-------|
| `for` | `reversed`, `limit: n`, `offset: n`, range `(a..b)` | Sub-tag `else` (quando array vazio); cria objeto `forloop`; suporta `break`/`continue` |
| `break` | — | Interrompe `for`; usa `BreakInterrupt` |
| `continue` | — | Pula iteração em `for`; usa `ContinueInterrupt` |
| `cycle` | nome opcional: `{% cycle "name": v1, v2 %}` | Estado em `registers[:cycle]`; precisa estar dentro de `for` |
| `tablerow` | `cols: n`, `limit: n`, `offset: n`, range `(a..b)` | Gera HTML de tabela (`<tr class="rowN">`, `<td class="colN">`); cria objeto `tablerowloop`; suporta `break` |

### Tags de inclusão de templates

| Tag | Sintaxe | Escopo | Status |
|-----|---------|--------|--------|
| `include` | `{% include 'arquivo' [with var] [for array] [as alias] [key: val] %}` | Compartilhado (leak de variáveis) | **Deprecated** |
| `render` | `{% render 'arquivo' [with var] [for array] [as alias] [key: val] %}` | Isolado (sem acesso a vars do pai, exceto globals) | Atual |

### Tags de texto / estrutura

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `raw` | `{% raw %}...{% endraw %}` | Output literal, bypassa renderização |
| `comment` | `{% comment %}...{% endcomment %}` | Ignorado; suporta nesting de `comment`/`raw` |
| `#` (inline comment) | `{%# comentário %}` | Linha única; cada linha precisa de `#` |
| `liquid` | `{% liquid tag1\ntag2 %}` | Multi-linha sem delimitadores `{% %}`; cada linha é uma tag |
| `doc` | `{% doc %}...{% enddoc %}` | Documentação LiquidDoc; ignorada pelo renderer |

---

## Filtros (StandardFilters)

### String

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `downcase` | `string \| downcase` | |
| `upcase` | `string \| upcase` | |
| `capitalize` | `string \| capitalize` | Capitaliza primeira letra, minúsculas no resto |
| `escape` | `string \| escape` | HTML escape (`<`, `>`, `&`, `"`, `'`); alias `h` |
| `escape_once` | `string \| escape_once` | HTML escape sem re-escapar já escapados |
| `url_encode` | `string \| url_encode` | CGI escape; espaços viram `+` |
| `url_decode` | `string \| url_decode` | CGI unescape |
| `base64_encode` | `string \| base64_encode` | Base64 estrito |
| `base64_decode` | `string \| base64_decode` | Levanta erro se inválido |
| `base64_url_safe_encode` | `string \| base64_url_safe_encode` | URL-safe Base64 |
| `base64_url_safe_decode` | `string \| base64_url_safe_decode` | URL-safe Base64 decode |
| `slice` | `string \| slice: offset[, length]` | Também funciona em arrays |
| `truncate` | `string \| truncate: n[, ellipsis]` | Default ellipsis `"..."`, incluso na contagem |
| `truncatewords` | `string \| truncatewords: n[, ellipsis]` | Default 15 palavras |
| `split` | `string \| split: separador` | Retorna array |
| `strip` | `string \| strip` | Remove whitespace dos dois lados |
| `lstrip` | `string \| lstrip` | Remove whitespace da esquerda |
| `rstrip` | `string \| rstrip` | Remove whitespace da direita |
| `strip_html` | `string \| strip_html` | Remove tags HTML + `<script>`, `<style>`, comentários |
| `strip_newlines` | `string \| strip_newlines` | Remove `\n` e `\r\n` |
| `squish` | `string \| squish` | Strip + colapsa whitespace interno em um espaço |
| `newline_to_br` | `string \| newline_to_br` | Converte `\n` em `<br />` |
| `replace` | `string \| replace: old, new` | Substitui todas as ocorrências |
| `replace_first` | `string \| replace_first: old, new` | Substitui só a primeira |
| `replace_last` | `string \| replace_last: old, new` | Substitui só a última |
| `remove` | `string \| remove: sub` | Remove todas as ocorrências |
| `remove_first` | `string \| remove_first: sub` | Remove só a primeira |
| `remove_last` | `string \| remove_last: sub` | Remove só a última |
| `append` | `string \| append: suffix` | Concatena no final |
| `prepend` | `string \| prepend: prefix` | Concatena no início |

### Array

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `size` | `array \| size` | Também funciona em strings |
| `first` | `array \| first` | |
| `last` | `array \| last` | |
| `join` | `array \| join[: glue]` | Default glue `" "` |
| `split` | `string \| split: sep` | Inverso de `join` |
| `reverse` | `array \| reverse` | |
| `sort` | `array \| sort[: property]` | Case-sensitive; nil-safe (nils vão pro final) |
| `sort_natural` | `array \| sort_natural[: property]` | Case-insensitive |
| `uniq` | `array \| uniq[: property]` | Remove duplicatas |
| `compact` | `array \| compact[: property]` | Remove nils |
| `map` | `array \| map: property` | Extrai propriedade de cada item |
| `concat` | `array \| concat: outro_array` | Combina arrays (não dedup) |
| `where` | `array \| where: prop[, valor]` | Filtra por propriedade; sem valor = truthy |
| `reject` | `array \| reject: prop[, valor]` | Inverso de `where` |
| `find` | `array \| find: prop[, valor]` | Primeiro match |
| `find_index` | `array \| find_index: prop[, valor]` | Índice do primeiro match |
| `has` | `array \| has: prop[, valor]` | `true` se algum item satisfaz |
| `sum` | `array \| sum[: property]` | Soma numérica |
| `slice` | `array \| slice: offset[, length]` | Fatia de array |

### Matemática

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `abs` | `number \| abs` | |
| `plus` | `number \| plus: n` | |
| `minus` | `number \| minus: n` | |
| `times` | `number \| times: n` | |
| `divided_by` | `number \| divided_by: n` | Tipo do resultado = tipo do divisor; levanta `ZeroDivisionError` |
| `modulo` | `number \| modulo: n` | Levanta `ZeroDivisionError` |
| `round` | `number \| round[: casas]` | Default 0 casas |
| `ceil` | `number \| ceil` | |
| `floor` | `number \| floor` | |
| `at_least` | `number \| at_least: n` | `max(input, n)` |
| `at_most` | `number \| at_most: n` | `min(input, n)` |

### Data

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `date` | `date \| date: format` | Formato strftime; retorna input se vazio/inválido |

### Misc / Default

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `default` | `var \| default: val[, allow_false: bool]` | **Keyword argument** `allow_false`; retorna default se nil, false, ou empty |

---

## Sistema de Filtros

| Feature | Descrição |
|---------|-----------|
| Filtros posicionais | `{{ val \| filter: arg1, arg2 }}` |
| **Filtros com keyword args** | `{{ val \| default: fallback, allow_false: true }}` — passados como hash ao método |
| Positional + keyword misturados | Suportado no modo `strict2_parse` |
| `register_filter(module)` | Registra módulo Ruby como fonte de filtros |
| `strict_filters` | Se `true`, levanta `UndefinedFilter` para filtros desconhecidos |
| `global_filter` | Proc aplicado ao output de toda expressão antes de renderizar |

---

## Expressões e Operadores

### Literais

| Literal | Exemplo |
|---------|---------|
| nil/null | `nil`, `null` |
| boolean | `true`, `false` |
| inteiro | `42`, `-1` |
| float | `3.14` |
| string | `"texto"` ou `'texto'` |
| range | `(1..10)` |
| blank | `blank` → `''` (compara como string vazia) |
| empty | `empty` → `''` (compara como string vazia) |

### Operadores de comparação

| Operador | Comportamento |
|----------|--------------|
| `==` | Igualdade (nil-safe) |
| `!=`, `<>` | Desigualdade |
| `<`, `>`, `<=`, `>=` | Comparação numérica/string |
| `contains` | String: `include?`; Array: `include?` |

### Operadores booleanos

| Operador | Comportamento |
|----------|--------------|
| `and` | Curto-circuito, avalia da esquerda para a direita sem precedência |
| `or` | Curto-circuito |

### Truthiness

| Valor | Truthy? |
|-------|---------|
| `false` | falsy |
| `nil` | falsy |
| `0` | **truthy** |
| `""` | **truthy** |
| `[]` | **truthy** |
| qualquer outro | truthy |

### Acesso a variáveis

| Sintaxe | Descrição |
|---------|-----------|
| `variavel` | Lookup em escopos empilhados |
| `obj.prop` | Acesso a propriedade |
| `obj[key]` | Acesso por chave string |
| `array[0]` | Acesso por índice inteiro |
| `array.first`, `array.last` | Atalhos |
| `array.size`, `hash.size` | Tamanho |
| `forloop.index`, etc. | Propriedades de loops |

---

## Drops (protocolo de objetos customizados)

| Feature | Descrição |
|---------|-----------|
| `Drop` base class | Classe base; métodos públicos são acessíveis por nome |
| `invoke_drop(key)` / `[key]` | Invoca método ou chama `liquid_method_missing` |
| `liquid_method_missing(name)` | Catch-all; levanta `UndefinedDropMethod` se `strict_variables` |
| `key?(_name)` | Sempre retorna `true` por padrão |
| `invokable_methods` | Whitelist de métodos públicos menos blacklist |
| Blacklist de sistema | Métodos de `Drop` + `Enumerable` (exceto `sort`, `count`, `first`, `min`, `max`) |
| `to_liquid` | Retorna `self`; usado para lazy conversion |
| `context=` | Injeta contexto de render no drop |

### ForloopDrop (objeto `forloop`)

| Campo | Descrição |
|-------|-----------|
| `index` | 1-based |
| `index0` | 0-based |
| `rindex` | 1-based reverso |
| `rindex0` | 0-based reverso |
| `first` | boolean |
| `last` | boolean |
| `length` | total de iterações |
| `parentloop` | drop do loop pai (ou nil) |
| `name` | identificador do loop |

### TablerowloopDrop (objeto `tablerowloop`)

| Campo | Descrição |
|-------|-----------|
| `index`, `index0` | iteração 1-based e 0-based |
| `rindex`, `rindex0` | reverso |
| `first`, `last` | boolean |
| `col` | coluna 1-based |
| `col0` | coluna 0-based |
| `col_first`, `col_last` | boolean |
| `row` | linha 1-based |
| `length` | total de iterações |

---

## Erros

| Tipo | Uso |
|------|-----|
| `Liquid::Error` | Base |
| `SyntaxError` | Erro de parse |
| `ArgumentError` | Argumento inválido para filtro/tag |
| `ContextError` | Erro em operação de contexto |
| `FileSystemError` | Erro ao carregar arquivo |
| `StandardError` | Erro genérico de runtime |
| `StackLevelError` | Stack overflow de includes |
| `MemoryError` | Resource limits atingidos |
| `ZeroDivisionError` | Divisão por zero |
| `FloatDomainError` | Operação float inválida |
| `UndefinedVariable` | Variável não encontrada (strict mode) |
| `UndefinedDropMethod` | Método de drop não encontrado (strict mode) |
| `UndefinedFilter` | Filtro não registrado (strict_filters) |
| `MethodOverrideError` | Tentativa de sobrescrever método proibido |
| `DisabledError` | Tag desabilitada (Disableable) |
| `InternalError` | Erro interno do engine |
| `TemplateEncodingError` | Encoding inválido no template |

**Atributos comuns:** `line_number`, `template_name`, `markup_context`

---

## Modos de Erro (error_mode)

| Modo | Comportamento |
|------|--------------|
| `:lax` | Ignora silenciosamente erros de sintaxe na maioria dos casos (Liquid 2.5 compat) |
| `:warn` | **Default**; warnings para sintaxe inválida |
| `:strict` | Levanta erro para a maioria das tags com sintaxe incorreta |
| `:strict2` | Levanta erro para todas as tags com sintaxe incorreta (modo mais rigoroso) |

---

## Resource Limits

| Limite | Descrição |
|--------|-----------|
| `render_length_limit` | Bytes máximos no output de um template |
| `render_score_limit` | Score de render por template (cada nó conta) |
| `assign_score_limit` | Score de assign por template (baseado em bytesize) |
| `cumulative_render_score_limit` | Score acumulado de render (múltiplos renders) |
| `cumulative_assign_score_limit` | Score acumulado de assign |
| `MemoryError` | Levantado quando qualquer limite é excedido |

**Scoring do assign:** String → bytesize; Array/Hash → soma recursiva dos elementos; outros → 1.

---

## Environment / Configuração

| Feature | Descrição |
|---------|-----------|
| `error_mode` | `:lax`, `:warn`, `:strict`, `:strict2` |
| `file_system` | Implementação plugável de filesystem para `include`/`render` |
| `exception_renderer` | Proc para interceptar exceções |
| `default_resource_limits` | Hash com limites padrão (ver acima) |
| `register_tag(name, klass)` | Registra tag customizada |
| `register_filter(module)` | Registra módulo de filtros |
| `Environment.build {}` | Builder imutável (freeze após construção) |
| `Environment.dangerously_override` | Override temporário do environment default (bloco) |

### Opções de render (por chamada)

| Opção | Descrição |
|-------|-----------|
| `filters:` | Módulos de filtro adicionais para o render |
| `registers:` | Hash de registros estáticos |
| `global_filter:` | Proc aplicada a todo output de variável |
| `exception_renderer:` | Override por render |
| `strict_variables:` | Levanta `UndefinedVariable` para vars inexistentes |
| `strict_filters:` | Levanta `UndefinedFilter` para filtros desconhecidos |

---

## Template API

| Método | Descrição |
|--------|-----------|
| `Template.parse(source, options)` | Parse em classe (usa `Environment.default`) |
| `template.render(*args)` | Render com variáveis / contexto / drops |
| `template.render!(*args)` | Render com rethrow de erros |
| `template.errors` | Array de erros acumulados |
| `template.warnings` | Warnings do parse |
| `template.resource_limits` | Objeto `ResourceLimits` |
| `template.root` | Nó raiz da parse tree |
| `template.name` | Nome do template (para mensagens de erro) |

---

## File System

| Interface | Descrição |
|-----------|-----------|
| `BlankFileSystem` | Default; levanta erro em qualquer `include`/`render` |
| `LocalFileSystem.new(root, pattern)` | Lê do disco; pattern default `_%s.liquid` |
| `LocalFileSystem#read_template_file(path)` | Lê arquivo; valida path (sem path traversal) |
| Interface customizada | `read_template_file(path)` — qualquer objeto que responda a esse método |

---

## Análise Estática (ParseTreeVisitor)

| Feature | Descrição |
|---------|-----------|
| `ParseTreeVisitor.for(node, callbacks)` | Cria visitor para nó |
| `visitor.add_callback_for(*classes) { \|node, ctx\| }` | Registra callback por classe |
| `visitor.visit(context)` | Percorre árvore recursivamente |
| `node.nodelist` | Lista de nós filhos (interface padrão) |
| Nodes podem ter `ParseTreeVisitor` próprio | Via `Node::ParseTreeVisitor` inner class |

---

## Profiler

| Feature | Descrição |
|---------|-----------|
| `Template.parse(source, profile: true)` | Habilita profiling |
| `template.profiler` | Objeto `Liquid::Profiler` após render |
| Profiler reporta | Tempo por tag/nó durante a renderização |

---

## Context interno

| Feature | Descrição |
|---------|-----------|
| `scopes` | Stack de hashes de variáveis |
| `environments` | Variáveis globais (do caller) |
| `registers` | Storage de dois níveis: `static` + `changes` |
| `stack(scope) { }` | Push/pop temporário de escopo |
| `new_isolated_subcontext()` | Sub-contexto isolado (para `render` tag) |
| `add_filters(module)` | Adiciona filtros ao contexto |
| `invoke(filter_name, input, *args)` | Invoca filtro |
| `find_variable(name)` | Lookup com strict_variables |
| `tag_disabled?(name)` | Suporte a Disableable tags |
| `with_disabled_tags([names]) { }` | Desabilita tags temporariamente (Disabler) |

---

## Mecanismo de Disablement de Tags

| Feature | Descrição |
|---------|-----------|
| `Tag::Disableable` | Mixin que faz a tag verificar se está desabilitada antes de renderizar |
| `Tag::Disabler` | Mixin que desabilita tags listadas em `disabled_tags` durante seu render |
| Usado em | `render` desabilita `include` dentro de partials renderizados com `render` |

---

## Whitespace Control

| Marcador | Efeito |
|----------|--------|
| `{%-` | Remove whitespace antes da tag |
| `-%}` | Remove whitespace depois da tag |
| `{{-` | Remove whitespace antes do output |
| `-}}` | Remove whitespace depois do output |

---

## Itens ausentes desta implementação Go (vs. Ruby)

> Resumo rápido para cruzamento. Ver `macro-checklist.md` para rastreamento detalhado.

| Item | Grupo na macro-checklist |
|------|--------------------------|
| `echo` tag | T1 |
| `liquid` tag (multi-linha) | T1 |
| `#` inline comment | T1 |
| `increment` / `decrement` | T1 |
| `render` tag (escopo isolado) | T2 |
| Sub-contexto isolado | T2 |
| `doc` tag | — (não listado) |
| `ifchanged` tag | — (não listado) |
| Filter keyword arguments (`allow_false: true`) | — (não listado, citado no README) |
| `squish` filtro | — (não listado) |
| `base64_url_safe_encode` / `base64_url_safe_decode` | — (não listado) |
| `url_decode` filtro | — (não listado) |
| `strict_variables` opção de render | — (parcialmente via engine) |
| `strict_filters` opção de render | — (não listado) |
| `global_filter` opção de render | — (não listado) |
| Error modes `:lax`, `:warn`, `:strict`, `:strict2` | — (citado no README) |
| Resource limits (render_score, assign_score, etc.) | C4 (parcial) |
| Profiler | — (não listado) |
| `ParseTreeVisitor` API pública | A1 (parcial) |
| `ForloopDrop` como tipo público | D1 |
| `TablerowloopDrop` como tipo público | D1 |
| `liquid_method_missing` em drops | D1 |
| `Disableable`/`Disabler` para tags | — (não listado) |
| `Template.render` com `registers:` | — (não listado) |
| `cumulative_*_limit` em resource limits | — (não listado) |
| `TemplateEncodingError` | — (não listado) |
| `ForloopDrop.parentloop` | — (não listado) |
| `tablerow` com `cols` correto + HTML gerado | — (verificar se impl está ok) |
| `cycle` com nome explícito | — (verificar se impl está ok) |
| `case` com `or` em `when` | — (verificar se impl está ok) |
