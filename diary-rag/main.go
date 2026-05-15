package main

import "fmt"

func main() {
	rootDir := "/home/morket/0mni/1. Life Admin/12. Logging/11. Diary"
	allChunks, err := GetAllChunks(rootDir)
	if err != nil {
		fmt.Printf("Failed to get chunks: %v\n", err)
	}
	WriteJson(allChunks, "chunks.json")
	EmbeddingModel := "nomic-embed-text-v1.5"
	EmbeddingURL := "http://192.168.1.5:5001/v1/embeddings"
	var justTexts []string
	for _, chunk := range allChunks {
		justTexts = append(justTexts, chunk.Text)
	}
	embeddings, err := GetEmbeddings(justTexts, EmbeddingURL, EmbeddingModel)
	WriteJson(embeddings, "embeddings.json")
	chunksWithEmbeddings, err := AttachEmbeddingsToChunks(allChunks, embeddings)
	if err != nil {
		panic(err)
	}
	WriteJson(chunksWithEmbeddings, "chunksWithEmbeddings.json")
}
