# LiquidJS — Mapeamento Completo de Features

> Referência extraída diretamente do código-fonte em `.example-repositories/liquid-js/liquidjs` (src/).
> Organizada por domínio. Usada como base para comparação com a implementação Go.

---

## Tags

### Tags de output / expressão

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `{{ }}` | `{{ expressao }}` | Output de variável ou expressão com filtros |
| `echo` | `{% echo expressao %}` | Equivalente a `{{ }}`; usável dentro de `{% liquid %}`; value é opcional (se vazio não emite nada) |

### Tags de variável / estado

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `assign` | `{% assign var = valor %}` | Cria variável no scope inferior (`ctx.bottom()`); suporta filtros no valor |
| `capture` | `{% capture var %}...{% endcapture %}` | Captura output como string; nome pode ser identifier ou string quoted |
| `increment` | `{% increment var %}` | Stored em `context.environments`; começa em 0, emite valor atual depois incrementa |
| `decrement` | `{% decrement var %}` | Stored em `context.environments`; começa em 0, decrementa antes de emitir (emite -1, -2, …) |

### Tags condicionais

| Tag | Sub-tags | Notas |
|-----|----------|-------|
| `if` | `elsif`, `else`, `endif` | Operadores: `==`, `!=`, `>`, `<`, `>=`, `<=`, `contains`, `not`, `and`, `or`; `else` não pode ser duplicado; `elsif` não pode aparecer após `else` |
| `unless` | `elsif`, `else`, `endunless` | Inverte condição inicial; `elsif` usa `isTruthy` (não invertida); `else` vai para `elseTemplates` |
| `case` | `when`, `else`, `endcase` | `when` suporta múltiplos valores separados por `or` ou `,`; `else` aceita apenas um; branches após `else` são ignorados |

### Tags de iteração

| Tag | Opções | Notas |
|-----|--------|-------|
| `for` | `offset: n\|continue`, `limit: n`, `reversed` | Sub-tag `else` quando coleção vazia; cria objeto `forloop`; suporta `break`/`continue`; `offset: continue` usa `continueKey` register para continuar do ponto anterior; por padrão aplica modifiers na ordem: `offset` → `limit` → `reversed` (ou ordem de declaração se `orderedFilterParameters: true`) |
| `break` | — | Seta `ctx.breakCalled = true`; interrompe `for` no final da iteração atual |
| `continue` | — | Seta `ctx.continueCalled = true`; pula restante da iteração atual em `for` |
| `cycle` | `[group:] v1, v2, [v3, ...]` | Estado em register `'cycle'`; chave: `"cycle:{group}:{candidates}"`; grupo é opcional; sem grupo a key inclui lista de candidatos |
| `tablerow` | `cols: n`, `offset: n`, `limit: n` | Gera HTML de tabela (`<tr class="rowN">`, `<td class="colN">`); cria objeto `tablerowloop`; sem `cols` usa largura = tamanho da coleção |

### Tags de inclusão de templates

| Tag | Sintaxe | Escopo | Notas |
|-----|---------|--------|-------|
| `include` | `{% include 'arquivo' [with var] [key: val ...] %}` | Compartilhado (non-isolated, leak de variáveis) | `with var` adiciona `filepath` como chave no scope; `jekyllInclude: true` injeta tudo no objeto `include` em vez do scope direto |
| `render` | `{% render 'arquivo' [with var [as alias]] [for collection [as alias]] [key: val ...] %}` | Isolado (`ctx.spawn()`) | `with` expõe única variável; `for` itera coleção expondo item + `forloop`; `as alias` renomeia a variável exposta |

### Tags de layout / herança

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `layout` | `{% layout 'arquivo' [key: val ...] %}` | Herança de layout; consome tokens restantes do template como templates filhos; `{% layout none %}` (ou `{% layout %}` sem arquivo) desabilita layout e renderiza direto |
| `block` | `{% block nome %}...{% endblock %}` | Define bloco sobreponível no contexto de layout; suporta `{{ block.super }}` para renderizar conteúdo do pai |

### Tags de texto / estrutura

| Tag | Sintaxe | Notas |
|-----|---------|-------|
| `raw` | `{% raw %}...{% endraw %}` | Output literal, bypassa renderização; tokens internos são emitidos como texto |
| `comment` | `{% comment %}...{% endcomment %}` | Ignorado; consome tokens mas não renderiza |
| `#` (inline comment) | `{%# comentário %}` | Linha única; multi-linha exige `#` em cada linha (`/\n\s*[^#\s]/g` lança erro) |
| `liquid` | `{% liquid tag1\ntag2\n... %}` | Multi-tag sem delimitadores `{% %}`; usa `readLiquidTagTokens` |

