package search

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
)

// ChunkWithEmbedding is a text chunk paired with its vector embedding.
type ChunkWithEmbedding struct {
	Text      string            `json:"text"`
	Metadata  map[string]string `json:"metadata"`
	Embedding []float32         `json:"embedding"`
}

// SearchResult is a matching chunk with its similarity score and metadata.
type SearchResult struct {
	Text     string  `json:"text"`
	Score    float64 `json:"score"`
	Date     string  `json:"date"`
	Journal  string  `json:"journal"`
	Filename string  `json:"filename"`
	Path     string  `json:"path"`
}

// Searcher holds the embedding index for cosine-similarity search.
type Searcher struct {
	chunks []ChunkWithEmbedding
}

// LoadFromFile reads a JSON array of ChunkWithEmbedding from disk.
func LoadFromFile(path string) (*Searcher, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read corpus %s: %w", path, err)
	}
	var chunks []ChunkWithEmbedding
	if err := json.Unmarshal(data, &chunks); err != nil {
		return nil, fmt.Errorf("failed to parse corpus %s: %w", path, err)
	}
	return &Searcher{chunks: chunks}, nil
}

// Search finds the topK most similar chunks by cosine similarity to the query vector.
func (s *Searcher) Search(query []float32, topK int) []SearchResult {
	type scored struct {
		idx   int
		score float64
	}

	scores := make([]scored, len(s.chunks))
	for i, chunk := range s.chunks {
		scores[i] = scored{i, cosineSimilarity(query, chunk.Embedding)}
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	if topK > len(scores) {
		topK = len(scores)
	}

	results := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		chunk := s.chunks[scores[i].idx]
		results[i] = SearchResult{
			Text:     chunk.Text,
			Score:    scores[i].score,
			Date:     chunk.Metadata["date"],
			Journal:  chunk.Metadata["journal"],
			Filename: chunk.Metadata["filename"],
			Path:     chunk.Metadata["path"],
		}
	}
	return results
}

// ChunkCount returns the number of indexed chunks.
func (s *Searcher) ChunkCount() int {
	return len(s.chunks)
}

func cosineSimilarity(a, b []float32) float64 {
	var dot, magA, magB float64
	for i := range a {
		da := float64(a[i])
		db := float64(b[i])
		dot += da * db
		magA += da * da
		magB += db * db
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}
