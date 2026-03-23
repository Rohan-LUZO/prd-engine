package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"prd-engine/internal/domain"
	parser "prd-engine/internal/storage/file"
	"sort"
	"strconv"
	"strings"
)

// FileModuleRepository stores each module under BasePath as a directory of versioned markdown files.
// Layout: BasePath/<moduleID>/v1.md, v2.md, ... (one file per version; never overwrite).
type FileModuleRepository struct {
	BasePath string // e.g. "docs/modules"
}

func NewFileModuleRepository(basePath string) *FileModuleRepository {
	return &FileModuleRepository{
		BasePath: basePath,
	}
}

// GetLatest loads the highest version number for the module, then reads that version's file.
func (r *FileModuleRepository) GetLatest(moduleID string) (*domain.Module, error) {
	versions, err := r.ListVersions(moduleID)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found for module %s", moduleID)
	}
	// ListVersions returns sorted ascending, so last element is latest
	latestVersion := versions[len(versions)-1]
	return r.GetByVersion(moduleID, latestVersion)
}

// GetByVersion returns a specific version of a module
func (r *FileModuleRepository) GetByVersion(moduleID string, version int) (*domain.Module, error) {
	filePath := r.moduleFilePath(moduleID, version)

	module, err := parser.ParseModuleFromFile(filePath)
	if err != nil {
		return nil, err
	}

	return module, nil
}

// ListModuleIDs returns all module IDs by listing direct subdirectories of BasePath.
// Each subdirectory name is treated as a module ID (e.g. "checkout-flow", "onboarding").
func (r *FileModuleRepository) ListModuleIDs() ([]string, error) {
	entries, err := os.ReadDir(r.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if name == "" || name == "." || name == ".." {
			continue
		}
		ids = append(ids, name)
	}
	return ids, nil
}

// ListVersions scans the module directory for files named v<N>.md and returns
// the version numbers sorted ascending (e.g. [1, 2, 3]). Non-matching files are ignored.
func (r *FileModuleRepository) ListVersions(moduleID string) ([]int, error) {
	dirPath := r.moduleDirPath(moduleID)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []int{}, nil
		}
		return nil, err
	}
	var versions []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name() // expect v1.md, v2.md, ...
		if !strings.HasPrefix(name, "v") || !strings.HasSuffix(name, ".md") {
			continue
		}
		versionStr := strings.TrimSuffix(strings.TrimPrefix(name, "v"), ".md")
		version, err := strconv.Atoi(versionStr)
		if err != nil {
			continue
		}
		versions = append(versions, version)
	}
	sort.Ints(versions)
	return versions, nil
}

// SaveNewVersion writes a NEW version file (never overwrites)
func (r *FileModuleRepository) SaveNewVersion(module *domain.Module) error {
	dirPath := r.moduleDirPath(module.ID)

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return fmt.Errorf("create module dir: %w", err)
	}

	filePath := r.moduleFilePath(module.ID, module.Version)

	if err := parser.WriteModuleToFile(filePath, module); err != nil {
		return err
	}

	return nil
}

// ---------- path helpers (internal) ----------

// moduleDirPath returns BasePath/moduleID (e.g. docs/modules/checkout-flow).
func (r *FileModuleRepository) moduleDirPath(moduleID string) string {
	return filepath.Join(r.BasePath, moduleID)
}

// moduleFilePath returns BasePath/moduleID/v<N>.md for a given version.
func (r *FileModuleRepository) moduleFilePath(moduleID string, version int) string {
	filename := fmt.Sprintf("v%d.md", version)
	return filepath.Join(r.moduleDirPath(moduleID), filename)
}