---

## Filtros

### String

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `append` | `v \| append: arg` | Concatena `arg` ao final; exige exatamente 2 argumentos |
| `prepend` | `v \| prepend: arg` | Prepende `arg` ao início; exige exatamente 2 argumentos |
| `downcase` | `v \| downcase` | `toLowerCase()` |
| `upcase` | `v \| upcase` | `toUpperCase()` |
| `capitalize` | `v \| capitalize` | Primeira letra maiúscula, resto minúsculas |
| `remove` | `v \| remove: arg` | Remove todas as ocorrências de `arg` |
| `remove_first` | `v \| remove_first: l` | Remove primeira ocorrência |
| `remove_last` | `v \| remove_last: l` | Remove última ocorrência (busca com `lastIndexOf`) |
| `replace` | `v \| replace: pattern, replacement` | Substitui todas as ocorrências |
| `replace_first` | `v \| replace_first: arg1, arg2` | Substitui primeira ocorrência |
| `replace_last` | `v \| replace_last: arg1, arg2` | Substitui última ocorrência (busca com `lastIndexOf`) |
| `split` | `v \| split: arg` | Divide por delimitador; trailing empty strings removidas (comportamento Ruby) |
| `lstrip` | `v \| lstrip[: chars]` | Strip esquerda; sem argumento: whitespace; com argumento: conjunto de chars |
| `rstrip` | `v \| rstrip[: chars]` | Strip direita; sem argumento: whitespace; com argumento: conjunto de chars |
| `strip` | `v \| strip[: chars]` | Strip ambos lados; sem argumento: `trim()`; com argumento: conjunto de chars |
| `strip_newlines` | `v \| strip_newlines` | Remove `\r?\n` |
| `truncate` | `v \| truncate[: l[, o]]` | Trunca para `l` chars (default 50); sufixo `o` (default `'...'`) |
| `truncatewords` | `v \| truncatewords[: words[, o]]` | Trunca para `words` palavras (default 15); sufixo `o` (default `'...'`); `words <= 0` usa 1 |
| `normalize_whitespace` | `v \| normalize_whitespace` | Colapsa whitespace em espaço único |
| `number_of_words` | `input \| number_of_words[: mode]` | Conta palavras; modos: `'cjk'` (CJK + não-CJK), `'auto'` (CJK se presente, else split), default (split por espaço) |
| `array_to_sentence_string` | `array \| array_to_sentence_string[: connector]` | `"a, b, and c"`; `connector` default `'and'`; 0 → `''`, 1 → `a`, 2 → `a and b` |

### Math

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `abs` | `v \| abs` | `Math.abs` |
| `at_least` | `v \| at_least: min` | `Math.max(v, min)` |
| `at_most` | `v \| at_most: max` | `Math.min(v, max)` |
| `ceil` | `v \| ceil` | `Math.ceil` |
| `floor` | `v \| floor` | `Math.floor` |
| `round` | `v \| round[: decimals]` | Arredonda para `decimals` casas (default 0); usa `Math.round(v * 10^d) / 10^d` |
| `divided_by` | `v \| divided_by: divisor[, integerArithmetic]` | Divisão; `integerArithmetic=true` usa `Math.floor` (integer division) |
| `minus` | `v \| minus: arg` | Subtração |
| `plus` | `v \| plus: arg` | Adição |
| `modulo` | `v \| modulo: arg` | Módulo (`%`) |
| `times` | `v \| times: arg` | Multiplicação |

### HTML

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `escape` | `v \| escape` | HTML escape: `&`→`&amp;`, `<`→`&lt;`, `>`→`&gt;`, `"`→`&#34;`, `'`→`&#39;` |
| `xml_escape` | `v \| xml_escape` | Alias para `escape` |
| `escape_once` | `v \| escape_once` | Unescape então escape (idempotente; não re-escapa entidades já escapadas) |
| `newline_to_br` | `v \| newline_to_br` | Substitui `\r?\n` por `<br />\n` |
| `strip_html` | `v \| strip_html` | Remove tags `<script>`, `<style>`, tags HTML e comentários HTML |

