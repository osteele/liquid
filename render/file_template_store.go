package render

import (
	"os"
)

type FileTemplateStore struct{}

func (tl *FileTemplateStore) ReadTemplate(filename string) ([]byte, error) {
	source, err := os.ReadFile(filename) // #nosec G304 - template paths are trusted
	return source, err
}
