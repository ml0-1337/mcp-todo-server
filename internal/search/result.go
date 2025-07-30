package search

// Result represents a search result
type Result struct {
	ID      string
	Task    string
	Score   float64
	Snippet string
}