### URL

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `url_encode` | `v \| url_encode` | `encodeURIComponent`, espaços viram `+` |
| `url_decode` | `v \| url_decode` | `decodeURIComponent`, `+` vira espaço |
| `cgi_escape` | `v \| cgi_escape` | `encodeURIComponent` com `+` e hex maiúsculo para `!'()*` |
| `uri_escape` | `v \| uri_escape` | `encodeURI` preservando `[` e `]` |
| `slugify` | `v \| slugify[: mode[, cased]]` | Slugifica string; modos: `'raw'`, `'default'`, `'pretty'`, `'ascii'`, `'latin'`, `'none'`; `cased=false` por padrão (lowercase) |

### Date

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `date` | `v \| date[: format[, timezoneOffset]]` | Formatação strftime; `'now'`/`'today'` → hora atual; string numérica/número → epoch seconds; respeita `opts.preserveTimezones`; default format via `opts.dateFormat` |
| `date_to_xmlschema` | `v \| date_to_xmlschema` | `date` com formato `'%Y-%m-%dT%H:%M:%S%:z'` |
| `date_to_rfc822` | `v \| date_to_rfc822` | `date` com formato `'%a, %d %b %Y %H:%M:%S %z'` |
| `date_to_string` | `v \| date_to_string[: type[, style]]` | Mês abreviado (`%b`); `type='ordinal'`, `style='US'` usa formato americano |
| `date_to_long_string` | `v \| date_to_long_string[: type[, style]]` | Mês completo (`%B`); mesmas opções de `date_to_string` |

### Array

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `join` | `arr \| join[: sep]` | Junta com separador (default: `' '`); nil → `' '` |
| `first` | `arr \| first` | Primeiro elemento; string/array retorna `''` se não array-like |
| `last` | `arr \| last` | Último elemento; string/array retorna `''` se não array-like |
| `reverse` | `arr \| reverse` | Cria cópia invertida (`[...arr].reverse()`) |
| `sort` | `arr \| sort[: property]` | Ordena; `property` pode ser path com `.`; `<` para comparar |
| `sort_natural` | `arr \| sort_natural[: property]` | Ordenação case-insensitive |
| `size` | `v \| size` | Comprimento de string ou array; `0` se nil |
| `map` | `arr \| map: property` | Extrai propriedade de cada item |
| `sum` | `arr \| sum[: property]` | Soma valores (NaN → 0); com `property` soma a propriedade de cada item |
| `compact` | `arr \| compact` | Remove valores nil (usando `toValue`) |
| `concat` | `arr \| concat: arr2` | Concatena dois arrays; `arr2` default `[]` |
| `push` | `arr \| push: item` | Adiciona item ao final (retorna novo array) |
| `unshift` | `arr \| unshift: item` | Adiciona item ao início (retorna novo array) |
| `pop` | `arr \| pop` | Remove último elemento (retorna novo array) |
| `shift` | `arr \| shift` | Remove primeiro elemento (retorna novo array) |
| `slice` | `arr \| slice: begin[, length]` | Fatia string ou array; `begin < 0` → desde o fim; `length` default `1` |
| `where` | `arr \| where: property[, expected]` | Filtra itens onde `property == expected`; sem `expected` → filtra truthy; com `jekyllWhere: true` também aceita array membership |
| `reject` | `arr \| reject: property[, expected]` | Inverso de `where` |
| `where_exp` | `arr \| where_exp: itemName, exp` | Filtra por expressão Liquid; `exp` avaliada com `itemName` no scope |
| `reject_exp` | `arr \| reject_exp: itemName, exp` | Inverso de `where_exp` |
| `group_by` | `arr \| group_by: property` | Agrupa em `[{name, items}, ...]` |
| `group_by_exp` | `arr \| group_by_exp: itemName, exp` | Agrupa por expressão Liquid |
| `has` | `arr \| has: property[, expected]` | Retorna boolean; `true` se algum item combina |
| `has_exp` | `arr \| has_exp: itemName, exp` | Retorna boolean via expressão |
| `find` | `arr \| find: property[, expected]` | Retorna primeiro item que combina (ou `undefined`) |
| `find_exp` | `arr \| find_exp: itemName, exp` | Retorna primeiro item via expressão |
| `find_index` | `arr \| find_index: property[, expected]` | Retorna índice do primeiro item que combina (ou `undefined`) |
| `find_index_exp` | `arr \| find_index_exp: itemName, exp` | Retorna índice via expressão |
| `uniq` | `arr \| uniq` | Remove duplicatas (`new Set`) |
| `sample` | `v \| sample[: count]` | Amostra aleatória; `count=1` retorna item único; `count>1` retorna array; aceita string ou array |

