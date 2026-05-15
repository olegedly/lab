package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	var allChunks []Chunk
	rootDir := "/home/morket/0mni/1. Life Admin/12. Logging/11. Diary"

	processor := func(content string, filename string, path string) error {
		chunks := ParseNoteToChunks(content, filename, path)
		allChunks = append(allChunks, chunks...)
		return nil
	}

	err := TraverseAndProcessMarkdown(rootDir, processor)
	if err != nil {
		fmt.Printf("Traversal failed: %v\n", err)
		return
	}

	fmt.Printf("Successfully processed %d chunks from all markdown files.\n", len(allChunks))

	// Convert the slice of chunks into pretty-printed JSON
	jsonData, err := json.MarshalIndent(allChunks, "", "  ")
	if err != nil {
		fmt.Printf("Failed to encode chunks to JSON: %v\n", err)
		return
	}

	// Write the JSON data to a file in the current directory
	outputFile := "chunks.json"
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Failed to write JSON file to disk: %v\n", err)
		return
	}

	// Log success
	fmt.Printf("Successfully saved chunks to %s\n", outputFile)
}
