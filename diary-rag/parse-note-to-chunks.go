package main

import (
	"strconv"
	"strings"
)

// Chunk represents a single text segment and its metadata for RAG ingestion.
type Chunk struct {
	Text     string            `json:"text"`
	Metadata map[string]string `json:"metadata"`
}

// ParseNoteToChunks processes a single markdown note into an array of Chunks.
func ParseNoteToChunks(content string, filename string, path string) []Chunk {
	// Normalize line endings to avoid issues with Windows (\r\n) files
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	var chunks []Chunk
	var journalType string
	var date string
	var bodyLines []string
	captureContent := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// Skip YAML frontmatter dashes completely
		if trimmedLine == "---" {
			continue
		}

		// Extract the first word of the journal field and lowercase it
		if strings.HasPrefix(line, "journal:") {
			journalRaw := strings.TrimSpace(strings.TrimPrefix(line, "journal:"))
			words := strings.Fields(journalRaw)
			if len(words) > 0 {
				journalType = strings.ToLower(words[0])
			}
			continue
		}

		// Identify the date and start capturing the body afterwards
		if strings.HasPrefix(line, "journal-date:") {
			date = strings.TrimSpace(strings.TrimPrefix(line, "journal-date:"))
			captureContent = true
			continue
		}

		// Stop capturing when we hit the calendar navigation footer
		if strings.HasPrefix(line, "```calendar-nav") {
			break
		}

		// Collect the lines in between
		if captureContent {
			bodyLines = append(bodyLines, line)
		}
	}

	// Join the captured lines back together and trim surrounding whitespace
	bodyText := strings.TrimSpace(strings.Join(bodyLines, "\n"))

	// Split by double newlines to isolate individual paragraphs
	paragraphs := strings.Split(bodyText, "\n\n")

	var chunkIndex = 0
	for _, p := range paragraphs {
		// Trim each paragraph and ignore any empty ones created by extra line breaks
		cleanParagraph := strings.TrimSpace(p)
		if cleanParagraph != "" {
			chunks = append(chunks, Chunk{
				Text: cleanParagraph,
				Metadata: map[string]string{
					"journal":     journalType,
					"date":        date,
					"globalIndex": strconv.Itoa(GlobalIndex),
					"chunkIndex":  strconv.Itoa(chunkIndex),
					"path":        path,
					"filename":    filename,
				},
			})
			chunkIndex += 1
			GlobalIndex += 1
		}
	}

	return chunks
}