### Base64

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `base64_encode` | `v \| base64_encode` | Encode Base64; stringifica input antes |
| `base64_decode` | `v \| base64_decode` | Decode Base64; stringifica input antes |

### Misc

| Filtro | Assinatura | Notas |
|--------|-----------|-------|
| `default` | `v \| default: defaultValue[, allow_false: true]` | Retorna `defaultValue` se `v` for falsy ou empty (string/array vazia, objeto sem chaves); `allow_false: true` deixa `false` passar |
| `json` | `v \| json[: space]` | `JSON.stringify(v, null, space)`; `space` default `0` |
| `jsonify` | `v \| jsonify[: space]` | Alias para `json` |
| `inspect` | `v \| inspect[: space]` | `JSON.stringify` com proteção a referências circulares (`'[Circular]'`) |
| `to_integer` | `v \| to_integer` | `Number(v)` |
| `raw` | `v \| raw` | Passa valor sem evaluation (`{ raw: true, handler: identify }`) |

---

## Drops (Objetos Especiais)

### Drop (classe base abstrata)

| Método | Notas |
|--------|-------|
| `liquidMethodMissing(key, context)` | Chamado quando propriedade não encontrada; retorna `undefined` por padrão; pode ser sobrescrito |

### ForloopDrop (criado em `for`)

| Propriedade/Método | Tipo | Notas |
|--------------------|------|-------|
| `length` | number | Total de iterações |
| `name` | string | `"variavel-colecao"` |
| `index` | number | 1-based índice atual |
| `index0` | number | 0-based índice atual |
| `rindex` | number | Iterações restantes incluindo atual |
| `rindex0` | number | Iterações restantes excluindo atual |
| `first` | boolean | `i === 0` |
| `last` | boolean | `i === length - 1` |

### TablerowloopDrop (criado em `tablerow`, extends ForloopDrop)

Herda todas as propriedades de `ForloopDrop`, adiciona:

| Propriedade/Método | Tipo | Notas |
|--------------------|------|-------|
| `row` | number | Linha atual (1-based): `Math.floor(i / cols) + 1` |
| `col` | number | Coluna atual (1-based): `col0 + 1` |
| `col0` | number | Coluna atual (0-based): `i % cols` |
| `col_first` | boolean | `col0 === 0` |
| `col_last` | boolean | `col === cols` |

### EmptyDrop (`empty` literal)

| Comportamento | Notas |
|---------------|-------|
| `== empty` | outro EmptyDrop → `false`; string/array → `.length === 0`; objeto → `Object.keys().length === 0` |
| `valueOf()` | `''` |
| Comparações `>`, `<`, `>=`, `<=` | sempre `false` |

### BlankDrop (`blank` literal, extends EmptyDrop)

| Comportamento | Notas |
|---------------|-------|
| `== blank` | `false` → `true`; nil → `true`; string só whitespace → `true`; herda lógica de EmptyDrop para outros |

### NullDrop (`null`/`nil` literal)

| Comportamento | Notas |
|---------------|-------|
| `== null` | equals qualquer nil value |
| `valueOf()` | `null` |
| Comparações | sempre `false` |

### BlockDrop (disponível como `block` dentro de `{% block %}`)

| Propriedade | Notas |
|-------------|-------|
| `block.super` | Renderiza o conteúdo do bloco pai (do layout) |

---

## Operadores

| Operador | Tipo | Comportamento |
|----------|------|---------------|
| `==` | binário | Igualdade; respeita interface `Comparable`; arrays comparam elemento a elemento |
| `!=` | binário | Desigualdade (`!equals`) |
| `>` | binário | Maior que; respeita `Comparable.gt/lt` |
| `<` | binário | Menor que; respeita `Comparable.lt/gt` |
| `>=` | binário | Maior ou igual; respeita `Comparable.geq/leq` |
| `<=` | binário | Menor ou igual; respeita `Comparable.leq/geq` |
| `contains` | binário | String: `indexOf > -1`; array: `some(equals)`; outros com `indexOf`: `indexOf > -1` |
| `not` | unário | `isFalsy(v, ctx)` |
| `and` | binário | `isTruthy(l) && isTruthy(r)` |
| `or` | binário | `isTruthy(l) \|\| isTruthy(r)` |

