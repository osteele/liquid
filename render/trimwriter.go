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
	w         io.Writer
	buf       bytes.Buffer
	trimRight bool
}

// This violates the letter of the protocol by returning the count of the
// bytes, rather than the actual number of bytes written. We can't know the
// number of bytes written until later, and it won't in general be the same
// as the argument length (that's the whole point of trimming), but speaking
// truthfully here would cause some callers to return io.ErrShortWrite, ruining
// this as an io.Writer.
func (tw *trimWriter) Write(b []byte) (int, error) {
	n := len(b)
	if tw.trimRight {
		b = bytes.TrimLeftFunc(b, unicode.IsSpace)
	} else if tw.buf.Len() > 0 {
		if err := tw.Flush(); err != nil {
			return 0, err
		}
	}
	nonWS := bytes.TrimRightFunc(b, unicode.IsSpace)
	if len(nonWS) < len(b) {
		if _, err := tw.buf.Write(b[len(nonWS):]); err != nil {
			return 0, err
		}
	}
	_, err := tw.w.Write(nonWS)
	return n, err
}
func (tw *trimWriter) Flush() (err error) {
	if tw.buf.Len() > 0 {
		_, err = tw.buf.WriteTo(tw.w)
		tw.buf.Reset()
	}
	return
}

func (tw *trimWriter) TrimLeft(f bool) {
	if !f && tw.buf.Len() > 0 {
		if err := tw.Flush(); err != nil {
			panic(err)
		}
	}
	tw.buf.Reset()
	tw.trimRight = false
}

func (tw *trimWriter) TrimRight(f bool) {
	tw.trimRight = f
}
