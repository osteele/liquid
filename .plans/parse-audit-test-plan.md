# Test Plan: ParseTemplateAudit — Complete Coverage

## Status: 0 tests exist today. This plan maps all tests needed.

---

## Inventory of Existing Tests (0)

No tests exist yet for `ParseTemplateAudit` / `ParseStringAudit`.

---

## 1. Basic API Contract — `ParseResult` Structure

These tests validate the invariants that must hold regardless of template content.

| ID | Test | What it validates |
|---|---|---|
| B01 | Clean template → `ParseResult` non-nil | Return value is never nil |
| B02 | Fatal-error template → `ParseResult` non-nil | Still returns struct even when Template=nil |
| B03 | `ParseResult.Diagnostics` is non-nil even for clean template | Never nil, empty slice `[]Diagnostic{}` |
| B04 | `ParseResult.Diagnostics` is non-nil for fatal error | Populated with error diagnostic |
| B05 | Clean template → `ParseResult.Template` non-nil | Compiled template usable |
| B06 | Fatal-error template → `ParseResult.Template` nil | Not usable |
| B07 | Non-fatal error template → `ParseResult.Template` non-nil | Parsing recovered, template usable |
| B08 | `ParseTemplateAudit([]byte)` and `ParseStringAudit(string)` return identical results | Both variants equivalent |
| B09 | `ParseStringAudit("")` — empty source → Template non-nil, Diagnostics empty | Handles empty input |
| B10 | `ParseStringAudit` with only whitespace → Template non-nil, no diagnostics | Whitespace-only not an error |
| B11 | `ParseStringAudit` with only text (no tags) → Template non-nil, Diagnostics empty | Static text always valid |
| B12 | `ParseResult` returned by `ParseTemplateAudit` is JSON-serializable without error | `json.Marshal` succeeds |

---

## 2. Fatal Errors — `Template = nil`

Template becomes nil only for structural block errors that prevent a coherent AST.

### 2.1 `unclosed-tag` — block opened but never closed

| ID | Test | What it validates |
|---|---|---|
| U01 | `{% if x %}` with no `{% endif %}` | Template=nil, code="unclosed-tag", severity="error" |
| U02 | `{% unless x %}` with no `{% endunless %}` | Template=nil, code="unclosed-tag" |
| U03 | `{% for x in items %}` with no `{% endfor %}` | Template=nil, code="unclosed-tag" |
| U04 | `{% case x %}{% when "a" %}` with no `{% endcase %}` | Template=nil, code="unclosed-tag" |
| U05 | `{% capture x %}` with no `{% endcapture %}` | Template=nil, code="unclosed-tag" |
| U06 | `{% tablerow x in items %}` with no `{% endtablerow %}` | Template=nil, code="unclosed-tag" |
| U07 | Nested unclosed: `{% if %}{% for %}` with no `{% endfor %}{% endif %}` | Template=nil, code="unclosed-tag" referring to the innermost open tag |
| U08 | Multiple consecutive opens with no closes | Template=nil, at least one unclosed-tag diagnostic |
| U09 | `unclosed-tag` diagnostic — Code field equals `"unclosed-tag"` | Exact code string |
| U10 | `unclosed-tag` diagnostic — Severity equals `"error"` | Exact severity string |
| U11 | `unclosed-tag` diagnostic — Message mentions the tag name (e.g. `"if"`) | Message clarity |
| U12 | `unclosed-tag` diagnostic — Source contains the opening tag source | Correct source excerpt |
| U13 | `unclosed-tag` diagnostic — Range.Start points to the opening tag line | Accurate range origin |
| U14 | `unclosed-tag` diagnostic — Related is non-empty and points to EOF or expected close | LSP Related populated |
| U15 | `unclosed-tag` Related[0].Message mentions expected closing tag | Message clarity in Related |
| U16 | `unclosed-tag` on line 3 of a multi-line template — Range.Start.Line=3 | Correct line number |
| U17 | `unclosed-tag` diagnostic — `Source` does not contain the entire template | Only the tag excerpt |

### 2.2 `unexpected-tag` — closing tag with no matching open

