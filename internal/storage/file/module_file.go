package parser

import (
	"errors"
	"fmt"
	"os"
	"prd-engine/internal/domain"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	ErrInvalidFrontmatter = errors.New("invalid frontmatter")
)

// frontmatter is a FILE concern, not a domain concern
type frontmatter struct {
	ID       string                  `yaml:"id"`
	Version  int                     `yaml:"version"`
	Title    string                  `yaml:"title"`
	Order    int                     `yaml:"order"`
	Surfaces []domain.ProductSurface `yaml:"surfaces"`

	CreatedBy string    `yaml:"created_by"`
	CreatedAt time.Time `yaml:"created_at"`
	UpdatedBy string    `yaml:"updated_by"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

// ParseModuleFromFile reads a versioned markdown file and converts it into a domain.Module.
// File format: "---\n" + YAML frontmatter (id, version, title, order, surfaces, created_by, etc.) + "---\n\n" + markdown body.
// SplitN(content, "---", 3) gives [beforeFirst, yamlBlock, body]; we parse the middle and use the rest as content.
func ParseModuleFromFile(filePath string) (*domain.Module, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return nil, ErrInvalidFrontmatter
	}
	// Exactly two "---" delimiters: first after opening ---, second before body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, ErrInvalidFrontmatter
	}
	var fm frontmatter
	if err := yaml.Unmarshal([]byte(parts[1]), &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	module := &domain.Module{
		ID:       fm.ID,
		Version:  fm.Version,
		Title:    fm.Title,
		Order:    fm.Order,
		Surfaces: fm.Surfaces,
		Content:  strings.TrimSpace(parts[2]),

		CreatedBy: fm.CreatedBy,
		CreatedAt: fm.CreatedAt,
		UpdatedBy: fm.UpdatedBy,
		UpdatedAt: fm.UpdatedAt,
	}

	return module, nil
}

// WriteModuleToFile writes a new version file in the same format as ParseModuleFromFile expects.
// It refuses to overwrite: if the file already exists, returns an error (versioning safety).
func WriteModuleToFile(filePath string, module *domain.Module) error {
	fm := frontmatter{
		ID:        module.ID,
		Version:   module.Version,
		Title:     module.Title,
		Order:     module.Order,
		Surfaces:  module.Surfaces,
		CreatedBy: module.CreatedBy,
		CreatedAt: module.CreatedAt,
		UpdatedBy: module.UpdatedBy,
		UpdatedAt: module.UpdatedAt,
	}

	yamlBytes, err := yaml.Marshal(&fm)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}

	var builder strings.Builder
	builder.WriteString("---\n")
	builder.Write(yamlBytes)
	builder.WriteString("---\n\n")
	builder.WriteString(strings.TrimSpace(module.Content))
	builder.WriteString("\n")

	// Versioning safety: never overwrite an existing version file
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file already exists: %s", filePath)
	}

	if err := os.WriteFile(filePath, []byte(builder.String()), 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
