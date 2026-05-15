package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ProcessMarkdownFunc defines the signature for the callback function.
// It receives the file's text content, its filename, and its full path.
type ProcessMarkdownFunc func(content string, filename string, path string) error

var GlobalIndex = 0

// TraverseAndProcessMarkdown recursively walks the tree starting at rootDir.
// For every ".md" file it encounters, it reads the content and calls processFn.
func TraverseAndProcessMarkdown(rootDir string, processFn ProcessMarkdownFunc) error {
	// WalkDir is more efficient than Walk because it doesn't read file stats
	// for every file unless explicitly requested.
	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		// 1. Handle any errors walking to this file/dir (e.g. permissions)
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return err
		}

		// 2. Skip directories (we only care about files)
		if d.IsDir() {
			return nil
		}

		// 3. Check if the file has a .md extension (case-insensitive)
		if strings.ToLower(filepath.Ext(path)) == ".md" {
			// 4. Read the file content
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", path, err)
			}

			// 5. Extract just the filename (e.g., "note.md")
			filename := d.Name()

			// 6. Call the provided callback function
			err = processFn(string(contentBytes), filename, path)
			if err != nil {
				return fmt.Errorf("error processing file %s: %w", path, err)
			}
		}

		return nil
	})
}