| ID | Test | What it validates |
|---|---|---|
| X01 | `{% endif %}` at top level with no `{% if %}` | Template=nil, code="unexpected-tag", severity="error" |
| X02 | `{% endfor %}` at top level with no `{% for %}` | Template=nil, code="unexpected-tag" |
| X03 | `{% endunless %}` with no `{% unless %}` | Template=nil, code="unexpected-tag" |
| X04 | `{% endcase %}` with no `{% case %}` | Template=nil, code="unexpected-tag" |
| X05 | `{% endcapture %}` with no `{% capture %}` | Template=nil, code="unexpected-tag" |
| X06 | `{% else %}` outside any block | Template=nil, code="unexpected-tag" |
| X07 | `{% elsif x %}` outside any block | Template=nil, code="unexpected-tag" |
| X08 | `{% when "a" %}` outside any `{% case %}` | Template=nil, code="unexpected-tag" |
| X09 | Well-formed `{% if %}…{% endif %}` followed by extra `{% endif %}` | Template=nil, second `{% endif %}` → unexpected-tag |
| X10 | `unexpected-tag` diagnostic — Code field equals `"unexpected-tag"` | Exact code |
| X11 | `unexpected-tag` diagnostic — Severity equals `"error"` | Exact severity |
| X12 | `unexpected-tag` diagnostic — Source contains the bad tag source | Correct source excerpt |
| X13 | `unexpected-tag` diagnostic — Range.Start.Line correct | Accurate line |
| X14 | `unexpected-tag` diagnostic — Message mentions the tag kind | Message clarity |

---

## 3. Non-Fatal Errors — `Template != nil`, `syntax-error`

Expression-level errors: the block structure is intact, the bad token is replaced with a no-op `ASTBroken` node and parsing continues.

### 3.1 Single syntax-error on `{{ }}`

| ID | Test | What it validates |
|---|---|---|
| S01 | `{{ | bad }}` — invalid expression in object | Template non-nil, code="syntax-error", severity="error" |
| S02 | `{{ product.price | | round }}` — double pipe | Template non-nil, code="syntax-error" |
| S03 | `{{ }}` — empty object expression | Template non-nil, code="syntax-error" (if engine rejects it) |
| S04 | `syntax-error` diagnostic — Code="syntax-error" | Exact code |
| S05 | `syntax-error` diagnostic — Severity="error" | Exact severity |
| S06 | `syntax-error` diagnostic — Source contains `{{ ... }}` delimiters | Source field |
| S07 | `syntax-error` diagnostic — Range.Start.Line correct | Line position |
| S08 | `syntax-error` diagnostic — Range.Start.Column correct | Column position |
| S09 | `syntax-error` diagnostic — Message contains description of error | Non-empty message |

### 3.2 Syntax-error on tag args

| ID | Test | What it validates |
|---|---|---|
| ST01 | `{% assign x = | bad %}` — broken expression in assign | Template non-nil, code="syntax-error" |
| ST02 | `{% if | condition %}` — broken expression in if args | Template non-nil, code="syntax-error" |
| ST03 | `{% for %}` — missing iteration spec | Template non-nil, code="syntax-error" or similar |
| ST04 | Tag-level syntax-error — Source contains `{% ... %}` delimiters in diagnostic | Correct source |

### 3.3 Multiple syntax-errors in same template

| ID | Test | What it validates |
|---|---|---|
| SM01 | Two bad `{{ }}` objects in template | Template non-nil, len(Diagnostics)=2, both code="syntax-error" |
| SM02 | Three bad `{{ }}` objects | Template non-nil, len(Diagnostics)=3 |
| SM03 | Mix of bad `{{ }}` and bad `{% tag %}` | Template non-nil, all syntax-errors collected |
| SM04 | Text between two bad expressions is rendered correctly | Output is correct for valid surrounding content |
| SM05 | ASTBroken renders as empty string — no output from bad node | Broken node outputs nothing |
| SM06 | Two syntax-errors on different lines — each Diagnostic has correct distinct Range | Different Ranges |
| SM07 | All Diagnostics in multi-error result have distinct source fields | No duplicate entries |
| SM08 | Multiple syntax-errors — `len(Diagnostics)` matches count of bad tokens | Accurate accumulation |

### 3.4 Rendering a non-fatal template (integration)

| ID | Test | What it validates |
|---|---|---|
| SR01 | Template with syntax-error renders cleanly (ASTBroken → empty string) | No runtime panic or error |
| SR02 | `{{ bad | | body }}text after{{valid_var}}` — "text after" and valid var rendered | Render continues after broken node |
| SR03 | Template returned by ParseTemplateAudit can be used with `RenderAudit` | Full pipeline works |

---

## 4. `undefined-filter` — Static Analysis

