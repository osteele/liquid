# Test Plan: RenderAudit — Complete Coverage

## Status: 34 tests exist today. This plan maps ~200+ new tests needed.

---

## Inventory of Existing Tests (34)

| # | Test | Covers |
|---|---|---|
| 1 | TraceVariables_simple | Name, Value, Kind basic |
| 2 | TraceVariables_noTrace | Flag off → 0 expressions |
| 3 | TraceVariables_filterPipeline | 1 filter (upcase), Input/Output |
| 4 | TraceVariables_depth | Depth=1 inside if |
| 5 | TraceConditions_if_taken | Branches, Executed=true/false |
| 6 | TraceConditions_else_taken | Else branch executed |
| 7 | TraceConditions_unless | Branch kind="unless" |
| 8 | TraceIterations_basic | Variable, Collection, Length |
| 9 | TraceAssignments_assign | Variable, Value |
| 10 | TraceAssignments_capture | Variable, Value string |
| 11 | Combined | assign + variable together |
| 12 | Error_strictVariables | AuditError returned |
| 13 | ResultNonNilOnError | result never nil |
| 14 | Validate_emptyIF | empty-block diagnostic |
| 15 | Validate_nonEmpty | No false positive |
| 16 | Position_lineNumber | Start.Line correct |
| 17 | AssignSourceLoc | Range.Start.Line > 0 |
| 18 | AssignFilterPipeline | Pipeline in assign |
| 19 | MaxIterItems_TracedCount | Truncated, TracedCount |
| 20 | NoMaxIterItems_AllTraced | No truncation |
| 21 | ConditionComparisons_simple | Operator >=, Left, Right, Result |
| 22 | ConditionComparisons_else_noComparisons | Else without Items |
| 23 | ConditionComparisons_equality | Operator == |
| 24 | ConditionComparisons_groupTrace_and | GroupTrace and with 2 children |
| 25 | Diagnostic_undefinedVariable | Code, Severity |
| 26 | Diagnostic_argumentError | divided_by 0 |
| 27 | Diagnostic_typeMismatch | string == int |
| 28 | Diagnostic_notIterable | for over string |
| 29 | Diagnostic_nilDereference | a.b.c with b=nil |
| 30 | ConditionComparisons_expressionField | Expression not empty |
| 31 | Diagnostic_typeMismatch_hasRange | Range populated |
| 32 | Diagnostic_nilDereference_hasRange | Range populated |
| 33 | Diagnostic_notIterable_hasRange | Range span |
| 34 | Validate_UndefinedFilter | undefined-filter |

---

## 1. VariableTrace — `{{ expr }}`

### 1.1 Basic Attributes (Name, Parts, Value)

| ID | Test | What it validates |
|---|---|---|
| V01 | `{{ x }}` — simple variable | Name="x", Parts=["x"], Value correct |
| V02 | `{{ customer.name }}` — dot access | Name="customer.name", Parts=["customer","name"] |
| V03 | `{{ a.b.c.d }}` — deep dot access | Parts=["a","b","c","d"], Name="a.b.c.d" |
| V04 | `{{ items[0] }}` — array index access | Name contains bracket, Value correct |
| V05 | `{{ "literal" }}` — string literal | Name="\"literal\"" or similar, Value="literal" |
| V06 | `{{ 42 }}` — integer literal | Value=42 (int) |
| V07 | `{{ 3.14 }}` — float literal | Value=3.14 (float64) |
| V08 | `{{ true }}` — boolean literal | Value=true |
| V09 | `{{ false }}` — boolean literal | Value=false |
| V10 | `{{ nil }}` — nil literal | Value=nil |
| V11 | `{{ blank }}` — blank | Value="" or nil |
| V12 | `{{ empty }}` — empty | Corresponding value |
| V13 | Undefined variable (without strict) | Value=nil, no error |
| V14 | Undefined variable (with strict) | Value=nil, Error populated, Diagnostic code="undefined-variable" |
| V15 | Multiple variables in template | len(Expressions) correct, all KindVariable |
| V16 | `{{ hash["key"] }}` — bracket string access | Value correct |

