package main

import (
	"fmt"

	"diary-rag/pkg/parser"
)

// ChunkWithEmbedding pairs a Chunk with its vector embedding.
type ChunkWithEmbedding struct {
	parser.Chunk
	Embedding []float32 `json:"embedding"`
}

// AttachEmbeddingsToChunks zips chunks and embeddings into a single slice.
func AttachEmbeddingsToChunks(chunks []parser.Chunk, embeddings [][]float32) ([]ChunkWithEmbedding, error) {
	if len(chunks) != len(embeddings) {
		return nil, fmt.Errorf("chunks and embeddings count mismatch: %d vs %d", len(chunks), len(embeddings))
	}
	result := make([]ChunkWithEmbedding, len(chunks))
	for i := range chunks {
		result[i] = ChunkWithEmbedding{chunks[i], embeddings[i]}
	}
	return result, nil
}
