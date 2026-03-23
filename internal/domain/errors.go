package domain

import "errors"

var (
	// Module lifecycle
	ErrModuleNotFound      = errors.New("module not found")
	ErrModuleAlreadyExists = errors.New("module already exists")

	// Validation / rules
	ErrDuplicateOrder  = errors.New("module order already exists")
	ErrInvalidModuleID = errors.New("invalid module id")
	ErrInvalidSurface  = errors.New("invalid product surface")
	ErrInvalidOrder    = errors.New("invalid module order")

	// Versioning
	ErrNoVersionsFound = errors.New("no module versions found")
)
