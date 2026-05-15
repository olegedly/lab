package main

import (
	"fmt"
)

func GetAllChunks(rootDir string) ([]Chunk, error) {
	var allChunks []Chunk

	processor := func(content string, filename string, path string) error {
		chunks := ParseNoteToChunks(content, filename, path)
		allChunks = append(allChunks, chunks...)
		return nil
	}

	err := TraverseAndProcessMarkdown(rootDir, processor)
	if err != nil {
		return allChunks, err
	}

	fmt.Printf("Successfully processed %d chunks from all markdown files.\n", len(allChunks))

	return allChunks, nil
}
