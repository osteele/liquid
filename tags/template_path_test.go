package tags

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveTemplatePath(t *testing.T) {
	source := filepath.Join("templates", "pages", "index.liquid")
	path, err := resolveTemplatePath(source, filepath.Join("partials", "card.liquid"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join("templates", "pages", "partials", "card.liquid"), path)

	for _, name := range []string{"", "..", filepath.Join("..", "secret.liquid"), filepath.Join(string(filepath.Separator), "secret.liquid")} {
		_, err := resolveTemplatePath(source, name)
		require.Errorf(t, err, "expected %q to be rejected", name)
	}
}
