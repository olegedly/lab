package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EmbeddingRequest represents the JSON payload sent to LM Studio.
type EmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// EmbeddingData represents a single embedding item inside the response data array.
type EmbeddingData struct {
	Embedding []float32 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbeddingResponse represents the JSON returned by LM Studio.
type EmbeddingResponse struct {
	Data []EmbeddingData `json:"data"`
}

// GetEmbeddings sends an array of texts to the LM Studio server and returns their vector embeddings.
func GetEmbeddings(texts []string, url string, model string) ([][]float32, error) {
	// Prepare the request payload
	reqBody := EmbeddingRequest{
		Model: model,
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close() // Always close the response body when done

	// Check for non-200 HTTP statuses
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the JSON response
	var resBody EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&resBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract embeddings and sort them based on their original index
	embeddings := make([][]float32, len(resBody.Data))
	for _, item := range resBody.Data {
		if item.Index >= 0 && item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}

	return embeddings, nil
}
