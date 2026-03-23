package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"prd-engine/internal/auth"
	"prd-engine/internal/domain"
	"prd-engine/internal/http/middleware"
	"prd-engine/internal/repository"
	"prd-engine/internal/service"

	"github.com/gin-gonic/gin"
)

const testToken = "test-token"

func testRouter(t *testing.T, basePath string) *gin.Engine {
	t.Helper()

	userStore := auth.NewMemoryUserStore(map[string]*auth.User{
		testToken: {Username: "testuser", FullName: "Test User", Token: testToken, Roles: []string{"admin"}},
	})

	repo := repository.NewFileModuleRepository(basePath)
	moduleService := service.NewModuleService(repo)
	moduleHandler := NewModuleHandler(moduleService)
	exportHandler := NewExportHandler(moduleService)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.Auth(userStore))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		modules := api.Group("/modules")
		{
			modules.POST("", moduleHandler.SaveModule)
			modules.GET("/:id", moduleHandler.GetLatest)
			modules.GET("/:id/history", moduleHandler.GetHistory)
			modules.GET("/:id/headers", moduleHandler.GetHeaders)
		}
		api.GET("/export", exportHandler.Export)
	}

	return r
}

func authHeader() string {
	return "Bearer " + testToken
}

func TestHealth(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("health: got status %d, want 200", rec.Code)
	}
	var body map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("health: decode body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("health: got status %q, want ok", body["status"])
	}
}

func TestAuthRejectsMissingToken(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/modules/any", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token: got status %d, want 401", rec.Code)
	}
}

func TestAuthRejectsInvalidToken(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/modules/any", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("invalid token: got status %d, want 401", rec.Code)
	}
}

func TestSaveModule(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	body := map[string]any{
		"id":       "test-module",
		"title":    "Test Module",
		"order":    1,
		"surfaces": []string{string(domain.SurfaceCustomerApp)},
		"content":  "# Hello\n\nSome content.",
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/modules", bytes.NewReader(raw))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("SaveModule: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		StatusCode int `json:"statusCode"`
		Data       struct {
			ID      string `json:"ID"`
			Title   string `json:"Title"`
			Version int    `json:"Version"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("SaveModule: decode: %v", err)
	}
	if resp.Data.ID != "test-module" {
		t.Errorf("SaveModule: got id %q", resp.Data.ID)
	}
	if resp.Data.Version != 1 {
		t.Errorf("SaveModule: got version %d, want 1", resp.Data.Version)
	}
}

func TestSaveModuleInvalidPayload(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	req := httptest.NewRequest(http.MethodPost, "/api/modules", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("SaveModule invalid: got status %d, want 400", rec.Code)
	}
}

func TestGetLatest(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	// Create module first
	createModule(t, r, dir, "get-latest-mod", "Get Latest Module", 1, "# H1\nContent")

	req := httptest.NewRequest(http.MethodGet, "/api/modules/get-latest-mod", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetLatest: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		Data struct {
			ID      string `json:"ID"`
			Title   string `json:"Title"`
			Content string `json:"Content"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("GetLatest: decode: %v", err)
	}
	if resp.Data.ID != "get-latest-mod" {
		t.Errorf("GetLatest: got id %q", resp.Data.ID)
	}
	if resp.Data.Title != "Get Latest Module" {
		t.Errorf("GetLatest: got title %q", resp.Data.Title)
	}
}

func TestGetLatestNotFound(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/modules/nonexistent", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("GetLatest not found: got status %d, want 404", rec.Code)
	}
}