### 1.2 Filter Pipeline

| ID | Test | What it validates |
|---|---|---|
| VP01 | No filters | Pipeline=[] (empty, not nil) |
| VP02 | One filter without args (`upcase`) | len(Pipeline)=1, Filter, Input, Output |
| VP03 | One filter with 1 arg (`truncate: 10`) | Args=[10], Input/Output correct |
| VP04 | One filter with multiple args (`truncate: 10, "..."`) | Args=[10, "..."] |
| VP05 | Chain of 2 filters (`upcase \| truncate: 5`) | len(Pipeline)=2, Output[0]=Input[1] |
| VP06 | Chain of 3+ filters (`downcase \| replace \| truncate`) | Complete chaining |
| VP07 | `default` filter with nil value | Input=nil, Output=default value |
| VP08 | `split` filter (returns array) | Output is []string |
| VP09 | `size` filter on string | Output is int |
| VP10 | `size` filter on array | Output is int |
| VP11 | `times` filter (math) | Input/Output numeric |
| VP12 | `round` filter | Input float → Output int |
| VP13 | `join` filter on array | Input=[]string, Output=string |
| VP14 | `date` filter | Input=string/time, Output=formatted string |
| VP15 | `first` filter on array | Output is first element |
| VP16 | `last` filter on array | Output is last element |
| VP17 | `map` filter on array of maps | Output is slice of values |
| VP18 | `where` filter on array | Output is filtered array |
| VP19 | `sort` filter on array | Input unordered → Output ordered |
| VP20 | `reverse` filter on array | Output inverted |
| VP21 | `compact` filter (removes nils) | Input with nils → Output without nils |
| VP22 | `uniq` filter | Removes duplicates |
| VP23 | Undefined filter with LaxFilters | No error, value passthrough |
| VP24 | Pipeline with erroring filter (e.g. `divided_by: 0`) | Error on Expression, Diagnostic |

### 1.3 Source and Range

| ID | Test | What it validates |
|---|---|---|
| VR01 | Source contains `{{ ... }}` delimiters | Exact source |
| VR02 | Range.Start.Line correct (1st line) | Line=1 |
| VR03 | Range.Start.Line correct (3rd line) | Line=3 |
| VR04 | Range.Start.Column correct | Precise column |
| VR05 | Range.End > Range.Start | End is after Start |
| VR06 | Range.End.Column = Start.Column + len(source) (single-line) | Precise calculation |
| VR07 | Multiple expressions in same template, Ranges do not overlap | No overlap |

### 1.4 Depth

| ID | Test | What it validates |
|---|---|---|
| VD01 | Top-level variable | Depth=0 |
| VD02 | Inside `{% if %}` | Depth=1 |
| VD03 | Inside `{% for %}` | Depth=1 |
| VD04 | Inside nested `{% if %}{% if %}` | Depth=2 |
| VD05 | Inside nested `{% for %}{% if %}` | Depth=2 |
| VD06 | After exiting block, returns to Depth=0 | Correct Depth |

---

## 2. ConditionTrace — `{% if %}`, `{% unless %}`, `{% case %}`

### 2.1 Branch Structure

| ID | Test | What it validates |
|---|---|---|
| C01 | `{% if x %}...{% endif %}` (no else) | 1 branch, kind="if" |
| C02 | `{% if x %}...{% else %}...{% endif %}` | 2 branches: "if" + "else" |
| C03 | `{% if x %}...{% elsif y %}...{% endif %}` | 2 branches: "if" + "elsif" |
| C04 | `{% if x %}...{% elsif y %}...{% else %}...{% endif %}` | 3 branches: "if"+"elsif"+"else" |
| C05 | `{% if x %}...{% elsif y %}...{% elsif z %}...{% else %}...{% endif %}` | 4 branches |
| C06 | `{% unless x %}...{% endunless %}` | 1 branch, kind="unless" |
| C07 | `{% unless x %}...{% else %}...{% endunless %}` | 2 branches: "unless"+"else" |
| C08 | `{% case x %}{% when "a" %}...{% when "b" %}...{% endcase %}` | 2 branches with kind="when" |
| C09 | `{% case x %}{% when "a" %}...{% when "b" %}...{% else %}...{% endcase %}` | 3 branches: "when"+"when"+"else" |
| C10 | `{% case x %}{% when "a","b" %}...{% endcase %}` | When with multiple values |

