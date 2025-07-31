package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

// Test 1: GetEnv returns environment variable when set
func TestGetEnv_WithEnvironmentVariable(t *testing.T) {
	// Set up test environment variable
	testKey := "TEST_ENV_VAR"
	testValue := "test_value"
	os.Setenv(testKey, testValue)
	defer os.Unsetenv(testKey)

	// Call GetEnv
	result := GetEnv(testKey, "fallback")

	// Assert result equals test value
	if result != testValue {
		t.Errorf("GetEnv(%s, fallback) = %s; want %s", testKey, result, testValue)
	}
}

// Test 2: GetEnv returns fallback when variable not set
func TestGetEnv_WithoutEnvironmentVariable(t *testing.T) {
	// Ensure test key doesn't exist
	testKey := "NON_EXISTENT_ENV_VAR"
	os.Unsetenv(testKey)

	// Call GetEnv with fallback
	fallback := "default_value"
	result := GetEnv(testKey, fallback)

	// Assert result equals fallback
	if result != fallback {
		t.Errorf("GetEnv(%s, %s) = %s; want %s", testKey, fallback, result, fallback)
	}
}

// Test 3: FileExists returns true for existing file
func TestFileExists_WithExistingFile(t *testing.T) {
	// Create a temporary file
	tempFile, err := ioutil.TempFile("", "test-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test FileExists
	exists := FileExists(tempFile.Name())

	// Assert file exists
	if !exists {
		t.Errorf("FileExists(%s) = false; want true", tempFile.Name())
	}
}

// Test 4: FileExists returns false for non-existent file
func TestFileExists_WithNonExistentFile(t *testing.T) {
	// Use a path that definitely doesn't exist
	nonExistentPath := "/tmp/this-file-should-not-exist-12345.txt"

	// Ensure it doesn't exist
	os.Remove(nonExistentPath)

	// Test FileExists
	exists := FileExists(nonExistentPath)

	// Assert file doesn't exist
	if exists {
		t.Errorf("FileExists(%s) = true; want false", nonExistentPath)
	}
}

// Test 5: IsDirectory returns true for directory
func TestIsDirectory_WithDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test IsDirectory
	isDir := IsDirectory(tempDir)

	// Assert it's a directory
	if !isDir {
		t.Errorf("IsDirectory(%s) = false; want true", tempDir)
	}
}

