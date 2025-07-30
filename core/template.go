package core

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	interrors "github.com/user/mcp-todo-server/internal/errors"
)

// Template represents a todo template with metadata and content
type Template struct {
	Name        string   `yaml:"template_name"`
	Description string   `yaml:"description"`
	Variables   []string `yaml:"variables"`
	Content     string   // The template content after frontmatter
}

// TemplateManager handles loading and managing todo templates
type TemplateManager struct {
	templatesDir string
}

// NewTemplateManager creates a new template manager
func NewTemplateManager(templatesDir string) *TemplateManager {
	return &TemplateManager{
		templatesDir: templatesDir,
	}
}

// LoadTemplate loads a template by name from the templates directory
func (tm *TemplateManager) LoadTemplate(name string) (*Template, error) {
	// Construct template file path
	templatePath := filepath.Join(tm.templatesDir, name+".md")

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		return nil, interrors.NewNotFoundError("template", name)
	}

	// Read template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to read template")
	}

	// Parse template file
	contentStr := string(content)
	parts := strings.Split(contentStr, "---\n")
	if len(parts) < 3 {
		return nil, interrors.NewValidationError("template", name, "invalid template format: missing frontmatter delimiters")
	}

	// Parse YAML frontmatter
	var template Template
	err = yaml.Unmarshal([]byte(parts[1]), &template)
	if err != nil {
		return nil, interrors.Wrap(err, "failed to parse template frontmatter")
	}

	// Set the content (everything after frontmatter)
	template.Content = parts[2]

	// Validate template name matches filename
	if template.Name != name {
		return nil, interrors.NewValidationError("template_name", template.Name, fmt.Sprintf("template name mismatch: file '%s' contains template '%s'", name, template.Name))
	}

	return &template, nil
}

// ListTemplates returns a list of available template names
func (tm *TemplateManager) ListTemplates() ([]string, error) {
	// Read directory
	files, err := ioutil.ReadDir(tm.templatesDir)
	if err != nil {
		if os.IsNotExist(err) {
			// No templates directory exists
			return []string{}, nil
		}
		return nil, interrors.Wrap(err, "failed to read templates directory")
	}

	// Collect template names
	var templates []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			// Remove .md extension
			templateName := strings.TrimSuffix(file.Name(), ".md")
			templates = append(templates, templateName)
		}
	}

	return templates, nil
}

// ExecuteTemplate processes a template with the given variables
// CreateFromTemplate creates a new todo from a template
func (tm *TemplateManager) CreateFromTemplate(templateName, task, priority, todoType string) (*Todo, error) {
	// Load template
	tmpl, err := tm.LoadTemplate(templateName)
	if err != nil {
		return nil, err
	}

	// Prepare variables
	vars := map[string]interface{}{
		"task":     task,
		"priority": priority,
		"type":     todoType,
	}

	// Execute template
	_, err = tm.ExecuteTemplate(tmpl, vars)
	if err != nil {
		return nil, err
	}

	// Create todo with template content
	todo := &Todo{
		ID:       generateBaseID(task),
		Task:     task,
		Started:  time.Now(),
		Status:   "in_progress",
		Priority: priority,
		Type:     todoType,
	}

	// Note: In a real implementation, we'd write the todo with the template content
	// For now, just return the todo object
	return todo, nil
}

func (tm *TemplateManager) ExecuteTemplate(tmpl *Template, vars map[string]interface{}) (string, error) {
	// Parse the template content
	t, err := template.New(tmpl.Name).Parse(tmpl.Content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute the template with variables
	var buf bytes.Buffer
	err = t.Execute(&buf, vars)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