### 2.2 Executed flag

| ID | Test | What it validates |
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

| ID | Test | What it validates |
|---|---|---|
| CC01 | `x == 1` | Operator="==", Left, Right, Result |
| CC02 | `x != 1` | Operator="!=" |
| CC03 | `x > 5` | Operator=">" |
| CC04 | `x < 5` | Operator="<" |
| CC05 | `x >= 5` | Operator=">=" |
| CC06 | `x <= 5` | Operator="<=" |
| CC07 | `arr contains "a"` | Operator="contains" |
| CC08 | `str contains "sub"` | Operator="contains" on string |
| CC09 | Expression with true result | Result=true |
| CC10 | Expression with false result | Result=false |
| CC11 | Left/Right are correct types (int, string, bool, nil) | Types preserved |
| CC12 | Expression field contains raw text | Expression="x == 1" or similar |
| CC13 | Simple truthiness `{% if x %}` (no operator) | ComparisonTrace with implicit operator == true |

### 2.4 GroupTrace (and/or)

| ID | Test | What it validates |
|---|---|---|
| CG01 | `a and b` (both true) | Operator="and", Result=true, 2 Items |
| CG02 | `a and b` (one false) | Result=false |
| CG03 | `a or b` (both false) | Operator="or", Result=false |
| CG04 | `a or b` (one true) | Result=true |
| CG05 | `a and b and c` (3 terms) | Correct nesting |
| CG06 | `a or b or c` (3 terms) | Correct nesting |
| CG07 | `a and b or c` (mixed) | Correct precedence (Liquid: left-to-right, no precedence) |
| CG08 | `a > 1 and b < 10` with comparisons inside the group | Group.Items[0].Comparison != nil |
| CG09 | Nested group `(a and b) or c` — Liquid has no parentheses, but tests the flat left-to-right | Correct structure |

### 2.5 Branch Range and Source

| ID | Test | What it validates |
|---|---|---|
| CB01 | Branch[0].Range.Start points to `{% if ... %}` | Precise range |
| CB02 | Branch[1].Range.Start points to `{% else %}` | Range for else |
| CB03 | Branch for elsif points to `{% elsif ... %}` | Range for elsif |
| CB04 | Condition Expression.Range covers from {% if %} to {% endif %} | Total range |
| CB05 | Condition Expression.Source contains the if header | Precise source |

### 2.6 Depth in Conditions

| ID | Test | What it validates |
|---|---|---|
| CD01 | Top-level if | Depth=0 |
| CD02 | If inside for | Depth=1 |
| CD03 | Nested if | Depth reflects nesting |

### 2.7 Condition with Error

| ID | Test | What it validates |
|---|---|---|
| CR01 | Undefined variable in condition (strict) | Diagnostic, render continues |

---

## 3. IterationTrace — `{% for %}`, `{% tablerow %}`

### 3.1 Basic Attributes

| ID | Test | What it validates |
|---|---|---|
| I01 | `{% for item in items %}` | Variable="item", Collection="items" |
| I02 | Empty collection | Length=0, TracedCount=0 |
| I03 | Collection with 1 item | Length=1, TracedCount=1 |
| I04 | Collection with 100 items | Length=100 |
| I05 | `{% for item in hash %}` — iteration over map | Works, correct Length |
| I06 | `{% for i in (1..5) %}` — range | Length=5, Variable="i", Collection="(1..5)" |
| I07 | `{% for i in (1..0) %}` — empty/inverted range | Length=0 |

### 3.2 Limit, Offset, Reversed

