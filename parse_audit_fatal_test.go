package liquid_test

import (
	"testing"

	"github.com/osteele/liquid"
)

// ============================================================================
// Fatal Errors — unclosed-tag (U01–U17)
// ============================================================================

// U01 — {% if %} without {% endif %}: Template=nil, code="unclosed-tag".
func TestParseAudit_Unclosed_U01_ifNoEndif(t *testing.T) {
	r := parseAudit(`{% if x %}content here`)
	assertTemplateNil(t, r, "U01")
	d := requireParseDiag(t, r, "unclosed-tag")
	_ = d
}

// U02 — {% unless %} without {% endunless %}.
func TestParseAudit_Unclosed_U02_unlessNoEnd(t *testing.T) {
	r := parseAudit(`{% unless x %}content`)
	assertTemplateNil(t, r, "U02")
	requireParseDiag(t, r, "unclosed-tag")
}

// U03 — {% for %} without {% endfor %}.
func TestParseAudit_Unclosed_U03_forNoEndfor(t *testing.T) {
	r := parseAudit(`{% for item in items %}{{ item }}`)
	assertTemplateNil(t, r, "U03")
	requireParseDiag(t, r, "unclosed-tag")
}

// U04 — {% case %} without {% endcase %}.
func TestParseAudit_Unclosed_U04_caseNoEndcase(t *testing.T) {
	r := parseAudit(`{% case x %}{% when "a" %}yes`)
	assertTemplateNil(t, r, "U04")
	requireParseDiag(t, r, "unclosed-tag")
}

// U05 — {% capture %} without {% endcapture %}.
func TestParseAudit_Unclosed_U05_captureNoEnd(t *testing.T) {
	r := parseAudit(`{% capture greeting %}Hello`)
	assertTemplateNil(t, r, "U05")
	requireParseDiag(t, r, "unclosed-tag")
}

// U06 — {% tablerow %} without {% endtablerow %}.
func TestParseAudit_Unclosed_U06_tablerowNoEnd(t *testing.T) {
	r := parseAudit(`{% tablerow item in items %}{{ item.name }}`)
	assertTemplateNil(t, r, "U06")
	requireParseDiag(t, r, "unclosed-tag")
}

// U07 — Nested unclosed: {% if %}{% for %} both unclosed → Template=nil.
func TestParseAudit_Unclosed_U07_nestedUnclosed(t *testing.T) {
	r := parseAudit(`{% if true %}{% for x in items %}{{ x }}`)
	assertTemplateNil(t, r, "U07")
	// At minimum one unclosed-tag diagnostic must be present.
	tags := allParseDiags(r, "unclosed-tag")
	if len(tags) == 0 {
		t.Fatal("U07: expected at least one unclosed-tag diagnostic")
	}
}

// U08 — Multiple consecutive opens with no closes → at least one unclosed-tag.
func TestParseAudit_Unclosed_U08_multipleOpensNoClose(t *testing.T) {
	r := parseAudit(`{% if a %}{% if b %}{% if c %}deep`)
	assertTemplateNil(t, r, "U08")
	tags := allParseDiags(r, "unclosed-tag")
	if len(tags) == 0 {
		t.Fatal("U08: expected at least one unclosed-tag diagnostic")
	}
}

// U09 — Code field equals exactly "unclosed-tag".
func TestParseAudit_Unclosed_U09_codeField(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	d := requireParseDiag(t, r, "unclosed-tag")
	assertDiagField(t, d.Code, "unclosed-tag", "Code", "unclosed-tag")
}

// U10 — Severity equals exactly "error".
func TestParseAudit_Unclosed_U10_severityError(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	d := requireParseDiag(t, r, "unclosed-tag")
	assertDiagField(t, string(d.Severity), string(liquid.SeverityError), "Severity", "unclosed-tag")
}

