package render

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// FileTemplateStore reads templates beneath Root. An empty Root uses the current directory.
type FileTemplateStore struct {
	Root string
}

func (tl *FileTemplateStore) ReadTemplate(filename string) ([]byte, error) {
	rootPath := tl.Root
	if rootPath == "" {
		rootPath = "."
	}

	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	absFilename := filename
	if !filepath.IsAbs(absFilename) {
		absFilename = filepath.Join(absRoot, absFilename)
	}
	rel, err := filepath.Rel(absRoot, filepath.Clean(absFilename))
	if err != nil || !filepath.IsLocal(rel) {
		return nil, fmt.Errorf("template path %q is outside root %q", filename, rootPath)
	}

	root, err := os.OpenRoot(absRoot)
	if err != nil {
		return nil, err
	}

	source, readErr := root.ReadFile(rel)
	closeErr := root.Close()

	return source, errors.Join(readErr, closeErr)
}
