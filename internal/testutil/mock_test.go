package testutil_test

import "testing"

// mockTB implements a minimal testing.TB interface for testing our test utilities
type mockTB struct {
	*testing.T
	errorCalled bool
	fatalCalled bool
}

func newMockTB(t *testing.T) *mockTB {
	return &mockTB{T: t}
}

func (m *mockTB) Errorf(format string, args ...interface{}) {
	m.errorCalled = true
	// Don't call the real Errorf to avoid failing the test
	m.Logf("Mock Errorf: "+format, args...)
}

func (m *mockTB) Fatalf(format string, args ...interface{}) {
	m.fatalCalled = true
	// Don't call the real Fatalf to avoid stopping the test
	m.Logf("Mock Fatalf: "+format, args...)
}

func (m *mockTB) Error(args ...interface{}) {
	m.errorCalled = true
	allArgs := append([]interface{}{"Mock Error:"}, args...)
	m.Log(allArgs...)
}

func (m *mockTB) Fatal(args ...interface{}) {
	m.fatalCalled = true
	allArgs := append([]interface{}{"Mock Fatal:"}, args...)
	m.Log(allArgs...)
}