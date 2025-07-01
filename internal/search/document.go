package search

import (
	"time"
)

// Document represents a todo in the search index
type Document struct {
	ID        string    `json:"id"`
	Task      string    `json:"task"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	Type      string    `json:"type"`
	Started   time.Time `json:"started"`
	Completed time.Time `json:"completed,omitempty"`
	Content   string    `json:"content"`
	Findings  string    `json:"findings"`
	Tests     string    `json:"tests"`
}