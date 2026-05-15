package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// WriteJson marshals data to indented JSON and writes it to a file.
func WriteJson(data any, outputFile string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	if err := os.WriteFile(outputFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", outputFile, err)
	}
	fmt.Printf("Saved %s\n", outputFile)
	return nil
}
