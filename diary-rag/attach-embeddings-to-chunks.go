package main

import "fmt"

type ChunkWithEmbedding struct {
	Chunk
	Emedding []float32 `json:"embedding"`
}

func AttachEmbeddingsToChunks(chunks []Chunk, embeddings [][]float32) ([]ChunkWithEmbedding, error) {
	if len(chunks) != len(embeddings) {
		return nil, fmt.Errorf("Chunks and embeddings count not the same")
	}
	chunksWithEmbeddings := make([]ChunkWithEmbedding, len(chunks))
	for index, chunk := range chunks {
		chunksWithEmbeddings[index] = ChunkWithEmbedding{chunk, embeddings[index]}
	}
	return chunksWithEmbeddings, nil
}