Operadores são customizáveis via opção `operators`.

---

## Truthiness

| Modo | Falsy | Truthy |
|------|-------|--------|
| default (Liquid) | `false`, `undefined`, `null` | tudo o resto (inclusive `0`, `""`, `[]`) |
| `jsTruthy: true` | qualquer falsy JavaScript (`0`, `""`, `null`, `undefined`, `false`, `NaN`) | tudo o resto |

---

## Controle de Whitespace

| Delimitador | Comportamento |
|-------------|---------------|
| `{%-` / `-%}` | Trim whitespace antes/depois de tags |
| `{{-` / `-}}` | Trim whitespace antes/depois de outputs |
| `trimTagRight: true` | Global: trim direita de `{% %}` até `\n` (inclusive) |
| `trimTagLeft: true` | Global: trim esquerda de `{% %}` |
| `trimOutputRight: true` | Global: trim direita de `{{ }}` até `\n` (inclusive) |
| `trimOutputLeft: true` | Global: trim esquerda de `{{ }}` |
| `greedy: true` | (default) trim consome todos os espaços/`\n` consecutivos |

---

## Herança de Layout

| Mecanismo | Notas |
|-----------|-------|
| `{% layout 'arquivo' %}` | Template filho declara layout pai; tokens do filho são processados em `blockMode: STORE` |
| `{% layout none %}` | Desativa layout; renderiza `blockMode: OUTPUT` direto |
| `{% block nome %}...{% endblock %}` | Define bloco sobreponível; em modo `STORE` guarda render function |
| `{{ block.super }}` | Acessa conteúdo do bloco pai via `BlockDrop.super()` |
| Bloco anônimo (`blocks['']`) | Conteúdo do filho fora de `{% block %}` vira bloco anônimo; disponível no pai como `{{ content_for_layout }}` equivalente |

---

## Opções (LiquidOptions)

### Filesystem / templates

| Opção | Default | Notas |
|-------|---------|-------|
| `root` | `['.']` | Diretório(s) base para templates |
| `partials` | `root` | Diretório(s) para partials (`include`, `render`) |
| `layouts` | `root` | Diretório(s) para layouts |
| `relativeReference` | `true` | Permite referências relativas (path precisa estar dentro do root) |
| `extname` | `''` | Extensão adicionada ao lookup se filepath não tiver extensão |
| `cache` | `false` | Cache de templates; `true`, número (LRU size), ou objeto `LiquidCache` |
| `fs` | node fs | Implementação customizada do filesystem |
| `templates` | — | Mapa de templates em memória; ignora fs e root quando definido |
| `dynamicPartials` | `true` | Trata nome de arquivo como expressão Liquid; `false` trata como literal |

### Comportamento de escopo / variáveis

| Opção | Default | Notas |
|-------|---------|-------|
| `globals` | `{}` | Scope global passado para todos os templates (inclusive partials/layouts) |
| `strictVariables` | `false` | Lança erro em variável undefined |
| `strictFilters` | `false` | Lança erro em filtro undefined; se `false`, filtro undefined é pulado |
| `ownPropertyOnly` | `true` | Ignora propriedades herdadas do prototype |
| `lenientIf` | `false` | Com `strictVariables: true`, permite variável undefined em `if`/`elsif`/`unless`/`default` sem erro |
| `catchAllErrors` | `false` | Coleta todos os erros em vez de parar no primeiro |
| `jsTruthy` | `false` | Usa truthiness JavaScript em vez de Liquid |
| `jekyllInclude` | `false` | `include` injeta variáveis em objeto `include` do scope |
| `jekyllWhere` | `false` | `where` também faz array membership check |

### Output

| Opção | Default | Notas |
|-------|---------|-------|
| `outputEscape` | `undefined` | Escape padrão aplicado a outputs: `'escape'`, `'json'`, ou função |
| `keepOutputType` | `false` | Preserva tipo do output (não converte para string) |

### Delimitadores

| Opção | Default | Notas |
|-------|---------|-------|
| `tagDelimiterLeft` | `'{%'` | Delimitador esquerdo de tags |
| `tagDelimiterRight` | `'%}'` | Delimitador direito de tags |
| `outputDelimiterLeft` | `'{{'` | Delimitador esquerdo de outputs |
| `outputDelimiterRight` | `'}}'` | Delimitador direito de outputs |
| `keyValueSeparator` | `':'` | Separador key/value em hash arguments |

### Whitespace