Detected by static AST walk after successful parse. Filter name not registered in the engine.

| ID | Test | What it validates |
|---|---|---|
| F01 | `{{ x | no_such_filter }}` — unknown filter | Diagnostics contains code="undefined-filter" |
| F02 | `{{ x | upcase }}` — valid filter | No undefined-filter diagnostic |
| F03 | `{{ x | no_such \| upcase }}` — one bad filter in chain | One undefined-filter diagnostic |
| F04 | `{{ x | one_bad \| two_bad }}` — two unknown filters in chain | Two undefined-filter diagnostics |
| F05 | `{{ x \| bad }}` and `{{ y \| also_bad }}` — two bad objects | Two undefined-filter diagnostics |
| F06 | `{% assign x = val \| bad_filter %}` — unknown filter in assign | One undefined-filter diagnostic |
| F07 | `{% capture x %}{{ val \| bad_filter }}{% endcapture %}` — unknown in capture | One undefined-filter diagnostic |
| F08 | `undefined-filter` diagnostic — Code="undefined-filter" | Exact code |
| F09 | `undefined-filter` diagnostic — Severity="error" | Exact severity |
| F10 | `undefined-filter` diagnostic — Source contains the full expression | Source field |
| F11 | `undefined-filter` diagnostic — Range points to the expression | Line correct |
| F12 | `undefined-filter` diagnostic — Message mentions the filter name | Mentions the bad name |
| F13 | `undefined-filter` co-exists with `syntax-error` — both in Diagnostics | Multiple distinct codes |
| F14 | Engine with custom registered filter — that filter does not produce undefined-filter | No false positive |
| F15 | `WithLaxFilters()` context: undefined filter — undefined-filter still detected at parse time (parse is filter-agnostic) | Parse is independent of render options |
| F16 | `undefined-filter` template — Template is still non-nil (non-fatal) | Parse recovery works |

---

## 5. `empty-block` — Static Analysis

Detected by static AST walk. Block with no meaningful content in any branch.

| ID | Test | What it validates |
|---|---|---|
| E01 | `{% if true %}{% endif %}` — completely empty if | Diagnostics contains code="empty-block", severity="info" |
| E02 | `{% unless true %}{% endunless %}` — empty unless | code="empty-block" |
| E03 | `{% for x in items %}{% endfor %}` — empty for | code="empty-block" |
| E04 | `{% tablerow x in items %}{% endtablerow %}` — empty tablerow | code="empty-block" |
| E05 | `{% if true %}{% else %}{% endif %}` — both branches empty | code="empty-block" |
| E06 | `{% if true %}content{% else %}{% endif %}` — else empty but if branch has content — NOT empty-block | No diagnostic |
| E07 | `{% if true %}{% else %}content{% endif %}` — if empty but else has content — NOT empty-block | No diagnostic |
| E08 | `{% if true %}  {% endif %}` — whitespace only inside if | Behavior documented (does whitespace-only count?) |
| E09 | `{% if true %}{{ x }}{% endif %}` — has expression inside | No empty-block |
| E10 | `{% if true %}{% assign x = 1 %}{% endif %}` — has tag inside | No empty-block |
| E11 | `{% if true %}text{% endif %}` — has text inside | No empty-block |
| E12 | `empty-block` co-exists with `undefined-filter` — both in Diagnostics | Multiple distinct codes |
| E13 | `empty-block` diagnostic — Code="empty-block" | Exact code |
| E14 | `empty-block` diagnostic — Severity="info" | Exactly "info" not "warning" or "error" |
| E15 | `empty-block` diagnostic — Source contains `{% if ... %}` header | Source field |
| E16 | `empty-block` diagnostic — Range.Start.Line correct | Accurate line |
| E17 | `empty-block` diagnostic — Message mentions the block name | Message clarity |
| E18 | Two empty-blocks in same template — two empty-block diagnostics | Accumulation |
| E19 | Nested empty block: `{% if true %}{% for x in items %}{% endfor %}{% endif %}` — inner for empty | Inner block detected |
| E20 | `{% case x %}{% when "a" %}{% endcase %}` — when branch empty | Detects empty when (if implementable) |
| E21 | `empty-block` template — Template is still non-nil | Non-fatal |

---

## 6. Multiple Diagnostics — Accumulation

Tests that verify that multiple different issues are collected together in a single `Diagnostics` slice.

