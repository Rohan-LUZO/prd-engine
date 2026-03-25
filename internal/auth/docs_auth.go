package auth

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type DocsAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func LoadDocsAuth(path string) (*DocsAuth, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read docs auth file: %w", err)
	}
	var cfg DocsAuth
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse docs auth file: %w", err)
	}
	return &cfg, nil
}