| Opção | Default | Notas |
|-------|---------|-------|
| `trimTagRight` | `false` | Trim direita de `{% %}` até `\n` (inclusive) |
| `trimTagLeft` | `false` | Trim esquerda de `{% %}` |
| `trimOutputRight` | `false` | Trim direita de `{{ }}` até `\n` (inclusive) |
| `trimOutputLeft` | `false` | Trim esquerda de `{{ }}` |
| `greedy` | `true` | Trim consome todos os espaços/`\n` consecutivos |

### Date

| Opção | Default | Notas |
|-------|---------|-------|
| `timezoneOffset` | local | Nome de timezone JS ou offset numérico para filtro `date` |
| `dateFormat` | `'%A, %B %-e, %Y at %-l:%M %P %z'` | Formato padrão quando `date` não recebe formato |
| `locale` | sistema | Locale padrão usado pelo filtro `date` |
| `preserveTimezones` | `false` | Preserva timezone da string de input no filtro `date` |

### Limites (DoS)

| Opção | Default | Notas |
|-------|---------|-------|
| `parseLimit` | `Infinity` | Máximo de chars processados por chamada `parse()` |
| `renderLimit` | `Infinity` | Máximo de ms por chamada `render()` |
| `memoryLimit` | `Infinity` | Máximo de alocações (arrays, strings, etc.) |

### Misc

| Opção | Default | Notas |
|-------|---------|-------|
| `orderedFilterParameters` | `false` | Respeita ordem de declaração dos modifiers `offset/limit/reversed` no `for` |
| `operators` | defaultOperators | Objeto de operadores para condicionais customizáveis |

---

## RenderOptions (overrides por chamada)

| Opção | Notas |
|-------|-------|
| `globals` | Globals apenas para este render |
| `strictVariables` | Override de `strictVariables` |
| `ownPropertyOnly` | Override de `ownPropertyOnly` |
| `templateLimit` | Máximo de templates renderizados por chamada |
| `renderLimit` | Máximo de ms por chamada |
| `memoryLimit` | Máximo de alocações por chamada |

---

## API da Classe Liquid

### Parse

| Método | Retorno | Notas |
|--------|---------|-------|
| `parse(html, filepath?)` | `Template[]` | Síncrono |
| `parseFile(file, lookupType?)` | `Promise<Template[]>` | Assíncrono |
| `parseFileSync(file, lookupType?)` | `Template[]` | Síncrono |

### Render

| Método | Retorno | Notas |
|--------|---------|-------|
| `render(tpl, scope?, opts?)` | `Promise<any>` | |
| `renderSync(tpl, scope?, opts?)` | `any` | |
| `renderToNodeStream(tpl, scope?, opts?)` | `NodeJS.ReadableStream` | Streaming |
| `parseAndRender(html, scope?, opts?)` | `Promise<any>` | Parse + render num passo |
| `parseAndRenderSync(html, scope?, opts?)` | `any` | |
| `renderFile(file, ctx?, opts?)` | `Promise<any>` | |
| `renderFileSync(file, ctx?, opts?)` | `any` | |
| `renderFileToNodeStream(file, scope?, opts?)` | `Promise<NodeJS.ReadableStream>` | |

### Eval

| Método | Retorno | Notas |
|--------|---------|-------|
| `evalValue(str, scope?)` | `Promise<any>` | Avalia expressão Liquid |
| `evalValueSync(str, scope?)` | `any` | |

### Análise Estática

| Método | Retorno | Notas |
|--------|---------|-------|
| `analyze(template, options?)` | `Promise<StaticAnalysis>` | Analisa variáveis, partials, etc. |
| `analyzeSync(template, options?)` | `StaticAnalysis` | |
| `parseAndAnalyze(html, filename?, options?)` | `Promise<StaticAnalysis>` | |
| `parseAndAnalyzeSync(html, filename?, options?)` | `StaticAnalysis` | |
| `variables(template, options?)` | `Promise<string[]>` | Lista variáveis (sem propriedades) |
| `variablesSync(template, options?)` | `string[]` | |

### Extensibilidade

| Método | Notas |
|--------|-------|
| `registerFilter(name, filter)` | Registra filtro customizado |
| `registerTag(name, tag)` | Registra tag customizada (classe ou objeto de implementação) |
| `plugin(fn)` | Registra plugin (`fn` recebe `this=liquidInstance, arg=LiquidClass`) |
| `express()` | Retorna adapter para Express.js view engine |
