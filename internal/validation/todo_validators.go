package validation

// Priority constants
const (
	PriorityHigh   = "high"
	PriorityMedium = "medium"
	PriorityLow    = "low"
)

// TodoType constants
const (
	TypeFeature    = "feature"
	TypeBug        = "bug"
	TypeRefactor   = "refactor"
	TypeResearch   = "research"
	TypeMultiPhase = "multi-phase"
	TypePhase      = "phase"
	TypeSubtask    = "subtask"
)

// Format constants
const (
	FormatFull    = "full"
	FormatSummary = "summary"
	FormatList    = "list"
)

// Operation constants
const (
	OperationAppend  = "append"
	OperationReplace = "replace"
	OperationPrepend = "prepend"
	OperationToggle  = "toggle"
)

// IsValidPriority validates priority values
func IsValidPriority(p string) bool {
	return p == PriorityHigh || p == PriorityMedium || p == PriorityLow
}

// IsValidTodoType validates todo type values
func IsValidTodoType(t string) bool {
	return t == TypeFeature || t == TypeBug || t == TypeRefactor || 
		t == TypeResearch || t == TypeMultiPhase || t == TypePhase || t == TypeSubtask
}

// IsValidFormat validates format values
func IsValidFormat(f string) bool {
	return f == FormatFull || f == FormatSummary || f == FormatList
}

// IsValidOperation validates operation values
func IsValidOperation(o string) bool {
	return o == OperationAppend || o == OperationReplace || o == OperationPrepend || o == OperationToggle
}

// GetValidPriorities returns all valid priority values
func GetValidPriorities() []string {
	return []string{PriorityHigh, PriorityMedium, PriorityLow}
}

// GetValidTodoTypes returns all valid todo type values
func GetValidTodoTypes() []string {
	return []string{TypeFeature, TypeBug, TypeRefactor, TypeResearch, TypeMultiPhase, TypePhase, TypeSubtask}
}

// GetValidFormats returns all valid format values
func GetValidFormats() []string {
	return []string{FormatFull, FormatSummary, FormatList}
}

// GetValidOperations returns all valid operation values
func GetValidOperations() []string {
	return []string{OperationAppend, OperationReplace, OperationPrepend, OperationToggle}
}