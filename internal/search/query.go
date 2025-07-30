package search

import (
	"strings"
)

// sanitizeQuery cleans up the search query to handle special characters safely
// This prevents regex injection attacks and ensures queries don't break the search engine
func sanitizeQuery(query string) string {
	// First, handle potential regex patterns - remove forward slashes that would indicate regex
	if strings.HasPrefix(query, "/") && strings.HasSuffix(query, "/") {
		// Don't execute as regex - treat as literal by removing slashes
		query = query[1 : len(query)-1]
	}

	// Remove null bytes and other control characters
	query = strings.ReplaceAll(query, "\x00", " ") // Replace null with space instead of removing
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\r", " ")
	query = strings.ReplaceAll(query, "\t", " ")

	// Replace special characters that break queries with spaces
	// Keep alphanumeric, spaces, and some safe punctuation
	var result []rune
	for _, r := range query {
		switch {
		case r >= 'a' && r <= 'z':
			result = append(result, r)
		case r >= 'A' && r <= 'Z':
			result = append(result, r)
		case r >= '0' && r <= '9':
			result = append(result, r)
		case r == ' ':
			result = append(result, r)
		case r == '-' || r == '_':
			// Keep hyphens and underscores as they're common in identifiers
			result = append(result, r)
		default:
			// Replace other special characters with space
			result = append(result, ' ')
		}
	}

	// Convert back to string and clean up extra spaces
	cleaned := string(result)

	// Replace multiple consecutive spaces with single space
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}

	// Trim leading and trailing spaces
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}
