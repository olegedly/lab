package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"diary-rag/pkg/embed"
	"diary-rag/pkg/parser"
	"diary-rag/pkg/traverse"
)

func main() {
	rootDir := flag.String("dir", "/home/morket/0mni/1. Life Admin/12. Logging/11. Diary", "root directory of markdown diary files")
	embedURL := flag.String("embed-url", "http://192.168.1.5:5001/v1/embeddings", "embedding API URL")
	embedModel := flag.String("embed-model", "nomic-embed-text-v1.5", "embedding model name")
	outputDir := flag.String("output", "output", "output directory for JSON files")
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	var allChunks []parser.Chunk
	if err := traverse.TraverseAndProcessMarkdown(*rootDir, func(content, filename, path string) error {
		chunks := parser.ParseNoteToChunks(content, filename, path)
		allChunks = append(allChunks, chunks...)
		return nil
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to process markdown: %v\n", err)
		os.Exit(1)
	}

	// Assign global indices now that all chunks are collected
	for i := range allChunks {
		allChunks[i].Metadata["globalIndex"] = strconv.Itoa(i)
	}

	fmt.Printf("Successfully processed %d chunks.\n", len(allChunks))

	if err := WriteJson(allChunks, filepath.Join(*outputDir, "chunks.json")); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	justTexts := make([]string, len(allChunks))
	for i, chunk := range allChunks {
		justTexts[i] = chunk.Text
	}

	embeddings, err := embed.GetEmbeddings(justTexts, *embedURL, *embedModel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get embeddings: %v\n", err)
		os.Exit(1)
	}

	if err := WriteJson(embeddings, filepath.Join(*outputDir, "embeddings.json")); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	chunksWithEmbeddings, err := AttachEmbeddingsToChunks(allChunks, embeddings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to attach embeddings: %v\n", err)
		os.Exit(1)
	}

	if err := WriteJson(chunksWithEmbeddings, filepath.Join(*outputDir, "chunksWithEmbeddings.json")); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
