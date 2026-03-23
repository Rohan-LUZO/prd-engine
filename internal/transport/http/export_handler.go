package http

import (
	"prd-engine/internal/export"
	"prd-engine/internal/http/response"
	"prd-engine/internal/service"

	"github.com/gin-gonic/gin"
)

// ExportHandler handles combined-document export (PDF or JSON).
type ExportHandler struct {
	moduleService *service.ModuleService
}

// NewExportHandler returns an export handler.
func NewExportHandler(moduleService *service.ModuleService) *ExportHandler {
	return &ExportHandler{moduleService: moduleService}
}

// Export combines all modules (latest versions), then returns PDF or JSON.
// GET /api/export?format=pdf  → application/pdf, attachment
// GET /api/export?format=json or no format → JSON with full document + audit
func (h *ExportHandler) Export(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	title := c.DefaultQuery("title", "PRD Documentation")

	doc, err := h.moduleService.GetCombinedDocument(title)
	if err != nil {
		response.Send(
			c,
			response.New(500).
				WithMessage(err.Error()).
				WithUserMessage(
					"Export failed",
					"Could not build the combined document. Please try again.",
				).
				WithTag("EXPORT"),
		)
		return
	}

	switch format {
	case "pdf":
		c.Header("Content-Type", "application/pdf")
		c.Header("Content-Disposition", `attachment; filename="prd-documentation.pdf"`)

		exportDoc := toExportDocument(doc)
		if err := export.BuildPDF(exportDoc, c.Writer); err != nil {
			response.Send(
				c,
				response.New(500).
					WithMessage(err.Error()).
					WithUserMessage(
						"PDF generation failed",
						"Could not generate the PDF. Please try again.",
					).
					WithTag("EXPORT_PDF"),
			)
			return
		}
	default:
		response.Send(
			c,
			response.New(200).
				WithMessage("Document exported successfully").
				WithData(doc),
		)
	}
}

// toExportDocument maps the service layer CombinedDocument to the export package's Document type
// (which has no dependency on the service package, so the PDF builder stays decoupled).
func toExportDocument(d *service.CombinedDocument) *export.Document {
	if d == nil {
		return nil
	}
	out := &export.Document{
		Title:       d.Title,
		GeneratedAt: d.GeneratedAt,
		Modules:     make([]export.Module, len(d.Modules)),
	}
	for i := range d.Modules {
		m := &d.Modules[i]
		out.Modules[i] = export.Module{
			ID:        m.ID,
			Title:     m.Title,
			Order:     m.Order,
			Version:   m.Version,
			Content:   m.Content,
			UpdatedBy: m.UpdatedBy,
			UpdatedAt: m.UpdatedAt,
			CreatedBy: m.CreatedBy,
			CreatedAt: m.CreatedAt,
		}
	}
	return out
}
