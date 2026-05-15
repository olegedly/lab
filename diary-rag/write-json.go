package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteJson(chunks []Chunk, outputFile string) {
	// Convert the slice of chunks into pretty-printed JSON
	jsonData, err := json.MarshalIndent(chunks, "", "  ")
	if err != nil {
		fmt.Printf("Failed to encode chunks to JSON: %v\n", err)
		return
	}

	// Write the JSON data to a file in the current directory
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Failed to write JSON file to disk: %v\n", err)
		return
	}

	// Log success
	fmt.Printf("Successfully saved chunks to %s\n", outputFile)
}
