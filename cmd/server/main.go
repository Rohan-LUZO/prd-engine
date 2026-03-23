package main

import (
	"log"
	"path/filepath"

	"prd-engine/internal/auth"
	"prd-engine/internal/http/middleware"
	"prd-engine/internal/repository"
	"prd-engine/internal/service"
	httpTransport "prd-engine/internal/transport/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// ---------- config ----------
	baseDocsPath := "docs/modules"
	usersFilePath := "config/users.yaml"
	port := ":8080"

	// ---------- wiring ----------
	userStore, err := auth.NewFileUserStore(usersFilePath)
	if err != nil {
		log.Fatalf("failed to load users: %v", err)
	}

	repo := repository.NewFileModuleRepository(baseDocsPath)
	moduleService := service.NewModuleService(repo)
	moduleHandler := httpTransport.NewModuleHandler(moduleService)
	exportHandler := httpTransport.NewExportHandler(moduleService)

	// ---------- router ----------
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(middleware.Recovery())

	// ---- Public: Swagger / OpenAPI (no auth) ----
	r.GET("/swagger.yaml", func(c *gin.Context) {
		path := filepath.Join("docs", "openapi.yaml")
		c.Header("Content-Type", "application/x-yaml")
		c.File(path)
	})
	r.GET("/docs", serveSwaggerUI)

	// ---- Protected routes (Bearer token required) ----
	protected := r.Group("/")
	protected.Use(middleware.Auth(userStore))

	protected.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := protected.Group("/api")
	modules := api.Group("/modules")
	{
		modules.POST("", moduleHandler.SaveModule)
		modules.GET("", moduleHandler.SaveModule)
		modules.GET("/:id", moduleHandler.GetLatest)
		modules.GET("/:id/history", moduleHandler.GetHistory)
		modules.GET("/:id/headers", moduleHandler.GetHeaders)
	}
	api.GET("/export", exportHandler.Export)

	// ---- auth routes (future) ----
	// auth := api.Group("/auth")
	// {
	// 	auth.POST("/login", authHandler.Login)
	// 	auth.POST("/logout", authHandler.Logout)
	// }

	// ---------- start server ----------
	log.Println("🚀 PRD Engine running on", port)
	log.Println("   API docs: http://localhost" + port + "/docs")
	log.Println("   OpenAPI spec: http://localhost" + port + "/swagger.yaml")
	if err := r.Run(port); err != nil {
		log.Fatal(err)
	}
}

// serveSwaggerUI returns an HTML page that loads Swagger UI with the OpenAPI spec at /swagger.yaml.
func serveSwaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	// Use Swagger UI 5.x from CDN; spec URL is same-origin so /swagger.yaml works.
	c.Writer.WriteString(swaggerUIHTML)
}

const swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>PRD Engine API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      window.ui = SwaggerUIBundle({
        url: window.location.origin + "/swagger.yaml",
        dom_id: "#swagger-ui",
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ]
      });
    };
  </script>
</body>
</html>
`
