package omnisearch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "test query" {
			t.Errorf("query = %q, want %q", got, "test query")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Item{
			{Basename: "note a", Excerpt: "excerpt a", Score: 0.5, Path: "notes/a.md"},
			{Basename: "note b", Excerpt: "excerpt b", Score: 0.9, Path: "notes/b.md"},
		})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "/vault", http.DefaultClient)
	results, err := client.Search(context.Background(), "test query")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}

	if results[0].Basename != "note b" || results[0].Score != 0.9 {
		t.Error("expected results sorted by score descending")
	}
	if results[0].AbsPath != "/vault/notes/b.md" {
		t.Errorf("AbsPath = %q, want %q", results[0].AbsPath, "/vault/notes/b.md")
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "/vault", http.DefaultClient)
	results, err := client.Search(context.Background(), "nothing")
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("len = %d, want 0", len(results))
	}
}

func TestSearch_BadStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "/vault", http.DefaultClient)
	_, err := client.Search(context.Background(), "test")
	if err == nil {
		t.Fatal("Search() expected error for 500 status")
	}
}

func TestSearch_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "/vault", http.DefaultClient)
	_, err := client.Search(context.Background(), "test")
	if err == nil {
		t.Fatal("Search() expected error for invalid JSON")
	}
}

func TestSearch_NetworkError(t *testing.T) {
	client := &HTTPClient{
		baseURL:   "http://127.0.0.1:1",
		vaultPath: "/vault",
		httpDoer: &mockDoer{fn: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("connection refused")
		}},
	}
	_, err := client.Search(context.Background(), "test")
	if err == nil {
		t.Fatal("Search() expected error for network failure")
	}
}

type mockDoer struct {
	fn func(*http.Request) (*http.Response, error)
}

func (m *mockDoer) Do(req *http.Request) (*http.Response, error) {
	return m.fn(req)
}
