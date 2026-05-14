package note

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRead_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	reader := NewReader()
	content, err := reader.Read(context.Background(), path)
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if content != "hello world" {
		t.Errorf("content = %q, want %q", content, "hello world")
	}
}

func TestRead_NotFound(t *testing.T) {
	reader := NewReader()
	_, err := reader.Read(context.Background(), "/nonexistent/path.md")
	if err == nil {
		t.Fatal("Read() expected error for nonexistent file")
	}
}
