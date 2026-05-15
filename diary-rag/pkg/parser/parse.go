package parser

import (
	"strconv"
	"strings"
)

// ParseNoteToChunks splits a markdown note into Chunks, one per paragraph
// after extracting frontmatter metadata (journal type, date) and trimming the
// calendar-navigation footer.
func ParseNoteToChunks(content string, filename string, path string) []Chunk {
	// Normalize line endings before splitting
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")

	var chunks []Chunk
	var journalType string
	var date string
	var bodyLines []string
	captureContent := false

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		if trimmedLine == "---" {
			continue
		}

		if strings.HasPrefix(line, "journal:") {
			journalRaw := strings.TrimSpace(strings.TrimPrefix(line, "journal:"))
			words := strings.Fields(journalRaw)
			if len(words) > 0 {
				journalType = strings.ToLower(words[0])
			}
			continue
		}

		if strings.HasPrefix(line, "journal-date:") {
			date = strings.TrimSpace(strings.TrimPrefix(line, "journal-date:"))
			captureContent = true
			continue
		}

		// Stop at the calendar navigation block that some日记 files append
		if strings.HasPrefix(line, "```calendar-nav") {
			break
		}

		if captureContent {
			bodyLines = append(bodyLines, line)
		}
	}

	bodyText := strings.TrimSpace(strings.Join(bodyLines, "\n"))
	paragraphs := strings.Split(bodyText, "\n\n")

	var chunkIndex int
	for _, p := range paragraphs {
		cleanParagraph := strings.TrimSpace(p)
		if cleanParagraph != "" {
			chunks = append(chunks, Chunk{
				Text: cleanParagraph,
				Metadata: map[string]string{
					"journal":    journalType,
					"date":       date,
					"chunkIndex": strconv.Itoa(chunkIndex),
					"path":       path,
					"filename":   filename,
				},
			})
			chunkIndex++
		}
	}

	return chunks
}