| ID | Test | What it validates |
|---|---|---|
| IL01 | `limit:3` with 5 items | Limit=ptr(3), Length=5 |
| IL02 | `offset:2` with 5 items | Offset=ptr(2), Length=5 |
| IL03 | `limit:2 offset:1` combined | Both populated |
| IL04 | `reversed` with array | Reversed=true |
| IL05 | Without limit/offset/reversed | Limit=nil, Offset=nil, Reversed=false |
| IL06 | `limit:0` | Limit=ptr(0) |
| IL07 | `offset:continue` (if supported) | Tests or documents limitation |

### 3.3 MaxIterationTraceItems / Truncation

| ID | Test | What it validates |
|---|---|---|
| IT01 | MaxIterItems=0, 10 items | Truncated=false, TracedCount=10 |
| IT02 | MaxIterItems=5, 10 items | Truncated=true, TracedCount=5 |
| IT03 | MaxIterItems=10, 5 items | Truncated=false, TracedCount=5 |
| IT04 | MaxIterItems=1, 100 items | Truncated=true, TracedCount=1, inner expressions limited |
| IT05 | MaxIterItems limits inner expressions but not output | Output complete even with truncation |
| IT06 | Nested for — each for has its own TracedCount | Correct per block |
| IT07 | MaxIterItems with empty collection | Truncated=false, TracedCount=0 |

### 3.4 Inner Expressions in For

| ID | Test | What it validates |
|---|---|---|
| IF01 | For with `{{ item }}` — variable trace appears per iteration | len(var expressions) = Length |
| IF02 | For with `{% if item > 2 %}` — condition trace per iteration | Conditions inside for |
| IF03 | For with `{% assign x = item %}` — assign per iteration | Assignments |
| IF04 | Nested for — inner expressions of inner for | Correct Depth |
| IF05 | MaxIterItems truncates inner expressions | No traces after the cutoff |
| IF06 | `forloop` variables (forloop.first, forloop.last, forloop.index) | Accessible as variables |

### 3.5 Tablerow

| ID | Test | What it validates |
|---|---|---|
| TR01 | `{% tablerow item in items %}` | IterationTrace with Variable, Collection |
| TR02 | Tablerow with `cols:3` | Correct output, trace recorded |
| TR03 | Tablerow with limit/offset | Limit/Offset populated |

### 3.6 Source, Range, Depth

| ID | Test | What it validates |
|---|---|---|
| IR01 | Source contains `{% for ... %}` | Precise source |
| IR02 | Range.Start/End correct | Positions |
| IR03 | Depth=0 top-level, 1 nested | Correct depth |

### 3.7 For with Errors / Edge Cases

| ID | Test | What it validates |
|---|---|---|
| IE01 | For over int — not-iterable | Diagnostic warning |
| IE02 | For over bool — not-iterable | Diagnostic warning |
| IE03 | For over nil — not-iterable or zero iterations | Behavior |
| IE04 | For over string — not-iterable | Diagnostic warning |
| IE05 | For-else (`{% for %}...{% else %}...{% endfor %}`) empty collection | Else executed |

---

## 4. AssignmentTrace — `{% assign %}`

### 4.1 Basic Attributes

| ID | Test | What it validates |
|---|---|---|
| A01 | `{% assign x = "hello" %}` | Variable="x", Value="hello" |
| A02 | `{% assign x = 42 %}` | Value=42 (int) |
| A03 | `{% assign x = 3.14 %}` | Value=3.14 (float) |
| A04 | `{% assign x = true %}` | Value=true |
| A05 | `{% assign x = false %}` | Value=false |
| A06 | `{% assign x = nil %}` | Value=nil |
| A07 | `{% assign x = var %}` — from another variable | Value resolves to binding value |
| A08 | `{% assign x = a.b.c %}` — dot access | Value resolves to nested |
| A09 | Path field empty for simple assign | Path=nil or [] |

### 4.2 Pipeline in Assign