| ID | Test | What it validates |
|---|---|---|
| M01 | `undefined-filter` + `empty-block` in same template | len(Diagnostics)=2, distinct codes |
| M02 | Two `undefined-filter` + one `empty-block` | len(Diagnostics)=3 |
| M03 | `syntax-error` + `undefined-filter` | Both codes present in result |
| M04 | `syntax-error` + `empty-block` | Both present |
| M05 | `syntax-error` + `undefined-filter` + `empty-block` | All three present |
| M06 | Three `undefined-filter` for three different bad filters on different lines | len = 3, each Diagnostic has distinct Range |
| M07 | Two `empty-block` on separate blocks | len = 2 |
| M08 | Mixed template: clean sections + bad sections | Only bad sections produce diagnostics |
| M09 | Multiple diagnostics from different categories — source order preserved | Diagnostics in source order (by Range.Start.Line) |
| M10 | Single fatal error template — only one diagnostic (not duplicated) | len(Diagnostics)=1 |
| M11 | Fatal error template — no static analysis diagnostics (AST walk skipped when Template=nil) | No false undefined-filter or empty-block |

---

## 7. Diagnostic Field Completeness

For each diagnostic code, all fields must be well-formed.

### 7.1 All codes — shared field checks

| ID | Test | What it validates |
|---|---|---|
| DF01 | Every Diagnostic has non-empty Code | No blank code |
| DF02 | Every Diagnostic has Severity in {"error","warning","info"} | Valid severity values |
| DF03 | Every Diagnostic has non-empty Message | Useful human message |
| DF04 | Every Diagnostic has Source != "" | Source excerpt always populated |
| DF05 | Every Diagnostic has Range.Start.Line >= 1 | 1-based, never 0 |
| DF06 | Every Diagnostic has Range.Start.Column >= 1 | 1-based, never 0 |
| DF07 | Every Diagnostic has Range.End.Line >= Range.Start.Line | End not before Start |
| DF08 | err-severity diagnostics have Severity="error" | Correct mapping |
| DF09 | info-severity diagnostics have Severity="info" | Correct mapping |

### 7.2 `unclosed-tag` — Related field

| ID | Test | What it validates |
|---|---|---|
| DF10 | `unclosed-tag` — Related is non-nil and non-empty | At least one related entry |
| DF11 | `unclosed-tag` Related[0].Range.Start.Line >= 1 | Valid line |
| DF12 | `unclosed-tag` Related[0].Message non-empty | Explains what was expected |
| DF13 | `syntax-error` — Related field is nil (not used for syntax errors) | No spurious Related |
| DF14 | `undefined-filter` — Related field is nil | Not used |
| DF15 | `empty-block` — Related field is nil | Not used |

---

## 8. Range and Position Precision

| ID | Test | What it validates |
|---|---|---|
| P01 | `{{ bad | | }}` on line 1, column 1 — Range.Start.Line=1, Column=1 | First line/col |
| P02 | Three-line template, bad expression on line 3 — Range.Start.Line=3 | Line tracking |
| P03 | Template starting with text before bad expression — Start.Column > 1 | Column offset |
| P04 | Bad expression `{{ bad | | }}` is 14 chars — End.Column = Start.Column + 14 | Span calculation |
| P05 | Two diagnostics on different lines — each has distinct Range | No Range sharing |
| P06 | `{% if bad_expr %}` — Range points to the tag line, not EOF | Tag position |
| P07 | `unclosed-tag` Range.Start.Line = line of opening tag, not EOF | Opening tag range |
| P08 | `unclosed-tag` Related[0].Range.Start.Line = EOF line | EOF related range |
| P09 | Multi-line template with 10 lines, expression on line 7 | Line=7 |
| P10 | Template with Windows line endings (`\r\n`) — line numbers still correct | CRLF tolerance |
| P11 | Template with tabs before expression — Column counts correctly | Tab handling |
| P12 | `ParseTemplateLocation(src, "myfile.html", 5)` — Diagnostic.Range.Start.Line accounts for base line | Line offset |

---

## 9. `ParseTemplateAudit` vs `ParseStringAudit`

| ID | Test | What it validates |
|---|---|---|
| PA01 | Same malformed source via `ParseTemplateAudit([]byte)` and `ParseStringAudit(string)` — identical Diagnostics | Variant parity |
| PA02 | Same clean source via both variants — both return Template non-nil | Variant parity |
| PA03 | `ParseStringAudit` is callable without explicit engine — uses engine's config | Engine config applies |
| PA04 | `ParseTemplateAudit` with nil source → no panic | Robustness |

