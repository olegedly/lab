package parser

// Chunk represents a single text segment and its metadata for RAG ingestion.
type Chunk struct {
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata"`
}
