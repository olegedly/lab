package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func WriteJson(data any, outputFile string) {
	// Convert the data into pretty-printed JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Failed to encode data to JSON: %v\n", err)
		return
	}

	// Write the JSON data to a file in the current directory
	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		fmt.Printf("Failed to write JSON file to disk: %v\n", err)
		return
	}

	// Log success
	fmt.Printf("Successfully saved data to %s\n", outputFile)
}
