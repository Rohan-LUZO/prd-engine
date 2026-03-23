# PRD Engine – API documentation

## OpenAPI (Swagger) spec

- **Spec file:** [openapi.yaml](./openapi.yaml) – OpenAPI 3.0 definition for all APIs.

## Viewing the docs

1. **Start the server** from the project root:
   ```bash
   go run ./cmd/server
   ```
2. Open in a browser:
   - **Swagger UI:** http://localhost:8080/docs  
     Interactive API docs; use “Authorize” to set your Bearer token.
   - **Raw spec:** http://localhost:8080/swagger.yaml  

The docs and spec are **public** (no authentication). The API routes under `/api` and `/health` require a Bearer token.

## API overview

| Method | Path | Description |
|--------|------|-------------|
| GET | /health | Health check |
| POST | /api/modules | Create or update a module (new version) |
| GET | /api/modules/:id | Get latest version of a module |
| GET | /api/modules/:id/history | List version numbers for a module |
| GET | /api/modules/:id/headers | Get markdown headings (optional ?version=) |
| GET | /api/export | Export combined document (?format=json\|pdf, &title=...) |

All `/api/*` and `/health` requests require: `Authorization: Bearer <token>`.