---

## 10. Integration — `ParseTemplateAudit → RenderAudit` Pipeline

| ID | Test | What it validates |
|---|---|---|
| I01 | Parse-clean template → render with `RenderAudit` succeeds | Full pipeline happy path |
| I02 | Parse with `syntax-error` (non-fatal) → template renders as if broken node = empty string | ASTBroken → empty render |
| I03 | Parse with `undefined-filter` → render with `RenderAudit` — no runtime panic (filter not invoked) | Graceful runtime |
| I04 | `allDiags = parseResult.Diagnostics + auditResult.Diagnostics` — no overlap of identical errors | Parse and render diag types are distinct |
| I05 | Parse with `empty-block` → render with `RenderAudit` — output is empty string (block does nothing) | Consistent behavior |
| I06 | Parse fatal → Template=nil → `RenderAudit` not callable (guarded by caller) | Nil template not panic |
| I07 | Complete pipeline: parse (clean) → render (strict vars) → collect all diagnostics | End-to-end scenario |
| I08 | `ParseStringAudit` followed by `Template.Validate()` — Validate diagnostics not duplicated from parse | No double-reporting |

---

## 11. Engine Configuration Interaction

| ID | Test | What it validates |
|---|---|---|
| EC01 | Engine with `RegisterFilter("my_filter", fn)` → `{{ x | my_filter }}` → no `undefined-filter` | Custom filter recognized |
| EC02 | Engine without custom filter → `{{ x | my_filter }}` → `undefined-filter` | Default engine |
| EC03 | Two engines: one has filter, one doesn't — same source gives different Diagnostics | Engine-specific |
| EC04 | `ParseStringAudit` on engine configured with `SetTrimTagLeft(true)` — no crash | Config doesn't break audit |
| EC05 | `ParseStringAudit` on engine with `StrictVariables()` — parse is not affected (strict is render-time) | Parse is independent |
| EC06 | `LaxFilters()` on engine — `undefined-filter` still detected at parse time (static walk is unconditional) | Parse walk independent of render-time lax flag (or documents if lax suppresses it) |

---

## 12. JSON Serialization of `ParseResult`

| ID | Test | What it validates |
|---|---|---|
| J01 | Clean `ParseResult` with no diagnostics serializes to `{"diagnostics":[]}` | Empty array not null |
| J02 | `ParseResult.Template` is `omitempty` — absent from JSON when nil | omitempty on nil template |
| J03 | Diagnostic JSON keys are snake_case: `"start_line"` or nested struct keys | JSON tag correctness |
| J04 | `Diagnostic.Related` absent from JSON when nil/empty | omitempty on Related |
| J05 | `Diagnostic.Range` always present in JSON (not omitted even when zero) | Required field |
| J06 | Full round-trip: Marshal `ParseResult` → Unmarshal → re-Marshal → same JSON | Serialization stability |
| J07 | `Diagnostic.Severity` serializes as string (not int) | string "error" not 0 |
| J08 | `Position.Line` and `Position.Column` serialize as numbers in JSON | Correct types |

---

## 13. Edge Cases and Robustness

| ID | Test | What it validates |
|---|---|---|
| ED01 | Empty source `""` → no diagnostics, Template non-nil | Edge case |
| ED02 | Source with only a comment `{% comment %}...{% endcomment %}` → no diagnostics | Comment is valid |
| ED03 | `{% raw %}{{ not_parsed }}{% endraw %}` → no syntax-error for the raw content | Raw tag bypass |
| ED04 | Very long template (5000+ tokens) → no crash, diagnostics correct | Performance / robustness |
| ED05 | Template with Unicode in string literals — no crash | Unicode tolerance |
| ED06 | Template with Unicode in variable names (if supported) — no crash | Unicode identifiers |
| ED07 | Template with whitespace-control `{%- if -%}` without close — `unclosed-tag` | Trim markers don't confuse |
| ED08 | Template with correctly nested blocks but maximum depth (10+) — no crash | Deep nesting |
| ED09 | Template with `{% liquid assign x = 1 %}` multi-line tag — no crash | liquid tag |
| ED10 | Multiple `{% assign x = | bad %}` — each produces its own syntax-error | Per-node recovery |
| ED11 | `{{ x | unknown_filter }}` + `{% if true %}{% endif %}` — both diagnostics present | Multiple static issues |
| ED12 | Template that is valid Liquid but uses `continue` or `break` — no crash | Iteration control tags |
| ED13 | Template with `{% increment x %}{% decrement x %}` — no false diagnostics | Count tags |
| ED14 | Template with `{% cycle "a","b" %}` inside for — no false diagnostics | Cycle tag |

