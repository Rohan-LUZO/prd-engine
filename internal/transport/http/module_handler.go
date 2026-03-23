package http

import (
	"strconv"

	"prd-engine/internal/domain"
	"prd-engine/internal/http/response"
	"prd-engine/internal/service"

	"github.com/gin-gonic/gin"
)

type ModuleHandler struct {
	service *service.ModuleService
}

func NewModuleHandler(service *service.ModuleService) *ModuleHandler {
	return &ModuleHandler{
		service: service,
	}
}

// ----------- DTOs (HTTP only) -----------
// These structs match the JSON request/response shapes; they are not part of the domain.

type saveModuleRequest struct {
	ID       string                  `json:"id"`
	Title    string                  `json:"title" binding:"required"`
	Order    int                     `json:"order" binding:"required"`
	Surfaces []domain.ProductSurface `json:"surfaces" binding:"required"`
	Content  string                  `json:"content" binding:"required"`
}

type moduleResponse struct {
	*domain.Module
}

type moduleHeadersResponse struct {
	ID      string                 `json:"id"`
	Headers []service.ModuleHeader `json:"headers"`
}

// ----------- Handlers -----------

// POST /modules
func (h *ModuleHandler) SaveModule(c *gin.Context) {
	var req saveModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Send(
			c,
			response.New(400).
				WithMessage("Invalid request payload").
				WithErrors(err.Error()).
				WithUserMessage(
					"Invalid input",
					"Please check the form fields and try again",
				),
		)
		return
	}

	var moduleId string
	if req.ID == "" {
		moduleId = service.SlugifyHeading(req.Title) // your slugify function
	}

	updatedBy := c.GetString("user")

	module, err := h.service.SaveModule(service.SaveModuleInput{
		ID:        moduleId,
		Title:     req.Title,
		Order:     req.Order,
		Surfaces:  req.Surfaces,
		Content:   req.Content,
		UpdatedBy: updatedBy,
	})

	if err != nil {
		response.Send(
			c,
			response.New(422).
				WithMessage(err.Error()).
				WithUserMessage(
					"Unable to save module",
					"Please verify the module details and try again",
				).
				WithTag("MODULE_SAVE"),
		)
		return
	}

	response.Send(
		c,
		response.New(200).
			WithMessage("Module saved successfully").
			WithData(moduleResponse{Module: module}),
	)
}

// GET /modules
func (h *ModuleHandler) ListModules(c *gin.Context) {
	modules, err := h.service.ListModules()

	if err != nil {
		response.Send(
			c,
			response.New(422).
				WithMessage("Something went wrong while fetching modules").
				WithData(nil),
		)
	}

	response.Send(
		c,
		response.New(200).
			WithMessage("Modules listed successfully").
			WithData(modules),
	)
}

// GET /modules/:id
func (h *ModuleHandler) GetLatest(c *gin.Context) {
	id := c.Param("id")

	module, err := h.service.GetLatest(id)
	if err != nil {
		response.Send(
			c,
			response.New(404).
				WithMessage(err.Error()).
				WithUserMessage(
					"Module not found",
					"The requested module does not exist",
				),
		)
		return
	}

	response.Send(
		c,
		response.New(200).
			WithMessage("Module fetched successfully").
			WithData(moduleResponse{Module: module}),
	)
}

// GET /modules/:id/history
func (h *ModuleHandler) GetHistory(c *gin.Context) {
	id := c.Param("id")

	versions, err := h.service.GetHistory(id)
	if err != nil {
		response.Send(
			c,
			response.New(404).
				WithMessage(err.Error()).
				WithUserMessage(
					"History not found",
					"No versions available for this module",
				),
		)
		return
	}

	response.Send(
		c,
		response.New(200).
			WithMessage("Module history fetched successfully").
			WithData(gin.H{
				"id":       id,
				"versions": versions,
			}),
	)
}

// GET /modules/:id/headers returns the table-of-contents for the module (markdown headings).
// Query "version" is optional: if omitted, latest version is used; if set, that version's headings are returned.
func (h *ModuleHandler) GetHeaders(c *gin.Context) {
	id := c.Param("id")
	versionStr := c.Query("version")

	var (
		headers []service.ModuleHeader
		err     error
	)
	if versionStr == "" {
		headers, err = h.service.GetHeaders(id)
	} else {
		v, parseErr := strconv.Atoi(versionStr)
		if parseErr != nil || v <= 0 {
			response.Send(
				c,
				response.New(400).
					WithMessage("invalid version parameter").
					WithUserMessage(
						"Invalid version",
						"Version must be a positive integer",
					),
			)
			return
		}

		headers, err = h.service.GetHeadersByVersion(id, v)
	}

	if err != nil {
		response.Send(
			c,
			response.New(404).
				WithMessage(err.Error()).
				WithUserMessage(
					"Module not found",
					"The requested module does not exist",
				),
		)
		return
	}

	response.Send(
		c,
		response.New(200).
			WithMessage("Module headers fetched successfully").
			WithData(moduleHeadersResponse{
				ID:      id,
				Headers: headers,
			}),
	)
}
