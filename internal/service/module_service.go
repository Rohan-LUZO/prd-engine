package service

import (
	"sort"

	"prd-engine/internal/domain"
	"prd-engine/internal/repository"
	"strings"
	"time"
)

// ModuleService implements the business logic for PRD modules: create/update (with versioning),
// fetch latest, history, headers extraction, and combined-document export.
type ModuleService struct {
	repo *repository.FileModuleRepository
}

func NewModuleService(repo *repository.FileModuleRepository) *ModuleService {
	return &ModuleService{
		repo: repo,
	}
}

// SaveModuleInput is the input for saving (creating or updating) a module.
type SaveModuleInput struct {
	ID       string
	Title    string
	Order    int
	Surfaces []domain.ProductSurface
	Content  string

	UpdatedBy string
}

// SaveModule creates a new module (version 1) or saves a new version of an existing module.
//   - If GetLatest fails (e.g. module does not exist), we treat it as a new module: Version=1, CreatedBy/At set.
//   - If GetLatest succeeds, we bump Version to latest.Version+1 and preserve original CreatedBy/CreatedAt;
//     UpdatedBy/UpdatedAt are always set to the current user and now.
//
// After building the module, we validate order and persist via SaveNewVersion (writes a new file, never overwrites).
func (s *ModuleService) SaveModule(input SaveModuleInput) (*domain.Module, error) {
	now := time.Now()

	latest, err := s.repo.GetLatest(input.ID)
	isNewModule := false
	if err != nil {
		isNewModule = true
	}

	var module domain.Module
	if isNewModule {
		module = domain.Module{
			ID:       input.ID,
			Version:  1,
			Title:    input.Title,
			Order:    input.Order,
			Surfaces: input.Surfaces,
			Content:  input.Content,

			CreatedBy: input.UpdatedBy,
			CreatedAt: now,
			UpdatedBy: input.UpdatedBy,
			UpdatedAt: now,
		}
	} else {
		// New version: increment version, keep original creator/time, set current updater/time
		module = domain.Module{
			ID:       latest.ID,
			Version:  latest.Version + 1,
			Title:    input.Title,
			Order:    input.Order,
			Surfaces: input.Surfaces,
			Content:  input.Content,

			CreatedBy: latest.CreatedBy,
			CreatedAt: latest.CreatedAt,
			UpdatedBy: input.UpdatedBy,
			UpdatedAt: now,
		}
	}

	if err := s.validateOrder(module.ID, module.Order); err != nil {
		return nil, err
	}

	if err := s.repo.SaveNewVersion(&module); err != nil {
		return nil, err
	}

	return &module, nil
}

// List modules list all the modules that are available
func (s *ModuleService) ListModules() ([]string, error) {
	return s.repo.ListModuleIDs()
}

func (s *ModuleService) GetLatest(moduleID string) (*domain.Module, error) {
	return s.repo.GetLatest(moduleID)
}

func (s *ModuleService) GetHistory(moduleID string) ([]int, error) {
	return s.repo.ListVersions(moduleID)
}

// ModuleHeader represents a single markdown heading inside a module's content.
// This is a read-model primarily used by the HTTP/API layer.
type ModuleHeader struct {
	Level  int    `json:"level"`  // 1 = H1, 2 = H2, etc.
	Text   string `json:"text"`   // visible text of the heading
	Anchor string `json:"anchor"` // slug that frontend can use for in-page navigation
}

// GetHeaders returns the list of markdown headers for the latest version
// of the given module. This is useful for frontends that want to render
// a navigation sidebar or table-of-contents.
func (s *ModuleService) GetHeaders(moduleID string) ([]ModuleHeader, error) {
	module, err := s.repo.GetLatest(moduleID)
	if err != nil {
		return nil, err
	}

	return extractModuleHeaders(module.Content), nil
}

// GetHeadersByVersion returns the list of markdown headers for a specific
// version of the given module.
func (s *ModuleService) GetHeadersByVersion(moduleID string, version int) ([]ModuleHeader, error) {
	module, err := s.repo.GetByVersion(moduleID, version)
	if err != nil {
		return nil, err
	}

	return extractModuleHeaders(module.Content), nil
}

// extractModuleHeaders parses the markdown content and returns a flat list
// of headings (H1–H3). You can expand this later if you need deeper levels.
func extractModuleHeaders(content string) []ModuleHeader {
	lines := strings.Split(content, "\n")
	var headers []ModuleHeader

	for _, rawLine := range lines {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}

		// Count leading '#' characters (markdown headings)
		level := 0
		for i := 0; i < len(line) && line[i] == '#'; i++ {
			level++
		}

		// Only keep H1–H3 and require a space after the hashes (`## Heading`)
		if level == 0 || level > 3 {
			continue
		}
		if len(line) <= level || line[level] != ' ' {
			continue
		}

		text := strings.TrimSpace(line[level:])
		if text == "" {
			continue
		}

		anchor := SlugifyHeading(text)

		headers = append(headers, ModuleHeader{
			Level:  level,
			Text:   text,
			Anchor: anchor,
		})
	}

	return headers
}

// slugifyHeading converts a heading text into a URL-friendly anchor.
// Example: "My Heading 1!" -> "my-heading-1"
func SlugifyHeading(text string) string {
	// Lowercase
	s := strings.ToLower(strings.TrimSpace(text))

	// Replace spaces and consecutive whitespace with single dashes
	s = strings.Join(strings.Fields(s), "-")

	// Remove any character that's not a-z, 0-9, or '-'
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}

	// Avoid empty anchor
	if b.Len() == 0 {
		return "section"
	}

	return b.String()
}

// CombinedModuleEntry is one module's latest version, for export or audit.
type CombinedModuleEntry struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Order     int       `json:"order"`
	Version   int       `json:"version"`
	Content   string    `json:"content"`
	UpdatedBy string    `json:"updatedBy"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
}

// CombinedDocument holds all modules (latest versions only), sorted by Order,
// for export as PDF or JSON. Includes versioning and who-updated-what for audit.
type CombinedDocument struct {
	Title       string                `json:"title"`
	GeneratedAt time.Time             `json:"generatedAt"`
	Modules     []CombinedModuleEntry `json:"modules"`
}

// GetCombinedDocument returns all modules' latest versions, sorted by Order,
// so the app can export one combined document (e.g. PDF) with full audit trail.
// Modules that fail to load (missing files, parse errors) are skipped so export
// can still succeed for the rest.
func (s *ModuleService) GetCombinedDocument(title string) (*CombinedDocument, error) {
	ids, err := s.repo.ListModuleIDs()
	if err != nil {
		return nil, err
	}

	var entries []CombinedModuleEntry
	for _, id := range ids {
		m, err := s.repo.GetLatest(id)
		if err != nil {
			// Skip broken or empty modules so we still return a valid document
			continue
		}
		entries = append(entries, CombinedModuleEntry{
			ID:        m.ID,
			Title:     m.Title,
			Order:     m.Order,
			Version:   m.Version,
			Content:   m.Content,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			CreatedAt: m.CreatedAt,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Order < entries[j].Order
	})

	docTitle := title
	if docTitle == "" {
		docTitle = "PRD Documentation"
	}

	return &CombinedDocument{
		Title:       docTitle,
		GeneratedAt: time.Now(),
		Modules:     entries,
	}, nil
}

// validateOrder enforces that order is a positive integer. The currentModuleID is reserved
// for future use (e.g. uniqueness checks across modules).
func (s *ModuleService) validateOrder(currentModuleID string, order int) error {
	if order <= 0 {
		return domain.ErrInvalidOrder
	}
	return nil
}
