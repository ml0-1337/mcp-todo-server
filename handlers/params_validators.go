package handlers

// isValidPriority validates priority values
func isValidPriority(p string) bool {
	return p == "high" || p == "medium" || p == "low"
}

// isValidTodoType validates todo type values
func isValidTodoType(t string) bool {
	return t == "feature" || t == "bug" || t == "refactor" || t == "research" || t == "multi-phase" || t == "phase" || t == "subtask"
}

// isValidFormat validates format values
func isValidFormat(f string) bool {
	return f == "full" || f == "summary" || f == "list"
}

// isValidOperation validates operation values
func isValidOperation(o string) bool {
	return o == "append" || o == "replace" || o == "prepend" || o == "toggle"
}