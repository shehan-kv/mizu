package email

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path"
	"path/filepath"
	"strings"
)

//go:embed templates/*.html
var embeddedTemplates embed.FS

// TemplateManager holds parsed templates mapped by name.
// Embedded templates are loaded first. If a custom templates
// directory is provided at startup, matching templates are
// replaced with the disk versions.
type TemplateManager struct {
	templates map[string]*template.Template
}

// NewTemplateManager loads embedded templates and optionally
// overrides them with templates found in customDir.
// Pass an empty string for customDir to use embedded templates only.
func NewTemplateManager(customDir string) (*TemplateManager, error) {
	m := &TemplateManager{
		templates: make(map[string]*template.Template),
	}

	if err := m.loadEmbedded(); err != nil {
		return nil, fmt.Errorf("email.TemplateManager: load embedded: %w", err)
	}

	if customDir != "" {
		if err := m.loadFromDisk(customDir); err != nil {
			return nil, fmt.Errorf("email.TemplateManager: load custom: %w", err)
		}
	}

	return m, nil
}

// Get returns the parsed template for the given name.
// Returns an error if the template is not found.
func (m *TemplateManager) Get(name string) (*template.Template, error) {
	t, ok := m.templates[name]
	if !ok {
		return nil, fmt.Errorf("email.TemplateManager: template not found: %s", name)
	}
	return t, nil
}

// loadEmbedded parses all embedded templates and stores them by name.
// The name is derived from the filename without the extension.
func (m *TemplateManager) loadEmbedded() error {
	entries, err := embeddedTemplates.ReadDir("templates")
	if err != nil {
		return fmt.Errorf("read embedded templates dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := templateName(entry.Name())
		path := path.Join("templates", entry.Name())

		content, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read embedded template %s: %w", path, err)
		}

		t, err := template.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse embedded template %s: %w", name, err)
		}

		m.templates[name] = t
	}

	return nil
}

// loadFromDisk reads .html files from dir and replaces any matching
// embedded templates. Files that don't match an embedded template name
// are added as new templates.
func (m *TemplateManager) loadFromDisk(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read custom templates dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".html" {
			continue
		}

		name := templateName(entry.Name())
		path := path.Join(dir, entry.Name())

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read custom template %s: %w", path, err)
		}

		t, err := template.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parse custom template %s: %w", name, err)
		}

		m.templates[name] = t
	}

	return nil
}

// templateName derives the template name from a filename
// by stripping the extension.
func templateName(filename string) string {
	return strings.TrimSuffix(filename, filepath.Ext(filename))
}
