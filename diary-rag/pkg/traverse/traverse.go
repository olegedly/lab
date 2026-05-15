package traverse

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ProcessMarkdownFunc is called for each .md file found during traversal.
type ProcessMarkdownFunc func(content string, filename string, path string) error

// TraverseAndProcessMarkdown walks rootDir recursively and calls processFn for every .md file.
func TraverseAndProcessMarkdown(rootDir string, processFn ProcessMarkdownFunc) error {
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}
			if err := processFn(string(content), d.Name(), path); err != nil {
				return fmt.Errorf("error processing file %s: %w", path, err)
			}
		}
		return nil
	})
}