---

## 14. `ParseTemplate` vs `ParseTemplateAudit` — Behavioral Parity

These tests ensure `ParseTemplateAudit` produces exactly the same compiled template as `ParseTemplate` for clean inputs, and that its diagnosed non-fatal errors would have caused a real error in `ParseTemplate`.

| ID | Test | What it validates |
|---|---|---|
| PB01 | Clean source: `ParseTemplate` succeeds, `ParseTemplateAudit.Template` is non-nil — both render identically | Same outcome |
| PB02 | Fatal source: `ParseTemplate` returns error, `ParseTemplateAudit.Template` is nil | Both signal failure |
| PB03 | Non-fatal source with syntax-error: `ParseTemplate` returns error, `ParseTemplateAudit.Template` is non-nil | Audit recovers where normal parse fails |
| PB04 | Clean template from audit: `Template.Render(vars)` output same as `eng.ParseAndRenderString(src, vars)` | Output parity |
| PB05 | `ParseTemplateAudit` on clean source → no diagnostics — same as parsing with `ParseTemplate` (no error) | No false alarms |

---

## 15. `Validate()` Overlap (if `Validate()` is implemented)

`Validate()` is a method on a compiled `*Template` and performs static analysis of the AST. `ParseTemplateAudit` performs the same analysis at parse time. These tests ensure there is no double-reporting when both are used.

| ID | Test | What it validates |
|---|---|---|
| VA01 | `ParseStringAudit` on template with `empty-block` → diagnostic present in ParseResult.Diagnostics | Parse-time detection |
| VA02 | Same template: `ParseString` + `Validate()` → diagnostic present in AuditResult.Diagnostics | Validate-time detection |
| VA03 | Full pipeline: `ParseStringAudit` + `RenderAudit` → empty-block appears in parse diags, NOT again in render diags | No duplication |
| VA04 | `undefined-filter` via `ParseStringAudit`: in parse diags | Parse-time detection |
| VA05 | `undefined-filter` via `ParseString` + `Validate()`: in validate diags | Validate-time detection |

---

## Summary Count

| Category | Tests |
|---|---|
| 1. Basic API contract | 12 |
| 2. Fatal errors (unclosed-tag) | 17 |
| 2. Fatal errors (unexpected-tag) | 14 |
| 3. Non-fatal syntax-error (single) | 9 |
| 3. Non-fatal syntax-error (tag args) | 4 |
| 3. Non-fatal syntax-error (multiple) | 8 |
| 3. Non-fatal syntax-error (rendering) | 3 |
| 4. undefined-filter | 16 |
| 5. empty-block | 21 |
| 6. Multiple diagnostics accumulation | 11 |
| 7. Diagnostic field completeness | 15 |
| 8. Range and position precision | 12 |
| 9. ParseTemplateAudit vs ParseStringAudit | 4 |
| 10. Integration pipeline | 8 |
| 11. Engine config interaction | 6 |
| 12. JSON serialization | 8 |
| 13. Edge cases and robustness | 14 |
| 14. Behavioral parity with ParseTemplate | 5 |
| 15. Validate() overlap | 5 |
| **Total** | **192** |

---

## Implementation Priority

1. **Basic API contract** — B01–B12 must pass before anything else is useful
2. **Fatal errors** — U-series and X-series; most critical user-facing behavior
3. **Non-fatal syntax-error** — S/ST/SM-series; validates ASTBroken recovery
4. **undefined-filter** — F-series; static analysis pass
5. **empty-block** — E-series; static analysis pass
6. **Multiple diagnostics accumulation** — M-series; validates aggregation
7. **Diagnostic field completeness** — DF-series; validates all fields
8. **Range and position precision** — P-series; validates scanner-level data
9. **JSON serialization** — J-series; validates public contract
10. **Integration pipeline** — I-series; validates ParseTemplateAudit→RenderAudit
11. **Engine config interaction** — EC-series; validates engine scoping
12. **Edge cases** — ED-series; stress and corner cases
13. **Behavioral parity** — PB-series; regression guard
14. **Validate() overlap** — VA-series; anti-duplication guard
