package tags

import (
	"fmt"
	"path/filepath"
)

func resolveTemplatePath(sourceFile, templateName string) (string, error) {
	if templateName == "" {
		return "", fmt.Errorf("template path is empty")
	}
	if filepath.IsAbs(templateName) || filepath.VolumeName(templateName) != "" {
		return "", fmt.Errorf("template path %q must be relative", templateName)
	}

	clean := filepath.Clean(templateName)
	if !filepath.IsLocal(clean) {
		return "", fmt.Errorf("template path %q escapes its source directory", templateName)
	}

	return filepath.Join(filepath.Dir(sourceFile), clean), nil
}
