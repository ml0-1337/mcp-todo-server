package handlers

import (
	"github.com/user/mcp-todo-server/internal/validation"
)

// isValidPriority validates priority values
func isValidPriority(p string) bool {
	return validation.IsValidPriority(p)
}

// isValidTodoType validates todo type values
func isValidTodoType(t string) bool {
	return validation.IsValidTodoType(t)
}

// isValidFormat validates format values
func isValidFormat(f string) bool {
	return validation.IsValidFormat(f)
}

// isValidOperation validates operation values
func isValidOperation(o string) bool {
	return validation.IsValidOperation(o)
}
