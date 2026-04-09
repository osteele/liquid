package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// empty-block — Static Analysis (E01–E21)
// ============================================================================

// E01 — {% if true %}{% endif %}: completely empty if → empty-block diagnostic.
func TestParseAudit_EmptyBlock_E01_emptyIf(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	requireParseDiag(t, r, "empty-block")
}

// E02 — {% unless true %}{% endunless %}: empty unless → empty-block.
func TestParseAudit_EmptyBlock_E02_emptyUnless(t *testing.T) {
	r := parseAudit(`{% unless true %}{% endunless %}`)
	requireParseDiag(t, r, "empty-block")
}

// E03 — {% for x in items %}{% endfor %}: empty for → empty-block.
func TestParseAudit_EmptyBlock_E03_emptyFor(t *testing.T) {
	r := parseAudit(`{% for x in items %}{% endfor %}`)
	requireParseDiag(t, r, "empty-block")
}

// E04 — {% tablerow x in items %}{% endtablerow %}: empty tablerow → empty-block.
func TestParseAudit_EmptyBlock_E04_emptyTablerow(t *testing.T) {
	r := parseAudit(`{% tablerow x in items %}{% endtablerow %}`)
	requireParseDiag(t, r, "empty-block")
}

// E05 — {% if true %}{% else %}{% endif %}: both branches empty → empty-block.
func TestParseAudit_EmptyBlock_E05_bothBranchesEmpty(t *testing.T) {
	r := parseAudit(`{% if true %}{% else %}{% endif %}`)
	requireParseDiag(t, r, "empty-block")
}

// E06 — {% if true %}content{% else %}{% endif %}: else empty but if has content → NOT empty-block.
func TestParseAudit_EmptyBlock_E06_ifHasContentElseEmpty(t *testing.T) {
	r := parseAudit(`{% if true %}content{% else %}{% endif %}`)
	d := firstParseDiag(r, "empty-block")
	if d != nil {
		t.Errorf("E06: unexpected empty-block diagnostic when if branch has content")
	}
}

// E07 — {% if true %}{% else %}content{% endif %}: if empty but else has content → NOT empty-block.
func TestParseAudit_EmptyBlock_E07_elseHasContentIfEmpty(t *testing.T) {
	r := parseAudit(`{% if true %}{% else %}content{% endif %}`)
	d := firstParseDiag(r, "empty-block")
	if d != nil {
		t.Errorf("E07: unexpected empty-block diagnostic when else branch has content")
	}
}

// E08 — {% if true %}   {% endif %}: whitespace-only body.
// This test documents the behavior (may or may not count as empty; both are acceptable).
func TestParseAudit_EmptyBlock_E08_whitespaceOnlyBody(t *testing.T) {
	r := parseAudit("{% if true %}   \n   {% endif %}")
	assertParseResultNonNil(t, r, "E08")
	// Only assert no crash; behavior (empty-block or not) is implementation-defined.
	// Log the decision so it is visible in test output.
	d := firstParseDiag(r, "empty-block")
	t.Logf("E08: whitespace-only body detected as empty-block: %v", d != nil)
}

// E09 — {% if true %}{{ x }}{% endif %}: has expression inside → NOT empty-block.
func TestParseAudit_EmptyBlock_E09_hasExpressionInside(t *testing.T) {
	r := parseAudit(`{% if true %}{{ x }}{% endif %}`)
	d := firstParseDiag(r, "empty-block")
	if d != nil {
		t.Errorf("E09: unexpected empty-block diagnostic when block contains {{ x }}")
	}
}

// E10 — {% if true %}{% assign x = 1 %}{% endif %}: has tag inside → NOT empty-block.
func TestParseAudit_EmptyBlock_E10_hasTagInside(t *testing.T) {
	r := parseAudit(`{% if true %}{% assign x = 1 %}{% endif %}`)
	d := firstParseDiag(r, "empty-block")
	if d != nil {
		t.Error("E10: unexpected empty-block diagnostic when block contains {%% assign %%}")
	}
}

// E11 — {% if true %}text{% endif %}: has static text inside → NOT empty-block.
func TestParseAudit_EmptyBlock_E11_hasTextInside(t *testing.T) {
	r := parseAudit(`{% if true %}hello{% endif %}`)
	d := firstParseDiag(r, "empty-block")
	if d != nil {
		t.Errorf("E11: unexpected empty-block diagnostic when block contains static text")
	}
}

