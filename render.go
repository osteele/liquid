package main

import "io"

type Context struct {
	Variables map[string]interface{}
}

func (n *ASTSeq) Render(w io.Writer, ctx Context) error {
	for _, c := range n.Children {
		if err := c.Render(w, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (n *ASTChunks) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte("{chunks}"))
	return err
}

func (n *ASTText) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte(n.chunk.Source))
	return err
}

func (n *ASTControlTag) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte("{control}"))
	return err
}

func (n *ASTObject) Render(w io.Writer, _ Context) error {
	_, err := w.Write([]byte("{object}"))
	return err
}