// Test 6: IsDirectory returns false for file
func TestIsDirectory_WithFile(t *testing.T) {
	// Create a temporary file
	tempFile, err := ioutil.TempFile("", "test-file-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test IsDirectory
	isDir := IsDirectory(tempFile.Name())

	// Assert it's not a directory
	if isDir {
		t.Errorf("IsDirectory(%s) = true; want false", tempFile.Name())
	}
}

// Test 7: FindProjectRoot finds .claude directory
func TestFindProjectRoot_WithClaudeDirectory(t *testing.T) {
	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory
	claudeDir := filepath.Join(tempDir, ".claude")
	err = os.Mkdir(claudeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Create a subdirectory to start search from
	subDir := filepath.Join(tempDir, "src", "pkg")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Find project root from subdirectory
	root, err := FindProjectRoot(subDir)

	// Assert no error and correct root found
	if err != nil {
		t.Errorf("FindProjectRoot(%s) error = %v; want nil", subDir, err)
	}
	if root != tempDir {
		t.Errorf("FindProjectRoot(%s) = %s; want %s", subDir, root, tempDir)
	}
}

// Test 8: FindProjectRoot finds .git directory
func TestFindProjectRoot_WithGitDirectory(t *testing.T) {
	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create a subdirectory to start search from
	subDir := filepath.Join(tempDir, "src", "pkg")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Find project root from subdirectory
	root, err := FindProjectRoot(subDir)

	// Assert no error and correct root found
	if err != nil {
		t.Errorf("FindProjectRoot(%s) error = %v; want nil", subDir, err)
	}
	if root != tempDir {
		t.Errorf("FindProjectRoot(%s) = %s; want %s", subDir, root, tempDir)
	}
}

// Test 9: FindProjectRoot finds go.mod file
func TestFindProjectRoot_WithGoModFile(t *testing.T) {
	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create go.mod file
	goModPath := filepath.Join(tempDir, "go.mod")
	err = ioutil.WriteFile(goModPath, []byte("module test\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod file: %v", err)
	}

	// Create a subdirectory to start search from
	subDir := filepath.Join(tempDir, "cmd", "app")
	err = os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Find project root from subdirectory
	root, err := FindProjectRoot(subDir)

	// Assert no error and correct root found
	if err != nil {
		t.Errorf("FindProjectRoot(%s) error = %v; want nil", subDir, err)
	}
	if root != tempDir {
		t.Errorf("FindProjectRoot(%s) = %s; want %s", subDir, root, tempDir)
	}
}

// Test 10: FindProjectRoot returns error when no markers found
func TestFindProjectRoot_NoMarkersFound(t *testing.T) {
	// Create temp directory without any markers
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Find project root (should fail)
	_, err = FindProjectRoot(tempDir)

	// Assert error is returned
	if err == nil {
		t.Errorf("FindProjectRoot(%s) error = nil; want error", tempDir)
	}
}

// Test 11: FindProjectRoot handles filesystem root correctly
func TestFindProjectRoot_FilesystemRoot(t *testing.T) {
	// Start from root directory (no project markers exist there)
	_, err := FindProjectRoot("/")

	// Assert error is returned
	if err == nil {
		t.Errorf("FindProjectRoot(/) error = nil; want error")
	}
}

// Test 12: ResolveTodoPath uses project root .claude/todos
func TestResolveTodoPath_WithProjectRoot(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp project directory
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Resolve todo path
	todoPath, err := ResolveTodoPath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTodoPath() error = %v; want nil", err)
	}

	// Resolve symlinks for comparison (macOS /var vs /private/var)
	resolvedTodoPath, _ := filepath.EvalSymlinks(todoPath)
	expectedPath := filepath.Join(tempDir, ".claude", "todos")
	resolvedExpectedPath, _ := filepath.EvalSymlinks(expectedPath)

	if resolvedTodoPath != resolvedExpectedPath {
		t.Errorf("ResolveTodoPath() = %s; want %s", todoPath, expectedPath)
	}

	// Verify directory was created
	if !IsDirectory(todoPath) {
		t.Errorf("ResolveTodoPath() did not create directory at %s", todoPath)
	}
}

// Test 13: ResolveTodoPath creates directory if missing
func TestResolveTodoPath_CreatesDirectory(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp project directory
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory marker
	claudeDir := filepath.Join(tempDir, ".claude")
	err = os.Mkdir(claudeDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Ensure todos directory doesn't exist yet
	todosDir := filepath.Join(tempDir, ".claude", "todos")
	if FileExists(todosDir) {
		t.Fatalf("todos directory already exists")
	}

	// Resolve todo path
	todoPath, err := ResolveTodoPath()

	// Assert no error
	if err != nil {
		t.Errorf("ResolveTodoPath() error = %v; want nil", err)
	}

	// Verify directory was created
	if !IsDirectory(todoPath) {
		t.Errorf("ResolveTodoPath() did not create directory at %s", todoPath)
	}

	// Verify correct permissions
	info, err := os.Stat(todoPath)
	if err == nil {
		mode := info.Mode().Perm()
		if mode != 0750 {
			t.Errorf("Directory created with wrong permissions: %o; want 750", mode)
		}
	}
}

// Test 14: ResolveTodoPath respects CLAUDE_TODO_PATH override
func TestResolveTodoPath_WithEnvironmentOverride(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create custom path
	customPath := "/tmp/custom-todos"

	// Set environment variable
	os.Setenv("CLAUDE_TODO_PATH", customPath)
	defer os.Unsetenv("CLAUDE_TODO_PATH")

	// Resolve todo path
	todoPath, err := ResolveTodoPath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTodoPath() error = %v; want nil", err)
	}
	if todoPath != customPath {
		t.Errorf("ResolveTodoPath() = %s; want %s", todoPath, customPath)
	}
}

// Test 15: ResolveTodoPath uses current dir if no project root
func TestResolveTodoPath_NoProjectRoot(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp directory without project markers
	tempDir, err := ioutil.TempDir("", "no-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Resolve todo path
	todoPath, err := ResolveTodoPath()

	// Assert no error
	if err != nil {
		t.Errorf("ResolveTodoPath() error = %v; want nil", err)
	}

	// Should use current directory when no project root found
	expectedPath := filepath.Join(tempDir, ".claude", "todos")
	resolvedTodoPath, _ := filepath.EvalSymlinks(todoPath)
	resolvedExpectedPath, _ := filepath.EvalSymlinks(expectedPath)

	if resolvedTodoPath != resolvedExpectedPath {
		t.Errorf("ResolveTodoPath() = %s; want %s", todoPath, expectedPath)
	}
}

// Test 16: ResolveTemplatePath auto mode finds project templates
func TestResolveTemplatePath_AutoModeProjectTemplates(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp project directory
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create project templates directory
	templatesDir := filepath.Join(tempDir, ".claude", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Resolve template path (auto mode is default)
	templatePath, err := ResolveTemplatePath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTemplatePath() error = %v; want nil", err)
	}

	resolvedTemplatePath, _ := filepath.EvalSymlinks(templatePath)
	resolvedExpectedPath, _ := filepath.EvalSymlinks(templatesDir)

	if resolvedTemplatePath != resolvedExpectedPath {
		t.Errorf("ResolveTemplatePath() = %s; want %s", templatePath, templatesDir)
	}
}

// Test 17: ResolveTemplatePath auto mode falls back to user templates
func TestResolveTemplatePath_AutoModeFallbackToUser(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp project directory
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create temp home directory
	tempHome, err := ioutil.TempDir("", "home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create user templates directory (but not project templates)
	userTemplatesDir := filepath.Join(tempHome, ".claude", "templates")
	err = os.MkdirAll(userTemplatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user templates directory: %v", err)
	}

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Resolve template path (should fallback to user)
	templatePath, err := ResolveTemplatePath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTemplatePath() error = %v; want nil", err)
	}

	if templatePath != userTemplatesDir {
		t.Errorf("ResolveTemplatePath() = %s; want %s", templatePath, userTemplatesDir)
	}
}

// Test 18: ResolveTemplatePath project mode only checks project
func TestResolveTemplatePath_ProjectMode(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Set mode to project
	os.Setenv("CLAUDE_TEMPLATE_MODE", "project")
	defer os.Unsetenv("CLAUDE_TEMPLATE_MODE")

	// Create temp project directory
	tempDir, err := ioutil.TempDir("", "project-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .git directory to mark project root
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Create project templates directory
	templatesDir := filepath.Join(tempDir, ".claude", "templates")
	err = os.MkdirAll(templatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	// Change to project directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Resolve template path
	templatePath, err := ResolveTemplatePath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTemplatePath() error = %v; want nil", err)
	}

	resolvedTemplatePath, _ := filepath.EvalSymlinks(templatePath)
	resolvedExpectedPath, _ := filepath.EvalSymlinks(templatesDir)

	if resolvedTemplatePath != resolvedExpectedPath {
		t.Errorf("ResolveTemplatePath() = %s; want %s", templatePath, templatesDir)
	}
}

// Test 19: ResolveTemplatePath user mode only checks home
func TestResolveTemplatePath_UserMode(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Set mode to user
	os.Setenv("CLAUDE_TEMPLATE_MODE", "user")
	defer os.Unsetenv("CLAUDE_TEMPLATE_MODE")

	// Create temp home directory
	tempHome, err := ioutil.TempDir("", "home-*")
	if err != nil {
		t.Fatalf("Failed to create temp home directory: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable
	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", oldHome)

	// Create user templates directory
	userTemplatesDir := filepath.Join(tempHome, ".claude", "templates")
	err = os.MkdirAll(userTemplatesDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create user templates directory: %v", err)
	}

	// Resolve template path
	templatePath, err := ResolveTemplatePath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTemplatePath() error = %v; want nil", err)
	}

	if templatePath != userTemplatesDir {
		t.Errorf("ResolveTemplatePath() = %s; want %s", templatePath, userTemplatesDir)
	}
}

// Test 20: ResolveTemplatePath respects CLAUDE_TEMPLATE_PATH override
func TestResolveTemplatePath_EnvironmentOverride(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Set custom template path
	customPath := "/tmp/custom-templates"
	os.Setenv("CLAUDE_TEMPLATE_PATH", customPath)
	defer os.Unsetenv("CLAUDE_TEMPLATE_PATH")

	// Resolve template path
	templatePath, err := ResolveTemplatePath()

	// Assert no error and correct path
	if err != nil {
		t.Errorf("ResolveTemplatePath() error = %v; want nil", err)
	}

	if templatePath != customPath {
		t.Errorf("ResolveTemplatePath() = %s; want %s", templatePath, customPath)
	}
}

// Test 21: ResolveTodoPathFromWorkingDir creates todo directory
func TestResolveTodoPathFromWorkingDir(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp directory to use as working directory
	tempDir, err := ioutil.TempDir("", "working-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve todo path from working directory
	todoPath, err := ResolveTodoPathFromWorkingDir(tempDir)

	// Assert no error
	if err != nil {
		t.Errorf("ResolveTodoPathFromWorkingDir(%s) error = %v; want nil", tempDir, err)
	}

	// Verify correct path
	expectedPath := filepath.Join(tempDir, ".claude", "todos")
	if todoPath != expectedPath {
		t.Errorf("ResolveTodoPathFromWorkingDir(%s) = %s; want %s", tempDir, todoPath, expectedPath)
	}

	// Verify directory was created
	if !IsDirectory(todoPath) {
		t.Errorf("ResolveTodoPathFromWorkingDir() did not create directory at %s", todoPath)
	}

	// Verify correct permissions
	info, err := os.Stat(todoPath)
	if err == nil {
		mode := info.Mode().Perm()
		if mode != 0750 {
			t.Errorf("Directory created with wrong permissions: %o; want 750", mode)
		}
	}
}

// Test 22: ResolveTodoPathFromWorkingDir handles existing directory
func TestResolveTodoPathFromWorkingDir_ExistingDirectory(t *testing.T) {
	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp directory structure
	tempDir, err := ioutil.TempDir("", "working-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Pre-create the .claude/todos directory
	todoDir := filepath.Join(tempDir, ".claude", "todos")
	err = os.MkdirAll(todoDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create todo directory: %v", err)
	}

	// Add a file to verify directory is not recreated
	testFile := filepath.Join(todoDir, "test.md")
	err = ioutil.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Resolve todo path from working directory
	todoPath, err := ResolveTodoPathFromWorkingDir(tempDir)

	// Assert no error
	if err != nil {
		t.Errorf("ResolveTodoPathFromWorkingDir(%s) error = %v; want nil", tempDir, err)
	}

	// Verify correct path
	if todoPath != todoDir {
		t.Errorf("ResolveTodoPathFromWorkingDir(%s) = %s; want %s", tempDir, todoPath, todoDir)
	}

	// Verify test file still exists (directory wasn't recreated)
	if !FileExists(testFile) {
		t.Errorf("Test file was removed, directory may have been recreated")
	}
}

// Test 23: ResolveTodoPathFromWorkingDir handles permission errors
func TestResolveTodoPathFromWorkingDir_PermissionError(t *testing.T) {
	// Skip if running as root
	if os.Geteuid() == 0 {
		t.Skip("Test cannot run as root")
	}

	// Suppress log output during test
	log.SetOutput(ioutil.Discard)
	defer log.SetOutput(os.Stderr)

	// Create temp directory
	tempDir, err := ioutil.TempDir("", "working-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory with no write permissions
	claudeDir := filepath.Join(tempDir, ".claude")
	err = os.Mkdir(claudeDir, 0555) // Read and execute only
	if err != nil {
		t.Fatalf("Failed to create .claude directory: %v", err)
	}

	// Try to resolve todo path (should fail due to permissions)
	_, err = ResolveTodoPathFromWorkingDir(tempDir)

	// Assert error is returned
	if err == nil {
		t.Errorf("ResolveTodoPathFromWorkingDir(%s) error = nil; want permission error", tempDir)
	}
}