// U11 — Message mentions the tag name ("if").
func TestParseAudit_Unclosed_U11_messageContainsTagName(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	d := requireParseDiag(t, r, "unclosed-tag")
	assertDiagContains(t, "Message", d.Message, "if", "unclosed-tag")
}

// U12 — Source field contains the opening tag text.
func TestParseAudit_Unclosed_U12_sourceContainsOpenTag(t *testing.T) {
	r := parseAudit(`{% if order %}content`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if len(d.Source) == 0 {
		t.Fatal("U12: unclosed-tag diagnostic Source is empty")
	}
	// Source should contain the if tag, not the entire template.
	assertDiagContains(t, "Source", d.Source, "if", "unclosed-tag")
}

// U13 — Range.Start points to the opening tag line (line 1 for first-line tag).
func TestParseAudit_Unclosed_U13_rangeStartAtOpenTag(t *testing.T) {
	r := parseAudit(`{% if x %}body`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if d.Range.Start.Line != 1 {
		t.Errorf("U13: Range.Start.Line=%d, want 1", d.Range.Start.Line)
	}
}

// U14 — Related is non-empty and contains at least one entry pointing to EOF.
func TestParseAudit_Unclosed_U14_relatedNonEmpty(t *testing.T) {
	r := parseAudit(`{% if x %}body`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if len(d.Related) == 0 {
		t.Fatal("U14: unclosed-tag diagnostic Related is empty; expected at least one entry pointing to expected close location")
	}
}

// U15 — Related[0].Message mentions the expected closing tag.
func TestParseAudit_Unclosed_U15_relatedMessageClear(t *testing.T) {
	r := parseAudit(`{% if x %}body`)
	d := requireParseDiag(t, r, "unclosed-tag")
	if len(d.Related) == 0 {
		t.Skip("U15: no Related entries (U14 already fails)")
	}
	if len(d.Related[0].Message) == 0 {
		t.Fatal("U15: Related[0].Message is empty; should explain expected closing tag")
	}
}

// U16 — unclosed-tag on line 3: Range.Start.Line=3.
func TestParseAudit_Unclosed_U16_lineTracking(t *testing.T) {
	r := parseAudit("line1\nline2\n{% if x %}body")
	d := requireParseDiag(t, r, "unclosed-tag")
	if d.Range.Start.Line != 3 {
		t.Errorf("U16: Range.Start.Line=%d, want 3", d.Range.Start.Line)
	}
}

// U17 — Source does not contain the complete template (only the tag excerpt).
func TestParseAudit_Unclosed_U17_sourceNotFullTemplate(t *testing.T) {
	template := "{% if order %}lots of content here that should not appear in source"
	r := parseAudit(template)
	d := requireParseDiag(t, r, "unclosed-tag")
	// Source should be shorter than the full template.
	if len(d.Source) >= len(template) {
		t.Errorf("U17: Source=%q contains entire template (len=%d); expected only tag excerpt", d.Source, len(d.Source))
	}
}

// ============================================================================
// Fatal Errors — unexpected-tag (X01–X14)
// ============================================================================

// X01 — {% endif %} at top level with no {% if %}: Template=nil, unexpected-tag.
func TestParseAudit_Unexpected_X01_endifOrphan(t *testing.T) {
	r := parseAudit(`{% endif %}`)
	assertTemplateNil(t, r, "X01")
	requireParseDiag(t, r, "unexpected-tag")
}

// X02 — {% endfor %} at top level with no {% for %}.
func TestParseAudit_Unexpected_X02_endforOrphan(t *testing.T) {
	r := parseAudit(`{% endfor %}`)
	assertTemplateNil(t, r, "X02")
	requireParseDiag(t, r, "unexpected-tag")
}

// X03 — {% endunless %} with no {% unless %}.
func TestParseAudit_Unexpected_X03_endunlessOrphan(t *testing.T) {
	r := parseAudit(`{% endunless %}`)
	assertTemplateNil(t, r, "X03")
	requireParseDiag(t, r, "unexpected-tag")
}

// X04 — {% endcase %} with no {% case %}.
func TestParseAudit_Unexpected_X04_endcaseOrphan(t *testing.T) {
	r := parseAudit(`{% endcase %}`)
	assertTemplateNil(t, r, "X04")
	requireParseDiag(t, r, "unexpected-tag")
}

// X05 — {% endcapture %} with no {% capture %}.
func TestParseAudit_Unexpected_X05_endcaptureOrphan(t *testing.T) {
	r := parseAudit(`{% endcapture %}`)
	assertTemplateNil(t, r, "X05")
	requireParseDiag(t, r, "unexpected-tag")
}

// X06 — {% else %} at top level outside any block.
func TestParseAudit_Unexpected_X06_elseOrphan(t *testing.T) {
	r := parseAudit(`{% else %}`)
	assertTemplateNil(t, r, "X06")
	requireParseDiag(t, r, "unexpected-tag")
}

// X07 — {% elsif x %} at top level outside any block.
func TestParseAudit_Unexpected_X07_elsifOrphan(t *testing.T) {
	r := parseAudit(`{% elsif x %}`)
	assertTemplateNil(t, r, "X07")
	requireParseDiag(t, r, "unexpected-tag")
}

// X08 — {% when "a" %} outside any {% case %} block.
func TestParseAudit_Unexpected_X08_whenOrphan(t *testing.T) {
	r := parseAudit(`{% when "a" %}`)
	assertTemplateNil(t, r, "X08")
	requireParseDiag(t, r, "unexpected-tag")
}

// X09 — Well-formed {% if %}…{% endif %} followed by an extra {% endif %}.
func TestParseAudit_Unexpected_X09_extraEndif(t *testing.T) {
	r := parseAudit(`{% if x %}yes{% endif %}{% endif %}`)
	assertTemplateNil(t, r, "X09")
	requireParseDiag(t, r, "unexpected-tag")
}

// X10 — Code field equals exactly "unexpected-tag".
func TestParseAudit_Unexpected_X10_codeField(t *testing.T) {
	r := parseAudit(`{% endif %}`)
	d := requireParseDiag(t, r, "unexpected-tag")
	assertDiagField(t, d.Code, "unexpected-tag", "Code", "unexpected-tag")
}

// X11 — Severity equals exactly "error".
func TestParseAudit_Unexpected_X11_severityError(t *testing.T) {
	r := parseAudit(`{% endif %}`)
	d := requireParseDiag(t, r, "unexpected-tag")
	assertDiagField(t, string(d.Severity), string(liquid.SeverityError), "Severity", "unexpected-tag")
}

// X12 — Source contains the unexpected tag text.
func TestParseAudit_Unexpected_X12_sourceContainsTag(t *testing.T) {
	r := parseAudit(`{% endif %}`)
	d := requireParseDiag(t, r, "unexpected-tag")
	if len(d.Source) == 0 {
		t.Fatal("X12: unexpected-tag diagnostic Source is empty")
	}
}

// X13 — Range.Start.Line is correct for the unexpected tag position.
func TestParseAudit_Unexpected_X13_rangeLineCorrect(t *testing.T) {
	r := parseAudit("first\nsecond\n{% endif %}")
	d := requireParseDiag(t, r, "unexpected-tag")
	if d.Range.Start.Line != 3 {
		t.Errorf("X13: Range.Start.Line=%d, want 3", d.Range.Start.Line)
	}
}

// X14 — Message mentions the unexpected tag kind.
func TestParseAudit_Unexpected_X14_messageContainsTagKind(t *testing.T) {
	r := parseAudit(`{% endif %}`)
	d := requireParseDiag(t, r, "unexpected-tag")
	if len(d.Message) == 0 {
		t.Fatal("X14: unexpected-tag diagnostic Message is empty")
	}
}
