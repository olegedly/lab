package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"diary-rag/pkg/embed"
	"diary-rag/pkg/parser"
	"diary-rag/pkg/search"
	"diary-rag/pkg/traverse"
)

// BuildEmbeddingIndex scans diary markdown files, computes embeddings, and writes the corpus to outputDir.
func BuildEmbeddingIndex(rootDir, embedURL, embedModel, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	var allChunks []parser.Chunk
	if err := traverse.TraverseAndProcessMarkdown(rootDir, func(content, filename, path string) error {
		relPath := strings.TrimPrefix(path, rootDir+"/")
		chunks := parser.ParseNoteToChunks(content, filename, relPath)
		allChunks = append(allChunks, chunks...)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to process markdown: %w", err)
	}

	for i := range allChunks {
		allChunks[i].Metadata["globalIndex"] = strconv.Itoa(i)
	}

	fmt.Printf("Processing %d chunks.\n", len(allChunks))

	if err := WriteJson(allChunks, filepath.Join(outputDir, "chunks.json")); err != nil {
		return err
	}

	justTexts := make([]string, len(allChunks))
	for i, chunk := range allChunks {
		justTexts[i] = chunk.Text
	}

	embeddings, err := embed.GetEmbeddings(justTexts, embedURL, embedModel)
	if err != nil {
		return fmt.Errorf("failed to get embeddings: %w", err)
	}

	if err := WriteJson(embeddings, filepath.Join(outputDir, "embeddings.json")); err != nil {
		return err
	}

	return WriteJson(attachEmbeddings(allChunks, embeddings), filepath.Join(outputDir, "chunksWithEmbeddings.json"))
}

func attachEmbeddings(chunks []parser.Chunk, embeddings [][]float32) []search.ChunkWithEmbedding {
	result := make([]search.ChunkWithEmbedding, len(chunks))
	for i := range chunks {
		result[i] = search.ChunkWithEmbedding{
			Text:      chunks[i].Text,
			Metadata:  chunks[i].Metadata,
			Embedding: embeddings[i],
		}
	}
	return result
}
