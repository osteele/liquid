package render

import (
	"bytes"
	"io"
	"unicode"
)

// A trimWriter provides whitespace control around a wrapped io.Writer.
// The caller should call TrimLeft/TrimRight (greedy) or TrimLeftNonGreedy/
// TrimRightNonGreedy (non-greedy) respectively before and after processing a
// tag or expression, and Flush() at completion.
type trimWriter struct {
	w             io.Writer
	buf           bytes.Buffer
	trim          bool // greedy right-trim pending
	trimNonGreedy bool // non-greedy right-trim pending
}

// isInlineBlank returns true for space and horizontal-tab characters only.
// Used by non-greedy trim (mirrors LiquidJS INLINE_BLANK mask).
func isInlineBlank(r rune) bool {
	return r == ' ' || r == '\t'
}

// nonGreedyTrimLeft removes leading inline-blank (space/tab) characters from b,
// then removes at most one trailing newline.
func nonGreedyTrimLeft(b []byte) []byte {
	i := 0
	for i < len(b) && (b[i] == ' ' || b[i] == '\t') {
		i++
	}
	if i < len(b) && b[i] == '\n' {
		i++
	}
	return b[i:]
}

// Write writes b to the current buffer. If a trim flag is set,
// a prefix whitespace trim on b is performed before writing it to
// the buffer and the trim flag is unset. If no trim flag was set,
// the current buffer is flushed before b is written.
// Write only returns the bytes written to w during a flush.
func (tw *trimWriter) Write(b []byte) (n int, err error) {
	if tw.trim {
		b = bytes.TrimLeftFunc(b, unicode.IsSpace)
		tw.trim = false
		tw.trimNonGreedy = false
	} else if tw.trimNonGreedy {
		b = nonGreedyTrimLeft(b)
		tw.trimNonGreedy = false
	} else if n, err = tw.Flush(); err != nil {
		return n, err
	}

	_, err = tw.buf.Write(b)

	return
}

// TrimLeft trims all whitespace before the trim node, i.e. the whitespace
// suffix of the current buffer (greedy). It then writes the current buffer to
// w and resets the buffer.
func (tw *trimWriter) TrimLeft() error {
	tw.trimNonGreedy = false
	_, err := tw.w.Write(bytes.TrimRightFunc(tw.buf.Bytes(), unicode.IsSpace))
	tw.buf.Reset()

	return err
}

// TrimLeftNonGreedy trims only trailing inline-blank (space/tab) characters
// from the current buffer, then writes the buffer to w and resets it.
func (tw *trimWriter) TrimLeftNonGreedy() error {
	tw.trimNonGreedy = false
	_, err := tw.w.Write(bytes.TrimRightFunc(tw.buf.Bytes(), isInlineBlank))
	tw.buf.Reset()

	return err
}

// TrimRight sets the greedy trim flag on the trimWriter. This will cause a
// full (all whitespace) prefix trim on any subsequent write.
func (tw *trimWriter) TrimRight() {
	tw.trim = true
	tw.trimNonGreedy = false // greedy overrides non-greedy
}

// TrimRightNonGreedy sets the non-greedy trim flag when no greedy flag is
// already pending. The next write will trim only leading inline blanks plus
// at most one newline.
func (tw *trimWriter) TrimRightNonGreedy() {
	if !tw.trim {
		tw.trimNonGreedy = true
	}
}

// Flush flushes the current buffer into w.
func (tw *trimWriter) Flush() (int, error) {
	if tw.buf.Len() > 0 {
		n, err := tw.buf.WriteTo(tw.w)
		tw.buf.Reset()

		return int(n), err
	}

	return 0, nil
}
