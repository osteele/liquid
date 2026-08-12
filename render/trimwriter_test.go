package render

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTrimWriterWriteContract(t *testing.T) {
	var output bytes.Buffer
	w := trimWriter{w: &output}
	w.TrimRight()

	n, err := w.Write([]byte("  text"))
	require.NoError(t, err)
	require.Equal(t, len("  text"), n)
	_, err = w.Flush()
	require.NoError(t, err)
	require.Equal(t, "text", output.String())
}
