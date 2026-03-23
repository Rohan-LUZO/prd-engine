package auth

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// User represents a single authenticated user loaded from file.
type User struct {
	Username string   `yaml:"username" json:"username"`
	FullName string   `yaml:"full_name" json:"fullName"`
	Token    string   `yaml:"token" json:"-"`
	Roles    []string `yaml:"roles" json:"roles"`
}

// UserStore defines how authentication data is retrieved.
type UserStore interface {
	// FindByToken returns a user for the given token, or nil if not found.
	FindByToken(token string) (*User, error)
}

// FileUserStore loads users from a YAML file at startup and keeps them in memory.
//
// Example users.yaml:
//
//   - username: alice
//     full_name: Alice Smith
//     token: alice-secret-token
//     roles: [admin]
//
//   - username: bob
//     full_name: Bob Jones
//     token: bob-token
//     roles: [editor]
type FileUserStore struct {
	usersByToken map[string]*User
}

// NewFileUserStore loads users from the provided YAML file path.
func NewFileUserStore(path string) (*FileUserStore, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read users file: %w", err)
	}

	var users []*User
	if err := yaml.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("parse users file: %w", err)
	}

	usersByToken := make(map[string]*User, len(users))
	for _, u := range users {
		if u == nil || u.Token == "" {
			continue
		}
		usersByToken[u.Token] = u
	}

	return &FileUserStore{
		usersByToken: usersByToken,
	}, nil
}

func (s *FileUserStore) FindByToken(token string) (*User, error) {
	if token == "" {
		return nil, nil
	}

	if u, ok := s.usersByToken[token]; ok {
		return u, nil
	}

	return nil, nil
}

