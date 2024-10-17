package render

import (
	"bytes"
	"io"
	"unicode"
)

// A trimWriter provides whitespace control around a wrapped io.Writer.
// The caller should call TrimLeft(bool) and TrimRight(bool) respectively
// before and after processing a tag or expression, and Flush() at completion.
type trimWriter struct {
	w    io.Writer
	buf  bytes.Buffer
	trim bool
}

// Write writes b to the current buffer. If the trim flag is set,
// a prefix whitespace trim on b is performed before writing it to
// the buffer and the trim flag is unset. If the trim flag was not
// set, the current buffer is flushed before b is written.
// Write only returns the bytes written to w during a flush.
func (tw *trimWriter) Write(b []byte) (n int, err error) {
	if tw.trim {
		b = bytes.TrimLeftFunc(b, unicode.IsSpace)
		tw.trim = false
	} else if n, err = tw.Flush(); err != nil {
		return n, err
	}
	_, err = tw.buf.Write(b)
	return
}

// TrimLeft trims all whitespaces before the trim node, i.e. the whitespace
// suffix of the current buffer. It then writes the current buffer to w and
// resets the buffer.
func (tw *trimWriter) TrimLeft() error {
	_, err := tw.w.Write(bytes.TrimRightFunc(tw.buf.Bytes(), unicode.IsSpace))
	tw.buf.Reset()
	return err
}

// TrimRight sets the trim flag on the trimWriter. This will cause a prefix
// whitespace trim on any subsequent write.
func (tw *trimWriter) TrimRight() {
	tw.trim = true
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
