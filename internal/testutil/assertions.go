package testutil

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/user/mcp-todo-server/core"
	"github.com/user/mcp-todo-server/internal/errors"
)

// AssertTodoEqual checks if two todos are equal
func AssertTodoEqual(t testing.TB, expected, actual *core.Todo) {
	t.Helper()

	if expected == nil && actual == nil {
		return
	}

	if expected == nil || actual == nil {
		t.Errorf("Todo mismatch: expected %v, got %v", expected, actual)
		return
	}

	if expected.ID != actual.ID {
		t.Errorf("Todo ID mismatch: expected %s, got %s", expected.ID, actual.ID)
	}

	if expected.Task != actual.Task {
		t.Errorf("Todo Task mismatch: expected %s, got %s", expected.Task, actual.Task)
	}

	if expected.Priority != actual.Priority {
		t.Errorf("Todo Priority mismatch: expected %s, got %s", expected.Priority, actual.Priority)
	}

	if expected.Type != actual.Type {
		t.Errorf("Todo Type mismatch: expected %s, got %s", expected.Type, actual.Type)
	}

	if expected.Status != actual.Status {
		t.Errorf("Todo Status mismatch: expected %s, got %s", expected.Status, actual.Status)
	}

	if expected.ParentID != actual.ParentID {
		t.Errorf("Todo ParentID mismatch: expected %s, got %s", expected.ParentID, actual.ParentID)
	}
}

// AssertErrorIs checks if an error matches the expected error
func AssertErrorIs(t testing.TB, err, target error) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Errorf("Error mismatch: expected error to be %v, got %v", target, err)
	}
}

// AssertErrorAs checks if an error can be cast to the expected type
func AssertErrorAs(t testing.TB, err error, target interface{}) {
	t.Helper()

	if !errors.As(err, target) {
		t.Errorf("Error type mismatch: expected error to be %T, got %T", target, err)
	}
}

// AssertErrorCategory checks if an error has the expected category
func AssertErrorCategory(t testing.TB, err error, expected errors.ErrorCategory) {
	t.Helper()

	actual := errors.GetCategory(err)
	if actual != expected {
		t.Errorf("Error category mismatch: expected %v, got %v", expected, actual)
	}
}

// AssertErrorContains checks if an error message contains a substring
func AssertErrorContains(t testing.TB, err error, substr string) {
	t.Helper()

	if err == nil {
		t.Errorf("Expected error containing '%s', but got nil", substr)
		return
	}

	if !strings.Contains(err.Error(), substr) {
		t.Errorf("Error message '%s' does not contain '%s'", err.Error(), substr)
	}
}

// AssertNotNil fails if the value is nil
func AssertNotNil(t testing.TB, value interface{}, message string) {
	t.Helper()

	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		t.Errorf("%s: expected non-nil value", message)
	}
}

// AssertNil fails if the value is not nil
func AssertNil(t testing.TB, value interface{}, message string) {
	t.Helper()

	if value != nil && !(reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		t.Errorf("%s: expected nil, got %v", message, value)
	}
}

// AssertTrue fails if the condition is false
func AssertTrue(t testing.TB, condition bool, message string) {
	t.Helper()

	if !condition {
		t.Errorf("%s: expected true", message)
	}
}

// AssertFalse fails if the condition is true
func AssertFalse(t testing.TB, condition bool, message string) {
	t.Helper()

	if condition {
		t.Errorf("%s: expected false", message)
	}
}

// AssertLen checks if a slice/map/string has the expected length
func AssertLen(t testing.TB, value interface{}, expected int, message string) {
	t.Helper()

	v := reflect.ValueOf(value)
	var actual int

	switch v.Kind() {
	case reflect.Slice, reflect.Map, reflect.Array, reflect.String:
		actual = v.Len()
	default:
		t.Fatalf("%s: cannot get length of %T", message, value)
		return
	}

	if actual != expected {
		t.Errorf("%s: expected length %d, got %d", message, expected, actual)
	}
}

// AssertPanic checks if a function panics
func AssertPanic(t testing.TB, fn func(), message string) {
	t.Helper()

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("%s: expected panic but function completed normally", message)
		}
	}()

	fn()
}

// AssertNoPanic checks if a function does not panic
func AssertNoPanic(t testing.TB, fn func(), message string) {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("%s: unexpected panic: %v", message, r)
		}
	}()

	fn()
}

// AssertEventually checks if a condition becomes true within a timeout
func AssertEventually(t testing.TB, condition func() bool, timeout time.Duration, interval time.Duration, message string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}

	t.Errorf("%s: condition did not become true within %v", message, timeout)
}

// RequireNoError is like AssertNoError but stops test execution
func RequireNoError(t testing.TB, err error, message string) {
	t.Helper()

	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// RequireError is like AssertError but stops test execution
func RequireError(t testing.TB, err error, message string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// CompareSlices compares two slices for equality
func CompareSlices[T comparable](t testing.TB, expected, actual []T, message string) {
	t.Helper()

	if len(expected) != len(actual) {
		t.Errorf("%s: slice length mismatch: expected %d, got %d", message, len(expected), len(actual))
		return
	}

	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("%s: element mismatch at index %d: expected %v, got %v",
				message, i, expected[i], actual[i])
		}
	}
}
