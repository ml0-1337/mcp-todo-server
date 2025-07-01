package handlers

// TodoCreateParams represents parameters for todo_create
type TodoCreateParams struct {
	Task     string
	Priority string
	Type     string
	Template string
	ParentID string
}

// TodoCreateMultiParams represents parameters for todo_create_multi
type TodoCreateMultiParams struct {
	Parent   TodoCreateInfo   `json:"parent"`
	Children []TodoCreateInfo `json:"children"`
}

// TodoCreateInfo represents information for creating a todo in bulk operations
type TodoCreateInfo struct {
	Task     string `json:"task"`
	Priority string `json:"priority,omitempty"`
	Type     string `json:"type,omitempty"`
}

// TodoReadParams represents parameters for todo_read
type TodoReadParams struct {
	ID     string
	Filter TodoFilter
	Format string
}

// TodoFilter represents filter options for todo_read
type TodoFilter struct {
	Status   string
	Priority string
	Days     int
}

// TodoUpdateParams represents parameters for todo_update
type TodoUpdateParams struct {
	ID        string
	Section   string
	Operation string
	Content   string
	Metadata  TodoMetadata
}

// TodoMetadata represents metadata updates
type TodoMetadata struct {
	Status      string
	Priority    string
	CurrentTest string
}

// TodoSearchParams represents parameters for todo_search
type TodoSearchParams struct {
	Query   string
	Scope   []string
	Filters SearchFilters
	Limit   int
}

// SearchFilters represents search filter options
type SearchFilters struct {
	Status   string
	DateFrom string
	DateTo   string
}

// TodoArchiveParams represents parameters for todo_archive
type TodoArchiveParams struct {
	ID      string
	Quarter string
}