func TestGetHistory(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	createModule(t, r, dir, "history-mod", "History Module", 1, "v1 content")
	createModule(t, r, dir, "history-mod", "History Module", 1, "v2 content")

	req := httptest.NewRequest(http.MethodGet, "/api/modules/history-mod/history", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetHistory: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		Data struct {
			ID       string `json:"id"`
			Versions []int  `json:"versions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("GetHistory: decode: %v", err)
	}
	if resp.Data.ID != "history-mod" {
		t.Errorf("GetHistory: got id %q", resp.Data.ID)
	}
	if len(resp.Data.Versions) < 2 {
		t.Errorf("GetHistory: got %d versions, want at least 2", len(resp.Data.Versions))
	}
}

func TestGetHeaders(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	createModule(t, r, dir, "headers-mod", "Headers Module", 1, "# First\n## Second\n\nBody")

	req := httptest.NewRequest(http.MethodGet, "/api/modules/headers-mod/headers", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetHeaders: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		Data struct {
			ID      string `json:"id"`
			Headers []struct {
				Level  int    `json:"level"`
				Text   string `json:"text"`
				Anchor string `json:"anchor"`
			} `json:"headers"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("GetHeaders: decode: %v", err)
	}
	if resp.Data.ID != "headers-mod" {
		t.Errorf("GetHeaders: got id %q", resp.Data.ID)
	}
	if len(resp.Data.Headers) < 2 {
		t.Errorf("GetHeaders: got %d headers, want at least 2", len(resp.Data.Headers))
	}
}

func TestGetHeadersByVersion(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	createModule(t, r, dir, "ver-mod", "Version Module", 1, "# Only H1")

	req := httptest.NewRequest(http.MethodGet, "/api/modules/ver-mod/headers?version=1", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GetHeadersByVersion: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		Data struct {
			Headers []struct {
				Text string `json:"text"`
			} `json:"headers"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("GetHeadersByVersion: decode: %v", err)
	}
	if len(resp.Data.Headers) == 0 {
		t.Error("GetHeadersByVersion: expected at least one header")
	}
}

func TestGetHeadersInvalidVersion(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	req := httptest.NewRequest(http.MethodGet, "/api/modules/some-id/headers?version=abc", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GetHeaders invalid version: got status %d, want 400", rec.Code)
	}
}

func TestExportJSON(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	createModule(t, r, dir, "export-mod", "Export Module", 1, "Content for export")

	req := httptest.NewRequest(http.MethodGet, "/api/export?format=json", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Export JSON: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		StatusCode int `json:"statusCode"`
		Data       struct {
			Title       string `json:"title"`
			GeneratedAt string `json:"generatedAt"`
			Modules     []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"modules"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Export JSON: decode: %v", err)
	}
	if len(resp.Data.Modules) < 1 {
		t.Errorf("Export JSON: got %d modules, want at least 1", len(resp.Data.Modules))
	}
	found := false
	for _, m := range resp.Data.Modules {
		if m.ID == "export-mod" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Export JSON: expected export-mod in modules")
	}
}

func TestExportPDF(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)

	createModule(t, r, dir, "pdf-mod", "PDF Module", 1, "# PDF Section\n\nText.")

	req := httptest.NewRequest(http.MethodGet, "/api/export?format=pdf", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Export PDF: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("Export PDF: Content-Type %q, want application/pdf", ct)
	}
	if cd := rec.Header().Get("Content-Disposition"); cd == "" {
		t.Error("Export PDF: missing Content-Disposition")
	}
	// PDF magic bytes
	body := rec.Body.Bytes()
	if len(body) < 5 || string(body[:4]) != "%PDF" {
		t.Errorf("Export PDF: body does not look like PDF (first bytes: %q)", body[:min(20, len(body))])
	}
}

// createModule creates a module via POST /api/modules (same router must be used).
func createModule(t *testing.T, r *gin.Engine, _ string, id, title string, order int, content string) {
	t.Helper()

	body := map[string]any{
		"id":       id,
		"title":    title,
		"order":    order,
		"surfaces": []string{string(domain.SurfaceCustomerApp)},
		"content":  content,
	}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/modules", bytes.NewReader(raw))
	req.Header.Set("Authorization", authHeader())
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("createModule %q: got status %d, body: %s", id, rec.Code, rec.Body.Bytes())
	}
}

func TestExportEmptyDocument(t *testing.T) {
	dir := t.TempDir()
	r := testRouter(t, dir)
	// No modules created

	req := httptest.NewRequest(http.MethodGet, "/api/export", nil)
	req.Header.Set("Authorization", authHeader())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Export empty: got status %d, body: %s", rec.Code, rec.Body.Bytes())
	}

	var resp struct {
		Data struct {
			Modules []any `json:"modules"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Export empty: decode: %v", err)
	}
	// Empty export returns empty or nil modules slice
	if len(resp.Data.Modules) != 0 {
		t.Errorf("Export empty: expected no modules, got %d", len(resp.Data.Modules))
	}
}