| ID | Test | What it validates |
|---|---|---|
| AP01 | `{% assign x = name \| upcase %}` — 1 filter | Pipeline[0] complete |
| AP02 | `{% assign x = price \| times: 0.9 \| round %}` — chain | 2 FilterSteps |
| AP03 | `{% assign x = arr \| sort \| first %}` — array filters | Correct pipeline |
| AP04 | `{% assign x = "a,b,c" \| split: "," %}` — returns array | Value=["a","b","c"], step Output |
| AP05 | No filters | Pipeline=[] (empty) |
| AP06 | Erroring filter — assign with error | Error/Diagnostic |

### 4.3 Source, Range, Depth

| ID | Test | What it validates |
|---|---|---|
| AR01 | Source contains complete `{% assign ... %}` | Exact string |
| AR02 | Precise range | Line, Column |
| AR03 | Assign inside if | Depth=1 |
| AR04 | Assign inside for | Depth=1, repeated per iteration |

### 4.4 Multiple Assigns

| ID | Test | What it validates |
|---|---|---|
| AM01 | 3 assigns in sequence | 3 Expressions with KindAssignment, in order |
| AM02 | Assign followed by use (`{% assign x %}{{ x }}`) | Assignment before Variable in list |
| AM03 | Reassign same variable | Two assignment traces, different values |

---

## 5. CaptureTrace — `{% capture %}`

### 5.1 Basic Attributes

| ID | Test | What it validates |
|---|---|---|
| CP01 | `{% capture x %}Hello{% endcapture %}` | Variable="x", Value="Hello" |
| CP02 | Capture with expressions inside (`{% capture x %}{{ name }}!{% endcapture %}`) | Value contains rendered result |
| CP03 | Capture with multiple lines | Value contains all content |
| CP04 | Empty capture (`{% capture x %}{% endcapture %}`) | Value="" |
| CP05 | Capture with tags inside (`{% capture x %}{% if true %}yes{% endif %}{% endcapture %}`) | Value="yes" |

### 5.2 Source, Range, Depth

| ID | Test | What it validates |
|---|---|---|
| CPR01 | Source contains `{% capture ... %}` | Precise string |
| CPR02 | Range points to opening of capture | Correct position |
| CPR03 | Depth inside block | Correct depth |

### 5.3 Capture with Inner Traces

| ID | Test | What it validates |
|---|---|---|
| CPI01 | Capture with `{{ var }}` inside — inner variable trace appears? | Inner expressions traced |
| CPI02 | Capture with `{% if %}` inside — inner condition trace | Conditions in array |
| CPI03 | Capture followed by `{{ x }}` using the captured value | Variable trace with capture value |

---

## 6. Diagnostics — Complete Catalog

### 6.1 Runtime Errors

| ID | Test | Diagnostic Code | What it validates |
|---|---|---|---|
| D01 | `{{ ghost }}` with StrictVariables | `undefined-variable` | severity=warning, message contains name |
| D02 | `{{ a.b }}` with a undefined, StrictVariables | `undefined-variable` | Path |
| D03 | Multiple undefined variables | Multiple diagnostics | Accumulates all |
| D04 | `{{ x \| divided_by: 0 }}` | `argument-error` | severity=error |
| D05 | `{{ x \| modulo: 0 }}` | `argument-error` | severity=error |
| D06 | Filter with invalid argument | `argument-error` | Descriptive message |
| D07 | `{% if "str" == 1 %}` type mismatch | `type-mismatch` | severity=warning, message with types |
| D08 | `{% if "str" > 1 %}` type mismatch with > | `type-mismatch` | Operator in message |
| D09 | `{% if nil == 1 %}` — nil vs int | No diagnostic (nil comparison is normal) | No warning |
| D10 | `{% for x in 42 %}` not iterable int | `not-iterable` | severity=warning |
| D11 | `{% for x in true %}` not iterable bool | `not-iterable` | severity=warning |
| D12 | `{% for x in "str" %}` not iterable string | `not-iterable` | severity=warning |
| D13 | `{{ a.b.c }}` with b=nil | `nil-dereference` | severity=warning, message with path |
| D14 | `{{ a.b.c.d }}` with b=nil (deep nil) | `nil-dereference` | Property="c" |
| D15 | `{{ nil_var }}` — simple nil | No diagnostic | Normal behavior |
| D16 | `{% if nil_var %}` — nil in condition | No diagnostic | Normal behavior |

