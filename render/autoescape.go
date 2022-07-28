package render

import (
	"io"
	"strings"
)

// HtmlEscaper is a Replacer that escapes HTML markup characters. Copied from Go standard library because it's
// not exposed.
var HtmlEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`'`, "&#39;", // "&#39;" is shorter than "&apos;" and apos was not in HTML until HTML5.
	`<`, "&lt;",
	`>`, "&gt;",
	`"`, "&#34;", // "&#34;" is shorter than "&quot;".
)

// Replacer interface is used for auto-escape.
type Replacer interface {
	WriteString(io.Writer, string) (int, error)
}
