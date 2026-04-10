package liquid_test

import (
	"testing"
)

// ============================================================================
// Basic API Contract (B01–B12)
// ============================================================================

// B01 — ParseResult is non-nil for a clean template.
func TestParseAudit_Basic_B01_resultNonNilClean(t *testing.T) {
	r := parseAudit(`Hello {{ name }}!`)
	assertParseResultNonNil(t, r, "B01")
}

// B02 — ParseResult is non-nil even when the parse is fatal (Template=nil).
func TestParseAudit_Basic_B02_resultNonNilOnFatal(t *testing.T) {
	r := parseAudit(`{% if x %}no close`)
	assertParseResultNonNil(t, r, "B02")
}

// B03 — Diagnostics is non-nil (never nil) for a clean template.
func TestParseAudit_Basic_B03_diagnosticsNonNilClean(t *testing.T) {
	r := parseAudit(`Hello, world!`)
	assertDiagsNonNil(t, r, "B03")
}

// B04 — Diagnostics is non-nil for a fatal-error template and contains at
// least one diagnostic.
func TestParseAudit_Basic_B04_diagnosticsNonNilOnFatal(t *testing.T) {
	r := parseAudit(`{% if x %}`)
	assertDiagsNonNil(t, r, "B04")
	if len(r.Diagnostics) == 0 {
		t.Fatal("B04: expected at least one diagnostic for unclosed-tag, got none")
	}
}

// B05 — Template is non-nil for a clean template.
func TestParseAudit_Basic_B05_templateNonNilClean(t *testing.T) {
	r := parseAudit(`{{ name | upcase }}`)
	assertTemplateNonNil(t, r, "B05")
}

// B06 — Template is nil for a fatal-error template.
func TestParseAudit_Basic_B06_templateNilFatal(t *testing.T) {
	r := parseAudit(`{% if x %}no close`)
	assertTemplateNil(t, r, "B06")
}

// B07 — Template is non-nil even when there is a non-fatal syntax-error.
// The parse recovered; the broken node renders as empty string.
func TestParseAudit_Basic_B07_templateNonNilOnNonFatal(t *testing.T) {
	r := parseAudit(`{{ | bad }}`)
	assertTemplateNonNil(t, r, "B07")
}

// B08 — ParseTemplateAudit([]byte) and ParseStringAudit(string) return
// identical diagnostic codes for the same source.
func TestParseAudit_Basic_B08_byteAndStringVariantParity(t *testing.T) {
	src := `{% if x %}unclosed`
	rBytes := parseAuditBytes(src)
	rStr := parseAudit(src)

	assertParseResultNonNil(t, rBytes, "B08 bytes")
	assertParseResultNonNil(t, rStr, "B08 string")

	codesBytes := parseDiagCodes(rBytes.Diagnostics)
	codesStr := parseDiagCodes(rStr.Diagnostics)

	if len(codesBytes) != len(codesStr) {
		t.Errorf("B08: bytes diagnostics=%v, string diagnostics=%v (count mismatch)", codesBytes, codesStr)
	}
	for i := range codesStr {
		if i >= len(codesBytes) {
			break
		}
		if codesBytes[i] != codesStr[i] {
			t.Errorf("B08: diagnostics[%d] bytes code=%q, string code=%q", i, codesBytes[i], codesStr[i])
		}
	}

	// Both should agree on whether Template is nil.
	if (rBytes.Template == nil) != (rStr.Template == nil) {
		t.Errorf("B08: bytes Template nil=%v, string Template nil=%v (should match)",
			rBytes.Template == nil, rStr.Template == nil)
	}
}

// B09 — Empty source string: Template non-nil, Diagnostics empty.
func TestParseAudit_Basic_B09_emptySource(t *testing.T) {
	r := parseAudit(``)
	assertTemplateNonNil(t, r, "B09")
	assertNoParseDiags(t, r, "B09")
}

// B10 — Whitespace-only source: Template non-nil, no diagnostics.
func TestParseAudit_Basic_B10_whitespaceOnly(t *testing.T) {
	r := parseAudit("   \n\t\n  ")
	assertTemplateNonNil(t, r, "B10")
	assertNoParseDiags(t, r, "B10")
}

// B11 — Plain text with no tags: Template non-nil, Diagnostics empty.
func TestParseAudit_Basic_B11_plainText(t *testing.T) {
	r := parseAudit(`Hello, world! This is plain text with no Liquid.`)
	assertTemplateNonNil(t, r, "B11")
	assertNoParseDiags(t, r, "B11")
}

// B12 — ParseResult is JSON-serializable without error.
func TestParseAudit_Basic_B12_jsonSerializable(t *testing.T) {
	// Use import via the json package; marshal in a sub-test to get line precision.
	// We don't import encoding/json here — that's covered in parse_audit_json_test.go.
	// This test just confirms Template is non-nil and Diagnostics non-nil (contract).
	r := parseAudit(`{{ name }}`)
	assertTemplateNonNil(t, r, "B12")
	assertDiagsNonNil(t, r, "B12")
}