### 6.2 Diagnostics Range and Source

| ID | Test | What it validates |
|---|---|---|
| DR01 | Every Diagnostic has Range.Start.Line >= 1 | Never 0 |
| DR02 | Diagnostic.Source is the raw excerpt | Includes delimiters |
| DR03 | Diagnostic on line 5 of template | Line=5 |
| DR04 | Multiple diagnostics on different lines | Each with its own Range |

### 6.3 Diagnostics and Expressions together

| ID | Test | What it validates |
|---|---|---|
| DE01 | Variable with error → Expression.Error != nil and Diagnostic in array | Dual reference |
| DE02 | Expression.Error and Diagnostics[i] have the same Code | Consistency |
| DE03 | Multiple errors: len(Diagnostics) == len(AuditError.Errors()) | Parity |

### 6.4 Render Continues After Error

| ID | Test | What it validates |
|---|---|---|
| DC01 | Template with error in the middle — output before and after captured | Partial output |
| DC02 | 3 variables: 1st OK, 2nd errors, 3rd OK — output contains 1st and 3rd | Render does not stop |
| DC03 | Multiple different errors in same template | All accumulated |
| DC04 | AuditError.Error() contains count | "N error(s)" |
| DC05 | AuditError.Errors() returns slice with correct types | SourceError interface |

---

## 7. AuditOptions — Granular Control

| ID | Test | What it validates |
|---|---|---|
| O01 | All flags false → Expressions=[] | No trace |
| O02 | Only TraceVariables → only KindVariable | Only variables |
| O03 | Only TraceConditions → only KindCondition | Only conditions |
| O04 | Only TraceIterations → only KindIteration | Only iterations |
| O05 | Only TraceAssignments → KindAssignment + KindCapture | Assigns and captures |
| O06 | All flags true → all kinds present | Everything |
| O07 | Diagnostics always present regardless of flags | Errors don't depend on trace |
| O08 | MaxIterationTraceItems=0 with all flags → no limit | Unlimited |
| O09 | Flags do not affect Output | Output identical to normal Render |

---

## 8. AuditResult — General Structure

### 8.1 Output

| ID | Test | What it validates |
|---|---|---|
| R01 | Output equal to normal Render (simple template) | Parity |
| R02 | Output equal to normal Render (complex template) | Parity |
| R03 | Partial output when there is an error | Content before error |
| R04 | Empty output for empty template | Output="" |

### 8.2 Expressions Ordering

| ID | Test | What it validates |
|---|---|---|
| RO01 | assign before variable in array | Execution order |
| RO02 | for → inner expressions repeated per iteration | Linearized |
| RO03 | if(true) → inner expressions present; if(false) → absent | Only executed branch |
| RO04 | case → only expressions from active when | Correct branch |
| RO05 | Nested: if inside for → correct execution | Linear order |

### 8.3 JSON Serialization

| ID | Test | What it validates |
|---|---|---|
| RJ01 | AuditResult serializes to JSON without error | json.Marshal OK |
| RJ02 | JSON has correct keys (snake_case: traced_count, etc.) | Correct tags |
| RJ03 | JSON omitempty works (nil fields omitted) | No pollution |
| RJ04 | JSON roundtrip: Marshal → Unmarshal → equal | Stability |

---

## 9. Validate — Static Analysis

