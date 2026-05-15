package main

import "fmt"

func main() {
	rootDir := "/home/morket/0mni/1. Life Admin/12. Logging/11. Diary"
	allChunks, err := GetAllChunks(rootDir)
	if err != nil {
		fmt.Printf("Failed to get chunks: %v\n", err)
	}
	WriteJson(allChunks, "chunks.json")
}
