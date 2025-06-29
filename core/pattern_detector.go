package core

import (
	"regexp"
	"strings"
)

// PatternHint represents a suggestion based on detected patterns
type PatternHint struct {
	Pattern       string `json:"pattern"`       // The pattern type detected
	SuggestedType string `json:"suggestedType"` // Suggested todo type
	Message       string `json:"message"`       // Helpful message for the user
	Number        string `json:"number,omitempty"` // Extracted number if applicable
}

var (
	// Compile patterns once for performance
	phasePattern     = regexp.MustCompile(`(?i)^phase\s+(\d+(?:\.\d+)?)\b`)
	partPattern      = regexp.MustCompile(`(?i)^part\s+(\d+)(?:\s+of\s+\d+)?\b`)
	stepPattern      = regexp.MustCompile(`(?i)^step\s+(\d+)\b`)
	bracketPattern   = regexp.MustCompile(`^\[(\d+)\]\s+`)
	numberDotPattern = regexp.MustCompile(`^(\d+)\.\s+`)
	numberParenPattern = regexp.MustCompile(`^(\d+)\)\s+`)
)

// DetectPattern analyzes a todo title for common patterns
func DetectPattern(title string) *PatternHint {
	title = strings.TrimSpace(title)
	
	// Check for phase pattern
	if matches := phasePattern.FindStringSubmatch(title); matches != nil {
		return &PatternHint{
			Pattern:       "phase",
			SuggestedType: "phase",
			Message:       "This looks like a phase. Consider using type 'phase' with a parent_id.",
			Number:        matches[1],
		}
	}
	
	// Check for part pattern
	if matches := partPattern.FindStringSubmatch(title); matches != nil {
		return &PatternHint{
			Pattern:       "part",
			SuggestedType: "phase",
			Message:       "This looks like a multi-part task. Consider using type 'phase' with a parent_id.",
			Number:        matches[1],
		}
	}
	
	// Check for step pattern
	if matches := stepPattern.FindStringSubmatch(title); matches != nil {
		return &PatternHint{
			Pattern:       "step",
			SuggestedType: "subtask",
			Message:       "This looks like a step. Consider using type 'subtask' with a parent_id.",
			Number:        matches[1],
		}
	}
	
	// Check for numbered patterns
	if matches := bracketPattern.FindStringSubmatch(title); matches != nil {
		return &PatternHint{
			Pattern:       "numbered",
			SuggestedType: "subtask",
			Message:       "This looks like a numbered task. Consider using type 'subtask' with a parent_id.",
			Number:        matches[1],
		}
	}
	
	if matches := numberDotPattern.FindStringSubmatch(title); matches != nil {
		return &PatternHint{
			Pattern:       "numbered",
			SuggestedType: "subtask",
			Message:       "This looks like a numbered task. Consider using type 'subtask' with a parent_id.",
			Number:        matches[1],
		}
	}
	
	if matches := numberParenPattern.FindStringSubmatch(title); matches != nil {
		return &PatternHint{
			Pattern:       "numbered",
			SuggestedType: "subtask",
			Message:       "This looks like a numbered task. Consider using type 'subtask' with a parent_id.",
			Number:        matches[1],
		}
	}
	
	// No pattern detected
	return nil
}

// FindSimilarTodos finds todos with similar titles or patterns
func FindSimilarTodos(todos []*Todo, title string) []string {
	var similar []string
	
	// Extract common prefix (up to first colon or dash)
	prefix := extractPrefix(title)
	if prefix == "" {
		return similar
	}
	
	// Normalize the prefix for comparison
	normalizedPrefix := strings.ToLower(strings.TrimSpace(prefix))
	
	for _, todo := range todos {
		if todo.ID == "" {
			continue
		}
		
		todoPrefix := extractPrefix(todo.Task)
		if todoPrefix == "" {
			continue
		}
		
		// Normalize todo prefix
		normalizedTodoPrefix := strings.ToLower(strings.TrimSpace(todoPrefix))
		
		// Check if they have the same pattern type (Phase, Step, Part, etc.)
		if normalizedPrefix == normalizedTodoPrefix {
			similar = append(similar, todo.ID)
		}
	}
	
	return similar
}

// extractPrefix extracts a common prefix from a title
func extractPrefix(title string) string {
	// For phase/part/step patterns, extract just the pattern type
	if matches := phasePattern.FindStringSubmatch(title); matches != nil {
		return "Phase"
	}
	if matches := partPattern.FindStringSubmatch(title); matches != nil {
		return "Part"
	}
	if matches := stepPattern.FindStringSubmatch(title); matches != nil {
		return "Step"
	}
	
	// For numbered patterns, group them together
	if bracketPattern.MatchString(title) || numberDotPattern.MatchString(title) || numberParenPattern.MatchString(title) {
		return "Numbered"
	}
	
	// For other patterns, look for common prefixes before separators
	separators := []string{":", " - ", " — "}
	
	for _, sep := range separators {
		if idx := strings.Index(title, sep); idx > 0 {
			prefix := strings.TrimSpace(title[:idx])
			// Only return if it's a meaningful prefix (not too long)
			if len(prefix) < 30 {
				return prefix
			}
		}
	}
	
	return ""
}