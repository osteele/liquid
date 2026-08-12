package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFileTemplateStoreConfinesReadsToRoot(t *testing.T) {
	parent := t.TempDir()
	rootPath := filepath.Join(parent, "templates")
	require.NoError(t, os.Mkdir(rootPath, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(rootPath, "inside.liquid"), []byte("inside"), 0o600))
	outsidePath := filepath.Join(parent, "outside.liquid")
	require.NoError(t, os.WriteFile(outsidePath, []byte("outside"), 0o600))

	store := &FileTemplateStore{Root: rootPath}
	source, err := store.ReadTemplate("inside.liquid")
	require.NoError(t, err)
	require.Equal(t, "inside", string(source))

	_, err = store.ReadTemplate("../outside.liquid")
	require.Error(t, err)

	linkPath := filepath.Join(rootPath, "linked.liquid")
	if err := os.Symlink(outsidePath, linkPath); err == nil {
		_, err = store.ReadTemplate("linked.liquid")
		require.Error(t, err)
	}
}
