package note

import (
	"context"
	"os"
)

// Reader reads note files from the filesystem.
type Reader interface {
	Read(ctx context.Context, filepath string) (string, error)
}

// OSReader reads files from the local filesystem.
type OSReader struct{}

// NewReader creates an OSReader.
func NewReader() *OSReader {
	return &OSReader{}
}

// Read returns the contents of a file.
func (r *OSReader) Read(_ context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
