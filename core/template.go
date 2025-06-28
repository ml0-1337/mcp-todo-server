package core

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"gopkg.in/yaml.v3"
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
		return nil, fmt.Errorf("template not found: %s", name)
	}
	
	// Read template file
	content, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}
	
	// Parse template file
	contentStr := string(content)
	parts := strings.Split(contentStr, "---\n")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid template format: missing frontmatter delimiters")
	}
	
	// Parse YAML frontmatter
	var template Template
	err = yaml.Unmarshal([]byte(parts[1]), &template)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template frontmatter: %w", err)
	}
	
	// Set the content (everything after frontmatter)
	template.Content = parts[2]
	
	// Validate template name matches filename
	if template.Name != name {
		return nil, fmt.Errorf("template name mismatch: file '%s' contains template '%s'", name, template.Name)
	}
	
	return &template, nil
}

// ListTemplates returns a list of available template names
func (tm *TemplateManager) ListTemplates() ([]string, error) {
	// Read directory
	files, err := ioutil.ReadDir(tm.templatesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read templates directory: %w", err)
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