// E12 — empty-block co-exists with undefined-filter in same template.
func TestParseAudit_EmptyBlock_E12_coexistsWithUndefinedFilter(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}{{ x | totally_unknown_xyz }}`)
	assertParseResultNonNil(t, r, "E12")
	hasEmptyBlock := firstParseDiag(r, "empty-block") != nil
	hasUndefinedFilter := firstParseDiag(r, "undefined-filter") != nil
	if !hasEmptyBlock {
		t.Error("E12: expected empty-block diagnostic")
	}
	if !hasUndefinedFilter {
		t.Error("E12: expected undefined-filter diagnostic")
	}
}

// E13 — Code field equals exactly "empty-block".
func TestParseAudit_EmptyBlock_E13_codeField(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	d := requireParseDiag(t, r, "empty-block")
	assertDiagField(t, d.Code, "empty-block", "Code", "empty-block")
}

// E14 — Severity equals exactly "info" (not warning or error).
func TestParseAudit_EmptyBlock_E14_severityInfo(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	d := requireParseDiag(t, r, "empty-block")
	assertDiagField(t, string(d.Severity), string(liquid.SeverityInfo), "Severity", "empty-block")
}

// E15 — Source contains the block opening tag header.
func TestParseAudit_EmptyBlock_E15_sourceContainsHeader(t *testing.T) {
	r := parseAudit(`{% if debug %}{% endif %}`)
	d := requireParseDiag(t, r, "empty-block")
	if len(d.Source) == 0 {
		t.Fatal("E15: empty-block Source is empty")
	}
	assertDiagContains(t, "Source", d.Source, "if", "empty-block")
}

// E16 — Range.Start.Line is correct for the empty block.
func TestParseAudit_EmptyBlock_E16_rangeStartLine(t *testing.T) {
	r := parseAudit("text before\n{% if true %}{% endif %}")
	d := requireParseDiag(t, r, "empty-block")
	if d.Range.Start.Line != 2 {
		t.Errorf("E16: Range.Start.Line=%d, want 2", d.Range.Start.Line)
	}
}

// E17 — Message mentions the block name ("if", "for", etc.).
func TestParseAudit_EmptyBlock_E17_messageContainsBlockName(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	d := requireParseDiag(t, r, "empty-block")
	if len(d.Message) == 0 {
		t.Fatal("E17: empty-block Message is empty")
	}
	// Should mention "if" in the message.
	assertDiagContains(t, "Message", d.Message, "if", "empty-block")
}

// E18 — Two empty-blocks in same template → two empty-block diagnostics.
func TestParseAudit_EmptyBlock_E18_twoEmptyBlocks(t *testing.T) {
	r := parseAudit(`{% if a %}{% endif %}{% if b %}{% endif %}`)
	blocks := allParseDiags(r, "empty-block")
	if len(blocks) != 2 {
		t.Errorf("E18: expected 2 empty-block diagnostics, got %d", len(blocks))
	}
}

// E19 — Nested empty block: inner for inside if is empty.
// {% if true %}{% for x in items %}{% endfor %}{% endif %}
// The inner for is empty → at least empty-block for the for.
func TestParseAudit_EmptyBlock_E19_nestedEmptyBlock(t *testing.T) {
	r := parseAudit(`{% if true %}content{% for x in items %}{% endfor %}{% endif %}`)
	blocks := allParseDiags(r, "empty-block")
	if len(blocks) == 0 {
		t.Error("E19: expected at least one empty-block for the empty for loop inside if")
	}
}

// E20 — {% case x %}{% when "a" %}{% endcase %}: empty when branch (if detectable).
// This test is advisory; behavior depends on implementation depth for case branches.
func TestParseAudit_EmptyBlock_E20_emptyCaseWhen(t *testing.T) {
	r := parseAudit(`{% case x %}{% when "a" %}{% endcase %}`)
	assertParseResultNonNil(t, r, "E20")
	// The behavior (whether empty-block is detected on case/when) is implementation-defined.
	t.Logf("E20: empty-block count for empty case/when: %d", len(allParseDiags(r, "empty-block")))
}

// E21 — Template is still non-nil when there are only empty-block diagnostics.
func TestParseAudit_EmptyBlock_E21_templateNonNilForEmptyBlock(t *testing.T) {
	r := parseAudit(`{% if true %}{% endif %}`)
	assertTemplateNonNil(t, r, "E21")
}