| ID | Test | What it validates |
|---|---|---|
| VA01 | `{% if true %}{% endif %}` | empty-block info |
| VA02 | `{% unless true %}{% endunless %}` | empty-block info |
| VA03 | `{% for x in items %}{% endfor %}` | empty-block info |
| VA04 | `{% case x %}{% when "a" %}{% endcase %}` | empty-block (if detectable) |
| VA05 | Normal template without issues | Diagnostics=[] |
| VA06 | `{{ x \| no_such_filter }}` | undefined-filter error |
| VA07 | `{{ x \| upcase }}` — valid filter | No diagnostic |
| VA08 | Validate returns Output="" (does not render) | Empty output |
| VA09 | Validate returns Expressions=[] (no execution) | No expressions |
| VA10 | Multiple empty-blocks | All detected |
| VA11 | Empty-block with whitespace (`{% if true %} {% endif %}`) — not empty if it has text | No false positive |
| VA12 | Nested empty block | Detects inner block |

---

## 10. RenderOptions — Interaction with Audit

| ID | Test | What it validates |
|---|---|---|
| RO01 | WithStrictVariables → undefined-variable as diagnostic | Warning captured |
| RO02 | Without StrictVariables → undefined var without diagnostic | Nothing |
| RO03 | WithLaxFilters → unknown filter without error | No diagnostic |
| RO04 | WithGlobals(map) → global variables accessible | Correct value |
| RO05 | WithContext with cancel → clean behavior | No panic |
| RO06 | WithSizeLimit → output truncated but complete trace | All expressions |
| RO07 | WithErrorHandler → handler called AND diagnostic created | Both |

---

## 11. Position & Range — Precision

| ID | Test | What it validates |
|---|---|---|
| P01 | First line, first column | Line=1, Column=1 |
| P02 | Third line | Line=3 |
| P03 | Column > 1 (expression with indent/text before) | Correct column |
| P04 | Range.End.Column for single-line expression | Precise end |
| P05 | Multiple expressions — each Range is unique | No overlap |
| P06 | Template with tabs — column counts bytes/chars correctly | Consistent |
| P07 | Multiline template — expression on last line | Correct line |
| P08 | Expression after long tag — correct Column | Column offset |

---

## 12. Edge Cases and Complex Scenarios

| ID | Test | What it validates |
|---|---|---|
| E01 | Empty template — no crash, result OK | Output="", Expressions=[], Diagnostics=[] |
| E02 | Template with only text — no traces | 0 expressions |
| E03 | Extremely long template (1000+ expressions) | No crash, OK performance |
| E04 | Nested for 3 deep | Correct depth increments |
| E05 | If → for → if → variable | Depth=3 for innermost variable |
| E06 | For with break/continue | Interrupted iterations reflected in trace |
| E07 | `{% comment %}...{% endcomment %}` — no trace | Ignored |
| E08 | `{% raw %}{{ not_parsed }}{% endraw %}` — no trace | Raw ignored |
| E09 | Include/render tag (if template store configured) | Expressions from included? |
| E10 | Increment/decrement tags | No crash |
| E11 | `{% liquid assign x = 1 \n echo x %}` (liquid tag) | Correct traces |
| E12 | Template with Unicode | Values preserve Unicode |
| E13 | Template with whitespace control `{%- if -%}` | Output trimmed, traces still present |
| E14 | Bindings with complex types (structs, interfaces, Drops) | Value resolves correctly |
| E15 | Cycle tag inside for | No crash |

---

## 13. Spec Example — End-to-End

| ID | Test | What it validates |
|---|---|---|
| S01 | Complete template from spec (assign + variable + if + for) | Exact output, Expressions in spec order, all fields |
| S02 | S01 output serialized to JSON and verified against expected | JSON match |

---

## Summary Count

| Category | New Tests | Existing |
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

## Implementation Priority

1. **VariableTrace complete** — most used type with the most fields
2. **ConditionTrace complete** — most complex with branches, groups, comparisons
3. **IterationTrace complete** — has truncation, inner expressions, limit/offset/reversed
4. **AssignmentTrace complete** — pipeline, path
5. **CaptureTrace complete** — relatively simple
6. **Advanced Diagnostics** — dual reference, render continues, multiple errors
7. **Granular AuditOptions** — flag isolation
8. **Advanced Validate** — more static patterns
9. **Position/Range precision** — exact columns
10. **Edge cases** — stress, Unicode, whitespace control, etc.
11. **E2E spec example** — golden test
