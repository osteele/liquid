package render

import (
	"bytes"
	"io"
	"unicode"
)

type trimWriter struct {
	w         io.Writer
	buf       []byte
	trimRight bool
}

func (tw *trimWriter) Write(b []byte) (int, error) {
	if tw.trimRight {
		b = bytes.TrimLeftFunc(b, unicode.IsSpace)
	} else if len(tw.buf) > 0 {
		_, err := tw.w.Write(tw.buf)
		tw.buf = []byte{}
		if err != nil {
			return 0, err
		}
	}
	nonWS := bytes.TrimRightFunc(b, unicode.IsSpace)
	if len(nonWS) < len(b) {
		tw.buf = append(tw.buf, b[len(nonWS):]...)
	}
	return tw.w.Write(nonWS)
}
func (tw *trimWriter) Flush() (err error) {
	if tw.buf != nil {
		_, err = tw.w.Write(tw.buf)
		tw.buf = []byte{}
	}
	return
}

func (tw *trimWriter) TrimLeft(f bool) {
	if !f {
		if err := tw.Flush(); err != nil {
			panic(err)
		}
	}
	tw.buf = []byte{}
	tw.trimRight = false
}

func (tw *trimWriter) TrimRight(f bool) {
	tw.trimRight = f
}
